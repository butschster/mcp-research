package service

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/google/uuid"
)

// Block document normalization.
//
// The single entry point is NormalizeBlockDocument, which every writer calls
// before persisting. It is deliberately forgiving: the author is usually an AI
// agent over MCP, so an unknown block type or malformed data is DROPPED and the
// rest of the document is kept. A bad payload degrades to fewer blocks rather
// than failing the call — the caller checks what came back to see what stored.
//
// Adding a block type means touching four places that must agree, and a hook
// (.claude/hooks/lint-blocks.py) fails the edit when they drift:
//  1. domain.BlockType constant
//  2. a normalizer here
//  3. a branch in frontend/components/blocks/BlockRenderer.vue
//  4. the catalog in internal/docs/blocks.md
//
// blockNormalizer validates and clamps one block's data, returning ok=false to
// drop the block.
type blockNormalizer func(data map[string]any) (map[string]any, bool)

var blockNormalizers = map[domain.BlockType]blockNormalizer{
	domain.BlockParagraph:  normParagraph,
	domain.BlockHeading:    normHeading,
	domain.BlockList:       normList,
	domain.BlockTable:      normTable,
	domain.BlockQuote:      normQuote,
	domain.BlockCode:       normCode,
	domain.BlockCallout:    normCallout,
	domain.BlockDivider:    normDivider,
	domain.BlockImage:      normImage,
	domain.BlockChecklist:  normChecklist,
	domain.BlockMermaid:    normMermaid,
	domain.BlockHTML:       normHTML,
	domain.BlockTaskRef:    normTaskRef,
	domain.BlockTranscript: normTranscript,
}

// ParseStoredBlockDocument reads a document the server itself wrote, without
// normalizing it.
//
// Normalization is for input: it strips `state`, because ticks are server-owned
// and an author must not be able to assert them. Running it over stored content
// on the way OUT would therefore erase exactly what a reader came for — which is
// how the markdown export first came to show every box unticked.
func ParseStoredBlockDocument(raw string) (*domain.BlockDocument, error) {
	var doc domain.BlockDocument
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &doc); err != nil {
		var bare []domain.Block
		if err2 := json.Unmarshal([]byte(strings.TrimSpace(raw)), &bare); err2 != nil {
			return nil, fmt.Errorf("content is not a block document: %w", err)
		}
		doc.Blocks = bare
	}
	if doc.Version == 0 {
		doc.Version = domain.BlockDocumentVersion
	}
	return &doc, nil
}

// NormalizeBlockDocument parses and normalizes a stored or incoming document.
// It never returns an error for bad blocks — only for input that is not a
// document at all, so the caller can tell "you sent something else entirely"
// from "some blocks were dropped".
func NormalizeBlockDocument(raw string) (*domain.BlockDocument, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("content is empty: a block document needs at least one block")
	}

	var doc domain.BlockDocument
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		// A bare array is accepted too: agents reach for it, and refusing would
		// buy nothing.
		var bare []domain.Block
		if err2 := json.Unmarshal([]byte(raw), &bare); err2 != nil {
			return nil, fmt.Errorf("content is not a block document: expected {\"version\":1,\"blocks\":[...]} or a bare blocks array")
		}
		doc = domain.BlockDocument{Blocks: bare}
	}

	doc.Version = domain.BlockDocumentVersion

	kept := make([]domain.Block, 0, len(doc.Blocks))
	seenIDs := make(map[string]bool, len(doc.Blocks))
	for _, b := range doc.Blocks {
		if len(kept) >= domain.MaxBlocks {
			break
		}
		norm, ok := blockNormalizers[b.Type]
		if !ok {
			continue // unknown type — dropped, never an error
		}
		data, ok := norm(b.Data)
		if !ok {
			continue
		}
		kept = append(kept, domain.Block{
			ID:   blockID(b.ID, seenIDs),
			Type: b.Type,
			Data: data,
		})
	}
	doc.Blocks = kept

	if len(doc.Blocks) == 0 {
		return nil, fmt.Errorf("no valid blocks: every block was dropped as unknown or malformed")
	}
	return &doc, nil
}

