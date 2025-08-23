package provider

import (
	"context"
	"fmt"
	"log"

	"github.com/EgorLis/MicroserviceExampleGo/provider/internal/domain/events"
)

type Client struct {
	handlers []*handler
}

func New(psp PSP, pub events.Publisher, db Database, cons []Consumer) *Client {
	handlers := make([]*handler, 0, len(cons))
	for idx, con := range cons {
		handlers = append(handlers, newHandler(con, pub, db, psp, fmt.Sprintf("provider: handler[%d]", idx)))
	}

	return &Client{
		handlers: handlers,
	}
}

func (c *Client) Run(ctx context.Context) {

	log.Println("provider: started")

	for _, h := range c.handlers {
		go h.run(ctx)
	}

	<-ctx.Done()

	log.Println("provider: closed")
}
