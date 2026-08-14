package handlers

import (
	"archive/zip"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/butschster/mcp-research/internal/service"
)

// ExportObsidian streams a research as a zip shaped like an Obsidian vault.
//
// It hangs off the existing `GET /api/researches/{id}/export` route behind
// `format=obsidian`, so it inherits that route's `wrapRead` wrapper and with it
// the ownership check — a research belonging to somebody else is a 404 here for
// the same reason it is one everywhere else.
func (h *ExportHandler) ExportObsidian(w http.ResponseWriter, r *http.Request) {
	if h.obsidian == nil {
		writeError(w, http.StatusInternalServerError, "obsidian export not configured")
		return
	}

	opts := service.VaultOptions{
		Sessions:  boolParam(r, "sessions", true),
		Tasks:     boolParam(r, "tasks", true),
		Roadmaps:  boolParam(r, "roadmaps", true),
		HTML:      boolParam(r, "html", true),
		Revisions: boolParam(r, "revisions", false),
	}

	// The whole vault is built before a byte of the response is written: once the
	// zip has started there is no way to turn a failure into a status code, and a
	// truncated archive that reports 200 is worse than an error.
	vault, err := h.obsidian.Vault(r.Context(), r.PathValue("id"), opts)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	filename := vault.Root + ".zip"
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", contentDisposition(filename))
	w.WriteHeader(http.StatusOK)

	zw := zip.NewWriter(w)
	modified := time.Now().UTC()
	for _, f := range vault.Files {
		fw, err := zw.CreateHeader(&zip.FileHeader{
			// No wrapper folder inside the archive: every extractor worth using
			// already creates one named after the zip, and a second one just
			// buries the vault a level deeper.
			Name:     f.Path,
			Method:   zip.Deflate,
			Modified: modified,
		})
		if err != nil {
			h.log.Error("obsidian export: create zip entry", "path", f.Path, "error", err)
			return
		}
		if _, err := fw.Write(f.Content); err != nil {
			// The client hung up, or the write failed: nothing useful can be said
			// to a response that is already half a zip file.
			h.log.Warn("obsidian export: write interrupted", "path", f.Path, "error", err)
			return
		}
	}
	if err := zw.Close(); err != nil {
		h.log.Error("obsidian export: close archive", "error", err)
	}
}

// boolParam reads a query flag.
//
// Only a recognized value moves an option; anything else keeps the default, so
// `revisions=off` — the obvious guess for someone typing a URL from the docs —
// does not silently ship the opposite of what it asked for.
func boolParam(r *http.Request, name string, def bool) bool {
	switch strings.ToLower(strings.TrimSpace(r.URL.Query().Get(name))) {
	case "0", "false", "no", "off", "n":
		return false
	case "1", "true", "yes", "on", "y":
		return true
	default:
		return def
	}
}

// contentDisposition names the file twice: an ASCII form every client
// understands, and the RFC 5987 form that carries a Cyrillic research name
// intact.
func contentDisposition(filename string) string {
	return fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`,
		asciiFallback(filename), url.PathEscape(filename))
}

// asciiFallback keeps the shape of the name for a client that ignores
// filename*: non-ASCII runes become '_' rather than disappearing, so a
// fully-Cyrillic name still produces a distinguishable file.
func asciiFallback(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r == '"' || r == '\\' || r < 0x20:
			b.WriteRune('_')
		case r > 0x7e:
			b.WriteRune('_')
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
