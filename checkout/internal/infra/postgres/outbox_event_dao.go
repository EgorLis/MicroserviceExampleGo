package postgres

import "time"

type OutboxEventRow struct {
	ID            int64     `db:"id"`
	AggregateType string    `db:"aggregate_type"`
	AggregateID   string    `db:"aggregate_id"`
	EventType     string    `db:"event_type"`
	EventVersion  string    `db:"event_version"`
	Key           string    `db:"key"`
	Payload       []byte    `db:"payload"`
	Headers       []byte    `db:"headers"`
	Status        string    `db:"status"`
	Attempt       int       `db:"attempt"`
	NextAttemptAt time.Time `db:"next_attempt_at"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}
