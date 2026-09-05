package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

// MemoryItem is one independently editable note. Legacy notes have no known
// author or creation time: those facts must not be inferred from the research.
type MemoryItem struct {
	ID          string     `json:"id"`
	Text        string     `json:"text"`
	CreatedAt   *time.Time `json:"created_at"`
	SessionID   string     `json:"session_id,omitempty"`
	SessionCode string     `json:"session_code,omitempty"`
	Author      string     `json:"author"` // agent, user, or unknown for legacy notes
	Version     int        `json:"version"`
}

// Memory accepts v1 portable dumps on import, but always emits structured
// items. The database and every live API use only the structured form.
type Memory []MemoryItem

func (m *Memory) UnmarshalJSON(data []byte) error {
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	items := make(Memory, 0, len(raw))
	for _, value := range raw {
		if string(value) == "null" {
			return fmt.Errorf("memory items must be strings or objects, not null")
		}
		var item MemoryItem
		if len(value) > 0 && value[0] == '"' {
			if err := json.Unmarshal(value, &item.Text); err != nil {
				return err
			}
			item.Author = "unknown"
		} else if err := json.Unmarshal(value, &item); err != nil {
			return err
		}
		items = append(items, item)
	}
	*m = items
	return nil
}
