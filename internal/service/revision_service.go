package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/butschster/mcp-research/internal/auth"
	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/storage"
)

// Revision history for entries.
//
// Every write that changes what an entry says appends a snapshot. Restoring an
// old revision writes a new one equal to it, so history is append-only and a
// revision number is never reused.
//
// What deliberately does NOT produce a revision:
//   - a write that changes nothing (an agent rewriting identical text)
//   - a state-only block patch (ticking a checkbox is not an edit to the
//     document; it would otherwise bury the real edits under checkbox noise)

type authorKey struct{}

// WithAuthor marks who is writing, for revisions recorded further down the call
// stack. MCP sets agent, the REST layer sets human or agent depending on the
// credential, import and restore set themselves.
func WithAuthor(ctx context.Context, kind domain.AuthorKind) context.Context {
	if !kind.Valid() {
		return ctx
	}
	return context.WithValue(ctx, authorKey{}, kind)
}

// AuthorFromContext returns who is writing. The default is agent: everything
// that reaches this code without saying otherwise came in over MCP, which is
// how essentially all content in this product is written.
func AuthorFromContext(ctx context.Context) domain.AuthorKind {
	if k, ok := ctx.Value(authorKey{}).(domain.AuthorKind); ok && k.Valid() {
		return k
	}
	return domain.AuthorAgent
}

// revisionNote carries the author's intent for one write. Zero value means
// "figure it out from the context".
type revisionNote struct {
	kind    domain.AuthorKind
	summary string
	skip    bool

	// sessionID is resolved BEFORE the write's transaction opens. The pool runs
	// a single connection (MaxOpenConns(1)), so a query issued from inside a
	// transaction waits for a connection the transaction itself is holding —
	// a deadlock, not a slow query. Anything recordRevision needs from the
	// database has to be gathered out here.
	sessionID  string
	sessionSet bool
}

func (n revisionNote) authorKind(ctx context.Context) domain.AuthorKind {
	if n.kind.Valid() {
		return n.kind
	}
	return AuthorFromContext(ctx)
}

// recordRevision appends a snapshot of the entry as it now stands.
//
// q is the transaction the entry was written in, when there is one: a revision
// that lands while the write it describes is rolled back would be a lie about
// the document's history. The markdown path opens one for exactly this reason;
// the block path already has one.
//
// A revision that would duplicate the newest one is not written. This is what
// keeps a status toggle or a re-save of identical text out of the history.
func (s *EntryService) recordRevision(ctx context.Context, q storage.Querier, entry *domain.Entry, note revisionNote) error {
	if s.revisions == nil || note.skip {
		return nil
	}

	candidate := &domain.EntryRevision{
		EntryID:     entry.ID,
		ResearchID:  entry.ResearchID,
		Title:       entry.Title,
		Description: entry.Description,
		Content:     entry.Content,
		Type:        entry.Type,
		Status:      entry.Status,
		Tags:        entry.Tags,
		AuthorKind:  note.authorKind(ctx),
		SessionID:   note.sessionID,
		UserID:      auth.UserIDFromContext(ctx),
		Summary:     note.summary,
	}
	if !note.sessionSet {
		// Only reachable outside a transaction (entry creation); inside one this
		// would deadlock on the single connection.
		candidate.SessionID = s.activeSessionID(ctx, entry)
	}

	latest, err := s.revisions.Latest(ctx, q, entry.ID)
	if err != nil {
		return fmt.Errorf("read latest revision: %w", err)
	}
	if latest != nil && latest.SameContent(candidate) {
		return nil
	}

	if err := s.revisions.Create(ctx, q, candidate); err != nil {
		return err
	}
	if s.revisionLimit > 0 {
		if err := s.revisions.Trim(ctx, q, entry.ID, s.revisionLimit); err != nil {
			return err
		}
	}
	return nil
}

// resolveSession fills in the session a revision will belong to, before the
// write's transaction opens. See the comment on revisionNote.sessionID.
func (s *EntryService) resolveSession(ctx context.Context, entry *domain.Entry, note revisionNote) revisionNote {
	if note.skip || note.sessionSet {
		return note
	}
	note.sessionID = s.activeSessionID(ctx, entry)
	note.sessionSet = true
	return note
}

