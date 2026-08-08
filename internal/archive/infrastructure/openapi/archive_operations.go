package openapi

import (
	"github.com/yovannylopez/docsy-main/pkg/openapi"
)

func setupGetMyWorkspaceOperation(generator *openapi.Generator) {
	spec := generator.GetSpec()
	if pathItem, exists := spec.Paths["/api/v1/archive/workspaces/me"]; exists && pathItem.Get != nil {
		op := pathItem.Get
		op.Summary = "Get personal workspace"
		op.Description = "Ensures and returns the authenticated user's personal archive workspace"
		op.OperationID = "getMyArchiveWorkspace"
		op.Tags = []string{"archive"}
		op.Security = []map[string][]string{{"BearerAuth": {}}}
		op.Responses = map[string]openapi.Response{
			"200": {
				Description: "Workspace retrieved successfully",
				Content: map[string]openapi.MediaType{
					"application/json": {Schema: &openapi.Schema{Ref: "#/components/schemas/WorkspaceResponse"}},
				},
			},
			"401": {Description: "Unauthorized", Content: openapi.JSONErrorRefContent(nil)},
			"500": {Description: "Internal server error", Content: openapi.JSONErrorRefContent(nil)},
		}
		spec.Paths["/api/v1/archive/workspaces/me"] = pathItem
	}
}

func setupListWorkspacesOperation(generator *openapi.Generator) {
	spec := generator.GetSpec()
	if pathItem, exists := spec.Paths["/api/v1/archive/workspaces"]; exists && pathItem.Get != nil {
		op := pathItem.Get
		op.Summary = "List workspaces"
		op.Description = "Lists all workspaces the authenticated user belongs to"
		op.OperationID = "listArchiveWorkspaces"
		op.Tags = []string{"archive"}
		op.Security = []map[string][]string{{"BearerAuth": {}}}
		op.Responses = map[string]openapi.Response{
			"200": {Description: "Workspaces retrieved successfully"},
			"401": {Description: "Unauthorized", Content: openapi.JSONErrorRefContent(nil)},
		}
		spec.Paths["/api/v1/archive/workspaces"] = pathItem
	}
}

func setupCreateHouseholdOperation(generator *openapi.Generator) {
	spec := generator.GetSpec()
	if pathItem, exists := spec.Paths["/api/v1/archive/workspaces/household"]; exists && pathItem.Post != nil {
		op := pathItem.Post
		op.Summary = "Create household workspace"
		op.OperationID = "createArchiveHousehold"
		op.Tags = []string{"archive"}
		op.Security = []map[string][]string{{"BearerAuth": {}}}
		op.RequestBody = &openapi.RequestBody{
			Required: true,
			Content: map[string]openapi.MediaType{
				"application/json": {Schema: &openapi.Schema{Ref: "#/components/schemas/CreateHouseholdRequest"}},
			},
		}
		op.Responses = map[string]openapi.Response{
			"201": {
				Description: "Household created",
				Content: map[string]openapi.MediaType{
					"application/json": {Schema: &openapi.Schema{Ref: "#/components/schemas/WorkspaceResponse"}},
				},
			},
			"400": {Description: "Invalid input", Content: openapi.JSONErrorRefContent(nil)},
			"401": {Description: "Unauthorized", Content: openapi.JSONErrorRefContent(nil)},
		}
		spec.Paths["/api/v1/archive/workspaces/household"] = pathItem
	}
}

func setupListMembersOperation(generator *openapi.Generator) {
	spec := generator.GetSpec()
	if pathItem, exists := spec.Paths["/api/v1/archive/workspaces/{id}/members"]; exists && pathItem.Get != nil {
		op := pathItem.Get
		op.Summary = "List workspace members"
		op.OperationID = "listArchiveWorkspaceMembers"
		op.Tags = []string{"archive"}
		op.Security = []map[string][]string{{"BearerAuth": {}}}
		op.Responses = map[string]openapi.Response{
			"200": {Description: "Members retrieved successfully"},
			"403": {Description: "Forbidden", Content: openapi.JSONErrorRefContent(nil)},
			"404": {Description: "Not found", Content: openapi.JSONErrorRefContent(nil)},
			"401": {Description: "Unauthorized", Content: openapi.JSONErrorRefContent(nil)},
		}
		spec.Paths["/api/v1/archive/workspaces/{id}/members"] = pathItem
	}
}

func setupInviteMemberOperation(generator *openapi.Generator) {
	setupArchiveMemberBodyOperation(
		generator,
		"/api/v1/archive/workspaces/{id}/members",
		"post",
		"Invite workspace member",
		"inviteArchiveWorkspaceMember",
		"#/components/schemas/InviteMemberRequest",
		"201",
		"Member invited",
	)
}

