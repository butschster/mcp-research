package service

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"testing"

	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/storage"
)

func setupRoadmapService(t *testing.T) (*RoadmapService, *mockNotifier, *sql.DB, context.Context) {
	t.Helper()
	db := setupTestDB(t)
	notifier := &mockNotifier{}
	svc := NewRoadmapService(
		storage.NewRoadmapRepository(db),
		storage.NewRoadmapNodeRepository(db),
		storage.NewRoadmapEdgeRepository(db),
		storage.NewResearchRepository(db),
		notifier,
		slog.Default(),
	)
	return svc, notifier, db, context.Background()
}

func TestRoadmapService_Create(t *testing.T) {
	t.Run("creates empty roadmap", func(t *testing.T) {
		svc, notifier, db, ctx := setupRoadmapService(t)
		r := createTestResearch(t, db)
		notifier.reset()

		rm, err := svc.Create(ctx, CreateRoadmapRequest{
			ResearchID:  r.ID,
			Title:       "My Roadmap",
			Description: "A test roadmap",
			Statuses:    []string{"todo", "doing", "done"},
		})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if rm.ID == "" {
			t.Fatal("expected non-empty ID")
		}
		if rm.Code == "" {
			t.Fatal("expected auto-generated code")
		}
		if rm.Status != domain.RoadmapActive {
			t.Errorf("expected status active, got %s", rm.Status)
		}
		if len(rm.Statuses) != 3 {
			t.Errorf("expected 3 statuses, got %d", len(rm.Statuses))
		}
		if !notifier.hasEvent("roadmap.created") {
			t.Error("expected roadmap.created event")
		}
	})

	t.Run("creates roadmap with nodes and edges", func(t *testing.T) {
		svc, _, db, ctx := setupRoadmapService(t)
		r := createTestResearch(t, db)

		rm, err := svc.Create(ctx, CreateRoadmapRequest{
			ResearchID: r.ID,
			Title:      "Full Graph",
			Statuses:   []string{"not_started", "completed"},
			Nodes: []CreateRoadmapNodeRequest{
				{TempID: "n1", Title: "Step A", NodeType: "step", Status: "not_started"},
				{TempID: "n2", Title: "Step B", NodeType: "milestone", Status: "not_started"},
			},
			Edges: []CreateRoadmapEdgeRequest{
				{SourceNodeRef: "n1", TargetNodeRef: "n2", Label: "next", EdgeType: "default"},
			},
		})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if len(rm.Nodes) != 2 {
			t.Errorf("expected 2 nodes, got %d", len(rm.Nodes))
		}
		if len(rm.Edges) != 1 {
			t.Errorf("expected 1 edge, got %d", len(rm.Edges))
		}

		// Verify edge references were resolved from temp_id to real IDs
		edge := rm.Edges[0]
		if edge.SourceNodeID == "n1" || edge.TargetNodeID == "n2" {
			t.Error("expected temp_ids to be resolved to real UUIDs")
		}
		if edge.SourceNodeID != rm.Nodes[0].ID {
			t.Errorf("edge source: got %s, want %s", edge.SourceNodeID, rm.Nodes[0].ID)
		}
		if edge.TargetNodeID != rm.Nodes[1].ID {
			t.Errorf("edge target: got %s, want %s", edge.TargetNodeID, rm.Nodes[1].ID)
		}
	})

	t.Run("research not found", func(t *testing.T) {
		svc, _, _, ctx := setupRoadmapService(t)

		_, err := svc.Create(ctx, CreateRoadmapRequest{
			ResearchID: "nonexistent",
			Title:      "Fail",
		})
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("default node type is step", func(t *testing.T) {
		svc, _, db, ctx := setupRoadmapService(t)
		r := createTestResearch(t, db)

		rm, err := svc.Create(ctx, CreateRoadmapRequest{
			ResearchID: r.ID,
			Title:      "Defaults",
			Nodes: []CreateRoadmapNodeRequest{
				{TempID: "n1", Title: "No Type"},
			},
		})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if rm.Nodes[0].NodeType != "step" {
			t.Errorf("expected default node_type 'step', got %s", rm.Nodes[0].NodeType)
		}
	})
}

func TestRoadmapService_Get(t *testing.T) {
	t.Run("returns roadmap with nodes and edges", func(t *testing.T) {
		svc, _, db, ctx := setupRoadmapService(t)
		r := createTestResearch(t, db)

		created, _ := svc.Create(ctx, CreateRoadmapRequest{
			ResearchID: r.ID,
			Title:      "Get Test",
			Nodes: []CreateRoadmapNodeRequest{
				{TempID: "n1", Title: "A", NodeType: "step"},
				{TempID: "n2", Title: "B", NodeType: "step"},
			},
			Edges: []CreateRoadmapEdgeRequest{
				{SourceNodeRef: "n1", TargetNodeRef: "n2"},
			},
		})

		got, err := svc.Get(ctx, created.ID)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if got.Title != "Get Test" {
			t.Errorf("expected title 'Get Test', got %s", got.Title)
		}
		if len(got.Nodes) != 2 {
			t.Errorf("expected 2 nodes, got %d", len(got.Nodes))
		}
		if len(got.Edges) != 1 {
			t.Errorf("expected 1 edge, got %d", len(got.Edges))
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc, _, _, ctx := setupRoadmapService(t)

		_, err := svc.Get(ctx, "nonexistent")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestRoadmapService_List(t *testing.T) {
	svc, _, db, ctx := setupRoadmapService(t)
	r := createTestResearch(t, db)

	_, _ = svc.Create(ctx, CreateRoadmapRequest{ResearchID: r.ID, Title: "RM 1"})
	_, _ = svc.Create(ctx, CreateRoadmapRequest{ResearchID: r.ID, Title: "RM 2"})

	list, err := svc.List(ctx, r.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 roadmaps, got %d", len(list))
	}
}

func TestRoadmapService_Update(t *testing.T) {
	t.Run("partial update", func(t *testing.T) {
		svc, notifier, db, ctx := setupRoadmapService(t)
		r := createTestResearch(t, db)
		rm, _ := svc.Create(ctx, CreateRoadmapRequest{
			ResearchID: r.ID, Title: "Original", Description: "Old",
			Statuses: []string{"a"},
		})
		notifier.reset()

		updated, err := svc.Update(ctx, rm.ID, UpdateRoadmapRequest{
			Title:       ptr("Updated"),
			Description: ptr("New"),
			Statuses:    []string{"x", "y"},
		})
		if err != nil {
			t.Fatalf("update: %v", err)
		}
		if updated.Title != "Updated" {
			t.Errorf("title: got %s, want Updated", updated.Title)
		}
		if updated.Description != "New" {
			t.Errorf("description: got %s, want New", updated.Description)
		}
		if len(updated.Statuses) != 2 {
			t.Errorf("statuses: got %d, want 2", len(updated.Statuses))
		}
		if !notifier.hasEvent("roadmap.updated") {
			t.Error("expected roadmap.updated event")
		}
	})

	t.Run("update status", func(t *testing.T) {
		svc, _, db, ctx := setupRoadmapService(t)
		r := createTestResearch(t, db)
		rm, _ := svc.Create(ctx, CreateRoadmapRequest{ResearchID: r.ID, Title: "Status Test"})

		updated, err := svc.Update(ctx, rm.ID, UpdateRoadmapRequest{
			Status: ptr(domain.RoadmapCompleted),
		})
		if err != nil {
			t.Fatalf("update: %v", err)
		}
		if updated.Status != domain.RoadmapCompleted {
			t.Errorf("status: got %s, want completed", updated.Status)
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc, _, _, ctx := setupRoadmapService(t)

		_, err := svc.Update(ctx, "nonexistent", UpdateRoadmapRequest{Title: ptr("x")})
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestRoadmapService_Delete(t *testing.T) {
	t.Run("deletes roadmap", func(t *testing.T) {
		svc, notifier, db, ctx := setupRoadmapService(t)
		r := createTestResearch(t, db)
		rm, _ := svc.Create(ctx, CreateRoadmapRequest{ResearchID: r.ID, Title: "Delete Me"})
		notifier.reset()

		if err := svc.Delete(ctx, rm.ID); err != nil {
			t.Fatalf("delete: %v", err)
		}
		if !notifier.hasEvent("roadmap.deleted") {
			t.Error("expected roadmap.deleted event")
		}

		_, err := svc.Get(ctx, rm.ID)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound after delete, got %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc, _, _, ctx := setupRoadmapService(t)

		err := svc.Delete(ctx, "nonexistent")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestRoadmapService_AddNodes(t *testing.T) {
	t.Run("adds nodes and edges to existing roadmap", func(t *testing.T) {
		svc, _, db, ctx := setupRoadmapService(t)
		r := createTestResearch(t, db)

		rm, _ := svc.Create(ctx, CreateRoadmapRequest{
			ResearchID: r.ID, Title: "Extend Me",
			Nodes: []CreateRoadmapNodeRequest{
				{TempID: "n1", Title: "Existing", NodeType: "step"},
			},
		})

		existingNodeID := rm.Nodes[0].ID

		updated, err := svc.AddNodes(ctx, rm.ID, []CreateRoadmapNodeRequest{
			{TempID: "n2", Title: "New Node", NodeType: "milestone"},
		}, []CreateRoadmapEdgeRequest{
			{SourceNodeRef: existingNodeID, TargetNodeRef: "n2", Label: "then"},
		})
		if err != nil {
			t.Fatalf("add nodes: %v", err)
		}
		if len(updated.Nodes) != 2 {
			t.Errorf("expected 2 nodes, got %d", len(updated.Nodes))
		}
		if len(updated.Edges) != 1 {
			t.Errorf("expected 1 edge, got %d", len(updated.Edges))
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc, _, _, ctx := setupRoadmapService(t)

		_, err := svc.AddNodes(ctx, "nonexistent",
			[]CreateRoadmapNodeRequest{{Title: "X", NodeType: "step"}},
			nil,
		)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestRoadmapService_UpdateNode(t *testing.T) {
	t.Run("updates node fields", func(t *testing.T) {
		svc, notifier, db, ctx := setupRoadmapService(t)
		r := createTestResearch(t, db)

		rm, _ := svc.Create(ctx, CreateRoadmapRequest{
			ResearchID: r.ID, Title: "Node Test",
			Nodes: []CreateRoadmapNodeRequest{
				{TempID: "n1", Title: "Original", NodeType: "step", Status: "pending"},
			},
		})
		notifier.reset()

		node, err := svc.UpdateNode(ctx, rm.Nodes[0].ID, UpdateRoadmapNodeRequest{
			Title:       ptr("Updated Node"),
			Description: ptr("New description"),
			NodeType:    ptr("milestone"),
			Status:      ptr("completed"),
			PositionX:   ptr(float64(100)),
			PositionY:   ptr(float64(200)),
		})
		if err != nil {
			t.Fatalf("update node: %v", err)
		}
		if node.Title != "Updated Node" {
			t.Errorf("title: got %s, want 'Updated Node'", node.Title)
		}
		if node.Description != "New description" {
			t.Errorf("description: got %s, want 'New description'", node.Description)
		}
		if node.NodeType != "milestone" {
			t.Errorf("node_type: got %s, want milestone", node.NodeType)
		}
		if node.Status != "completed" {
			t.Errorf("status: got %s, want completed", node.Status)
		}
		if node.PositionX != 100 {
			t.Errorf("position_x: got %f, want 100", node.PositionX)
		}
		if !notifier.hasEvent("roadmap.updated") {
			t.Error("expected roadmap.updated event")
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc, _, _, ctx := setupRoadmapService(t)

		_, err := svc.UpdateNode(ctx, "nonexistent", UpdateRoadmapNodeRequest{
			Title: ptr("x"),
		})
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestRoadmapService_RemoveNodes(t *testing.T) {
	t.Run("removes nodes and cascades edges", func(t *testing.T) {
		svc, notifier, db, ctx := setupRoadmapService(t)
		r := createTestResearch(t, db)

		rm, _ := svc.Create(ctx, CreateRoadmapRequest{
			ResearchID: r.ID, Title: "Remove Test",
			Nodes: []CreateRoadmapNodeRequest{
				{TempID: "n1", Title: "A", NodeType: "step"},
				{TempID: "n2", Title: "B", NodeType: "step"},
				{TempID: "n3", Title: "C", NodeType: "step"},
			},
			Edges: []CreateRoadmapEdgeRequest{
				{SourceNodeRef: "n1", TargetNodeRef: "n2"},
				{SourceNodeRef: "n2", TargetNodeRef: "n3"},
			},
		})
		notifier.reset()

		// Remove middle node
		err := svc.RemoveNodes(ctx, rm.ID, []string{rm.Nodes[1].ID})
		if err != nil {
			t.Fatalf("remove nodes: %v", err)
		}
		if !notifier.hasEvent("roadmap.updated") {
			t.Error("expected roadmap.updated event")
		}

		// Verify
		got, _ := svc.Get(ctx, rm.ID)
		if len(got.Nodes) != 2 {
			t.Errorf("expected 2 nodes, got %d", len(got.Nodes))
		}
		if len(got.Edges) != 0 {
			t.Errorf("expected 0 edges (both cascaded), got %d", len(got.Edges))
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc, _, _, ctx := setupRoadmapService(t)

		err := svc.RemoveNodes(ctx, "nonexistent", []string{"x"})
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}