// activeSessionID is the session a revision belongs to: the one running when the
// write happened, not the one that first produced the entry. An entry created in
// SS1 and rewritten during SS3 contributes its rewrite to SS3, which is what
// "what did this session change" has to mean to be worth anything.
func (s *EntryService) activeSessionID(ctx context.Context, entry *domain.Entry) string {
	if s.sessions != nil {
		if active, err := s.sessions.FindActive(ctx, entry.ResearchID); err == nil && active != nil {
			return active.ID
		}
	}
	return entry.SessionID
}

// History returns an entry's revisions, newest first, without content.
func (s *EntryService) History(ctx context.Context, entryID string) (*domain.Entry, []*domain.EntryRevision, error) {
	entry, err := s.Get(ctx, entryID)
	if err != nil {
		return nil, nil, err
	}
	if s.revisions == nil {
		return entry, []*domain.EntryRevision{}, nil
	}
	revs, err := s.revisions.ListByEntry(ctx, entry.ID, 0)
	if err != nil {
		return nil, nil, fmt.Errorf("list revisions: %w", err)
	}
	s.enrichSessions(ctx, revs)
	return entry, revs, nil
}

// LatestRevision returns the newest revision of an entry without its content,
// or nil when the entry has none.
//
// This is what puts provenance on the entry page itself: who last wrote the
// document and when, answered without opening the history. Content is stripped
// because the caller already has it — it is the entry.
func (s *EntryService) LatestRevision(ctx context.Context, entry *domain.Entry) *domain.EntryRevision {
	if s.revisions == nil || entry == nil {
		return nil
	}
	rev, err := s.revisions.Latest(ctx, nil, entry.ID)
	if err != nil || rev == nil {
		return nil
	}
	rev.Content = ""
	s.enrichSessions(ctx, []*domain.EntryRevision{rev})
	return rev
}

// Revision returns one revision with its content.
func (s *EntryService) Revision(ctx context.Context, entryID string, number int) (*domain.EntryRevision, error) {
	// Ownership first: Get refuses an entry the caller may not open, and a
	// revision must not be reachable where the entry is not.
	entry, err := s.Get(ctx, entryID)
	if err != nil {
		return nil, err
	}
	if s.revisions == nil {
		return nil, ErrNotFound
	}
	rev, err := s.revisions.FindByNumber(ctx, entry.ID, number)
	if err != nil {
		return nil, fmt.Errorf("find revision: %w", err)
	}
	if rev == nil {
		return nil, ErrNotFound
	}
	s.enrichSessions(ctx, []*domain.EntryRevision{rev})
	return rev, nil
}

// EntryDiff is the comparison of two revisions of one entry.
type EntryDiff struct {
	EntryID   string                `json:"entry_id"`
	EntryCode string                `json:"entry_code,omitempty"`
	From      *domain.EntryRevision `json:"from"`
	To        *domain.EntryRevision `json:"to"`
	Title     *DiffResult           `json:"title,omitempty"`
	Content   DiffResult            `json:"content"`
	Summary   string                `json:"summary"`
	// Fields is what changed outside the body: a retitle, a status flip, an
	// edited description or tag list. Without it a status change renders as a
	// revision whose diff says "nothing changed", which reads as a broken diff
	// rather than as an accurate account of the write.
	Fields []FieldChange `json:"fields,omitempty"`
}

