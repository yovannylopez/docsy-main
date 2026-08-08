package ports

import "context"

// UserRef is a minimal user identity for archive invitations.
type UserRef struct {
	ID          string
	Email       string
	DisplayName string
}

// UserDirectory looks up users for workspace invitations (no auth coupling in domain).
type UserDirectory interface {
	FindByEmail(ctx context.Context, email string) (*UserRef, error)
}
