package kotlin

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/antigravity/kogen/internal/core/domain"
)

func ToPascalCase(s string) string {
	parts := splitWords(s)
	for i := range parts {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + strings.ToLower(parts[i][1:])
		}
	}
	return strings.Join(parts, "")
}

func ToCamelCase(s string) string {
	pascal := ToPascalCase(s)
	if len(pascal) == 0 {
		return ""
	}
	return strings.ToLower(pascal[:1]) + pascal[1:]
}

func ToSnakeCase(s string) string {
	parts := splitWords(s)
	for i := range parts {
		parts[i] = strings.ToLower(parts[i])
	}
	return strings.Join(parts, "_")
}

func ToKebabCase(s string) string {
	parts := splitWords(s)
	for i := range parts {
		parts[i] = strings.ToLower(parts[i])
	}
	return strings.Join(parts, "-")
}

func splitWords(s string) []string {
	var words []string
	var current []rune

	for i, r := range s {
		if r == '_' || r == '-' || r == '.' || r == '/' || r == ' ' {
			if len(current) > 0 {
				words = append(words, string(current))
				current = nil
			}
			continue
		}
		if unicode.IsUpper(r) && i > 0 {
			prev := rune(s[i-1])
			if !unicode.IsUpper(prev) && prev != '_' && prev != '-' && prev != '.' {
				if len(current) > 0 {
					words = append(words, string(current))
					current = nil
				}
			}
		}
		current = append(current, r)
	}
	if len(current) > 0 {
		words = append(words, string(current))
	}
	return words
}

func MapDBMLTypeToKotlin(col domain.Column, doc *domain.DBMLDocument) string {
	rawType := strings.ToLower(col.Type)
	if strings.Contains(rawType, "(") {
		rawType = rawType[:strings.Index(rawType, "(")]
	}
	rawType = strings.TrimSpace(rawType)

	// Check if type matches an Enum
	if doc != nil {
		enumObj := doc.FindEnum(col.Type)
		if enumObj != nil {
			return ToPascalCase(enumObj.Name)
		}
	}

	switch rawType {
	case "int", "integer", "int4", "smallint":
		return "Int"
	case "bigint", "int8", "long":
		return "Long"
	case "varchar", "text", "char", "string", "citext":
		return "String"
	case "boolean", "bool":
		return "Boolean"
	case "uuid":
		return "String" // or java.util.UUID
	case "timestamp", "datetime", "timestamptz":
		return "String" // ISO 8601 string representation for native compatibility
	case "date":
		return "String"
	case "decimal", "numeric", "float", "double", "real":
		return "Double"
	case "json", "jsonb":
		return "String"
	default:
		return "String"
	}
}

func MapDBMLTypeToSQL(col domain.Column, dbType domain.DatabaseType) string {
	rawType := strings.ToLower(col.Type)
	baseType := rawType
	param := ""
	if idx := strings.Index(rawType, "("); idx != -1 {
		baseType = rawType[:idx]
		param = rawType[idx:]
	}

	switch dbType {
	case domain.DBPostgreSQL:
		switch baseType {
		case "int", "integer":
			if col.Increment {
				return "SERIAL"
			}
			return "INTEGER"
		case "bigint":
			if col.Increment {
				return "BIGSERIAL"
			}
			return "BIGINT"
		case "varchar", "string":
			if param != "" {
				return "VARCHAR" + param
			}
			return "VARCHAR(255)"
		case "uuid":
			return "UUID"
		case "timestamp", "datetime":
			return "TIMESTAMP WITH TIME ZONE"
		case "boolean", "bool":
			return "BOOLEAN"
		default:
			return strings.ToUpper(col.Type)
		}

	case domain.DBSQLite, domain.DBMemory:
		switch baseType {
		case "int", "integer", "bigint":
			if col.PK && col.Increment {
				return "INTEGER PRIMARY KEY AUTOINCREMENT"
			}
			return "INTEGER"
		case "boolean", "bool":
			return "INTEGER"
		default:
			return "TEXT"
		}

	case domain.DBMySQL, domain.DBMariaDB:
		switch baseType {
		case "int", "integer":
			if col.Increment {
				return "INT AUTO_INCREMENT"
			}
			return "INT"
		case "varchar", "string":
			if param != "" {
				return "VARCHAR" + param
			}
			return "VARCHAR(255)"
		case "uuid":
			return "VARCHAR(36)"
		case "boolean", "bool":
			return "TINYINT(1)"
		default:
			return strings.ToUpper(col.Type)
		}

	default:
		return strings.ToUpper(col.Type)
	}
}

func GetPKColumn(t domain.Table) domain.Column {
	for _, c := range t.Columns {
		if c.PK {
			return c
		}
	}
	// Fallback to first column or synthesize id
	if len(t.Columns) > 0 {
		return t.Columns[0]
	}
	return domain.Column{Name: "id", Type: "uuid", PK: true}
}

func SanitizePackageName(pkg string) string {
	pkg = strings.TrimSpace(strings.ToLower(pkg))
	pkg = strings.ReplaceAll(pkg, "-", "_")
	if pkg == "" {
		return "com.example.api"
	}
	return pkg
}

func FormatDocComment(note string, prefix string) string {
	if strings.TrimSpace(note) == "" {
		return ""
	}
	lines := strings.Split(note, "\n")
	var result []string
	result = append(result, prefix+"/**")
	for _, l := range lines {
		result = append(result, fmt.Sprintf("%s * %s", prefix, strings.TrimSpace(l)))
	}
	result = append(result, prefix+" */")
	return strings.Join(result, "\n") + "\n"
}
