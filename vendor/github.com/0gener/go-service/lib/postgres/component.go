package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

func (c *Component) Configure(ctx context.Context) error {
	var err error

	if c.connectionString == "" {
		return ErrConnectionStringRequired
	}

	c.pool, err = pgxpool.New(ctx, c.connectionString)
	if err != nil {
		return err
	}

	c.NotifyStatus(components.CONFIGURED)
	return nil
}

func (c *Component) Start(ctx context.Context) error {
	err := c.retryPing(ctx)
	if err != nil {
		return err
	}

	c.Logger().Info("reached database successfully")

	err = c.runMigrations()
	if err != nil {
		return err
	}

	c.NotifyStatus(components.STARTED)
	return nil
}

func (c *Component) Shutdown(_ context.Context) error {
	if c.pool != nil {
		c.pool.Close()
	}
	c.NotifyStatus(components.STOPPED)
	return nil
}

func (c *Component) Pool() *pgxpool.Pool {
	return c.pool
}

func (c *Component) retryPing(ctx context.Context) error {
	return retry.Do(func() error {
		return c.pool.Ping(context.TODO())
	},
		retry.Context(ctx),
		retry.DelayType(retry.BackOffDelay),
		retry.Attempts(pingAttempts),
		retry.Delay(pingInitialDelay),
		retry.MaxDelay(pingMaxDelay),
		retry.LastErrorOnly(true),
		retry.OnRetry(c.onPingRetry),
	)
}

func (c *Component) onPingRetry(n uint, err error) {
	c.Logger().Warn(
		fmt.Sprintf("retrying ping attempt #%d for connection string: %s", n+1, c.connectionString),
		zap.Error(err),
	)
}

func (c *Component) runMigrations() error {
	if c.migrationsDir == "" {
		c.Logger().Warn("skipping migrations because migrations directory is not set")
		return nil
	}

	// Open a standard database connection using the native driver
	db, err := sql.Open("postgres", c.connectionString)
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
	source, err := (&file.File{}).Open("file://" + c.migrationsDir)
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

	c.Logger().Info("migrations successfully applied")

	return nil
}
