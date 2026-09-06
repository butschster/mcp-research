package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/dovod-app/app/internal/auth"
	"github.com/dovod-app/app/internal/domain"
)

// ResolveSlug is the read the MCP write tools are built on: they hold a slug
// and the service methods take an id, so this is the step between. Anything it
// hands back is about to be edited or deleted, which makes it worth the same
// scrutiny as Load rather than less.

func TestSkillResolveSlug_RefusesAShareVisitor(t *testing.T) {
	k := newSkillKit(t)
	owner, _, research, _, _ := k.sharedResearch(t, domain.TeamEditor)

	sk, err := k.skills.CreatePrivate(owner, research.ID, skillInput("House rule"))
	if err != nil {
		t.Fatalf("create private: %v", err)
	}

	// A share resolves to viewer on exactly this research, which is right for
	// findings and wrong for methodology — the same reason Load refuses one.
	// Without this the tools would resolve a slug for a stranger and then hand
	// the id to Update, where the team check is the only thing left standing.
	visitor := auth.WithShare(context.Background(), &auth.Share{
		ID: "share-1", ResearchID: research.ID,
	})
	if _, err := k.skills.ResolveSlug(visitor, research.ID, sk.Slug); !errors.Is(err, ErrNotFound) {
		t.Errorf("share visitor: want ErrNotFound, got %v", err)
	}
}

func TestSkillResolveSlug_DoesNotCrossResearches(t *testing.T) {
	k := newSkillKit(t)
	owner, _, mine, _, teamID := k.sharedResearch(t, domain.TeamEditor)

	theirs, _, err := k.research.Create(owner, CreateResearchRequest{
		TeamID: teamID, Name: "Second research", Goal: "Other",
	})
	if err != nil {
		t.Fatalf("create second research: %v", err)
	}
	secret, err := k.skills.CreatePrivate(owner, mine.ID, skillInput("House rule"))
	if err != nil {
		t.Fatalf("create private: %v", err)
	}

	if _, err := k.skills.ResolveSlug(owner, theirs.ID, secret.Slug); !errors.Is(err, ErrNotFound) {
		t.Errorf("cross-research resolve: want ErrNotFound, got %v", err)
	}
}

// A viewer may resolve — it is a read — and must still be refused by the write
// that follows. Testing the pair together is the point: resolving is the step
// that could have been mistaken for the permission check.
func TestSkillResolveSlug_ViewerResolvesAndStillCannotWrite(t *testing.T) {
	k := newSkillKit(t)
	owner, viewer, research, _, _ := k.sharedResearch(t, domain.TeamViewer)

	sk, err := k.skills.CreatePrivate(owner, research.ID, skillInput("House rule"))
	if err != nil {
		t.Fatalf("create private: %v", err)
	}

	got, err := k.skills.ResolveSlug(viewer, research.ID, sk.Slug)
	if err != nil {
		t.Fatalf("viewer resolve: %v", err)
	}
	if got.ID != sk.ID {
		t.Fatalf("resolved %q, want %q", got.ID, sk.ID)
	}
	if _, err := k.skills.Update(viewer, got.ID, skillInput("Rewritten")); !errors.Is(err, ErrForbidden) {
		t.Errorf("viewer update: want ErrForbidden, got %v", err)
	}
	if err := k.skills.Delete(viewer, got.ID); !errors.Is(err, ErrForbidden) {
		t.Errorf("viewer delete: want ErrForbidden, got %v", err)
	}
}

// The slug resolves to what the research follows before what it merely could
// attach — the rule Load already obeys, restated here because a write landing
// on the wrong one of two rows sharing a slug edits a skill nobody was looking
// at.
func TestSkillResolveSlug_PrefersTheAttachedRow(t *testing.T) {
	k := newSkillKit(t)
	owner, _, research, _, _ := k.sharedResearch(t, domain.TeamEditor)

	builtins, err := k.skills.LoadBuiltinSkills(owner)
	if err != nil || builtins == 0 {
		t.Fatalf("load builtins: %d, %v", builtins, err)
	}
	forked, err := k.skills.Fork(owner, research.ID, "evidence-grading", SkillInput{
		Name:        "Grading evidence, our way",
		Description: "Use when weighing a source we are going to cite.",
		Body:        "Our house rules for grading.",
	})
	if err != nil {
		t.Fatalf("fork: %v", err)
	}

	got, err := k.skills.ResolveSlug(owner, research.ID, "evidence-grading")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if got.ID != forked.ID {
		t.Errorf("resolved the built-in, not the fork this research follows")
	}
	if got.Tier != domain.SkillTeam {
		t.Errorf("tier = %q, want team", got.Tier)
	}
}

