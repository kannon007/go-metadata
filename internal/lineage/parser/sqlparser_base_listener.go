// Code generated from C:/Users/pxl/IdeaProjects/go-metadata/core/lineage/grammar/SQLParser.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // SQLParser
import "github.com/antlr4-go/antlr/v4"

// BaseSQLParserListener is a complete listener for a parse tree produced by SQLParser.
type BaseSQLParserListener struct{}

var _ SQLParserListener = &BaseSQLParserListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseSQLParserListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseSQLParserListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseSQLParserListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseSQLParserListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterSqlStatements is called when production sqlStatements is entered.
func (s *BaseSQLParserListener) EnterSqlStatements(ctx *SqlStatementsContext) {}

// ExitSqlStatements is called when production sqlStatements is exited.
func (s *BaseSQLParserListener) ExitSqlStatements(ctx *SqlStatementsContext) {}

// EnterSqlStatement is called when production sqlStatement is entered.
func (s *BaseSQLParserListener) EnterSqlStatement(ctx *SqlStatementContext) {}

// ExitSqlStatement is called when production sqlStatement is exited.
func (s *BaseSQLParserListener) ExitSqlStatement(ctx *SqlStatementContext) {}

// EnterDmlStatement is called when production dmlStatement is entered.
func (s *BaseSQLParserListener) EnterDmlStatement(ctx *DmlStatementContext) {}

// ExitDmlStatement is called when production dmlStatement is exited.
func (s *BaseSQLParserListener) ExitDmlStatement(ctx *DmlStatementContext) {}

// EnterDdlStatement is called when production ddlStatement is entered.
func (s *BaseSQLParserListener) EnterDdlStatement(ctx *DdlStatementContext) {}

// ExitDdlStatement is called when production ddlStatement is exited.
func (s *BaseSQLParserListener) ExitDdlStatement(ctx *DdlStatementContext) {}

// EnterSelectStatement is called when production selectStatement is entered.
func (s *BaseSQLParserListener) EnterSelectStatement(ctx *SelectStatementContext) {}

// ExitSelectStatement is called when production selectStatement is exited.
func (s *BaseSQLParserListener) ExitSelectStatement(ctx *SelectStatementContext) {}

// EnterQueryExpression is called when production queryExpression is entered.
func (s *BaseSQLParserListener) EnterQueryExpression(ctx *QueryExpressionContext) {}

// ExitQueryExpression is called when production queryExpression is exited.
func (s *BaseSQLParserListener) ExitQueryExpression(ctx *QueryExpressionContext) {}

// EnterQueryTerm is called when production queryTerm is entered.
func (s *BaseSQLParserListener) EnterQueryTerm(ctx *QueryTermContext) {}

// ExitQueryTerm is called when production queryTerm is exited.
func (s *BaseSQLParserListener) ExitQueryTerm(ctx *QueryTermContext) {}

// EnterWithClause is called when production withClause is entered.
func (s *BaseSQLParserListener) EnterWithClause(ctx *WithClauseContext) {}

// ExitWithClause is called when production withClause is exited.
func (s *BaseSQLParserListener) ExitWithClause(ctx *WithClauseContext) {}

// EnterCteDefinition is called when production cteDefinition is entered.
func (s *BaseSQLParserListener) EnterCteDefinition(ctx *CteDefinitionContext) {}

// ExitCteDefinition is called when production cteDefinition is exited.
func (s *BaseSQLParserListener) ExitCteDefinition(ctx *CteDefinitionContext) {}

// EnterSelectClause is called when production selectClause is entered.
func (s *BaseSQLParserListener) EnterSelectClause(ctx *SelectClauseContext) {}

// ExitSelectClause is called when production selectClause is exited.
func (s *BaseSQLParserListener) ExitSelectClause(ctx *SelectClauseContext) {}

// EnterSelectElements is called when production selectElements is entered.
func (s *BaseSQLParserListener) EnterSelectElements(ctx *SelectElementsContext) {}

// ExitSelectElements is called when production selectElements is exited.
func (s *BaseSQLParserListener) ExitSelectElements(ctx *SelectElementsContext) {}

// EnterSelectAll is called when production selectAll is entered.
func (s *BaseSQLParserListener) EnterSelectAll(ctx *SelectAllContext) {}

// ExitSelectAll is called when production selectAll is exited.
func (s *BaseSQLParserListener) ExitSelectAll(ctx *SelectAllContext) {}

// EnterSelectTableAll is called when production selectTableAll is entered.
func (s *BaseSQLParserListener) EnterSelectTableAll(ctx *SelectTableAllContext) {}

// ExitSelectTableAll is called when production selectTableAll is exited.
func (s *BaseSQLParserListener) ExitSelectTableAll(ctx *SelectTableAllContext) {}

// EnterSelectExpr is called when production selectExpr is entered.
func (s *BaseSQLParserListener) EnterSelectExpr(ctx *SelectExprContext) {}

// ExitSelectExpr is called when production selectExpr is exited.
func (s *BaseSQLParserListener) ExitSelectExpr(ctx *SelectExprContext) {}

// EnterFromClause is called when production fromClause is entered.
func (s *BaseSQLParserListener) EnterFromClause(ctx *FromClauseContext) {}

// ExitFromClause is called when production fromClause is exited.
func (s *BaseSQLParserListener) ExitFromClause(ctx *FromClauseContext) {}

// EnterTableReferences is called when production tableReferences is entered.
func (s *BaseSQLParserListener) EnterTableReferences(ctx *TableReferencesContext) {}

// ExitTableReferences is called when production tableReferences is exited.
func (s *BaseSQLParserListener) ExitTableReferences(ctx *TableReferencesContext) {}

// EnterTableReference is called when production tableReference is entered.
func (s *BaseSQLParserListener) EnterTableReference(ctx *TableReferenceContext) {}

// ExitTableReference is called when production tableReference is exited.
func (s *BaseSQLParserListener) ExitTableReference(ctx *TableReferenceContext) {}

// EnterTableNameFactor is called when production tableNameFactor is entered.
func (s *BaseSQLParserListener) EnterTableNameFactor(ctx *TableNameFactorContext) {}

// ExitTableNameFactor is called when production tableNameFactor is exited.
func (s *BaseSQLParserListener) ExitTableNameFactor(ctx *TableNameFactorContext) {}

