package database

import (
	"context"
	"database/sql"
	"log/slog"
)

type PGClient struct {
	Logger *slog.Logger
	Driver *sql.DB
}

type Option func(*PGClient) error

func WithDriver(driver *sql.DB) Option {
	return func(c *PGClient) error {
		c.Driver = driver
		return nil
	}
}

func WithLogger(logger *slog.Logger) Option {
	return func(c *PGClient) error {
		c.Logger = logger
		return nil
	}
}

func (p *PGClient) Run(ctx context.Context) error {
	p.Logger.Info("database client running")
	return nil
}

func (p *PGClient) Shutdown(ctx context.Context) error {
	p.Logger.Info("shutting down database client")
	if p.Driver != nil {
		return p.Driver.Close()
	}
	return nil
}

func NewClient(ctx context.Context, opts ...Option) (*PGClient, error) {
	client := &PGClient{}
	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil, err
		}
	}

	return client, nil
}
