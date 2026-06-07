package responses

import (
	"github.com/labstack/echo/v4"

	domerrs "github.com/yovannylopez/docsy-main/pkg/errors"
	"github.com/yovannylopez/docsy-main/pkg/http_status"
)

const (
	msgInternalGenericES    = "An internal error occurred. Please try again later."
	msgExternalGenericES    = "An external service did not respond correctly. Please try again later."
	msgServiceUnavailableES = "The service is temporarily unavailable. Please try again later."
	msgDatabaseGenericES    = "The data operation could not be completed. Please try again later."
)

// ToHTTPAppError looks up a *domerrs.AppError in the err chain and translates it to the *AppError
// of this package (type + HTTP code + client-safe message).
func ToHTTPAppError(err error) (*AppError, bool) {
	dom, ok := domerrs.GetAppError(err)
	if !ok || dom == nil {
		return nil, false
	}

	respType, httpCode := domainTypeToResponse(dom.Type)
	clientMsg := clientMessageFromDomain(dom, httpCode)

	out := NewAppError(respType, httpCode, clientMsg)

	details := domainDetailsForClient(dom, httpCode >= 500)
	if dom.Details != "" {
		if details == nil {
			details = map[string]any{}
		}
		details["details"] = dom.Details
	}
	if len(details) > 0 {
		out = out.WithDetails(details)
	}

	return out, true
}

func domainTypeToResponse(t domerrs.ErrorType) (ErrorType, int) {
	switch t {
	case domerrs.ErrorTypeValidation:
		return ValidationError, http_status.BadRequest.Code
	case domerrs.ErrorTypeNotFound:
		return NotFoundError, http_status.NotFound.Code
	case domerrs.ErrorTypeUnauthorized:
		return AuthenticationError, http_status.Unauthorized.Code
	case domerrs.ErrorTypeForbidden:
		return AuthorizationError, http_status.Forbidden.Code
	case domerrs.ErrorTypeConflict:
		return ConflictError, http_status.Conflict.Code
	case domerrs.ErrorTypeInternal:
		return InternalServerError, http_status.InternalError.Code
	case domerrs.ErrorTypeDatabase:
		return DatabaseError, http_status.InternalError.Code
	case domerrs.ErrorTypeExternal:
		return ServiceError, http_status.BadGateway.Code
	case domerrs.ErrorTypeServiceUnavailable:
		return ServiceError, http_status.ServiceUnavailable.Code
	default:
		return InternalServerError, http_status.InternalError.Code
	}
}

func clientMessageFromDomain(dom *domerrs.AppError, httpCode int) string {
	if dom.UserMessage != "" {
		return dom.UserMessage
	}

	switch dom.Type {
	case domerrs.ErrorTypeDatabase:
		return msgDatabaseGenericES
	case domerrs.ErrorTypeInternal:
		return msgInternalGenericES
	case domerrs.ErrorTypeExternal:
		return msgExternalGenericES
	case domerrs.ErrorTypeServiceUnavailable:
		return msgServiceUnavailableES
	default:
		if httpCode >= 500 {
			return msgInternalGenericES
		}
		return dom.Message
	}
}

// MapDomainError translates a domain error (pkg/errors.AppError) to an Echo HTTP response.
// Covers all ErrorType values defined in pkg/errors; if the error is not a domain AppError
// it responds with a generic 500 Internal Server Error.
func MapDomainError(c echo.Context, err error) error {
	appErr, ok := ToHTTPAppError(err)
	if !ok {
		return InternalError(c, "internal server error")
	}

	st, found := http_status.LookupByCode(appErr.Code)
	if !found {
		st = http_status.Custom(appErr.Code, appErr.Message)
	}

	return EchoAppError(c, &st, appErr)
}

func domainDetailsForClient(dom *domerrs.AppError, serverSide bool) map[string]any {
	if serverSide {
		return nil
	}
	d := make(map[string]any)
	if dom.Code != "" {
		d["domain_code"] = dom.Code
	}
	if dom.Operation != "" {
		d["operation"] = dom.Operation
	}
	if dom.Resource != "" {
		d["resource"] = dom.Resource
	}
	if len(d) == 0 {
		return nil
	}
	return d
}
