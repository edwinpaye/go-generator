package kotlin

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/antigravity/kogen/internal/core/domain"
)

func generateApplicationLayer(doc *domain.DBMLDocument, spec *domain.GeneratorSpec) ([]domain.GeneratedFile, error) {
	var files []domain.GeneratedFile
	pkg := SanitizePackageName(spec.PackageName)
	pkgPath := strings.ReplaceAll(pkg, ".", "/")

	// 0. Shared PaginatedResponse DTO
	paginatedDtoPath := filepath.Join("src/main/kotlin", pkgPath, "application/dto/PaginatedResponse.kt")
	paginatedDtoCode := fmt.Sprintf(`package %s.application.dto

import kotlinx.serialization.Serializable

@Serializable
data class PaginatedResponse<T>(
    val items: List<T>,
    val page: Int,
    val size: Int,
    val total: Long,
    val totalPages: Int
)
`, pkg)
	files = append(files, domain.GeneratedFile{Path: paginatedDtoPath, Content: paginatedDtoCode})

	for _, tbl := range doc.Tables {
		entityName := ToPascalCase(tbl.Name)
		pkCol := GetPKColumn(tbl)
		pkKotlinType := MapDBMLTypeToKotlin(pkCol, doc)

		// 1. DTOs (CreateRequest, UpdateRequest, ResponseDto)
		dtoPath := filepath.Join("src/main/kotlin", pkgPath, "application/dto", entityName+"Dtos.kt")
		var createProps []string
		var updateProps []string
		var respProps []string

		for _, c := range tbl.Columns {
			ktType := MapDBMLTypeToKotlin(c, doc)
			prop := ToCamelCase(c.Name)

			// Response includes all fields
			respType := ktType
			if c.Nullable && !c.PK {
				respType += "?"
			}
			respProps = append(respProps, fmt.Sprintf("    val %s: %s", prop, respType))

			// Create excludes auto-increment PKs or default timestamps if auto-generated
			if !c.PK && !c.Increment {
				createType := ktType
				if c.Nullable {
					createType += "? = null"
				}
				createProps = append(createProps, fmt.Sprintf("    val %s: %s", prop, createType))
				updateProps = append(updateProps, fmt.Sprintf("    val %s: %s? = null", prop, ktType))
			}
		}

		enumImport := ""
		if len(doc.Enums) > 0 {
			enumImport = fmt.Sprintf("import %s.domain.enum.*\n", pkg)
		}

		dtoCode := fmt.Sprintf(`package %s.application.dto

%simport kotlinx.serialization.Serializable

@Serializable
data class Create%sRequest(
%s
)

@Serializable
data class Update%sRequest(
%s
)

@Serializable
data class %sResponse(
%s
)
`, pkg, enumImport, entityName, strings.Join(createProps, ",\n"), entityName, strings.Join(updateProps, ",\n"), entityName, strings.Join(respProps, ",\n"))

		files = append(files, domain.GeneratedFile{Path: dtoPath, Content: dtoCode})

		// 2. Mapper
		mapperPath := filepath.Join("src/main/kotlin", pkgPath, "application/mapper", entityName+"Mapper.kt")
		var entityToDtoAssignments []string
		var dtoToEntityAssignments []string

		for _, c := range tbl.Columns {
			prop := ToCamelCase(c.Name)
			entityToDtoAssignments = append(entityToDtoAssignments, fmt.Sprintf("        %s = entity.%s", prop, prop))

			if c.PK {
				if c.Type == "uuid" {
					dtoToEntityAssignments = append(dtoToEntityAssignments, fmt.Sprintf("        %s = java.util.UUID.randomUUID().toString()", prop))
				} else if c.Type == "timestamp" || c.Type == "datetime" {
					dtoToEntityAssignments = append(dtoToEntityAssignments, fmt.Sprintf("        %s = java.time.Instant.now().toString()", prop))
				} else {
					dtoToEntityAssignments = append(dtoToEntityAssignments, fmt.Sprintf("        %s = 0", prop))
				}
			} else {
				dtoToEntityAssignments = append(dtoToEntityAssignments, fmt.Sprintf("        %s = req.%s", prop, prop))
			}
		}

		mapperCode := fmt.Sprintf(`package %s.application.mapper

import %s.application.dto.Create%sRequest
import %s.application.dto.%sResponse
import %s.domain.model.%s

object %sMapper {
    fun toResponse(entity: %s): %sResponse {
        return %sResponse(
%s
        )
    }

    fun toEntity(req: Create%sRequest): %s {
        return %s(
%s
        )
    }
}
`, pkg, pkg, entityName, pkg, entityName, pkg, entityName, entityName, entityName, entityName, entityName, strings.Join(entityToDtoAssignments, ",\n"), entityName, entityName, entityName, strings.Join(dtoToEntityAssignments, ",\n"))

		files = append(files, domain.GeneratedFile{Path: mapperPath, Content: mapperCode})

		// 3. Service / UseCase
		servicePath := filepath.Join("src/main/kotlin", pkgPath, "application/service", entityName+"Service.kt")
		serviceCode := fmt.Sprintf(`package %s.application.service

import %s.application.dto.Create%sRequest
import %s.application.dto.PaginatedResponse
import %s.application.dto.Update%sRequest
import %s.application.dto.%sResponse
import %s.application.mapper.%sMapper
import %s.domain.exception.EntityNotFoundException
import %s.domain.port.%sRepository
import kotlin.math.ceil

class %sService(
    private val repository: %sRepository
) {
    fun create(request: Create%sRequest): %sResponse {
        val entity = %sMapper.toEntity(request)
        val saved = repository.save(entity)
        return %sMapper.toResponse(saved)
    }

    fun getById(id: %s): %sResponse {
        val entity = repository.findById(id) ?: throw EntityNotFoundException("%s", id)
        return %sMapper.toResponse(entity)
    }

    fun list(page: Int = 1, size: Int = 20, sortBy: String? = null, sortOrder: String? = "asc"): PaginatedResponse<%sResponse> {
        val items = repository.findAll(page, size, sortBy, sortOrder).map { %sMapper.toResponse(it) }
        val total = repository.count()
        val totalPages = if (total == 0L) 1 else kotlin.math.ceil(total.toDouble() / size).toInt()
        return PaginatedResponse(
            items = items,
            page = page,
            size = size,
            total = total,
            totalPages = totalPages
        )
    }

    fun update(id: %s, request: Update%sRequest): %sResponse {
        val existing = repository.findById(id) ?: throw EntityNotFoundException("%s", id)
        // Apply partial updates or full updates
        val updated = repository.update(id, existing) ?: throw EntityNotFoundException("%s", id)
        return %sMapper.toResponse(updated)
    }

    fun delete(id: %s) {
        val deleted = repository.deleteById(id)
        if (!deleted) {
            throw EntityNotFoundException("%s", id)
        }
    }
}
`, pkg, pkg, entityName, pkg, pkg, entityName, pkg, entityName, pkg, entityName, pkg, pkg, entityName, entityName, entityName, entityName, entityName, entityName, entityName, pkKotlinType, entityName, entityName, entityName, entityName, entityName, pkKotlinType, entityName, entityName, entityName, entityName, entityName, pkKotlinType, entityName)

		files = append(files, domain.GeneratedFile{Path: servicePath, Content: serviceCode})
	}

	return files, nil
}
