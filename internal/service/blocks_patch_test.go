package service

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/storage"
	"github.com/uptrace/bun"
)

func patchFixture(t *testing.T) (*EntryService, context.Context, *domain.Entry) {
	t.Helper()
	db := setupTestDB(t)
	log := slog.Default()

	researchRepo := storage.NewResearchRepository(db)
	sectionRepo := storage.NewSectionRepository(db)
	entryRepo := storage.NewEntryRepository(db)
	blockRepo := storage.NewBlockRepository(db)
	crossrefRepo := storage.NewCrossRefRepository(db)
	researchSvc := NewResearchService(researchRepo, sectionRepo, storage.NewTeamRepository(db), testAccess(db), &mockNotifier{}, log)
	entrySvc := NewEntryService(entryRepo, sectionRepo, researchRepo, testAccess(db), nil, blockRepo, storage.NewEntryRevisionRepository(db), crossrefRepo, nil, &mockNotifier{}, log)

	ctx := context.Background()
	research, sections, err := researchSvc.Create(ctx, CreateResearchRequest{
		Name: "R", Goal: "T",
		Sections: []CreateSectionRequest{{Name: "s1", DisplayName: "S1"}},
	})
	if err != nil {
		t.Fatalf("create research: %v", err)
	}

	doc := `{"blocks":[
		{"id":"aaaa1111","type":"heading","data":{"level":2,"text":"Runbook"}},
		{"id":"bbbb2222","type":"paragraph","data":{"text":"Before you start."}},
		{"id":"cccc3333","type":"checklist","data":{"items":[
			{"key":"k1","text":"Back up the database"},
			{"key":"k2","text":"Run the migration on a copy"}
		]}}
	]}`
	entry, err := entrySvc.Create(ctx, CreateEntryRequest{
		ResearchID: research.ID,
		SectionID:  sections[0].ID,
		Title:      "Runbook",
		Content:    doc,
		Type:       domain.EntryBlocks,
	})
	if err != nil {
		t.Fatalf("create entry: %v", err)
	}
	return entrySvc, ctx, entry
}

func blocksOf(t *testing.T, svc *EntryService, ctx context.Context, id string) []domain.Block {
	t.Helper()
	entry, err := svc.Get(ctx, id)
	if err != nil {
		t.Fatalf("get entry: %v", err)
	}
	doc, err := svc.LoadBlockDocument(ctx, entry)
	if err != nil {
		t.Fatalf("load document: %v", err)
	}
	return doc.Blocks
}

func TestPatchBlocks_Structural(t *testing.T) {
	svc, ctx, entry := patchFixture(t)

	t.Run("insert lands where the anchor says", func(t *testing.T) {
		if _, err := svc.PatchBlocks(ctx, entry.ID, PatchBlocksRequest{Ops: []BlockOp{{
			Op: OpInsert, ID: "dddd4444", Type: "paragraph",
			Data: map[string]any{"text": "Inserted."}, After: "aaaa1111",
		}}}); err != nil {
			t.Fatalf("patch: %v", err)
		}
		got := blocksOf(t, svc, ctx, entry.ID)
		if len(got) != 4 || got[1].ID != "dddd4444" {
			t.Fatalf("order = %v", ids(got))
		}
	})

	t.Run("ops see each other, in order", func(t *testing.T) {
		if _, err := svc.PatchBlocks(ctx, entry.ID, PatchBlocksRequest{Ops: []BlockOp{
			{Op: OpInsert, ID: "eeee5555", Type: "paragraph", Data: map[string]any{"text": "First."}, At: "start"},
			{Op: OpInsert, Type: "paragraph", Data: map[string]any{"text": "Second."}, After: "eeee5555"},
		}}); err != nil {
			t.Fatalf("patch: %v", err)
		}
		got := blocksOf(t, svc, ctx, entry.ID)
		if got[0].ID != "eeee5555" || got[1].Data["text"] != "Second." {
			t.Fatalf("order = %v", ids(got))
		}
	})

	t.Run("update replaces one block and leaves the rest", func(t *testing.T) {
		before := blocksOf(t, svc, ctx, entry.ID)
		if _, err := svc.PatchBlocks(ctx, entry.ID, PatchBlocksRequest{Ops: []BlockOp{{
			Op: OpUpdate, ID: "bbbb2222", Data: map[string]any{"text": "Rewritten."},
		}}}); err != nil {
			t.Fatalf("patch: %v", err)
		}
		after := blocksOf(t, svc, ctx, entry.ID)
		if len(after) != len(before) {
			t.Fatalf("block count changed: %d → %d", len(before), len(after))
		}
		for _, b := range after {
			if b.ID == "bbbb2222" && b.Data["text"] != "Rewritten." {
				t.Errorf("text = %v", b.Data["text"])
			}
		}
	})

	t.Run("move and delete address by id", func(t *testing.T) {
		if _, err := svc.PatchBlocks(ctx, entry.ID, PatchBlocksRequest{Ops: []BlockOp{
			{Op: OpMove, ID: "cccc3333", At: "start"},
			{Op: OpDelete, ID: "dddd4444"},
		}}); err != nil {
			t.Fatalf("patch: %v", err)
		}
		got := blocksOf(t, svc, ctx, entry.ID)
		if got[0].ID != "cccc3333" {
			t.Errorf("first block = %q, want the moved one", got[0].ID)
		}
		if indexOf(got, "dddd4444") >= 0 {
			t.Error("deleted block still present")
		}
	})
}

