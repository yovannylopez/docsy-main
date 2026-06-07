package test_utils

import (
	"time"

	"github.com/yovannylopez/docsy-main/internal/auth/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
)

// CloneLoginRequest returns a shallow copy of r (safe for string-only DTO).
func CloneLoginRequest(r *dtos.LoginRequest) *dtos.LoginRequest {
	if r == nil {
		return nil
	}
	c := *r
	return &c
}

// CloneUser returns a deep copy of u (independent Roles slice and string pointers).
func CloneUser(u *entities.User) *entities.User {
	if u == nil {
		return nil
	}
	out := *u
	out.Username = cloneStringPtr(u.Username)
	out.IdentificationNumber = cloneStringPtr(u.IdentificationNumber)
	out.IdentificationType = cloneStringPtr(u.IdentificationType)
	out.Phone = cloneStringPtr(u.Phone)
	out.LastLoginAt = cloneTimePtr(u.LastLoginAt)
	out.LastFailedLoginAt = cloneTimePtr(u.LastFailedLoginAt)
	out.LockedUntil = cloneTimePtr(u.LockedUntil)
	out.MFASecret = cloneStringPtr(u.MFASecret)
	out.CreatedBy = cloneStringPtr(u.CreatedBy)
	out.UpdatedBy = cloneStringPtr(u.UpdatedBy)
	if u.Roles != nil {
		out.Roles = make([]entities.Role, len(u.Roles))
		for i := range u.Roles {
			out.Roles[i] = CloneRole(&u.Roles[i])
		}
	}
	if u.PermissionNames != nil {
		out.PermissionNames = append([]string(nil), u.PermissionNames...)
	}
	return &out
}

// CloneRole returns a deep copy of r.
func CloneRole(r *entities.Role) entities.Role {
	if r == nil {
		return entities.Role{}
	}
	out := *r
	out.Description = cloneStringPtr(r.Description)
	return out
}

// CloneSession returns a deep copy of s.
func CloneSession(s *entities.Session) *entities.Session {
	if s == nil {
		return nil
	}
	out := *s
	out.AccessTokenJTI = cloneStringPtr(s.AccessTokenJTI)
	out.UserAgent = cloneStringPtr(s.UserAgent)
	out.IPAddress = cloneStringPtr(s.IPAddress)
	out.Location = cloneStringPtr(s.Location)
	out.DeviceFingerprint = cloneStringPtr(s.DeviceFingerprint)
	out.RevokedAt = cloneTimePtr(s.RevokedAt)
	out.RevokedReason = cloneStringPtr(s.RevokedReason)
	return &out
}

// CloneAuthToken returns a deep copy of t.
func CloneAuthToken(t *entities.AuthToken) *entities.AuthToken {
	if t == nil {
		return nil
	}
	out := *t
	return &out
}

// CloneAuditLog returns a deep copy of a (independent pointers, maps, and slices).
func CloneAuditLog(a *entities.AuditLog) *entities.AuditLog {
	if a == nil {
		return nil
	}
	out := *a
	out.UserID = cloneStringPtr(a.UserID)
	out.SessionID = cloneStringPtr(a.SessionID)
	out.Resource = cloneStringPtr(a.Resource)
	out.ResourceID = cloneStringPtr(a.ResourceID)
	out.Message = cloneStringPtr(a.Message)
	out.IPAddress = cloneStringPtr(a.IPAddress)
	out.UserAgent = cloneStringPtr(a.UserAgent)
	out.RequestID = cloneStringPtr(a.RequestID)
	if a.PreviousData != nil {
		m := mapsCloneShallow(*a.PreviousData)
		out.PreviousData = &m
	}
	if a.NewData != nil {
		m := mapsCloneShallow(*a.NewData)
		out.NewData = &m
	}
	if a.ChangedFields != nil {
		out.ChangedFields = append([]string(nil), a.ChangedFields...)
	}
	return &out
}

func mapsCloneShallow(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func cloneStringPtr(p *string) *string {
	if p == nil {
		return nil
	}
	v := *p
	return &v
}

func cloneTimePtr(p *time.Time) *time.Time {
	if p == nil {
		return nil
	}
	v := *p
	return &v
}
