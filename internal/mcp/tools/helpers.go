package tools

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/butschster/mcp-research/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var slugRegexp = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)

func isValidSlug(s string) bool {
	return slugRegexp.MatchString(s)
}

func successResult(data any) (*mcp.CallToolResult, any, error) {
	text, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return errorResult(fmt.Sprintf("marshal response: %v", err))
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: sanitizeUTF8(string(text))},
		},
	}, nil, nil
}

// sanitizeUTF8 removes U+FFFD replacement characters and any invalid UTF-8 sequences.
func sanitizeUTF8(s string) string {
	// Fast path: if no replacement chars and valid UTF-8, return as-is
	if !strings.ContainsRune(s, utf8.RuneError) && utf8.ValidString(s) {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size <= 1 {
			// Skip invalid byte
			i++
			continue
		}
		if r == utf8.RuneError {
			// Skip U+FFFD replacement character
			i += size
			continue
		}
		b.WriteRune(r)
		i += size
	}
	return b.String()
}

func errorResult(msg string) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			&mcp.TextContent{Text: msg},
		},
	}, nil, nil
}

func validationErrorResult(errs []string) (*mcp.CallToolResult, any, error) {
	return errorResult("Validation errors:\n- " + strings.Join(errs, "\n- "))
}

// derefStr returns the string value of a *string, or empty string if nil.
func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// derefFloat64 returns the float64 value of a *float64, or 0 if nil.
func derefFloat64(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}

func nodeFromInput(n RoadmapCreateNode) service.CreateRoadmapNodeRequest {
	return service.CreateRoadmapNodeRequest{
		TempID:      n.TempID,
		Title:       n.Title,
		Description: derefStr(n.Description),
		NodeType:    derefStr(n.NodeType),
		Status:      derefStr(n.Status),
		PositionX:   derefFloat64(n.PositionX),
		PositionY:   derefFloat64(n.PositionY),
		ParentID:    derefStr(n.ParentID),
		RefType:     derefStr(n.RefType),
		RefID:       derefStr(n.RefID),
		Metadata:    derefStr(n.Metadata),
		Stage:       derefStr(n.Stage),
		NodeDate:    derefStr(n.NodeDate),
	}
}

func edgeFromInput(e RoadmapCreateEdge) service.CreateRoadmapEdgeRequest {
	return service.CreateRoadmapEdgeRequest{
		SourceNodeRef: e.SourceNodeRef,
		TargetNodeRef: e.TargetNodeRef,
		Label:         derefStr(e.Label),
		EdgeType:      derefStr(e.EdgeType),
	}
}
