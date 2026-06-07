package usecases

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yovannylopez/docsy-main/internal/users/domain/entities"
	userstest "github.com/yovannylopez/docsy-main/internal/users/test_utils"
)

// wrapError wraps errors from external packages
func wrapError(err error, operation string) error {
	if err != nil {
		return fmt.Errorf("mock repository %s operation failed: %w", operation, err)
	}

	return nil
}

// MockUserRepository is a mock of the user repository for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) FindByUsername(ctx context.Context, username string) (*entities.User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(*entities.User), wrapError(args.Error(1), "FindByUsername")
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*entities.User), wrapError(args.Error(1), "FindByEmail")
}

func (m *MockUserRepository) Create(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return wrapError(args.Error(0), "Create")
}

func (m *MockUserRepository) Update(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return wrapError(args.Error(0), "Update")
}

func (m *MockUserRepository) Delete(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return wrapError(args.Error(0), "Delete")
}

func (m *MockUserRepository) GetRoleByName(ctx context.Context, roleName string) (*entities.Role, error) {
	args := m.Called(ctx, roleName)
	return args.Get(0).(*entities.Role), wrapError(args.Error(1), "GetRoleByName")
}

func (m *MockUserRepository) GetAllUsers(ctx context.Context, limit, offset int) ([]entities.User, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]entities.User), wrapError(args.Error(1), "GetAllUsers")
}

func (m *MockUserRepository) GetUsersByRole(ctx context.Context, roleName string, limit, offset int) ([]entities.User, error) {
	args := m.Called(ctx, roleName, limit, offset)
	return args.Get(0).([]entities.User), wrapError(args.Error(1), "GetUsersByRole")
}

func (m *MockUserRepository) GetActiveUsers(ctx context.Context, limit, offset int) ([]entities.User, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]entities.User), wrapError(args.Error(1), "GetActiveUsers")
}

func (m *MockUserRepository) GetTotalUsersCount(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), wrapError(args.Error(1), "GetTotalUsersCount")
}

func (m *MockUserRepository) SearchUsers(ctx context.Context, query string, activo *bool, limit, offset int) ([]entities.User, error) {
	args := m.Called(ctx, query, activo, limit, offset)
	return args.Get(0).([]entities.User), wrapError(args.Error(1), "SearchUsers")
}

func (m *MockUserRepository) CountSearchUsers(ctx context.Context, query string, activo *bool) (int, error) {
	args := m.Called(ctx, query, activo)
	return args.Int(0), wrapError(args.Error(1), "CountSearchUsers")
}

func (m *MockUserRepository) AssignRoleToUser(ctx context.Context, userID, roleID string) error {
	args := m.Called(ctx, userID, roleID)
	return wrapError(args.Error(0), "AssignRoleToUser")
}

func (m *MockUserRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID string) error {
	args := m.Called(ctx, userID, roleID)
	return wrapError(args.Error(0), "RemoveRoleFromUser")
}

func (m *MockUserRepository) GetUserRoles(ctx context.Context, userID string) ([]entities.Role, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]entities.Role), wrapError(args.Error(1), "GetUserRoles")
}

func (m *MockUserRepository) FindByID(ctx context.Context, userID string) (*entities.User, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*entities.User), wrapError(args.Error(1), "FindByID")
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return wrapError(args.Error(0), "UpdateLastLogin")
}

func (m *MockUserRepository) IncrementFailedLoginAttempts(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return wrapError(args.Error(0), "IncrementFailedLoginAttempts")
}

func (m *MockUserRepository) ResetFailedLoginAttempts(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return wrapError(args.Error(0), "ResetFailedLoginAttempts")
}

func (m *MockUserRepository) LockUserAccount(ctx context.Context, userID string, until *time.Time) error {
	args := m.Called(ctx, userID, until)
	return wrapError(args.Error(0), "LockUserAccount")
}

func (m *MockUserRepository) UnlockUserAccount(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return wrapError(args.Error(0), "UnlockUserAccount")
}

func (m *MockUserRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	args := m.Called(ctx, userID, passwordHash)
	return wrapError(args.Error(0), "UpdatePassword")
}

func (m *MockUserRepository) SetMustChangePassword(ctx context.Context, userID string, mustChange bool) error {
	args := m.Called(ctx, userID, mustChange)
	return wrapError(args.Error(0), "SetMustChangePassword")
}

