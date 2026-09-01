package config

import (
	"os"

	"github.com/antigravity/kogen/internal/core/domain"
	"github.com/spf13/viper"
)

type ConfigAdapter struct{}

func NewConfigAdapter() *ConfigAdapter {
	return &ConfigAdapter{}
}

func (c *ConfigAdapter) LoadConfig(configPath string) (*domain.GeneratorSpec, error) {
	v := viper.New()

	// Default Spec Values
	spec := &domain.GeneratorSpec{
		ProjectName:       "sample-backend",
		PackageName:       "com.example.api",
		TargetDir:         "./output/sample-backend",
		Database:          domain.DBPostgreSQL,
		Architecture:      domain.ArchHexagonalSingleModule,
		NativeTarget:      domain.NativeGraalVM,
		IncludeInMemory:   true,
		IncludeOpenAPI:    true,
		IncludeDocker:     true,
		IncludeMigrations: true,
		ServerPort:        8080,
		DryRun:            false,
		ForceOverwrite:    false,
	}

	if configPath != "" {
		if _, err := os.Stat(configPath); err == nil {
			v.SetConfigFile(configPath)
			if err := v.ReadInConfig(); err == nil {
				if err := v.Unmarshal(spec); err != nil {
					return nil, err
				}
			}
		}
	}

	return spec, nil
}
