package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/butschster/mcp-research/internal/auth"
	"github.com/butschster/mcp-research/internal/config"
	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/service"
	"github.com/butschster/mcp-research/internal/storage"
	"github.com/google/uuid"
)

// This handler resolved a short code straight from the repository, so any
// authenticated caller who knew a research id could read any entry of it —
// title, status and full content. The test is here rather than in the service
// package on purpose: EntryService.GetByIDOrCode was always scoped correctly,
// and what went wrong was a handler reaching past it.
func TestResolveCode_DoesNotCrossUsers(t *testing.T) {
	log := slog.Default()
	db, err := storage.NewDB(config.Config{}, log)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	userRepo := storage.NewUserRepository(db)
	researchRepo := storage.NewResearchRepository(db)
	sectionRepo := storage.NewSectionRepository(db)
	entryRepo := storage.NewEntryRepository(db)
	blockRepo := storage.NewBlockRepository(db)
	crossrefRepo := storage.NewCrossRefRepository(db)

	researchSvc := service.NewResearchService(researchRepo, sectionRepo, nopNotifier{}, log)
	entrySvc := service.NewEntryService(entryRepo, sectionRepo, researchRepo, nil, blockRepo, crossrefRepo, nil, nopNotifier{}, log)
	handler := NewEntryHandler(entrySvc, entryRepo, researchRepo, log)

	alice := &domain.User{ID: uuid.New().String(), Email: "alice@test.com", PasswordHash: "x", Name: "Alice"}
	mallory := &domain.User{ID: uuid.New().String(), Email: "mallory@test.com", PasswordHash: "x", Name: "Mallory"}
	for _, u := range []*domain.User{alice, mallory} {
		if err := userRepo.Create(t.Context(), u); err != nil {
			t.Fatalf("create user: %v", err)
		}
	}

	ctxAlice := auth.WithUser(t.Context(), alice)
	research, sections, err := researchSvc.Create(ctxAlice, service.CreateResearchRequest{
		Name: "Alice's Research", Goal: "Test",
		Sections: []service.CreateSectionRequest{{Name: "s1", DisplayName: "S1"}},
	})
	if err != nil {
		t.Fatalf("create research: %v", err)
	}
	secret, err := entrySvc.Create(ctxAlice, service.CreateEntryRequest{
		ResearchID: research.ID,
		SectionID:  sections[0].ID,
		Title:      "Private findings",
		Content:    "The passphrase is hunter2.",
	})
	if err != nil {
		t.Fatalf("create entry: %v", err)
	}

	call := func(user *domain.User) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/api/researches/"+research.ID+"/entries/by-code/"+secret.Code, nil)
		req.SetPathValue("id", research.ID)
		req.SetPathValue("code", secret.Code)
		if user != nil {
			req = req.WithContext(auth.WithUser(req.Context(), user))
		}
		rec := httptest.NewRecorder()
		handler.ResolveCode(rec, req)
		return rec
	}

	t.Run("the owner reads their entry", func(t *testing.T) {
		rec := call(alice)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
		}
		var got struct {
			Data domain.Entry `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Data.ID != secret.ID {
			t.Errorf("id = %q, want %q", got.Data.ID, secret.ID)
		}
	})

	t.Run("another user gets nothing", func(t *testing.T) {
		rec := call(mallory)
		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404", rec.Code)
		}
		// Body matters as much as the status: a 500 carrying the row would leak
		// just as well as a 200.
		for _, leak := range []string{"hunter2", "Private findings", secret.ID} {
			if strings.Contains(rec.Body.String(), leak) {
				t.Errorf("response leaked %q: %s", leak, rec.Body.String())
			}
		}
	})
}

type nopNotifier struct{}

func (nopNotifier) Notify(service.Event) {}
