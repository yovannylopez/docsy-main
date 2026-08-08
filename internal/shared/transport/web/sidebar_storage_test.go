package web

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yovannylopez/docsy-main/pkg/constants"
)

func TestFormatByteSize(t *testing.T) {
	assert.Equal(t, "512 B", FormatByteSize(512))
	assert.Equal(t, "1.0 KB", FormatByteSize(constants.BytesPerKB))
	assert.Equal(t, "1.5 MB", FormatByteSize(int64(1.5*float64(constants.BytesPerMB))))
	assert.Equal(t, "10.0 GB", FormatByteSize(int64(constants.DefaultStorageQuotaBytes)))
}

func TestNewSidebarStorageData(t *testing.T) {
	data := NewSidebarStorageData(130023424, constants.DefaultStorageQuotaBytes, 1)
	assert.True(t, data.Show)
	assert.Contains(t, data.Summary, "de")
	assert.Contains(t, data.Summary, "usados")
	assert.Equal(t, 1, data.Percent)
}
