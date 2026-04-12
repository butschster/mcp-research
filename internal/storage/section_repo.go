package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/butschster/mcp-research/internal/domain"
)

type SectionRepository struct {
	db *sql.DB
}

func NewSectionRepository(db *sql.DB) *SectionRepository {
	return &SectionRepository{db: db}
}

func (r *SectionRepository) Create(ctx context.Context, section *domain.Section) error {
	now := time.Now().UTC().Format(time.DateTime)

	if section.Code == "" {
		code, err := NextCode(ctx, r.db, "sections", "S", "research_id", section.ResearchID)
		if err != nil {
			return fmt.Errorf("generate code: %w", err)
		}
		section.Code = code
	}

	statusesJSON := marshalEntryStatuses(section.AllowedEntryStatuses)

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO sections (id, code, research_id, name, display_name, description, status, position, allowed_entry_statuses, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		section.ID, section.Code, section.ResearchID, section.Name, section.DisplayName,
		section.Description, section.Status, section.Position, statusesJSON,
		now, now,
	)
	if err != nil {
		return fmt.Errorf("insert section: %w", err)
	}
	section.CreatedAt, _ = time.Parse(time.DateTime, now)
	section.UpdatedAt = section.CreatedAt
	return nil
}

func (r *SectionRepository) CreateTx(ctx context.Context, tx *sql.Tx, section *domain.Section) error {
	now := time.Now().UTC().Format(time.DateTime)

	if section.Code == "" {
		code, err := NextCode(ctx, tx, "sections", "S", "research_id", section.ResearchID)
		if err != nil {
			return fmt.Errorf("generate code: %w", err)
		}
		section.Code = code
	}

	statusesJSON := marshalEntryStatuses(section.AllowedEntryStatuses)

	_, err := tx.ExecContext(ctx,
		`INSERT INTO sections (id, code, research_id, name, display_name, description, status, position, allowed_entry_statuses, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		section.ID, section.Code, section.ResearchID, section.Name, section.DisplayName,
		section.Description, section.Status, section.Position, statusesJSON,
		now, now,
	)
	if err != nil {
		return fmt.Errorf("insert section: %w", err)
	}
	section.CreatedAt, _ = time.Parse(time.DateTime, now)
	section.UpdatedAt = section.CreatedAt
	return nil
}

func (r *SectionRepository) Update(ctx context.Context, section *domain.Section) error {
	now := time.Now().UTC().Format(time.DateTime)
	statusesJSON := marshalEntryStatuses(section.AllowedEntryStatuses)
	_, err := r.db.ExecContext(ctx,
		`UPDATE sections SET name=?, display_name=?, description=?, status=?, position=?, code=?, allowed_entry_statuses=?, updated_at=?
		 WHERE id=?`,
		section.Name, section.DisplayName, section.Description,
		section.Status, section.Position, section.Code, statusesJSON, now, section.ID,
	)
	if err != nil {
		return fmt.Errorf("update section: %w", err)
	}
	section.UpdatedAt, _ = time.Parse(time.DateTime, now)
	return nil
}

func (r *SectionRepository) FindByID(ctx context.Context, id string) (*domain.Section, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, code, research_id, name, display_name, description, status, position, allowed_entry_statuses, created_at, updated_at
		 FROM sections WHERE id=?`, id)
	return r.scanSection(row)
}

func (r *SectionRepository) FindByResearch(ctx context.Context, researchID string) ([]*domain.Section, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, code, research_id, name, display_name, description, status, position, allowed_entry_statuses, created_at, updated_at
		 FROM sections WHERE research_id=? ORDER BY position ASC`, researchID)
	if err != nil {
		return nil, fmt.Errorf("query sections: %w", err)
	}
	defer rows.Close()

	var result []*domain.Section
	for rows.Next() {
		s, err := r.scanSectionRow(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, rows.Err()
}

func (r *SectionRepository) CountEntriesBySection(ctx context.Context, sectionID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM entries WHERE section_id=?", sectionID).Scan(&count)
	return count, err
}

func (r *SectionRepository) FindByResearchAndName(ctx context.Context, researchID, name string) (*domain.Section, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, code, research_id, name, display_name, description, status, position, allowed_entry_statuses, created_at, updated_at
		 FROM sections WHERE research_id=? AND name=?`, researchID, name)
	return r.scanSection(row)
}

func (r *SectionRepository) scanSection(row *sql.Row) (*domain.Section, error) {
	var s domain.Section
	var createdAt, updatedAt string
	var statusesJSON sql.NullString
	err := row.Scan(
		&s.ID, &s.Code, &s.ResearchID, &s.Name, &s.DisplayName,
		&s.Description, &s.Status, &s.Position, &statusesJSON,
		&createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan section: %w", err)
	}
	s.AllowedEntryStatuses = unmarshalEntryStatuses(statusesJSON)
	s.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
	s.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)
	return &s, nil
}

func (r *SectionRepository) scanSectionRow(rows *sql.Rows) (*domain.Section, error) {
	var s domain.Section
	var createdAt, updatedAt string
	var statusesJSON sql.NullString
	err := rows.Scan(
		&s.ID, &s.Code, &s.ResearchID, &s.Name, &s.DisplayName,
		&s.Description, &s.Status, &s.Position, &statusesJSON,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan section row: %w", err)
	}
	s.AllowedEntryStatuses = unmarshalEntryStatuses(statusesJSON)
	s.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
	s.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)
	return &s, nil
}

// marshalEntryStatuses encodes the statuses slice to a JSON string for storage.
func marshalEntryStatuses(statuses []string) *string {
	if len(statuses) == 0 {
		return nil
	}
	b, _ := json.Marshal(statuses)
	s := string(b)
	return &s
}

// unmarshalEntryStatuses decodes the JSON string from the DB into a slice.
// Returns DefaultEntryStatuses when NULL.
func unmarshalEntryStatuses(raw sql.NullString) []string {
	if !raw.Valid || raw.String == "" {
		return append([]string{}, domain.DefaultEntryStatuses...)
	}
	var result []string
	if err := json.Unmarshal([]byte(raw.String), &result); err != nil {
		return append([]string{}, domain.DefaultEntryStatuses...)
	}
	return result
}
