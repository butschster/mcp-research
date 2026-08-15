package domain

import "time"

// ShareScope names how much of a research a link reaches. Only ShareResearch is
// issued today — the narrower forms are in the schema so they will not need a
// second migration, and the service refuses them rather than enforcing them
// halfway.
type ShareScope string

const (
	ShareResearch ShareScope = "research"
)

// ShareInclude is what the creator chose to put in the link beyond the content
// itself. Entries, sections and roadmaps are the research; these four are the
// parts a person might reasonably not want a client to see.
//
// The zero value is the safe one: nothing extra. A flag that has to be set to
// reveal something cannot leak it by being forgotten.
type ShareInclude struct {
	// Sessions exposes the interview transcript — questions and answers.
	Sessions bool `json:"sessions"`
	// Tasks exposes the agent's own todo list, which is working state rather
	// than a result.
	Tasks bool `json:"tasks"`
	// Roadmaps are part of the research proper, so this defaults to true at
	// creation; it is here so a link can leave them out.
	Roadmaps bool `json:"roadmaps"`
	// Export lets the visitor download markdown. Without it they may read the
	// research but not walk away with a file.
	Export bool `json:"export"`
}

// Share is a read-only capability over one research, addressed by an
// unguessable token rather than by who is holding it.
//
// The token itself is never in this struct: it exists in the response to the
// one request that created it and nowhere else. What is stored is its SHA-256,
// which the repository takes and returns separately for exactly that reason.
type Share struct {
	ID         string `json:"id"`
	ResearchID string `json:"research_id"`
	// UserID is who created the link. A record, not a permission — the link
	// outlives its creator's session and dies only with revocation, expiry or
	// the research.
	UserID   string       `json:"user_id,omitempty"`
	Scope    ShareScope   `json:"scope"`
	TargetID string       `json:"target_id,omitempty"`
	Label    string       `json:"label"`
	Include  ShareInclude `json:"include"`
	// HasPassword rather than the hash: whether a link is protected is useful
	// to the owner listing their shares, and the hash is useful to nobody.
	HasPassword bool       `json:"has_password"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	RevokedAt   *time.Time `json:"revoked_at,omitempty"`
	LastSeenAt  *time.Time `json:"last_seen_at,omitempty"`
	ViewCount   int        `json:"view_count"`
	CreatedAt   time.Time  `json:"created_at"`

	// CreatorName is resolved per request for the manage-shares list, so a team
	// can tell who handed the link out. Not stored on the share.
	CreatorName string `json:"created_by_name,omitempty"`
	// ResearchCode names the research the way every URL in the web UI does.
	ResearchCode string `json:"research_code,omitempty"`
}

// Active reports whether the link still works: not revoked, not expired.
//
// Callers must not use this to explain a refusal — revoked, expired and never
// existed are one indistinguishable answer to whoever is holding the URL.
func (s *Share) Active(now time.Time) bool {
	if s == nil || s.RevokedAt != nil {
		return false
	}
	return s.ExpiresAt == nil || s.ExpiresAt.After(now)
}
