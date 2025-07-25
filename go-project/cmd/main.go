package main

import (
	"go-project/internal/app"
	"log"
)

func main() {
	application := app.NewApp()

	if err := application.Start(); err != nil {
		log.Fatalf("Failed to start the application: %v", err)
	}

	// Application logic goes here

	defer application.Stop()
}
