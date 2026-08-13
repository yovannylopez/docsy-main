package web

import (
	"strings"

	"github.com/labstack/echo/v4"

	authentities "github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
)

// UserFromEchoContext returns the authenticated user set by WebAuthMiddleware.
func UserFromEchoContext(c echo.Context) *authentities.User {
	u, ok := c.Get("user").(*authentities.User)
	if !ok {
		return nil
	}
	return u
}

// CurrentUserID returns the authenticated user ID or an empty string.
func CurrentUserID(c echo.Context) string {
	if u := UserFromEchoContext(c); u != nil {
		return u.ID
	}
	return ""
}

// CurrentUserIsSuperAdmin reports whether the authenticated user has the super_admin role.
func CurrentUserIsSuperAdmin(c echo.Context) bool {
	u := UserFromEchoContext(c)
	if u == nil {
		return false
	}
	for i := range u.Roles {
		if u.Roles[i].Name == "super_admin" {
			return true
		}
	}
	return false
}

// DisplayUserName builds a display name from the user entity.
func DisplayUserName(user *authentities.User) string {
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

// AppLayoutFromEcho builds layout data using the authenticated user from context.
func AppLayoutFromEcho(c echo.Context, title, subtitle, activeRoute string) AppLayoutData {
	data := NewAppLayoutData(title, subtitle, DisplayUserName(UserFromEchoContext(c)), activeRoute)
	if v, ok := c.Get(ContextKeySidebarStorage).(SidebarStorageData); ok {
		data.Storage = v
	}
	if n, ok := c.Get(ContextKeyAlertCount).(int); ok && n > 0 {
		data.AlertCount = n
	}
	return data
}

// IsHTMXRequest reports whether the request was made by HTMX.
func IsHTMXRequest(c echo.Context) bool {
	return c.Request().Header.Get("HX-Request") == "true"
}

// DerefString returns the string value or empty when the pointer is nil.
func DerefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