// EnterSubqueryFactor is called when production subqueryFactor is entered.
func (s *BaseSQLParserListener) EnterSubqueryFactor(ctx *SubqueryFactorContext) {}

// ExitSubqueryFactor is called when production subqueryFactor is exited.
func (s *BaseSQLParserListener) ExitSubqueryFactor(ctx *SubqueryFactorContext) {}

// EnterLateralSubqueryFactor is called when production lateralSubqueryFactor is entered.
func (s *BaseSQLParserListener) EnterLateralSubqueryFactor(ctx *LateralSubqueryFactorContext) {}

// ExitLateralSubqueryFactor is called when production lateralSubqueryFactor is exited.
func (s *BaseSQLParserListener) ExitLateralSubqueryFactor(ctx *LateralSubqueryFactorContext) {}

// EnterUnnestFactor is called when production unnestFactor is entered.
func (s *BaseSQLParserListener) EnterUnnestFactor(ctx *UnnestFactorContext) {}

// ExitUnnestFactor is called when production unnestFactor is exited.
func (s *BaseSQLParserListener) ExitUnnestFactor(ctx *UnnestFactorContext) {}

// EnterTableValuedFunctionFactor is called when production tableValuedFunctionFactor is entered.
func (s *BaseSQLParserListener) EnterTableValuedFunctionFactor(ctx *TableValuedFunctionFactorContext) {
}

// ExitTableValuedFunctionFactor is called when production tableValuedFunctionFactor is exited.
func (s *BaseSQLParserListener) ExitTableValuedFunctionFactor(ctx *TableValuedFunctionFactorContext) {
}

// EnterTableSample is called when production tableSample is entered.
func (s *BaseSQLParserListener) EnterTableSample(ctx *TableSampleContext) {}

// ExitTableSample is called when production tableSample is exited.
func (s *BaseSQLParserListener) ExitTableSample(ctx *TableSampleContext) {}

// EnterJoinPart is called when production joinPart is entered.
func (s *BaseSQLParserListener) EnterJoinPart(ctx *JoinPartContext) {}

// ExitJoinPart is called when production joinPart is exited.
func (s *BaseSQLParserListener) ExitJoinPart(ctx *JoinPartContext) {}

// EnterTemporalJoinClause is called when production temporalJoinClause is entered.
func (s *BaseSQLParserListener) EnterTemporalJoinClause(ctx *TemporalJoinClauseContext) {}

// ExitTemporalJoinClause is called when production temporalJoinClause is exited.
func (s *BaseSQLParserListener) ExitTemporalJoinClause(ctx *TemporalJoinClauseContext) {}

// EnterJoinType is called when production joinType is entered.
func (s *BaseSQLParserListener) EnterJoinType(ctx *JoinTypeContext) {}

// ExitJoinType is called when production joinType is exited.
func (s *BaseSQLParserListener) ExitJoinType(ctx *JoinTypeContext) {}

// EnterWhereClause is called when production whereClause is entered.
func (s *BaseSQLParserListener) EnterWhereClause(ctx *WhereClauseContext) {}

// ExitWhereClause is called when production whereClause is exited.
func (s *BaseSQLParserListener) ExitWhereClause(ctx *WhereClauseContext) {}

// EnterGroupByClause is called when production groupByClause is entered.
func (s *BaseSQLParserListener) EnterGroupByClause(ctx *GroupByClauseContext) {}

// ExitGroupByClause is called when production groupByClause is exited.
func (s *BaseSQLParserListener) ExitGroupByClause(ctx *GroupByClauseContext) {}

// EnterGroupByElements is called when production groupByElements is entered.
func (s *BaseSQLParserListener) EnterGroupByElements(ctx *GroupByElementsContext) {}

// ExitGroupByElements is called when production groupByElements is exited.
func (s *BaseSQLParserListener) ExitGroupByElements(ctx *GroupByElementsContext) {}

// EnterGroupByElement is called when production groupByElement is entered.
func (s *BaseSQLParserListener) EnterGroupByElement(ctx *GroupByElementContext) {}

// ExitGroupByElement is called when production groupByElement is exited.
func (s *BaseSQLParserListener) ExitGroupByElement(ctx *GroupByElementContext) {}

// EnterHavingClause is called when production havingClause is entered.
func (s *BaseSQLParserListener) EnterHavingClause(ctx *HavingClauseContext) {}

// ExitHavingClause is called when production havingClause is exited.
func (s *BaseSQLParserListener) ExitHavingClause(ctx *HavingClauseContext) {}

// EnterOrderByClause is called when production orderByClause is entered.
func (s *BaseSQLParserListener) EnterOrderByClause(ctx *OrderByClauseContext) {}

// ExitOrderByClause is called when production orderByClause is exited.
func (s *BaseSQLParserListener) ExitOrderByClause(ctx *OrderByClauseContext) {}

// EnterOrderByElement is called when production orderByElement is entered.
func (s *BaseSQLParserListener) EnterOrderByElement(ctx *OrderByElementContext) {}

// ExitOrderByElement is called when production orderByElement is exited.
func (s *BaseSQLParserListener) ExitOrderByElement(ctx *OrderByElementContext) {}

// EnterLimitClause is called when production limitClause is entered.
func (s *BaseSQLParserListener) EnterLimitClause(ctx *LimitClauseContext) {}

// ExitLimitClause is called when production limitClause is exited.
func (s *BaseSQLParserListener) ExitLimitClause(ctx *LimitClauseContext) {}

// EnterInsertStatement is called when production insertStatement is entered.
func (s *BaseSQLParserListener) EnterInsertStatement(ctx *InsertStatementContext) {}

// ExitInsertStatement is called when production insertStatement is exited.
func (s *BaseSQLParserListener) ExitInsertStatement(ctx *InsertStatementContext) {}

// EnterPartitionSpec is called when production partitionSpec is entered.
func (s *BaseSQLParserListener) EnterPartitionSpec(ctx *PartitionSpecContext) {}

// ExitPartitionSpec is called when production partitionSpec is exited.
func (s *BaseSQLParserListener) ExitPartitionSpec(ctx *PartitionSpecContext) {}

