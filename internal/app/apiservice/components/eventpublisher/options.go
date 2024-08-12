package eventpublisher

import "github.com/0gener/go-service/components"

// WithEventsTopicARN configures the topic ARN where events will be published.
func WithEventsTopicARN(topicArn string) components.Option {
	return func(component components.Component) error {
		publisherComponent, err := components.AsComponent[*Component](component)
		if err != nil {
			return err
		}

		publisherComponent.topicArn = topicArn
		return nil
	}
}
