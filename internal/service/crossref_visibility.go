package service

import (
	"context"

	"github.com/butschster/mcp-research/internal/domain"
)

// VisibleCrossRefs returns the references with everything the reader may not
// follow stripped out: the target ids blanked and `resolved` set to false, so
// the `[[…]]` renders as plain text rather than as a link into nothing.
//
// This is where cross-research resolution is decided, and it is decided per
// reader rather than per author.
//
// It used to be decided when the reference was written, by refusing to store a
// target the author could not read. That had it backwards in both directions:
// a colleague who *could* open the target still saw plain text, because the
// author's permissions had been frozen into the row; and an author who lost
// access later kept a working link. Storing the resolution and filtering it on
// the way out fixes both, and closes the id-harvesting the old rule existed to
// prevent — writing [[R1:E1]], [[R1:E2]], … now reads back exactly as
// unresolved as it looked going in.
func (a *Access) VisibleCrossRefs(ctx context.Context, refs []domain.CrossRef) []domain.CrossRef {
	mayRead := a.readMemo(ctx)

	out := make([]domain.CrossRef, 0, len(refs))
	for _, ref := range refs {
		// A reference with no target research is a same-research one, and the
		// reader is already holding the source.
		target := ref.TargetResearchID
		if target == "" {
			target = ref.SourceResearchID
		}
		if !ref.Resolved || mayRead(target) {
			out = append(out, ref)
			continue
		}

		ref.Resolved = false
		ref.TargetEntryID = ""
		ref.TargetResearchID = ""
		ref.TargetRoadmapID = ""
		ref.TargetNodeID = ""
		out = append(out, ref)
	}
	return out
}

// VisibleIncomingCrossRefs drops the references *into* something the reader can
// see that come *from* something they cannot.
//
// Blanking is the wrong treatment here, which is why this is a second function
// rather than a flag on the first. An outgoing reference is text the reader is
// already holding — `[[R2:E5]]` is in the entry they are looking at, and
// rendering it inert is the honest answer. An incoming one is nothing they
// possess: the row exists only because somebody else wrote it, and even the
// stripped version announces that an unseen research cites this entry. For a
// share visitor that is the whole shape of the workspace behind the link.
func (a *Access) VisibleIncomingCrossRefs(ctx context.Context, refs []domain.CrossRef) []domain.CrossRef {
	mayRead := a.readMemo(ctx)

	out := make([]domain.CrossRef, 0, len(refs))
	for _, ref := range refs {
		if ref.SourceResearchID == "" || !mayRead(ref.SourceResearchID) {
			continue
		}
		out = append(out, ref)
	}
	return out
}

// readMemo answers "may this reader open that research" once per research. An
// entry that points at a colleague's research twenty times asks once.
func (a *Access) readMemo(ctx context.Context) func(string) bool {
	readable := make(map[string]bool, 4)
	return func(researchID string) bool {
		if researchID == "" {
			return false
		}
		if seen, ok := readable[researchID]; ok {
			return seen
		}
		ok := a.Read(ctx, researchID) == nil
		readable[researchID] = ok
		return ok
	}
}
