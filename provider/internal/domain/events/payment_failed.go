package events

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/EgorLis/MicroserviceExampleGo/provider/internal/shared/event"
)

type PaymentFailed struct {
	EventType    string `json:"event_type"`
	EventVersion int    `json:"event_version"`
	PaymentID    string `json:"payment_id"`
	MerchantID   string `json:"merchant_id"`
	OrderID      string `json:"order_id"`
	Amount       string `json:"amount"`
	Currency     string `json:"currency"`
	OccurredAt   string `json:"occurred_at"`
	ErrorDetails string `json:"error_details"`
}

// Конструктор события из доменного объекта
func NewPaymentFailedEvent(evn event.Envelope, errDetails error) (event.Envelope, error) {
	var payload PaymentFailed

	if err := json.Unmarshal(evn.Payload, &payload); err != nil {
		return event.Envelope{}, fmt.Errorf("invalid JSON err:%v", err)
	}

	payload.EventType = string(event.PaymentFailedEvent)
	payload.OccurredAt = time.Now().UTC().Format(time.RFC3339Nano)
	payload.ErrorDetails = errDetails.Error()

	value, err := json.Marshal(payload)
	if err != nil {
		return event.Envelope{}, err
	}

	return event.Envelope{
		Type:    event.PaymentFailedEvent,
		Key:     payload.PaymentID, // партиционирование по payment_id
		Payload: value,
		Headers: evn.Headers,
	}, nil
}
