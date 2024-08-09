package controller

import (
	"context"
	"github.com/0gener/go-service/components"
	httpComp "github.com/0gener/go-service/lib/http"
	"github.com/0gener/tilt-example/internal/components/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

const ComponentName = "items_controller"

type Component struct {
	components.BaseComponent

	httpComponent *httpComp.Component
	repo          *repository.Component
}

func New() *Component {
	return &Component{
		BaseComponent: *components.NewBaseComponent(ComponentName),
	}
}

func (c *Component) Configure(_ context.Context) error {
	var err error
	c.httpComponent, err = components.AsComponent[*httpComp.Component](c.Dependency(httpComp.ComponentName))
	if err != nil {
		return err
	}

	c.httpComponent.RegisterRoutes(
		httpComp.Route{
			RelativePath: "/v1/items",
			HTTPMethod:   http.MethodPost,
			Handlers:     []gin.HandlerFunc{c.createItem},
		},
		httpComp.Route{
			RelativePath: "/v1/items",
			HTTPMethod:   http.MethodGet,
			Handlers:     []gin.HandlerFunc{c.getItems},
		},
	)

	c.repo, err = components.AsComponent[*repository.Component](c.Dependency(repository.ComponentName))
	if err != nil {
		return err
	}

	c.NotifyStatus(components.CONFIGURED)
	return nil
}

func (c *Component) createItem(ctx *gin.Context) {
	var req CreateItemRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, &ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	insertItem := repository.InsertItem{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
	}
	err := c.repo.InsertItem(context.Background(), insertItem)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, &ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, &CreateItemResponse{
		ID:          insertItem.ID,
		Name:        insertItem.Name,
		Description: insertItem.Description,
	})
}

func (c *Component) getItems(ctx *gin.Context) {
	items, err := c.repo.GetItems(context.Background())
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, MapItemsToResponse(items))
}
