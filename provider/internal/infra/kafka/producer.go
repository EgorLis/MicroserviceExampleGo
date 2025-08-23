package kafka

import (
	"context"
	"log"
	"time"

	"github.com/EgorLis/MicroserviceExampleGo/provider/internal/config"
	"github.com/EgorLis/MicroserviceExampleGo/provider/internal/shared/event"
	"github.com/segmentio/kafka-go"
)

type Producer struct {
	w   *kafka.Writer
	cfg config.KafkaProducer
}

func newProducer(cfg config.Kafka) *Producer {
	return &Producer{
		w: &kafka.Writer{
			Addr:                   kafka.TCP(cfg.Brokers...),
			Balancer:               &kafka.Hash{},
			RequiredAcks:           kafka.RequireOne,          // быстрее, чем RequireAll
			AllowAutoTopicCreation: true,                      // удобно в dev
			BatchTimeout:           cfg.Producer.BatchTimeout, // "linger": маленький -> быстрее flush
			BatchSize:              cfg.Producer.BatchSize,    // мгновенно пушить одиночки
			BatchBytes:             0,                         // не ограничиваем по байтам
			Compression:            kafka.Snappy,              // опционально, CPU vs сеть
			MaxAttempts:            5,                         // ретраи внутри writer
			WriteTimeout:           3 * time.Second,
			ReadTimeout:            3 * time.Second,
			// Async: false  // синхронная запись (получаешь ошибку сразу)
		}, cfg: cfg.Producer,
	}
}

func (p *Producer) close() error {
	log.Println("Kafka producer: closed...")
	return p.w.Close()
}

func (p *Producer) Publish(ctx context.Context, event event.Envelope) error {
	msg := p.toKafkaMessage(event)
	return p.w.WriteMessages(ctx, msg)
}