// The whole point of the strict/forgiving split: a whole-document write drops a
// bad block, a patch refuses and writes nothing.
func TestPatchBlocks_StrictAndAtomic(t *testing.T) {
	svc, ctx, entry := patchFixture(t)
	before := blocksOf(t, svc, ctx, entry.ID)

	cases := []struct {
		name string
		ops  []BlockOp
		want string
	}{
		{"unknown block id", []BlockOp{{Op: OpUpdate, ID: "nope0000", Data: map[string]any{"text": "x"}}}, "no block"},
		{"unknown anchor", []BlockOp{{Op: OpInsert, Type: "paragraph", Data: map[string]any{"text": "x"}, After: "nope0000"}}, "no block"},
		{"unknown op", []BlockOp{{Op: "frobnicate", ID: "aaaa1111"}}, "unknown op"},
		{"unknown type", []BlockOp{{Op: OpInsert, Type: "spreadsheet", Data: map[string]any{"text": "x"}}}, "rejected"},
		{"data the type rejects", []BlockOp{{Op: OpUpdate, ID: "bbbb2222", Data: map[string]any{"txt": "typo in the field name"}}}, "rejected"},
		{"two anchors", []BlockOp{{Op: OpInsert, Type: "paragraph", Data: map[string]any{"text": "x"}, After: "aaaa1111", Before: "bbbb2222"}}, "one of after"},
		{"a good op after a bad one is not applied", []BlockOp{
			{Op: OpInsert, Type: "paragraph", Data: map[string]any{"text": "fine"}},
			{Op: OpDelete, ID: "nope0000"},
		}, "no block"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.PatchBlocks(ctx, entry.ID, PatchBlocksRequest{Ops: tc.ops})
			if err == nil {
				t.Fatal("expected an error, got none")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %q, want it to mention %q", err, tc.want)
			}
			if got := blocksOf(t, svc, ctx, entry.ID); len(got) != len(before) {
				t.Errorf("document changed after a failed patch: %d → %d blocks", len(before), len(got))
			}
		})
	}

	t.Run("the document cannot be emptied", func(t *testing.T) {
		ops := make([]BlockOp, 0, len(before))
		for _, b := range before {
			ops = append(ops, BlockOp{Op: OpDelete, ID: b.ID})
		}
		if _, err := svc.PatchBlocks(ctx, entry.ID, PatchBlocksRequest{Ops: ops}); err == nil {
			t.Fatal("expected a refusal")
		}
	})
}

func TestPatchBlocks_Rev(t *testing.T) {
	svc, ctx, entry := patchFixture(t)
	rev := DocumentRev(entry.Content)

	if _, err := svc.PatchBlocks(ctx, entry.ID, PatchBlocksRequest{
		Rev: rev,
		Ops: []BlockOp{{Op: OpUpdate, ID: "bbbb2222", Data: map[string]any{"text": "First writer."}}},
	}); err != nil {
		t.Fatalf("patch with a fresh rev: %v", err)
	}

	// The second writer read the same document and is now stale.
	_, err := svc.PatchBlocks(ctx, entry.ID, PatchBlocksRequest{
		Rev: rev,
		Ops: []BlockOp{{Op: OpUpdate, ID: "bbbb2222", Data: map[string]any{"text": "Second writer."}}},
	})
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
	for _, b := range blocksOf(t, svc, ctx, entry.ID) {
		if b.ID == "bbbb2222" && b.Data["text"] != "First writer." {
			t.Errorf("stale write landed: %v", b.Data["text"])
		}
	}
}

