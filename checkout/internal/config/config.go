package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

var Version = "unknown"

type Config struct {
	HTTPAddr       string
	PGDsn          string
	RedisAddr      string
	PaymentTimeout time.Duration
}

func LoadConfig() *Config {

	timeoutMs, err := strconv.Atoi(getenv("PAYMENT_TIMEOUT_MS", "500"))
	if err != nil {
		log.Fatalf("error: cant initialize PAYMENT_TIMEOUT_MS")
	}

	payTimeout := time.Duration(timeoutMs) * time.Millisecond

	return &Config{
		HTTPAddr:       getenv("HTTP_ADDR", ":7081"),
		PGDsn:          getenv("PG_DSN", "postgres://app:app@localhost:5432/app?sslmode=disable"),
		RedisAddr:      getenv("REDIS_ADDR", "localhost:6379"),
		PaymentTimeout: payTimeout,
	}
}

func getenv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
