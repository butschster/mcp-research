package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMemoryJSON(t *testing.T) {
	for _, data := range []string{`["same","same",""]`, `[{"id":"x","text":"hi","author":"user","created_at":"2026-09-01T12:30:45Z","version":1}]`, `null`, `[]`} {
		var memory Memory
		if err := json.Unmarshal([]byte(data), &memory); err != nil {
			t.Fatal(err)
		}
		encoded, err := json.Marshal(memory)
		if err != nil {
			t.Fatal(err)
		}
		if len(memory) > 0 && !strings.HasPrefix(string(encoded), "[{") {
			t.Fatalf("exported strings instead of objects: %s", encoded)
		}
	}
	for _, data := range []string{`[null]`, `[42]`, `[true]`, `{}`, `[{"created_at":"invalid"}]`} {
		var memory Memory
		if err := json.Unmarshal([]byte(data), &memory); err == nil {
			t.Fatalf("invalid memory accepted: %s", data)
		}
	}
}
