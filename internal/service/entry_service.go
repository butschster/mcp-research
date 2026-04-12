package service

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"regexp"
	"strings"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/storage"
	"github.com/google/uuid"
)

var refPattern = regexp.MustCompile(`\[\[([^\]]+)\]\]`)

// Matches markdown links [title](url) and bare URLs
var mdLinkPattern = regexp.MustCompile(`\[([^\]]*)\]\((https?://[^)]+)\)`)
var bareLinkPattern = regexp.MustCompile(`(?:^|[\s(])((https?://)[^\s)<>]+)`)

// CrossRefParser parses [[...]] references from text and stores them.
type CrossRefParser interface {
	ParseCrossRefs(ctx context.Context, sourceType, sourceID, researchID, text string)
}

type CreateEntryRequest struct {
	ResearchID string
	SectionID  string
	SessionID  string
	Content    string
	Title      string
	Description string
	Status     domain.EntryStatus
	Tags       []string
}

type UpdateEntryRequest struct {
	Title       *string
	Content     *string
	Description *string
	Status      *domain.EntryStatus
	Tags        []string
	TextReplace *TextReplace
	SessionID   *string
}

type TextReplace struct {
	From string
	To   string
}

type EntryService struct {
	entries       *storage.EntryRepository
	sections      *storage.SectionRepository
	researches    *storage.ResearchRepository
	sessions      *storage.SessionRepository
	crossrefs     *storage.CrossRefRepository
	externalLinks *storage.ExternalLinkRepository
	roadmaps      *storage.RoadmapRepository
	roadmapNodes  *storage.RoadmapNodeRepository
	events        EventNotifier
	log           *slog.Logger
}

func NewEntryService(entries *storage.EntryRepository, sections *storage.SectionRepository, researches *storage.ResearchRepository, sessions *storage.SessionRepository, crossrefs *storage.CrossRefRepository, externalLinks *storage.ExternalLinkRepository, events EventNotifier, log *slog.Logger) *EntryService {
	return &EntryService{entries: entries, sections: sections, researches: researches, sessions: sessions, crossrefs: crossrefs, externalLinks: externalLinks, events: events, log: log}
}

// SetRoadmapRepos enables [[RM1]] and [[RM1:N3]] cross-reference resolution.
func (s *EntryService) SetRoadmapRepos(roadmaps *storage.RoadmapRepository, nodes *storage.RoadmapNodeRepository) {
	s.roadmaps = roadmaps
	s.roadmapNodes = nodes
}

func (s *EntryService) Create(ctx context.Context, req CreateEntryRequest) (*domain.Entry, error) {
	// Validate research exists and current user has access
	if err := validateResearchAccess(ctx, s.researches, req.ResearchID); err != nil {
		return nil, fmt.Errorf("research %s: %w", req.ResearchID, err)
	}

	// Validate section exists and belongs to research
	section, err := s.sections.FindByID(ctx, req.SectionID)
	if err != nil {
		return nil, fmt.Errorf("find section: %w", err)
	}
	if section == nil {
		return nil, fmt.Errorf("section %s: %w", req.SectionID, ErrNotFound)
	}
	if section.ResearchID != req.ResearchID {
		return nil, fmt.Errorf("section %s does not belong to research %s", req.SectionID, req.ResearchID)
	}

	req.Content = normalizeContent(req.Content)

	if req.Content == "" {
		return nil, fmt.Errorf("content is required")
	}

	title := req.Title
	if title == "" {
		title = autoTitle(req.Content)
	}

	description := req.Description
	if description == "" {
		description = autoDescription(req.Content)
	}

	status := req.Status
	if status == "" {
		status = domain.EntryDraft
	}

	tags := req.Tags
	if tags == nil {
		tags = []string{}
	}

	// Auto-assign active session if not specified
	sessionID := req.SessionID
	if sessionID == "" && s.sessions != nil {
		if active, _ := s.sessions.FindActive(ctx, req.ResearchID); active != nil {
			sessionID = active.ID
		}
	}

	entry := &domain.Entry{
		ID:          uuid.New().String(),
		ResearchID:  req.ResearchID,
		SectionID:   req.SectionID,
		SessionID:   sessionID,
		Title:       title,
		Content:     req.Content,
		Description: description,
		Status:      status,
		Tags:        tags,
	}

	if err := s.entries.Create(ctx, entry); err != nil {
		return nil, fmt.Errorf("create entry: %w", err)
	}

	s.updateCrossRefs(ctx, entry)
	s.updateExternalLinks(ctx, entry)
	s.events.Notify(Event{Type: "entry.created", ResearchID: entry.ResearchID, EntityID: entry.ID, Entity: "entry"})
	return entry, nil
}

