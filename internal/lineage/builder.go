package lineage

import (
	"go-metadata/internal/lineage/ast"
	"go-metadata/internal/lineage/parser"

	"github.com/antlr4-go/antlr/v4"
)

// scopeMarker is used to mark the beginning of a query scope on the stack.
type scopeMarker struct {
	queryType string // "select", "subquery", "exists", etc.
}

// ASTBuilder builds custom AST from ANTLR parse tree.
type ASTBuilder struct {
	*parser.BaseSQLParserListener
	stack      []interface{}
	sourceSQL  string // Original SQL string for extracting text with spaces
	queryDepth int    // Track nested query depth
}

// NewASTBuilder creates a new AST builder.
func NewASTBuilder() *ASTBuilder {
	return &ASTBuilder{
		stack: make([]interface{}, 0),
	}
}

// NewASTBuilderWithSource creates a new AST builder with source SQL.
func NewASTBuilderWithSource(sql string) *ASTBuilder {
	return &ASTBuilder{
		stack:     make([]interface{}, 0),
		sourceSQL: sql,
	}
}

// getSourceText extracts text from original SQL using token positions.
func (b *ASTBuilder) getSourceText(ctx antlr.ParserRuleContext) string {
	if b.sourceSQL == "" || ctx == nil {
		return ctx.GetText()
	}
	start := ctx.GetStart()
	stop := ctx.GetStop()
	if start == nil || stop == nil {
		return ctx.GetText()
	}
	startIdx := start.GetStart()
	stopIdx := stop.GetStop()
	if startIdx < 0 || stopIdx < 0 || stopIdx >= len(b.sourceSQL) {
		return ctx.GetText()
	}
	return b.sourceSQL[startIdx : stopIdx+1]
}

func (b *ASTBuilder) push(v interface{}) {
	b.stack = append(b.stack, v)
}

func (b *ASTBuilder) pop() interface{} {
	if len(b.stack) == 0 {
		return nil
	}
	v := b.stack[len(b.stack)-1]
	b.stack = b.stack[:len(b.stack)-1]
	return v
}

func (b *ASTBuilder) peek() interface{} {
	if len(b.stack) == 0 {
		return nil
	}
	return b.stack[len(b.stack)-1]
}

// Result returns the built AST.
func (b *ASTBuilder) Result() ast.Statement {
	if len(b.stack) == 0 {
		return nil
	}
	if stmt, ok := b.stack[0].(ast.Statement); ok {
		return stmt
	}
	return nil
}

// EnterQueryTerm is called when entering queryTerm - push a scope marker.
func (b *ASTBuilder) EnterQueryTerm(ctx *parser.QueryTermContext) {
	// Only push marker if this is a real SELECT (not parenthesized expression)
	if ctx.SelectClause() != nil {
		b.queryDepth++
		b.push(&scopeMarker{queryType: "select"})
	}
}

// ExitQueryTerm is called when exiting queryTerm (the main SELECT structure).
func (b *ASTBuilder) ExitQueryTerm(ctx *parser.QueryTermContext) {
	// Skip if this is a parenthesized query expression
	if ctx.SelectClause() == nil {
		return
	}

	stmt := &ast.SelectStmt{
		SelectList: make([]*ast.AliasedExpr, 0),
	}

	// Process SELECT clause for DISTINCT
	if ctx.SelectClause() != nil {
		selectCtx := ctx.SelectClause().(*parser.SelectClauseContext)
		if selectCtx.DISTINCT() != nil {
			stmt.Distinct = true
		}
	}

	// Collect all items from stack until we hit our scope marker
	var fromClause *ast.FromClause
	var withClause *ast.WithClause
	selectExprs := make([]*ast.AliasedExpr, 0)

	// Pop items from stack until we hit the scope marker
loop:
	for len(b.stack) > 0 {
		item := b.pop()
		switch v := item.(type) {
		case *scopeMarker:
			// Found our marker, stop collecting
			break loop
		case *ast.FromClause:
			fromClause = v
		case *ast.WithClause:
			withClause = v
		case *ast.AliasedExpr:
			selectExprs = append([]*ast.AliasedExpr{v}, selectExprs...)
		case *ast.TableSource:
			// TableSource should be wrapped in FromClause, but handle it anyway
			if fromClause == nil {
				fromClause = &ast.FromClause{Tables: []*ast.TableSource{v}}
			} else {
				fromClause.Tables = append([]*ast.TableSource{v}, fromClause.Tables...)
			}
		case *ast.SelectStmt:
			// This is a subquery result, skip it (already processed)
			continue
		case *ast.ColumnRefExpr, *ast.BinaryExpr, *ast.LiteralExpr, *ast.FunctionCallExpr:
			// Skip expressions from ON/WHERE/HAVING conditions
			continue
		default:
			// Skip unknown items
			continue
		}
	}

	stmt.SelectList = selectExprs
	stmt.From = fromClause
	stmt.WithClause = withClause

	b.queryDepth--
	b.push(stmt)
}

