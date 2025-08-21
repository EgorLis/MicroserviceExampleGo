package payment

import (
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