func setupUpdateMemberRoleOperation(generator *openapi.Generator) {
	setupArchiveMemberBodyOperation(
		generator,
		"/api/v1/archive/workspaces/{id}/members/{userId}",
		"patch",
		"Update member role",
		"updateArchiveMemberRole",
		"#/components/schemas/UpdateMemberRoleRequest",
		"200",
		"Role updated",
	)
}

func setupArchiveMemberBodyOperation(
	generator *openapi.Generator,
	path, method, summary, operationID, schemaRef, successCode, successDesc string,
) {
	spec := generator.GetSpec()
	pathItem, exists := spec.Paths[path]
	if !exists {
		return
	}
	var op **openapi.Operation
	switch method {
	case "post":
		if pathItem.Post == nil {
			return
		}
		op = &pathItem.Post
	case "patch":
		if pathItem.Patch == nil {
			return
		}
		op = &pathItem.Patch
	default:
		return
	}
	(*op).Summary = summary
	(*op).OperationID = operationID
	(*op).Tags = []string{"archive"}
	(*op).Security = []map[string][]string{{"BearerAuth": {}}}
	(*op).RequestBody = &openapi.RequestBody{
		Required: true,
		Content: map[string]openapi.MediaType{
			"application/json": {Schema: &openapi.Schema{Ref: schemaRef}},
		},
	}
	(*op).Responses = map[string]openapi.Response{
		successCode: {Description: successDesc},
		"400":       {Description: "Invalid input", Content: openapi.JSONErrorRefContent(nil)},
		"403":       {Description: "Forbidden", Content: openapi.JSONErrorRefContent(nil)},
		"404":       {Description: "Not found", Content: openapi.JSONErrorRefContent(nil)},
		"401":       {Description: "Unauthorized", Content: openapi.JSONErrorRefContent(nil)},
	}
	spec.Paths[path] = pathItem
}

func setupRemoveMemberOperation(generator *openapi.Generator) {
	spec := generator.GetSpec()
	if pathItem, exists := spec.Paths["/api/v1/archive/workspaces/{id}/members/{userId}"]; exists && pathItem.Delete != nil {
		op := pathItem.Delete
		op.Summary = "Remove workspace member"
		op.OperationID = "removeArchiveWorkspaceMember"
		op.Tags = []string{"archive"}
		op.Security = []map[string][]string{{"BearerAuth": {}}}
		op.Responses = map[string]openapi.Response{
			"200": {Description: "Member removed"},
			"403": {Description: "Forbidden", Content: openapi.JSONErrorRefContent(nil)},
			"404": {Description: "Not found", Content: openapi.JSONErrorRefContent(nil)},
			"401": {Description: "Unauthorized", Content: openapi.JSONErrorRefContent(nil)},
		}
		spec.Paths["/api/v1/archive/workspaces/{id}/members/{userId}"] = pathItem
	}
}

func setupListCategoriesOperation(generator *openapi.Generator) {
	spec := generator.GetSpec()
	if pathItem, exists := spec.Paths["/api/v1/archive/categories"]; exists && pathItem.Get != nil {
		op := pathItem.Get
		op.Summary = "List document categories"
		op.Description = "Returns active document categories for the archive module"
		op.OperationID = "listArchiveCategories"
		op.Tags = []string{"archive"}
		op.Security = []map[string][]string{{"BearerAuth": {}}}
		op.Responses = map[string]openapi.Response{
			"200": {Description: "Categories retrieved successfully"},
			"401": {Description: "Unauthorized", Content: openapi.JSONErrorRefContent(nil)},
		}
		spec.Paths["/api/v1/archive/categories"] = pathItem
	}
}

func setupListDocumentsOperation(generator *openapi.Generator) {
	spec := generator.GetSpec()
	if pathItem, exists := spec.Paths["/api/v1/archive/documents"]; exists && pathItem.Get != nil {
		op := pathItem.Get
		op.Summary = "List documents"
		op.Description = "Lists documents in the caller's personal workspace with optional filters"
		op.OperationID = "listArchiveDocuments"
		op.Tags = []string{"archive"}
		op.Security = []map[string][]string{{"BearerAuth": {}}}
		op.Parameters = []openapi.Parameter{
			{Name: "workspace_id", In: "query", Schema: &openapi.Schema{Type: "string"}},
			{Name: "category", In: "query", Schema: &openapi.Schema{Type: "string"}},
			{Name: "q", In: "query", Schema: &openapi.Schema{Type: "string"}},
			{Name: "from", In: "query", Schema: &openapi.Schema{Type: "string", Format: "date"}},
			{Name: "to", In: "query", Schema: &openapi.Schema{Type: "string", Format: "date"}},
			{Name: "due_before", In: "query", Schema: &openapi.Schema{Type: "string", Format: "date"}},
			{Name: "status", In: "query", Schema: &openapi.Schema{Type: "string"}},
			{Name: "limit", In: "query", Schema: &openapi.Schema{Type: "integer"}},
			{Name: "offset", In: "query", Schema: &openapi.Schema{Type: "integer"}},
		}
		op.Responses = map[string]openapi.Response{
			"200": {Description: "Documents retrieved successfully"},
			"400": {Description: "Invalid query", Content: openapi.JSONErrorRefContent(nil)},
			"401": {Description: "Unauthorized", Content: openapi.JSONErrorRefContent(nil)},
		}
		spec.Paths["/api/v1/archive/documents"] = pathItem
	}
}

