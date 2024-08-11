package repository

import "github.com/google/uuid"

type InsertItem struct {
	ID          uuid.UUID
	Name        string
	Description *string
}

type Item struct {
	ID          uuid.UUID
	Name        string
	Description *string
}
