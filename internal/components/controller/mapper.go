package controller

import "github.com/0gener/tilt-example/internal/components/repository"

func MapItemsToResponse(items []repository.Item) ListItemsResponse {
	responseItems := make(ListItemsResponse, len(items))

	for i, item := range items {
		responseItems[i] = ItemResponse{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
		}
	}

	return responseItems
}
