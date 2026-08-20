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
	// A task_ref is a working list, not a project plan: past a few dozen rows
	// nobody reads it, and every row costs a resolution.
	MaxTaskRefs = 50
	// A one-hour call is a few hundred turns. The cap is generous enough that a
	// real transcript survives it whole and low enough that a runaway paste
	// cannot bloat the row.
	MaxTranscriptTurns    = 500
	MaxTranscriptSpeaker  = 120
	MaxTranscriptStamp    = 40
	MaxTranscriptTurnText = 5000
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
	// BlockChecklist is the one block whose state a human owns. Items carry a
	// stable key for the same reason blocks carry an id: inserting an item above
	// a ticked one must not move the tick. The ticks themselves are never in the
	// author's payload — see service.normChecklist.
	BlockChecklist BlockType = "checklist"
	// BlockMermaid holds a mermaid source drawn as a diagram. It is deliberately
	// separate from BlockCode: a diagram is a figure with a caption, not a
	// listing, and an exporter has to know which one it is looking at.
	BlockMermaid BlockType = "mermaid"
	// BlockHTML holds a self-contained HTML document rendered in a sandboxed
	// iframe. This is what the standalone `artifact` entry type became: as a
	// block it composes with prose instead of taking over the whole entry.
	BlockHTML BlockType = "html"
	// BlockTaskRef projects existing tasks into a document as a checklist.
	//
	// It holds REFERENCES and nothing else. The product already has Task — a
	// status, a priority, a board, four MCP tools — and a checklist with state of
	// its own would create a second place where "what has to be done" lives, with
	// neither authoritative: the board would not know what was ticked in an
	// article, and an agent reading tasks over MCP would not see it either. So a
	// tick here is a status change on the task, and the document is a view.
	//
	// Contrast BlockChecklist, whose state genuinely belongs to the document.
	BlockTaskRef BlockType = "task_ref"
	// BlockTranscript stores a conversation that happened outside the tool — a
	// call, a meeting, an interview — as structured turns.
	//
	// The product models the interview it runs itself as sessions and questions.
	// This is the other kind: dropped into a code block a transcript would be
	// stored and lose everything — speakers unsearchable, no single line to quote
	// or link, nothing for "make a task out of what he said at 14:32" to point at.
	BlockTranscript BlockType = "transcript"
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
