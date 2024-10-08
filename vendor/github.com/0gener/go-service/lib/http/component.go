package http

import (
	"context"
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	ginprometheus "github.com/zsais/go-gin-prometheus"
	"go.uber.org/zap"
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
	ignoredRoutes     map[string]bool
	middlewares       gin.HandlersChain
	metricsMiddleware gin.HandlerFunc
}

type Route struct {
	RelativePath string
	HTTPMethod   string
	Handlers     []gin.HandlerFunc
	IgnoreLogs   bool
}

func New() *Component {
	gin.SetMode(gin.ReleaseMode)

	return &Component{
		BaseComponent: *components.NewBaseComponent(ComponentName),
		router:        gin.New(),
		ignoredRoutes: make(map[string]bool),
	}
}

func (component *Component) DefaultOptions() []components.Option {
	return []components.Option{
		WithServerHost(defaultServerHost),
		WithServerPort(defaultServerPort),
		WithMetricsMiddleware(
			ginprometheus.NewPrometheus("gin").HandlerFunc(),
		),
		WithMiddlewares([]gin.HandlerFunc{
			gin.Recovery(),
			component.RequestLoggerMiddleware(),
		}),
		WithRoute(Route{
			RelativePath: metricsPath,
			HTTPMethod:   http.MethodGet,
			Handlers: []gin.HandlerFunc{
				gin.WrapF(promhttp.Handler().ServeHTTP),
			},
			IgnoreLogs: true,
		}),
	}
}

// Configure performs initial component setup.
func (component *Component) Configure(_ context.Context) error {
	component.middlewares = append(component.middlewares, component.metricsMiddleware)
	component.router.Use(component.middlewares...)
	component.RegisterRoutes(component.routes...)
	component.NotifyStatus(components.CONFIGURED)
	return nil
}

// Start performs actions required to begin the component lifecycle.
func (component *Component) Start(_ context.Context) error {
	addr := fmt.Sprintf("%s:%d", component.serverHost, component.serverPort)
	component.server = &http.Server{
		Addr:    addr,
		Handler: component.router,
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	component.Logger().Info(fmt.Sprintf("server started, listening on %q", addr))
	component.NotifyStatus(components.STARTED)

	if err = component.server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		component.NotifyError(err)
	}

	return nil
}

// Shutdown performs actions required to gracefully shutdown the component lifecycle.
func (component *Component) Shutdown(ctx context.Context) error {
	if component.server == nil {
		component.NotifyStatus(components.STOPPED)
		return nil
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := component.server.Shutdown(shutdownCtx); err != nil && errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("shutting down HTTP server: %w", err)
	}

	component.NotifyStatus(components.STOPPED)

	return nil
}

func (component *Component) RegisterRoutes(routes ...Route) {
	for _, route := range routes {
		component.router.Handle(route.HTTPMethod, route.RelativePath, route.Handlers...)

		if route.IgnoreLogs {
			component.ignoredRoutes[route.RelativePath] = true
		}
	}
}

// GetServerHost returns the server host.
func (component *Component) GetServerHost() string {
	return component.serverHost
}

// GetServerPort returns the server port.
func (component *Component) GetServerPort() int {
	return component.serverPort
}

// GetRoutes returns the registered routes.
func (component *Component) GetRoutes() []Route {
	return component.routes
}

func (component *Component) RequestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if component.ignoredRoutes[c.Request.URL.Path] {
			// Skip logging for this route
			c.Next()
			return
		}

		startTime := time.Now()

		c.Next()

		duration := time.Since(startTime)
		statusCode := c.Writer.Status()

		component.Logger().Info("HTTP request processed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", statusCode),
			zap.Duration("duration", duration),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		)
	}
}
