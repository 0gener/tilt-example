package http

import (
	"github.com/0gener/go-service/components"
	"github.com/gin-gonic/gin"
)

// WithServerHost configures the server host.
func WithServerHost(serverHost string) components.Option {
	return func(component components.Component) error {
		httpComponent, err := components.AsComponent[*Component](component)
		if err != nil {
			return err
		}

		httpComponent.serverHost = serverHost
		return nil
	}
}

// WithServerPort configures the server port.
func WithServerPort(serverPort int) components.Option {
	return func(component components.Component) error {
		httpComponent, err := components.AsComponent[*Component](component)
		if err != nil {
			return err
		}

		httpComponent.serverPort = serverPort
		return nil
	}
}

// WithRoute configures a route to be handled.
func WithRoute(route Route) components.Option {
	return func(component components.Component) error {
		httpComponent, err := components.AsComponent[*Component](component)
		if err != nil {
			return err
		}

		httpComponent.routes = append(httpComponent.routes, route)
		return nil
	}
}

// WithMiddlewares configures the middlewares.
func WithMiddlewares(middlewares gin.HandlersChain) components.Option {
	return func(component components.Component) error {
		httpComponent, err := components.AsComponent[*Component](component)
		if err != nil {
			return err
		}

		httpComponent.middlewares = append(httpComponent.middlewares, middlewares...)
		return nil
	}
}

// WithMetricsMiddleware overrides the default metrics middleware.
func WithMetricsMiddleware(middleware gin.HandlerFunc) components.Option {
	return func(component components.Component) error {
		httpComponent, err := components.AsComponent[*Component](component)
		if err != nil {
			return err
		}

		httpComponent.metricsMiddleware = middleware
		return nil
	}
}
