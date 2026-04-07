package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/storage"
	"github.com/google/uuid"
)

// --- Request DTOs ---

type CreateRoadmapNodeRequest struct {
	TempID      string  // Client-provided temp ID for edge references during bulk create
	Title       string
	Description string
	NodeType    string
	Status      string
	PositionX   float64
	PositionY   float64
	ParentID    string
}

type CreateRoadmapEdgeRequest struct {
	SourceNodeRef string // TempID or real node ID
	TargetNodeRef string // TempID or real node ID
	Label         string
	EdgeType      string
}

type CreateRoadmapRequest struct {
	ResearchID  string
	Title       string
	Description string
	Statuses    []string
	Nodes       []CreateRoadmapNodeRequest
	Edges       []CreateRoadmapEdgeRequest
}

type UpdateRoadmapRequest struct {
	Title       *string
	Description *string
	Statuses    []string
	Status      *domain.RoadmapStatus
}

type UpdateRoadmapNodeRequest struct {
	Title       *string
	Description *string
	NodeType    *string
	Status      *string
	PositionX   *float64
	PositionY   *float64
	ParentID    *string
}

// --- Service ---

type RoadmapService struct {
	roadmaps   *storage.RoadmapRepository
	nodes      *storage.RoadmapNodeRepository
	edges      *storage.RoadmapEdgeRepository
	researches *storage.ResearchRepository
	events     EventNotifier
	log        *slog.Logger
}

func NewRoadmapService(
	roadmaps *storage.RoadmapRepository,
	nodes *storage.RoadmapNodeRepository,
	edges *storage.RoadmapEdgeRepository,
	researches *storage.ResearchRepository,
	events EventNotifier,
	log *slog.Logger,
) *RoadmapService {
	return &RoadmapService{
		roadmaps:   roadmaps,
		nodes:      nodes,
		edges:      edges,
		researches: researches,
		events:     events,
		log:        log,
	}
}

// Create creates a roadmap with initial nodes and edges in one call.
func (s *RoadmapService) Create(ctx context.Context, req CreateRoadmapRequest) (*domain.Roadmap, error) {
	if err := validateResearchAccess(ctx, s.researches, req.ResearchID); err != nil {
		return nil, fmt.Errorf("research %s: %w", req.ResearchID, err)
	}

	statuses := req.Statuses
	if statuses == nil {
		statuses = []string{}
	}

	rm := &domain.Roadmap{
		ID:          uuid.New().String(),
		ResearchID:  req.ResearchID,
		Title:       req.Title,
		Description: req.Description,
		Statuses:    statuses,
		Status:      domain.RoadmapActive,
	}

	if err := s.roadmaps.Create(ctx, rm); err != nil {
		return nil, fmt.Errorf("create roadmap: %w", err)
	}

	// Create nodes, building a tempID -> realID map for edge resolution
	tempToReal := make(map[string]string)
	var nodes []*domain.RoadmapNode
	for _, nr := range req.Nodes {
		nodeType := nr.NodeType
		if nodeType == "" {
			nodeType = "step"
		}
		node := &domain.RoadmapNode{
			ID:          uuid.New().String(),
			RoadmapID:   rm.ID,
			Title:       nr.Title,
			Description: nr.Description,
			NodeType:    nodeType,
			Status:      nr.Status,
			PositionX:   nr.PositionX,
			PositionY:   nr.PositionY,
			ParentID:    nr.ParentID,
		}
		if err := s.nodes.Create(ctx, node); err != nil {
			return nil, fmt.Errorf("create node %q: %w", nr.Title, err)
		}
		if nr.TempID != "" {
			tempToReal[nr.TempID] = node.ID
		}
		nodes = append(nodes, node)
	}

	// Create edges, resolving temp IDs
	var edges []*domain.RoadmapEdge
	for _, er := range req.Edges {
		sourceID := er.SourceNodeRef
		if real, ok := tempToReal[sourceID]; ok {
			sourceID = real
		}
		targetID := er.TargetNodeRef
		if real, ok := tempToReal[targetID]; ok {
			targetID = real
		}
		edgeType := er.EdgeType
		if edgeType == "" {
			edgeType = "default"
		}
		edge := &domain.RoadmapEdge{
			ID:           uuid.New().String(),
			RoadmapID:    rm.ID,
			SourceNodeID: sourceID,
			TargetNodeID: targetID,
			Label:        er.Label,
			EdgeType:     edgeType,
		}
		if err := s.edges.Create(ctx, edge); err != nil {
			return nil, fmt.Errorf("create edge: %w", err)
		}
		edges = append(edges, edge)
	}

	rm.Nodes = nodes
	rm.Edges = edges

	s.events.Notify(Event{Type: "roadmap.created", ResearchID: rm.ResearchID, EntityID: rm.ID, Entity: "roadmap"})
	return rm, nil
}

