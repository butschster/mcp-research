package storage

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/google/uuid"
)

func TestRepositoryCoverage_AnnotationQueueIsolation(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	entry, other := contractEntry(t, db, nil), contractEntry(t, db, nil)
	repo := NewAnnotationRepository(db)
	annotations := []*domain.Annotation{
		{ID: uuid.NewString(), ResearchID: entry.ResearchID, EntryID: entry.ID, Kind: domain.AnnotationVerify, Status: domain.AnnotationOpen, Body: "first"},
		{ID: uuid.NewString(), ResearchID: entry.ResearchID, EntryID: entry.ID, Kind: domain.AnnotationDig, Status: domain.AnnotationOpen, Body: "second"},
		{ID: uuid.NewString(), ResearchID: entry.ResearchID, EntryID: entry.ID, Kind: domain.AnnotationVerify, Status: domain.AnnotationClosed, Body: "closed"},
		{ID: uuid.NewString(), ResearchID: other.ResearchID, EntryID: other.ID, Kind: domain.AnnotationVerify, Status: domain.AnnotationOpen, Body: "other research"},
	}
	for i, a := range annotations {
		if err := repo.Create(ctx, a); err != nil {
			t.Fatal(err)
		}
		if _, err := db.NewUpdate().Table("annotations").Set("created_at=?", time.Date(2026, 1, 1, 0, 0, i, 0, time.UTC).Format(time.DateTime)).Where("id=?", a.ID).Exec(ctx); err != nil {
			t.Fatal(err)
		}
	}
	for _, a := range []*domain.Annotation{annotations[0], annotations[3]} {
		got, err := repo.FindByCode(ctx, a.ResearchID, "A1")
		if err != nil || got == nil || got.ID != a.ID {
			t.Fatalf("scoped code: %+v %v", got, err)
		}
	}
	if got, err := repo.FindByCode(ctx, other.ResearchID, "A2"); err != nil || got != nil {
		t.Fatalf("code escaped research: %+v %v", got, err)
	}
	counts, err := repo.CountByStatus(ctx, entry.ResearchID)
	if err != nil || !reflect.DeepEqual(counts, map[domain.AnnotationStatus]int{domain.AnnotationOpen: 2, domain.AnnotationClosed: 1}) {
		t.Fatalf("counts: %+v %v", counts, err)
	}
	open, verify := domain.AnnotationOpen, domain.AnnotationVerify
	for _, tc := range []struct {
		filter AnnotationFilter
		ids    []string
	}{
		{AnnotationFilter{Status: &open, Kind: &verify, EntryID: entry.ID}, []string{annotations[0].ID}},
		{AnnotationFilter{Limit: 1, Offset: 1}, []string{annotations[1].ID}},
		{AnnotationFilter{EntryID: other.ID}, nil},
		{AnnotationFilter{Limit: 1, Offset: 20}, nil},
	} {
		got, err := repo.FindByResearch(ctx, entry.ResearchID, tc.filter)
		if err != nil {
			t.Fatal(err)
		}
		var ids []string
		for _, a := range got {
			ids = append(ids, a.ID)
			if a.EntryTitle != entry.Title || a.EntryCode != entry.Code {
				t.Fatalf("entry join: %+v", a)
			}
		}
		if !reflect.DeepEqual(ids, tc.ids) {
			t.Fatalf("filter %+v: got %v want %v", tc.filter, ids, tc.ids)
		}
	}
	if err := repo.Delete(ctx, annotations[0].ID); err != nil {
		t.Fatal(err)
	}
	if got, err := repo.FindByID(ctx, annotations[0].ID); err != nil || got != nil {
		t.Fatalf("deleted annotation: %+v %v", got, err)
	}
	if got, err := repo.FindByID(ctx, annotations[3].ID); err != nil || got == nil {
		t.Fatalf("other annotation removed: %+v %v", got, err)
	}
	if err := NewEntryRepository(db).Delete(ctx, entry.ID); err != nil {
		t.Fatal(err)
	}
	if got, err := repo.CountByStatus(ctx, entry.ResearchID); err != nil || len(got) != 0 {
		t.Fatalf("cascade counts: %+v %v", got, err)
	}
}

