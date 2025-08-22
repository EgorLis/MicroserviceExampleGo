package outbox

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/config"
	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/domain/events"
	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/shared/event"
)

type Repository interface {
	PickBatch(ctx context.Context, count int) (map[int64]event.Envelope, error)
	MarkSent(ctx context.Context, ids []int64) error
	MarkFailed(ctx context.Context, ids []int64) error
	ResetEvents(ctx context.Context) error
}

type Worker struct {
	repo Repository
	pub  events.Publisher
	cfg  config.Outbox
}

func New(cfg config.Outbox, pub events.Publisher, repo Repository) *Worker {
	return &Worker{
		cfg:  cfg,
		pub:  pub,
		repo: repo,
	}
}

func (w *Worker) Run(ctx context.Context) {
	tickerPoll := time.NewTicker(w.cfg.PollInterval)
	tickerReset := time.NewTicker(w.cfg.ResetEventsInterval)
	defer func() {
		tickerPoll.Stop()
		tickerReset.Stop()
	}()

	for {
		select {
		case <-tickerPoll.C:
			pollCtx, cancel := context.WithTimeout(ctx, w.cfg.PollTimeout)

			w.PollBatch(pollCtx)

			cancel()
		case <-tickerReset.C:
			ctxReset, cancel := context.WithTimeout(ctx, w.cfg.ResetEventsTimeout)

			if err := w.repo.ResetEvents(ctxReset); err != nil {
				log.Printf("worker: error while reset events:%v", err)
			}

			cancel()
		case <-ctx.Done():
			log.Println("Outbox worker closed...")
			return
		}
	}
}

func (w *Worker) PollBatch(ctx context.Context) {
	envs, err := w.repo.PickBatch(ctx, w.cfg.BatchSize)
	if err != nil {
		log.Printf("worker: error pick batch:%v", err)
		return
	}

	if len(envs) == 0 {
		return
	}

	var (
		muSent   sync.Mutex
		muFailed sync.Mutex
		sent     = make([]int64, 0, len(envs))
		failed   = make([]int64, 0, len(envs))
		wg       sync.WaitGroup
	)

	semaphore := make(chan struct{}, w.cfg.MaxParallel)

	wg.Add(len(envs))
	for id, env := range envs {
		go func(int64, event.Envelope) {
			select {
			case semaphore <- struct{}{}:
				// заняли
			case <-ctx.Done():
				wg.Done()
				return
			}

			defer func() {
				<-semaphore // освободить
				wg.Done()
			}()

			if err := w.pub.Publish(ctx, env); err != nil {
				log.Printf("worker: kafka: publish failed key=%s, event_id=%d error:%v", env.Key, id, err)
				muFailed.Lock()
				failed = append(failed, id)
				muFailed.Unlock()
				return
			}
			muSent.Lock()
			sent = append(sent, id)
			muSent.Unlock()
			log.Printf("worker: kafka: published key=%s, event_id=%d", env.Key, id)

		}(id, env)
	}

	wg.Wait()

	if len(sent) > 0 {
		if err := w.repo.MarkSent(ctx, sent); err != nil {
			log.Printf("worker: error update sent:%v", err)
		}
	}

	if len(failed) > 0 {
		if err := w.repo.MarkFailed(ctx, failed); err != nil {
			log.Printf("worker: error update failed:%v", err)
		}
	}
}
