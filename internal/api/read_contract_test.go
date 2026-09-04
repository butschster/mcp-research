package api

import (
	"encoding/json"
	"net/url"
	"strings"
	"testing"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/storage"
	"github.com/google/uuid"
)

func TestReadContract_SearchScopeAndEmptyResults(t *testing.T) {
	s := newShareServer(t)
	for _, tc := range []struct {
		query string
		ids   []string
	}{
		{"q=finding&research=" + s.research.ID, []string{s.entry.ID}},
		{"q=FINDING&research=" + s.research.ID, []string{s.entry.ID}},
		{"q=" + url.QueryEscape("' OR 1=1 --") + "&research=" + s.research.ID, nil},
		{"q=unmatched", nil}, {"q=x", nil}, {"q=", nil},
	} {
		code, body := s.get("/api/search?" + tc.query)
		if code != 200 {
			t.Fatalf("search status %d: %s", code, body)
		}
		var result struct {
			Entries []domain.Entry `json:"entries"`
		}
		if err := json.Unmarshal([]byte(body), &result); err != nil {
			t.Fatal(err)
		}
		if result.Entries == nil || len(result.Entries) != len(tc.ids) {
			t.Fatalf("search %s: %s", tc.query, body)
		}
		for i, e := range result.Entries {
			if e.ID != tc.ids[i] || e.Content != "" {
				t.Fatalf("search scope/projection: %+v", e)
			}
		}
	}
}

func TestReadContract_LinksGraphAndAnnotationQueue(t *testing.T) {
	s := newShareServer(t)
	ctx := t.Context()
	links := storage.NewExternalLinkRepository(s.db)
	link := domain.ExternalLink{ID: uuid.NewString(), SourceType: "entry", SourceID: s.entry.ID, ResearchID: s.research.ID, URL: "https://example.invalid/source", Title: "Evidence", Domain: "example.invalid"}
	if err := links.ReplaceForSource(ctx, "entry", s.entry.ID, []domain.ExternalLink{link}); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/api/researches/" + s.research.ID + "/links", "/api/entries/" + s.entry.ID + "/links"} {
		code, body := s.get(path)
		if code != 200 || !strings.Contains(body, link.URL) || !strings.Contains(body, s.entry.Code) {
			t.Fatalf("link projection %s: %d %s", path, code, body)
		}
	}
	if code, body := s.get("/api/researches/" + s.other.ID + "/links"); code != 200 || strings.Contains(body, link.URL) {
		t.Fatalf("link scope: %d %s", code, body)
	}
	for _, path := range []string{"/api/researches/missing/links", "/api/entries/missing/links", "/api/researches/missing/graph"} {
		if code, body := s.get(path); code != 404 {
			t.Fatalf("missing %s: %d %s", path, code, body)
		}
	}
	code, body := s.get("/api/researches/" + s.research.ID + "/graph")
	if code != 200 {
		t.Fatalf("graph: %d %s", code, body)
	}
	var graph struct {
		Nodes []struct{ ID, Type string }
		Edges []struct{ Source, Target, Type string }
	}
	if err := json.Unmarshal([]byte(body), &graph); err != nil {
		t.Fatal(err)
	}
	ids := map[string]bool{}
	for _, node := range graph.Nodes {
		if ids[node.ID] {
			t.Fatalf("duplicate graph node: %s", node.ID)
		}
		ids[node.ID] = true
	}
	if !ids[s.entry.ID] || !ids[s.sectionID] || !ids[s.sessionID] || strings.Contains(body, "Private finding") {
		t.Fatalf("graph scope: %s", body)
	}
	sectionEdge := false
	for _, edge := range graph.Edges {
		if edge.Type == "section" && edge.Source == s.sectionID && edge.Target == s.entry.ID {
			sectionEdge = true
		}
	}
	if !sectionEdge {
		t.Fatal("graph omitted the entry's section edge")
	}
	ann := &domain.Annotation{ID: uuid.NewString(), EntryID: s.entry.ID, ResearchID: s.research.ID, Body: "Check this", Kind: domain.AnnotationVerify, Status: domain.AnnotationOpen}
	if err := storage.NewAnnotationRepository(s.db).Create(ctx, ann); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/api/researches/" + s.research.ID + "/annotations?status=open&kind=verify", "/api/entries/" + s.entry.ID + "/annotations"} {
		code, body := s.get(path)
		if code != 200 || !strings.Contains(body, ann.ID) {
			t.Fatalf("annotation list %s: %d %s", path, code, body)
		}
	}
	if code, body := s.get("/api/researches/" + s.other.ID + "/annotations"); code != 200 || strings.Contains(body, ann.ID) {
		t.Fatalf("annotation scope: %d %s", code, body)
	}
	for _, query := range []string{"status=invalid", "kind=invalid"} {
		if code, body := s.get("/api/researches/" + s.research.ID + "/annotations?" + query); code != 400 {
			t.Fatalf("invalid annotation filter: %d %s", code, body)
		}
	}
}
