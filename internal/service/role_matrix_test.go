package service

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"testing"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/storage"
)

// The ownership suite next door asks one question: can a stranger reach this?
// Teams add a second, and it is the one a single-owner model never had to
// answer — a colleague who *can* see the research still may not be allowed to
// change it.
//
// So this is a matrix, not a list: every entity, crossed with every role. A
// missed call site in the sweep from `validateResearchAccess` to
// `Access.Write` is a viewer who can write, and nothing else in the suite
// would notice.

type roleKit struct {
	db       *sql.DB
	research *ResearchService
	section  *SectionService
	entry    *EntryService
	session  *SessionService
	task     *TaskService
	roadmap  *RoadmapService
	team     *TeamService
	teamRepo *storage.TeamRepository
}

func newRoleKit(t *testing.T) *roleKit {
	t.Helper()
	db := setupTestDB(t)
	log := slog.Default()
	notifier := &mockNotifier{}
	access := testAccess(db)

	researchRepo := storage.NewResearchRepository(db)
	sectionRepo := storage.NewSectionRepository(db)
	entryRepo := storage.NewEntryRepository(db)
	sessionRepo := storage.NewSessionRepository(db)
	teamRepo := storage.NewTeamRepository(db)

	entrySvc := NewEntryService(entryRepo, sectionRepo, researchRepo, access, sessionRepo,
		storage.NewBlockRepository(db), storage.NewEntryRevisionRepository(db),
		storage.NewCrossRefRepository(db), storage.NewExternalLinkRepository(db), notifier, log)

	return &roleKit{
		db:       db,
		research: NewResearchService(researchRepo, sectionRepo, teamRepo, access, notifier, log),
		section:  NewSectionService(sectionRepo, entryRepo, researchRepo, access, notifier, log),
		entry:    entrySvc,
		session: NewSessionService(db, sessionRepo, storage.NewQuestionRepository(db), researchRepo,
			access, entrySvc, notifier, log),
		task: NewTaskService(storage.NewTaskRepository(db), researchRepo, access, entrySvc, notifier, log),
		roadmap: NewRoadmapService(storage.NewRoadmapRepository(db), storage.NewRoadmapNodeRepository(db),
			storage.NewRoadmapEdgeRepository(db), researchRepo, access, notifier, log),
		team:     NewTeamService(teamRepo, storage.NewTeamInviteRepository(db), storage.NewUserRepository(db), researchRepo, notifier, log),
		teamRepo: teamRepo,
	}
}

// sharedResearch sets up the shape every case here needs: an owner with a
// non-personal team, a research in it, and a second user holding `role`.
func (k *roleKit) sharedResearch(t *testing.T, role domain.TeamRole) (owner, member context.Context, research *domain.Research, section *domain.Section, teamID string) {
	t.Helper()
	ownerUser := createTestUser(t, k.db, "owner-"+string(role)+"@test.com", "Owner")
	memberUser := createTestUser(t, k.db, "member-"+string(role)+"@test.com", "Member")
	owner, member = userCtx(ownerUser), userCtx(memberUser)

	team, err := k.team.Create(owner, "Shared "+string(role))
	if err != nil {
		t.Fatalf("create team: %v", err)
	}
	teamID = team.ID
	addToTeam(t, k.db, teamID, memberUser.ID, role)

	research, sections, err := k.research.Create(owner, CreateResearchRequest{
		TeamID: teamID,
		Name:   "Shared research",
		Goal:   "Roles",
		Sections: []CreateSectionRequest{
			{Name: "s1", DisplayName: "Section one"},
		},
	})
	if err != nil {
		t.Fatalf("create research: %v", err)
	}
	return owner, member, research, sections[0], teamID
}

func TestRoles_EveryMemberCanRead(t *testing.T) {
	for _, role := range []domain.TeamRole{domain.TeamViewer, domain.TeamEditor, domain.TeamOwner} {
		t.Run(string(role), func(t *testing.T) {
			k := newRoleKit(t)
			owner, member, research, section, _ := k.sharedResearch(t, role)

			if _, err := k.entry.Create(owner, CreateEntryRequest{
				ResearchID: research.ID, SectionID: section.ID,
				Title: "Seed", Content: "body",
			}); err != nil {
				t.Fatalf("seed entry: %v", err)
			}

			reads := map[string]func() error{
				"research get": func() error { _, err := k.research.Get(member, research.ID); return err },
				"research list": func() error {
					list, err := k.research.List(member, storage.ResearchFilter{})
					if err != nil {
						return err
					}
					for _, r := range list {
						if r.ID == research.ID {
							return nil
						}
					}
					t.Fatal("shared research missing from the member's list")
					return nil
				},
				"sections":  func() error { _, err := k.section.List(member, research.ID); return err },
				"entries":   func() error { _, err := k.entry.ListByResearch(member, research.ID, storage.EntryFilter{}); return err },
				"tasks":     func() error { _, err := k.task.List(member, research.ID, storage.TaskFilter{}); return err },
				"sessions":  func() error { _, err := k.session.ListByResearch(member, research.ID); return err },
				"roadmaps":  func() error { _, err := k.roadmap.List(member, research.ID); return err },
				"team read": func() error { _, err := k.team.Members(member, researchTeam(t, k, research.ID)); return err },
			}
			for name, call := range reads {
				if err := call(); err != nil {
					t.Errorf("%s: a %s should be able to read: %v", name, role, err)
				}
			}
		})
	}
}

