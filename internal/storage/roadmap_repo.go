package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/butschster/mcp-research/internal/domain"
)

type RoadmapRepository struct {
	db *sql.DB
}

func NewRoadmapRepository(db *sql.DB) *RoadmapRepository {
	return &RoadmapRepository{db: db}
}

// Create inserts a roadmap record.
func (r *RoadmapRepository) Create(ctx context.Context, rm *domain.Roadmap) error {
	now := time.Now().UTC().Format(time.DateTime)

	if rm.Code == "" {
		code, err := NextCode(ctx, r.db, "roadmaps", "RM", "research_id", rm.ResearchID)
		if err != nil {
			return fmt.Errorf("generate code: %w", err)
		}
		rm.Code = code
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO roadmaps (id, code, research_id, title, description, statuses, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		rm.ID, rm.Code, rm.ResearchID, rm.Title, rm.Description,
		marshalJSON(rm.Statuses), rm.Status,
		now, now,
	)
	if err != nil {
		return fmt.Errorf("insert roadmap: %w", err)
	}
	rm.CreatedAt, _ = time.Parse(time.DateTime, now)
	rm.UpdatedAt = rm.CreatedAt
	return nil
}

// Update updates roadmap metadata.
func (r *RoadmapRepository) Update(ctx context.Context, rm *domain.Roadmap) error {
	now := time.Now().UTC().Format(time.DateTime)
	_, err := r.db.ExecContext(ctx,
		`UPDATE roadmaps SET title=?, description=?, statuses=?, status=?, updated_at=? WHERE id=?`,
		rm.Title, rm.Description, marshalJSON(rm.Statuses), rm.Status, now, rm.ID,
	)
	if err != nil {
		return fmt.Errorf("update roadmap: %w", err)
	}
	rm.UpdatedAt, _ = time.Parse(time.DateTime, now)
	return nil
}

// FindByID returns a roadmap without nodes/edges.
func (r *RoadmapRepository) FindByID(ctx context.Context, id string) (*domain.Roadmap, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, code, research_id, title, description, statuses, status, created_at, updated_at
		 FROM roadmaps WHERE id=?`, id)
	return r.scanRoadmap(row)
}

// FindByCode returns a roadmap by its short code (e.g. RM1).
func (r *RoadmapRepository) FindByCode(ctx context.Context, code string) (*domain.Roadmap, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, code, research_id, title, description, statuses, status, created_at, updated_at
		 FROM roadmaps WHERE code=?`, code)
	return r.scanRoadmap(row)
}

// FindByResearch returns all roadmaps for a research.
func (r *RoadmapRepository) FindByResearch(ctx context.Context, researchID string) ([]*domain.Roadmap, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, code, research_id, title, description, statuses, status, created_at, updated_at
		 FROM roadmaps WHERE research_id=? ORDER BY created_at ASC`, researchID)
	if err != nil {
		return nil, fmt.Errorf("query roadmaps: %w", err)
	}
	defer rows.Close()

	var result []*domain.Roadmap
	for rows.Next() {
		rm, err := r.scanRoadmapRow(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, rm)
	}
	return result, rows.Err()
}

// Delete removes a roadmap (cascade deletes nodes and edges).
func (r *RoadmapRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM roadmaps WHERE id=?", id)
	return err
}

func (r *RoadmapRepository) scanRoadmap(row *sql.Row) (*domain.Roadmap, error) {
	var rm domain.Roadmap
	var createdAt, updatedAt string
	var statuses sql.NullString

	err := row.Scan(
		&rm.ID, &rm.Code, &rm.ResearchID, &rm.Title, &rm.Description,
		&statuses, &rm.Status,
		&createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan roadmap: %w", err)
	}
	rm.Statuses = unmarshalStringSlice(statuses)
	rm.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
	rm.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)
	return &rm, nil
}

func (r *RoadmapRepository) scanRoadmapRow(rows *sql.Rows) (*domain.Roadmap, error) {
	var rm domain.Roadmap
	var createdAt, updatedAt string
	var statuses sql.NullString

	err := rows.Scan(
		&rm.ID, &rm.Code, &rm.ResearchID, &rm.Title, &rm.Description,
		&statuses, &rm.Status,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan roadmap row: %w", err)
	}
	rm.Statuses = unmarshalStringSlice(statuses)
	rm.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
	rm.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)
	return &rm, nil
}

// --- Roadmap Nodes ---

type RoadmapNodeRepository struct {
	db *sql.DB
}

func NewRoadmapNodeRepository(db *sql.DB) *RoadmapNodeRepository {
	return &RoadmapNodeRepository{db: db}
}

// Create inserts a node.
func (r *RoadmapNodeRepository) Create(ctx context.Context, node *domain.RoadmapNode) error {
	now := time.Now().UTC().Format(time.DateTime)

	if node.Code == "" {
		code, err := NextCode(ctx, r.db, "roadmap_nodes", "N", "roadmap_id", node.RoadmapID)
		if err != nil {
			return fmt.Errorf("generate node code: %w", err)
		}
		node.Code = code
	}

	var parentID *string
	if node.ParentID != "" {
		parentID = &node.ParentID
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO roadmap_nodes (id, code, roadmap_id, title, description, node_type, status, position_x, position_y, parent_id, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		node.ID, node.Code, node.RoadmapID, node.Title, node.Description,
		node.NodeType, node.Status, node.PositionX, node.PositionY, parentID,
		now, now,
	)
	if err != nil {
		return fmt.Errorf("insert roadmap node: %w", err)
	}
	node.CreatedAt, _ = time.Parse(time.DateTime, now)
	node.UpdatedAt = node.CreatedAt
	return nil
}

// Update updates a node.
func (r *RoadmapNodeRepository) Update(ctx context.Context, node *domain.RoadmapNode) error {
	now := time.Now().UTC().Format(time.DateTime)

	var parentID *string
	if node.ParentID != "" {
		parentID = &node.ParentID
	}

	_, err := r.db.ExecContext(ctx,
		`UPDATE roadmap_nodes SET title=?, description=?, node_type=?, status=?, position_x=?, position_y=?, parent_id=?, updated_at=? WHERE id=?`,
		node.Title, node.Description, node.NodeType, node.Status,
		node.PositionX, node.PositionY, parentID, now, node.ID,
	)
	if err != nil {
		return fmt.Errorf("update roadmap node: %w", err)
	}
	node.UpdatedAt, _ = time.Parse(time.DateTime, now)
	return nil
}

// FindByID returns a single node.
func (r *RoadmapNodeRepository) FindByID(ctx context.Context, id string) (*domain.RoadmapNode, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, code, roadmap_id, title, description, node_type, status, position_x, position_y, parent_id, created_at, updated_at
		 FROM roadmap_nodes WHERE id=?`, id)
	return r.scanNode(row)
}

