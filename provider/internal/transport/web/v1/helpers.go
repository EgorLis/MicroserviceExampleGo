package v1

import (
	"encoding/json"
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

func toRFC3339(t time.Time) string {
	return t.Truncate(time.Second).UTC().Format(time.RFC3339)
}
