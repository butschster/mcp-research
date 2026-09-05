package service

import (
	"strings"
	"testing"

	"github.com/dovod-app/app/internal/domain"
)

// A task_ref holds references and nothing else. Everything below is about the
// two ways that can go wrong: a reference that should not have been kept, and a
// tick this block is not allowed to own.

func TestNormTaskRef_KeepsOnlyValidReferences(t *testing.T) {
	doc := normDoc(t, `[{"type":"task_ref","data":{"tasks":[
		"T1", "t4", "  T7 ",
		"9F2C1A44-BE07-4E44-9BC7-1D2E3F405162",
		"E3", "T", "TX", "", "42", 7, null
	]}}]`)

	if len(doc.Blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(doc.Blocks))
	}
	got := taskRefCodes(doc.Blocks[0].Data)
	want := []string{"T1", "T4", "T7", "9f2c1a44-be07-4e44-9bc7-1d2e3f405162"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("refs = %v, want %v", got, want)
	}
}

func TestNormTaskRef_CollapsesDuplicates(t *testing.T) {
	// The same task twice is one row. Two checkboxes for one task would show a
	// half-done list the moment either was ticked.
	doc := normDoc(t, `[{"type":"task_ref","data":{"tasks":["T1","t1","T1 ","T2"]}}]`)
	if got := taskRefCodes(doc.Blocks[0].Data); len(got) != 2 {
		t.Errorf("refs = %v, want two", got)
	}
}

func TestNormTaskRef_DroppedWhenNothingResolvable(t *testing.T) {
	_, err := NormalizeBlockDocument(`[{"type":"task_ref","data":{"tasks":["E3","nope"]}}]`)
	if err == nil {
		t.Fatal("a block with no usable reference should be dropped, leaving no document")
	}
}

func TestNormTaskRef_CapsTheList(t *testing.T) {
	var refs []string
	for i := 1; i <= domain.MaxTaskRefs+40; i++ {
		refs = append(refs, `"T`+itoa(i)+`"`)
	}
	doc := normDoc(t, `[{"type":"task_ref","data":{"tasks":[`+strings.Join(refs, ",")+`]}}]`)
	if got := len(taskRefCodes(doc.Blocks[0].Data)); got != domain.MaxTaskRefs {
		t.Errorf("kept %d refs, want the cap of %d", got, domain.MaxTaskRefs)
	}
}

func TestNormTaskRef_ShowProgressIsStamped(t *testing.T) {
	// Absent means yes, so the renderer never has to tell a missing key from a
	// false one.
	doc := normDoc(t, `[{"type":"task_ref","data":{"tasks":["T1"]}}]`)
	if v, ok := doc.Blocks[0].Data["show_progress"].(bool); !ok || !v {
		t.Errorf("show_progress = %v, want true", doc.Blocks[0].Data["show_progress"])
	}
	off := normDoc(t, `[{"type":"task_ref","data":{"tasks":["T1"],"show_progress":false}}]`)
	if v, _ := off.Blocks[0].Data["show_progress"].(bool); v {
		t.Error("show_progress:false was not honoured")
	}
}

func TestNormTaskRef_RejectsAuthorState(t *testing.T) {
	// The block has no state of its own — the task holds it. An author asserting
	// one would be asserting a task status through a document.
	doc := normDoc(t, `[{"type":"task_ref","data":{"tasks":["T1"],"state":{"T1":true},"checked":true}}]`)
	d := doc.Blocks[0].Data
	if _, ok := d["state"]; ok {
		t.Error("state survived normalization")
	}
	if _, ok := d["checked"]; ok {
		t.Error("checked survived normalization")
	}
}

func TestTaskRefText_IndexesTheNoteAndTheCodes(t *testing.T) {
	doc := normDoc(t, `[{"type":"task_ref","data":{"tasks":["T4"],"note":"Before the call — see [[E3]]"}}]`)
	text := BlockPlainText(doc)
	if !strings.Contains(text, "[[E3]]") {
		t.Errorf("the note's reference is not indexed:\n%s", text)
	}
	if !strings.Contains(text, "T4") {
		t.Errorf("the task code is not indexed:\n%s", text)
	}
}

func TestTaskRefMarkdown_WithoutAResolverClaimsNothing(t *testing.T) {
	// An exporter that cannot see the tasks must not draw boxes. Every `- [ ]`
	// would read as "nobody has started", which it never asked.
	doc := normDoc(t, `[{"type":"task_ref","data":{"tasks":["T4","T7"],"note":"Before the call"}}]`)
	md := BlockDocumentToMarkdown(doc)
	if strings.Contains(md, "- [ ]") || strings.Contains(md, "- [x]") {
		t.Errorf("an unresolved block drew checkboxes:\n%s", md)
	}
	for _, want := range []string{"Before the call", "[[T4]]", "[[T7]]"} {
		if !strings.Contains(md, want) {
			t.Errorf("missing %q:\n%s", want, md)
		}
	}
}

