package usersrepository

import (
	"context"
	"fmt"
	"github.com/0gener/go-service/components"
	"github.com/0gener/go-service/lib/postgres"
)

const ComponentName = "users_repository"

type Component struct {
	components.BaseComponent

	postgresComponent *postgres.Component
}

func New() *Component {
	return &Component{
		BaseComponent: *components.NewBaseComponent(ComponentName),
	}
}

func (c *Component) Configure(_ context.Context) error {
	var err error
	c.postgresComponent, err = components.AsComponent[*postgres.Component](c.Dependency(postgres.ComponentName))
	if err != nil {
		return err
	}

	c.NotifyStatus(components.CONFIGURED)
	return nil
}

func (c *Component) InsertUser(ctx context.Context, user User) error {
	query := `
        INSERT INTO users (name, email, age)
        VALUES ($1, $2, $3)
    `

	_, err := c.postgresComponent.Pool().Exec(ctx, query, user.Name, user.Email, user.Age)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	return nil
}