func setupCreateDocumentOperation(generator *openapi.Generator) {
	spec := generator.GetSpec()
	if pathItem, exists := spec.Paths["/api/v1/archive/documents"]; exists && pathItem.Post != nil {
		op := pathItem.Post
		op.Summary = "Create document"
		op.Description = "Creates a document metadata record in the personal workspace"
		op.OperationID = "createArchiveDocument"
		op.Tags = []string{"archive"}
		op.Security = []map[string][]string{{"BearerAuth": {}}}
		op.RequestBody = &openapi.RequestBody{
			Required: true,
			Content: map[string]openapi.MediaType{
				"application/json": {Schema: &openapi.Schema{Ref: "#/components/schemas/CreateDocumentRequest"}},
			},
		}
		op.Responses = map[string]openapi.Response{
			"201": {
				Description: "Document created",
				Content: map[string]openapi.MediaType{
					"application/json": {Schema: &openapi.Schema{Ref: "#/components/schemas/DocumentResponse"}},
				},
			},
			"400": {Description: "Invalid input", Content: openapi.JSONErrorRefContent(nil)},
			"401": {Description: "Unauthorized", Content: openapi.JSONErrorRefContent(nil)},
		}
		spec.Paths["/api/v1/archive/documents"] = pathItem
	}
}

func setupGetDocumentOperation(generator *openapi.Generator) {
	spec := generator.GetSpec()
	if pathItem, exists := spec.Paths["/api/v1/archive/documents/{id}"]; exists && pathItem.Get != nil {
		op := pathItem.Get
		op.Summary = "Get document"
		op.OperationID = "getArchiveDocument"
		op.Tags = []string{"archive"}
		op.Security = []map[string][]string{{"BearerAuth": {}}}
		op.Responses = map[string]openapi.Response{
			"200": {
				Description: "Document retrieved",
				Content: map[string]openapi.MediaType{
					"application/json": {Schema: &openapi.Schema{Ref: "#/components/schemas/DocumentResponse"}},
				},
			},
			"404": {Description: "Not found", Content: openapi.JSONErrorRefContent(nil)},
			"401": {Description: "Unauthorized", Content: openapi.JSONErrorRefContent(nil)},
		}
		spec.Paths["/api/v1/archive/documents/{id}"] = pathItem
	}
}

func setupUpdateDocumentOperation(generator *openapi.Generator) {
	spec := generator.GetSpec()
	if pathItem, exists := spec.Paths["/api/v1/archive/documents/{id}"]; exists && pathItem.Patch != nil {
		op := pathItem.Patch
		op.Summary = "Update document"
		op.OperationID = "updateArchiveDocument"
		op.Tags = []string{"archive"}
		op.Security = []map[string][]string{{"BearerAuth": {}}}
		op.RequestBody = &openapi.RequestBody{
			Required: true,
			Content: map[string]openapi.MediaType{
				"application/json": {Schema: &openapi.Schema{Ref: "#/components/schemas/UpdateDocumentRequest"}},
			},
		}
		op.Responses = map[string]openapi.Response{
			"200": {
				Description: "Document updated",
				Content: map[string]openapi.MediaType{
					"application/json": {Schema: &openapi.Schema{Ref: "#/components/schemas/DocumentResponse"}},
				},
			},
			"400": {Description: "Invalid input", Content: openapi.JSONErrorRefContent(nil)},
			"404": {Description: "Not found", Content: openapi.JSONErrorRefContent(nil)},
			"401": {Description: "Unauthorized", Content: openapi.JSONErrorRefContent(nil)},
		}
		spec.Paths["/api/v1/archive/documents/{id}"] = pathItem
	}
}

func setupArchiveDocumentOperation(generator *openapi.Generator) {
	spec := generator.GetSpec()
	if pathItem, exists := spec.Paths["/api/v1/archive/documents/{id}/archive"]; exists && pathItem.Post != nil {
		op := pathItem.Post
		op.Summary = "Archive document"
		op.Description = "Soft-archives a document (status=archived)"
		op.OperationID = "archiveArchiveDocument"
		op.Tags = []string{"archive"}
		op.Security = []map[string][]string{{"BearerAuth": {}}}
		op.Responses = map[string]openapi.Response{
			"200": {
				Description: "Document archived",
				Content: map[string]openapi.MediaType{
					"application/json": {Schema: &openapi.Schema{Ref: "#/components/schemas/DocumentResponse"}},
				},
			},
			"404": {Description: "Not found", Content: openapi.JSONErrorRefContent(nil)},
			"401": {Description: "Unauthorized", Content: openapi.JSONErrorRefContent(nil)},
		}
		spec.Paths["/api/v1/archive/documents/{id}/archive"] = pathItem
	}
}

