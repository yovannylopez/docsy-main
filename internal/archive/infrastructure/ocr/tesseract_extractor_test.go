package ocr

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sharedconfig "github.com/yovannylopez/docsy-main/internal/shared/infrastructure/config"
)

func TestTesseractExtractor_Available_Disabled(t *testing.T) {
	e := NewTesseractExtractor(sharedconfig.OCRConfig{Enabled: false, TesseractBin: "tesseract"})
	assert.False(t, e.Available(context.Background()))
}

func TestTesseractExtractor_Available_MissingBinary(t *testing.T) {
	e := NewTesseractExtractor(sharedconfig.OCRConfig{
		Enabled:      true,
		TesseractBin: "tesseract-definitely-not-installed-xyz",
		Timeout:      5 * time.Second,
	})
	assert.False(t, e.Available(context.Background()))
}

func TestTesseractExtractor_ExtractImage_SkipIfNoBinary(t *testing.T) {
	if testing.Short() {
		t.Skip("short")
	}
	if _, err := exec.LookPath("tesseract"); err != nil {
		t.Skip("tesseract not in PATH")
	}
	e := NewTesseractExtractor(sharedconfig.OCRConfig{
		Enabled:      true,
		TesseractBin: "tesseract",
		Lang:         "eng",
		Timeout:      20 * time.Second,
	})
	require.True(t, e.Available(context.Background()))

	// Minimal valid 1x1 PNG (black pixel).
	png := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x00, 0x00, 0x00, 0x00, 0x3a, 0x7e, 0x9b, 0x55, 0x00, 0x00, 0x00,
		0x0a, 0x49, 0x44, 0x41, 0x54, 0x08, 0xd7, 0x63, 0x60, 0x00, 0x00, 0x00,
		0x02, 0x00, 0x01, 0xe2, 0x21, 0xbc, 0x33, 0x00, 0x00, 0x00, 0x00, 0x49,
		0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
	}
	_, err := e.ExtractText(context.Background(), "dot.png", "image/png", png)
	// May return empty text; just ensure it does not panic / hard-fail on binary missing.
	assert.NoError(t, err)
}

func TestUsefulText(t *testing.T) {
	assert.False(t, usefulText("abc"))
	assert.True(t, usefulText("Este es un texto suficientemente largo para contar como capa de PDF util."))
}
