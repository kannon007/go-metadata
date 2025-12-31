package lineage

// resolver handles the core lineage resolution logic.
type resolver struct {
	catalog    Catalog
	aliasMap   map[string]string // alias -> table name
	lineageCtx *lineageContext
}

// lineageContext maintains the context during lineage resolution.
type lineageContext struct {
	columns []ColumnLineage
}

// newResolver creates a new resolver with the given catalog.
func newResolver(catalog Catalog) *resolver {
	return &resolver{
		catalog:  catalog,
		aliasMap: make(map[string]string),
		lineageCtx: &lineageContext{
			columns: make([]ColumnLineage, 0),
		},
	}
}

// Resolve parses the SQL and extracts lineage information.
func (r *resolver) Resolve(sql string) (*LineageResult, error) {
	// TODO: Implement SQL parsing using TiDB Parser
	// 1. Parse SQL to AST
	// 2. Normalize AST (expand *, bind aliases)
	// 3. Extract column dependencies
	// 4. Build lineage graph

	return &LineageResult{
		Columns: r.lineageCtx.columns,
	}, nil
}

// resolveTableAlias registers a table alias mapping.
func (r *resolver) resolveTableAlias(alias, tableName string) {
	r.aliasMap[alias] = tableName
}

// getTableName resolves an alias to its actual table name.
func (r *resolver) getTableName(aliasOrTable string) string {
	if tableName, ok := r.aliasMap[aliasOrTable]; ok {
		return tableName
	}
	return aliasOrTable
}

// addColumnLineage adds a column lineage entry to the context.
func (r *resolver) addColumnLineage(target ColumnRef, sources []ColumnRef, operators []string) {
	r.lineageCtx.columns = append(r.lineageCtx.columns, ColumnLineage{
		Target:    target,
		Sources:   sources,
		Operators: operators,
	})
}

// expandStar expands SELECT * using catalog information.
func (r *resolver) expandStar(db, table string) ([]string, error) {
	schema, err := r.catalog.GetTableSchema(db, table)
	if err != nil {
		return nil, err
	}
	return schema.Columns, nil
}
