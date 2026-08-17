package service

import (
	"context"
	"embed"
	"fmt"
	"strings"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/google/uuid"
)

//go:embed skills_data/*.md
var builtinSkillsFS embed.FS

// LoadBuiltinSkills upserts the skills we ship into the database at boot.
//
// Two rules make this safe to run on every start:
//
//   - It only ever touches rows with no team and no research. A team that
//     forked a built-in owns a separate row with the same slug, and an upgrade
//     must never overwrite what somebody edited — which is the whole reason
//     editing a built-in forks instead of writing through.
//   - It matches on slug within that tier, so the second boot updates the row
//     the first one wrote instead of inserting a duplicate. The three partial
//     unique indexes in migration 021 exist because a plain UNIQUE would not
//     have caught it: SQLite treats NULLs as distinct.
//
// A built-in that disappears from the binary is left in place rather than
// deleted: a research may still be following it, and taking a methodology away
// mid-research is worse than carrying a stale row.
func (s *SkillService) LoadBuiltinSkills(ctx context.Context) (int, error) {
	entries, err := builtinSkillsFS.ReadDir("skills_data")
	if err != nil {
		return 0, fmt.Errorf("read builtin skills: %w", err)
	}

	var n int
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		raw, err := builtinSkillsFS.ReadFile("skills_data/" + e.Name())
		if err != nil {
			return n, fmt.Errorf("read %s: %w", e.Name(), err)
		}
		sk, err := parseBuiltinSkill(string(raw))
		if err != nil {
			return n, fmt.Errorf("%s: %w", e.Name(), err)
		}
		// The budgets are enforced on what we ship, not only on what users
		// write. A built-in over the description cap would silently eat the
		// index budget it exists to protect.
		if err := validateSkill(SkillInput{Name: sk.Name, Description: sk.Description, Body: sk.Body}); err != nil {
			return n, fmt.Errorf("%s: %w", e.Name(), err)
		}

		existing, err := s.skills.FindBuiltinBySlug(ctx, sk.Slug)
		if err != nil {
			return n, err
		}
		if existing == nil {
			sk.ID = uuid.New().String()
			if err := s.skills.Create(ctx, sk); err != nil {
				return n, fmt.Errorf("create builtin %s: %w", sk.Slug, err)
			}
			n++
			continue
		}
		existing.Name = sk.Name
		existing.Description = sk.Description
		existing.Body = sk.Body
		if err := s.skills.Update(ctx, existing); err != nil {
			return n, fmt.Errorf("update builtin %s: %w", sk.Slug, err)
		}
		// The ambient flag is not part of Update, because it is structural
		// rather than editorial. Setting it here keeps a shipped change to it
		// effective without widening the update path users also travel.
		if err := s.skills.SetAmbient(ctx, existing.ID, sk.Ambient); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

// parseBuiltinSkill reads the frontmatter shape this team already writes in
// .claude/skills/*/SKILL.md: a `---` fenced block of `key: value` lines
// followed by the body.
//
// It is deliberately not a YAML parser. The fields are four scalars, the files
// are ours, and a dependency that can interpret a colon in a description as
// structure is a worse trade than a strict reader that refuses what it does not
// understand.
func parseBuiltinSkill(raw string) (*domain.Skill, error) {
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

	sk := &domain.Skill{Tier: domain.SkillBuiltin, Version: 1, Body: body}
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
			sk.Slug = value
		case "name":
			sk.Name = value
		case "description":
			sk.Description = value
		case "ambient":
			sk.Ambient = value == "true"
		default:
			return nil, fmt.Errorf("unknown frontmatter key %q", key)
		}
	}
	if sk.Slug == "" {
		return nil, fmt.Errorf("frontmatter has no slug")
	}
	return sk, nil
}