// EnterPartitionElement is called when production partitionElement is entered.
func (s *BaseSQLParserListener) EnterPartitionElement(ctx *PartitionElementContext) {}

// ExitPartitionElement is called when production partitionElement is exited.
func (s *BaseSQLParserListener) ExitPartitionElement(ctx *PartitionElementContext) {}

// EnterColumnList is called when production columnList is entered.
func (s *BaseSQLParserListener) EnterColumnList(ctx *ColumnListContext) {}

// ExitColumnList is called when production columnList is exited.
func (s *BaseSQLParserListener) ExitColumnList(ctx *ColumnListContext) {}

// EnterValuesClause is called when production valuesClause is entered.
func (s *BaseSQLParserListener) EnterValuesClause(ctx *ValuesClauseContext) {}

// ExitValuesClause is called when production valuesClause is exited.
func (s *BaseSQLParserListener) ExitValuesClause(ctx *ValuesClauseContext) {}

// EnterValueRow is called when production valueRow is entered.
func (s *BaseSQLParserListener) EnterValueRow(ctx *ValueRowContext) {}

// ExitValueRow is called when production valueRow is exited.
func (s *BaseSQLParserListener) ExitValueRow(ctx *ValueRowContext) {}

// EnterOnDuplicateKeyUpdate is called when production onDuplicateKeyUpdate is entered.
func (s *BaseSQLParserListener) EnterOnDuplicateKeyUpdate(ctx *OnDuplicateKeyUpdateContext) {}

// ExitOnDuplicateKeyUpdate is called when production onDuplicateKeyUpdate is exited.
func (s *BaseSQLParserListener) ExitOnDuplicateKeyUpdate(ctx *OnDuplicateKeyUpdateContext) {}

// EnterUpdateStatement is called when production updateStatement is entered.
func (s *BaseSQLParserListener) EnterUpdateStatement(ctx *UpdateStatementContext) {}

// ExitUpdateStatement is called when production updateStatement is exited.
func (s *BaseSQLParserListener) ExitUpdateStatement(ctx *UpdateStatementContext) {}

// EnterUpdateElement is called when production updateElement is entered.
func (s *BaseSQLParserListener) EnterUpdateElement(ctx *UpdateElementContext) {}

// ExitUpdateElement is called when production updateElement is exited.
func (s *BaseSQLParserListener) ExitUpdateElement(ctx *UpdateElementContext) {}

// EnterDeleteStatement is called when production deleteStatement is entered.
func (s *BaseSQLParserListener) EnterDeleteStatement(ctx *DeleteStatementContext) {}

// ExitDeleteStatement is called when production deleteStatement is exited.
func (s *BaseSQLParserListener) ExitDeleteStatement(ctx *DeleteStatementContext) {}

// EnterTruncateStatement is called when production truncateStatement is entered.
func (s *BaseSQLParserListener) EnterTruncateStatement(ctx *TruncateStatementContext) {}

// ExitTruncateStatement is called when production truncateStatement is exited.
func (s *BaseSQLParserListener) ExitTruncateStatement(ctx *TruncateStatementContext) {}

// EnterMergeStatement is called when production mergeStatement is entered.
func (s *BaseSQLParserListener) EnterMergeStatement(ctx *MergeStatementContext) {}

// ExitMergeStatement is called when production mergeStatement is exited.
func (s *BaseSQLParserListener) ExitMergeStatement(ctx *MergeStatementContext) {}

// EnterMergeClause is called when production mergeClause is entered.
func (s *BaseSQLParserListener) EnterMergeClause(ctx *MergeClauseContext) {}

// ExitMergeClause is called when production mergeClause is exited.
func (s *BaseSQLParserListener) ExitMergeClause(ctx *MergeClauseContext) {}

// EnterMergeUpdateClause is called when production mergeUpdateClause is entered.
func (s *BaseSQLParserListener) EnterMergeUpdateClause(ctx *MergeUpdateClauseContext) {}

// ExitMergeUpdateClause is called when production mergeUpdateClause is exited.
func (s *BaseSQLParserListener) ExitMergeUpdateClause(ctx *MergeUpdateClauseContext) {}

// EnterMergeInsertClause is called when production mergeInsertClause is entered.
func (s *BaseSQLParserListener) EnterMergeInsertClause(ctx *MergeInsertClauseContext) {}

// ExitMergeInsertClause is called when production mergeInsertClause is exited.
func (s *BaseSQLParserListener) ExitMergeInsertClause(ctx *MergeInsertClauseContext) {}

// EnterCreateStatement is called when production createStatement is entered.
func (s *BaseSQLParserListener) EnterCreateStatement(ctx *CreateStatementContext) {}

// ExitCreateStatement is called when production createStatement is exited.
func (s *BaseSQLParserListener) ExitCreateStatement(ctx *CreateStatementContext) {}

// EnterCreateTableStatement is called when production createTableStatement is entered.
func (s *BaseSQLParserListener) EnterCreateTableStatement(ctx *CreateTableStatementContext) {}

// ExitCreateTableStatement is called when production createTableStatement is exited.
func (s *BaseSQLParserListener) ExitCreateTableStatement(ctx *CreateTableStatementContext) {}

// EnterTableElementList is called when production tableElementList is entered.
func (s *BaseSQLParserListener) EnterTableElementList(ctx *TableElementListContext) {}

// ExitTableElementList is called when production tableElementList is exited.
func (s *BaseSQLParserListener) ExitTableElementList(ctx *TableElementListContext) {}

// EnterTableElement is called when production tableElement is entered.
func (s *BaseSQLParserListener) EnterTableElement(ctx *TableElementContext) {}

// ExitTableElement is called when production tableElement is exited.
func (s *BaseSQLParserListener) ExitTableElement(ctx *TableElementContext) {}

// EnterWatermarkDefinition is called when production watermarkDefinition is entered.
func (s *BaseSQLParserListener) EnterWatermarkDefinition(ctx *WatermarkDefinitionContext) {}

// ExitWatermarkDefinition is called when production watermarkDefinition is exited.
func (s *BaseSQLParserListener) ExitWatermarkDefinition(ctx *WatermarkDefinitionContext) {}

// EnterColumnDefinition is called when production columnDefinition is entered.
func (s *BaseSQLParserListener) EnterColumnDefinition(ctx *ColumnDefinitionContext) {}

