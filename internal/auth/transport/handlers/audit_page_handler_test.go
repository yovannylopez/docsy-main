package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yovannylopez/docsy-main/internal/auth/domain"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/auth/mocks"
	authtest "github.com/yovannylopez/docsy-main/internal/auth/test_utils"
	"github.com/yovannylopez/docsy-main/internal/shared/infrastructure/templates"
)

func TestAuditPageHandler_List_ReturnsHTML(t *testing.T) {
	mockUC := mocks.NewListAuditLogsUseCase(t)
	handler := NewAuditPageHandler(mockUC)

	renderer, err := templates.NewRenderer()
	require.NoError(t, err)
	e := echo.New()
	e.Renderer = renderer

	req := httptest.NewRequest(http.MethodGet, "/auditoria?limit=10&offset=0", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &entities.User{ID: uuid.NewString(), Email: "admin@test.com", PermissionNames: []string{"audit.read"}})

	mockLogs := []entities.AuditLog{
		authtest.TransportAuditLogRow(uuid.NewString(), domain.AuditActionCreate, domain.AuditResultSuccess, authtest.StringPtr("users"), nil),
	}
	mockUC.On("Execute", mock.Anything, mock.MatchedBy(func(f *dtos.AuditLogFilters) bool {
		return f.Limit == 10 && f.Offset == 0
	})).Return(mockLogs, 1, nil)

	err = handler.List(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Logs de auditoría")
	mockUC.AssertExpectations(t)
}

func TestAuditPageHandler_List_ReturnsTablePartialForHTMX(t *testing.T) {
	mockUC := mocks.NewListAuditLogsUseCase(t)
	handler := NewAuditPageHandler(mockUC)

	renderer, err := templates.NewRenderer()
	require.NoError(t, err)
	e := echo.New()
	e.Renderer = renderer

	req := httptest.NewRequest(http.MethodGet, "/auditoria", nil)
	req.Header.Set("HX-Request", "true")
	req.Header.Set("HX-Target", "#audit-table")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &entities.User{ID: uuid.NewString(), Email: "admin@test.com"})

	mockUC.On("Execute", mock.Anything, mock.Anything).Return([]entities.AuditLog{}, 0, nil)

	err = handler.List(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Mostrando 0 de 0 logs")
	assert.NotContains(t, rec.Body.String(), "Filtros de búsqueda")
}
