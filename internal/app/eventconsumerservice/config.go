package eventconsumerservice

import (
	"github.com/0gener/go-service/utils"
)

const (
	EnvAWSEndpoint    = "AWS_ENDPOINT"
	EnvEventsQueueURL = "EVENTS_QUEUE_URL"
)

type Config struct {
	AWSEndpoint    string
	EventsQueueURL string
}

func loadConfig() (*Config, error) {
	eventsQueueUrl, err := utils.GetRequiredString(EnvEventsQueueURL)
	if err != nil {
		return nil, err
	}

	return &Config{
		AWSEndpoint:    utils.GetStringOrDefault(EnvAWSEndpoint, ""),
		EventsQueueURL: eventsQueueUrl,
	}, nil
}
