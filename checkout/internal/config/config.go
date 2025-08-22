package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

var Version = "unknown"

type Config struct {
	HTTP   HTTP     `mapstructure:"http"`
	Redis  Redis    `mapstructure:"redis"`
	DB     Database `mapstructure:"database"`
	Kafka  Kafka    `mapstructure:"kafka"`
	Outbox Outbox   `mapstructure:"outbox"`
}

type HTTP struct {
	Addr           string        `mapstructure:"addr"`
	PaymentTimeout time.Duration `mapstructure:"payment_timeout"`
}

type Database struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Name string `mapstructure:"name"`
	User string
	Pass string
}

type Redis struct {
	Addr   string `mapstructure:"addr"`
	Prefix string `mapstructure:"prefix"`
	DB     int    `mapstructure:"db"`
	Pass   string
}

type Kafka struct {
	Brokers       string        `mapstructure:"brokers"`
	PaymentsTopic string        `mapstructure:"payments_topic"`
	ClientID      string        `mapstructure:"client_id"`
	BatchSize     int           `mapstructure:"batch_size"`
	BatchTimeout  time.Duration `mapstructure:"batch_timeout"`
}

type Outbox struct {
	PollInterval        time.Duration `mapstructure:"poll_interval"`
	PollTimeout         time.Duration `mapstructure:"poll_timeout"`
	BatchSize           int           `mapstructure:"batch_size"`
	ResetEventsInterval time.Duration `mapstructure:"reset_events_interval"`
	ResetEventsTimeout  time.Duration `mapstructure:"reset_events_timeout"`
	MaxParallel         int           `mapstructure:"max_parallel"`
}

func LoadConfig() (*Config, error) {
	if _, err := os.Stat(".env"); err == nil {
		// пытаемся загрузить .env
		if err := godotenv.Load(); err != nil {
			return nil, errors.New(".env file not found, skipping")
		}
	}

	v := viper.New()

	// ищем файл config.yaml
	v.SetConfigFile(os.Getenv("CONFIG_PATH"))

	// читаем ENV (с префиксом APP_)
	//v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// читаем config.yaml
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// дополняем секретами из ENV
	cfg.DB.User = v.GetString("pg.user")
	cfg.DB.Pass = v.GetString("pg.pass")
	cfg.Redis.Pass = v.GetString("redis.pass")

	// env override для Docker
	if brokers := v.GetString("kafka.brokers"); brokers != "" {
		cfg.Kafka.Brokers = brokers
	}

	if postgresHost := v.GetString("pg.host"); postgresHost != "" {
		cfg.DB.Host = postgresHost
	}

	if redisAddr := v.GetString("redis.addr"); redisAddr != "" {
		cfg.Redis.Addr = redisAddr
	}

	return cfg, nil
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		c.DB.User,
		c.DB.Pass,
		c.DB.Host,
		c.DB.Port,
		c.DB.Name,
	)
}
