package main

import (
	"github.com/0gener/tilt-example/internal/app/apiservice"
)

func main() {
	err := apiservice.Bootstrap()
	if err != nil {
		panic(err)
	}
}
