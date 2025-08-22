package v1

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/EgorLis/MicroserviceExampleGo/provider/internal/domain/events"
)

type HealthHandler struct {
	Version string
	DB      interface {
		Ping() error
		Statistic(ctx context.Context) (events.Statistic, error)
	}
}

func (h *HealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{"ok"})
}

func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	err := h.DB.Ping()
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "")
		return
	}
	writeJSON(w, http.StatusOK, "ready")
}

func (h *HealthHandler) VersionInfo(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, versionResponse{h.Version})
}

func (h *HealthHandler) Stats(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(300*time.Millisecond))
	defer cancel()
	stats, err := h.DB.Statistic(ctx)
	if err != nil {
		if isTimeout(err) {
			writeError(w, http.StatusGatewayTimeout, "")
			return
		}
		log.Printf("http: stats error:%v", err)
		writeError(w, http.StatusServiceUnavailable, "")
		return
	}
	writeJSON(w, http.StatusOK, stats)
}
