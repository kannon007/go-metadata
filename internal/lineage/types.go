// Package lineage provides SQL data lineage parsing capabilities.
package lineage

// ColumnRef represents a reference to a column in a table.
type ColumnRef struct {
	Database string `json:"database,omitempty"`
	Table    string `json:"table"`
	Column   string `json:"column"`
}

// ColumnLineage represents the lineage of a single target column.
type ColumnLineage struct {
	Target    ColumnRef   `json:"target"`
	Sources   []ColumnRef `json:"sources"`
	Operators []string    `json:"operators"`
}

// LineageResult represents the complete lineage result for a SQL statement.
type LineageResult struct {
	Columns []ColumnLineage `json:"columns"`
}
