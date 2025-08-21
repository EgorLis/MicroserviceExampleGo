package redisidem

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/config"
	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/domain/idempotency"
	"github.com/redis/go-redis/v9"
)

type Store struct {
	rdb    *redis.Client
	prefix string
}

func New(cfg *config.Redis) (*Store, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Pass,
		DB:       cfg.DB,
	})

	// Проверим соединение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Проверка соединения
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis not responding: %v", err)
	}

	return &Store{rdb: rdb, prefix: cfg.Prefix}, nil
}

func (s *Store) Ping() error {
	// Проверим соединение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Проверка соединения
	if err := s.rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis not responding: %v", err)
	}
	return nil
}

func (s *Store) Close() {
	s.rdb.Close()
	log.Println("Redis closed...")
}

func (s *Store) Reserve(ctx context.Context, merchantID, idemKey, bodyHash string, ttl time.Duration) (bool, error) {
	rec := idempotency.Record{State: idempotency.StateInProgress, BodyHash: bodyHash, UpdatedAt: time.Now().Unix()}
	b, _ := json.Marshal(rec)
	return s.rdb.SetNX(ctx, s.key(merchantID, idemKey), b, ttl).Result()
}

func (s *Store) Finalize(ctx context.Context, merchantID, idemKey, bodyHash string, httpCode int, paymentID string, resp map[string]any, ttl time.Duration) error {
	rec := idempotency.Record{
		State:     idempotency.StateDone,
		BodyHash:  bodyHash,
		PaymentID: paymentID,
		HTTPCode:  httpCode,
		Response:  resp,
		UpdatedAt: time.Now().Unix(),
	}
	b, _ := json.Marshal(rec)
	return s.rdb.Set(ctx, s.key(merchantID, idemKey), b, ttl).Err()
}

func (s *Store) Load(ctx context.Context, merchantID, idemKey string) (*idempotency.Record, error) {
	raw, err := s.rdb.Get(ctx, s.key(merchantID, idemKey)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var rec idempotency.Record
	if err := json.Unmarshal(raw, &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

func (s *Store) key(merchantID, idemKey string) string {
	return s.prefix + merchantID + ":" + idemKey
}
