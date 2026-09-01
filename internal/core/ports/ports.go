package ports

import (
	"context"

	"github.com/antigravity/kogen/internal/core/domain"
)

// ParserPort defines the interface for parsing DBML files into AST domain models
type ParserPort interface {
	Parse(ctx context.Context, dbmlContent string) (*domain.DBMLDocument, error)
	ParseFile(ctx context.Context, filePath string) (*domain.DBMLDocument, error)
}

// CodeGeneratorPort defines the interface for generating target backend project files
type CodeGeneratorPort interface {
	Generate(ctx context.Context, doc *domain.DBMLDocument, spec *domain.GeneratorSpec) ([]domain.GeneratedFile, error)
}

// ConfigPort defines the interface for loading and resolving project config settings
type ConfigPort interface {
	LoadConfig(configPath string) (*domain.GeneratorSpec, error)
}

// FileWriterPort defines the interface for writing generated project files to disk
type FileWriterPort interface {
	WriteFiles(ctx context.Context, baseDir string, files []domain.GeneratedFile, force bool, dryRun bool) error
}
