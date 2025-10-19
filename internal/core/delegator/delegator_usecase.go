package delegator

import (
	"context"
	"delegator/internal/models"
	"delegator/pkg/domain"
	"log/slog"
	"time"
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
func (uc *UseCaseImpl) Create(ctx context.Context, data []domain.TzktApiDelegationsResponse) error {
	uc.logger.Info("processing API responses", "total", len(data))
	createDTOs := make([]domain.CreateDelegationDTO, 0, len(data))

	for i, apiResponse := range data {
		uc.logger.Info("processing delegation", "index", i, "type", apiResponse.Type, "status", apiResponse.Status, "level", apiResponse.Level)
		if apiResponse.Type != "delegation" || apiResponse.Status != "applied" {
			uc.logger.Info("skipping delegation", "reason", "wrong type or status", "type", apiResponse.Type, "status", apiResponse.Status)
			continue
		}

		timestamp, err := time.Parse("2006-01-02T15:04:05Z", apiResponse.Timestamp)
		if err != nil {
			uc.logger.Warn("failed to parse timestamp", "timestamp", apiResponse.Timestamp, "error", err)
			continue
		}

		delegatorAddress := ""
		if apiResponse.Sender != nil {
			delegatorAddress = apiResponse.Sender.Address
		}

		if delegatorAddress == "" {
			uc.logger.Warn("missing delegator address", "delegator", delegatorAddress)
			continue
		}

		bakerAddress := ""
		isUndelegation := apiResponse.NewDelegate == nil

		if !isUndelegation {
			bakerAddress = apiResponse.NewDelegate.Address
		} else {
			bakerAddress = "UNDELEGATED"
		}

		baker := models.Baker{
			Address:   bakerAddress,
			FirstSeen: timestamp,
			LastSeen:  timestamp,
		}

		delegation := models.Delegation{
			Delegator:       delegatorAddress,
			BakerID:         bakerAddress,
			Amount:          apiResponse.Amount,
			Timestamp:       timestamp,
			Level:           apiResponse.Level,
			OperationHash:   &apiResponse.Hash,
			IsNewDelegation: !isUndelegation && apiResponse.PrevDelegate == nil,
			PreviousBaker:   nil,
			IndexedAt:       time.Now(),
		}

		if apiResponse.PrevDelegate != nil {
			delegation.PreviousBaker = &apiResponse.PrevDelegate.Address
		}

		createDTO := domain.CreateDelegationDTO{
			Baker:      baker,
			Delegation: delegation,
		}

		createDTOs = append(createDTOs, createDTO)
	}

	if len(createDTOs) == 0 {
		uc.logger.Info("no valid delegations to create")
		return nil
	}

	return uc.repository.Create(ctx, createDTOs)
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