// Diff compares two revisions. from defaults to the one before to; to defaults
// to the newest.
func (s *EntryService) Diff(ctx context.Context, entryID string, from, to int) (*EntryDiff, error) {
	entry, revs, err := s.History(ctx, entryID)
	if err != nil {
		return nil, err
	}
	if len(revs) == 0 {
		return nil, ErrNotFound
	}
	if to <= 0 {
		to = revs[0].Revision
	}
	if from <= 0 {
		from = to - 1
	}

	target, err := s.Revision(ctx, entryID, to)
	if err != nil {
		return nil, err
	}

	var base *domain.EntryRevision
	if from > 0 {
		base, err = s.Revision(ctx, entryID, from)
		if err != nil {
			return nil, err
		}
	}

	beforeText, beforeTitle := "", ""
	if base != nil {
		beforeText = s.revisionText(base)
		beforeTitle = base.Title
	}
	content := DiffText(beforeText, s.revisionText(target))

	out := &EntryDiff{
		EntryID:   entry.ID,
		EntryCode: entry.Code,
		From:      base,
		To:        target,
		Content:   content,
		Summary:   DiffSummary(content),
	}
	if beforeTitle != target.Title {
		titleDiff := DiffText(beforeTitle, target.Title)
		out.Title = &titleDiff
	}
	// Only against a predecessor. With none, every field reads as
	// changed-from-empty — which is the entry being born, not a change to it,
	// and a client that rendered it would announce "title changed from nothing"
	// about a document's first version.
	if base != nil {
		out.Fields = changedFields(base, target)
	}
	return out, nil
}

// FieldChange is a field that changed without the body changing — a status
// flip, a retitle, a tag edit.
type FieldChange struct {
	Field  string `json:"field"`
	Before string `json:"before"`
	After  string `json:"after"`
}

// changedFields reports the metadata a diff of the content alone would show as
// "nothing changed".
//
// A status flip records a revision (rightly — it changed the entry), and its
// content diff is empty. Without this the history showed a revision summarized
// "Updated status" beside a diff pane reading "Nothing changed", which reads as
// a bug in the diff rather than as the truth about the write.
func changedFields(before, after *domain.EntryRevision) []FieldChange {
	if after == nil {
		return nil
	}
	var out []FieldChange
	add := func(field, was, now string) {
		if was != now {
			out = append(out, FieldChange{Field: field, Before: was, After: now})
		}
	}
	var (
		wasTitle, wasDescription, wasStatus, wasTags string
	)
	if before != nil {
		wasTitle = before.Title
		wasDescription = before.Description
		wasStatus = string(before.Status)
		wasTags = strings.Join(before.Tags, ", ")
	}
	add("title", wasTitle, after.Title)
	add("description", wasDescription, after.Description)
	add("status", wasStatus, string(after.Status))
	add("tags", wasTags, strings.Join(after.Tags, ", "))
	return out
}

// revisionText is what a human compares. A block document is diffed through its
// markdown projection, never its JSON: a reader wants to know that a paragraph
// changed, not that a field moved inside an object.
func (s *EntryService) revisionText(rev *domain.EntryRevision) string {
	if rev == nil {
		return ""
	}
	if rev.Type != domain.EntryBlocks {
		return rev.Content
	}
	doc, err := ParseStoredBlockDocument(rev.Content)
	if err != nil {
		return rev.Content
	}
	return BlockDocumentToMarkdown(doc)
}

// Restore writes the content of an earlier revision back onto the entry.
//
// It goes through the ordinary update path, which means normalization, block
// rows, cross-references and external links are rebuilt exactly as they are for
// any other write — and a new revision is appended. History is never rewritten:
// restoring revision 2 onto a document at revision 7 produces revision 8.
func (s *EntryService) Restore(ctx context.Context, entryID string, number int) (*domain.Entry, error) {
	rev, err := s.Revision(ctx, entryID, number)
	if err != nil {
		return nil, err
	}

	entryType := rev.Type
	status := rev.Status
	req := UpdateEntryRequest{
		Type:        &entryType,
		Title:       &rev.Title,
		Content:     &rev.Content,
		Description: &rev.Description,
		Status:      &status,
		Tags:        rev.Tags,
	}
	if req.Tags == nil {
		req.Tags = []string{}
	}

	return s.update(ctx, entryID, req, revisionNote{
		kind:    domain.AuthorRestore,
		summary: fmt.Sprintf("Restored revision %d", number),
	})
}

