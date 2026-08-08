package dtos

// OCRSuggestionResponse holds ephemeral field suggestions from OCR (SDD 012).
// Dates are YYYY-MM-DD for HTML date inputs; Amount is a decimal pesos string for the form.
type OCRSuggestionResponse struct {
	Title           string          `json:"title,omitempty"`
	Issuer          string          `json:"issuer,omitempty"`
	DocumentDate    string          `json:"document_date,omitempty"`
	DueDate         string          `json:"due_date,omitempty"`
	Amount          string          `json:"amount,omitempty"`
	AmountCents     *int64          `json:"amount_cents,omitempty"`
	Currency        string          `json:"currency,omitempty"`
	ReferenceNumber string          `json:"reference_number,omitempty"`
	Notes           string          `json:"notes,omitempty"`
	ExtraFields     []ExtraFieldDTO `json:"extra_fields,omitempty"`
	RawExcerpt      string          `json:"raw_excerpt,omitempty"`
	Confidence      float64         `json:"confidence,omitempty"`
}
