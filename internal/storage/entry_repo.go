package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/dovod-app/app/internal/domain"
	"github.com/uptrace/bun"
)

type EntryFilter struct {
	Status    *domain.EntryStatus
	Tag       string // filter by tag (JSON array contains)
	SessionID string // filter by session
}

type EntryRepository struct {
	db *bun.DB
}

func NewEntryRepository(db *bun.DB) *EntryRepository {
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

	_, err := r.db.NewInsert().Table("entries").Model(&map[string]any{
		"id":           entry.ID,
		"code":         entry.Code,
		"research_id":  entry.ResearchID,
		"section_id":   entry.SectionID,
		"session_id":   sessionID,
		"entry_type":   entry.Type,
		"title":        entry.Title,
		"content":      entry.Content,
		"description":  entry.Description,
		"status":       entry.Status,
		"tags":         marshalJSON(entry.Tags),
		"metadata":     marshalObject(entry.Metadata),
		"spec_version": entry.SpecVersion,
		"created_at":   now,
		"updated_at":   now,
	}).Exec(ctx)
	if err != nil {
		return fmt.Errorf("insert entry: %w", err)
	}
	entry.CreatedAt, _ = time.Parse(time.DateTime, now)
	entry.UpdatedAt = entry.CreatedAt
	return nil
}

func (r *EntryRepository) FindByCode(ctx context.Context, researchID, code string) (*domain.Entry, error) {
	row := selectRow(ctx, r.db.NewSelect().
		ColumnExpr("id, code, research_id, section_id, session_id, entry_type, title, content, description, status, tags, metadata, spec_version, created_at, updated_at").
		TableExpr("entries").
		Where("research_id=? AND code=?", researchID, code))
	return r.scanEntry(row, true)
}

func (r *EntryRepository) Update(ctx context.Context, entry *domain.Entry) error {
	return r.UpdateTx(ctx, nil, entry)
}

// UpdateTx is Update inside a caller's transaction. A block document is written
// as rows and as the projection in entries.content, and those two must land
// together or not at all.
func (r *EntryRepository) UpdateTx(ctx context.Context, q Querier, entry *domain.Entry) error {
	if q == nil {
		q = r.db
	}
	now := time.Now().UTC().Format(time.DateTime)
	var sessionID *string
	if entry.SessionID != "" {
		sessionID = &entry.SessionID
	}

	_, err := q.NewUpdate().
		Table("entries").
		Set("entry_type=?", entry.Type).
		Set("title=?", entry.Title).
		Set("content=?", entry.Content).
		Set("description=?", entry.Description).
		Set("status=?", entry.Status).
		Set("tags=?", marshalJSON(entry.Tags)).
		Set("metadata=?", marshalObject(entry.Metadata)).
		Set("spec_version=?", entry.SpecVersion).
		Set("code=?", entry.Code).
		Set("session_id=?", sessionID).
		Set("updated_at=?", now).
		Where("id=?", entry.ID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("update entry: %w", err)
	}
	entry.UpdatedAt, _ = time.Parse(time.DateTime, now)
	return nil
}

func (r *EntryRepository) FindByID(ctx context.Context, id string) (*domain.Entry, error) {
	return r.FindByIDQuery(ctx, r.db, id)
}

// FindByIDQuery reads an entry through the caller's connection or transaction.
// Snapshot readers use this together with the revision repository so the entry
// projection and its revision number cannot straddle a concurrent commit.
func (r *EntryRepository) FindByIDQuery(ctx context.Context, q Querier, id string) (*domain.Entry, error) {
	row := selectRow(ctx, q.NewSelect().
		ColumnExpr("id, code, research_id, section_id, session_id, entry_type, title, content, description, status, tags, metadata, spec_version, created_at, updated_at").
		TableExpr("entries").
		Where("id=?", id))
	return r.scanEntry(row, true)
}

