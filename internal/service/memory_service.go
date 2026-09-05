package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/butschster/mcp-research/internal/auth"
	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/storage"
)

func (s *ResearchService) memoryAccess(ctx context.Context, id string, write bool) (string, error) {
	if auth.ShareFromContext(ctx) != nil {
		return "", ErrForbidden
	}
	researchID, err := s.ResolveID(ctx, id)
	if err != nil {
		return "", err
	}
	if write {
		err = s.access.Write(ctx, researchID)
	}
	return researchID, err
}

func (s *ResearchService) ListMemory(ctx context.Context, id string) (domain.Memory, error) {
	researchID, err := s.memoryAccess(ctx, id, false)
	if err != nil {
		return nil, err
	}
	return storage.NewMemoryRepository(s.researches.DB()).List(ctx, researchID)
}

func validateMemoryText(text string) error {
	if strings.TrimSpace(text) == "" {
		return fmt.Errorf("memory text must not be empty: %w", ErrValidation)
	}
	if len(text) > 64000 {
		return fmt.Errorf("memory text exceeds 64000 bytes: %w", ErrValidation)
	}
	return nil
}

func (s *ResearchService) AddMemory(ctx context.Context, id, text, sessionID string) (*domain.MemoryItem, error) {
	researchID, err := s.memoryAccess(ctx, id, true)
	if err != nil {
		return nil, err
	}
	item, err := s.prepareMemory(ctx, researchID, text, sessionID)
	if err != nil {
		return nil, err
	}
	if err := storage.NewMemoryRepository(s.researches.DB()).Create(ctx, researchID, item); err != nil {
		return nil, err
	}
	s.memoryChanged(ctx, researchID, item.ID)
	return item, nil
}

func (s *ResearchService) prepareMemory(ctx context.Context, researchID, text, sessionID string) (*domain.MemoryItem, error) {
	if err := validateMemoryText(text); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	item := &domain.MemoryItem{Text: text, Author: "agent", CreatedAt: &now}
	if AuthorFromContext(ctx) == domain.AuthorHuman {
		item.Author = "user"
	}
	// Only an explicitly identified research session is provenance. The most
	// recently active session could belong to a different concurrent agent.
	if sessionID != "" {
		repo := storage.NewSessionRepository(s.researches.DB())
		session, err := repo.FindByID(ctx, sessionID)
		if err != nil {
			return nil, err
		}
		if session == nil && isCode(sessionID) {
			session, err = repo.FindByCodeAndResearch(ctx, sessionID, researchID)
		}
		if err != nil {
			return nil, err
		}
		if session == nil || session.ResearchID != researchID {
			return nil, ErrNotFound
		}
		item.SessionID, item.SessionCode = session.ID, session.Code
	}
	return item, nil
}

func (s *ResearchService) UpdateMemory(ctx context.Context, id, itemID, text string, version int) error {
	researchID, err := s.memoryAccess(ctx, id, true)
	if err != nil {
		return err
	}
	if err := validateMemoryText(text); err != nil {
		return err
	}
	if version < 1 || itemID == "" {
		return fmt.Errorf("item id and current version are required: %w", ErrValidation)
	}
	err = storage.NewMemoryRepository(s.researches.DB()).Update(ctx, researchID, itemID, text, version)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if errors.Is(err, storage.ErrMemoryConflict) {
		return fmt.Errorf("%s: %w", err, ErrConflict)
	}
	if err != nil {
		return err
	}
	s.memoryChanged(ctx, researchID, itemID)
	return nil
}

func (s *ResearchService) DeleteMemory(ctx context.Context, id string, ids []string) error {
	researchID, err := s.memoryAccess(ctx, id, true)
	if err != nil {
		return err
	}
	if len(ids) == 0 || len(ids) > 500 {
		return fmt.Errorf("select 1–500 memory item IDs: %w", ErrValidation)
	}
	for _, id := range ids {
		if id == "" {
			return fmt.Errorf("empty memory item ID: %w", ErrValidation)
		}
	}
	if err := storage.NewMemoryRepository(s.researches.DB()).Delete(ctx, researchID, ids); err != nil {
		return err
	}
	s.memoryChanged(ctx, researchID, "")
	return nil
}

func (s *ResearchService) memoryChanged(ctx context.Context, researchID, itemID string) {
	emit(ctx, s.events, Event{Type: "research.updated", ResearchID: researchID, EntityID: itemID, Entity: "memory"})
}
