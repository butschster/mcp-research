package ws

import (
	"encoding/json"
	"log/slog"
	"sync"
)

// Event is broadcast to all connected WebSocket clients.
type Event struct {
	Type       string `json:"type"`        // e.g. "research.updated", "entry.created"
	ResearchID string `json:"research_id"` // scope
	EntityID   string `json:"entity_id"`   // ID of the changed entity
	Entity     string `json:"entity"`      // "research", "section", "entry", "session", "question", "task"
}

// Hub manages WebSocket connections and broadcasts events.
type Hub struct {
	mu      sync.RWMutex
	clients map[*Client]struct{}
	log     *slog.Logger
}

func NewHub(log *slog.Logger) *Hub {
	return &Hub{
		clients: make(map[*Client]struct{}),
		log:     log,
	}
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

// Broadcast sends an event to all connected clients.
func (h *Hub) Broadcast(event Event) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for c := range h.clients {
		select {
		case c.send <- data:
		default:
			// Client buffer full, skip
		}
	}
}

func (h *Hub) count() int {
	return len(h.clients)
}
