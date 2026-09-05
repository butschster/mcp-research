package api

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
)

// A generated document can still be wrong about the part a person writes. The
// first version of this one said `POST /api/sessions/{id}/questions` answers
// 201 and it answers 200; that a template fork answers 201 and it answers 200;
// that `GET /api/auth/me` returns `{user, teams}` and it returns a bare user.
// None of that is caught by generating the paths.
//
// So this walks the product the way a client would and, for every call, insists
// the status it actually got is one the document lists for that operation.

type contractClient struct {
	t     *testing.T
	s     *specServer
	doc   map[string]any
	token string
	// seen records which operations were actually exercised. Several calls
	// below are guarded on an id the step before returned, and a guard that
	// silently skips is a test that passes by not looking.
	seen map[string]bool
}

// call performs a request and checks the status against the document.
//
// `template` is the OpenAPI path — `/api/researches/{id}` — and `path` is the
// concrete URL. Keeping both is the point: the first is how the operation is
// named in the document, the second is what the mux is asked.
func (c *contractClient) call(method, template, path string, body any) map[string]any {
	c.t.Helper()

	var payload string
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			c.t.Fatal(err)
		}
		payload = string(raw)
	}
	r := httptest.NewRequest(method, path, strings.NewReader(payload))
	r.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		r.Header.Set("Authorization", "Bearer "+c.token)
	}
	w := httptest.NewRecorder()
	c.s.mux.ServeHTTP(w, r)
	if c.seen == nil {
		c.seen = map[string]bool{}
	}
	c.seen[strings.ToLower(method)+" "+template] = true

	item, ok := c.doc["paths"].(map[string]any)[template].(map[string]any)
	if !ok {
		c.t.Fatalf("%s %s is not in the document at all", method, template)
	}
	opDoc, ok := item[strings.ToLower(method)].(map[string]any)
	if !ok {
		c.t.Fatalf("the document has no %s for %s", method, template)
	}
	responses, _ := opDoc["responses"].(map[string]any)
	got := fmt.Sprint(w.Code)
	if _, listed := responses[got]; !listed {
		var documented []string
		for status := range responses {
			documented = append(documented, status)
		}
		c.t.Errorf("%s %s answered %s; the document lists %v\n  body: %s",
			method, template, got, documented, truncate(w.Body.String(), 200))
	}

	var out map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	c.checkShape(method, template, got, opDoc, w.Body.Bytes())
	return out
}

// checkShape compares the top-level keys of what came back with the top-level
// properties the document declares for that status.
//
// Checking the status alone was not enough: the first generated document said
// the three exports were wrapped in `data` and none of them are, that
// `sessions/{id}/changes` returns two arrays of documents when it returns two
// integers, and that the share-create token sits beside `data` when it sits
// inside it. Every one of those passed the status check and would have broken
// the first client to read the document.
func (c *contractClient) checkShape(method, template, status string, opDoc map[string]any, body []byte) {
	c.t.Helper()

	response, _ := opDoc["responses"].(map[string]any)[status].(map[string]any)
	if response == nil {
		return
	}
	content, _ := response["content"].(map[string]any)
	media, _ := content["application/json"].(map[string]any)
	schema, _ := media["schema"].(map[string]any)
	if schema == nil {
		return // a download, or a response with no declared body
	}
	schema = c.resolve(schema)
	if t, _ := schema["type"].(string); t != "object" {
		return // an array or a scalar; the key comparison does not apply
	}
	props, _ := schema["properties"].(map[string]any)
	if len(props) == 0 {
		return // deliberately unspecified
	}

	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		return // not a JSON object; nothing to compare
	}

	var undeclared []string
	for key := range got {
		if _, ok := props[key]; !ok {
			undeclared = append(undeclared, key)
		}
	}
	sort.Strings(undeclared)
	if len(undeclared) > 0 {
		c.t.Errorf("%s %s answered %s with keys the document does not declare: %v",
			method, template, status, undeclared)
	}

	// The other direction, for the keys the document promises unconditionally.
	var promised []string
	for _, raw := range asSlice(schema["required"]) {
		key, _ := raw.(string)
		if key == "" {
			continue
		}
		if _, ok := got[key]; !ok {
			promised = append(promised, key)
		}
	}
	sort.Strings(promised)
	if len(promised) > 0 {
		c.t.Errorf("%s %s answered %s without keys the document marks required: %v",
			method, template, status, promised)
	}
}

// resolve follows a single $ref into components/schemas.
func (c *contractClient) resolve(schema map[string]any) map[string]any {
	ref, _ := schema["$ref"].(string)
	if ref == "" {
		return schema
	}
	name := strings.TrimPrefix(ref, "#/components/schemas/")
	components, _ := c.doc["components"].(map[string]any)
	schemas, _ := components["schemas"].(map[string]any)
	if target, ok := schemas[name].(map[string]any); ok {
		return target
	}
	return schema
}

