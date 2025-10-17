package delegator

import (
	"context"
	"delegator/pkg/domain"
	"log/slog"
)

type UseCaseImpl struct {
	logger     *slog.Logger
	repository domain.Repository
}

type UseCaseOption func(*UseCaseImpl)

func UseCaseWithLogger(logger *slog.Logger) UseCaseOption {
	return func(u *UseCaseImpl) {
		u.logger = logger
	}
}

func (uc *UseCaseImpl) Create(ctx context.Context, data []byte) error {
	return uc.repository.Create(ctx, data)
}

func (uc *UseCaseImpl) GetDelegations(ctx context.Context) ([]domain.Delegations, error) {
	_, err := uc.repository.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
