package payment

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

type PaymentStatus string

const (
	StatusPending    PaymentStatus = "PENDING"
	StatusProcessing PaymentStatus = "PROCESSING"
	StatusSucceeded  PaymentStatus = "SUCCEEDED"
	StatusFailed     PaymentStatus = "FAILED"
)

type Payment struct {
	ID          string
	MerchantID  string
	OrderID     string
	Amount      decimal.Decimal
	Currency    string
	MethodToken string
	Status      PaymentStatus
	PSPRef      *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Repository interface {
	InsertPayment(ctx context.Context, payment Payment) error
	GetPaymentByID(ctx context.Context, id string) (Payment, error)
	GetPaymentByUniqKeys(ctx context.Context, merchantID, orderID string) (Payment, error)
}
