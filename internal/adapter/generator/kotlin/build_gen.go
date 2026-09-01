package kotlin

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/antigravity/kogen/internal/core/domain"
)

func generateBuildAndBootstrap(doc *domain.DBMLDocument, spec *domain.GeneratorSpec) ([]domain.GeneratedFile, error) {
	var files []domain.GeneratedFile
	pkg := SanitizePackageName(spec.PackageName)
	pkgPath := strings.ReplaceAll(pkg, ".", "/")

	// 1. build.gradle.kts
	buildGradle := fmt.Sprintf(`plugins {
    kotlin("jvm") version "1.9.22"
    kotlin("plugin.serialization") version "1.9.22"
    id("org.graalvm.buildtools.native") version "0.10.1"
    id("application")
}

group = "%s"
version = "1.0.0"

java {
    sourceCompatibility = JavaVersion.VERSION_21
    targetCompatibility = JavaVersion.VERSION_21
}

repositories {
    mavenCentral()
}

dependencies {
    // Ktor Server Core & CIO Native Engine
    implementation("io.ktor:ktor-server-core-jvm:2.3.8")
    implementation("io.ktor:ktor-server-cio-jvm:2.3.8")
    implementation("io.ktor:ktor-server-content-negotiation-jvm:2.3.8")
    implementation("io.ktor:ktor-serialization-kotlinx-json-jvm:2.3.8")
    implementation("io.ktor:ktor-server-status-pages-jvm:2.3.8")
    implementation("io.ktor:ktor-server-cors-jvm:2.3.8")

    // JetBrains Exposed ORM & Database Driver Infrastructure
    implementation("org.jetbrains.exposed:exposed-core:0.47.0")
    implementation("org.jetbrains.exposed:exposed-dao:0.47.0")
    implementation("org.jetbrains.exposed:exposed-jdbc:0.47.0")
    implementation("com.zaxxer:HikariCP:5.1.0")
    implementation("org.xerial:sqlite-jdbc:3.45.1.0")
    implementation("org.postgresql:postgresql:42.7.1")

    // Logging & Diagnostics
    implementation("ch.qos.logback:logback-classic:1.4.14")

    // Testing
    testImplementation("kotlin:kotlin-test-junit5:1.9.22")
    testImplementation("io.ktor:ktor-server-tests-jvm:2.3.8")
}

application {
    mainClass.set("%s.ApplicationKt")
}

graalvmNative {
    binaries {
        named("main") {
            imageName.set("%s-native")
            mainClass.set("%s.ApplicationKt")
            verbose.set(true)
            buildArgs.add("-H:+ReportExceptionStackTraces")
        }
    }
}

tasks.withType<org.jetbrains.kotlin.gradle.tasks.KotlinCompile> {
    kotlinOptions {
        jvmTarget = "21"
        freeCompilerArgs = listOf("-Xjsr305=strict")
    }
}
`, pkg, pkg, spec.ProjectName, pkg)
	files = append(files, domain.GeneratedFile{Path: "build.gradle.kts", Content: buildGradle})

	// 2. settings.gradle.kts
	settingsGradle := fmt.Sprintf(`rootProject.name = "%s"
`, spec.ProjectName)
	files = append(files, domain.GeneratedFile{Path: "settings.gradle.kts", Content: settingsGradle})

	// 3. gradle.properties
	gradleProps := `org.gradle.jvmargs=-Xmx2048m -XX:MaxMetaspaceSize=512m
kotlin.code.style=official
`
	files = append(files, domain.GeneratedFile{Path: "gradle.properties", Content: gradleProps})

	// 4. application.yaml
	appYaml := fmt.Sprintf(`server:
  port: ${PORT:%d}

database:
  type: "%s"
  url: "${DATABASE_URL:jdbc:postgresql://localhost:5432/%s}"
  user: "${DB_USER:postgres}"
  password: "${DB_PASSWORD:postgres}"

logging:
  level: "INFO"
`, spec.ServerPort, string(spec.Database), strings.ToLower(spec.ProjectName))
	files = append(files, domain.GeneratedFile{Path: "src/main/resources/application.yaml", Content: appYaml})

	// 5. Main Application Entry Point Application.kt
	var routeRegistrations []string
	var serviceInstantiations []string

	for _, tbl := range doc.Tables {
		entityName := ToPascalCase(tbl.Name)
		serviceVar := ToCamelCase(entityName) + "Service"
		repoVar := ToCamelCase(entityName) + "Repo"

		serviceInstantiations = append(serviceInstantiations, fmt.Sprintf("    val %s = %s.infrastructure.persistence.memory.InMemory%sRepository()\n    val %s = %s.application.service.%sService(%s)",
			repoVar, pkg, entityName, serviceVar, pkg, entityName, repoVar))

		routeRegistrations = append(routeRegistrations, fmt.Sprintf("        configure%sRoutes(%s)", entityName, serviceVar))
	}

	var routeImports []string
	for _, tbl := range doc.Tables {
		entityName := ToPascalCase(tbl.Name)
		routeImports = append(routeImports, fmt.Sprintf("import %s.infrastructure.web.routes.configure%sRoutes", pkg, entityName))
	}

	appKtCode := fmt.Sprintf(`package %s

import %s.infrastructure.web.error.configureErrorHandling
%s
import io.ktor.serialization.kotlinx.json.*
import io.ktor.server.application.*
import io.ktor.server.cio.*
import io.ktor.server.engine.*
import io.ktor.server.plugins.contentnegotiation.*
import io.ktor.server.plugins.statuspages.*
import io.ktor.server.routing.*
import kotlinx.serialization.json.Json

fun main() {
    val port = System.getenv("PORT")?.toIntOrNull() ?: %d
    println("🚀 Starting Native Kotlin Hexagonal REST API on port $port...")

    embeddedServer(CIO, port = port) {
        install(ContentNegotiation) {
            json(Json {
                prettyPrint = true
                isLenient = true
                ignoreUnknownKeys = true
            })
        }

        install(StatusPages) {
            configureErrorHandling()
        }

%s

        routing {
%s
        }
    }.start(wait = true)
}
`, pkg, pkg, strings.Join(routeImports, "\n"), spec.ServerPort, strings.Join(serviceInstantiations, "\n"), strings.Join(routeRegistrations, "\n"))

	appKtPath := filepath.Join("src/main/kotlin", pkgPath, "Application.kt")
	files = append(files, domain.GeneratedFile{Path: appKtPath, Content: appKtCode})

	// 6. Dockerfile (GraalVM Native Multi-Stage Build)
	dockerfile := fmt.Sprintf(`FROM ghcr.io/graalvm/native-image-community:21 AS builder
WORKDIR /app
COPY . .
RUN ./gradlew nativeCompile --no-daemon

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /app/build/native/nativeCompile/%s-native /app/server
EXPOSE %d
ENV PORT=%d
ENTRYPOINT ["/app/server"]
`, spec.ProjectName, spec.ServerPort, spec.ServerPort)
	files = append(files, domain.GeneratedFile{Path: "Dockerfile", Content: dockerfile})

	// 7. Gradle Wrapper Script (gradlew)
	gradlewScript := `#!/usr/bin/env sh
exec gradle "$@"
`
	files = append(files, domain.GeneratedFile{Path: "gradlew", Content: gradlewScript, IsExecutable: true})

	// 8. Gradle Wrapper Windows Batch (gradlew.bat)
	gradlewBat := `@rem Gradle Wrapper script for Windows
@if "%DEBUG%" == "" @echo off
gradle %*
`
	files = append(files, domain.GeneratedFile{Path: "gradlew.bat", Content: gradlewBat, IsExecutable: true})

	// 9. gradle-wrapper.properties
	wrapperProps := `distributionBase=GRADLE_USER_HOME
distributionPath=wrapper/dists
distributionUrl=https\://services.gradle.org/distributions/gradle-8.5-bin.zip
networkTimeout=10000
validateDistributionUrl=true
zipStoreBase=GRADLE_USER_HOME
zipStorePath=wrapper/dists
`
	files = append(files, domain.GeneratedFile{Path: "gradle/wrapper/gradle-wrapper.properties", Content: wrapperProps})

	// 10. Integration Test Suite ApplicationTest.kt
	firstTable := "users"
	if len(doc.Tables) > 0 {
		firstTable = ToKebabCase(doc.Tables[0].Name)
	}

	testCode := fmt.Sprintf(`package %s

import io.ktor.client.request.*
import io.ktor.client.statement.*
import io.ktor.http.*
import io.ktor.server.testing.*
import kotlin.test.*

class ApplicationTest {
    @Test
    fun testHealthAndRoutes() = testApplication {
        client.get("/api/v1/%s").apply {
            assertEquals(HttpStatusCode.OK, status)
        }
    }
}
`, pkg, firstTable)

	testPath := filepath.Join("src/test/kotlin", pkgPath, "ApplicationTest.kt")
	files = append(files, domain.GeneratedFile{Path: testPath, Content: testCode})

	// 11. README.md
	readme := fmt.Sprintf(`# %s - Native Kotlin Hexagonal REST API

This repository contains a high-performance Enterprise-Grade Kotlin REST API backend generated by **kogen** CLI.

## Architecture
- **Hexagonal Architecture (Ports & Adapters)** / Clean Architecture
- **Domain Layer**: Immutable Entities, Value Objects, Domain Enums, Repository Ports, Invariants.
- **Application Layer**: Use Cases / Services, Request & Response DTOs, Mappers.
- **Infrastructure Layer**: In-Memory Repositories for zero-dependency testing, SQL persistence adapters, Ktor REST routes, RFC 7807 Exception Handlers.
- **Native Binary Compilation**: Native executable target via GraalVM Native Image plugin.

## Getting Started

### Local Development (In-Memory Mode / Zero-Dependency)
Run the application locally:
`+"```bash"+`
./gradlew run
`+"```"+`
The REST API will be live at `+"`http://localhost:%d`"+`.

### Run Automated Tests
`+"```bash"+`
./gradlew test
`+"```"+`

### Compile Native Executable (JetBrains Native / GraalVM)
`+"```bash"+`
./gradlew nativeCompile
`+"```"+`
Run native binary:
`+"```bash"+`
./build/native/nativeCompile/%s-native
`+"```"+`

### Build Docker Native Container
`+"```bash"+`
docker build -t %s:latest .
docker run -p %d:%d %s:latest
`+"```"+`

### Documentation
- OpenAPI Spec: `+"`docs/openapi.yaml`"+`
- Database Migrations: `+"`src/main/resources/db/migration/V1__init_schema.sql`"+`
`, spec.ProjectName, spec.ServerPort, spec.ProjectName, spec.ProjectName, spec.ServerPort, spec.ServerPort, spec.ProjectName)

	files = append(files, domain.GeneratedFile{Path: "README.md", Content: readme})

	return files, nil
}
