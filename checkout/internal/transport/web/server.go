package web

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/config"
	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/infra/kafka"
	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/infra/postgres"
	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/infra/redisidem"
	v1 "github.com/EgorLis/MicroserviceExampleGo/checkout/internal/transport/web/v1"
)

type Server struct {
	server *http.Server
	cfg    config.HTTP
}

func New(cfg config.HTTP, db *postgres.PaymentsRepo, idemStore *redisidem.Store, kafkaProducer *kafka.Producer) *Server {
	healthHandler := &v1.HealthHandler{Version: config.Version, DBPinger: db, CachePinger: idemStore}
	paymentsHandler := &v1.PaymentsHandler{Cfg: cfg, IdemStore: idemStore, Repo: db, Publisher: kafkaProducer}
	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           newRouter(healthHandler, paymentsHandler),
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		MaxHeaderBytes:    1 << 20,
		ReadHeaderTimeout: 2 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	return &Server{server: srv, cfg: cfg}
}

func (ws *Server) Run() {
	log.Printf("server started on %s", ws.server.Addr)
	if err := ws.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

func (ws *Server) Close(ctx context.Context) {
	if err := ws.server.Shutdown(ctx); err != nil {
		log.Printf("server forced to shutdown: %v", err)
	}

	log.Println("server exited gracefully")
}

func newRouter(hh *v1.HealthHandler, ph *v1.PaymentsHandler) http.Handler {
	mux := http.NewServeMux()

	// health
	mux.HandleFunc("GET /healthz", hh.Liveness)
	mux.HandleFunc("GET /readyz", hh.Readiness)
	mux.HandleFunc("GET /version", hh.VersionInfo)

	// payments
	mux.HandleFunc("POST /v1/payments", limitBody(16<<10, ph.Create)) // 16 KB
	mux.HandleFunc("GET /v1/payments/{id}", ph.Get)

	loggedMux := loggingMiddleware(mux)

	return loggedMux
}

func limitBody(n int64, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, n)
		h(w, r)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// оборачиваем ResponseWriter чтобы перехватить код ответа
		lrw := &loggingResponseWriter{w, http.StatusOK}
		next.ServeHTTP(lrw, r)

		log.Printf("%s %s %d %s %s",
			r.Proto,
			r.Method,
			lrw.statusCode,
			r.URL.Path,
			time.Since(start),
		)
	})
}

// helper для перехвата статуса
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
