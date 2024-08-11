package eventconsumerservice

import (
	"github.com/0gener/go-service/lib/http"
	"github.com/0gener/go-service/lib/probes"
	"github.com/0gener/go-service/service"
)

const (
	ServiceName = "eventconsumerservice"
)

func Bootstrap() error {
	svc, err := service.New(
		ServiceName,
		service.WithComponent(http.New()),
		service.WithComponent(probes.New(),
			probes.WithMonitoredComponents(),
		),
	)
	if err != nil {
		return err
	}

	return svc.Run()
}
