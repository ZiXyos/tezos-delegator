package services

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHTTPHandler(t *testing.T) {
	t.Parallel()

	type args struct {
		opts []HandlerOptions
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "Create_Handler_Without_Options",
			args: args{opts: []HandlerOptions{}},
		},
		{
			name: "Create_Handler_With_Logger",
			args: args{opts: []HandlerOptions{
				HandlerWithLogger(slog.New(slog.NewJSONHandler(os.Stdout, nil))),
			}},
		},
		{
			name: "Create_Handler_With_Client",
			args: args{opts: []HandlerOptions{
				HandlerWithClient(&http.Client{}),
			}},
		},
		{
			name: "Create_Handler_With_BaseURL",
			args: args{opts: []HandlerOptions{
				HandlerWithBaseURL("https://api.tzkt.io/v1/"),
			}},
		},
		{
			name: "Create_Handler_With_All_Options",
			args: args{opts: []HandlerOptions{
				HandlerWithLogger(slog.New(slog.NewJSONHandler(os.Stdout, nil))),
				HandlerWithClient(&http.Client{}),
				HandlerWithBaseURL("https://api.tzkt.io/v1/"),
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := NewHTTPHandler(tt.args.opts...)
			assert.NotNil(t, handler)
		})
	}
}

func TestHandlerWithLogger(t *testing.T) {
	t.Parallel()

	type args struct {
		logger *slog.Logger
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "Set_Logger_Option",
			args: args{logger: slog.New(slog.NewJSONHandler(os.Stdout, nil))},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := &HTTPHandler{}
			option := HandlerWithLogger(tt.args.logger)
			option(handler)

			assert.Equal(t, tt.args.logger, handler.logger)
		})
	}
}

func TestHandlerWithClient(t *testing.T) {
	t.Parallel()

	type args struct {
		client *http.Client
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "Set_Client_Option",
			args: args{client: &http.Client{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := &HTTPHandler{}
			option := HandlerWithClient(tt.args.client)
			option(handler)

			assert.Equal(t, tt.args.client, handler.client)
		})
	}
}

func TestHandlerWithBaseURL(t *testing.T) {
	t.Parallel()

	type args struct {
		baseURL string
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "Set_BaseURL_Option",
			args: args{baseURL: "https://api.tzkt.io/v1/"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := &HTTPHandler{}
			option := HandlerWithBaseURL(tt.args.baseURL)
			option(handler)

			assert.Equal(t, tt.args.baseURL, handler.baseURL)
		})
	}
}

func TestHTTPHandler_GetDelegations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		mockResponse   string
		mockStatusCode int
		expectedError  bool
		expectedCount  int
	}{
		{
			name: "Get_Delegations_Success",
			mockResponse: `[
				{
					"type": "delegation",
					"status": "applied",
					"timestamp": "2023-01-01T12:00:00Z",
					"level": 1000,
					"hash": "ophash123",
					"amount": 100000,
					"sender": {"address": "tz1delegator"},
					"newDelegate": {"address": "tz1baker"}
				}
			]`,
			mockStatusCode: http.StatusOK,
			expectedError:  false,
			expectedCount:  1,
		},
		{
			name:           "Get_Delegations_Empty_Response",
			mockResponse:   `[]`,
			mockStatusCode: http.StatusOK,
			expectedError:  false,
			expectedCount:  0,
		},
		{
			name:           "Get_Delegations_Server_Error",
			mockResponse:   "",
			mockStatusCode: http.StatusInternalServerError,
			expectedError:  true,
			expectedCount:  0,
		},
		{
			name:           "Get_Delegations_Invalid_JSON",
			mockResponse:   `invalid json`,
			mockStatusCode: http.StatusOK,
			expectedError:  true,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.mockStatusCode)
				if tt.mockResponse != "" {
					w.Write([]byte(tt.mockResponse))
				}
			}))
			defer server.Close()

			handler := NewHTTPHandler(
				HandlerWithLogger(slog.New(slog.NewJSONHandler(os.Stdout, nil))),
				HandlerWithClient(server.Client()),
				HandlerWithBaseURL(server.URL+"/"),
			)

			result, err := handler.GetDelegations()

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.expectedCount)
			}
		})
	}
}

