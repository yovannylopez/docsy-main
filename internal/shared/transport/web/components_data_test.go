package web

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewPaginationData_FirstPage(t *testing.T) {
	q := url.Values{"q": []string{"ana"}}
	data := NewPaginationData(0, 10, 25, "/usuarios", q)

	require.Equal(t, 1, data.PageStart)
	require.Equal(t, 10, data.PageEnd)
	require.True(t, data.HasNext)
	require.False(t, data.HasPrevious)
	require.Contains(t, data.NextURL, "offset=10")
	require.Contains(t, data.NextURL, "q=ana")
	require.Contains(t, data.PrevURL, "offset=0")
}

func TestNewPaginationData_LastPage(t *testing.T) {
	data := NewPaginationData(20, 10, 25, "/usuarios", nil)

	require.Equal(t, 21, data.PageStart)
	require.Equal(t, 25, data.PageEnd)
	require.False(t, data.HasNext)
	require.True(t, data.HasPrevious)
	require.Contains(t, data.PrevURL, "offset=10")
}

func TestNewPaginationData_EmptyResult(t *testing.T) {
	data := NewPaginationData(0, 10, 0, "/usuarios", nil)

	require.Equal(t, 0, data.PageStart)
	require.Equal(t, 0, data.PageEnd)
	require.False(t, data.HasNext)
	require.False(t, data.HasPrevious)
}