// The requirement this whole design exists for: a human ticks, an agent rewrites
// the document, and the tick is still there.
func TestChecklistStateSurvivesRewrite(t *testing.T) {
	svc, ctx, entry := patchFixture(t)

	tick := func(item string, checked bool) {
		t.Helper()
		if _, err := svc.PatchBlocks(ctx, entry.ID, PatchBlocksRequest{Ops: []BlockOp{{
			Op: OpSetState, ID: "cccc3333", Item: item, Checked: &checked,
		}}}); err != nil {
			t.Fatalf("set_state: %v", err)
		}
	}
	checkedItems := func() map[string]bool {
		t.Helper()
		out := map[string]bool{}
		for _, b := range blocksOf(t, svc, ctx, entry.ID) {
			for _, item := range checklistItems(b.Data) {
				if item.Checked {
					out[item.Key] = true
				}
			}
		}
		return out
	}

	tick("k1", true)
	if !checkedItems()["k1"] {
		t.Fatal("the tick did not stick")
	}

	t.Run("an agent rewriting the document with ids intact keeps it", func(t *testing.T) {
		rewritten := `{"blocks":[
			{"id":"aaaa1111","type":"heading","data":{"level":2,"text":"Runbook, revised"}},
			{"id":"bbbb2222","type":"paragraph","data":{"text":"Rewritten intro."}},
			{"id":"cccc3333","type":"checklist","data":{"items":[
				{"key":"k1","text":"Back up the database"},
				{"key":"k2","text":"Run the migration on a copy"},
				{"key":"k3","text":"Announce the window"}
			]}}
		]}`
		updated, err := svc.Update(ctx, entry.ID, UpdateEntryRequest{Content: &rewritten})
		if err != nil {
			t.Fatalf("update: %v", err)
		}
		if !checkedItems()["k1"] {
			t.Error("the human's tick was lost to an agent rewrite")
		}
		if updated.BlockReport == nil || updated.BlockReport.StateLost != 0 {
			t.Errorf("report = %+v, want no loss", updated.BlockReport)
		}
	})

	t.Run("an agent that mints fresh ids is caught by the item text", func(t *testing.T) {
		regenerated := `{"blocks":[
			{"type":"heading","data":{"level":2,"text":"Runbook"}},
			{"type":"checklist","data":{"items":[
				{"text":"Back up the database"},
				{"text":"Run the migration on a copy"}
			]}}
		]}`
		updated, err := svc.Update(ctx, entry.ID, UpdateEntryRequest{Content: &regenerated})
		if err != nil {
			t.Fatalf("update: %v", err)
		}
		if len(checkedItems()) != 1 {
			t.Errorf("ticks after a re-identified rewrite = %v, want the one that was set", checkedItems())
		}
		if updated.BlockReport == nil || updated.BlockReport.Reidentified == 0 {
			t.Errorf("report = %+v, want it to say ids were not recognized", updated.BlockReport)
		}
	})

	t.Run("an author cannot write state directly", func(t *testing.T) {
		forged := `{"blocks":[
			{"id":"cccc3333","type":"checklist","data":{
				"items":[{"key":"k1","text":"Back up the database"},{"key":"k2","text":"Run the migration on a copy"}],
				"state":{"k2":true}
			}}
		]}`
		if _, err := svc.Update(ctx, entry.ID, UpdateEntryRequest{Content: &forged}); err != nil {
			t.Fatalf("update: %v", err)
		}
		if checkedItems()["k2"] {
			t.Error("state sent by an author was stored; it must have exactly one origin")
		}
	})
}

func TestSetState_Refusals(t *testing.T) {
	svc, ctx, entry := patchFixture(t)
	yes := true

	for _, tc := range []struct {
		name string
		op   BlockOp
		want string
	}{
		{"a block that is not a checklist", BlockOp{Op: OpSetState, ID: "bbbb2222", Item: "k1", Checked: &yes}, "not a checklist"},
		{"an item the block does not have", BlockOp{Op: OpSetState, ID: "cccc3333", Item: "k9", Checked: &yes}, "no item"},
		{"a block that does not exist", BlockOp{Op: OpSetState, ID: "nope0000", Item: "k1", Checked: &yes}, "no block"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.PatchBlocks(ctx, entry.ID, PatchBlocksRequest{Ops: []BlockOp{tc.op}})
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want it to mention %q", err, tc.want)
			}
		})
	}
}

