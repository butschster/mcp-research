package service

// EventNotifier is called after data mutations to notify listeners.
type EventNotifier interface {
	Notify(event Event)
}

type Event struct {
	Type       string // e.g. "research.created", "entry.updated"
	ResearchID string
	EntityID   string
	Entity     string // "research", "section", "entry", "session", "question", "task"
}

// NoopNotifier does nothing (default when no WebSocket hub is connected).
type NoopNotifier struct{}

func (NoopNotifier) Notify(Event) {}
