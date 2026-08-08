package repositories

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/yovannylopez/docsy-main/internal/archive/domain/entities"
)

// extraFieldsJSONB serializes ExtraFields for PostgreSQL JSONB.
type extraFieldsJSONB entities.ExtraFields

// Value implements driver.Valuer.
func (e extraFieldsJSONB) Value() (driver.Value, error) {
	if len(e) == 0 {
		return nil, nil
	}
	b, err := json.Marshal([]entities.ExtraField(e))
	if err != nil {
		return nil, fmt.Errorf("marshal extra_fields: %w", err)
	}
	return b, nil
}

// Scan implements sql.Scanner.
func (e *extraFieldsJSONB) Scan(value any) error {
	if value == nil {
		*e = nil
		return nil
	}
	var raw []byte
	switch v := value.(type) {
	case []byte:
		raw = v
	case string:
		raw = []byte(v)
	default:
		return fmt.Errorf("extra_fields: unsupported Scan type %T", value)
	}
	if len(raw) == 0 || string(raw) == "null" {
		*e = nil
		return nil
	}
	var fields []entities.ExtraField
	if err := json.Unmarshal(raw, &fields); err != nil {
		return fmt.Errorf("unmarshal extra_fields: %w", err)
	}
	*e = extraFieldsJSONB(fields)
	return nil
}
