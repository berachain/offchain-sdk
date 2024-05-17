package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/berachain/offchain-sdk/log"
)

// 10 seconds is a stable default.
const defaultReadHeaderTimeout = 10 * time.Second

// Handler is a handler.
type Handler struct {
	Path    string
	Handler http.Handler
}

type Middleware func(http.Handler) http.Handler

// Server is a server, that currently only supports HTTP.
type Server struct {
	cfg    *Config
	logger log.Logger

	mux    *http.ServeMux
	srv    *http.Server
	closer sync.Once

	middlewares []Middleware
}

// New creates a new server.
func New(cfg *Config, logger log.Logger, middlewares ...Middleware) *Server {
	return &Server{
		cfg:         cfg,
		logger:      logger,
		mux:         http.NewServeMux(),
		middlewares: middlewares,
	}
}

// RegisterHandler registers a handler.
func (s *Server) RegisterHandler(h *Handler) {
	s.mux.Handle(h.Path, h.Handler)
}

// RegisterMiddleware registers a middleware.
func (s *Server) RegisterMiddleware(m Middleware) {
	s.middlewares = append(s.middlewares, m)
}

// applyMiddlewares applies the middlewares to the server in reverse order,
// so that the first middleware is the outermost one.
func (s *Server) applyMiddlewares() http.Handler {
	var h http.Handler = s.mux
	for i := len(s.middlewares) - 1; i >= 0; i-- {
		h = s.middlewares[i](h)
	}
	return h
}

// Start starts the server. It is blocking so must run in a go-routine.
func (s *Server) Start(ctx context.Context) {
	s.srv = &http.Server{
		Addr:              fmt.Sprintf("%s:%d", s.cfg.HTTP.Host, s.cfg.HTTP.Port),
		Handler:           s.applyMiddlewares(),
		ReadHeaderTimeout: defaultReadHeaderTimeout,
	}

	if err := s.srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		s.logger.Error("HTTP server errored", "err", err)
	} else {
		s.logger.Info("HTTP server closed")
	}

	<-ctx.Done()
	s.Stop()
}

// Stop stops the server.
func (s *Server) Stop() {
	s.closer.Do(func() {
		if err := s.srv.Close(); err != nil {
			s.logger.Error("HTTP server close error", "err", err)
		}
	})
}
