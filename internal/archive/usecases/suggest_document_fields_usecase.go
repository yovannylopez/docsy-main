package usecases

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/yovannylopez/docsy-main/internal/archive/domain/dtos"
	domainerrors "github.com/yovannylopez/docsy-main/internal/archive/domain/errors"
	"github.com/yovannylopez/docsy-main/internal/archive/domain/ports"
	authdomain "github.com/yovannylopez/docsy-main/internal/auth/domain"
	authports "github.com/yovannylopez/docsy-main/internal/auth/domain/ports"
)

const auditResourceOCR = "ocr_suggestion"

// SuggestDocumentFieldsUseCase extracts text via OCR and parses metadata suggestions.
type SuggestDocumentFieldsUseCase struct {
	extractor ports.OCRTextExtractor
	maxBytes  int64
	auditRepo authports.AuditRepository
}

// NewSuggestDocumentFieldsUseCase creates the use case.
func NewSuggestDocumentFieldsUseCase(
	extractor ports.OCRTextExtractor,
	maxBytes int64,
	auditRepo authports.AuditRepository,
) *SuggestDocumentFieldsUseCase {
	return &SuggestDocumentFieldsUseCase{
		extractor: extractor,
		maxBytes:  maxBytes,
		auditRepo: auditRepo,
	}
}

// Execute validates the file, runs OCR and returns field suggestions.
func (uc *SuggestDocumentFieldsUseCase) Execute(
	ctx context.Context,
	userID string,
	originalName, contentType string,
	data []byte,
) (*dtos.OCRSuggestionResponse, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, domainerrors.ErrUserIDRequired
	}
	if uc.extractor == nil || !uc.extractor.Available(ctx) {
		return nil, domainerrors.ErrOCRUnavailable
	}
	if len(data) == 0 {
		return nil, domainerrors.ErrFileRequired
	}
	if uc.maxBytes > 0 && int64(len(data)) > uc.maxBytes {
		return nil, domainerrors.ErrFileTooLarge
	}

	name := sanitizeOriginalName(originalName)
	ct, err := normalizeContentType(contentType, name, data)
	if err != nil {
		return nil, err
	}
	if !isOCRSupportedContentType(ct) {
		return nil, domainerrors.ErrOCRUnsupportedType
	}

	raw, err := uc.extractor.ExtractText(ctx, name, ct, data)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domainerrors.ErrOCRFailed, err)
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, domainerrors.ErrOCRNoText
	}

	parsed := parseOCRFields(raw)
	resp := &dtos.OCRSuggestionResponse{
		Title:           parsed.Title,
		Issuer:          parsed.Issuer,
		DocumentDate:    parsed.DocumentDate,
		DueDate:         parsed.DueDate,
		Amount:          parsed.Amount,
		AmountCents:     parsed.AmountCents,
		Currency:        parsed.Currency,
		ReferenceNumber: parsed.ReferenceNumber,
		Notes:           parsed.Notes,
		ExtraFields:     parsed.ExtraFields,
		RawExcerpt:      parsed.RawExcerpt,
		Confidence:      parsed.Confidence,
	}

	logArchiveAction(
		ctx, uc.auditRepo, userID,
		authdomain.AuditActionArchiveOCRSuggested,
		auditResourceOCR, filepath.Ext(name),
		"Archive OCR suggestions generated successfully",
	)
	return resp, nil
}

func isOCRSupportedContentType(ct string) bool {
	switch ct {
	case mimeApplicationPDF,
		mimeImageJPEG, mimeImagePNG, mimeImageWebP, mimeImageGIF, mimeImageTIFF:
		return true
	default:
		return false
	}
}
