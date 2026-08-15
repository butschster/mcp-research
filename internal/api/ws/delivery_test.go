package ws

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"
)

// countingAuth answers like fakeAuth but records how often it was asked, which
// is the point of the cache in front of it.
type countingAuth struct {
	fakeAuth
	calls atomic.Int64
}

func (c *countingAuth) CanReadResearch(ctx context.Context, userID, researchID string) bool {
	c.calls.Add(1)
	return c.fakeAuth.CanReadResearch(ctx, userID, researchID)
}

func (h *Hub) attach(userID string, buf int) *Client {
	c := &Client{hub: h, send: make(chan []byte, buf), userID: userID}
	h.Register(c)
	return c
}

func received(t *testing.T, c *Client) (Event, bool) {
	t.Helper()
	select {
	case data := <-c.send:
		var e Event
		if err := json.Unmarshal(data, &e); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		return e, true
	default:
		return Event{}, false
	}
}

// A directed event is the one message that must reach somebody the ordinary
// rule now refuses: "you have lost access". If it went through the research
// check it would be dropped for exactly the person it is written for.
func TestDeliver_DirectedEventReachesOnlyItsTarget(t *testing.T) {
	hub := quietHub()
	// Deliberately an authorizer that grants nobody anything.
	hub.SetAuthorizer(fakeAuth{}, true)

	target := hub.attach("alice", 4)
	other := hub.attach("bob", 4)

	hub.deliver(Event{Type: "access.revoked", Entity: "team", EntityID: "t1", TargetUserID: "alice"})

	if _, ok := received(t, target); !ok {
		t.Error("the person who lost access was not told")
	}
	if _, ok := received(t, other); ok {
		t.Error("a directed event reached somebody it was not addressed to")
	}
}

// The target id is a routing decision, not something a recipient should be
// handed: they learn nothing from it that the delivery did not already say.
func TestDeliver_DirectedEventDoesNotShipTheTargetID(t *testing.T) {
	hub := quietHub()
	hub.SetAuthorizer(fakeAuth{}, true)
	c := hub.attach("alice", 4)

	hub.deliver(Event{Type: "access.revoked", Entity: "team", EntityID: "t1", TargetUserID: "alice"})

	data := <-c.send
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if _, present := raw["target_user_id"]; present {
		t.Errorf("target_user_id was serialised: %s", data)
	}
}

// Events name the research by id; every URL names it by short code. Carrying
// both is what stopped six pages from silently dropping everything they got.
func TestDeliver_FillsInTheShortCode(t *testing.T) {
	hub := quietHub()
	hub.SetCodeLookup(codes{"r1": "R7"})
	c := hub.attach("", 4)

	hub.deliver(Event{Type: "entry.updated", ResearchID: "r1", EntityID: "e1", Entity: "entry"})

	got, ok := received(t, c)
	if !ok {
		t.Fatal("nothing delivered")
	}
	if got.ResearchCode != "R7" {
		t.Errorf("research_code = %q, want %q", got.ResearchCode, "R7")
	}
}

type codes map[string]string

func (c codes) Code(_ context.Context, researchID string) string { return c[researchID] }

// One entry update with twenty readers connected used to be twenty serialized
// queries against a database that permits one connection at a time.
func TestDeliver_AsksTheAuthorizerOncePerReaderPerResearch(t *testing.T) {
	auth := &countingAuth{fakeAuth: fakeAuth{research: map[string][]string{"r1": {"alice"}}}}
	hub := quietHub()
	hub.SetAuthorizer(auth, true)
	hub.attach("alice", 16)

	for i := 0; i < 8; i++ {
		hub.deliver(Event{Type: "entry.updated", ResearchID: "r1", EntityID: "e1", Entity: "entry"})
	}
	if n := auth.calls.Load(); n != 1 {
		t.Errorf("authorizer asked %d times, want 1", n)
	}
}