// The codes are what a client branches on, and both transports now read them
// from here. A message edit must not be able to change what either of them
// does.
func TestSkillErrorCode_NamesTheRefusalsClientsBranchOn(t *testing.T) {
	cases := []struct {
		err  error
		want string
	}{
		{ErrSkillCapReached, "skill_cap_reached"},
		{ErrSkillAlreadyAttached, "already_attached"},
		{ErrSkillSlugTaken, "slug_taken"},
		{ErrSkillInUse, "skill_in_use"},
		{ErrSkillAmbient, "not_allowed"},
		{ErrSkillBuiltinWrite, "not_allowed"},
		{ErrSkillNotPrivate, "not_allowed"},
		{ErrNotFound, ""},
		{ErrForbidden, ""},
		{nil, ""},
	}
	for _, c := range cases {
		if got := SkillErrorCode(c.err); got != c.want {
			t.Errorf("SkillErrorCode(%v) = %q, want %q", c.err, got, c.want)
		}
	}
	// Delete wraps the in-use error with the count, so errors.Is has to be what
	// the switch uses rather than equality.
	if got := SkillErrorCode(errors.Join(ErrSkillInUse, errors.New("3"))); got != "skill_in_use" {
		t.Errorf("wrapped in-use = %q, want skill_in_use", got)
	}
}

// A migrated instruction carries a description a migration invented, flagged so
// the UI can ask somebody to rewrite it. Update clears that flag because a
// caller sending a description has looked at it — which stopped being true when
// the MCP tools started sending back the stored one to make a partial write.
func TestSkillUpdate_ABodyOnlyEditDoesNotClaimTheTriggerWasReviewed(t *testing.T) {
	k := newSkillKit(t)
	owner, _, research, _, _ := k.sharedResearch(t, domain.TeamEditor)

	sk, err := k.skills.CreatePrivate(owner, research.ID, skillInput("Legacy instruction"))
	if err != nil {
		t.Fatalf("create private: %v", err)
	}
	// What the instruction migration produces.
	sk.NeedsTrigger = true
	if err := k.repo.Update(owner, sk); err != nil {
		t.Fatalf("set needs_trigger: %v", err)
	}

	// The partial write the MCP tool makes: a new body, the stored description
	// handed back verbatim.
	after, err := k.skills.Update(owner, sk.ID, SkillInput{
		Name:                 sk.Name,
		Description:          sk.Description,
		Body:                 "A better body, written by an agent that never read the trigger line.",
		DescriptionUntouched: true,
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if !after.NeedsTrigger {
		t.Error("a body-only edit cleared needs_trigger — the prompt to rewrite a generated trigger line is gone, and the generated line is still there")
	}

	// Sending one clears it, which is the whole point of the flag.
	reviewed, err := k.skills.Update(owner, sk.ID, SkillInput{
		Name:        sk.Name,
		Description: "Use when writing anything into this research.",
		Body:        after.Body,
	})
	if err != nil {
		t.Fatalf("update with description: %v", err)
	}
	if reviewed.NeedsTrigger {
		t.Error("a written trigger line did not clear needs_trigger")
	}
}

// Fork and CopyHere write the copy before attaching it, and the attach can be
// refused by the cap. The copy used to be left behind — holding the source's
// slug, attached to nothing, and shadowing the built-in for every later lookup,
// which made one refused call poison that slug for the life of the research.
func TestSkill_ARefusedCopyLeavesNothingBehind(t *testing.T) {
	k := newSkillKit(t)
	owner, _, research, _, _ := k.sharedResearch(t, domain.TeamEditor)
	if _, err := k.skills.LoadBuiltinSkills(owner); err != nil {
		t.Fatalf("load builtins: %v", err)
	}

	for i := 0; i < domain.SkillsPerResearch; i++ {
		if _, err := k.skills.CreatePrivate(owner, research.ID, skillInput(fmt.Sprintf("Filler %d", i))); err != nil {
			t.Fatalf("filler %d: %v", i, err)
		}
	}

	for _, tc := range []struct {
		name string
		call func() error
	}{
		{"copy", func() error {
			_, err := k.skills.CopyHere(owner, research.ID, "evidence-grading")
			return err
		}},
		{"fork", func() error {
			_, err := k.skills.Fork(owner, research.ID, "evidence-grading", SkillInput{
				Name: "Ours", Description: "Use when weighing a source we will cite.", Body: "x",
			})
			return err
		}},
	} {
		t.Run(tc.name+" refused at the cap", func(t *testing.T) {
			if err := tc.call(); !errors.Is(err, ErrSkillCapReached) {
				t.Fatalf("want ErrSkillCapReached, got %v", err)
			}
			// The slug must still mean the built-in. A copy left behind would
			// resolve first, and the agent would read a body it never wrote.
			sk, err := k.skills.Load(owner, research.ID, "evidence-grading")
			if err != nil {
				t.Fatalf("load after refusal: %v", err)
			}
			if sk.Tier != domain.SkillBuiltin {
				t.Errorf("tier = %q, want builtin — a refused %s left its copy behind", sk.Tier, tc.name)
			}
		})
	}

	// And the operation must still be possible once there is room, which the
	// orphan made permanently false: CopyHere short-circuits on an already
	// private source and would have answered "nothing to do" forever.
	if err := k.skills.Detach(owner, research.ID, "filler-0"); err != nil {
		t.Fatalf("detach: %v", err)
	}
	copied, err := k.skills.CopyHere(owner, research.ID, "evidence-grading")
	if err != nil {
		t.Fatalf("copy after freeing a slot: %v", err)
	}
	if copied.Tier != domain.SkillPrivate {
		t.Errorf("tier = %q, want private", copied.Tier)
	}
	attached, err := k.skills.ListAttached(owner, research.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	var found bool
	for _, sk := range attached {
		if sk.ID == copied.ID {
			found = true
		}
	}
	if !found {
		t.Error("the copy was made but never attached")
	}
}
