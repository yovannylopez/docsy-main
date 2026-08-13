package handlers

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDocumentsDatePresetURLs(t *testing.T) {
	now := time.Date(2026, 8, 12, 15, 0, 0, 0, time.UTC)
	monthURL, yearURL, last90URL := documentsDatePresetURLs(docsListURLParams{
		WorkspaceID: "ws-1",
		Status:      "active",
		View:        "grid",
	}, now)

	monthQ, err := url.Parse(monthURL)
	require.NoError(t, err)
	assert.Equal(t, "2026-08-01", monthQ.Query().Get("from"))
	assert.Equal(t, "2026-08-12", monthQ.Query().Get("to"))

	yearQ, err := url.Parse(yearURL)
	require.NoError(t, err)
	assert.Equal(t, "2026-01-01", yearQ.Query().Get("from"))
	assert.Equal(t, "2026-08-12", yearQ.Query().Get("to"))

	lastQ, err := url.Parse(last90URL)
	require.NoError(t, err)
	assert.Equal(t, "2026-05-14", lastQ.Query().Get("from"))
	assert.Equal(t, "2026-08-12", lastQ.Query().Get("to"))
	assert.Equal(t, "ws-1", lastQ.Query().Get("workspace_id"))
}
