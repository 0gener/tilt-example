package postgres

import (
	"context"
	"database/sql"
	"errors"
	"github.com/0gener/go-service/components"
	"github.com/avast/retry-go"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"time"
)

const (
	ComponentName = "postgres"

	pingAttempts     = 8
	pingInitialDelay = 200 * time.Millisecond
	pingMaxDelay     = 3 * time.Second
)

var (
	ErrConnectionStringRequired = errors.New("connection string is required")
)

type Component struct {
	components.BaseComponent
	pool *pgxpool.Pool

	connectionString string
	migrationsDir    string
}

func New() *Component {
	return &Component{
		BaseComponent: *components.NewBaseComponent(ComponentName),
	}
}

func (component *Component) Configure(ctx context.Context) error {
	var err error

	if component.connectionString == "" {
		return ErrConnectionStringRequired
	}

	component.pool, err = pgxpool.New(ctx, component.connectionString)
	if err != nil {
		return err
	}

	component.NotifyStatus(components.CONFIGURED)
	return nil
}

func (component *Component) Start(ctx context.Context) error {
	err := component.retryPing(ctx)
	if err != nil {
		return err
	}

	component.Logger().Info("reached database successfully")

	err = component.runMigrations()
	if err != nil {
		return err
	}

	component.NotifyStatus(components.STARTED)
	return nil
}

func (component *Component) Shutdown(_ context.Context) error {
	if component.pool != nil {
		component.pool.Close()
	}
	component.NotifyStatus(components.STOPPED)
	return nil
}

func (component *Component) Monitor(ctx context.Context) error {
	return component.pool.Ping(ctx)
}

func (component *Component) Pool() *pgxpool.Pool {
	return component.pool
}

func (component *Component) retryPing(ctx context.Context) error {
	return retry.Do(func() error {
		return component.pool.Ping(context.TODO())
	},
		retry.Context(ctx),
		retry.DelayType(retry.BackOffDelay),
		retry.Attempts(pingAttempts),
		retry.Delay(pingInitialDelay),
		retry.MaxDelay(pingMaxDelay),
		retry.LastErrorOnly(true),
		retry.OnRetry(component.onPingRetry),
	)
}

func (component *Component) onPingRetry(n uint, err error) {
	component.Logger().Warn(
		"retrying ping attempt",
		zap.Error(err),
		zap.Uint("attempt", n+1),
		zap.String("connection_string", component.connectionString),
	)
}

func (component *Component) runMigrations() error {
	if component.migrationsDir == "" {
		component.Logger().Warn("skipping migrations because migrations directory is not set")
		return nil
	}

	// Open a standard database connection using the native driver
	db, err := sql.Open("postgres", component.connectionString)
	if err != nil {
		return err
	}
	defer db.Close()

	// Create an instance of the database driver
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	// Create an instance of the file source
	source, err := (&file.File{}).Open("file://" + component.migrationsDir)
	if err != nil {
		return err
	}

	// Use migrate.NewWithInstance to create the migration instance
	m, err := migrate.NewWithInstance("file", source, "postgres", driver)
	if err != nil {
		return err
	}

	// Run the migrations and handle any errors
	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	component.Logger().Info("migrations successfully applied")

	return nil
}
