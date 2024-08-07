package main

import "github.com/0gener/tilt-example/internal"

func main() {
	err := internal.Bootstrap()
	if err != nil {
		panic(err)
	}
}
