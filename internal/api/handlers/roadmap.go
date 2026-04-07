package handlers

import (
	"log/slog"
	"net/http"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/service"
)

type RoadmapHandler struct {
	roadmap  *service.RoadmapService
	research *service.ResearchService
	log      *slog.Logger
}

func NewRoadmapHandler(roadmap *service.RoadmapService, research *service.ResearchService, log *slog.Logger) *RoadmapHandler {
	return &RoadmapHandler{roadmap: roadmap, research: research, log: log}
}

func (h *RoadmapHandler) ListByResearch(w http.ResponseWriter, r *http.Request) {
	researchID, err := h.research.ResolveID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	roadmaps, err := h.roadmap.List(r.Context(), researchID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":  roadmaps,
		"count": len(roadmaps),
	})
}

func (h *RoadmapHandler) Get(w http.ResponseWriter, r *http.Request) {
	rm, err := h.roadmap.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": rm})
}

func (h *RoadmapHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ResearchID  string   `json:"research_id"`
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Statuses    []string `json:"statuses"`
		Nodes       []struct {
			TempID      string  `json:"temp_id"`
			Title       string  `json:"title"`
			Description string  `json:"description"`
			NodeType    string  `json:"node_type"`
			Status      string  `json:"status"`
			PositionX   float64 `json:"position_x"`
			PositionY   float64 `json:"position_y"`
			ParentID    string  `json:"parent_id"`
		} `json:"nodes"`
		Edges []struct {
			Source   string `json:"source"`
			Target   string `json:"target"`
			Label    string `json:"label"`
			EdgeType string `json:"edge_type"`
		} `json:"edges"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	if input.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	var nodes []service.CreateRoadmapNodeRequest
	for _, n := range input.Nodes {
		nodes = append(nodes, service.CreateRoadmapNodeRequest{
			TempID: n.TempID, Title: n.Title, Description: n.Description,
			NodeType: n.NodeType, Status: n.Status,
			PositionX: n.PositionX, PositionY: n.PositionY, ParentID: n.ParentID,
		})
	}

	var edges []service.CreateRoadmapEdgeRequest
	for _, e := range input.Edges {
		edges = append(edges, service.CreateRoadmapEdgeRequest{
			SourceNodeRef: e.Source, TargetNodeRef: e.Target,
			Label: e.Label, EdgeType: e.EdgeType,
		})
	}

	rm, err := h.roadmap.Create(r.Context(), service.CreateRoadmapRequest{
		ResearchID: input.ResearchID, Title: input.Title,
		Description: input.Description, Statuses: input.Statuses,
		Nodes: nodes, Edges: edges,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"data": rm})
}

func (h *RoadmapHandler) Update(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title       *string  `json:"title"`
		Description *string  `json:"description"`
		Statuses    []string `json:"statuses"`
		Status      *string  `json:"status"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}

	req := service.UpdateRoadmapRequest{
		Title:       input.Title,
		Description: input.Description,
		Statuses:    input.Statuses,
	}
	if input.Status != nil {
		s := domain.RoadmapStatus(*input.Status)
		req.Status = &s
	}

	rm, err := h.roadmap.Update(r.Context(), r.PathValue("id"), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": rm})
}

func (h *RoadmapHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.roadmap.Delete(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

func (h *RoadmapHandler) UpdateNode(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title       *string  `json:"title"`
		Description *string  `json:"description"`
		NodeType    *string  `json:"node_type"`
		Status      *string  `json:"status"`
		PositionX   *float64 `json:"position_x"`
		PositionY   *float64 `json:"position_y"`
		ParentID    *string  `json:"parent_id"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}

	node, err := h.roadmap.UpdateNode(r.Context(), r.PathValue("nodeId"), service.UpdateRoadmapNodeRequest{
		Title: input.Title, Description: input.Description,
		NodeType: input.NodeType, Status: input.Status,
		PositionX: input.PositionX, PositionY: input.PositionY,
		ParentID: input.ParentID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": node})
}
