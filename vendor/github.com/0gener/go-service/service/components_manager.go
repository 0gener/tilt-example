package service

import (
	"context"

	"github.com/0gener/go-service/components"
	"go.uber.org/zap"
)

type ComponentsManager interface {
	Register(component components.Component, opts ...components.Option)
	GetComponent(name string) components.Component
	Configure(ctx context.Context, logger *zap.Logger) error
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
}
