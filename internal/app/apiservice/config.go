package apiservice

import (
	"github.com/0gener/go-service/utils"
)

const (
	EnvDatabaseConnectionString = "DATABASE_CONNECTION_STRING"
	EnvDatabaseMigrationsDir    = "DATABASE_MIGRATIONS_DIR"
)

type Config struct {
	Database DatabaseConfig
}

type DatabaseConfig struct {
	ConnectionString string
	MigrationsDir    string
}

func loadConfig() (*Config, error) {
	connectionString, err := utils.GetRequiredString(EnvDatabaseConnectionString)
	if err != nil {
		return nil, err
	}

	migrationsDir, err := utils.GetRequiredString(EnvDatabaseMigrationsDir)
	if err != nil {
		return nil, err
	}

	return &Config{
		Database: DatabaseConfig{
			ConnectionString: connectionString,
			MigrationsDir:    migrationsDir,
		},
	}, nil
}
