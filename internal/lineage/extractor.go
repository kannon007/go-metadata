package lineage

import (
	"fmt"
	"go-metadata/internal/lineage/ast"
)

// Extractor extracts column lineage from AST nodes.
type Extractor struct {
	catalog  Catalog
	scope    *Scope
	lineages []ColumnLineage
}

// Scope maintains the current resolution context.
type Scope struct {
	parent     *Scope
	tableAlias map[string]*ast.TableRef // alias -> table
	cteMap     map[string]*ast.SelectStmt
	columns    map[string][]string // table -> columns (from catalog)
}

// NewExtractor creates a new lineage extractor.
func NewExtractor(catalog Catalog) *Extractor {
	return &Extractor{
		catalog:  catalog,
		scope:    newScope(nil),
		lineages: make([]ColumnLineage, 0),
	}
}

func newScope(parent *Scope) *Scope {
	return &Scope{
		parent:     parent,
		tableAlias: make(map[string]*ast.TableRef),
		cteMap:     make(map[string]*ast.SelectStmt),
		columns:    make(map[string][]string),
	}
}

// Extract extracts lineage from a statement.
func (e *Extractor) Extract(stmt ast.Statement) (*LineageResult, error) {
	switch s := stmt.(type) {
	case *ast.SelectStmt:
		return e.extractSelect(s, "")
	case *ast.InsertStmt:
		return e.extractInsert(s)
	default:
		return &LineageResult{Columns: e.lineages}, nil
	}
}

// extractSelect extracts lineage from SELECT statement.
func (e *Extractor) extractSelect(stmt *ast.SelectStmt, targetTable string) (*LineageResult, error) {
	// Process WITH clause (CTEs)
	if stmt.WithClause != nil {
		for _, cte := range stmt.WithClause.CTEs {
			e.scope.cteMap[cte.Name] = cte.Query
		}
	}

	// Process FROM clause to build table alias map
	if stmt.From != nil {
		for _, ts := range stmt.From.Tables {
			e.registerTableSource(ts)
		}
	}

	// Process SELECT list
	for i, selectExpr := range stmt.SelectList {
		// Handle * and table.* specially - expand to multiple columns
		if starExpr, ok := selectExpr.Expr.(*ast.StarExpr); ok {
			e.expandStarExpr(starExpr, targetTable)
			continue
		}

		targetCol := ""
		if selectExpr.Alias != "" {
			targetCol = selectExpr.Alias
		} else if colRef, ok := selectExpr.Expr.(*ast.ColumnRefExpr); ok {
			targetCol = colRef.Column
		} else {
			targetCol = fmt.Sprintf("_col%d", i)
		}

		sources, operators := e.extractExprSources(selectExpr.Expr)

		target := ColumnRef{
			Table:  targetTable,
			Column: targetCol,
		}

		e.lineages = append(e.lineages, ColumnLineage{
			Target:    target,
			Sources:   sources,
			Operators: operators,
		})
	}

	return &LineageResult{Columns: e.lineages}, nil
}

// extractInsert extracts lineage from INSERT statement.
func (e *Extractor) extractInsert(stmt *ast.InsertStmt) (*LineageResult, error) {
	if stmt.Select == nil {
		return &LineageResult{Columns: e.lineages}, nil
	}

	targetTable := stmt.Table.Table

	// Process the SELECT part
	selectResult, err := e.extractSelect(stmt.Select, targetTable)
	if err != nil {
		return nil, err
	}

	// Map columns if INSERT has explicit column list
	if len(stmt.Columns) > 0 && len(selectResult.Columns) > 0 {
		for i := range selectResult.Columns {
			if i < len(stmt.Columns) {
				e.lineages[i].Target.Table = targetTable
				e.lineages[i].Target.Column = stmt.Columns[i]
			}
		}
	}

	return &LineageResult{Columns: e.lineages}, nil
}

// expandStarExpr expands a * or table.* expression to individual column lineages.
func (e *Extractor) expandStarExpr(starExpr *ast.StarExpr, targetTable string) {
	if starExpr.Table != "" {
		// table.* - expand columns from specific table
		alias := starExpr.Table
		if cols, ok := e.scope.columns[alias]; ok {
			tableName := e.resolveTableAlias(alias)
			for _, col := range cols {
				e.lineages = append(e.lineages, ColumnLineage{
					Target: ColumnRef{
						Table:  targetTable,
						Column: col,
					},
					Sources: []ColumnRef{{
						Table:  tableName,
						Column: col,
					}},
					Operators: []string{col},
				})
			}
		}
	} else {
		// * - expand columns from all tables in scope
		for alias, cols := range e.scope.columns {
			tableName := e.resolveTableAlias(alias)
			for _, col := range cols {
				e.lineages = append(e.lineages, ColumnLineage{
					Target: ColumnRef{
						Table:  targetTable,
						Column: col,
					},
					Sources: []ColumnRef{{
						Table:  tableName,
						Column: col,
					}},
					Operators: []string{col},
				})
			}
		}
	}
}

