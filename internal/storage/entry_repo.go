package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/butschster/mcp-research/internal/domain"
)

type EntryFilter struct {
	Status    *domain.EntryStatus
	Tag       string // filter by tag (JSON array contains)
	SessionID string // filter by session
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

	var sessionID *string
	if entry.SessionID != "" {
		sessionID = &entry.SessionID
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO entries (id, code, research_id, section_id, session_id, entry_type, title, content, description, status, tags, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		entry.ID, entry.Code, entry.ResearchID, entry.SectionID, sessionID, entry.Type,
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
		`SELECT id, code, research_id, section_id, session_id, entry_type, title, content, description, status, tags, created_at, updated_at
		 FROM entries WHERE research_id=? AND code=?`, researchID, code)
	return r.scanEntry(row, true)
}

func (r *EntryRepository) Update(ctx context.Context, entry *domain.Entry) error {
	now := time.Now().UTC().Format(time.DateTime)
	var sessionID *string
	if entry.SessionID != "" {
		sessionID = &entry.SessionID
	}

	_, err := r.db.ExecContext(ctx,
		`UPDATE entries SET entry_type=?, title=?, content=?, description=?, status=?, tags=?, code=?, session_id=?, updated_at=?
		 WHERE id=?`,
		entry.Type, entry.Title, entry.Content, entry.Description,
		entry.Status, marshalJSON(entry.Tags), entry.Code, sessionID,
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
		`SELECT id, code, research_id, section_id, session_id, entry_type, title, content, description, status, tags, created_at, updated_at
		 FROM entries WHERE id=?`, id)
	return r.scanEntry(row, true)
}

// SearchEntries performs a full-text search across entry title, description, and content.
// Returns entries without content for efficiency, ordered by relevance (title > description > content).
// SearchEntries matches title, description and content. userID scopes the search
// the way validateResearchAccess scopes everything else: an empty userID means
// auth is off and there is nothing to scope by, and a research with no owner
// stays visible to everyone. Without this the endpoint returned every user's
// entries, and its content LIKE made the search box an oracle over their text.
func (r *EntryRepository) SearchEntries(ctx context.Context, query string, limit int, userID string) ([]*domain.Entry, error) {
	if query == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 20
	}
	pattern := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, code, research_id, section_id, session_id, entry_type, title, description, status, tags, created_at, updated_at,
		        CASE
		          WHEN title LIKE ? THEN 3
		          WHEN description LIKE ? THEN 2
		          WHEN content LIKE ? THEN 1
		          ELSE 0
		        END AS relevance
		 FROM entries
		 WHERE (title LIKE ? OR description LIKE ? OR content LIKE ?)
		   AND (? = '' OR EXISTS (
		         SELECT 1 FROM researches res
		          WHERE res.id = entries.research_id
		            AND (res.user_id IS NULL OR res.user_id = '' OR res.user_id = ?)))
		 ORDER BY relevance DESC, created_at DESC
		 LIMIT ?`,
		pattern, pattern, pattern,
		pattern, pattern, pattern,
		userID, userID,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("search entries: %w", err)
	}
	defer rows.Close()

	var result []*domain.Entry
	for rows.Next() {
		var e domain.Entry
		var tags sql.NullString
		var sessionID sql.NullString
		var createdAt, updatedAt string
		var relevance int
		if err := rows.Scan(
			&e.ID, &e.Code, &e.ResearchID, &e.SectionID, &sessionID, &e.Type,
			&e.Title, &e.Description, &e.Status,
			&tags, &createdAt, &updatedAt, &relevance,
		); err != nil {
			return nil, fmt.Errorf("scan search entry: %w", err)
		}
		e.SessionID = sessionID.String
		e.Tags = unmarshalStringSlice(tags)
		e.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
		e.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)
		result = append(result, &e)
	}
	return result, rows.Err()
}

// FindBySection returns entries without content for token efficiency.
func (r *EntryRepository) FindBySection(ctx context.Context, researchID, sectionID string, filter EntryFilter) ([]*domain.Entry, error) {
	query := `SELECT id, code, research_id, section_id, session_id, entry_type, title, description, status, tags, created_at, updated_at
		 FROM entries WHERE research_id=? AND section_id=?`
	args := []any{researchID, sectionID}

	if filter.Status != nil {
		query += " AND status=?"
		args = append(args, *filter.Status)
	}
	if filter.Tag != "" {
		query += ` AND EXISTS (SELECT 1 FROM json_each(tags) WHERE json_each.value=?)`
		args = append(args, filter.Tag)
	}
	if filter.SessionID != "" {
		query += " AND session_id=?"
		args = append(args, filter.SessionID)
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
	query := `SELECT id, code, research_id, section_id, session_id, entry_type, title, description, status, tags, created_at, updated_at
		 FROM entries WHERE research_id=?`
	args := []any{researchID}

	if filter.Status != nil {
		query += " AND status=?"
		args = append(args, *filter.Status)
	}
	if filter.Tag != "" {
		query += ` AND EXISTS (SELECT 1 FROM json_each(tags) WHERE json_each.value=?)`
		args = append(args, filter.Tag)
	}
	if filter.SessionID != "" {
		query += " AND session_id=?"
		args = append(args, filter.SessionID)
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

// TagCount represents a tag and how many entries have it.
type TagCount struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

// FindTagsByResearch returns all unique tags in a research with their entry counts.
func (r *EntryRepository) FindTagsByResearch(ctx context.Context, researchID string) ([]TagCount, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT json_each.value AS tag, COUNT(*) AS cnt
		 FROM entries, json_each(entries.tags)
		 WHERE entries.research_id=?
		 GROUP BY json_each.value
		 ORDER BY cnt DESC, tag ASC`, researchID)
	if err != nil {
		return nil, fmt.Errorf("query tags: %w", err)
	}
	defer rows.Close()

	var result []TagCount
	for rows.Next() {
		var tc TagCount
		if err := rows.Scan(&tc.Tag, &tc.Count); err != nil {
			return nil, fmt.Errorf("scan tag count: %w", err)
		}
		result = append(result, tc)
	}
	return result, rows.Err()
}

