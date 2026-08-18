package service

import (
	"context"
	"errors"
	"testing"

	"github.com/butschster/mcp-research/internal/auth"
	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/storage"
)

// A section declares what its documents record. These tests are written from
// the two decisions that shape the feature: the vocabulary is closed, and the
// declaration is a lens rather than a rule the stored documents are held to
// retroactively.

type metaKit struct {
	*roleKit
	ctx      context.Context
	research *domain.Research
	section  *domain.Section
}

func newMetaKit(t *testing.T) *metaKit {
	t.Helper()
	k := newRoleKit(t)
	ctx := context.Background()
	research, sections, err := k.research.Create(ctx, CreateResearchRequest{
		Name: "Platform", Goal: "Specify the payload",
		Sections: []CreateSectionRequest{{Name: "specifications", DisplayName: "Specifications"}},
	})
	if err != nil {
		t.Fatalf("create research: %v", err)
	}
	return &metaKit{roleKit: k, ctx: ctx, research: research, section: sections[0]}
}

func (k *metaKit) declare(t *testing.T, specs ...domain.FieldSpec) *domain.Section {
	t.Helper()
	sec, err := k.roleKit.section.Update(k.ctx, k.section.ID, UpdateSectionRequest{FieldSpec: &specs})
	if err != nil {
		t.Fatalf("declare field spec: %v", err)
	}
	k.section = sec
	return sec
}

func (k *metaKit) write(t *testing.T, meta map[string]any) *domain.Entry {
	t.Helper()
	entry, err := k.entry.Create(k.ctx, CreateEntryRequest{
		ResearchID: k.research.ID, SectionID: k.section.ID,
		Title: "SPEC-01", Content: "the body", Metadata: meta,
	})
	if err != nil {
		t.Fatalf("create entry: %v", err)
	}
	return entry
}

func specStage() domain.FieldSpec {
	return domain.FieldSpec{Key: "stage", Label: "Stage", Type: domain.FieldEnum,
		Options: []string{"draft", "in-review", "agreed"}}
}

