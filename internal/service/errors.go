package service

import "errors"

var (
	ErrNotFound             = errors.New("not found")
	ErrDuplicateSectionName = errors.New("section name already exists in this research")
	ErrSectionHasNoEntries  = errors.New("cannot complete a section with no entries")
	ErrTextReplaceNotFound  = errors.New("text_replace: from string not found in content")
	ErrQuestionDepthLimit   = errors.New("question nesting depth limit exceeded (max 3 levels)")
	ErrAnswerRequired       = errors.New("answered questions must have a non-empty answer")
	ErrMutualExclusion      = errors.New("mutually exclusive fields provided")
)
