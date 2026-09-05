package service

import (
	"context"

	"github.com/dovod-app/app/internal/domain"
	"github.com/dovod-app/app/internal/storage"
)

// Telling a writer what its write did to marks it never mentioned.
//
// This is the same contract as BlockSaveReport, and it exists for the same
// reason: the loss is otherwise silent. An agent asked to tighten a paragraph
// has no idea that the sentence it dropped was the one a person had flagged as
// unverified, and finding out weeks later — from a queue full of orphans nobody
// can reconstruct — is finding out too late to act.
//
// Both halves run OUTSIDE the write's transaction. The pool is a single
// connection (MaxOpenConns(1)), so a query issued from inside one deadlocks
// against it; the same constraint that made revisionNote resolve its session up
// front applies here.

// SetAnnotations enables anchor reporting on entry writes.
//
// Optional on purpose: a server assembled without it behaves exactly as it did
// before this feature existed, which is what keeps the wiring in main.go honest
// rather than making every test construct an annotation repository.
func (s *EntryService) SetAnnotations(annotations *storage.AnnotationRepository) {
	s.annotations = annotations
}

// captureAnchors records where marks sit BEFORE a write. Called with the entry
// as it was read, before any field on it has been touched.
func (s *EntryService) captureAnchors(ctx context.Context, entry *domain.Entry) map[string]domain.AnchorState {
	anns := s.loadAnchored(ctx, entry)
	if len(anns) == 0 {
		return nil
	}
	return AnchorStates(anns)
}

// anchorReport places the same marks against the document as it now stands and
// reports only what this write changed.
func (s *EntryService) anchorReport(ctx context.Context, entry *domain.Entry, before map[string]domain.AnchorState) *domain.AnnotationReport {
	if before == nil && s.annotations == nil {
		return nil
	}
	anns := s.loadAnchored(ctx, entry)
	if len(anns) == 0 {
		return nil
	}
	report := AnnotationReportFor(before, anns)
	if report.Empty() {
		return nil
	}
	return &report
}

// loadAnchored reads this entry's marks and places them against the document as
// it stands at the moment of the call.
//
// It re-reads the entry rather than trusting the one in hand: the caller's copy
// is mid-write, and for a blocks entry its Content is the document the author
// sent, not the one the server assembled from rows. Placing marks against the
// former would report drift on state the server was about to fix.
func (s *EntryService) loadAnchored(ctx context.Context, entry *domain.Entry) []*domain.Annotation {
	if s.annotations == nil || entry == nil {
		return nil
	}
	anns, err := s.annotations.FindByEntry(ctx, entry.ID)
	if err != nil || len(anns) == 0 {
		return nil
	}

	stored, err := s.entries.FindByID(ctx, entry.ID)
	if err != nil || stored == nil {
		stored = entry
	}

	var doc *domain.BlockDocument
	var markdown string
	if stored.Type == domain.EntryBlocks {
		if d, derr := s.LoadBlockDocument(ctx, stored); derr == nil {
			doc = d
		}
	} else {
		markdown = stored.Content
	}

	ResolveAnchors(anns, doc, markdown)
	return anns
}
