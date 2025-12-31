// Package biz provides business logic and domain models.
package biz

import "time"

// TableMetadata represents metadata for a database table.
type TableMetadata struct {
	ID          string            `json:"id"`
	Database    string            `json:"database"`
	Schema      string            `json:"schema"`
	Name        string            `json:"name"`
	Type        string            `json:"type"` // table, view, etc.
	Comment     string            `json:"comment"`
	Columns     []*ColumnMetadata `json:"columns"`
	Indexes     []*IndexMetadata  `json:"indexes"`
	RowCount    int64             `json:"row_count"`
	DataSize    int64             `json:"data_size"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	CollectedAt time.Time         `json:"collected_at"`
}

// ColumnMetadata represents metadata for a table column.
type ColumnMetadata struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	DataType     string `json:"data_type"`
	Length       int    `json:"length"`
	Precision    int    `json:"precision"`
	Scale        int    `json:"scale"`
	Nullable     bool   `json:"nullable"`
	DefaultValue string `json:"default_value"`
	Comment      string `json:"comment"`
	IsPrimaryKey bool   `json:"is_primary_key"`
	IsForeignKey bool   `json:"is_foreign_key"`
	Position     int    `json:"position"`
}

// IndexMetadata represents metadata for a table index.
type IndexMetadata struct {
	Name      string   `json:"name"`
	Type      string   `json:"type"` // btree, hash, etc.
	Columns   []string `json:"columns"`
	IsUnique  bool     `json:"is_unique"`
	IsPrimary bool     `json:"is_primary"`
}

// DatabaseMetadata represents metadata for a database.
type DatabaseMetadata struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Charset     string           `json:"charset"`
	Collation   string           `json:"collation"`
	Tables      []*TableMetadata `json:"tables"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	CollectedAt time.Time        `json:"collected_at"`
}
