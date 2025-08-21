package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/config"
	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/store/postgres"
	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/store/redisidem"
	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/web"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	defer log.Println("exit...")

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed load config: %v", err)
	}

	log.Println("init modules...")

	dsn := cfg.GetDSN()

	postgres, err := postgres.NewPaymentsRepo(dsn)
	if err != nil {
		log.Fatalf("failed init postgres: %v", err)
	}
	defer postgres.Close()
	log.Println("Postgres is initialized")
	log.Println("Run Postgres Migrations")

	err = postgres.RunMigrations()
	if err != nil {
		log.Fatalf("failed migrations: %v", err)
	}
	log.Println("Postgres Migrations ended")

	redis, err := redisidem.New(cfg.Redis)
	if err != nil {
		log.Fatalf("failed init redis: %v", err)
	}
	defer redis.Close()
	log.Println("Redis is initialized")

	server := web.New(cfg.HTTP, postgres, redis)
	go server.Run()
	log.Println("The server is initialized")

	<-ctx.Done()
	log.Println("shutting down...")

	// Контекст с таймаутом (например, 5 секунд)
	ctxSrv, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	server.Close(ctxSrv)
}