func TestTaskRefMarkdown_ResolvedIsATaskList(t *testing.T) {
	doc := normDoc(t, `[{"type":"task_ref","data":{"tasks":["T1","T2","T3"]}}]`)
	md := BlockDocumentToMarkdownWith(doc, MarkdownOptions{Tasks: NewTaskRefResolver([]*domain.Task{
		{ID: "id1", Code: "T1", Title: "Migrate", Status: domain.TaskCompleted},
		{ID: "id2", Code: "T2", Title: "Measure", Status: domain.TaskBlocked},
		{ID: "id3", Code: "T3", Title: "Ship", Status: domain.TaskPending},
	})})

	if !strings.Contains(md, "- [x] T1 — Migrate") {
		t.Errorf("a completed task is not ticked:\n%s", md)
	}
	if !strings.Contains(md, "- [ ] T3 — Ship") {
		t.Errorf("a pending task is not listed:\n%s", md)
	}
	// blocked is neither ticked nor silent: `- [ ]` alone reads as untouched.
	if !strings.Contains(md, "- [ ] T2 — Measure *(blocked)*") {
		t.Errorf("a blocked task lost its status:\n%s", md)
	}
	if !strings.Contains(md, "*1 of 3 done*") {
		t.Errorf("no progress line:\n%s", md)
	}
}

func TestTaskRefMarkdown_DropsReferencesThatResolveToNothing(t *testing.T) {
	doc := normDoc(t, `[{"type":"task_ref","data":{"tasks":["T1","T9"]}}]`)
	md := BlockDocumentToMarkdownWith(doc, MarkdownOptions{Tasks: NewTaskRefResolver([]*domain.Task{
		{ID: "id1", Code: "T1", Title: "Migrate", Status: domain.TaskPending},
	})})
	if strings.Contains(md, "T9") {
		t.Errorf("a deleted task left a ghost row:\n%s", md)
	}
	if !strings.Contains(md, "*0 of 1 done*") {
		t.Errorf("the progress line counted the ghost:\n%s", md)
	}
}

func TestTaskRefMarkdown_EveryReferenceGoneSaysSo(t *testing.T) {
	doc := normDoc(t, `[{"type":"task_ref","data":{"tasks":["T9"]}}]`)
	md := BlockDocumentToMarkdownWith(doc, MarkdownOptions{Tasks: NewTaskRefResolver(nil)})
	if !strings.Contains(md, "every reference in this list has been removed") {
		t.Errorf("an emptied block said nothing:\n%s", md)
	}
}

func TestTaskRefResolver_MatchesByUUIDToo(t *testing.T) {
	const id = "9f2c1a44-be07-4e44-9bc7-1d2e3f405162"
	rows := NewTaskRefResolver([]*domain.Task{
		{ID: id, Code: "T4", Title: "Migrate", Status: domain.TaskCompleted},
	})([]string{id})
	if len(rows) != 1 || rows[0].Code != "T4" || !rows[0].Done {
		t.Errorf("uuid reference did not resolve: %+v", rows)
	}
}

// itoa without importing strconv into a test file that reads better without it.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

func TestTaskRefDescription_FallsBackToTheNote(t *testing.T) {
	doc := normDoc(t, `[{"type":"task_ref","data":{"tasks":["T1"],"note":"Close before Friday"}}]`)
	if got := BlockDocumentDescription(doc, ""); got != "Close before Friday" {
		t.Errorf("description = %q, want the note", got)
	}
}

func TestTaskRef_FilesItsCodesAsReferences(t *testing.T) {
	// The block that IS a reference has to file one, or the task's own page can
	// never find the document that plans around it.
	doc := normDoc(t, `[{"type":"task_ref","data":{"tasks":["T4","9f2c1a44-be07-4e44-9bc7-1d2e3f405162"]}}]`)
	text := BlockPlainText(doc)
	if !strings.Contains(text, "[[T4]]") {
		t.Errorf("a task code was not filed as a reference:\n%s", text)
	}
	// A uuid is not a syntax this product resolves; filing it would add a row
	// that can only ever read "unresolved".
	if strings.Contains(text, "[[9f2c1a44") {
		t.Errorf("a uuid was filed as a reference:\n%s", text)
	}
}

func TestTaskRef_ReportsWhatTheCapThrewAway(t *testing.T) {
	var refs []string
	for i := 1; i <= domain.MaxTaskRefs+7; i++ {
		refs = append(refs, `"T`+itoa(i)+`"`)
	}
	doc := normDoc(t, `[{"type":"task_ref","data":{"tasks":[`+strings.Join(refs, ",")+`]}}]`)
	if got := intOr(doc.Blocks[0].Data, "truncated", 0); got != 7 {
		t.Errorf("truncated = %d, want 7", got)
	}
	if md := BlockDocumentToMarkdown(doc); !strings.Contains(md, "7 further references were not stored") {
		t.Errorf("the export does not say the list is incomplete:\n%s", md[:200])
	}
}

func TestTruncation_SurvivesASecondNormalization(t *testing.T) {
	// A create normalizes the author's payload and then normalizes the document
	// that produced. The second pass drops nothing — it is looking at 50 stored
	// references — so without carrying the count forward the warning vanished on
	// the very write that earned it.
	var refs []string
	for i := 1; i <= domain.MaxTaskRefs+7; i++ {
		refs = append(refs, `"T`+itoa(i)+`"`)
	}
	once := normDoc(t, `[{"type":"task_ref","data":{"tasks":[`+strings.Join(refs, ",")+`]}}]`)
	stored, err := MarshalBlockDocument(once)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	twice := normDoc(t, stored)
	if got := intOr(twice.Blocks[0].Data, "truncated", 0); got != 7 {
		t.Errorf("after a second pass truncated = %d, want 7", got)
	}
	// And a third: normalizing a normal document must be a no-op.
	third, _ := MarshalBlockDocument(twice)
	if got := intOr(normDoc(t, third).Blocks[0].Data, "truncated", 0); got != 7 {
		t.Errorf("normalization is not idempotent: truncated = %d, want 7", got)
	}
}