// ExitColumnDefinition is called when production columnDefinition is exited.
func (s *BaseSQLParserListener) ExitColumnDefinition(ctx *ColumnDefinitionContext) {}

// EnterDataType is called when production dataType is entered.
func (s *BaseSQLParserListener) EnterDataType(ctx *DataTypeContext) {}

// ExitDataType is called when production dataType is exited.
func (s *BaseSQLParserListener) ExitDataType(ctx *DataTypeContext) {}

// EnterPrimitiveType is called when production primitiveType is entered.
func (s *BaseSQLParserListener) EnterPrimitiveType(ctx *PrimitiveTypeContext) {}

// ExitPrimitiveType is called when production primitiveType is exited.
func (s *BaseSQLParserListener) ExitPrimitiveType(ctx *PrimitiveTypeContext) {}

// EnterStructField is called when production structField is entered.
func (s *BaseSQLParserListener) EnterStructField(ctx *StructFieldContext) {}

// ExitStructField is called when production structField is exited.
func (s *BaseSQLParserListener) ExitStructField(ctx *StructFieldContext) {}

// EnterColumnConstraint is called when production columnConstraint is entered.
func (s *BaseSQLParserListener) EnterColumnConstraint(ctx *ColumnConstraintContext) {}

// ExitColumnConstraint is called when production columnConstraint is exited.
func (s *BaseSQLParserListener) ExitColumnConstraint(ctx *ColumnConstraintContext) {}

// EnterIdentityOptions is called when production identityOptions is entered.
func (s *BaseSQLParserListener) EnterIdentityOptions(ctx *IdentityOptionsContext) {}

// ExitIdentityOptions is called when production identityOptions is exited.
func (s *BaseSQLParserListener) ExitIdentityOptions(ctx *IdentityOptionsContext) {}

// EnterReferentialAction is called when production referentialAction is entered.
func (s *BaseSQLParserListener) EnterReferentialAction(ctx *ReferentialActionContext) {}

// ExitReferentialAction is called when production referentialAction is exited.
func (s *BaseSQLParserListener) ExitReferentialAction(ctx *ReferentialActionContext) {}

// EnterReferentialActionType is called when production referentialActionType is entered.
func (s *BaseSQLParserListener) EnterReferentialActionType(ctx *ReferentialActionTypeContext) {}

// ExitReferentialActionType is called when production referentialActionType is exited.
func (s *BaseSQLParserListener) ExitReferentialActionType(ctx *ReferentialActionTypeContext) {}

// EnterNotEnforced is called when production notEnforced is entered.
func (s *BaseSQLParserListener) EnterNotEnforced(ctx *NotEnforcedContext) {}

// ExitNotEnforced is called when production notEnforced is exited.
func (s *BaseSQLParserListener) ExitNotEnforced(ctx *NotEnforcedContext) {}

// EnterTableConstraint is called when production tableConstraint is entered.
func (s *BaseSQLParserListener) EnterTableConstraint(ctx *TableConstraintContext) {}

// ExitTableConstraint is called when production tableConstraint is exited.
func (s *BaseSQLParserListener) ExitTableConstraint(ctx *TableConstraintContext) {}

// EnterPartitionedByClause is called when production partitionedByClause is entered.
func (s *BaseSQLParserListener) EnterPartitionedByClause(ctx *PartitionedByClauseContext) {}

// ExitPartitionedByClause is called when production partitionedByClause is exited.
func (s *BaseSQLParserListener) ExitPartitionedByClause(ctx *PartitionedByClauseContext) {}

// EnterClusteredByClause is called when production clusteredByClause is entered.
func (s *BaseSQLParserListener) EnterClusteredByClause(ctx *ClusteredByClauseContext) {}

// ExitClusteredByClause is called when production clusteredByClause is exited.
func (s *BaseSQLParserListener) ExitClusteredByClause(ctx *ClusteredByClauseContext) {}

// EnterDistributedByClause is called when production distributedByClause is entered.
func (s *BaseSQLParserListener) EnterDistributedByClause(ctx *DistributedByClauseContext) {}

// ExitDistributedByClause is called when production distributedByClause is exited.
func (s *BaseSQLParserListener) ExitDistributedByClause(ctx *DistributedByClauseContext) {}

// EnterSortedByClause is called when production sortedByClause is entered.
func (s *BaseSQLParserListener) EnterSortedByClause(ctx *SortedByClauseContext) {}

// ExitSortedByClause is called when production sortedByClause is exited.
func (s *BaseSQLParserListener) ExitSortedByClause(ctx *SortedByClauseContext) {}

// EnterTableInheritsClause is called when production tableInheritsClause is entered.
func (s *BaseSQLParserListener) EnterTableInheritsClause(ctx *TableInheritsClauseContext) {}

// ExitTableInheritsClause is called when production tableInheritsClause is exited.
func (s *BaseSQLParserListener) ExitTableInheritsClause(ctx *TableInheritsClauseContext) {}

// EnterEngineClause is called when production engineClause is entered.
func (s *BaseSQLParserListener) EnterEngineClause(ctx *EngineClauseContext) {}

// ExitEngineClause is called when production engineClause is exited.
func (s *BaseSQLParserListener) ExitEngineClause(ctx *EngineClauseContext) {}

// EnterCharsetClause is called when production charsetClause is entered.
func (s *BaseSQLParserListener) EnterCharsetClause(ctx *CharsetClauseContext) {}

// ExitCharsetClause is called when production charsetClause is exited.
func (s *BaseSQLParserListener) ExitCharsetClause(ctx *CharsetClauseContext) {}

// EnterCollateClause is called when production collateClause is entered.
func (s *BaseSQLParserListener) EnterCollateClause(ctx *CollateClauseContext) {}

// ExitCollateClause is called when production collateClause is exited.
func (s *BaseSQLParserListener) ExitCollateClause(ctx *CollateClauseContext) {}

// EnterTablespaceClause is called when production tablespaceClause is entered.
func (s *BaseSQLParserListener) EnterTablespaceClause(ctx *TablespaceClauseContext) {}

// ExitTablespaceClause is called when production tablespaceClause is exited.
func (s *BaseSQLParserListener) ExitTablespaceClause(ctx *TablespaceClauseContext) {}

