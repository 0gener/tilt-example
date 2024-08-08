package components

import (
	"context"
	"fmt"
	"github.com/0gener/go-service/internal/logfields"
	"github.com/hashicorp/go-multierror"
	"go.uber.org/zap"
)

type Manager struct {
	components        []Component
	componentsOptions map[string][]Option

	runtimeErrChan chan error
}

func NewManager() *Manager {
	return &Manager{
		components:        []Component{},
		componentsOptions: make(map[string][]Option),

		runtimeErrChan: make(chan error),
	}
}

func (m *Manager) Register(component Component, opts ...Option) {
	m.components = append(m.components, component)
	m.componentsOptions[component.Name()] = opts
}

// Configure configures all registered components.
func (m *Manager) Configure(ctx context.Context, logger *zap.Logger) error {
	for _, component := range m.components {
		componentName := component.Name()
		component.SetDependencyManager(m)
		component.SetLogger(logger.With(zap.String(logfields.Component, componentName)))

		component.Logger().Info("configuring component")
		opts := append(component.DefaultOptions(), m.componentsOptions[componentName]...)
		if err := Options(opts...)(component); err != nil {
			return fmt.Errorf("failed to apply options to component: %w", err)
		}

		if err := m.executeAndWaitForStatus(ctx, component, func() error {
			return component.Configure()
		}, CONFIGURED); err != nil {
			component.Logger().Error("failed to configure component", zap.Error(err))
			return fmt.Errorf("failed to configure component: %w", err)
		}
	}

	return nil
}

// Start starts all registered components.
// This will block execution until a component errors or until the context is finished.
func (m *Manager) Start(ctx context.Context) error {
	for _, component := range m.components {
		component.Logger().Info("starting component")
		if err := m.executeAndWaitForStatus(ctx, component, component.Start, STARTED); err != nil {
			component.Logger().Error("failed to start component", zap.Error(err))
			return fmt.Errorf("failed to start component: %w", err)
		}

		go m.listenToComponentRuntimeErrors(ctx, component)
	}

	return m.wait(ctx)
}

// Shutdown shutdowns all registered components.
func (m *Manager) Shutdown(ctx context.Context) error {
	var err error

	for _, component := range m.components {
		component.Logger().Info("shutting down component")
		if shutdownErr := m.executeAndWaitForStatus(ctx, component, component.Shutdown, STOPPED); shutdownErr != nil {
			err = multierror.Append(err, shutdownErr)
			component.Logger().Warn("failed to shutdown component", zap.Error(shutdownErr))
		}
	}

	return err
}

// SendRuntimeError sends an error to the internal runtime error channel. This is supposed to be used only in tests.
func (m *Manager) SendRuntimeError(err error) {
	m.runtimeErrChan <- err
}

// LoadComponent returns a component by name.
func (m *Manager) LoadComponent(name string) Component {
	for _, component := range m.components {
		if component.Name() == name {
			return component
		}
	}

	return nil
}

// Wait blocks the execution until a component errors or until the context is finished.
func (m *Manager) wait(ctx context.Context) error {
	select {
	case err := <-m.runtimeErrChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (m *Manager) executeAndWaitForStatus(ctx context.Context, component Component, fn func() error, expectedStatus Status) error {
	errChan := make(chan error, 1)
	statusChan := component.StatusChan()

	go func() {
		if err := fn(); err != nil {
			errChan <- err
		}
	}()

	select {
	case newStatus := <-statusChan:
		if newStatus != expectedStatus {
			return fmt.Errorf("%s: not the expected status: got %v, expected %v", component.Name(), newStatus, expectedStatus)
		}
	case err := <-errChan:
		return fmt.Errorf("%s: %w", component.Name(), err)
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

func (m *Manager) listenToComponentRuntimeErrors(ctx context.Context, component Component) {
	for {
		select {
		case <-ctx.Done():
			return
		case err := <-component.ErrorChan():
			if err == nil {
				continue
			}
			select {
			case m.runtimeErrChan <- fmt.Errorf("%s: %w", component.Name(), err):
			case <-ctx.Done():
				return
			}
		}
	}
}
