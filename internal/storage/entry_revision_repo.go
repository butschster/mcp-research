package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/dovod-app/app/internal/domain"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type EntryRevisionRepository struct {
	db *bun.DB
}

func NewEntryRevisionRepository(db *bun.DB) *EntryRevisionRepository {
	return &EntryRevisionRepository{db: db}
}

// DB exposes the handle so a service can open the transaction that writes an
// entry and its revision together. Nothing else should reach for it.
func (r *EntryRevisionRepository) DB() *bun.DB { return r.db }

// Create appends a revision, numbering it inside the caller's transaction.
//
// The number comes from MAX(revision)+1 read in the same transaction as the
// insert, which is what keeps two concurrent writers from claiming the same one.
// The UNIQUE(entry_id, revision) index is the backstop if that ever stops being
// true.
func (r *EntryRevisionRepository) Create(ctx context.Context, q Querier, rev *domain.EntryRevision) error {
	if q == nil {
		q = r.db
	}
	if rev.ID == "" {
		rev.ID = uuid.New().String()
	}
	if rev.Revision == 0 {
		var last sql.NullInt64
		if err := selectRow(ctx, q.NewSelect().
			ColumnExpr("MAX(revision)").
			TableExpr("entry_revisions").
			Where("entry_id=?", rev.EntryID)).
			Scan(&last); err != nil {
			return fmt.Errorf("next revision: %w", err)
		}
		rev.Revision = int(last.Int64) + 1
	}
	now := time.Now().UTC().Format(time.DateTime)

	_, err := q.NewInsert().Table("entry_revisions").Model(&map[string]any{
		"id":           rev.ID,
		"entry_id":     rev.EntryID,
		"research_id":  rev.ResearchID,
		"revision":     rev.Revision,
		"title":        rev.Title,
		"description":  rev.Description,
		"content":      rev.Content,
		"entry_type":   rev.Type,
		"status":       rev.Status,
		"tags":         marshalJSON(rev.Tags),
		"metadata":     marshalObject(rev.Metadata),
		"spec_version": rev.SpecVersion,
		"author_kind":  rev.AuthorKind,
		"session_id":   nullable(rev.SessionID),
		"user_id":      nullable(rev.UserID),
		"summary":      rev.Summary,
		"created_at":   now,
	}).Exec(ctx)
	if err != nil {
		return fmt.Errorf("insert revision: %w", err)
	}
	rev.CreatedAt, _ = time.Parse(time.DateTime, now)
	return nil
}

// Latest returns the newest revision of an entry, content included, or nil when
// the entry has no history yet.
func (r *EntryRevisionRepository) Latest(ctx context.Context, q Querier, entryID string) (*domain.EntryRevision, error) {
	if q == nil {
		q = r.db
	}
	row := selectRow(ctx, q.NewSelect().
		ColumnExpr(revisionColumns).
		TableExpr("entry_revisions").
		Where("entry_id=?", entryID).
		OrderExpr("revision DESC").
		Limit(1))
	return scanRevisionRow(row, true)
}

// FindByNumber returns one revision with its content.
func (r *EntryRevisionRepository) FindByNumber(ctx context.Context, entryID string, revision int) (*domain.EntryRevision, error) {
	row := selectRow(ctx, r.db.NewSelect().
		ColumnExpr(revisionColumns).
		TableExpr("entry_revisions").
		Where("entry_id=? AND revision=?", entryID, revision))
	return scanRevisionRow(row, true)
}

// ListByEntry returns an entry's history, newest first. Content is left out:
// a history listing of a long document would otherwise be mostly copies of it.
func (r *EntryRevisionRepository) ListByEntry(ctx context.Context, entryID string, limit int) ([]*domain.EntryRevision, error) {
	query := r.db.NewSelect().
		ColumnExpr(revisionColumns).
		Table("entry_revisions").
		Where("entry_id=?", entryID).
		Order("revision DESC")
	if limit > 0 {
		query.Limit(limit)
	}
	rows, err := query.Rows(ctx)
	if err != nil {
		return nil, fmt.Errorf("list revisions: %w", err)
	}
	defer rows.Close()
	return scanRevisions(rows, false)
}

