// Code generated from C:/Users/pxl/IdeaProjects/go-metadata/core/lineage/grammar/SQLParser.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // SQLParser
import "github.com/antlr4-go/antlr/v4"

type BaseSQLParserVisitor struct {
	*antlr.BaseParseTreeVisitor
}

func (v *BaseSQLParserVisitor) VisitSqlStatements(ctx *SqlStatementsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitSqlStatement(ctx *SqlStatementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitDmlStatement(ctx *DmlStatementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitDdlStatement(ctx *DdlStatementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitSelectStatement(ctx *SelectStatementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitQueryExpression(ctx *QueryExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitQueryTerm(ctx *QueryTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitWithClause(ctx *WithClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitCteDefinition(ctx *CteDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitSelectClause(ctx *SelectClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitSelectElements(ctx *SelectElementsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitSelectAll(ctx *SelectAllContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitSelectTableAll(ctx *SelectTableAllContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitSelectExpr(ctx *SelectExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitFromClause(ctx *FromClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitTableReferences(ctx *TableReferencesContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitTableReference(ctx *TableReferenceContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitTableNameFactor(ctx *TableNameFactorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitSubqueryFactor(ctx *SubqueryFactorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitLateralSubqueryFactor(ctx *LateralSubqueryFactorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitUnnestFactor(ctx *UnnestFactorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitTableValuedFunctionFactor(ctx *TableValuedFunctionFactorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitTableSample(ctx *TableSampleContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitJoinPart(ctx *JoinPartContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitTemporalJoinClause(ctx *TemporalJoinClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitJoinType(ctx *JoinTypeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitWhereClause(ctx *WhereClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitGroupByClause(ctx *GroupByClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitGroupByElements(ctx *GroupByElementsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitGroupByElement(ctx *GroupByElementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitHavingClause(ctx *HavingClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitOrderByClause(ctx *OrderByClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitOrderByElement(ctx *OrderByElementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitLimitClause(ctx *LimitClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitInsertStatement(ctx *InsertStatementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitPartitionSpec(ctx *PartitionSpecContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitPartitionElement(ctx *PartitionElementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitColumnList(ctx *ColumnListContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitValuesClause(ctx *ValuesClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitValueRow(ctx *ValueRowContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitOnDuplicateKeyUpdate(ctx *OnDuplicateKeyUpdateContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitUpdateStatement(ctx *UpdateStatementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitUpdateElement(ctx *UpdateElementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitDeleteStatement(ctx *DeleteStatementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitTruncateStatement(ctx *TruncateStatementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitMergeStatement(ctx *MergeStatementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitMergeClause(ctx *MergeClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitMergeUpdateClause(ctx *MergeUpdateClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitMergeInsertClause(ctx *MergeInsertClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitCreateStatement(ctx *CreateStatementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitCreateTableStatement(ctx *CreateTableStatementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitTableElementList(ctx *TableElementListContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitTableElement(ctx *TableElementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitWatermarkDefinition(ctx *WatermarkDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitColumnDefinition(ctx *ColumnDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitDataType(ctx *DataTypeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitPrimitiveType(ctx *PrimitiveTypeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitStructField(ctx *StructFieldContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitColumnConstraint(ctx *ColumnConstraintContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitIdentityOptions(ctx *IdentityOptionsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitReferentialAction(ctx *ReferentialActionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitReferentialActionType(ctx *ReferentialActionTypeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitNotEnforced(ctx *NotEnforcedContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitTableConstraint(ctx *TableConstraintContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitPartitionedByClause(ctx *PartitionedByClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitClusteredByClause(ctx *ClusteredByClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitDistributedByClause(ctx *DistributedByClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitSortedByClause(ctx *SortedByClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitTableInheritsClause(ctx *TableInheritsClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitEngineClause(ctx *EngineClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitCharsetClause(ctx *CharsetClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitCollateClause(ctx *CollateClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitTablespaceClause(ctx *TablespaceClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitTtlClause(ctx *TtlClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitLifecycleClause(ctx *LifecycleClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitRowFormatClause(ctx *RowFormatClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitStoredAsClause(ctx *StoredAsClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitLocationClause(ctx *LocationClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitTablePropertiesClause(ctx *TablePropertiesClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitWithOptionsClause(ctx *WithOptionsClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitPropertyList(ctx *PropertyListContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitProperty(ctx *PropertyContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitCreateViewStatement(ctx *CreateViewStatementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitCreateDatabaseStatement(ctx *CreateDatabaseStatementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitCreateIndexStatement(ctx *CreateIndexStatementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitIndexColumn(ctx *IndexColumnContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitDropStatement(ctx *DropStatementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitAlterStatement(ctx *AlterStatementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitAlterTableAction(ctx *AlterTableActionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitCastExpr(ctx *CastExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitExtractExpr(ctx *ExtractExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitTypeCastExpr(ctx *TypeCastExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitExistsExpr(ctx *ExistsExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitBitwiseNotExpr(ctx *BitwiseNotExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitParenExpr(ctx *ParenExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitConcatExpr(ctx *ConcatExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitBetweenExpr(ctx *BetweenExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitColumnExpr(ctx *ColumnExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitVariableExpr(ctx *VariableExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitArrayAccessExpr(ctx *ArrayAccessExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitUnaryMinusExpr(ctx *UnaryMinusExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitLiteralExpr(ctx *LiteralExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitStructExpr(ctx *StructExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitLikeExpr(ctx *LikeExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitScalarSubqueryExpr(ctx *ScalarSubqueryExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitFuncExpr(ctx *FuncExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitAddSubExpr(ctx *AddSubExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitArrayExpr(ctx *ArrayExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitParameterExpr(ctx *ParameterExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitInExpr(ctx *InExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitMemberExpr(ctx *MemberExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitQuantifiedComparisonExpr(ctx *QuantifiedComparisonExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitMapExpr(ctx *MapExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitOrExpr(ctx *OrExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitComparisonExpr(ctx *ComparisonExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitBitwiseExpr(ctx *BitwiseExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitNotExpr(ctx *NotExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitIsNullExpr(ctx *IsNullExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitUnaryPlusExpr(ctx *UnaryPlusExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitCaseExpr(ctx *CaseExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitSystemVariableExpr(ctx *SystemVariableExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitIntervalExpr(ctx *IntervalExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitMulDivExpr(ctx *MulDivExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitIsBooleanExpr(ctx *IsBooleanExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitAndExpr(ctx *AndExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitCastExpression(ctx *CastExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitExtractExpression(ctx *ExtractExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitIntervalExpression(ctx *IntervalExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitCaseExpression(ctx *CaseExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitArrayConstructor(ctx *ArrayConstructorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitMapConstructor(ctx *MapConstructorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitStructConstructor(ctx *StructConstructorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitFunctionCall(ctx *FunctionCallContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitOverClause(ctx *OverClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitPartitionByClause(ctx *PartitionByClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitWindowFrame(ctx *WindowFrameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitWindowFrameBound(ctx *WindowFrameBoundContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitColumnRef(ctx *ColumnRefContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitExpressionList(ctx *ExpressionListContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitTableName(ctx *TableNameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitDatabaseName(ctx *DatabaseNameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitColumnName(ctx *ColumnNameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitFunctionName(ctx *FunctionNameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitAlias(ctx *AliasContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitIdentifier(ctx *IdentifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitNonReservedKeyword(ctx *NonReservedKeywordContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitLiteral(ctx *LiteralContext) interface{} {
	return v.VisitChildren(ctx)
}
