package delegator

import (
	"context"
	"delegator/internal/models"
	"delegator/pkg/domain"
	"log/slog"
)

type Repository struct {
	logger *slog.Logger
}

func (r Repository) Create(ctx context.Context, data []byte) error {
	//TODO implement me
	panic("implement me")
}

func (r Repository) FindOneByID(ctx context.Context, id domain.ID) (models.Delegation, error) {
	//TODO implement me
	panic("implement me")
}

func (r Repository) FindAll(ctx context.Context) ([]models.Delegation, error) {
	//TODO implement me
	panic("implement me")
}

type RepositoryOptions func(*Repository)

func RepositoryWithLogger(logger *slog.Logger) RepositoryOptions {
	return func(r *Repository) {
		r.logger = logger
	}
}
