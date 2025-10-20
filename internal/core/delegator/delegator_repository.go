package delegator

import (
	"context"
	"delegator/internal/models"
	"delegator/pkg/domain"
	"log/slog"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (r *Repository) FindAll(ctx context.Context) ([]models.Delegation, error) {
	r.logger.Info("delegator repository FindAll")
	res, err := gorm.G[models.Delegation](r.dbClient).Find(ctx)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (r *Repository) GetLastProcessedLevel(ctx context.Context) (int64, error) {
	var maxLevel int64
	err := r.dbClient.Model(&models.Delegation{}).
		Select("COALESCE(MAX(level), 0)").
		Scan(&maxLevel).Error
	if err != nil {
		r.logger.Warn("error getting last processed level", "error", err)
		return 0, err
	}
	r.logger.Info("last processed level", "level", maxLevel)
	return maxLevel, nil
}

func (r *Repository) CountDelegations(ctx context.Context) (int64, error) {
	var count int64
	err := r.dbClient.Model(&models.Delegation{}).Count(&count).Error
	if err != nil {
		r.logger.Warn("error counting delegations", "error", err)
		return 0, err
	}
	return count, nil
}

func (r *Repository) createDelegation(ctx context.Context, delegation models.Delegation) error {
	r.logger.Info("attempting to create delegation", "delegator", delegation.Delegator, "level", delegation.Level, "hash", delegation.OperationHash)
	res := gorm.WithResult()
	err := gorm.G[models.Delegation](r.dbClient, res).Create(ctx, &delegation)
	if err != nil {
		r.logger.Warn("error while creating delegator", "error", err, "delegator", delegation.Delegator, "level", delegation.Level)
		return err
	}

	r.logger.Info("successfully created delegation", "id", delegation.ID, "delegator", delegation.Delegator, "level", delegation.Level)
	return nil
}

func (r *Repository) createBaker(ctx context.Context, baker models.Baker) error {
	err := r.dbClient.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "address"}},
			DoUpdates: clause.AssignmentColumns([]string{"last_seen"}),
		}).
		Create(&baker).Error

	if err != nil {
		r.logger.Warn("error while creating/updating baker", "error", err, "address", baker.Address)
		return err
	}

	r.logger.Info("created/updated baker", "address", baker.Address)
	return nil
}

func NewRepository(opts ...RepositoryOptions) *Repository {
	r := &Repository{}
	for _, opt := range opts {
		opt(r)
	}

	return r
}
