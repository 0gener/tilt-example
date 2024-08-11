package probes

import "errors"

var (
	ErrNotMonitoredComponent = errors.New("component is not a monitored component")
)
