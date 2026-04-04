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
	sections *storage.SectionRepository
	entries  *storage.EntryRepository
	log      *slog.Logger
}

func NewSectionService(sections *storage.SectionRepository, entries *storage.EntryRepository, log *slog.Logger) *SectionService {
	return &SectionService{sections: sections, entries: entries, log: log}
}

func (s *SectionService) List(ctx context.Context, researchID string) ([]*domain.Section, error) {
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

	// Guard: section can only be completed if it has entries
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
		section.DisplayName = *req.DisplayName
	}
	if req.Description != nil {
		section.Description = *req.Description
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

	return section, nil
}

func (s *SectionService) CountEntries(ctx context.Context, sectionID string) (int, error) {
	return s.entries.CountBySection(ctx, sectionID)
}
