package delegator

import (
	"context"
	"delegator/pkg/domain"
	"log/slog"
)

// UseCaseImpl represent the use case implementation tf the delegator.
type UseCaseImpl struct {
	logger     *slog.Logger
	repository domain.Repository
}

// UseCaseOption represent the Option function to load option.
type UseCaseOption func(*UseCaseImpl)

// UseCaseWithLogger inject the logger to the use case.
func UseCaseWithLogger(logger *slog.Logger) UseCaseOption {
	return func(u *UseCaseImpl) {
		u.logger = logger
	}
}

// UseCaseWithRepository inject the repository to the use case.
func UseCaseWithRepository(repository domain.Repository) UseCaseOption {
	return func(u *UseCaseImpl) {
		u.repository = repository
	}
}

// Create will create a new delegation.
func (uc *UseCaseImpl) Create(ctx context.Context, data []byte) error {
	return uc.repository.Create(ctx, data)
}

// GetDelegations return all delegations.
func (uc *UseCaseImpl) GetDelegations(ctx context.Context) (domain.ApiResponse[domain.DelegationsResponseType], error) {
	delegations, err := uc.repository.FindAll(ctx)
	if err != nil {
		return domain.ApiResponse[domain.DelegationsResponseType]{}, err
	}

	res := make([]domain.DelegationsResponseType, len(delegations))
	for i, delegation := range delegations {
		res[i] = domain.DelegationsResponseType{
			Timestamp: delegation.Timestamp,
			Amount:    delegation.Amount,
			Delegator: delegation.Delegator,
			Level:     delegation.Level,
		}
	}

	return domain.ApiResponse[domain.DelegationsResponseType]{
		Data: res,
	}, nil
}

// NewUseCase create a new use case for the delegator.
func NewUseCase(opts ...UseCaseOption) *UseCaseImpl {
	uc := &UseCaseImpl{}
	for _, opt := range opts {
		opt(uc)
	}

	return uc
}
