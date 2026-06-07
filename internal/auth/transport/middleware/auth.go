package middleware

import (
	"slices"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/ports"
	"github.com/yovannylopez/docsy-main/pkg/responses"
)

// AuthMiddleware validates the access JWT and places the user in the Echo context.
type AuthMiddleware struct {
	authService ports.AuthenticationService
}

// NewAuthMiddleware creates a new AuthMiddleware instance
func NewAuthMiddleware(authService ports.AuthenticationService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

// Authenticate validates the access JWT token (not refresh).
func (m *AuthMiddleware) Authenticate() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return responses.Unauthorized(c, "No authorization header")
			}

			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || !strings.EqualFold(tokenParts[0], "Bearer") {
				return responses.Unauthorized(c, "Invalid authorization header")
			}

			user, err := m.authService.ValidateToken(c.Request().Context(), tokenParts[1])
			if err != nil {
				return responses.Unauthorized(c, "Invalid token")
			}

			c.Set("user", user)

			return next(c)
		}
	}
}

// InjectXUserIDHeader fills X-User-ID for handlers that still read the header.
func InjectXUserIDHeader() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if u, ok := c.Get("user").(*entities.User); ok && u != nil {
				c.Request().Header.Set("X-User-ID", u.ID)
			}
			return next(c)
		}
	}
}

// userIsSuperAdmin reports whether the user has the super_admin role.
func userIsSuperAdmin(u *entities.User) bool {
	for i := range u.Roles {
		if u.Roles[i].Name == "super_admin" {
			return true
		}
	}
	return false
}

// userHasPermission checks the permission by name (e.g. correspondence.read).
func userHasPermission(u *entities.User, permission string) bool {
	return slices.Contains(u.PermissionNames, permission)
}

// RequirePermission requires a specific permission after Authenticate.
func RequirePermission(permission string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			u, ok := c.Get("user").(*entities.User)
			if !ok || u == nil {
				return responses.Unauthorized(c, "User not authenticated")
			}
			if userIsSuperAdmin(u) {
				return next(c)
			}
			if !userHasPermission(u, permission) {
				return responses.Forbidden(c, "Permission denied")
			}
			return next(c)
		}
	}
}
