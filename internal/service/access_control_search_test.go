package service

import (
	"errors"
	"log/slog"
	"testing"

	"github.com/butschster/mcp-research/internal/auth"
	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/storage"
)

// The three leaks these cover were all reachable by an authenticated user against
// another user's research, and all three bypassed validateResearchAccess by going
// to a repository directly. They live in one file because they are one class.

func TestAccessControl_Search(t *testing.T) {
	db := setupTestDB(t)
	log := slog.Default()

	researchRepo := storage.NewResearchRepository(db)
	sectionRepo := storage.NewSectionRepository(db)
	entryRepo := storage.NewEntryRepository(db)
	blockRepo := storage.NewBlockRepository(db)
	crossrefRepo := storage.NewCrossRefRepository(db)
	researchSvc := NewResearchService(researchRepo, sectionRepo, storage.NewTeamRepository(db), testAccess(db), &mockNotifier{}, log)
	entrySvc := NewEntryService(entryRepo, sectionRepo, researchRepo, testAccess(db), nil, blockRepo, storage.NewEntryRevisionRepository(db), crossrefRepo, nil, &mockNotifier{}, log)

	userA, userB := setupTwoUsers(t, db)
	ctxA := userCtx(userA)

	research, sections, _ := researchSvc.Create(ctxA, CreateResearchRequest{
		Name: "Alice's Research", Goal: "Test",
		Sections: []CreateSectionRequest{{Name: "s1", DisplayName: "S1"}},
	})
	if _, err := entrySvc.Create(ctxA, CreateEntryRequest{
		ResearchID: research.ID,
		SectionID:  sections[0].ID,
		Title:      "Zaphod deployment notes",
		Content:    "The passphrase is hunter2 and the bastion is at 10.0.0.1",
	}); err != nil {
		t.Fatalf("create entry: %v", err)
	}

	t.Run("owner finds their own entry", func(t *testing.T) {
		found, err := entryRepo.SearchEntries(ctxA, "Zaphod", 20, userA.ID)
		if err != nil {
			t.Fatalf("search: %v", err)
		}
		if len(found) != 1 {
			t.Fatalf("owner got %d results, want 1", len(found))
		}
	})

	t.Run("another user finds nothing", func(t *testing.T) {
		for _, q := range []string{"Zaphod", "hunter2", "bastion"} {
			found, err := entryRepo.SearchEntries(userCtx(userB), q, 20, userB.ID)
			if err != nil {
				t.Fatalf("search %q: %v", q, err)
			}
			if len(found) != 0 {
				// content LIKE makes the search box an oracle: a hit on a term
				// that appears only in the body confirms the body.
				t.Errorf("query %q leaked %d of another user's entries", q, len(found))
			}
		}
	})

	t.Run("no user means no scoping, as everywhere else", func(t *testing.T) {
		found, err := entryRepo.SearchEntries(ctxA, "Zaphod", 20, "")
		if err != nil {
			t.Fatalf("search: %v", err)
		}
		if len(found) != 1 {
			t.Fatalf("auth-disabled search got %d results, want 1", len(found))
		}
	})
}

