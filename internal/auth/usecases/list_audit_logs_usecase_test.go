package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yovannylopez/docsy-main/internal/auth/domain"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/auth/mocks"
)

func TestListAuditLogsUseCase_Execute_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := mocks.NewAuditRepository(t)
	useCase := NewListAuditLogsUseCase(mockRepo)

	filters := &dtos.AuditLogFilters{
		Limit:  20,
		Offset: 0,
	}

	mockLogs := []entities.AuditLog{
		{
			ID:        uuid.NewString(),
			Action:    "create",
			Resource:  stringPtr("comunicaciones"),
			Result:    domain.AuditResultSuccess,
			CreatedAt: time.Now(),
		},
		{
			ID:        uuid.NewString(),
			Action:    "update",
			Resource:  stringPtr("empresas"),
			Result:    domain.AuditResultSuccess,
			CreatedAt: time.Now(),
		},
	}
	total := 2

	mockRepo.On("List", ctx, mock.MatchedBy(func(f *dtos.AuditLogFilters) bool {
		return f.Limit == 20 && f.Offset == 0
	})).Return(mockLogs, total, nil)

	// Act
	logs, count, err := useCase.Execute(ctx, filters)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, logs)
	assert.Equal(t, len(mockLogs), len(logs))
	assert.Equal(t, total, count)
	mockRepo.AssertExpectations(t)
}

func TestListAuditLogsUseCase_Execute_WithFilters(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := mocks.NewAuditRepository(t)
	useCase := NewListAuditLogsUseCase(mockRepo)

	userID := uuid.NewString()
	action := "update"
	resource := "comunicaciones"
	startDate := time.Now().AddDate(0, 0, -7)
	endDate := time.Now()

	filters := &dtos.AuditLogFilters{
		UserID:    &userID,
		Action:    &action,
		Resource:  &resource,
		StartDate: &startDate,
		EndDate:   &endDate,
		Limit:     10,
		Offset:    0,
	}

	mockLogs := []entities.AuditLog{
		{
			ID:        uuid.NewString(),
			UserID:    &userID,
			Action:    action,
			Resource:  &resource,
			Result:    domain.AuditResultSuccess,
			CreatedAt: time.Now(),
		},
	}
	total := 1

	mockRepo.On("List", ctx, mock.MatchedBy(func(f *dtos.AuditLogFilters) bool {
		return f.UserID != nil && *f.UserID == userID &&
			f.Action != nil && *f.Action == action &&
			f.Resource != nil && *f.Resource == resource
	})).Return(mockLogs, total, nil)

	// Act
	logs, count, err := useCase.Execute(ctx, filters)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, logs)
	assert.Equal(t, 1, len(logs))
	assert.Equal(t, total, count)
	mockRepo.AssertExpectations(t)
}

func TestListAuditLogsUseCase_Execute_RepositoryError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := mocks.NewAuditRepository(t)
	useCase := NewListAuditLogsUseCase(mockRepo)

	filters := &dtos.AuditLogFilters{
		Limit:  20,
		Offset: 0,
	}

	mockRepo.On("List", ctx, mock.Anything).Return(nil, 0, errors.New("database error"))

	// Act
	logs, count, err := useCase.Execute(ctx, filters)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, logs)
	assert.Equal(t, 0, count)
	assert.Contains(t, err.Error(), "failed to list audit logs")
	mockRepo.AssertExpectations(t)
}

func TestListAuditLogsUseCase_Execute_NilFilters(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := mocks.NewAuditRepository(t)
	useCase := NewListAuditLogsUseCase(mockRepo)

	// Act
	logs, count, err := useCase.Execute(ctx, nil)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, logs)
	assert.Equal(t, 0, count)
	assert.Contains(t, err.Error(), "filters cannot be nil")
	mockRepo.AssertExpectations(t)
}

func TestListAuditLogsUseCase_Execute_InvalidLimit(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := mocks.NewAuditRepository(t)
	useCase := NewListAuditLogsUseCase(mockRepo)

	tests := []struct {
		name    string
		filters *dtos.AuditLogFilters
	}{
		{
			name: "limit too large",
			filters: &dtos.AuditLogFilters{
				Limit:  101,
				Offset: 0,
			},
		},
		{
			name: "negative offset",
			filters: &dtos.AuditLogFilters{
				Limit:  20,
				Offset: -1,
			},
		},
		{
			name: "negative limit",
			filters: &dtos.AuditLogFilters{
				Limit:  -1,
				Offset: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			logs, count, err := useCase.Execute(ctx, tt.filters)

			// Assert
			assert.Error(t, err)
			assert.Nil(t, logs)
			assert.Equal(t, 0, count)
		})
	}
}

func TestListAuditLogsUseCase_Execute_InvalidDateRange(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := mocks.NewAuditRepository(t)
	useCase := NewListAuditLogsUseCase(mockRepo)

	startDate := time.Now()
	endDate := time.Now().AddDate(0, 0, -7) // End date before start date

	filters := &dtos.AuditLogFilters{
		StartDate: &startDate,
		EndDate:   &endDate,
		Limit:     20,
		Offset:    0,
	}

	// Act
	logs, count, err := useCase.Execute(ctx, filters)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, logs)
	assert.Equal(t, 0, count)
	assert.Contains(t, err.Error(), "start_date cannot be after end_date")
	mockRepo.AssertExpectations(t)
}

func TestListAuditLogsUseCase_Execute_InvalidResult(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := mocks.NewAuditRepository(t)
	useCase := NewListAuditLogsUseCase(mockRepo)

	invalidResult := "invalid_result"
	filters := &dtos.AuditLogFilters{
		Result: &invalidResult,
		Limit:  20,
		Offset: 0,
	}

	// Act
	logs, count, err := useCase.Execute(ctx, filters)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, logs)
	assert.Equal(t, 0, count)
	assert.Contains(t, err.Error(), "invalid result value")
	mockRepo.AssertExpectations(t)
}

func TestListAuditLogsUseCase_Execute_ApplyDefaults(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := mocks.NewAuditRepository(t)
	useCase := NewListAuditLogsUseCase(mockRepo)

	filters := &dtos.AuditLogFilters{
		Limit:  0, // Should default to 20
		Offset: 0,
	}

	mockLogs := []entities.AuditLog{}
	total := 0

	mockRepo.On("List", ctx, mock.MatchedBy(func(f *dtos.AuditLogFilters) bool {
		return f.Limit == 20 && f.Offset == 0
	})).Return(mockLogs, total, nil)

	// Act
	logs, count, err := useCase.Execute(ctx, filters)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, logs)
	assert.Equal(t, total, count)
	mockRepo.AssertExpectations(t)
}

func TestListAuditLogsUseCase_Execute_ValidResults(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := mocks.NewAuditRepository(t)
	useCase := NewListAuditLogsUseCase(mockRepo)

	validResults := []string{
		domain.AuditResultSuccess,
		domain.AuditResultFailure,
		domain.AuditResultError,
	}

	for _, res := range validResults {
		t.Run("result_"+res, func(t *testing.T) {
			r := res
			filters := &dtos.AuditLogFilters{
				Result: &r,
				Limit:  20,
				Offset: 0,
			}

			mockLogs := []entities.AuditLog{}
			total := 0

			mockRepo.On("List", ctx, mock.Anything).Return(mockLogs, total, nil)

			// Act
			logs, count, err := useCase.Execute(ctx, filters)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, logs)
			assert.Equal(t, total, count)
		})
	}
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
