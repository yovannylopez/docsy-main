package entities

// ExtraField is a curated key/value metadata item (stored in JSONB).
type ExtraField struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Value string `json:"value"`
}

// ExtraFields is an ordered list of curated metadata fields.
type ExtraFields []ExtraField
