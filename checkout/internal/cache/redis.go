package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

func NewRedisCache(addr, password string, db int) (*Redis, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,     // "localhost:6379" или "redis:6379" в docker
		Password: password, // если пусто, то ""
		DB:       db,       // 0 — дефолт
	})

	// Проверим соединение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Проверка соединения
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis not responding: %v", err)
	}

	return &Redis{client: rdb}, nil
}

func (r *Redis) Ping() error {
	// Проверим соединение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Проверка соединения
	if err := r.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis not responding: %v", err)
	}
	return nil
}

func (r *Redis) Close() {
	r.client.Close()
	log.Println("Redis closed...")
}
