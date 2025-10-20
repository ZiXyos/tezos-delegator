package routes

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewRouteRegistry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "Create_Route_Registry",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			registry := NewRouteRegistry()
			assert.NotNil(t, registry)
			assert.NotNil(t, registry.registrars)
			assert.Empty(t, registry.registrars)
		})
	}
}

func TestRouteRegistry_AddRegistrar(t *testing.T) {
	t.Parallel()

	type args struct {
		registrarCount int
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "Add_Single_Registrar",
			args: args{registrarCount: 1},
		},
		{
			name: "Add_Multiple_Registrars",
			args: args{registrarCount: 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			registry := NewRouteRegistry()

			// Add the specified number of registrars
			for i := 0; i < tt.args.registrarCount; i++ {
				registrar := func(engine *gin.Engine) {
					// Mock registrar function
				}
				registry.AddRegistrar(registrar)
			}

			assert.Len(t, registry.registrars, tt.args.registrarCount)
		})
	}
}

func TestRouteRegistry_RegisterAll(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		registrarCount int
	}{
		{
			name:           "Register_All_Single_Registrar",
			registrarCount: 1,
		},
		{
			name:           "Register_All_Multiple_Registrars",
			registrarCount: 3,
		},
		{
			name:           "Register_All_No_Registrars",
			registrarCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			registry := NewRouteRegistry()
			engine := gin.New()
			callCount := 0

			// Add the specified number of registrars
			for i := 0; i < tt.registrarCount; i++ {
				registrar := func(engine *gin.Engine) {
					callCount++
					assert.NotNil(t, engine)
				}
				registry.AddRegistrar(registrar)
			}

			registry.RegisterAll(engine)

			assert.Equal(t, tt.registrarCount, callCount)
		})
	}
}

func TestCreateRouteRegistrar(t *testing.T) {
	t.Parallel()

	type args struct {
		registrarCount int
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "Create_Registrar_Single",
			args: args{registrarCount: 1},
		},
		{
			name: "Create_Registrar_Multiple",
			args: args{registrarCount: 3},
		},
		{
			name: "Create_Registrar_None",
			args: args{registrarCount: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			engine := gin.New()
			callCount := 0

			// Create mock registrars
			var registrars []RouteRegistrar
			for i := 0; i < tt.args.registrarCount; i++ {
				registrar := func(engine *gin.Engine) {
					callCount++
					assert.NotNil(t, engine)
				}
				registrars = append(registrars, registrar)
			}

			// Create the combined registrar
			combinedRegistrar := CreateRouteRegistrar(registrars...)
			assert.NotNil(t, combinedRegistrar)

			// Execute the combined registrar
			combinedRegistrar(engine)

			assert.Equal(t, tt.args.registrarCount, callCount)
		})
	}
}

func TestRouteRegistrar_Type(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "RouteRegistrar_Function_Type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var registrar RouteRegistrar = func(engine *gin.Engine) {
				// Mock function
			}

			assert.NotNil(t, registrar)

			// Test that it can be called
			engine := gin.New()
			assert.NotPanics(t, func() {
				registrar(engine)
			})
		})
	}
}