package kafka

import (
	"github.com/EgorLis/MicroserviceExampleGo/provider/internal/shared/event"
	"github.com/segmentio/kafka-go"
)

func (p *Producer) toKafkaMessage(evt event.Envelope) kafka.Message {
	evt.Headers["client-id"] = p.cfg.ClientID
	headers := make([]kafka.Header, 0, len(evt.Headers))
	for k, v := range evt.Headers {
		headers = append(headers, kafka.Header{Key: k, Value: []byte(v)})
	}

	var topic string
	switch evt.Type {
	case event.PaymentProcessedEvent:
		topic = p.cfg.PaymentsProcessedTopic
	}

	return kafka.Message{
		Topic:   topic,
		Key:     []byte(evt.Key),
		Value:   evt.Payload,
		Headers: headers,
	}
}

func (c *Consumer) toEvent(msg kafka.Message) event.Envelope {
	headers := make(map[string]string, len(msg.Headers))

	for _, msgHeader := range msg.Headers {
		headers[msgHeader.Key] = string(msgHeader.Value)
	}

	return event.Envelope{
		Type:    event.PaymentCreatedEvent,
		Payload: msg.Value,
		Headers: headers,
		Key:     string(msg.Key),
	}
}
