package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/dovod-app/app/internal/auth"
	"github.com/dovod-app/app/internal/domain"
	"github.com/dovod-app/app/internal/storage"
)

type UpdateSectionRequest struct {
	DisplayName *string
	Description *string
	Status      *domain.SectionStatus
	Position    *int
	// FieldSpec replaces the section's whole declaration. A pointer so an
	// omitted spec (leave it alone) is distinguishable from an empty one, which
	// removes every field — and removing a field never deletes the values
	// documents already carry under it.
	FieldSpec *[]domain.FieldSpec
}

type SectionService struct {
	sections   *storage.SectionRepository
	entries    *storage.EntryRepository
	researches *storage.ResearchRepository
	access     *Access
	events     EventNotifier
	log        *slog.Logger
}

func NewSectionService(sections *storage.SectionRepository, entries *storage.EntryRepository, researches *storage.ResearchRepository, access *Access, events EventNotifier, log *slog.Logger) *SectionService {
	return &SectionService{sections: sections, entries: entries, researches: researches, access: access, events: events, log: log}
}

func (s *SectionService) List(ctx context.Context, researchID string) ([]*domain.Section, error) {
	if err := s.access.Read(ctx, researchID); err != nil {
		return nil, err
	}
	sections, err := s.sections.FindByResearch(ctx, researchID)
	if err != nil {
		return nil, err
	}
	redactSectionsForShare(ctx, sections)
	return sections, nil
}

func (s *SectionService) Get(ctx context.Context, id string) (*domain.Section, error) {
	section, err := s.sections.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find section: %w", err)
	}
	if section == nil {
		return nil, ErrNotFound
	}
	if err := s.access.Read(ctx, section.ResearchID); err != nil {
		return nil, ErrNotFound
	}
	redactSectionForShare(ctx, section)
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
	if err := s.access.Write(ctx, section.ResearchID); err != nil {
		return nil, err
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
	if req.FieldSpec != nil {
		specs := *req.FieldSpec
		if errs := domain.ValidateFieldSpecs(specs); len(errs) > 0 {
			return nil, fmt.Errorf("%w: %s", ErrInvalidFieldSpec, strings.Join(errs, "; "))
		}
		for i := range specs {
			specs[i].Label = normalizeTitle(specs[i].Label)
			specs[i].Help = normalizeTitle(specs[i].Help)
		}
		// The version is bumped whenever the declaration changes, and only then.
		// Documents record the version they were validated against, so without
		// this a spec edit would silently restate history: a document written
		// under the old rules would be indistinguishable from one breaking the
		// new ones.
		if !sameFieldSpec(section.FieldSpec, specs) {
			section.SpecVersion++
		}
		section.FieldSpec = specs
	}

	if err := s.sections.Update(ctx, section); err != nil {
		return nil, fmt.Errorf("update section: %w", err)
	}

	emit(ctx, s.events, Event{Type: "section.updated", ResearchID: section.ResearchID, EntityID: section.ID, Entity: "section"})
	return section, nil
}

func (s *SectionService) CountEntries(ctx context.Context, sectionID string) (int, error) {
	return s.entries.CountBySection(ctx, sectionID)
}

// sameFieldSpec reports whether a submitted declaration is the one already
// stored, so saving a form without changing anything does not bump the version
// and re-date every document's compliance.
func sameFieldSpec(a, b []domain.FieldSpec) bool {
	if len(a) != len(b) {
		return false
	}
	ja, err := json.Marshal(a)
	if err != nil {
		return false
	}
	jb, err := json.Marshal(b)
	if err != nil {
		return false
	}
	return string(ja) == string(jb)
}

// redactSectionForShare hides what a team decided to record about its
// documents. The values are stripped on the entry; the declaration is stripped
// here, because a list of twelve field labels with nothing in them still says
// what the team tracks.
func redactSectionForShare(ctx context.Context, section *domain.Section) {
	if section == nil || auth.ShareFromContext(ctx) == nil {
		return
	}
	section.FieldSpec = nil
	section.SpecVersion = 0
}

func redactSectionsForShare(ctx context.Context, sections []*domain.Section) {
	if auth.ShareFromContext(ctx) == nil {
		return
	}
	for _, sec := range sections {
		redactSectionForShare(ctx, sec)
	}
}
