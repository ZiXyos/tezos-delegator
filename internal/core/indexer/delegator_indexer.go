package indexer

import (
	"context"
	"delegator/pkg/domain"
	"log/slog"
)

type DelegatorIndexer struct {
	logger *slog.Logger

	DelegationHandler domain.DelegationService
}

type Options func(*DelegatorIndexer)

func WithLogger(logger *slog.Logger) Options {
	return func(i *DelegatorIndexer) {
		i.logger = logger
	}
}

func WithDelegationHandler(delegationHandler domain.DelegationService) Options {
	return func(i *DelegatorIndexer) {
		i.DelegationHandler = delegationHandler
	}
}

func (d *DelegatorIndexer) Run(ctx context.Context) error {
	d.logger.Info("starting delegator indexer")

	data, err := d.DelegationHandler.GetDelegations()
	if err != nil {
		d.logger.Info("error getting delegations", "error", err)
		return err
	}

	d.logger.Info("got delegations", "data", data)
	return nil
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
