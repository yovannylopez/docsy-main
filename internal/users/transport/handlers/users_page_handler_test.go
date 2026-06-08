package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	authentities "github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/shared/infrastructure/templates"
	"github.com/yovannylopez/docsy-main/internal/users/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/users/usecases"
)

func newTestUsersPageHandler(mockRepo *MockUserRepository) *UsersPageHandler {
	return NewUsersPageHandler(
		usecases.NewGetUsersUseCase(mockRepo),
		usecases.NewCreateUsersUseCase(mockRepo, &stubPasswordHasher{}, noopAuditRepo{}),
		usecases.NewUpdateUserUseCase(mockRepo),
		usecases.NewSearchUsersUseCase(mockRepo),
		usecases.NewGetUserByIDUseCase(mockRepo),
	)
}

func newEchoWithPageRenderer(t *testing.T) *echo.Echo {
	t.Helper()
	renderer, err := templates.NewRenderer()
	require.NoError(t, err)
	e := echo.New()
	e.Renderer = renderer
	return e
}

func TestUsersPageHandler_ListUsers_ReturnsHTML(t *testing.T) {
	mockRepo := &MockUserRepository{}
	mockRepo.On("GetAllUsers", mock.Anything, 10, 0).Return([]entities.User{}, nil)
	mockRepo.On("GetTotalUsersCount", mock.Anything).Return(0, nil)

	h := newTestUsersPageHandler(mockRepo)
	e := newEchoWithPageRenderer(t)

	req := httptest.NewRequest(http.MethodGet, "/usuarios", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &authentities.User{ID: "admin-1", Email: "admin@test.com", FirstName: "Admin"})

	err := h.ListUsers(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Usuarios del sistema")
	mockRepo.AssertExpectations(t)
}

func TestUsersPageHandler_ShowCreate_ReturnsHTML(t *testing.T) {
	h := newTestUsersPageHandler(&MockUserRepository{})
	e := newEchoWithPageRenderer(t)

	req := httptest.NewRequest(http.MethodGet, "/usuarios/nuevo", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &authentities.User{ID: "admin-1", Email: "admin@test.com"})

	err := h.ShowCreate(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Crear usuario")
}

func TestUsersPageHandler_ShowEdit_UserNotFound(t *testing.T) {
	mockRepo := &MockUserRepository{}
	mockRepo.On("FindByID", mock.Anything, "missing").Return((*entities.User)(nil), nil)

	h := newTestUsersPageHandler(mockRepo)
	e := newEchoWithPageRenderer(t)

	req := httptest.NewRequest(http.MethodGet, "/usuarios/missing/editar", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/usuarios/:id/editar")
	c.SetParamNames("id")
	c.SetParamValues("missing")
	c.Set("user", &authentities.User{ID: "admin-1"})

	err := h.ShowEdit(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestUsersPageHandler_ShowEdit_ReturnsHTML(t *testing.T) {
	mockRepo := &MockUserRepository{}
	user := &entities.User{
		ID:        "user-1",
		Email:     "ana@test.com",
		FirstName: "Ana",
		LastName:  "García",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	mockRepo.On("FindByID", mock.Anything, "user-1").Return(user, nil)

	h := newTestUsersPageHandler(mockRepo)
	e := newEchoWithPageRenderer(t)

	req := httptest.NewRequest(http.MethodGet, "/usuarios/user-1/editar", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/usuarios/:id/editar")
	c.SetParamNames("id")
	c.SetParamValues("user-1")
	c.Set("user", &authentities.User{ID: "admin-1"})

	err := h.ShowEdit(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "ana@test.com")
}
