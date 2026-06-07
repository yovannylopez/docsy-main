package test_utils

import (
	"time"

	"github.com/yovannylopez/docsy-main/internal/auth/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
)

// MinimalLoginResponse is a small login payload for transport-layer tests.
func MinimalLoginResponse(userID, email, accessToken string) *dtos.LoginResponse {
	return &dtos.LoginResponse{
		User:  &dtos.UserResponse{ID: userID, Email: email},
		Token: &dtos.TokenResponse{AccessToken: accessToken, TokenType: "Bearer"},
	}
}

// TransportAuditLogRow builds an audit log row for handler List tests from EmptyAuditLog (Object Mother + clone).
func TransportAuditLogRow(id, action, result string, resource, userID *string) entities.AuditLog {
	stubs := NewAuthStubs()
	a := CloneAuditLog(stubs.Entities.EmptyAuditLog)
	a.ID = id
	a.Action = action
	a.Result = result
	a.Resource = cloneStringPtr(resource)
	a.UserID = cloneStringPtr(userID)
	a.CreatedAt = time.Now()
	return *a
}
