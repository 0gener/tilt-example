package eventconsumerservice

import (
	"github.com/0gener/go-service/lib/awsmessaging"
	"github.com/0gener/go-service/lib/http"
	"github.com/0gener/go-service/lib/probes"
	"github.com/0gener/go-service/service"
	"github.com/0gener/tilt-example/internal/app/eventconsumerservice/components/eventhandler"
)

const (
	ServiceName = "eventconsumerservice"
)

func Bootstrap() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	svc, err := service.New(
		ServiceName,
		service.WithComponent(http.New()),
		service.WithComponent(awsmessaging.New(),
			awsmessaging.WithAWSEndpoint(cfg.AWSEndpoint),
		),
		service.WithComponent(eventhandler.New(),
			eventhandler.WithEventsQueueURL(cfg.EventsQueueURL),
		),
		service.WithComponent(probes.New(),
			probes.WithMonitoredComponents(),
		),
	)
	if err != nil {
		return err
	}

	return svc.Run()
}
