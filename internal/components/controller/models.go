package controller

import "github.com/google/uuid"

type ErrorResponse struct {
	Error string `json:"error"`
}

type CreateItemRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
}

type CreateItemResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
}

type ListItemsResponse []ItemResponse

type ItemResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
}
