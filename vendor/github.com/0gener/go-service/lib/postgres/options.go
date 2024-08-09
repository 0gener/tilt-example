package postgres

import "github.com/0gener/go-service/components"

// WithConnectionString configures the database connection string.
func WithConnectionString(connectionString string) components.Option {
	return func(component components.Component) error {
		psqlComponent, err := components.AsComponent[*Component](component)
		if err != nil {
			return err
		}

		psqlComponent.connectionString = connectionString
		return nil
	}
}

// WithMigrationsDir configures the database migrations directory.
func WithMigrationsDir(migrationsDir string) components.Option {
	return func(component components.Component) error {
		psqlComponent, err := components.AsComponent[*Component](component)
		if err != nil {
			return err
		}

		psqlComponent.migrationsDir = migrationsDir
		return nil
	}
}
