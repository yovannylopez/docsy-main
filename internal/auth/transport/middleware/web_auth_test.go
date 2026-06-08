package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/shared/infrastructure/templates"
)

func TestWebAuthMiddleware_RequirePermission_DeniesWithoutPermission(t *testing.T) {
	renderer, err := templates.NewRenderer()
	require.NoError(t, err)

	e := echo.New()
	e.Renderer = renderer
	mw := NewWebAuthMiddleware(nil)

	h := mw.RequirePermission("users.read")(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/usuarios", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &entities.User{
		ID:              "u1",
		Email:           "user@test.com",
		PermissionNames: []string{"other.perm"},
	})

	err = h(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), "No tienes permiso")
}

func TestWebAuthMiddleware_RequirePermission_AllowsWithPermission(t *testing.T) {
	e := echo.New()
	mw := NewWebAuthMiddleware(nil)

	h := mw.RequirePermission("users.read")(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/usuarios", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &entities.User{
		ID:              "u1",
		PermissionNames: []string{"users.read"},
	})

	err := h(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}
