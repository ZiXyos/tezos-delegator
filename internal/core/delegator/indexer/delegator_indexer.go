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
	repository        domain.Repository
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

func WithRepository(repository domain.Repository) Options {
	return func(i *DelegatorIndexer) {
		i.repository = repository
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
			}
		}
	}
}

func (d *DelegatorIndexer) indexOnce(ctx context.Context) error {
	count, err := d.repository.CountDelegations(ctx)
	if err != nil {
		d.logger.Warn("failed to count delegations", "error", err)
		return err
	}

	var data []domain.TzktApiDelegationsResponse
	if count == 0 {
		d.logger.Info("database is empty, fetching initial batch of recent delegations")
		data, err = d.DelegationHandler.GetDelegationsFromLevel(0, 1000)
	} else {
		lastLevel, err := d.repository.GetLastProcessedLevel(ctx)
		if err != nil {
			d.logger.Warn("failed to get last processed level", "error", err)
			return err
		}

		d.logger.Info("fetching new delegations", "lastLevel", lastLevel)
		data, err = d.DelegationHandler.GetDelegationsFromLevel(lastLevel, 100)
	}

	if err != nil {
		return err
	}

	if len(data) == 0 {
		d.logger.Info("no new delegations found")
		return nil
	}

	d.logger.Info("processing delegations", "count", len(data))
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
