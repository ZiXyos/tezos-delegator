package indexer

import (
	"context"
	"delegator/pkg/domain"
	"log/slog"
	"time"
)

type DelegatorIndexer struct {
	logger *slog.Logger

	delegatorUseCase  domain.UseCase
	DelegationHandler domain.DelegationService
}

type Options func(*DelegatorIndexer)

func WithLogger(logger *slog.Logger) Options {
	return func(i *DelegatorIndexer) {
		i.logger = logger
	}
}

func WithDelegatorUseCase(delegatorUseCase domain.UseCase) Options {
	return func(i *DelegatorIndexer) {
		i.delegatorUseCase = delegatorUseCase
	}
}

func WithDelegationHandler(delegationHandler domain.DelegationService) Options {
	return func(i *DelegatorIndexer) {
		i.DelegationHandler = delegationHandler
	}
}

func (d *DelegatorIndexer) Run(ctx context.Context) error {
	d.logger.Info("starting delegator indexer")

	if err := d.indexOnce(ctx); err != nil {
		d.logger.Warn("initial indexing failed", "error", err)
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			d.logger.Info("indexer stopping due to context cancellation")
			return ctx.Err()
		case <-ticker.C:
			if err := d.indexOnce(ctx); err != nil {
				d.logger.Warn("indexing failed", "error", err)
				// Continue running even if one iteration fails
			}
		}
	}
}

func (d *DelegatorIndexer) indexOnce(ctx context.Context) error {
	data, err := d.DelegationHandler.GetDelegations()
	if err != nil {
		return err
	}

	d.logger.Info("got delegations", "count", len(data))
	return d.delegatorUseCase.Create(ctx, data)
}

func (d *DelegatorIndexer) Shutdown(ctx context.Context) error {
	d.logger.Info("shutting down delegator indexer")
	return nil
}

func NewDelegatorIndexer(options ...Options) *DelegatorIndexer {
	i := &DelegatorIndexer{}
	for _, option := range options {
		option(i)
	}

	return i
}
