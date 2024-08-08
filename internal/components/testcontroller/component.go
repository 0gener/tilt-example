package testcontroller

import (
	"github.com/0gener/go-service/components"
	httpComp "github.com/0gener/go-service/lib/http"
	"github.com/gin-gonic/gin"
	"net/http"
)

const ComponentName = "testcontroller"

type Component struct {
	components.BaseComponent

	httpComponent *httpComp.Component
}

func New() *Component {
	return &Component{
		BaseComponent: *components.NewBaseComponent(ComponentName),
	}
}

func (c *Component) Configure() error {
	var err error
	c.httpComponent, err = components.AsComponent[*httpComp.Component](c.Dependency(httpComp.ComponentName))
	if err != nil {
		return err
	}

	c.httpComponent.RegisterRoutes(
		httpComp.Route{
			RelativePath: "/v1/test",
			HTTPMethod:   http.MethodGet,
			Handlers:     []gin.HandlerFunc{handleTest},
		},
	)

	c.NotifyStatus(components.CONFIGURED)
	return nil
}

func handleTest(c *gin.Context) {
	c.Status(http.StatusOK)
}
