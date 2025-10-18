package domain

import (
	"context"
	"delegator/internal/models"
)

type Repository interface {
	Create(ctx context.Context, data []byte) error
	FindOneByID(ctx context.Context, id ID) error
	FindAll(ctx context.Context) ([]models.Delegation, error)
}

type UseCase interface {
	Create(ctx context.Context, data []byte) error
	GetDelegations(ctx context.Context) (ApiResponse[DelegationsResponseType], error)
}
