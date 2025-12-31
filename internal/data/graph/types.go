// Package graph provides interfaces and types for graph database operations.
package graph

// NodeType represents the type of a graph node.
type NodeType string

const (
	NodeTypeDatabase NodeType = "database"
	NodeTypeTable    NodeType = "table"
	NodeTypeColumn   NodeType = "column"
	NodeTypeJob      NodeType = "job"
)

// EdgeType represents the type of a graph edge.
type EdgeType string

const (
	EdgeTypeContains   EdgeType = "contains"    // 包含关系
	EdgeTypeDependsOn  EdgeType = "depends_on"  // 依赖关系
	EdgeTypeProducedBy EdgeType = "produced_by" // 产出关系
)

// Node represents a graph node.
type Node struct {
	ID         string         `json:"id"`
	Type       NodeType       `json:"type"`
	Name       string         `json:"name"`
	Database   string         `json:"database,omitempty"`
	Table      string         `json:"table,omitempty"`
	Column     string         `json:"column,omitempty"`
	Properties map[string]any `json:"properties,omitempty"`
}

// Edge represents a graph edge.
type Edge struct {
	ID         string         `json:"id"`
	Type       EdgeType       `json:"type"`
	SourceID   string         `json:"source_id"`
	TargetID   string         `json:"target_id"`
	Properties map[string]any `json:"properties,omitempty"`
}

// LineageGraph represents a lineage graph containing nodes and edges.
type LineageGraph struct {
	Nodes []*Node `json:"nodes"`
	Edges []*Edge `json:"edges"`
}
