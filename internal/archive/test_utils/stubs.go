// Package test_utils provides Object Mother factories for archive bounded context tests.
package test_utils

import (
	"time"

	"github.com/yovannylopez/docsy-main/internal/archive/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/archive/domain/entities"
)

// ArchiveStubs holds factories for archive module tests.
type ArchiveStubs struct{}

// NewArchiveStubs returns the archive Object Mother.
func NewArchiveStubs() *ArchiveStubs {
	return &ArchiveStubs{}
}

// PersonalWorkspace returns a personal workspace owned by user-1.
func (s *ArchiveStubs) PersonalWorkspace() *entities.Workspace {
	now := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	return &entities.Workspace{
		ID:          "ws-1",
		Name:        "Mi archivo",
		Type:        entities.WorkspaceTypePersonal,
		OwnerUserID: "user-1",
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// HouseholdWorkspace returns a household workspace owned by user-1.
func (s *ArchiveStubs) HouseholdWorkspace() *entities.Workspace {
	ws := s.PersonalWorkspace()
	ws.ID = "ws-home-1"
	ws.Name = "Hogar López"
	ws.Type = entities.WorkspaceTypeHousehold
	return ws
}

// OwnerMember returns owner membership for ws-1 / user-1.
func (s *ArchiveStubs) OwnerMember() *entities.WorkspaceMember {
	return &entities.WorkspaceMember{
		ID:          "m-1",
		WorkspaceID: "ws-1",
		UserID:      "user-1",
		Role:        entities.WorkspaceRoleOwner,
		JoinedAt:    time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC),
	}
}

// ActiveDocument returns a sample active document in ws-1.
func (s *ArchiveStubs) ActiveDocument() *entities.Document {
	now := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	return &entities.Document{
		ID:           "d-1",
		WorkspaceID:  "ws-1",
		CategoryCode: "taxes",
		Title:        "Predial",
		Status:       entities.DocumentStatusActive,
		Currency:     entities.DefaultDocumentCurrency,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// CreateDocumentRequest returns a valid create request.
func (s *ArchiveStubs) CreateDocumentRequest() *dtos.CreateDocumentRequest {
	return &dtos.CreateDocumentRequest{
		CategoryCode: "taxes",
		Title:        "Predial",
	}
}

// InviteMemberRequest returns a valid invite request.
func (s *ArchiveStubs) InviteMemberRequest() *dtos.InviteMemberRequest {
	return &dtos.InviteMemberRequest{
		Email: "member@example.com",
		Role:  entities.WorkspaceRoleMember,
	}
}

// DocumentFile returns sample attachment metadata.
func (s *ArchiveStubs) DocumentFile() *entities.DocumentFile {
	return &entities.DocumentFile{
		ID:           "f-1",
		DocumentID:   "d-1",
		StorageKey:   "ws-1/d-1/f-1_predial.pdf",
		OriginalName: "predial.pdf",
		ContentType:  "application/pdf",
		SizeBytes:    128,
		UploadedAt:   time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC),
	}
}

// CloneWorkspace returns a shallow copy of a workspace.
func CloneWorkspace(w *entities.Workspace) *entities.Workspace {
	if w == nil {
		return nil
	}
	cp := *w
	return &cp
}

// CloneDocument returns a shallow copy of a document.
func CloneDocument(d *entities.Document) *entities.Document {
	if d == nil {
		return nil
	}
	cp := *d
	return &cp
}