// EnterTtlClause is called when production ttlClause is entered.
func (s *BaseSQLParserListener) EnterTtlClause(ctx *TtlClauseContext) {}

// ExitTtlClause is called when production ttlClause is exited.
func (s *BaseSQLParserListener) ExitTtlClause(ctx *TtlClauseContext) {}

// EnterLifecycleClause is called when production lifecycleClause is entered.
func (s *BaseSQLParserListener) EnterLifecycleClause(ctx *LifecycleClauseContext) {}

// ExitLifecycleClause is called when production lifecycleClause is exited.
func (s *BaseSQLParserListener) ExitLifecycleClause(ctx *LifecycleClauseContext) {}

// EnterRowFormatClause is called when production rowFormatClause is entered.
func (s *BaseSQLParserListener) EnterRowFormatClause(ctx *RowFormatClauseContext) {}

// ExitRowFormatClause is called when production rowFormatClause is exited.
func (s *BaseSQLParserListener) ExitRowFormatClause(ctx *RowFormatClauseContext) {}

// EnterStoredAsClause is called when production storedAsClause is entered.
func (s *BaseSQLParserListener) EnterStoredAsClause(ctx *StoredAsClauseContext) {}

// ExitStoredAsClause is called when production storedAsClause is exited.
func (s *BaseSQLParserListener) ExitStoredAsClause(ctx *StoredAsClauseContext) {}

// EnterLocationClause is called when production locationClause is entered.
func (s *BaseSQLParserListener) EnterLocationClause(ctx *LocationClauseContext) {}

// ExitLocationClause is called when production locationClause is exited.
func (s *BaseSQLParserListener) ExitLocationClause(ctx *LocationClauseContext) {}

// EnterTablePropertiesClause is called when production tablePropertiesClause is entered.
func (s *BaseSQLParserListener) EnterTablePropertiesClause(ctx *TablePropertiesClauseContext) {}

// ExitTablePropertiesClause is called when production tablePropertiesClause is exited.
func (s *BaseSQLParserListener) ExitTablePropertiesClause(ctx *TablePropertiesClauseContext) {}

// EnterWithOptionsClause is called when production withOptionsClause is entered.
func (s *BaseSQLParserListener) EnterWithOptionsClause(ctx *WithOptionsClauseContext) {}

// ExitWithOptionsClause is called when production withOptionsClause is exited.
func (s *BaseSQLParserListener) ExitWithOptionsClause(ctx *WithOptionsClauseContext) {}

// EnterPropertyList is called when production propertyList is entered.
func (s *BaseSQLParserListener) EnterPropertyList(ctx *PropertyListContext) {}

// ExitPropertyList is called when production propertyList is exited.
func (s *BaseSQLParserListener) ExitPropertyList(ctx *PropertyListContext) {}

// EnterProperty is called when production property is entered.
func (s *BaseSQLParserListener) EnterProperty(ctx *PropertyContext) {}

// ExitProperty is called when production property is exited.
func (s *BaseSQLParserListener) ExitProperty(ctx *PropertyContext) {}

// EnterCreateViewStatement is called when production createViewStatement is entered.
func (s *BaseSQLParserListener) EnterCreateViewStatement(ctx *CreateViewStatementContext) {}

// ExitCreateViewStatement is called when production createViewStatement is exited.
func (s *BaseSQLParserListener) ExitCreateViewStatement(ctx *CreateViewStatementContext) {}

// EnterCreateDatabaseStatement is called when production createDatabaseStatement is entered.
func (s *BaseSQLParserListener) EnterCreateDatabaseStatement(ctx *CreateDatabaseStatementContext) {}

// ExitCreateDatabaseStatement is called when production createDatabaseStatement is exited.
func (s *BaseSQLParserListener) ExitCreateDatabaseStatement(ctx *CreateDatabaseStatementContext) {}

// EnterCreateIndexStatement is called when production createIndexStatement is entered.
func (s *BaseSQLParserListener) EnterCreateIndexStatement(ctx *CreateIndexStatementContext) {}

// ExitCreateIndexStatement is called when production createIndexStatement is exited.
func (s *BaseSQLParserListener) ExitCreateIndexStatement(ctx *CreateIndexStatementContext) {}

// EnterIndexColumn is called when production indexColumn is entered.
func (s *BaseSQLParserListener) EnterIndexColumn(ctx *IndexColumnContext) {}

// ExitIndexColumn is called when production indexColumn is exited.
func (s *BaseSQLParserListener) ExitIndexColumn(ctx *IndexColumnContext) {}

// EnterDropStatement is called when production dropStatement is entered.
func (s *BaseSQLParserListener) EnterDropStatement(ctx *DropStatementContext) {}

// ExitDropStatement is called when production dropStatement is exited.
func (s *BaseSQLParserListener) ExitDropStatement(ctx *DropStatementContext) {}

// EnterAlterStatement is called when production alterStatement is entered.
func (s *BaseSQLParserListener) EnterAlterStatement(ctx *AlterStatementContext) {}

// ExitAlterStatement is called when production alterStatement is exited.
func (s *BaseSQLParserListener) ExitAlterStatement(ctx *AlterStatementContext) {}

// EnterAlterTableAction is called when production alterTableAction is entered.
func (s *BaseSQLParserListener) EnterAlterTableAction(ctx *AlterTableActionContext) {}

// ExitAlterTableAction is called when production alterTableAction is exited.
func (s *BaseSQLParserListener) ExitAlterTableAction(ctx *AlterTableActionContext) {}

// EnterCastExpr is called when production castExpr is entered.
func (s *BaseSQLParserListener) EnterCastExpr(ctx *CastExprContext) {}

// ExitCastExpr is called when production castExpr is exited.
func (s *BaseSQLParserListener) ExitCastExpr(ctx *CastExprContext) {}

// EnterExtractExpr is called when production extractExpr is entered.
func (s *BaseSQLParserListener) EnterExtractExpr(ctx *ExtractExprContext) {}

// ExitExtractExpr is called when production extractExpr is exited.
func (s *BaseSQLParserListener) ExitExtractExpr(ctx *ExtractExprContext) {}

