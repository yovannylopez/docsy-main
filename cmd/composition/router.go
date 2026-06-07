package composition

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	authmiddleware "github.com/yovannylopez/docsy-main/internal/auth/transport/middleware"
	authRoutes "github.com/yovannylopez/docsy-main/internal/auth/transport/routes"
	coreMiddleware "github.com/yovannylopez/docsy-main/internal/shared/transport/middleware"
	sharedRoutes "github.com/yovannylopez/docsy-main/internal/shared/transport/routes"
	usersRoutes "github.com/yovannylopez/docsy-main/internal/users/transport/routes"
)

const (
	corsMaxAge = 24 * time.Hour
)

// Router contains the routes of the application.
// To add a new module, register its routes in SetupRoutes.
type Router struct {
	echo         *echo.Echo
	container    *Container
	authRoutes   *authRoutes.AuthRoutes
	auditRoutes  *authRoutes.AuditRoutes
	healthRoutes *sharedRoutes.HealthRoutes
}

// NewRouter creates a new instance of Router with the dependencies of the base modules.
func NewRouter(e *echo.Echo, container *Container) *Router {
	return &Router{
		echo:         e,
		container:    container,
		authRoutes:   authRoutes.NewAuthRoutes(container.CreateAuthHandler(), container.AuthContainer.GetMFAHandler()),
		auditRoutes:  authRoutes.NewAuditRoutes(container.AuthContainer.GetAuditHandler()),
		healthRoutes: sharedRoutes.NewHealthRoutes(container.CreateHealthHandler()),
	}
}

// SetupRoutes configures global middleware and registers all the routes of the application.
func (r *Router) SetupRoutes() {
	r.echo.HTTPErrorHandler = coreMiddleware.CentralHTTPErrorHandler
	r.echo.Use(coreMiddleware.RequestIDMiddleware())
	r.echo.Use(middleware.Logger())
	r.echo.Use(middleware.Recover())
	r.setupCORS()

	api := r.echo.Group("/api")

	// Public routes (without authentication)
	public := api.Group("/public")
	r.healthRoutes.Setup(public)

	v1 := api.Group("/v1")

	// Authentication with rate limiting
	v1Auth := v1.Group("/auth")
	v1Auth.Use(r.container.AuthRateLimit)
	r.authRoutes.Setup(v1Auth)

	// Protected routes by JWT
	protected := v1.Group("")
	protected.Use(r.container.AuthContainer.AuthHTTPMiddleware.Authenticate())
	protected.Use(authmiddleware.InjectXUserIDHeader())

	// Auth routes that require an authenticated user (e.g. change-password)
	protectedAuth := protected.Group("/auth")
	r.authRoutes.SetupProtected(protectedAuth)

	usersRoutes.RegisterUserRoutes(protected, r.container.UsersContainer.GetUsersHandler())
	r.auditRoutes.Setup(protected)

	// ─── Register here the routes of your business modules ──────────────────
	// Example when adding the "products" module with scaffold_module.sh:
	//   productsRoutes.NewProductsRoutes(r.echo, r.container.Products).Register(
	//       r.container.AuthContainer.AuthHTTPMiddleware.Authenticate(),
	//       authmiddleware.InjectXUserIDHeader(),
	//   )
}

// setupCORS configures CORS reading allowed origins from the ALLOWED_ORIGINS environment variable
// ALLOWED_ORIGINS (comma-separated). By default it allows http://localhost:3000.
func (r *Router) setupCORS() {
	rawOrigins := os.Getenv("ALLOWED_ORIGINS")
	if rawOrigins == "" {
		rawOrigins = "http://localhost:3000"
	}

	allowedOrigins := strings.Split(rawOrigins, ",")

	r.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-User-ID"},
		AllowCredentials: true,
		MaxAge:           int(corsMaxAge.Seconds()),
	}))
}
