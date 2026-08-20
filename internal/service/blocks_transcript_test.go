package service

import (
	"strings"
	"testing"

	"github.com/butschster/mcp-research/internal/domain"
)

func TestNormTranscript_KeepsTurnsWithText(t *testing.T) {
	doc := normDoc(t, `[{"type":"transcript","data":{
		"title":"Infrastructure call",
		"turns":[
			{"speaker":"Пётр","text":"Контур закрыт.","ts":"00:03:12"},
			{"speaker":"Аня","text":"Сканер внутрь — см. [[E4]]."},
			{"speaker":"Тишина","text":"   "},
			{"text":"An unattributed line stays."},
			"not a turn",
			{"speaker":"Ghost"}
		]}}]`)

	if len(doc.Blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(doc.Blocks))
	}
	turns := transcriptTurns(doc.Blocks[0].Data)
	if len(turns) != 3 {
		t.Fatalf("kept %d turns, want 3: %+v", len(turns), turns)
	}
	if turns[0].Speaker != "Пётр" || turns[0].Stamp != "00:03:12" {
		t.Errorf("first turn = %+v", turns[0])
	}
	if turns[2].Speaker != "" {
		t.Errorf("an unattributed turn invented a speaker: %+v", turns[2])
	}
	if doc.Blocks[0].Data["title"] != "Infrastructure call" {
		t.Errorf("title = %v", doc.Blocks[0].Data["title"])
	}
}

func TestNormTranscript_DroppedWhenNoTurnSpoke(t *testing.T) {
	_, err := NormalizeBlockDocument(`[{"type":"transcript","data":{"title":"Empty","turns":[{"speaker":"A"},{"text":"  "}]}}]`)
	if err == nil {
		t.Fatal("a transcript with no turns should be dropped, leaving no document")
	}
}

func TestNormTranscript_CapsTurns(t *testing.T) {
	var turns []string
	for i := 0; i < domain.MaxTranscriptTurns+50; i++ {
		turns = append(turns, `{"speaker":"A","text":"line"}`)
	}
	doc := normDoc(t, `[{"type":"transcript","data":{"turns":[`+strings.Join(turns, ",")+`]}}]`)
	if got := len(transcriptTurns(doc.Blocks[0].Data)); got != domain.MaxTranscriptTurns {
		t.Errorf("kept %d turns, want the cap of %d", got, domain.MaxTranscriptTurns)
	}
}

func TestNormTranscript_ClampsTheLabels(t *testing.T) {
	long := strings.Repeat("я", domain.MaxTranscriptSpeaker*2)
	doc := normDoc(t, `[{"type":"transcript","data":{"turns":[{"speaker":"`+long+`","text":"x","ts":"`+long+`"}]}}]`)
	turn := transcriptTurns(doc.Blocks[0].Data)[0]
	if len(turn.Speaker) > domain.MaxTranscriptSpeaker || len(turn.Stamp) > domain.MaxTranscriptStamp {
		t.Errorf("labels not clamped: speaker %d bytes, ts %d bytes", len(turn.Speaker), len(turn.Stamp))
	}
	// A clamp on a rune boundary, or the row stores invalid UTF-8.
	if !strings.HasPrefix(turn.Speaker, "я") || strings.ContainsRune(turn.Speaker, '\uFFFD') {
		t.Errorf("speaker cut mid-rune: %q", turn.Speaker)
	}
}

func TestNormTranscript_SpeakerIsOneLine(t *testing.T) {
	doc := normDoc(t, `[{"type":"transcript","data":{"turns":[{"speaker":"Anna\nSmith","text":"line one\nline two"}]}}]`)
	turn := transcriptTurns(doc.Blocks[0].Data)[0]
	if strings.Contains(turn.Speaker, "\n") {
		t.Errorf("speaker kept a newline: %q", turn.Speaker)
	}
	// The turn itself is prose and keeps its shape.
	if !strings.Contains(turn.Text, "\n") {
		t.Errorf("turn text lost its line break: %q", turn.Text)
	}
}

