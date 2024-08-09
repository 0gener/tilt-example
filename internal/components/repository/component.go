package repository

import (
	"context"
	"fmt"
	"github.com/0gener/go-service/components"
	"github.com/0gener/go-service/lib/postgres"
)

const ComponentName = "items_repository"

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

func (c *Component) InsertItem(ctx context.Context, item InsertItem) error {
	query := `
        INSERT INTO items (id, name, description)
        VALUES ($1, $2, $3)
    `

	_, err := c.postgresComponent.Pool().Exec(ctx, query, item.ID, item.Name, item.Description)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	return nil
}

func (c *Component) GetItems(ctx context.Context) ([]Item, error) {
	query := `
		SELECT id, name, description
		FROM items
	`

	rows, err := c.postgresComponent.Pool().Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve items: %w", err)
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item

		err = rows.Scan(&item.ID, &item.Name, &item.Description)
		if err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}

		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred during rows iteration: %w", err)
	}

	return items, nil
}
