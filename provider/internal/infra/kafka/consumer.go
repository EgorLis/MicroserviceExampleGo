package kafka

import (
	"context"
	"errors"
	"io"
	"log"

	"github.com/EgorLis/MicroserviceExampleGo/provider/internal/config"
	"github.com/EgorLis/MicroserviceExampleGo/provider/internal/shared/event"
	"github.com/EgorLis/MicroserviceExampleGo/provider/internal/shared/helpers"
	"github.com/segmentio/kafka-go"
)

type consumer struct {
	logPrefix     string
	cons          *kafka.Reader
	readMsgChan   chan kafka.Message
	commitMsgChan chan kafka.Message
}

func newConsumer(cfg config.Kafka, logPrefix string) *consumer {
	return &consumer{
		logPrefix: logPrefix,
		cons: kafka.NewReader(kafka.ReaderConfig{
			Brokers: cfg.Brokers,
			GroupID: cfg.Consumer.GroupID,
			Topic:   cfg.Consumer.PaymentsInitiatedTopic,
		}),
		readMsgChan:   make(chan kafka.Message, 1),
		commitMsgChan: make(chan kafka.Message, 1),
	}
}

func (c *consumer) run(ctx context.Context) {
	go c.startReadMsg(ctx)

	<-ctx.Done()
}

func (c *consumer) close() error {
	log.Printf("%s closed...", c.logPrefix)
	return c.cons.Close()
}

func (c *consumer) startReadMsg(ctx context.Context) {
	log.Printf("%s: read messages started", c.logPrefix)
	defer func() { log.Printf("%s: read messages closed", c.logPrefix) }()
	for {
		msg, err := c.cons.FetchMessage(ctx)
		if err != nil {
			log.Printf("%s: error while reading a msg:%v", c.logPrefix, err)

			if errors.Is(err, io.EOF) || helpers.IsTimeout(err) {
				return
			}

			continue
		}
		select {
		case <-ctx.Done():
			return
		case c.readMsgChan <- msg:
		}
	}
}

// Функция блокирует поток пока консьюмер не прочитал новое сообщение или отменился контекст
func (c *consumer) ConsumeEvent(ctx context.Context) (evn event.Envelope, err error) {
	select {
	case <-ctx.Done():
		return event.Envelope{}, errors.New("context deadline")
	case msg := <-c.readMsgChan:

		select {
		case <-ctx.Done():
			return event.Envelope{}, errors.New("context deadline")
		case c.commitMsgChan <- msg:
			evn := c.toEvent(msg)
			return evn, nil
		}
	}
}

// Функция блокирует поток пока сообщение не поступило в канал или отменился контекст
func (c *consumer) FinalizeEvent(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case msg := <-c.commitMsgChan:
		return c.cons.CommitMessages(ctx, msg)
	}
}
