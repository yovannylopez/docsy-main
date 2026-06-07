package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	authdtos "github.com/yovannylopez/docsy-main/internal/auth/domain/dtos"
	authentities "github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	sharedtest "github.com/yovannylopez/docsy-main/internal/shared/test_utils"
	"github.com/yovannylopez/docsy-main/internal/users/domain/entities"
	userstest "github.com/yovannylopez/docsy-main/internal/users/test_utils"
	"github.com/yovannylopez/docsy-main/internal/users/usecases"
	"github.com/yovannylopez/docsy-main/pkg/http_status"
)

// stubPasswordHasher is a stub for tests that do not verify the hash
type stubPasswordHasher struct{}

func (s *stubPasswordHasher) HashPassword(password string) (string, error) {
	return "hashed_" + password, nil
}

// noopAuditRepo satisfies auth AuditRepository for handler tests that do not assert audit calls.
type noopAuditRepo struct{}

func (noopAuditRepo) LogAction(_ context.Context, _ *authentities.AuditLog) error { return nil }

func (noopAuditRepo) GetUserAuditLogs(_ context.Context, _ string, _, _ int) ([]authentities.AuditLog, error) {
	return nil, nil
}

func (noopAuditRepo) GetSessionAuditLogs(_ context.Context, _ string, _, _ int) ([]authentities.AuditLog, error) {
	return nil, nil
}

func (noopAuditRepo) List(_ context.Context, _ *authdtos.AuditLogFilters) ([]authentities.AuditLog, int, error) {
	return nil, 0, nil
}

// newTestHandler creates a UsersHandler with use cases injected using the given mock
func newTestHandler(mockRepo *MockUserRepository) *UsersHandler {
	return NewUsersHandler(
		usecases.NewGetUsersUseCase(mockRepo),
		usecases.NewCreateUsersUseCase(mockRepo, &stubPasswordHasher{}, noopAuditRepo{}),
		usecases.NewUpdateUserUseCase(mockRepo),
		usecases.NewSearchUsersUseCase(mockRepo),
		usecases.NewGetUserByIDUseCase(mockRepo),
	)
}

// MockUserRepository is a mock of the user repository for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) FindByUsername(ctx context.Context, username string) (*entities.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) FindByID(ctx context.Context, userID string) (*entities.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Update(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) GetAllUsers(ctx context.Context, limit, offset int) ([]entities.User, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]entities.User), args.Error(1)
}

func (m *MockUserRepository) GetTotalUsersCount(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *MockUserRepository) SearchUsers(ctx context.Context, query string, activo *bool, limit, offset int) ([]entities.User, error) {
	args := m.Called(ctx, query, activo, limit, offset)
	return args.Get(0).([]entities.User), args.Error(1)
}

func (m *MockUserRepository) CountSearchUsers(ctx context.Context, query string, activo *bool) (int, error) {
	args := m.Called(ctx, query, activo)
	return args.Int(0), args.Error(1)
}

func (m *MockUserRepository) GetUsersByRole(ctx context.Context, roleName string, limit, offset int) ([]entities.User, error) {
	args := m.Called(ctx, roleName, limit, offset)
	return args.Get(0).([]entities.User), args.Error(1)
}

func (m *MockUserRepository) GetActiveUsers(ctx context.Context, limit, offset int) ([]entities.User, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]entities.User), args.Error(1)
}

func (m *MockUserRepository) GetRoleByName(ctx context.Context, roleName string) (*entities.Role, error) {
	args := m.Called(ctx, roleName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Role), args.Error(1)
}

func (m *MockUserRepository) AssignRoleToUser(ctx context.Context, userID, roleID string) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *MockUserRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID string) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserRoles(ctx context.Context, userID string) ([]entities.Role, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]entities.Role), args.Error(1)
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) IncrementFailedLoginAttempts(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) ResetFailedLoginAttempts(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) LockUserAccount(ctx context.Context, userID string, until *time.Time) error {
	args := m.Called(ctx, userID, until)
	return args.Error(0)
}

func (m *MockUserRepository) UnlockUserAccount(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	args := m.Called(ctx, userID, passwordHash)
	return args.Error(0)
}

func (m *MockUserRepository) SetMustChangePassword(ctx context.Context, userID string, mustChange bool) error {
	args := m.Called(ctx, userID, mustChange)
	return args.Error(0)
}

