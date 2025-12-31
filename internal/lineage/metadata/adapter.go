package metadata

import (
	"go-metadata/internal/lineage"
)

// CatalogAdapter adapts a metadata.Provider to the lineage.Catalog interface.
type CatalogAdapter struct {
	provider Provider
}

// NewCatalogAdapter creates a new catalog adapter.
func NewCatalogAdapter(provider Provider) *CatalogAdapter {
	return &CatalogAdapter{provider: provider}
}

// GetTableSchema implements lineage.Catalog interface.
func (a *CatalogAdapter) GetTableSchema(db, table string) (*lineage.TableSchema, error) {
	schema, err := a.provider.GetTableSchema(db, table)
	if err != nil {
		return nil, err
	}

	// Convert metadata.TableSchema to lineage.TableSchema
	return &lineage.TableSchema{
		Database: schema.Database,
		Table:    schema.Table,
		Columns:  schema.GetColumnNames(),
	}, nil
}

// Provider returns the underlying metadata provider.
func (a *CatalogAdapter) Provider() Provider {
	return a.provider
}
