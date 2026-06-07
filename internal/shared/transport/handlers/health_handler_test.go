package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yovannylopez/docsy-main/pkg/http_status"
	"github.com/yovannylopez/docsy-main/pkg/logging"
)

func newEchoContext(path string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func openTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	db, err := sqlx.Connect("sqlite3", ":memory:")
	require.NoError(t, err)
	return db
}

func parseJSON(t *testing.T, body string) map[string]any {
	t.Helper()
	var data map[string]any
	require.NoError(t, json.Unmarshal([]byte(body), &data))
	return data
}

func TestMain(m *testing.M) {
	// Initialize logger to avoid nil pointer in logging during tests
	_ = logging.Init(false)
	os.Exit(m.Run())
}

func TestHealthHandler_HealthCheck_Healthy(t *testing.T) {
	db := openTestDB(t)
	t.Cleanup(func() { _ = db.Close() })

	h := NewHealthHandler(db, nil)
	c, rec := newEchoContext("/health")

	err := h.HealthCheck(c)
	assert.NoError(t, err)
	assert.Equal(t, http_status.OK.Code, rec.Code)

	payload := parseJSON(t, rec.Body.String())
	assert.Equal(t, "ok", payload["status"])
	services := payload["services"].(map[string]any)
	dbInfo := services["database"].(map[string]any)
	assert.Equal(t, "healthy", dbInfo["status"])
	assert.Equal(t, "", dbInfo["error"])
	stats := dbInfo["stats"].(map[string]any)
	// Minimal sanity checks on stats keys
	assert.Contains(t, stats, "max_open_connections")
	assert.Contains(t, stats, "open_connections")
}

func TestHealthHandler_HealthCheck_Unhealthy(t *testing.T) {
	db := openTestDB(t)
	require.NoError(t, db.Close()) // closed DB to force ping error

	h := NewHealthHandler(db, nil)
	c, rec := newEchoContext("/health")

	start := time.Now()
	err := h.HealthCheck(c)
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.Equal(t, http_status.ServiceUnavailable.Code, rec.Code)
	// Should fail fast, not sleep for timeout
	assert.Less(t, duration, 2*time.Second)

	payload := parseJSON(t, rec.Body.String())
	services := payload["services"].(map[string]any)
	dbInfo := services["database"].(map[string]any)
	assert.Equal(t, "unhealthy", dbInfo["status"])
	assert.NotEmpty(t, dbInfo["error"])
}

func TestHealthHandler_ReadyCheck_Ready(t *testing.T) {
	db := openTestDB(t)
	t.Cleanup(func() { _ = db.Close() })

	h := NewHealthHandler(db, nil)
	c, rec := newEchoContext("/ready")

	err := h.ReadyCheck(c)
	assert.NoError(t, err)
	assert.Equal(t, http_status.OK.Code, rec.Code)

	payload := parseJSON(t, rec.Body.String())
	assert.Equal(t, "ready", payload["status"])
}

func TestHealthHandler_ReadyCheck_NotReady(t *testing.T) {
	db := openTestDB(t)
	require.NoError(t, db.Close())

	h := NewHealthHandler(db, nil)
	c, rec := newEchoContext("/ready")

	err := h.ReadyCheck(c)
	assert.NoError(t, err)
	assert.Equal(t, http_status.ServiceUnavailable.Code, rec.Code)

	payload := parseJSON(t, rec.Body.String())
	assert.Equal(t, "not ready", payload["status"])
	assert.Equal(t, "database not available", payload["error"])
}