// ExitSelectAll is called when exiting selectAll (*).
func (b *ASTBuilder) ExitSelectAll(ctx *parser.SelectAllContext) {
	b.push(&ast.AliasedExpr{
		Expr: &ast.StarExpr{Table: ""},
	})
}

// ExitSelectTableAll is called when exiting selectTableAll (table.*).
func (b *ASTBuilder) ExitSelectTableAll(ctx *parser.SelectTableAllContext) {
	tableName := ""
	if ctx.TableName() != nil {
		tableName = getText(ctx.TableName())
	}
	b.push(&ast.AliasedExpr{
		Expr: &ast.StarExpr{Table: tableName},
	})
}

// ExitSelectExpr is called when exiting selectExpr.
func (b *ASTBuilder) ExitSelectExpr(ctx *parser.SelectExprContext) {
	alias := ""
	if ctx.Alias() != nil {
		alias = getIdentifierText(getText(ctx.Alias()))
	}

	expr := b.pop()
	if e, ok := expr.(ast.Expression); ok {
		b.push(&ast.AliasedExpr{
			Expr:  e,
			Alias: alias,
		})
	}
}

// ExitColumnExpr is called when exiting columnExpr.
func (b *ASTBuilder) ExitColumnExpr(ctx *parser.ColumnExprContext) {
	colRef := ctx.ColumnRef().(*parser.ColumnRefContext)
	table := ""
	column := ""

	if colRef.TableName() != nil {
		table = getIdentifierText(getText(colRef.TableName()))
	}
	if colRef.ColumnName() != nil {
		column = getIdentifierText(getText(colRef.ColumnName()))
	}

	b.push(&ast.ColumnRefExpr{
		Table:   table,
		Column:  column,
		RawText: b.getSourceText(ctx),
	})
}

// ExitFuncExpr is called when exiting funcExpr.
func (b *ASTBuilder) ExitFuncExpr(ctx *parser.FuncExprContext) {
	funcCall := ctx.FunctionCall().(*parser.FunctionCallContext)
	funcName := ""
	if funcCall.FunctionName() != nil {
		funcName = getIdentifierText(getText(funcCall.FunctionName()))
	}

	// Collect arguments from stack
	args := make([]ast.Expression, 0)
	if funcCall.ExpressionList() != nil {
		exprList := funcCall.ExpressionList().(*parser.ExpressionListContext)
		argCount := len(exprList.AllExpression())
		for i := 0; i < argCount; i++ {
			if expr, ok := b.pop().(ast.Expression); ok {
				args = append([]ast.Expression{expr}, args...)
			}
		}
	}

	distinct := funcCall.DISTINCT() != nil

	// Get raw expression text with spaces preserved
	rawText := b.getSourceText(ctx)

	b.push(&ast.FunctionCallExpr{
		Name:     funcName,
		Args:     args,
		Distinct: distinct,
		RawText:  rawText,
	})
}

// ExitAddSubExpr is called when exiting addSubExpr.
func (b *ASTBuilder) ExitAddSubExpr(ctx *parser.AddSubExprContext) {
	right := b.pop().(ast.Expression)
	left := b.pop().(ast.Expression)
	op := ctx.GetOp().GetText()

	b.push(&ast.BinaryExpr{
		Left:     left,
		Operator: op,
		Right:    right,
		RawText:  b.getSourceText(ctx),
	})
}

