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

const sqlWorkspaceIDEq = "workspace_id = $1"

// DocumentRepositoryAdapter implements DocumentRepository with sqlx.
type DocumentRepositoryAdapter struct {
	db *sqlx.DB
}

// NewDocumentRepositoryAdapter creates the adapter.
func NewDocumentRepositoryAdapter(db *sqlx.DB) *DocumentRepositoryAdapter {
	return &DocumentRepositoryAdapter{db: db}
}

type documentRow struct {
	ID              string           `db:"id"`
	WorkspaceID     string           `db:"workspace_id"`
	CategoryCode    string           `db:"category_code"`
	Title           string           `db:"title"`
	DocumentDate    sql.NullTime     `db:"document_date"`
	DueDate         sql.NullTime     `db:"due_date"`
	Issuer          sql.NullString   `db:"issuer"`
	ReferenceNumber sql.NullString   `db:"reference_number"`
	AmountCents     sql.NullInt64    `db:"amount_cents"`
	Currency        string           `db:"currency"`
	Notes           sql.NullString   `db:"notes"`
	ExtraFields     extraFieldsJSONB `db:"extra_fields"`
	Status          string           `db:"status"`
	CreatedBy       sql.NullString   `db:"created_by"`
	UpdatedBy       sql.NullString   `db:"updated_by"`
	CreatedAt       time.Time        `db:"created_at"`
	UpdatedAt       time.Time        `db:"updated_at"`
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
		ExtraFields:  entities.ExtraFields(r.ExtraFields),
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
	where := []string{sqlWorkspaceIDEq}
	args := []any{workspaceID}
	argN := 2

	status := filter.Status
	if status == "" {
		status = entities.DocumentStatusActive
	}
	if status != entities.DocumentStatusAll {
		where = append(where, fmt.Sprintf("status = $%d", argN))
		args = append(args, status)
		argN++
	}
	if filter.Category != "" {
		where = append(where, fmt.Sprintf("category_code = $%d", argN))
		args = append(args, filter.Category)
		argN++
	}
	where, args, argN = appendSearchQueryFilter(where, args, argN, filter.Query)
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
	if filter.DueFrom != nil {
		where = append(where, fmt.Sprintf("due_date IS NOT NULL AND due_date >= $%d", argN))
		args = append(args, filter.DueFrom.Format("2006-01-02"))
		argN++
	}
	if filter.DueTo != nil {
		where = append(where, fmt.Sprintf("due_date IS NOT NULL AND due_date <= $%d", argN))
		args = append(args, filter.DueTo.Format("2006-01-02"))
		argN++
	}
	if filter.DueBefore != nil {
		where = append(where, fmt.Sprintf("due_date IS NOT NULL AND due_date <= $%d", argN))
		args = append(args, filter.DueBefore.Format("2006-01-02"))
		argN++
	}
	switch strings.ToLower(strings.TrimSpace(filter.DueAlert)) {
	case entities.DueAlertExpired:
		where = append(where, fmt.Sprintf("due_date IS NOT NULL AND due_date < $%d", argN))
		args = append(args, time.Now().Format("2006-01-02"))
		argN++
	case entities.DueAlertUpcoming:
		today := time.Now().Format("2006-01-02")
		horizon := time.Now().AddDate(0, 0, entities.DueSoonWindowDays).Format("2006-01-02")
		where = append(where, fmt.Sprintf("due_date IS NOT NULL AND due_date >= $%d AND due_date <= $%d", argN, argN+1))
		args = append(args, today, horizon)
		argN += 2
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
		       reference_number, amount_cents, currency, notes, extra_fields, status,
		       created_by, updated_by, created_at, updated_at
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
		       reference_number, amount_cents, currency, notes, extra_fields, status,
		       created_by, updated_by, created_at, updated_at
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
			reference_number, amount_cents, currency, notes, extra_fields, status,
			created_by, updated_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		)`

	_, err := r.db.ExecContext(ctx, q,
		doc.ID, doc.WorkspaceID, doc.CategoryCode, doc.Title,
		nullDate(doc.DocumentDate), nullDate(doc.DueDate),
		nullStr(doc.Issuer), nullStr(doc.ReferenceNumber), nullInt64(doc.AmountCents),
		doc.Currency, nullStr(doc.Notes), extraFieldsJSONB(doc.ExtraFields), doc.Status,
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
			reference_number = $7, amount_cents = $8, currency = $9, notes = $10,
			extra_fields = $11, status = $12, updated_by = $13, updated_at = $14
		WHERE id = $1 AND workspace_id = $15`

	res, err := r.db.ExecContext(ctx, q,
		doc.ID, doc.CategoryCode, doc.Title,
		nullDate(doc.DocumentDate), nullDate(doc.DueDate),
		nullStr(doc.Issuer), nullStr(doc.ReferenceNumber), nullInt64(doc.AmountCents),
		doc.Currency, nullStr(doc.Notes), extraFieldsJSONB(doc.ExtraFields), doc.Status,
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

// CategoryExists reports whether a category code is active for the workspace (system or custom).
func (r *DocumentRepositoryAdapter) CategoryExists(ctx context.Context, workspaceID, code string) (bool, error) {
	const q = `
		SELECT EXISTS(
			SELECT 1 FROM archive_document_categories
			WHERE code = $1 AND is_active = true
			  AND (is_system = true OR workspace_id = $2)
		)`
	var ok bool
	if err := r.db.GetContext(ctx, &ok, q, code, workspaceID); err != nil {
		return false, fmt.Errorf("check category: %w", err)
	}
	return ok, nil
}

// ListCategories returns active system categories plus custom ones for the workspace.
func (r *DocumentRepositoryAdapter) ListCategories(ctx context.Context, workspaceID string) ([]entities.DocumentCategory, error) {
	const q = `
		SELECT code, workspace_id, label_es, sort_order, is_active, is_system, created_at, updated_at
		FROM archive_document_categories
		WHERE is_active = true
		  AND (is_system = true OR workspace_id = $1)
		ORDER BY sort_order ASC, label_es ASC, code ASC`

	var cats []entities.DocumentCategory
	if err := r.db.SelectContext(ctx, &cats, q, workspaceID); err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	return cats, nil
}

// FindCategory loads one active category visible in the workspace.
func (r *DocumentRepositoryAdapter) FindCategory(ctx context.Context, workspaceID, code string) (*entities.DocumentCategory, error) {
	const q = `
		SELECT code, workspace_id, label_es, sort_order, is_active, is_system, created_at, updated_at
		FROM archive_document_categories
		WHERE code = $1 AND is_active = true
		  AND (is_system = true OR workspace_id = $2)`
	var cat entities.DocumentCategory
	if err := r.db.GetContext(ctx, &cat, q, code, workspaceID); err != nil {
		return nil, fmt.Errorf("find category: %w", err)
	}
	return &cat, nil
}

// CreateCategory inserts a custom workspace category.
func (r *DocumentRepositoryAdapter) CreateCategory(ctx context.Context, cat *entities.DocumentCategory) error {
	const q = `
		INSERT INTO archive_document_categories (
			code, workspace_id, label_es, sort_order, is_active, is_system, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.ExecContext(
		ctx, q,
		cat.Code, nullStr(cat.WorkspaceID), cat.LabelES, cat.SortOrder,
		cat.IsActive, cat.IsSystem, cat.CreatedAt, cat.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create category: %w", err)
	}
	return nil
}

// UpdateCategory updates label/sort of a custom category.
func (r *DocumentRepositoryAdapter) UpdateCategory(ctx context.Context, cat *entities.DocumentCategory) error {
	const q = `
		UPDATE archive_document_categories
		SET label_es = $3, sort_order = $4, updated_at = $5
		WHERE code = $1 AND workspace_id = $2 AND is_system = false AND is_active = true`
	res, err := r.db.ExecContext(ctx, q, cat.Code, *cat.WorkspaceID, cat.LabelES, cat.SortOrder, cat.UpdatedAt)
	if err != nil {
		return fmt.Errorf("update category: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("update category: no rows affected")
	}
	return nil
}

// UpdateSystemCategory updates the label of a global system category.
func (r *DocumentRepositoryAdapter) UpdateSystemCategory(ctx context.Context, cat *entities.DocumentCategory) error {
	const q = `
		UPDATE archive_document_categories
		SET label_es = $2, sort_order = $3, updated_at = $4
		WHERE code = $1 AND is_system = true AND workspace_id IS NULL AND is_active = true`
	res, err := r.db.ExecContext(ctx, q, cat.Code, cat.LabelES, cat.SortOrder, cat.UpdatedAt)
	if err != nil {
		return fmt.Errorf("update system category: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("update system category: no rows affected")
	}
	return nil
}

// DeactivateCategory soft-deactivates a custom category.
func (r *DocumentRepositoryAdapter) DeactivateCategory(ctx context.Context, workspaceID, code string) error {
	const q = `
		UPDATE archive_document_categories
		SET is_active = false, updated_at = NOW()
		WHERE code = $1 AND workspace_id = $2 AND is_system = false AND is_active = true`
	res, err := r.db.ExecContext(ctx, q, code, workspaceID)
	if err != nil {
		return fmt.Errorf("deactivate category: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("deactivate category: no rows affected")
	}
	return nil
}

// CountCustomCategories counts active custom categories for a workspace.
func (r *DocumentRepositoryAdapter) CountCustomCategories(ctx context.Context, workspaceID string) (int, error) {
	const q = `
		SELECT COUNT(*) FROM archive_document_categories
		WHERE workspace_id = $1 AND is_system = false AND is_active = true`
	var n int
	if err := r.db.GetContext(ctx, &n, q, workspaceID); err != nil {
		return 0, fmt.Errorf("count custom categories: %w", err)
	}
	return n, nil
}

// CountByCategory returns document counts per category_code for a workspace.
func (r *DocumentRepositoryAdapter) CountByCategory(ctx context.Context, workspaceID string, status string) (map[string]int, error) {
	where := []string{sqlWorkspaceIDEq}
	args := []any{workspaceID}
	argN := 2

	if status == "" {
		status = entities.DocumentStatusActive
	}
	if status != entities.DocumentStatusAll {
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

// CountDueAlerts returns upcoming and expired document counts for a workspace.
func (r *DocumentRepositoryAdapter) CountDueAlerts(ctx context.Context, workspaceID, status string) (upcoming, expired int, err error) {
	where := []string{sqlWorkspaceIDEq, "due_date IS NOT NULL"}
	args := []any{workspaceID}
	argN := 2

	if status == "" {
		status = entities.DocumentStatusActive
	}
	if status != entities.DocumentStatusAll {
		where = append(where, fmt.Sprintf("status = $%d", argN))
		args = append(args, status)
		argN++
	}

	todayIdx := argN
	horizonIdx := argN + 1
	args = append(args,
		time.Now().Format("2006-01-02"),
		time.Now().AddDate(0, 0, entities.DueSoonWindowDays).Format("2006-01-02"),
	)

	q := fmt.Sprintf(`
		SELECT
			COUNT(*) FILTER (WHERE due_date < $%d) AS expired,
			COUNT(*) FILTER (WHERE due_date >= $%d AND due_date <= $%d) AS upcoming
		FROM archive_documents
		WHERE %s`,
		todayIdx, todayIdx, horizonIdx, strings.Join(where, " AND "),
	)

	var row struct {
		Expired  int `db:"expired"`
		Upcoming int `db:"upcoming"`
	}
	if err := r.db.GetContext(ctx, &row, q, args...); err != nil {
		return 0, 0, fmt.Errorf("count due alerts: %w", err)
	}
	return row.Upcoming, row.Expired, nil
}

// CountDueAlertsByCategory returns upcoming/expired due counts keyed by category_code.
func (r *DocumentRepositoryAdapter) CountDueAlertsByCategory(
	ctx context.Context,
	workspaceID, status string,
) (map[string]dtos.CategoryDueAlertCounts, error) {
	where := []string{sqlWorkspaceIDEq, "due_date IS NOT NULL"}
	args := []any{workspaceID}
	argN := 2

	if status == "" {
		status = entities.DocumentStatusActive
	}
	if status != entities.DocumentStatusAll {
		where = append(where, fmt.Sprintf("status = $%d", argN))
		args = append(args, status)
		argN++
	}

	todayIdx := argN
	horizonIdx := argN + 1
	args = append(args,
		time.Now().Format("2006-01-02"),
		time.Now().AddDate(0, 0, entities.DueSoonWindowDays).Format("2006-01-02"),
	)

	q := fmt.Sprintf(`
		SELECT
			category_code,
			COUNT(*) FILTER (WHERE due_date < $%d) AS expired,
			COUNT(*) FILTER (WHERE due_date >= $%d AND due_date <= $%d) AS upcoming
		FROM archive_documents
		WHERE %s
		GROUP BY category_code`,
		todayIdx, todayIdx, horizonIdx, strings.Join(where, " AND "),
	)

	type row struct {
		CategoryCode string `db:"category_code"`
		Expired      int    `db:"expired"`
		Upcoming     int    `db:"upcoming"`
	}
	var rows []row
	if err := r.db.SelectContext(ctx, &rows, q, args...); err != nil {
		return nil, fmt.Errorf("count due alerts by category: %w", err)
	}

	out := make(map[string]dtos.CategoryDueAlertCounts, len(rows))
	for _, row := range rows {
		if row.Expired == 0 && row.Upcoming == 0 {
			continue
		}
		out[row.CategoryCode] = dtos.CategoryDueAlertCounts{
			Upcoming: row.Upcoming,
			Expired:  row.Expired,
		}
	}
	return out, nil
}

func appendSearchQueryFilter(where []string, args []any, argN int, rawQuery string) ([]string, []any, int) {
	tokens := entities.TokenizeSearchQuery(rawQuery)
	if len(tokens) == 0 {
		return where, args, argN
	}
	for _, token := range tokens {
		pattern := "%" + escapeILIKEPattern(token) + "%"
		textMatch := fmt.Sprintf(
			`(title ILIKE $%d ESCAPE '\'
			  OR COALESCE(issuer, '') ILIKE $%d ESCAPE '\'
			  OR COALESCE(reference_number, '') ILIKE $%d ESCAPE '\'
			  OR COALESCE(notes, '') ILIKE $%d ESCAPE '\'
			  OR COALESCE(extra_fields::text, '') ILIKE $%d ESCAPE '\'
			  OR EXISTS (
			    SELECT 1 FROM archive_document_categories c
			    WHERE c.code = archive_documents.category_code
			      AND c.is_active = true
			      AND (c.is_system = true OR c.workspace_id = archive_documents.workspace_id)
			      AND c.label_es ILIKE $%d ESCAPE '\'
			  ))`,
			argN, argN, argN, argN, argN, argN,
		)
		if entities.IsYearToken(token) {
			yearIdx := argN + 1
			where = append(where, fmt.Sprintf(
				`(%s OR to_char(document_date, 'YYYY') = $%d OR to_char(due_date, 'YYYY') = $%d)`,
				textMatch, yearIdx, yearIdx,
			))
			args = append(args, pattern, token)
			argN += 2
			continue
		}
		where = append(where, textMatch)
		args = append(args, pattern)
		argN++
	}
	return where, args, argN
}

func escapeILIKEPattern(s string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(s)
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
