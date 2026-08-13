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
//  4. the catalog in internal/docs/llms/blocks.md
//
// blockNormalizer validates and clamps one block's data, returning ok=false to
// drop the block.
type blockNormalizer func(data map[string]any) (map[string]any, bool)

var blockNormalizers = map[domain.BlockType]blockNormalizer{
	domain.BlockParagraph: normParagraph,
	domain.BlockHeading:   normHeading,
	domain.BlockList:      normList,
	domain.BlockTable:     normTable,
	domain.BlockQuote:     normQuote,
	domain.BlockCode:      normCode,
	domain.BlockCallout:   normCallout,
	domain.BlockDivider:   normDivider,
	domain.BlockImage:     normImage,
	domain.BlockHTML:      normHTML,
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
