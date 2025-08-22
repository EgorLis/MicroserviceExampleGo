package v1

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

// helpers
func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

func isTimeout(err error) bool {
	return errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled)
}

func toRFC3339(t time.Time) string {
	return t.Truncate(time.Second).UTC().Format(time.RFC3339)
}