// MarshalBlockDocument renders a document back to the stored string form.
func MarshalBlockDocument(doc *domain.BlockDocument) (string, error) {
	b, err := json.Marshal(doc)
	if err != nil {
		return "", fmt.Errorf("marshal block document: %w", err)
	}
	return string(b), nil
}

// ArtifactToBlockDocument wraps a bare HTML document as a single html block.
// This is what the `artifact` entry type normalizes into.
//
// The document's own <title> and <meta name="description"> are lifted onto the
// block, so an artifact still names itself the way it always did — the entry
// title is derived from the block, not from the raw HTML.
func ArtifactToBlockDocument(html string) *domain.BlockDocument {
	data := map[string]any{
		"html": clampStr(sanitizeUTF8(html), domain.MaxInlineHTML),
	}
	if t := clampStr(normalizeTitle(htmlTitle(html)), domain.MaxHeadingText); t != "" {
		data["title"] = t
	}
	if d := clampStr(normalizeContent(htmlMetaDescription(html)), domain.MaxCaptionText); d != "" {
		data["caption"] = d
	}
	return &domain.BlockDocument{
		Version: domain.BlockDocumentVersion,
		Blocks:  []domain.Block{{Type: domain.BlockHTML, Data: data}},
	}
}

// blockIDRule is the same rule in words, for an error a caller can act on.
const blockIDRule = "4 to 16 lowercase letters or digits"

var blockIDPattern = regexp.MustCompile(`^[a-z0-9]{4,16}$`)

// blockID keeps a caller's id when it is safe and unique within the document, and
// mints one otherwise. Keeping it is what lets state attached to a block survive
// the agent rewriting the document; minting is what makes every block addressable
// even when the author did not think about ids at all.
func blockID(given string, seen map[string]bool) string {
	id := strings.ToLower(strings.TrimSpace(given))
	if !blockIDPattern.MatchString(id) || seen[id] {
		id = newBlockID()
		for seen[id] {
			id = newBlockID()
		}
	}
	seen[id] = true
	return id
}

func newBlockID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")[:domain.BlockIDLength]
}

// --- per-block normalizers -------------------------------------------------

func normParagraph(d map[string]any) (map[string]any, bool) {
	text := clampStr(normalizeContent(str(d, "text")), domain.MaxBlockText)
	if text == "" {
		return nil, false
	}
	return map[string]any{"text": text}, true
}

func normHeading(d map[string]any) (map[string]any, bool) {
	text := clampStr(normalizeTitle(str(d, "text")), domain.MaxHeadingText)
	if text == "" {
		return nil, false
	}
	level := intOr(d, "level", 2)
	if !domain.HeadingLevels[level] {
		level = 2
	}
	return map[string]any{"level": level, "text": text}, true
}

func normList(d map[string]any) (map[string]any, bool) {
	style := str(d, "style")
	if style != domain.ListOrdered {
		style = domain.ListUnordered
	}
	raw, _ := d["items"].([]any)
	items := make([]any, 0, len(raw))
	for _, it := range raw {
		if len(items) >= domain.MaxListItems {
			break
		}
		s, ok := it.(string)
		if !ok {
			continue
		}
		s = clampStr(normalizeContent(s), domain.MaxBlockText)
		if s == "" {
			continue // blank items are noise, not content
		}
		items = append(items, s)
	}
	if len(items) == 0 {
		return nil, false
	}
	return map[string]any{"style": style, "items": items}, true
}

func normTable(d map[string]any) (map[string]any, bool) {
	rawRows, _ := d["rows"].([]any)
	rows := make([]any, 0, len(rawRows))
	for _, r := range rawRows {
		if len(rows) >= domain.MaxTableRows {
			break
		}
		rawCells, ok := r.([]any)
		if !ok {
			continue
		}
		cells := make([]any, 0, len(rawCells))
		for _, c := range rawCells {
			if len(cells) >= domain.MaxTableCols {
				break
			}
			s, _ := c.(string)
			cells = append(cells, clampStr(normalizeContent(s), domain.MaxBlockText))
		}
		if len(cells) == 0 {
			continue
		}
		rows = append(rows, cells)
	}
	if len(rows) == 0 {
		return nil, false
	}
	return map[string]any{
		"header": boolOr(d, "header", true),
		"rows":   rows,
	}, true
}

