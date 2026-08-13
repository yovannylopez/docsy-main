package entities

import "time"

// Document status values.
const (
	DocumentStatusActive   = "active"
	DocumentStatusArchived = "archived"
	DocumentStatusAll      = "all" // list filter: include active and archived
)

// DefaultDocumentCurrency is the default ISO currency code for amounts.
const DefaultDocumentCurrency = "COP"

// Document holds archive document metadata (binaries come in iteration C).
type Document struct {
	ID              string      `json:"id" db:"id"`
	WorkspaceID     string      `json:"workspace_id" db:"workspace_id"`
	CategoryCode    string      `json:"category_code" db:"category_code"`
	Title           string      `json:"title" db:"title"`
	DocumentDate    *time.Time  `json:"document_date,omitempty" db:"document_date"`
	DueDate         *time.Time  `json:"due_date,omitempty" db:"due_date"`
	Issuer          *string     `json:"issuer,omitempty" db:"issuer"`
	ReferenceNumber *string     `json:"reference_number,omitempty" db:"reference_number"`
	AmountCents     *int64      `json:"amount_cents,omitempty" db:"amount_cents"`
	Currency        string      `json:"currency" db:"currency"`
	Notes           *string     `json:"notes,omitempty" db:"notes"`
	ExtraFields     ExtraFields `json:"extra_fields,omitempty" db:"extra_fields"`
	Status          string      `json:"status" db:"status"`
	CreatedBy       *string     `json:"created_by,omitempty" db:"created_by"`
	UpdatedBy       *string     `json:"updated_by,omitempty" db:"updated_by"`
	CreatedAt       time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at" db:"updated_at"`
}
