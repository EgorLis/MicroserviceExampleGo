package provider

import (
	"context"
	"log"

	"github.com/EgorLis/MicroserviceExampleGo/provider/internal/domain/events"
	"github.com/EgorLis/MicroserviceExampleGo/provider/internal/shared/event"
)

type Database interface {
	InsertProcessedEvent(ctx context.Context, payment events.PaymentProcessed) error
}

type PSP interface {
	DecidePayment() (status string, pspRef *string)
}

type Client struct {
	pub events.Publisher
	psp PSP
	db  Database
}

func New(psp PSP, pub events.Publisher, db Database) *Client {
	return &Client{
		psp: psp,
		pub: pub,
		db:  db,
	}
}

func (c *Client) ProvidePayment(ctx context.Context, event event.Envelope) error {
	log.Printf("provider: consumed payment_id=%s", event.Key)
	status, pspRef := c.psp.DecidePayment()
	newEvent, err := events.NewPaymentProcessedEvent(event, string(status), pspRef)
	if err != nil {
		log.Printf("provider: can't create processed event, error:%v", err)
		return err
	}

	if err = c.db.InsertProcessedEvent(ctx, events.PaymentProcessed{
		PaymentID: newEvent.Key,
		Status:    string(status),
		PSPRef:    pspRef,
	}); err != nil {
		log.Printf("provider: database error:%v", err)
		return err
	}

	if err = c.pub.Publish(ctx, newEvent); err != nil {
		log.Printf("provider: publisher error:%v", err)
		return err
	}

	log.Printf("provider: published payment.processed payment_id=%s status=%s", newEvent.Key, status)

	return nil
}