func (m *MockUserRepository) EnableMFA(ctx context.Context, userID, secret string) error {
	args := m.Called(ctx, userID, secret)
	return args.Error(0)
}

func (m *MockUserRepository) DisableMFA(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) VerifyUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) CreateVerificationToken(ctx context.Context, token *entities.VerificationToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockUserRepository) FindVerificationToken(ctx context.Context, tokenHash string) (*entities.VerificationToken, error) {
	args := m.Called(ctx, tokenHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.VerificationToken), args.Error(1)
}

func (m *MockUserRepository) MarkTokenAsUsed(ctx context.Context, tokenID string) error {
	args := m.Called(ctx, tokenID)
	return args.Error(0)
}

func (m *MockUserRepository) CleanExpiredTokens(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// TestGetUsers_Success tests the GET /api/v1/users endpoint with successful pagination
func TestGetUsers_Success(t *testing.T) {
	// Set up mock
	mockRepo := new(MockUserRepository)

	stubs := userstest.NewUsersStubs()
	users := []entities.User{*stubs.UserJohnPerez(), *stubs.UserMariaGarcia()}

	// Configure mock expectations
	mockRepo.On("GetAllUsers", mock.Anything, 10, 0).Return(users, nil)
	mockRepo.On("GetTotalUsersCount", mock.Anything).Return(150, nil)

	// Create handler
	handler := newTestHandler(mockRepo)

	// Configure Echo
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users?limit=10&offset=0", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute handler
	err := handler.GetUsers(c)

	// Verify results
	assert.NoError(t, err)
	assert.Equal(t, http_status.OK.Code, rec.Code)

	// Verify JSON response
	var response map[string]any
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify paginated response structure
	assert.Contains(t, response, "data")
	assert.Contains(t, response, "pagination")
	assert.Contains(t, response, "message")
	assert.Contains(t, response, "status")

	// Verify pagination metadata
	pagination := response["pagination"].(map[string]any)
	assert.Equal(t, float64(150), pagination["total"])
	assert.Equal(t, float64(10), pagination["limit"])
	assert.Equal(t, float64(0), pagination["offset"])
	assert.Equal(t, float64(15), pagination["total_pages"])
	assert.Equal(t, float64(1), pagination["current_page"])
	assert.Equal(t, true, pagination["has_next"])
	assert.Equal(t, false, pagination["has_previous"])

	// Verify mock was called correctly
	mockRepo.AssertExpectations(t)
}

// TestGetUsers_InvalidLimit tests the endpoint with an invalid limit
func TestGetUsers_InvalidLimit(t *testing.T) {
	// Set up mock
	mockRepo := new(MockUserRepository)

	// Create handler
	handler := newTestHandler(mockRepo)

	// Configure Echo
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users?limit=invalid", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute handler
	err := handler.GetUsers(c)

	// Verify results
	assert.NoError(t, err)
	assert.Equal(t, http_status.BadRequest.Code, rec.Code)

	// Verify JSON response
	var response map[string]any
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
	assert.Contains(t, response["error"], "parameter 'limit' must be an integer")
}

// TestGetUsers_InvalidOffset tests the endpoint with an invalid offset
func TestGetUsers_InvalidOffset(t *testing.T) {
	// Set up mock
	mockRepo := new(MockUserRepository)

	// Create handler
	handler := newTestHandler(mockRepo)

	// Configure Echo
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users?offset=invalid", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute handler
	err := handler.GetUsers(c)

	// Verify results
	assert.NoError(t, err)
	assert.Equal(t, http_status.BadRequest.Code, rec.Code)

	// Verify JSON response
	var response map[string]any
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
	assert.Contains(t, response["error"], "parameter 'offset' must be an integer")
}

// TestGetUsers_LimitTooHigh tests the endpoint with a limit that is too high
func TestGetUsers_LimitTooHigh(t *testing.T) {
	// Set up mock
	mockRepo := new(MockUserRepository)

	// Create handler
	handler := newTestHandler(mockRepo)

	// Configure Echo
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users?limit=200", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute handler
	err := handler.GetUsers(c)

	// Verify results
	assert.NoError(t, err)
	assert.Equal(t, http_status.BadRequest.Code, rec.Code)

	// Verify JSON response
	var response map[string]any
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
	assert.Contains(t, response["error"], "maximum 100")
}

// TestGetUsers_NegativeOffset tests the endpoint with a negative offset
func TestGetUsers_NegativeOffset(t *testing.T) {
	// Set up mock
	mockRepo := new(MockUserRepository)

	// Create handler
	handler := newTestHandler(mockRepo)

	// Configure Echo
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users?offset=-5", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute handler
	err := handler.GetUsers(c)

	// Verify results
	assert.NoError(t, err)
	assert.Equal(t, http_status.BadRequest.Code, rec.Code)

	// Verify JSON response
	var response map[string]any
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
	assert.Contains(t, response["error"], "negative offset")
}

// TestGetUsers_DefaultPagination tests the endpoint without pagination parameters
func TestGetUsers_DefaultPagination(t *testing.T) {
	// Set up mock
	mockRepo := new(MockUserRepository)

	stubs := userstest.NewUsersStubs()
	users := []entities.User{*stubs.UserJohnPerez()}

	// Configure mock expectations
	mockRepo.On("GetAllUsers", mock.Anything, 10, 0).Return(users, nil)
	mockRepo.On("GetTotalUsersCount", mock.Anything).Return(50, nil)

	// Create handler
	handler := newTestHandler(mockRepo)

	// Configure Echo
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute handler
	err := handler.GetUsers(c)

	// Verify results
	assert.NoError(t, err)
	assert.Equal(t, http_status.OK.Code, rec.Code)

	// Verify JSON response
	var response map[string]any
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify pagination metadata with default values
	pagination := response["pagination"].(map[string]any)
	assert.Equal(t, float64(10), pagination["limit"])
	assert.Equal(t, float64(0), pagination["offset"])

	// Verify mock was called correctly
	mockRepo.AssertExpectations(t)
}

// TestSearchUsers_Success tests the GET /api/v1/users/search endpoint with successful search
func TestSearchUsers_Success(t *testing.T) {
	// Set up mock
	mockRepo := new(MockUserRepository)

	stubs := userstest.NewUsersStubs()
	users := []entities.User{*stubs.UserJohnDoeSearch(), *stubs.UserJaneSmithSearch()}

	query := "john"
	activo := true
	limit := 10
	offset := 0

	// Configure mock expectations
	mockRepo.On("SearchUsers", mock.Anything, query, &activo, limit, offset).Return(users, nil)
	mockRepo.On("CountSearchUsers", mock.Anything, query, &activo).Return(2, nil)

	// Create handler
	handler := newTestHandler(mockRepo)

	// Configure Echo
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/search?q=john&limit=10&offset=0&activo=true", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute handler
	err := handler.SearchUsers(c)

	// Verify results
	assert.NoError(t, err)
	assert.Equal(t, http_status.OK.Code, rec.Code)

	// Verify JSON response
	var response map[string]any
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify response structure
	assert.Contains(t, response, "data")
	assert.Contains(t, response, "pagination")
	assert.Contains(t, response, "message")
	assert.Contains(t, response, "status")

	// Verify pagination metadata
	pagination := response["pagination"].(map[string]any)
	assert.Equal(t, float64(2), pagination["total"])
	assert.Equal(t, float64(10), pagination["limit"])
	assert.Equal(t, float64(0), pagination["offset"])

	// Verify mock was called correctly
	mockRepo.AssertExpectations(t)
}

// TestSearchUsers_WithoutActivoFilter tests search without the active filter
func TestSearchUsers_WithoutActivoFilter(t *testing.T) {
	// Set up mock
	mockRepo := new(MockUserRepository)

	stubs := userstest.NewUsersStubs()
	users := []entities.User{*stubs.UserTestGeneric()}

	query := "test"
	limit := 10
	offset := 0

	// Configure mock expectations (activo is nil)
	mockRepo.On("SearchUsers", mock.Anything, query, (*bool)(nil), limit, offset).Return(users, nil)
	mockRepo.On("CountSearchUsers", mock.Anything, query, (*bool)(nil)).Return(1, nil)

	// Create handler
	handler := newTestHandler(mockRepo)

	// Configure Echo
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/search?q=test&limit=10&offset=0", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute handler
	err := handler.SearchUsers(c)

	// Verify results
	assert.NoError(t, err)
	assert.Equal(t, http_status.OK.Code, rec.Code)

	// Verify mock was called correctly
	mockRepo.AssertExpectations(t)
}

// TestSearchUsers_EmptyQuery tests the error when the q parameter is missing
func TestSearchUsers_EmptyQuery(t *testing.T) {
	// Set up mock
	mockRepo := new(MockUserRepository)

	// Create handler
	handler := newTestHandler(mockRepo)

	// Configure Echo
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/search?limit=10&offset=0", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute handler
	err := handler.SearchUsers(c)

	// Verify results
	assert.NoError(t, err) // The handler returns BadRequest, not an error
	assert.Equal(t, http_status.BadRequest.Code, rec.Code)

	// Verify mock was not called
	mockRepo.AssertExpectations(t)
}

// TestSearchUsers_InvalidActivo tests the error when activo has an invalid value
func TestSearchUsers_InvalidActivo(t *testing.T) {
	// Set up mock
	mockRepo := new(MockUserRepository)

	// Create handler
	handler := newTestHandler(mockRepo)

	// Configure Echo
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/search?q=test&activo=invalid", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute handler
	err := handler.SearchUsers(c)

	// Verify results
	assert.NoError(t, err) // The handler returns BadRequest, not an error
	assert.Equal(t, http_status.BadRequest.Code, rec.Code)

	// Verify mock was not called
	mockRepo.AssertExpectations(t)
}

// TestSearchUsers_RepositoryError tests repository error handling
func TestSearchUsers_RepositoryError(t *testing.T) {
	// Set up mock
	mockRepo := new(MockUserRepository)

	query := "test"
	activo := true
	limit := 10
	offset := 0

	// Configure mock expectations to fail
	mockRepo.On("SearchUsers", mock.Anything, query, &activo, limit, offset).
		Return([]entities.User{}, errors.New("database error"))

	// Create handler
	handler := newTestHandler(mockRepo)

	// Configure Echo
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/search?q=test&limit=10&offset=0&activo=true", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute handler
	err := handler.SearchUsers(c)

	// Verify results: the error propagates (without HTTPErrorHandler there is no 500 response in the recorder)
	assert.Error(t, err)
	assert.Equal(t, http_status.OK.Code, rec.Code)

	// Verify mock was called correctly
	mockRepo.AssertExpectations(t)
}

func TestSearchUsers_RepositoryError_WithCentralHTTPErrorHandler(t *testing.T) {
	mockRepo := new(MockUserRepository)

	query := "test"
	activo := true
	limit := 10
	offset := 0

	mockRepo.On("SearchUsers", mock.Anything, query, &activo, limit, offset).
		Return([]entities.User{}, errors.New("database error"))

	handler := newTestHandler(mockRepo)

	e := sharedtest.NewEchoWithCentralHTTPErrorHandler()
	e.GET("/api/v1/users/search", handler.SearchUsers)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/search?q=test&limit=10&offset=0&activo=true", nil)
	rec := sharedtest.ServeEcho(e, req)

	sharedtest.AssertCentralJSONInternalServerError(t, rec, "")
	mockRepo.AssertExpectations(t)
}

// TestSearchUsers_WithPagination tests search with custom pagination
func TestSearchUsers_WithPagination(t *testing.T) {
	// Set up mock
	mockRepo := new(MockUserRepository)

	stubs := userstest.NewUsersStubs()
	users := []entities.User{*stubs.UserTestGeneric()}

	query := "test"
	limit := 5
	offset := 10

	// Configure mock expectations
	mockRepo.On("SearchUsers", mock.Anything, query, (*bool)(nil), limit, offset).Return(users, nil)
	mockRepo.On("CountSearchUsers", mock.Anything, query, (*bool)(nil)).Return(50, nil)

	// Create handler
	handler := newTestHandler(mockRepo)

	// Configure Echo
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/search?q=test&limit=5&offset=10", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute handler
	err := handler.SearchUsers(c)

	// Verify results
	assert.NoError(t, err)
	assert.Equal(t, http_status.OK.Code, rec.Code)

	// Verify JSON response
	var response map[string]any
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify pagination metadata
	pagination := response["pagination"].(map[string]any)
	assert.Equal(t, float64(50), pagination["total"])
	assert.Equal(t, float64(5), pagination["limit"])
	assert.Equal(t, float64(10), pagination["offset"])

	// Verify mock was called correctly
	mockRepo.AssertExpectations(t)
}

// stringPtr is a helper function to create string pointers