func (s *EntryService) Get(ctx context.Context, id string) (*domain.Entry, error) {
	entry, err := s.entries.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find entry: %w", err)
	}
	if entry == nil {
		return nil, ErrNotFound
	}
	if err := validateResearchAccess(ctx, s.researches, entry.ResearchID); err != nil {
		return nil, ErrNotFound
	}
	return entry, nil
}

// GetByIDOrCode resolves an entry by UUID or short code within a research.
func (s *EntryService) GetByIDOrCode(ctx context.Context, researchID, idOrCode string) (*domain.Entry, error) {
	if err := validateResearchAccess(ctx, s.researches, researchID); err != nil {
		return nil, ErrNotFound
	}
	// Try UUID first
	entry, err := s.entries.FindByID(ctx, idOrCode)
	if err != nil {
		return nil, fmt.Errorf("find entry: %w", err)
	}
	// If not found and looks like a code, try by code
	if entry == nil && isCode(idOrCode) {
		entry, err = s.entries.FindByCode(ctx, researchID, idOrCode)
		if err != nil {
			return nil, fmt.Errorf("find entry by code: %w", err)
		}
	}
	if entry == nil {
		return nil, ErrNotFound
	}
	return entry, nil
}

func (s *EntryService) List(ctx context.Context, researchID, sectionID string, filter storage.EntryFilter) ([]*domain.Entry, error) {
	if err := validateResearchAccess(ctx, s.researches, researchID); err != nil {
		return nil, err
	}
	return s.entries.FindBySection(ctx, researchID, sectionID, filter)
}

func (s *EntryService) ListByResearch(ctx context.Context, researchID string, filter storage.EntryFilter) ([]*domain.Entry, error) {
	if err := validateResearchAccess(ctx, s.researches, researchID); err != nil {
		return nil, err
	}
	return s.entries.FindByResearch(ctx, researchID, filter)
}

func (s *EntryService) Update(ctx context.Context, id string, req UpdateEntryRequest) (*domain.Entry, error) {
	entry, err := s.entries.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find entry: %w", err)
	}
	if entry == nil {
		return nil, ErrNotFound
	}
	if err := validateResearchAccess(ctx, s.researches, entry.ResearchID); err != nil {
		return nil, ErrNotFound
	}

	if req.Title != nil {
		entry.Title = *req.Title
	}
	if req.Content != nil {
		entry.Content = normalizeContent(*req.Content)
	}
	if req.Description != nil {
		entry.Description = *req.Description
	}
	if req.Status != nil {
		entry.Status = *req.Status
	}
	if req.Tags != nil {
		entry.Tags = req.Tags
	}
	if req.SessionID != nil {
		entry.SessionID = *req.SessionID
	}

	// text_replace
	if req.TextReplace != nil {
		if !strings.Contains(entry.Content, req.TextReplace.From) {
			return nil, ErrTextReplaceNotFound
		}
		entry.Content = strings.Replace(entry.Content, req.TextReplace.From, req.TextReplace.To, 1)
	}

	if err := s.entries.Update(ctx, entry); err != nil {
		return nil, fmt.Errorf("update entry: %w", err)
	}

	s.updateCrossRefs(ctx, entry)
	s.updateExternalLinks(ctx, entry)
	s.events.Notify(Event{Type: "entry.updated", ResearchID: entry.ResearchID, EntityID: entry.ID, Entity: "entry"})
	return entry, nil
}

