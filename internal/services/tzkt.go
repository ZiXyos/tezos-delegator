package services

import (
	"context"
	"log/slog"
)

type Tzkt struct {
	logger  *slog.Logger
	handler *HTTPHandler
}

type Options func(*Tzkt)

func WithLogger(logger *slog.Logger) Options {
	return func(h *Tzkt) {
		h.logger = logger
	}
}

func WithHandler(handler *HTTPHandler) Options {
	return func(h *Tzkt) {
		h.handler = handler
	}
}

func (t *Tzkt) Run(ctx context.Context) error {
	t.logger.Info("starting tzkt components")
	go func() {
		data, err := t.handler.GetDelegations()
		if err != nil {
			t.logger.Warn("error getting delegations: ", err)
		}
		t.logger.Info("got delegations: ", data)
	}()
	return nil
}

func (t *Tzkt) Shutdown(ctx context.Context) error {
	t.logger.Info("shutting down tzkt components")
	return nil
}

func NewTzktClient(opts ...Options) *Tzkt {
	t := &Tzkt{}
	for _, opt := range opts {
		opt(t)
	}

	return t
}
