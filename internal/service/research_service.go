package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/butschster/mcp-research/internal/auth"
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
	// TeamID puts the research in a team other than the creator's personal
	// one. Empty is the ordinary case and the only one a solo user ever hits.
	TeamID string
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
	Memory      []string // replace entire memory
	AddMemory   *string  // append single entry
}

type ResearchService struct {
	researches *storage.ResearchRepository
	sections   *storage.SectionRepository
	teams      *storage.TeamRepository
	access     *Access
	events     EventNotifier
	log        *slog.Logger
}

func NewResearchService(researches *storage.ResearchRepository, sections *storage.SectionRepository, teams *storage.TeamRepository, access *Access, events EventNotifier, log *slog.Logger) *ResearchService {
	return &ResearchService{researches: researches, sections: sections, teams: teams, access: access, events: events, log: log}
}

func (s *ResearchService) Create(ctx context.Context, req CreateResearchRequest) (*domain.Research, []*domain.Section, error) {
	teamID, err := s.resolveCreateTeam(ctx, req.TeamID)
	if err != nil {
		return nil, nil, err
	}

	research := &domain.Research{
		ID:          uuid.New().String(),
		UserID:      auth.UserIDFromContext(ctx),
		TeamID:      teamID,
		Name:        normalizeTitle(req.Name),
		Description: normalizeContent(req.Description),
		Goal:        normalizeContent(req.Goal),
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
			Name:        normalizeTitle(sec.Name),
			DisplayName: normalizeTitle(sec.DisplayName),
			Description: normalizeContent(sec.Description),
			Status:      domain.SectionDraft,
			Position:    sec.Position,
		}
		if err := s.sections.Create(ctx, section); err != nil {
			return nil, nil, fmt.Errorf("create section %s: %w", sec.Name, err)
		}
		sections = append(sections, section)
	}

	s.decorate(ctx, research)
	emit(ctx, s.events, Event{Type: "research.created", ResearchID: research.ID, EntityID: research.ID, Entity: "research"})
	return research, sections, nil
}

// resolveCreateTeam decides which team a new research lands in.
//
// A research with no team would be reachable by nobody once auth is on, so
// there is no "unassigned" outcome here: either the caller named a team they
// may write to, or it goes to their personal one.
func (s *ResearchService) resolveCreateTeam(ctx context.Context, requested string) (string, error) {
	uid := auth.UserIDFromContext(ctx)
	if uid == "" {
		// Local mode: no users, one team, everything in it.
		return domain.LocalTeamID, nil
	}

	if requested != "" {
		role, ok, err := s.teams.FindRole(ctx, requested, uid)
		if err != nil {
			return "", fmt.Errorf("find team role: %w", err)
		}
		if !ok {
			return "", ErrNotFound
		}
		if !role.CanWrite() {
			return "", ErrForbidden
		}
		return requested, nil
	}

	team, err := s.teams.FindPersonal(ctx, uid)
	if err != nil {
		return "", fmt.Errorf("find personal team: %w", err)
	}
	if team == nil {
		// Every user gets one at registration and the migration backfilled the
		// rest, so this is a broken account rather than a missing feature.
		return "", fmt.Errorf("user %s has no personal team", uid)
	}
	return team.ID, nil
}

// decorate fills the per-request fields — the owning team's name and what the
// caller may do — so a reader can tell whose team a research is in and the UI
// can hide controls instead of letting them fail on click.
func (s *ResearchService) decorate(ctx context.Context, research *domain.Research) {
	if research == nil {
		return
	}
	research.Role = s.access.RoleOrEmpty(ctx, research.ID)
	if research.TeamID == "" {
		return
	}
	if team, err := s.teams.FindByID(ctx, research.TeamID); err == nil && team != nil {
		research.TeamName = team.Name
		research.TeamPersonal = team.Personal
	}
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
	// Membership, not ownership: a colleague in the same team sees this, and
	// the creator of a research they were since removed from does not.
	if err := s.access.Read(ctx, research.ID); err != nil {
		return nil, err
	}
	s.decorate(ctx, research)
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

// List returns the researches of every team the caller belongs to. Filtering
// by creator instead would hide a colleague's work in a shared team, which is
// the whole point of having one.
func (s *ResearchService) List(ctx context.Context, filter storage.ResearchFilter) ([]*domain.Research, error) {
	uid := auth.UserIDFromContext(ctx)
	if uid != "" {
		filter.MemberOf = &uid
	}

	researches, err := s.researches.FindAll(ctx, filter)
	if err != nil {
		return nil, err
	}
	if researches == nil {
		// An empty list is now an ordinary state — a new member of a team that
		// holds nothing yet — and `"data": null` is not a list.
		researches = []*domain.Research{}
	}
	if uid == "" || len(researches) == 0 {
		return researches, nil
	}

	// One query for the caller's teams rather than two per research: the list
	// is the busiest read in the product.
	teams, err := s.teams.ListByUser(ctx, uid)
	if err != nil {
		return researches, nil
	}
	byID := make(map[string]*domain.Team, len(teams))
	for _, t := range teams {
		byID[t.ID] = t
	}
	for _, r := range researches {
		if t := byID[r.TeamID]; t != nil {
			r.TeamName = t.Name
			r.TeamPersonal = t.Personal
			r.Role = t.Role
		}
	}
	return researches, nil
}

func (s *ResearchService) Update(ctx context.Context, id string, req UpdateResearchRequest) (*domain.Research, error) {
	// Mutual exclusion: memory and add_memory
	if req.Memory != nil && req.AddMemory != nil {
		return nil, fmt.Errorf("memory and add_memory: %w", ErrMutualExclusion)
	}

	research, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.access.Write(ctx, research.ID); err != nil {
		return nil, err
	}

	if req.Name != nil {
		research.Name = normalizeTitle(*req.Name)
	}
	if req.Description != nil {
		research.Description = normalizeContent(*req.Description)
	}
	if req.Goal != nil {
		research.Goal = normalizeContent(*req.Goal)
	}
	if req.Status != nil {
		research.Status = *req.Status
	}
	if req.Instruction != nil {
		research.Instruction = normalizeContent(*req.Instruction)
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

	emit(ctx, s.events, Event{Type: "research.updated", ResearchID: research.ID, EntityID: research.ID, Entity: "research"})
	return research, nil
}

func (s *ResearchService) AddSection(ctx context.Context, researchID string, req CreateSectionRequest) (*domain.Section, error) {
	if err := s.access.Write(ctx, researchID); err != nil {
		return nil, err
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
		Name:        normalizeTitle(req.Name),
		DisplayName: normalizeTitle(req.DisplayName),
		Description: normalizeContent(req.Description),
		Status:      domain.SectionDraft,
		Position:    req.Position,
	}

	if err := s.sections.Create(ctx, section); err != nil {
		return nil, fmt.Errorf("create section: %w", err)
	}

	emit(ctx, s.events, Event{Type: "section.created", ResearchID: researchID, EntityID: section.ID, Entity: "section"})
	return section, nil
}
