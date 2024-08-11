package probes

import (
	"context"
	"fmt"
	"github.com/0gener/go-service/components"
	httpComponent "github.com/0gener/go-service/lib/http"
	"github.com/gin-gonic/gin"
	"net/http"
	"sync/atomic"
)

const (
	ComponentName = "probes"
)

type MonitoredComponent interface {
	components.Component
	Monitor(ctx context.Context) error
}

type Component struct {
	components.BaseComponent
	http *httpComponent.Component

	isLive                  atomic.Bool
	monitoredComponentNames []string
	monitoredComponents     []MonitoredComponent
}

func New() *Component {
	return &Component{
		BaseComponent: *components.NewBaseComponent(ComponentName),
	}
}

func (component *Component) Configure(_ context.Context) error {
	var err error
	component.http, err = components.AsComponent[*httpComponent.Component](component.Dependency(httpComponent.ComponentName))
	if err != nil {
		return err
	}

	if err = component.loadMonitoredComponents(); err != nil {
		return err
	}

	component.http.RegisterRoutes(
		httpComponent.Route{
			RelativePath: "/v1/live",
			HTTPMethod:   http.MethodGet,
			Handlers:     []gin.HandlerFunc{component.handleLive},
			IgnoreLogs:   true,
		},
		httpComponent.Route{
			RelativePath: "/v1/ready",
			HTTPMethod:   http.MethodGet,
			Handlers:     []gin.HandlerFunc{component.handleReady},
			IgnoreLogs:   true,
		},
	)

	component.NotifyStatus(components.CONFIGURED)
	return nil
}

func (component *Component) Start(_ context.Context) error {
	component.isLive.Store(true)
	component.NotifyStatus(components.STARTED)
	return nil
}

func (component *Component) Shutdown(_ context.Context) error {
	component.isLive.Store(false)
	component.NotifyStatus(components.STOPPED)
	return nil
}

func (component *Component) loadMonitoredComponents() error {
	for _, monitoredComponentName := range component.monitoredComponentNames {
		comp := component.Dependency(monitoredComponentName)

		if monitoredComponent, ok := comp.(MonitoredComponent); ok {
			component.monitoredComponents = append(component.monitoredComponents, monitoredComponent)
		} else {
			return fmt.Errorf("%s: %w", monitoredComponentName, ErrNotMonitoredComponent)
		}
	}

	return nil
}

func (component *Component) handleLive(c *gin.Context) {
	if !component.isLive.Load() {
		c.Status(http.StatusServiceUnavailable)
		return
	}

	c.Status(http.StatusOK)
}

func (component *Component) handleReady(c *gin.Context) {
	statusCode := http.StatusOK
	componentsMap := map[string]string{}

	for _, monitoredComponent := range component.monitoredComponents {
		componentStatus := "ready"
		if err := monitoredComponent.Monitor(c.Request.Context()); err != nil {
			componentStatus = "not_ready"
			statusCode = http.StatusServiceUnavailable
		}
		componentsMap[monitoredComponent.Name()] = componentStatus
	}

	c.JSON(statusCode, componentsMap)
}