// The cache must never be the reason somebody keeps seeing what they lost, so
// anything that could have changed a membership drops all of it.
func TestDeliver_MembershipChangeForgetsEveryVerdict(t *testing.T) {
	auth := &countingAuth{fakeAuth: fakeAuth{
		research: map[string][]string{"r1": {"alice"}},
		teams:    map[string][]string{"t1": {"alice"}},
	}}
	hub := quietHub()
	hub.SetAuthorizer(auth, true)
	hub.attach("alice", 16)

	entry := Event{Type: "entry.updated", ResearchID: "r1", EntityID: "e1", Entity: "entry"}
	hub.deliver(entry)
	hub.deliver(entry)
	if n := auth.calls.Load(); n != 1 {
		t.Fatalf("setup: authorizer asked %d times, want 1", n)
	}

	// Through Broadcast, because that is where the flush lives — it must not
	// depend on the event surviving the queue.
	hub.Broadcast(Event{Type: "team.member_removed", Entity: "team", EntityID: "t1"})
	hub.deliver(entry)
	if n := auth.calls.Load(); n != 2 {
		t.Errorf("authorizer asked %d times after a membership change, want 2", n)
	}

	hub.Broadcast(Event{Type: "research.transferred", ResearchID: "r1", EntityID: "r1", Entity: "research"})
	hub.deliver(entry)
	if n := auth.calls.Load(); n != 3 {
		t.Errorf("authorizer asked %d times after a transfer, want 3", n)
	}
}

// A reader whose buffer has backed up must not hold up everybody else's.
func TestDeliver_AFullClientIsSkippedNotWaitedFor(t *testing.T) {
	hub := quietHub()
	stuck := hub.attach("", 1)
	healthy := hub.attach("", 4)

	stuck.send <- []byte("already full")

	hub.deliver(Event{Type: "entry.updated", ResearchID: "r1", EntityID: "e1", Entity: "entry"})

	if _, ok := received(t, healthy); !ok {
		t.Error("a healthy client was starved by a full one")
	}
}

// A socket outlived the credential that opened it: the per-event check asks
// whether the user may read, never whether their token is still good.
func TestCredentialStillValid(t *testing.T) {
	hub := quietHub()
	hub.SetAuthorizer(fakeAuth{}, true)
	hub.SetTokenValidator(fakeValidator{valid: map[string]string{"alice-key": "alice"}})

	cases := map[string]struct {
		client *Client
		want   bool
	}{
		"the credential still resolves to them": {&Client{userID: "alice", token: "alice-key"}, true},
		"the key was deleted":                   {&Client{userID: "alice", token: "revoked-key"}, false},
		"it now resolves to somebody else":      {&Client{userID: "bob", token: "alice-key"}, false},
		"there is no credential to check":       {&Client{userID: "alice"}, false},
	}
	for name, tc := range cases {
		if got := hub.credentialStillValid(tc.client); got != tc.want {
			t.Errorf("%s: valid=%v, want %v", name, got, tc.want)
		}
	}

	// With auth off there is nothing to revoke and nothing to check.
	local := quietHub()
	if !local.credentialStillValid(&Client{}) {
		t.Error("local mode disconnected a connection over a credential it never asked for")
	}

	// A lookup that failed is not a revocation. Closing on it would tell
	// somebody still signed in that their session ended — and the client treats
	// that verdict as terminal and stops reconnecting, so the tab never
	// recovers without a reload.
	flaky := quietHub()
	flaky.SetAuthorizer(fakeAuth{}, true)
	flaky.SetTokenValidator(fakeValidator{broken: true})
	if !flaky.credentialStillValid(&Client{userID: "alice", token: "alice-key"}) {
		t.Error("a failed lookup was treated as a revoked credential")
	}
}

// The flush that keeps the verdict cache honest must not travel with the event:
// delivery is best-effort and a full queue drops events, so a dropped
// membership change would leave a cached "may read" standing for a full TTL
// after the access it describes was taken away.
func TestBroadcast_ForgetsVerdictsEvenWhenTheQueueIsFull(t *testing.T) {
	hub := quietHub()
	hub.verdicts.put("alice", "r1", true)

	// Fill the queue without the drain goroutine emptying it faster than we can
	// write, by draining nothing: the events pile up and the rest are dropped.
	for i := 0; i < broadcastQueue*2; i++ {
		hub.Broadcast(Event{Type: "entry.updated", ResearchID: "r1", EntityID: "e1", Entity: "entry"})
	}
	hub.Broadcast(Event{Type: "team.member_removed", Entity: "team", EntityID: "t1"})

	if _, cached := hub.verdicts.get("alice", "r1"); cached {
		t.Error("a membership change left a cached verdict standing")
	}
}
