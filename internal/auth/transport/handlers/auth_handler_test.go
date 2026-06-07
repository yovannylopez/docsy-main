package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yovannylopez/docsy-main/internal/auth/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/auth/mocks"
	authtest "github.com/yovannylopez/docsy-main/internal/auth/test_utils"
	"github.com/yovannylopez/docsy-main/pkg/http_status"
)

func newTestAuthHandler(t *testing.T) (*AuthHandler, *mocks.LoginService) {
	t.Helper()
	loginSvc := mocks.NewLoginService(t)
	authSvc := mocks.NewAuthenticationService(t)
	return NewAuthHandler(loginSvc, authSvc, mocks.NewChangePasswordService(t)), loginSvc
}

func TestAuthHandler_Login_BindError(t *testing.T) {
	h, _ := newTestAuthHandler(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString("{"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Login(c)
	require.NoError(t, err)
	assert.Equal(t, http_status.BadRequest.Code, rec.Code)
}

func TestAuthHandler_Login_MissingFields(t *testing.T) {
	h, _ := newTestAuthHandler(t)

	body, _ := json.Marshal(map[string]string{"email": "", "password": ""})
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Login(c)
	require.NoError(t, err)
	assert.Equal(t, http_status.BadRequest.Code, rec.Code)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	h, loginSvc := newTestAuthHandler(t)

	stubs := authtest.NewAuthStubs()
	loginDTO := stubs.CreateMockLoginRequest("a@b.com", "secret")
	resp := authtest.MinimalLoginResponse("u1", loginDTO.Email, "at")
	loginSvc.On("Login", mock.Anything, mock.MatchedBy(func(r *dtos.LoginRequest) bool {
		return r.Email == loginDTO.Email && r.Password == loginDTO.Password
	}), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(resp, nil)

	body, _ := json.Marshal(loginDTO)
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("X-Real-IP", "10.0.0.1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Login(c)
	require.NoError(t, err)
	assert.Equal(t, http_status.OK.Code, rec.Code)
	loginSvc.AssertExpectations(t)
}

func TestAuthHandler_Login_Errors(t *testing.T) {
	cases := []struct {
		name       string
		loginErr   error
		wantStatus int
	}{
		{"invalid credentials", errors.New("invalid credentials"), http_status.BadRequest.Code},
		{"account locked", errors.New("account is locked"), http_status.BadRequest.Code},
		{"inactive", errors.New("account is not active"), http_status.BadRequest.Code},
		{"other", errors.New("db down"), http_status.InternalError.Code},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h, loginSvc := newTestAuthHandler(t)

			loginSvc.On("Login", mock.Anything, mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
				Return((*dtos.LoginResponse)(nil), tc.loginErr)

			stubs := authtest.NewAuthStubs()
			body, _ := json.Marshal(stubs.CreateMockLoginRequest("a@b.com", "x"))
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := h.Login(c)
			require.NoError(t, err)
			assert.Equal(t, tc.wantStatus, rec.Code)
		})
	}
}