// SearchEntries matches title, description and content, and returns entries
// without their content, ordered by relevance (title > description > content).
//
// userID scopes it the way Access scopes everything else: by **team
// membership**, not by who created the research. Scoping by creator was both
// too narrow — a colleague's entries in a shared team never appeared, which is
// the point of having one — and too wide, because an ownerless research
// matched everybody while `research.Get` refused it, turning the search box
// into an oracle over text its reader could not open.
//
// An empty userID means auth is off and there is nothing to scope by.
// SearchEntries finds entries whose title, description or content matches.
//
// researchID scopes the search to one research. It is optional because the
// command palette searches everything, and it is not optional in practice for
// anyone reading a sixty-entry research: tags are the only in-research filter
// and they exist only if the agent applied them.
func (r *EntryRepository) SearchEntries(ctx context.Context, query string, limit int, userID string, researchID string) ([]*domain.Entry, error) {
	if query == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 20
	}
	pattern := "%" + query + "%"
	rows, err := r.db.NewSelect().
		ColumnExpr("id, code, research_id, section_id, session_id, entry_type, title, description, status, tags, metadata, spec_version, created_at, updated_at, CASE WHEN LOWER(title) LIKE LOWER(?) THEN 3 WHEN LOWER(description) LIKE LOWER(?) THEN 2 WHEN LOWER(content) LIKE LOWER(?) THEN 1 ELSE 0 END AS relevance", pattern, pattern, pattern).
		TableExpr("entries").
		Where("(LOWER(title) LIKE LOWER(?) OR LOWER(description) LIKE LOWER(?) OR LOWER(content) LIKE LOWER(?)) AND (? = '' OR research_id = ?) AND (? = '' OR EXISTS ( SELECT 1 FROM researches res WHERE res.id = entries.research_id AND (res.team_id = 'team-local' OR res.team_id IN (SELECT team_id FROM team_members WHERE user_id = ?))))", pattern, pattern, pattern, researchID, researchID, userID, userID).
		OrderExpr("relevance DESC, created_at DESC").
		Limit(limit).
		Rows(ctx)
	if err != nil {
		return nil, fmt.Errorf("search entries: %w", err)
	}
	defer rows.Close()

	var result []*domain.Entry
	for rows.Next() {
		var e domain.Entry
		var tags sql.NullString
		var metadata sql.NullString
		var sessionID sql.NullString
		var createdAt, updatedAt string
		var relevance int
		if err := rows.Scan(
			&e.ID, &e.Code, &e.ResearchID, &e.SectionID, &sessionID, &e.Type,
			&e.Title, &e.Description, &e.Status,
			&tags, &metadata, &e.SpecVersion, &createdAt, &updatedAt, &relevance,
		); err != nil {
			return nil, fmt.Errorf("scan search entry: %w", err)
		}
		e.SessionID = sessionID.String
		e.Tags = unmarshalStringSlice(tags)
		e.Metadata = unmarshalObject(metadata)
		e.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
		e.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)
		result = append(result, &e)
	}
	return result, rows.Err()
}

