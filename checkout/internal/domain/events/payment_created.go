package events

import (
	"encoding/json"
	"time"

	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/domain/payment"
	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/shared/event"
)

type PaymentCreated struct {
	EventType    string `json:"event_type"`
	EventVersion int    `json:"event_version"`
	PaymentID    string `json:"payment_id"`
	MerchantID   string `json:"merchant_id"`
	OrderID      string `json:"order_id"`
	Amount       string `json:"amount"`
	Currency     string `json:"currency"`
	Status       string `json:"status"`
	OccurredAt   string `json:"occurred_at"`
}

// Конструктор события из доменного объекта
func NewPaymentCreatedEvent(pay payment.Payment) (event.Envelope, error) {
	payload := PaymentCreated{
		EventType:    string(event.PaymentCreatedEvent),
		EventVersion: 1,
		PaymentID:    pay.ID,
		MerchantID:   pay.MerchantID,
		OrderID:      pay.OrderID,
		Amount:       pay.Amount.StringFixed(2),
		Currency:     pay.Currency,
		Status:       string(pay.Status),
		OccurredAt:   time.Now().UTC().Format(time.RFC3339Nano),
	}

	value, err := json.Marshal(payload)
	if err != nil {
		return event.Envelope{}, err
	}

	return event.Envelope{
		Type:    event.PaymentCreatedEvent,
		Key:     pay.ID, // партиционирование по payment_id
		Payload: value,
		Headers: map[string]string{
			"content-type": "application/json",
		},
	}, nil
}
