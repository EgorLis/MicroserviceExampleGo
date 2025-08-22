package events

import (
	"context"

	"github.com/EgorLis/MicroserviceExampleGo/provider/internal/shared/event"
)

type Publisher interface {
	Publish(ctx context.Context, event event.Envelope) error
}
