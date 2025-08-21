package events

type EventType int

const (
	PaymentCreatedEvent EventType = iota
)

// Унифицированное событие для отправки
type Event struct {
	Type    EventType
	Key     string
	Value   []byte
	Headers map[string]string
}
