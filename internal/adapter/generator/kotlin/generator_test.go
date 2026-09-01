package kotlin

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/antigravity/kogen/internal/core/domain"
)

func TestGenerator_Generate(t *testing.T) {
	doc := &domain.DBMLDocument{
		Project: &domain.Project{Name: "TestBackend"},
		Enums: []domain.Enum{
			{Name: "Status", Values: []domain.EnumValue{{Name: "ACTIVE"}, {Name: "INACTIVE"}}},
		},
		Tables: []domain.Table{
			{
				Name: "users",
				Columns: []domain.Column{
					{Name: "id", Type: "uuid", PK: true, Nullable: false},
					{Name: "email", Type: "varchar(255)", Nullable: false, Unique: true},
					{Name: "status", Type: "Status", Nullable: true},
				},
			},
		},
	}

	spec := &domain.GeneratorSpec{
		ProjectName:       "TestBackend",
		PackageName:       "com.example.test",
		TargetDir:         "./output/test-backend",
		Database:          domain.DBPostgreSQL,
		Architecture:      domain.ArchHexagonalSingleModule,
		NativeTarget:      domain.NativeGraalVM,
		IncludeInMemory:   true,
		IncludeOpenAPI:    true,
		IncludeDocker:     true,
		IncludeMigrations: true,
		ServerPort:        8080,
	}

	gen := NewGenerator()
	files, err := gen.Generate(context.Background(), doc, spec)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(files) == 0 {
		t.Fatalf("Expected generated files, got 0")
	}

	// Verify key generated files exist
	foundApp := false
	foundDomain := false
	foundOpenAPI := false
	foundMigration := false

	for _, f := range files {
		p := filepath.ToSlash(f.Path)
		if p == "src/main/kotlin/com/example/test/Application.kt" {
			foundApp = true
		}
		if p == "src/main/kotlin/com/example/test/domain/model/Users.kt" {
			foundDomain = true
		}
		if p == "docs/openapi.yaml" {
			foundOpenAPI = true
		}
		if p == "src/main/resources/db/migration/V1__init_schema.sql" {
			foundMigration = true
		}
	}

	if !foundApp || !foundDomain || !foundOpenAPI || !foundMigration {
		t.Errorf("Missing expected files in generated output: app=%v, domain=%v, openapi=%v, migration=%v",
			foundApp, foundDomain, foundOpenAPI, foundMigration)
	}
}
