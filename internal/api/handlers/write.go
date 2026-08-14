package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/service"
)

// WriteHandler provides REST write endpoints mirroring MCP tools.
type WriteHandler struct {
	research *service.ResearchService
	section  *service.SectionService
	entry    *service.EntryService
	session  *service.SessionService
	task     *service.TaskService
	log      *slog.Logger
}

func NewWriteHandler(
	research *service.ResearchService,
	section *service.SectionService,
	entry *service.EntryService,
	session *service.SessionService,
	task *service.TaskService,
	log *slog.Logger,
) *WriteHandler {
	return &WriteHandler{research: research, section: section, entry: entry, session: session, task: task, log: log}
}

// --- Research ---

func (h *WriteHandler) CreateResearch(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Goal        string   `json:"goal"`
		Tags        []string `json:"tags"`
		Sections    []struct {
			Name        string `json:"name"`
			DisplayName string `json:"display_name"`
			Description string `json:"description"`
			Position    int    `json:"position"`
		} `json:"sections"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	if input.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	var sections []service.CreateSectionRequest
	for _, s := range input.Sections {
		sections = append(sections, service.CreateSectionRequest{
			Name: s.Name, DisplayName: s.DisplayName,
			Description: s.Description, Position: s.Position,
		})
	}

	research, created, err := h.research.Create(r.Context(), service.CreateResearchRequest{
		Name: input.Name, Description: input.Description,
		Goal: input.Goal, Tags: input.Tags, Sections: sections,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"data": map[string]any{
			"research_id": research.ID, "code": research.Code,
			"name": research.Name, "status": research.Status,
			"sections_created": len(created),
		},
	})
}

func (h *WriteHandler) UpdateResearch(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var input struct {
		Name        *string  `json:"name"`
		Description *string  `json:"description"`
		Goal        *string  `json:"goal"`
		Status      *string  `json:"status"`
		Instruction *string  `json:"instruction"`
		Tags        []string `json:"tags"`
		Memory      []string `json:"memory"`
		AddMemory   *string  `json:"add_memory"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}

	var status *domain.ResearchStatus
	if input.Status != nil {
		s := domain.ResearchStatus(*input.Status)
		status = &s
	}

	research, err := h.research.Update(r.Context(), id, service.UpdateResearchRequest{
		Name: input.Name, Description: input.Description, Goal: input.Goal,
		Status: status, Instruction: input.Instruction, Tags: input.Tags,
		Memory: input.Memory, AddMemory: input.AddMemory,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": research})
}

// --- Sections ---

func (h *WriteHandler) AddSection(w http.ResponseWriter, r *http.Request) {
	researchID := r.PathValue("id")
	var input struct {
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		Description string `json:"description"`
		Position    int    `json:"position"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	if input.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	section, err := h.research.AddSection(r.Context(), researchID, service.CreateSectionRequest{
		Name: input.Name, DisplayName: input.DisplayName,
		Description: input.Description, Position: input.Position,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": section})
}

func (h *WriteHandler) UpdateSection(w http.ResponseWriter, r *http.Request) {
	sectionID := r.PathValue("sectionId")
	var input struct {
		DisplayName *string `json:"display_name"`
		Description *string `json:"description"`
		Status      *string `json:"status"`
		Position    *int    `json:"position"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}

	var status *domain.SectionStatus
	if input.Status != nil {
		s := domain.SectionStatus(*input.Status)
		status = &s
	}

	section, err := h.section.Update(r.Context(), sectionID, service.UpdateSectionRequest{
		DisplayName: input.DisplayName, Description: input.Description,
		Status: status, Position: input.Position,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": section})
}

// --- Entries ---

func (h *WriteHandler) CreateEntry(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ResearchID  string   `json:"research_id"`
		SectionID   string   `json:"section_id"`
		SessionID   string   `json:"session_id"`
		EntryType   string   `json:"entry_type"`
		Content     string   `json:"content"`
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Status      string   `json:"status"`
		Tags        []string `json:"tags"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	if input.ResearchID == "" || input.SectionID == "" || input.Content == "" {
		writeError(w, http.StatusBadRequest, "research_id, section_id, and content are required")
		return
	}

	entry, err := h.entry.Create(r.Context(), service.CreateEntryRequest{
		ResearchID: input.ResearchID, SectionID: input.SectionID, SessionID: input.SessionID,
		Type:    domain.EntryType(input.EntryType),
		Content: input.Content, Title: input.Title, Description: input.Description,
		Status: domain.EntryStatus(input.Status), Tags: input.Tags,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"data": map[string]any{
			"entry_id": entry.ID, "code": entry.Code,
			"title": entry.Title, "status": entry.Status, "entry_type": entry.Type,
		},
	})
}

func (h *WriteHandler) UpdateEntry(w http.ResponseWriter, r *http.Request) {
	entryID := r.PathValue("id")
	var input struct {
		EntryType   *string  `json:"entry_type"`
		Title       *string  `json:"title"`
		Content     *string  `json:"content"`
		Description *string  `json:"description"`
		Status      *string  `json:"status"`
		Tags        []string `json:"tags"`
		TextReplace *struct {
			From string `json:"from"`
			To   string `json:"to"`
		} `json:"text_replace"`
		SessionID *string `json:"session_id"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}

	var status *domain.EntryStatus
	if input.Status != nil {
		s := domain.EntryStatus(*input.Status)
		status = &s
	}

	var entryType *domain.EntryType
	if input.EntryType != nil {
		t := domain.EntryType(*input.EntryType)
		entryType = &t
	}
	var textReplace *service.TextReplace
	if input.TextReplace != nil {
		textReplace = &service.TextReplace{From: input.TextReplace.From, To: input.TextReplace.To}
	}

	entry, err := h.entry.Update(r.Context(), entryID, service.UpdateEntryRequest{
		Type:  entryType,
		Title: input.Title, Content: input.Content, Description: input.Description,
		Status: status, Tags: input.Tags, TextReplace: textReplace, SessionID: input.SessionID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": entry})
}

// --- Tasks ---

func (h *WriteHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ResearchID  string `json:"research_id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Priority    string `json:"priority"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	if input.ResearchID == "" || input.Title == "" {
		writeError(w, http.StatusBadRequest, "research_id and title are required")
		return
	}

	task, err := h.task.Create(r.Context(), service.CreateTaskRequest{
		ResearchID: input.ResearchID, Title: input.Title,
		Description: input.Description, Priority: domain.Priority(input.Priority),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": task})
}

func (h *WriteHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("id")
	var input struct {
		Title    *string `json:"title"`
		Status   *string `json:"status"`
		Priority *string `json:"priority"`
		Result   *string `json:"result"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}

	var status *domain.TaskStatus
	if input.Status != nil {
		s := domain.TaskStatus(*input.Status)
		status = &s
	}
	var priority *domain.Priority
	if input.Priority != nil {
		p := domain.Priority(*input.Priority)
		priority = &p
	}

	task, err := h.task.Update(r.Context(), taskID, service.UpdateTaskRequest{
		Title: input.Title, Status: status, Priority: priority, Result: input.Result,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": task})
}

func (h *WriteHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("id")
	if err := h.task.Delete(r.Context(), taskID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

// PatchEntry applies block operations to a blocks entry. A write, so it is
// registered with wrap(); the checkbox in the web UI posts here too, which keeps
// the browser and an agent indistinguishable to the server.
func (h *WriteHandler) PatchEntry(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var input struct {
		Rev string `json:"rev"`
		Ops []struct {
			Op      string         `json:"op"`
			ID      string         `json:"id"`
			Type    string         `json:"type"`
			Data    map[string]any `json:"data"`
			After   string         `json:"after"`
			Before  string         `json:"before"`
			At      string         `json:"at"`
			Item    string         `json:"item"`
			Checked *bool          `json:"checked"`
		} `json:"ops"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	if len(input.Ops) == 0 {
		writeError(w, http.StatusBadRequest, "ops is required")
		return
	}

	ops := make([]service.BlockOp, 0, len(input.Ops))
	for _, o := range input.Ops {
		ops = append(ops, service.BlockOp{
			Op: o.Op, ID: o.ID, Type: o.Type, Data: o.Data,
			After: o.After, Before: o.Before, At: o.At,
			Item: o.Item, Checked: o.Checked,
		})
	}

	entry, err := h.entry.PatchBlocks(r.Context(), id, service.PatchBlocksRequest{Ops: ops, Rev: input.Rev})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			writeError(w, http.StatusNotFound, "entry not found")
		case errors.Is(err, service.ErrConflict):
			// The one case a client can act on by itself: re-read and re-anchor.
			writeError(w, http.StatusConflict, err.Error())
		default:
			writeError(w, http.StatusBadRequest, err.Error())
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": entry, "rev": service.DocumentRev(entry.Content)})
}

func (h *WriteHandler) DeleteEntry(w http.ResponseWriter, r *http.Request) {
	entryID := r.PathValue("id")
	if err := h.entry.Delete(r.Context(), entryID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

// --- Sessions ---

func (h *WriteHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ResearchID string `json:"research_id"`
		Title      string `json:"title"`
		Focus      string `json:"focus"`
		Questions  []struct {
			Text      string `json:"text"`
			Area      string `json:"area"`
			Rationale string `json:"rationale"`
			Priority  string `json:"priority"`
			ParentID  string `json:"parent_id"`
			Position  int    `json:"position"`
		} `json:"questions"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	if input.ResearchID == "" || input.Title == "" {
		writeError(w, http.StatusBadRequest, "research_id and title are required")
		return
	}

	var questions []service.CreateQuestionRequest
	for _, q := range input.Questions {
		questions = append(questions, service.CreateQuestionRequest{
			Text: q.Text, Area: q.Area, Rationale: q.Rationale,
			Priority: domain.Priority(q.Priority), ParentID: q.ParentID, Position: q.Position,
		})
	}

	session, _, err := h.session.Create(r.Context(), service.CreateSessionRequest{
		ResearchID: input.ResearchID, Title: input.Title,
		Focus: input.Focus, Questions: questions,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": session})
}

func (h *WriteHandler) UpdateSession(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	var input struct {
		Title   *string `json:"title"`
		Focus   *string `json:"focus"`
		Status  *string `json:"status"`
		Notes   *string `json:"notes"`
		AddNote *string `json:"add_note"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}

	var status *domain.SessionStatus
	if input.Status != nil {
		s := domain.SessionStatus(*input.Status)
		status = &s
	}

	session, err := h.session.Update(r.Context(), sessionID, service.UpdateSessionRequest{
		Title: input.Title, Focus: input.Focus, Status: status,
		Notes: input.Notes, AddNote: input.AddNote,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": session})
}

// --- Questions ---

func (h *WriteHandler) UpdateQuestion(w http.ResponseWriter, r *http.Request) {
	questionID := r.PathValue("questionId")
	var input struct {
		Status *string `json:"status"`
		Answer *string `json:"answer"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}

	var status *domain.QuestionStatus
	if input.Status != nil {
		s := domain.QuestionStatus(*input.Status)
		status = &s
	}

	question, err := h.session.UpdateQuestion(r.Context(), questionID, status, input.Answer)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": question})
}

func (h *WriteHandler) AddQuestions(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	var input struct {
		Questions []struct {
			Text     string `json:"text"`
			Area     string `json:"area"`
			Priority string `json:"priority"`
		} `json:"questions"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	if len(input.Questions) == 0 {
		writeError(w, http.StatusBadRequest, "at least one question is required")
		return
	}

	var requests []service.CreateQuestionRequest
	for _, q := range input.Questions {
		requests = append(requests, service.CreateQuestionRequest{
			Text:     q.Text,
			Area:     q.Area,
			Priority: domain.Priority(q.Priority),
		})
	}

	questions, err := h.session.AddQuestions(r.Context(), sessionID, requests)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": questions, "count": len(questions)})
}

// --- Helpers ---

func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return false
	}
	return true
}