// EnterTypeCastExpr is called when production typeCastExpr is entered.
func (s *BaseSQLParserListener) EnterTypeCastExpr(ctx *TypeCastExprContext) {}

// ExitTypeCastExpr is called when production typeCastExpr is exited.
func (s *BaseSQLParserListener) ExitTypeCastExpr(ctx *TypeCastExprContext) {}

// EnterExistsExpr is called when production existsExpr is entered.
func (s *BaseSQLParserListener) EnterExistsExpr(ctx *ExistsExprContext) {}

// ExitExistsExpr is called when production existsExpr is exited.
func (s *BaseSQLParserListener) ExitExistsExpr(ctx *ExistsExprContext) {}

// EnterBitwiseNotExpr is called when production bitwiseNotExpr is entered.
func (s *BaseSQLParserListener) EnterBitwiseNotExpr(ctx *BitwiseNotExprContext) {}

// ExitBitwiseNotExpr is called when production bitwiseNotExpr is exited.
func (s *BaseSQLParserListener) ExitBitwiseNotExpr(ctx *BitwiseNotExprContext) {}

// EnterParenExpr is called when production parenExpr is entered.
func (s *BaseSQLParserListener) EnterParenExpr(ctx *ParenExprContext) {}

// ExitParenExpr is called when production parenExpr is exited.
func (s *BaseSQLParserListener) ExitParenExpr(ctx *ParenExprContext) {}

// EnterConcatExpr is called when production concatExpr is entered.
func (s *BaseSQLParserListener) EnterConcatExpr(ctx *ConcatExprContext) {}

// ExitConcatExpr is called when production concatExpr is exited.
func (s *BaseSQLParserListener) ExitConcatExpr(ctx *ConcatExprContext) {}

// EnterBetweenExpr is called when production betweenExpr is entered.
func (s *BaseSQLParserListener) EnterBetweenExpr(ctx *BetweenExprContext) {}

// ExitBetweenExpr is called when production betweenExpr is exited.
func (s *BaseSQLParserListener) ExitBetweenExpr(ctx *BetweenExprContext) {}

// EnterColumnExpr is called when production columnExpr is entered.
func (s *BaseSQLParserListener) EnterColumnExpr(ctx *ColumnExprContext) {}

// ExitColumnExpr is called when production columnExpr is exited.
func (s *BaseSQLParserListener) ExitColumnExpr(ctx *ColumnExprContext) {}

// EnterVariableExpr is called when production variableExpr is entered.
func (s *BaseSQLParserListener) EnterVariableExpr(ctx *VariableExprContext) {}

// ExitVariableExpr is called when production variableExpr is exited.
func (s *BaseSQLParserListener) ExitVariableExpr(ctx *VariableExprContext) {}

// EnterArrayAccessExpr is called when production arrayAccessExpr is entered.
func (s *BaseSQLParserListener) EnterArrayAccessExpr(ctx *ArrayAccessExprContext) {}

// ExitArrayAccessExpr is called when production arrayAccessExpr is exited.
func (s *BaseSQLParserListener) ExitArrayAccessExpr(ctx *ArrayAccessExprContext) {}

// EnterUnaryMinusExpr is called when production unaryMinusExpr is entered.
func (s *BaseSQLParserListener) EnterUnaryMinusExpr(ctx *UnaryMinusExprContext) {}

// ExitUnaryMinusExpr is called when production unaryMinusExpr is exited.
func (s *BaseSQLParserListener) ExitUnaryMinusExpr(ctx *UnaryMinusExprContext) {}

// EnterLiteralExpr is called when production literalExpr is entered.
func (s *BaseSQLParserListener) EnterLiteralExpr(ctx *LiteralExprContext) {}

// ExitLiteralExpr is called when production literalExpr is exited.
func (s *BaseSQLParserListener) ExitLiteralExpr(ctx *LiteralExprContext) {}

// EnterStructExpr is called when production structExpr is entered.
func (s *BaseSQLParserListener) EnterStructExpr(ctx *StructExprContext) {}

// ExitStructExpr is called when production structExpr is exited.
func (s *BaseSQLParserListener) ExitStructExpr(ctx *StructExprContext) {}

// EnterLikeExpr is called when production likeExpr is entered.
func (s *BaseSQLParserListener) EnterLikeExpr(ctx *LikeExprContext) {}

// ExitLikeExpr is called when production likeExpr is exited.
func (s *BaseSQLParserListener) ExitLikeExpr(ctx *LikeExprContext) {}

// EnterScalarSubqueryExpr is called when production scalarSubqueryExpr is entered.
func (s *BaseSQLParserListener) EnterScalarSubqueryExpr(ctx *ScalarSubqueryExprContext) {}

// ExitScalarSubqueryExpr is called when production scalarSubqueryExpr is exited.
func (s *BaseSQLParserListener) ExitScalarSubqueryExpr(ctx *ScalarSubqueryExprContext) {}

// EnterFuncExpr is called when production funcExpr is entered.
func (s *BaseSQLParserListener) EnterFuncExpr(ctx *FuncExprContext) {}

// ExitFuncExpr is called when production funcExpr is exited.
func (s *BaseSQLParserListener) ExitFuncExpr(ctx *FuncExprContext) {}

// EnterAddSubExpr is called when production addSubExpr is entered.
func (s *BaseSQLParserListener) EnterAddSubExpr(ctx *AddSubExprContext) {}

// ExitAddSubExpr is called when production addSubExpr is exited.
func (s *BaseSQLParserListener) ExitAddSubExpr(ctx *AddSubExprContext) {}

// EnterArrayExpr is called when production arrayExpr is entered.
func (s *BaseSQLParserListener) EnterArrayExpr(ctx *ArrayExprContext) {}

// ExitArrayExpr is called when production arrayExpr is exited.
func (s *BaseSQLParserListener) ExitArrayExpr(ctx *ArrayExprContext) {}

// EnterParameterExpr is called when production parameterExpr is entered.
func (s *BaseSQLParserListener) EnterParameterExpr(ctx *ParameterExprContext) {}

// ExitParameterExpr is called when production parameterExpr is exited.
func (s *BaseSQLParserListener) ExitParameterExpr(ctx *ParameterExprContext) {}

// EnterInExpr is called when production inExpr is entered.
func (s *BaseSQLParserListener) EnterInExpr(ctx *InExprContext) {}

