package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/dovod-app/app/internal/auth"
	"github.com/dovod-app/app/internal/domain"
	"github.com/dovod-app/app/internal/storage"
)

const (
	localViewerID     = "local"
	maxSeenBatchItems = 2000
)

// EntryUpdates is the reader's current queue for one research.
type EntryUpdates struct {
	Entries []*domain.EntryUpdate `json:"entries"`
	New     int                   `json:"new"`
	Changed int                   `json:"changed"`
	Count   int                   `json:"count"`
}

// EntryViewService keeps document read state separate from EntryService. An
// entry is shared research content; its checkpoint belongs to one reader and
// must never leak into exports, MCP responses, or public share payloads.
type EntryViewService struct {
	views   *storage.EntryViewRepository
	entries *storage.EntryRepository
	access  *Access
	events  EventNotifier
}

func NewEntryViewService(
	views *storage.EntryViewRepository,
	entries *storage.EntryRepository,
	access *Access,
	events EventNotifier,
) *EntryViewService {
	return &EntryViewService{views: views, entries: entries, access: access, events: events}
}

// List returns only documents this reader has never seen or whose numbered
// history has advanced since they saw it.
func (s *EntryViewService) List(ctx context.Context, researchID string) (*EntryUpdates, error) {
	if auth.ShareFromContext(ctx) != nil {
		return nil, ErrNotFound
	}
	if err := s.access.Read(ctx, researchID); err != nil {
		return nil, err
	}
	viewerID, _ := entryViewer(ctx)
	entries, err := s.views.ListUpdates(ctx, viewerID, researchID)
	if err != nil {
		return nil, err
	}
	out := &EntryUpdates{Entries: entries, Count: len(entries)}
	for _, entry := range entries {
		if entry.Kind == "new" {
			out.New++
		} else {
			out.Changed++
		}
	}
	return out, nil
}

// StateAt describes one exact snapshot already loaded by the entry handler.
// Passing currentRevision in is deliberate: asking the database for "latest"
// again could race with the document response and claim the reader was shown a
// revision that was written between the two reads.
func (s *EntryViewService) StateAt(
	ctx context.Context,
	entry *domain.Entry,
	currentRevision int,
) (*domain.EntryViewState, error) {
	if entry == nil || currentRevision < 1 || auth.ShareFromContext(ctx) != nil {
		return nil, nil
	}
	if err := s.access.Read(ctx, entry.ResearchID); err != nil {
		return nil, err
	}
	viewerID, _ := entryViewer(ctx)
	view, err := s.views.Find(ctx, viewerID, entry.ID)
	if err != nil {
		return nil, err
	}
	state := &domain.EntryViewState{CurrentRevision: currentRevision}
	if view == nil {
		state.Kind = "new"
		state.UnseenRevisions = currentRevision
		return state, nil
	}
	state.SeenRevision = view.SeenRevision
	if view.SeenRevision < currentRevision {
		state.Kind = "changed"
		state.UnseenRevisions = currentRevision - view.SeenRevision
	} else {
		state.Kind = "seen"
	}
	return state, nil
}

// MarkSeen advances a checkpoint to the revision the client actually rendered.
// Read permission is enough: viewers are readers too, and this changes none of
// the team's content.
func (s *EntryViewService) MarkSeen(ctx context.Context, entryID string, revision int) error {
	if revision < 1 {
		return fmt.Errorf("revision must be positive: %w", ErrValidation)
	}
	if auth.ShareFromContext(ctx) != nil {
		return ErrNotFound
	}
	entry, err := s.entries.FindByID(ctx, entryID)
	if err != nil {
		return fmt.Errorf("find entry: %w", err)
	}
	if entry == nil {
		return ErrNotFound
	}
	if err := s.access.Read(ctx, entry.ResearchID); err != nil {
		return err
	}
	viewerID, userID := entryViewer(ctx)
	ok, err := s.views.MarkSeen(ctx, viewerID, userID, entry.ResearchID, entry.ID, revision)
	if err != nil {
		return err
	}
	if !ok {
		// Same answer for a revision that does not exist and one that belongs to
		// some other document: the supplied pair is not visible to this caller.
		return ErrNotFound
	}
	s.notify(ctx, entry.ResearchID, entry.ID)
	return nil
}

// MarkSeenMany advances exactly the entry/revision pairs the updates page
// displayed. A revision written after that snapshot remains in the queue.
func (s *EntryViewService) MarkSeenMany(
	ctx context.Context,
	researchID string,
	targets []domain.SeenRevision,
) error {
	if auth.ShareFromContext(ctx) != nil {
		return ErrNotFound
	}
	if err := s.access.Read(ctx, researchID); err != nil {
		return err
	}
	if len(targets) > maxSeenBatchItems {
		return fmt.Errorf("at most %d documents may be marked at once: %w", maxSeenBatchItems, ErrValidation)
	}

	// Duplicate rows make the reported operation ambiguous and add no value.
	// Collapse exact duplicates, reject two revisions for one entry.
	unique := make([]domain.SeenRevision, 0, len(targets))
	byEntry := make(map[string]int, len(targets))
	for _, target := range targets {
		if target.EntryID == "" || target.Revision < 1 {
			return fmt.Errorf("entry_id and a positive revision are required: %w", ErrValidation)
		}
		if previous, ok := byEntry[target.EntryID]; ok {
			if previous != target.Revision {
				return fmt.Errorf("entry %s appears with two revisions: %w", target.EntryID, ErrValidation)
			}
			continue
		}
		byEntry[target.EntryID] = target.Revision
		unique = append(unique, target)
	}

	viewerID, userID := entryViewer(ctx)
	if err := s.views.MarkSeenMany(ctx, viewerID, userID, researchID, unique); err != nil {
		if errors.Is(err, storage.ErrEntryViewTargetNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("mark document updates seen: %w", err)
	}
	s.notify(ctx, researchID, researchID)
	return nil
}

func entryViewer(ctx context.Context) (viewerID, userID string) {
	userID = auth.UserIDFromContext(ctx)
	if userID == "" {
		return localViewerID, ""
	}
	return userID, userID
}

func (s *EntryViewService) notify(ctx context.Context, researchID, entityID string) {
	if s.events == nil {
		return
	}
	uid := auth.UserIDFromContext(ctx)
	emit(ctx, s.events, Event{
		Type:         "entry_view.updated",
		ResearchID:   researchID,
		EntityID:     entityID,
		Entity:       "entry_view",
		TargetUserID: uid,
	})
}
