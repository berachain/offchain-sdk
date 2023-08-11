package server

import (
	"context"
	"fmt"
	"net/http"
)

// Handler is a handler.
type Handler struct {
	Path    string
	Handler http.Handler
}

// Server is a server.
type Server struct {
	cfg *Config
	mux *http.ServeMux
}

// New creates a new server.
func New(cfg *Config) *Server {
	return &Server{
		mux: http.NewServeMux(),
		cfg: cfg,
	}
}

// RegisterHandler registers a handler.
func (s *Server) RegisterHandler(h Handler) {
	s.mux.Handle(h.Path, h.Handler)
}

// Start starts the server.
func (s *Server) Start(_ context.Context) {
	if err := http.ListenAndServe(fmt.Sprintf(":%d", s.cfg.HTTP.Port), s.mux); err != nil { //nolint:gosec // todo fix.
		panic(err)
	}
}

// Stop stops the server.
func (s *Server) Stop() {}
