package middleware

import (
	"github.com/labstack/echo/v4"

	"github.com/yovannylopez/docsy-main/internal/archive/domain/ports"
	weblayout "github.com/yovannylopez/docsy-main/internal/shared/transport/web"
)

// InjectSidebarStorage loads storage usage into the echo context for AppLayoutData.
func InjectSidebarStorage(uc ports.GetStorageUsageService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if uc != nil {
				userID := weblayout.CurrentUserID(c)
				if userID != "" {
					usage, err := uc.Execute(c.Request().Context(), userID)
					if err == nil && usage != nil {
						c.Set(weblayout.ContextKeySidebarStorage, weblayout.NewSidebarStorageData(usage.UsedBytes, usage.QuotaBytes, usage.Percent))
					}
				}
			}
			return next(c)
		}
	}
}
