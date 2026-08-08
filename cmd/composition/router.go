package composition

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	archivemw "github.com/yovannylopez/docsy-main/internal/archive/transport/middleware"
	archiveRoutes "github.com/yovannylopez/docsy-main/internal/archive/transport/routes"
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
	echo             *echo.Echo
	container        *Container
	authRoutes       *authRoutes.AuthRoutes
	webAuthRoutes    *authRoutes.WebAuthRoutes
	webAuditRoutes   *authRoutes.WebAuditRoutes
	webUsersRoutes   *usersRoutes.WebUsersRoutes
	webArchiveRoutes *archiveRoutes.WebArchiveRoutes
	auditRoutes      *authRoutes.AuditRoutes
	healthRoutes     *sharedRoutes.HealthRoutes
}

// NewRouter creates a new instance of Router with the dependencies of the base modules.
func NewRouter(e *echo.Echo, container *Container) *Router {
	webAuthMW := authmiddleware.NewWebAuthMiddleware(container.AuthContainer.AuthUseCase)
	sidebarStorageMW := archivemw.InjectSidebarStorage(container.ArchiveContainer.GetStorageUsageUC())
	return &Router{
		echo:      e,
		container: container,
		authRoutes: authRoutes.NewAuthRoutes(
			container.CreateAuthHandler(),
			container.AuthContainer.GetMFAHandler(),
		),
		webAuthRoutes: authRoutes.NewWebAuthRoutes(
			container.CreateLoginPageHandler(),
			webAuthMW,
			container.AuthRateLimit,
			sidebarStorageMW,
		),
		webUsersRoutes: usersRoutes.NewWebUsersRoutes(
			container.UsersContainer.GetUsersPageHandler(),
			webAuthMW,
			sidebarStorageMW,
		),
		webAuditRoutes: authRoutes.NewWebAuditRoutes(
			container.CreateAuditPageHandler(),
			webAuthMW,
			sidebarStorageMW,
		),
		webArchiveRoutes: archiveRoutes.NewWebArchiveRoutes(
			container.ArchiveContainer.GetArchivePageHandler(),
			webAuthMW,
			sidebarStorageMW,
		),
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
	r.setupStatic()

	// Server-rendered HTML routes (outside /api)
	r.webAuthRoutes.Setup(r.echo)
	r.webAuthRoutes.SetupProtected(r.echo)
	r.webUsersRoutes.Setup(r.echo)
	r.webAuditRoutes.Setup(r.echo)
	r.webArchiveRoutes.Setup(r.echo)

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
	archiveRoutes.RegisterArchiveRoutes(protected, r.container.ArchiveContainer.GetArchiveHandler())

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
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-User-ID", "HX-Request", "HX-Redirect"},
		AllowCredentials: true,
		MaxAge:           int(corsMaxAge.Seconds()),
	}))
}

func (r *Router) setupStatic() {
	staticPath, err := resolveStaticPath()
	if err != nil {
		panic(fmt.Sprintf("static assets not found: %v", err))
	}
	r.echo.Static("/static", staticPath)
}

func resolveStaticPath() (string, error) {
	if envPath := os.Getenv("STATIC_PATH"); envPath != "" {
		if abs, err := filepath.Abs(envPath); err == nil {
			if info, err2 := os.Stat(abs); err2 == nil && info.IsDir() {
				return abs, nil
			}
		}
	}

	candidates := []string{
		"web/static",
		"../web/static",
		"../../web/static",
	}
	for _, c := range candidates {
		if abs, err := filepath.Abs(c); err == nil {
			if info, err2 := os.Stat(abs); err2 == nil && info.IsDir() {
				return abs, nil
			}
		}
	}

	if wd, err := os.Getwd(); err == nil {
		dir := wd
		for i := 0; i < 5; i++ {
			candidate := filepath.Join(dir, "web", "static")
			if info, err2 := os.Stat(candidate); err2 == nil && info.IsDir() {
				if abs, err3 := filepath.Abs(candidate); err3 == nil {
					return abs, nil
				}
				return candidate, nil
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	return "", fmt.Errorf("no static folder 'web/static' found; configure STATIC_PATH or run from project root")
}
