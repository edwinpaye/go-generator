package cli

import (
	"context"
	"fmt"

	"github.com/antigravity/kogen/internal/core/domain"
	"github.com/spf13/cobra"
)

var (
	schemaPath     string
	targetDir      string
	projectName    string
	packageName    string
	dbTarget       string
	serverPort     int
	dryRun         bool
	forceOverwrite bool
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a full Kotlin Hexagonal REST API project from a DBML schema",
	RunE: func(cmd *cobra.Command, args []string) error {
		spec, err := configAdapter.LoadConfig(cfgFile)
		if err != nil {
			return err
		}

		// Flag Overrides
		if schemaPath != "" {
			// Spec update
		} else if cmd.Flag("schema").Value.String() == "" {
			schemaPath = "./schema.dbml"
		}

		if targetDir != "" {
			spec.TargetDir = targetDir
		}
		if projectName != "" {
			spec.ProjectName = projectName
		}
		if packageName != "" {
			spec.PackageName = packageName
		}
		if dbTarget != "" {
			spec.Database = domain.DatabaseType(dbTarget)
		}
		if serverPort != 0 {
			spec.ServerPort = serverPort
		}
		if cmd.Flags().Changed("dry-run") {
			spec.DryRun = dryRun
		}
		if cmd.Flags().Changed("force") {
			spec.ForceOverwrite = forceOverwrite
		}

		fmt.Printf("⚡ Generating Enterprise Kotlin Backend for DBML: '%s' -> Target: '%s'\n", schemaPath, spec.TargetDir)
		files, err := generatorService.GenerateFromDBML(context.Background(), schemaPath, spec)
		if err != nil {
			return err
		}

		fmt.Printf("✅ Successfully generated %d project files in '%s'!\n", len(files), spec.TargetDir)
		return nil
	},
}

func init() {
	generateCmd.Flags().StringVarP(&schemaPath, "schema", "s", "./schema.dbml", "Path to source DBML schema file")
	generateCmd.Flags().StringVarP(&targetDir, "output", "o", "./output/sample-backend", "Target output directory")
	generateCmd.Flags().StringVarP(&projectName, "name", "n", "sample-backend", "Project name")
	generateCmd.Flags().StringVarP(&packageName, "package", "p", "com.example.api", "Kotlin base package name")
	generateCmd.Flags().StringVarP(&dbTarget, "db", "d", "postgres", "Target database (postgres, mysql, mariadb, sqlite, memory, oracle, mssql)")
	generateCmd.Flags().IntVar(&serverPort, "port", 8080, "HTTP server port")
	generateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview generated files without writing to disk")
	generateCmd.Flags().BoolVar(&forceOverwrite, "force", false, "Force overwrite existing files in target output directory")
}
