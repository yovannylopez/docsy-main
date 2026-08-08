package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/yovannylopez/docsy-main/internal/auth/domain"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/ports"
	weblayout "github.com/yovannylopez/docsy-main/internal/shared/transport/web"
	"github.com/yovannylopez/docsy-main/pkg/constants"
)

const (
	accessTokenCookieName = "access_token"
	secondsPerHour        = 60 * 60
)

const (
	loginPageTitle    = "Hola de nuevo"
	loginPageSubtitle = "Ingresa tus credenciales para acceder a tu cuenta."

	msgLoginFailed     = "Credenciales inválidas. Verifica tu correo y contraseña."
	msgAccountLocked   = "La cuenta está bloqueada temporalmente. Intenta más tarde."
	msgAccountInactive = "La cuenta no está activa. Contacta al administrador."
	msgFieldsRequired  = "El correo electrónico y la contraseña son obligatorios."
	msgInternalError   = "Ocurrió un error interno. Intenta de nuevo."
)

// LoginPageData holds view data for the login page and HTMX partials.
type LoginPageData struct {
	Title      string
	Subtitle   string
	Email      string
	Error      string
	ThemeClass string
}

// LoginResultData holds token data rendered after a successful login.
type LoginResultData struct {
	AccessToken  string
	RefreshToken string
}

type loginFormRequest struct {
	Email    string `form:"email"`
	Password string `form:"password"`
}

// LoginPageHandler serves server-rendered auth pages (HTMX).
type LoginPageHandler struct {
	loginService ports.LoginService
	authService  ports.AuthenticationService
}

// NewLoginPageHandler creates a LoginPageHandler.
func NewLoginPageHandler(loginService ports.LoginService, authService ports.AuthenticationService) *LoginPageHandler {
	return &LoginPageHandler{
		loginService: loginService,
		authService:  authService,
	}
}

// ShowLogin renders the full login page.
func (h *LoginPageHandler) ShowLogin(c echo.Context) error {
	data := LoginPageData{
		Title:    loginPageTitle,
		Subtitle: loginPageSubtitle,
	}
	return c.Render(http.StatusOK, "login", data)
}

// SubmitLogin handles HTMX form login (application/x-www-form-urlencoded).
func (h *LoginPageHandler) SubmitLogin(c echo.Context) error {
	var form loginFormRequest
	if err := c.Bind(&form); err != nil {
		return h.renderLoginAlert(c, http.StatusUnprocessableEntity, msgFieldsRequired, form.Email)
	}

	form.Email = strings.TrimSpace(form.Email)
	if form.Email == "" || form.Password == "" {
		return h.renderLoginAlert(c, http.StatusUnprocessableEntity, msgFieldsRequired, form.Email)
	}

	request := &dtos.LoginRequest{
		Email:    form.Email,
		Password: form.Password,
	}

	responseData, err := h.loginService.Login(
		c.Request().Context(),
		request,
		c.Request().UserAgent(),
		h.getClientIP(c),
	)
	if err != nil {
		return h.handleLoginError(c, err, form.Email)
	}

	if responseData != nil && responseData.MFARequired {
		return c.Render(http.StatusOK, "mfa_unavailable", nil)
	}

	if responseData == nil || responseData.Token == nil {
		return h.renderLoginAlert(c, http.StatusInternalServerError, msgInternalError, form.Email)
	}

	setAccessTokenCookie(c, responseData.Token.AccessToken)
	c.Response().Header().Set("HX-Redirect", "/")

	return c.Render(http.StatusOK, "login_result", LoginResultData{
		AccessToken:  responseData.Token.AccessToken,
		RefreshToken: responseData.Token.RefreshToken,
	})
}

// ShowHome renders the protected home placeholder.
func (h *LoginPageHandler) ShowHome(c echo.Context) error {
	return c.Render(http.StatusOK, "home", buildAppLayoutData(c, "Inicio", "Panel principal", "/"))
}

