package idempotency

import "time"

type State string

const (
	StateInProgress State = "IN_PROGRESS"
	StateError      State = "ERROR"
	StateDone       State = "DONE"
)

const TTL = time.Duration(24 * time.Hour)

type Record struct {
	State     State          `json:"state"`
	BodyHash  string         `json:"body_hash"`
	PaymentID string         `json:"payment_id,omitempty"`
	HTTPCode  int            `json:"http_code,omitempty"`
	Response  map[string]any `json:"response,omitempty"`
	UpdatedAt int64          `json:"updated_at"`
}
