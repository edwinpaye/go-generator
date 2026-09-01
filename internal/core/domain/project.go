package domain

type DatabaseType string

const (
	DBPostgreSQL DatabaseType = "postgres"
	DBMySQL      DatabaseType = "mysql"
	DBMariaDB    DatabaseType = "mariadb"
	DBSQLite     DatabaseType = "sqlite"
	DBMemory     DatabaseType = "memory"
	DBOracle     DatabaseType = "oracle"
	DBMSSQL      DatabaseType = "mssql"
)

type ArchitectureStyle string

const (
	ArchHexagonalMultiModule  ArchitectureStyle = "multi-module"
	ArchHexagonalSingleModule ArchitectureStyle = "hexagonal"
)

type NativeTarget string

const (
	NativeGraalVM      NativeTarget = "graalvm"
	NativeKotlinNative NativeTarget = "kotlin-native"
)

// GeneratorSpec defines configuration for the target Kotlin project to generate
type GeneratorSpec struct {
	ProjectName       string            `mapstructure:"project_name" yaml:"project_name"`
	PackageName       string            `mapstructure:"package_name" yaml:"package_name"`
	TargetDir         string            `mapstructure:"target_dir" yaml:"target_dir"`
	Database          DatabaseType      `mapstructure:"database" yaml:"database"`
	Architecture      ArchitectureStyle `mapstructure:"architecture" yaml:"architecture"`
	NativeTarget      NativeTarget      `mapstructure:"native_target" yaml:"native_target"`
	IncludeInMemory   bool              `mapstructure:"include_in_memory" yaml:"include_in_memory"`
	IncludeOpenAPI    bool              `mapstructure:"include_openapi" yaml:"include_openapi"`
	IncludeDocker     bool              `mapstructure:"include_docker" yaml:"include_docker"`
	IncludeMigrations bool              `mapstructure:"include_migrations" yaml:"include_migrations"`
	ServerPort        int               `mapstructure:"server_port" yaml:"server_port"`
	DryRun            bool              `mapstructure:"dry_run" yaml:"dry_run"`
	ForceOverwrite    bool              `mapstructure:"force_overwrite" yaml:"force_overwrite"`
}

// GeneratedFile represents a code artifact emitted by the generator
type GeneratedFile struct {
	Path        string
	Content     string
	IsExecutable bool
}
