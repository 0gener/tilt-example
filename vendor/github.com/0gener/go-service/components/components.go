package components

import (
	"reflect"

	"go.uber.org/zap"
)

type Status int

const (
	UNLOADED Status = iota
	CONFIGURED
	STARTED
	STOPPED
)

func (s Status) String() string {
	m := map[Status]string{
		UNLOADED:   "UNLOADED",
		CONFIGURED: "CONFIGURED",
		STARTED:    "STARTED",
		STOPPED:    "STOPPED",
	}

	return m[s]
}

// AsComponent performs the type assertion to the desired component type
func AsComponent[T any](component interface{}) (T, error) {
	var target T
	if reflect.TypeOf(component) != reflect.TypeOf(target) {
		return target, &ErrWrongComponent{ExpectedType: reflect.TypeOf(target).String()}
	}
	return component.(T), nil
}

// Component represents a unit of functionality that can be added to a Service.
type Component interface {
	// DefaultOptions returns the component's default options.
	DefaultOptions() []Option

	// Configure performs initial component setup.
	Configure() error

	// Start performs actions required to begin the component lifecycle.
	Start() error

	// Shutdown performs actions required to gracefully shutdown the component lifecycle.
	Shutdown() error

	// Name returns the name of the component.
	Name() string

	// Logger returns the component logger.
	Logger() *zap.Logger

	// SetLogger sets the component's logger.
	SetLogger(logger *zap.Logger)

	// ErrorChan returns the component's channel for outputting errors.
	ErrorChan() <-chan error

	// StatusChan returns the component's channel for outputting status updates.
	StatusChan() <-chan Status
}
