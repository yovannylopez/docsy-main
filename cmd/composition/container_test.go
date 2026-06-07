package composition

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	authContainer "github.com/yovannylopez/docsy-main/internal/auth/infrastructure/container"
	"github.com/yovannylopez/docsy-main/internal/auth/transport/handlers"
	sharedHandlers "github.com/yovannylopez/docsy-main/internal/shared/transport/handlers"
	usersHandlers "github.com/yovannylopez/docsy-main/internal/users/transport/handlers"
)

func TestContainer_Close(t *testing.T) {
	t.Run("close with nil database connection", func(t *testing.T) {
		container := &Container{
			DB: nil,
		}
		err := container.Close()
		assert.NoError(t, err)
	})
}

func TestContainer_CreateAuthHandler(t *testing.T) {
	t.Run("create auth handler", func(t *testing.T) {
		container := &Container{
			AuthContainer: &authContainer.AuthContainer{
				AuthHandler: &handlers.AuthHandler{},
			},
		}

		handler := container.CreateAuthHandler()
		assert.NotNil(t, handler)
		assert.IsType(t, &handlers.AuthHandler{}, handler)
	})
}

func TestContainer_CreateHealthHandler(t *testing.T) {
	t.Run("create health handler", func(t *testing.T) {
		container := &Container{
			HealthHandler: &sharedHandlers.HealthHandler{},
		}

		handler := container.CreateHealthHandler()
		assert.NotNil(t, handler)
		assert.IsType(t, &sharedHandlers.HealthHandler{}, handler)
	})
}

func TestContainer_CreateUserHandler(t *testing.T) {
	t.Run("create user handler", func(t *testing.T) {
		container := &Container{
			UserHandler: &usersHandlers.UsersHandler{},
		}

		handler := container.CreateUserHandler()
		assert.NotNil(t, handler)
		assert.IsType(t, &usersHandlers.UsersHandler{}, handler)
	})
}

func TestContainer_EdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		setupContainer func() *Container
		testFunction   func(*Container) any
		expectedNil    bool
	}{
		{
			name: "create health handler with nil health handler",
			setupContainer: func() *Container {
				return &Container{
					HealthHandler: nil,
				}
			},
			testFunction: func(c *Container) any {
				return c.CreateHealthHandler()
			},
			expectedNil: true,
		},
		{
			name: "create user handler with nil user handler",
			setupContainer: func() *Container {
				return &Container{
					UserHandler: nil,
				}
			},
			testFunction: func(c *Container) any {
				return c.CreateUserHandler()
			},
			expectedNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := tt.setupContainer()
			result := tt.testFunction(container)

			if tt.expectedNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

// MockContainerForTest simulates the container for tests
type MockContainerForTest struct {
	*Container
	MockDB *sqlx.DB
}

// NewMockContainerForTest creates a mock of the container for tests
func NewMockContainerForTest() *MockContainerForTest {
	return &MockContainerForTest{
		Container: &Container{
			AuthContainer: &authContainer.AuthContainer{},
			HealthHandler: &sharedHandlers.HealthHandler{},
			UserHandler:   &usersHandlers.UsersHandler{},
		},
		MockDB: &sqlx.DB{},
	}
}

func BenchmarkContainer_CreateAuthHandler(b *testing.B) {
	container := &Container{
		AuthContainer: &authContainer.AuthContainer{
			AuthHandler: &handlers.AuthHandler{},
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = container.CreateAuthHandler()
	}
}

func BenchmarkContainer_CreateHealthHandler(b *testing.B) {
	container := &Container{
		HealthHandler: &sharedHandlers.HealthHandler{},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = container.CreateHealthHandler()
	}
}

func BenchmarkContainer_CreateUserHandler(b *testing.B) {
	container := &Container{
		UserHandler: &usersHandlers.UsersHandler{},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = container.CreateUserHandler()
	}
}
