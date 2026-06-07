package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/auth/mocks"
	authtest "github.com/yovannylopez/docsy-main/internal/auth/test_utils"
	"github.com/yovannylopez/docsy-main/pkg/http_status"
)

func TestAuthMiddleware_Authenticate_MissingHeader(t *testing.T) {
	svc := mocks.NewAuthenticationService(t)
	m := NewAuthMiddleware(svc)

	e := echo.New()
	h := m.Authenticate()(func(c echo.Context) error {
		return c.String(http_status.OK.Code, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h(c)
	require.NoError(t, err)
	assert.Equal(t, http_status.Unauthorized.Code, rec.Code)
}

func TestAuthMiddleware_Authenticate_InvalidBearer(t *testing.T) {
	svc := mocks.NewAuthenticationService(t)
	m := NewAuthMiddleware(svc)

	e := echo.New()
	h := m.Authenticate()(func(c echo.Context) error {
		return c.String(http_status.OK.Code, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Basic xxx")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h(c)
	require.NoError(t, err)
	assert.Equal(t, http_status.Unauthorized.Code, rec.Code)
}

func TestAuthMiddleware_Authenticate_InvalidToken(t *testing.T) {
	svc := mocks.NewAuthenticationService(t)
	svc.On("ValidateToken", mock.Anything, "bad").Return((*entities.User)(nil), errors.New("invalid"))
	m := NewAuthMiddleware(svc)

	e := echo.New()
	h := m.Authenticate()(func(c echo.Context) error {
		return c.String(http_status.OK.Code, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer bad")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h(c)
	require.NoError(t, err)
	assert.Equal(t, http_status.Unauthorized.Code, rec.Code)
}

func TestAuthMiddleware_Authenticate_Success(t *testing.T) {
	svc := mocks.NewAuthenticationService(t)
	stubs := authtest.NewAuthStubs()
	u := authtest.CloneUser(stubs.Entities.ValidUser)
	u.ID = "u1"
	u.Email = "a@b.com"
	svc.On("ValidateToken", mock.Anything, "good").Return(u, nil)
	m := NewAuthMiddleware(svc)

	e := echo.New()
	var seen *entities.User
	h := m.Authenticate()(func(c echo.Context) error {
		seen, _ = c.Get("user").(*entities.User)
		return c.String(http_status.OK.Code, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer good")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h(c)
	require.NoError(t, err)
	assert.Equal(t, http_status.OK.Code, rec.Code)
	require.NotNil(t, seen)
	assert.Equal(t, u.ID, seen.ID)
}

func TestInjectXUserIDHeader_SetsHeader(t *testing.T) {
	e := echo.New()
	stubs := authtest.NewAuthStubs()
	u := authtest.CloneUser(stubs.Entities.EmptyUser)
	u.ID = "user-42"
	h := InjectXUserIDHeader()(func(c echo.Context) error {
		assert.Equal(t, "user-42", c.Request().Header.Get("X-User-ID"))
		return c.NoContent(http_status.NoContent.Code)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", u)
	err := h(c)
	require.NoError(t, err)
}

func TestRequirePermission_AllowsSuperAdmin(t *testing.T) {
	e := echo.New()
	stubs := authtest.NewAuthStubs()
	u := authtest.CloneUser(stubs.Entities.EmptyUser)
	u.Roles = []entities.Role{{Name: "super_admin"}}
	h := RequirePermission("any.perm")(func(c echo.Context) error {
		return c.NoContent(http_status.NoContent.Code)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", u)
	err := h(c)
	require.NoError(t, err)
	assert.Equal(t, http_status.NoContent.Code, rec.Code)
}

func TestRequirePermission_DeniesWithoutPerm(t *testing.T) {
	e := echo.New()
	stubs := authtest.NewAuthStubs()
	u := authtest.CloneUser(stubs.Entities.EmptyUser)
	u.PermissionNames = []string{"other.perm"}
	h := RequirePermission("users.read")(func(c echo.Context) error {
		return c.NoContent(http_status.OK.Code)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", u)
	err := h(c)
	require.NoError(t, err)
	assert.Equal(t, http_status.Forbidden.Code, rec.Code)
}