func TestRepositoryCoverage_ScopedCodesAndEdgeDeletion(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	entries := []*domain.Entry{contractEntry(t, db, nil), contractEntry(t, db, nil)}
	roadmaps, nodes, edges, sessions := NewRoadmapRepository(db), NewRoadmapNodeRepository(db), NewRoadmapEdgeRepository(db), NewSessionRepository(db)
	var roadmapIDs []string
	for _, e := range entries {
		rm := &domain.Roadmap{ID: uuid.NewString(), ResearchID: e.ResearchID, Title: "map"}
		if err := roadmaps.Create(ctx, rm); err != nil {
			t.Fatal(err)
		}
		roadmapIDs = append(roadmapIDs, rm.ID)
		n1 := &domain.RoadmapNode{ID: uuid.NewString(), RoadmapID: rm.ID, Title: "first"}
		n2 := &domain.RoadmapNode{ID: uuid.NewString(), RoadmapID: rm.ID, Title: "second"}
		for _, n := range []*domain.RoadmapNode{n1, n2} {
			if err := nodes.Create(ctx, n); err != nil {
				t.Fatal(err)
			}
		}
		edge := &domain.RoadmapEdge{ID: uuid.NewString(), RoadmapID: rm.ID, SourceNodeID: n1.ID, TargetNodeID: n2.ID}
		if err := edges.Create(ctx, edge); err != nil {
			t.Fatal(err)
		}
		if got, err := roadmaps.FindByCodeAndResearch(ctx, rm.Code, e.ResearchID); err != nil || got == nil || got.ID != rm.ID {
			t.Fatalf("roadmap code: %+v %v", got, err)
		}
		if got, err := nodes.FindByCode(ctx, rm.ID, "N1"); err != nil || got == nil || got.ID != n1.ID {
			t.Fatalf("node code: %+v %v", got, err)
		}
		s := &domain.Session{ID: uuid.NewString(), ResearchID: e.ResearchID, Title: "session"}
		if err := sessions.Create(ctx, s); err != nil {
			t.Fatal(err)
		}
		if got, err := sessions.FindByCodeAndResearch(ctx, s.Code, e.ResearchID); err != nil || got == nil || got.ID != s.ID {
			t.Fatalf("session code: %+v %v", got, err)
		}
	}
	for _, scope := range []string{"missing", entries[0].ResearchID} {
		if got, err := roadmaps.FindByCodeAndResearch(ctx, "RM999", scope); err != nil || got != nil {
			t.Fatalf("missing roadmap: %+v %v", got, err)
		}
		if got, err := sessions.FindByCodeAndResearch(ctx, "SS999", scope); err != nil || got != nil {
			t.Fatalf("missing session: %+v %v", got, err)
		}
	}
	if got, err := nodes.FindByCode(ctx, "missing", "N1"); err != nil || got != nil {
		t.Fatalf("node escaped roadmap: %+v %v", got, err)
	}
	unique := &domain.Roadmap{ID: uuid.NewString(), ResearchID: entries[0].ResearchID, Code: "RM900", Title: "unique"}
	if err := roadmaps.Create(ctx, unique); err != nil {
		t.Fatal(err)
	}
	if got, err := roadmaps.FindByCode(ctx, unique.Code); err != nil || got == nil || got.ID != unique.ID {
		t.Fatalf("legacy code lookup: %+v %v", got, err)
	}
	if got, err := roadmaps.FindByCodeAndResearch(ctx, unique.Code, entries[1].ResearchID); err != nil || got != nil {
		t.Fatalf("cross-research code leak: %+v %v", got, err)
	}
	if err := edges.DeleteByRoadmap(ctx, roadmapIDs[0]); err != nil {
		t.Fatal(err)
	}
	for i, id := range roadmapIDs {
		got, err := edges.FindByRoadmap(ctx, id)
		if err != nil || len(got) != i {
			t.Fatalf("edge deletion scope %d: %+v %v", i, got, err)
		}
	}
	revisions := NewEntryRevisionRepository(db)
	if n, err := revisions.CountByEntry(ctx, entries[0].ID); err != nil || n != 0 {
		t.Fatalf("initial revision count: %d %v", n, err)
	}
	for i := 1; i <= 2; i++ {
		if err := revisions.Create(ctx, nil, &domain.EntryRevision{EntryID: entries[0].ID, ResearchID: entries[0].ResearchID, Revision: i}); err != nil {
			t.Fatal(err)
		}
	}
	if n, err := revisions.CountByEntry(ctx, entries[0].ID); err != nil || n != 2 {
		t.Fatalf("revision count: %d %v", n, err)
	}
	if n, err := revisions.CountByEntry(ctx, entries[1].ID); err != nil || n != 0 {
		t.Fatalf("cross-entry count: %d %v", n, err)
	}
}

