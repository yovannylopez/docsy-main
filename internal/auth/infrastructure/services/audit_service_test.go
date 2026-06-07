package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yovannylopez/docsy-main/internal/auth/domain"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/auth/mocks"
)

func TestAuditService_LogCreate_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := mocks.NewAuditRepository(t)
	service := NewAuditService(mockRepo)

	userID := uuid.NewString()
	resource := "communications"
	resourceID := uuid.NewString()
	newData := map[string]any{
		"id":      resourceID,
		"subject": "Test communication",
	}
	ipAddress := "192.168.1.100"
	userAgent := "Mozilla/5.0"

	mockRepo.On("LogAction", ctx, mock.MatchedBy(func(log *entities.AuditLog) bool {
		return log.Action == domain.AuditActionCreate &&
			log.Resource != nil && *log.Resource == resource &&
			log.ResourceID != nil && *log.ResourceID == resourceID &&
			log.UserID != nil && *log.UserID == userID &&
			log.Result == domain.AuditResultSuccess &&
			log.NewData != nil &&
			log.PreviousData == nil
	})).Return(nil)

	// Act
	err := service.LogCreate(ctx, userID, resource, resourceID, newData, &ipAddress, &userAgent)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAuditService_LogUpdate_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := mocks.NewAuditRepository(t)
	service := NewAuditService(mockRepo)

	userID := uuid.NewString()
	resource := "companies"
	resourceID := uuid.NewString()
	oldData := map[string]any{
		"id":    resourceID,
		"name":  "Old Company",
		"email": "old@example.com",
	}
	newData := map[string]any{
		"id":    resourceID,
		"name":  "New Company",
		"email": "new@example.com",
	}
	ipAddress := "192.168.1.100"
	userAgent := "Mozilla/5.0"

	mockRepo.On("LogAction", ctx, mock.AnythingOfType("*entities.AuditLog")).Return(nil).Run(func(args mock.Arguments) {
		log := args.Get(1).(*entities.AuditLog)
		assert.Equal(t, domain.AuditActionUpdate, log.Action)
		assert.NotNil(t, log.Resource)
		assert.Equal(t, resource, *log.Resource)
		assert.NotNil(t, log.ResourceID)
		assert.Equal(t, resourceID, *log.ResourceID)
		assert.NotNil(t, log.UserID)
		assert.Equal(t, userID, *log.UserID)
		assert.Equal(t, domain.AuditResultSuccess, log.Result)
		assert.NotNil(t, log.NewData)
		assert.NotNil(t, log.PreviousData)
		assert.Greater(t, len(log.ChangedFields), 0)
	})

	// Act
	err := service.LogUpdate(ctx, userID, resource, resourceID, oldData, newData, &ipAddress, &userAgent)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAuditService_LogDelete_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := mocks.NewAuditRepository(t)
	service := NewAuditService(mockRepo)

	userID := uuid.NewString()
	resource := "dependencies"
	resourceID := uuid.NewString()
	oldData := map[string]any{
		"id":          resourceID,
		"name":        "Dependency Test",
		"description": "Test description",
	}
	ipAddress := "192.168.1.100"
	userAgent := "Mozilla/5.0"

	mockRepo.On("LogAction", ctx, mock.MatchedBy(func(log *entities.AuditLog) bool {
		return log.Action == domain.AuditActionDelete &&
			log.Resource != nil && *log.Resource == resource &&
			log.ResourceID != nil && *log.ResourceID == resourceID &&
			log.UserID != nil && *log.UserID == userID &&
			log.Result == domain.AuditResultSuccess &&
			log.PreviousData != nil &&
			log.NewData == nil
	})).Return(nil)

	// Act
	err := service.LogDelete(ctx, userID, resource, resourceID, oldData, &ipAddress, &userAgent)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAuditService_LogRead_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := mocks.NewAuditRepository(t)
	service := NewAuditService(mockRepo)

	userID := uuid.NewString()
	resource := "users"
	resourceID := uuid.NewString()
	ipAddress := "192.168.1.100"
	userAgent := "Mozilla/5.0"

	mockRepo.On("LogAction", ctx, mock.MatchedBy(func(log *entities.AuditLog) bool {
		return log.Action == domain.AuditActionRead &&
			log.Resource != nil && *log.Resource == resource &&
			log.ResourceID != nil && *log.ResourceID == resourceID &&
			log.UserID != nil && *log.UserID == userID &&
			log.Result == domain.AuditResultSuccess
	})).Return(nil)

	// Act
	err := service.LogRead(ctx, userID, resource, &resourceID, &ipAddress, &userAgent)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAuditService_LogRead_WithoutResourceID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := mocks.NewAuditRepository(t)
	service := NewAuditService(mockRepo)

	userID := uuid.NewString()
	resource := "users"
	ipAddress := "192.168.1.100"
	userAgent := "Mozilla/5.0"

	mockRepo.On("LogAction", ctx, mock.MatchedBy(func(log *entities.AuditLog) bool {
		return log.Action == domain.AuditActionRead &&
			log.Resource != nil && *log.Resource == resource &&
			log.ResourceID == nil &&
			log.UserID != nil && *log.UserID == userID
	})).Return(nil)

	// Act
	err := service.LogRead(ctx, userID, resource, nil, &ipAddress, &userAgent)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAuditService_LogAction_RepositoryError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := mocks.NewAuditRepository(t)
	service := NewAuditService(mockRepo)

	userID := uuid.NewString()
	resource := "communications"
	resourceID := uuid.NewString()
	newData := map[string]any{"id": resourceID}
	ipAddress := "192.168.1.100"
	userAgent := "Mozilla/5.0"

	mockRepo.On("LogAction", ctx, mock.Anything).Return(errors.New("database error"))

	// Act
	err := service.LogCreate(ctx, userID, resource, resourceID, newData, &ipAddress, &userAgent)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	mockRepo.AssertExpectations(t)
}

