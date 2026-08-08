package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalDocumentStorage_PutGetDelete(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocalDocumentStorage(dir)
	require.NoError(t, err)

	ctx := context.Background()
	key := "ws/doc/file.pdf"
	payload := []byte("%PDF-1.4 test")

	require.NoError(t, store.Put(ctx, key, "application/pdf", payload))
	got, _, err := store.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, payload, got)

	require.NoError(t, store.Delete(ctx, key))
	_, _, err = store.Get(ctx, key)
	require.Error(t, err)
}

func TestLocalDocumentStorage_RejectsPathTraversal(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocalDocumentStorage(dir)
	require.NoError(t, err)

	err = store.Put(context.Background(), "../outside.txt", "text/plain", []byte("x"))
	require.Error(t, err)

	// Ensure nothing was written outside the temp root.
	_, err = os.Stat(filepath.Join(filepath.Dir(dir), "outside.txt"))
	assert.True(t, os.IsNotExist(err))
}
