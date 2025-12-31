// Package ast defines the abstract syntax tree nodes for SQL statements.
package ast

// Node is the base interface for all AST nodes.
type Node interface {
	Accept(visitor Visitor) interface{}
}

// Statement represents a SQL statement.
type Statement interface {
	Node
	statementNode()
}

// Expression represents a SQL expression.
type Expression interface {
	Node
	expressionNode()
}

// TableRef represents a table reference.
type TableRef struct {
	Database string
	Table    string
	Alias    string
}

func (t *TableRef) Accept(visitor Visitor) interface{} {
	return visitor.VisitTableRef(t)
}

// ColumnRefExpr represents a column reference expression.
type ColumnRefExpr struct {
	Table   string
	Column  string
	RawText string // Original expression text, e.g., "u.name" or "id"
}

func (c *ColumnRefExpr) Accept(visitor Visitor) interface{} {
	return visitor.VisitColumnRef(c)
}
func (c *ColumnRefExpr) expressionNode() {}

// FunctionCallExpr represents a function call expression.
type FunctionCallExpr struct {
	Name     string
	Args     []Expression
	Distinct bool
	Over     *WindowSpec
	RawText  string // Original expression text, e.g., "SUM(amount)"
}

func (f *FunctionCallExpr) Accept(visitor Visitor) interface{} {
	return visitor.VisitFunctionCall(f)
}
func (f *FunctionCallExpr) expressionNode() {}

// WindowSpec represents a window specification.
type WindowSpec struct {
	PartitionBy []Expression
	OrderBy     []*OrderByElement
}

// OrderByElement represents an ORDER BY element.
type OrderByElement struct {
	Expr Expression
	Desc bool
}

// BinaryExpr represents a binary expression.
type BinaryExpr struct {
	Left     Expression
	Operator string
	Right    Expression
	RawText  string // Original expression text
}

func (b *BinaryExpr) Accept(visitor Visitor) interface{} {
	return visitor.VisitBinaryExpr(b)
}
func (b *BinaryExpr) expressionNode() {}

// CaseExpr represents a CASE expression.
type CaseExpr struct {
	Operand  Expression
	WhenList []*WhenClause
	Else     Expression
	RawText  string // Original expression text
}

func (c *CaseExpr) Accept(visitor Visitor) interface{} {
	return visitor.VisitCaseExpr(c)
}
func (c *CaseExpr) expressionNode() {}

// WhenClause represents a WHEN clause in CASE expression.
type WhenClause struct {
	Condition Expression
	Result    Expression
}

// LiteralExpr represents a literal value.
type LiteralExpr struct {
	Value string
	Type  string // "string", "number", "bool", "null"
}

func (l *LiteralExpr) Accept(visitor Visitor) interface{} {
	return visitor.VisitLiteral(l)
}
func (l *LiteralExpr) expressionNode() {}

// SubqueryExpr represents a subquery expression.
type SubqueryExpr struct {
	Query *SelectStmt
}

func (s *SubqueryExpr) Accept(visitor Visitor) interface{} {
	return visitor.VisitSubquery(s)
}
func (s *SubqueryExpr) expressionNode() {}

// StarExpr represents a * or table.* expression.
type StarExpr struct {
	Table string // empty for *, non-empty for table.*
}

func (s *StarExpr) Accept(visitor Visitor) interface{} {
	return visitor.VisitStar(s)
}
func (s *StarExpr) expressionNode() {}

// AliasedExpr represents an expression with an optional alias.
type AliasedExpr struct {
	Expr  Expression
	Alias string
}

func (a *AliasedExpr) Accept(visitor Visitor) interface{} {
	return visitor.VisitAliasedExpr(a)
}
func (a *AliasedExpr) expressionNode() {}