// The projection in entries.content is what search, both exports and the web UI
// read. It must always describe the rows beside it.
func TestProjectionMatchesRows(t *testing.T) {
	svc, ctx, entry := patchFixture(t)
	yes := true
	if _, err := svc.PatchBlocks(ctx, entry.ID, PatchBlocksRequest{Ops: []BlockOp{
		{Op: OpSetState, ID: "cccc3333", Item: "k1", Checked: &yes},
		{Op: OpInsert, Type: "paragraph", Data: map[string]any{"text": "Tail."}},
	}}); err != nil {
		t.Fatalf("patch: %v", err)
	}

	stored, err := svc.Get(ctx, entry.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	fromRows, err := svc.LoadBlockDocument(ctx, stored)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	want, err := MarshalBlockDocument(fromRows)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if stored.Content != want {
		t.Errorf("projection drifted from the rows:\n stored: %s\n rows:   %s", stored.Content, want)
	}

	var doc map[string]any
	if err := json.Unmarshal([]byte(stored.Content), &doc); err != nil {
		t.Fatalf("projection is not a document: %v", err)
	}
}

func ids(blocks []domain.Block) []string {
	out := make([]string, 0, len(blocks))
	for _, b := range blocks {
		out = append(out, b.ID)
	}
	return out
}

// Every one of these was a real defect found by reviewing the first cut.
func TestBlockDefectsFoundInReview(t *testing.T) {
	t.Run("unticking the last item of a block sticks", func(t *testing.T) {
		svc, ctx, entry := patchFixture(t)
		on, off := true, false
		if _, err := svc.PatchBlocks(ctx, entry.ID, PatchBlocksRequest{Ops: []BlockOp{
			{Op: OpSetState, ID: "cccc3333", Item: "k1", Checked: &on},
		}}); err != nil {
			t.Fatalf("tick: %v", err)
		}
		if _, err := svc.PatchBlocks(ctx, entry.ID, PatchBlocksRequest{Ops: []BlockOp{
			{Op: OpSetState, ID: "cccc3333", Item: "k1", Checked: &off},
		}}); err != nil {
			t.Fatalf("untick: %v", err)
		}
		for _, b := range blocksOf(t, svc, ctx, entry.ID) {
			for _, item := range checklistItems(b.Data) {
				if item.Checked {
					t.Errorf("item %q is still ticked after being unticked", item.Key)
				}
			}
		}
	})

	t.Run("a rewrite that re-keys items reports the ticks it killed", func(t *testing.T) {
		svc, ctx, entry := patchFixture(t)
		on := true
		if _, err := svc.PatchBlocks(ctx, entry.ID, PatchBlocksRequest{Ops: []BlockOp{
			{Op: OpSetState, ID: "cccc3333", Item: "k1", Checked: &on},
		}}); err != nil {
			t.Fatalf("tick: %v", err)
		}
		// Block id kept, items sent as plain strings — which the catalog allows,
		// and which mints fresh keys. The ticks cannot survive; the report used to
		// claim they had.
		rewritten := `{"blocks":[
			{"id":"cccc3333","type":"checklist","data":{"items":["Something else entirely","And another"]}}
		]}`
		updated, err := svc.Update(ctx, entry.ID, UpdateEntryRequest{Content: &rewritten})
		if err != nil {
			t.Fatalf("update: %v", err)
		}
		if updated.BlockReport == nil || updated.BlockReport.StateLost != 1 {
			t.Errorf("report = %+v, want state_lost 1", updated.BlockReport)
		}
	})

	t.Run("one tick revives at most one item", func(t *testing.T) {
		svc, ctx, entry := patchFixture(t)
		on := true
		if _, err := svc.PatchBlocks(ctx, entry.ID, PatchBlocksRequest{Ops: []BlockOp{
			{Op: OpSetState, ID: "cccc3333", Item: "k1", Checked: &on},
		}}); err != nil {
			t.Fatalf("tick: %v", err)
		}
		// Two checklists, fresh ids, both containing the ticked text.
		rewritten := `{"blocks":[
			{"type":"checklist","data":{"title":"Prod","items":[{"text":"Back up the database"}]}},
			{"type":"checklist","data":{"title":"Staging","items":[{"text":"back up   the Database"}]}}
		]}`
		if _, err := svc.Update(ctx, entry.ID, UpdateEntryRequest{Content: &rewritten}); err != nil {
			t.Fatalf("update: %v", err)
		}
		ticks := 0
		for _, b := range blocksOf(t, svc, ctx, entry.ID) {
			for _, item := range checklistItems(b.Data) {
				if item.Checked {
					ticks++
				}
			}
		}
		if ticks != 1 {
			t.Errorf("ticks = %d, want 1 — one tick must not multiply across blocks", ticks)
		}
	})

	t.Run("an id the caller chose is used or refused, never swapped", func(t *testing.T) {
		svc, ctx, entry := patchFixture(t)
		_, err := svc.PatchBlocks(ctx, entry.ID, PatchBlocksRequest{Ops: []BlockOp{
			{Op: OpInsert, ID: "todo-1", Type: "paragraph", Data: map[string]any{"text": "x"}},
		}})
		if err == nil || !strings.Contains(err.Error(), "not usable") {
			t.Fatalf("error = %v, want a refusal naming the id rule", err)
		}
		_, err = svc.PatchBlocks(ctx, entry.ID, PatchBlocksRequest{Ops: []BlockOp{
			{Op: OpInsert, ID: "aaaa1111", Type: "paragraph", Data: map[string]any{"text": "x"}},
		}})
		if err == nil || !strings.Contains(err.Error(), "already used") {
			t.Fatalf("error = %v, want a refusal for a duplicate id", err)
		}
	})

	t.Run("a document whose rows are missing is not truncated by a patch", func(t *testing.T) {
		svc, ctx, entry := patchFixture(t)
		// Exactly the state of every blocks entry written before migration 017.
		if err := svc.blocks.DeleteForEntry(ctx, nil, entry.ID); err != nil {
			t.Fatalf("clear rows: %v", err)
		}
		if _, err := svc.PatchBlocks(ctx, entry.ID, PatchBlocksRequest{Ops: []BlockOp{
			{Op: OpInsert, Type: "paragraph", Data: map[string]any{"text": "appended"}},
		}}); err != nil {
			t.Fatalf("patch: %v", err)
		}
		got := blocksOf(t, svc, ctx, entry.ID)
		if len(got) != 4 {
			t.Fatalf("document holds %d blocks, want the original 3 plus the new one", len(got))
		}
	})

	t.Run("ticks survive an export and import round trip", func(t *testing.T) {
		svc, ctx, entry := patchFixture(t)
		on := true
		if _, err := svc.PatchBlocks(ctx, entry.ID, PatchBlocksRequest{Ops: []BlockOp{
			{Op: OpSetState, ID: "cccc3333", Item: "k1", Checked: &on},
		}}); err != nil {
			t.Fatalf("tick: %v", err)
		}
		exported, err := svc.Get(ctx, entry.ID)
		if err != nil {
			t.Fatalf("get: %v", err)
		}

		// Import re-creates the entry from the exported content.
		research, sections, _ := NewResearchService(storage.NewResearchRepository(dbOf(svc)),
			storage.NewSectionRepository(dbOf(svc)),
			storage.NewTeamRepository(dbOf(svc)),
			testAccess(dbOf(svc)),
			&mockNotifier{},
			slog.Default(),
		).Create(ctx, CreateResearchRequest{Name: "Imported", Goal: "T",
			Sections: []CreateSectionRequest{{Name: "s1", DisplayName: "S1"}}})
		imported, err := svc.Create(ctx, CreateEntryRequest{
			ResearchID: research.ID, SectionID: sections[0].ID,
			Title: "Runbook", Content: exported.Content, Type: domain.EntryBlocks,
		})
		if err != nil {
			t.Fatalf("import: %v", err)
		}
		ticks := 0
		for _, b := range blocksOf(t, svc, ctx, imported.ID) {
			for _, item := range checklistItems(b.Data) {
				if item.Checked {
					ticks++
				}
			}
		}
		if ticks != 1 {
			t.Errorf("ticks after a round trip = %d, want 1", ticks)
		}
	})
}

func dbOf(svc *EntryService) *bun.DB { return svc.blocks.DB() }
