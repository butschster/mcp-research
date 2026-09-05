package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dovod-app/app/internal/domain"
	"github.com/uptrace/bun"
)

// AnnotationFilter narrows a queue read. Anchor state is deliberately absent:
// it is computed from the document on read, so it cannot be a WHERE clause.
type AnnotationFilter struct {
	Status  *domain.AnnotationStatus
	Kind    *domain.AnnotationKind
	EntryID string
	Limit   int
	Offset  int
}

type AnnotationRepository struct {
	db *bun.DB
}

func NewAnnotationRepository(db *bun.DB) *AnnotationRepository {
	return &AnnotationRepository{db: db}
}

const annotationColumns = `id, code, research_id, entry_id, block_id,
	quote_exact, quote_prefix, quote_suffix, anchored_revision,
	kind, body, author_kind, user_id,
	status, resolution, resolved_revision, session_id, task_id, rejections, attempts,
	created_at, updated_at, answered_at, closed_at`

func (r *AnnotationRepository) Create(ctx context.Context, a *domain.Annotation) error {
	now := time.Now().UTC().Format(time.DateTime)

	if a.Code == "" {
		code, err := NextCode(ctx, r.db, "annotations", "A", "research_id", a.ResearchID)
		if err != nil {
			return fmt.Errorf("generate code: %w", err)
		}
		a.Code = code
	}

	_, err := r.db.NewInsert().Table("annotations").Model(&map[string]any{
		"id":                a.ID,
		"code":              a.Code,
		"research_id":       a.ResearchID,
		"entry_id":          a.EntryID,
		"block_id":          a.BlockID,
		"quote_exact":       a.Quote.Exact,
		"quote_prefix":      a.Quote.Prefix,
		"quote_suffix":      a.Quote.Suffix,
		"anchored_revision": a.AnchoredRevision,
		"kind":              a.Kind,
		"body":              a.Body,
		"author_kind":       a.AuthorKind,
		"user_id":           nullString(a.UserID),
		"status":            a.Status,
		"resolution":        a.Resolution,
		"resolved_revision": a.ResolvedRevision,
		"session_id":        nullString(a.SessionID),
		"task_id":           nullString(a.TaskID),
		"rejections":        marshalRejections(a.Rejections),
		"attempts":          a.Attempts,
		"created_at":        now,
		"updated_at":        now,
		"answered_at":       nil,
		"closed_at":         nil,
	}).Exec(ctx)
	if err != nil {
		return fmt.Errorf("insert annotation: %w", err)
	}
	a.CreatedAt, _ = time.Parse(time.DateTime, now)
	a.UpdatedAt = a.CreatedAt
	return nil
}

// Update writes everything a caller may change. The anchor columns are included
// on purpose: a human who re-points a drifted mark at the sentence it now
// belongs to is re-anchoring it, and that has to persist.
func (r *AnnotationRepository) Update(ctx context.Context, a *domain.Annotation) error {
	now := time.Now().UTC().Format(time.DateTime)

	_, err := r.db.NewUpdate().
		Table("annotations").
		Set("block_id=?", a.BlockID).
		Set("quote_exact=?", a.Quote.Exact).
		Set("quote_prefix=?", a.Quote.Prefix).
		Set("quote_suffix=?", a.Quote.Suffix).
		Set("anchored_revision=?", a.AnchoredRevision).
		Set("kind=?", a.Kind).
		Set("body=?", a.Body).
		Set("status=?", a.Status).
		Set("resolution=?", a.Resolution).
		Set("resolved_revision=?", a.ResolvedRevision).
		Set("session_id=?", nullString(a.SessionID)).
		Set("task_id=?", nullString(a.TaskID)).
		Set("rejections=?", marshalRejections(a.Rejections)).
		Set("attempts=?", a.Attempts).
		Set("updated_at=?", now).
		Set("answered_at=?", nullTime(a.AnsweredAt)).
		Set("closed_at=?", nullTime(a.ClosedAt)).
		Where("id=?", a.ID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("update annotation: %w", err)
	}
	a.UpdatedAt, _ = time.Parse(time.DateTime, now)
	return nil
}

