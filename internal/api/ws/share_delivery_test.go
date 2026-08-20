package ws

import (
	"testing"

	"github.com/butschster/mcp-research/internal/auth"
	"github.com/butschster/mcp-research/internal/domain"
)

// A share connection is the one reader on the hub with no account. Every rule
// in visible() is written in terms of a user id, so the failure mode to guard
// against is not "a share sees too little" — it is a share falling through into
// a rule that was never meant for it.

func (h *Hub) attachShare(researchID string, include domain.ShareInclude, buf int) *Client {
	c := &Client{
		hub:   h,
		send:  make(chan []byte, buf),
		share: &auth.Share{ID: "share-1", ResearchID: researchID, Include: include},
	}
	h.Register(c)
	return c
}

func TestShareDelivery_OnlyItsOwnResearch(t *testing.T) {
	hub := quietHub()
	hub.SetAuthorizer(fakeAuth{}, true)

	visitor := hub.attachShare("r1", domain.ShareInclude{Roadmaps: true}, 8)

	hub.deliver(Event{Type: "entry.updated", Entity: "entry", EntityID: "e1", ResearchID: "r1"})
	if _, ok := received(t, visitor); !ok {
		t.Error("a share visitor did not get an update to the research they are reading")
	}

	hub.deliver(Event{Type: "entry.updated", Entity: "entry", EntityID: "e9", ResearchID: "r2"})
	if _, ok := received(t, visitor); ok {
		t.Error("a share visitor got an event for a research outside the link")
	}
}

// With auth off every ordinary connection sees everything, which is what that
// mode has always done. A share must not inherit it: the whole point is a
// stranger on a public URL.
func TestShareDelivery_AuthDisabledDoesNotOpenTheStream(t *testing.T) {
	hub := quietHub()
	hub.SetAuthorizer(fakeAuth{}, false)

	visitor := hub.attachShare("r1", domain.ShareInclude{}, 8)
	local := hub.attach("", 8)

	hub.deliver(Event{Type: "entry.updated", Entity: "entry", EntityID: "e9", ResearchID: "r2"})

	if _, ok := received(t, visitor); ok {
		t.Error("with auth off, a share visitor received another research's event")
	}
	if _, ok := received(t, local); !ok {
		t.Error("the local single-user connection stopped seeing its own events")
	}
}

func TestShareDelivery_IncludeFlagsFilterTheStream(t *testing.T) {
	hub := quietHub()
	hub.SetAuthorizer(fakeAuth{}, true)

	// Content only: the link excludes sessions, tasks and roadmaps.
	visitor := hub.attachShare("r1", domain.ShareInclude{}, 16)

	for _, e := range []Event{
		{Type: "session.updated", Entity: "session", EntityID: "s1", ResearchID: "r1"},
		{Type: "question.answered", Entity: "question", EntityID: "q1", ResearchID: "r1"},
		{Type: "task.updated", Entity: "task", EntityID: "t1", ResearchID: "r1"},
		{Type: "roadmap.updated", Entity: "roadmap", EntityID: "rm1", ResearchID: "r1"},
		{Type: "share.created", Entity: "share", EntityID: "sh1", ResearchID: "r1"},
		{Type: "team.updated", Entity: "team", EntityID: "team1"},
		// A mark is one person saying they do not believe a sentence. There is
		// no include flag for it and never will be — and `parent_code` names the
		// document being disputed, so the default `return true` at the bottom of
		// visibleToShare was handing a stranger the shape of the argument.
		{Type: "annotation.created", Entity: "annotation", EntityID: "a1", ResearchID: "r1", ParentID: "e1", ParentCode: "E7"},
		{Type: "annotation.answered", Entity: "annotation", EntityID: "a1", ResearchID: "r1", ParentID: "e1", ParentCode: "E7"},
		{Type: "annotation.deleted", Entity: "annotation", EntityID: "a1", ResearchID: "r1", ParentID: "e1", ParentCode: "E7"},
	} {
		hub.deliver(e)
		if got, ok := received(t, visitor); ok {
			t.Errorf("a link excluding %s delivered %s: %+v", e.Entity, e.Type, got)
		}
	}

	// And the content it does include still arrives.
	hub.deliver(Event{Type: "entry.created", Entity: "entry", EntityID: "e1", ResearchID: "r1"})
	if _, ok := received(t, visitor); !ok {
		t.Error("the flags filtered out the content the link exists to show")
	}
}

func TestShareDelivery_NoDirectedEventsAndNoActorIdentity(t *testing.T) {
	hub := quietHub()
	hub.SetAuthorizer(fakeAuth{}, true)

	visitor := hub.attachShare("r1", domain.ShareInclude{}, 8)
	member := hub.attach("alice", 8)

	// A directed message names a user. A visitor is not one, and the access
	// revocation it carries is not their business.
	hub.deliver(Event{Type: "access.revoked", Entity: "research", ResearchID: "r1", TargetUserID: "alice"})
	if _, ok := received(t, visitor); ok {
		t.Error("a directed event reached a share visitor")
	}
	drain(member)

	hub.deliver(Event{
		Type: "entry.updated", Entity: "entry", EntityID: "e1", ResearchID: "r1",
		ActorUserID: "alice-user-id", ActorClientID: "tab-7",
	})
	got, ok := received(t, visitor)
	if !ok {
		t.Fatal("the visitor got no event at all")
	}
	if got.ActorUserID != "" || got.ActorClientID != "" {
		t.Errorf("the event named an account and a tab inside the owner's organisation: %+v", got)
	}
}

func drain(c *Client) {
	for {
		select {
		case <-c.send:
		default:
			return
		}
	}
}