func TestRoles_ViewerCannotWriteAnything(t *testing.T) {
	k := newRoleKit(t)
	owner, viewer, research, section, _ := k.sharedResearch(t, domain.TeamViewer)

	entry, err := k.entry.Create(owner, CreateEntryRequest{
		ResearchID: research.ID, SectionID: section.ID, Title: "Seed", Content: "body",
	})
	if err != nil {
		t.Fatalf("seed entry: %v", err)
	}
	sess, _, err := k.session.Create(owner, CreateSessionRequest{
		ResearchID: research.ID, Title: "S", Focus: "F",
		Questions: []CreateQuestionRequest{{Text: "Q?"}},
	})
	if err != nil {
		t.Fatalf("seed session: %v", err)
	}
	task, err := k.task.Create(owner, CreateTaskRequest{ResearchID: research.ID, Title: "T"})
	if err != nil {
		t.Fatalf("seed task: %v", err)
	}
	rm, err := k.roadmap.Create(owner, CreateRoadmapRequest{ResearchID: research.ID, Title: "RM"})
	if err != nil {
		t.Fatalf("seed roadmap: %v", err)
	}
	questions, err := k.session.ListQuestions(owner, sess.ID, storage.QuestionFilter{})
	if err != nil || len(questions) == 0 {
		t.Fatalf("seed questions: %v", err)
	}

	writes := map[string]func(context.Context) error{
		"research update": func(ctx context.Context) error {
			_, err := k.research.Update(ctx, research.ID, UpdateResearchRequest{Name: ptr("Renamed")})
			return err
		},
		"add section": func(ctx context.Context) error {
			_, err := k.research.AddSection(ctx, research.ID, CreateSectionRequest{Name: "s2", DisplayName: "S2"})
			return err
		},
		"section update": func(ctx context.Context) error {
			_, err := k.section.Update(ctx, section.ID, UpdateSectionRequest{DisplayName: ptr("Renamed")})
			return err
		},
		"entry create": func(ctx context.Context) error {
			_, err := k.entry.Create(ctx, CreateEntryRequest{
				ResearchID: research.ID, SectionID: section.ID, Title: "New", Content: "x",
			})
			return err
		},
		"entry update": func(ctx context.Context) error {
			_, err := k.entry.Update(ctx, entry.ID, UpdateEntryRequest{Title: ptr("Renamed")})
			return err
		},
		"entry delete": func(ctx context.Context) error { return k.entry.Delete(ctx, entry.ID) },
		"entry patch": func(ctx context.Context) error {
			_, err := k.entry.PatchBlocks(ctx, entry.ID, PatchBlocksRequest{})
			return err
		},
		"crossref rebuild": func(ctx context.Context) error {
			_, err := k.entry.RebuildCrossRefs(ctx, research.ID)
			return err
		},
		"session create": func(ctx context.Context) error {
			_, _, err := k.session.Create(ctx, CreateSessionRequest{ResearchID: research.ID, Title: "S2"})
			return err
		},
		"session update": func(ctx context.Context) error {
			_, err := k.session.Update(ctx, sess.ID, UpdateSessionRequest{Title: ptr("Renamed")})
			return err
		},
		"question add": func(ctx context.Context) error {
			_, err := k.session.AddQuestions(ctx, sess.ID, []CreateQuestionRequest{{Text: "Another?"}})
			return err
		},
		"question update": func(ctx context.Context) error {
			answer := "an answer"
			_, err := k.session.UpdateQuestion(ctx, questions[0].ID, nil, &answer)
			return err
		},
		"task create": func(ctx context.Context) error {
			_, err := k.task.Create(ctx, CreateTaskRequest{ResearchID: research.ID, Title: "T2"})
			return err
		},
		"task update": func(ctx context.Context) error {
			_, err := k.task.Update(ctx, task.ID, UpdateTaskRequest{Title: ptr("Renamed")})
			return err
		},
		"task delete": func(ctx context.Context) error { return k.task.Delete(ctx, task.ID) },
		"roadmap create": func(ctx context.Context) error {
			_, err := k.roadmap.Create(ctx, CreateRoadmapRequest{ResearchID: research.ID, Title: "RM2"})
			return err
		},
		"roadmap update": func(ctx context.Context) error {
			_, err := k.roadmap.Update(ctx, rm.ID, UpdateRoadmapRequest{Title: ptr("Renamed")})
			return err
		},
		"roadmap add nodes": func(ctx context.Context) error {
			_, err := k.roadmap.AddNodes(ctx, rm.ID, []CreateRoadmapNodeRequest{{Title: "N"}}, nil)
			return err
		},
		"roadmap delete": func(ctx context.Context) error { return k.roadmap.Delete(ctx, rm.ID) },
	}

	for name, call := range writes {
		if err := call(viewer); !errors.Is(err, ErrForbidden) {
			t.Errorf("%s: a viewer must be refused with ErrForbidden, got %v", name, err)
		}
	}
}

