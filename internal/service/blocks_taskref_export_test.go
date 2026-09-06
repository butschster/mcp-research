package service

import (
	"context"
	"strings"
	"testing"

	"github.com/dovod-app/app/internal/domain"
)

// A task_ref block is the one block whose export depends on data the exporter is
// not always allowed to see. These tests pin both halves: it resolves when the
// tasks are in the vault, and it says nothing about them when they are not.

func seedTaskRefVault(t *testing.T, f *obsidianFixture) *domain.Research {
	t.Helper()
	ctx := context.Background()

	research, sections, err := f.research.Create(ctx, CreateResearchRequest{
		Name:     "Rollout",
		Goal:     "Ship it",
		Sections: []CreateSectionRequest{{Name: "plan", DisplayName: "Plan", Position: 1}},
	})
	if err != nil {
		t.Fatalf("create research: %v", err)
	}

	done, err := f.task.Create(ctx, CreateTaskRequest{
		ResearchID: research.ID, Title: "Migrate the schema", Priority: domain.PriorityHigh,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	if _, err := f.task.Update(ctx, done.ID, UpdateTaskRequest{Status: ptr(domain.TaskCompleted)}); err != nil {
		t.Fatalf("complete task: %v", err)
	}
	if _, err := f.task.Create(ctx, CreateTaskRequest{
		ResearchID: research.ID, Title: "Measure the fallout", Priority: domain.PriorityMedium,
	}); err != nil {
		t.Fatalf("create task 2: %v", err)
	}

	if _, err := f.entry.Create(ctx, CreateEntryRequest{
		ResearchID: research.ID,
		SectionID:  sections[0].ID,
		Type:       domain.EntryBlocks,
		Title:      "Before the call",
		Content: `{"version":1,"blocks":[
			{"type":"paragraph","data":{"text":"What is left."},"id":"para0001"},
			{"type":"task_ref","data":{"tasks":["T1","T2"],"note":"Close before Friday"},"id":"tref0001"}
		]}`,
	}); err != nil {
		t.Fatalf("create entry: %v", err)
	}
	return research
}

func TestObsidian_TaskRefResolvesWhenTheVaultCarriesTasks(t *testing.T) {
	f := setupObsidian(t)
	research := seedTaskRefVault(t, f)

	v := vaultOf(t, f, research.ID, DefaultVaultOptions())
	note := noteContaining(t, v, "Before the call")

	for _, want := range []string{
		"Close before Friday",
		"- [x] T1 — Migrate the schema",
		"- [ ] T2 — Measure the fallout",
		"1 of 2 done",
	} {
		if !strings.Contains(note, want) {
			t.Errorf("missing %q from the note:\n%s", want, note)
		}
	}
}

func TestObsidian_TaskRefLeaksNoTitleWhenTasksAreExcluded(t *testing.T) {
	// This is the share case in the shape the vault sees it: clampForShare turns
	// a link with tasks switched off into Tasks:false, and from there the block
	// must degrade to references. A title printed here would hand over through a
	// document exactly what the flag refused through the route.
	f := setupObsidian(t)
	research := seedTaskRefVault(t, f)

	v := vaultOf(t, f, research.ID, VaultOptions{HTML: true})
	note := noteContaining(t, v, "Before the call")

	for _, leaked := range []string{"Migrate the schema", "Measure the fallout", "- [x]", "done"} {
		if strings.Contains(note, leaked) {
			t.Errorf("a vault built without tasks printed %q:\n%s", leaked, note)
		}
	}
	// The author's own note is theirs and stays; the references stay as links.
	if !strings.Contains(note, "Close before Friday") {
		t.Errorf("the author's note was dropped with the tasks:\n%s", note)
	}
	if !strings.Contains(note, "T1") || !strings.Contains(note, "T2") {
		t.Errorf("the references themselves were dropped:\n%s", note)
	}
}

// noteContaining returns the one vault file whose body mentions needle.
func noteContaining(t *testing.T, v *Vault, needle string) string {
	t.Helper()
	for _, f := range v.Files {
		if strings.Contains(string(f.Content), needle) {
			return string(f.Content)
		}
	}
	t.Fatalf("no vault file mentions %q; files: %v", needle, pathsOf(v))
	return ""
}
