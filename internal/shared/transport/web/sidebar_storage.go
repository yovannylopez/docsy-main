package web

import (
	"fmt"
	"strconv"

	"github.com/yovannylopez/docsy-main/pkg/constants"
)

// ContextKeySidebarStorage is the echo.Context key for sidebar storage usage.
const ContextKeySidebarStorage = "sidebar_storage"

const storagePercentMax = 100

// SidebarStorageData drives the sidebar storage indicator.
type SidebarStorageData struct {
	Show       bool
	Percent    int
	UsedLabel  string
	QuotaLabel string
	Summary    string
}

// NewSidebarStorageData formats used/quota for the sidebar.
func NewSidebarStorageData(usedBytes, quotaBytes int64, percent int) SidebarStorageData {
	if quotaBytes <= 0 && usedBytes <= 0 {
		return SidebarStorageData{}
	}
	if percent < 0 {
		percent = 0
	}
	if percent > storagePercentMax {
		percent = storagePercentMax
	}
	usedLabel := FormatByteSize(usedBytes)
	quotaLabel := FormatByteSize(quotaBytes)
	summary := usedLabel + " de " + quotaLabel + " usados"
	if quotaBytes <= 0 {
		summary = usedLabel + " usados"
		quotaLabel = "—"
	}
	return SidebarStorageData{
		Show:       true,
		Percent:    percent,
		UsedLabel:  usedLabel,
		QuotaLabel: quotaLabel,
		Summary:    summary,
	}
}

// FormatByteSize renders a human-readable size (binary units).
func FormatByteSize(n int64) string {
	if n < 0 {
		n = 0
	}
	switch {
	case n < constants.BytesPerKB:
		return strconv.FormatInt(n, 10) + " B"
	case n < constants.BytesPerMB:
		return formatOneDecimal(float64(n)/float64(constants.BytesPerKB)) + " KB"
	case n < constants.BytesPerGB:
		return formatOneDecimal(float64(n)/float64(constants.BytesPerMB)) + " MB"
	default:
		return formatOneDecimal(float64(n)/float64(constants.BytesPerGB)) + " GB"
	}
}

func formatOneDecimal(v float64) string {
	return fmt.Sprintf("%.1f", v)
}
