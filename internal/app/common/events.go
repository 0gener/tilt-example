package common

import "github.com/google/uuid"

type ItemCreatedEvent struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
}
