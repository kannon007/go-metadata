package metadata

import (
	"go-metadata/internal/lineage/parser"
	"strings"

	"github.com/antlr4-go/antlr/v4"
)

// DDLParser parses DDL statements to extract table schema metadata.
type DDLParser struct {
	provider *MemoryProvider
}

// NewDDLParser creates a new DDL parser.
func NewDDLParser() *DDLParser {
	return &DDLParser{
		provider: NewMemoryProvider(),
	}
}

// ParseDDL parses a DDL statement and returns the extracted table schema.
func (p *DDLParser) ParseDDL(sql string) (*TableSchema, error) {
	input := antlr.NewInputStream(sql)
	lexer := parser.NewSQLLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	sqlParser := parser.NewSQLParser(stream)

	// Remove default error listeners and add custom one
	sqlParser.RemoveErrorListeners()
	errorListener := &ddlErrorListener{}
	sqlParser.AddErrorListener(errorListener)

	tree := sqlParser.SqlStatements()

	if errorListener.hasError {
		return nil, errorListener.err
	}

	// Walk the tree to extract schema
	extractor := &ddlSchemaExtractor{
		sourceSQL: sql,
	}
	antlr.ParseTreeWalkerDefault.Walk(extractor, tree)

	if extractor.schema == nil {
		return nil, nil // Not a CREATE TABLE statement
	}

	return extractor.schema, nil
}

// ParseMultipleDDL parses multiple DDL statements and returns all extracted schemas.
func (p *DDLParser) ParseMultipleDDL(sql string) ([]*TableSchema, error) {
	var schemas []*TableSchema

	// Remove line comments first
	lines := strings.Split(sql, "\n")
	var cleanLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "--") {
			continue
		}
		cleanLines = append(cleanLines, line)
	}
	cleanSQL := strings.Join(cleanLines, "\n")

	// Split by semicolon
	statements := strings.Split(cleanSQL, ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		// Skip if statement is only whitespace or comments
		if !containsDDLKeyword(stmt) {
			continue
		}

		schema, err := p.ParseDDL(stmt)
		if err != nil {
			// Log error but continue parsing other statements
			continue
		}
		if schema != nil {
			schemas = append(schemas, schema)
		}
	}

	return schemas, nil
}

// containsDDLKeyword checks if the statement contains a DDL keyword
func containsDDLKeyword(stmt string) bool {
	upper := strings.ToUpper(stmt)
	keywords := []string{"CREATE", "DROP", "ALTER", "INSERT", "SELECT"}
	for _, kw := range keywords {
		if strings.Contains(upper, kw) {
			return true
		}
	}
	return false
}

// ddlErrorListener captures parsing errors.
type ddlErrorListener struct {
	*antlr.DefaultErrorListener
	hasError bool
	err      error
}

func (l *ddlErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{},
	line, column int, msg string, e antlr.RecognitionException) {
	l.hasError = true
	l.err = &ParseError{Line: line, Column: column, Message: msg}
}

// ParseError represents a DDL parsing error.
type ParseError struct {
	Line    int
	Column  int
	Message string
}

func (e *ParseError) Error() string {
	return e.Message
}

// ddlSchemaExtractor extracts table schema from CREATE TABLE statements.
type ddlSchemaExtractor struct {
	*parser.BaseSQLParserListener
	schema    *TableSchema
	sourceSQL string
}

// EnterCreateTableStatement is called when entering a CREATE TABLE statement.
func (e *ddlSchemaExtractor) EnterCreateTableStatement(ctx *parser.CreateTableStatementContext) {
	e.schema = &TableSchema{
		Columns:   make([]ColumnSchema, 0),
		TableType: "TABLE",
	}

	// Extract table name
	if ctx.TableName() != nil {
		tableNameCtx := ctx.TableName().(*parser.TableNameContext)
		if tableNameCtx.DatabaseName() != nil {
			e.schema.Database = getIdentifierText(tableNameCtx.DatabaseName().GetText())
		}
		if tableNameCtx.Identifier() != nil {
			e.schema.Table = getIdentifierText(tableNameCtx.Identifier().GetText())
		}
	}

	// Check table type
	if ctx.EXTERNAL() != nil {
		e.schema.TableType = "EXTERNAL"
	} else if ctx.TEMPORARY() != nil || ctx.TEMP() != nil {
		e.schema.TableType = "TEMPORARY"
	}

	// Extract comment
	if ctx.COMMENT() != nil && ctx.STRING_LITERAL() != nil {
		e.schema.Comment = unquoteString(ctx.STRING_LITERAL().GetText())
	}
}

// EnterColumnDefinition is called when entering a column definition.
func (e *ddlSchemaExtractor) EnterColumnDefinition(ctx *parser.ColumnDefinitionContext) {
	if e.schema == nil {
		return
	}

	col := ColumnSchema{
		Nullable: true, // Default to nullable
	}

	// Extract column name
	if ctx.Identifier() != nil {
		col.Name = getIdentifierText(ctx.Identifier().GetText())
	}

	// Extract data type
	if ctx.DataType() != nil {
		col.DataType = ctx.DataType().GetText()
	}

	// Extract column constraints
	for _, constraint := range ctx.AllColumnConstraint() {
		constraintCtx := constraint.(*parser.ColumnConstraintContext)

		if constraintCtx.NOT() != nil && constraintCtx.NULL() != nil {
			col.Nullable = false
		}
		if constraintCtx.PRIMARY() != nil {
			col.PrimaryKey = true
			col.Nullable = false
		}
		if constraintCtx.DEFAULT() != nil {
			// Extract default expression
			if constraintCtx.Expression() != nil {
				col.DefaultExpr = constraintCtx.Expression().GetText()
			}
		}
	}

	// Extract column comment (defined at ColumnDefinition level)
	if ctx.COMMENT() != nil && ctx.STRING_LITERAL() != nil {
		col.Comment = unquoteString(ctx.STRING_LITERAL().GetText())
	}

	e.schema.Columns = append(e.schema.Columns, col)
}

// EnterCreateViewStatement is called when entering a CREATE VIEW statement.
func (e *ddlSchemaExtractor) EnterCreateViewStatement(ctx *parser.CreateViewStatementContext) {
	e.schema = &TableSchema{
		Columns:   make([]ColumnSchema, 0),
		TableType: "VIEW",
	}

	// Extract view name
	if ctx.TableName() != nil {
		tableNameCtx := ctx.TableName().(*parser.TableNameContext)
		if tableNameCtx.DatabaseName() != nil {
			e.schema.Database = getIdentifierText(tableNameCtx.DatabaseName().GetText())
		}
		if tableNameCtx.Identifier() != nil {
			e.schema.Table = getIdentifierText(tableNameCtx.Identifier().GetText())
		}
	}

	// Extract column names if specified
	for _, id := range ctx.AllIdentifier() {
		col := ColumnSchema{
			Name: getIdentifierText(id.GetText()),
		}
		e.schema.Columns = append(e.schema.Columns, col)
	}

	// Extract comment
	if ctx.COMMENT() != nil && ctx.STRING_LITERAL() != nil {
		e.schema.Comment = unquoteString(ctx.STRING_LITERAL().GetText())
	}
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

// unquoteString removes quotes from a string literal.
func unquoteString(text string) string {
	if len(text) >= 2 {
		first := text[0]
		last := text[len(text)-1]
		if (first == '\'' && last == '\'') ||
			(first == '"' && last == '"') {
			return text[1 : len(text)-1]
		}
	}
	return text
}
