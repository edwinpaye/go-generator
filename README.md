# ⚡ `kogen`: Enterprise-Grade Hexagonal Native Kotlin REST API Generator

[![Go Version](https://img.shields.io/badge/Go-1.27.0-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![Kotlin Version](https://img.shields.io/badge/Kotlin-1.9.22-7F52FF?style=flat-square&logo=kotlin)](https://kotlinlang.org/)
[![Architecture](https://img.shields.io/badge/Architecture-Hexagonal%20%2F%20Clean-FF6F00?style=flat-square)](https://en.wikipedia.org/wiki/Hexagonal_architecture_(software))
[![DBML Spec](https://img.shields.io/badge/DBML-%40dbml%2Fcore%3A10-2B3A42?style=flat-square)](https://dbml.dbdiagram.io/)
[![Native Compilation](https://img.shields.io/badge/GraalVM-Native%20Image-ED8B00?style=flat-square&logo=graalvm)](https://www.graalvm.org/)

`kogen` is an enterprise-grade Go (Golang) CLI tool engineered with **Modular Hexagonal Architecture (Ports & Adapters)**. It parses Database Markup Language (DBML v1/v2 `@dbml/core:10`) schemas and instantly generates a production-ready, high-performance **Kotlin REST API backend** compiled to native binary via **JetBrains / GraalVM Native Image**.

---

## 📐 System Architecture

### 1. `kogen` Golang CLI Architecture (Ports & Adapters)
`kogen` follows pure Hexagonal Architecture principles to decouple the core schema domain logic from parser implementations, code generator engines, configuration sources, and CLI commands:

```
                          ┌──────────────────────────────────────────────────┐
                          │                    CLI Adapter                   │
                          │        (Cobra: init, template, validate,         │
                          │             inspect, generate, version)          │
                          └────────────────────────┬─────────────────────────┘
                                                   │
                                                   ▼
┌───────────────────────┐  ┌────────────────────────────────────────────────┐  ┌───────────────────────┐
│     Parser Adapter    │  │                   Core Domain                  │  │   Generator Adapter   │
│ (DBML Lexer & Parser) │─►│ - DBMLDocument, Table, TablePartial, Enum      │◄─│ (Kotlin Code Engine:  │
│  - Partial Mixins     │  │ - TableGroup, Ref, Index, Project              │  │  Domain, App, Infra,  │
│  - Postgres Schemas   │  │ - GeneratorSpec, GeneratedFile                 │  │  OpenAPI, Migrations) │
└───────────────────────┘  └───────────────────────┬────────────────────────┘  └───────────────────────┘
                                                   │
                                                   ▼
                          ┌──────────────────────────────────────────────────┐
                          │               File System Writer                 │
                          │    (Disk Writer, Dry-Run, Force Overwrite)       │
                          └──────────────────────────────────────────────────┘
```

### 2. Generated Kotlin Backend Architecture
The generated Kotlin REST API application follows Clean Hexagonal Architecture:

```
project-root/
├── src/main/kotlin/com/example/api/
│   ├── domain/                         # Core Domain Layer (Zero Framework Dependencies)
│   │   ├── enum/                       # Domain Enums (UserStatus.kt, AccountType.kt)
│   │   ├── exception/                  # Domain Exceptions (EntityNotFoundException.kt)
│   │   ├── model/                      # Immutable Entities & Invariants (Users.kt)
│   │   └── port/                       # Repository Port Interfaces (UsersRepository.kt)
│   │
│   ├── application/                    # Application Layer (Use Cases & DTOs)
│   │   ├── dto/                        # Request/Response DTOs & PaginatedResponse.kt
│   │   ├── mapper/                     # DTO <-> Entity Mappers (UsersMapper.kt)
│   │   └── service/                    # Transactional Use Case Services (UsersService.kt)
│   │
│   └── infrastructure/                 # Infrastructure Layer (Adapters & Web)
│       ├── persistence/
│       │   ├── memory/                 # ConcurrentHashMap Repositories (In-Memory Testing)
│       │   └── sql/                    # JetBrains Exposed ORM Table Mappings & Repositories
│       └── web/
│           ├── error/                  # RFC 7807 ProblemDetails Error Handler
│           └── routes/                 # Ktor REST API Routes (UsersRoutes.kt)
│
├── src/test/kotlin/com/example/api/    # JUnit 5 & Ktor Integration Test Suite
├── src/main/resources/
│   ├── application.yaml                # Application & Server Config
│   └── db/migration/V1__init_schema.sql# Flyway/Liquibase DDL Migration Script
├── docs/openapi.yaml                   # OpenAPI 3.0 Specification
├── build.gradle.kts                    # Gradle Build & GraalVM Native Plugin Config
├── gradlew & gradlew.bat               # Gradle Executable Wrappers
└── Dockerfile                          # Multi-stage GraalVM Native Image Container
```

---

## 🌟 Key Features

- **Full DBML Feature Support (`@dbml/core:10`)**:
  - `Project`: Schema headers, notes, and metadata settings.
  - `TablePartial`: Reusable schema mixins (`mixin AuditFields`) merged at parse-time.
  - `Enum`: Custom database enums with optional PostgreSQL schema namespaces (`auth.UserStatus`).
  - `TableGroup`: Bounded context groupings mapped to sub-packages or domain modules.
  - `PostgreSQL Schemas`: Schema-qualified identifiers (`auth.users`, `public.orders`, `bank.accounts`).
  - `Ref`: Foreign Key relationships (1:1, 1:N, N:M) with `ON DELETE` and `ON UPDATE` triggers.
  - `Indexes`: Single, composite, and unique index definitions.
- **Multi-Database Support**:
  - Primary Drivers: PostgreSQL, MySQL, MariaDB, Oracle, MS SQL Server, SQLite.
  - **In-Memory Storage Mode**: Thread-safe `ConcurrentHashMap` repository implementation generated for every entity for instant, zero-dependency local execution and unit tests.
- **JetBrains Native / GraalVM Compilation**: Pre-configured `build.gradle.kts` with `org.graalvm.buildtools.native` producing ultra-fast, native binary executables.
- **Auto-Generated Documentation**: OpenAPI 3.0 specification (`docs/openapi.yaml`).
- **Auto-Generated Database Migrations**: Flyway/Liquibase compatible SQL DDL scripts (`V1__init_schema.sql`).
- **High Productivity Developer Tools**:
  - Interactive CLI commands (`init`, `template`, `validate`, `inspect`, `generate`).
  - Industry domain templates (`ecommerce`, `fintech`, `saas`, `healthcare`).

---

## 🚀 Installation & Build

### Prerequisites
- **Go**: 1.22+ (Go 1.27 supported)
- **Java**: JDK 17 or JDK 21 (for building generated Kotlin projects)

### Build `kogen` Binary from Source
```bash
git clone https://github.com/antigravity/kogen.git
cd go-generator
go build -o kogen.exe ./cmd/kogen
```

Verify installation:
```bash
.\kogen.exe version
# Output: kogen CLI v1.0.0 (Enterprise Kotlin Native Backend Generator)
```

---

## 📖 CLI Command Tutorial & Usage Guide

### 1. Initialize a New Workspace (`kogen init`)
Create a starter `schema.dbml` and `kogen.yaml` configuration file in the current directory:

```bash
.\kogen.exe init
```

*Output:*
```
✔ Created 'schema.dbml'
✔ Created 'kogen.yaml'
🚀 Workspace initialized! Run `kogen generate` to create your Kotlin REST API backend.
```

---

### 2. Generate Domain Templates (`kogen template`)
Instantly spin up industry-standard DBML schema templates for specific verticals:

```bash
# Available domains: ecommerce, fintech, saas, healthcare
.\kogen.exe template -d fintech
```

*Output:*
```
✨ Successfully generated 'schema-fintech.dbml' domain DBML template!
```

---

### 3. Inspect DBML Schema Structure (`kogen inspect`)
Visualize tables, mixins, enums, bounded contexts, and column definitions before generating code:

```bash
.\kogen.exe inspect -s schema-fintech.dbml
```

*Output:*
```
================================================================
📦 Project: FintechDomain
   Note: Fintech Core Banking & Digital Wallet DBML Schema
================================================================

📊 Summary:
  • Tables:       2
  • Partials:     1
  • Enums:        2
  • TableGroups:  1
  • References:   0

🏷️  Enums:
  - bank.AccountType (SAVINGS | CHECKING | INVESTMENT | CREDIT)
  - bank.TransactionStatus (PENDING | COMPLETED | FAILED | REVERSED)

🧩 TablePartials (Mixins):
  - Audit (2 columns)

📁 TableGroups (Bounded Contexts):
  - Banking: [bank.accounts, bank.transactions]

📋 Tables & Columns:
  • Table: bank.accounts
      - id                   uuid            [PK] [NOT NULL]
      - account_number       varchar(50)     [UNIQUE]
      - type                 bank.AccountType
      - balance              decimal(18,4)  
      - currency             varchar(3)     
      - created_at           timestamp      
      - updated_at           timestamp      
  • Table: bank.transactions
      - id                   uuid            [PK] [NOT NULL]
      - account_id           uuid           
      - amount               decimal(18,4)  
      - status               bank.TransactionStatus
      - reference_code       varchar(100)    [UNIQUE]
      - created_at           timestamp      
      - updated_at           timestamp      

================================================================
```

---

### 4. Validate DBML Schema Syntax (`kogen validate`)
Run semantic AST checks to verify foreign key integrity, missing enum definitions, and schema correctness:

```bash
.\kogen.exe validate -s schema.dbml
```

*Output:*
```
🔍 Validated DBML schema 'schema.dbml'
Summary: Project: ECommerceBackend, Tables: 2, Partials: 1, Enums: 1, TableGroups: 2, Refs: 1

✨ Schema is valid!
```

---

### 5. Generate Kotlin Backend Project (`kogen generate`)
Generate the complete, native-compilable Kotlin Hexagonal REST API backend project:

```bash
.\kogen.exe generate \
  --schema ./schema-fintech.dbml \
  --name fintech-banking-api \
  --package com.enterprise.fintech \
  --output ./output/fintech-banking-api \
  --db postgres \
  --port 8080 \
  --force
```

#### Command Line Flags:
| Flag | Short | Default | Description |
| :--- | :--- | :--- | :--- |
| `--schema` | `-s` | `./schema.dbml` | Path to source DBML schema file |
| `--output` | `-o` | `./output/sample-backend` | Target project directory |
| `--name` | `-n` | `sample-backend` | Target Kotlin project name |
| `--package` | `-p` | `com.example.api` | Kotlin base package name |
| `--db` | `-d` | `postgres` | Target database (`postgres`, `mysql`, `mariadb`, `sqlite`, `memory`, `oracle`, `mssql`) |
| `--port` | | `8080` | Server HTTP port |
| `--config` | `-c` | `""` | Path to `kogen.yaml` configuration file |
| `--dry-run` | | `false` | Preview generated files without writing to disk |
| `--force` | | `false` | Force overwrite existing target files |

---

## 📝 Writing DBML Schemas for `kogen`

Here is an example DBML file demonstrating all supported `@dbml/core:10` features:

```dbml
Project ECommerce {
  Note: 'Enterprise E-Commerce Microservice Schema'
}

// 1. Enum Definitions with optional Postgres Schema Scope
Enum auth.UserStatus {
  ACTIVE [note: 'Account is active']
  SUSPENDED [note: 'Suspended by system']
  PENDING
}

// 2. TablePartials (Mixins for Reusable Audit Fields)
TablePartial AuditFields {
  created_at timestamp [not null, default: 'now()']
  updated_at timestamp [not null]
}

// 3. Schema-qualified Tables with Mixins & Indexes
Table auth.users {
  id uuid [pk, increment]
  email varchar(255) [not null, unique]
  status auth.UserStatus [default: 'ACTIVE']
  mixin AuditFields

  indexes {
    email [unique, name: 'idx_user_email']
    (status, created_at) [name: 'idx_status_created']
  }
}

Table orders {
  id uuid [pk]
  user_id uuid [not null, ref: > auth.users.id]
  total_amount decimal(10,2) [not null]
  mixin AuditFields
}

// 4. Bounded Context Table Groups
TableGroup AuthDomain {
  auth.users
}

TableGroup OrderDomain {
  orders
}

// 5. Standalone Foreign Key References
Ref fk_orders_user: orders.user_id > auth.users.id [delete: cascade]
```

---

## 🏃 Running the Generated Kotlin Backend

Navigate to your generated backend project directory:
```bash
cd ./output/fintech-banking-api
```

### 1. Run Application Locally (In-Memory Zero-Dependency Mode)
```bash
./gradlew run
```
The server will start at `http://localhost:8080`.

### 2. Run Integration Tests
```bash
./gradlew test
```

### 3. Compile Native Executable (JetBrains Native / GraalVM)
```bash
./gradlew nativeCompile
```
Run the compiled native binary:
```bash
./build/native/nativeCompile/fintech-banking-api-native
```

### 4. Build Multi-Stage GraalVM Docker Container
```bash
docker build -t fintech-banking-api:latest .
docker run -p 8080:8080 fintech-banking-api:latest
```

---

## 🧪 Testing `kogen` CLI Engine

Run Golang unit tests for parser and generator adapters:
```bash
cd go-generator
go test ./... -v
```

## 📑 Documentation & Deep Dive Guides

- [Database Configuration Guide](docs/DATABASE_CONFIG_GUIDE.md) — Complete guide on configuring HikariCP, PostgreSQL SSL, MySQL, Oracle Wallet, MS SQL Azure, SQLite WAL mode, and environment variables.
- [OpenAPI Spec](docs/openapi.yaml) — Auto-generated OpenAPI 3.0 specification.
- [Database Migrations](src/main/resources/db/migration/V1__init_schema.sql) — Flyway/Liquibase SQL DDL script.

---

## 📜 License

Distributed under the MIT License. See `LICENSE` for details.
