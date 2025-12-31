package lineage

import "errors"

var (
	// ErrUnsupportedSQL is returned when the SQL syntax is not supported.
	ErrUnsupportedSQL = errors.New("unsupported SQL syntax")

	// ErrTableNotFound is returned when a table is not found in the catalog.
	ErrTableNotFound = errors.New("table not found in catalog")

	// ErrColumnNotFound is returned when a column is not found in the table.
	ErrColumnNotFound = errors.New("column not found in table")
)