func asSlice(v any) []any {
	out, _ := v.([]any)
	return out
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func str(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k].(string); ok && v != "" {
			return v
		}
	}
	return ""
}

func TestOpenAPI_DocumentedStatusesMatchReality(t *testing.T) {
	s := newSpecServer(t)
	c := &contractClient{t: t, s: s, doc: specOf(t, nil, s.mux)}

	// --- signing in ---
	reg := c.call("POST", "/api/auth/register", "/api/auth/register", map[string]any{
		"email": "contract@test.com", "password": "hunter2hunter2", "name": "Person",
	})
	c.token, _ = reg["token"].(string)
	if c.token == "" {
		t.Fatalf("registration returned no token: %v", reg)
	}
	c.call("GET", "/api/auth/me", "/api/auth/me", nil)
	c.call("GET", "/api/auth/api-keys", "/api/auth/api-keys", nil)
	key := c.call("POST", "/api/auth/api-keys", "/api/auth/api-keys", map[string]any{"name": "a script"})
	if id := str(key, "id"); id != "" {
		c.call("DELETE", "/api/auth/api-keys/{id}", "/api/auth/api-keys/"+id, nil)
	}
	c.call("GET", "/api/auth/info", "/api/auth/info", nil)
	c.call("GET", "/api/health", "/api/health", nil)

	// --- a research with a section ---
	created := c.call("POST", "/api/researches", "/api/researches", map[string]any{
		"name": "Contract", "goal": "Check the document against the server",
		"sections": []map[string]any{{"name": "findings", "display_name": "Findings"}},
	})
	researchID := str(annData(created), "research_id", "id")
	if researchID == "" {
		t.Fatalf("no research id: %v", created)
	}
	research := "/api/researches/" + researchID

	c.call("GET", "/api/researches", "/api/researches", nil)
	got := c.call("GET", "/api/researches/{id}", research, nil)
	sections, _ := annData(got)["sections"].([]any)
	if len(sections) == 0 {
		t.Fatalf("the research came back with no sections: %v", got)
	}
	sectionID := str(sections[0].(map[string]any), "id")

	c.call("PUT", "/api/researches/{id}", research, map[string]any{"goal": "A newer goal"})
	c.call("POST", "/api/researches/{id}/sections", research+"/sections",
		map[string]any{"name": "open-questions", "display_name": "Open questions"})
	c.call("PUT", "/api/sections/{sectionId}", "/api/sections/"+sectionID,
		map[string]any{"description": "What we found"})

	// --- a document ---
	entry := c.call("POST", "/api/entries", "/api/entries", map[string]any{
		"research_id": researchID, "section_id": sectionID,
		"title": "A finding", "content": "the body", "tags": []string{"contract"},
	})
	entryID := str(annData(entry), "entry_id", "id")
	if entryID == "" {
		t.Fatalf("no entry id: %v", entry)
	}
	e := "/api/entries/" + entryID

	c.call("GET", "/api/entries/{id}", e, nil)
	c.call("PUT", "/api/entries/{id}", e, map[string]any{"content": "a longer body"})
	c.call("GET", "/api/entries/{id}/revisions", e+"/revisions", nil)
	c.call("GET", "/api/entries/{id}/revisions/{revision}", e+"/revisions/1", nil)
	c.call("GET", "/api/entries/{id}/related", e+"/related", nil)
	c.call("GET", "/api/entries/{id}/crossrefs", e+"/crossrefs", nil)
	c.call("GET", "/api/entries/{id}/links", e+"/links", nil)
	c.call("GET", "/api/entries/{id}/annotations", e+"/annotations", nil)
	c.call("PUT", "/api/entries/{id}/seen", e+"/seen", nil)
	c.call("GET", "/api/researches/{id}/entries", research+"/entries", nil)
	c.call("GET", "/api/researches/{id}/sections/{sectionId}/entries",
		research+"/sections/"+sectionID+"/entries", nil)
	c.call("GET", "/api/researches/{id}/tags", research+"/tags", nil)
	c.call("GET", "/api/researches/{id}/updates", research+"/updates", nil)
	c.call("POST", "/api/researches/{id}/updates/seen", research+"/updates/seen", nil)
	c.call("GET", "/api/researches/{id}/crossrefs", research+"/crossrefs", nil)
	c.call("POST", "/api/researches/{id}/crossrefs/rebuild", research+"/crossrefs/rebuild", nil)
	c.call("GET", "/api/researches/{id}/links", research+"/links", nil)
	c.call("GET", "/api/researches/{id}/graph", research+"/graph", nil)
	c.call("GET", "/api/researches/{id}/resume", research+"/resume", nil)
	c.call("GET", "/api/search", "/api/search?q=finding", nil)
	// The document says this takes a short code. It used to answer 200 with
	// nothing, which reads as "the research is empty".
	byCode := c.call("GET", "/api/search", "/api/search?q=finding&research="+str(annData(created), "code"), nil)
	byID := c.call("GET", "/api/search", "/api/search?q=finding&research="+researchID, nil)
	if len(asSlice(byCode["entries"])) != len(asSlice(byID["entries"])) ||
		len(asSlice(byID["entries"])) == 0 {
		t.Errorf("search scoped by short code found %d and by id %d; they must agree and not be empty",
			len(asSlice(byCode["entries"])), len(asSlice(byID["entries"])))
	}
	c.call("GET", "/api/metadata/schema", "/api/metadata/schema", nil)

	// --- an annotation ---
	ann := c.call("POST", "/api/entries/{id}/annotations", e+"/annotations", map[string]any{
		"kind": "question", "body": "Is this still true?", "quote": "a longer body",
	})
	if id := str(annData(ann), "id"); id != "" {
		c.call("PUT", "/api/annotations/{id}", "/api/annotations/"+id, map[string]any{"status": "closed"})
		c.call("DELETE", "/api/annotations/{id}", "/api/annotations/"+id, nil)
	}
	c.call("GET", "/api/researches/{id}/annotations", research+"/annotations", nil)

	// --- a task ---
	task := c.call("POST", "/api/tasks", "/api/tasks", map[string]any{
		"research_id": researchID, "title": "Chase it", "priority": "high",
	})
	if id := str(annData(task), "id"); id != "" {
		c.call("PUT", "/api/tasks/{id}", "/api/tasks/"+id, map[string]any{"status": "completed"})
		c.call("DELETE", "/api/tasks/{id}", "/api/tasks/"+id, nil)
	}
	c.call("GET", "/api/researches/{id}/tasks", research+"/tasks", nil)

	// --- a session and its questions ---
	sess := c.call("POST", "/api/sessions", "/api/sessions", map[string]any{
		"research_id": researchID, "title": "Initial exploration", "focus": "pricing",
		"questions": []map[string]any{{"text": "What does it cost?", "area": "pricing"}},
	})
	sessionID := str(annData(sess), "id")
	if sessionID == "" {
		t.Fatalf("no session id: %v", sess)
	}
	c.call("GET", "/api/researches/{id}/sessions", research+"/sessions", nil)
	c.call("GET", "/api/researches/{id}/sessions/{sessionId}", research+"/sessions/"+sessionID, nil)
	c.call("PUT", "/api/sessions/{id}", "/api/sessions/"+sessionID, map[string]any{"focus": "packaging"})
	added := c.call("POST", "/api/sessions/{id}/questions", "/api/sessions/"+sessionID+"/questions",
		map[string]any{"questions": []map[string]any{{"text": "And support?", "area": "pricing"}}})
	if list, _ := added["data"].([]any); len(list) > 0 {
		if qid := str(list[0].(map[string]any), "id"); qid != "" {
			c.call("PUT", "/api/questions/{questionId}", "/api/questions/"+qid,
				map[string]any{"answer": "It is included.", "status": "answered"})
		}
	}
	c.call("GET", "/api/sessions/{id}/changes", "/api/sessions/"+sessionID+"/changes", nil)

	// --- a roadmap ---
	rm := c.call("POST", "/api/roadmaps", "/api/roadmaps", map[string]any{
		"research_id": researchID, "title": "Plan",
		"nodes": []map[string]any{{"temp_id": "n1", "title": "Root"}},
	})
	if id := str(annData(rm), "id"); id != "" {
		c.call("GET", "/api/roadmaps/{id}", "/api/roadmaps/"+id, nil)
		c.call("PUT", "/api/roadmaps/{id}", "/api/roadmaps/"+id, map[string]any{"title": "A better plan"})
		c.call("DELETE", "/api/roadmaps/{id}", "/api/roadmaps/"+id, nil)
	}
	c.call("GET", "/api/researches/{id}/roadmaps", research+"/roadmaps", nil)

	// --- teams ---
	c.call("GET", "/api/teams", "/api/teams", nil)
	team := c.call("POST", "/api/teams", "/api/teams", map[string]any{"name": "A second team"})
	teamID := str(annData(team), "id")
	if teamID == "" {
		t.Fatalf("no team id: %v", team)
	}
	c.call("GET", "/api/teams/{id}", "/api/teams/"+teamID, nil)
	c.call("PUT", "/api/teams/{id}", "/api/teams/"+teamID, map[string]any{"name": "Renamed"})
	c.call("GET", "/api/teams/{id}/members", "/api/teams/"+teamID+"/members", nil)
	c.call("GET", "/api/teams/{id}/invites", "/api/teams/"+teamID+"/invites", nil)
	invite := c.call("POST", "/api/teams/{id}/invites", "/api/teams/"+teamID+"/invites",
		map[string]any{"email": "colleague@test.com", "role": "editor"})
	if id := str(annData(invite), "id"); id != "" {
		c.call("DELETE", "/api/invites/{id}", "/api/invites/"+id, nil)
	}
	c.call("GET", "/api/teams/{id}/skills", "/api/teams/"+teamID+"/skills", nil)
	c.call("GET", "/api/teams/{id}/templates", "/api/teams/"+teamID+"/templates", nil)

	// --- skills ---
	c.call("GET", "/api/researches/{id}/skills", research+"/skills", nil)
	c.call("GET", "/api/researches/{id}/skills/library", research+"/skills/library", nil)
	skill := c.call("POST", "/api/researches/{id}/skills", research+"/skills", map[string]any{
		"slug": "how-we-work", "name": "How we work", "content": "# Method\n\nAsk first.",
	})
	if id := str(annData(skill), "id"); id != "" {
		c.call("GET", "/api/skills/{skillId}", "/api/skills/"+id, nil)
		c.call("DELETE", "/api/skills/{skillId}", "/api/skills/"+id, nil)
	}

	// --- templates and share links ---
	c.call("GET", "/api/templates", "/api/templates", nil)
	c.call("GET", "/api/researches/{id}/templates/draft", research+"/templates/draft", nil)
	share := c.call("POST", "/api/researches/{id}/shares", research+"/shares",
		map[string]any{"include": map[string]any{"tasks": true}})
	c.call("GET", "/api/researches/{id}/shares", research+"/shares", nil)
	if token := str(share, "token"); token != "" {
		c.call("GET", "/api/shared/{token}", "/api/shared/"+token, nil)
	}
	if id := str(annData(share), "id"); id != "" {
		c.call("DELETE", "/api/shares/{id}", "/api/shares/"+id, nil)
	}

	// --- exports ---
	c.call("GET", "/api/researches/{id}/export", research+"/export", nil)
	c.call("GET", "/api/researches/{id}/export/portable", research+"/export/portable", nil)
	c.call("GET", "/api/researches/{id}/sessions/{sessionId}/export",
		research+"/sessions/"+sessionID+"/export", nil)
	c.call("GET", "/api/entries/{id}/markdown", e+"/markdown", nil)

	// --- admin ---
	c.call("POST", "/api/admin/backfill-codes", "/api/admin/backfill-codes", nil)

	// --- the refusals, which are the half a generated document gets wrong ---
	//
	// These five creates name their parent in the body rather than the path, so
	// they have no path parameter and were documented without a 404 — while
	// answering 404 for every unknown or other-team parent, which is the most
	// common way a create call fails.
	const nobody = "00000000-0000-0000-0000-000000000000"
	c.call("POST", "/api/entries", "/api/entries", map[string]any{
		"research_id": nobody, "section_id": nobody, "title": "x", "content": "x",
	})
	c.call("POST", "/api/tasks", "/api/tasks", map[string]any{
		"research_id": nobody, "title": "x",
	})
	c.call("POST", "/api/sessions", "/api/sessions", map[string]any{
		"research_id": nobody, "title": "x",
	})
	c.call("POST", "/api/roadmaps", "/api/roadmaps", map[string]any{
		"research_id": nobody, "title": "x",
	})
	// And a record that is simply not there, addressed through the path.
	c.call("GET", "/api/researches/{id}", "/api/researches/"+nobody, nil)
	c.call("GET", "/api/entries/{id}", "/api/entries/"+nobody, nil)
	// A body that is not JSON at all.
	c.call("POST", "/api/tasks", "/api/tasks", "not an object")

	// --- deleting the document last, so everything above had one ---
	c.call("DELETE", "/api/entries/{id}", e, nil)

	// A guard above that silently skipped its call would leave this test green
	// while checking nothing. The floor is what the walk reaches when every id
	// it needs came back.
	const floor = 60
	if len(c.seen) < floor {
		t.Fatalf("only %d operations were exercised, expected at least %d — a step returned no id and its calls were skipped",
			len(c.seen), floor)
	}
	t.Logf("checked %d of %d documented operations against the running server",
		len(c.seen), countOperations(c.doc))
}

func countOperations(doc map[string]any) int {
	n := 0
	paths, _ := doc["paths"].(map[string]any)
	for _, item := range paths {
		methods, _ := item.(map[string]any)
		for _, raw := range methods {
			if _, ok := raw.(map[string]any); ok {
				n++
			}
		}
	}
	return n
}

// annData reaches into the `data` envelope most routes use.
func annData(m map[string]any) map[string]any {
	if d, ok := m["data"].(map[string]any); ok {
		return d
	}
	return m
}
