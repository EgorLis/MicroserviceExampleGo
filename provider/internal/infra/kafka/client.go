package kafka

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/EgorLis/MicroserviceExampleGo/provider/internal/config"
)

type Client struct {
	producer  *Producer
	consumers []*consumer
	cfg       config.Kafka
}

func NewClient(cfg config.Kafka) *Client {
	consumers := make([]*consumer, 0, cfg.Consumer.Partitions)

	for i := range cfg.Consumer.Partitions {
		consumers = append(consumers, newConsumer(cfg, fmt.Sprintf("kafka consumer[%d]", i)))
	}

	return &Client{
		producer:  newProducer(cfg),
		consumers: consumers,
		cfg:       cfg,
	}
}

func (c *Client) Run(ctx context.Context) {
	log.Println("kafka client: started...")

	for _, cons := range c.consumers {
		go cons.run(ctx)
	}

	<-ctx.Done()
}

func (c *Client) Close() {
	wg := &sync.WaitGroup{}

	wg.Go(func() {
		if err := c.producer.close(); err != nil {
			log.Printf("kafka client: error while closing producer:%v", err)
		}
	})

	for _, con := range c.consumers {
		wg.Go(func() {
			if err := con.close(); err != nil {
				log.Printf("kafka client: error while closing consumer:%v", err)
			}
		})
	}

	wg.Wait()

	log.Println("kafka client: closed...")
}

func (c *Client) GetProducer() *Producer {
	return c.producer
}

func (c *Client) GetConsumers() []*consumer {
	return c.consumers
}
