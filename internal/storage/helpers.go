package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

func marshalJSON(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "[]"
	}
	return string(data)
}

func unmarshalStringSlice(s sql.NullString) []string {
	if !s.Valid || s.String == "" || s.String == "null" {
		return []string{}
	}
	var result []string
	if err := json.Unmarshal([]byte(s.String), &result); err != nil {
		return []string{}
	}
	if result == nil {
		return []string{}
	}
	return result
}

func parseTime(s string) (t sql.NullTime) {
	for _, layout := range []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
	} {
		if parsed, err := fmt.Sscanf(s, layout); err == nil && parsed > 0 {
			break
		}
	}
	return
}
