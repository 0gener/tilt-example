package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/0gener/go-service/components"
	"os/signal"
	"sync"
	"syscall"

	"go.uber.org/zap"
)

var (
	// ErrServiceNameRequired is thrown a service is instantiated without a service name.
	ErrServiceNameRequired = errors.New("service name is required")

	// ErrInterrupted is thrown when the service is interrupted.
	ErrInterrupted = errors.New("interrupted")
)

// Service represents a framework for managing the lifecycle of a set of Components.
type Service struct {
	name              string
	lifecycleMutex    sync.Mutex
	logger            *zap.Logger
	lifecycleErrChan  chan error
	componentsManager ComponentsManager

	runCtx      context.Context
	shutdownCtx context.Context
}

// New constructs a new Service.
func New(name string, opts ...Option) (*Service, error) {
	if name == "" {
		return nil, ErrServiceNameRequired
	}

	service := &Service{
		name:             name,
		lifecycleErrChan: make(chan error, 1),
	}

	opts = append([]Option{DefaultOptions()}, opts...)
	if err := Options(opts...)(service); err != nil {
		return nil, err
	}

	return service, nil
}

// Run starts the lifecycle of the service. This is a blocking statement.
// It blocks here until a lifecycle error occurs. A lifecycle error could be:
// - an error during configuration/startup
// - an error from a long-running start routine
// - a context cancellation (timeout/interrupt signal)
func (s *Service) Run() error {
	var lifecycleErr error
	if err := s.runLifecycle(s.runCtx); err != nil {
		lifecycleErr = fmt.Errorf("service lifecycle: %w", err)
	}

	return s.runShutdown(s.shutdownCtx, lifecycleErr)
}

// GetComponent returns a registered component by name.
func (s *Service) GetComponent(name string) components.Component {
	return s.componentsManager.GetComponent(name)
}

func (s *Service) runLifecycle(ctx context.Context) error {
	notifyCtx, notifyCtxStop := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer notifyCtxStop()

	lifecycleCtx, cancelLifecycleCtx := context.WithCancel(notifyCtx)
	defer cancelLifecycleCtx()

	go func() {
		if err := s.configureWithContext(lifecycleCtx); err != nil {
			s.lifecycleErrChan <- err
			return
		}

		if err := s.startWithContext(lifecycleCtx); err != nil {
			s.lifecycleErrChan <- err
		}
	}()

	// Wait for shutdown signals
	var err error
	select {
	// Lifecycle error
	case lifecycleErr := <-s.lifecycleErrChan:
		err = lifecycleErr

	// Context cancelled by caller
	case <-ctx.Done():
		s.logger.Warn("run context cancelled by caller")
		err = ctx.Err()

	// Interrupt signal received
	case <-notifyCtx.Done():
		s.logger.Warn("interrupt signal received, shutting down")
		err = ErrInterrupted
	}

	return err
}

func (s *Service) runShutdown(ctx context.Context, lifecycleErr error) error {
	notifyCtx, notifyCtxStop := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer notifyCtxStop()

	shutdownCtx, cancelShutdownCtx := context.WithCancel(notifyCtx)
	defer cancelShutdownCtx()

	done := make(chan struct{})

	go func() {
		s.shutdown(shutdownCtx)
		close(done)
	}()

	select {
	// Interrupt signal received
	case <-notifyCtx.Done():
		s.logger.Warn("force stop requested, forcing service shutdown")
		cancelShutdownCtx()
	// Context cancelled by caller
	case <-ctx.Done():
		s.logger.Warn("shutdown context cancelled by caller")

	// Shutdown was completed successfully
	case <-done:
		s.logger.Info("service shutdown complete")
	}

	if lifecycleErr != nil {
		return lifecycleErr
	}

	return nil
}

// configureWithContext configures the service with the provided context.
func (s *Service) configureWithContext(ctx context.Context) error {
	s.lifecycleMutex.Lock()
	defer s.lifecycleMutex.Unlock()

	if err := s.componentsManager.Configure(ctx, s.logger); err != nil {
		return err
	}

	return nil
}

// startWithContext starts the service with the provided context.
func (s *Service) startWithContext(ctx context.Context) error {
	s.lifecycleMutex.Lock()
	defer s.lifecycleMutex.Unlock()

	if err := s.componentsManager.Start(ctx); err != nil {
		return err
	}

	return nil
}

// shutdown stops the service with the provided context.
func (s *Service) shutdown(ctx context.Context) {
	s.lifecycleMutex.Lock()
	defer s.lifecycleMutex.Unlock()

	_ = s.componentsManager.Shutdown(ctx)
}
