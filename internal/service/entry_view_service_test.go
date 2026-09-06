package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/dovod-app/app/internal/auth"
	"github.com/dovod-app/app/internal/domain"
	"github.com/dovod-app/app/internal/storage"
)

func entryViewFixture(t *testing.T, role domain.TeamRole) (*roleKit, *EntryViewService, context.Context, context.Context, *domain.Research, *domain.Section) {
	t.Helper()
	k := newRoleKit(t)
	owner, member, research, section, _ := k.sharedResearch(t, role)
	views := NewEntryViewService(
		storage.NewEntryViewRepository(k.db),
		storage.NewEntryRepository(k.db),
		testAccess(k.db),
		k.events,
	)
	return k, views, owner, member, research, section
}

func TestEntryViews_ArePersonalAndViewersCanAdvanceTheirOwnCheckpoint(t *testing.T) {
	k, views, owner, member, research, section := entryViewFixture(t, domain.TeamViewer)
	entry := mustEntry(t, k.entry, owner, research.ID, section.ID, "First version.")

	ownerUpdates, err := views.List(owner, research.ID)
	if err != nil {
		t.Fatalf("list owner updates: %v", err)
	}
	memberUpdates, err := views.List(member, research.ID)
	if err != nil {
		t.Fatalf("list member updates: %v", err)
	}
	for who, updates := range map[string]*EntryUpdates{"owner": ownerUpdates, "member": memberUpdates} {
		if updates.Count != 1 || updates.New != 1 || updates.Changed != 0 {
			t.Fatalf("%s initial queue = %+v, want one new document", who, updates)
		}
		got := updates.Entries[0]
		if got.EntryID != entry.ID || got.Kind != "new" || got.SeenRevision != 0 || got.CurrentRevision != 1 {
			t.Errorf("%s initial update = %+v", who, got)
		}
	}

	// A viewer cannot edit team content, but can record their own read state.
	if err := views.MarkSeen(member, entry.ID, 1); err != nil {
		t.Fatalf("viewer marks seen: %v", err)
	}
	last := k.events.events[len(k.events.events)-1]
	if last.Type != "entry_view.updated" || last.TargetUserID != auth.UserIDFromContext(member) {
		t.Errorf("view event must be private to its reader: %+v", last)
	}
	memberUpdates, _ = views.List(member, research.ID)
	if memberUpdates.Count != 0 {
		t.Fatalf("viewer's queue should be empty after viewing: %+v", memberUpdates)
	}
	ownerUpdates, _ = views.List(owner, research.ID)
	if ownerUpdates.Count != 1 || ownerUpdates.Entries[0].Kind != "new" {
		t.Fatalf("one reader's checkpoint changed another's queue: %+v", ownerUpdates)
	}

	if _, err := k.entry.Update(owner, entry.ID, UpdateEntryRequest{Content: ptr("Second version.")}); err != nil {
		t.Fatalf("update entry: %v", err)
	}
	memberUpdates, _ = views.List(member, research.ID)
	if memberUpdates.Count != 1 {
		t.Fatalf("updated document missing from queue: %+v", memberUpdates)
	}
	changed := memberUpdates.Entries[0]
	if changed.Kind != "changed" || changed.SeenRevision != 1 || changed.CurrentRevision != 2 || changed.UnseenRevisions != 1 {
		t.Errorf("changed update = %+v", changed)
	}

	state, err := views.StateAt(member, entry, 2)
	if err != nil {
		t.Fatalf("read entry view state: %v", err)
	}
	if state.Kind != "changed" || state.SeenRevision != 1 || state.CurrentRevision != 2 {
		t.Errorf("entry view state = %+v", state)
	}

}

func TestEntryViews_MarkAllUsesTheDisplayedRevisionSnapshot(t *testing.T) {
	k, views, owner, _, research, section := entryViewFixture(t, domain.TeamEditor)
	first := mustEntry(t, k.entry, owner, research.ID, section.ID, "First v1")
	second := mustEntry(t, k.entry, owner, research.ID, section.ID, "Second v1")

	snapshot, err := views.List(owner, research.ID)
	if err != nil {
		t.Fatalf("list snapshot: %v", err)
	}
	targets := make([]domain.SeenRevision, 0, len(snapshot.Entries))
	for _, update := range snapshot.Entries {
		targets = append(targets, domain.SeenRevision{EntryID: update.EntryID, Revision: update.CurrentRevision})
	}

	// A new revision lands after the page has loaded but before the click.
	if _, err := k.entry.Update(owner, first.ID, UpdateEntryRequest{Content: ptr("First v2")}); err != nil {
		t.Fatalf("concurrent update: %v", err)
	}
	if err := views.MarkSeenMany(owner, research.ID, targets); err != nil {
		t.Fatalf("mark displayed snapshot: %v", err)
	}

	remaining, err := views.List(owner, research.ID)
	if err != nil {
		t.Fatalf("list remaining: %v", err)
	}
	if remaining.Count != 1 || remaining.Entries[0].EntryID != first.ID {
		t.Fatalf("only the concurrent update should remain: %+v", remaining.Entries)
	}
	got := remaining.Entries[0]
	if got.Kind != "changed" || got.SeenRevision != 1 || got.CurrentRevision != 2 {
		t.Errorf("concurrent update checkpoint = %+v", got)
	}
	if secondState, _ := views.StateAt(owner, second, 1); secondState.Kind != "seen" {
		t.Errorf("unchanged snapshot entry was not marked seen: %+v", secondState)
	}
}

