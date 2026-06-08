package container

import (
	authPorts "github.com/yovannylopez/docsy-main/internal/auth/domain/ports"

	"github.com/yovannylopez/docsy-main/internal/users/domain/ports"
	usersRepositories "github.com/yovannylopez/docsy-main/internal/users/infrastructure/repositories"
	usersHandlers "github.com/yovannylopez/docsy-main/internal/users/transport/handlers"
	"github.com/yovannylopez/docsy-main/internal/users/usecases"
)

// UsersContainer manages the dependencies of the users module
type UsersContainer struct {
	userProfileRepo    ports.UserProfileRepository
	passwordHasher     ports.PasswordHasher
	getUsersUseCase    *usecases.GetUsersUseCase
	createUsersUseCase *usecases.CreateUsersUseCase
	updateUserUseCase  *usecases.UpdateUserUseCase
	searchUsersUseCase *usecases.SearchUsersUseCase
	getUserByIDUseCase *usecases.GetUserByIDUseCase
}

// NewUsersContainer creates a new instance of the users container.
// authUserRepo must be injected from the composition root (auth implementation).
func NewUsersContainer(
	authUserRepo authPorts.UserRepository,
	passwordHasher ports.PasswordHasher,
	auditRepo authPorts.AuditRepository,
) *UsersContainer {
	userProfileRepo := usersRepositories.NewUserProfileRepositoryAdapter(authUserRepo)

	return &UsersContainer{
		userProfileRepo:    userProfileRepo,
		passwordHasher:     passwordHasher,
		getUsersUseCase:    usecases.NewGetUsersUseCase(userProfileRepo),
		createUsersUseCase: usecases.NewCreateUsersUseCase(userProfileRepo, passwordHasher, auditRepo),
		updateUserUseCase:  usecases.NewUpdateUserUseCase(userProfileRepo),
		searchUsersUseCase: usecases.NewSearchUsersUseCase(userProfileRepo),
		getUserByIDUseCase: usecases.NewGetUserByIDUseCase(userProfileRepo),
	}
}

// GetUserProfileRepository returns the user profile repository
func (c *UsersContainer) GetUserProfileRepository() ports.UserProfileRepository {
	return c.userProfileRepo
}

// GetUserRepository returns the user repository (alias for compatibility)
func (c *UsersContainer) GetUserRepository() ports.UserProfileRepository {
	return c.userProfileRepo
}

// GetPasswordHasher returns the password hasher
func (c *UsersContainer) GetPasswordHasher() ports.PasswordHasher {
	return c.passwordHasher
}

// GetUsersHandler returns the users handler configured with injected use cases
func (c *UsersContainer) GetUsersHandler() *usersHandlers.UsersHandler {
	return usersHandlers.NewUsersHandler(
		c.getUsersUseCase,
		c.createUsersUseCase,
		c.updateUserUseCase,
		c.searchUsersUseCase,
		c.getUserByIDUseCase,
	)
}

// GetUsersPageHandler returns the server-rendered users page handler.
func (c *UsersContainer) GetUsersPageHandler() *usersHandlers.UsersPageHandler {
	return usersHandlers.NewUsersPageHandler(
		c.getUsersUseCase,
		c.createUsersUseCase,
		c.updateUserUseCase,
		c.searchUsersUseCase,
		c.getUserByIDUseCase,
	)
}