func TestRoles_EditorCanWriteButNotAdminister(t *testing.T) {
	k := newRoleKit(t)
	_, editor, research, section, teamID := k.sharedResearch(t, domain.TeamEditor)

	if _, err := k.entry.Create(editor, CreateEntryRequest{
		ResearchID: research.ID, SectionID: section.ID, Title: "Editor's entry", Content: "body",
	}); err != nil {
		t.Fatalf("an editor must be able to write: %v", err)
	}
	if _, err := k.task.Create(editor, CreateTaskRequest{ResearchID: research.ID, Title: "T"}); err != nil {
		t.Fatalf("an editor must be able to create a task: %v", err)
	}
	if _, err := k.research.Update(editor, research.ID, UpdateResearchRequest{Goal: ptr("New goal")}); err != nil {
		t.Fatalf("an editor must be able to update the research: %v", err)
	}

	// Managing the team is where an editor stops.
	other := createTestUser(t, k.db, "outsider@test.com", "Outsider")
	if _, err := k.team.CreateInvite(editor, teamID, "someone@test.com", domain.TeamViewer); !errors.Is(err, ErrForbidden) {
		t.Errorf("an editor must not invite: %v", err)
	}
	if err := k.team.UpdateRole(editor, teamID, other.ID, domain.TeamEditor); !errors.Is(err, ErrForbidden) {
		t.Errorf("an editor must not change roles: %v", err)
	}
	if err := k.team.Delete(editor, teamID); !errors.Is(err, ErrForbidden) {
		t.Errorf("an editor must not delete the team: %v", err)
	}
	if err := k.team.TransferResearch(editor, research.ID, teamID); !errors.Is(err, ErrForbidden) {
		t.Errorf("an editor must not move the research: %v", err)
	}
}

func TestRoles_RemovedMemberLosesEverything(t *testing.T) {
	k := newRoleKit(t)
	owner, member, research, _, teamID := k.sharedResearch(t, domain.TeamEditor)

	memberUser := createTestUser(t, k.db, "second@test.com", "Second")
	addToTeam(t, k.db, teamID, memberUser.ID, domain.TeamOwner)

	if _, err := k.research.Get(member, research.ID); err != nil {
		t.Fatalf("member should start with access: %v", err)
	}

	var removed string
	members, err := k.team.Members(owner, teamID)
	if err != nil {
		t.Fatalf("members: %v", err)
	}
	for _, m := range members {
		if m.Email == "member-editor@test.com" {
			removed = m.UserID
		}
	}
	if removed == "" {
		t.Fatal("did not find the member to remove")
	}
	if err := k.team.RemoveMember(owner, teamID, removed); err != nil {
		t.Fatalf("remove member: %v", err)
	}

	// Not "forbidden" — gone. Someone outside the team must not learn the
	// research is there.
	if _, err := k.research.Get(member, research.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("a removed member must get ErrNotFound, got %v", err)
	}
	list, err := k.research.List(member, storage.ResearchFilter{})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	for _, r := range list {
		if r.ID == research.ID {
			t.Error("a removed member still sees the research in their list")
		}
	}
}

// researchTeam is the team a research belongs to, read back through the
// service so the test asserts on what callers actually get.
func researchTeam(t *testing.T, k *roleKit, researchID string) string {
	t.Helper()
	r, err := storage.NewResearchRepository(k.db).FindByID(context.Background(), researchID)
	if err != nil || r == nil {
		t.Fatalf("find research: %v", err)
	}
	return r.TeamID
}
