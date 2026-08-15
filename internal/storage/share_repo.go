package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/butschster/mcp-research/internal/domain"
)

type ShareRepository struct {
	db *sql.DB
}

func NewShareRepository(db *sql.DB) *ShareRepository {
	return &ShareRepository{db: db}
}

// shareColumns joins in the two names a share is displayed with. The manage
// list has to say who handed a link out, and the research code is what every
// URL in the web UI is built from.
const shareColumns = `s.id, s.research_id, COALESCE(s.user_id, ''), s.scope, s.target_id, s.label, s.include,
	s.password_hash <> '', s.expires_at, s.revoked_at, s.last_seen_at, s.view_count, s.created_at,
	COALESCE(u.name, u.email, ''), COALESCE(r.code, '')`

const shareFrom = ` FROM shares s
	LEFT JOIN users u ON u.id = s.user_id
	LEFT JOIN researches r ON r.id = s.research_id`

// Create stores a share. The token hash and the password hash are passed
// separately from the struct and never live on it: the struct is what gets
// serialised into an API response, and a secret that is never in it cannot be
// leaked by adding a field to the response later.
func (r *ShareRepository) Create(ctx context.Context, sh *domain.Share, tokenHash, passwordHash string) error {
	now := time.Now().UTC().Format(time.DateTime)
	include, err := json.Marshal(sh.Include)
	if err != nil {
		return fmt.Errorf("marshal include: %w", err)
	}
	var expires any
	if sh.ExpiresAt != nil {
		expires = sh.ExpiresAt.UTC().Format(time.DateTime)
	}
	// NULL rather than "": there are no users in local mode, and an empty
	// string is not one of them as far as the foreign key is concerned.
	var creator any
	if sh.UserID != "" {
		creator = sh.UserID
	}
	_, err = r.db.ExecContext(ctx,
		`INSERT INTO shares (id, token_hash, research_id, user_id, scope, target_id, label,
		                     include, password_hash, expires_at, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		sh.ID, tokenHash, sh.ResearchID, creator, string(sh.Scope), sh.TargetID, sh.Label,
		string(include), passwordHash, expires, now)
	if err != nil {
		return fmt.Errorf("insert share: %w", err)
	}
	sh.CreatedAt, _ = time.Parse(time.DateTime, now)
	sh.HasPassword = passwordHash != ""
	return nil
}

// FindByHash is the lookup a visited link performs. It returns the share
// whatever its state — revoked, expired or live — because deciding which of
// those a caller is told is the service's job, and the answer is the same for
// all three.
func (r *ShareRepository) FindByHash(ctx context.Context, tokenHash string) (*domain.Share, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+shareColumns+shareFrom+` WHERE s.token_hash = ?`, tokenHash)
	return scanShare(row)
}

// PasswordHash returns the bcrypt hash for a share, separately from the share
// itself so it cannot ride along into a response.
func (r *ShareRepository) PasswordHash(ctx context.Context, id string) (string, error) {
	var hash string
	err := r.db.QueryRowContext(ctx, `SELECT password_hash FROM shares WHERE id = ?`, id).Scan(&hash)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("read share password: %w", err)
	}
	return hash, nil
}

func (r *ShareRepository) FindByID(ctx context.Context, id string) (*domain.Share, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+shareColumns+shareFrom+` WHERE s.id = ?`, id)
	return scanShare(row)
}

// ListByResearch returns every share ever made for a research, revoked and
// expired ones included.
//
// Filtering the dead ones away would be the wrong kindness: "who did I give
// this to, and did I take it back" is the question this list answers, and a row
// that silently disappears on revocation leaves no record that it existed.
func (r *ShareRepository) ListByResearch(ctx context.Context, researchID string) ([]*domain.Share, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+shareColumns+shareFrom+` WHERE s.research_id = ? ORDER BY s.created_at DESC`,
		researchID)
	if err != nil {
		return nil, fmt.Errorf("query shares: %w", err)
	}
	defer rows.Close()

	var result []*domain.Share
	for rows.Next() {
		sh, err := scanShare(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, sh)
	}
	return result, rows.Err()
}

// CountActive is how many links are live right now, for the badge that tells an
// owner a research is exposed without making them open a dialog to find out.
func (r *ShareRepository) CountActive(ctx context.Context, researchID string) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM shares
		  WHERE research_id = ? AND revoked_at IS NULL
		    AND (expires_at IS NULL OR expires_at > ?)`,
		researchID, time.Now().UTC().Format(time.DateTime)).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("count active shares: %w", err)
	}
	return n, nil
}

// Revoke closes a share. The WHERE clause carries the guard, so revoking twice
// is reported as "already gone" rather than silently succeeding and moving the
// timestamp.
func (r *ShareRepository) Revoke(ctx context.Context, id string) error {
	now := time.Now().UTC().Format(time.DateTime)
	res, err := r.db.ExecContext(ctx,
		`UPDATE shares SET revoked_at=? WHERE id=? AND revoked_at IS NULL`, now, id)
	if err != nil {
		return fmt.Errorf("revoke share: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// TouchSeen records that the link was opened. Errors are dropped on purpose:
// this is analytics, and a visitor must not be refused a page because a counter
// could not be written.
func (r *ShareRepository) TouchSeen(ctx context.Context, id string) {
	now := time.Now().UTC().Format(time.DateTime)
	r.db.ExecContext(ctx,
		`UPDATE shares SET view_count = view_count + 1, last_seen_at = ? WHERE id = ?`, now, id)
}

func scanShare(s scanner) (*domain.Share, error) {
	var sh domain.Share
	var scope, include, createdAt string
	var expiresAt, revokedAt, lastSeenAt sql.NullString
	err := s.Scan(&sh.ID, &sh.ResearchID, &sh.UserID, &scope, &sh.TargetID, &sh.Label, &include,
		&sh.HasPassword, &expiresAt, &revokedAt, &lastSeenAt, &sh.ViewCount, &createdAt,
		&sh.CreatorName, &sh.ResearchCode)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan share: %w", err)
	}
	sh.Scope = domain.ShareScope(scope)
	// A row whose include column is unreadable falls back to the zero value,
	// which includes nothing. Failing towards less is the only safe direction.
	if err := json.Unmarshal([]byte(include), &sh.Include); err != nil {
		sh.Include = domain.ShareInclude{}
	}
	sh.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
	if expiresAt.Valid {
		t, _ := time.Parse(time.DateTime, expiresAt.String)
		sh.ExpiresAt = &t
	}
	if revokedAt.Valid {
		t, _ := time.Parse(time.DateTime, revokedAt.String)
		sh.RevokedAt = &t
	}
	if lastSeenAt.Valid {
		t, _ := time.Parse(time.DateTime, lastSeenAt.String)
		sh.LastSeenAt = &t
	}
	return &sh, nil
}
