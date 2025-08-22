package event

type EnvelopeType string

const (
	PaymentCreatedEvent   EnvelopeType = "payment.created"
	PaymentProcessedEvent EnvelopeType = "payments.processed"
)

type Envelope struct {
	Type    EnvelopeType      // "payment.created"
	Key     string            // routing key (e.g. payment_id)
	Payload []byte            // JSON/Proto
	Headers map[string]string // метаданные
}
