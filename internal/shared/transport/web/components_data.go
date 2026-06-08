package web

import (
	"fmt"
	"net/url"
	"strconv"
)

// ConfirmationModalData holds view data for the confirmation modal partial.
type ConfirmationModalData struct {
	ModalID            string
	Title              string
	Message            string
	ConfirmText        string
	CancelText         string
	ConfirmURL         string
	ConfirmMethod      string // post, delete, patch
	ConfirmHXTarget    string
	ConfirmHXSwap      string
	HeaderClass        string
	ConfirmButtonClass string
}

// SuccessModalData holds view data for the success modal partial.
type SuccessModalData struct {
	ModalID            string
	Title              string
	Summary            string
	Sections           []SuccessModalSection
	PrimaryActionURL   string
	PrimaryActionLabel string
}

// SuccessModalSection groups detail rows in the success modal.
type SuccessModalSection struct {
	Title string
	Rows  []SuccessModalRow
}

// SuccessModalRow is a label/value pair in the success modal.
type SuccessModalRow struct {
	Label string
	Value string
}

// ModalShellData holds view data for the generic modal shell partial.
type ModalShellData struct {
	ModalID       string
	ModalTitle    string
	ModalMessage  string
	ConfirmText   string
	CancelText    string
	ConfirmURL    string
	ConfirmMethod string
	HeaderTone    string // warning, danger, success, info
}

// FormFieldData describes a single form input for partials/form-field.
type FormFieldData struct {
	ID       string
	Name     string
	Label    string
	Type     string
	Value    string
	Invalid  bool
	Error    string
	Disabled bool
}

// FormFieldPageData wraps FormFieldData for template execution.
type FormFieldPageData struct {
	Field FormFieldData
}

// AlertSuccessData holds view data for the success alert partial.
type AlertSuccessData struct {
	SuccessTitle   string
	SuccessMessage string
}

// PaginationData holds view data for table pagination partials.
type PaginationData struct {
	Offset      int
	Limit       int
	Total       int
	HasPrevious bool
	HasNext     bool
	PrevURL     string
	NextURL     string
	PageStart   int
	PageEnd     int
}

// NewPaginationData builds pagination view data with prev/next URLs.
func NewPaginationData(offset, limit, total int, basePath string, query url.Values) PaginationData {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	pageStart := 0
	pageEnd := 0
	if total > 0 {
		pageStart = offset + 1
		pageEnd = offset + limit
		if pageEnd > total {
			pageEnd = total
		}
	}

	hasPrevious := offset > 0
	hasNext := offset+limit < total

	return PaginationData{
		Offset:      offset,
		Limit:       limit,
		Total:       total,
		HasPrevious: hasPrevious,
		HasNext:     hasNext,
		PrevURL:     paginationURL(basePath, query, max(0, offset-limit)),
		NextURL:     paginationURL(basePath, query, offset+limit),
		PageStart:   pageStart,
		PageEnd:     pageEnd,
	}
}

func paginationURL(basePath string, query url.Values, offset int) string {
	q := cloneQuery(query)
	q.Set("offset", strconv.Itoa(offset))
	encoded := q.Encode()
	if encoded == "" {
		return basePath
	}
	return fmt.Sprintf("%s?%s", basePath, encoded)
}

func cloneQuery(query url.Values) url.Values {
	if query == nil {
		return url.Values{}
	}
	out := make(url.Values, len(query))
	for k, vals := range query {
		cp := make([]string, len(vals))
		copy(cp, vals)
		out[k] = cp
	}
	return out
}
