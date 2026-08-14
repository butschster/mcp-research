package service

import (
	"context"
	"errors"
	"unicode"

	"github.com/butschster/mcp-research/internal/auth"
	"github.com/butschster/mcp-research/internal/storage"
)

var (
	ErrNotFound             = errors.New("not found")
	ErrDuplicateSectionName = errors.New("section name already exists in this research")
	ErrSectionHasNoEntries  = errors.New("cannot complete a section with no entries")
	ErrTextReplaceNotFound  = errors.New("text_replace: from string not found in content")
	ErrTextReplaceOnBlocks  = errors.New("text_replace does not work on a blocks entry: it would edit the stored JSON as text. Send the whole document in content instead")
	ErrQuestionDepthLimit   = errors.New("question nesting depth limit exceeded (max 3 levels)")
	ErrAnswerRequired       = errors.New("answered questions must have a non-empty answer")
	ErrMutualExclusion      = errors.New("mutually exclusive fields provided")
)

// validateResearchAccess checks that the research exists and the current user owns it.
func validateResearchAccess(ctx context.Context, repo *storage.ResearchRepository, researchID string) error {
	research, err := repo.FindByID(ctx, researchID)
	if err != nil {
		return err
	}
	if research == nil {
		return ErrNotFound
	}
	if uid := auth.UserIDFromContext(ctx); uid != "" && research.UserID != "" && research.UserID != uid {
		return ErrNotFound
	}
	return nil
}

// isCode returns true if s looks like a short code (e.g. R1, E23, SS1, T5, Q3) rather than a UUID.
func isCode(s string) bool {
	if len(s) < 2 || len(s) > 10 {
		return false
	}
	// Determine prefix length: SS, RM have 2-char prefix, others have 1-char
	prefixLen := 0
	switch {
	case len(s) >= 2 && (s[:2] == "SS" || s[:2] == "RM"):
		prefixLen = 2
	case s[0] == 'R' || s[0] == 'E' || s[0] == 'S' || s[0] == 'T' || s[0] == 'Q' || s[0] == 'N':
		prefixLen = 1
	default:
		return false
	}
	for _, c := range s[prefixLen:] {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}