func (r *AnnotationRepository) FindByID(ctx context.Context, id string) (*domain.Annotation, error) {
	row := selectRow(ctx, r.db.NewSelect().ColumnExpr(annotationColumns).TableExpr("annotations").Where("id=?", id))
	return scanAnnotationRow(row)
}

// FindByCode resolves A7 within a research, so [[A7]] and a code typed by a
// person both land somewhere.
func (r *AnnotationRepository) FindByCode(ctx context.Context, researchID, code string) (*domain.Annotation, error) {
	row := selectRow(ctx, r.db.NewSelect().
		ColumnExpr(annotationColumns).
		TableExpr("annotations").
		Where("research_id=? AND code=?", researchID, code))
	return scanAnnotationRow(row)
}

// FindByEntry is the read the entry page makes: everything marked in this
// document, oldest first so the order on screen matches the order of writing.
func (r *AnnotationRepository) FindByEntry(ctx context.Context, entryID string) ([]*domain.Annotation, error) {
	rows, err := r.db.NewSelect().
		ColumnExpr(annotationColumns).
		TableExpr("annotations").
		Where("entry_id=?", entryID).
		OrderExpr("created_at ASC").
		Rows(ctx)
	if err != nil {
		return nil, fmt.Errorf("query annotations: %w", err)
	}
	defer rows.Close()
	return collectAnnotations(rows)
}

