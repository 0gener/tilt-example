package testcontroller

import (
	"github.com/0gener/go-service/components"
	"github.com/0gener/go-service/lib/http"
)

// WithHTTPComponent configures the http component.
func WithHTTPComponent(httpComponent *http.Component) components.Option {
	return func(component components.Component) error {
		comp, err := components.AsComponent[*Component](component)
		if err != nil {
			return err
		}

		comp.httpComponent = httpComponent
		return nil
	}
}