func normQuote(d map[string]any) (map[string]any, bool) {
	text := clampStr(normalizeContent(str(d, "text")), domain.MaxBlockText)
	if text == "" {
		return nil, false
	}
	out := map[string]any{"text": text}
	if cite := clampStr(normalizeTitle(str(d, "cite")), domain.MaxHeadingText); cite != "" {
		out["cite"] = cite
	}
	return out, true
}

func normCode(d map[string]any) (map[string]any, bool) {
	// Code is stored verbatim: expanding literal \n here would corrupt a snippet
	// that legitimately contains a backslash escape.
	code := clampStr(sanitizeUTF8(str(d, "code")), domain.MaxCodeText)
	if strings.TrimSpace(code) == "" {
		return nil, false
	}
	out := map[string]any{"code": code}
	if lang := clampStr(slugish(str(d, "language")), domain.MaxLanguageIdent); lang != "" {
		out["language"] = lang
	}
	return out, true
}

func normCallout(d map[string]any) (map[string]any, bool) {
	text := clampStr(normalizeContent(str(d, "text")), domain.MaxCalloutText)
	if text == "" {
		return nil, false
	}
	variant := str(d, "variant")
	switch variant {
	case domain.CalloutInfo, domain.CalloutWarning, domain.CalloutSuccess, domain.CalloutDanger:
	default:
		variant = domain.CalloutInfo
	}
	out := map[string]any{"variant": variant, "text": text}
	if title := clampStr(normalizeTitle(str(d, "title")), domain.MaxHeadingText); title != "" {
		out["title"] = title
	}
	return out, true
}

func normDivider(map[string]any) (map[string]any, bool) {
	return map[string]any{}, true
}

// normChecklist accepts items as plain strings or as {key, text} objects, and
// mints a key for any item that arrives without one — the same preserve-or-mint
// rule as block ids, one level down.
//
// It deliberately does NOT read `state`. Ticks are server-owned and live in the
// block's row; dropping whatever an author sends is what gives that field a
// single origin, and it is why an agent rewriting a document cannot clear them.
// itemKeyPattern is looser than blockIDPattern on purpose: an item key only has
// to be unique inside its block, and an author writing "k1" or "backup" should
// keep it rather than have it silently replaced — a replaced key is a lost tick.
var itemKeyPattern = regexp.MustCompile(`^[a-z0-9_-]{1,24}$`)

// itemKey preserves a safe, unused key and mints one otherwise, the same
// preserve-or-mint rule as block ids.
func itemKey(given string, seen map[string]bool) string {
	key := strings.ToLower(strings.TrimSpace(given))
	if !itemKeyPattern.MatchString(key) || seen[key] {
		key = newBlockID()
	}
	seen[key] = true
	return key
}

func normChecklist(d map[string]any) (map[string]any, bool) {
	raw, _ := d["items"].([]any)
	seen := map[string]bool{}
	items := make([]any, 0, len(raw))
	for _, it := range raw {
		var key, text string
		switch v := it.(type) {
		case string:
			text = v
		case map[string]any:
			key, _ = v["key"].(string)
			text, _ = v["text"].(string)
		default:
			continue
		}
		text = clampStr(normalizeContent(text), domain.MaxBlockText)
		if strings.TrimSpace(text) == "" {
			continue
		}
		items = append(items, map[string]any{"key": itemKey(key, seen), "text": text})
		if len(items) >= domain.MaxListItems {
			break
		}
	}
	if len(items) == 0 {
		return nil, false
	}
	out := map[string]any{"items": items}
	if t := clampStr(normalizeTitle(str(d, "title")), domain.MaxHeadingText); t != "" {
		out["title"] = t
	}
	return out, true
}

func normMermaid(d map[string]any) (map[string]any, bool) {
	// Verbatim, like code: mermaid has its own escapes, and expanding literal
	// \n here would rewrite a label that contains one.
	code := clampStr(sanitizeUTF8(str(d, "code")), domain.MaxMermaidText)
	if strings.TrimSpace(code) == "" {
		return nil, false
	}
	out := map[string]any{"code": code}
	if cap := clampStr(normalizeContent(str(d, "caption")), domain.MaxCaptionText); cap != "" {
		out["caption"] = cap
	}
	return out, true
}

