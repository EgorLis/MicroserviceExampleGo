package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/app"
)

func main() {
	a, err := app.Build()
	if err != nil {
		log.Println("app build error:")
		panic(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := a.Run(ctx); err != nil {
		log.Println("app error:", err)
	}
}