// ListBySession returns every revision written while a session was active,
// oldest first. This is what "what did this session change" is built from.
func (r *EntryRevisionRepository) ListBySession(ctx context.Context, sessionID string) ([]*domain.EntryRevision, error) {
	rows, err := r.db.NewSelect().
		ColumnExpr(revisionColumns).
		TableExpr("entry_revisions").
		Where("session_id=?", sessionID).
		OrderExpr("created_at, revision").
		Rows(ctx)
	if err != nil {
		return nil, fmt.Errorf("list session revisions: %w", err)
	}
	defer rows.Close()
	return scanRevisions(rows, false)
}

// CountByEntry reports how many revisions an entry has.
func (r *EntryRevisionRepository) CountByEntry(ctx context.Context, entryID string) (int, error) {
	var n int
	if err := selectRow(ctx, r.db.NewSelect().
		ColumnExpr("COUNT(*)").
		TableExpr("entry_revisions").
		Where("entry_id=?", entryID)).
		Scan(&n); err != nil {
		return 0, fmt.Errorf("count revisions: %w", err)
	}
	return n, nil
}

// Trim keeps the newest `keep` revisions, revision 1, and every revision that
// is still a reader's "last seen" comparison base.
//
// Revision 1 survives because it is the only record of what the entry looked
// like when it was created, and that is the one snapshot a reader asks for
// months later. Called only when a limit is configured; the default keeps
// everything, which for kilobyte-sized documents is the honest choice.
func (r *EntryRevisionRepository) Trim(ctx context.Context, q Querier, entryID string, keep int) error {
	if keep <= 0 {
		return nil
	}
	if q == nil {
		q = r.db
	}
	newest := q.NewSelect().Column("revision").Table("entry_revisions").Where("entry_id=?", entryID).Order("revision DESC").Limit(keep)
	// The derived table also supports MySQL, which rejects LIMIT directly in
	// an IN subquery and modifying the table selected by an unwrapped subquery.
	retained := q.NewSelect().Column("revision").TableExpr("(?) AS retained", newest)
	seen := q.NewSelect().Column("seen_revision").Table("entry_views").Where("entry_id=?", entryID)
	_, err := q.NewDelete().Table("entry_revisions").Where("entry_id=?", entryID).
		Where("revision <> 1").Where("revision NOT IN (?)", seen).Where("revision NOT IN (?)", retained).Exec(ctx)
	if err != nil {
		return fmt.Errorf("trim revisions: %w", err)
	}
	return nil
}

const revisionColumns = `id, entry_id, research_id, revision, title, description, content, entry_type, status, tags, metadata, spec_version, author_kind, session_id, user_id, summary, created_at`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanRevisionRow(row rowScanner, withContent bool) (*domain.EntryRevision, error) {
	rev, err := scanRevision(row, withContent)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return rev, nil
}

func scanRevisions(rows *sql.Rows, withContent bool) ([]*domain.EntryRevision, error) {
	out := []*domain.EntryRevision{}
	for rows.Next() {
		rev, err := scanRevision(rows, withContent)
		if err != nil {
			return nil, err
		}
		out = append(out, rev)
	}
	return out, rows.Err()
}

func scanRevision(s rowScanner, withContent bool) (*domain.EntryRevision, error) {
	var (
		rev       domain.EntryRevision
		content   string
		tags      sql.NullString
		metadata  sql.NullString
		sessionID sql.NullString
		userID    sql.NullString
		createdAt string
	)
	err := s.Scan(
		&rev.ID, &rev.EntryID, &rev.ResearchID, &rev.Revision,
		&rev.Title, &rev.Description, &content, &rev.Type, &rev.Status,
		&tags, &metadata, &rev.SpecVersion,
		&rev.AuthorKind, &sessionID, &userID, &rev.Summary, &createdAt,
	)
	if err != nil {
		return nil, err
	}
	if withContent {
		rev.Content = content
	}
	rev.Tags = unmarshalStringSlice(tags)
	rev.Metadata = unmarshalObject(metadata)
	rev.SessionID = sessionID.String
	rev.UserID = userID.String
	rev.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
	return &rev, nil
}

func nullable(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
