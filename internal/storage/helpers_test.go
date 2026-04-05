package storage

import (
	"database/sql"
	"testing"
)

func TestMarshalJSON(t *testing.T) {
	t.Run("nil slice", func(t *testing.T) {
		result := marshalJSON(nil)
		if result != "null" {
			t.Errorf("expected null, got %s", result)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		result := marshalJSON([]string{})
		if result != "[]" {
			t.Errorf("expected [], got %s", result)
		}
	})

	t.Run("non-empty slice", func(t *testing.T) {
		result := marshalJSON([]string{"a", "b", "c"})
		if result != `["a","b","c"]` {
			t.Errorf("expected [\"a\",\"b\",\"c\"], got %s", result)
		}
	})
}

func TestUnmarshalStringSlice(t *testing.T) {
	t.Run("invalid NullString", func(t *testing.T) {
		result := unmarshalStringSlice(sql.NullString{Valid: false})
		if len(result) != 0 {
			t.Errorf("expected empty slice, got %v", result)
		}
	})

	t.Run("null JSON", func(t *testing.T) {
		result := unmarshalStringSlice(sql.NullString{Valid: true, String: "null"})
		if len(result) != 0 {
			t.Errorf("expected empty slice, got %v", result)
		}
	})

	t.Run("empty string", func(t *testing.T) {
		result := unmarshalStringSlice(sql.NullString{Valid: true, String: ""})
		if len(result) != 0 {
			t.Errorf("expected empty slice, got %v", result)
		}
	})

	t.Run("empty JSON array", func(t *testing.T) {
		result := unmarshalStringSlice(sql.NullString{Valid: true, String: "[]"})
		if len(result) != 0 {
			t.Errorf("expected empty slice, got %v", result)
		}
	})

	t.Run("valid JSON array", func(t *testing.T) {
		result := unmarshalStringSlice(sql.NullString{Valid: true, String: `["x","y"]`})
		if len(result) != 2 {
			t.Fatalf("expected 2 elements, got %d", len(result))
		}
		if result[0] != "x" || result[1] != "y" {
			t.Errorf("expected [x y], got %v", result)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		result := unmarshalStringSlice(sql.NullString{Valid: true, String: "not-json"})
		if len(result) != 0 {
			t.Errorf("expected empty slice for invalid JSON, got %v", result)
		}
	})
}
