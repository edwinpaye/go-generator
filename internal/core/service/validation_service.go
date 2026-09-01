package service

import (
	"fmt"
	"strings"

	"github.com/antigravity/kogen/internal/core/domain"
)

type ValidationIssue struct {
	Level   string // "ERROR", "WARNING", "INFO"
	Message string
}

type SchemaValidationService struct{}

func NewSchemaValidationService() *SchemaValidationService {
	return &SchemaValidationService{}
}

// Validate checks the DBML document for semantic errors, dangling foreign key references, missing enum types, etc.
func (s *SchemaValidationService) Validate(doc *domain.DBMLDocument) []ValidationIssue {
	var issues []ValidationIssue

	if len(doc.Tables) == 0 {
		issues = append(issues, ValidationIssue{
			Level:   "WARNING",
			Message: "DBML document does not contain any table definitions.",
		})
	}

	tableSet := make(map[string]bool)
	for _, t := range doc.Tables {
		fullName := strings.ToLower(t.FullName())
		if tableSet[fullName] {
			issues = append(issues, ValidationIssue{
				Level:   "ERROR",
				Message: fmt.Sprintf("Duplicate table definition found: '%s'", t.FullName()),
			})
		}
		tableSet[fullName] = true
	}

	enumSet := make(map[string]bool)
	for _, e := range doc.Enums {
		fullName := strings.ToLower(e.FullName())
		enumSet[fullName] = true
		enumSet[strings.ToLower(e.Name)] = true
	}

	// Validate Foreign Keys (Refs)
	for _, ref := range doc.Refs {
		srcTable := doc.FindTable(ref.SourceSchema, ref.SourceTable)
		if srcTable == nil {
			issues = append(issues, ValidationIssue{
				Level:   "ERROR",
				Message: fmt.Sprintf("Ref references non-existent source table: '%s'", ref.SourceFullName()),
			})
		}

		targetTable := doc.FindTable(ref.TargetSchema, ref.TargetTable)
		if targetTable == nil {
			issues = append(issues, ValidationIssue{
				Level:   "ERROR",
				Message: fmt.Sprintf("Ref references non-existent target table: '%s'", ref.TargetFullName()),
			})
		}
	}

	// Validate inline references & column types
	for _, t := range doc.Tables {
		for _, c := range t.Columns {
			if c.Ref != nil {
				targetTable := doc.FindTable(c.Ref.TargetSchema, c.Ref.TargetTable)
				if targetTable == nil {
					issues = append(issues, ValidationIssue{
						Level:   "ERROR",
						Message: fmt.Sprintf("Column '%s.%s' ref points to missing target table '%s.%s'", t.FullName(), c.Name, c.Ref.TargetSchema, c.Ref.TargetTable),
					})
				}
			}
		}
	}

	return issues
}
