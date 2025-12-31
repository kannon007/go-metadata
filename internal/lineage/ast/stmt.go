package ast

// SelectStmt represents a SELECT statement.
type SelectStmt struct {
	WithClause *WithClause
	Distinct   bool
	SelectList []*AliasedExpr
	From       *FromClause
	Where      Expression
	GroupBy    []Expression
	Having     Expression
	OrderBy    []*OrderByElement
	Limit      Expression
	Offset     Expression
}

func (s *SelectStmt) Accept(visitor Visitor) interface{} {
	return visitor.VisitSelectStmt(s)
}
func (s *SelectStmt) statementNode() {}

// WithClause represents a WITH clause (CTE).
type WithClause struct {
	Recursive bool
	CTEs      []*CTE
}

// CTE represents a Common Table Expression.
type CTE struct {
	Name  string
	Query *SelectStmt
}

// FromClause represents a FROM clause.
type FromClause struct {
	Tables []*TableSource
}

// TableSource represents a table source in FROM clause.
type TableSource struct {
	Table    *TableRef
	Subquery *SelectStmt
	Alias    string
	Joins    []*JoinClause
}

// JoinClause represents a JOIN clause.
type JoinClause struct {
	Type      string // INNER, LEFT, RIGHT, FULL, CROSS
	Table     *TableSource
	Condition Expression
}

// InsertStmt represents an INSERT statement.
type InsertStmt struct {
	Table   *TableRef
	Columns []string
	Select  *SelectStmt
	Values  [][]Expression
}

func (i *InsertStmt) Accept(visitor Visitor) interface{} {
	return visitor.VisitInsertStmt(i)
}
func (i *InsertStmt) statementNode() {}

// UpdateStmt represents an UPDATE statement.
type UpdateStmt struct {
	Table       *TableRef
	Assignments []*Assignment
	Where       Expression
}

func (u *UpdateStmt) Accept(visitor Visitor) interface{} {
	return visitor.VisitUpdateStmt(u)
}
func (u *UpdateStmt) statementNode() {}

// Assignment represents a column assignment in UPDATE.
type Assignment struct {
	Column string
	Value  Expression
}

// DeleteStmt represents a DELETE statement.
type DeleteStmt struct {
	Table *TableRef
	Where Expression
}

func (d *DeleteStmt) Accept(visitor Visitor) interface{} {
	return visitor.VisitDeleteStmt(d)
}
func (d *DeleteStmt) statementNode() {}
