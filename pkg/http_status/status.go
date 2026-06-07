// Package http_status defines the Status type with common HTTP status codes, user-friendly descriptions
// for API responses (Spanish), and classification utilities (2xx, 4xx, 5xx).
// In runtime, it only depends on net/http. For unified JSON responses in the project, it is typically
// used together with pkg/responses (field Status *http_status.Status).
package http_status

import (
	"maps"
	"net/http"
)

// EnvelopeInternalServerErrorMessageEN is the fixed text for the "message" field for errors
// not mapped in the central Echo handler (internal/shared/transport/middleware).
// Keeps a stable English message for clients; does not match Description (Spanish).
const EnvelopeInternalServerErrorMessageEN = "Internal server error"

// Status groups an HTTP code and a user-friendly description (e.g. JSON).
type Status struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
}

// Common HTTP status codes (English descriptions, first letter capitalized).
var (
	// Success responses
	OK        = Status{http.StatusOK, "Operation successful"}
	Created   = Status{http.StatusCreated, "Resource created successfully"}
	Accepted  = Status{http.StatusAccepted, "Request accepted"}
	NoContent = Status{http.StatusNoContent, "No content"}

	// Client errors
	BadRequest          = Status{http.StatusBadRequest, "Invalid request"}
	Unauthorized        = Status{http.StatusUnauthorized, "Unauthorized"}
	Forbidden           = Status{http.StatusForbidden, "Forbidden"}
	NotFound            = Status{http.StatusNotFound, "Resource not found"}
	MethodNotAllowed    = Status{http.StatusMethodNotAllowed, "Method not allowed"}
	Conflict            = Status{http.StatusConflict, "Conflict"}
	UnprocessableEntity = Status{http.StatusUnprocessableEntity, "Unprocessable entity"}
	TooManyRequests     = Status{http.StatusTooManyRequests, "Too many requests"}

	// Server errors
	InternalError      = Status{http.StatusInternalServerError, "Internal server error"}
	NotImplemented     = Status{http.StatusNotImplemented, "Not implemented"}
	BadGateway         = Status{http.StatusBadGateway, "Bad gateway"}
	ServiceUnavailable = Status{http.StatusServiceUnavailable, "Service unavailable"}
	GatewayTimeout     = Status{http.StatusGatewayTimeout, "Gateway timeout"}
)

// commonStatusCodes is the immutable catalog of the package; CommonStatusCodes returns a copy.
var commonStatusCodes map[string]Status

// byCode indexes the predefined Status by numeric code (e.g. LookupByCode).
var byCode map[int]Status

func init() {
	commonStatusCodes = map[string]Status{
		"OK":                  OK,
		"Created":             Created,
		"Accepted":            Accepted,
		"NoContent":           NoContent,
		"BadRequest":          BadRequest,
		"Unauthorized":        Unauthorized,
		"Forbidden":           Forbidden,
		"NotFound":            NotFound,
		"MethodNotAllowed":    MethodNotAllowed,
		"Conflict":            Conflict,
		"UnprocessableEntity": UnprocessableEntity,
		"TooManyRequests":     TooManyRequests,
		"InternalError":       InternalError,
		"NotImplemented":      NotImplemented,
		"BadGateway":          BadGateway,
		"ServiceUnavailable":  ServiceUnavailable,
		"GatewayTimeout":      GatewayTimeout,
	}

	byCode = make(map[int]Status, len(commonStatusCodes))
	for _, s := range commonStatusCodes {
		byCode[s.Code] = s
	}
}

// Custom creates a Status with arbitrary code and description.
func Custom(code int, description string) Status {
	return Status{Code: code, Description: description}
}

// IsSuccess checks if the status code is successful (2xx)
func (s Status) IsSuccess() bool {
	return s.Code >= 200 && s.Code < 300
}

// IsClientError checks if the status code is a client error (4xx)
func (s Status) IsClientError() bool {
	return s.Code >= 400 && s.Code < 500
}

// IsServerError checks if the status code is a server error (5xx)
func (s Status) IsServerError() bool {
	return s.Code >= 500 && s.Code < 600
}

// GetHTTPStatusText returns the official HTTP status code text
func (s Status) GetHTTPStatusText() string {
	return http.StatusText(s.Code)
}

// CommonStatusCodes returns a copy of the predefined status codes map.
// Modifying the returned map does not alter the internal package catalog.
func CommonStatusCodes() map[string]Status {
	return maps.Clone(commonStatusCodes)
}

// LookupByCode returns the predefined Status associated with the HTTP code, if it exists in the catalog.
func LookupByCode(code int) (Status, bool) {
	s, ok := byCode[code]
	return s, ok
}

// Ptr returns a pointer to a copy of s (useful for *http_status.Status signatures).
func Ptr(s Status) *Status {
	x := s
	return &x
}
