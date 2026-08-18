package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/butschster/mcp-research/internal/domain"
)

func marshalJSON(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "[]"
	}
	return string(data)
}

// marshalObject renders a value map for a JSON column. It differs from
// marshalJSON only in its zero value: an object column defaults to `{}`, and a
// column holding `[]` would fail every json_extract read against it.
func marshalObject(v map[string]any) string {
	if len(v) == 0 {
		return "{}"
	}
	data, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(data)
}

// unmarshalObject reads a JSON object column back. A row written before the
// column existed holds NULL or the empty string, and both mean "no values".
func unmarshalObject(s sql.NullString) map[string]any {
	if !s.Valid || s.String == "" || s.String == "null" {
		return nil
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(s.String), &result); err != nil {
		return nil
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// unmarshalFieldSpec reads a section's field declaration. A malformed value
// yields no fields rather than an error: a section whose spec cannot be parsed
// must still list its entries.
func unmarshalFieldSpec(s sql.NullString) []domain.FieldSpec {
	if !s.Valid || s.String == "" || s.String == "null" {
		return nil
	}
	var result []domain.FieldSpec
	if err := json.Unmarshal([]byte(s.String), &result); err != nil {
		return nil
	}
	return result
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
