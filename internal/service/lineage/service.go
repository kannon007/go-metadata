// Package lineage provides lineage query service.
package lineage

import (
	"context"

	"go-metadata/internal/data/graph"
	lineageCore "go-metadata/internal/lineage"
)

// Service provides lineage query operations.
type Service struct {
	analyzer *lineageCore.Analyzer
	graphDB  graph.GraphDB
}

// NewService creates a new lineage service.
func NewService(analyzer *lineageCore.Analyzer, graphDB graph.GraphDB) *Service {
	return &Service{
		analyzer: analyzer,
		graphDB:  graphDB,
	}
}

// AnalyzeSQL analyzes SQL statement and extracts lineage.
func (s *Service) AnalyzeSQL(ctx context.Context, sql string) (*lineageCore.LineageResult, error) {
	if s.analyzer == nil {
		return nil, nil
	}
	return s.analyzer.Analyze(sql)
}

// GetColumnLineage retrieves column-level lineage.
func (s *Service) GetColumnLineage(ctx context.Context, database, table, column string, depth int) (*graph.LineageGraph, error) {
	if s.graphDB == nil {
		return nil, nil
	}
	
	nodeID := buildColumnNodeID(database, table, column)
	return s.graphDB.GetLineage(ctx, nodeID, depth)
}

// GetTableLineage retrieves table-level lineage.
func (s *Service) GetTableLineage(ctx context.Context, database, table string, depth int) (*graph.LineageGraph, error) {
	if s.graphDB == nil {
		return nil, nil
	}
	
	nodeID := buildTableNodeID(database, table)
	return s.graphDB.GetLineage(ctx, nodeID, depth)
}

// buildColumnNodeID builds a node ID for a column.
func buildColumnNodeID(database, table, column string) string {
	return database + "." + table + "." + column
}

// buildTableNodeID builds a node ID for a table.
func buildTableNodeID(database, table string) string {
	return database + "." + table
}