// registerTableSource registers a table source in the current scope.
func (e *Extractor) registerTableSource(ts *ast.TableSource) {
	if ts.Table != nil {
		alias := ts.Alias
		if alias == "" {
			alias = ts.Table.Table
		}
		e.scope.tableAlias[alias] = ts.Table

		// Load columns from catalog
		if e.catalog != nil {
			schema, err := e.catalog.GetTableSchema(ts.Table.Database, ts.Table.Table)
			if err == nil {
				e.scope.columns[alias] = schema.Columns
			}
		}
	}

	// Process joins
	for _, join := range ts.Joins {
		if join.Table != nil {
			e.registerTableSource(join.Table)
		}
	}
}

// extractExprSources extracts source columns and operators from an expression.
func (e *Extractor) extractExprSources(expr ast.Expression) ([]ColumnRef, []string) {
	sources := make([]ColumnRef, 0)
	operators := make([]string, 0)

	switch ex := expr.(type) {
	case *ast.ColumnRefExpr:
		tableName := e.resolveColumnTable(ex.Table, ex.Column)
		sources = append(sources, ColumnRef{
			Table:  tableName,
			Column: ex.Column,
		})
		// Use raw expression text as operator
		if ex.RawText != "" {
			operators = append(operators, ex.RawText)
		} else if ex.Table != "" {
			operators = append(operators, ex.Table+"."+ex.Column)
		} else {
			operators = append(operators, ex.Column)
		}

	case *ast.FunctionCallExpr:
		// Use raw expression text as operator
		if ex.RawText != "" {
			operators = append(operators, ex.RawText)
		} else {
			operators = append(operators, ex.Name)
		}
		for _, arg := range ex.Args {
			argSources, _ := e.extractExprSources(arg)
			sources = append(sources, argSources...)
		}

	case *ast.BinaryExpr:
		// Use raw expression text as operator
		if ex.RawText != "" {
			operators = append(operators, ex.RawText)
		} else {
			operators = append(operators, ex.Operator)
		}
		leftSources, _ := e.extractExprSources(ex.Left)
		rightSources, _ := e.extractExprSources(ex.Right)
		sources = append(sources, leftSources...)
		sources = append(sources, rightSources...)

	case *ast.CaseExpr:
		// Use raw expression text as operator
		if ex.RawText != "" {
			operators = append(operators, ex.RawText)
		} else {
			operators = append(operators, "case")
		}
		if ex.Operand != nil {
			opSources, _ := e.extractExprSources(ex.Operand)
			sources = append(sources, opSources...)
		}
		for _, when := range ex.WhenList {
			condSources, _ := e.extractExprSources(when.Condition)
			resultSources, _ := e.extractExprSources(when.Result)
			sources = append(sources, condSources...)
			sources = append(sources, resultSources...)
		}
		if ex.Else != nil {
			elseSources, _ := e.extractExprSources(ex.Else)
			sources = append(sources, elseSources...)
		}

	case *ast.StarExpr:
		operators = append(operators, "star")
		// Expand * using catalog
		if ex.Table != "" {
			tableName := e.resolveTableAlias(ex.Table)
			if cols, ok := e.scope.columns[ex.Table]; ok {
				for _, col := range cols {
					sources = append(sources, ColumnRef{
						Table:  tableName,
						Column: col,
					})
				}
			}
		} else {
			// Expand all tables
			for alias, cols := range e.scope.columns {
				tableName := e.resolveTableAlias(alias)
				for _, col := range cols {
					sources = append(sources, ColumnRef{
						Table:  tableName,
						Column: col,
					})
				}
			}
		}

	case *ast.SubqueryExpr:
		operators = append(operators, "subquery")
		// Recursively extract from subquery
		subExtractor := NewExtractor(e.catalog)
		subExtractor.scope = newScope(e.scope)
		subResult, _ := subExtractor.extractSelect(ex.Query, "")
		for _, col := range subResult.Columns {
			sources = append(sources, col.Sources...)
		}

	case *ast.AliasedExpr:
		return e.extractExprSources(ex.Expr)
	}

	return sources, operators
}

// resolveTableAlias resolves a table alias to the actual table name.
func (e *Extractor) resolveTableAlias(alias string) string {
	if alias == "" {
		return ""
	}
	if tableRef, ok := e.scope.tableAlias[alias]; ok {
		return tableRef.Table
	}
	// Check parent scope
	if e.scope.parent != nil {
		if tableRef, ok := e.scope.parent.tableAlias[alias]; ok {
			return tableRef.Table
		}
	}
	return alias
}

// resolveColumnTable resolves the table name for a column.
// If tableHint is provided, use it. Otherwise, try to find the table that contains this column.
func (e *Extractor) resolveColumnTable(tableHint, column string) string {
	// If table hint is provided, resolve it
	if tableHint != "" {
		return e.resolveTableAlias(tableHint)
	}

	// If only one table in scope, use it
	if len(e.scope.tableAlias) == 1 {
		for _, tableRef := range e.scope.tableAlias {
			return tableRef.Table
		}
	}

	// Try to find the table that contains this column using catalog
	for alias, cols := range e.scope.columns {
		for _, col := range cols {
			if col == column {
				return e.resolveTableAlias(alias)
			}
		}
	}

	return ""
}
