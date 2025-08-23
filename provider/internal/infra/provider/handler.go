package provider

import (
	"context"
	"log"
	"time"

	"github.com/EgorLis/MicroserviceExampleGo/provider/internal/domain/events"
	"github.com/EgorLis/MicroserviceExampleGo/provider/internal/shared/event"
	"github.com/EgorLis/MicroserviceExampleGo/provider/internal/shared/helpers"
)

type Database interface {
	InsertProcessedEvent(ctx context.Context, payment events.PaymentProcessed) error
}

type PSP interface {
	DecidePayment() (status string, pspRef *string)
}

type Consumer interface {
	ConsumeEvent(ctx context.Context) (evn event.Envelope, err error)
	FinalizeEvent(ctx context.Context) error
}

type handler struct {
	logPrefix string
	consumer  Consumer
	pub       events.Publisher
	db        Database
	psp       PSP

	evnChan chan event.Envelope
}

func newHandler(con Consumer, pub events.Publisher, db Database, psp PSP, logPrefix string) *handler {
	evnChan := make(chan event.Envelope, 1)

	return &handler{
		logPrefix: logPrefix,
		consumer:  con,
		pub:       pub,
		db:        db,
		psp:       psp,
		evnChan:   evnChan,
	}
}

func (h *handler) run(ctx context.Context) {
	go h.startReadEvents(ctx)
	go h.startProcessEvents(ctx)

	<-ctx.Done()
}

func (h *handler) startProcessEvents(ctx context.Context) {
	log.Printf("%s: process events started", h.logPrefix)
	defer func() { log.Printf("%s: process events closed", h.logPrefix) }()

	for {
		select {
		case <-ctx.Done():
			return
		case evn := <-h.evnChan:
			// пытаемся провести платеж
			if err := h.providePayment(ctx, evn); err != nil {
				if helpers.IsTimeout(err) {
					return
				}
				// если неудачно, то пишем в топик брокера
				if paymentFailedEvn, err := events.NewPaymentFailedEvent(evn, err); err != nil {
					log.Printf("%s: error while create new payment failed event:%v", h.logPrefix, err)
				} else {
					if err = h.pub.Publish(ctx, paymentFailedEvn); err != nil {
						log.Printf("%s: publisher error:%v", h.logPrefix, err)
						continue
					}
					log.Printf("%s: published payment.failed payment_id=%s", h.logPrefix, paymentFailedEvn.Key)
				}
			}

			if err := h.consumer.FinalizeEvent(ctx); err != nil {
				log.Printf("%s: error while finalize a event: %v", h.logPrefix, err)
				if helpers.IsTimeout(err) {
					return
				}
			}
		}
	}
}

func (h *handler) startReadEvents(ctx context.Context) {
	log.Printf("%s: read events started", h.logPrefix)
	defer func() { log.Printf("%s: read events closed", h.logPrefix) }()

	for {
		evn, err := h.consumer.ConsumeEvent(ctx)
		if err != nil {
			return
		}
		select {
		case h.evnChan <- evn:
		case <-ctx.Done():
			return
		}

	}
}

func (h *handler) providePayment(ctx context.Context, event event.Envelope) error {
	log.Printf("%s: consumed payment_id=%s", h.logPrefix, event.Key)
	status, pspRef := h.psp.DecidePayment()
	newEvent, err := events.NewPaymentProcessedEvent(event, string(status), pspRef)
	if err != nil {
		log.Printf("%s: can't create processed event, error:%v", h.logPrefix, err)
		return err
	}

	attempt, err := h.retray(0, func() error {
		return h.db.InsertProcessedEvent(ctx, events.PaymentProcessed{
			PaymentID: newEvent.Key,
			Status:    string(status),
			PSPRef:    pspRef,
		})
	})

	if err != nil {
		log.Printf("%s: database error:%v", h.logPrefix, err)
		return err
	}

	_, err = h.retray(attempt, func() error {
		return h.pub.Publish(ctx, newEvent)
	})

	if err != nil {
		log.Printf("%s: publisher error:%v", h.logPrefix, err)
		return err
	}

	log.Printf("%s: published payment.processed payment_id=%s status=%s", h.logPrefix, newEvent.Key, status)

	return nil
}

func (h *handler) retray(attempt int, fn func() error) (int, error) {
	var lastErr error

	for curAttempt := attempt; curAttempt < 4; curAttempt++ {
		time.Sleep(5 * time.Second * time.Duration(curAttempt))
		err := fn()
		if err == nil {
			return attempt, err
		}
		log.Printf("%s attempt=%d error while provide a payment:%v", h.logPrefix, curAttempt+1, err)
		lastErr = err
	}

	return -1, lastErr
}
