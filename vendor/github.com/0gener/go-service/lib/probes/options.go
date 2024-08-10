package probes

import "github.com/0gener/go-service/components"

// WithMonitoredComponents specifies a set of component names that will be monitored for
// readiness by the probes component.
func WithMonitoredComponents(componentNames ...string) components.Option {
	return func(component components.Component) error {
		probeComponent, err := components.AsComponent[*Component](component)
		if err != nil {
			return err
		}

		probeComponent.monitoredComponentNames = componentNames

		return nil
	}
}
