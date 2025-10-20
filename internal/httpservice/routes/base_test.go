package routes

import (
	"delegator/mocks"
	"delegator/pkg/domain"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegisterBaseRoutes(t *testing.T) {
	t.Parallel()

	type args struct {
		setupMocks func() (domain.UseCase, *slog.Logger)
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "Register_Base_Routes_Success",
			args: args{
				setupMocks: func() (domain.UseCase, *slog.Logger) {
					mockUseCase := mocks.NewMockUseCase(t)
					logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
					return mockUseCase, logger
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)
			router := gin.New()
			useCase, logger := tt.args.setupMocks()

			assert.NotPanics(t, func() {
				RegisterBaseRoutes(router, logger, useCase)
			})

			// Verify routes are registered
			routes := router.Routes()
			assert.Len(t, routes, 2) // health + delegations endpoints

			// Check health endpoint exists
			healthFound := false
			delegationsFound := false
			for _, route := range routes {
				if route.Path == "/health" && route.Method == "GET" {
					healthFound = true
				}
				if route.Path == "/xtz/delegations" && route.Method == "GET" {
					delegationsFound = true
				}
			}
			assert.True(t, healthFound)
			assert.True(t, delegationsFound)
		})
	}
}

func TestHealthEndpoint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Health_Check_Success",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"data":"ok"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)
			router := gin.New()
			mockUseCase := mocks.NewMockUseCase(t)
			logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

			RegisterBaseRoutes(router, logger, mockUseCase)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/health", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestDelegationsEndpoint(t *testing.T) {
	t.Parallel()

	type args struct {
		setupMocks func() domain.UseCase
	}

	tests := []struct {
		name           string
		args           args
		expectedStatus int
		checkResponse  bool
	}{
		{
			name: "Delegations_Success",
			args: args{
				setupMocks: func() domain.UseCase {
					mockUseCase := mocks.NewMockUseCase(t)
					response := domain.ApiResponse[domain.DelegationsResponseType]{
						Data: []domain.DelegationsResponseType{
							{
								Timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
								Amount:    100000,
								Delegator: "tz1delegator1",
								Level:     1000,
							},
						},
					}
					mockUseCase.EXPECT().GetDelegations(mock.Anything).Return(response, nil).Once()
					return mockUseCase
				},
			},
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name: "Delegations_Error",
			args: args{
				setupMocks: func() domain.UseCase {
					mockUseCase := mocks.NewMockUseCase(t)
					mockUseCase.EXPECT().GetDelegations(mock.Anything).Return(
						domain.ApiResponse[domain.DelegationsResponseType]{}, 
						errors.New("database error"),
					).Once()
					return mockUseCase
				},
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)
			router := gin.New()
			useCase := tt.args.setupMocks()
			logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

			RegisterBaseRoutes(router, logger, useCase)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/xtz/delegations", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkResponse && tt.expectedStatus == http.StatusOK {
				assert.Contains(t, w.Body.String(), "tz1delegator1")
				assert.Contains(t, w.Body.String(), "100000")
			}

			if tt.expectedStatus == http.StatusInternalServerError {
				assert.Contains(t, w.Body.String(), "failed to get delegations")
			}
		})
	}
}

func TestCreateDelegatorRegistrar(t *testing.T) {
	t.Parallel()

	type args struct {
		logger   *slog.Logger
		useCase  domain.UseCase
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "Create_Delegator_Registrar_Success",
			args: args{
				logger:  slog.New(slog.NewJSONHandler(os.Stdout, nil)),
				useCase: mocks.NewMockUseCase(t),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			registrar := CreateDelegatorRegistrar(tt.args.logger, tt.args.useCase)
			assert.NotNil(t, registrar)

			// Test that the registrar can be called
			gin.SetMode(gin.TestMode)
			engine := gin.New()

			assert.NotPanics(t, func() {
				registrar(engine)
			})

			// Verify routes are registered
			routes := engine.Routes()
			assert.Len(t, routes, 2) // health + delegations endpoints
		})
	}
}

func TestRouteRegistrar_Integration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "Full_Integration_Test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)
			engine := gin.New()
			logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
			mockUseCase := mocks.NewMockUseCase(t)

			// Mock successful response
			response := domain.ApiResponse[domain.DelegationsResponseType]{
				Data: []domain.DelegationsResponseType{
					{
						Timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						Amount:    100000,
						Delegator: "tz1delegator1",
						Level:     1000,
					},
				},
			}
			mockUseCase.EXPECT().GetDelegations(mock.Anything).Return(response, nil).Once()

			// Create and register routes
			registrar := CreateDelegatorRegistrar(logger, mockUseCase)
			registrar(engine)

			// Test health endpoint
			w1 := httptest.NewRecorder()
			req1, _ := http.NewRequest("GET", "/health", nil)
			engine.ServeHTTP(w1, req1)
			assert.Equal(t, http.StatusOK, w1.Code)

			// Test delegations endpoint
			w2 := httptest.NewRecorder()
			req2, _ := http.NewRequest("GET", "/xtz/delegations", nil)
			engine.ServeHTTP(w2, req2)
			assert.Equal(t, http.StatusOK, w2.Code)
		})
	}
}