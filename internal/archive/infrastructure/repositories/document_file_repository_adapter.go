package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/yovannylopez/docsy-main/internal/archive/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/archive/domain/ports"
)

// DocumentFileRepositoryAdapter implements DocumentFileRepository with sqlx.
type DocumentFileRepositoryAdapter struct {
	db *sqlx.DB
}

// NewDocumentFileRepositoryAdapter creates the adapter.
func NewDocumentFileRepositoryAdapter(db *sqlx.DB) *DocumentFileRepositoryAdapter {
	return &DocumentFileRepositoryAdapter{db: db}
}

type documentFileRow struct {
	ID           string         `db:"id"`
	DocumentID   string         `db:"document_id"`
	StorageKey   string         `db:"storage_key"`
	OriginalName string         `db:"original_name"`
	ContentType  string         `db:"content_type"`
	SizeBytes    int64          `db:"size_bytes"`
	UploadedBy   sql.NullString `db:"uploaded_by"`
	UploadedAt   time.Time      `db:"uploaded_at"`
}

func (r documentFileRow) toEntity() entities.DocumentFile {
	f := entities.DocumentFile{
		ID:           r.ID,
		DocumentID:   r.DocumentID,
		StorageKey:   r.StorageKey,
		OriginalName: r.OriginalName,
		ContentType:  r.ContentType,
		SizeBytes:    r.SizeBytes,
		UploadedAt:   r.UploadedAt,
	}
	if r.UploadedBy.Valid {
		s := r.UploadedBy.String
		f.UploadedBy = &s
	}
	return f
}

// Create inserts attachment metadata.
func (r *DocumentFileRepositoryAdapter) Create(ctx context.Context, file *entities.DocumentFile) error {
	const q = `
		INSERT INTO archive_document_files (
			id, document_id, storage_key, original_name, content_type, size_bytes, uploaded_by, uploaded_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.ExecContext(ctx, q,
		file.ID, file.DocumentID, file.StorageKey, file.OriginalName,
		file.ContentType, file.SizeBytes, nullStr(file.UploadedBy), file.UploadedAt,
	)
	if err != nil {
		return fmt.Errorf("insert document file: %w", err)
	}
	return nil
}

// ListByDocument returns attachments for a document ordered by upload time.
func (r *DocumentFileRepositoryAdapter) ListByDocument(ctx context.Context, documentID string) ([]entities.DocumentFile, error) {
	const q = `
		SELECT id, document_id, storage_key, original_name, content_type, size_bytes, uploaded_by, uploaded_at
		FROM archive_document_files
		WHERE document_id = $1
		ORDER BY uploaded_at DESC, id DESC`

	var rows []documentFileRow
	if err := r.db.SelectContext(ctx, &rows, q, documentID); err != nil {
		return nil, fmt.Errorf("list document files: %w", err)
	}
	out := make([]entities.DocumentFile, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.toEntity())
	}
	return out, nil
}

// FindByID returns one attachment or nil.
func (r *DocumentFileRepositoryAdapter) FindByID(ctx context.Context, documentID, fileID string) (*entities.DocumentFile, error) {
	const q = `
		SELECT id, document_id, storage_key, original_name, content_type, size_bytes, uploaded_by, uploaded_at
		FROM archive_document_files
		WHERE id = $1 AND document_id = $2
		LIMIT 1`

	var row documentFileRow
	err := r.db.GetContext(ctx, &row, q, fileID, documentID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get document file: %w", err)
	}
	f := row.toEntity()
	return &f, nil
}

// Delete removes attachment metadata.
func (r *DocumentFileRepositoryAdapter) Delete(ctx context.Context, documentID, fileID string) error {
	const q = `DELETE FROM archive_document_files WHERE id = $1 AND document_id = $2`
	res, err := r.db.ExecContext(ctx, q, fileID, documentID)
	if err != nil {
		return fmt.Errorf("delete document file: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("delete document file: no rows affected")
	}
	return nil
}

// CountByDocument returns how many attachments a document has.
func (r *DocumentFileRepositoryAdapter) CountByDocument(ctx context.Context, documentID string) (int, error) {
	const q = `SELECT COUNT(*) FROM archive_document_files WHERE document_id = $1`
	var n int
	if err := r.db.GetContext(ctx, &n, q, documentID); err != nil {
		return 0, fmt.Errorf("count document files: %w", err)
	}
	return n, nil
}

// SumSizeBytesForUser totals attachment bytes across workspaces the user belongs to.
func (r *DocumentFileRepositoryAdapter) SumSizeBytesForUser(ctx context.Context, userID string) (int64, error) {
	const q = `
		SELECT COALESCE(SUM(f.size_bytes), 0)
		FROM archive_document_files f
		INNER JOIN archive_documents d ON d.id = f.document_id
		INNER JOIN archive_workspace_members m ON m.workspace_id = d.workspace_id
		WHERE m.user_id = $1`

	var total int64
	if err := r.db.GetContext(ctx, &total, q, userID); err != nil {
		return 0, fmt.Errorf("sum size bytes for user: %w", err)
	}
	return total, nil
}

// FindPrimaryByDocumentIDs returns the newest attachment per document.
func (r *DocumentFileRepositoryAdapter) FindPrimaryByDocumentIDs(
	ctx context.Context,
	documentIDs []string,
) (map[string]ports.DocumentPrimaryFile, error) {
	out := make(map[string]ports.DocumentPrimaryFile)
	if len(documentIDs) == 0 {
		return out, nil
	}

	q, args, err := sqlx.In(`
		SELECT DISTINCT ON (document_id) document_id, original_name, content_type
		FROM archive_document_files
		WHERE document_id IN (?)
		ORDER BY document_id, uploaded_at DESC, id DESC`, documentIDs)
	if err != nil {
		return nil, fmt.Errorf("build primary files query: %w", err)
	}
	q = r.db.Rebind(q)

	type row struct {
		DocumentID   string `db:"document_id"`
		OriginalName string `db:"original_name"`
		ContentType  string `db:"content_type"`
	}
	var rows []row
	if err := r.db.SelectContext(ctx, &rows, q, args...); err != nil {
		return nil, fmt.Errorf("find primary files: %w", err)
	}
	for _, row := range rows {
		out[row.DocumentID] = ports.DocumentPrimaryFile{
			DocumentID:   row.DocumentID,
			OriginalName: row.OriginalName,
			ContentType:  row.ContentType,
		}
	}
	return out, nil
}
