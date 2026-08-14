package ws

import (
	"time"

	"github.com/butschster/mcp-research/internal/service"
)

// HubNotifier bridges service.EventNotifier to the WebSocket Hub.
type HubNotifier struct {
	hub *Hub
}

func NewHubNotifier(hub *Hub) *HubNotifier {
	return &HubNotifier{hub: hub}
}

func (n *HubNotifier) Notify(event service.Event) {
	n.hub.Broadcast(Event{
		Type:          event.Type,
		ResearchID:    event.ResearchID,
		EntityID:      event.EntityID,
		Entity:        event.Entity,
		ActorUserID:   event.ActorUserID,
		ActorClientID: event.ActorClientID,
		Reason:        event.Reason,
		Name:          event.Name,
		TargetUserID:  event.TargetUserID,
		At:            time.Now().UnixMilli(),
	})
}