// ShowProfile renders the profile placeholder page.
func (h *LoginPageHandler) ShowProfile(c echo.Context) error {
	return c.Render(http.StatusOK, "placeholder", buildAppLayoutData(c, "Perfil", "Mi perfil", "/perfil"))
}

// ShowSettings renders the settings placeholder page.
func (h *LoginPageHandler) ShowSettings(c echo.Context) error {
	return c.Render(http.StatusOK, "placeholder", buildAppLayoutData(c, "Configuración", "Preferencias de la aplicación", "/configuracion"))
}

func buildAppLayoutData(c echo.Context, title, subtitle, activeRoute string) weblayout.AppLayoutData {
	return weblayout.AppLayoutFromEcho(c, title, subtitle, activeRoute)
}

// SubmitLogout closes the session and redirects to login.
func (h *LoginPageHandler) SubmitLogout(c echo.Context) error {
	token := bearerToken(c)
	if token == "" {
		if cookie, err := c.Cookie("access_token"); err == nil {
			token = cookie.Value
		}
	}

	if token != "" {
		_ = h.authService.Logout(
			c.Request().Context(),
			token,
			c.Request().UserAgent(),
			h.getClientIP(c),
		)
	}

	clearAccessTokenCookie(c)

	if isHTMXRequest(c) {
		c.Response().Header().Set("HX-Redirect", "/login")
		return c.NoContent(http.StatusOK)
	}

	return c.Redirect(http.StatusFound, "/login")
}

func (h *LoginPageHandler) handleLoginError(c echo.Context, err error, email string) error {
	switch {
	case errors.Is(err, domain.ErrInvalidCredentials):
		return h.renderLoginAlert(c, http.StatusUnprocessableEntity, msgLoginFailed, email)
	case err.Error() == "account is locked":
		return h.renderLoginAlert(c, http.StatusUnprocessableEntity, msgAccountLocked, email)
	case err.Error() == "account is not active":
		return h.renderLoginAlert(c, http.StatusUnprocessableEntity, msgAccountInactive, email)
	default:
		return h.renderLoginAlert(c, http.StatusInternalServerError, msgInternalError, email)
	}
}

func (h *LoginPageHandler) renderLoginAlert(c echo.Context, status int, message, email string) error {
	return c.Render(status, "partials/alerts", LoginPageData{
		Error: message,
		Email: email,
	})
}

func (h *LoginPageHandler) getClientIP(c echo.Context) string {
	realIP := c.Request().Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	forwardedFor := c.Request().Header.Get("X-Forwarded-For")
	if forwardedFor != "" {
		ips := strings.Split(forwardedFor, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	return c.RealIP()
}

func isHTMXRequest(c echo.Context) bool {
	return c.Request().Header.Get("HX-Request") == "true"
}

func cookieSecure(c echo.Context) bool {
	if c.Scheme() == "https" {
		return true
	}
	return strings.EqualFold(c.Request().Header.Get("X-Forwarded-Proto"), "https")
}

func setAccessTokenCookie(c echo.Context, token string) {
	maxAge := constants.AccessTokenExpirationHours * secondsPerHour
	c.SetCookie(&http.Cookie{ //nolint:gosec // Secure follows request scheme; HttpOnly and SameSiteLax required for local HTTP dev
		Name:     accessTokenCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   cookieSecure(c),
		SameSite: http.SameSiteLaxMode,
	})
}

func clearAccessTokenCookie(c echo.Context) {
	c.SetCookie(&http.Cookie{ //nolint:gosec // Secure follows request scheme; HttpOnly and SameSiteLax required for local HTTP dev
		Name:     accessTokenCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   cookieSecure(c),
		SameSite: http.SameSiteLaxMode,
	})
}

func displayName(user *entities.User) string {
	if user == nil {
		return ""
	}
	if strings.TrimSpace(user.FirstName) != "" {
		name := strings.TrimSpace(user.FirstName)
		if strings.TrimSpace(user.LastName) != "" {
			name += " " + strings.TrimSpace(user.LastName)
		}
		return name
	}
	return user.Email
}
