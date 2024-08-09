package testcontroller

import (
	"context"
	"fmt"
	"github.com/0gener/go-service/components"
	httpComp "github.com/0gener/go-service/lib/http"
	"github.com/0gener/tilt-example/internal/components/usersrepository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

const ComponentName = "test_controller"

type Component struct {
	components.BaseComponent

	httpComponent *httpComp.Component
	usersRepo     *usersrepository.Component
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
			RelativePath: "/v1/test",
			HTTPMethod:   http.MethodGet,
			Handlers:     []gin.HandlerFunc{c.handleTest},
		},
	)

	c.usersRepo, err = components.AsComponent[*usersrepository.Component](c.Dependency(usersrepository.ComponentName))
	if err != nil {
		return err
	}

	c.NotifyStatus(components.CONFIGURED)
	return nil
}

func (c *Component) handleTest(ctx *gin.Context) {
	err := c.usersRepo.InsertUser(context.Background(), usersrepository.User{
		Name:  uuid.NewString(),
		Email: fmt.Sprintf("%s@test.com", uuid.NewString()),
	})
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}
