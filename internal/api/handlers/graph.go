package handlers

import (
	"log/slog"
	"net/http"

	"github.com/butschster/mcp-research/internal/service"
	"github.com/butschster/mcp-research/internal/storage"
)

// GraphHandler builds a knowledge graph from all entities and their connections.
type GraphHandler struct {
	research  *service.ResearchService
	section   *service.SectionService
	entry     *service.EntryService
	session   *service.SessionService
	task      *service.TaskService
	entries   *storage.EntryRepository
	crossrefs *storage.CrossRefRepository
	log       *slog.Logger
}

func NewGraphHandler(
	research *service.ResearchService,
	section *service.SectionService,
	entry *service.EntryService,
	session *service.SessionService,
	task *service.TaskService,
	entries *storage.EntryRepository,
	crossrefs *storage.CrossRefRepository,
	log *slog.Logger,
) *GraphHandler {
	return &GraphHandler{
		research: research, section: section, entry: entry,
		session: session, task: task, entries: entries,
		crossrefs: crossrefs, log: log,
	}
}

type graphNode struct {
	ID     string   `json:"id"`
	Code   string   `json:"code"`
	Label  string   `json:"label"`
	Type   string   `json:"type"` // entry, section, session, question, task
	Status string   `json:"status"`
	Tags   []string `json:"tags,omitempty"`
	Group  string   `json:"group,omitempty"` // section name for entries, session title for questions
}

type graphEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type"` // crossref, tag, section, session
	Label  string `json:"label,omitempty"`
}

// Get returns the full knowledge graph for a research.
func (h *GraphHandler) Get(w http.ResponseWriter, r *http.Request) {
	idOrCode := r.PathValue("id")

	researchID, err := h.research.ResolveID(r.Context(), idOrCode)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	ctx := r.Context()
	var nodes []graphNode
	var edges []graphEdge
	nodeSet := map[string]bool{}

	// Helper to register a node
	addNode := func(n graphNode) {
		if !nodeSet[n.ID] {
			nodeSet[n.ID] = true
			nodes = append(nodes, n)
		}
	}

	// 1. Sections
	sections, err := h.section.List(ctx, researchID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	sectionNames := map[string]string{} // id -> display_name
	for _, s := range sections {
		name := s.DisplayName
		if name == "" {
			name = s.Name
		}
		sectionNames[s.ID] = name
		addNode(graphNode{
			ID: s.ID, Code: s.Code, Label: name,
			Type: "section", Status: string(s.Status),
		})
	}

	// 2. Entries (with tags)
	allEntries, err := h.entries.FindByResearch(ctx, researchID, storage.EntryFilter{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	entryTags := map[string][]string{}  // entry_id -> tags
	tagEntries := map[string][]string{} // tag -> []entry_id
	for _, e := range allEntries {
		group := sectionNames[e.SectionID]
		addNode(graphNode{
			ID: e.ID, Code: e.Code, Label: e.Title,
			Type: "entry", Status: string(e.Status),
			Tags: e.Tags, Group: group,
		})
		// Entry -> Section edge
		edges = append(edges, graphEdge{
			Source: e.SectionID, Target: e.ID,
			Type: "section",
		})
		// Entry -> Session edge
		if e.SessionID != "" {
			edges = append(edges, graphEdge{
				Source: e.SessionID, Target: e.ID,
				Type: "session", Label: "produced",
			})
		}
		entryTags[e.ID] = e.Tags
		for _, tag := range e.Tags {
			tagEntries[tag] = append(tagEntries[tag], e.ID)
		}
	}

	// 3. Sessions & Questions
	if h.session != nil {
		sessionList, err := h.session.ListByResearch(ctx, researchID)
		if err == nil {
			for _, sess := range sessionList {
				addNode(graphNode{
					ID: sess.ID, Code: sess.Code, Label: sess.Title,
					Type: "session", Status: string(sess.Status),
				})

				questions, err := h.session.ListQuestions(ctx, sess.ID, storage.QuestionFilter{})
				if err == nil {
					for _, q := range questions {
						addNode(graphNode{
							ID: q.ID, Code: q.Code, Label: q.Text,
							Type: "question", Status: string(q.Status),
							Group: sess.Title,
						})
						edges = append(edges, graphEdge{
							Source: sess.ID, Target: q.ID,
							Type: "session",
						})
					}
				}
			}
		}
	}

	// 4. Tasks
	if h.task != nil {
		tasks, err := h.task.List(ctx, researchID, storage.TaskFilter{})
		if err == nil {
			for _, t := range tasks {
				addNode(graphNode{
					ID: t.ID, Code: t.Code, Label: t.Title,
					Type: "task", Status: string(t.Status),
				})
			}
		}
	}

	// 5. Cross-references
	refs, err := h.crossrefs.FindByResearch(ctx, researchID)
	if err == nil {
		for _, ref := range refs {
			if !ref.Resolved {
				continue
			}
			sourceID := ref.SourceID
			if !nodeSet[sourceID] {
				continue
			}
			if ref.TargetEntryID != "" && nodeSet[ref.TargetEntryID] {
				edges = append(edges, graphEdge{
					Source: sourceID, Target: ref.TargetEntryID,
					Type: "crossref", Label: ref.TargetRef,
				})
			}
		}
	}

	// 6. Tag-based connections (entries sharing tags)
	for tag, entryIDs := range tagEntries {
		if len(entryIDs) < 2 || len(entryIDs) > 20 {
			// Skip tags that are too common (noise) or unique
			continue
		}
		for i := 0; i < len(entryIDs); i++ {
			for j := i + 1; j < len(entryIDs); j++ {
				edges = append(edges, graphEdge{
					Source: entryIDs[i], Target: entryIDs[j],
					Type: "tag", Label: tag,
				})
			}
		}
	}

	if nodes == nil {
		nodes = []graphNode{}
	}
	if edges == nil {
		edges = []graphEdge{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"nodes": nodes,
		"edges": edges,
	})
}
