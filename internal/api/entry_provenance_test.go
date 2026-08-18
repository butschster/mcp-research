package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/butschster/mcp-research/internal/domain"
)

// Provenance rides at the top level of the entry response, beside `data` rather
// than inside it — which is easy to get wrong from the outside and was, by me,
// while adding a field to it. A test pins the shape as well as the contents.
func TestEntryPayload_ProvenanceRidesBesideData(t *testing.T) {
	s := newShareServer(t)

	code, body := s.get("/api/researches/" + s.research.ID + "/entries/" + s.entry.ID)
	if code != http.StatusOK {
		t.Fatalf("status %d: %s", code, body)
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(body), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if out["revision"] == nil || out["author_kind"] == nil || out["revised_at"] == nil {
		t.Fatalf("provenance is missing from the top level: %v", body)
	}
	if data, ok := out["data"].(map[string]any); ok {
		if data["revision"] != nil {
			t.Error("provenance is also inside `data`; one place or the other, not both")
		}
	}
}

// A share visitor gets none of it — including the name, which is the field this
// branch added and the most identifying thing in the block.
func TestEntryPayload_ProvenanceIsWithheldFromAShare(t *testing.T) {
	s := newShareServer(t)
	token := s.newShare(domain.ShareInclude{Sessions: true, Tasks: true, Roadmaps: true, Export: true})

	code, body := s.get("/api/shared/" + token + "/researches/" + s.research.ID + "/entries/" + s.entry.ID)
	if code != http.StatusOK {
		t.Fatalf("status %d: %s", code, body)
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(body), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, key := range []string{"revision", "author_kind", "author_name", "revised_at", "revision_session"} {
		if out[key] != nil {
			t.Errorf("a share link was given %q", key)
		}
	}
}
