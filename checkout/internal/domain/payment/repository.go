package payment

import "context"

type Repository interface {
	InsertPayment(ctx context.Context, payment Payment) error
	GetPaymentByID(ctx context.Context, id string) (Payment, error)
	GetPaymentByUniqKeys(ctx context.Context, merchantID, orderID string) (Payment, error)
}
