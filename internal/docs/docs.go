// Package docs embeds shared documentation files used by both the MCP prompts
// and the LLMs documentation endpoints. Single source of truth for all docs.
package docs

import "embed"

//go:embed *.md *.txt
var FS embed.FS

// ReadFile reads a file from the embedded docs filesystem.
func ReadFile(name string) ([]byte, error) {
	return FS.ReadFile(name)
}
