package kotlin

import (
	"context"

	"github.com/antigravity/kogen/internal/core/domain"
)

type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) Generate(ctx context.Context, doc *domain.DBMLDocument, spec *domain.GeneratorSpec) ([]domain.GeneratedFile, error) {
	var allFiles []domain.GeneratedFile

	// 1. Generate Domain Layer
	domainFiles, err := generateDomainLayer(doc, spec)
	if err != nil {
		return nil, err
	}
	allFiles = append(allFiles, domainFiles...)

	// 2. Generate Application Layer
	appFiles, err := generateApplicationLayer(doc, spec)
	if err != nil {
		return nil, err
	}
	allFiles = append(allFiles, appFiles...)

	// 3. Generate Infrastructure Layer
	infraFiles, err := generateInfrastructureLayer(doc, spec)
	if err != nil {
		return nil, err
	}
	allFiles = append(allFiles, infraFiles...)

	// 4. Generate OpenAPI Specification
	if spec.IncludeOpenAPI {
		openApiFiles, err := generateOpenAPISpec(doc, spec)
		if err != nil {
			return nil, err
		}
		allFiles = append(allFiles, openApiFiles...)
	}

	// 5. Generate Database Migrations
	if spec.IncludeMigrations {
		migrationFiles, err := generateMigrations(doc, spec)
		if err != nil {
			return nil, err
		}
		allFiles = append(allFiles, migrationFiles...)
	}

	// 6. Generate Build & Bootstrap Infrastructure
	buildFiles, err := generateBuildAndBootstrap(doc, spec)
	if err != nil {
		return nil, err
	}
	allFiles = append(allFiles, buildFiles...)

	return allFiles, nil
}
