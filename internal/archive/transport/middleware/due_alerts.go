package middleware

import (
	"github.com/labstack/echo/v4"

	"github.com/yovannylopez/docsy-main/internal/archive/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/archive/domain/ports"
	weblayout "github.com/yovannylopez/docsy-main/internal/shared/transport/web"
)

// InjectDueAlerts loads upcoming+expired due counts into the echo context for AppLayoutData.
func InjectDueAlerts(uc ports.ListDocumentsService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if uc != nil {
				userID := weblayout.CurrentUserID(c)
				if userID != "" {
					upcoming, expired, err := uc.CountDueAlerts(
						c.Request().Context(),
						userID,
						"",
						entities.DocumentStatusActive,
					)
					if err == nil {
						total := upcoming + expired
						if total > 0 {
							c.Set(weblayout.ContextKeyAlertCount, total)
						}
					}
				}
			}
			return next(c)
		}
	}
}
