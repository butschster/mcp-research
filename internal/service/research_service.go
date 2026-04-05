package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/storage"
	"github.com/google/uuid"
)

type CreateResearchRequest struct {
	Name        string
	Description string
	Goal        string
	Tags        []string
	Sections    []CreateSectionRequest
}

type CreateSectionRequest struct {
	Name        string
	DisplayName string
	Description string
	Position    int
}

type UpdateResearchRequest struct {
	Name        *string
	Description *string
	Goal        *string
	Status      *domain.ResearchStatus
	Instruction *string
	Tags        []string
	Memory      []string  // replace entire memory
	AddMemory   *string   // append single entry
}

type ResearchService struct {
	researches *storage.ResearchRepository
	sections   *storage.SectionRepository
	events     EventNotifier
	log        *slog.Logger
}

func NewResearchService(researches *storage.ResearchRepository, sections *storage.SectionRepository, events EventNotifier, log *slog.Logger) *ResearchService {
	return &ResearchService{researches: researches, sections: sections, events: events, log: log}
}

func (s *ResearchService) Create(ctx context.Context, req CreateResearchRequest) (*domain.Research, []*domain.Section, error) {
	research := &domain.Research{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Goal:        req.Goal,
		Status:      domain.ResearchActive,
		Memory:      []string{},
		Tags:        req.Tags,
	}
	if research.Tags == nil {
		research.Tags = []string{}
	}

	if err := s.researches.Create(ctx, research); err != nil {
		return nil, nil, fmt.Errorf("create research: %w", err)
	}

	var sections []*domain.Section
	for _, sec := range req.Sections {
		section := &domain.Section{
			ID:          uuid.New().String(),
			ResearchID:  research.ID,
			Name:        sec.Name,
			DisplayName: sec.DisplayName,
			Description: sec.Description,
			Status:      domain.SectionDraft,
			Position:    sec.Position,
		}
		if err := s.sections.Create(ctx, section); err != nil {
			return nil, nil, fmt.Errorf("create section %s: %w", sec.Name, err)
		}
		sections = append(sections, section)
	}

	s.events.Notify(Event{Type: "research.created", ResearchID: research.ID, EntityID: research.ID, Entity: "research"})
	return research, sections, nil
}

func (s *ResearchService) Get(ctx context.Context, id string) (*domain.Research, error) {
	research, err := s.researches.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find research: %w", err)
	}
	// If not found by UUID, try by short code
	if research == nil && isCode(id) {
		research, err = s.researches.FindByCode(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("find research by code: %w", err)
		}
	}
	if research == nil {
		return nil, ErrNotFound
	}
	return research, nil
}

// ResolveID resolves a UUID or short code to a research UUID.
func (s *ResearchService) ResolveID(ctx context.Context, idOrCode string) (string, error) {
	r, err := s.Get(ctx, idOrCode)
	if err != nil {
		return "", err
	}
	return r.ID, nil
}

func (s *ResearchService) List(ctx context.Context, filter storage.ResearchFilter) ([]*domain.Research, error) {
	return s.researches.FindAll(ctx, filter)
}

func (s *ResearchService) Update(ctx context.Context, id string, req UpdateResearchRequest) (*domain.Research, error) {
	// Mutual exclusion: memory and add_memory
	if req.Memory != nil && req.AddMemory != nil {
		return nil, fmt.Errorf("memory and add_memory: %w", ErrMutualExclusion)
	}

	research, err := s.researches.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find research: %w", err)
	}
	if research == nil {
		return nil, ErrNotFound
	}

	if req.Name != nil {
		research.Name = *req.Name
	}
	if req.Description != nil {
		research.Description = *req.Description
	}
	if req.Goal != nil {
		research.Goal = *req.Goal
	}
	if req.Status != nil {
		research.Status = *req.Status
	}
	if req.Instruction != nil {
		research.Instruction = *req.Instruction
	}
	if req.Tags != nil {
		research.Tags = req.Tags
	}
	if req.Memory != nil {
		research.Memory = req.Memory
	}
	if req.AddMemory != nil {
		research.Memory = append(research.Memory, *req.AddMemory)
	}

	if err := s.researches.Update(ctx, research); err != nil {
		return nil, fmt.Errorf("update research: %w", err)
	}

	s.events.Notify(Event{Type: "research.updated", ResearchID: research.ID, EntityID: research.ID, Entity: "research"})
	return research, nil
}

func (s *ResearchService) AddSection(ctx context.Context, researchID string, req CreateSectionRequest) (*domain.Section, error) {
	exists, err := s.researches.Exists(ctx, researchID)
	if err != nil {
		return nil, fmt.Errorf("check research: %w", err)
	}
	if !exists {
		return nil, ErrNotFound
	}

	// Check for duplicate name
	existing, err := s.sections.FindByResearchAndName(ctx, researchID, req.Name)
	if err != nil {
		return nil, fmt.Errorf("check section name: %w", err)
	}
	if existing != nil {
		return nil, ErrDuplicateSectionName
	}

	section := &domain.Section{
		ID:          uuid.New().String(),
		ResearchID:  researchID,
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Status:      domain.SectionDraft,
		Position:    req.Position,
	}

	if err := s.sections.Create(ctx, section); err != nil {
		return nil, fmt.Errorf("create section: %w", err)
	}

	s.events.Notify(Event{Type: "section.created", ResearchID: researchID, EntityID: section.ID, Entity: "section"})
	return section, nil
}
