package ast

// Visitor defines the interface for visiting AST nodes.
type Visitor interface {
	// Statements
	VisitSelectStmt(stmt *SelectStmt) interface{}
	VisitInsertStmt(stmt *InsertStmt) interface{}
	VisitUpdateStmt(stmt *UpdateStmt) interface{}
	VisitDeleteStmt(stmt *DeleteStmt) interface{}

	// Expressions
	VisitColumnRef(expr *ColumnRefExpr) interface{}
	VisitFunctionCall(expr *FunctionCallExpr) interface{}
	VisitBinaryExpr(expr *BinaryExpr) interface{}
	VisitCaseExpr(expr *CaseExpr) interface{}
	VisitLiteral(expr *LiteralExpr) interface{}
	VisitSubquery(expr *SubqueryExpr) interface{}
	VisitStar(expr *StarExpr) interface{}
	VisitAliasedExpr(expr *AliasedExpr) interface{}
	VisitTableRef(ref *TableRef) interface{}
}

// BaseVisitor provides default implementations for Visitor interface.
type BaseVisitor struct{}

func (v *BaseVisitor) VisitSelectStmt(stmt *SelectStmt) interface{}   { return nil }
func (v *BaseVisitor) VisitInsertStmt(stmt *InsertStmt) interface{}   { return nil }
func (v *BaseVisitor) VisitUpdateStmt(stmt *UpdateStmt) interface{}   { return nil }
func (v *BaseVisitor) VisitDeleteStmt(stmt *DeleteStmt) interface{}   { return nil }
func (v *BaseVisitor) VisitColumnRef(expr *ColumnRefExpr) interface{} { return nil }
func (v *BaseVisitor) VisitFunctionCall(expr *FunctionCallExpr) interface{} {
	return nil
}
func (v *BaseVisitor) VisitBinaryExpr(expr *BinaryExpr) interface{} { return nil }
func (v *BaseVisitor) VisitCaseExpr(expr *CaseExpr) interface{}     { return nil }
func (v *BaseVisitor) VisitLiteral(expr *LiteralExpr) interface{}   { return nil }
func (v *BaseVisitor) VisitSubquery(expr *SubqueryExpr) interface{} { return nil }
func (v *BaseVisitor) VisitStar(expr *StarExpr) interface{}         { return nil }
func (v *BaseVisitor) VisitAliasedExpr(expr *AliasedExpr) interface{} {
	return nil
}
func (v *BaseVisitor) VisitTableRef(ref *TableRef) interface{} { return nil }
