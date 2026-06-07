package types

import (
	"database/sql/driver"
	"encoding/json"
)

// JSONB maneja campos JSONB de PostgreSQL en la capa de infraestructura
type JSONB map[string]any

// Value implementa driver.Valuer para JSONB
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implementa sql.Scanner para JSONB
func (j *JSONB) Scan(value any) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte(value.(string)), j)
	}
	return json.Unmarshal(bytes, j)
}
