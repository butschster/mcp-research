package storage

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

// Querier lets repository operations use either a database or the caller's
// transaction, including Bun's query builders.
type Querier = bun.IDB

// NextCode generates the next short code for a table with the given prefix, scoped by a parent ID.
func NextCode(ctx context.Context, q Querier, table, prefix, scopeColumn, scopeValue string) (string, error) {
	num, err := reserveCode(ctx, q, table, prefix, scopeColumn, scopeValue)
	if err != nil {
		return "", fmt.Errorf("next code for %s: %w", table, err)
	}
	return fmt.Sprintf("%s%d", prefix, num), nil
}

// NextCodeGlobal generates the next short code for a table with the given prefix, globally scoped.
func NextCodeGlobal(ctx context.Context, q Querier, table, prefix string) (string, error) {
	num, err := reserveCode(ctx, q, table, prefix, "", "")
	if err != nil {
		return "", fmt.Errorf("next code for %s: %w", table, err)
	}
	return fmt.Sprintf("%s%d", prefix, num), nil
}

// Reserve under a row lock (a write lock on SQLite). A counter survives deletes
// and concurrent creates; MAX only seeds it from legacy or explicitly set codes.
func reserveCode(ctx context.Context, db Querier, table, prefix, scopeColumn, scopeValue string) (int, error) {
	key := table + ":" + prefix + ":" + scopeValue
	var next int
	allocate := func(q Querier) error {
		insert := q.NewInsert().Table("storage_counters").Model(&map[string]any{"scope_key": key, "value": 0})
		if _, err := onConflict(insert, q, []string{"scope_key"}).Exec(ctx); err != nil {
			return err
		}
		// A no-op UPDATE acquires the row lock on every dialect before MAX.
		if _, err := q.NewUpdate().Table("storage_counters").Set("value=value").Where("scope_key=?", key).Exec(ctx); err != nil {
			return err
		}
		query := codeQuery(q, table, prefix)
		if scopeColumn != "" {
			query.Where("?=?", bun.Ident(scopeColumn), scopeValue)
		}
		var max int
		if err := query.Scan(ctx, &max); err != nil {
			return err
		}
		if _, err := q.NewUpdate().Table("storage_counters").Set("value=CASE WHEN value < ? THEN ? ELSE value END + 1", max, max).Where("scope_key=?", key).Exec(ctx); err != nil {
			return err
		}
		return q.NewSelect().Column("value").Table("storage_counters").Where("scope_key=?", key).Scan(ctx, &next)
	}
	switch db.(type) {
	case bun.Tx, *bun.Tx:
		err := allocate(db)
		return next, err
	}
	err := db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error { return allocate(tx) })
	return next, err
}

func codeQuery(q Querier, table, prefix string) *bun.SelectQuery {
	castType := "INTEGER"
	if q.Dialect().Name() == dialect.MySQL {
		castType = "SIGNED"
	}
	return q.NewSelect().Table(table).
		ColumnExpr("COALESCE(MAX(CAST(SUBSTR(code, ?) AS "+castType+")), 0)", len(prefix)+1).
		Where("code LIKE ?", prefix+"%")
}

// BackfillCodes generates codes for all records that don't have one yet.
func BackfillCodes(ctx context.Context, db *bun.DB) (int, error) {
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

func backfillGlobal(ctx context.Context, db *bun.DB, table, prefix string) (int, error) {
	// Collect IDs first, then close cursor before updating (avoids MaxOpenConns(1) deadlock)
	rows, err := db.NewSelect().Column("id").Table(table).Where("code = ''").Order("created_at ASC").Rows(ctx)
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
		if _, err := db.NewUpdate().Table(table).Set("code=?", code).Where("id=?", id).Exec(ctx); err != nil {
			return 0, err
		}
	}
	return len(ids), nil
}

func backfillScoped(ctx context.Context, db *bun.DB, table, prefix, scopeColumn string) (int, error) {
	rows, err := db.NewSelect().Column("id", scopeColumn).Table(table).Where("code = ''").Order("created_at ASC").Rows(ctx)
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
		if _, err := db.NewUpdate().Table(table).Set("code=?", code).Where("id=?", i.id).Exec(ctx); err != nil {
			return 0, err
		}
	}
	return len(items), nil
}
