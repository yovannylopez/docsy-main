package entities

import (
	authEntities "github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
)

// User re-exports the User entity from the auth context (consolidation Option A)
type User = authEntities.User

// Role re-exports the Role entity from the auth context (consolidation Option A)
type Role = authEntities.Role

// PasswordHistory re-exports the PasswordHistory entity from the auth context
type PasswordHistory = authEntities.PasswordHistory

// VerificationToken re-exports the VerificationToken entity from the auth context
type VerificationToken = authEntities.VerificationToken
