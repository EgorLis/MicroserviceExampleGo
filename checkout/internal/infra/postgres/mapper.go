package postgres

import (
	"encoding/json"
	"fmt"

	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/domain/payment"
	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/shared/event"
)

// db -> domain
func PaymentRowToDomain(row PaymentRow) payment.Payment {
	return payment.Payment{
		ID: row.ID, MerchantID: row.MerchantID,
		OrderID: row.OrderID, Amount: row.Amount,
		Currency: row.Currency, Status: payment.PaymentStatus(row.Status),
		PSPRef: row.PSPRef, MethodToken: row.MethodToken,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

// domain -> row
func PaymentToRow(p payment.Payment) PaymentRow {
	return PaymentRow{
		ID: p.ID, MerchantID: p.MerchantID,
		OrderID: p.OrderID, Amount: p.Amount,
		Currency: p.Currency, Status: string(p.Status),
		PSPRef: p.PSPRef, MethodToken: p.MethodToken,
		CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}

// row -> envelope
func OutboxRowToEnvelope(row OutboxEventRow) (event.Envelope, error) {
	headers := map[string]string{}
	if err := json.Unmarshal(row.Headers, &headers); err != nil {
		return event.Envelope{}, fmt.Errorf("invalid headers, err:%v", err)
	}

	eventType, err := event.StringToType(row.EventType)
	if err != nil {
		return event.Envelope{}, fmt.Errorf("invalid event type, err:%v", err)
	}

	return event.Envelope{
		Type:    eventType,
		Payload: row.Payload,
		Headers: headers,
		Key:     row.Key,
	}, nil
}

// envelope -> row
func EnvelopeToRow(env event.Envelope) (OutboxEventRow, error) {
	headers, err := json.Marshal(env.Headers)
	if err != nil {
		return OutboxEventRow{}, err
	}

	return OutboxEventRow{
		AggregateType: "payment",
		AggregateID:   env.Key,
		EventType:     string(env.Type),
		Key:           env.Key,
		Payload:       env.Payload,
		Headers:       headers,
	}, nil
}
