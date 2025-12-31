package tests

import (
	"fmt"
	"go-metadata/internal/lineage/parser"
	"testing"

	"github.com/antlr4-go/antlr/v4"
)

// DebugListener prints all enter/exit events.
type DebugListener struct {
	*parser.BaseSQLParserListener
}

func (l *DebugListener) EnterSelectStatement(ctx *parser.SelectStatementContext) {
	fmt.Println(">>> EnterSelectStatement")
}

func (l *DebugListener) ExitSelectStatement(ctx *parser.SelectStatementContext) {
	fmt.Println("<<< ExitSelectStatement")
}

func (l *DebugListener) EnterSelectExpr(ctx *parser.SelectExprContext) {
	fmt.Println(">>> EnterSelectExpr")
}

func (l *DebugListener) ExitSelectExpr(ctx *parser.SelectExprContext) {
	fmt.Printf("<<< ExitSelectExpr: %s\n", ctx.GetText())
}

func (l *DebugListener) EnterColumnExpr(ctx *parser.ColumnExprContext) {
	fmt.Println(">>> EnterColumnExpr")
}

func (l *DebugListener) ExitColumnExpr(ctx *parser.ColumnExprContext) {
	fmt.Printf("<<< ExitColumnExpr: %s\n", ctx.GetText())
}

func (l *DebugListener) EnterFromClause(ctx *parser.FromClauseContext) {
	fmt.Println(">>> EnterFromClause")
}

func (l *DebugListener) ExitFromClause(ctx *parser.FromClauseContext) {
	fmt.Println("<<< ExitFromClause")
}

func (l *DebugListener) EnterTableNameFactor(ctx *parser.TableNameFactorContext) {
	fmt.Println(">>> EnterTableNameFactor")
}

func (l *DebugListener) ExitTableNameFactor(ctx *parser.TableNameFactorContext) {
	fmt.Printf("<<< ExitTableNameFactor: %s\n", ctx.GetText())
}

// TestDebug_ParseTree prints the parse tree for debugging.
func TestDebug_ParseTree(t *testing.T) {
	sql := `SELECT id, name FROM users u WHERE EXISTS (SELECT 1 FROM orders o WHERE o.user_id = u.id)`

	input := antlr.NewInputStream(sql)
	lexer := parser.NewSQLLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := parser.NewSQLParser(stream)

	tree := p.SqlStatements()

	fmt.Printf("Parse Tree: %s\n", tree.ToStringTree(nil, p))

	fmt.Println("\n=== Walking with DebugListener ===")
	listener := &DebugListener{}
	antlr.ParseTreeWalkerDefault.Walk(listener, tree)
}
