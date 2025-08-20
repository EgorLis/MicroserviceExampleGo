package idempotency

import "errors"

var (
	ErrNotFound     = errors.New("idempotency record not found")
	ErrBodyMismatch = errors.New("idempotency key reused with different payload")
)
