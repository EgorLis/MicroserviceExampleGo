package postgres

import (
	"time"

	"github.com/shopspring/decimal"
)

type PaymentRow struct {
	ID          string          `db:"payment_id"`
	MerchantID  string          `db:"merchant_id"`
	OrderID     string          `db:"order_id"`
	Amount      decimal.Decimal `db:"amount"`
	Currency    string          `db:"currency"`
	MethodToken string          `db:"method_token"`
	Status      string          `db:"status"`
	PSPRef      *string         `db:"psp_reference"`
	CreatedAt   time.Time       `db:"created_at"`
	UpdatedAt   time.Time       `db:"updated_at"`
}
