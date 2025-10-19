package domain

import (
	"context"
	"delegator/internal/models"
)

type CreateDelegationDTO struct {
	Baker      models.Baker
	Delegation models.Delegation
}

type Repository interface {
	Create(ctx context.Context, delegationToCreate []CreateDelegationDTO) error
	FindOneByID(ctx context.Context, id ID) error
	FindAll(ctx context.Context) ([]models.Delegation, error)
	GetLastProcessedLevel(ctx context.Context) (int64, error)
	CountDelegations(ctx context.Context) (int64, error)
}

type UseCase interface {
	Create(ctx context.Context, data []TzktApiDelegationsResponse) error // should be a dto here instead of the api resp
	GetDelegations(ctx context.Context) (ApiResponse[DelegationsResponseType], error)
}
