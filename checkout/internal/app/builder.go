package app

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/config"
	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/infra/kafka"
	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/infra/outbox"
	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/infra/postgres"
	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/infra/redisidem"
	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/transport/web"
)

type App struct {
	config   *config.Config
	postgres *postgres.PaymentsRepo
	redis    *redisidem.Store
	kafka    *kafka.Producer
	worker   *outbox.Worker
	server   *web.Server
}

func Build() (*App, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed load config: %w", err)
	}

	dsn := cfg.GetDSN()
	postgres, err := postgres.NewPaymentsRepo(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed init postgres: %w", err)
	}

	redis, err := redisidem.New(cfg.Redis)
	if err != nil {
		return nil, fmt.Errorf("failed init redis: %w", err)
	}

	kafka := kafka.NewProducer(cfg.Kafka)

	worker := outbox.New(cfg.Outbox, kafka, postgres)

	server := web.New(cfg.HTTP, postgres, redis, kafka)

	return &App{
		config:   cfg,
		postgres: postgres,
		redis:    redis,
		kafka:    kafka,
		worker:   worker,
		server:   server,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	log.Println("app: start application...")
	err := a.postgres.RunMigrations()
	if err != nil {
		log.Fatalf("failed migrations: %v", err)
	}

	go a.server.Run()
	go a.worker.Run(ctx)

	<-ctx.Done()
	log.Println("app: stop application...")
	// graceful stop
	stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	a.redis.Close()
	a.server.Close(stopCtx)
	a.postgres.Close()
	a.kafka.Close()

	return nil
}