func setupListDocumentFilesOperation(generator *openapi.Generator) {
	spec := generator.GetSpec()
	if pathItem, exists := spec.Paths["/api/v1/archive/documents/{id}/files"]; exists && pathItem.Get != nil {
		op := pathItem.Get
		op.Summary = "List document files"
		op.OperationID = "listArchiveDocumentFiles"
		op.Tags = []string{"archive"}
		op.Security = []map[string][]string{{"BearerAuth": {}}}
		op.Responses = map[string]openapi.Response{
			"200": {Description: "Files retrieved"},
			"404": {Description: "Document not found", Content: openapi.JSONErrorRefContent(nil)},
			"401": {Description: "Unauthorized", Content: openapi.JSONErrorRefContent(nil)},
		}
		spec.Paths["/api/v1/archive/documents/{id}/files"] = pathItem
	}
}

func setupUploadDocumentFileOperation(generator *openapi.Generator) {
	spec := generator.GetSpec()
	if pathItem, exists := spec.Paths["/api/v1/archive/documents/{id}/files"]; exists && pathItem.Post != nil {
		op := pathItem.Post
		op.Summary = "Upload document file"
		op.Description = "Uploads an attachment (multipart field name: file). Allowed: PDF, JPG/JPEG, PNG, TIFF, WebP, GIF, DOC/DOCX, XLS/XLSX"
		op.OperationID = "uploadArchiveDocumentFile"
		op.Tags = []string{"archive"}
		op.Security = []map[string][]string{{"BearerAuth": {}}}
		op.RequestBody = &openapi.RequestBody{
			Required: true,
			Content: map[string]openapi.MediaType{
				"multipart/form-data": {
					Schema: &openapi.Schema{
						Type: "object",
						Properties: map[string]*openapi.Schema{
							"file": {Type: "string", Format: "binary"},
						},
						Required: []string{"file"},
					},
				},
			},
		}
		op.Responses = map[string]openapi.Response{
			"201": {
				Description: "File uploaded",
				Content: map[string]openapi.MediaType{
					"application/json": {Schema: &openapi.Schema{Ref: "#/components/schemas/DocumentFileResponse"}},
				},
			},
			"400": {Description: "Invalid file", Content: openapi.JSONErrorRefContent(nil)},
			"404": {Description: "Document not found", Content: openapi.JSONErrorRefContent(nil)},
			"401": {Description: "Unauthorized", Content: openapi.JSONErrorRefContent(nil)},
		}
		spec.Paths["/api/v1/archive/documents/{id}/files"] = pathItem
	}
}

func setupDownloadDocumentFileOperation(generator *openapi.Generator) {
	spec := generator.GetSpec()
	if pathItem, exists := spec.Paths["/api/v1/archive/documents/{id}/files/{fileId}"]; exists && pathItem.Get != nil {
		op := pathItem.Get
		op.Summary = "Download document file"
		op.OperationID = "downloadArchiveDocumentFile"
		op.Tags = []string{"archive"}
		op.Security = []map[string][]string{{"BearerAuth": {}}}
		op.Responses = map[string]openapi.Response{
			"200": {Description: "Binary file content"},
			"404": {Description: "Not found", Content: openapi.JSONErrorRefContent(nil)},
			"401": {Description: "Unauthorized", Content: openapi.JSONErrorRefContent(nil)},
		}
		spec.Paths["/api/v1/archive/documents/{id}/files/{fileId}"] = pathItem
	}
}

func setupDeleteDocumentFileOperation(generator *openapi.Generator) {
	spec := generator.GetSpec()
	if pathItem, exists := spec.Paths["/api/v1/archive/documents/{id}/files/{fileId}"]; exists && pathItem.Delete != nil {
		op := pathItem.Delete
		op.Summary = "Delete document file"
		op.OperationID = "deleteArchiveDocumentFile"
		op.Tags = []string{"archive"}
		op.Security = []map[string][]string{{"BearerAuth": {}}}
		op.Responses = map[string]openapi.Response{
			"200": {Description: "File deleted"},
			"404": {Description: "Not found", Content: openapi.JSONErrorRefContent(nil)},
			"401": {Description: "Unauthorized", Content: openapi.JSONErrorRefContent(nil)},
		}
		spec.Paths["/api/v1/archive/documents/{id}/files/{fileId}"] = pathItem
	}
}