// ExitInExpr is called when production inExpr is exited.
func (s *BaseSQLParserListener) ExitInExpr(ctx *InExprContext) {}

// EnterMemberExpr is called when production memberExpr is entered.
func (s *BaseSQLParserListener) EnterMemberExpr(ctx *MemberExprContext) {}

// ExitMemberExpr is called when production memberExpr is exited.
func (s *BaseSQLParserListener) ExitMemberExpr(ctx *MemberExprContext) {}

// EnterQuantifiedComparisonExpr is called when production quantifiedComparisonExpr is entered.
func (s *BaseSQLParserListener) EnterQuantifiedComparisonExpr(ctx *QuantifiedComparisonExprContext) {}

// ExitQuantifiedComparisonExpr is called when production quantifiedComparisonExpr is exited.
func (s *BaseSQLParserListener) ExitQuantifiedComparisonExpr(ctx *QuantifiedComparisonExprContext) {}

// EnterMapExpr is called when production mapExpr is entered.
func (s *BaseSQLParserListener) EnterMapExpr(ctx *MapExprContext) {}

// ExitMapExpr is called when production mapExpr is exited.
func (s *BaseSQLParserListener) ExitMapExpr(ctx *MapExprContext) {}

// EnterOrExpr is called when production orExpr is entered.
func (s *BaseSQLParserListener) EnterOrExpr(ctx *OrExprContext) {}

// ExitOrExpr is called when production orExpr is exited.
func (s *BaseSQLParserListener) ExitOrExpr(ctx *OrExprContext) {}

// EnterComparisonExpr is called when production comparisonExpr is entered.
func (s *BaseSQLParserListener) EnterComparisonExpr(ctx *ComparisonExprContext) {}

// ExitComparisonExpr is called when production comparisonExpr is exited.
func (s *BaseSQLParserListener) ExitComparisonExpr(ctx *ComparisonExprContext) {}

// EnterBitwiseExpr is called when production bitwiseExpr is entered.
func (s *BaseSQLParserListener) EnterBitwiseExpr(ctx *BitwiseExprContext) {}

// ExitBitwiseExpr is called when production bitwiseExpr is exited.
func (s *BaseSQLParserListener) ExitBitwiseExpr(ctx *BitwiseExprContext) {}

// EnterNotExpr is called when production notExpr is entered.
func (s *BaseSQLParserListener) EnterNotExpr(ctx *NotExprContext) {}

// ExitNotExpr is called when production notExpr is exited.
func (s *BaseSQLParserListener) ExitNotExpr(ctx *NotExprContext) {}

// EnterIsNullExpr is called when production isNullExpr is entered.
func (s *BaseSQLParserListener) EnterIsNullExpr(ctx *IsNullExprContext) {}

// ExitIsNullExpr is called when production isNullExpr is exited.
func (s *BaseSQLParserListener) ExitIsNullExpr(ctx *IsNullExprContext) {}

// EnterUnaryPlusExpr is called when production unaryPlusExpr is entered.
func (s *BaseSQLParserListener) EnterUnaryPlusExpr(ctx *UnaryPlusExprContext) {}

// ExitUnaryPlusExpr is called when production unaryPlusExpr is exited.
func (s *BaseSQLParserListener) ExitUnaryPlusExpr(ctx *UnaryPlusExprContext) {}

// EnterCaseExpr is called when production caseExpr is entered.
func (s *BaseSQLParserListener) EnterCaseExpr(ctx *CaseExprContext) {}

// ExitCaseExpr is called when production caseExpr is exited.
func (s *BaseSQLParserListener) ExitCaseExpr(ctx *CaseExprContext) {}

// EnterSystemVariableExpr is called when production systemVariableExpr is entered.
func (s *BaseSQLParserListener) EnterSystemVariableExpr(ctx *SystemVariableExprContext) {}

// ExitSystemVariableExpr is called when production systemVariableExpr is exited.
func (s *BaseSQLParserListener) ExitSystemVariableExpr(ctx *SystemVariableExprContext) {}

// EnterIntervalExpr is called when production intervalExpr is entered.
func (s *BaseSQLParserListener) EnterIntervalExpr(ctx *IntervalExprContext) {}

// ExitIntervalExpr is called when production intervalExpr is exited.
func (s *BaseSQLParserListener) ExitIntervalExpr(ctx *IntervalExprContext) {}

// EnterMulDivExpr is called when production mulDivExpr is entered.
func (s *BaseSQLParserListener) EnterMulDivExpr(ctx *MulDivExprContext) {}

// ExitMulDivExpr is called when production mulDivExpr is exited.
func (s *BaseSQLParserListener) ExitMulDivExpr(ctx *MulDivExprContext) {}

// EnterIsBooleanExpr is called when production isBooleanExpr is entered.
func (s *BaseSQLParserListener) EnterIsBooleanExpr(ctx *IsBooleanExprContext) {}

// ExitIsBooleanExpr is called when production isBooleanExpr is exited.
func (s *BaseSQLParserListener) ExitIsBooleanExpr(ctx *IsBooleanExprContext) {}

// EnterAndExpr is called when production andExpr is entered.
func (s *BaseSQLParserListener) EnterAndExpr(ctx *AndExprContext) {}

// ExitAndExpr is called when production andExpr is exited.
func (s *BaseSQLParserListener) ExitAndExpr(ctx *AndExprContext) {}

// EnterCastExpression is called when production castExpression is entered.
func (s *BaseSQLParserListener) EnterCastExpression(ctx *CastExpressionContext) {}

// ExitCastExpression is called when production castExpression is exited.
func (s *BaseSQLParserListener) ExitCastExpression(ctx *CastExpressionContext) {}

// EnterExtractExpression is called when production extractExpression is entered.
func (s *BaseSQLParserListener) EnterExtractExpression(ctx *ExtractExpressionContext) {}

// ExitExtractExpression is called when production extractExpression is exited.
func (s *BaseSQLParserListener) ExitExtractExpression(ctx *ExtractExpressionContext) {}

// EnterIntervalExpression is called when production intervalExpression is entered.
func (s *BaseSQLParserListener) EnterIntervalExpression(ctx *IntervalExpressionContext) {}

