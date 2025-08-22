package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EgorLis/MicroserviceExampleGo/provider/internal/config"
	"github.com/EgorLis/MicroserviceExampleGo/provider/internal/infra/kafka"
	"github.com/EgorLis/MicroserviceExampleGo/provider/internal/infra/postgres"
	"github.com/EgorLis/MicroserviceExampleGo/provider/internal/infra/provider"
	"github.com/EgorLis/MicroserviceExampleGo/provider/internal/infra/psp"
	"github.com/EgorLis/MicroserviceExampleGo/provider/internal/transport/web"
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

	pspSimulator := psp.New(&cfg.PSP)
	log.Println("PSP Simulator is initialized")

	kafkaProducer := kafka.NewProducer(cfg.Kafka)
	defer func() {
		err := kafkaProducer.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	log.Println("Kafka producer is initialized")

	provider := provider.New(pspSimulator, kafkaProducer, postgres)

	log.Println("Payment provider is initialized")

	kafkaConsumer := kafka.NewConsumer(cfg.Kafka, provider)
	go kafkaConsumer.Run(ctx)

	log.Println("Kafka consumer is initialized")

	server := web.New(cfg.HTTP, postgres, kafkaProducer)
	go server.Run()
	log.Println("The server is initialized")

	<-ctx.Done()
	log.Println("shutting down...")

	// Контекст с таймаутом (например, 5 секунд)
	ctxSrv, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	server.Close(ctxSrv)
}
