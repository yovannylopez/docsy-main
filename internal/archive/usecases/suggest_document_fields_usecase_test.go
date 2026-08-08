package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainerrors "github.com/yovannylopez/docsy-main/internal/archive/domain/errors"
)

type stubOCRExtractor struct {
	available bool
	text      string
	err       error
}

func (s stubOCRExtractor) Available(_ context.Context) bool { return s.available }

func (s stubOCRExtractor) ExtractText(_ context.Context, _, _ string, _ []byte) (string, error) {
	return s.text, s.err
}

func pngSample(extra int) []byte {
	base := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	out := make([]byte, 0, len(base)+extra)
	out = append(out, base...)
	if extra > 0 {
		out = append(out, make([]byte, extra)...)
	}
	return out
}

func TestSuggestDocumentFieldsUseCase_Success(t *testing.T) {
	uc := NewSuggestDocumentFieldsUseCase(stubOCRExtractor{
		available: true,
		text:      "FACTURA\nACME SAS\nFecha: 01/01/2026\nTotal a pagar $50.000\nFactura No. A-123",
	}, 1024*1024, nil)

	got, err := uc.Execute(context.Background(), "user-1", "factura.png", "image/png", pngSample(64))
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "2026-01-01", got.DocumentDate)
	assert.Equal(t, "50000", got.Amount)
	assert.Equal(t, "A-123", got.ReferenceNumber)
}

func TestSuggestDocumentFieldsUseCase_Unavailable(t *testing.T) {
	uc := NewSuggestDocumentFieldsUseCase(stubOCRExtractor{available: false}, 1024, nil)
	_, err := uc.Execute(context.Background(), "u1", "a.png", "image/png", pngSample(0))
	assert.ErrorIs(t, err, domainerrors.ErrOCRUnavailable)
}

func TestSuggestDocumentFieldsUseCase_UnsupportedOffice(t *testing.T) {
	hdr := []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}
	ole := make([]byte, len(hdr)+32)
	copy(ole, hdr)
	uc := NewSuggestDocumentFieldsUseCase(stubOCRExtractor{available: true, text: "x"}, 1024, nil)
	_, err := uc.Execute(context.Background(), "u1", "a.doc", "application/msword", ole)
	assert.ErrorIs(t, err, domainerrors.ErrOCRUnsupportedType)
}

func TestSuggestDocumentFieldsUseCase_NoText(t *testing.T) {
	uc := NewSuggestDocumentFieldsUseCase(stubOCRExtractor{available: true, text: "  "}, 1024, nil)
	_, err := uc.Execute(context.Background(), "u1", "a.png", "image/png", pngSample(16))
	assert.ErrorIs(t, err, domainerrors.ErrOCRNoText)
}

func TestSuggestDocumentFieldsUseCase_ExtractError(t *testing.T) {
	uc := NewSuggestDocumentFieldsUseCase(stubOCRExtractor{
		available: true,
		err:       errors.New("boom"),
	}, 1024, nil)
	_, err := uc.Execute(context.Background(), "u1", "a.png", "image/png", pngSample(16))
	assert.ErrorIs(t, err, domainerrors.ErrOCRFailed)
}