// FindByResearch is the queue.
//
// It joins entries for the code and title because every consumer needs them and
// the alternative is a lookup per row on the client. The join is also what keeps
// a mark on a deleted entry from appearing at all — the row is gone by cascade,
// but an inner join makes that structural rather than incidental.
func (r *AnnotationRepository) FindByResearch(ctx context.Context, researchID string, filter AnnotationFilter) ([]*domain.Annotation, error) {
	query := r.db.NewSelect().
		ColumnExpr(prefixColumns("a.", annotationColumns)+", e.code, e.title, e.entry_type").
		TableExpr("annotations a").
		Join("JOIN entries e ON e.id = a.entry_id").
		Where("a.research_id=?", researchID)

	if filter.Status != nil {
		query.Where("a.status=?", *filter.Status)
	}
	if filter.Kind != nil {
		query.Where("a.kind=?", *filter.Kind)
	}
	if filter.EntryID != "" {
		query.Where("a.entry_id=?", filter.EntryID)
	}

	// Grouped by entry, then by age. The grouping is not cosmetic: it is what
	// lets a caller take a batch that is one document rather than fifteen
	// scattered ones, which is the difference between one read of the document
	// and fifteen.
	query.Order("e.code", "a.created_at ASC")

	if filter.Limit > 0 {
		query.Limit(filter.Limit)
		if filter.Offset > 0 {
			query.Offset(filter.Offset)
		}
	}

	rows, err := query.Rows(ctx)
	if err != nil {
		return nil, fmt.Errorf("query annotations: %w", err)
	}
	defer rows.Close()

	var out []*domain.Annotation
	for rows.Next() {
		a, err := scanAnnotationWithEntry(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// CountByStatus backs the badge on the research page.
func (r *AnnotationRepository) CountByStatus(ctx context.Context, researchID string) (map[domain.AnnotationStatus]int, error) {
	rows, err := r.db.NewSelect().
		ColumnExpr("status, COUNT(*)").
		TableExpr("annotations").
		Where("research_id=?", researchID).
		GroupExpr("status").
		Rows(ctx)
	if err != nil {
		return nil, fmt.Errorf("count annotations: %w", err)
	}
	defer rows.Close()

	out := make(map[domain.AnnotationStatus]int)
	for rows.Next() {
		var status domain.AnnotationStatus
		var n int
		if err := rows.Scan(&status, &n); err != nil {
			return nil, fmt.Errorf("scan count: %w", err)
		}
		out[status] = n
	}
	return out, rows.Err()
}

func (r *AnnotationRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().Table("annotations").Where("id=?", id).Exec(ctx)
	return err
}

// prefixColumns qualifies a column list for a join.
func prefixColumns(prefix, columns string) string {
	parts := strings.Split(columns, ",")
	for i, p := range parts {
		parts[i] = prefix + strings.TrimSpace(p)
	}
	return strings.Join(parts, ", ")
}

// marshalRejections stores the list. A malformed value can only come from this
// process, so failing to encode it is a bug rather than a condition — and an
// empty array keeps the column readable by anything that queries it directly.
func marshalRejections(rs []domain.Rejection) string {
	if len(rs) == 0 {
		return "[]"
	}
	b, err := json.Marshal(rs)
	if err != nil {
		return "[]"
	}
	return string(b)
}

func unmarshalRejections(raw string) []domain.Rejection {
	if raw == "" || raw == "[]" {
		return nil
	}
	var rs []domain.Rejection
	if err := json.Unmarshal([]byte(raw), &rs); err != nil {
		return nil
	}
	return rs
}

func nullString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullTime(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.UTC().Format(time.DateTime)
}

type annotationScanner interface {
	Scan(dest ...any) error
}

func scanAnnotationRow(row scanner) (*domain.Annotation, error) {
	a, err := scanAnnotation(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return a, err
}

func collectAnnotations(rows *sql.Rows) ([]*domain.Annotation, error) {
	var out []*domain.Annotation
	for rows.Next() {
		a, err := scanAnnotation(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func scanAnnotation(s annotationScanner) (*domain.Annotation, error) {
	var a domain.Annotation
	var userID, sessionID, taskID sql.NullString
	var resolvedRevision sql.NullInt64
	var createdAt, updatedAt, rejections string
	var answeredAt, closedAt sql.NullString

	err := s.Scan(
		&a.ID, &a.Code, &a.ResearchID, &a.EntryID, &a.BlockID,
		&a.Quote.Exact, &a.Quote.Prefix, &a.Quote.Suffix, &a.AnchoredRevision,
		&a.Kind, &a.Body, &a.AuthorKind, &userID,
		&a.Status, &a.Resolution, &resolvedRevision, &sessionID, &taskID, &rejections, &a.Attempts,
		&createdAt, &updatedAt, &answeredAt, &closedAt,
	)
	if err != nil {
		return nil, err
	}
	fillAnnotation(&a, userID, sessionID, taskID, resolvedRevision, rejections, createdAt, updatedAt, answeredAt, closedAt)
	return &a, nil
}

func scanAnnotationWithEntry(rows *sql.Rows) (*domain.Annotation, error) {
	var a domain.Annotation
	var userID, sessionID, taskID sql.NullString
	var resolvedRevision sql.NullInt64
	var createdAt, updatedAt, rejections string
	var answeredAt, closedAt sql.NullString

	err := rows.Scan(
		&a.ID, &a.Code, &a.ResearchID, &a.EntryID, &a.BlockID,
		&a.Quote.Exact, &a.Quote.Prefix, &a.Quote.Suffix, &a.AnchoredRevision,
		&a.Kind, &a.Body, &a.AuthorKind, &userID,
		&a.Status, &a.Resolution, &resolvedRevision, &sessionID, &taskID, &rejections, &a.Attempts,
		&createdAt, &updatedAt, &answeredAt, &closedAt,
		&a.EntryCode, &a.EntryTitle, &a.EntryType,
	)
	if err != nil {
		return nil, fmt.Errorf("scan annotation: %w", err)
	}
	fillAnnotation(&a, userID, sessionID, taskID, resolvedRevision, rejections, createdAt, updatedAt, answeredAt, closedAt)
	return &a, nil
}

func fillAnnotation(a *domain.Annotation, userID, sessionID, taskID sql.NullString, resolvedRevision sql.NullInt64, rejections, createdAt, updatedAt string, answeredAt, closedAt sql.NullString) {
	a.Rejections = unmarshalRejections(rejections)
	a.UserID = userID.String
	a.SessionID = sessionID.String
	a.TaskID = taskID.String
	if resolvedRevision.Valid {
		n := int(resolvedRevision.Int64)
		a.ResolvedRevision = &n
	}
	a.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
	a.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)
	if answeredAt.Valid {
		t, _ := time.Parse(time.DateTime, answeredAt.String)
		a.AnsweredAt = &t
	}
	if closedAt.Valid {
		t, _ := time.Parse(time.DateTime, closedAt.String)
		a.ClosedAt = &t
	}
}
