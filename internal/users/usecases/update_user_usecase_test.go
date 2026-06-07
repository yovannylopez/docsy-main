package usecases

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yovannylopez/docsy-main/internal/users/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/users/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/users/mocks"
	userstest "github.com/yovannylopez/docsy-main/internal/users/test_utils"
)

func TestUpdateUserUseCase_Execute(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		mockRepo := new(mocks.UserProfileRepository)
		useCase := NewUpdateUserUseCase(mockRepo)

		stubs := userstest.NewUsersStubs()
		userID := "user-123"
		firstName := "Jane"
		req := stubs.UpdateRequestFirstName(firstName)

		existingUser := stubs.UserForUpdate(userID, "old@example.com", "John")

		mockRepo.On("FindByID", mock.Anything, userID).Return(existingUser, nil)
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *entities.User) bool {
			return u.ID == userID && u.FirstName == firstName
		})).Return(nil)

		// Act
		resp, err := useCase.Execute(context.Background(), userID, req, "admin-id")

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, firstName, resp.FirstName)
		mockRepo.AssertExpectations(t)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		// Arrange
		mockRepo := new(mocks.UserProfileRepository)
		useCase := NewUpdateUserUseCase(mockRepo)

		mockRepo.On("FindByID", mock.Anything, "user-999").Return(nil, nil) // Returns nil user, nil error for not found? Or error?
		// Looking at implementation: if existingUser == nil { return nil, errors.New("user not found") }

		// Act
		resp, err := useCase.Execute(context.Background(), "user-999", &dtos.UpdateUserRequest{}, "admin-id")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "user not found")
	})

	t.Run("EmailConflict", func(t *testing.T) {
		// Arrange
		mockRepo := new(mocks.UserProfileRepository)
		useCase := NewUpdateUserUseCase(mockRepo)

		stubs := userstest.NewUsersStubs()
		userID := "user-123"
		newEmail := "other@example.com"
		req := stubs.UpdateRequestEmail(newEmail)

		existingUser := stubs.UserForUpdate(userID, "old@example.com", "John")
		otherUser := stubs.UserForUpdate("user-456", newEmail, "Other")

		mockRepo.On("FindByID", mock.Anything, userID).Return(existingUser, nil)
		mockRepo.On("FindByEmail", mock.Anything, newEmail).Return(otherUser, nil)

		// Act
		resp, err := useCase.Execute(context.Background(), userID, req, "admin-id")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "email is already in use")
	})
}
