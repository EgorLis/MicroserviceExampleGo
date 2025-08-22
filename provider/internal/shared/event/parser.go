package event

import "errors"

func StringToType(s string) (EnvelopeType, error) {
	switch s {
	case string(PaymentCreatedEvent):
		return PaymentCreatedEvent, nil
	case string(PaymentProcessedEvent):
		return PaymentProcessedEvent, nil
	default:
		return EnvelopeType(""), errors.New("invalid envelope type")
	}
}
