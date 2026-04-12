package service

import (
	"strings"
	"unicode/utf8"
)

// normalizeContent replaces literal escape sequences (e.g. backslash-n, backslash-t)
// with their actual characters. MCP clients sometimes send these instead of real
// newlines/tabs in markdown content. It also strips invalid UTF-8 sequences and
// U+FFFD replacement characters to prevent storing corrupted text.
func normalizeContent(s string) string {
	s = strings.ReplaceAll(s, `\n`, "\n")
	s = strings.ReplaceAll(s, `\t`, "\t")
	s = sanitizeUTF8(s)
	return s
}

// sanitizeUTF8 removes invalid UTF-8 bytes and U+FFFD replacement characters.
func sanitizeUTF8(s string) string {
	if !strings.ContainsRune(s, utf8.RuneError) && utf8.ValidString(s) {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size <= 1 {
			i++
			continue
		}
		if r == utf8.RuneError {
			i += size
			continue
		}
		b.WriteRune(r)
		i += size
	}
	return b.String()
}
