package idempotency

import (
	"context"
	"time"
)

// Доменный интерфейс для идемпотентности
type Store interface {
	Reserve(ctx context.Context, merchantID, key, bodyHash string, ttl time.Duration) (created bool, err error)
	Finalize(ctx context.Context, merchantID, key, bodyHash string, httpCode int, paymentID string, resp map[string]any, ttl time.Duration) error
	Load(ctx context.Context, merchantID, key string) (*Record, error)
}
