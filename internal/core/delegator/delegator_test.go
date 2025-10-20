package delegator

import (
	"context"
	"delegator/internal/models"
	"delegator/mocks"
	"delegator/pkg/domain"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateNewDelegator(t *testing.T) {
	t.Parallel()

	type args struct {
		opts []Option
	}

	tests := []struct {
		name    string
		args    args
		wantNil bool
	}{
		{
			name: "Create_Delegator_With_Logger",
			args: args{opts: []Option{
				WithLogger(
					slog.New(
						slog.NewJSONHandler(os.Stdout, nil),
					),
				),
			}},
			wantNil: false,
		},
		{
			name:    "Create_Delegator_Without_Options",
			args:    args{opts: []Option{}},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := NewDelegator(tt.args.opts...)

			if tt.wantNil {
				assert.Nil(t, got)
			} else {
				assert.NotNil(t, got)
				assert.Equal(t, "delegator", got.Name())
			}
		})
	}
}

func TestUseCaseImpl_Create(t *testing.T) {
	t.Parallel()

	type fields struct {
		logger     *slog.Logger
		repository func(t *testing.T) domain.Repository
	}
	type args struct {
		ctx  context.Context
		data []domain.TzktApiDelegationsResponse
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Create_Valid_Delegations",
			fields: fields{
				logger: slog.New(slog.NewJSONHandler(os.Stdout, nil)),
				repository: func(t *testing.T) domain.Repository {
					mockRepo := mocks.NewMockRepository(t)
					mockRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(dtos []domain.CreateDelegationDTO) bool {
						return len(dtos) == 1 && dtos[0].Delegation.Delegator == "tz1delegator"
					})).Return(nil).Once()
					return mockRepo
				},
			},
			args: args{
				ctx: context.Background(),
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
						PrevDelegate: nil,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Create_Skip_Non_Delegation_Type",
			fields: fields{
				logger: slog.New(slog.NewJSONHandler(os.Stdout, nil)),
				repository: func(t *testing.T) domain.Repository {
					mockRepo := mocks.NewMockRepository(t)
					return mockRepo
				},
			},
			args: args{
				ctx: context.Background(),
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
			},
			wantErr: false,
		},
		{
			name: "Create_Skip_Non_Applied_Status",
			fields: fields{
				logger: slog.New(slog.NewJSONHandler(os.Stdout, nil)),
				repository: func(t *testing.T) domain.Repository {
					mockRepo := mocks.NewMockRepository(t)
					return mockRepo
				},
			},
			args: args{
				ctx: context.Background(),
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
			},
			wantErr: false,
		},
		{
			name: "Create_Undelegation",
			fields: fields{
				logger: slog.New(slog.NewJSONHandler(os.Stdout, nil)),
				repository: func(t *testing.T) domain.Repository {
					mockRepo := mocks.NewMockRepository(t)
					mockRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(dtos []domain.CreateDelegationDTO) bool {
						return len(dtos) == 1 && dtos[0].Baker.Address == "UNDELEGATED"
					})).Return(nil).Once()
					return mockRepo
				},
			},
			args: args{
				ctx: context.Background(),
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
						NewDelegate:  nil,
						PrevDelegate: &domain.Account{Address: "tz1oldbaker"},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc := &UseCaseImpl{
				logger:     tt.fields.logger,
				repository: tt.fields.repository(t),
			}

			err := uc.Create(tt.args.ctx, tt.args.data)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUseCaseImpl_GetDelegations(t *testing.T) {
	t.Parallel()

	type fields struct {
		logger     *slog.Logger
		repository func(t *testing.T) domain.Repository
	}
	type args struct {
		ctx context.Context
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    domain.ApiResponse[domain.DelegationsResponseType]
		wantErr bool
	}{
		{
			name: "GetDelegations_Success",
			fields: fields{
				logger: slog.New(slog.NewJSONHandler(os.Stdout, nil)),
				repository: func(t *testing.T) domain.Repository {
					mockRepo := mocks.NewMockRepository(t)
					delegations := []models.Delegation{
						{
							ID:        uuid.New(),
							Delegator: "tz1delegator",
							Amount:    100000,
							Timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
							Level:     1000,
						},
					}
					mockRepo.EXPECT().FindAll(mock.Anything).Return(delegations, nil).Once()
					return mockRepo
				},
			},
			args: args{
				ctx: context.Background(),
			},
			want: domain.ApiResponse[domain.DelegationsResponseType]{
				Data: []domain.DelegationsResponseType{
					{
						Timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						Amount:    100000,
						Delegator: "tz1delegator",
						Level:     1000,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "GetDelegations_Repository_Error",
			fields: fields{
				logger: slog.New(slog.NewJSONHandler(os.Stdout, nil)),
				repository: func(t *testing.T) domain.Repository {
					mockRepo := mocks.NewMockRepository(t)
					mockRepo.EXPECT().FindAll(mock.Anything).Return(nil, assert.AnError).Once()
					return mockRepo
				},
			},
			args: args{
				ctx: context.Background(),
			},
			want:    domain.ApiResponse[domain.DelegationsResponseType]{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc := &UseCaseImpl{
				logger:     tt.fields.logger,
				repository: tt.fields.repository(t),
			}

			got, err := uc.GetDelegations(tt.args.ctx)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.want, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestNewUseCase(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	mockRepo := mocks.NewMockRepository(t)

	uc := NewUseCase(
		UseCaseWithLogger(logger),
		UseCaseWithRepository(mockRepo),
	)

	assert.NotNil(t, uc)
	assert.Equal(t, logger, uc.logger)
	assert.Equal(t, mockRepo, uc.repository)
}
