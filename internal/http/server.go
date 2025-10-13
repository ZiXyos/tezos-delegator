package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Server struct {
	logger *slog.Logger

	engine *gin.Engine // should have a domain engine here
	server *http.Server
}

type Options func(*Server)

func WithHTTPServer(conf any) Options {
	return func(h *Server) {
		if h.engine == nil {
			panic(errors.New("ErrEngineErrorOrder"))
		}
		h.server = &http.Server{
			Addr:         ":8088",
			Handler:      h.engine,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		}
	}
}

func WithEngine(engine *gin.Engine) Options {
	return func(h *Server) {
		h.engine = engine
	}
}

func WithLogger(logger *slog.Logger) Options {
	return func(h *Server) {
		h.logger = logger
	}
}

func WithRoutes(routerRegister func(engine *gin.Engine)) Options {
	return func(h *Server) {
		routerRegister(h.engine)
	}
}

func NewHTTPServer(opts ...Options) *Server {
	h := &Server{}
	for _, opt := range opts {
		opt(h)
	}

	return h
}

func (s *Server) Run(ctx context.Context) error {
	s.logger.Info("starting http server", "addr", "localhost:8088")

	go func() {
		if err := s.server.ListenAndServe(); err != nil {
			s.logger.Warn("http server stopped", "error", err)
		}
	}()

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Warn("http server failed to shutdown", "error", err)
		return err
	}

	return nil
}
