// Package test_utils provides Object Mother factories for users bounded context tests.
package test_utils

import (
	authtest "github.com/yovannylopez/docsy-main/internal/auth/test_utils"
	"github.com/yovannylopez/docsy-main/internal/users/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/users/domain/entities"
)

// UsersStubs holds factories for users module tests.
type UsersStubs struct {
	auth *authtest.AuthStubs
}

// NewUsersStubs returns a users test mother (reuses auth entity shapes via Clone).
func NewUsersStubs() *UsersStubs {
	return &UsersStubs{auth: authtest.NewAuthStubs()}
}

// UserJohnPerez returns a user matching legacy get-users success test fixtures.
func (s *UsersStubs) UserJohnPerez() *entities.User {
	u := authtest.CloneUser(s.auth.Entities.ValidUser)
	u.ID = "user-1"
	u.Email = "test1@example.com"
	u.FirstName = "John"
	u.LastName = "Perez"
	u.Phone = nil
	u.IsActive = true
	u.IsVerified = true
	u.LastLoginAt = nil
	u.Username = nil
	return u
}

// UserMariaGarcia returns a second list user for get-users success.
func (s *UsersStubs) UserMariaGarcia() *entities.User {
	u := authtest.CloneUser(s.auth.Entities.ValidUser)
	u.ID = "user-2"
	u.Email = "test2@example.com"
	u.FirstName = "Maria"
	u.LastName = "Garcia"
	p := "+1234567890"
	u.Phone = &p
	u.IsActive = true
	u.IsVerified = false
	u.LastLoginAt = nil
	u.Username = nil
	return u
}

// UserJohnDoeSearch returns a user for search success tests.
func (s *UsersStubs) UserJohnDoeSearch() *entities.User {
	u := authtest.CloneUser(s.auth.Entities.ValidUser)
	u.ID = "user-1"
	u.Email = "john@example.com"
	u.FirstName = "John"
	u.LastName = "Doe"
	u.Username = authtest.StringPtr("johndoe")
	u.Phone = nil
	u.IsActive = true
	u.IsVerified = true
	u.LastLoginAt = nil
	return u
}

// UserJaneSmithSearch returns a second user for search success tests.
func (s *UsersStubs) UserJaneSmithSearch() *entities.User {
	u := authtest.CloneUser(s.auth.Entities.ValidUser)
	u.ID = "user-2"
	u.Email = "jane@example.com"
	u.FirstName = "Jane"
	u.LastName = "Smith"
	u.Username = authtest.StringPtr("janesmith")
	p := "+1234567890"
	u.Phone = &p
	u.IsActive = true
	u.IsVerified = false
	u.LastLoginAt = nil
	return u
}

// UserTestGeneric returns a minimal active user used in several search tests.
func (s *UsersStubs) UserTestGeneric() *entities.User {
	u := authtest.CloneUser(s.auth.Entities.ValidUser)
	u.ID = "user-1"
	u.Email = "test@example.com"
	u.FirstName = "Test"
	u.LastName = "User"
	u.Username = nil
	u.Phone = nil
	u.IsActive = true
	u.IsVerified = true
	u.LastLoginAt = nil
	return u
}

// SearchRequestWithActivo builds SearchUsersRequest with activo filter.
func (s *UsersStubs) SearchRequestWithActivo(q string, activo bool, limit, offset int) *dtos.SearchUsersRequest {
	a := activo
	return &dtos.SearchUsersRequest{Q: q, Limit: limit, Offset: offset, Activo: &a}
}

// SearchRequestNoActivo builds SearchUsersRequest without activo filter.
func (s *UsersStubs) SearchRequestNoActivo(q string, limit, offset int) *dtos.SearchUsersRequest {
	return &dtos.SearchUsersRequest{Q: q, Limit: limit, Offset: offset, Activo: nil}
}

// CreateUsersRequestStandard matches create use case success test (full optional fields).
func (s *UsersStubs) CreateUsersRequestStandard() *dtos.CreateUsersRequest {
	idType := "CC"
	idNum := "123456"
	username := "jdoe"
	return &dtos.CreateUsersRequest{
		Users: []dtos.CreateUserRequest{
			{
				Email:                "test@example.com",
				Password:             "Password123!",
				FirstName:            "John",
				LastName:             "Doe",
				RoleName:             "admin",
				IdentificationType:   &idType,
				IdentificationNumber: &idNum,
				Username:             &username,
			},
		},
	}
}

// CreateUsersRequestMinimal matches duplicate-email test (no optional IDs).
func (s *UsersStubs) CreateUsersRequestMinimal() *dtos.CreateUsersRequest {
	return &dtos.CreateUsersRequest{
		Users: []dtos.CreateUserRequest{
			{
				Email:     "test@example.com",
				Password:  "Password123!",
				FirstName: "John",
				LastName:  "Doe",
				RoleName:  "admin",
			},
		},
	}
}

// UserExistingByEmail returns a minimal user for "already exists" scenarios.
func (s *UsersStubs) UserExistingByEmail(id, email string) *entities.User {
	u := authtest.CloneUser(s.auth.Entities.ValidUser)
	u.ID = id
	u.Email = email
	u.FirstName = "Existing"
	u.LastName = "User"
	return u
}

// UserForUpdate returns a base user for update use case tests.
func (s *UsersStubs) UserForUpdate(userID, email, firstName string) *entities.User {
	u := authtest.CloneUser(s.auth.Entities.ValidUser)
	u.ID = userID
	u.Email = email
	u.FirstName = firstName
	u.LastName = "User"
	return u
}

// UpdateRequestFirstName returns UpdateUserRequest with first name set.
func (s *UsersStubs) UpdateRequestFirstName(name string) *dtos.UpdateUserRequest {
	return &dtos.UpdateUserRequest{FirstName: &name}
}

// UpdateRequestEmail returns UpdateUserRequest with email set.
func (s *UsersStubs) UpdateRequestEmail(email string) *dtos.UpdateUserRequest {
	return &dtos.UpdateUserRequest{Email: &email}
}
