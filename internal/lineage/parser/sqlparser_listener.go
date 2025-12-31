// Code generated from C:/Users/pxl/IdeaProjects/go-metadata/core/lineage/grammar/SQLParser.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // SQLParser
import "github.com/antlr4-go/antlr/v4"

// SQLParserListener is a complete listener for a parse tree produced by SQLParser.
type SQLParserListener interface {
	antlr.ParseTreeListener

	// EnterSqlStatements is called when entering the sqlStatements production.
	EnterSqlStatements(c *SqlStatementsContext)

	// EnterSqlStatement is called when entering the sqlStatement production.
	EnterSqlStatement(c *SqlStatementContext)

	// EnterDmlStatement is called when entering the dmlStatement production.
	EnterDmlStatement(c *DmlStatementContext)

	// EnterDdlStatement is called when entering the ddlStatement production.
	EnterDdlStatement(c *DdlStatementContext)

	// EnterSelectStatement is called when entering the selectStatement production.
	EnterSelectStatement(c *SelectStatementContext)

	// EnterQueryExpression is called when entering the queryExpression production.
	EnterQueryExpression(c *QueryExpressionContext)

	// EnterQueryTerm is called when entering the queryTerm production.
	EnterQueryTerm(c *QueryTermContext)

	// EnterWithClause is called when entering the withClause production.
	EnterWithClause(c *WithClauseContext)

	// EnterCteDefinition is called when entering the cteDefinition production.
	EnterCteDefinition(c *CteDefinitionContext)

	// EnterSelectClause is called when entering the selectClause production.
	EnterSelectClause(c *SelectClauseContext)

	// EnterSelectElements is called when entering the selectElements production.
	EnterSelectElements(c *SelectElementsContext)

	// EnterSelectAll is called when entering the selectAll production.
	EnterSelectAll(c *SelectAllContext)

	// EnterSelectTableAll is called when entering the selectTableAll production.
	EnterSelectTableAll(c *SelectTableAllContext)

	// EnterSelectExpr is called when entering the selectExpr production.
	EnterSelectExpr(c *SelectExprContext)

	// EnterFromClause is called when entering the fromClause production.
	EnterFromClause(c *FromClauseContext)

	// EnterTableReferences is called when entering the tableReferences production.
	EnterTableReferences(c *TableReferencesContext)

	// EnterTableReference is called when entering the tableReference production.
	EnterTableReference(c *TableReferenceContext)

	// EnterTableNameFactor is called when entering the tableNameFactor production.
	EnterTableNameFactor(c *TableNameFactorContext)

	// EnterSubqueryFactor is called when entering the subqueryFactor production.
	EnterSubqueryFactor(c *SubqueryFactorContext)

	// EnterLateralSubqueryFactor is called when entering the lateralSubqueryFactor production.
	EnterLateralSubqueryFactor(c *LateralSubqueryFactorContext)

	// EnterUnnestFactor is called when entering the unnestFactor production.
	EnterUnnestFactor(c *UnnestFactorContext)

	// EnterTableValuedFunctionFactor is called when entering the tableValuedFunctionFactor production.
	EnterTableValuedFunctionFactor(c *TableValuedFunctionFactorContext)

	// EnterTableSample is called when entering the tableSample production.
	EnterTableSample(c *TableSampleContext)

	// EnterJoinPart is called when entering the joinPart production.
	EnterJoinPart(c *JoinPartContext)

	// EnterTemporalJoinClause is called when entering the temporalJoinClause production.
	EnterTemporalJoinClause(c *TemporalJoinClauseContext)

	// EnterJoinType is called when entering the joinType production.
	EnterJoinType(c *JoinTypeContext)

	// EnterWhereClause is called when entering the whereClause production.
	EnterWhereClause(c *WhereClauseContext)

	// EnterGroupByClause is called when entering the groupByClause production.
	EnterGroupByClause(c *GroupByClauseContext)

	// EnterGroupByElements is called when entering the groupByElements production.
	EnterGroupByElements(c *GroupByElementsContext)

	// EnterGroupByElement is called when entering the groupByElement production.
	EnterGroupByElement(c *GroupByElementContext)

	// EnterHavingClause is called when entering the havingClause production.
	EnterHavingClause(c *HavingClauseContext)

	// EnterOrderByClause is called when entering the orderByClause production.
	EnterOrderByClause(c *OrderByClauseContext)

	// EnterOrderByElement is called when entering the orderByElement production.
	EnterOrderByElement(c *OrderByElementContext)

	// EnterLimitClause is called when entering the limitClause production.
	EnterLimitClause(c *LimitClauseContext)

	// EnterInsertStatement is called when entering the insertStatement production.
	EnterInsertStatement(c *InsertStatementContext)

	// EnterPartitionSpec is called when entering the partitionSpec production.
	EnterPartitionSpec(c *PartitionSpecContext)

	// EnterPartitionElement is called when entering the partitionElement production.
	EnterPartitionElement(c *PartitionElementContext)

	// EnterColumnList is called when entering the columnList production.
	EnterColumnList(c *ColumnListContext)

	// EnterValuesClause is called when entering the valuesClause production.
	EnterValuesClause(c *ValuesClauseContext)

	// EnterValueRow is called when entering the valueRow production.
	EnterValueRow(c *ValueRowContext)

	// EnterOnDuplicateKeyUpdate is called when entering the onDuplicateKeyUpdate production.
	EnterOnDuplicateKeyUpdate(c *OnDuplicateKeyUpdateContext)

	// EnterUpdateStatement is called when entering the updateStatement production.
	EnterUpdateStatement(c *UpdateStatementContext)

	// EnterUpdateElement is called when entering the updateElement production.
	EnterUpdateElement(c *UpdateElementContext)

	// EnterDeleteStatement is called when entering the deleteStatement production.
	EnterDeleteStatement(c *DeleteStatementContext)

	// EnterTruncateStatement is called when entering the truncateStatement production.
	EnterTruncateStatement(c *TruncateStatementContext)

	// EnterMergeStatement is called when entering the mergeStatement production.
	EnterMergeStatement(c *MergeStatementContext)

	// EnterMergeClause is called when entering the mergeClause production.
	EnterMergeClause(c *MergeClauseContext)

	// EnterMergeUpdateClause is called when entering the mergeUpdateClause production.
	EnterMergeUpdateClause(c *MergeUpdateClauseContext)

	// EnterMergeInsertClause is called when entering the mergeInsertClause production.
	EnterMergeInsertClause(c *MergeInsertClauseContext)

	// EnterCreateStatement is called when entering the createStatement production.
	EnterCreateStatement(c *CreateStatementContext)

	// EnterCreateTableStatement is called when entering the createTableStatement production.
	EnterCreateTableStatement(c *CreateTableStatementContext)

	// EnterTableElementList is called when entering the tableElementList production.
	EnterTableElementList(c *TableElementListContext)

	// EnterTableElement is called when entering the tableElement production.
	EnterTableElement(c *TableElementContext)

	// EnterWatermarkDefinition is called when entering the watermarkDefinition production.
	EnterWatermarkDefinition(c *WatermarkDefinitionContext)

	// EnterColumnDefinition is called when entering the columnDefinition production.
	EnterColumnDefinition(c *ColumnDefinitionContext)

	// EnterDataType is called when entering the dataType production.
	EnterDataType(c *DataTypeContext)

	// EnterPrimitiveType is called when entering the primitiveType production.
	EnterPrimitiveType(c *PrimitiveTypeContext)

	// EnterStructField is called when entering the structField production.
	EnterStructField(c *StructFieldContext)

	// EnterColumnConstraint is called when entering the columnConstraint production.
	EnterColumnConstraint(c *ColumnConstraintContext)

	// EnterIdentityOptions is called when entering the identityOptions production.
	EnterIdentityOptions(c *IdentityOptionsContext)

	// EnterReferentialAction is called when entering the referentialAction production.
	EnterReferentialAction(c *ReferentialActionContext)

	// EnterReferentialActionType is called when entering the referentialActionType production.
	EnterReferentialActionType(c *ReferentialActionTypeContext)

	// EnterNotEnforced is called when entering the notEnforced production.
	EnterNotEnforced(c *NotEnforcedContext)

	// EnterTableConstraint is called when entering the tableConstraint production.
	EnterTableConstraint(c *TableConstraintContext)

	// EnterPartitionedByClause is called when entering the partitionedByClause production.
	EnterPartitionedByClause(c *PartitionedByClauseContext)

	// EnterClusteredByClause is called when entering the clusteredByClause production.
	EnterClusteredByClause(c *ClusteredByClauseContext)

	// EnterDistributedByClause is called when entering the distributedByClause production.
	EnterDistributedByClause(c *DistributedByClauseContext)

	// EnterSortedByClause is called when entering the sortedByClause production.
	EnterSortedByClause(c *SortedByClauseContext)

	// EnterTableInheritsClause is called when entering the tableInheritsClause production.
	EnterTableInheritsClause(c *TableInheritsClauseContext)

	// EnterEngineClause is called when entering the engineClause production.
	EnterEngineClause(c *EngineClauseContext)

	// EnterCharsetClause is called when entering the charsetClause production.
	EnterCharsetClause(c *CharsetClauseContext)

	// EnterCollateClause is called when entering the collateClause production.
	EnterCollateClause(c *CollateClauseContext)

	// EnterTablespaceClause is called when entering the tablespaceClause production.
	EnterTablespaceClause(c *TablespaceClauseContext)

	// EnterTtlClause is called when entering the ttlClause production.
	EnterTtlClause(c *TtlClauseContext)

	// EnterLifecycleClause is called when entering the lifecycleClause production.
	EnterLifecycleClause(c *LifecycleClauseContext)

	// EnterRowFormatClause is called when entering the rowFormatClause production.
	EnterRowFormatClause(c *RowFormatClauseContext)

	// EnterStoredAsClause is called when entering the storedAsClause production.
	EnterStoredAsClause(c *StoredAsClauseContext)

	// EnterLocationClause is called when entering the locationClause production.
	EnterLocationClause(c *LocationClauseContext)

	// EnterTablePropertiesClause is called when entering the tablePropertiesClause production.
	EnterTablePropertiesClause(c *TablePropertiesClauseContext)

	// EnterWithOptionsClause is called when entering the withOptionsClause production.
	EnterWithOptionsClause(c *WithOptionsClauseContext)

	// EnterPropertyList is called when entering the propertyList production.
	EnterPropertyList(c *PropertyListContext)

	// EnterProperty is called when entering the property production.
	EnterProperty(c *PropertyContext)

	// EnterCreateViewStatement is called when entering the createViewStatement production.
	EnterCreateViewStatement(c *CreateViewStatementContext)

	// EnterCreateDatabaseStatement is called when entering the createDatabaseStatement production.
	EnterCreateDatabaseStatement(c *CreateDatabaseStatementContext)

	// EnterCreateIndexStatement is called when entering the createIndexStatement production.
	EnterCreateIndexStatement(c *CreateIndexStatementContext)

	// EnterIndexColumn is called when entering the indexColumn production.
	EnterIndexColumn(c *IndexColumnContext)

	// EnterDropStatement is called when entering the dropStatement production.
	EnterDropStatement(c *DropStatementContext)

	// EnterAlterStatement is called when entering the alterStatement production.
	EnterAlterStatement(c *AlterStatementContext)

	// EnterAlterTableAction is called when entering the alterTableAction production.
	EnterAlterTableAction(c *AlterTableActionContext)

	// EnterCastExpr is called when entering the castExpr production.
	EnterCastExpr(c *CastExprContext)

	// EnterExtractExpr is called when entering the extractExpr production.
	EnterExtractExpr(c *ExtractExprContext)

	// EnterTypeCastExpr is called when entering the typeCastExpr production.
	EnterTypeCastExpr(c *TypeCastExprContext)

	// EnterExistsExpr is called when entering the existsExpr production.
	EnterExistsExpr(c *ExistsExprContext)

	// EnterBitwiseNotExpr is called when entering the bitwiseNotExpr production.
	EnterBitwiseNotExpr(c *BitwiseNotExprContext)

	// EnterParenExpr is called when entering the parenExpr production.
	EnterParenExpr(c *ParenExprContext)

	// EnterConcatExpr is called when entering the concatExpr production.
	EnterConcatExpr(c *ConcatExprContext)

	// EnterBetweenExpr is called when entering the betweenExpr production.
	EnterBetweenExpr(c *BetweenExprContext)

	// EnterColumnExpr is called when entering the columnExpr production.
	EnterColumnExpr(c *ColumnExprContext)

	// EnterVariableExpr is called when entering the variableExpr production.
	EnterVariableExpr(c *VariableExprContext)

	// EnterArrayAccessExpr is called when entering the arrayAccessExpr production.
	EnterArrayAccessExpr(c *ArrayAccessExprContext)

	// EnterUnaryMinusExpr is called when entering the unaryMinusExpr production.
	EnterUnaryMinusExpr(c *UnaryMinusExprContext)

	// EnterLiteralExpr is called when entering the literalExpr production.
	EnterLiteralExpr(c *LiteralExprContext)

	// EnterStructExpr is called when entering the structExpr production.
	EnterStructExpr(c *StructExprContext)

	// EnterLikeExpr is called when entering the likeExpr production.
	EnterLikeExpr(c *LikeExprContext)

	// EnterScalarSubqueryExpr is called when entering the scalarSubqueryExpr production.
	EnterScalarSubqueryExpr(c *ScalarSubqueryExprContext)

	// EnterFuncExpr is called when entering the funcExpr production.
	EnterFuncExpr(c *FuncExprContext)

	// EnterAddSubExpr is called when entering the addSubExpr production.
	EnterAddSubExpr(c *AddSubExprContext)

	// EnterArrayExpr is called when entering the arrayExpr production.
	EnterArrayExpr(c *ArrayExprContext)

	// EnterParameterExpr is called when entering the parameterExpr production.
	EnterParameterExpr(c *ParameterExprContext)

	// EnterInExpr is called when entering the inExpr production.
	EnterInExpr(c *InExprContext)

	// EnterMemberExpr is called when entering the memberExpr production.
	EnterMemberExpr(c *MemberExprContext)

	// EnterQuantifiedComparisonExpr is called when entering the quantifiedComparisonExpr production.
	EnterQuantifiedComparisonExpr(c *QuantifiedComparisonExprContext)

	// EnterMapExpr is called when entering the mapExpr production.
	EnterMapExpr(c *MapExprContext)

	// EnterOrExpr is called when entering the orExpr production.
	EnterOrExpr(c *OrExprContext)

	// EnterComparisonExpr is called when entering the comparisonExpr production.
	EnterComparisonExpr(c *ComparisonExprContext)

	// EnterBitwiseExpr is called when entering the bitwiseExpr production.
	EnterBitwiseExpr(c *BitwiseExprContext)

	// EnterNotExpr is called when entering the notExpr production.
	EnterNotExpr(c *NotExprContext)

	// EnterIsNullExpr is called when entering the isNullExpr production.
	EnterIsNullExpr(c *IsNullExprContext)

	// EnterUnaryPlusExpr is called when entering the unaryPlusExpr production.
	EnterUnaryPlusExpr(c *UnaryPlusExprContext)

	// EnterCaseExpr is called when entering the caseExpr production.
	EnterCaseExpr(c *CaseExprContext)

	// EnterSystemVariableExpr is called when entering the systemVariableExpr production.
	EnterSystemVariableExpr(c *SystemVariableExprContext)

	// EnterIntervalExpr is called when entering the intervalExpr production.
	EnterIntervalExpr(c *IntervalExprContext)

	// EnterMulDivExpr is called when entering the mulDivExpr production.
	EnterMulDivExpr(c *MulDivExprContext)

	// EnterIsBooleanExpr is called when entering the isBooleanExpr production.
	EnterIsBooleanExpr(c *IsBooleanExprContext)

	// EnterAndExpr is called when entering the andExpr production.
	EnterAndExpr(c *AndExprContext)

	// EnterCastExpression is called when entering the castExpression production.
	EnterCastExpression(c *CastExpressionContext)

	// EnterExtractExpression is called when entering the extractExpression production.
	EnterExtractExpression(c *ExtractExpressionContext)

	// EnterIntervalExpression is called when entering the intervalExpression production.
	EnterIntervalExpression(c *IntervalExpressionContext)

	// EnterCaseExpression is called when entering the caseExpression production.
	EnterCaseExpression(c *CaseExpressionContext)

	// EnterArrayConstructor is called when entering the arrayConstructor production.
	EnterArrayConstructor(c *ArrayConstructorContext)

	// EnterMapConstructor is called when entering the mapConstructor production.
	EnterMapConstructor(c *MapConstructorContext)

	// EnterStructConstructor is called when entering the structConstructor production.
	EnterStructConstructor(c *StructConstructorContext)

	// EnterFunctionCall is called when entering the functionCall production.
	EnterFunctionCall(c *FunctionCallContext)

	// EnterOverClause is called when entering the overClause production.
	EnterOverClause(c *OverClauseContext)

	// EnterPartitionByClause is called when entering the partitionByClause production.
	EnterPartitionByClause(c *PartitionByClauseContext)

	// EnterWindowFrame is called when entering the windowFrame production.
	EnterWindowFrame(c *WindowFrameContext)

	// EnterWindowFrameBound is called when entering the windowFrameBound production.
	EnterWindowFrameBound(c *WindowFrameBoundContext)

	// EnterColumnRef is called when entering the columnRef production.
	EnterColumnRef(c *ColumnRefContext)

	// EnterExpressionList is called when entering the expressionList production.
	EnterExpressionList(c *ExpressionListContext)

	// EnterTableName is called when entering the tableName production.
	EnterTableName(c *TableNameContext)

	// EnterDatabaseName is called when entering the databaseName production.
	EnterDatabaseName(c *DatabaseNameContext)

	// EnterColumnName is called when entering the columnName production.
	EnterColumnName(c *ColumnNameContext)

	// EnterFunctionName is called when entering the functionName production.
	EnterFunctionName(c *FunctionNameContext)

	// EnterAlias is called when entering the alias production.
	EnterAlias(c *AliasContext)

	// EnterIdentifier is called when entering the identifier production.
	EnterIdentifier(c *IdentifierContext)

	// EnterNonReservedKeyword is called when entering the nonReservedKeyword production.
	EnterNonReservedKeyword(c *NonReservedKeywordContext)

	// EnterLiteral is called when entering the literal production.
	EnterLiteral(c *LiteralContext)

	// ExitSqlStatements is called when exiting the sqlStatements production.
	ExitSqlStatements(c *SqlStatementsContext)

	// ExitSqlStatement is called when exiting the sqlStatement production.
	ExitSqlStatement(c *SqlStatementContext)

	// ExitDmlStatement is called when exiting the dmlStatement production.
	ExitDmlStatement(c *DmlStatementContext)

	// ExitDdlStatement is called when exiting the ddlStatement production.
	ExitDdlStatement(c *DdlStatementContext)

	// ExitSelectStatement is called when exiting the selectStatement production.
	ExitSelectStatement(c *SelectStatementContext)

	// ExitQueryExpression is called when exiting the queryExpression production.
	ExitQueryExpression(c *QueryExpressionContext)

	// ExitQueryTerm is called when exiting the queryTerm production.
	ExitQueryTerm(c *QueryTermContext)

	// ExitWithClause is called when exiting the withClause production.
	ExitWithClause(c *WithClauseContext)

	// ExitCteDefinition is called when exiting the cteDefinition production.
	ExitCteDefinition(c *CteDefinitionContext)

	// ExitSelectClause is called when exiting the selectClause production.
	ExitSelectClause(c *SelectClauseContext)

	// ExitSelectElements is called when exiting the selectElements production.
	ExitSelectElements(c *SelectElementsContext)

	// ExitSelectAll is called when exiting the selectAll production.
	ExitSelectAll(c *SelectAllContext)

	// ExitSelectTableAll is called when exiting the selectTableAll production.
	ExitSelectTableAll(c *SelectTableAllContext)

	// ExitSelectExpr is called when exiting the selectExpr production.
	ExitSelectExpr(c *SelectExprContext)

	// ExitFromClause is called when exiting the fromClause production.
	ExitFromClause(c *FromClauseContext)

	// ExitTableReferences is called when exiting the tableReferences production.
	ExitTableReferences(c *TableReferencesContext)

	// ExitTableReference is called when exiting the tableReference production.
	ExitTableReference(c *TableReferenceContext)

	// ExitTableNameFactor is called when exiting the tableNameFactor production.
	ExitTableNameFactor(c *TableNameFactorContext)

	// ExitSubqueryFactor is called when exiting the subqueryFactor production.
	ExitSubqueryFactor(c *SubqueryFactorContext)

	// ExitLateralSubqueryFactor is called when exiting the lateralSubqueryFactor production.
	ExitLateralSubqueryFactor(c *LateralSubqueryFactorContext)

	// ExitUnnestFactor is called when exiting the unnestFactor production.
	ExitUnnestFactor(c *UnnestFactorContext)

	// ExitTableValuedFunctionFactor is called when exiting the tableValuedFunctionFactor production.
	ExitTableValuedFunctionFactor(c *TableValuedFunctionFactorContext)

	// ExitTableSample is called when exiting the tableSample production.
	ExitTableSample(c *TableSampleContext)

	// ExitJoinPart is called when exiting the joinPart production.
	ExitJoinPart(c *JoinPartContext)

	// ExitTemporalJoinClause is called when exiting the temporalJoinClause production.
	ExitTemporalJoinClause(c *TemporalJoinClauseContext)

	// ExitJoinType is called when exiting the joinType production.
	ExitJoinType(c *JoinTypeContext)

	// ExitWhereClause is called when exiting the whereClause production.
	ExitWhereClause(c *WhereClauseContext)

	// ExitGroupByClause is called when exiting the groupByClause production.
	ExitGroupByClause(c *GroupByClauseContext)

	// ExitGroupByElements is called when exiting the groupByElements production.
	ExitGroupByElements(c *GroupByElementsContext)

	// ExitGroupByElement is called when exiting the groupByElement production.
	ExitGroupByElement(c *GroupByElementContext)

	// ExitHavingClause is called when exiting the havingClause production.
	ExitHavingClause(c *HavingClauseContext)

	// ExitOrderByClause is called when exiting the orderByClause production.
	ExitOrderByClause(c *OrderByClauseContext)

	// ExitOrderByElement is called when exiting the orderByElement production.
	ExitOrderByElement(c *OrderByElementContext)

	// ExitLimitClause is called when exiting the limitClause production.
	ExitLimitClause(c *LimitClauseContext)

	// ExitInsertStatement is called when exiting the insertStatement production.
	ExitInsertStatement(c *InsertStatementContext)

	// ExitPartitionSpec is called when exiting the partitionSpec production.
	ExitPartitionSpec(c *PartitionSpecContext)

	// ExitPartitionElement is called when exiting the partitionElement production.
	ExitPartitionElement(c *PartitionElementContext)

	// ExitColumnList is called when exiting the columnList production.
	ExitColumnList(c *ColumnListContext)

	// ExitValuesClause is called when exiting the valuesClause production.
	ExitValuesClause(c *ValuesClauseContext)

	// ExitValueRow is called when exiting the valueRow production.
	ExitValueRow(c *ValueRowContext)

	// ExitOnDuplicateKeyUpdate is called when exiting the onDuplicateKeyUpdate production.
	ExitOnDuplicateKeyUpdate(c *OnDuplicateKeyUpdateContext)

	// ExitUpdateStatement is called when exiting the updateStatement production.
	ExitUpdateStatement(c *UpdateStatementContext)

	// ExitUpdateElement is called when exiting the updateElement production.
	ExitUpdateElement(c *UpdateElementContext)

	// ExitDeleteStatement is called when exiting the deleteStatement production.
	ExitDeleteStatement(c *DeleteStatementContext)

	// ExitTruncateStatement is called when exiting the truncateStatement production.
	ExitTruncateStatement(c *TruncateStatementContext)

	// ExitMergeStatement is called when exiting the mergeStatement production.
	ExitMergeStatement(c *MergeStatementContext)

	// ExitMergeClause is called when exiting the mergeClause production.
	ExitMergeClause(c *MergeClauseContext)

	// ExitMergeUpdateClause is called when exiting the mergeUpdateClause production.
	ExitMergeUpdateClause(c *MergeUpdateClauseContext)

	// ExitMergeInsertClause is called when exiting the mergeInsertClause production.
	ExitMergeInsertClause(c *MergeInsertClauseContext)

	// ExitCreateStatement is called when exiting the createStatement production.
	ExitCreateStatement(c *CreateStatementContext)

	// ExitCreateTableStatement is called when exiting the createTableStatement production.
	ExitCreateTableStatement(c *CreateTableStatementContext)

	// ExitTableElementList is called when exiting the tableElementList production.
	ExitTableElementList(c *TableElementListContext)

	// ExitTableElement is called when exiting the tableElement production.
	ExitTableElement(c *TableElementContext)

	// ExitWatermarkDefinition is called when exiting the watermarkDefinition production.
	ExitWatermarkDefinition(c *WatermarkDefinitionContext)

	// ExitColumnDefinition is called when exiting the columnDefinition production.
	ExitColumnDefinition(c *ColumnDefinitionContext)

	// ExitDataType is called when exiting the dataType production.
	ExitDataType(c *DataTypeContext)

	// ExitPrimitiveType is called when exiting the primitiveType production.
	ExitPrimitiveType(c *PrimitiveTypeContext)

	// ExitStructField is called when exiting the structField production.
	ExitStructField(c *StructFieldContext)

	// ExitColumnConstraint is called when exiting the columnConstraint production.
	ExitColumnConstraint(c *ColumnConstraintContext)

	// ExitIdentityOptions is called when exiting the identityOptions production.
	ExitIdentityOptions(c *IdentityOptionsContext)

	// ExitReferentialAction is called when exiting the referentialAction production.
	ExitReferentialAction(c *ReferentialActionContext)

	// ExitReferentialActionType is called when exiting the referentialActionType production.
	ExitReferentialActionType(c *ReferentialActionTypeContext)

	// ExitNotEnforced is called when exiting the notEnforced production.
	ExitNotEnforced(c *NotEnforcedContext)

	// ExitTableConstraint is called when exiting the tableConstraint production.
	ExitTableConstraint(c *TableConstraintContext)

	// ExitPartitionedByClause is called when exiting the partitionedByClause production.
	ExitPartitionedByClause(c *PartitionedByClauseContext)

	// ExitClusteredByClause is called when exiting the clusteredByClause production.
	ExitClusteredByClause(c *ClusteredByClauseContext)

	// ExitDistributedByClause is called when exiting the distributedByClause production.
	ExitDistributedByClause(c *DistributedByClauseContext)

	// ExitSortedByClause is called when exiting the sortedByClause production.
	ExitSortedByClause(c *SortedByClauseContext)

	// ExitTableInheritsClause is called when exiting the tableInheritsClause production.
	ExitTableInheritsClause(c *TableInheritsClauseContext)

	// ExitEngineClause is called when exiting the engineClause production.
	ExitEngineClause(c *EngineClauseContext)

	// ExitCharsetClause is called when exiting the charsetClause production.
	ExitCharsetClause(c *CharsetClauseContext)

	// ExitCollateClause is called when exiting the collateClause production.
	ExitCollateClause(c *CollateClauseContext)

	// ExitTablespaceClause is called when exiting the tablespaceClause production.
	ExitTablespaceClause(c *TablespaceClauseContext)

	// ExitTtlClause is called when exiting the ttlClause production.
	ExitTtlClause(c *TtlClauseContext)

	// ExitLifecycleClause is called when exiting the lifecycleClause production.
	ExitLifecycleClause(c *LifecycleClauseContext)

	// ExitRowFormatClause is called when exiting the rowFormatClause production.
	ExitRowFormatClause(c *RowFormatClauseContext)

	// ExitStoredAsClause is called when exiting the storedAsClause production.
	ExitStoredAsClause(c *StoredAsClauseContext)

	// ExitLocationClause is called when exiting the locationClause production.
	ExitLocationClause(c *LocationClauseContext)

	// ExitTablePropertiesClause is called when exiting the tablePropertiesClause production.
	ExitTablePropertiesClause(c *TablePropertiesClauseContext)

	// ExitWithOptionsClause is called when exiting the withOptionsClause production.
	ExitWithOptionsClause(c *WithOptionsClauseContext)

	// ExitPropertyList is called when exiting the propertyList production.
	ExitPropertyList(c *PropertyListContext)

	// ExitProperty is called when exiting the property production.
	ExitProperty(c *PropertyContext)

	// ExitCreateViewStatement is called when exiting the createViewStatement production.
	ExitCreateViewStatement(c *CreateViewStatementContext)

	// ExitCreateDatabaseStatement is called when exiting the createDatabaseStatement production.
	ExitCreateDatabaseStatement(c *CreateDatabaseStatementContext)

	// ExitCreateIndexStatement is called when exiting the createIndexStatement production.
	ExitCreateIndexStatement(c *CreateIndexStatementContext)

	// ExitIndexColumn is called when exiting the indexColumn production.
	ExitIndexColumn(c *IndexColumnContext)

	// ExitDropStatement is called when exiting the dropStatement production.
	ExitDropStatement(c *DropStatementContext)

	// ExitAlterStatement is called when exiting the alterStatement production.
	ExitAlterStatement(c *AlterStatementContext)

	// ExitAlterTableAction is called when exiting the alterTableAction production.
	ExitAlterTableAction(c *AlterTableActionContext)

	// ExitCastExpr is called when exiting the castExpr production.
	ExitCastExpr(c *CastExprContext)

	// ExitExtractExpr is called when exiting the extractExpr production.
	ExitExtractExpr(c *ExtractExprContext)

	// ExitTypeCastExpr is called when exiting the typeCastExpr production.
	ExitTypeCastExpr(c *TypeCastExprContext)

	// ExitExistsExpr is called when exiting the existsExpr production.
	ExitExistsExpr(c *ExistsExprContext)

	// ExitBitwiseNotExpr is called when exiting the bitwiseNotExpr production.
	ExitBitwiseNotExpr(c *BitwiseNotExprContext)

	// ExitParenExpr is called when exiting the parenExpr production.
	ExitParenExpr(c *ParenExprContext)

	// ExitConcatExpr is called when exiting the concatExpr production.
	ExitConcatExpr(c *ConcatExprContext)

	// ExitBetweenExpr is called when exiting the betweenExpr production.
	ExitBetweenExpr(c *BetweenExprContext)

	// ExitColumnExpr is called when exiting the columnExpr production.
	ExitColumnExpr(c *ColumnExprContext)

	// ExitVariableExpr is called when exiting the variableExpr production.
	ExitVariableExpr(c *VariableExprContext)

	// ExitArrayAccessExpr is called when exiting the arrayAccessExpr production.
	ExitArrayAccessExpr(c *ArrayAccessExprContext)

	// ExitUnaryMinusExpr is called when exiting the unaryMinusExpr production.
	ExitUnaryMinusExpr(c *UnaryMinusExprContext)

	// ExitLiteralExpr is called when exiting the literalExpr production.
	ExitLiteralExpr(c *LiteralExprContext)

	// ExitStructExpr is called when exiting the structExpr production.
	ExitStructExpr(c *StructExprContext)

	// ExitLikeExpr is called when exiting the likeExpr production.
	ExitLikeExpr(c *LikeExprContext)

	// ExitScalarSubqueryExpr is called when exiting the scalarSubqueryExpr production.
	ExitScalarSubqueryExpr(c *ScalarSubqueryExprContext)

	// ExitFuncExpr is called when exiting the funcExpr production.
	ExitFuncExpr(c *FuncExprContext)

	// ExitAddSubExpr is called when exiting the addSubExpr production.
	ExitAddSubExpr(c *AddSubExprContext)

	// ExitArrayExpr is called when exiting the arrayExpr production.
	ExitArrayExpr(c *ArrayExprContext)

	// ExitParameterExpr is called when exiting the parameterExpr production.
	ExitParameterExpr(c *ParameterExprContext)

	// ExitInExpr is called when exiting the inExpr production.
	ExitInExpr(c *InExprContext)

	// ExitMemberExpr is called when exiting the memberExpr production.
	ExitMemberExpr(c *MemberExprContext)

	// ExitQuantifiedComparisonExpr is called when exiting the quantifiedComparisonExpr production.
	ExitQuantifiedComparisonExpr(c *QuantifiedComparisonExprContext)

	// ExitMapExpr is called when exiting the mapExpr production.
	ExitMapExpr(c *MapExprContext)

	// ExitOrExpr is called when exiting the orExpr production.
	ExitOrExpr(c *OrExprContext)

	// ExitComparisonExpr is called when exiting the comparisonExpr production.
	ExitComparisonExpr(c *ComparisonExprContext)

	// ExitBitwiseExpr is called when exiting the bitwiseExpr production.
	ExitBitwiseExpr(c *BitwiseExprContext)

	// ExitNotExpr is called when exiting the notExpr production.
	ExitNotExpr(c *NotExprContext)

	// ExitIsNullExpr is called when exiting the isNullExpr production.
	ExitIsNullExpr(c *IsNullExprContext)

	// ExitUnaryPlusExpr is called when exiting the unaryPlusExpr production.
	ExitUnaryPlusExpr(c *UnaryPlusExprContext)

	// ExitCaseExpr is called when exiting the caseExpr production.
	ExitCaseExpr(c *CaseExprContext)

	// ExitSystemVariableExpr is called when exiting the systemVariableExpr production.
	ExitSystemVariableExpr(c *SystemVariableExprContext)

	// ExitIntervalExpr is called when exiting the intervalExpr production.
	ExitIntervalExpr(c *IntervalExprContext)

	// ExitMulDivExpr is called when exiting the mulDivExpr production.
	ExitMulDivExpr(c *MulDivExprContext)

	// ExitIsBooleanExpr is called when exiting the isBooleanExpr production.
	ExitIsBooleanExpr(c *IsBooleanExprContext)

	// ExitAndExpr is called when exiting the andExpr production.
	ExitAndExpr(c *AndExprContext)

	// ExitCastExpression is called when exiting the castExpression production.
	ExitCastExpression(c *CastExpressionContext)

	// ExitExtractExpression is called when exiting the extractExpression production.
	ExitExtractExpression(c *ExtractExpressionContext)

	// ExitIntervalExpression is called when exiting the intervalExpression production.
	ExitIntervalExpression(c *IntervalExpressionContext)

	// ExitCaseExpression is called when exiting the caseExpression production.
	ExitCaseExpression(c *CaseExpressionContext)

	// ExitArrayConstructor is called when exiting the arrayConstructor production.
	ExitArrayConstructor(c *ArrayConstructorContext)

	// ExitMapConstructor is called when exiting the mapConstructor production.
	ExitMapConstructor(c *MapConstructorContext)

	// ExitStructConstructor is called when exiting the structConstructor production.
	ExitStructConstructor(c *StructConstructorContext)

	// ExitFunctionCall is called when exiting the functionCall production.
	ExitFunctionCall(c *FunctionCallContext)

	// ExitOverClause is called when exiting the overClause production.
	ExitOverClause(c *OverClauseContext)

	// ExitPartitionByClause is called when exiting the partitionByClause production.
	ExitPartitionByClause(c *PartitionByClauseContext)

	// ExitWindowFrame is called when exiting the windowFrame production.
	ExitWindowFrame(c *WindowFrameContext)

	// ExitWindowFrameBound is called when exiting the windowFrameBound production.
	ExitWindowFrameBound(c *WindowFrameBoundContext)

	// ExitColumnRef is called when exiting the columnRef production.
	ExitColumnRef(c *ColumnRefContext)

	// ExitExpressionList is called when exiting the expressionList production.
	ExitExpressionList(c *ExpressionListContext)

	// ExitTableName is called when exiting the tableName production.
	ExitTableName(c *TableNameContext)

	// ExitDatabaseName is called when exiting the databaseName production.
	ExitDatabaseName(c *DatabaseNameContext)

	// ExitColumnName is called when exiting the columnName production.
	ExitColumnName(c *ColumnNameContext)

	// ExitFunctionName is called when exiting the functionName production.
	ExitFunctionName(c *FunctionNameContext)

	// ExitAlias is called when exiting the alias production.
	ExitAlias(c *AliasContext)

	// ExitIdentifier is called when exiting the identifier production.
	ExitIdentifier(c *IdentifierContext)

	// ExitNonReservedKeyword is called when exiting the nonReservedKeyword production.
	ExitNonReservedKeyword(c *NonReservedKeywordContext)

	// ExitLiteral is called when exiting the literal production.
	ExitLiteral(c *LiteralContext)
}
