package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yovannylopez/docsy-main/internal/auth/mocks"
	"github.com/yovannylopez/docsy-main/pkg/http_status"
)

func TestAuthHandler_Logout_PassesClientContext(t *testing.T) {
	loginSvc := mocks.NewLoginService(t)
	authSvc := mocks.NewAuthenticationService(t)
	h := NewAuthHandler(loginSvc, authSvc, mocks.NewChangePasswordService(t))

	authSvc.On(
		"Logout",
		mock.Anything,
		"access-token-value",
		"TestUA/1.0",
		"10.0.0.5",
	).Return(nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer access-token-value")
	req.Header.Set("User-Agent", "TestUA/1.0")
	req.Header.Set("X-Real-IP", "10.0.0.5")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Logout(c)
	require.NoError(t, err)
	assert.Equal(t, http_status.OK.Code, rec.Code)
	authSvc.AssertExpectations(t)
}
