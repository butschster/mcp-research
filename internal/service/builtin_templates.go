package service

import (
	"context"
	"embed"
	"fmt"
	"strings"

	"github.com/dovod-app/app/internal/domain"
	"github.com/google/uuid"
)

//go:embed templates_data/*.md
var builtinTemplatesFS embed.FS

// LoadBuiltinTemplates upserts the methodologies we ship, at every boot.
//
// The same two rules as the built-in skills, for the same reasons:
//
//   - only rows with no team are touched, so a team that forked one keeps its
//     copy through every upgrade;
//   - matching is by slug within that tier, so the second boot updates rather
//     than inserting a duplicate — which a plain UNIQUE would not have caught,
//     since SQLite treats NULLs as distinct.
//
// A template naming a skill that does not exist is refused — a broken
// methodology is worth finding at startup rather than from somebody's kickoff.
// But one bad file only removes itself: the rest load, and every problem is
// reported together, because a server running with no methodologies at all is
// the quietest possible failure.
func (s *TemplateService) LoadBuiltinTemplates(ctx context.Context) (int, error) {
	entries, err := builtinTemplatesFS.ReadDir("templates_data")
	if err != nil {
		return 0, fmt.Errorf("read builtin templates: %w", err)
	}

	// Errors are collected rather than returned on the first one. Returning
	// early meant a mistyped skill slug in the alphabetically-first file left a
	// server running with *no* methodologies at all, `template_list` answering
	// an empty array, and one ERROR line in a log nobody reads. Load what is
	// good, and report everything that is not.
	var problems []string
	var n int
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		raw, err := builtinTemplatesFS.ReadFile("templates_data/" + e.Name())
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s: read: %v", e.Name(), err))
			continue
		}
		tp, err := parseBuiltinTemplate(string(raw))
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s: %v", e.Name(), err))
			continue
		}
		if err := s.validate(ctx, TemplateInput{
			Name: tp.Name, Description: tp.Description, WhenToUse: tp.WhenToUse,
			WhenNotToUse: tp.WhenNotToUse, Body: tp.Body, Skills: tp.Skills,
		}); err != nil {
			problems = append(problems, fmt.Sprintf("%s: %v", e.Name(), err))
			continue
		}
		var unknown string
		for _, slug := range tp.Skills {
			ok, err := s.skills.BuiltinSkillExists(ctx, slug)
			if err != nil {
				return n, err
			}
			if !ok {
				unknown = slug
				break
			}
		}
		if unknown != "" {
			problems = append(problems, fmt.Sprintf("%s: %v: %q", e.Name(), ErrTemplateUnknownSkill, unknown))
			continue
		}

		// Only a row this loader wrote. A global template the operator added
		// through the API lives in the same tier under the same unique index,
		// and matching on the tier alone would have overwritten their text with
		// ours the first time a shipped file took the same slug. Now the insert
		// collides instead, and the clash is reported rather than resolved
		// silently in our favour.
		existing, err := s.templates.FindBuiltinBySlug(ctx, tp.Slug)
		if err != nil {
			return n, err
		}
		if existing == nil {
			tp.ID = uuid.New().String()
			if err := s.templates.Create(ctx, tp); err != nil {
				problems = append(problems, fmt.Sprintf("%s: create: %v", e.Name(), err))
				continue
			}
			n++
			continue
		}
		// Only write when something actually changed. Update bumps the version,
		// and a research stamps the version it followed — so an unconditional
		// refresh made every restart look like a rewrite of the methodology.
		// Three boots of an unchanged binary took `literature-review` to v3 and
		// left every research created before them reading as out of date.
		existingBody, err := s.templates.Body(ctx, existing.ID)
		if err != nil {
			return n, err
		}
		if sameTemplate(existing, existingBody, tp) {
			continue
		}
		existing.Name = tp.Name
		existing.Description = tp.Description
		existing.WhenToUse = tp.WhenToUse
		existing.WhenNotToUse = tp.WhenNotToUse
		existing.Body = tp.Body
		existing.Skills = tp.Skills
		if err := s.templates.Update(ctx, existing); err != nil {
			return n, fmt.Errorf("update builtin template %s: %w", tp.Slug, err)
		}
		n++
	}
	if len(problems) > 0 {
		// "n written", not "n loaded". They are different numbers and the
		// difference is the whole message: an unchanged boot writes nothing and
		// reported "(0 loaded)", which reads at three in the morning as a server
		// with no methodologies at all when in fact every one of them is present
		// and healthy.
		return n, fmt.Errorf("%w (%d of %d written, the rest were already current): %s",
			ErrTemplateBuiltinLoad, n, len(entries)-len(problems), strings.Join(problems, "; "))
	}
	return n, nil
}

// sameTemplate is what stops an unchanged boot from bumping the version. It
// compares everything a reader of the methodology would notice; `updated_at` and
// the version itself are deliberately not part of it.
func sameTemplate(existing *domain.Template, existingBody string, next *domain.Template) bool {
	if existing.Name != next.Name || existing.Description != next.Description ||
		existing.WhenToUse != next.WhenToUse || existing.WhenNotToUse != next.WhenNotToUse ||
		existingBody != next.Body || len(existing.Skills) != len(next.Skills) {
		return false
	}
	for i := range existing.Skills {
		if existing.Skills[i] != next.Skills[i] {
			return false
		}
	}
	return true
}

// parseBuiltinTemplate reads the same frontmatter shape the built-in skills use.
// Deliberately not a YAML parser: the fields are five scalars plus a flat list,
// the files are ours, and a dependency that can read a colon in a sentence as
// structure is a worse trade than a strict reader that refuses what it does not
// understand.
func parseBuiltinTemplate(raw string) (*domain.Template, error) {
	text := strings.ReplaceAll(raw, "\r\n", "\n")
	if !strings.HasPrefix(text, "---\n") {
		return nil, fmt.Errorf("missing frontmatter")
	}
	end := strings.Index(text[4:], "\n---\n")
	if end < 0 {
		return nil, fmt.Errorf("unterminated frontmatter")
	}
	head := text[4 : 4+end]
	body := strings.TrimSpace(text[4+end+5:])

	tp := &domain.Template{
		Tier:    domain.TemplateGlobal,
		Source:  domain.TemplateSourceBuiltin,
		Version: 1,
		Body:    body,
		Skills:  []string{},
	}
	for _, line := range strings.Split(head, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			return nil, fmt.Errorf("frontmatter line is not key: value: %q", line)
		}
		value = strings.TrimSpace(value)
		switch strings.TrimSpace(key) {
		case "slug":
			tp.Slug = value
		case "name":
			tp.Name = value
		case "description":
			tp.Description = value
		case "when_to_use":
			tp.WhenToUse = value
		case "when_not_to_use":
			tp.WhenNotToUse = value
		case "skills":
			tp.Skills = parseSlugList(value)
		default:
			return nil, fmt.Errorf("unknown frontmatter key %q", key)
		}
	}
	if tp.Slug == "" {
		return nil, fmt.Errorf("frontmatter has no slug")
	}
	return tp, nil
}

// parseSlugList reads `[a, b]` or `a, b`, which is as much list syntax as these
// files need.
func parseSlugList(v string) []string {
	v = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(v, "["), "]"))
	if v == "" {
		return []string{}
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
