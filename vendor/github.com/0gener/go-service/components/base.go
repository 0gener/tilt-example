package components

import (
	"sync"

	"go.uber.org/zap"
)

// BaseComponent provides common functionality for components.
type BaseComponent struct {
	name       string
	logger     *zap.Logger
	errorChan  chan error
	statusChan chan Status
	status     Status
	mu         sync.RWMutex
	manager    *Manager
}

func NewBaseComponent(name string) *BaseComponent {
	return &BaseComponent{
		name:       name,
		errorChan:  make(chan error, 1),
		statusChan: make(chan Status, 1),
		status:     UNLOADED,
	}
}

func (bc *BaseComponent) Name() string {
	return bc.name
}

func (bc *BaseComponent) DefaultOptions() []Option {
	return []Option{}
}

func (bc *BaseComponent) Configure() error {
	bc.NotifyStatus(CONFIGURED)
	return nil
}

func (bc *BaseComponent) Start() error {
	bc.NotifyStatus(STARTED)
	return nil
}

func (bc *BaseComponent) Shutdown() error {
	bc.NotifyStatus(STOPPED)
	return nil
}

func (bc *BaseComponent) Logger() *zap.Logger {
	return bc.logger
}

func (bc *BaseComponent) SetLogger(logger *zap.Logger) {
	bc.logger = logger
}

func (bc *BaseComponent) ErrorChan() <-chan error {
	return bc.errorChan
}

func (bc *BaseComponent) StatusChan() <-chan Status {
	return bc.statusChan
}

// NotifyError sends an error to the error channel.
func (bc *BaseComponent) NotifyError(err error) {
	bc.errorChan <- err
}

// NotifyStatus sends a status update to the status channel.
func (bc *BaseComponent) NotifyStatus(status Status) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.status = status
	bc.statusChan <- status
}

// GetStatus returns the current status of the component.
func (bc *BaseComponent) GetStatus() Status {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.status
}

// SetDependencyManager sets the component's manager.
func (bc *BaseComponent) SetDependencyManager(manager *Manager) {
	bc.manager = manager
}

// Dependency loads a dependency.
func (bc *BaseComponent) Dependency(name string) Component {
	return bc.manager.LoadComponent(name)
}
