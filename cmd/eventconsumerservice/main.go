package main

import (
	"github.com/0gener/tilt-example/internal/app/eventconsumerservice"
)

func main() {
	err := eventconsumerservice.Bootstrap()
	if err != nil {
		panic(err)
	}
}
