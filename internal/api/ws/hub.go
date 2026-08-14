package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
)

// Event is broadcast to the WebSocket clients allowed to see it.
type Event struct {
	Type       string `json:"type"`        // e.g. "research.updated", "entry.created"
	ResearchID string `json:"research_id"` // scope
	EntityID   string `json:"entity_id"`   // ID of the changed entity
	Entity     string `json:"entity"`      // "research", "section", "entry", "session", "question", "task", "team"
}

// Authorizer answers, for one connected reader, whether an event is theirs to
// see. It is the same question every HTTP read asks, asked again at send time
// rather than at subscribe time — which is what makes a removed member stop
// receiving updates on the connection they already have open.
type Authorizer interface {
	// CanReadResearch reports whether the user may read the research.
	CanReadResearch(ctx context.Context, userID, researchID string) bool
	// IsTeamMember reports whether the user belongs to the team.
	IsTeamMember(ctx context.Context, userID, teamID string) bool
}

// Hub manages WebSocket connections and delivers events to the clients
// entitled to them.
//
// Delivery is decided per client per event. The alternative — deciding once,
// when a client subscribes — is what makes a revoked membership keep receiving
// updates until the tab is closed, and "removing a member revokes access
// immediately" is the property this whole feature rests on.
type Hub struct {
	mu      sync.RWMutex
	clients map[*Client]struct{}
	auth    Authorizer
	// authEnabled mirrors the server's configuration. With auth off there are
	// no users to scope by and everything is one local team, so every client
	// sees everything — which is the behaviour that mode has always had.
	authEnabled bool
	log         *slog.Logger
}

func NewHub(log *slog.Logger) *Hub {
	return &Hub{
		clients: make(map[*Client]struct{}),
		log:     log,
	}
}

// SetAuthorizer turns on per-client scoping. It is called after construction
// because the services the authorizer reads are wired later; until it is, the
// hub refuses to deliver anything to an identified client rather than
// defaulting to delivering everything.
func (h *Hub) SetAuthorizer(auth Authorizer, authEnabled bool) {
	h.mu.Lock()
	h.auth = auth
	h.authEnabled = authEnabled
	h.mu.Unlock()
}

// AuthEnabled reports whether connections must carry a token.
func (h *Hub) AuthEnabled() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.authEnabled
}

func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
	h.log.Debug("ws client connected", "clients", h.count())
}

func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
	h.log.Debug("ws client disconnected", "clients", h.count())
}

// Broadcast sends an event to every client entitled to it.
func (h *Hub) Broadcast(event Event) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	h.mu.RLock()
	auth, authEnabled := h.auth, h.authEnabled
	clients := make([]*Client, 0, len(h.clients))
	for c := range h.clients {
		clients = append(clients, c)
	}
	h.mu.RUnlock()

	// Outside the lock: the check reads the database, and holding the hub
	// lock across it would stall every other connection.
	ctx := context.Background()
	for _, c := range clients {
		if !visible(ctx, auth, authEnabled, c.userID, event) {
			continue
		}
		select {
		case c.send <- data:
		default:
			// Client buffer full, skip
		}
	}
}

// visible is the delivery rule, kept in one place so a new event type cannot
// be added without deciding who it is for.
func visible(ctx context.Context, auth Authorizer, authEnabled bool, userID string, event Event) bool {
	if !authEnabled {
		return true
	}
	// With auth on, an unidentified connection sees nothing. So does every
	// connection if the hub was never given an authorizer — failing closed is
	// the only safe direction for a broadcast.
	if userID == "" || auth == nil {
		return false
	}

	if event.Entity == "team" {
		return auth.IsTeamMember(ctx, userID, event.EntityID)
	}
	if event.ResearchID == "" {
		// An event with no scope cannot be shown to anyone: there is nothing
		// to check it against, and "when in doubt, send it" is how the hub
		// became a public activity feed.
		return false
	}
	return auth.CanReadResearch(ctx, userID, event.ResearchID)
}

func (h *Hub) count() int {
	return len(h.clients)
}
