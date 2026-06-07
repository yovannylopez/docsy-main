package entities

import "time"

// AuditModel contains the standard audit fields for all configuration tables.
// This model should be embedded in structs that represent tables with automatic auditing.
//
// Usage:
//
//	type MyEntity struct {
//	    ID   string `json:"id" db:"id"`
//	    Name string `json:"name" db:"name"`
//	    entities.AuditModel
//	}
//
// The fields are compatible with sqlx and are automatically mapped to the
// created_at and updated_at columns in the database.
type AuditModel struct {
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
