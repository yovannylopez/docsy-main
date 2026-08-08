package ocr

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	sharedconfig "github.com/yovannylopez/docsy-main/internal/shared/infrastructure/config"
)

const (
	minUsefulPDFTextLen   = 40
	minUsefulLetterCount  = 20
	defaultOCRTimeoutSecs = 30
	tempFileMode          = 0o600
	pdfRenderDPI          = "200"
	mimeApplicationPDF    = "application/pdf"
	defaultTesseractBin   = "tesseract"
	defaultPDFToTextBin   = "pdftotext"
	defaultPDFToPPMBin    = "pdftoppm"
	defaultOCRLang        = "spa+eng"
)

// TesseractExtractor runs Tesseract CLI (and Poppler helpers for PDF).
type TesseractExtractor struct {
	enabled      bool
	tesseractBin string
	pdftotextBin string
	pdftoppmBin  string
	lang         string
	timeout      time.Duration
}

// NewTesseractExtractor builds an extractor from OCR config.
func NewTesseractExtractor(cfg sharedconfig.OCRConfig) *TesseractExtractor {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultOCRTimeoutSecs * time.Second
	}
	lang := strings.TrimSpace(cfg.Lang)
	if lang == "" {
		lang = defaultOCRLang
	}
	tess := strings.TrimSpace(cfg.TesseractBin)
	if tess == "" {
		tess = defaultTesseractBin
	}
	pdftotext := strings.TrimSpace(cfg.PDFToTextBin)
	if pdftotext == "" {
		pdftotext = defaultPDFToTextBin
	}
	pdftoppm := strings.TrimSpace(cfg.PDFToPPMBin)
	if pdftoppm == "" {
		pdftoppm = defaultPDFToPPMBin
	}
	return &TesseractExtractor{
		enabled:      cfg.Enabled,
		tesseractBin: tess,
		pdftotextBin: pdftotext,
		pdftoppmBin:  pdftoppm,
		lang:         lang,
		timeout:      timeout,
	}
}

// Available reports whether OCR can run on this host.
func (e *TesseractExtractor) Available(_ context.Context) bool {
	if e == nil || !e.enabled {
		return false
	}
	return lookPath(e.tesseractBin) != ""
}

// ExtractText extracts plain text from PDF or image bytes.
func (e *TesseractExtractor) ExtractText(ctx context.Context, filename, contentType string, data []byte) (string, error) {
	if e == nil || !e.enabled {
		return "", fmt.Errorf("ocr disabled")
	}
	if lookPath(e.tesseractBin) == "" {
		return "", fmt.Errorf("tesseract binary not found: %s", e.tesseractBin)
	}

	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	ct := strings.ToLower(strings.TrimSpace(contentType))
	ext := strings.ToLower(filepath.Ext(filename))
	if ct == mimeApplicationPDF || ext == ".pdf" {
		return e.extractPDF(ctx, data)
	}
	return e.runTesseractOnImage(ctx, data, ext)
}

func (e *TesseractExtractor) extractPDF(ctx context.Context, data []byte) (string, error) {
	if lookPath(e.pdftotextBin) != "" {
		text, err := e.runPDFToText(ctx, data)
		if err == nil && usefulText(text) {
			return text, nil
		}
	}
	if lookPath(e.pdftoppmBin) == "" {
		return "", fmt.Errorf("pdftoppm not found; install poppler-utils for scanned PDFs")
	}
	img, err := e.runPDFToPPMFirstPage(ctx, data)
	if err != nil {
		return "", err
	}
	return e.runTesseractOnImage(ctx, img, ".png")
}

func (e *TesseractExtractor) runPDFToText(ctx context.Context, data []byte) (string, error) {
	dir, err := os.MkdirTemp("", "docsy-ocr-pdf-*")
	if err != nil {
		return "", fmt.Errorf("temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(dir) }()

	pdfPath := filepath.Join(dir, "in.pdf")
	if err := os.WriteFile(pdfPath, data, tempFileMode); err != nil {
		return "", fmt.Errorf("write pdf: %w", err)
	}
	//nolint:gosec // binary from config; paths are under a private temp dir
	cmd := exec.CommandContext(ctx, e.pdftotextBin, "-layout", "-f", "1", "-l", "1", pdfPath, "-")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("pdftotext: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}

func (e *TesseractExtractor) runPDFToPPMFirstPage(ctx context.Context, data []byte) ([]byte, error) {
	dir, err := os.MkdirTemp("", "docsy-ocr-ppm-*")
	if err != nil {
		return nil, fmt.Errorf("temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(dir) }()

	pdfPath := filepath.Join(dir, "in.pdf")
	if err := os.WriteFile(pdfPath, data, tempFileMode); err != nil {
		return nil, fmt.Errorf("write pdf: %w", err)
	}
	outPrefix := filepath.Join(dir, "page")
	//nolint:gosec // binary from config; paths are under a private temp dir
	cmd := exec.CommandContext(ctx, e.pdftoppmBin, "-png", "-f", "1", "-l", "1", "-r", pdfRenderDPI, pdfPath, outPrefix)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("pdftoppm: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	// pdftoppm writes page-1.png
	candidates := []string{
		outPrefix + "-1.png",
		outPrefix + ".png",
	}
	for _, p := range candidates {
		b, err := os.ReadFile(p) //nolint:gosec // path under private temp dir
		if err == nil && len(b) > 0 {
			return b, nil
		}
	}
	entries, _ := os.ReadDir(dir)
	for _, ent := range entries {
		if strings.HasSuffix(strings.ToLower(ent.Name()), ".png") {
			return os.ReadFile(filepath.Join(dir, ent.Name())) //nolint:gosec // temp dir only
		}
	}
	return nil, fmt.Errorf("pdftoppm produced no png")
}

func (e *TesseractExtractor) runTesseractOnImage(ctx context.Context, data []byte, ext string) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("empty image data")
	}
	dir, err := os.MkdirTemp("", "docsy-ocr-img-*")
	if err != nil {
		return "", fmt.Errorf("temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(dir) }()

	if ext == "" || ext == "." {
		ext = ".png"
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	imgPath := filepath.Join(dir, "in"+ext)
	if err := os.WriteFile(imgPath, data, tempFileMode); err != nil {
		return "", fmt.Errorf("write image: %w", err)
	}

	//nolint:gosec // binary from config; image path under private temp dir
	cmd := exec.CommandContext(ctx, e.tesseractBin, imgPath, "stdout", "-l", e.lang, "--psm", "6")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("tesseract: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}

func usefulText(s string) bool {
	s = strings.TrimSpace(s)
	if len(s) < minUsefulPDFTextLen {
		return false
	}
	letters := 0
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= 'á' && r <= 'ú') || r == 'ñ' || r == 'Ñ' {
			letters++
		}
	}
	return letters >= minUsefulLetterCount
}

func lookPath(bin string) string {
	if bin == "" {
		return ""
	}
	if filepath.IsAbs(bin) {
		if st, err := os.Stat(bin); err == nil && !st.IsDir() {
			return bin
		}
		return ""
	}
	p, err := exec.LookPath(bin)
	if err != nil {
		return ""
	}
	return p
}
