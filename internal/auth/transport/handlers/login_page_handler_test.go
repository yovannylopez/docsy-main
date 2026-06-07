package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yovannylopez/docsy-main/internal/auth/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/auth/mocks"
	authtest "github.com/yovannylopez/docsy-main/internal/auth/test_utils"
	"github.com/yovannylopez/docsy-main/internal/shared/infrastructure/templates"
)

func newTestLoginPageHandler(t *testing.T) (*LoginPageHandler, *mocks.LoginService) {
	t.Helper()
	loginSvc := mocks.NewLoginService(t)
	authSvc := mocks.NewAuthenticationService(t)
	return NewLoginPageHandler(loginSvc, authSvc), loginSvc
}

func newEchoWithRenderer(t *testing.T) *echo.Echo {
	t.Helper()
	renderer, err := templates.NewRenderer()
	require.NoError(t, err)

	e := echo.New()
	e.Renderer = renderer
	return e
}

func TestLoginPageHandler_ShowLogin_ReturnsHTML(t *testing.T) {
	h, _ := newTestLoginPageHandler(t)
	e := newEchoWithRenderer(t)

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ShowLogin(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get(echo.HeaderContentType), "text/html")
	assert.Contains(t, rec.Body.String(), "Hola de nuevo")
}

func TestLoginPageHandler_SubmitLogin_InvalidCredentials(t *testing.T) {
	h, loginSvc := newTestLoginPageHandler(t)
	loginSvc.On("Login", mock.Anything, mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
		Return((*dtos.LoginResponse)(nil), &domainError{msg: "invalid credentials"})

	e := newEchoWithRenderer(t)
	form := url.Values{}
	form.Set("email", "a@b.com")
	form.Set("password", "wrong")

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SubmitLogin(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	assert.Contains(t, rec.Body.String(), "Credenciales inválidas")
}

type domainError struct{ msg string }

func (e *domainError) Error() string { return e.msg }

func TestLoginPageHandler_SubmitLogin_Success(t *testing.T) {
	h, loginSvc := newTestLoginPageHandler(t)

	stubs := authtest.NewAuthStubs()
	loginDTO := stubs.CreateMockLoginRequest("a@b.com", "secret")
	resp := authtest.MinimalLoginResponse("u1", loginDTO.Email, "access-token-123")
	resp.Token.RefreshToken = "refresh-token-456"

	loginSvc.On("Login", mock.Anything, mock.MatchedBy(func(r *dtos.LoginRequest) bool {
		return r.Email == loginDTO.Email && r.Password == loginDTO.Password
	}), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(resp, nil)

	e := newEchoWithRenderer(t)
	form := url.Values{}
	form.Set("email", loginDTO.Email)
	form.Set("password", loginDTO.Password)

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SubmitLogin(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "/", rec.Header().Get("HX-Redirect"))
	assert.Contains(t, rec.Body.String(), "access-token-123")
	assert.Contains(t, rec.Body.String(), "sessionStorage.setItem('access_token'")

	cookies := rec.Result().Cookies()
	var accessCookie *http.Cookie
	for _, ck := range cookies {
		if ck.Name == "access_token" {
			accessCookie = ck
			break
		}
	}
	require.NotNil(t, accessCookie)
	assert.Equal(t, "access-token-123", accessCookie.Value)
	assert.True(t, accessCookie.HttpOnly)
	loginSvc.AssertExpectations(t)
}

func TestLoginPageHandler_SubmitLogin_EmptyFields(t *testing.T) {
	h, _ := newTestLoginPageHandler(t)
	e := newEchoWithRenderer(t)

	form := url.Values{}
	form.Set("email", "")
	form.Set("password", "")

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SubmitLogin(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	assert.Contains(t, rec.Body.String(), "obligatorios")
}

func TestLoginPageHandler_ShowHome_IncludesUserNameAndProfileMenu(t *testing.T) {
	h, _ := newTestLoginPageHandler(t)
	e := newEchoWithRenderer(t)

	stubs := authtest.NewAuthStubs()
	user := authtest.CloneUser(stubs.Entities.ValidUser)
	user.FirstName = "Ana"
	user.LastName = "García"

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", user)

	err := h.ShowHome(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Ana García")
	assert.Contains(t, rec.Body.String(), "Variantes de color")
	assert.Contains(t, rec.Body.String(), "Cerrar sesión")
}
