// Package graph provides interfaces and types for graph database operations.
package graph

import "context"

// Config represents the configuration for a graph database connection.
type Config struct {
	Type     string // "nebula" or "neo4j"
	Host     string
	Port     int
	User     string
	Password string
	Space    string // NebulaGraph space 或 Neo4j database
}

// GraphDB defines the interface for graph database operations.
type GraphDB interface {
	// Connect establishes a connection to the graph database.
	Connect(ctx context.Context) error
	// Close closes the connection to the graph database.
	Close() error

	// Node operations

	// CreateNode creates a new node in the graph.
	CreateNode(ctx context.Context, node *Node) error
	// GetNode retrieves a node by its ID.
	GetNode(ctx context.Context, id string) (*Node, error)
	// UpdateNode updates an existing node.
	UpdateNode(ctx context.Context, node *Node) error
	// DeleteNode deletes a node by its ID.
	DeleteNode(ctx context.Context, id string) error

	// Edge operations

	// CreateEdge creates a new edge in the graph.
	CreateEdge(ctx context.Context, edge *Edge) error
	// GetEdge retrieves an edge by its ID.
	GetEdge(ctx context.Context, id string) (*Edge, error)
	// DeleteEdge deletes an edge by its ID.
	DeleteEdge(ctx context.Context, id string) error

	// Query operations

	// GetUpstream retrieves upstream nodes and edges for a given node.
	GetUpstream(ctx context.Context, nodeID string, depth int) ([]*Node, []*Edge, error)
	// GetDownstream retrieves downstream nodes and edges for a given node.
	GetDownstream(ctx context.Context, nodeID string, depth int) ([]*Node, []*Edge, error)
	// GetLineage retrieves the complete lineage graph for a given node.
	GetLineage(ctx context.Context, nodeID string, depth int) (*LineageGraph, error)

	// Batch operations

	// BatchCreateNodes creates multiple nodes in a single operation.
	BatchCreateNodes(ctx context.Context, nodes []*Node) error
	// BatchCreateEdges creates multiple edges in a single operation.
	BatchCreateEdges(ctx context.Context, edges []*Edge) error
}