// ExitMulDivExpr is called when exiting mulDivExpr.
func (b *ASTBuilder) ExitMulDivExpr(ctx *parser.MulDivExprContext) {
	right := b.pop().(ast.Expression)
	left := b.pop().(ast.Expression)
	op := ctx.GetOp().GetText()

	b.push(&ast.BinaryExpr{
		Left:     left,
		Operator: op,
		Right:    right,
		RawText:  b.getSourceText(ctx),
	})
}

// ExitLiteralExpr is called when exiting literalExpr.
func (b *ASTBuilder) ExitLiteralExpr(ctx *parser.LiteralExprContext) {
	literal := ctx.Literal().(*parser.LiteralContext)
	value := getText(literal)
	litType := "string"

	if literal.NUMBER() != nil {
		litType = "number"
	} else if literal.TRUE() != nil || literal.FALSE() != nil {
		litType = "bool"
	} else if literal.NULL() != nil {
		litType = "null"
	}

	b.push(&ast.LiteralExpr{
		Value: value,
		Type:  litType,
	})
}

// ExitCaseExpr is called when exiting caseExpr.
func (b *ASTBuilder) ExitCaseExpr(ctx *parser.CaseExprContext) {
	caseExprCtx := ctx.CaseExpression().(*parser.CaseExpressionContext)

	caseExpr := &ast.CaseExpr{
		WhenList: make([]*ast.WhenClause, 0),
	}

	// Count WHEN clauses
	whenCount := len(caseExprCtx.AllWHEN())

	// Pop ELSE if exists
	if caseExprCtx.ELSE() != nil {
		if elseExpr, ok := b.pop().(ast.Expression); ok {
			caseExpr.Else = elseExpr
		}
	}

	// Pop WHEN/THEN pairs
	for i := 0; i < whenCount; i++ {
		result := b.pop().(ast.Expression)
		condition := b.pop().(ast.Expression)
		caseExpr.WhenList = append([]*ast.WhenClause{{
			Condition: condition,
			Result:    result,
		}}, caseExpr.WhenList...)
	}

	// Pop operand if exists (CASE expr WHEN ...)
	allExprs := caseExprCtx.AllExpression()
	if len(allExprs) > whenCount*2 {
		if operand, ok := b.pop().(ast.Expression); ok {
			caseExpr.Operand = operand
		}
	}

	// Set raw expression text with spaces preserved
	caseExpr.RawText = b.getSourceText(ctx)

	b.push(caseExpr)
}

// ExitScalarSubqueryExpr is called when exiting scalarSubqueryExpr.
// This handles scalar subqueries in SELECT list like: (SELECT SUM(amount) FROM orders)
func (b *ASTBuilder) ExitScalarSubqueryExpr(ctx *parser.ScalarSubqueryExprContext) {
	// Pop the inner SelectStmt that was built by the nested query
	var query *ast.SelectStmt
	if stmt, ok := b.pop().(*ast.SelectStmt); ok {
		query = stmt
	}

	// Push a SubqueryExpr that wraps the inner query
	b.push(&ast.SubqueryExpr{
		Query: query,
	})
}

// ExitFromClause is called when exiting fromClause.
func (b *ASTBuilder) ExitFromClause(ctx *parser.FromClauseContext) {
	from := &ast.FromClause{
		Tables: make([]*ast.TableSource, 0),
	}

	// Collect table sources from stack
	for {
		if ts, ok := b.peek().(*ast.TableSource); ok {
			b.pop()
			from.Tables = append([]*ast.TableSource{ts}, from.Tables...)
		} else {
			break
		}
	}

	b.push(from)
}

// ExitTableNameFactor is called when exiting tableNameFactor.
func (b *ASTBuilder) ExitTableNameFactor(ctx *parser.TableNameFactorContext) {
	tableNameCtx := ctx.TableName().(*parser.TableNameContext)

	table := ""
	database := ""

	if tableNameCtx.DatabaseName() != nil {
		database = getIdentifierText(getText(tableNameCtx.DatabaseName()))
	}
	if tableNameCtx.Identifier() != nil {
		table = getIdentifierText(getText(tableNameCtx.Identifier()))
	}

	alias := ""
	if ctx.Alias() != nil {
		aliasCtx := ctx.Alias().(*parser.AliasContext)
		if aliasCtx.Identifier() != nil {
			alias = getIdentifierText(getText(aliasCtx.Identifier()))
		}
	}

	b.push(&ast.TableSource{
		Table: &ast.TableRef{
			Database: database,
			Table:    table,
			Alias:    alias,
		},
		Alias: alias,
		Joins: make([]*ast.JoinClause, 0),
	})
}

