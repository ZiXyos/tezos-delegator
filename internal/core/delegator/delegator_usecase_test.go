package delegator

import (
	"context"
	"delegator/internal/models"
	"delegator/mocks"
	"delegator/pkg/domain"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUseCaseImpl_Create_Comprehensive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		data        []domain.TzktApiDelegationsResponse
		setupMocks  func(*mocks.MockRepository)
		wantErr     bool
		expectedErr string
	}{
		{
			name: "Valid_Single_Delegation",
			data: []domain.TzktApiDelegationsResponse{
				{
					Type:      "delegation",
					Status:    "applied",
					Timestamp: "2023-01-01T12:00:00Z",
					Level:     1000,
					Hash:      "ophash123",
					Amount:    100000,
					Sender: &domain.Account{
						Address: "tz1delegator",
					},
					NewDelegate: &domain.Account{
						Address: "tz1baker",
					},
				},
			},
			setupMocks: func(repo *mocks.MockRepository) {
				repo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(dtos []domain.CreateDelegationDTO) bool {
					return len(dtos) == 1 &&
						dtos[0].Delegation.Delegator == "tz1delegator" &&
						dtos[0].Baker.Address == "tz1baker" &&
						dtos[0].Delegation.Amount == 100000 &&
						dtos[0].Delegation.Level == 1000 &&
						dtos[0].Delegation.IsNewDelegation == true
				})).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "Valid_Undelegation",
			data: []domain.TzktApiDelegationsResponse{
				{
					Type:      "delegation",
					Status:    "applied",
					Timestamp: "2023-01-01T12:00:00Z",
					Level:     1000,
					Hash:      "ophash123",
					Amount:    0,
					Sender: &domain.Account{
						Address: "tz1delegator",
					},
					NewDelegate:  nil, // Undelegation
					PrevDelegate: &domain.Account{Address: "tz1oldbaker"},
				},
			},
			setupMocks: func(repo *mocks.MockRepository) {
				repo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(dtos []domain.CreateDelegationDTO) bool {
					return len(dtos) == 1 &&
						dtos[0].Baker.Address == "UNDELEGATED" &&
						*dtos[0].Delegation.PreviousBaker == "tz1oldbaker" &&
						dtos[0].Delegation.IsNewDelegation == false
				})).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "Valid_Redelegation",
			data: []domain.TzktApiDelegationsResponse{
				{
					Type:      "delegation",
					Status:    "applied",
					Timestamp: "2023-01-01T12:00:00Z",
					Level:     1000,
					Hash:      "ophash123",
					Amount:    100000,
					Sender: &domain.Account{
						Address: "tz1delegator",
					},
					NewDelegate: &domain.Account{
						Address: "tz1newbaker",
					},
					PrevDelegate: &domain.Account{
						Address: "tz1oldbaker",
					},
				},
			},
			setupMocks: func(repo *mocks.MockRepository) {
				repo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(dtos []domain.CreateDelegationDTO) bool {
					return len(dtos) == 1 &&
						dtos[0].Baker.Address == "tz1newbaker" &&
						*dtos[0].Delegation.PreviousBaker == "tz1oldbaker" &&
						dtos[0].Delegation.IsNewDelegation == false
				})).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "Invalid_Timestamp_Format",
			data: []domain.TzktApiDelegationsResponse{
				{
					Type:      "delegation",
					Status:    "applied",
					Timestamp: "invalid-timestamp",
					Level:     1000,
					Hash:      "ophash123",
					Amount:    100000,
					Sender: &domain.Account{
						Address: "tz1delegator",
					},
					NewDelegate: &domain.Account{
						Address: "tz1baker",
					},
				},
			},
			setupMocks: func(repo *mocks.MockRepository) {
				// Should not call Create since delegation is skipped due to invalid timestamp
			},
			wantErr: false, // Function returns nil even if some delegations are skipped
		},
		{
			name: "Missing_Sender_Address",
			data: []domain.TzktApiDelegationsResponse{
				{
					Type:        "delegation",
					Status:      "applied",
					Timestamp:   "2023-01-01T12:00:00Z",
					Level:       1000,
					Hash:        "ophash123",
					Amount:      100000,
					Sender:      nil, // Missing sender
					NewDelegate: &domain.Account{Address: "tz1baker"},
				},
			},
			setupMocks: func(repo *mocks.MockRepository) {
				// Should not call Create since delegation is skipped due to missing sender
			},
			wantErr: false,
		},
		{
			name: "Empty_Sender_Address",
			data: []domain.TzktApiDelegationsResponse{
				{
					Type:      "delegation",
					Status:    "applied",
					Timestamp: "2023-01-01T12:00:00Z",
					Level:     1000,
					Hash:      "ophash123",
					Amount:    100000,
					Sender: &domain.Account{
						Address: "", // Empty address
					},
					NewDelegate: &domain.Account{Address: "tz1baker"},
				},
			},
			setupMocks: func(repo *mocks.MockRepository) {
				// Should not call Create since delegation is skipped due to empty sender address
			},
			wantErr: false,
		},
		{
			name: "Non_Delegation_Type",
			data: []domain.TzktApiDelegationsResponse{
				{
					Type:      "origination",
					Status:    "applied",
					Timestamp: "2023-01-01T12:00:00Z",
					Level:     1000,
					Hash:      "ophash123",
					Amount:    100000,
				},
			},
			setupMocks: func(repo *mocks.MockRepository) {
				// Should not call Create since type is not "delegation"
			},
			wantErr: false,
		},
		{
			name: "Non_Applied_Status",
			data: []domain.TzktApiDelegationsResponse{
				{
					Type:      "delegation",
					Status:    "failed",
					Timestamp: "2023-01-01T12:00:00Z",
					Level:     1000,
					Hash:      "ophash123",
					Amount:    100000,
				},
			},
			setupMocks: func(repo *mocks.MockRepository) {
				// Should not call Create since status is not "applied"
			},
			wantErr: false,
		},
		{
			name: "Repository_Error",
			data: []domain.TzktApiDelegationsResponse{
				{
					Type:      "delegation",
					Status:    "applied",
					Timestamp: "2023-01-01T12:00:00Z",
					Level:     1000,
					Hash:      "ophash123",
					Amount:    100000,
					Sender: &domain.Account{
						Address: "tz1delegator",
					},
					NewDelegate: &domain.Account{
						Address: "tz1baker",
					},
				},
			},
			setupMocks: func(repo *mocks.MockRepository) {
				repo.EXPECT().Create(mock.Anything, mock.Anything).Return(errors.New("database error")).Once()
			},
			wantErr:     true,
			expectedErr: "database error",
		},
		{
			name: "Empty_Data_List",
			data: []domain.TzktApiDelegationsResponse{},
			setupMocks: func(repo *mocks.MockRepository) {
				// Should not call Create since data is empty
			},
			wantErr: false,
		},
		{
			name: "Mixed_Valid_Invalid_Delegations",
			data: []domain.TzktApiDelegationsResponse{
				{
					Type:      "delegation",
					Status:    "applied",
					Timestamp: "2023-01-01T12:00:00Z",
					Level:     1000,
					Hash:      "ophash123",
					Amount:    100000,
					Sender: &domain.Account{
						Address: "tz1delegator",
					},
					NewDelegate: &domain.Account{
						Address: "tz1baker",
					},
				},
				{
					Type:      "origination", // Invalid type
					Status:    "applied",
					Timestamp: "2023-01-01T12:00:00Z",
					Level:     1001,
					Hash:      "ophash124",
					Amount:    200000,
				},
				{
					Type:      "delegation",
					Status:    "applied",
					Timestamp: "2023-01-01T12:00:00Z",
					Level:     1002,
					Hash:      "ophash125",
					Amount:    300000,
					Sender: &domain.Account{
						Address: "tz1delegator2",
					},
					NewDelegate: &domain.Account{
						Address: "tz1baker2",
					},
				},
			},
			setupMocks: func(repo *mocks.MockRepository) {
				repo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(dtos []domain.CreateDelegationDTO) bool {
					// Should only have 2 valid delegations (skipping the origination)
					return len(dtos) == 2
				})).Return(nil).Once()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
			mockRepo := mocks.NewMockRepository(t)

			if tt.setupMocks != nil {
				tt.setupMocks(mockRepo)
			}

			uc := &UseCaseImpl{
				logger:     logger,
				repository: mockRepo,
			}

			err := uc.Create(context.Background(), tt.data)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != "" {
					assert.Contains(t, err.Error(), tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUseCaseImpl_GetDelegations_Comprehensive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupMocks     func(*mocks.MockRepository)
		expectedResult domain.ApiResponse[domain.DelegationsResponseType]
		wantErr        bool
		expectedErr    string
	}{
		{
			name: "Success_With_Delegations",
			setupMocks: func(repo *mocks.MockRepository) {
				delegations := []models.Delegation{
					{
						ID:        uuid.New(),
						Delegator: "tz1delegator1",
						Amount:    100000,
						Timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						Level:     1000,
					},
					{
						ID:        uuid.New(),
						Delegator: "tz1delegator2",
						Amount:    200000,
						Timestamp: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						Level:     1001,
					},
				}
				repo.EXPECT().FindAll(mock.Anything).Return(delegations, nil).Once()
			},
			expectedResult: domain.ApiResponse[domain.DelegationsResponseType]{
				Data: []domain.DelegationsResponseType{
					{
						Timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						Amount:    100000,
						Delegator: "tz1delegator1",
						Level:     1000,
					},
					{
						Timestamp: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						Amount:    200000,
						Delegator: "tz1delegator2",
						Level:     1001,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Success_Empty_Result",
			setupMocks: func(repo *mocks.MockRepository) {
				repo.EXPECT().FindAll(mock.Anything).Return([]models.Delegation{}, nil).Once()
			},
			expectedResult: domain.ApiResponse[domain.DelegationsResponseType]{
				Data: []domain.DelegationsResponseType{},
			},
			wantErr: false,
		},
		{
			name: "Repository_Error",
			setupMocks: func(repo *mocks.MockRepository) {
				repo.EXPECT().FindAll(mock.Anything).Return(nil, errors.New("database connection failed")).Once()
			},
			expectedResult: domain.ApiResponse[domain.DelegationsResponseType]{},
			wantErr:        true,
			expectedErr:    "database connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
			mockRepo := mocks.NewMockRepository(t)

			if tt.setupMocks != nil {
				tt.setupMocks(mockRepo)
			}

			uc := &UseCaseImpl{
				logger:     logger,
				repository: mockRepo,
			}

			result, err := uc.GetDelegations(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != "" {
					assert.Contains(t, err.Error(), tt.expectedErr)
				}
				assert.Equal(t, tt.expectedResult, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestUseCaseOptions(t *testing.T) {
	t.Parallel()

	t.Run("UseCaseWithLogger", func(t *testing.T) {
		logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
		uc := &UseCaseImpl{}

		option := UseCaseWithLogger(logger)
		option(uc)

		assert.Equal(t, logger, uc.logger)
	})

	t.Run("UseCaseWithRepository", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)
		uc := &UseCaseImpl{}

		option := UseCaseWithRepository(mockRepo)
		option(uc)

		assert.Equal(t, mockRepo, uc.repository)
	})
}

func TestNewUseCase_Comprehensive(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	mockRepo := mocks.NewMockRepository(t)

	tests := []struct {
		name    string
		options []UseCaseOption
		checks  func(*testing.T, *UseCaseImpl)
	}{
		{
			name: "With_All_Options",
			options: []UseCaseOption{
				UseCaseWithLogger(logger),
				UseCaseWithRepository(mockRepo),
			},
			checks: func(t *testing.T, uc *UseCaseImpl) {
				assert.Equal(t, logger, uc.logger)
				assert.Equal(t, mockRepo, uc.repository)
			},
		},
		{
			name:    "With_No_Options",
			options: []UseCaseOption{},
			checks: func(t *testing.T, uc *UseCaseImpl) {
				assert.Nil(t, uc.logger)
				assert.Nil(t, uc.repository)
			},
		},
		{
			name: "With_Logger_Only",
			options: []UseCaseOption{
				UseCaseWithLogger(logger),
			},
			checks: func(t *testing.T, uc *UseCaseImpl) {
				assert.Equal(t, logger, uc.logger)
				assert.Nil(t, uc.repository)
			},
		},
		{
			name: "With_Repository_Only",
			options: []UseCaseOption{
				UseCaseWithRepository(mockRepo),
			},
			checks: func(t *testing.T, uc *UseCaseImpl) {
				assert.Nil(t, uc.logger)
				assert.Equal(t, mockRepo, uc.repository)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc := NewUseCase(tt.options...)

			assert.NotNil(t, uc)
			if tt.checks != nil {
				tt.checks(t, uc)
			}
		})
	}
}
