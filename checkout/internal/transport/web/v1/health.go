package v1

import (
	"net/http"
)

type HealthHandler struct {
	Version  string
	DBPinger interface {
		Ping() error
	}
	CachePinger interface {
		Ping() error
	}
}

func (h *HealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{"ok"})
}

func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	err := h.DBPinger.Ping()
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	err = h.CachePinger.Ping()
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, "ready")
}

func (h *HealthHandler) VersionInfo(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, versionResponse{h.Version})
}
