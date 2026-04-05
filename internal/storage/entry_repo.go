package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/butschster/mcp-research/internal/domain"
)

type EntryFilter struct {
	Status *domain.EntryStatus
}

type EntryRepository struct {
	db *sql.DB
}

func NewEntryRepository(db *sql.DB) *EntryRepository {
	return &EntryRepository{db: db}
}

func (r *EntryRepository) Create(ctx context.Context, entry *domain.Entry) error {
	now := time.Now().UTC().Format(time.DateTime)

	// Auto-assign short code within the research
	if entry.Code == "" {
		code, err := NextCode(ctx, r.db, "entries", "E", "research_id", entry.ResearchID)
		if err != nil {
			return fmt.Errorf("generate code: %w", err)
		}
		entry.Code = code
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO entries (id, code, research_id, section_id, title, content, description, status, tags, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		entry.ID, entry.Code, entry.ResearchID, entry.SectionID,
		entry.Title, entry.Content, entry.Description,
		entry.Status, marshalJSON(entry.Tags),
		now, now,
	)
	if err != nil {
		return fmt.Errorf("insert entry: %w", err)
	}
	entry.CreatedAt, _ = time.Parse(time.DateTime, now)
	entry.UpdatedAt = entry.CreatedAt
	return nil
}

func (r *EntryRepository) FindByCode(ctx context.Context, researchID, code string) (*domain.Entry, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, code, research_id, section_id, title, content, description, status, tags, created_at, updated_at
		 FROM entries WHERE research_id=? AND code=?`, researchID, code)
	return r.scanEntry(row, true)
}

func (r *EntryRepository) Update(ctx context.Context, entry *domain.Entry) error {
	now := time.Now().UTC().Format(time.DateTime)
	_, err := r.db.ExecContext(ctx,
		`UPDATE entries SET title=?, content=?, description=?, status=?, tags=?, code=?, updated_at=?
		 WHERE id=?`,
		entry.Title, entry.Content, entry.Description,
		entry.Status, marshalJSON(entry.Tags), entry.Code,
		now, entry.ID,
	)
	if err != nil {
		return fmt.Errorf("update entry: %w", err)
	}
	entry.UpdatedAt, _ = time.Parse(time.DateTime, now)
	return nil
}

func (r *EntryRepository) FindByID(ctx context.Context, id string) (*domain.Entry, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, code, research_id, section_id, title, content, description, status, tags, created_at, updated_at
		 FROM entries WHERE id=?`, id)
	return r.scanEntry(row, true)
}

// FindBySection returns entries without content for token efficiency.
func (r *EntryRepository) FindBySection(ctx context.Context, researchID, sectionID string, filter EntryFilter) ([]*domain.Entry, error) {
	query := `SELECT id, code, research_id, section_id, title, description, status, tags, created_at, updated_at
		 FROM entries WHERE research_id=? AND section_id=?`
	args := []any{researchID, sectionID}

	if filter.Status != nil {
		query += " AND status=?"
		args = append(args, *filter.Status)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query entries: %w", err)
	}
	defer rows.Close()

	var result []*domain.Entry
	for rows.Next() {
		e, err := r.scanEntryRowNoContent(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, rows.Err()
}

// FindByResearch returns entries without content for token efficiency.
func (r *EntryRepository) FindByResearch(ctx context.Context, researchID string, filter EntryFilter) ([]*domain.Entry, error) {
	query := `SELECT id, code, research_id, section_id, title, description, status, tags, created_at, updated_at
		 FROM entries WHERE research_id=?`
	args := []any{researchID}

	if filter.Status != nil {
		query += " AND status=?"
		args = append(args, *filter.Status)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query entries: %w", err)
	}
	defer rows.Close()

	var result []*domain.Entry
	for rows.Next() {
		e, err := r.scanEntryRowNoContent(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, rows.Err()
}

// FindByResearchWithContent returns all entries with content for cross-reference scanning.
func (r *EntryRepository) FindByResearchWithContent(ctx context.Context, researchID string) ([]*domain.Entry, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, code, research_id, section_id, title, content, description, status, tags, created_at, updated_at
		 FROM entries WHERE research_id=? ORDER BY created_at`,
		researchID,
	)
	if err != nil {
		return nil, fmt.Errorf("query entries with content: %w", err)
	}
	defer rows.Close()

	var result []*domain.Entry
	for rows.Next() {
		var e domain.Entry
		var tags sql.NullString
		var createdAt, updatedAt string
		err := rows.Scan(
			&e.ID, &e.Code, &e.ResearchID, &e.SectionID,
			&e.Title, &e.Content, &e.Description,
			&e.Status, &tags,
			&createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan entry with content: %w", err)
		}
		e.Tags = unmarshalStringSlice(tags)
		e.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
		e.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)
		result = append(result, &e)
	}
	return result, rows.Err()
}

func (r *EntryRepository) CountBySection(ctx context.Context, sectionID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM entries WHERE section_id=?", sectionID).Scan(&count)
	return count, err
}

func (r *EntryRepository) scanEntry(row *sql.Row, withContent bool) (*domain.Entry, error) {
	var e domain.Entry
	var tags sql.NullString
	var createdAt, updatedAt string

	var err error
	if withContent {
		err = row.Scan(
			&e.ID, &e.Code, &e.ResearchID, &e.SectionID,
			&e.Title, &e.Content, &e.Description,
			&e.Status, &tags,
			&createdAt, &updatedAt,
		)
	} else {
		err = row.Scan(
			&e.ID, &e.Code, &e.ResearchID, &e.SectionID,
			&e.Title, &e.Description,
			&e.Status, &tags,
			&createdAt, &updatedAt,
		)
	}

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan entry: %w", err)
	}
	e.Tags = unmarshalStringSlice(tags)
	e.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
	e.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)
	return &e, nil
}

func (r *EntryRepository) scanEntryRowNoContent(rows *sql.Rows) (*domain.Entry, error) {
	var e domain.Entry
	var tags sql.NullString
	var createdAt, updatedAt string

	err := rows.Scan(
		&e.ID, &e.Code, &e.ResearchID, &e.SectionID,
		&e.Title, &e.Description,
		&e.Status, &tags,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan entry row: %w", err)
	}
	e.Tags = unmarshalStringSlice(tags)
	e.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
	e.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)
	return &e, nil
}
