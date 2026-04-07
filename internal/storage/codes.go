package storage

import (
	"context"
	"database/sql"
	"fmt"
)

// Querier is an interface for *sql.DB and *sql.Tx.
type Querier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// NextCode generates the next short code for a table with the given prefix, scoped by a parent ID.
func NextCode(ctx context.Context, q Querier, table, prefix, scopeColumn, scopeValue string) (string, error) {
	var maxNum int
	query := fmt.Sprintf(
		`SELECT COALESCE(MAX(CAST(SUBSTR(code, %d) AS INTEGER)), 0) FROM %s WHERE %s=? AND code LIKE '%s%%'`,
		len(prefix)+1, table, scopeColumn, prefix,
	)
	err := q.QueryRowContext(ctx, query, scopeValue).Scan(&maxNum)
	if err != nil {
		return "", fmt.Errorf("next code for %s: %w", table, err)
	}
	return fmt.Sprintf("%s%d", prefix, maxNum+1), nil
}

// NextCodeGlobal generates the next short code for a table with the given prefix, globally scoped.
func NextCodeGlobal(ctx context.Context, q Querier, table, prefix string) (string, error) {
	var maxNum int
	query := fmt.Sprintf(
		`SELECT COALESCE(MAX(CAST(SUBSTR(code, %d) AS INTEGER)), 0) FROM %s WHERE code LIKE '%s%%'`,
		len(prefix)+1, table, prefix,
	)
	err := q.QueryRowContext(ctx, query).Scan(&maxNum)
	if err != nil {
		return "", fmt.Errorf("next code for %s: %w", table, err)
	}
	return fmt.Sprintf("%s%d", prefix, maxNum+1), nil
}

// BackfillCodes generates codes for all records that don't have one yet.
func BackfillCodes(ctx context.Context, db *sql.DB) (int, error) {
	total := 0

	n, err := backfillGlobal(ctx, db, "researches", "R")
	if err != nil {
		return total, fmt.Errorf("backfill researches: %w", err)
	}
	total += n

	n, err = backfillScoped(ctx, db, "sections", "S", "research_id")
	if err != nil {
		return total, fmt.Errorf("backfill sections: %w", err)
	}
	total += n

	n, err = backfillScoped(ctx, db, "entries", "E", "research_id")
	if err != nil {
		return total, fmt.Errorf("backfill entries: %w", err)
	}
	total += n

	n, err = backfillScoped(ctx, db, "sessions", "SS", "research_id")
	if err != nil {
		return total, fmt.Errorf("backfill sessions: %w", err)
	}
	total += n

	n, err = backfillScoped(ctx, db, "questions", "Q", "session_id")
	if err != nil {
		return total, fmt.Errorf("backfill questions: %w", err)
	}
	total += n

	n, err = backfillScoped(ctx, db, "tasks", "T", "research_id")
	if err != nil {
		return total, fmt.Errorf("backfill tasks: %w", err)
	}
	total += n

	n, err = backfillScoped(ctx, db, "roadmaps", "RM", "research_id")
	if err != nil {
		return total, fmt.Errorf("backfill roadmaps: %w", err)
	}
	total += n

	n, err = backfillScoped(ctx, db, "roadmap_nodes", "N", "roadmap_id")
	if err != nil {
		return total, fmt.Errorf("backfill roadmap nodes: %w", err)
	}
	total += n

	return total, nil
}

func backfillGlobal(ctx context.Context, db *sql.DB, table, prefix string) (int, error) {
	// Collect IDs first, then close cursor before updating (avoids MaxOpenConns(1) deadlock)
	rows, err := db.QueryContext(ctx,
		fmt.Sprintf(`SELECT id FROM %s WHERE code = '' ORDER BY created_at ASC`, table))
	if err != nil {
		return 0, err
	}
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return 0, err
		}
		ids = append(ids, id)
	}
	rows.Close()

	for _, id := range ids {
		code, err := NextCodeGlobal(ctx, db, table, prefix)
		if err != nil {
			return 0, err
		}
		if _, err := db.ExecContext(ctx,
			fmt.Sprintf(`UPDATE %s SET code=? WHERE id=?`, table), code, id); err != nil {
			return 0, err
		}
	}
	return len(ids), nil
}

func backfillScoped(ctx context.Context, db *sql.DB, table, prefix, scopeColumn string) (int, error) {
	rows, err := db.QueryContext(ctx,
		fmt.Sprintf(`SELECT id, %s FROM %s WHERE code = '' ORDER BY created_at ASC`, scopeColumn, table))
	if err != nil {
		return 0, err
	}
	type item struct{ id, scope string }
	var items []item
	for rows.Next() {
		var i item
		if err := rows.Scan(&i.id, &i.scope); err != nil {
			rows.Close()
			return 0, err
		}
		items = append(items, i)
	}
	rows.Close()

	for _, i := range items {
		code, err := NextCode(ctx, db, table, prefix, scopeColumn, i.scope)
		if err != nil {
			return 0, err
		}
		if _, err := db.ExecContext(ctx,
			fmt.Sprintf(`UPDATE %s SET code=? WHERE id=?`, table), code, i.id); err != nil {
			return 0, err
		}
	}
	return len(items), nil
}