// FindByRoadmap returns all nodes for a roadmap.
func (r *RoadmapNodeRepository) FindByRoadmap(ctx context.Context, roadmapID string) ([]*domain.RoadmapNode, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, code, roadmap_id, title, description, node_type, status, position_x, position_y, parent_id, created_at, updated_at
		 FROM roadmap_nodes WHERE roadmap_id=? ORDER BY created_at ASC`, roadmapID)
	if err != nil {
		return nil, fmt.Errorf("query roadmap nodes: %w", err)
	}
	defer rows.Close()

	var result []*domain.RoadmapNode
	for rows.Next() {
		n, err := r.scanNodeRow(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, n)
	}
	return result, rows.Err()
}

// Delete removes a node (cascade deletes edges).
func (r *RoadmapNodeRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM roadmap_nodes WHERE id=?", id)
	return err
}

func (r *RoadmapNodeRepository) scanNode(row *sql.Row) (*domain.RoadmapNode, error) {
	var n domain.RoadmapNode
	var createdAt, updatedAt string
	var parentID sql.NullString

	err := row.Scan(
		&n.ID, &n.Code, &n.RoadmapID, &n.Title, &n.Description,
		&n.NodeType, &n.Status, &n.PositionX, &n.PositionY, &parentID,
		&createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan roadmap node: %w", err)
	}
	if parentID.Valid {
		n.ParentID = parentID.String
	}
	n.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
	n.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)
	return &n, nil
}

func (r *RoadmapNodeRepository) scanNodeRow(rows *sql.Rows) (*domain.RoadmapNode, error) {
	var n domain.RoadmapNode
	var createdAt, updatedAt string
	var parentID sql.NullString

	err := rows.Scan(
		&n.ID, &n.Code, &n.RoadmapID, &n.Title, &n.Description,
		&n.NodeType, &n.Status, &n.PositionX, &n.PositionY, &parentID,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan roadmap node row: %w", err)
	}
	if parentID.Valid {
		n.ParentID = parentID.String
	}
	n.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
	n.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)
	return &n, nil
}

// --- Roadmap Edges ---

type RoadmapEdgeRepository struct {
	db *sql.DB
}

func NewRoadmapEdgeRepository(db *sql.DB) *RoadmapEdgeRepository {
	return &RoadmapEdgeRepository{db: db}
}

// Create inserts an edge.
func (r *RoadmapEdgeRepository) Create(ctx context.Context, edge *domain.RoadmapEdge) error {
	now := time.Now().UTC().Format(time.DateTime)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO roadmap_edges (id, roadmap_id, source_node_id, target_node_id, label, edge_type, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		edge.ID, edge.RoadmapID, edge.SourceNodeID, edge.TargetNodeID,
		edge.Label, edge.EdgeType, now,
	)
	if err != nil {
		return fmt.Errorf("insert roadmap edge: %w", err)
	}
	edge.CreatedAt, _ = time.Parse(time.DateTime, now)
	return nil
}

// FindByRoadmap returns all edges for a roadmap.
func (r *RoadmapEdgeRepository) FindByRoadmap(ctx context.Context, roadmapID string) ([]*domain.RoadmapEdge, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, roadmap_id, source_node_id, target_node_id, label, edge_type, created_at
		 FROM roadmap_edges WHERE roadmap_id=? ORDER BY created_at ASC`, roadmapID)
	if err != nil {
		return nil, fmt.Errorf("query roadmap edges: %w", err)
	}
	defer rows.Close()

	var result []*domain.RoadmapEdge
	for rows.Next() {
		e, err := r.scanEdgeRow(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, rows.Err()
}

// DeleteByRoadmap removes all edges for a roadmap.
func (r *RoadmapEdgeRepository) DeleteByRoadmap(ctx context.Context, roadmapID string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM roadmap_edges WHERE roadmap_id=?", roadmapID)
	return err
}

func (r *RoadmapEdgeRepository) scanEdgeRow(rows *sql.Rows) (*domain.RoadmapEdge, error) {
	var e domain.RoadmapEdge
	var createdAt string

	err := rows.Scan(
		&e.ID, &e.RoadmapID, &e.SourceNodeID, &e.TargetNodeID,
		&e.Label, &e.EdgeType, &createdAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan roadmap edge row: %w", err)
	}
	e.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
	return &e, nil
}
