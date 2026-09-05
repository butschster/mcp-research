package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/dovod-app/app/internal/domain"
	"github.com/uptrace/bun"
)

// The reads behind the resume summary.
//
// They live together because they share one contract the ordinary list methods
// do not: each takes a hard LIMIT and each states its ORDER BY in full. A
// summary that asked for everything and sliced in Go would grow with the
// research, which is the cost this feature exists to remove.
//
// Two ordering rules hold throughout:
//
//   - **Never order by a short code.** `E10` sorts before `E2` as a string, so
//     a top-five taken that way is not the five a person would name. Order by
//     the thing the code stands for — a timestamp, a position, a priority.
//   - **Always break the tie on id.** Two rows written in the same second are
//     otherwise returned in whatever order the engine likes, and the same
//     request answers differently on SQLite and PostgreSQL.

// taskPriorityOrder sorts high, medium, low — as an expression rather than a
// stored rank, matching what FindByResearch already does.
const taskPriorityOrder = "CASE priority WHEN 'high' THEN 0 WHEN 'medium' THEN 1 WHEN 'low' THEN 2 ELSE 3 END"

// FindForResume returns the newest-touched tasks in the given statuses.
//
// Recency, not creation order: the question this answers is "where did we stop",
// and a task nobody has touched in a month is not where anybody stopped.
func (r *TaskRepository) FindForResume(ctx context.Context, researchID string, statuses []domain.TaskStatus, limit int) ([]*domain.Task, error) {
	if limit <= 0 || len(statuses) == 0 {
		return nil, nil
	}
	rows, err := r.db.NewSelect().
		ColumnExpr("id, code, research_id, title, description, status, priority, result, created_at, updated_at, completed_at").
		Table("tasks").
		Where("research_id=?", researchID).
		Where("status IN (?)", bun.In(statuses)).
		OrderExpr(taskPriorityOrder).
		OrderExpr("updated_at DESC").
		OrderExpr("id").
		Limit(limit).
		Rows(ctx)
	if err != nil {
		return nil, fmt.Errorf("query resume tasks: %w", err)
	}
	defer rows.Close()

	var out []*domain.Task
	for rows.Next() {
		t, err := r.scanTaskRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// FindForResume returns a session's outstanding questions in interview order.
//
// Position, not recency: questions were asked in an order somebody chose, and
// continuing an interview means taking the next one, not the newest one.
func (r *QuestionRepository) FindForResume(ctx context.Context, sessionID string, statuses []domain.QuestionStatus, limit int) ([]*domain.Question, error) {
	if limit <= 0 || sessionID == "" || len(statuses) == 0 {
		return nil, nil
	}
	rows, err := r.db.NewSelect().
		ColumnExpr("id, code, session_id, text, area, rationale, priority, status, answer, parent_id, position, created_at, updated_at").
		Table("questions").
		Where("session_id=?", sessionID).
		Where("status IN (?)", bun.In(statuses)).
		OrderExpr("position ASC").
		OrderExpr("id").
		Limit(limit).
		Rows(ctx)
	if err != nil {
		return nil, fmt.Errorf("query resume questions: %w", err)
	}
	defer rows.Close()

	var out []*domain.Question
	for rows.Next() {
		q, err := r.scanQuestionRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, q)
	}
	return out, rows.Err()
}

// FindForResume returns the oldest marks in one status, with their entry.
//
// The caller owns the scoping: this filters on `a.research_id` and joins the
// entry without re-checking that the entry belongs to the same research. That
// holds because an annotation takes its research from the entry it is created
// on and nothing ever moves an entry between researches — but a caller passing
// an id it has not authorised gets no protection here.
//
// Oldest first, and deliberately not the entry-grouped order the queue read
// uses: that groups by `e.code` as a string, which is right for working a batch
// document by document and wrong for "what has been waiting longest".
func (r *AnnotationRepository) FindForResume(ctx context.Context, researchID string, status domain.AnnotationStatus, limit int) ([]*domain.Annotation, error) {
	if limit <= 0 {
		return nil, nil
	}
	rows, err := r.db.NewSelect().
		ColumnExpr(prefixColumns("a.", annotationColumns)+", e.code, e.title, e.entry_type").
		TableExpr("annotations a").
		Join("JOIN entries e ON e.id = a.entry_id").
		Where("a.research_id=?", researchID).
		Where("a.status=?", status).
		OrderExpr("a.created_at ASC").
		OrderExpr("a.id").
		Limit(limit).
		Rows(ctx)
	if err != nil {
		return nil, fmt.Errorf("query resume annotations: %w", err)
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

// FindRecentlyUpdated returns the most recently changed entries, without
// content. It is a "what moved while you were away" list, not a change log:
// a deleted document is not here, and two writes in one second are one row.
func (r *EntryRepository) FindRecentlyUpdated(ctx context.Context, researchID string, since time.Time, limit int) ([]*domain.Entry, error) {
	if limit <= 0 {
		return nil, nil
	}
	rows, err := r.db.NewSelect().
		ColumnExpr("id, code, research_id, section_id, session_id, entry_type, title, description, status, tags, metadata, spec_version, created_at, updated_at").
		Table("entries").
		Where("research_id=?", researchID).
		Where("updated_at >= ?", since).
		OrderExpr("updated_at DESC").
		OrderExpr("id").
		Limit(limit).
		Rows(ctx)
	if err != nil {
		return nil, fmt.Errorf("query recent entries: %w", err)
	}
	defer rows.Close()

	var out []*domain.Entry
	for rows.Next() {
		e, err := r.scanEntryRowNoContent(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// CountUpdatedSince counts the entries touched in a window.
//
// Deliberately not a count of all entries: the list it sits beside is "what
// changed", and pairing it with the size of the whole research would label a
// research of two hundred documents "Changed 200" forever.
func (r *EntryRepository) CountUpdatedSince(ctx context.Context, researchID string, since time.Time) (int, error) {
	n, err := r.db.NewSelect().
		Table("entries").
		Where("research_id=?", researchID).
		Where("updated_at >= ?", since).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count recent entries: %w", err)
	}
	return n, nil
}

// LatestRevision is the head of one entry's history: which revision it is at
// and who wrote that revision.
type LatestRevision struct {
	Revision   int
	AuthorKind domain.AuthorKind
}

// LatestByEntries returns the newest revision of each entry, with its author.
//
// It takes ids, not a research, so the caller owns the scoping — pass only ids
// you have already authorised. Every caller today derives them from a
// research-scoped list.
//
// Two queries, both bounded by the caller's limit, rather than one per entry:
// the list this feeds is meant to be cheap, and a round trip per row is how a
// summary ends up costing more than the documents it summarises.
//
// The second query addresses exact (entry_id, revision) pairs with an OR chain
// instead of a window function or a row-value IN. Both of those exist on all
// three engines with different spellings, and this list is at most fifteen
// entries — a portable query beats a clever one at that size.
func (r *EntryRevisionRepository) LatestByEntries(ctx context.Context, entryIDs []string) (map[string]LatestRevision, error) {
	out := map[string]LatestRevision{}
	if len(entryIDs) == 0 {
		return out, nil
	}
	rows, err := r.db.NewSelect().
		ColumnExpr("entry_id, MAX(revision)").
		TableExpr("entry_revisions").
		Where("entry_id IN (?)", bun.In(entryIDs)).
		GroupExpr("entry_id").
		Rows(ctx)
	if err != nil {
		return nil, fmt.Errorf("latest revisions: %w", err)
	}
	for rows.Next() {
		var id string
		var rev int
		if err := rows.Scan(&id, &rev); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan latest revision: %w", err)
		}
		out[id] = LatestRevision{Revision: rev}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()
	if len(out) == 0 {
		return out, nil
	}

	authors := r.db.NewSelect().
		ColumnExpr("entry_id, revision, author_kind").
		TableExpr("entry_revisions")
	authors = authors.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
		for id, head := range out {
			q = q.WhereOr("entry_id=? AND revision=?", id, head.Revision)
		}
		return q
	})
	arows, err := authors.Rows(ctx)
	if err != nil {
		return nil, fmt.Errorf("latest revision authors: %w", err)
	}
	defer arows.Close()

	for arows.Next() {
		var id string
		var rev int
		var kind string
		if err := arows.Scan(&id, &rev, &kind); err != nil {
			return nil, fmt.Errorf("scan revision author: %w", err)
		}
		if head, ok := out[id]; ok && head.Revision == rev {
			head.AuthorKind = domain.AuthorKind(kind)
			out[id] = head
		}
	}
	return out, arows.Err()
}
