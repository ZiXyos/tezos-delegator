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
	"gorm.io/gorm"
)

func TestNewRepository(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	var db *gorm.DB = nil

	repo := NewRepository(
		RepositoryWithLogger(logger),
		RepositoryWithDBClient(db),
	)

	assert.NotNil(t, repo)
	assert.Equal(t, logger, repo.logger)
	assert.Equal(t, db, repo.dbClient)
}

func TestRepositoryWithLogger(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	repo := &Repository{}
	
	option := RepositoryWithLogger(logger)
	option(repo)
	
	assert.Equal(t, logger, repo.logger)
}

func TestRepositoryWithDBClient(t *testing.T) {
	t.Parallel()

	var db *gorm.DB = nil
	repo := &Repository{}
	
	option := RepositoryWithDBClient(db)
	option(repo)
	
	assert.Equal(t, db, repo.dbClient)
}


// Test the Repository interface behavior using mocks
func TestRepositoryInterface_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		delegations    []domain.CreateDelegationDTO
		mockSetup      func(*mocks.MockRepository)
		expectedError  error
	}{
		{
			name: "Success_Single_Delegation",
			delegations: []domain.CreateDelegationDTO{
				{
					Baker: models.Baker{
						Address:   "tz1baker",
						FirstSeen: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						LastSeen:  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					},
					Delegation: models.Delegation{
						Delegator: "tz1delegator",
						BakerID:   "tz1baker",
						Amount:    100000,
						Level:     1000,
					},
				},
			},
			mockSetup: func(repo *mocks.MockRepository) {
				repo.EXPECT().Create(context.Background(), 
					[]domain.CreateDelegationDTO{
						{
							Baker: models.Baker{
								Address:   "tz1baker",
								FirstSeen: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
								LastSeen:  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
							},
							Delegation: models.Delegation{
								Delegator: "tz1delegator",
								BakerID:   "tz1baker",
								Amount:    100000,
								Level:     1000,
							},
						},
					}).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name: "Error_Database_Failure",
			delegations: []domain.CreateDelegationDTO{
				{
					Baker: models.Baker{Address: "tz1baker"},
					Delegation: models.Delegation{
						Delegator: "tz1delegator",
						BakerID:   "tz1baker",
					},
				},
			},
			mockSetup: func(repo *mocks.MockRepository) {
				repo.EXPECT().Create(context.Background(), 
					[]domain.CreateDelegationDTO{
						{
							Baker: models.Baker{Address: "tz1baker"},
							Delegation: models.Delegation{
								Delegator: "tz1delegator",
								BakerID:   "tz1baker",
							},
						},
					}).Return(errors.New("database error")).Once()
			},
			expectedError: errors.New("database error"),
		},
		{
			name:        "Success_Empty_List",
			delegations: []domain.CreateDelegationDTO{},
			mockSetup: func(repo *mocks.MockRepository) {
				repo.EXPECT().Create(context.Background(), 
					[]domain.CreateDelegationDTO{}).Return(nil).Once()
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := mocks.NewMockRepository(t)
			tt.mockSetup(mockRepo)

			err := mockRepo.Create(context.Background(), tt.delegations)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRepositoryInterface_FindAll(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		mockSetup       func(*mocks.MockRepository)
		expectedResult  []models.Delegation
		expectedError   error
	}{
		{
			name: "Success_With_Delegations",
			mockSetup: func(repo *mocks.MockRepository) {
				delegations := []models.Delegation{
					{
						ID:        uuid.New(),
						Delegator: "tz1delegator1",
						BakerID:   "tz1baker1",
						Amount:    100000,
						Timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						Level:     1000,
					},
					{
						ID:        uuid.New(),
						Delegator: "tz1delegator2", 
						BakerID:   "tz1baker2",
						Amount:    200000,
						Timestamp: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						Level:     1001,
					},
				}
				repo.EXPECT().FindAll(context.Background()).Return(delegations, nil).Once()
			},
			expectedResult: []models.Delegation{
				{
					ID:        uuid.New(),
					Delegator: "tz1delegator1",
					BakerID:   "tz1baker1", 
					Amount:    100000,
					Timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					Level:     1000,
				},
				{
					ID:        uuid.New(),
					Delegator: "tz1delegator2",
					BakerID:   "tz1baker2",
					Amount:    200000,
					Timestamp: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
					Level:     1001,
				},
			},
			expectedError: nil,
		},
		{
			name: "Success_Empty_Result",
			mockSetup: func(repo *mocks.MockRepository) {
				repo.EXPECT().FindAll(context.Background()).Return([]models.Delegation{}, nil).Once()
			},
			expectedResult: []models.Delegation{},
			expectedError:  nil,
		},
		{
			name: "Error_Database_Failure",
			mockSetup: func(repo *mocks.MockRepository) {
				repo.EXPECT().FindAll(context.Background()).Return(nil, errors.New("connection failed")).Once()
			},
			expectedResult: nil,
			expectedError:  errors.New("connection failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := mocks.NewMockRepository(t)
			tt.mockSetup(mockRepo)

			result, err := mockRepo.FindAll(context.Background())

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, len(tt.expectedResult))
			}
		})
	}
}

func TestRepositoryInterface_CountDelegations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		mockSetup      func(*mocks.MockRepository)
		expectedCount  int64
		expectedError  error
	}{
		{
			name: "Success_With_Count",
			mockSetup: func(repo *mocks.MockRepository) {
				repo.EXPECT().CountDelegations(context.Background()).Return(int64(42), nil).Once()
			},
			expectedCount: 42,
			expectedError: nil,
		},
		{
			name: "Success_Zero_Count",
			mockSetup: func(repo *mocks.MockRepository) {
				repo.EXPECT().CountDelegations(context.Background()).Return(int64(0), nil).Once()
			},
			expectedCount: 0,
			expectedError: nil,
		},
		{
			name: "Error_Database_Failure",
			mockSetup: func(repo *mocks.MockRepository) {
				repo.EXPECT().CountDelegations(context.Background()).Return(int64(0), errors.New("count query failed")).Once()
			},
			expectedCount: 0,
			expectedError: errors.New("count query failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := mocks.NewMockRepository(t)
			tt.mockSetup(mockRepo)

			count, err := mockRepo.CountDelegations(context.Background())

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedCount, count)
		})
	}
}

func TestRepositoryInterface_GetLastProcessedLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		mockSetup      func(*mocks.MockRepository)
		expectedLevel  int64
		expectedError  error
	}{
		{
			name: "Success_With_Level",
			mockSetup: func(repo *mocks.MockRepository) {
				repo.EXPECT().GetLastProcessedLevel(context.Background()).Return(int64(1000), nil).Once()
			},
			expectedLevel: 1000,
			expectedError: nil,
		},
		{
			name: "Success_Zero_Level",
			mockSetup: func(repo *mocks.MockRepository) {
				repo.EXPECT().GetLastProcessedLevel(context.Background()).Return(int64(0), nil).Once()
			},
			expectedLevel: 0,
			expectedError: nil,
		},
		{
			name: "Error_Database_Failure",
			mockSetup: func(repo *mocks.MockRepository) {
				repo.EXPECT().GetLastProcessedLevel(context.Background()).Return(int64(0), errors.New("level query failed")).Once()
			},
			expectedLevel: 0,
			expectedError: errors.New("level query failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := mocks.NewMockRepository(t)
			tt.mockSetup(mockRepo)

			level, err := mockRepo.GetLastProcessedLevel(context.Background())

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedLevel, level)
		})
	}
}