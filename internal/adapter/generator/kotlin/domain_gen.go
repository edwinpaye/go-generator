package kotlin

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/antigravity/kogen/internal/core/domain"
)

func generateDomainLayer(doc *domain.DBMLDocument, spec *domain.GeneratorSpec) ([]domain.GeneratedFile, error) {
	var files []domain.GeneratedFile
	pkg := SanitizePackageName(spec.PackageName)
	pkgPath := strings.ReplaceAll(pkg, ".", "/")

	// 1. Generate Domain Enums
	for _, enumObj := range doc.Enums {
		enumName := ToPascalCase(enumObj.Name)
		filePath := filepath.Join("src/main/kotlin", pkgPath, "domain/enum", enumName+".kt")

		var valLines []string
		for _, v := range enumObj.Values {
			valLines = append(valLines, fmt.Sprintf("    %s,", strings.ToUpper(v.Name)))
		}

		code := fmt.Sprintf(`package %s.domain.enum

/**
 * Domain Enum for %s
 */
enum class %s {
%s
}
`, pkg, enumName, enumName, strings.Join(valLines, "\n"))

		files = append(files, domain.GeneratedFile{Path: filePath, Content: code})
	}

	// 2. Generate Domain Exceptions
	excPath := filepath.Join("src/main/kotlin", pkgPath, "domain/exception/DomainExceptions.kt")
	excCode := fmt.Sprintf(`package %s.domain.exception

sealed class DomainException(message: String, cause: Throwable? = null) : RuntimeException(message, cause)

class EntityNotFoundException(val entityName: String, val id: Any) : 
    DomainException("$entityName with ID '$id' was not found.")

class DuplicateKeyException(val entityName: String, val keyName: String, val value: Any) : 
    DomainException("$entityName with $keyName '$value' already exists.")

class DomainValidationException(message: String) : DomainException(message)
`, pkg)
	files = append(files, domain.GeneratedFile{Path: excPath, Content: excCode})

	// 3. Generate Entities & Repository Ports for each Table
	for _, tbl := range doc.Tables {
		entityName := ToPascalCase(tbl.Name)
		pkCol := GetPKColumn(tbl)
		pkKotlinType := MapDBMLTypeToKotlin(pkCol, doc)

		// Entity Class
		entityPath := filepath.Join("src/main/kotlin", pkgPath, "domain/model", entityName+".kt")
		var fieldDecls []string
		for _, c := range tbl.Columns {
			ktType := MapDBMLTypeToKotlin(c, doc)
			fieldProp := ToCamelCase(c.Name)
			if c.Nullable && !c.PK {
				ktType += "?"
			}
			docComment := FormatDocComment(c.Note, "    ")
			fieldDecls = append(fieldDecls, fmt.Sprintf("%s    val %s: %s", docComment, fieldProp, ktType))
		}

		docComment := FormatDocComment(tbl.Note, "")
		enumImport := ""
		if len(doc.Enums) > 0 {
			enumImport = fmt.Sprintf("import %s.domain.enum.*\n\n", pkg)
		}

		entityCode := fmt.Sprintf(`package %s.domain.model

%s%sdata class %s(
%s
) {
    init {
        // Domain invariants & validation logic
    }
}
`, pkg, enumImport, docComment, entityName, strings.Join(fieldDecls, ",\n"))
		files = append(files, domain.GeneratedFile{Path: entityPath, Content: entityCode})

		// Repository Port Interface
		repoPath := filepath.Join("src/main/kotlin", pkgPath, "domain/port", entityName+"Repository.kt")
		repoCode := fmt.Sprintf(`package %s.domain.port

import %s.domain.model.%s

interface %sRepository {
    fun save(entity: %s): %s
    fun findById(id: %s): %s?
    fun findAll(page: Int = 1, size: Int = 20, sortBy: String? = null, sortOrder: String? = "asc"): List<%s>
    fun count(): Long
    fun update(id: %s, entity: %s): %s?
    fun deleteById(id: %s): Boolean
}
`, pkg, pkg, entityName, entityName, entityName, entityName, pkKotlinType, entityName, entityName, pkKotlinType, entityName, entityName, pkKotlinType)
		files = append(files, domain.GeneratedFile{Path: repoPath, Content: repoCode})
	}

	return files, nil
}