func TestHTTPHandler_GetDelegationsFromLevel(t *testing.T) {
	t.Parallel()

	type args struct {
		lastLevel int64
		limit     int
	}

	tests := []struct {
		name           string
		args           args
		mockResponse   string
		mockStatusCode int
		expectedError  bool
		expectedCount  int
		checkURL       func(string) bool
	}{
		{
			name: "Get_Delegations_From_Level_Success",
			args: args{lastLevel: 1000, limit: 50},
			mockResponse: `[
				{
					"type": "delegation",
					"status": "applied",
					"timestamp": "2023-01-01T12:00:00Z",
					"level": 1001,
					"hash": "ophash123",
					"amount": 100000,
					"sender": {"address": "tz1delegator"},
					"newDelegate": {"address": "tz1baker"}
				}
			]`,
			mockStatusCode: http.StatusOK,
			expectedError:  false,
			expectedCount:  1,
			checkURL: func(url string) bool {
				return strings.Contains(url, "level.gt=1000") && strings.Contains(url, "limit=50") && strings.Contains(url, "sort.asc=level")
			},
		},
		{
			name: "Get_Delegations_From_Level_Zero_Success",
			args: args{lastLevel: 0, limit: 100},
			mockResponse: `[
				{
					"type": "delegation",
					"status": "applied",
					"timestamp": "2023-01-01T12:00:00Z",
					"level": 1000,
					"hash": "ophash123",
					"amount": 100000,
					"sender": {"address": "tz1delegator"},
					"newDelegate": {"address": "tz1baker"}
				}
			]`,
			mockStatusCode: http.StatusOK,
			expectedError:  false,
			expectedCount:  1,
			checkURL: func(url string) bool {
				return !strings.Contains(url, "level.gt=") && strings.Contains(url, "limit=100") && strings.Contains(url, "sort.desc=level")
			},
		},
		{
			name:           "Get_Delegations_From_Level_Server_Error",
			args:           args{lastLevel: 1000, limit: 50},
			mockResponse:   "",
			mockStatusCode: http.StatusInternalServerError,
			expectedError:  true,
			expectedCount:  0,
			checkURL:       func(url string) bool { return true },
		},
		{
			name:           "Get_Delegations_From_Level_Invalid_JSON",
			args:           args{lastLevel: 1000, limit: 50},
			mockResponse:   `invalid json`,
			mockStatusCode: http.StatusOK,
			expectedError:  true,
			expectedCount:  0,
			checkURL:       func(url string) bool { return true },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create mock server
			var requestURL string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestURL = r.URL.String()
				w.WriteHeader(tt.mockStatusCode)
				if tt.mockResponse != "" {
					w.Write([]byte(tt.mockResponse))
				}
			}))
			defer server.Close()

			handler := NewHTTPHandler(
				HandlerWithLogger(slog.New(slog.NewJSONHandler(os.Stdout, nil))),
				HandlerWithClient(server.Client()),
				HandlerWithBaseURL(server.URL+"/"),
			)

			result, err := handler.GetDelegationsFromLevel(tt.args.lastLevel, tt.args.limit)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.expectedCount)
			}

			// Check URL construction
			assert.True(t, tt.checkURL(requestURL), "URL check failed for: %s", requestURL)
		})
	}
}

func TestHTTPHandler_GetDelegationsFromLevel_HTTPClientError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "Get_Delegations_HTTP_Client_Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create handler with invalid base URL to trigger HTTP client error
			handler := NewHTTPHandler(
				HandlerWithLogger(slog.New(slog.NewJSONHandler(os.Stdout, nil))),
				HandlerWithClient(&http.Client{}),
				HandlerWithBaseURL("http://invalid-url-that-does-not-exist.local/"),
			)

			result, err := handler.GetDelegationsFromLevel(0, 100)

			assert.Error(t, err)
			assert.Nil(t, result)
		})
	}
}

func TestHTTPHandler_GetDelegationsFromLevel_BodyReadError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "Get_Delegations_Body_Read_Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create mock server that returns a body that errors on read
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Length", "1")
				// Don't write anything, causing a read error
			}))
			defer server.Close()

			handler := NewHTTPHandler(
				HandlerWithLogger(slog.New(slog.NewJSONHandler(os.Stdout, nil))),
				HandlerWithClient(server.Client()),
				HandlerWithBaseURL(server.URL+"/"),
			)

			result, err := handler.GetDelegationsFromLevel(0, 100)

			assert.Error(t, err)
			assert.Nil(t, result)
		})
	}
}

func TestHTTPHandler_BaseURL_Configuration(t *testing.T) {
	t.Parallel()

	type args struct {
		baseURL string
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "Set_Standard_TzKT_URL",
			args: args{baseURL: "https://api.tzkt.io/v1/"},
		},
		{
			name: "Set_Alternative_URL",
			args: args{baseURL: "https://mainnet.api.tez.ie/v1/"},
		},
		{
			name: "Set_Custom_URL",
			args: args{baseURL: "https://custom.tzkt.instance/api/v1/"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := &HTTPHandler{
				baseURL: tt.args.baseURL,
			}

			assert.Equal(t, tt.args.baseURL, handler.baseURL)
		})
	}
}