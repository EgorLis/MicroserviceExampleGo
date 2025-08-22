package payment

import (
	"context"

	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/shared/event"
)

type Repository interface {
	InsertPayment(ctx context.Context, payment Payment, out event.Envelope) error
	GetPaymentByID(ctx context.Context, id string) (Payment, error)
	GetPaymentByUniqKeys(ctx context.Context, merchantID, orderID string) (Payment, error)
}