// FindBySection returns entries without content for token efficiency.
func (r *EntryRepository) FindBySection(ctx context.Context, researchID, sectionID string, filter EntryFilter) ([]*domain.Entry, error) {
	query := r.filteredEntries(researchID, filter).Where("section_id=?", sectionID)
	rows, err := query.Rows(ctx)
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
	rows, err := r.filteredEntries(researchID, filter).Rows(ctx)
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

func (r *EntryRepository) filteredEntries(researchID string, filter EntryFilter) *bun.SelectQuery {
	query := r.db.NewSelect().
		ColumnExpr("id, code, research_id, section_id, session_id, entry_type, title, description, status, tags, metadata, spec_version, created_at, updated_at").
		Table("entries").
		Where("research_id=?", researchID).
		Order("created_at DESC")
	if filter.Status != nil {
		query.Where("status=?", *filter.Status)
	}
	if filter.SessionID != "" {
		query.Where("session_id=?", filter.SessionID)
	}
	if filter.Tag != "" {
		query.Where("EXISTS (?)", r.db.NewSelect().
			ColumnExpr("1").
			TableExpr("?", tagValues(r.db, "entries.tags")).
			Where("jt.value=?", filter.Tag))
	}
	return query
}

// TagCount represents a tag and how many entries have it.
type TagCount struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

// FindTagsByResearch returns all unique tags in a research with their entry counts.
func (r *EntryRepository) FindTagsByResearch(ctx context.Context, researchID string) ([]TagCount, error) {
	rows, err := r.db.NewSelect().
		ColumnExpr("jt.value AS tag, COUNT(*) AS cnt").
		Table("entries").
		TableExpr("?", tagValues(r.db, "entries.tags")).
		Where("entries.research_id=?", researchID).
		GroupExpr("jt.value").
		OrderExpr("cnt DESC, tag ASC").
		Rows(ctx)
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

// FindRelatedByTags returns entries sharing at least one tag with the given
// entry, excluding it, ordered by how many tags they share.
//
// userID scopes it exactly as SearchEntries does, and for the same reason it
// had to stop scoping by creator: a teammate saw none of a colleague's related
// entries, while someone removed from a team went on seeing every entry they
// had created there.
// FindRelatedByTags finds entries sharing tags with one the caller is holding.
//
// researchID confines the answer to a single research. It is empty for an
// ordinary reader, whose membership does the scoping — and it is not empty for
// a share visitor, who has no membership at all: the `? = ”` clause below
// disables the team filter for a caller with no user id, which for a share
// would have made this endpoint return every tagged entry on the server.
func (r *EntryRepository) FindRelatedByTags(ctx context.Context, entryID string, tags []string, userID, researchID string) ([]*domain.Entry, error) {
	if len(tags) == 0 {
		return nil, nil
	}

	query := r.db.NewSelect().
		ColumnExpr("e.id, e.code, e.research_id, e.section_id, e.session_id, e.entry_type, e.title, e.description, e.status, e.tags, e.created_at, e.updated_at, COUNT(*) AS shared").
		TableExpr("entries e").
		TableExpr("?", tagValues(r.db, "e.tags")).
		Where("jt.value IN (?)", bun.In(tags)).
		Where("e.id != ?", entryID)
	if researchID != "" {
		query.Where("e.research_id=?", researchID)
	}
	if userID != "" {
		query.Where("EXISTS (?)", r.db.NewSelect().
			ColumnExpr("1").
			TableExpr("researches res").
			Where("res.id=e.research_id").
			Where("(res.team_id='team-local' OR res.team_id IN (?))", r.db.NewSelect().
				Column("team_id").
				Table("team_members").
				Where("user_id=?", userID)))
	}
	rows, err := query.GroupExpr("e.id, e.code, e.research_id, e.section_id, e.session_id, e.entry_type, e.title, e.description, e.status, e.tags, e.created_at, e.updated_at").
		OrderExpr("shared DESC, e.created_at DESC").
		Rows(ctx)
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
	rows, err := r.db.NewSelect().
		ColumnExpr("id, code, research_id, section_id, session_id, entry_type, title, content, description, status, tags, metadata, spec_version, created_at, updated_at").
		TableExpr("entries").
		Where("research_id=?", researchID).
		OrderExpr("created_at").
		Rows(ctx)
	if err != nil {
		return nil, fmt.Errorf("query entries with content: %w", err)
	}
	defer rows.Close()

	var result []*domain.Entry
	for rows.Next() {
		var e domain.Entry
		var tags sql.NullString
		var metadata sql.NullString
		var sessionID sql.NullString
		var createdAt, updatedAt string
		err := rows.Scan(
			&e.ID, &e.Code, &e.ResearchID, &e.SectionID, &sessionID, &e.Type,
			&e.Title, &e.Content, &e.Description,
			&e.Status, &tags, &metadata, &e.SpecVersion,
			&createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan entry with content: %w", err)
		}
		e.SessionID = sessionID.String
		e.Tags = unmarshalStringSlice(tags)
		e.Metadata = unmarshalObject(metadata)
		e.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
		e.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)
		result = append(result, &e)
	}
	return result, rows.Err()
}

func (r *EntryRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().Table("entries").Where("id=?", id).Exec(ctx)
	return err
}

func (r *EntryRepository) CountBySection(ctx context.Context, sectionID string) (int, error) {
	var count int
	err := selectRow(ctx, r.db.NewSelect().
		ColumnExpr("COUNT(*)").
		TableExpr("entries").
		Where("section_id=?", sectionID)).
		Scan(&count)
	return count, err
}

func (r *EntryRepository) scanEntry(row scanner, withContent bool) (*domain.Entry, error) {
	var e domain.Entry
	var tags sql.NullString
	var metadata sql.NullString
	var sessionID sql.NullString
	var createdAt, updatedAt string

	var err error
	if withContent {
		err = row.Scan(
			&e.ID, &e.Code, &e.ResearchID, &e.SectionID, &sessionID, &e.Type,
			&e.Title, &e.Content, &e.Description,
			&e.Status, &tags, &metadata, &e.SpecVersion,
			&createdAt, &updatedAt,
		)
	} else {
		err = row.Scan(
			&e.ID, &e.Code, &e.ResearchID, &e.SectionID, &sessionID, &e.Type,
			&e.Title, &e.Description,
			&e.Status, &tags, &metadata, &e.SpecVersion,
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
	e.Metadata = unmarshalObject(metadata)
	e.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
	e.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)
	return &e, nil
}

func (r *EntryRepository) scanEntryRowNoContent(rows *sql.Rows) (*domain.Entry, error) {
	var e domain.Entry
	var tags sql.NullString
	var metadata sql.NullString
	var sessionID sql.NullString
	var createdAt, updatedAt string

	err := rows.Scan(
		&e.ID, &e.Code, &e.ResearchID, &e.SectionID, &sessionID, &e.Type,
		&e.Title, &e.Description,
		&e.Status, &tags, &metadata, &e.SpecVersion,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan entry row: %w", err)
	}
	e.SessionID = sessionID.String
	e.Tags = unmarshalStringSlice(tags)
	e.Metadata = unmarshalObject(metadata)
	e.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
	e.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)
	return &e, nil
}
