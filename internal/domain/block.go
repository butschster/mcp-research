package domain

// Block document model for entries whose Type is EntryBlocks.
//
// A document is {version, blocks: [{type, data}]} and the array order is the
// render order. Every block is a typed unit the frontend renders with a
// per-type branch; `data` is kept as raw JSON here so the domain package stays
// free of per-block structs and the normalizer owns the shape.
//
// The format is deliberately forgiving on input: the author is usually an AI
// agent over MCP, so an unknown type or malformed data is DROPPED rather than
// rejected. A bad payload degrades to fewer blocks instead of failing the whole
// call. See service.NormalizeBlockDocument.

// BlockDocumentVersion is the current envelope version.
const BlockDocumentVersion = 1

// Soft caps so a runaway payload cannot bloat a row or the page.
const (
	MaxBlocks        = 400
	MaxBlockText     = 20000
	MaxListItems     = 200
	MaxTableRows     = 200
	MaxTableCols     = 20
	MaxInlineHTML    = 200000
	MaxHeadingText   = 300
	MaxCaptionText   = 1000
	MaxURLLength     = 2000
	MaxCalloutText   = 5000
	MaxCodeText      = 100000
	MaxMermaidText   = 20000
	MaxLanguageIdent = 40
)

// BlockType enumerates the block kinds the renderer knows. An unknown type is
// dropped by the normalizer, so this list and the renderer must stay in step —
// enforced by .claude/hooks/lint-blocks.py.
type BlockType string

const (
	BlockParagraph BlockType = "paragraph"
	BlockHeading   BlockType = "heading"
	BlockList      BlockType = "list"
	BlockTable     BlockType = "table"
	BlockQuote     BlockType = "quote"
	BlockCode      BlockType = "code"
	BlockCallout   BlockType = "callout"
	BlockDivider   BlockType = "divider"
	BlockImage     BlockType = "image"
	// BlockMermaid holds a mermaid source drawn as a diagram. It is deliberately
	// separate from BlockCode: a diagram is a figure with a caption, not a
	// listing, and an exporter has to know which one it is looking at.
	BlockMermaid BlockType = "mermaid"
	// BlockHTML holds a self-contained HTML document rendered in a sandboxed
	// iframe. This is what the standalone `artifact` entry type became: as a
	// block it composes with prose instead of taking over the whole entry.
	BlockHTML BlockType = "html"
)

// Callout variants.
const (
	CalloutInfo    = "info"
	CalloutWarning = "warning"
	CalloutSuccess = "success"
	CalloutDanger  = "danger"
)

// List styles.
const (
	ListUnordered = "unordered"
	ListOrdered   = "ordered"
)

// HeadingLevels are the levels an author may use. Level 1 is reserved for the
// entry title, which is rendered by the page rather than the document.
var HeadingLevels = map[int]bool{2: true, 3: true, 4: true}

// BlockIDLength is the length of a generated block id: short enough not to bloat
// a document an agent reads, long enough that a collision inside one document is
// not a practical concern.
const BlockIDLength = 8

// Block is one unit of a block document.
type Block struct {
	// ID is stable across updates: anything that has to point AT a block rather
	// than at its content — a checkbox's state, a comment, a deep link — needs an
	// identity that survives inserting a paragraph above it, which a position in
	// the array does not. The server fills it in when absent and preserves it when
	// present, so an agent that round-trips a document keeps whatever is attached
	// to its blocks.
	ID   string         `json:"id,omitempty"`
	Type BlockType      `json:"type"`
	Data map[string]any `json:"data"`
}

// BlockDocument is the stored content of an EntryBlocks entry.
type BlockDocument struct {
	Version int     `json:"version"`
	Blocks  []Block `json:"blocks"`
}
