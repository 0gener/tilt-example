package eventhandler

import "github.com/0gener/go-service/components"

// WithEventsQueueURL configures the queue URL where events will be consumed from.
func WithEventsQueueURL(queueUrl string) components.Option {
	return func(component components.Component) error {
		eventHandlerComponent, err := components.AsComponent[*Component](component)
		if err != nil {
			return err
		}

		eventHandlerComponent.queueUrl = queueUrl
		return nil
	}
}
