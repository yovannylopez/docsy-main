package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/yovannylopez/docsy-main/internal/archive/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/archive/domain/entities"
)

// DocumentRepositoryAdapter implements DocumentRepository with sqlx.
type DocumentRepositoryAdapter struct {
	db *sqlx.DB
}

// NewDocumentRepositoryAdapter creates the adapter.
func NewDocumentRepositoryAdapter(db *sqlx.DB) *DocumentRepositoryAdapter {
	return &DocumentRepositoryAdapter{db: db}
}

type documentRow struct {
	ID              string         `db:"id"`
	WorkspaceID     string         `db:"workspace_id"`
	CategoryCode    string         `db:"category_code"`
	Title           string         `db:"title"`
	DocumentDate    sql.NullTime   `db:"document_date"`
	DueDate         sql.NullTime   `db:"due_date"`
	Issuer          sql.NullString `db:"issuer"`
	ReferenceNumber sql.NullString `db:"reference_number"`
	AmountCents     sql.NullInt64  `db:"amount_cents"`
	Currency        string         `db:"currency"`
	Notes           sql.NullString `db:"notes"`
	Status          string         `db:"status"`
	CreatedBy       sql.NullString `db:"created_by"`
	UpdatedBy       sql.NullString `db:"updated_by"`
	CreatedAt       time.Time      `db:"created_at"`
	UpdatedAt       time.Time      `db:"updated_at"`
}

func (r documentRow) toEntity() entities.Document {
	doc := entities.Document{
		ID:           r.ID,
		WorkspaceID:  r.WorkspaceID,
		CategoryCode: r.CategoryCode,
		Title:        r.Title,
		Currency:     r.Currency,
		Status:       r.Status,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}
	if r.DocumentDate.Valid {
		t := r.DocumentDate.Time
		doc.DocumentDate = &t
	}
	if r.DueDate.Valid {
		t := r.DueDate.Time
		doc.DueDate = &t
	}
	if r.Issuer.Valid {
		s := r.Issuer.String
		doc.Issuer = &s
	}
	if r.ReferenceNumber.Valid {
		s := r.ReferenceNumber.String
		doc.ReferenceNumber = &s
	}
	if r.AmountCents.Valid {
		v := r.AmountCents.Int64
		doc.AmountCents = &v
	}
	if r.Notes.Valid {
		s := r.Notes.String
		doc.Notes = &s
	}
	if r.CreatedBy.Valid {
		s := r.CreatedBy.String
		doc.CreatedBy = &s
	}
	if r.UpdatedBy.Valid {
		s := r.UpdatedBy.String
		doc.UpdatedBy = &s
	}
	return doc
}

