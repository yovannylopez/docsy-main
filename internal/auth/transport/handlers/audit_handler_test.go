package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
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
	sharedtest "github.com/yovannylopez/docsy-main/internal/shared/test_utils"
	"github.com/yovannylopez/docsy-main/pkg/http_status"
)

func TestAuditHandler_List_Success(t *testing.T) {
	// Arrange
	mockUC := mocks.NewListAuditLogsUseCase(t)
	handler := NewAuditHandler(mockUC)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auditoria?limit=20&offset=0", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockLogs := []entities.AuditLog{
		authtest.TransportAuditLogRow(uuid.NewString(), "create", domain.AuditResultSuccess, authtest.StringPtr("communications"), nil),
	}
	total := 1

	mockUC.On("Execute", mock.Anything, mock.MatchedBy(func(f *dtos.AuditLogFilters) bool {
		return f.Limit == 20 && f.Offset == 0
	})).Return(mockLogs, total, nil)

	// Act
	err := handler.List(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http_status.OK.Code, rec.Code)

	var response map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	assert.Equal(t, "audit logs retrieved successfully", response["message"])
	assert.NotNil(t, response["data"])
	assert.NotNil(t, response["pagination"])
	mockUC.AssertExpectations(t)
}

func TestAuditHandler_List_WithFilters(t *testing.T) {
	// Arrange
	mockUC := mocks.NewListAuditLogsUseCase(t)
	handler := NewAuditHandler(mockUC)

	userID := uuid.NewString()
	action := "update"
	resource := "companies"

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auditoria?user_id="+userID+"&action="+action+"&resource="+resource+"&limit=10&offset=0", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockLogs := []entities.AuditLog{
		authtest.TransportAuditLogRow(uuid.NewString(), action, domain.AuditResultSuccess, &resource, &userID),
	}
	total := 1

	mockUC.On("Execute", mock.Anything, mock.MatchedBy(func(f *dtos.AuditLogFilters) bool {
		return f.UserID != nil && *f.UserID == userID &&
			f.Action != nil && *f.Action == action &&
			f.Resource != nil && *f.Resource == resource
	})).Return(mockLogs, total, nil)

	// Act
	err := handler.List(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http_status.OK.Code, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestAuditHandler_List_WithDateFilters(t *testing.T) {
	// Arrange
	mockUC := mocks.NewListAuditLogsUseCase(t)
	handler := NewAuditHandler(mockUC)

	startDate := "2024-01-01"
	endDate := "2024-12-31"

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auditoria?start_date="+startDate+"&end_date="+endDate+"&limit=20&offset=0", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockLogs := []entities.AuditLog{}
	total := 0

	mockUC.On("Execute", mock.Anything, mock.MatchedBy(func(f *dtos.AuditLogFilters) bool {
		return f.StartDate != nil && f.EndDate != nil
	})).Return(mockLogs, total, nil)

	// Act
	err := handler.List(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http_status.OK.Code, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestAuditHandler_List_InvalidDateFormat(t *testing.T) {
	// Arrange
	mockUC := mocks.NewListAuditLogsUseCase(t)
	handler := NewAuditHandler(mockUC)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auditoria?start_date=invalid-date&limit=20&offset=0", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Act
	err := handler.List(c)

	// Assert
	assert.NoError(t, err) // responses.BadRequest returns nil
	assert.Equal(t, http_status.BadRequest.Code, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestAuditHandler_List_InvalidLimit(t *testing.T) {
	// Arrange
	mockUC := mocks.NewListAuditLogsUseCase(t)
	handler := NewAuditHandler(mockUC)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auditoria?limit=101&offset=0", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Act
	err := handler.List(c)

	// Assert
	assert.NoError(t, err) // responses.BadRequest returns nil
	assert.Equal(t, http_status.BadRequest.Code, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestAuditHandler_List_UseCaseError(t *testing.T) {
	// Arrange
	mockUC := mocks.NewListAuditLogsUseCase(t)
	handler := NewAuditHandler(mockUC)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auditoria?limit=20&offset=0", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockUC.On("Execute", mock.Anything, mock.Anything).Return(nil, 0, assert.AnError)

	// Act
	err := handler.List(c)

	// Assert: the error propagates to Echo's HTTPErrorHandler (no response is written here)
	assert.Error(t, err)
	assert.Equal(t, http_status.OK.Code, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestAuditHandler_List_UseCaseError_WithCentralHTTPErrorHandler(t *testing.T) {
	mockUC := mocks.NewListAuditLogsUseCase(t)
	handler := NewAuditHandler(mockUC)
	mockUC.On("Execute", mock.Anything, mock.Anything).Return(nil, 0, assert.AnError)

	e := sharedtest.NewEchoWithCentralHTTPErrorHandler()
	e.GET("/api/v1/auditoria", handler.List)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auditoria?limit=20&offset=0", nil)
	rec := sharedtest.ServeEcho(e, req)

	sharedtest.AssertCentralJSONInternalServerError(t, rec, "unknown")
	mockUC.AssertExpectations(t)
}

func TestAuditHandler_List_AllFilters(t *testing.T) {
	// Arrange
	mockUC := mocks.NewListAuditLogsUseCase(t)
	handler := NewAuditHandler(mockUC)

	userID := uuid.NewString()
	sessionID := uuid.NewString()
	action := "delete"
	resource := "dependencies"
	resourceID := uuid.NewString()
	result := domain.AuditResultSuccess
	message := "test message"
	startDate := "2024-01-01"
	endDate := "2024-12-31"

	// Build URL with properly encoded parameters
	u, _ := url.Parse("/api/v1/auditoria")
	q := u.Query()
	q.Set("user_id", userID)
	q.Set("session_id", sessionID)
	q.Set("action", action)
	q.Set("resource", resource)
	q.Set("resource_id", resourceID)
	q.Set("result", result)
	q.Set("message", message)
	q.Set("start_date", startDate)
	q.Set("end_date", endDate)
	q.Set("limit", "50")
	q.Set("offset", "10")
	u.RawQuery = q.Encode()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, u.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockLogs := []entities.AuditLog{}
	total := 0

	mockUC.On("Execute", mock.Anything, mock.MatchedBy(func(f *dtos.AuditLogFilters) bool {
		return f.UserID != nil && *f.UserID == userID &&
			f.SessionID != nil && *f.SessionID == sessionID &&
			f.Action != nil && *f.Action == action &&
			f.Resource != nil && *f.Resource == resource &&
			f.ResourceID != nil && *f.ResourceID == resourceID &&
			f.Result != nil && *f.Result == result &&
			f.Message != nil && *f.Message == message &&
			f.StartDate != nil &&
			f.EndDate != nil &&
			f.Limit == 50 &&
			f.Offset == 10
	})).Return(mockLogs, total, nil)

	// Act
	err := handler.List(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http_status.OK.Code, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestAuditHandler_List_EmptyResults(t *testing.T) {
	// Arrange
	mockUC := mocks.NewListAuditLogsUseCase(t)
	handler := NewAuditHandler(mockUC)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auditoria?limit=20&offset=0", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockLogs := []entities.AuditLog{}
	total := 0

	mockUC.On("Execute", mock.Anything, mock.Anything).Return(mockLogs, total, nil)

	// Act
	err := handler.List(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http_status.OK.Code, rec.Code)

	var response map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	assert.NotNil(t, response["data"])
	data := response["data"].([]any)
	assert.Equal(t, 0, len(data))
	mockUC.AssertExpectations(t)
}

func TestAuditHandler_List_DefaultPagination(t *testing.T) {
	// Arrange
	mockUC := mocks.NewListAuditLogsUseCase(t)
	handler := NewAuditHandler(mockUC)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auditoria", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockLogs := []entities.AuditLog{}
	total := 0

	mockUC.On("Execute", mock.Anything, mock.MatchedBy(func(f *dtos.AuditLogFilters) bool {
		return f.Limit == 20 && f.Offset == 0
	})).Return(mockLogs, total, nil)

	// Act
	err := handler.List(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http_status.OK.Code, rec.Code)
	mockUC.AssertExpectations(t)
}
