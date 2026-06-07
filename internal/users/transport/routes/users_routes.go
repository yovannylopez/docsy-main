package routes

import (
	"github.com/labstack/echo/v4"

	authmw "github.com/yovannylopez/docsy-main/internal/auth/transport/middleware"
	"github.com/yovannylopez/docsy-main/internal/users/transport/handlers"
)

// RegisterUserRoutes registers the routes for the users module under the protected /api/v1 group.
func RegisterUserRoutes(g *echo.Group, userHandler *handlers.UsersHandler) {
	ug := g.Group("/users")

	ug.GET("", userHandler.GetUsers, authmw.RequirePermission("users.read"))
	ug.GET("/search", userHandler.SearchUsers, authmw.RequirePermission("users.read"))
	ug.GET("/profile", userHandler.GetUserProfile, authmw.RequirePermission("users.read"))
	ug.PUT("/profile", userHandler.UpdateUserProfile, authmw.RequirePermission("users.update"))
	ug.POST("/reset-password", userHandler.ResetPassword, authmw.RequirePermission("users.update"))
	ug.POST("", userHandler.CreateUser, authmw.RequirePermission("users.create"))
	ug.GET("/:id", userHandler.GetUserByID, authmw.RequirePermission("users.read"))
	ug.PATCH("/:id", userHandler.PatchUser, authmw.RequirePermission("users.update"))
}
