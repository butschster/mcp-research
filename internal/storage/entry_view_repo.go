package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/butschster/mcp-research/internal/domain"
)

var ErrEntryViewTargetNotFound = errors.New("entry view target not found")

// EntryViewRepository stores the one mutable fact behind the update queue:
// which numbered snapshot a particular reader has seen.
type EntryViewRepository struct {
	db *sql.DB
}

func NewEntryViewRepository(db *sql.DB) *EntryViewRepository {
	return &EntryViewRepository{db: db}
}

// Find returns the reader's checkpoint for one document, or nil when the
// document is new to them.
func (r *EntryViewRepository) Find(ctx context.Context, viewerID, entryID string) (*domain.EntryView, error) {
	var (
		view   domain.EntryView
		userID sql.NullString
		seenAt string
	)
	err := r.db.QueryRowContext(ctx,
		`SELECT viewer_id, user_id, entry_id, seen_revision, seen_at
		   FROM entry_views WHERE viewer_id=? AND entry_id=?`, viewerID, entryID,
	).Scan(&view.ViewerID, &userID, &view.EntryID, &view.SeenRevision, &seenAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find entry view: %w", err)
	}
	view.UserID = userID.String
	view.SeenAt, _ = time.Parse(time.DateTime, seenAt)
	return &view, nil
}

// ListUpdates returns only documents ahead of this reader's checkpoint. The
// MAX(revision) aggregation is scoped to one research before it joins entries,
// so a queue read does not scan the history of every research on the server.
func (r *EntryViewRepository) ListUpdates(ctx context.Context, viewerID, researchID string) ([]*domain.EntryUpdate, error) {
	rows, err := r.db.QueryContext(ctx, `
		WITH latest AS (
			SELECT entry_id, MAX(revision) AS current_revision
			  FROM entry_revisions
			 WHERE research_id=?
			 GROUP BY entry_id
		)
		SELECT e.id, e.code, e.research_id, e.section_id, e.title,
		       e.description, e.entry_type, e.status,
		       latest.current_revision, COALESCE(v.seen_revision, 0), current_rev.created_at
		  FROM entries e
		  JOIN latest ON latest.entry_id = e.id
		  JOIN entry_revisions current_rev
		    ON current_rev.entry_id=e.id AND current_rev.revision=latest.current_revision
		  LEFT JOIN entry_views v ON v.viewer_id=? AND v.entry_id=e.id
		 WHERE e.research_id=?
		   AND (v.seen_revision IS NULL OR v.seen_revision < latest.current_revision)
		 ORDER BY CASE WHEN v.seen_revision IS NULL THEN 0 ELSE 1 END,
		          current_rev.created_at DESC, e.code`, researchID, viewerID, researchID)
	if err != nil {
		return nil, fmt.Errorf("list entry updates: %w", err)
	}
	defer rows.Close()

	updates := []*domain.EntryUpdate{}
	for rows.Next() {
		var (
			u         domain.EntryUpdate
			updatedAt string
		)
		if err := rows.Scan(
			&u.EntryID, &u.EntryCode, &u.ResearchID, &u.SectionID,
			&u.Title, &u.Description, &u.Type, &u.Status,
			&u.CurrentRevision, &u.SeenRevision, &updatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan entry update: %w", err)
		}
		u.UnseenRevisions = u.CurrentRevision - u.SeenRevision
		if u.SeenRevision == 0 {
			u.Kind = "new"
		} else {
			u.Kind = "changed"
		}
		u.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)
		updates = append(updates, &u)
	}
	return updates, rows.Err()
}

// MarkSeen advances one exact document checkpoint. It succeeds only when the
// requested revision belongs to the named entry and research. The UPSERT never
// moves backwards, so a late request from another tab cannot resurrect an old
// unread range.
func (r *EntryViewRepository) MarkSeen(
	ctx context.Context,
	viewerID, userID, researchID, entryID string,
	revision int,
) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin entry view: %w", err)
	}
	defer tx.Rollback()

	ok, err := markSeen(ctx, tx, viewerID, userID, researchID, entryID, revision)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit entry view: %w", err)
	}
	return true, nil
}

// MarkSeenMany advances a page snapshot as one operation. Every pair is
// validated by the INSERT ... SELECT below; one bad pair rolls the whole batch
// back instead of making "Mark all" only partly true.
func (r *EntryViewRepository) MarkSeenMany(
	ctx context.Context,
	viewerID, userID, researchID string,
	targets []domain.SeenRevision,
) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin entry views: %w", err)
	}
	defer tx.Rollback()

	for _, target := range targets {
		ok, err := markSeen(ctx, tx, viewerID, userID, researchID, target.EntryID, target.Revision)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("entry %s revision %d: %w", target.EntryID, target.Revision, ErrEntryViewTargetNotFound)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit entry views: %w", err)
	}
	return nil
}

func markSeen(
	ctx context.Context,
	tx *sql.Tx,
	viewerID, userID, researchID, entryID string,
	revision int,
) (bool, error) {
	now := time.Now().UTC().Format(time.DateTime)
	var storedUser any
	if userID != "" {
		storedUser = userID
	}
	result, err := tx.ExecContext(ctx, `
		INSERT INTO entry_views (viewer_id, user_id, entry_id, seen_revision, seen_at)
		SELECT ?, ?, er.entry_id, er.revision, ?
		  FROM entry_revisions er
		 WHERE er.research_id=? AND er.entry_id=? AND er.revision=?
		ON CONFLICT(viewer_id, entry_id) DO UPDATE SET
			seen_revision = MAX(entry_views.seen_revision, excluded.seen_revision),
			seen_at = CASE
				WHEN excluded.seen_revision >= entry_views.seen_revision THEN excluded.seen_at
				ELSE entry_views.seen_at
			END`,
		viewerID, storedUser, now, researchID, entryID, revision,
	)
	if err != nil {
		return false, fmt.Errorf("mark entry seen: %w", err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("entry view affected rows: %w", err)
	}
	return n > 0, nil
}
