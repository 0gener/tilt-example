package internal

import (
	"github.com/0gener/go-service/lib/http"
	"github.com/0gener/go-service/service"
	"github.com/0gener/tilt-example/internal/components/testcontroller"
)

const (
	ServiceName = "tiltexample"
)

func Bootstrap() error {
	svc, err := service.New(
		ServiceName,
		service.WithComponent(http.New()),
		service.WithComponent(testcontroller.New()),
	)
	if err != nil {
		return err
	}

	return svc.Run()
}
