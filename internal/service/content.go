package service

import "strings"

// normalizeContent replaces literal escape sequences (e.g. backslash-n, backslash-t)
// with their actual characters. MCP clients sometimes send these instead of real
// newlines/tabs in markdown content.
func normalizeContent(s string) string {
	s = strings.ReplaceAll(s, `\n`, "\n")
	s = strings.ReplaceAll(s, `\t`, "\t")
	return s
}
