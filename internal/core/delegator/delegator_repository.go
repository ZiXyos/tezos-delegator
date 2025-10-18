package delegator

import (
	"context"
	"database/sql"
	"delegator/internal/models"
	"delegator/pkg/domain"
	"log/slog"
)

type Repository struct {
	logger *slog.Logger

	dbClient *sql.DB // TODO: implement driver on domain.
}

type RepositoryOptions func(*Repository)

func RepositoryWithLogger(logger *slog.Logger) RepositoryOptions {
	return func(r *Repository) {
		r.logger = logger
	}
}

func RepositoryWithDBClient(db *sql.DB) RepositoryOptions {
	return func(r *Repository) {
		r.dbClient = db
	}
}

func (r *Repository) Create(ctx context.Context, data []byte) error {
	//TODO implement me
	panic("implement me")
}

func (r *Repository) FindOneByID(ctx context.Context, id domain.ID) error {
	//TODO implement me
	panic("implement me")
}

func (r *Repository) FindAll(ctx context.Context) ([]models.Delegation, error) {
	r.logger.Info("delegator repository FindAll")

	return nil, nil
}

func NewRepository(opts ...RepositoryOptions) *Repository {
	r := &Repository{}
	for _, opt := range opts {
		opt(r)
	}

	return r
}
