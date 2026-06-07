package repositories

import (
	"time"

	"github.com/yovannylopez/docsy-main/internal/auth/domain"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	authtest "github.com/yovannylopez/docsy-main/internal/auth/test_utils"
)

// repoUserForCreate builds a user for sqlmock Create paths (Object Mother: EmptyUser + clone + deltas).
func repoUserForCreate(now time.Time) *entities.User {
	stubs := authtest.NewAuthStubs()
	u := authtest.CloneUser(stubs.Entities.EmptyUser)
	u.ID = "u1"
	u.Email = "a@b.co"
	u.PasswordHash = "h"
	u.FirstName = "F"
	u.LastName = "L"
	u.IsActive = true
	u.IsVerified = false
	u.PasswordChangedAt = now
	u.MustChangePassword = false
	u.CreatedAt = now
	u.UpdatedAt = now
	return u
}

// repoUserForUpdateSuccess builds a user for Update success sqlmock.
func repoUserForUpdateSuccess(now time.Time) *entities.User {
	stubs := authtest.NewAuthStubs()
	u := authtest.CloneUser(stubs.Entities.EmptyUser)
	u.ID = "u1"
	u.Email = "n@x.co"
	u.FirstName = "A"
	u.LastName = "B"
	u.IsActive = true
	u.IsVerified = true
	u.MFAEnabled = false
	u.UpdatedAt = now
	u.UpdatedBy = nil
	return u
}

// repoUserForUpdateNoRows builds a user for Update when no rows affected.
func repoUserForUpdateNoRows(now time.Time) *entities.User {
	stubs := authtest.NewAuthStubs()
	u := authtest.CloneUser(stubs.Entities.EmptyUser)
	u.ID = "ghost"
	u.Email = "x@y.z"
	u.FirstName = "A"
	u.LastName = "B"
	u.IsActive = true
	u.IsVerified = false
	u.MFAEnabled = false
	u.UpdatedAt = now
	return u
}

// repoSessionForCreate builds a session for sqlmock Create/Update paths (clone EmptySession + deltas).
func repoSessionForCreate(id, userID, refreshHash string, now time.Time) *entities.Session {
	stubs := authtest.NewAuthStubs()
	s := authtest.CloneSession(stubs.Entities.EmptySession)
	s.ID = id
	s.UserID = userID
	s.RefreshTokenHash = refreshHash
	s.CreatedAt = now
	s.LastUsedAt = now
	s.ExpiresAt = now
	s.IsActive = true
	return s
}

// repoAuditLogForInsert builds an audit log for LogAction sqlmock (clone EmptyAuditLog + deltas).
func repoAuditLogForInsert(id, action, result string, createdAt time.Time, changedFields []string) *entities.AuditLog {
	stubs := authtest.NewAuthStubs()
	a := authtest.CloneAuditLog(stubs.Entities.EmptyAuditLog)
	a.ID = id
	a.Action = action
	a.Result = result
	a.CreatedAt = createdAt
	if changedFields != nil {
		a.ChangedFields = append([]string(nil), changedFields...)
	}
	return a
}

// repoAuditLogWithJSONMaps sets PreviousData/NewData on a cloned empty audit row (maps owned by caller).
func repoAuditLogWithJSONMaps(id string, createdAt time.Time, previousData, newData map[string]any) *entities.AuditLog {
	stubs := authtest.NewAuthStubs()
	a := authtest.CloneAuditLog(stubs.Entities.EmptyAuditLog)
	a.ID = id
	a.Action = "update"
	a.Result = domain.AuditResultSuccess
	a.CreatedAt = createdAt
	a.PreviousData = &previousData
	a.NewData = &newData
	a.ChangedFields = []string{}
	return a
}