// Get returns a roadmap with all nodes and edges.
func (s *RoadmapService) Get(ctx context.Context, id string) (*domain.Roadmap, error) {
	rm, err := s.roadmaps.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find roadmap: %w", err)
	}
	if rm == nil {
		return nil, ErrNotFound
	}
	if err := validateResearchAccess(ctx, s.researches, rm.ResearchID); err != nil {
		return nil, ErrNotFound
	}

	rm.Nodes, err = s.nodes.FindByRoadmap(ctx, rm.ID)
	if err != nil {
		return nil, fmt.Errorf("find nodes: %w", err)
	}
	rm.Edges, err = s.edges.FindByRoadmap(ctx, rm.ID)
	if err != nil {
		return nil, fmt.Errorf("find edges: %w", err)
	}

	return rm, nil
}

// List returns all roadmaps for a research (without nodes/edges).
func (s *RoadmapService) List(ctx context.Context, researchID string) ([]*domain.Roadmap, error) {
	if err := validateResearchAccess(ctx, s.researches, researchID); err != nil {
		return nil, err
	}
	return s.roadmaps.FindByResearch(ctx, researchID)
}

// Update updates roadmap metadata.
func (s *RoadmapService) Update(ctx context.Context, id string, req UpdateRoadmapRequest) (*domain.Roadmap, error) {
	rm, err := s.roadmaps.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find roadmap: %w", err)
	}
	if rm == nil {
		return nil, ErrNotFound
	}
	if err := validateResearchAccess(ctx, s.researches, rm.ResearchID); err != nil {
		return nil, ErrNotFound
	}

	if req.Title != nil {
		rm.Title = *req.Title
	}
	if req.Description != nil {
		rm.Description = *req.Description
	}
	if req.Statuses != nil {
		rm.Statuses = req.Statuses
	}
	if req.Status != nil {
		rm.Status = *req.Status
	}

	if err := s.roadmaps.Update(ctx, rm); err != nil {
		return nil, fmt.Errorf("update roadmap: %w", err)
	}

	s.events.Notify(Event{Type: "roadmap.updated", ResearchID: rm.ResearchID, EntityID: rm.ID, Entity: "roadmap"})
	return rm, nil
}

// Delete removes a roadmap and all its nodes/edges.
func (s *RoadmapService) Delete(ctx context.Context, id string) error {
	rm, err := s.roadmaps.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("find roadmap: %w", err)
	}
	if rm == nil {
		return ErrNotFound
	}
	if err := validateResearchAccess(ctx, s.researches, rm.ResearchID); err != nil {
		return ErrNotFound
	}

	if err := s.roadmaps.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete roadmap: %w", err)
	}

	s.events.Notify(Event{Type: "roadmap.deleted", ResearchID: rm.ResearchID, EntityID: id, Entity: "roadmap"})
	return nil
}

