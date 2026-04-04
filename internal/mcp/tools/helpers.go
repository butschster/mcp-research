package tools

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

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
			&mcp.TextContent{Text: string(text)},
		},
	}, nil, nil
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
