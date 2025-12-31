package graph

import "errors"

// Graph database-specific errors.
var (
	// ErrConnectionFailed is returned when the graph database connection fails.
	ErrConnectionFailed = errors.New("failed to connect to graph database")

	// ErrNodeNotFound is returned when the specified node is not found.
	ErrNodeNotFound = errors.New("node not found")

	// ErrEdgeNotFound is returned when the specified edge is not found.
	ErrEdgeNotFound = errors.New("edge not found")

	// ErrInvalidNodeType is returned when the node type is invalid.
	ErrInvalidNodeType = errors.New("invalid node type")

	// ErrInvalidEdgeType is returned when the edge type is invalid.
	ErrInvalidEdgeType = errors.New("invalid edge type")

	// ErrDuplicateNode is returned when attempting to create a node that already exists.
	ErrDuplicateNode = errors.New("node already exists")

	// ErrDuplicateEdge is returned when attempting to create an edge that already exists.
	ErrDuplicateEdge = errors.New("edge already exists")

	// ErrInvalidQuery is returned when the graph query is invalid.
	ErrInvalidQuery = errors.New("invalid graph query")

	// ErrConnectionClosed is returned when attempting to use a closed connection.
	ErrConnectionClosed = errors.New("connection is closed")

	// ErrSpaceNotFound is returned when the graph space/database is not found.
	ErrSpaceNotFound = errors.New("graph space not found")
)
