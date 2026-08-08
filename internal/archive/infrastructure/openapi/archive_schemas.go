// Package openapi defines OpenAPI documentation for the archive API.
package openapi

import (
	"github.com/yovannylopez/docsy-main/internal/archive/domain/dtos"
	"github.com/yovannylopez/docsy-main/pkg/openapi"
)

// SetupArchiveSpec registers OpenAPI schemas and operations for archive.
func SetupArchiveSpec(generator *openapi.Generator) {
	schemaGen := openapi.NewSchemaGenerator(generator)

	schemaGen.GenerateSchemaFromStruct("WorkspaceResponse", dtos.WorkspaceResponse{})
	schemaGen.GenerateSchemaFromStruct("CreateHouseholdRequest", dtos.CreateHouseholdRequest{})
	schemaGen.GenerateSchemaFromStruct("InviteMemberRequest", dtos.InviteMemberRequest{})
	schemaGen.GenerateSchemaFromStruct("UpdateMemberRoleRequest", dtos.UpdateMemberRoleRequest{})
	schemaGen.GenerateSchemaFromStruct("WorkspaceMemberResponse", dtos.WorkspaceMemberResponse{})
	schemaGen.GenerateSchemaFromStruct("DocumentResponse", dtos.DocumentResponse{})
	schemaGen.GenerateSchemaFromStruct("CreateDocumentRequest", dtos.CreateDocumentRequest{})
	schemaGen.GenerateSchemaFromStruct("UpdateDocumentRequest", dtos.UpdateDocumentRequest{})
	schemaGen.GenerateSchemaFromStruct("DocumentCategoryResponse", dtos.DocumentCategoryResponse{})
	schemaGen.GenerateSchemaFromStruct("DocumentFileResponse", dtos.DocumentFileResponse{})

	setupGetMyWorkspaceOperation(generator)
	setupListWorkspacesOperation(generator)
	setupCreateHouseholdOperation(generator)
	setupListMembersOperation(generator)
	setupInviteMemberOperation(generator)
	setupUpdateMemberRoleOperation(generator)
	setupRemoveMemberOperation(generator)
	setupListCategoriesOperation(generator)
	setupListDocumentsOperation(generator)
	setupCreateDocumentOperation(generator)
	setupGetDocumentOperation(generator)
	setupUpdateDocumentOperation(generator)
	setupArchiveDocumentOperation(generator)
	setupListDocumentFilesOperation(generator)
	setupUploadDocumentFileOperation(generator)
	setupDownloadDocumentFileOperation(generator)
	setupDeleteDocumentFileOperation(generator)
}
