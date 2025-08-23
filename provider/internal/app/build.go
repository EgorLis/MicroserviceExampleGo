package app

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/EgorLis/MicroserviceExampleGo/provider/internal/config"
	"github.com/EgorLis/MicroserviceExampleGo/provider/internal/infra/kafka"
	"github.com/EgorLis/MicroserviceExampleGo/provider/internal/infra/postgres"
	"github.com/EgorLis/MicroserviceExampleGo/provider/internal/infra/provider"
	"github.com/EgorLis/MicroserviceExampleGo/provider/internal/infra/psp"
	"github.com/EgorLis/MicroserviceExampleGo/provider/internal/transport/web"
)

type App struct {
	config   *config.Config
	postgres *postgres.PaymentsRepo
	pspSim   *psp.Simulator
	kafka    *kafka.Client
	provider *provider.Client
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

	pspSimulator := psp.New(&cfg.PSP)

	kafka := kafka.NewClient(cfg.Kafka)

	cons := kafka.GetConsumers()
	adapters := make([]provider.Consumer, 0, len(cons))
	for _, con := range cons {
		adapters = append(adapters, con)
	}

	provider := provider.New(pspSimulator, kafka.GetProducer(), postgres, adapters)

	server := web.New(cfg.HTTP, postgres)

	return &App{
		config:   cfg,
		postgres: postgres,
		pspSim:   pspSimulator,
		kafka:    kafka,
		provider: provider,
		server:   server,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	log.Println("app: start application...")
	err := a.postgres.RunMigrations()
	if err != nil {
		log.Fatalf("failed migrations: %v", err)
	}

	go a.provider.Run(ctx)
	go a.kafka.Run(ctx)
	go a.server.Run()

	<-ctx.Done()
	log.Println("app: stop application...")
	// graceful stop
	stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	a.server.Close(stopCtx)
	a.postgres.Close()
	a.kafka.Close()

	return nil
}