func TestAuditService_LogUpdate_CompareData(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := mocks.NewAuditRepository(t)
	service := NewAuditService(mockRepo)

	userID := uuid.NewString()
	resource := "companies"
	resourceID := uuid.NewString()

	// Struct to test comparison
	type TestStruct struct {
		ID      string
		Name    string
		Email   string
		Updated time.Time
	}

	oldData := &TestStruct{
		ID:      resourceID,
		Name:    "Old Company",
		Email:   "old@example.com",
		Updated: time.Now(),
	}

	newData := &TestStruct{
		ID:      resourceID,
		Name:    "New Company",             // Changed
		Email:   "old@example.com",         // No changes
		Updated: time.Now().Add(time.Hour), // Changed but ignored
	}

	ipAddress := "192.168.1.100"
	userAgent := "Mozilla/5.0"

	mockRepo.On("LogAction", ctx, mock.MatchedBy(func(log *entities.AuditLog) bool {
		return log.Action == domain.AuditActionUpdate &&
			log.ChangedFields != nil &&
			len(log.ChangedFields) > 0
	})).Return(nil)

	// Act
	err := service.LogUpdate(ctx, userID, resource, resourceID, oldData, newData, &ipAddress, &userAgent)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAuditService_LogCreate_NilData(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := mocks.NewAuditRepository(t)
	service := NewAuditService(mockRepo)

	userID := uuid.NewString()
	resource := "communications"
	resourceID := uuid.NewString()
	ipAddress := "192.168.1.100"
	userAgent := "Mozilla/5.0"

	mockRepo.On("LogAction", ctx, mock.MatchedBy(func(log *entities.AuditLog) bool {
		return log.NewData == nil
	})).Return(nil)

	// Act
	err := service.LogCreate(ctx, userID, resource, resourceID, nil, &ipAddress, &userAgent)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAuditService_LogAction_Custom(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := mocks.NewAuditRepository(t)
	service := NewAuditService(mockRepo)

	userID := uuid.NewString()
	action := "assign"
	resource := "communications"
	resourceID := uuid.NewString()
	result := domain.AuditResultSuccess
	message := "Assignment completed"
	ipAddress := "192.168.1.100"
	userAgent := "Mozilla/5.0"

	mockRepo.On("LogAction", ctx, mock.MatchedBy(func(log *entities.AuditLog) bool {
		return log.Action == action &&
			log.Resource != nil && *log.Resource == resource &&
			log.Result == result &&
			log.Message != nil && *log.Message == message
	})).Return(nil)

	// Act
	err := service.LogAction(ctx, userID, action, resource, &resourceID, result, message, &ipAddress, &userAgent, nil, nil)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAuditService_CompareData_Struct(t *testing.T) {
	// Arrange
	service := &AuditService{}

	type TestStruct struct {
		ID        string
		Name      string
		Email     string
		UpdatedAt time.Time
		Password  string // Should be ignored
	}

	oldData := &TestStruct{
		ID:        "123",
		Name:      "Old",
		Email:     "old@example.com",
		UpdatedAt: time.Now(),
		Password:  "secret",
	}

	newData := &TestStruct{
		ID:        "123",
		Name:      "New",                     // Changed
		Email:     "new@example.com",         // Changed
		UpdatedAt: time.Now().Add(time.Hour), // Changed but ignored
		Password:  "newsecret",               // Changed but ignored
	}

	// Act
	changedFields := service.compareData(oldData, newData)

	// Assert
	assert.Contains(t, changedFields, "Name")
	assert.Contains(t, changedFields, "Email")
	assert.NotContains(t, changedFields, "UpdatedAt")
	assert.NotContains(t, changedFields, "Password")
}

func TestAuditService_CompareData_NilValues(t *testing.T) {
	// Arrange
	service := &AuditService{}

	// Act
	changedFields1 := service.compareData(nil, map[string]any{"key": "value"})
	changedFields2 := service.compareData(map[string]any{"key": "value"}, nil)

	// Assert
	assert.Empty(t, changedFields1)
	assert.Empty(t, changedFields2)
}

func TestAuditService_CompareData_NoChanges(t *testing.T) {
	// Arrange
	service := &AuditService{}

	type TestStruct struct {
		ID   string
		Name string
	}

	oldData := &TestStruct{
		ID:   "123",
		Name: "Test",
	}

	newData := &TestStruct{
		ID:   "123",
		Name: "Test",
	}

	// Act
	changedFields := service.compareData(oldData, newData)

	// Assert
	assert.Empty(t, changedFields)
}

func TestAuditService_MarshalData_Error(t *testing.T) {
	// Arrange
	service := &AuditService{}

	// Create data that cannot be serialized (channel)
	unmarshalableData := make(chan int)

	// Act
	result, err := service.marshalData(unmarshalableData)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
}