func normImage(d map[string]any) (map[string]any, bool) {
	url, ok := safeURL(str(d, "url"))
	if !ok {
		return nil, false
	}
	out := map[string]any{"url": url}
	if alt := clampStr(normalizeTitle(str(d, "alt")), domain.MaxCaptionText); alt != "" {
		out["alt"] = alt
	}
	if cap := clampStr(normalizeContent(str(d, "caption")), domain.MaxCaptionText); cap != "" {
		out["caption"] = cap
	}
	return out, true
}

func normHTML(d map[string]any) (map[string]any, bool) {
	// Stored verbatim, rendered in a sandboxed iframe. No sanitizing here on
	// purpose: the isolation is the boundary, and stripping tags would break
	// legitimate documents while providing no guarantee.
	html := clampStr(sanitizeUTF8(str(d, "html")), domain.MaxInlineHTML)
	if strings.TrimSpace(html) == "" {
		return nil, false
	}
	out := map[string]any{"html": html}
	if title := clampStr(normalizeTitle(str(d, "title")), domain.MaxHeadingText); title != "" {
		out["title"] = title
	}
	if cap := clampStr(normalizeContent(str(d, "caption")), domain.MaxCaptionText); cap != "" {
		out["caption"] = cap
	}
	return out, true
}

// taskRefPattern is a task short code. Codes are per-research and the block
// lives in an entry, so the research is implied — a bare `T4` is the whole
// reference and there is no cross-research form on purpose: a document that
// checks off somebody else's task would be a write across a boundary the rest of
// the product keeps closed.
var taskRefPattern = regexp.MustCompile(`^T[0-9]{1,9}$`)

// uuidPattern accepts the other spelling of a task reference. An agent that has
// just called task_create holds the id and not the code, and making it fetch the
// code first buys nothing.
var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// normTaskRef keeps the references and nothing else.
//
// It validates the SHAPE of a reference and never whether the task exists:
// normalization is pure — it takes a document and returns a document, touching
// no other table — and a reference that resolves today may not tomorrow anyway.
// Resolution happens where the block is read, and a reference that resolves to
// nothing is dropped from the render rather than shown as a ghost.
func normTaskRef(d map[string]any) (map[string]any, bool) {
	raw, _ := d["tasks"].([]any)
	seen := make(map[string]bool, len(raw))
	refs := make([]any, 0, len(raw))
	// Say what was cut. A cap that drops silently reads to both the writer and
	// the reader as a document that is simply shorter than the one that was
	// sent — the create returns 200 and the page draws a confident list.
	dropped := 0
	for _, it := range raw {
		s, ok := it.(string)
		if !ok {
			continue
		}
		ref := strings.TrimSpace(s)
		// A code is upper-case, a uuid is lower-case, and an author types either.
		if up := strings.ToUpper(ref); taskRefPattern.MatchString(up) {
			ref = up
		} else if low := strings.ToLower(ref); uuidPattern.MatchString(low) {
			ref = low
		} else {
			continue
		}
		if seen[ref] {
			continue // the same task twice is one row, and two checkboxes for one
		}
		seen[ref] = true
		// Counted past the cap but not kept: a runaway payload must not be held
		// in memory to be told how long it was.
		if len(refs) >= domain.MaxTaskRefs {
			dropped++
			continue
		}
		refs = append(refs, ref)
	}
	if len(refs) == 0 {
		return nil, false
	}
	out := map[string]any{
		"tasks": refs,
		// Stamped rather than carried through: absent means "yes" for an author
		// who did not think about it, and the renderer should not have to know
		// that a missing key and a false are different things.
		"show_progress": boolOr(d, "show_progress", true),
	}
	if note := clampStr(normalizeContent(str(d, "note")), domain.MaxCaptionText); note != "" {
		out["note"] = note
	}
	if n := carriedTruncation(d, dropped); n > 0 {
		out[truncatedKey] = n
	}
	return out, true
}

// truncatedKey counts what a cap threw away.
//
// Carried forward rather than recomputed, because normalization runs TWICE on a
// create — once on the author's payload and once on the document that produced —
// and a second pass over 500 stored turns drops nothing and would erase the
// first pass's count. Idempotence is the property that matters: normalizing an
// already-normal document must give the same document back.
//
// The cost is that an author can assert a `truncated` nobody earned. It buys
// them nothing: the field only ever makes a document claim to be LESS complete
// than it is, which is the safe direction for a number whose whole job is to
// stop a document from looking finished when it is not.
const truncatedKey = "truncated"

