package service

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/dovod-app/app/internal/domain"
)

// Check process records before creating anything. Portable import is not a
// mechanism for applying current editorial caps to historical data, but it
// must reject ambiguous identities and broken session links rather than lose
// provenance or leave a half-imported research on a predictable input error.
func validateImportProcess(r domain.ExportResearch) error {
	sessions := map[string]bool{}
	for _, session := range r.Sessions {
		if session.Code != "" && sessions[session.Code] {
			return fmt.Errorf("duplicate session code %q: %w", session.Code, ErrValidation)
		}
		sessions[session.Code] = true
	}
	for _, item := range r.Memory {
		switch item.Author {
		case "", "unknown", "agent", "user":
		default:
			return fmt.Errorf("unknown memory author %q: %w", item.Author, ErrValidation)
		}
		if item.SessionCode != "" && !sessions[item.SessionCode] {
			return fmt.Errorf("memory names missing session %q: %w", item.SessionCode, ErrValidation)
		}
	}
	slugs := map[string]bool{}
	for _, skill := range r.PrivateSkills {
		if strings.TrimSpace(skill.Slug) == "" || utf8.RuneCountInString(skill.Slug) > 191 || strings.TrimSpace(skill.Name) == "" {
			return fmt.Errorf("private skill requires name and slug (max 191 characters): %w", ErrValidation)
		}
		if slugs[skill.Slug] {
			return fmt.Errorf("duplicate private skill slug %q: %w", skill.Slug, ErrValidation)
		}
		slugs[skill.Slug] = true
	}
	return nil
}
