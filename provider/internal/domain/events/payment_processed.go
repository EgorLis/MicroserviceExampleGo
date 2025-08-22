package events

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/EgorLis/MicroserviceExampleGo/provider/internal/shared/event"
)

type PaymentProcessed struct {
	EventType    string  `json:"event_type"`
	EventVersion int     `json:"event_version"`
	PaymentID    string  `json:"payment_id"`
	MerchantID   string  `json:"merchant_id"`
	OrderID      string  `json:"order_id"`
	Amount       string  `json:"amount"`
	Currency     string  `json:"currency"`
	Status       string  `json:"status"`
	PSPRef       *string `json:"psp_reference"`
	OccurredAt   string  `json:"occurred_at"`
}

// Конструктор события из доменного объекта
func NewPaymentProcessedEvent(evn event.Envelope, status string, pspRef *string) (event.Envelope, error) {
	var payload PaymentProcessed

	if err := json.Unmarshal(evn.Payload, &payload); err != nil {
		return event.Envelope{}, fmt.Errorf("invalid JSON err:%v", err)
	}

	payload.EventType = string(event.PaymentProcessedEvent)
	payload.Status = status
	payload.PSPRef = pspRef
	payload.OccurredAt = time.Now().UTC().Format(time.RFC3339Nano)

	value, err := json.Marshal(payload)
	if err != nil {
		return event.Envelope{}, err
	}

	return event.Envelope{
		Type:    event.PaymentProcessedEvent,
		Key:     payload.PaymentID, // партиционирование по payment_id
		Payload: value,
		Headers: evn.Headers,
	}, nil
}
