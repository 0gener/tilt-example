package controller

import (
	"context"
	"github.com/0gener/go-service/components"
	httpComp "github.com/0gener/go-service/lib/http"
	"github.com/0gener/tilt-example/internal/app/apiservice/components/eventpublisher"
	"github.com/0gener/tilt-example/internal/app/apiservice/components/repository"
	"github.com/0gener/tilt-example/internal/app/common"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

const ComponentName = "items_controller"

type Component struct {
	components.BaseComponent

	http            *httpComp.Component
	repo            *repository.Component
	eventsPublisher *eventpublisher.Component
}

func New() *Component {
	return &Component{
		BaseComponent: *components.NewBaseComponent(ComponentName),
	}
}

func (component *Component) Configure(_ context.Context) error {
	var err error
	component.http, err = components.AsComponent[*httpComp.Component](component.Dependency(httpComp.ComponentName))
	if err != nil {
		return err
	}

	component.http.RegisterRoutes(
		httpComp.Route{
			RelativePath: "/v1/items",
			HTTPMethod:   http.MethodPost,
			Handlers:     []gin.HandlerFunc{component.createItem},
		},
		httpComp.Route{
			RelativePath: "/v1/items",
			HTTPMethod:   http.MethodGet,
			Handlers:     []gin.HandlerFunc{component.getItems},
		},
	)

	component.repo, err = components.AsComponent[*repository.Component](component.Dependency(repository.ComponentName))
	if err != nil {
		return err
	}

	component.eventsPublisher, err = components.AsComponent[*eventpublisher.Component](component.Dependency(eventpublisher.ComponentName))
	if err != nil {
		return err
	}

	component.NotifyStatus(components.CONFIGURED)
	return nil
}

func (component *Component) createItem(c *gin.Context) {
	ctx := context.Background()
	var req CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, &ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	insertItem := repository.InsertItem{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
	}
	err := component.repo.InsertItem(ctx, insertItem)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	err = component.eventsPublisher.ItemCreatedEvent(ctx, common.ItemCreatedEvent{
		ID:          insertItem.ID,
		Name:        insertItem.Name,
		Description: insertItem.Description,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, &ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, &CreateItemResponse{
		ID:          insertItem.ID,
		Name:        insertItem.Name,
		Description: insertItem.Description,
	})
}

func (component *Component) getItems(c *gin.Context) {
	items, err := component.repo.GetItems(context.Background())
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, MapItemsToResponse(items))
}
