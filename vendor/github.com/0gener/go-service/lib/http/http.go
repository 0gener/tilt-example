package http

import (
	"context"
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	ginprometheus "github.com/zsais/go-gin-prometheus"
	"net"
	"net/http"
	"time"

	"github.com/0gener/go-service/components"
	"github.com/gin-gonic/gin"
)

const (
	ComponentName = "http"

	metricsPath       = "/metrics"
	defaultServerHost = ""
	defaultServerPort = 8080
)

type Component struct {
	components.BaseComponent

	router     *gin.Engine
	server     *http.Server
	serverHost string
	serverPort int

	routes            []Route
	middlewares       gin.HandlersChain
	metricsMiddleware gin.HandlerFunc
}

type Route struct {
	RelativePath string
	HTTPMethod   string
	Handlers     []gin.HandlerFunc
}

func New() *Component {
	gin.SetMode(gin.ReleaseMode)

	return &Component{
		BaseComponent: *components.NewBaseComponent(ComponentName),
		router:        gin.New(),
	}
}

func (c *Component) DefaultOptions() []components.Option {
	return []components.Option{
		WithServerHost(defaultServerHost),
		WithServerPort(defaultServerPort),
		WithMetricsMiddleware(
			ginprometheus.NewPrometheus("gin").HandlerFunc(),
		),
		WithMiddlewares([]gin.HandlerFunc{
			gin.Recovery(),
		}),
		WithRoute(Route{
			RelativePath: metricsPath,
			HTTPMethod:   http.MethodGet,
			Handlers: []gin.HandlerFunc{
				gin.WrapF(promhttp.Handler().ServeHTTP),
			},
		}),
	}
}

// Configure performs initial component setup.
func (c *Component) Configure() error {
	c.middlewares = append(c.middlewares, c.metricsMiddleware)
	c.router.Use(c.middlewares...)
	c.RegisterRoutes(c.routes...)
	c.NotifyStatus(components.CONFIGURED)
	return nil
}

// Start performs actions required to begin the component lifecycle.
func (c *Component) Start() error {
	addr := fmt.Sprintf("%s:%d", c.serverHost, c.serverPort)
	c.server = &http.Server{
		Addr:    addr,
		Handler: c.router,
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	c.Logger().Info(fmt.Sprintf("server started, listening on %q", addr))
	c.NotifyStatus(components.STARTED)

	if err = c.server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		c.NotifyError(err)
	}

	return nil
}

// Shutdown performs actions required to gracefully shutdown the component lifecycle.
func (c *Component) Shutdown() error {
	if c.server == nil {
		c.NotifyStatus(components.STOPPED)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := c.server.Shutdown(ctx); err != nil && errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("shutting down HTTP server: %w", err)
	}

	c.NotifyStatus(components.STOPPED)

	return nil
}

func (c *Component) RegisterRoutes(routes ...Route) {
	for _, route := range routes {
		c.router.Handle(route.HTTPMethod, route.RelativePath, route.Handlers...)
	}
}

// GetServerHost returns the server host.
func (c *Component) GetServerHost() string {
	return c.serverHost
}

// GetServerPort returns the server port.
func (c *Component) GetServerPort() int {
	return c.serverPort
}

// GetRoutes returns the registered routes.
func (c *Component) GetRoutes() []Route {
	return c.routes
}
