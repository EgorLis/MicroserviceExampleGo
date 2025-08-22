package event

func StringToType(s string) (EnvelopeType, error) {
	switch s {
	case string(PaymentCreatedEvent):
		return PaymentCreatedEvent, nil
	default:
		return EnvelopeType(""), nil
	}
}
