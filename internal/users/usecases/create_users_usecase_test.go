package usecases

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	authdomain "github.com/yovannylopez/docsy-main/internal/auth/domain"
	authentities "github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	authmocks "github.com/yovannylopez/docsy-main/internal/auth/mocks"
	"github.com/yovannylopez/docsy-main/internal/users/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/users/mocks"
	userstest "github.com/yovannylopez/docsy-main/internal/users/test_utils"
)

func TestCreateUsersUseCase_Execute(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(mocks.UserProfileRepository)
		mockHasher := new(mocks.PasswordHasher)
		mockAudit := authmocks.NewAuditRepository(t)
		mockHasher.On("HashPassword", mock.Anything).Return("hashed_password", nil)
		useCase := NewCreateUsersUseCase(mockRepo, mockHasher, mockAudit)

		stubs := userstest.NewUsersStubs()
		req := stubs.CreateUsersRequestStandard()
		wantUsername := req.Users[0].Username

		mockRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(nil, nil)
		mockRepo.On("FindByUsername", mock.Anything, "jdoe").Return(nil, nil)
		mockRepo.On("GetRoleByName", mock.Anything, "admin").Return(&entities.Role{Name: "admin"}, nil)
		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *entities.User) bool {
			return u.Email == "test@example.com" && u.Username == wantUsername && !u.IsActive && !u.IsVerified
		})).Return(nil)
		mockAudit.On("LogAction", mock.Anything, mock.MatchedBy(func(log *authentities.AuditLog) bool {
			return log.Action == authdomain.AuditActionUserCreated && log.Result == authdomain.AuditResultSuccess
		})).Return(nil)

		resp, err := useCase.Execute(context.Background(), req, "admin-id")

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 1, resp.TotalCreated)
		assert.Empty(t, resp.Errors)
		mockRepo.AssertExpectations(t)
		mockHasher.AssertExpectations(t)
		mockAudit.AssertExpectations(t)
	})

	t.Run("UserAlreadyExists", func(t *testing.T) {
		mockRepo := new(mocks.UserProfileRepository)
		mockHasher := new(mocks.PasswordHasher)
		useCase := NewCreateUsersUseCase(mockRepo, mockHasher, nil)

		stubs := userstest.NewUsersStubs()
		req := stubs.CreateUsersRequestMinimal()
		existingUser := stubs.UserExistingByEmail("1", "test@example.com")
		mockRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(existingUser, nil)

		resp, err := useCase.Execute(context.Background(), req, "admin-id")

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 0, resp.TotalCreated)
		assert.NotEmpty(t, resp.Errors)
		assert.Contains(t, resp.Errors[0].Error, "email is already in use")
	})
}