func TestRepositoryCoverage_TeamLibrariesAndUsage(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	u := contractUser(t, db)
	teams, skills, templates := NewTeamRepository(db), NewSkillRepository(db), NewTemplateRepository(db)
	var owned *domain.Skill
	var teamIDs []string
	for _, name := range []string{"first", "other"} {
		team := &domain.Team{ID: uuid.NewString(), Name: name, CreatedBy: u.ID}
		if err := teams.CreateWithOwner(ctx, team, u.ID); err != nil {
			t.Fatal(err)
		}
		teamIDs = append(teamIDs, team.ID)
		for _, title := range []string{"Zulu", "Alpha"} {
			sk := &domain.Skill{ID: uuid.NewString(), TeamID: team.ID, Slug: title, Name: title, Tier: domain.SkillTeam, Body: "日本語abc", Version: 1}
			if err := skills.Create(ctx, sk); err != nil {
				t.Fatal(err)
			}
			if owned == nil {
				owned = sk
			}
			tp := &domain.Template{ID: uuid.NewString(), TeamID: team.ID, Slug: title, Name: title, Tier: domain.TemplateTeam, Body: "one two three", Version: 1}
			if err := templates.Create(ctx, tp); err != nil {
				t.Fatal(err)
			}
		}
		got, err := skills.ListByTeam(ctx, team.ID)
		if err != nil || len(got) != 2 || got[0].Name != "Alpha" || got[1].Name != "Zulu" {
			t.Fatalf("skills sorted: %+v %v", got, err)
		}
		for _, sk := range got {
			if sk.TeamID != team.ID || sk.Body != "" || sk.BodyTokens != 2 {
				t.Fatalf("skill scope/projection: %+v", sk)
			}
		}
		tpls, err := templates.ListByTeam(ctx, team.ID)
		if err != nil || len(tpls) != 2 || tpls[0].Name != "Alpha" || tpls[1].Name != "Zulu" {
			t.Fatalf("templates sorted: %+v %v", tpls, err)
		}
		for _, tp := range tpls {
			if tp.TeamID != team.ID || tp.Body != "" || tp.BodyWords != 3 {
				t.Fatalf("template scope/projection: %+v", tp)
			}
		}
	}
	if err := teams.Rename(ctx, teamIDs[0], "new ' ? name"); err != nil {
		t.Fatal(err)
	}
	for i, want := range []string{"new ' ? name", "other"} {
		got, err := teams.FindByID(ctx, teamIDs[i])
		if err != nil || got == nil || got.Name != want {
			t.Fatalf("rename scope: %+v %v", got, err)
		}
	}
	if got, err := skills.ListByTeam(ctx, "missing"); err != nil || len(got) != 0 {
		t.Fatalf("missing skill library: %+v %v", got, err)
	}
	if got, err := templates.ListByTeam(ctx, "missing"); err != nil || len(got) != 0 {
		t.Fatalf("missing template library: %+v %v", got, err)
	}
	if n, err := skills.UsageCount(ctx, owned.ID); err != nil || n != 0 {
		t.Fatalf("unused: %d %v", n, err)
	}
	var researchIDs []string
	for i := 0; i < 2; i++ {
		e := contractEntry(t, db, nil)
		researchIDs = append(researchIDs, e.ResearchID)
		if err := skills.Attach(ctx, e.ResearchID, owned.ID, false); err != nil {
			t.Fatal(err)
		}
	}
	if err := skills.Attach(ctx, researchIDs[0], owned.ID, false); err != nil {
		t.Fatal(err)
	}
	if n, err := skills.UsageCount(ctx, owned.ID); err != nil || n != 2 {
		t.Fatalf("unique usage: %d %v", n, err)
	}
	if err := skills.Detach(ctx, researchIDs[0], owned.ID); err != nil {
		t.Fatal(err)
	}
	if n, err := skills.UsageCount(ctx, owned.ID); err != nil || n != 1 {
		t.Fatalf("detached usage: %d %v", n, err)
	}
}

func TestRepositoryCoverage_InviteSingleConsumption(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	owner, recipient := contractUser(t, db), contractUser(t, db)
	team := &domain.Team{ID: uuid.NewString(), Name: "invites", CreatedBy: owner.ID}
	if err := NewTeamRepository(db).CreateWithOwner(ctx, team, owner.ID); err != nil {
		t.Fatal(err)
	}
	repo := NewTeamInviteRepository(db)
	makeInvite := func(hash string, expires time.Time) *domain.TeamInvite {
		i := &domain.TeamInvite{ID: uuid.NewString(), TeamID: team.ID, Email: recipient.Email, Role: domain.TeamEditor, InvitedBy: owner.ID, ExpiresAt: expires}
		if err := repo.Create(ctx, i, hash); err != nil {
			t.Fatal(err)
		}
		return i
	}
	inv := makeInvite("pending", time.Now().Add(time.Hour))
	expired := makeInvite("expired", time.Now().Add(-time.Hour))
	revoked := makeInvite("revoked", time.Now().Add(time.Hour))
	if err := repo.Revoke(ctx, revoked.ID); err != nil {
		t.Fatal(err)
	}
	if err := repo.MarkAccepted(ctx, revoked.ID, recipient.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("revoked consumed: %v", err)
	}
	if err := repo.MarkAccepted(ctx, "missing", recipient.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("missing consumed: %v", err)
	}
	const contenders = 8
	results := make(chan error, contenders)
	var wg sync.WaitGroup
	for i := 0; i < contenders; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); results <- repo.MarkAccepted(ctx, inv.ID, recipient.ID) }()
	}
	wg.Wait()
	close(results)
	winners := 0
	for err := range results {
		if err == nil {
			winners++
		} else if !errors.Is(err, sql.ErrNoRows) {
			t.Error(err)
		}
	}
	if winners != 1 {
		t.Fatalf("consumed %d times, want 1", winners)
	}
	got, err := repo.FindByID(ctx, inv.ID)
	if err != nil || got == nil || got.AcceptedAt == nil || got.AcceptedBy != recipient.ID {
		t.Fatalf("acceptance: %+v %v", got, err)
	}
	open, err := repo.ListOpenByTeam(ctx, team.ID)
	if err != nil || len(open) != 1 || open[0].ID != expired.ID {
		t.Fatalf("open queue must retain expired but exclude consumed/revoked: %+v %v", open, err)
	}
}
