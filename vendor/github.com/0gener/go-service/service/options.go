package service

import (
	"context"

	"github.com/0gener/go-service/components"
	"github.com/0gener/go-service/internal/logfields"
	"go.uber.org/zap"
)

// Option configures a service.
type Option func(*Service) error

// Options applies the provided list of options.
func Options(opts ...Option) Option {
	return func(s *Service) error {
		for _, opt := range opts {
			if err := opt(s); err != nil {
				return err
			}
		}

		return nil
	}
}

// DefaultOptions returns a set of default options for a service.
func DefaultOptions() Option {
	return Options(
		WithDefaultLogger(),
		WithComponentsManager(components.NewManager()),
		WithRunContext(context.Background()),
		WithShutdownContext(context.Background()),
	)
}

// WithLogger attaches the provided logger to a service.
func WithLogger(l *zap.Logger) Option {
	return func(s *Service) error {
		s.logger = l
		return nil
	}
}

// WithDefaultLogger attaches the default logger to a service.
func WithDefaultLogger() Option {
	return func(s *Service) error {
		if s.logger != nil {
			return nil
		}
		l, err := zap.NewProduction()
		if err != nil {
			return err
		}
		s.logger = l.With(zap.String(logfields.ServiceName, s.name))

		return nil
	}
}

// WithComponentsManager registers a component to the service.
func WithComponentsManager(componentsManager ComponentsManager) Option {
	return func(s *Service) error {
		s.componentsManager = componentsManager
		return nil
	}
}

// WithComponent registers a component to the service.
func WithComponent(component components.Component, opts ...components.Option) Option {
	return func(s *Service) error {
		s.componentsManager.Register(component, opts...)

		return nil
	}
}

// WithRunContext attaches the provided context to the execution of the service.
func WithRunContext(ctx context.Context) Option {
	return func(s *Service) error {
		s.runCtx = ctx
		return nil
	}
}

// WithShutdownContext attaches the provided context to the shutdown of the service.
func WithShutdownContext(ctx context.Context) Option {
	return func(s *Service) error {
		s.shutdownCtx = ctx
		return nil
	}
}