func TestAccessControl_RoadmapRefData(t *testing.T) {
	db := setupTestDB(t)
	log := slog.Default()

	researchRepo := storage.NewResearchRepository(db)
	sectionRepo := storage.NewSectionRepository(db)
	entryRepo := storage.NewEntryRepository(db)
	blockRepo := storage.NewBlockRepository(db)
	crossrefRepo := storage.NewCrossRefRepository(db)
	roadmapRepo := storage.NewRoadmapRepository(db)
	nodeRepo := storage.NewRoadmapNodeRepository(db)
	edgeRepo := storage.NewRoadmapEdgeRepository(db)

	researchSvc := NewResearchService(researchRepo, sectionRepo, storage.NewTeamRepository(db), testAccess(db), &mockNotifier{}, log)
	entrySvc := NewEntryService(entryRepo, sectionRepo, researchRepo, testAccess(db), nil, blockRepo, storage.NewEntryRevisionRepository(db), crossrefRepo, nil, &mockNotifier{}, log)
	roadmapSvc := NewRoadmapService(roadmapRepo, nodeRepo, edgeRepo, researchRepo, testAccess(db), &mockNotifier{}, log)
	roadmapSvc.SetRefResolvers(entryRepo, nil, nil, nil, sectionRepo)

	userA, userB := setupTwoUsers(t, db)
	ctxA, ctxB := userCtx(userA), userCtx(userB)

	// A writes something private.
	researchA, sectionsA, _ := researchSvc.Create(ctxA, CreateResearchRequest{
		Name: "Alice's Research", Goal: "Test",
		Sections: []CreateSectionRequest{{Name: "s1", DisplayName: "S1"}},
	})
	secret, err := entrySvc.Create(ctxA, CreateEntryRequest{
		ResearchID: researchA.ID,
		SectionID:  sectionsA[0].ID,
		Title:      "Alice's private findings",
		Content:    "The passphrase is hunter2.",
	})
	if err != nil {
		t.Fatalf("create entry: %v", err)
	}

	// B builds a roadmap in his own research and points a node at A's entry.
	researchB, _, _ := researchSvc.Create(ctxB, CreateResearchRequest{Name: "Bob's Research", Goal: "Test"})
	roadmap, err := roadmapSvc.Create(ctxB, CreateRoadmapRequest{
		ResearchID: researchB.ID,
		Title:      "Bob's map",
		Nodes: []CreateRoadmapNodeRequest{{
			Title:   "peek",
			RefType: string(domain.RefTypeEntry),
			RefID:   secret.ID,
		}},
	})
	if err != nil {
		t.Fatalf("create roadmap: %v", err)
	}

	got, err := roadmapSvc.Get(ctxB, roadmap.ID)
	if err != nil {
		t.Fatalf("get roadmap: %v", err)
	}
	if len(got.Nodes) != 1 {
		t.Fatalf("nodes = %d, want 1", len(got.Nodes))
	}
	// The node is Bob's and must survive; only the resolved payload is withheld.
	if ref := got.Nodes[0].RefData; ref != nil {
		t.Errorf("ref data resolved across users: title=%q content=%q", ref.Title, ref.Content)
	}

	t.Run("the owner still gets their own ref data", func(t *testing.T) {
		ownRoadmap, err := roadmapSvc.Create(ctxA, CreateRoadmapRequest{
			ResearchID: researchA.ID,
			Title:      "Alice's map",
			Nodes: []CreateRoadmapNodeRequest{{
				Title:   "mine",
				RefType: string(domain.RefTypeEntry),
				RefID:   secret.ID,
			}},
		})
		if err != nil {
			t.Fatalf("create roadmap: %v", err)
		}
		got, err := roadmapSvc.Get(ctxA, ownRoadmap.ID)
		if err != nil {
			t.Fatalf("get roadmap: %v", err)
		}
		if got.Nodes[0].RefData == nil {
			t.Fatal("ref data withheld from the owner — the check is too strict")
		}
	})
}

