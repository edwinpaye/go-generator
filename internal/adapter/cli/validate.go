package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate DBML schema syntax, foreign key references, and enum resolutions",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := schemaPath
		if path == "" {
			path = "./schema.dbml"
		}

		doc, issues, err := generatorService.InspectDBML(context.Background(), path)
		if err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}

		fmt.Printf("🔍 Validated DBML schema '%s'\n", path)
		fmt.Printf("Summary: %s\n\n", doc.Summary())

		errCount := 0
		for _, issue := range issues {
			if issue.Level == "ERROR" {
				errCount++
				fmt.Printf("❌ [ERROR] %s\n", issue.Message)
			} else {
				fmt.Printf("⚠️ [%s] %s\n", issue.Level, issue.Message)
			}
		}

		if errCount > 0 {
			return fmt.Errorf("schema validation failed with %d error(s)", errCount)
		}

		fmt.Println("✨ Schema is valid!")
		return nil
	},
}

func init() {
	validateCmd.Flags().StringVarP(&schemaPath, "schema", "s", "./schema.dbml", "Path to source DBML schema file")
}
