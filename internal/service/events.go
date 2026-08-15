package service

import (
	"context"

	"github.com/butschster/mcp-research/internal/auth"
)

// EventNotifier is called after data mutations to notify listeners.
type EventNotifier interface {
	Notify(event Event)
}

type Event struct {
	Type       string // e.g. "research.created", "entry.updated"
	ResearchID string
	EntityID   string
	Entity     string // "research", "section", "entry", "session", "question", "task", "team"

	// ActorUserID is who caused the change. Empty when auth is off, and for
	// anything an agent did over stdio.
	ActorUserID string

	// ActorClientID is which *tab* caused it, and it is the field that actually
	// answers "was this me". The user id cannot: with auth off it is empty, so
	// a browser write and an agent write look identical; and with auth on two
	// tabs of one person share it, so one would suppress a change the other
	// made and must display. A client sends its own id with each write and
	// compares it to what comes back.
	ActorClientID string

	// Reason says why, for the events where the type alone does not — chiefly
	// access.revoked, where the difference between "you were removed from the
	// team" and "the research was moved" is the whole message.
	Reason string

	// Name is the human name of the entity in EntityID, carried only where the
	// recipient cannot look it up: by the time somebody is told they lost
	// access, fetching the thing they lost returns 404.
	Name string

	// TargetUserID addresses the event to one person instead of to everyone who
	// can read the research. It exists for the one case the research scope
	// cannot express: telling somebody they have just lost access. By the time
	// that event is sent, the ordinary check would already refuse to deliver it.
	TargetUserID string
}

// emit stamps the event with who caused it and hands it on.
//
// The actor is read here rather than at each call site because a call site that
// forgets it degrades silently — the client falls back to treating a remote
// change as its own and drops the refetch.
func emit(ctx context.Context, n EventNotifier, event Event) {
	if n == nil {
		return
	}
	if event.ActorUserID == "" {
		event.ActorUserID = auth.UserIDFromContext(ctx)
	}
	if event.ActorClientID == "" {
		event.ActorClientID = ClientIDFromContext(ctx)
	}
	n.Notify(event)
}

type clientIDKey struct{}

// WithClientID records which browser tab is making this request. It is set from
// the X-Client-Id header on writes and is opaque to the server — it exists only
// so the tab can recognise its own change coming back.
func WithClientID(ctx context.Context, id string) context.Context {
	if id == "" {
		return ctx
	}
	if len(id) > 64 {
		id = id[:64]
	}
	return context.WithValue(ctx, clientIDKey{}, id)
}

// ClientIDFromContext returns the calling tab's id, or "" for an agent, a
// script, or a browser too old to send one.
func ClientIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(clientIDKey{}).(string); ok {
		return id
	}
	return ""
}

// NoopNotifier does nothing (default when no WebSocket hub is connected).
type NoopNotifier struct{}

func (NoopNotifier) Notify(Event) {}
