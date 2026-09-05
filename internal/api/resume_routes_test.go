package api

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/butschster/mcp-research/internal/domain"
)

// The resume route, driven through the real mux.
//
// Two of these are about routing rather than about the summary: the endpoint
// must exist under the owner's prefix and must NOT exist under a share token.
// That is a property of where the handler is registered, and only the mux can
// answer it.

type resumeEnvelope struct {
	Data domain.ResearchResume `json:"data"`
}

func TestResumeRoute_ServesTheOwnerAndRefusesAShare(t *testing.T) {
	s := newShareServer(t)
	token := s.newShare(allIn())

	code, body := s.get("/api/researches/" + s.research.ID + "/resume")
	if code != 200 {
		t.Fatalf("owner resume: %d %s", code, body)
	}
	var env resumeEnvelope
	if err := json.Unmarshal([]byte(body), &env); err != nil {
		t.Fatalf("decode: %v — %s", err, body)
	}
	if env.Data.SchemaVersion != domain.ResumeSchemaVersion {
		t.Errorf("schema_version = %d", env.Data.SchemaVersion)
	}
	if env.Data.Research.Code != s.research.Code {
		t.Errorf("research = %+v", env.Data.Research)
	}
	// The fixture's research holds one task and two entries.
	if env.Data.Work.Pending.Total == nil || *env.Data.Work.Pending.Total != 1 {
		t.Errorf("pending total = %v, want 1", env.Data.Work.Pending.Total)
	}
	if env.Data.RecentEntries.Returned == 0 {
		t.Error("recent entries should carry the fixture's documents")
	}

	// The summary is working process — what is unfinished, what somebody
	// disputed — and a share link publishes a result, not a workplan. The route
	// is not on the share sub-mux at all, so this is a 404 from the router
	// rather than a refusal from the service.
	//
	// The middle path is the one that guards anything: it is the shape the
	// sub-mux actually serves, so mounting `resume` there turns it green. The
	// other two rewrite to `/api/api/...` and `/api/resume` and would 404
	// whatever is registered — they are here to pin the prefix arithmetic, not
	// the route list.
	for _, path := range []string{
		"/api/shared/" + token + "/api/researches/" + s.research.ID + "/resume",
		"/api/shared/" + token + "/researches/" + s.research.ID + "/resume",
		"/api/shared/" + token + "/resume",
	} {
		if code, body := s.get(path); code != 404 {
			t.Errorf("share resume %s: %d %s — the summary must not be reachable through a link", path, code, body)
		}
	}
}

func TestResumeRoute_ShortCodeAndBadInput(t *testing.T) {
	s := newShareServer(t)

	// Every link in the product is built from the short code, so the code is
	// the identity this route is actually called with.
	code, body := s.get("/api/researches/" + s.research.Code + "/resume")
	if code != 200 || !strings.Contains(body, s.research.Code) {
		t.Fatalf("resume by code: %d %s", code, body)
	}

	// A malformed limit is the caller's bug and is refused, rather than being
	// silently defaulted to a number they did not ask for.
	if code, body := s.get("/api/researches/" + s.research.ID + "/resume?limit=abc"); code != 400 {
		t.Errorf("limit=abc: %d %s, want 400", code, body)
	}
	// Out of range is a different case: "as many as you can" is a reasonable
	// thing to mean, so the service clamps it.
	code, body = s.get("/api/researches/" + s.research.ID + "/resume?limit=9999")
	if code != 200 {
		t.Fatalf("limit=9999: %d %s", code, body)
	}
	var env resumeEnvelope
	if err := json.Unmarshal([]byte(body), &env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if env.Data.RecentEntries.Returned > domain.ResumeMaxLimit {
		t.Errorf("returned %d items, cap is %d", env.Data.RecentEntries.Returned, domain.ResumeMaxLimit)
	}

	// An unknown research and a session belonging to another one read alike:
	// confirming that either exists is information about somebody else's work.
	if code, _ := s.get("/api/researches/R999/resume"); code != 404 {
		t.Errorf("unknown research: %d, want 404", code)
	}
	if code, _ := s.get("/api/researches/" + s.other.ID + "/resume?session_id=" + s.sessionID); code != 404 {
		t.Errorf("session from another research: %d, want 404", code)
	}
}

// The one thing a summary must never do is quietly acknowledge documents.
// Reading it twice leaves the personal new/changed queue exactly as it was.
func TestResumeRoute_DoesNotMarkAnythingSeen(t *testing.T) {
	s := newShareServer(t)

	before := s.mustUpdatesCount()
	if code, body := s.get("/api/researches/" + s.research.ID + "/resume"); code != 200 {
		t.Fatalf("resume: %d %s", code, body)
	}
	if after := s.mustUpdatesCount(); after != before {
		t.Errorf("the personal queue changed from %d to %d — reading a summary is not reading the documents", before, after)
	}
}

func (s *shareServer) mustUpdatesCount() int {
	s.t.Helper()
	code, body := s.get("/api/researches/" + s.research.ID + "/updates")
	if code != 200 {
		s.t.Fatalf("updates: %d %s", code, body)
	}
	var env struct {
		Data struct {
			Count int `json:"count"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &env); err != nil {
		s.t.Fatalf("decode updates: %v", err)
	}
	return env.Data.Count
}
