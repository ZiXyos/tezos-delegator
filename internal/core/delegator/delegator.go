package delegator

import (
	"context"
	"delegator/pkg/domain"
	"log/slog"
)

type DelegatorService struct {
	logger     *slog.Logger
	components []domain.Handler
	cancel     context.CancelFunc
}

type Option func(*DelegatorService)

func WithLogger(logger *slog.Logger) Option {
	return func(delegator *DelegatorService) {
		delegator.logger = logger
	}
}

func WithComponents(components ...domain.Handler) Option {
	return func(delegator *DelegatorService) {
		delegator.components = components
	}
}

func (d *DelegatorService) Run(ctx context.Context) error {
	ctx, d.cancel = context.WithCancel(ctx)

	d.logger.Info("starting service", "name", "delegator")

	for _, handler := range d.components {
		go func(h domain.Handler) {
			defer d.logger.Warn("stopping service", "name", "delegator")
			if err := h.Run(ctx); err != nil {
				d.logger.Warn("component failed", "name", "delegator")
			}
		}(handler)
	}

	<-ctx.Done()
	d.logger.Info("service is shuttin")
	return nil
}

func (d *DelegatorService) Stop(ctx context.Context) error {
	d.logger.Info("stopping service", "name", "delegatpr")

	if d.cancel != nil {
		d.cancel()
	}

	for _, handler := range d.components {
		if err := handler.Shutdown(ctx); err != nil {
			d.logger.Warn("component failed", "name", "delegatpr")
			return err
		}
		d.logger.Info("component stopped", "name", "delegatpr")
	}

	return nil
}

func (d *DelegatorService) Name() string {
	return "delegator"
}

func (d *DelegatorService) SetServiceID(serviceID string) {
	return
}

func NewDelegator(opts ...Option) *DelegatorService {
	d := &DelegatorService{}
	for _, opt := range opts {
		opt(d)
	}

	return d
}
