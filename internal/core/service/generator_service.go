package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/antigravity/kogen/internal/core/domain"
	"github.com/antigravity/kogen/internal/core/ports"
)

type GeneratorService struct {
	parser    ports.ParserPort
	generator ports.CodeGeneratorPort
	writer    ports.FileWriterPort
	validator *SchemaValidationService
}

func NewGeneratorService(parser ports.ParserPort, generator ports.CodeGeneratorPort, writer ports.FileWriterPort) *GeneratorService {
	return &GeneratorService{
		parser:    parser,
		generator: generator,
		writer:    writer,
		validator: NewSchemaValidationService(),
	}
}

// GenerateFromDBML performs parsing, partial mixin expansion, semantic validation, code generation, and writing
func (s *GeneratorService) GenerateFromDBML(ctx context.Context, schemaPath string, spec *domain.GeneratorSpec) ([]domain.GeneratedFile, error) {
	doc, err := s.parser.ParseFile(ctx, schemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DBML file: %w", err)
	}

	// Expand TablePartials (mixins) into target tables
	s.expandPartials(doc)

	// Validate DBML model
	issues := s.validator.Validate(doc)
	hasError := false
	for _, issue := range issues {
		if issue.Level == "ERROR" {
			hasError = true
			fmt.Printf("[ERROR] %s\n", issue.Message)
		} else {
			fmt.Printf("[%s] %s\n", issue.Level, issue.Message)
		}
	}
	if hasError {
		return nil, fmt.Errorf("DBML validation failed with critical errors")
	}

	// Generate target project files
	files, err := s.generator.Generate(ctx, doc, spec)
	if err != nil {
		return nil, fmt.Errorf("failed to generate code: %w", err)
	}

	// Write files to target output directory
	if err := s.writer.WriteFiles(ctx, spec.TargetDir, files, spec.ForceOverwrite, spec.DryRun); err != nil {
		return nil, fmt.Errorf("failed to write generated files: %w", err)
	}

	return files, nil
}

// InspectDBML returns parsed DBML document and validation issues without generating files
func (s *GeneratorService) InspectDBML(ctx context.Context, schemaPath string) (*domain.DBMLDocument, []ValidationIssue, error) {
	doc, err := s.parser.ParseFile(ctx, schemaPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse DBML file: %w", err)
	}
	s.expandPartials(doc)
	issues := s.validator.Validate(doc)
	return doc, issues, nil
}

// expandPartials injects partial columns into tables that declare them or include them as mixins
func (s *GeneratorService) expandPartials(doc *domain.DBMLDocument) {
	partialMap := make(map[string]domain.TablePartial)
	for _, p := range doc.TablePartials {
		partialMap[strings.ToLower(p.Name)] = p
	}

	for i := range doc.Tables {
		tbl := &doc.Tables[i]
		for _, partialName := range tbl.PartialsUsed {
			if partial, exists := partialMap[strings.ToLower(partialName)]; exists {
				// Append partial columns if not already existing
				existingCols := make(map[string]bool)
				for _, c := range tbl.Columns {
					existingCols[strings.ToLower(c.Name)] = true
				}
				for _, pCol := range partial.Columns {
					if !existingCols[strings.ToLower(pCol.Name)] {
						tbl.Columns = append(tbl.Columns, pCol)
					}
				}
			}
		}
	}
}
