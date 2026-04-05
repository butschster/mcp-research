package tools

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestIsValidSlug(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"simple lowercase", "abc", true},
		{"with hyphen", "a-b", true},
		{"with underscore", "a_b", true},
		{"letter then digit", "a1", true},
		{"starts with digit", "1abc", true},
		{"longer slug", "my-research-2024", true},
		{"empty string", "", false},
		{"starts with hyphen", "-abc", false},
		{"starts with underscore", "_abc", false},
		{"uppercase letters", "ABC", false},
		{"mixed case", "aBc", false},
		{"contains space", "a b", false},
		{"contains dot", "a.b", false},
		{"contains exclamation", "a!b", false},
		{"contains at sign", "a@b", false},
		{"only hyphen", "-", false},
		{"only underscore", "_", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidSlug(tt.input)
			if got != tt.want {
				t.Errorf("isValidSlug(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestSuccessResult(t *testing.T) {
	data := map[string]string{"key": "value"}

	result, second, err := successResult(data)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if second != nil {
		t.Fatalf("expected nil second return, got %v", second)
	}
	if result.IsError {
		t.Fatal("expected IsError to be false")
	}
	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content element, got %d", len(result.Content))
	}

	tc, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected *mcp.TextContent, got %T", result.Content[0])
	}

	// Verify the text is valid indented JSON matching the input
	var got map[string]string
	if err := json.Unmarshal([]byte(tc.Text), &got); err != nil {
		t.Fatalf("expected valid JSON, got unmarshal error: %v", err)
	}
	if got["key"] != "value" {
		t.Errorf("expected key=value, got key=%s", got["key"])
	}

	// Verify indentation (MarshalIndent with two spaces)
	expected, _ := json.MarshalIndent(data, "", "  ")
	if tc.Text != string(expected) {
		t.Errorf("expected indented JSON:\n%s\ngot:\n%s", expected, tc.Text)
	}
}

func TestSuccessResultUnmarshalableInput(t *testing.T) {
	// Channels cannot be marshaled to JSON
	ch := make(chan int)

	result, second, err := successResult(ch)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if second != nil {
		t.Fatalf("expected nil second return, got %v", second)
	}
	if !result.IsError {
		t.Fatal("expected IsError to be true for unmarshalable input")
	}

	tc, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected *mcp.TextContent, got %T", result.Content[0])
	}
	if !strings.Contains(tc.Text, "marshal response") {
		t.Errorf("expected error message containing 'marshal response', got %q", tc.Text)
	}
}

func TestErrorResult(t *testing.T) {
	msg := "something went wrong"

	result, second, err := errorResult(msg)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if second != nil {
		t.Fatalf("expected nil second return, got %v", second)
	}
	if !result.IsError {
		t.Fatal("expected IsError to be true")
	}
	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content element, got %d", len(result.Content))
	}

	tc, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected *mcp.TextContent, got %T", result.Content[0])
	}
	if tc.Text != msg {
		t.Errorf("expected text %q, got %q", msg, tc.Text)
	}
}

func TestValidationErrorResult(t *testing.T) {
	errs := []string{"field1 is required", "field2 must be positive"}

	result, second, err := validationErrorResult(errs)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if second != nil {
		t.Fatalf("expected nil second return, got %v", second)
	}
	if !result.IsError {
		t.Fatal("expected IsError to be true")
	}
	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content element, got %d", len(result.Content))
	}

	tc, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected *mcp.TextContent, got %T", result.Content[0])
	}

	expected := "Validation errors:\n- field1 is required\n- field2 must be positive"
	if tc.Text != expected {
		t.Errorf("expected text:\n%s\ngot:\n%s", expected, tc.Text)
	}
}

func TestValidationErrorResultSingleError(t *testing.T) {
	errs := []string{"name is required"}

	result, _, _ := validationErrorResult(errs)

	tc := result.Content[0].(*mcp.TextContent)
	expected := "Validation errors:\n- name is required"
	if tc.Text != expected {
		t.Errorf("expected text:\n%s\ngot:\n%s", expected, tc.Text)
	}
}
