package kafka

import (
	"context"
	"log"

	"github.com/EgorLis/MicroserviceExampleGo/provider/internal/config"
	"github.com/EgorLis/MicroserviceExampleGo/provider/internal/shared/event"
	"github.com/segmentio/kafka-go"
)

type PaymentProvider interface {
	ProvidePayment(ctx context.Context, event event.Envelope) error
}

type Consumer struct {
	prov    PaymentProvider
	cons    *kafka.Reader
	msgChan chan kafka.Message
}

func NewConsumer(cfg config.Kafka, prov PaymentProvider) *Consumer {
	return &Consumer{
		cons: kafka.NewReader(kafka.ReaderConfig{
			Brokers: cfg.Brokers,
			GroupID: cfg.GroupID,
			Topic:   cfg.PaymentsInitiatedTopic,
		}),
		prov:    prov,
		msgChan: make(chan kafka.Message),
	}
}

func (c *Consumer) Run(ctx context.Context) {
	go c.StartReadMsg(ctx)
	for {
		select {
		case <-ctx.Done():
			log.Println("Kafka consumer closed...")
			c.cons.Close()
			return
		case msg := <-c.msgChan:
			log.Printf("kafka consumer: read msg, key=%s", string(msg.Key))
			evn := c.toEvent(msg)
			// пока оставляю обработку ошибки на провайдера
			c.prov.ProvidePayment(ctx, evn)

			c.cons.CommitMessages(ctx, msg)
			log.Printf("kafka consumer: commited a msg, key=%s", evn.Key)
		}
	}
}

func (c *Consumer) StartReadMsg(ctx context.Context) error {
	for {
		msg, err := c.cons.ReadMessage(ctx)
		if err != nil {
			log.Printf("kafka consumer: error while reading a msg")
			return err
		}
		c.msgChan <- msg
	}
}
