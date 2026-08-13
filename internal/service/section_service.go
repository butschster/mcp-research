package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/storage"
)

type UpdateSectionRequest struct {
	DisplayName *string
	Description *string
	Status      *domain.SectionStatus
	Position    *int
}

type SectionService struct {
	sections   *storage.SectionRepository
	entries    *storage.EntryRepository
	researches *storage.ResearchRepository
	events     EventNotifier
	log        *slog.Logger
}

func NewSectionService(sections *storage.SectionRepository, entries *storage.EntryRepository, researches *storage.ResearchRepository, events EventNotifier, log *slog.Logger) *SectionService {
	return &SectionService{sections: sections, entries: entries, researches: researches, events: events, log: log}
}

func (s *SectionService) List(ctx context.Context, researchID string) ([]*domain.Section, error) {
	if err := validateResearchAccess(ctx, s.researches, researchID); err != nil {
		return nil, err
	}
	return s.sections.FindByResearch(ctx, researchID)
}

func (s *SectionService) Get(ctx context.Context, id string) (*domain.Section, error) {
	section, err := s.sections.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find section: %w", err)
	}
	if section == nil {
		return nil, ErrNotFound
	}
	if err := validateResearchAccess(ctx, s.researches, section.ResearchID); err != nil {
		return nil, ErrNotFound
	}
	return section, nil
}

func (s *SectionService) Update(ctx context.Context, id string, req UpdateSectionRequest) (*domain.Section, error) {
	section, err := s.sections.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find section: %w", err)
	}
	if section == nil {
		return nil, ErrNotFound
	}
	if err := validateResearchAccess(ctx, s.researches, section.ResearchID); err != nil {
		return nil, ErrNotFound
	}

	if req.Status != nil && *req.Status == domain.SectionCompleted {
		count, err := s.entries.CountBySection(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("count entries: %w", err)
		}
		if count == 0 {
			return nil, ErrSectionHasNoEntries
		}
	}

	if req.DisplayName != nil {
		section.DisplayName = normalizeTitle(*req.DisplayName)
	}
	if req.Description != nil {
		section.Description = normalizeContent(*req.Description)
	}
	if req.Status != nil {
		section.Status = *req.Status
	}
	if req.Position != nil {
		section.Position = *req.Position
	}

	if err := s.sections.Update(ctx, section); err != nil {
		return nil, fmt.Errorf("update section: %w", err)
	}

	s.events.Notify(Event{Type: "section.updated", ResearchID: section.ResearchID, EntityID: section.ID, Entity: "section"})
	return section, nil
}

func (s *SectionService) CountEntries(ctx context.Context, sectionID string) (int, error) {
	return s.entries.CountBySection(ctx, sectionID)
}
