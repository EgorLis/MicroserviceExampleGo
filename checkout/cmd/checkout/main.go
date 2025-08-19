package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/cache"
	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/config"
	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/db"
	myhttp "github.com/EgorLis/MicroserviceExampleGo/checkout/internal/http"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	defer log.Println("exit...")

	cfg := config.LoadConfig()

	log.Println("init modules...")

	postgres, err := db.NewPostgresDB(cfg.PGDsn)
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

	redis, err := cache.NewRedisCache(cfg.RedisAddr, "", 0)
	if err != nil {
		log.Fatalf("failed init redis: %v", err)
	}
	defer redis.Close()
	log.Println("Redis is initialized")

	server := myhttp.New(cfg.HTTPAddr, postgres, redis)
	go server.Run()
	log.Println("The server is initialized")

	<-ctx.Done()
	log.Println("shutting down...")

	// Контекст с таймаутом (например, 5 секунд)
	ctxSrv, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	server.Close(ctxSrv)
}