// ExitInsertStatement is called when exiting insertStatement.
func (b *ASTBuilder) ExitInsertStatement(ctx *parser.InsertStatementContext) {
	stmt := &ast.InsertStmt{}

	// Get table name
	tableNameCtx := ctx.TableName().(*parser.TableNameContext)
	table := ""
	database := ""

	if tableNameCtx.DatabaseName() != nil {
		database = getIdentifierText(getText(tableNameCtx.DatabaseName()))
	}
	if tableNameCtx.Identifier() != nil {
		table = getIdentifierText(getText(tableNameCtx.Identifier()))
	}

	stmt.Table = &ast.TableRef{
		Database: database,
		Table:    table,
	}

	// Get column list
	if ctx.ColumnList() != nil {
		colListCtx := ctx.ColumnList().(*parser.ColumnListContext)
		for _, id := range colListCtx.AllIdentifier() {
			stmt.Columns = append(stmt.Columns, getIdentifierText(getText(id)))
		}
	}

	// Get SELECT statement if exists
	if ctx.SelectStatement() != nil {
		if selectStmt, ok := b.pop().(*ast.SelectStmt); ok {
			stmt.Select = selectStmt
		}
	}

	b.push(stmt)
}

// ExitWithClause is called when exiting withClause.
func (b *ASTBuilder) ExitWithClause(ctx *parser.WithClauseContext) {
	wc := &ast.WithClause{
		Recursive: ctx.RECURSIVE() != nil,
		CTEs:      make([]*ast.CTE, 0),
	}

	// Collect CTEs from stack
	cteCount := len(ctx.AllCteDefinition())
	for i := 0; i < cteCount; i++ {
		if cte, ok := b.pop().(*ast.CTE); ok {
			wc.CTEs = append([]*ast.CTE{cte}, wc.CTEs...)
		}
	}

	b.push(wc)
}

// ExitCteDefinition is called when exiting cteDefinition.
func (b *ASTBuilder) ExitCteDefinition(ctx *parser.CteDefinitionContext) {
	name := ""
	// CTE name is the first identifier
	if len(ctx.AllIdentifier()) > 0 {
		name = getIdentifierText(getText(ctx.Identifier(0)))
	}

	var query *ast.SelectStmt
	if selectStmt, ok := b.pop().(*ast.SelectStmt); ok {
		query = selectStmt
	}

	b.push(&ast.CTE{
		Name:  name,
		Query: query,
	})
}

// getText extracts text from a parser rule context.
func getText(ctx antlr.ParserRuleContext) string {
	if ctx == nil {
		return ""
	}
	return ctx.GetText()
}

// getIdentifierText removes quotes from identifier.
func getIdentifierText(text string) string {
	if len(text) >= 2 {
		first := text[0]
		last := text[len(text)-1]
		if (first == '`' && last == '`') ||
			(first == '"' && last == '"') ||
			(first == '[' && last == ']') {
			return text[1 : len(text)-1]
		}
	}
	return text
}

// ParseSQL parses SQL string and returns AST.
func ParseSQL(sql string) (ast.Statement, error) {
	input := antlr.NewInputStream(sql)
	lexer := parser.NewSQLLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := parser.NewSQLParser(stream)

	// Remove default error listeners and add custom one
	p.RemoveErrorListeners()
	errorListener := &errorCollector{}
	p.AddErrorListener(errorListener)

	// Parse
	tree := p.SqlStatements()

	if errorListener.hasErrors() {
		return nil, ErrUnsupportedSQL
	}

	// Build AST with source SQL to preserve spaces
	builder := NewASTBuilderWithSource(sql)
	antlr.ParseTreeWalkerDefault.Walk(builder, tree)

	return builder.Result(), nil
}

// errorCollector collects parsing errors.
type errorCollector struct {
	*antlr.DefaultErrorListener
	errors []string
}

func (e *errorCollector) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{},
	line, column int, msg string, ex antlr.RecognitionException) {
	e.errors = append(e.errors, msg)
}

func (e *errorCollector) hasErrors() bool {
	return len(e.errors) > 0
}