// SessionEntryChange is one entry as a session left it.
type SessionEntryChange struct {
	EntryID      string                  `json:"entry_id"`
	EntryCode    string                  `json:"entry_code"`
	Title        string                  `json:"title"`
	SectionID    string                  `json:"section_id"`
	Created      bool                    `json:"created"`
	Revisions    []*domain.EntryRevision `json:"revisions"`
	FromRevision int                     `json:"from_revision"`
	ToRevision   int                     `json:"to_revision"`
	Diff         *DiffResult             `json:"diff,omitempty"`
	Summary      string                  `json:"summary"`
	// Fields is what the session changed outside the body. Without it a session
	// whose whole work was flipping statuses renders as "nothing changed" on
	// every card — the same defect the entry diff carries when `fields` is
	// dropped, one screen over.
	Fields []FieldChange `json:"fields,omitempty"`
}

// sessionRevisionGroups collects the session's revisions per entry, in the order
// the entries were first touched. Both the full change list and the count-only
// answer start here; only what follows differs.
func (s *EntryService) sessionRevisionGroups(
	ctx context.Context, researchID, sessionID string,
) ([]string, map[string][]*domain.EntryRevision, error) {
	if err := s.access.Read(ctx, researchID); err != nil {
		return nil, nil, err
	}
	if s.revisions == nil {
		return nil, map[string][]*domain.EntryRevision{}, nil
	}

	revs, err := s.revisions.ListBySession(ctx, sessionID)
	if err != nil {
		return nil, nil, fmt.Errorf("list session revisions: %w", err)
	}

	byEntry := map[string][]*domain.EntryRevision{}
	order := []string{}
	for _, rev := range revs {
		// A session belongs to one research; a revision carrying another
		// research's id would mean the session id was reused, and it must not
		// be able to pull content across the boundary.
		if rev.ResearchID != researchID {
			continue
		}
		if _, seen := byEntry[rev.EntryID]; !seen {
			order = append(order, rev.EntryID)
		}
		byEntry[rev.EntryID] = append(byEntry[rev.EntryID], rev)
	}
	return order, byEntry, nil
}

// SessionChangeCounts answers "how many entries did this session touch, and how
// many of them did it create" without diffing anything.
//
// The Changes tab carries that number before anyone opens it, and the full
// SessionChanges pays an O(n·m) LCS per touched entry: asking it for a badge put
// the whole diff computation on the critical path of every session page load,
// on a database limited to one connection. The entry lookup stays, because the
// list skips an entry whose row is gone and a count that disagreed with the list
// it labels would be its own small lie.
func (s *EntryService) SessionChangeCounts(
	ctx context.Context, researchID, sessionID string,
) (created, modified int, err error) {
	order, byEntry, err := s.sessionRevisionGroups(ctx, researchID, sessionID)
	if err != nil {
		return 0, 0, err
	}

	for _, entryID := range order {
		entry, err := s.entries.FindByID(ctx, entryID)
		if err != nil {
			return 0, 0, fmt.Errorf("find entry: %w", err)
		}
		if entry == nil {
			continue
		}
		if byEntry[entryID][0].Revision == 1 {
			created++
		} else {
			modified++
		}
	}
	return created, modified, nil
}

