package server

import "net/http"

// Handler is a handler.
type Handler struct {
	Path    string
	Handler http.Handler
}

// Server is a server.
type Server struct {
	handlers []Handler
}

// New creates a new server.
func New() *Server {
	return &Server{}
}

// RegisterHandler registers a handler.
func (s *Server) RegisterHandler(h Handler) {
	s.handlers = append(s.handlers, h)
}

// Start starts the server.
func (s *Server) Start() error {
	mux := http.NewServeMux()
	for _, h := range s.handlers {
		mux.Handle(h.Path, h.Handler)
	}
	return http.ListenAndServe(":8080", mux)
}

// Stop stops the server.
func (s *Server) Stop() error {
	return nil
}