func TestTextReplaceRefusedOnBlocks(t *testing.T) {
	db := setupTestDB(t)
	log := slog.Default()

	researchRepo := storage.NewResearchRepository(db)
	sectionRepo := storage.NewSectionRepository(db)
	entryRepo := storage.NewEntryRepository(db)
	blockRepo := storage.NewBlockRepository(db)
	crossrefRepo := storage.NewCrossRefRepository(db)
	researchSvc := NewResearchService(researchRepo, sectionRepo, storage.NewTeamRepository(db), testAccess(db), &mockNotifier{}, log)
	entrySvc := NewEntryService(entryRepo, sectionRepo, researchRepo, testAccess(db), nil, blockRepo, storage.NewEntryRevisionRepository(db), crossrefRepo, nil, &mockNotifier{}, log)

	userA, _ := setupTwoUsers(t, db)
	ctx := auth.WithUser(userCtx(userA), userA)

	research, sections, _ := researchSvc.Create(ctx, CreateResearchRequest{
		Name: "R", Goal: "T",
		Sections: []CreateSectionRequest{{Name: "s1", DisplayName: "S1"}},
	})

	doc := `{"blocks":[{"type":"heading","data":{"level":2,"text":"Heading"}}]}`
	entry, err := entrySvc.Create(ctx, CreateEntryRequest{
		ResearchID: research.ID,
		SectionID:  sections[0].ID,
		Title:      "Block document",
		Content:    doc,
		Type:       domain.EntryBlocks,
	})
	if err != nil {
		t.Fatalf("create blocks entry: %v", err)
	}

	// A quote in the replacement used to be enough to leave unparseable JSON in
	// the column, with a success returned to the caller.
	_, err = entrySvc.Update(ctx, entry.ID, UpdateEntryRequest{
		TextReplace: &TextReplace{From: "Heading", To: `He"ading`},
	})
	if !errors.Is(err, ErrTextReplaceOnBlocks) {
		t.Fatalf("expected ErrTextReplaceOnBlocks, got %v", err)
	}

	after, err := entrySvc.Get(ctx, entry.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if _, err := NormalizeBlockDocument(after.Content); err != nil {
		t.Fatalf("stored document no longer parses: %v", err)
	}

	t.Run("markdown entries keep text_replace", func(t *testing.T) {
		md, err := entrySvc.Create(ctx, CreateEntryRequest{
			ResearchID: research.ID,
			SectionID:  sections[0].ID,
			Title:      "Markdown",
			Content:    "hello world",
		})
		if err != nil {
			t.Fatalf("create markdown entry: %v", err)
		}
		updated, err := entrySvc.Update(ctx, md.ID, UpdateEntryRequest{
			TextReplace: &TextReplace{From: "world", To: "there"},
		})
		if err != nil {
			t.Fatalf("text_replace on markdown: %v", err)
		}
		if updated.Content != "hello there" {
			t.Errorf("content = %q, want %q", updated.Content, "hello there")
		}
	})
}

// Three more leaks of the same shape, found by auditing the block work: each
// returned another user's data to an authenticated caller who never touched
// their research.
func TestAccessControl_RelatedAndRefs(t *testing.T) {
	db := setupTestDB(t)
	log := slog.Default()

	researchRepo := storage.NewResearchRepository(db)
	sectionRepo := storage.NewSectionRepository(db)
	entryRepo := storage.NewEntryRepository(db)
	blockRepo := storage.NewBlockRepository(db)
	crossrefRepo := storage.NewCrossRefRepository(db)
	researchSvc := NewResearchService(researchRepo, sectionRepo, storage.NewTeamRepository(db), testAccess(db), &mockNotifier{}, log)
	entrySvc := NewEntryService(entryRepo, sectionRepo, researchRepo, testAccess(db), nil, blockRepo, storage.NewEntryRevisionRepository(db), crossrefRepo, nil, &mockNotifier{}, log)

	alice, mallory := setupTwoUsers(t, db)
	ctxA, ctxM := userCtx(alice), userCtx(mallory)

	researchA, sectionsA, _ := researchSvc.Create(ctxA, CreateResearchRequest{
		Name: "Alice's Research", Goal: "T",
		Sections: []CreateSectionRequest{{Name: "s1", DisplayName: "S1"}},
	})
	secret, err := entrySvc.Create(ctxA, CreateEntryRequest{
		ResearchID: researchA.ID, SectionID: sectionsA[0].ID,
		Title: "Alice's incident notes", Content: "The passphrase is hunter2.",
		Tags: []string{"security"},
	})
	if err != nil {
		t.Fatalf("create entry: %v", err)
	}

	researchM, sectionsM, _ := researchSvc.Create(ctxM, CreateResearchRequest{
		Name: "Mallory's Research", Goal: "T",
		Sections: []CreateSectionRequest{{Name: "s1", DisplayName: "S1"}},
	})

	t.Run("tags do not relate entries across users", func(t *testing.T) {
		mine, err := entrySvc.Create(ctxM, CreateEntryRequest{
			ResearchID: researchM.ID, SectionID: sectionsM[0].ID,
			Title: "My notes", Content: "Nothing here.", Tags: []string{"security"},
		})
		if err != nil {
			t.Fatalf("create entry: %v", err)
		}
		related, err := entryRepo.FindRelatedByTags(ctxM, mine.ID, []string{"security"}, mallory.ID)
		if err != nil {
			t.Fatalf("related: %v", err)
		}
		for _, e := range related {
			if e.ResearchID == researchA.ID {
				t.Errorf("related entries leaked %q from another user", e.Title)
			}
		}
	})

	t.Run("a cross-research reference does not cross a user", func(t *testing.T) {
		// Mallory writes [[R1:E1]] in her own entry. It must stay unresolved:
		// resolving it would store Alice's entry uuid in Mallory's crossrefs,
		// which is how a foreign id becomes harvestable.
		body := "See [[" + researchA.Code + ":" + secret.Code + "]]."
		mine, err := entrySvc.Create(ctxM, CreateEntryRequest{
			ResearchID: researchM.ID, SectionID: sectionsM[0].ID,
			Title: "Probe", Content: body,
		})
		if err != nil {
			t.Fatalf("create entry: %v", err)
		}
		refs, err := crossrefRepo.FindBySourceEntry(ctxM, mine.ID)
		if err != nil {
			t.Fatalf("crossrefs: %v", err)
		}
		for _, r := range refs {
			if r.TargetEntryID == secret.ID || r.TargetResearchID == researchA.ID {
				t.Errorf("crossref resolved across users: target_entry=%q target_research=%q", r.TargetEntryID, r.TargetResearchID)
			}
		}
	})

	t.Run("a uuid from another research is not readable through your own", func(t *testing.T) {
		if _, err := entrySvc.GetByIDOrCode(ctxM, researchM.ID, secret.ID); !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}