func TestMetadata_SectionWithNoDeclarationAcceptsNothing(t *testing.T) {
	k := newMetaKit(t)

	entry := k.write(t, map[string]any{"owner": "platform"})

	if len(entry.Metadata) != 0 {
		t.Fatalf("a section that declares nothing stored %v", entry.Metadata)
	}
	if entry.MetaReport == nil || len(entry.MetaReport.UnknownKeys) != 1 {
		t.Fatalf("the write did not report the refused key: %+v", entry.MetaReport)
	}
	// And the feature stays invisible where nobody asked for it.
	read, err := k.entry.Get(k.ctx, entry.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if read.MetaStatus != nil {
		t.Fatalf("a plain topic section grew a metadata status: %+v", read.MetaStatus)
	}
}

func TestMetadata_UndeclaredKeyIsDroppedAndDeclaredKeyIsKept(t *testing.T) {
	k := newMetaKit(t)
	k.declare(t, specStage())

	entry := k.write(t, map[string]any{"stage": "in-review", "owner": "platform"})

	if entry.Metadata["stage"] != "in-review" {
		t.Fatalf("declared key not stored: %v", entry.Metadata)
	}
	if _, ok := entry.Metadata["owner"]; ok {
		t.Fatal("an undeclared key was stored: the vocabulary is not closed")
	}
	if len(entry.MetaReport.UnknownKeys) != 1 || entry.MetaReport.UnknownKeys[0].Key != "owner" {
		t.Fatalf("unknown key not reported: %+v", entry.MetaReport)
	}
}

func TestMetadata_InvalidValueIsStoredAndFlagged(t *testing.T) {
	k := newMetaKit(t)
	k.declare(t, specStage())

	entry := k.write(t, map[string]any{"stage": "sent for review"})

	// Discarding a person's value to protect a type is the same mistake as
	// refusing the write.
	if entry.Metadata["stage"] != "sent for review" {
		t.Fatalf("an invalid value was dropped rather than flagged: %v", entry.Metadata)
	}
	if len(entry.MetaReport.InvalidValues) != 1 {
		t.Fatalf("invalid value not reported: %+v", entry.MetaReport)
	}

	// And it is still readable as invalid later, by the person who can fix it.
	read, _ := k.entry.Get(k.ctx, entry.ID)
	if read.MetaStatus == nil || len(read.MetaStatus.Issues) != 1 {
		t.Fatalf("a stored invalid value is not visible on read: %+v", read.MetaStatus)
	}
}

func TestMetadata_AddingAFieldMakesExistingDocumentsIncompleteAndRewritesNone(t *testing.T) {
	k := newMetaKit(t)
	entry := k.write(t, nil)

	k.declare(t, domain.FieldSpec{Key: "owner", Label: "Owner", Type: domain.FieldText, Required: true})

	read, err := k.entry.Get(k.ctx, entry.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if read.MetaStatus == nil || read.MetaStatus.Complete {
		t.Fatalf("the document should read as incomplete: %+v", read.MetaStatus)
	}
	if len(read.MetaStatus.MissingRequired) != 1 || read.MetaStatus.MissingRequired[0] != "owner" {
		t.Fatalf("missing required not reported: %+v", read.MetaStatus)
	}
	if read.UpdatedAt != entry.UpdatedAt {
		t.Fatal("declaring a field rewrote the document; the spec is a lens, not a migration")
	}
	// Completeness is computed, never stored: the entry row still carries no
	// values at all.
	if len(read.Metadata) != 0 {
		t.Fatalf("values appeared from nowhere: %v", read.Metadata)
	}
}

func TestMetadata_RemovingAFieldKeepsItsValues(t *testing.T) {
	k := newMetaKit(t)
	k.declare(t, specStage())
	entry := k.write(t, map[string]any{"stage": "agreed"})

	k.declare(t) // the section stops collecting anything

	read, _ := k.entry.Get(k.ctx, entry.ID)
	if read.Metadata["stage"] != "agreed" {
		t.Fatalf("removing a field deleted what was already recorded: %v", read.Metadata)
	}
	if read.MetaStatus == nil || len(read.MetaStatus.Orphaned) != 1 {
		t.Fatalf("the kept value is not surfaced as orphaned: %+v", read.MetaStatus)
	}
}

func TestMetadata_ExplicitUnknownAnswersARequiredField(t *testing.T) {
	k := newMetaKit(t)
	k.declare(t, domain.FieldSpec{Key: "owner", Label: "Owner", Type: domain.FieldText, Required: true})

	// A model that does not know must be able to say so, rather than filling the
	// field with something plausible.
	entry := k.write(t, map[string]any{"owner": nil})

	if len(entry.MetaReport.MissingRequired) != 0 {
		t.Fatalf("an explicit unknown did not answer the requirement: %+v", entry.MetaReport)
	}
	read, _ := k.entry.Get(k.ctx, entry.ID)
	if !read.MetaStatus.Complete {
		t.Fatalf("document reads incomplete despite an explicit unknown: %+v", read.MetaStatus)
	}
	if v, ok := read.Metadata["owner"]; !ok || v != nil {
		t.Fatalf("the unknown was not stored as one: %#v", read.Metadata)
	}
}

func TestMetadata_CompletedIsGatedAndOverridable(t *testing.T) {
	k := newMetaKit(t)
	k.declare(t, domain.FieldSpec{Key: "owner", Label: "Owner", Type: domain.FieldText, Required: true})
	entry := k.write(t, nil)

	completed := domain.EntryCompleted
	_, err := k.entry.Update(k.ctx, entry.ID, UpdateEntryRequest{Status: &completed})
	var incomplete *IncompleteMetadataError
	if !errors.As(err, &incomplete) {
		t.Fatalf("completing an incomplete document was allowed: %v", err)
	}
	if len(incomplete.Missing) != 1 || incomplete.Missing[0] != "owner" {
		t.Fatalf("the refusal does not name the field: %+v", incomplete.Missing)
	}

	// The override is a decision, not a retry.
	updated, err := k.entry.Update(k.ctx, entry.ID, UpdateEntryRequest{Status: &completed, AllowIncomplete: true})
	if err != nil {
		t.Fatalf("override refused: %v", err)
	}
	if updated.Status != domain.EntryCompleted {
		t.Fatalf("status = %q, want completed", updated.Status)
	}

	// Every other write is still accepted while required fields are empty.
	title := "SPEC-01 revised"
	if _, err := k.entry.Update(k.ctx, entry.ID, UpdateEntryRequest{Title: &title}); err != nil {
		t.Fatalf("an ordinary write was refused over metadata: %v", err)
	}
}

func TestMetadata_EditingOnlyMetadataLeavesARevision(t *testing.T) {
	k := newMetaKit(t)
	k.declare(t, specStage())
	entry := k.write(t, map[string]any{"stage": "draft"})

	_, revsBefore, err := k.entry.History(k.ctx, entry.ID)
	if err != nil {
		t.Fatalf("history: %v", err)
	}

	meta := map[string]any{"stage": "agreed"}
	if _, err := k.entry.Update(k.ctx, entry.ID, UpdateEntryRequest{Metadata: &meta}); err != nil {
		t.Fatalf("update: %v", err)
	}

	_, revsAfter, err := k.entry.History(k.ctx, entry.ID)
	if err != nil {
		t.Fatalf("history: %v", err)
	}
	// Without this the write is judged a no-op by SameContent and the change
	// does not merely go unrecorded — it disappears.
	if len(revsAfter) != len(revsBefore)+1 {
		t.Fatalf("a metadata-only edit produced %d revisions, want one more than %d", len(revsAfter), len(revsBefore))
	}
	if revsAfter[0].Metadata["stage"] != "agreed" {
		t.Fatalf("the revision did not capture the values: %v", revsAfter[0].Metadata)
	}
}

func TestMetadata_ReservedKeysAreRefusedWhenDeclared(t *testing.T) {
	k := newMetaKit(t)
	specs := []domain.FieldSpec{{Key: "status", Label: "Review status", Type: domain.FieldText}}
	_, err := k.roleKit.section.Update(k.ctx, k.section.ID, UpdateSectionRequest{FieldSpec: &specs})
	if err == nil {
		t.Fatal("a field keyed `status` was accepted; it would overwrite the system key in every export")
	}
}

func TestMetadata_CapsAreEnforced(t *testing.T) {
	k := newMetaKit(t)

	tooMany := make([]domain.FieldSpec, domain.MaxSectionFields+1)
	for i := range tooMany {
		tooMany[i] = domain.FieldSpec{Key: "f" + string(rune('a'+i)), Type: domain.FieldText}
	}
	if _, err := k.roleKit.section.Update(k.ctx, k.section.ID, UpdateSectionRequest{FieldSpec: &tooMany}); err == nil {
		t.Fatalf("more than %d fields was accepted", domain.MaxSectionFields)
	}

	tooManyRequired := make([]domain.FieldSpec, domain.MaxRequiredFields+1)
	for i := range tooManyRequired {
		tooManyRequired[i] = domain.FieldSpec{Key: "r" + string(rune('a'+i)), Type: domain.FieldText, Required: true}
	}
	if _, err := k.roleKit.section.Update(k.ctx, k.section.ID, UpdateSectionRequest{FieldSpec: &tooManyRequired}); err == nil {
		t.Fatalf("more than %d required fields was accepted", domain.MaxRequiredFields)
	}
}

func TestMetadata_SpecVersionMovesOnlyWhenTheDeclarationChanges(t *testing.T) {
	k := newMetaKit(t)
	first := k.declare(t, specStage())
	if first.SpecVersion != 1 {
		t.Fatalf("spec_version = %d after the first declaration, want 1", first.SpecVersion)
	}

	same := k.declare(t, specStage())
	if same.SpecVersion != 1 {
		t.Fatalf("saving an unchanged declaration bumped the version to %d", same.SpecVersion)
	}

	changed := k.declare(t, specStage(), domain.FieldSpec{Key: "owner", Type: domain.FieldText})
	if changed.SpecVersion != 2 {
		t.Fatalf("spec_version = %d after a real change, want 2", changed.SpecVersion)
	}
}

func TestMetadata_ShareLinkSeesNeitherValuesNorDeclaration(t *testing.T) {
	k := newShareKit(t)
	owner, _, research, section, _ := k.sharedResearch(t, domain.TeamViewer)

	specs := []domain.FieldSpec{{Key: "owner", Label: "Owner", Type: domain.FieldText}}
	if _, err := k.roleKit.section.Update(owner, section.ID, UpdateSectionRequest{FieldSpec: &specs}); err != nil {
		t.Fatalf("declare: %v", err)
	}
	entry, err := k.entry.Create(owner, CreateEntryRequest{
		ResearchID: research.ID, SectionID: section.ID,
		Title: "Internal", Content: "body", Metadata: map[string]any{"owner": "the acquisition team"},
	})
	if err != nil {
		t.Fatalf("create entry: %v", err)
	}

	result, err := k.shares.Create(owner, research.ID, CreateShareRequest{Include: allIncluded()})
	if err != nil {
		t.Fatalf("create share: %v", err)
	}
	ctx := visit(t, k, result.Token)

	read, err := k.entry.Get(ctx, entry.ID)
	if err != nil {
		t.Fatalf("share read: %v", err)
	}
	if len(read.Metadata) != 0 || read.MetaStatus != nil {
		t.Fatalf("a share link published document metadata: %v %+v", read.Metadata, read.MetaStatus)
	}

	listed, err := k.entry.ListByResearch(ctx, research.ID, storage.EntryFilter{})
	if err != nil {
		t.Fatalf("share list: %v", err)
	}
	for _, e := range listed {
		if len(e.Metadata) != 0 {
			t.Fatalf("the entry list published metadata: %v", e.Metadata)
		}
	}

	sections, err := k.roleKit.section.List(ctx, research.ID)
	if err != nil {
		t.Fatalf("share sections: %v", err)
	}
	for _, sec := range sections {
		// A declaration with no values still says what the team tracks.
		if len(sec.FieldSpec) != 0 {
			t.Fatalf("a share link published the field declaration: %v", sec.FieldSpec)
		}
	}
	_ = auth.ShareFromContext(ctx)
}

// The three regressions the review found, each written from the rule it broke.

func TestMetadata_AWriteDoesNotDeleteOrphanedValues(t *testing.T) {
	k := newMetaKit(t)
	k.declare(t, specStage(), domain.FieldSpec{Key: "owner", Label: "Owner", Type: domain.FieldText})
	entry := k.write(t, map[string]any{"stage": "agreed", "owner": "platform"})

	k.declare(t, specStage()) // the section stops collecting `owner`

	// Somebody edits a different field entirely.
	meta := map[string]any{"stage": "draft"}
	updated, err := k.entry.Update(k.ctx, entry.ID, UpdateEntryRequest{Metadata: &meta})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Metadata["owner"] != "platform" {
		t.Fatalf("an unrelated save destroyed an orphaned value: %v", updated.Metadata)
	}
	// And restating it is not reported as an unknown key, because it is not one.
	for _, u := range updated.MetaReport.UnknownKeys {
		if u.Key == "owner" {
			t.Fatal("a value the document already carried was reported as unknown")
		}
	}

	// A key that was never declared and never stored is still refused.
	meta2 := map[string]any{"stage": "draft", "invented": "x"}
	again, err := k.entry.Update(k.ctx, entry.ID, UpdateEntryRequest{Metadata: &meta2})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if _, ok := again.Metadata["invented"]; ok {
		t.Fatal("carrying orphans through opened the vocabulary")
	}
}

func TestMetadata_CompletingWhileFillingTheFieldInTheSameCall(t *testing.T) {
	k := newMetaKit(t)
	k.declare(t, domain.FieldSpec{Key: "owner", Label: "Owner", Type: domain.FieldText, Required: true})
	entry := k.write(t, nil)

	// "Finish this document and fill in its fields" is the natural call, and it
	// must not be refused for an incompleteness the same request fixes.
	completed := domain.EntryCompleted
	meta := map[string]any{"owner": "platform"}
	updated, err := k.entry.Update(k.ctx, entry.ID, UpdateEntryRequest{
		Status: &completed, Metadata: &meta,
	})
	if err != nil {
		t.Fatalf("completing while filling the field was refused: %v", err)
	}
	if updated.Status != domain.EntryCompleted {
		t.Fatalf("status = %q, want completed", updated.Status)
	}
}

func TestMetadata_RestorePutsTheValuesBack(t *testing.T) {
	k := newMetaKit(t)
	k.declare(t, specStage())
	entry := k.write(t, map[string]any{"stage": "agreed"})

	changed := map[string]any{"stage": "draft"}
	if _, err := k.entry.Update(k.ctx, entry.ID, UpdateEntryRequest{Metadata: &changed}); err != nil {
		t.Fatalf("update: %v", err)
	}

	restored, err := k.entry.Restore(k.ctx, entry.ID, 1)
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if restored.Metadata["stage"] != "agreed" {
		t.Fatalf("restore left the values alone: %v", restored.Metadata)
	}

	// And a revision written before a required field existed stays restorable.
	k.declare(t, specStage(), domain.FieldSpec{Key: "owner", Type: domain.FieldText, Required: true})
	if _, err := k.entry.Restore(k.ctx, entry.ID, 1); err != nil {
		t.Fatalf("a rule written later made an old revision unrestorable: %v", err)
	}
}