func TestTranscriptText_IndexesSpeakerWithLine(t *testing.T) {
	doc := normDoc(t, `[{"type":"transcript","data":{"title":"Call","turns":[
		{"speaker":"Пётр","text":"Контур закрыт."},
		{"speaker":"Аня","text":"Сканер внутрь — см. [[E4]]."}
	]}}]`)
	text := BlockPlainText(doc)
	for _, want := range []string{"Call", "Пётр", "Контур закрыт.", "[[E4]]"} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q from the index:\n%s", want, text)
		}
	}
	// The speaker leads the lines they said, so "who said the thing about the
	// gateway" is still answerable — but the name appears once per run, not once
	// per turn, because that is what the page draws.
	if !strings.Contains(text, "Пётр\nКонтур закрыт.") {
		t.Errorf("the speaker does not lead their line:\n%s", text)
	}
}

// A selection running across two consecutive turns by one speaker has to be a
// substring of the projection, or the mark it makes is born orphaned and is
// never drawn. The renderer prints the name once above the run; repeating it
// between every line put a word in the haystack the reader could not have
// selected.
func TestTranscriptText_GroupsRunsTheWayTheRendererDoes(t *testing.T) {
	doc := normDoc(t, `[{"type":"transcript","data":{"turns":[
		{"speaker":"Sam","text":"Three things went wrong."},
		{"speaker":"Sam","text":"The migration ran on the wrong replica."},
		{"speaker":"Sam","text":"Nobody was watching the queue depth."},
		{"speaker":"Anna","text":"Then we roll back."}
	]}}]`)
	text := BlockPlainText(doc)
	if got := strings.Count(text, "Sam"); got != 1 {
		t.Errorf("the speaker is named %d times, want once per run:\n%s", got, text)
	}
	// The gesture: select the second and third lines and mark them.
	selection := "The migration ran on the wrong replica.\nNobody was watching the queue depth."
	if !strings.Contains(text, selection) {
		t.Errorf("a mark across two turns by one speaker cannot anchor:\n%s", text)
	}
	if !strings.Contains(text, "Anna") {
		t.Errorf("a new speaker starts a new run:\n%s", text)
	}
}

func TestTranscript_ReportsWhatTheCapThrewAway(t *testing.T) {
	var turns []string
	for i := 0; i < domain.MaxTranscriptTurns+120; i++ {
		turns = append(turns, `{"speaker":"A","text":"line"}`)
	}
	doc := normDoc(t, `[{"type":"transcript","data":{"title":"Long call","turns":[`+strings.Join(turns, ",")+`]}}]`)
	if got := intOr(doc.Blocks[0].Data, "truncated", 0); got != 120 {
		t.Errorf("truncated = %d, want 120 — a silent cut reads as a shorter call", got)
	}
	if md := BlockDocumentToMarkdown(doc); !strings.Contains(md, "120 further turns were not stored") {
		t.Errorf("the export does not say the transcript is incomplete:\n%s", md[:200])
	}
}

func TestTranscriptMarkdown_ReadsAsATranscript(t *testing.T) {
	doc := normDoc(t, `[{"type":"transcript","data":{"title":"Infrastructure call","turns":[
		{"speaker":"Peter","text":"The perimeter is closed.","ts":"00:03:12"},
		{"speaker":"Anna","text":"Then the scanner goes inside."},
		{"text":"Unattributed."}
	]}}]`)
	md := BlockDocumentToMarkdown(doc)
	for _, want := range []string{
		"**Infrastructure call**",
		"**Peter** *(00:03:12)*: The perimeter is closed.",
		"**Anna**: Then the scanner goes inside.",
		"Unattributed.",
	} {
		if !strings.Contains(md, want) {
			t.Errorf("missing %q:\n%s", want, md)
		}
	}
	// An unattributed turn must not open with a stray colon.
	if strings.Contains(md, ": Unattributed.") {
		t.Errorf("unattributed turn got a label anyway:\n%s", md)
	}
}

func TestTranscriptTitle_NamesADocumentThatIsOnlyAConversation(t *testing.T) {
	doc := normDoc(t, `[{"type":"transcript","data":{"title":"Infrastructure call","turns":[{"speaker":"A","text":"The perimeter is closed."}]}}]`)
	if got := BlockDocumentTitle(doc); got != "Infrastructure call" {
		t.Errorf("title = %q, want the transcript's own", got)
	}
	if got := BlockDocumentDescription(doc, "Infrastructure call"); got != "The perimeter is closed." {
		t.Errorf("description = %q, want the opening line", got)
	}
}
