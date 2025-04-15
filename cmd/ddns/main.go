package main

import (
	"context"
	"log"

	ddnsApp "github.com/jorgesanchez-e/simple-ddns/internal/app/ddns"
)

func main() {
	ctx := context.Background()

	app, err := ddnsApp.New(ctx)
	if err != nil {
		log.Fatal(err)
	}

	done := app.Run(ctx)

	<-done
}
