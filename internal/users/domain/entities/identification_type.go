package entities

import (
	"fmt"
	"strings"
)

// ValidIdentificationTypes defines the valid identification types in the domain.
// Centralizes validation to avoid hardcoding in use cases and to facilitate
// region-based configuration in the future.
//
//nolint:gochecknoglobals
var ValidIdentificationTypes = map[string]bool{
	"cc":  true, // National identity card
	"ce":  true, // Foreign identity card
	"pa":  true, // Passport
	"nit": true, // Tax identification number
	"rut": true, // Single tax registry
}

// ErrInvalidIdentificationType is returned when the identification type is not valid
var ErrInvalidIdentificationType = fmt.Errorf("invalid identification type. Valid values: cc, ce, pa, nit, rut")

// ValidateIdentificationType validates that the identification type is valid.
// Accepts values in uppercase or lowercase.
// Returns nil if valid, ErrInvalidIdentificationType otherwise.
func ValidateIdentificationType(tipo *string) error {
	if tipo == nil || *tipo == "" {
		return nil
	}

	normalized := strings.ToLower(strings.TrimSpace(*tipo))
	if !ValidIdentificationTypes[normalized] {
		return ErrInvalidIdentificationType
	}

	return nil
}
