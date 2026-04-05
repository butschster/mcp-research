package service

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/storage"
	"github.com/google/uuid"
)

var refPattern = regexp.MustCompile(`\[\[([^\]]+)\]\]`)

type CreateEntryRequest struct {
	ResearchID string
	SectionID  string
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
}

type TextReplace struct {
	From string
	To   string
}

type EntryService struct {
	entries    *storage.EntryRepository
	sections   *storage.SectionRepository
	researches *storage.ResearchRepository
	crossrefs  *storage.CrossRefRepository
	events     EventNotifier
	log        *slog.Logger
}

func NewEntryService(entries *storage.EntryRepository, sections *storage.SectionRepository, researches *storage.ResearchRepository, crossrefs *storage.CrossRefRepository, events EventNotifier, log *slog.Logger) *EntryService {
	return &EntryService{entries: entries, sections: sections, researches: researches, crossrefs: crossrefs, events: events, log: log}
}

func (s *EntryService) Create(ctx context.Context, req CreateEntryRequest) (*domain.Entry, error) {
	// Validate research exists
	exists, err := s.researches.Exists(ctx, req.ResearchID)
	if err != nil {
		return nil, fmt.Errorf("check research: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("research %s: %w", req.ResearchID, ErrNotFound)
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

	entry := &domain.Entry{
		ID:          uuid.New().String(),
		ResearchID:  req.ResearchID,
		SectionID:   req.SectionID,
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
	return entry, nil
}

func (s *EntryService) List(ctx context.Context, researchID, sectionID string, filter storage.EntryFilter) ([]*domain.Entry, error) {
	return s.entries.FindBySection(ctx, researchID, sectionID, filter)
}

func (s *EntryService) ListByResearch(ctx context.Context, researchID string, filter storage.EntryFilter) ([]*domain.Entry, error) {
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

	if req.Title != nil {
		entry.Title = *req.Title
	}
	if req.Content != nil {
		entry.Content = *req.Content
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
	s.events.Notify(Event{Type: "entry.updated", ResearchID: entry.ResearchID, EntityID: entry.ID, Entity: "entry"})
	return entry, nil
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
		count++
	}
	return count, nil
}

// updateCrossRefs parses [[...]] references from entry content and stores them.
func (s *EntryService) updateCrossRefs(ctx context.Context, entry *domain.Entry) {
	if s.crossrefs == nil {
		return
	}

	matches := refPattern.FindAllStringSubmatch(entry.Content, -1)
	var refs []domain.CrossRef

	for _, m := range matches {
		raw := m[1]
		cr := domain.CrossRef{
			SourceEntryID:    entry.ID,
			SourceResearchID: entry.ResearchID,
			TargetRef:        raw,
		}

		researchCode, entryCode := parseRef(raw)

		if researchCode != "" {
			// Cross-research reference
			targetResearch, err := s.researches.FindByCode(ctx, researchCode)
			if err == nil && targetResearch != nil {
				cr.TargetResearchID = targetResearch.ID
				if entryCode != "" {
					targetEntry, err := s.entries.FindByCode(ctx, targetResearch.ID, entryCode)
					if err == nil && targetEntry != nil {
						cr.TargetEntryID = targetEntry.ID
						cr.Resolved = true
					}
				} else {
					cr.Resolved = true
				}
			}
		} else if entryCode != "" {
			// Same-research reference
			cr.TargetResearchID = entry.ResearchID
			targetEntry, err := s.entries.FindByCode(ctx, entry.ResearchID, entryCode)
			if err == nil && targetEntry != nil {
				cr.TargetEntryID = targetEntry.ID
				cr.Resolved = true
			}
		}

		refs = append(refs, cr)
	}

	if err := s.crossrefs.ReplaceForEntry(ctx, entry.ID, refs); err != nil {
		s.log.Error("failed to update crossrefs", "entry_id", entry.ID, "error", err)
	}
}

// parseRef splits "R2:E5" into (researchCode="R2", entryCode="E5"),
// "E5" into ("", "E5"), "R2" into ("R2", "").
func parseRef(ref string) (researchCode, entryCode string) {
	if idx := strings.IndexByte(ref, ':'); idx >= 0 {
		return ref[:idx], ref[idx+1:]
	}
	if len(ref) > 1 && ref[0] == 'R' {
		return ref, ""
	}
	return "", ref
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
