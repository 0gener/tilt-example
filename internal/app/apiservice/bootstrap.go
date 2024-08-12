package apiservice

import (
	"github.com/0gener/go-service/lib/awsmessaging"
	"github.com/0gener/go-service/lib/http"
	"github.com/0gener/go-service/lib/postgres"
	"github.com/0gener/go-service/lib/probes"
	"github.com/0gener/go-service/service"
	"github.com/0gener/tilt-example/internal/app/apiservice/components/controller"
	"github.com/0gener/tilt-example/internal/app/apiservice/components/eventpublisher"
	"github.com/0gener/tilt-example/internal/app/apiservice/components/repository"
)

const (
	ServiceName = "apiservice"
)

func Bootstrap() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	svc, err := service.New(
		ServiceName,
		service.WithComponent(http.New()),
		service.WithComponent(postgres.New(),
			postgres.WithConnectionString(cfg.Database.ConnectionString),
			postgres.WithMigrationsDir(cfg.Database.MigrationsDir),
		),
		service.WithComponent(awsmessaging.New(),
			awsmessaging.WithAWSEndpoint(cfg.AWSEndpoint),
		),
		service.WithComponent(eventpublisher.New(),
			eventpublisher.WithEventsTopicARN(cfg.EventsTopicARN),
		),
		service.WithComponent(repository.New()),
		service.WithComponent(controller.New()),
		service.WithComponent(probes.New(),
			probes.WithMonitoredComponents(postgres.ComponentName),
		),
	)
	if err != nil {
		return err
	}

	return svc.Run()
}
