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
	httpComponent := http.New()

	svc, err := service.New(
		ServiceName,
		service.WithComponent(
			httpComponent,
		),
		service.WithComponent(
			testcontroller.New(),
			testcontroller.WithHTTPComponent(httpComponent),
		),
	)
	if err != nil {
		return err
	}

	return svc.Run()
}
