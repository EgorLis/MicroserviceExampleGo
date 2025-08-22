package v1

import (
	"time"

	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/domain/payment"
)

// domain -> http
func ToResponse(p payment.Payment) PaymentResponse {
	return PaymentResponse{
		ID: p.ID, MerchantID: p.MerchantID,
		OrderID: p.OrderID, Amount: p.Amount.StringFixed(2),
		Currency: p.Currency, Status: string(p.Status),
		PSPRef: p.PSPRef, CreatedAt: toRFC3339(p.CreatedAt),
		UpdatedAt: toRFC3339(p.UpdatedAt),
	}
}

func toRFC3339(t time.Time) string {
	return t.Truncate(time.Second).UTC().Format(time.RFC3339)
}
