package storage

import (
	"context"
	"database/sql"
	"strings"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
	"github.com/uptrace/bun/schema"
)

// Bun has no cross-dialect INSERT SELECT upsert API for this monotonic update.
// Keep this small SQL adapter here; the source query and its values use Bun.
func advanceCheckpoint(ctx context.Context, db Querier, source *bun.SelectQuery) (sql.Result, error) {
	conflict := `ON CONFLICT(viewer_id, entry_id) DO UPDATE SET
		seen_at = CASE WHEN excluded.seen_revision >= entry_views.seen_revision THEN excluded.seen_at ELSE entry_views.seen_at END,
		seen_revision = CASE WHEN excluded.seen_revision > entry_views.seen_revision THEN excluded.seen_revision ELSE entry_views.seen_revision END`
	if db.Dialect().Name() == dialect.MySQL {
		// MySQL evaluates assignments left to right. Compare against the OLD
		// revision before advancing it, so delayed requests retain seen_at.
		conflict = `ON DUPLICATE KEY UPDATE
			seen_at = CASE WHEN VALUES(seen_revision) >= seen_revision THEN VALUES(seen_at) ELSE seen_at END,
			seen_revision = GREATEST(seen_revision, VALUES(seen_revision))`
	}
	return db.NewRaw("INSERT INTO entry_views (viewer_id, user_id, entry_id, seen_revision, seen_at) ? "+conflict, source).Exec(ctx)
}

func portableProjection(db Querier, expr string) string {
	switch db.Dialect().Name() {
	case dialect.PG:
		return strings.ReplaceAll(expr, "char(10)", "chr(10)")
	case dialect.MySQL:
		return strings.ReplaceAll(expr, "(LENGTH(s.body)+3)/4", "(CHAR_LENGTH(s.body)+3) DIV 4")
	default:
		return expr
	}
}

// tagValues exposes a JSON array as a one-column table named jt. Keeping the
// dialect here makes filtering, related documents and counts use one contract.
func tagValues(db Querier, column string) schema.QueryAppender {
	switch db.Dialect().Name() {
	case dialect.PG:
		return bun.SafeQuery("jsonb_array_elements_text(COALESCE(NULLIF(?, 'null'), '[]')::jsonb) AS jt(value)", bun.Ident(column))
	case dialect.MySQL:
		return bun.SafeQuery("JSON_TABLE(COALESCE(NULLIF(?, 'null'), '[]'), '$[*]' COLUMNS (value TEXT PATH '$')) AS jt", bun.Ident(column))
	default:
		return bun.SafeQuery("json_each(?) AS jt", bun.Ident(column))
	}
}

// onConflict targets the supplied key on SQLite/PostgreSQL. MySQL handles any
// duplicate unique key, so use this only when those keys identify the same row.
// Unlike INSERT IGNORE, a no-op update does not hide other invalid data.
func onConflict(q *bun.InsertQuery, db Querier, keys []string, updates ...string) *bun.InsertQuery {
	if db.Dialect().Name() == dialect.MySQL {
		q.On("DUPLICATE KEY UPDATE")
		if len(updates) == 0 {
			return q.Set("? = ?", bun.Ident(keys[0]), bun.Ident(keys[0]))
		}
		for _, col := range updates {
			q.Set("? = VALUES(?)", bun.Ident(col), bun.Ident(col))
		}
		return q
	}
	keySQL := strings.Join(keys, ", ") // fixed schema identifiers, never request input
	if len(updates) == 0 {
		return q.On("CONFLICT (" + keySQL + ") DO NOTHING")
	}
	q.On("CONFLICT (" + keySQL + ") DO UPDATE")
	for _, col := range updates {
		q.Set("? = ?", bun.Ident(col), bun.Ident("excluded."+col))
	}
	return q
}
