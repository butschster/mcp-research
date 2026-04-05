package service

import (
	"errors"
	"unicode"
)

var (
	ErrNotFound             = errors.New("not found")
	ErrDuplicateSectionName = errors.New("section name already exists in this research")
	ErrSectionHasNoEntries  = errors.New("cannot complete a section with no entries")
	ErrTextReplaceNotFound  = errors.New("text_replace: from string not found in content")
	ErrQuestionDepthLimit   = errors.New("question nesting depth limit exceeded (max 3 levels)")
	ErrAnswerRequired       = errors.New("answered questions must have a non-empty answer")
	ErrMutualExclusion      = errors.New("mutually exclusive fields provided")
)

// isCode returns true if s looks like a short code (e.g. R1, E23, SS1, T5, Q3) rather than a UUID.
func isCode(s string) bool {
	if len(s) < 2 || len(s) > 10 {
		return false
	}
	// Determine prefix length: SS has 2-char prefix, others have 1-char
	prefixLen := 0
	switch {
	case len(s) >= 2 && s[:2] == "SS":
		prefixLen = 2
	case s[0] == 'R' || s[0] == 'E' || s[0] == 'S' || s[0] == 'T' || s[0] == 'Q':
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
