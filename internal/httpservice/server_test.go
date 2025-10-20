package httpservice

import (
	"context"
	"delegator/conf"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewHTTPServer(t *testing.T) {
	t.Parallel()

	type args struct {
		opts []Options
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "Create_Server_Without_Options",
			args: args{opts: []Options{}},
		},
		{
			name: "Create_Server_With_Logger",
			args: args{opts: []Options{
				WithLogger(slog.New(slog.NewJSONHandler(os.Stdout, nil))),
			}},
		},
		{
			name: "Create_Server_With_Engine",
			args: args{opts: []Options{
				WithEngine(gin.New()),
			}},
		},
		{
			name: "Create_Server_With_All_Options",
			args: args{opts: []Options{
				WithLogger(slog.New(slog.NewJSONHandler(os.Stdout, nil))),
				WithEngine(gin.New()),
				WithHTTPServer(&conf.DelegatorConfig{
					HTTP: struct {
						Port         int `toml:"port"`
						ReadTimeout  int `toml:"read_timeout"`
						WriteTimeout int `toml:"write_timeout"`
					}{
						Port:         8080,
						ReadTimeout:  30,
						WriteTimeout: 30,
					},
				}),
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := NewHTTPServer(tt.args.opts...)
			assert.NotNil(t, server)
		})
	}
}

func TestWithLogger(t *testing.T) {
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

			server := &Server{}
			option := WithLogger(tt.args.logger)
			option(server)

			assert.Equal(t, tt.args.logger, server.logger)
		})
	}
}

func TestWithEngine(t *testing.T) {
	t.Parallel()

	type args struct {
		engine *gin.Engine
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "Set_Engine_Option",
			args: args{engine: gin.New()},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := &Server{}
			option := WithEngine(tt.args.engine)
			option(server)

			assert.Equal(t, tt.args.engine, server.engine)
		})
	}
}

func TestWithHTTPServer(t *testing.T) {
	t.Parallel()

	type args struct {
		config *conf.DelegatorConfig
	}

	tests := []struct {
		name        string
		args        args
		shouldPanic bool
		setupEngine bool
	}{
		{
			name: "Set_HTTPServer_Option_Success",
			args: args{config: &conf.DelegatorConfig{
				HTTP: struct {
					Port         int `toml:"port"`
					ReadTimeout  int `toml:"read_timeout"`
					WriteTimeout int `toml:"write_timeout"`
				}{
					Port:         8080,
					ReadTimeout:  30,
					WriteTimeout: 30,
				},
			}},
			shouldPanic: false,
			setupEngine: true,
		},
		{
			name: "Set_HTTPServer_Option_Panic_No_Engine",
			args: args{config: &conf.DelegatorConfig{
				HTTP: struct {
					Port         int `toml:"port"`
					ReadTimeout  int `toml:"read_timeout"`
					WriteTimeout int `toml:"write_timeout"`
				}{
					Port:         8080,
					ReadTimeout:  30,
					WriteTimeout: 30,
				},
			}},
			shouldPanic: true,
			setupEngine: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := &Server{}
			if tt.setupEngine {
				server.engine = gin.New()
			}

			option := WithHTTPServer(tt.args.config)

			if tt.shouldPanic {
				assert.Panics(t, func() {
					option(server)
				})
			} else {
				assert.NotPanics(t, func() {
					option(server)
				})
				assert.NotNil(t, server.server)
				assert.Equal(t, ":8080", server.server.Addr)
				assert.Equal(t, time.Duration(30)*time.Second, server.server.ReadTimeout)
				assert.Equal(t, time.Duration(30)*time.Second, server.server.WriteTimeout)
			}
		})
	}
}

func TestWithRoutes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "Set_Routes_Option",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			routeCalled := false
			routeRegister := func(engine *gin.Engine) {
				routeCalled = true
				assert.NotNil(t, engine)
			}

			server := &Server{engine: gin.New()}
			option := WithRoutes(routeRegister)
			option(server)

			assert.True(t, routeCalled)
		})
	}
}

func TestServer_Run(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupMock func() *Server
	}{
		{
			name: "Run_Server_Success",
			setupMock: func() *Server {
				engine := gin.New()
				config := &conf.DelegatorConfig{
					HTTP: struct {
						Port         int `toml:"port"`
						ReadTimeout  int `toml:"read_timeout"`
						WriteTimeout int `toml:"write_timeout"`
					}{
						Port:         8080,
						ReadTimeout:  30,
						WriteTimeout: 30,
					},
				}

				return NewHTTPServer(
					WithLogger(slog.New(slog.NewJSONHandler(os.Stdout, nil))),
					WithEngine(engine),
					WithHTTPServer(config),
				)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := tt.setupMock()
			ctx := context.Background()

			err := server.Run(ctx)
			assert.NoError(t, err)

			// Give the server a moment to start
			time.Sleep(10 * time.Millisecond)

			// Test shutdown
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			err = server.Shutdown(shutdownCtx)
			assert.NoError(t, err)
		})
	}
}

func TestServer_Shutdown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupMock func() *Server
	}{
		{
			name: "Shutdown_Server_Success",
			setupMock: func() *Server {
				engine := gin.New()
				config := &conf.DelegatorConfig{
					HTTP: struct {
						Port         int `toml:"port"`
						ReadTimeout  int `toml:"read_timeout"`
						WriteTimeout int `toml:"write_timeout"`
					}{
						Port:         8080,
						ReadTimeout:  30,
						WriteTimeout: 30,
					},
				}

				server := NewHTTPServer(
					WithLogger(slog.New(slog.NewJSONHandler(os.Stdout, nil))),
					WithEngine(engine),
					WithHTTPServer(config),
				)

				// Start the server first
				_ = server.Run(context.Background())
				time.Sleep(10 * time.Millisecond) // Give it time to start

				return server
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := tt.setupMock()
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			err := server.Shutdown(ctx)
			assert.NoError(t, err)
		})
	}
}