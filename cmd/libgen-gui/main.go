package main

import (
	"log"

	"libgen-gui/internal/app"
)

func main() {
	if err := app.Run(nil); err != nil {
		log.Fatal(err)
	}
}
