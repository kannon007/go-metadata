// Code generated from C:/Users/pxl/IdeaProjects/go-metadata/core/lineage/grammar/SQLParser.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // SQLParser
import "github.com/antlr4-go/antlr/v4"

// A complete Visitor for a parse tree produced by SQLParser.
type SQLParserVisitor interface {
	antlr.ParseTreeVisitor

	// Visit a parse tree produced by SQLParser#sqlStatements.
	VisitSqlStatements(ctx *SqlStatementsContext) interface{}

	// Visit a parse tree produced by SQLParser#sqlStatement.
	VisitSqlStatement(ctx *SqlStatementContext) interface{}

	// Visit a parse tree produced by SQLParser#dmlStatement.
	VisitDmlStatement(ctx *DmlStatementContext) interface{}

	// Visit a parse tree produced by SQLParser#ddlStatement.
	VisitDdlStatement(ctx *DdlStatementContext) interface{}

	// Visit a parse tree produced by SQLParser#selectStatement.
	VisitSelectStatement(ctx *SelectStatementContext) interface{}

	// Visit a parse tree produced by SQLParser#queryExpression.
	VisitQueryExpression(ctx *QueryExpressionContext) interface{}

	// Visit a parse tree produced by SQLParser#queryTerm.
	VisitQueryTerm(ctx *QueryTermContext) interface{}

	// Visit a parse tree produced by SQLParser#withClause.
	VisitWithClause(ctx *WithClauseContext) interface{}

	// Visit a parse tree produced by SQLParser#cteDefinition.
	VisitCteDefinition(ctx *CteDefinitionContext) interface{}

	// Visit a parse tree produced by SQLParser#selectClause.
	VisitSelectClause(ctx *SelectClauseContext) interface{}

	// Visit a parse tree produced by SQLParser#selectElements.
	VisitSelectElements(ctx *SelectElementsContext) interface{}

	// Visit a parse tree produced by SQLParser#selectAll.
	VisitSelectAll(ctx *SelectAllContext) interface{}

	// Visit a parse tree produced by SQLParser#selectTableAll.
	VisitSelectTableAll(ctx *SelectTableAllContext) interface{}

	// Visit a parse tree produced by SQLParser#selectExpr.
	VisitSelectExpr(ctx *SelectExprContext) interface{}

	// Visit a parse tree produced by SQLParser#fromClause.
	VisitFromClause(ctx *FromClauseContext) interface{}

	// Visit a parse tree produced by SQLParser#tableReferences.
	VisitTableReferences(ctx *TableReferencesContext) interface{}

	// Visit a parse tree produced by SQLParser#tableReference.
	VisitTableReference(ctx *TableReferenceContext) interface{}

	// Visit a parse tree produced by SQLParser#tableNameFactor.
	VisitTableNameFactor(ctx *TableNameFactorContext) interface{}

	// Visit a parse tree produced by SQLParser#subqueryFactor.
	VisitSubqueryFactor(ctx *SubqueryFactorContext) interface{}

	// Visit a parse tree produced by SQLParser#lateralSubqueryFactor.
	VisitLateralSubqueryFactor(ctx *LateralSubqueryFactorContext) interface{}

	// Visit a parse tree produced by SQLParser#unnestFactor.
	VisitUnnestFactor(ctx *UnnestFactorContext) interface{}

	// Visit a parse tree produced by SQLParser#tableValuedFunctionFactor.
	VisitTableValuedFunctionFactor(ctx *TableValuedFunctionFactorContext) interface{}

	// Visit a parse tree produced by SQLParser#tableSample.
	VisitTableSample(ctx *TableSampleContext) interface{}

	// Visit a parse tree produced by SQLParser#joinPart.
	VisitJoinPart(ctx *JoinPartContext) interface{}

	// Visit a parse tree produced by SQLParser#temporalJoinClause.
	VisitTemporalJoinClause(ctx *TemporalJoinClauseContext) interface{}

	// Visit a parse tree produced by SQLParser#joinType.
	VisitJoinType(ctx *JoinTypeContext) interface{}

	// Visit a parse tree produced by SQLParser#whereClause.
	VisitWhereClause(ctx *WhereClauseContext) interface{}

	// Visit a parse tree produced by SQLParser#groupByClause.
	VisitGroupByClause(ctx *GroupByClauseContext) interface{}

	// Visit a parse tree produced by SQLParser#groupByElements.
	VisitGroupByElements(ctx *GroupByElementsContext) interface{}

	// Visit a parse tree produced by SQLParser#groupByElement.
	VisitGroupByElement(ctx *GroupByElementContext) interface{}

	// Visit a parse tree produced by SQLParser#havingClause.
	VisitHavingClause(ctx *HavingClauseContext) interface{}

	// Visit a parse tree produced by SQLParser#orderByClause.
	VisitOrderByClause(ctx *OrderByClauseContext) interface{}

	// Visit a parse tree produced by SQLParser#orderByElement.
	VisitOrderByElement(ctx *OrderByElementContext) interface{}

	// Visit a parse tree produced by SQLParser#limitClause.
	VisitLimitClause(ctx *LimitClauseContext) interface{}

	// Visit a parse tree produced by SQLParser#insertStatement.
	VisitInsertStatement(ctx *InsertStatementContext) interface{}

	// Visit a parse tree produced by SQLParser#partitionSpec.
	VisitPartitionSpec(ctx *PartitionSpecContext) interface{}

	// Visit a parse tree produced by SQLParser#partitionElement.
	VisitPartitionElement(ctx *PartitionElementContext) interface{}

	// Visit a parse tree produced by SQLParser#columnList.
	VisitColumnList(ctx *ColumnListContext) interface{}

	// Visit a parse tree produced by SQLParser#valuesClause.
	VisitValuesClause(ctx *ValuesClauseContext) interface{}

	// Visit a parse tree produced by SQLParser#valueRow.
	VisitValueRow(ctx *ValueRowContext) interface{}

	// Visit a parse tree produced by SQLParser#onDuplicateKeyUpdate.
	VisitOnDuplicateKeyUpdate(ctx *OnDuplicateKeyUpdateContext) interface{}

	// Visit a parse tree produced by SQLParser#updateStatement.
	VisitUpdateStatement(ctx *UpdateStatementContext) interface{}

	// Visit a parse tree produced by SQLParser#updateElement.
	VisitUpdateElement(ctx *UpdateElementContext) interface{}

	// Visit a parse tree produced by SQLParser#deleteStatement.
	VisitDeleteStatement(ctx *DeleteStatementContext) interface{}

	// Visit a parse tree produced by SQLParser#truncateStatement.
	VisitTruncateStatement(ctx *TruncateStatementContext) interface{}

	// Visit a parse tree produced by SQLParser#mergeStatement.
	VisitMergeStatement(ctx *MergeStatementContext) interface{}

	// Visit a parse tree produced by SQLParser#mergeClause.
	VisitMergeClause(ctx *MergeClauseContext) interface{}

	// Visit a parse tree produced by SQLParser#mergeUpdateClause.
	VisitMergeUpdateClause(ctx *MergeUpdateClauseContext) interface{}

	// Visit a parse tree produced by SQLParser#mergeInsertClause.
	VisitMergeInsertClause(ctx *MergeInsertClauseContext) interface{}

	// Visit a parse tree produced by SQLParser#createStatement.
	VisitCreateStatement(ctx *CreateStatementContext) interface{}

	// Visit a parse tree produced by SQLParser#createTableStatement.
	VisitCreateTableStatement(ctx *CreateTableStatementContext) interface{}

	// Visit a parse tree produced by SQLParser#tableElementList.
	VisitTableElementList(ctx *TableElementListContext) interface{}

	// Visit a parse tree produced by SQLParser#tableElement.
	VisitTableElement(ctx *TableElementContext) interface{}

	// Visit a parse tree produced by SQLParser#watermarkDefinition.
	VisitWatermarkDefinition(ctx *WatermarkDefinitionContext) interface{}

	// Visit a parse tree produced by SQLParser#columnDefinition.
	VisitColumnDefinition(ctx *ColumnDefinitionContext) interface{}

	// Visit a parse tree produced by SQLParser#dataType.
	VisitDataType(ctx *DataTypeContext) interface{}

	// Visit a parse tree produced by SQLParser#primitiveType.
	VisitPrimitiveType(ctx *PrimitiveTypeContext) interface{}

	// Visit a parse tree produced by SQLParser#structField.
	VisitStructField(ctx *StructFieldContext) interface{}

	// Visit a parse tree produced by SQLParser#columnConstraint.
	VisitColumnConstraint(ctx *ColumnConstraintContext) interface{}

	// Visit a parse tree produced by SQLParser#identityOptions.
	VisitIdentityOptions(ctx *IdentityOptionsContext) interface{}

	// Visit a parse tree produced by SQLParser#referentialAction.
	VisitReferentialAction(ctx *ReferentialActionContext) interface{}

	// Visit a parse tree produced by SQLParser#referentialActionType.
	VisitReferentialActionType(ctx *ReferentialActionTypeContext) interface{}

	// Visit a parse tree produced by SQLParser#notEnforced.
	VisitNotEnforced(ctx *NotEnforcedContext) interface{}

	// Visit a parse tree produced by SQLParser#tableConstraint.
	VisitTableConstraint(ctx *TableConstraintContext) interface{}

	// Visit a parse tree produced by SQLParser#partitionedByClause.
	VisitPartitionedByClause(ctx *PartitionedByClauseContext) interface{}

	// Visit a parse tree produced by SQLParser#clusteredByClause.
	VisitClusteredByClause(ctx *ClusteredByClauseContext) interface{}

	// Visit a parse tree produced by SQLParser#distributedByClause.
	VisitDistributedByClause(ctx *DistributedByClauseContext) interface{}

	// Visit a parse tree produced by SQLParser#sortedByClause.
	VisitSortedByClause(ctx *SortedByClauseContext) interface{}

	// Visit a parse tree produced by SQLParser#tableInheritsClause.
	VisitTableInheritsClause(ctx *TableInheritsClauseContext) interface{}

	// Visit a parse tree produced by SQLParser#engineClause.
	VisitEngineClause(ctx *EngineClauseContext) interface{}

	// Visit a parse tree produced by SQLParser#charsetClause.
	VisitCharsetClause(ctx *CharsetClauseContext) interface{}

	// Visit a parse tree produced by SQLParser#collateClause.
	VisitCollateClause(ctx *CollateClauseContext) interface{}

	// Visit a parse tree produced by SQLParser#tablespaceClause.
	VisitTablespaceClause(ctx *TablespaceClauseContext) interface{}

	// Visit a parse tree produced by SQLParser#ttlClause.
	VisitTtlClause(ctx *TtlClauseContext) interface{}

	// Visit a parse tree produced by SQLParser#lifecycleClause.
	VisitLifecycleClause(ctx *LifecycleClauseContext) interface{}

	// Visit a parse tree produced by SQLParser#rowFormatClause.
	VisitRowFormatClause(ctx *RowFormatClauseContext) interface{}

	// Visit a parse tree produced by SQLParser#storedAsClause.
	VisitStoredAsClause(ctx *StoredAsClauseContext) interface{}

	// Visit a parse tree produced by SQLParser#locationClause.
	VisitLocationClause(ctx *LocationClauseContext) interface{}

	// Visit a parse tree produced by SQLParser#tablePropertiesClause.
	VisitTablePropertiesClause(ctx *TablePropertiesClauseContext) interface{}

	// Visit a parse tree produced by SQLParser#withOptionsClause.
	VisitWithOptionsClause(ctx *WithOptionsClauseContext) interface{}

	// Visit a parse tree produced by SQLParser#propertyList.
	VisitPropertyList(ctx *PropertyListContext) interface{}

	// Visit a parse tree produced by SQLParser#property.
	VisitProperty(ctx *PropertyContext) interface{}

	// Visit a parse tree produced by SQLParser#createViewStatement.
	VisitCreateViewStatement(ctx *CreateViewStatementContext) interface{}

	// Visit a parse tree produced by SQLParser#createDatabaseStatement.
	VisitCreateDatabaseStatement(ctx *CreateDatabaseStatementContext) interface{}

	// Visit a parse tree produced by SQLParser#createIndexStatement.
	VisitCreateIndexStatement(ctx *CreateIndexStatementContext) interface{}

	// Visit a parse tree produced by SQLParser#indexColumn.
	VisitIndexColumn(ctx *IndexColumnContext) interface{}

	// Visit a parse tree produced by SQLParser#dropStatement.
	VisitDropStatement(ctx *DropStatementContext) interface{}

	// Visit a parse tree produced by SQLParser#alterStatement.
	VisitAlterStatement(ctx *AlterStatementContext) interface{}

	// Visit a parse tree produced by SQLParser#alterTableAction.
	VisitAlterTableAction(ctx *AlterTableActionContext) interface{}

	// Visit a parse tree produced by SQLParser#castExpr.
	VisitCastExpr(ctx *CastExprContext) interface{}

	// Visit a parse tree produced by SQLParser#extractExpr.
	VisitExtractExpr(ctx *ExtractExprContext) interface{}

	// Visit a parse tree produced by SQLParser#typeCastExpr.
	VisitTypeCastExpr(ctx *TypeCastExprContext) interface{}

	// Visit a parse tree produced by SQLParser#existsExpr.
	VisitExistsExpr(ctx *ExistsExprContext) interface{}

	// Visit a parse tree produced by SQLParser#bitwiseNotExpr.
	VisitBitwiseNotExpr(ctx *BitwiseNotExprContext) interface{}

	// Visit a parse tree produced by SQLParser#parenExpr.
	VisitParenExpr(ctx *ParenExprContext) interface{}

	// Visit a parse tree produced by SQLParser#concatExpr.
	VisitConcatExpr(ctx *ConcatExprContext) interface{}

	// Visit a parse tree produced by SQLParser#betweenExpr.
	VisitBetweenExpr(ctx *BetweenExprContext) interface{}

	// Visit a parse tree produced by SQLParser#columnExpr.
	VisitColumnExpr(ctx *ColumnExprContext) interface{}

	// Visit a parse tree produced by SQLParser#variableExpr.
	VisitVariableExpr(ctx *VariableExprContext) interface{}

	// Visit a parse tree produced by SQLParser#arrayAccessExpr.
	VisitArrayAccessExpr(ctx *ArrayAccessExprContext) interface{}

	// Visit a parse tree produced by SQLParser#unaryMinusExpr.
	VisitUnaryMinusExpr(ctx *UnaryMinusExprContext) interface{}

	// Visit a parse tree produced by SQLParser#literalExpr.
	VisitLiteralExpr(ctx *LiteralExprContext) interface{}

	// Visit a parse tree produced by SQLParser#structExpr.
	VisitStructExpr(ctx *StructExprContext) interface{}

	// Visit a parse tree produced by SQLParser#likeExpr.
	VisitLikeExpr(ctx *LikeExprContext) interface{}

	// Visit a parse tree produced by SQLParser#scalarSubqueryExpr.
	VisitScalarSubqueryExpr(ctx *ScalarSubqueryExprContext) interface{}

	// Visit a parse tree produced by SQLParser#funcExpr.
	VisitFuncExpr(ctx *FuncExprContext) interface{}

	// Visit a parse tree produced by SQLParser#addSubExpr.
	VisitAddSubExpr(ctx *AddSubExprContext) interface{}

	// Visit a parse tree produced by SQLParser#arrayExpr.
	VisitArrayExpr(ctx *ArrayExprContext) interface{}

	// Visit a parse tree produced by SQLParser#parameterExpr.
	VisitParameterExpr(ctx *ParameterExprContext) interface{}

	// Visit a parse tree produced by SQLParser#inExpr.
	VisitInExpr(ctx *InExprContext) interface{}

	// Visit a parse tree produced by SQLParser#memberExpr.
	VisitMemberExpr(ctx *MemberExprContext) interface{}

	// Visit a parse tree produced by SQLParser#quantifiedComparisonExpr.
	VisitQuantifiedComparisonExpr(ctx *QuantifiedComparisonExprContext) interface{}

	// Visit a parse tree produced by SQLParser#mapExpr.
	VisitMapExpr(ctx *MapExprContext) interface{}

	// Visit a parse tree produced by SQLParser#orExpr.
	VisitOrExpr(ctx *OrExprContext) interface{}

	// Visit a parse tree produced by SQLParser#comparisonExpr.
	VisitComparisonExpr(ctx *ComparisonExprContext) interface{}

	// Visit a parse tree produced by SQLParser#bitwiseExpr.
	VisitBitwiseExpr(ctx *BitwiseExprContext) interface{}

	// Visit a parse tree produced by SQLParser#notExpr.
	VisitNotExpr(ctx *NotExprContext) interface{}

	// Visit a parse tree produced by SQLParser#isNullExpr.
	VisitIsNullExpr(ctx *IsNullExprContext) interface{}

	// Visit a parse tree produced by SQLParser#unaryPlusExpr.
	VisitUnaryPlusExpr(ctx *UnaryPlusExprContext) interface{}

	// Visit a parse tree produced by SQLParser#caseExpr.
	VisitCaseExpr(ctx *CaseExprContext) interface{}

	// Visit a parse tree produced by SQLParser#systemVariableExpr.
	VisitSystemVariableExpr(ctx *SystemVariableExprContext) interface{}

	// Visit a parse tree produced by SQLParser#intervalExpr.
	VisitIntervalExpr(ctx *IntervalExprContext) interface{}

	// Visit a parse tree produced by SQLParser#mulDivExpr.
	VisitMulDivExpr(ctx *MulDivExprContext) interface{}

	// Visit a parse tree produced by SQLParser#isBooleanExpr.
	VisitIsBooleanExpr(ctx *IsBooleanExprContext) interface{}

	// Visit a parse tree produced by SQLParser#andExpr.
	VisitAndExpr(ctx *AndExprContext) interface{}

	// Visit a parse tree produced by SQLParser#castExpression.
	VisitCastExpression(ctx *CastExpressionContext) interface{}

	// Visit a parse tree produced by SQLParser#extractExpression.
	VisitExtractExpression(ctx *ExtractExpressionContext) interface{}

	// Visit a parse tree produced by SQLParser#intervalExpression.
	VisitIntervalExpression(ctx *IntervalExpressionContext) interface{}

	// Visit a parse tree produced by SQLParser#caseExpression.
	VisitCaseExpression(ctx *CaseExpressionContext) interface{}

	// Visit a parse tree produced by SQLParser#arrayConstructor.
	VisitArrayConstructor(ctx *ArrayConstructorContext) interface{}

	// Visit a parse tree produced by SQLParser#mapConstructor.
	VisitMapConstructor(ctx *MapConstructorContext) interface{}

	// Visit a parse tree produced by SQLParser#structConstructor.
	VisitStructConstructor(ctx *StructConstructorContext) interface{}

	// Visit a parse tree produced by SQLParser#functionCall.
	VisitFunctionCall(ctx *FunctionCallContext) interface{}

	// Visit a parse tree produced by SQLParser#overClause.
	VisitOverClause(ctx *OverClauseContext) interface{}

	// Visit a parse tree produced by SQLParser#partitionByClause.
	VisitPartitionByClause(ctx *PartitionByClauseContext) interface{}

	// Visit a parse tree produced by SQLParser#windowFrame.
	VisitWindowFrame(ctx *WindowFrameContext) interface{}

	// Visit a parse tree produced by SQLParser#windowFrameBound.
	VisitWindowFrameBound(ctx *WindowFrameBoundContext) interface{}

	// Visit a parse tree produced by SQLParser#columnRef.
	VisitColumnRef(ctx *ColumnRefContext) interface{}

	// Visit a parse tree produced by SQLParser#expressionList.
	VisitExpressionList(ctx *ExpressionListContext) interface{}

	// Visit a parse tree produced by SQLParser#tableName.
	VisitTableName(ctx *TableNameContext) interface{}

	// Visit a parse tree produced by SQLParser#databaseName.
	VisitDatabaseName(ctx *DatabaseNameContext) interface{}

	// Visit a parse tree produced by SQLParser#columnName.
	VisitColumnName(ctx *ColumnNameContext) interface{}

	// Visit a parse tree produced by SQLParser#functionName.
	VisitFunctionName(ctx *FunctionNameContext) interface{}

	// Visit a parse tree produced by SQLParser#alias.
	VisitAlias(ctx *AliasContext) interface{}

	// Visit a parse tree produced by SQLParser#identifier.
	VisitIdentifier(ctx *IdentifierContext) interface{}

	// Visit a parse tree produced by SQLParser#nonReservedKeyword.
	VisitNonReservedKeyword(ctx *NonReservedKeywordContext) interface{}

	// Visit a parse tree produced by SQLParser#literal.
	VisitLiteral(ctx *LiteralContext) interface{}
}
