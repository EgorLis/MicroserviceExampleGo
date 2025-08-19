package http

import (
	"encoding/json"
	"net/http"

	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/config"
)

type Controller struct {
	// для дальнейшего добавления логеров и прочего
	db    DataBase
	cache Cache
}

type DataBase interface {
	Ping() error
}

type Cache interface {
	Ping() error
}

type healthMsg struct {
	Status string `json:"status"`
}
type versionMsg struct {
	Version string `json:"version"`
}

// Конструктор контроллера
func NewController(db DataBase, cache Cache) *Controller {
	return &Controller{db: db, cache: cache}
}

// методы-контроллера:

func (ch *Controller) GetHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthMsg{"ok"})
}

func (ch *Controller) GetReadyz(w http.ResponseWriter, r *http.Request) {
	err := ch.db.Ping()
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	err = ch.cache.Ping()
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, "ready")
}

func (ch *Controller) GetVersion(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, versionMsg{config.Version})
}

// helpers
func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}
