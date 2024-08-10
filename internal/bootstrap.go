package internal

import (
	"github.com/0gener/go-service/lib/http"
	"github.com/0gener/go-service/lib/postgres"
	"github.com/0gener/go-service/lib/probes"
	"github.com/0gener/go-service/service"
	"github.com/0gener/tilt-example/internal/components/controller"
	"github.com/0gener/tilt-example/internal/components/repository"
)

const (
	ServiceName = "tiltexample"
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