func (s *EntryService) Delete(ctx context.Context, id string) error {
	entry, err := s.entries.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("find entry: %w", err)
	}
	if entry == nil {
		return ErrNotFound
	}
	if err := validateResearchAccess(ctx, s.researches, entry.ResearchID); err != nil {
		return ErrNotFound
	}

	// Clean up cross-references and external links
	if s.crossrefs != nil {
		_ = s.crossrefs.ReplaceForSource(ctx, "entry", id, nil)
	}
	if s.externalLinks != nil {
		_ = s.externalLinks.ReplaceForSource(ctx, "entry", id, nil)
	}

	if err := s.entries.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete entry: %w", err)
	}
	s.events.Notify(Event{Type: "entry.deleted", ResearchID: entry.ResearchID, EntityID: id, Entity: "entry"})
	return nil
}

// RebuildCrossRefs rescans all entries in a research and rebuilds cross-references.
func (s *EntryService) RebuildCrossRefs(ctx context.Context, researchID string) (int, error) {
	entries, err := s.entries.FindByResearchWithContent(ctx, researchID)
	if err != nil {
		return 0, fmt.Errorf("fetch entries: %w", err)
	}

	count := 0
	for _, entry := range entries {
		s.updateCrossRefs(ctx, entry)
		s.updateExternalLinks(ctx, entry)
		count++
	}
	return count, nil
}

// updateCrossRefs parses [[...]] references from entry content and stores them.
func (s *EntryService) updateCrossRefs(ctx context.Context, entry *domain.Entry) {
	s.parseCrossRefs(ctx, "entry", entry.ID, entry.ResearchID, entry.Content)
}

// updateExternalLinks extracts URLs from entry content and stores them.
func (s *EntryService) updateExternalLinks(ctx context.Context, entry *domain.Entry) {
	s.parseExternalLinks(ctx, "entry", entry.ID, entry.ResearchID, entry.Content)
}

// ParseExternalLinks extracts URLs from text and stores them.
func (s *EntryService) ParseExternalLinks(ctx context.Context, sourceType, sourceID, researchID, text string) {
	s.parseExternalLinks(ctx, sourceType, sourceID, researchID, text)
}

func (s *EntryService) parseExternalLinks(ctx context.Context, sourceType, sourceID, researchID, text string) {
	if s.externalLinks == nil || text == "" {
		return
	}

	seen := make(map[string]bool)
	var links []domain.ExternalLink

	// Extract [title](url) markdown links
	for _, m := range mdLinkPattern.FindAllStringSubmatch(text, -1) {
		rawURL := m[2]
		if seen[rawURL] {
			continue
		}
		seen[rawURL] = true
		links = append(links, domain.ExternalLink{
			ID:         uuid.New().String(),
			SourceType: sourceType,
			SourceID:   sourceID,
			ResearchID: researchID,
			URL:        rawURL,
			Title:      m[1],
			Domain:     extractDomain(rawURL),
		})
	}

	// Extract bare URLs not already captured
	for _, m := range bareLinkPattern.FindAllStringSubmatch(text, -1) {
		rawURL := m[1]
		if seen[rawURL] {
			continue
		}
		seen[rawURL] = true
		links = append(links, domain.ExternalLink{
			ID:         uuid.New().String(),
			SourceType: sourceType,
			SourceID:   sourceID,
			ResearchID: researchID,
			URL:        rawURL,
			Title:      "",
			Domain:     extractDomain(rawURL),
		})
	}

	if err := s.externalLinks.ReplaceForSource(ctx, sourceType, sourceID, links); err != nil {
		s.log.Error("failed to update external links", "source_type", sourceType, "source_id", sourceID, "error", err)
	}
}

func extractDomain(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return u.Hostname()
}

// ParseCrossRefs extracts [[...]] references from text and stores them.
// Can be called for entries, questions, or tasks.
func (s *EntryService) ParseCrossRefs(ctx context.Context, sourceType, sourceID, researchID, text string) {
	s.parseCrossRefs(ctx, sourceType, sourceID, researchID, text)
}

