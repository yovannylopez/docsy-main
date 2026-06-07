package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yovannylopez/docsy-main/internal/users/domain/entities"
	userstest "github.com/yovannylopez/docsy-main/internal/users/test_utils"
)

// TestSearchUsersUseCase_Execute_Success tests the successful execution of the search use case
func TestSearchUsersUseCase_Execute_Success(t *testing.T) {
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

	// Create use case
	usecase := NewSearchUsersUseCase(mockRepo)

	request := stubs.SearchRequestWithActivo(query, activo, limit, offset)

	// Execute use case
	response, err := usecase.Execute(context.Background(), request)

	// Verify results
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response.Usuarios, 2)
	assert.Equal(t, limit, response.Limite)
	assert.Equal(t, offset, response.Offset)
	assert.Equal(t, 2, response.Total)

	// Verify user content
	assert.Equal(t, "John", response.Usuarios[0].PrimerNombre)
	assert.Equal(t, "Doe", response.Usuarios[0].SegundoNombre)
	assert.Equal(t, "john@example.com", response.Usuarios[0].Email)
	assert.True(t, response.Usuarios[0].EstaActivo)
	assert.True(t, response.Usuarios[0].EstaVerificado)

	assert.Equal(t, "Jane", response.Usuarios[1].PrimerNombre)
	assert.Equal(t, "Smith", response.Usuarios[1].SegundoNombre)
	assert.Equal(t, "jane@example.com", response.Usuarios[1].Email)
	assert.True(t, response.Usuarios[1].EstaActivo)
	assert.False(t, response.Usuarios[1].EstaVerificado)

	// Verify mock was called correctly
	mockRepo.AssertExpectations(t)
}

// TestSearchUsersUseCase_Execute_WithoutActivoFilter tests search without the active filter
func TestSearchUsersUseCase_Execute_WithoutActivoFilter(t *testing.T) {
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

	// Create use case
	usecase := NewSearchUsersUseCase(mockRepo)

	request := stubs.SearchRequestNoActivo(query, limit, offset)

	// Execute use case
	response, err := usecase.Execute(context.Background(), request)

	// Verify results
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response.Usuarios, 1)
	assert.Equal(t, 1, response.Total)

	// Verify mock was called correctly
	mockRepo.AssertExpectations(t)
}

// TestSearchUsersUseCase_Execute_EmptyQuery tests the error when the query is empty
func TestSearchUsersUseCase_Execute_EmptyQuery(t *testing.T) {
	// Set up mock
	mockRepo := new(MockUserRepository)

	// Create use case
	usecase := NewSearchUsersUseCase(mockRepo)

	stubs := userstest.NewUsersStubs()
	request := stubs.SearchRequestNoActivo("   ", 10, 0)

	// Execute use case
	response, err := usecase.Execute(context.Background(), request)

	// Verify results
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "search query cannot be empty")

	// Verify mock was not called
	mockRepo.AssertExpectations(t)
}

// TestSearchUsersUseCase_Execute_DefaultLimit tests that the default limit is used when it is 0
func TestSearchUsersUseCase_Execute_DefaultLimit(t *testing.T) {
	// Set up mock
	mockRepo := new(MockUserRepository)

	stubs := userstest.NewUsersStubs()
	users := []entities.User{*stubs.UserTestGeneric()}

	query := "test"
	defaultLimit := 10

	// Configure mock expectations (should use the default limit)
	mockRepo.On("SearchUsers", mock.Anything, query, (*bool)(nil), defaultLimit, 0).Return(users, nil)
	mockRepo.On("CountSearchUsers", mock.Anything, query, (*bool)(nil)).Return(1, nil)

	// Create use case
	usecase := NewSearchUsersUseCase(mockRepo)

	request := stubs.SearchRequestNoActivo(query, 0, 0)

	// Execute use case
	response, err := usecase.Execute(context.Background(), request)

	// Verify results
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, defaultLimit, response.Limite)

	// Verify mock was called correctly
	mockRepo.AssertExpectations(t)
}

// TestSearchUsersUseCase_Execute_NegativeOffset tests that a negative offset is adjusted to 0
func TestSearchUsersUseCase_Execute_NegativeOffset(t *testing.T) {
	// Set up mock
	mockRepo := new(MockUserRepository)

	users := []entities.User{}

	query := "test"

	// Configure mock expectations (offset must be 0)
	mockRepo.On("SearchUsers", mock.Anything, query, (*bool)(nil), 10, 0).Return(users, nil)
	mockRepo.On("CountSearchUsers", mock.Anything, query, (*bool)(nil)).Return(0, nil)

	// Create use case
	usecase := NewSearchUsersUseCase(mockRepo)

	stubs := userstest.NewUsersStubs()
	request := stubs.SearchRequestNoActivo(query, 10, -5)

	// Execute use case
	response, err := usecase.Execute(context.Background(), request)

	// Verify results
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 0, response.Offset)

	// Verify mock was called correctly
	mockRepo.AssertExpectations(t)
}

// TestSearchUsersUseCase_Execute_SearchError tests repository error handling in SearchUsers
func TestSearchUsersUseCase_Execute_SearchError(t *testing.T) {
	// Set up mock
	mockRepo := new(MockUserRepository)

	query := "test"
	activo := true
	limit := 10
	offset := 0

	// Configure mock expectations to fail
	mockRepo.On("SearchUsers", mock.Anything, query, &activo, limit, offset).
		Return([]entities.User{}, errors.New("database error"))

	// Create use case
	usecase := NewSearchUsersUseCase(mockRepo)

	stubs := userstest.NewUsersStubs()
	request := stubs.SearchRequestWithActivo(query, activo, limit, offset)

	// Execute use case
	response, err := usecase.Execute(context.Background(), request)

	// Verify results
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to search users from repository")

	// Verify mock was called correctly
	mockRepo.AssertExpectations(t)
}

// TestSearchUsersUseCase_Execute_CountError tests repository error handling in CountSearchUsers
func TestSearchUsersUseCase_Execute_CountError(t *testing.T) {
	// Set up mock
	mockRepo := new(MockUserRepository)

	stubs := userstest.NewUsersStubs()
	users := []entities.User{*stubs.UserTestGeneric()}

	query := "test"
	activo := true
	limit := 10
	offset := 0

	// Configure mock expectations
	mockRepo.On("SearchUsers", mock.Anything, query, &activo, limit, offset).Return(users, nil)
	mockRepo.On("CountSearchUsers", mock.Anything, query, &activo).
		Return(0, errors.New("database error"))

	// Create use case
	usecase := NewSearchUsersUseCase(mockRepo)

	request := stubs.SearchRequestWithActivo(query, activo, limit, offset)

	// Execute use case
	response, err := usecase.Execute(context.Background(), request)

	// Verify results
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to count search users")

	// Verify mock was called correctly
	mockRepo.AssertExpectations(t)
}
