package cli

import (
	"fmt"
	"os"

	"github.com/antigravity/kogen/internal/adapter/config"
	"github.com/antigravity/kogen/internal/adapter/dbml"
	"github.com/antigravity/kogen/internal/adapter/generator/kotlin"
	"github.com/antigravity/kogen/internal/adapter/writer"
	"github.com/antigravity/kogen/internal/core/service"
	"github.com/spf13/cobra"
)

var (
	cfgFile string

	// Core Services
	parserAdapter    *dbml.Parser
	generatorAdapter *kotlin.Generator
	writerAdapter    *writer.FSWriterAdapter
	configAdapter    *config.ConfigAdapter
	generatorService *service.GeneratorService

	RootCmd = &cobra.Command{
		Use:   "kogen",
		Short: "Enterprise-Grade Hexagonal Native Kotlin REST API Code Generator from DBML",
		Long: `kogen is a high-productivity CLI generator designed for enterprise Kotlin backends.
It parses DBML schemas (supporting DBML v1/v2, TablePartial mixins, Enums, Postgres schemas, TableGroups)
and outputs a complete, production-ready, GraalVM Native Image compiled Kotlin REST API project.`,
	}
)

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Initialize Hexagonal Adapters & Core Services
	parserAdapter = dbml.NewParser()
	generatorAdapter = kotlin.NewGenerator()
	writerAdapter = writer.NewFSWriterAdapter()
	configAdapter = config.NewConfigAdapter()

	generatorService = service.NewGeneratorService(parserAdapter, generatorAdapter, writerAdapter)

	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file path (default is ./kogen.yaml)")

	RootCmd.AddCommand(initCmd)
	RootCmd.AddCommand(generateCmd)
	RootCmd.AddCommand(validateCmd)
	RootCmd.AddCommand(inspectCmd)
	RootCmd.AddCommand(templateCmd)
	RootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of kogen CLI",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("kogen CLI v1.0.0 (Enterprise Kotlin Native Backend Generator)")
	},
}