func TestEntryViews_CheckpointsNeverMoveBackward(t *testing.T) {
	k, views, owner, _, research, section := entryViewFixture(t, domain.TeamEditor)
	entry := mustEntry(t, k.entry, owner, research.ID, section.ID, "v1")
	for revision := 2; revision <= 3; revision++ {
		body := fmt.Sprintf("v%d", revision)
		if _, err := k.entry.Update(owner, entry.ID, UpdateEntryRequest{Content: &body}); err != nil {
			t.Fatalf("update %d: %v", revision, err)
		}
	}
	if err := views.MarkSeen(owner, entry.ID, 3); err != nil {
		t.Fatalf("mark revision 3: %v", err)
	}
	if err := views.MarkSeen(owner, entry.ID, 1); err != nil {
		t.Fatalf("late mark revision 1: %v", err)
	}
	state, err := views.StateAt(owner, entry, 3)
	if err != nil {
		t.Fatalf("state: %v", err)
	}
	if state.Kind != "seen" || state.SeenRevision != 3 {
		t.Fatalf("late tab moved checkpoint backward: %+v", state)
	}
}

func TestEntryViews_BulkValidationIsAtomicAndAccessControlled(t *testing.T) {
	k, views, owner, member, research, section := entryViewFixture(t, domain.TeamViewer)
	first := mustEntry(t, k.entry, owner, research.ID, section.ID, "First")
	second := mustEntry(t, k.entry, owner, research.ID, section.ID, "Second")

	err := views.MarkSeenMany(member, research.ID, []domain.SeenRevision{
		{EntryID: first.ID, Revision: 1},
		{EntryID: second.ID, Revision: 99},
	})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("bad revision error = %v, want not found", err)
	}
	updates, _ := views.List(member, research.ID)
	if updates.Count != 2 {
		t.Fatalf("invalid batch partly committed: %+v", updates.Entries)
	}

	stranger := createTestUser(t, k.db, "stranger-updates@test.com", "Stranger")
	if _, err := views.List(userCtx(stranger), research.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("stranger list error = %v, want not found", err)
	}
	if err := views.MarkSeen(userCtx(stranger), first.ID, 1); !errors.Is(err, ErrNotFound) {
		t.Errorf("stranger mark error = %v, want not found", err)
	}
	if err := views.MarkSeenMany(userCtx(stranger), research.ID, []domain.SeenRevision{{EntryID: first.ID, Revision: 1}}); !errors.Is(err, ErrNotFound) {
		t.Errorf("stranger bulk mark error = %v, want not found", err)
	}

	shareCtx := auth.WithShare(t.Context(), &auth.Share{ID: "public-link", ResearchID: research.ID})
	if _, err := views.List(shareCtx, research.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("share list error = %v, want not found", err)
	}
	if err := views.MarkSeen(shareCtx, first.ID, 1); !errors.Is(err, ErrNotFound) {
		t.Errorf("share mark error = %v, want not found", err)
	}
	if err := views.MarkSeenMany(shareCtx, research.ID, []domain.SeenRevision{{EntryID: first.ID, Revision: 1}}); !errors.Is(err, ErrNotFound) {
		t.Errorf("share bulk mark error = %v, want not found", err)
	}
}

func TestEntryViews_TrimPreservesEveryReadersDiffBase(t *testing.T) {
	k, views, owner, member, research, section := entryViewFixture(t, domain.TeamViewer)
	k.entry.SetRevisionLimit(2)
	entry := mustEntry(t, k.entry, owner, research.ID, section.ID, "v1")

	if _, err := k.entry.Update(owner, entry.ID, UpdateEntryRequest{Content: ptr("v2")}); err != nil {
		t.Fatalf("update v2: %v", err)
	}
	if err := views.MarkSeen(member, entry.ID, 2); err != nil {
		t.Fatalf("member marks v2: %v", err)
	}
	for revision := 3; revision <= 6; revision++ {
		body := fmt.Sprintf("v%d", revision)
		if _, err := k.entry.Update(owner, entry.ID, UpdateEntryRequest{Content: &body}); err != nil {
			t.Fatalf("update %d: %v", revision, err)
		}
	}

	// Revision 2 is older than the normal retention window, but remains the
	// exact base the member needs for "changed since you last viewed it".
	diff, err := k.entry.Diff(member, entry.ID, 2, 6)
	if err != nil {
		t.Fatalf("diff from retained checkpoint: %v", err)
	}
	if diff.From == nil || diff.From.Revision != 2 || diff.To.Revision != 6 {
		t.Fatalf("retained diff = %+v -> %+v", diff.From, diff.To)
	}
}
