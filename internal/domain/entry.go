package domain

import "time"

type EntryStatus string

const (
	EntryDraft     EntryStatus = "draft"
	EntryActive    EntryStatus = "active"
	EntryCompleted EntryStatus = "completed"
	EntryArchived  EntryStatus = "archived"
)

// EntryType selects how Content is interpreted when rendered.
type EntryType string

const (
	// EntryMarkdown is the default: Content is markdown.
	EntryMarkdown EntryType = "markdown"
	// EntryBlocks means Content is a block document — see block.go. Structured,
	// safe to render, indexable and exportable to markdown.
	EntryBlocks EntryType = "blocks"
	// EntryArtifact is accepted on input as sugar for a blocks document holding a
	// single `html` block, and is normalized to EntryBlocks on the way in. It is
	// no longer a stored type. Kept so clients that learned it keep working.
	EntryArtifact EntryType = "artifact"
)

// Valid reports whether the type may be supplied by a caller.
func (t EntryType) Valid() bool {
	return t == EntryMarkdown || t == EntryBlocks || t == EntryArtifact
}

// Stored reports whether the type may be persisted. EntryArtifact is an input
// alias only — storing it would leave two rendering paths for one thing.
func (t EntryType) Stored() bool {
	return t == EntryMarkdown || t == EntryBlocks
}

// BlockSaveReport tells a writer what happened to state it never mentioned:
// how many blocks arrived with an unrecognized id, and how many ticks that cost.
// It is attached to the response of a write and never stored — an agent that
// dropped block ids has to be told, because the loss is otherwise silent.
type BlockSaveReport struct {
	Blocks         int `json:"blocks"`
	Reidentified   int `json:"blocks_reidentified"`
	StatePreserved int `json:"state_preserved"`
	StateLost      int `json:"state_lost"`
}

type Entry struct {
	ID          string      `json:"id"`
	Code        string      `json:"code"`
	ResearchID  string      `json:"research_id"`
	SectionID   string      `json:"section_id"`
	SessionID   string      `json:"session_id,omitempty"`
	Type        EntryType   `json:"entry_type"`
	Title       string      `json:"title"`
	Content     string      `json:"content,omitempty"`
	Description string      `json:"description"`
	Status      EntryStatus `json:"status"`
	Tags        []string    `json:"tags"`
	// Metadata holds the values for the fields this entry's section declares,
	// keyed by field key. A key the section no longer declares keeps its value
	// here rather than being deleted — see domain.OrphanedKeys.
	Metadata map[string]any `json:"metadata,omitempty"`
	// SpecVersion is the section spec version these values were last validated
	// against, so a reader can tell which rule the document was held to.
	SpecVersion int       `json:"spec_version,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	// Set on a write to a blocks entry, never persisted. See BlockSaveReport.
	BlockReport *BlockSaveReport `json:"block_report,omitempty"`
	// Set on a write that carried metadata, never persisted. See MetadataReport.
	MetaReport *MetadataReport `json:"metadata_report,omitempty"`
	// Computed on read against the section's current declaration, never
	// persisted — a stored completeness flag goes stale the moment the spec
	// changes.
	MetaStatus *MetadataStatus `json:"metadata_status,omitempty"`
}

// MetadataStatus is how a stored document stands against what its section
// currently declares. It is computed on read, every read.
type MetadataStatus struct {
	// MissingRequired names declared, required fields this document does not
	// answer. A document is incomplete, never invalid: a missing element has
	// never made a record inauthentic, it annotates it.
	MissingRequired []string `json:"missing_required,omitempty"`
	// Orphaned names values kept under keys the section no longer declares.
	Orphaned []string `json:"orphaned,omitempty"`
	// Issues are stored values that do not match their declared type. They were
	// kept rather than discarded — dropping a person's value to protect a type
	// is the same mistake as refusing the write — and they are re-checked on
	// every read so the reader who can fix one is the one who is told.
	Issues   []MetadataIssue `json:"issues,omitempty"`
	Complete bool            `json:"complete"`
	// SpecVersion is the section's current version, which may be ahead of the
	// entry's own — that difference is exactly what lazy top-up works through.
	SpecVersion int `json:"spec_version"`
}
