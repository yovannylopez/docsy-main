package dtos

import "time"

// ExtraFieldDTO is a curated metadata badge (key/label/value).
type ExtraFieldDTO struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Value string `json:"value"`
}

// DocumentResponse is the API/view DTO for a document.
type DocumentResponse struct {
	ID                  string          `json:"id"`
	WorkspaceID         string          `json:"workspace_id"`
	CategoryCode        string          `json:"category_code"`
	CategoryLabel       string          `json:"category_label,omitempty"`
	Title               string          `json:"title"`
	DocumentDate        *time.Time      `json:"document_date,omitempty"`
	DueDate             *time.Time      `json:"due_date,omitempty"`
	Issuer              *string         `json:"issuer,omitempty"`
	ReferenceNumber     *string         `json:"reference_number,omitempty"`
	AmountCents         *int64          `json:"amount_cents,omitempty"`
	Currency            string          `json:"currency"`
	Notes               *string         `json:"notes,omitempty"`
	ExtraFields         []ExtraFieldDTO `json:"extra_fields,omitempty"`
	Status              string          `json:"status"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
	PrimaryOriginalName string          `json:"primary_original_name,omitempty"`
	PrimaryContentType  string          `json:"primary_content_type,omitempty"`
}

// CreateDocumentRequest is the input to create a document.
type CreateDocumentRequest struct {
	WorkspaceID     string          `json:"workspace_id,omitempty"`
	CategoryCode    string          `json:"category_code"`
	Title           string          `json:"title"`
	DocumentDate    *time.Time      `json:"document_date,omitempty"`
	DueDate         *time.Time      `json:"due_date,omitempty"`
	Issuer          *string         `json:"issuer,omitempty"`
	ReferenceNumber *string         `json:"reference_number,omitempty"`
	AmountCents     *int64          `json:"amount_cents,omitempty"`
	Currency        string          `json:"currency,omitempty"`
	Notes           *string         `json:"notes,omitempty"`
	ExtraFields     []ExtraFieldDTO `json:"extra_fields,omitempty"`
}

// UpdateDocumentRequest is the input to partially update a document.
type UpdateDocumentRequest struct {
	WorkspaceID     string          `json:"workspace_id,omitempty"`
	CategoryCode    *string         `json:"category_code,omitempty"`
	Title           *string         `json:"title,omitempty"`
	DocumentDate    *time.Time      `json:"document_date,omitempty"`
	DueDate         *time.Time      `json:"due_date,omitempty"`
	ClearDueDate    bool            `json:"clear_due_date,omitempty"`
	Issuer          *string         `json:"issuer,omitempty"`
	ReferenceNumber *string         `json:"reference_number,omitempty"`
	AmountCents     *int64          `json:"amount_cents,omitempty"`
	ClearAmount     bool            `json:"clear_amount,omitempty"`
	Currency        *string         `json:"currency,omitempty"`
	Notes           *string         `json:"notes,omitempty"`
	ExtraFields     []ExtraFieldDTO `json:"extra_fields,omitempty"`
	SetExtraFields  bool            `json:"set_extra_fields,omitempty"`
}

// ListDocumentsFilter holds list filters.
type ListDocumentsFilter struct {
	WorkspaceID string
	Category    string
	Query       string
	From        *time.Time
	To          *time.Time
	DueBefore   *time.Time
	Status      string
	Limit       int
	Offset      int
}

// DocumentCategoryResponse is a category option for forms/API.
type DocumentCategoryResponse struct {
	Code      string `json:"code"`
	LabelES   string `json:"label_es"`
	SortOrder int    `json:"sort_order"`
}

// CategoryFolderResponse is a virtual folder (category) for the documents browser.
type CategoryFolderResponse struct {
	Code      string `json:"code"`
	LabelES   string `json:"label_es"`
	SortOrder int    `json:"sort_order"`
	Count     int    `json:"count"`
}