// AddNodes adds new nodes and edges to an existing roadmap.
func (s *RoadmapService) AddNodes(ctx context.Context, roadmapID string, nodeReqs []CreateRoadmapNodeRequest, edgeReqs []CreateRoadmapEdgeRequest) (*domain.Roadmap, error) {
	rm, err := s.roadmaps.FindByID(ctx, roadmapID)
	if err != nil {
		return nil, fmt.Errorf("find roadmap: %w", err)
	}
	if rm == nil {
		return nil, ErrNotFound
	}
	if err := validateResearchAccess(ctx, s.researches, rm.ResearchID); err != nil {
		return nil, ErrNotFound
	}

	tempToReal := make(map[string]string)
	for _, nr := range nodeReqs {
		nodeType := nr.NodeType
		if nodeType == "" {
			nodeType = "step"
		}
		node := &domain.RoadmapNode{
			ID:          uuid.New().String(),
			RoadmapID:   rm.ID,
			Title:       nr.Title,
			Description: nr.Description,
			NodeType:    nodeType,
			Status:      nr.Status,
			PositionX:   nr.PositionX,
			PositionY:   nr.PositionY,
			ParentID:    nr.ParentID,
		}
		if err := s.nodes.Create(ctx, node); err != nil {
			return nil, fmt.Errorf("create node %q: %w", nr.Title, err)
		}
		if nr.TempID != "" {
			tempToReal[nr.TempID] = node.ID
		}
	}

	for _, er := range edgeReqs {
		sourceID := er.SourceNodeRef
		if real, ok := tempToReal[sourceID]; ok {
			sourceID = real
		}
		targetID := er.TargetNodeRef
		if real, ok := tempToReal[targetID]; ok {
			targetID = real
		}
		edgeType := er.EdgeType
		if edgeType == "" {
			edgeType = "default"
		}
		edge := &domain.RoadmapEdge{
			ID:           uuid.New().String(),
			RoadmapID:    rm.ID,
			SourceNodeID: sourceID,
			TargetNodeID: targetID,
			Label:        er.Label,
			EdgeType:     edgeType,
		}
		if err := s.edges.Create(ctx, edge); err != nil {
			return nil, fmt.Errorf("create edge: %w", err)
		}
	}

	s.events.Notify(Event{Type: "roadmap.updated", ResearchID: rm.ResearchID, EntityID: rm.ID, Entity: "roadmap"})

	// Return full roadmap
	return s.Get(ctx, roadmapID)
}

// UpdateNode updates a single node.
func (s *RoadmapService) UpdateNode(ctx context.Context, nodeID string, req UpdateRoadmapNodeRequest) (*domain.RoadmapNode, error) {
	node, err := s.nodes.FindByID(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("find node: %w", err)
	}
	if node == nil {
		return nil, ErrNotFound
	}

	// Validate access via roadmap -> research
	rm, err := s.roadmaps.FindByID(ctx, node.RoadmapID)
	if err != nil {
		return nil, fmt.Errorf("find roadmap: %w", err)
	}
	if rm == nil {
		return nil, ErrNotFound
	}
	if err := validateResearchAccess(ctx, s.researches, rm.ResearchID); err != nil {
		return nil, ErrNotFound
	}

	if req.Title != nil {
		node.Title = *req.Title
	}
	if req.Description != nil {
		node.Description = *req.Description
	}
	if req.NodeType != nil {
		node.NodeType = *req.NodeType
	}
	if req.Status != nil {
		node.Status = *req.Status
	}
	if req.PositionX != nil {
		node.PositionX = *req.PositionX
	}
	if req.PositionY != nil {
		node.PositionY = *req.PositionY
	}
	if req.ParentID != nil {
		node.ParentID = *req.ParentID
	}

	if err := s.nodes.Update(ctx, node); err != nil {
		return nil, fmt.Errorf("update node: %w", err)
	}

	s.events.Notify(Event{Type: "roadmap.updated", ResearchID: rm.ResearchID, EntityID: rm.ID, Entity: "roadmap"})
	return node, nil
}

// RemoveNodes deletes nodes by IDs (edges cascade).
func (s *RoadmapService) RemoveNodes(ctx context.Context, roadmapID string, nodeIDs []string) error {
	rm, err := s.roadmaps.FindByID(ctx, roadmapID)
	if err != nil {
		return fmt.Errorf("find roadmap: %w", err)
	}
	if rm == nil {
		return ErrNotFound
	}
	if err := validateResearchAccess(ctx, s.researches, rm.ResearchID); err != nil {
		return ErrNotFound
	}

	for _, nodeID := range nodeIDs {
		if err := s.nodes.Delete(ctx, nodeID); err != nil {
			return fmt.Errorf("delete node %s: %w", nodeID, err)
		}
	}

	s.events.Notify(Event{Type: "roadmap.updated", ResearchID: rm.ResearchID, EntityID: rm.ID, Entity: "roadmap"})
	return nil
}
