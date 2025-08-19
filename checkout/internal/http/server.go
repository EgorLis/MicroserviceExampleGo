package http

import (
	"context"
	"log"
	"net/http"
	"time"
)

var server *WebServer = nil

type WebServer struct {
	server *http.Server
}

func New(addr string, db DataBase, cache Cache) *WebServer {
	controller := NewController(db, cache)
	s := &http.Server{
		Addr:           addr,
		Handler:        routes(controller),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	webs := WebServer{s}
	server = &webs
	return server
}

func (ws *WebServer) Run() {
	log.Printf("server started on %s", ws.server.Addr)
	if err := ws.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

func (ws *WebServer) Close(ctx context.Context) {
	if err := ws.server.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("server exited gracefully")
}

func routes(c *Controller) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", c.GetHealthz)
	mux.HandleFunc("GET /readyz", c.GetReadyz)
	mux.HandleFunc("GET /version", c.GetVersion)
	return mux
}
