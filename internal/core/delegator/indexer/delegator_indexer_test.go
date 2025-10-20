package indexer

import (
	"context"
	"delegator/mocks"
	"delegator/pkg/domain"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewDelegatorIndexer(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	mockUseCase := mocks.NewMockUseCase(t)
	mockDelegationHandler := mocks.NewMockDelegationService(t)
	mockRepository := mocks.NewMockRepository(t)

	indexer := NewDelegatorIndexer(
		WithLogger(logger),
		WithDelegatorUseCase(mockUseCase),
		WithDelegationHandler(mockDelegationHandler),
		WithRepository(mockRepository),
	)

	assert.NotNil(t, indexer)
	assert.Equal(t, logger, indexer.logger)
	assert.Equal(t, mockUseCase, indexer.delegatorUseCase)
	assert.Equal(t, mockDelegationHandler, indexer.DelegationHandler)
	assert.Equal(t, mockRepository, indexer.repository)
}

func TestDelegatorIndexer_indexOnce_EmptyDatabase(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	mockUseCase := mocks.NewMockUseCase(t)
	mockDelegationHandler := mocks.NewMockDelegationService(t)
	mockRepository := mocks.NewMockRepository(t)

	indexer := &DelegatorIndexer{
		logger:            logger,
		delegatorUseCase:  mockUseCase,
		DelegationHandler: mockDelegationHandler,
		repository:        mockRepository,
	}

	ctx := context.Background()
	testData := []domain.TzktApiDelegationsResponse{
		{
			Type:      "delegation",
			Status:    "applied",
			Timestamp: "2023-01-01T12:00:00Z",
			Level:     1000,
			Hash:      "ophash123",
			Amount:    100000,
		},
	}

	mockRepository.EXPECT().CountDelegations(ctx).Return(int64(0), nil).Once()
	mockDelegationHandler.EXPECT().GetDelegationsFromLevel(int64(0), 1000).Return(testData, nil).Once()
	mockUseCase.EXPECT().Create(ctx, testData).Return(nil).Once()

	err := indexer.indexOnce(ctx)
	assert.NoError(t, err)
}

func TestDelegatorIndexer_indexOnce_ExistingData(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	mockUseCase := mocks.NewMockUseCase(t)
	mockDelegationHandler := mocks.NewMockDelegationService(t)
	mockRepository := mocks.NewMockRepository(t)

	indexer := &DelegatorIndexer{
		logger:            logger,
		delegatorUseCase:  mockUseCase,
		DelegationHandler: mockDelegationHandler,
		repository:        mockRepository,
	}

	ctx := context.Background()
	testData := []domain.TzktApiDelegationsResponse{
		{
			Type:      "delegation",
			Status:    "applied",
			Timestamp: "2023-01-01T12:00:00Z",
			Level:     1001,
			Hash:      "ophash124",
			Amount:    200000,
		},
	}

	mockRepository.EXPECT().CountDelegations(ctx).Return(int64(5), nil).Once()
	mockRepository.EXPECT().GetLastProcessedLevel(ctx).Return(int64(1000), nil).Once()
	mockDelegationHandler.EXPECT().GetDelegationsFromLevel(int64(1000), 100).Return(testData, nil).Once()
	mockUseCase.EXPECT().Create(ctx, testData).Return(nil).Once()

	err := indexer.indexOnce(ctx)
	assert.NoError(t, err)
}

func TestDelegatorIndexer_indexOnce_NoNewData(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	mockUseCase := mocks.NewMockUseCase(t)
	mockDelegationHandler := mocks.NewMockDelegationService(t)
	mockRepository := mocks.NewMockRepository(t)

	indexer := &DelegatorIndexer{
		logger:            logger,
		delegatorUseCase:  mockUseCase,
		DelegationHandler: mockDelegationHandler,
		repository:        mockRepository,
	}

	ctx := context.Background()

	mockRepository.EXPECT().CountDelegations(ctx).Return(int64(5), nil).Once()
	mockRepository.EXPECT().GetLastProcessedLevel(ctx).Return(int64(1000), nil).Once()
	mockDelegationHandler.EXPECT().GetDelegationsFromLevel(int64(1000), 100).Return([]domain.TzktApiDelegationsResponse{}, nil).Once()

	err := indexer.indexOnce(ctx)
	assert.NoError(t, err)
}

func TestDelegatorIndexer_indexOnce_CountError(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	mockUseCase := mocks.NewMockUseCase(t)
	mockDelegationHandler := mocks.NewMockDelegationService(t)
	mockRepository := mocks.NewMockRepository(t)

	indexer := &DelegatorIndexer{
		logger:            logger,
		delegatorUseCase:  mockUseCase,
		DelegationHandler: mockDelegationHandler,
		repository:        mockRepository,
	}

	ctx := context.Background()
	expectedError := errors.New("database error")

	mockRepository.EXPECT().CountDelegations(ctx).Return(int64(0), expectedError).Once()

	err := indexer.indexOnce(ctx)
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
}

func TestDelegatorIndexer_indexOnce_GetLastLevelError(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	mockUseCase := mocks.NewMockUseCase(t)
	mockDelegationHandler := mocks.NewMockDelegationService(t)
	mockRepository := mocks.NewMockRepository(t)

	indexer := &DelegatorIndexer{
		logger:            logger,
		delegatorUseCase:  mockUseCase,
		DelegationHandler: mockDelegationHandler,
		repository:        mockRepository,
	}

	ctx := context.Background()
	expectedError := errors.New("get level error")

	// Mock expectations for get last level error
	mockRepository.EXPECT().CountDelegations(ctx).Return(int64(5), nil).Once()
	mockRepository.EXPECT().GetLastProcessedLevel(ctx).Return(int64(0), expectedError).Once()

	err := indexer.indexOnce(ctx)
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
}

func TestDelegatorIndexer_indexOnce_DelegationHandlerError(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	mockUseCase := mocks.NewMockUseCase(t)
	mockDelegationHandler := mocks.NewMockDelegationService(t)
	mockRepository := mocks.NewMockRepository(t)

	indexer := &DelegatorIndexer{
		logger:            logger,
		delegatorUseCase:  mockUseCase,
		DelegationHandler: mockDelegationHandler,
		repository:        mockRepository,
	}

	ctx := context.Background()
	expectedError := errors.New("delegation handler error")

	mockRepository.EXPECT().CountDelegations(ctx).Return(int64(0), nil).Once()
	mockDelegationHandler.EXPECT().GetDelegationsFromLevel(int64(0), 1000).Return(nil, expectedError).Once()

	err := indexer.indexOnce(ctx)
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
}

func TestDelegatorIndexer_indexOnce_UseCaseError(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	mockUseCase := mocks.NewMockUseCase(t)
	mockDelegationHandler := mocks.NewMockDelegationService(t)
	mockRepository := mocks.NewMockRepository(t)

	indexer := &DelegatorIndexer{
		logger:            logger,
		delegatorUseCase:  mockUseCase,
		DelegationHandler: mockDelegationHandler,
		repository:        mockRepository,
	}

	ctx := context.Background()
	testData := []domain.TzktApiDelegationsResponse{
		{
			Type:   "delegation",
			Status: "applied",
			Level:  1000,
		},
	}
	expectedError := errors.New("use case error")

	mockRepository.EXPECT().CountDelegations(ctx).Return(int64(0), nil).Once()
	mockDelegationHandler.EXPECT().GetDelegationsFromLevel(int64(0), 1000).Return(testData, nil).Once()
	mockUseCase.EXPECT().Create(ctx, testData).Return(expectedError).Once()

	err := indexer.indexOnce(ctx)
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
}

func TestDelegatorIndexer_Shutdown(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	indexer := &DelegatorIndexer{
		logger: logger,
	}

	ctx := context.Background()
	err := indexer.Shutdown(ctx)
	assert.NoError(t, err)
}

func TestDelegatorIndexer_Run_CancellationContext(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	mockUseCase := mocks.NewMockUseCase(t)
	mockDelegationHandler := mocks.NewMockDelegationService(t)
	mockRepository := mocks.NewMockRepository(t)

	indexer := &DelegatorIndexer{
		logger:            logger,
		delegatorUseCase:  mockUseCase,
		DelegationHandler: mockDelegationHandler,
		repository:        mockRepository,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	mockRepository.EXPECT().CountDelegations(mock.Anything).Return(int64(0), nil).Maybe()
	mockDelegationHandler.EXPECT().GetDelegationsFromLevel(mock.Anything, mock.Anything).Return([]domain.TzktApiDelegationsResponse{}, nil).Maybe()

	err := indexer.Run(ctx)
	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
}

func TestIndexerOptions(t *testing.T) {
	t.Parallel()

	t.Run("WithLogger", func(t *testing.T) {
		logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
		indexer := &DelegatorIndexer{}

		option := WithLogger(logger)
		option(indexer)

		assert.Equal(t, logger, indexer.logger)
	})

	t.Run("WithDelegatorUseCase", func(t *testing.T) {
		mockUseCase := mocks.NewMockUseCase(t)
		indexer := &DelegatorIndexer{}

		option := WithDelegatorUseCase(mockUseCase)
		option(indexer)

		assert.Equal(t, mockUseCase, indexer.delegatorUseCase)
	})

	t.Run("WithDelegationHandler", func(t *testing.T) {
		mockHandler := mocks.NewMockDelegationService(t)
		indexer := &DelegatorIndexer{}

		option := WithDelegationHandler(mockHandler)
		option(indexer)

		assert.Equal(t, mockHandler, indexer.DelegationHandler)
	})

	t.Run("WithRepository", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)
		indexer := &DelegatorIndexer{}

		option := WithRepository(mockRepo)
		option(indexer)

		assert.Equal(t, mockRepo, indexer.repository)
	})
}
