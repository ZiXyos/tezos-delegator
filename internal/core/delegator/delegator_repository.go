package delegator

import (
	"context"
	"delegator/internal/models"
	"delegator/pkg/domain"
	"log/slog"

	"gorm.io/gorm"
)

type Repository struct {
	logger *slog.Logger

	dbClient *gorm.DB // TODO: implement driver on domain.
}

type RepositoryOptions func(*Repository)

func RepositoryWithLogger(logger *slog.Logger) RepositoryOptions {
	return func(r *Repository) {
		r.logger = logger
	}
}

func RepositoryWithDBClient(db *gorm.DB) RepositoryOptions {
	return func(r *Repository) {
		r.dbClient = db
	}
}

func (r *Repository) Create(ctx context.Context, delegationToCreate []domain.CreateDelegationDTO) error {
	r.logger.Info("create delegator", slog.Int("count", len(delegationToCreate)))
	for _, delegation := range delegationToCreate {
		err := r.createBaker(ctx, delegation.Baker)
		if err != nil {
			return err
		}
		err = r.createDelegation(ctx, delegation.Delegation)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) FindOneByID(ctx context.Context, id domain.ID) error {
	//TODO implement me
	panic("implement me")
}

func (r *Repository) FindAll(ctx context.Context) ([]models.Delegation, error) {
	r.logger.Info("delegator repository FindAll")
	res, err := gorm.G[models.Delegation](r.dbClient).Find(ctx)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (r *Repository) createDelegation(ctx context.Context, delegation models.Delegation) error {
	res := gorm.WithResult()
	err := gorm.G[models.Delegation](r.dbClient, res).Create(ctx, &delegation)
	if err != nil {
		r.logger.Warn("error while creating delegator", "error", err)
		return err
	}

	r.logger.Info("create delegator", delegation.ID)
	return nil
}

func (r *Repository) createBaker(ctx context.Context, baker models.Baker) error {
	res := gorm.WithResult()
	err := gorm.G[models.Baker](r.dbClient, res).Create(ctx, &baker)
	if err != nil {
		r.logger.Warn("error while creating baker", "error", err)
		return err
	}

	r.logger.Info("create baker", baker.Address)
	return nil
}

func NewRepository(opts ...RepositoryOptions) *Repository {
	r := &Repository{}
	for _, opt := range opts {
		opt(r)
	}

	return r
}