// List returns documents matching filters for a workspace.
func (r *DocumentRepositoryAdapter) List(ctx context.Context, workspaceID string, filter dtos.ListDocumentsFilter) ([]entities.Document, int, error) {
	where := []string{"workspace_id = $1"}
	args := []any{workspaceID}
	argN := 2

	status := filter.Status
	if status == "" {
		status = entities.DocumentStatusActive
	}
	if status != "all" {
		where = append(where, fmt.Sprintf("status = $%d", argN))
		args = append(args, status)
		argN++
	}
	if filter.Category != "" {
		where = append(where, fmt.Sprintf("category_code = $%d", argN))
		args = append(args, filter.Category)
		argN++
	}
	if q := strings.TrimSpace(filter.Query); q != "" {
		where = append(where, fmt.Sprintf(
			"(title ILIKE $%d OR COALESCE(issuer,'') ILIKE $%d OR COALESCE(reference_number,'') ILIKE $%d)",
			argN, argN, argN,
		))
		args = append(args, "%"+q+"%")
		argN++
	}
	if filter.From != nil {
		where = append(where, fmt.Sprintf("document_date >= $%d", argN))
		args = append(args, filter.From.Format("2006-01-02"))
		argN++
	}
	if filter.To != nil {
		where = append(where, fmt.Sprintf("document_date <= $%d", argN))
		args = append(args, filter.To.Format("2006-01-02"))
		argN++
	}
	if filter.DueBefore != nil {
		where = append(where, fmt.Sprintf("due_date IS NOT NULL AND due_date <= $%d", argN))
		args = append(args, filter.DueBefore.Format("2006-01-02"))
		argN++
	}

	whereSQL := strings.Join(where, " AND ")

	countQ := "SELECT COUNT(*) FROM archive_documents WHERE " + whereSQL
	var total int
	if err := r.db.GetContext(ctx, &total, countQ, args...); err != nil {
		return nil, 0, fmt.Errorf("count documents: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	listQ := fmt.Sprintf(`
		SELECT id, workspace_id, category_code, title, document_date, due_date, issuer,
		       reference_number, amount_cents, currency, notes, status, created_by, updated_by, created_at, updated_at
		FROM archive_documents
		WHERE %s
		ORDER BY COALESCE(document_date, created_at::date) DESC, created_at DESC
		LIMIT $%d OFFSET $%d`, whereSQL, argN, argN+1)
	args = append(args, limit, offset)

	var rows []documentRow
	if err := r.db.SelectContext(ctx, &rows, listQ, args...); err != nil {
		return nil, 0, fmt.Errorf("list documents: %w", err)
	}

	out := make([]entities.Document, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.toEntity())
	}
	return out, total, nil
}

// FindByID returns a document in the workspace or nil.
func (r *DocumentRepositoryAdapter) FindByID(ctx context.Context, workspaceID, documentID string) (*entities.Document, error) {
	const q = `
		SELECT id, workspace_id, category_code, title, document_date, due_date, issuer,
		       reference_number, amount_cents, currency, notes, status, created_by, updated_by, created_at, updated_at
		FROM archive_documents
		WHERE id = $1 AND workspace_id = $2
		LIMIT 1`

	var row documentRow
	err := r.db.GetContext(ctx, &row, q, documentID, workspaceID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get document: %w", err)
	}
	doc := row.toEntity()
	return &doc, nil
}

// Create inserts a document.
func (r *DocumentRepositoryAdapter) Create(ctx context.Context, doc *entities.Document) error {
	const q = `
		INSERT INTO archive_documents (
			id, workspace_id, category_code, title, document_date, due_date, issuer,
			reference_number, amount_cents, currency, notes, status, created_by, updated_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		)`

	_, err := r.db.ExecContext(ctx, q,
		doc.ID, doc.WorkspaceID, doc.CategoryCode, doc.Title,
		nullDate(doc.DocumentDate), nullDate(doc.DueDate),
		nullStr(doc.Issuer), nullStr(doc.ReferenceNumber), nullInt64(doc.AmountCents),
		doc.Currency, nullStr(doc.Notes), doc.Status,
		nullStr(doc.CreatedBy), nullStr(doc.UpdatedBy),
		doc.CreatedAt, doc.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert document: %w", err)
	}
	return nil
}

// Update persists document changes.
func (r *DocumentRepositoryAdapter) Update(ctx context.Context, doc *entities.Document) error {
	const q = `
		UPDATE archive_documents SET
			category_code = $2, title = $3, document_date = $4, due_date = $5, issuer = $6,
			reference_number = $7, amount_cents = $8, currency = $9, notes = $10, status = $11,
			updated_by = $12, updated_at = $13
		WHERE id = $1 AND workspace_id = $14`

	res, err := r.db.ExecContext(ctx, q,
		doc.ID, doc.CategoryCode, doc.Title,
		nullDate(doc.DocumentDate), nullDate(doc.DueDate),
		nullStr(doc.Issuer), nullStr(doc.ReferenceNumber), nullInt64(doc.AmountCents),
		doc.Currency, nullStr(doc.Notes), doc.Status,
		nullStr(doc.UpdatedBy), doc.UpdatedAt, doc.WorkspaceID,
	)
	if err != nil {
		return fmt.Errorf("update document: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("update document: no rows affected")
	}
	return nil
}

// Delete hard-deletes a document row (attachments cascade via FK).
func (r *DocumentRepositoryAdapter) Delete(ctx context.Context, workspaceID, documentID string) error {
	const q = `DELETE FROM archive_documents WHERE id = $1 AND workspace_id = $2`
	res, err := r.db.ExecContext(ctx, q, documentID, workspaceID)
	if err != nil {
		return fmt.Errorf("delete document: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("delete document: no rows affected")
	}
	return nil
}

// CategoryExists reports whether a category code is active.
func (r *DocumentRepositoryAdapter) CategoryExists(ctx context.Context, code string) (bool, error) {
	const q = `SELECT EXISTS(SELECT 1 FROM archive_document_categories WHERE code = $1 AND is_active = true)`
	var ok bool
	if err := r.db.GetContext(ctx, &ok, q, code); err != nil {
		return false, fmt.Errorf("check category: %w", err)
	}
	return ok, nil
}

// ListCategories returns active categories ordered by sort_order.
func (r *DocumentRepositoryAdapter) ListCategories(ctx context.Context) ([]entities.DocumentCategory, error) {
	const q = `
		SELECT code, label_es, sort_order, is_active
		FROM archive_document_categories
		WHERE is_active = true
		ORDER BY sort_order ASC, code ASC`

	var cats []entities.DocumentCategory
	if err := r.db.SelectContext(ctx, &cats, q); err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	return cats, nil
}

// CountByCategory returns document counts per category_code for a workspace.
func (r *DocumentRepositoryAdapter) CountByCategory(ctx context.Context, workspaceID string, status string) (map[string]int, error) {
	where := []string{"workspace_id = $1"}
	args := []any{workspaceID}
	argN := 2

	if status == "" {
		status = entities.DocumentStatusActive
	}
	if status != "all" {
		where = append(where, fmt.Sprintf("status = $%d", argN))
		args = append(args, status)
	}

	q := "SELECT category_code, COUNT(*) AS n FROM archive_documents WHERE " +
		strings.Join(where, " AND ") +
		" GROUP BY category_code"

	type row struct {
		CategoryCode string `db:"category_code"`
		N            int    `db:"n"`
	}
	var rows []row
	if err := r.db.SelectContext(ctx, &rows, q, args...); err != nil {
		return nil, fmt.Errorf("count documents by category: %w", err)
	}

	out := make(map[string]int, len(rows))
	for _, row := range rows {
		out[row.CategoryCode] = row.N
	}
	return out, nil
}

func nullStr(s *string) any {
	if s == nil {
		return nil
	}
	return *s
}

func nullInt64(v *int64) any {
	if v == nil {
		return nil
	}
	return *v
}

func nullDate(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.Format("2006-01-02")
}
