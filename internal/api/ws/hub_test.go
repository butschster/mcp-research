package ws

import (
	"context"
	"testing"
)

// The hub used to send every event to every socket, on an endpoint that took
// no credential — so an anonymous listener got a live feed of which researches
// exist and when each one is touched. These assert the delivery rule directly,
// because it is the one place a new event type can leak by omission.

type fakeAuth struct {
	research map[string][]string // researchID -> user ids who may read it
	teams    map[string][]string // teamID -> member user ids
}

func (f fakeAuth) CanReadResearch(_ context.Context, userID, researchID string) bool {
	return contains(f.research[researchID], userID)
}

func (f fakeAuth) IsTeamMember(_ context.Context, userID, teamID string) bool {
	return contains(f.teams[teamID], userID)
}

func contains(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}

func TestHub_DeliversOnlyToReadersOfTheResearch(t *testing.T) {
	auth := fakeAuth{research: map[string][]string{"r1": {"alice", "bob"}}}
	event := Event{Type: "entry.created", ResearchID: "r1", EntityID: "e1", Entity: "entry"}

	cases := map[string]struct {
		user string
		want bool
	}{
		"a member":              {"alice", true},
		"another member":        {"bob", true},
		"someone else":          {"mallory", false},
		"an anonymous listener": {"", false},
	}
	for name, tc := range cases {
		if got := visible(t.Context(), auth, true, tc.user, event); got != tc.want {
			t.Errorf("%s: delivered=%v, want %v", name, got, tc.want)
		}
	}
}

func TestHub_TeamEventsFollowMembership(t *testing.T) {
	auth := fakeAuth{teams: map[string][]string{"t1": {"alice"}}}
	event := Event{Type: "team.member_removed", EntityID: "t1", Entity: "team"}

	if !visible(t.Context(), auth, true, "alice", event) {
		t.Error("a member should hear about their own team")
	}
	// The timing of a membership change is organisational information: an
	// outsider learning that a team gained someone at 14:07 is the leak team
	// events introduce.
	if visible(t.Context(), auth, true, "mallory", event) {
		t.Error("an outsider must not hear about a team's membership")
	}
}

// An event carrying no scope cannot be checked against anything, so it goes
// nowhere. "When in doubt, send it" is what made the hub a public feed.
func TestHub_UnscopedEventsGoNowhere(t *testing.T) {
	auth := fakeAuth{research: map[string][]string{"r1": {"alice"}}}
	event := Event{Type: "something.happened", EntityID: "x", Entity: "mystery"}

	if visible(t.Context(), auth, true, "alice", event) {
		t.Error("an event with no research and no team was delivered")
	}
}

// A hub that was never handed an authorizer must go quiet rather than fall
// back to broadcasting — failing closed is the only safe direction here.
func TestHub_WithoutAnAuthorizerSendsNothing(t *testing.T) {
	event := Event{Type: "entry.created", ResearchID: "r1", Entity: "entry"}
	if visible(t.Context(), nil, true, "alice", event) {
		t.Error("delivered with no authorizer configured")
	}
}

// With auth off there are no users to scope by, and this is the behaviour the
// local single-binary mode has always had.
func TestHub_LocalModeDeliversEverything(t *testing.T) {
	event := Event{Type: "entry.created", ResearchID: "r1", Entity: "entry"}
	if !visible(t.Context(), nil, false, "", event) {
		t.Error("local mode should still receive its own events")
	}
}