// SessionChanges reports what a session did to the research's entries.
//
// This is the honest answer to "what came out of this session", and it is a
// better one than the entries whose session_id happens to match: it covers
// entries the session edited without creating, it says how much changed, and it
// can show it.
func (s *EntryService) SessionChanges(ctx context.Context, researchID, sessionID string) ([]*SessionEntryChange, error) {
	order, byEntry, err := s.sessionRevisionGroups(ctx, researchID, sessionID)
	if err != nil {
		return nil, err
	}

	out := make([]*SessionEntryChange, 0, len(order))
	for _, entryID := range order {
		group := byEntry[entryID]
		last := group[len(group)-1]

		change := &SessionEntryChange{
			EntryID:      entryID,
			Title:        last.Title,
			Revisions:    group,
			ToRevision:   last.Revision,
			FromRevision: group[0].Revision,
			Created:      group[0].Revision == 1,
		}

		entry, err := s.entries.FindByID(ctx, entryID)
		if err != nil {
			return nil, fmt.Errorf("find entry: %w", err)
		}
		if entry == nil {
			// Unreachable in practice, and deliberately not dressed up as a
			// feature: entry_revisions cascades on entries(id), so deleting an
			// entry takes its revisions with it and this loop never sees them.
			// A session that wrote twelve entries which were later deleted
			// reports nothing — the record of that work is gone with the work.
			//
			// Keeping the history of a deleted entry needs the row to survive
			// its entry, which is a schema decision, not a rendering one.
			continue
		}
		change.EntryCode = entry.Code
		change.SectionID = entry.SectionID
		change.Title = entry.Title

		// The state before this session is the revision immediately preceding
		// the first one it wrote. For an entry the session created there is
		// none, and the diff is the whole document arriving.
		var base *domain.EntryRevision
		if group[0].Revision > 1 {
			base, err = s.revisions.FindByNumber(ctx, entryID, group[0].Revision-1)
			if err != nil {
				return nil, fmt.Errorf("find base revision: %w", err)
			}
		}
		target, err := s.revisions.FindByNumber(ctx, entryID, last.Revision)
		if err != nil {
			return nil, fmt.Errorf("find target revision: %w", err)
		}
		if target != nil {
			d := DiffText(s.revisionText(base), s.revisionText(target))
			change.Diff = &d
			change.Summary = DiffSummary(d)
			// Suppressed for an entry the session created: every field reads as
			// changed-from-empty, which is the entry's birth, not a change to it.
			if base != nil {
				change.Fields = changedFields(base, target)
			}
		}
		out = append(out, change)
	}

	// The groups hold the same revision pointers the query returned, so filling
	// session codes across the flattened list fills them everywhere they show.
	flat := make([]*domain.EntryRevision, 0, len(order))
	for _, entryID := range order {
		flat = append(flat, byEntry[entryID]...)
	}
	s.enrichSessions(ctx, flat)
	return out, nil
}

// enrichSessions fills session codes and titles for display, in one pass per
// distinct session rather than one query per revision.
func (s *EntryService) enrichSessions(ctx context.Context, revs []*domain.EntryRevision) {
	if s.sessions == nil {
		return
	}
	cache := map[string]*domain.Session{}
	for _, rev := range revs {
		if rev.SessionID == "" {
			continue
		}
		sess, seen := cache[rev.SessionID]
		if !seen {
			found, err := s.sessions.FindByID(ctx, rev.SessionID)
			if err != nil {
				continue
			}
			cache[rev.SessionID] = found
			sess = found
		}
		// The session must belong to the same research as the revision.
		//
		// entries.session_id is caller-supplied and unvalidated by history's
		// standards, so a user could point their own entry at a session id from
		// another user's research and read its code and title back out of their
		// own revision list. Ownership is decided by the research, and the
		// revision already carries it — so this comparison is the check, and it
		// costs no extra query.
		if sess != nil && sess.ResearchID == rev.ResearchID {
			rev.SessionCode = sess.Code
			rev.SessionTitle = sess.Title
		}
	}
}

// inTx runs fn inside a transaction, so an entry and the revision describing it
// land together or not at all.
func (s *EntryService) inTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	if s.revisions == nil {
		return fn(nil)
	}
	tx, err := s.revisions.DB().BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

// summarizeUpdate describes a write in a few words, for the history list. It is
// derived rather than asked for: an agent will not write changelog entries, and
// a history where every row says the same thing is a history nobody reads.
func summarizeUpdate(req UpdateEntryRequest) string {
	var parts []string
	if req.Content != nil || req.TextReplace != nil {
		parts = append(parts, "content")
	}
	if req.Title != nil {
		parts = append(parts, "title")
	}
	if req.Description != nil {
		parts = append(parts, "description")
	}
	if req.Status != nil {
		parts = append(parts, "status")
	}
	if req.Tags != nil {
		parts = append(parts, "tags")
	}
	if req.Type != nil {
		parts = append(parts, "type")
	}
	if len(parts) == 0 {
		return ""
	}
	return "Updated " + strings.Join(parts, ", ")
}