func (s *EntryService) parseCrossRefs(ctx context.Context, sourceType, sourceID, researchID, text string) {
	if s.crossrefs == nil || text == "" {
		return
	}

	matches := refPattern.FindAllStringSubmatch(text, -1)
	var refs []domain.CrossRef

	for _, m := range matches {
		raw := m[1]
		cr := domain.CrossRef{
			SourceType:       sourceType,
			SourceID:         sourceID,
			SourceResearchID: researchID,
			TargetRef:        raw,
		}

		kind, first, second := parseRef(raw)

		switch kind {
		case "roadmap":
			// [[RM1]] — link to a roadmap in the same research
			if s.roadmaps != nil {
				rm, err := s.roadmaps.FindByCode(ctx, first)
				if err == nil && rm != nil {
					cr.TargetRoadmapID = rm.ID
					cr.TargetResearchID = rm.ResearchID
					cr.Resolved = true
				}
			}
		case "node":
			// [[RM1:N3]] — link to a specific node in a roadmap
			if s.roadmaps != nil && s.roadmapNodes != nil {
				rm, err := s.roadmaps.FindByCode(ctx, first)
				if err == nil && rm != nil {
					cr.TargetRoadmapID = rm.ID
					cr.TargetResearchID = rm.ResearchID
					node, err := s.roadmapNodes.FindByCode(ctx, rm.ID, second)
					if err == nil && node != nil {
						cr.TargetNodeID = node.ID
						cr.Resolved = true
					}
				}
			}
		case "research":
			// [[R2]] — link to a research
			targetResearch, err := s.researches.FindByCode(ctx, first)
			if err == nil && targetResearch != nil {
				cr.TargetResearchID = targetResearch.ID
				cr.Resolved = true
			}
		case "entry":
			// [[E3]] or [[R2:E5]] — link to an entry
			if first != "" {
				// Cross-research: [[R2:E5]]
				targetResearch, err := s.researches.FindByCode(ctx, first)
				if err == nil && targetResearch != nil {
					cr.TargetResearchID = targetResearch.ID
					if second != "" {
						targetEntry, err := s.entries.FindByCode(ctx, targetResearch.ID, second)
						if err == nil && targetEntry != nil {
							cr.TargetEntryID = targetEntry.ID
							cr.Resolved = true
						}
					} else {
						cr.Resolved = true
					}
				}
			} else if second != "" {
				// Same-research: [[E3]]
				cr.TargetResearchID = researchID
				targetEntry, err := s.entries.FindByCode(ctx, researchID, second)
				if err == nil && targetEntry != nil {
					cr.TargetEntryID = targetEntry.ID
					cr.Resolved = true
				}
			}
		}

		refs = append(refs, cr)
	}

	if err := s.crossrefs.ReplaceForSource(ctx, sourceType, sourceID, refs); err != nil {
		s.log.Error("failed to update crossrefs", "source_type", sourceType, "source_id", sourceID, "error", err)
	}
}

// parseRef splits references:
//   "R2:E5"  → kind="entry",   first="R2",  second="E5"
//   "E5"     → kind="entry",   first="",    second="E5"
//   "R2"     → kind="research",first="R2",  second=""
//   "RM1"    → kind="roadmap", first="RM1", second=""
//   "RM1:N3" → kind="node",    first="RM1", second="N3"
func parseRef(ref string) (kind, first, second string) {
	if idx := strings.IndexByte(ref, ':'); idx >= 0 {
		left, right := ref[:idx], ref[idx+1:]
		if strings.HasPrefix(left, "RM") {
			return "node", left, right
		}
		return "entry", left, right
	}
	if strings.HasPrefix(ref, "RM") {
		return "roadmap", ref, ""
	}
	if len(ref) > 1 && ref[0] == 'R' {
		return "research", ref, ""
	}
	return "entry", "", ref
}

// autoTitle extracts title from the first non-empty line of content.
func autoTitle(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Strip leading markdown heading markers
		line = strings.TrimLeft(line, "# ")
		line = strings.TrimSpace(line)
		if len(line) > 100 {
			line = line[:100]
		}
		return line
	}
	return "Untitled"
}

// autoDescription extracts description from lines 2-5 of content.
func autoDescription(content string) string {
	lines := strings.Split(content, "\n")
	var descLines []string
	started := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !started {
			if line != "" {
				started = true // skip first non-empty line (title)
			}
			continue
		}
		if line == "" {
			continue
		}
		// Strip markdown
		line = strings.TrimLeft(line, "# *_>-")
		line = strings.TrimSpace(line)
		descLines = append(descLines, line)
		if len(descLines) >= 4 {
			break
		}
	}
	desc := strings.Join(descLines, " ")
	if len(desc) > 200 {
		desc = desc[:200]
	}
	return desc
}