// carriedTruncation is what this pass dropped plus what a previous one recorded.
func carriedTruncation(d map[string]any, dropped int) int {
	prior := intOr(d, truncatedKey, 0)
	if prior < 0 {
		prior = 0
	}
	return dropped + prior
}

// normTranscript clamps a conversation into turns.
//
// A turn with no text is dropped: a speaker who said nothing is not a turn, and
// an empty line in a rendered transcript reads as a transcription failure. A
// transcript with no turns left is dropped whole, like every other empty block.
func normTranscript(d map[string]any) (map[string]any, bool) {
	raw, _ := d["turns"].([]any)
	turns := make([]any, 0, len(raw))
	dropped := 0
	for _, it := range raw {
		m, ok := it.(map[string]any)
		if !ok {
			continue
		}
		text := clampStr(normalizeContent(str(m, "text")), domain.MaxTranscriptTurnText)
		if text == "" {
			continue
		}
		turn := map[string]any{"text": text}
		// The speaker is a label, not an entity: one line, plain text, indexed.
		// Binding it to a user would make a transcript an audit log, which it is
		// not — the person quoted may not have an account here at all.
		if sp := clampStr(normalizeTitle(str(m, "speaker")), domain.MaxTranscriptSpeaker); sp != "" {
			turn["speaker"] = sp
		}
		// Displayed, never parsed: `00:03:12`, `14:32` and an ISO stamp are all
		// legitimate, and deciding which one this is would only create a class of
		// transcript the product refuses.
		if ts := clampStr(normalizeTitle(str(m, "ts")), domain.MaxTranscriptStamp); ts != "" {
			turn["ts"] = ts
		}
		if len(turns) >= domain.MaxTranscriptTurns {
			dropped++
			continue
		}
		turns = append(turns, turn)
	}
	if len(turns) == 0 {
		return nil, false
	}
	out := map[string]any{"turns": turns}
	// A one-hour call runs past this cap, and a transcript that stops mid-sentence
	// while the header says "500 turns" is a document that lies about itself.
	if n := carriedTruncation(d, dropped); n > 0 {
		out[truncatedKey] = n
	}
	if t := clampStr(normalizeTitle(str(d, "title")), domain.MaxHeadingText); t != "" {
		out["title"] = t
	}
	return out, true
}

// --- helpers ---------------------------------------------------------------

func str(d map[string]any, key string) string {
	s, _ := d[key].(string)
	return s
}

func intOr(d map[string]any, key string, def int) int {
	switch v := d[key].(type) {
	case float64: // JSON numbers decode as float64
		return int(v)
	case int:
		return v
	}
	return def
}

func boolOr(d map[string]any, key string, def bool) bool {
	if b, ok := d[key].(bool); ok {
		return b
	}
	return def
}

func clampStr(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	// Cut on a rune boundary so a clamp cannot produce invalid UTF-8.
	cut := s[:max]
	for len(cut) > 0 && !isUTF8Boundary(cut) {
		cut = cut[:len(cut)-1]
	}
	return cut
}

func isUTF8Boundary(s string) bool {
	return len([]rune(s)) > 0 && len(string([]rune(s))) == len(s)
}

// safeURL accepts absolute http(s) and domain-relative URLs only, so a block
// cannot smuggle javascript: or data: into an href or src.
func safeURL(raw string) (string, bool) {
	u := clampStr(raw, domain.MaxURLLength)
	if u == "" {
		return "", false
	}
	lower := strings.ToLower(u)
	switch {
	case strings.HasPrefix(lower, "http://"), strings.HasPrefix(lower, "https://"):
		return u, true
	case strings.HasPrefix(u, "/") && !strings.HasPrefix(u, "//"):
		return u, true
	}
	return "", false
}

// slugish keeps a language identifier to the characters highlighters use.
func slugish(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '+', r == '#', r == '-', r == '_', r == '.':
			b.WriteRune(r)
		}
	}
	return b.String()
}