// ExitIntervalExpression is called when production intervalExpression is exited.
func (s *BaseSQLParserListener) ExitIntervalExpression(ctx *IntervalExpressionContext) {}

// EnterCaseExpression is called when production caseExpression is entered.
func (s *BaseSQLParserListener) EnterCaseExpression(ctx *CaseExpressionContext) {}

// ExitCaseExpression is called when production caseExpression is exited.
func (s *BaseSQLParserListener) ExitCaseExpression(ctx *CaseExpressionContext) {}

// EnterArrayConstructor is called when production arrayConstructor is entered.
func (s *BaseSQLParserListener) EnterArrayConstructor(ctx *ArrayConstructorContext) {}

// ExitArrayConstructor is called when production arrayConstructor is exited.
func (s *BaseSQLParserListener) ExitArrayConstructor(ctx *ArrayConstructorContext) {}

// EnterMapConstructor is called when production mapConstructor is entered.
func (s *BaseSQLParserListener) EnterMapConstructor(ctx *MapConstructorContext) {}

// ExitMapConstructor is called when production mapConstructor is exited.
func (s *BaseSQLParserListener) ExitMapConstructor(ctx *MapConstructorContext) {}

// EnterStructConstructor is called when production structConstructor is entered.
func (s *BaseSQLParserListener) EnterStructConstructor(ctx *StructConstructorContext) {}

// ExitStructConstructor is called when production structConstructor is exited.
func (s *BaseSQLParserListener) ExitStructConstructor(ctx *StructConstructorContext) {}

// EnterFunctionCall is called when production functionCall is entered.
func (s *BaseSQLParserListener) EnterFunctionCall(ctx *FunctionCallContext) {}

// ExitFunctionCall is called when production functionCall is exited.
func (s *BaseSQLParserListener) ExitFunctionCall(ctx *FunctionCallContext) {}

// EnterOverClause is called when production overClause is entered.
func (s *BaseSQLParserListener) EnterOverClause(ctx *OverClauseContext) {}

// ExitOverClause is called when production overClause is exited.
func (s *BaseSQLParserListener) ExitOverClause(ctx *OverClauseContext) {}

// EnterPartitionByClause is called when production partitionByClause is entered.
func (s *BaseSQLParserListener) EnterPartitionByClause(ctx *PartitionByClauseContext) {}

// ExitPartitionByClause is called when production partitionByClause is exited.
func (s *BaseSQLParserListener) ExitPartitionByClause(ctx *PartitionByClauseContext) {}

// EnterWindowFrame is called when production windowFrame is entered.
func (s *BaseSQLParserListener) EnterWindowFrame(ctx *WindowFrameContext) {}

// ExitWindowFrame is called when production windowFrame is exited.
func (s *BaseSQLParserListener) ExitWindowFrame(ctx *WindowFrameContext) {}

// EnterWindowFrameBound is called when production windowFrameBound is entered.
func (s *BaseSQLParserListener) EnterWindowFrameBound(ctx *WindowFrameBoundContext) {}

// ExitWindowFrameBound is called when production windowFrameBound is exited.
func (s *BaseSQLParserListener) ExitWindowFrameBound(ctx *WindowFrameBoundContext) {}

// EnterColumnRef is called when production columnRef is entered.
func (s *BaseSQLParserListener) EnterColumnRef(ctx *ColumnRefContext) {}

// ExitColumnRef is called when production columnRef is exited.
func (s *BaseSQLParserListener) ExitColumnRef(ctx *ColumnRefContext) {}

// EnterExpressionList is called when production expressionList is entered.
func (s *BaseSQLParserListener) EnterExpressionList(ctx *ExpressionListContext) {}

// ExitExpressionList is called when production expressionList is exited.
func (s *BaseSQLParserListener) ExitExpressionList(ctx *ExpressionListContext) {}

// EnterTableName is called when production tableName is entered.
func (s *BaseSQLParserListener) EnterTableName(ctx *TableNameContext) {}

// ExitTableName is called when production tableName is exited.
func (s *BaseSQLParserListener) ExitTableName(ctx *TableNameContext) {}

// EnterDatabaseName is called when production databaseName is entered.
func (s *BaseSQLParserListener) EnterDatabaseName(ctx *DatabaseNameContext) {}

// ExitDatabaseName is called when production databaseName is exited.
func (s *BaseSQLParserListener) ExitDatabaseName(ctx *DatabaseNameContext) {}

// EnterColumnName is called when production columnName is entered.
func (s *BaseSQLParserListener) EnterColumnName(ctx *ColumnNameContext) {}

// ExitColumnName is called when production columnName is exited.
func (s *BaseSQLParserListener) ExitColumnName(ctx *ColumnNameContext) {}

// EnterFunctionName is called when production functionName is entered.
func (s *BaseSQLParserListener) EnterFunctionName(ctx *FunctionNameContext) {}

// ExitFunctionName is called when production functionName is exited.
func (s *BaseSQLParserListener) ExitFunctionName(ctx *FunctionNameContext) {}

// EnterAlias is called when production alias is entered.
func (s *BaseSQLParserListener) EnterAlias(ctx *AliasContext) {}

// ExitAlias is called when production alias is exited.
func (s *BaseSQLParserListener) ExitAlias(ctx *AliasContext) {}

// EnterIdentifier is called when production identifier is entered.
func (s *BaseSQLParserListener) EnterIdentifier(ctx *IdentifierContext) {}

// ExitIdentifier is called when production identifier is exited.
func (s *BaseSQLParserListener) ExitIdentifier(ctx *IdentifierContext) {}

// EnterNonReservedKeyword is called when production nonReservedKeyword is entered.
func (s *BaseSQLParserListener) EnterNonReservedKeyword(ctx *NonReservedKeywordContext) {}

// ExitNonReservedKeyword is called when production nonReservedKeyword is exited.
func (s *BaseSQLParserListener) ExitNonReservedKeyword(ctx *NonReservedKeywordContext) {}

// EnterLiteral is called when production literal is entered.
func (s *BaseSQLParserListener) EnterLiteral(ctx *LiteralContext) {}

// ExitLiteral is called when production literal is exited.
func (s *BaseSQLParserListener) ExitLiteral(ctx *LiteralContext) {}
