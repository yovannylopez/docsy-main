package ports

import (
	"context"

	"github.com/yovannylopez/docsy-main/internal/archive/domain/dtos"
)

// OCRTextExtractor extracts plain text from a document file (image or PDF).
type OCRTextExtractor interface {
	// Available reports whether the OCR engine can run (enabled + binary present).
	Available(ctx context.Context) bool
	ExtractText(ctx context.Context, filename, contentType string, data []byte) (string, error)
}

// SuggestDocumentFieldsService suggests document metadata from an uploaded file via OCR.
type SuggestDocumentFieldsService interface {
	Execute(
		ctx context.Context,
		userID string,
		originalName, contentType string,
		data []byte,
	) (*dtos.OCRSuggestionResponse, error)
}
