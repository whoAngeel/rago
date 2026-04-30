package rest

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/whoAngeel/rago/internal/core/ports"
)

type Server struct {
	server *http.Server
	logger ports.Logger
}

func NewServer(host, port string, router http.Handler, logger ports.Logger) *Server {

	return &Server{
		server: &http.Server{
			Addr:         fmt.Sprintf("%s:%s", host, port),
			Handler:      router,
			ReadTimeout:  60 * time.Second,
			WriteTimeout: 60 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		logger: logger,
	}
}

func (s *Server) Start() error {
	s.logger.Info("server started at", "addr", s.server.Addr)
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down server")
	return s.server.Shutdown(ctx)
}
