package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/ports"
	weblayout "github.com/yovannylopez/docsy-main/internal/shared/transport/web"
)

const accessTokenCookie = "access_token"

// WebAuthMiddleware protects server-rendered HTML routes.
type WebAuthMiddleware struct {
	authService ports.AuthenticationService
}

// NewWebAuthMiddleware creates a WebAuthMiddleware instance.
func NewWebAuthMiddleware(authService ports.AuthenticationService) *WebAuthMiddleware {
	return &WebAuthMiddleware{authService: authService}
}

// RequireAuth validates the access token from cookie or Authorization header.
// Unauthenticated requests are redirected to /login.
func (m *WebAuthMiddleware) RequireAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := extractAccessToken(c)
			if token == "" {
				return c.Redirect(http.StatusFound, "/login")
			}

			user, err := m.authService.ValidateToken(c.Request().Context(), token)
			if err != nil {
				return c.Redirect(http.StatusFound, "/login")
			}

			c.Set("user", user)

			return next(c)
		}
	}
}

func extractAccessToken(c echo.Context) string {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2) //nolint:mnd
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return strings.TrimSpace(parts[1])
		}
	}

	if cookie, err := c.Cookie(accessTokenCookie); err == nil {
		return strings.TrimSpace(cookie.Value)
	}

	return ""
}

// RequirePermission requires a named RBAC permission for server-rendered routes.
// Must run after RequireAuth so the user is present in context.
func (m *WebAuthMiddleware) RequirePermission(permission string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			u, ok := c.Get("user").(*entities.User)
			if !ok || u == nil {
				return c.Redirect(http.StatusFound, "/login")
			}
			if !CheckUserPermission(u, permission) {
				data := weblayout.AppLayoutFromEcho(c, "Acceso denegado", "No tienes permiso para ver esta sección", c.Path())
				return c.Render(http.StatusForbidden, "forbidden", data)
			}
			return next(c)
		}
	}
}
