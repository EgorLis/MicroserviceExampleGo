package v1

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
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

func createPaymentID() string {
	return "pay_" + uuid.NewString()
}

func canonicalHash(v any) (string, error) {
	// сериализуем в JSON
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	// sha256
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func isTimeout(err error) bool {
	return errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled)
}
