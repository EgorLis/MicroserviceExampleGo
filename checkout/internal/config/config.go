package config

import "os"

var Version = "unknown"

type Config struct {
	HTTPAddr  string
	PGDsn     string
	RedisAddr string
}

func LoadConfig() *Config {
	return &Config{
		HTTPAddr:  getenv("HTTP_ADDR", ":7081"),
		PGDsn:     getenv("PG_DSN", "postgres://app:app@localhost:5432/app?sslmode=disable"),
		RedisAddr: getenv("REDIS_ADDR", "localhost:6379"),
	}
}

func getenv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
