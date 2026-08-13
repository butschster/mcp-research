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

// normalizeTitle sanitizes a single-line field (title, name, label). It strips
// invalid UTF-8 and collapses any real newline or tab into a space, but it does
// NOT expand literal backslash escapes: in a one-line field a backslash is data,
// not an escape. Running normalizeContent here would turn a title like
// `Path C:\notes\deep` into two lines and break the markdown heading it becomes
// on export.
func normalizeTitle(s string) string {
	s = sanitizeUTF8(s)
	s = strings.NewReplacer("\r\n", " ", "\n", " ", "\r", " ", "\t", " ").Replace(s)
	return strings.TrimSpace(s)
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
