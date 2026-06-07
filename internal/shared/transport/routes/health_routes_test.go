package routes

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yovannylopez/docsy-main/internal/shared/transport/handlers"
	"github.com/yovannylopez/docsy-main/pkg/http_status"
	"github.com/yovannylopez/docsy-main/pkg/logging"
)

func TestMain(m *testing.M) {
	_ = logging.Init(false)
	os.Exit(m.Run())
}

func openHealthRoutesTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	db, err := sqlx.Connect("sqlite3", ":memory:")
	require.NoError(t, err)
	return db
}

// TestHealthRoutes_Setup_RegistersHealthAndReady ejercita NewHealthRoutes y Setup de
// health_routes.go contra un HealthHandler real (misma estrategia que handlers_test).
func TestHealthRoutes_Setup_RegistersHealthAndReady(t *testing.T) {
	db := openHealthRoutesTestDB(t)
	t.Cleanup(func() { _ = db.Close() })

	hr := NewHealthRoutes(handlers.NewHealthHandler(db, nil))

	e := echo.New()
	public := e.Group("/api/public")
	hr.Setup(public)

	t.Run("GET health", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/public/health", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http_status.OK.Code, rec.Code)
	})

	t.Run("GET ready", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/public/ready", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http_status.OK.Code, rec.Code)
	})
}
