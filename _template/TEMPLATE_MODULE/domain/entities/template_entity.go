// Package entities defines the domain entities for the TEMPLATE_MODULE module.
// Replace TEMPLATE_MODULE with your module name (e.g. products, orders, invoices).
package entities

import "time"

// TemplateEntity represents the main domain entity for this module.
// Rename this to match your domain concept (e.g. Product, Order, Invoice).
type TemplateEntity struct {
	ID          string    `json:"id"          db:"id"`
	Name        string    `json:"name"        db:"name"`
	Description *string   `json:"description" db:"description"`
	IsActive    bool      `json:"is_active"   db:"is_active"`
	CreatedAt   time.Time `json:"created_at"  db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"  db:"updated_at"`
}
