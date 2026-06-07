package openapi

// Example texts for documenting HTTP 400 errors for query limit/offset.
// Must match the messages returned by pkg/pagination (Validate / ParseFromQuery)
// when using DefaultConfig or a Config with MaxLimit=100 and MinLimit=1.
const (
	// PaginationQueryErrorLimitAboveMax is the typical detail when limit > MaxPageSize (100).
	PaginationQueryErrorLimitAboveMax = "pagination: limit out of allowed range: maximum 100"
	// PaginationQueryErrorLimitBelowMin is the typical detail when limit < MinLimit (1).
	PaginationQueryErrorLimitBelowMin = "pagination: limit out of allowed range: minimum 1"
	// PaginationQueryErrorOffsetNegative is the detail when offset < 0.
	PaginationQueryErrorOffsetNegative = "pagination: negative offset"
)

const badRequestDescriptionES = "Invalid request"

// PaginationQueryBadRequestContent documents a typical 400 for invalid limit/offset
// (example: limit above the maximum allowed), aligned with pkg/responses for errors.
func PaginationQueryBadRequestContent() map[string]MediaType {
	return JSONErrorRefContent(ErrorEnvelopeExample(400, badRequestDescriptionES, PaginationQueryErrorLimitAboveMax))
}

// PaginationMetadataSchema describes pagination.Metadata from pkg/pagination ("pagination" block in OKPaginated).
func PaginationMetadataSchema() *Schema {
	return &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"total":        {Type: "integer"},
			"limit":        {Type: "integer"},
			"offset":       {Type: "integer"},
			"total_pages":  {Type: "integer"},
			"current_page": {Type: "integer"},
			"has_next":     {Type: "boolean"},
			"has_previous": {Type: "boolean"},
		},
	}
}

// OKPaginatedEnvelopeSchema documents responses.OKPaginated: status, message, data (array), pagination.
func OKPaginatedEnvelopeSchema(items *Schema) *Schema {
	return &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"status": {
				Type: "object",
				Properties: map[string]*Schema{
					"code":        {Type: "integer"},
					"description": {Type: "string"},
				},
			},
			"message":    {Type: "string"},
			"data":       {Type: "array", Items: items},
			"pagination": PaginationMetadataSchema(),
		},
	}
}
