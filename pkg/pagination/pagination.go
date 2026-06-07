// Package pagination parsea y valida limit/offset desde query strings, calcula metadata
// de página (total_pages, has_next, etc.) y construye respuestas JSON con bloque "pagination".
//
// DefaultConfig usa los mismos valores numéricos que pkg/constants (DefaultPageSize, MaxPageSize).
// Si cambian límites de negocio o esquema de listados, revisar también docs/specs/data_schema.md.
package pagination

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/yovannylopez/docsy-main/pkg/constants"
)

// Validation errors (compatible with errors.Is).
var (
	ErrLimitOutOfRange = errors.New("pagination: limit out of allowed range")
	ErrNegativeOffset  = errors.New("pagination: negative offset")
)

// Params represents pagination parameters
type Params struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// Metadata contains pagination information
type Metadata struct {
	Total       int  `json:"total"`
	Limit       int  `json:"limit"`
	Offset      int  `json:"offset"`
	TotalPages  int  `json:"total_pages"`
	CurrentPage int  `json:"current_page"`
	HasNext     bool `json:"has_next"`
	HasPrev     bool `json:"has_previous"`
}

// Response represents a paginated response
type Response struct {
	Data     any      `json:"data"`
	Metadata Metadata `json:"pagination"`
}

// Config holds configuration for the pagination parser
type Config struct {
	DefaultLimit int
	MaxLimit     int
	MinLimit     int
}

// DefaultConfig coincide con pkg/constants.DefaultPageSize y MaxPageSize.
var DefaultConfig = Config{
	DefaultLimit: constants.DefaultPageSize,
	MaxLimit:     constants.MaxPageSize,
	MinLimit:     1,
}

// Parser handles parsing and validation of pagination parameters
type Parser struct {
	config Config
}

// NewParser creates a new parser instance with custom configuration
func NewParser(config Config) *Parser {
	return &Parser{
		config: config,
	}
}

// NewDefaultParser creates a new parser instance with the default configuration
func NewDefaultParser() *Parser {
	return &Parser{
		config: DefaultConfig,
	}
}

// ParseFromQuery extracts and validates pagination parameters from query parameters
func (p *Parser) ParseFromQuery(limitStr, offsetStr string) (*Params, error) {
	params := &Params{
		Limit:  p.config.DefaultLimit,
		Offset: 0,
	}

	// Parse limit
	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return nil, fmt.Errorf("parameter 'limit' must be an integer: %w", err)
		}
		params.Limit = limit
	}

	// Parse offset
	if offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			return nil, fmt.Errorf("parameter 'offset' must be an integer: %w", err)
		}
		params.Offset = offset
	}

	// Validate parameters
	if err := p.Validate(params); err != nil {
		return nil, err
	}

	return params, nil
}

// Validate valida los parámetros de paginación.
func (p *Parser) Validate(params *Params) error {
	if params.Limit < p.config.MinLimit {
		return fmt.Errorf("%w: minimum %d", ErrLimitOutOfRange, p.config.MinLimit)
	}

	if params.Limit > p.config.MaxLimit {
		return fmt.Errorf("%w: maximum %d", ErrLimitOutOfRange, p.config.MaxLimit)
	}

	if params.Offset < 0 {
		return fmt.Errorf("%w", ErrNegativeOffset)
	}

	return nil
}

// CreateMetadata crea metadata de paginación.
// Si offset implica una página posterior a total_pages (p. ej. total=25, limit=10, offset=100),
// current_page se limita a total_pages para evitar metadatos incoherentes.
func CreateMetadata(params *Params, total int) Metadata {
	totalPages := (total + params.Limit - 1) / params.Limit
	if totalPages == 0 {
		totalPages = 1
	}

	currentPage := (params.Offset / params.Limit) + 1
	if currentPage < 1 {
		currentPage = 1
	}
	if currentPage > totalPages {
		currentPage = totalPages
	}

	return Metadata{
		Total:       total,
		Limit:       params.Limit,
		Offset:      params.Offset,
		TotalPages:  totalPages,
		CurrentPage: currentPage,
		HasNext:     currentPage < totalPages,
		HasPrev:     currentPage > 1,
	}
}

// CreateResponse crea una respuesta paginada
func CreateResponse(data any, params *Params, total int) Response {
	return Response{
		Data:     data,
		Metadata: CreateMetadata(params, total),
	}
}

// GetPageFromOffset calcula el número de página basado en offset y limit.
// Si limit <= 0, devuelve 1; use Validate antes de confiar en el resultado con datos reales.
func GetPageFromOffset(offset, limit int) int {
	if limit <= 0 {
		return 1
	}
	return (offset / limit) + 1
}

// GetOffsetFromPage calculates the offset based on page number and limit
func GetOffsetFromPage(page, limit int) int {
	if page <= 1 {
		return 0
	}
	return (page - 1) * limit
}