func (m *MockUserRepository) EnableMFA(ctx context.Context, userID, secret string) error {
	args := m.Called(ctx, userID, secret)
	return wrapError(args.Error(0), "EnableMFA")
}

func (m *MockUserRepository) DisableMFA(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return wrapError(args.Error(0), "DisableMFA")
}

func (m *MockUserRepository) VerifyUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return wrapError(args.Error(0), "VerifyUser")
}

func (m *MockUserRepository) CreateVerificationToken(ctx context.Context, token *entities.VerificationToken) error {
	args := m.Called(ctx, token)
	return wrapError(args.Error(0), "CreateVerificationToken")
}

func (m *MockUserRepository) FindVerificationToken(ctx context.Context, tokenHash string) (*entities.VerificationToken, error) {
	args := m.Called(ctx, tokenHash)
	return args.Get(0).(*entities.VerificationToken), wrapError(args.Error(1), "FindVerificationToken")
}

func (m *MockUserRepository) MarkTokenAsUsed(ctx context.Context, tokenID string) error {
	args := m.Called(ctx, tokenID)
	return wrapError(args.Error(0), "MarkTokenAsUsed")
}

func (m *MockUserRepository) CleanExpiredTokens(ctx context.Context) error {
	args := m.Called(ctx)
	return wrapError(args.Error(0), "CleanExpiredTokens")
}

// TestGetUsersUseCase_Execute_Success tests the successful execution of the use case
func TestGetUsersUseCase_Execute_Success(t *testing.T) {
	// Set up mock
	mockRepo := new(MockUserRepository)

	stubs := userstest.NewUsersStubs()
	users := []entities.User{*stubs.UserJohnPerez(), *stubs.UserMariaGarcia()}

	// Configure mock expectations
	mockRepo.On("GetAllUsers", mock.Anything, 10, 0).Return(users, nil)
	mockRepo.On("GetTotalUsersCount", mock.Anything).Return(150, nil)

	// Create use case
	usecase := NewGetUsersUseCase(mockRepo)

	// Create request
	request := &GetUsersRequest{
		Limit:  10,
		Offset: 0,
	}

	// Execute use case
	response, err := usecase.Execute(context.Background(), request)

	// Verify results
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response.Usuarios, 2)
	assert.Equal(t, 10, response.Limite)
	assert.Equal(t, 0, response.Offset)
	assert.Equal(t, 150, response.Total) // Actual total from the database

	// Verify user content
	assert.Equal(t, "John", response.Usuarios[0].PrimerNombre)
	assert.Equal(t, "Perez", response.Usuarios[0].SegundoNombre)
	assert.Equal(t, "test1@example.com", response.Usuarios[0].Email)
	assert.True(t, response.Usuarios[0].EstaActivo)
	assert.True(t, response.Usuarios[0].EstaVerificado)

	assert.Equal(t, "Maria", response.Usuarios[1].PrimerNombre)
	assert.Equal(t, "Garcia", response.Usuarios[1].SegundoNombre)
	assert.Equal(t, "test2@example.com", response.Usuarios[1].Email)
	assert.True(t, response.Usuarios[1].EstaActivo)
	assert.False(t, response.Usuarios[1].EstaVerificado)
	assert.Equal(t, "+1234567890", *response.Usuarios[1].Telefono)

	// Verify mock was called correctly
	mockRepo.AssertExpectations(t)
}

// TestGetUsersUseCase_Execute_RepositoryError tests repository error handling
func TestGetUsersUseCase_Execute_RepositoryError(t *testing.T) {
	// Set up mock
	mockRepo := new(MockUserRepository)

	// Configure mock expectations to fail
	mockRepo.On("GetAllUsers", mock.Anything, 10, 0).Return([]entities.User{}, errors.New("database error"))

	// Create use case
	usecase := NewGetUsersUseCase(mockRepo)

	// Create request
	request := &GetUsersRequest{
		Limit:  10,
		Offset: 0,
	}

	// Execute use case
	response, err := usecase.Execute(context.Background(), request)

	// Verify results
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to get users from repository")

	// Verify mock was called correctly
	mockRepo.AssertExpectations(t)
}
