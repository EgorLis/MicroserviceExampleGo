package kafka

import (
	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/shared/event"
	"github.com/segmentio/kafka-go"
)

func (p *Producer) toKafkaMessage(evt event.Envelope) kafka.Message {
	headers := make([]kafka.Header, 0, len(evt.Headers)+1)
	for k, v := range evt.Headers {
		headers = append(headers, kafka.Header{Key: k, Value: []byte(v)})
	}

	headers = append(headers, kafka.Header{Key: "client-id",
		Value: []byte(p.cfg.ClientID)})

	var topic string
	switch evt.Type {
	case event.PaymentCreatedEvent:
		topic = p.cfg.PaymentsTopic
	}

	return kafka.Message{
		Topic:   topic,
		Key:     []byte(evt.Key),
		Value:   evt.Payload,
		Headers: headers,
	}
}
