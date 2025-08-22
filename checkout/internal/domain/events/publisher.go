package events

import (
	"context"

	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/shared/event"
)

type Publisher interface {
	Publish(ctx context.Context, event event.Envelope) error
}