// FindRelatedByTags returns entries that share at least one tag with the given entry,
// excluding the entry itself. Results are ordered by number of shared tags (descending).
func (r *EntryRepository) FindRelatedByTags(ctx context.Context, entryID string, tags []string) ([]*domain.Entry, error) {
	if len(tags) == 0 {
		return nil, nil
	}

	// Build placeholders for tags
	placeholders := ""
	args := make([]any, 0, len(tags)+1)
	for i, t := range tags {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args = append(args, t)
	}
	args = append(args, entryID)

	query := fmt.Sprintf(
		`SELECT e.id, e.code, e.research_id, e.section_id, e.session_id, e.entry_type, e.title, e.description, e.status, e.tags, e.created_at, e.updated_at,
		        COUNT(*) as shared
		 FROM entries e, json_each(e.tags) jt
		 WHERE jt.value IN (%s) AND e.id != ?
		 GROUP BY e.id
		 ORDER BY shared DESC, e.created_at DESC`, placeholders)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query related entries: %w", err)
	}
	defer rows.Close()

	var result []*domain.Entry
	for rows.Next() {
		var e domain.Entry
		var tags sql.NullString
		var sessionID sql.NullString
		var createdAt, updatedAt string
		var shared int
		if err := rows.Scan(
			&e.ID, &e.Code, &e.ResearchID, &e.SectionID, &sessionID, &e.Type,
			&e.Title, &e.Description, &e.Status,
			&tags, &createdAt, &updatedAt, &shared,
		); err != nil {
			return nil, fmt.Errorf("scan related entry: %w", err)
		}
		e.SessionID = sessionID.String
		e.Tags = unmarshalStringSlice(tags)
		e.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
		e.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)
		result = append(result, &e)
	}
	return result, rows.Err()
}

// FindByResearchWithContent returns all entries with content for cross-reference scanning.
func (r *EntryRepository) FindByResearchWithContent(ctx context.Context, researchID string) ([]*domain.Entry, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, code, research_id, section_id, session_id, entry_type, title, content, description, status, tags, created_at, updated_at
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
		var sessionID sql.NullString
		var createdAt, updatedAt string
		err := rows.Scan(
			&e.ID, &e.Code, &e.ResearchID, &e.SectionID, &sessionID, &e.Type,
			&e.Title, &e.Content, &e.Description,
			&e.Status, &tags,
			&createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan entry with content: %w", err)
		}
		e.SessionID = sessionID.String
		e.Tags = unmarshalStringSlice(tags)
		e.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
		e.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)
		result = append(result, &e)
	}
	return result, rows.Err()
}

func (r *EntryRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM entries WHERE id=?", id)
	return err
}

func (r *EntryRepository) CountBySection(ctx context.Context, sectionID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM entries WHERE section_id=?", sectionID).Scan(&count)
	return count, err
}

func (r *EntryRepository) scanEntry(row *sql.Row, withContent bool) (*domain.Entry, error) {
	var e domain.Entry
	var tags sql.NullString
	var sessionID sql.NullString
	var createdAt, updatedAt string

	var err error
	if withContent {
		err = row.Scan(
			&e.ID, &e.Code, &e.ResearchID, &e.SectionID, &sessionID, &e.Type,
			&e.Title, &e.Content, &e.Description,
			&e.Status, &tags,
			&createdAt, &updatedAt,
		)
	} else {
		err = row.Scan(
			&e.ID, &e.Code, &e.ResearchID, &e.SectionID, &sessionID, &e.Type,
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
	e.SessionID = sessionID.String
	e.Tags = unmarshalStringSlice(tags)
	e.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
	e.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)
	return &e, nil
}

func (r *EntryRepository) scanEntryRowNoContent(rows *sql.Rows) (*domain.Entry, error) {
	var e domain.Entry
	var tags sql.NullString
	var sessionID sql.NullString
	var createdAt, updatedAt string

	err := rows.Scan(
		&e.ID, &e.Code, &e.ResearchID, &e.SectionID, &sessionID, &e.Type,
		&e.Title, &e.Description,
		&e.Status, &tags,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan entry row: %w", err)
	}
	e.SessionID = sessionID.String
	e.Tags = unmarshalStringSlice(tags)
	e.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
	e.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)
	return &e, nil
}
