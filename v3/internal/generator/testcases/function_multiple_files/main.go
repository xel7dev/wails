package main

import (
	"log"

	"github.com/xel7dev/wails/v3/pkg/application"
)

func main() {
	app := application.New(application.Options{
		Services: []application.Service{
			NewGreetService(),
		},
	})

	app.Window.New()

	err := app.Run()

	if err != nil {
		log.Fatal(err)
	}

}
