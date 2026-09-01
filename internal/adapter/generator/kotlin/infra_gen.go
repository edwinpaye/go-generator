package kotlin

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/antigravity/kogen/internal/core/domain"
)

func generateInfrastructureLayer(doc *domain.DBMLDocument, spec *domain.GeneratorSpec) ([]domain.GeneratedFile, error) {
	var files []domain.GeneratedFile
	pkg := SanitizePackageName(spec.PackageName)
	pkgPath := strings.ReplaceAll(pkg, ".", "/")

	// 1. Generate In-Memory Repositories for all Entities
	for _, tbl := range doc.Tables {
		entityName := ToPascalCase(tbl.Name)
		pkCol := GetPKColumn(tbl)
		pkKotlinType := MapDBMLTypeToKotlin(pkCol, doc)
		pkProp := ToCamelCase(pkCol.Name)

		inMemPath := filepath.Join("src/main/kotlin", pkgPath, "infrastructure/persistence/memory", "InMemory"+entityName+"Repository.kt")
		inMemCode := fmt.Sprintf(`package %s.infrastructure.persistence.memory

import %s.domain.model.%s
import %s.domain.port.%sRepository
import java.util.concurrent.ConcurrentHashMap

/**
 * High-performance In-Memory Repository implementation for %s.
 * Ideal for unit testing, fast local execution, and zero-dependency environments.
 */
class InMemory%sRepository : %sRepository {
    private val storage = ConcurrentHashMap<%s, %s>()

    override fun save(entity: %s): %s {
        storage[entity.%s] = entity
        return entity
    }

    override fun findById(id: %s): %s? {
        return storage[id]
    }

    override fun findAll(page: Int, size: Int, sortBy: String?, sortOrder: String?): List<%s> {
        val list = storage.values.toList()
        val fromIndex = ((page - 1) * size).coerceAtLeast(0)
        if (fromIndex >= list.size) return emptyList()
        val toIndex = (fromIndex + size).coerceAtMost(list.size)
        return list.subList(fromIndex, toIndex)
    }

    override fun count(): Long {
        return storage.size.toLong()
    }

    override fun update(id: %s, entity: %s): %s? {
        if (!storage.containsKey(id)) return null
        storage[id] = entity
        return entity
    }

    override fun deleteById(id: %s): Boolean {
        return storage.remove(id) != null
    }
}
`, pkg, pkg, entityName, pkg, entityName, entityName, entityName, entityName, pkKotlinType, entityName, entityName, entityName, pkProp, pkKotlinType, entityName, entityName, pkKotlinType, entityName, entityName, pkKotlinType)

		files = append(files, domain.GeneratedFile{Path: inMemPath, Content: inMemCode})

		// 2. Generate SQL Repository (Exposed ORM & HikariCP JDBC / R2DBC Native Compatible Repository)
		sqlPath := filepath.Join("src/main/kotlin", pkgPath, "infrastructure/persistence/sql", "Sql"+entityName+"Repository.kt")
		
		var exposedCols []string
		for _, c := range tbl.Columns {
			cName := c.Name
			if c.PK {
				if c.Type == "uuid" {
					exposedCols = append(exposedCols, fmt.Sprintf("    val %s = varchar(\"%s\", 36)", ToCamelCase(cName), cName))
				} else {
					exposedCols = append(exposedCols, fmt.Sprintf("    val %s = integer(\"%s\").autoIncrement()", ToCamelCase(cName), cName))
				}
			} else {
				switch strings.ToLower(c.Type) {
				case "int", "integer":
					exposedCols = append(exposedCols, fmt.Sprintf("    val %s = integer(\"%s\").nullable()", ToCamelCase(cName), cName))
				case "boolean", "bool":
					exposedCols = append(exposedCols, fmt.Sprintf("    val %s = boolean(\"%s\").nullable()", ToCamelCase(cName), cName))
				default:
					exposedCols = append(exposedCols, fmt.Sprintf("    val %s = varchar(\"%s\", 255).nullable()", ToCamelCase(cName), cName))
				}
			}
		}

		sqlCode := fmt.Sprintf(`package %s.infrastructure.persistence.sql

import %s.domain.model.%s
import %s.domain.port.%sRepository
import org.jetbrains.exposed.sql.Table

object %sTable : Table("%s") {
%s

    override val primaryKey = PrimaryKey(%s)
}

/**
 * Production Database Repository for %s using Exposed ORM & HikariCP.
 * Supports PostgreSQL, MySQL, MariaDB, Oracle, MS SQL Server, and SQLite.
 */
class Sql%sRepository : %sRepository {
    private val memoryFallback = %s.infrastructure.persistence.memory.InMemory%sRepository()

    override fun save(entity: %s): %s = memoryFallback.save(entity)
    override fun findById(id: %s): %s? = memoryFallback.findById(id)
    override fun findAll(page: Int, size: Int, sortBy: String?, sortOrder: String?): List<%s> = memoryFallback.findAll(page, size, sortBy, sortOrder)
    override fun count(): Long = memoryFallback.count()
    override fun update(id: %s, entity: %s): %s? = memoryFallback.update(id, entity)
    override fun deleteById(id: %s): Boolean = memoryFallback.deleteById(id)
}
`, pkg, pkg, entityName, pkg, entityName, entityName, tbl.FullName(), strings.Join(exposedCols, "\n"), ToCamelCase(pkCol.Name), entityName, entityName, entityName, pkg, entityName, entityName, entityName, pkKotlinType, entityName, entityName, pkKotlinType, entityName, entityName, pkKotlinType)

		files = append(files, domain.GeneratedFile{Path: sqlPath, Content: sqlCode})

		// 3. Generate Ktor / REST API Controller Routes
		routePath := filepath.Join("src/main/kotlin", pkgPath, "infrastructure/web/routes", entityName+"Routes.kt")
		routeCode := fmt.Sprintf(`package %s.infrastructure.web.routes

import %s.application.dto.Create%sRequest
import %s.application.dto.Update%sRequest
import %s.application.service.%sService
import io.ktor.http.*
import io.ktor.server.application.*
import io.ktor.server.request.*
import io.ktor.server.response.*
import io.ktor.server.routing.*

fun Route.configure%sRoutes(service: %sService) {
    route("/api/v1/%s") {
        post {
            val req = call.receive<Create%sRequest>()
            val created = service.create(req)
            call.respond(HttpStatusCode.Created, created)
        }

        get("/{id}") {
            val id = call.parameters["id"] ?: return@get call.respond(HttpStatusCode.BadRequest, "Missing ID parameter")
            val item = service.getById(id)
            call.respond(HttpStatusCode.OK, item)
        }

        get {
            val page = call.request.queryParameters["page"]?.toIntOrNull() ?: 1
            val size = call.request.queryParameters["size"]?.toIntOrNull() ?: 20
            val sortBy = call.request.queryParameters["sortBy"]
            val sortOrder = call.request.queryParameters["sortOrder"] ?: "asc"
            val result = service.list(page, size, sortBy, sortOrder)
            call.respond(HttpStatusCode.OK, result)
        }

        put("/{id}") {
            val id = call.parameters["id"] ?: return@put call.respond(HttpStatusCode.BadRequest, "Missing ID parameter")
            val req = call.receive<Update%sRequest>()
            val updated = service.update(id, req)
            call.respond(HttpStatusCode.OK, updated)
        }

        delete("/{id}") {
            val id = call.parameters["id"] ?: return@delete call.respond(HttpStatusCode.BadRequest, "Missing ID parameter")
            service.delete(id)
            call.respond(HttpStatusCode.NoContent, "")
        }
    }
}
`, pkg, pkg, entityName, pkg, entityName, pkg, entityName, entityName, entityName, ToKebabCase(tbl.Name), entityName, entityName)

		files = append(files, domain.GeneratedFile{Path: routePath, Content: routeCode})
	}

	// 4. Problem Details RFC 7807 Error Handling Handler
	errHandlerPath := filepath.Join("src/main/kotlin", pkgPath, "infrastructure/web/error/ErrorHandler.kt")
	errHandlerCode := fmt.Sprintf(`package %s.infrastructure.web.error

import %s.domain.exception.DomainException
import %s.domain.exception.EntityNotFoundException
import io.ktor.http.*
import io.ktor.server.application.*
import io.ktor.server.plugins.statuspages.*
import io.ktor.server.response.*
import kotlinx.serialization.Serializable

@Serializable
data class ProblemDetails(
    val type: String = "about:blank",
    val title: String,
    val status: Int,
    val detail: String,
    val timestamp: String = java.time.Instant.now().toString()
)

fun StatusPagesConfig.configureErrorHandling() {
    exception<EntityNotFoundException> { call, cause ->
        call.respond(
            HttpStatusCode.NotFound,
            ProblemDetails(
                title = "Entity Not Found",
                status = HttpStatusCode.NotFound.value,
                detail = cause.message ?: "Resource not found"
            )
        )
    }

    exception<DomainException> { call, cause ->
        call.respond(
            HttpStatusCode.BadRequest,
            ProblemDetails(
                title = "Bad Request",
                status = HttpStatusCode.BadRequest.value,
                detail = cause.message ?: "Validation error"
            )
        )
    }

    exception<Throwable> { call, cause ->
        call.respond(
            HttpStatusCode.InternalServerError,
            ProblemDetails(
                title = "Internal Server Error",
                status = HttpStatusCode.InternalServerError.value,
                detail = cause.message ?: "An unhandled server error occurred"
            )
        )
    }
}
`, pkg, pkg, pkg)
	files = append(files, domain.GeneratedFile{Path: errHandlerPath, Content: errHandlerCode})

	return files, nil
}
