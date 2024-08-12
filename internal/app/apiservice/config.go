package apiservice

import (
	"github.com/0gener/go-service/utils"
)

const (
	EnvAWSEndpoint              = "AWS_ENDPOINT"
	EnvEventsTopicARN           = "EVENTS_TOPIC_ARN"
	EnvDatabaseConnectionString = "DATABASE_CONNECTION_STRING"
	EnvDatabaseMigrationsDir    = "DATABASE_MIGRATIONS_DIR"
)

type Config struct {
	AWSEndpoint    string
	EventsTopicARN string
	Database       DatabaseConfig
}

type DatabaseConfig struct {
	ConnectionString string
	MigrationsDir    string
}

func loadConfig() (*Config, error) {
	eventsTopicArn, err := utils.GetRequiredString(EnvEventsTopicARN)
	if err != nil {
		return nil, err
	}

	connectionString, err := utils.GetRequiredString(EnvDatabaseConnectionString)
	if err != nil {
		return nil, err
	}

	migrationsDir, err := utils.GetRequiredString(EnvDatabaseMigrationsDir)
	if err != nil {
		return nil, err
	}

	return &Config{
		AWSEndpoint:    utils.GetStringOrDefault(EnvAWSEndpoint, ""),
		EventsTopicARN: eventsTopicArn,
		Database: DatabaseConfig{
			ConnectionString: connectionString,
			MigrationsDir:    migrationsDir,
		},
	}, nil
}
