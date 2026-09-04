package storage

import (
	"context"
	"database/sql"

	"github.com/uptrace/bun"
)

// selectRow preserves database/sql's positional scanners and ErrNoRows contract
// while letting Bun build the SELECT. Execution stays lazy until Scan, so there
// is no open cursor if the caller returns before scanning.
func selectRow(ctx context.Context, query *bun.SelectQuery) scanner {
	return bunRow{ctx: ctx, query: query}
}

type bunRow struct {
	ctx   context.Context
	query *bun.SelectQuery
}

func (r bunRow) Scan(dest ...any) error {
	rows, err := r.query.Rows(r.ctx)
	if err != nil {
		return err
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		return sql.ErrNoRows
	}
	return rows.Scan(dest...)
}
