package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/dovod-app/app/internal/domain"
	"github.com/dovod-app/app/internal/service"
)

// Dropping a markdown file into a section, in two calls.
//
// Named apart from ImportHandler in import.go, which is the whole-research JSON
// dump. Different granularity, different act, different trust: that one accepts
// a file this product wrote, this one accepts a file anybody wrote.
//
// The preview writes nothing and the commit takes fields rather than the file,
// so an edit somebody makes in the preview survives — and so the commit has no
// parameter through which a refused key could come back.

type MarkdownImportHandler struct {
	entries *service.EntryService
	log     *slog.Logger
}

func NewMarkdownImportHandler(entries *service.EntryService, log *slog.Logger) *MarkdownImportHandler {
	return &MarkdownImportHandler{entries: entries, log: log}
}

// maxUploadBytes bounds what the multipart reader will hold in memory, one step
// above the service's own file cap so an oversized file is still read far
// enough to be recognised and refused with the right sentence rather than by
// the request dying.
const maxUploadBytes = service.MaxImportFileBytes + (64 << 10)

// Preview parses one uploaded file and answers with what it made of it.
func (h *MarkdownImportHandler) Preview(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		// The size refusal has to be recognised here. Anything past the cap
		// fails inside ParseMultipartForm rather than reaching the service, so
		// a 10 MB file was being told its request was malformed — a sentence
		// about the wrong thing entirely, and the one case where the person
		// most needs to be told the actual limit.
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			writeServiceError(w, service.ErrImportTooLarge)
			return
		}
		writeError(w, http.StatusBadRequest, "expected a multipart upload with one file in the `file` field")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "no file in the `file` field")
		return
	}
	defer file.Close()

	// One more than the cap, so a file exactly at the limit still reads whole
	// and one byte over is recognisably over rather than silently truncated.
	data, err := io.ReadAll(io.LimitReader(file, service.MaxImportFileBytes+1))
	if err != nil {
		writeError(w, http.StatusBadRequest, "that file could not be read")
		return
	}

	parsed, err := h.entries.PreviewMarkdownImport(r.Context(), r.PathValue("id"), header.Filename, data)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": parsed})
}

type importCommitRequest struct {
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Status      domain.EntryStatus `json:"status"`
	Tags        []string           `json:"tags"`
	Body        string             `json:"body"`
	Metadata    map[string]any     `json:"metadata"`
}

// Commit writes the accepted preview as one entry.
func (h *MarkdownImportHandler) Commit(w http.ResponseWriter, r *http.Request) {
	var req importCommitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	entry, err := h.entries.ImportMarkdown(r.Context(), service.ImportEntryRequest{
		SectionID:   r.PathValue("id"),
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Tags:        req.Tags,
		Body:        req.Body,
		Metadata:    req.Metadata,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": entry})
}
