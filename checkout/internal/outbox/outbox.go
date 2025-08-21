package outbox

import (
	"context"

	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/config"
	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/domain/events"
	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/domain/payment"
)

type OutboxWorker struct {
	repo *payment.Repository
	pub  *events.Publisher
	cfg  *config.Outbox
}

func New(cfg *config.Outbox, pub *events.Publisher, repo *payment.Repository) *OutboxWorker {
	return &OutboxWorker{
		cfg:  cfg,
		pub:  pub,
		repo: repo,
	}
}

func Run(ctx context.Context) {

}
