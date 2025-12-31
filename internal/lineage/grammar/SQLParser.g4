parser grammar SQLParser;

options {
    tokenVocab = SQLLexer;
}

// ============================================
// Entry point
// ============================================
sqlStatements
    : sqlStatement (SEMI sqlStatement)* SEMI? EOF
    ;

sqlStatement
    : dmlStatement
    | ddlStatement
    ;

dmlStatement
    : selectStatement
    | insertStatement
    | updateStatement
    | deleteStatement
    | mergeStatement
    | truncateStatement
    ;

ddlStatement
    : createStatement
    | dropStatement
    | alterStatement
    ;

// ============================================
// SELECT statement with UNION support
// ============================================
selectStatement
    : queryExpression
    ;

queryExpression
    : queryTerm ((UNION | INTERSECT | EXCEPT | MINUS_SET) ALL? queryTerm)*
    ;

queryTerm
    : withClause? selectClause fromClause? whereClause? groupByClause? havingClause? orderByClause? limitClause?
    | LPAREN queryExpression RPAREN
    ;

withClause
    : WITH RECURSIVE? cteDefinition (COMMA cteDefinition)*
    ;

cteDefinition
    : identifier (LPAREN identifier (COMMA identifier)* RPAREN)? AS LPAREN selectStatement RPAREN
    ;

selectClause
    : SELECT (ALL | DISTINCT)? (TOP expression)? selectElements
    ;

selectElements
    : selectElement (COMMA selectElement)*
    ;

selectElement
    : STAR                                          # selectAll
    | tableName DOT STAR                            # selectTableAll
    | expression (AS? alias)?                       # selectExpr
    ;

// ============================================
// FROM clause
// ============================================
fromClause
    : FROM tableReferences
    ;

tableReferences
    : tableReference (COMMA tableReference)*
    ;

tableReference
    : tableFactor joinPart*
    ;

tableFactor
    : tableName temporalJoinClause? (AS? alias)? tableSample?       # tableNameFactor
    | LPAREN selectStatement RPAREN (AS? alias)?                    # subqueryFactor
    | LATERAL LPAREN selectStatement RPAREN (AS? alias)?            # lateralSubqueryFactor
    | UNNEST LPAREN expression RPAREN (AS? alias)?                  # unnestFactor
    | tableName LPAREN expressionList RPAREN (AS? alias)?           # tableValuedFunctionFactor
    ;

tableSample
    : TABLESAMPLE LPAREN (NUMBER PERCENT | NUMBER ROWS | BUCKET NUMBER OUT OF NUMBER) RPAREN
    ;

joinPart
    : joinType? JOIN tableFactor (ON expression | USING LPAREN identifier (COMMA identifier)* RPAREN)?
    ;

// Flink temporal table join
temporalJoinClause
    : FOR SYSTEM_TIME AS OF expression
    | FOR SYSTEM TIME AS OF expression
    ;

joinType
    : INNER
    | LEFT OUTER?
    | RIGHT OUTER?
    | FULL OUTER?
    | CROSS
    | NATURAL
    | LEFT SEMI_JOIN
    | LEFT ANTI
    ;

// ============================================
// WHERE clause
// ============================================
whereClause
    : WHERE expression
    ;

// ============================================
// GROUP BY clause
// ============================================
groupByClause
    : GROUP BY groupByElements
    ;

groupByElements
    : groupByElement (COMMA groupByElement)*
    ;

groupByElement
    : expression
    | LPAREN RPAREN                                 // GROUPING SETS ()
    | LPAREN expressionList RPAREN                  // GROUPING SETS (a, b)
    ;

// ============================================
// HAVING clause
// ============================================
havingClause
    : HAVING expression
    ;

// ============================================
// ORDER BY clause
// ============================================
orderByClause
    : ORDER BY orderByElement (COMMA orderByElement)*
    ;

orderByElement
    : expression (ASC | DESC)? (NULLS (FIRST | LAST))?
    ;

// ============================================
// LIMIT clause
// ============================================
limitClause
    : LIMIT expression (OFFSET expression)?
    | LIMIT expression COMMA expression
    | OFFSET expression (ROW | ROWS)? (FETCH (FIRST | NEXT) expression (ROW | ROWS) ONLY)?
    ;

// ============================================
// INSERT statement
// ============================================
insertStatement
    : INSERT (INTO | OVERWRITE)? TABLE? tableName partitionSpec? columnList? 
      (selectStatement | valuesClause)
      onDuplicateKeyUpdate?
    ;

partitionSpec
    : PARTITION LPAREN partitionElement (COMMA partitionElement)* RPAREN
    ;

partitionElement
    : identifier (EQ expression)?
    ;

columnList
    : LPAREN identifier (COMMA identifier)* RPAREN
    ;

valuesClause
    : VALUES valueRow (COMMA valueRow)*
    ;

valueRow
    : LPAREN expressionList RPAREN
    ;

onDuplicateKeyUpdate
    : ON DUPLICATE KEY UPDATE updateElement (COMMA updateElement)*
    ;

// ============================================
// UPDATE statement
// ============================================
updateStatement
    : UPDATE tableName (AS? alias)? SET updateElement (COMMA updateElement)* fromClause? whereClause?
    ;

updateElement
    : columnRef EQ expression
    ;

// ============================================
// DELETE statement
// ============================================
deleteStatement
    : DELETE FROM tableName (AS? alias)? whereClause?
    ;

// ============================================
// TRUNCATE statement
// ============================================
truncateStatement
    : TRUNCATE TABLE? tableName
    ;

// ============================================
// MERGE statement (Big Data / UPSERT)
// ============================================
mergeStatement
    : MERGE INTO tableName (AS? alias)?
      USING tableReference (AS? alias)?
      ON expression
      mergeClause+
    ;

mergeClause
    : WHEN MATCHED (AND expression)? THEN mergeUpdateClause
    | WHEN MATCHED (AND expression)? THEN DELETE
    | WHEN NOT MATCHED (AND expression)? THEN mergeInsertClause
    ;

mergeUpdateClause
    : UPDATE SET updateElement (COMMA updateElement)*
    ;

mergeInsertClause
    : INSERT columnList? VALUES LPAREN expressionList RPAREN
    ;

// ============================================
// CREATE statements (DDL)
// ============================================
createStatement
    : createTableStatement
    | createViewStatement
    | createDatabaseStatement
    | createIndexStatement
    ;

createTableStatement
    : CREATE (GLOBAL | LOCAL)? (TEMPORARY | TEMP | EXTERNAL | UNLOGGED)? TABLE (IF_P NOT EXISTS)? tableName
      (LPAREN tableElementList RPAREN)?
      tableInheritsClause?
      partitionedByClause?
      distributedByClause?
      clusteredByClause?
      sortedByClause?
      rowFormatClause?
      storedAsClause?
      locationClause?
      engineClause?
      tablePropertiesClause?
      withOptionsClause?
      charsetClause?
      collateClause?
      tablespaceClause?
      (COMMENT STRING_LITERAL)?
      ttlClause?
      lifecycleClause?
      (AS selectStatement)?
    ;

tableElementList
    : tableElement (COMMA tableElement)*
    ;

tableElement
    : watermarkDefinition
    | tableConstraint
    | columnDefinition
    ;

// Flink WATERMARK definition
watermarkDefinition
    : WATERMARK FOR identifier AS expression
    ;

columnDefinition
    : identifier dataType columnConstraint* (COMMENT STRING_LITERAL)?
    ;

dataType
    : primitiveType
    | ARRAY LT dataType GT
    | MAP LT dataType COMMA dataType GT
    | STRUCT LT structField (COMMA structField)* GT
    | ROW LPAREN structField (COMMA structField)* RPAREN
    ;

primitiveType
    : (INT | INTEGER | TINYINT | SMALLINT | BIGINT)
    | (FLOAT | DOUBLE | REAL)
    | (DECIMAL | NUMERIC) (LPAREN NUMBER (COMMA NUMBER)? RPAREN)?
    | (BOOLEAN | BOOL)
    | (STRING | VARCHAR | CHAR | TEXT) (LPAREN NUMBER RPAREN)?
    | (BINARY | VARBINARY | BLOB | BYTES)
    | (DATE | TIME | TIMESTAMP) (LPAREN NUMBER RPAREN)? (WITH TIME ZONE | WITHOUT TIME ZONE)?
    | (JSON | XML | RAW)
    | MULTISET LT dataType GT
    | identifier (LPAREN NUMBER (COMMA NUMBER)? RPAREN)?  // Custom types with optional precision
    ;

structField
    : identifier COLON dataType
    | identifier dataType
    ;

columnConstraint
    : NOT? NULL
    | DEFAULT expression
    | PRIMARY KEY notEnforced?
    | UNIQUE
    | AUTO_INCREMENT
    | IDENTITY identityOptions?
    | GENERATED (ALWAYS | BY DEFAULT) AS (IDENTITY identityOptions? | LPAREN expression RPAREN STORED?)
    | REFERENCES tableName (LPAREN identifier RPAREN)? referentialAction*
    | CHECK LPAREN expression RPAREN
    | METADATA (FROM STRING_LITERAL)? VIRTUAL?
    | COLLATE identifier
    | CHARACTER SET identifier
    ;

identityOptions
    : LPAREN (START WITH? NUMBER)? (INCREMENT BY? NUMBER)? (MINVALUE NUMBER | NO MINVALUE)? (MAXVALUE NUMBER | NO MAXVALUE)? (CYCLE | NO CYCLE)? (CACHE NUMBER)? RPAREN
    ;

referentialAction
    : ON DELETE referentialActionType
    | ON UPDATE referentialActionType
    ;

referentialActionType
    : CASCADE
    | SET NULL
    | SET DEFAULT
    | RESTRICT
    | NO ACTION
    ;

notEnforced
    : NOT ENFORCED
    ;

tableConstraint
    : PRIMARY KEY LPAREN identifier (COMMA identifier)* RPAREN notEnforced?
    | UNIQUE LPAREN identifier (COMMA identifier)* RPAREN notEnforced?
    | FOREIGN KEY LPAREN identifier (COMMA identifier)* RPAREN REFERENCES tableName (LPAREN identifier (COMMA identifier)* RPAREN)? notEnforced?
    | CHECK LPAREN expression RPAREN
    | CONSTRAINT identifier tableConstraint
    ;

partitionedByClause
    : PARTITIONED BY LPAREN columnDefinition (COMMA columnDefinition)* RPAREN
    | PARTITIONED BY LPAREN identifier (COMMA identifier)* RPAREN
    ;

clusteredByClause
    : CLUSTERED BY LPAREN identifier (COMMA identifier)* RPAREN
      (SORTED BY LPAREN orderByElement (COMMA orderByElement)* RPAREN)?
      (INTO NUMBER BUCKETS)?
    ;

// Doris/StarRocks distributed clause
distributedByClause
    : DISTRIBUTED BY (HASH LPAREN identifier (COMMA identifier)* RPAREN | RANDOM | BROADCAST)
      (BUCKETS NUMBER)?
    ;

// ClickHouse/Doris sorted/order by clause for table
sortedByClause
    : ORDER BY LPAREN identifier (COMMA identifier)* RPAREN
    | ORDER BY identifier
    ;

// PostgreSQL table inheritance
tableInheritsClause
    : INHERITS LPAREN tableName (COMMA tableName)* RPAREN
    ;

// MySQL/ClickHouse engine clause
engineClause
    : ENGINE EQ? identifier (LPAREN expressionList? RPAREN)?
    ;

// MySQL charset clause
charsetClause
    : (DEFAULT? (CHARSET | CHARACTER SET) EQ? identifier)
    ;

// MySQL/PostgreSQL collate clause
collateClause
    : (DEFAULT? COLLATE EQ? identifier)
    ;

// PostgreSQL/Oracle tablespace clause
tablespaceClause
    : TABLESPACE identifier
    ;

// ClickHouse TTL clause
ttlClause
    : TTL expression (COMMA expression)*
    ;

// MaxCompute/ODPS lifecycle clause
lifecycleClause
    : LIFECYCLE NUMBER
    ;

rowFormatClause
    : ROW_FORMAT SERDE STRING_LITERAL (WITH SERDEPROPERTIES LPAREN propertyList RPAREN)?
    | ROW_FORMAT identifier
    ;

storedAsClause
    : STORED AS identifier
    | STORED AS identifier identifier  // e.g., STORED AS INPUTFORMAT 'xxx' OUTPUTFORMAT 'yyy'
    ;

locationClause
    : LOCATION STRING_LITERAL
    ;

tablePropertiesClause
    : TBLPROPERTIES LPAREN propertyList RPAREN
    ;

// Flink WITH clause for connectors
withOptionsClause
    : WITH LPAREN propertyList RPAREN
    ;

propertyList
    : property (COMMA property)*
    ;

property
    : STRING_LITERAL EQ STRING_LITERAL
    | identifier EQ STRING_LITERAL
    ;

createViewStatement
    : CREATE (OR REPLACE)? (TEMPORARY | TEMP)? VIEW (IF_P NOT EXISTS)? tableName
      (LPAREN identifier (COMMA identifier)* RPAREN)?
      (COMMENT STRING_LITERAL)?
      AS selectStatement
    ;

createDatabaseStatement
    : CREATE (DATABASE | SCHEMA) (IF_P NOT EXISTS)? identifier
      (COMMENT STRING_LITERAL)?
      locationClause?
      (WITH propertyList)?
    ;

createIndexStatement
    : CREATE UNIQUE? INDEX (IF_P NOT EXISTS)? identifier
      ON tableName LPAREN indexColumn (COMMA indexColumn)* RPAREN
    ;

indexColumn
    : identifier (ASC | DESC)?
    ;

// ============================================
// DROP statements (DDL)
// ============================================
dropStatement
    : DROP TABLE (IF_P EXISTS)? tableName
    | DROP VIEW (IF_P EXISTS)? tableName
    | DROP (DATABASE | SCHEMA) (IF_P EXISTS)? identifier
    | DROP INDEX (IF_P EXISTS)? identifier (ON tableName)?
    ;

// ============================================
// ALTER statements (DDL)
// ============================================
alterStatement
    : ALTER TABLE tableName alterTableAction (COMMA alterTableAction)*
    | ALTER VIEW tableName AS selectStatement
    | ALTER (DATABASE | SCHEMA) identifier SET propertyList
    ;

alterTableAction
    : ADD COLUMN? columnDefinition
    | DROP COLUMN? identifier
    | RENAME COLUMN? identifier TO identifier
    | RENAME TO tableName
    | MODIFY COLUMN? columnDefinition
    | CHANGE COLUMN? identifier columnDefinition
    | ADD tableConstraint
    | DROP CONSTRAINT identifier
    | SET TBLPROPERTIES LPAREN propertyList RPAREN
    | ADD partitionSpec locationClause?
    | DROP partitionSpec
    ;

// ============================================
// Expressions
// ============================================
expression
    : LPAREN expression RPAREN                                      # parenExpr
    | LPAREN selectStatement RPAREN                                 # scalarSubqueryExpr
    | expression LBRACKET expression RBRACKET                       # arrayAccessExpr
    | expression DOT identifier                                     # memberExpr
    | functionCall                                                  # funcExpr
    | castExpression                                                # castExpr
    | extractExpression                                             # extractExpr
    | intervalExpression                                            # intervalExpr
    | expression DOUBLE_COLON dataType                              # typeCastExpr
    | MINUS expression                                              # unaryMinusExpr
    | PLUS expression                                               # unaryPlusExpr
    | TILDE expression                                              # bitwiseNotExpr
    | expression op=(STAR | DIV | MOD) expression                   # mulDivExpr
    | expression op=(PLUS | MINUS) expression                       # addSubExpr
    | expression CONCAT expression                                  # concatExpr
    | expression op=(AMPERSAND | PIPE | CARET) expression           # bitwiseExpr
    | expression op=(EQ | NEQ | LT | LTE | GT | GTE) expression     # comparisonExpr
    | expression op=(EQ | NEQ | LT | LTE | GT | GTE) (ALL | ANY | SOME) LPAREN selectStatement RPAREN # quantifiedComparisonExpr
    | expression IS NOT? NULL                                       # isNullExpr
    | expression IS NOT? (TRUE | FALSE | UNKNOWN)                   # isBooleanExpr
    | expression NOT? IN LPAREN (selectStatement | expressionList) RPAREN # inExpr
    | expression NOT? BETWEEN expression AND expression             # betweenExpr
    | expression NOT? (LIKE | ILIKE | RLIKE | REGEXP | SIMILAR TO) expression (ESCAPE expression)? # likeExpr
    | NOT expression                                                # notExpr
    | expression AND expression                                     # andExpr
    | expression OR expression                                      # orExpr
    | EXISTS LPAREN selectStatement RPAREN                          # existsExpr
    | caseExpression                                                # caseExpr
    | arrayConstructor                                              # arrayExpr
    | mapConstructor                                                # mapExpr
    | structConstructor                                             # structExpr
    | columnRef                                                     # columnExpr
    | literal                                                       # literalExpr
    | VARIABLE                                                      # variableExpr
    | SYSTEM_VARIABLE                                               # systemVariableExpr
    | QUESTION                                                      # parameterExpr
    ;

castExpression
    : (CAST | TRY_CAST) LPAREN expression AS dataType RPAREN
    | CONVERT LPAREN expression COMMA dataType RPAREN
    ;

extractExpression
    : EXTRACT LPAREN identifier FROM expression RPAREN
    ;

intervalExpression
    : INTERVAL expression identifier
    | INTERVAL STRING_LITERAL
    ;

caseExpression
    : CASE expression? (WHEN expression THEN expression)+ (ELSE expression)? END
    ;

arrayConstructor
    : ARRAY LBRACKET expressionList? RBRACKET
    | ARRAY LPAREN selectStatement RPAREN
    ;

mapConstructor
    : MAP LBRACKET (expression COMMA expression (COMMA expression COMMA expression)*)? RBRACKET
    ;

structConstructor
    : STRUCT LPAREN expressionList RPAREN
    | NAMED_STRUCT LPAREN expressionList RPAREN
    | ROW LPAREN expressionList RPAREN
    ;

functionCall
    : functionName LPAREN (DISTINCT? expressionList | STAR)? RPAREN overClause?
    | functionName LPAREN expressionList RPAREN WITHIN GROUP LPAREN orderByClause RPAREN
    ;

overClause
    : OVER LPAREN partitionByClause? orderByClause? windowFrame? RPAREN
    | OVER identifier
    ;

partitionByClause
    : PARTITION BY expressionList
    ;

windowFrame
    : (ROWS | RANGE | GROUPS) windowFrameBound
    | (ROWS | RANGE | GROUPS) BETWEEN windowFrameBound AND windowFrameBound
    ;

windowFrameBound
    : UNBOUNDED PRECEDING
    | UNBOUNDED FOLLOWING
    | CURRENT ROW
    | expression PRECEDING
    | expression FOLLOWING
    ;

columnRef
    : (tableName DOT)? columnName
    ;

expressionList
    : expression (COMMA expression)*
    ;

// ============================================
// Identifiers
// ============================================
tableName
    : (databaseName DOT)? identifier
    ;

databaseName
    : identifier
    ;

columnName
    : identifier
    ;

functionName
    : identifier
    ;

alias
    : identifier
    ;

identifier
    : IDENTIFIER
    | BACKTICK_IDENTIFIER
    | BRACKET_IDENTIFIER
    | DOUBLE_QUOTED_STRING
    // Non-reserved keywords can be used as identifiers
    | nonReservedKeyword
    ;

nonReservedKeyword
    : SELECT | FROM | WHERE | AND | OR | NOT | AS | ON | JOIN
    | INNER | LEFT | RIGHT | OUTER | CROSS | FULL | NATURAL
    | INSERT | INTO | VALUES | UPDATE | SET | DELETE | TRUNCATE
    | CREATE | TABLE | VIEW | DATABASE | SCHEMA | DROP | ALTER | INDEX
    | ADD | COLUMN | RENAME | TO | MODIFY | CHANGE | CONSTRAINT
    | PRIMARY | KEY | FOREIGN | REFERENCES | UNIQUE | CHECK | DEFAULT
    | GROUP | BY | ORDER | ASC | DESC | HAVING
    | LIMIT | OFFSET | FETCH | NEXT | ONLY | TOP
    | UNION | INTERSECT | EXCEPT | MINUS_SET | ALL | DISTINCT
    | CASE | WHEN | THEN | ELSE | END | NULL | IS | IN
    | BETWEEN | LIKE | ILIKE | RLIKE | REGEXP | SIMILAR | ESCAPE | EXISTS
    | TRUE | FALSE | UNKNOWN
    | WITH | RECURSIVE | OVER | PARTITION | ROWS | RANGE | GROUPS
    | UNBOUNDED | PRECEDING | FOLLOWING | CURRENT | ROW | FIRST | LAST | NULLS
    | MERGE | USING | MATCHED | UPSERT | OVERWRITE | REPLACE | IGNORE | DUPLICATE
    | LATERAL | UNNEST | EXPLODE | TABLESAMPLE | PERCENT | BUCKET | OUT | OF
    | CAST | CONVERT | TRY_CAST | EXTRACT | INTERVAL | AT | ZONE
    | TIME | TIMESTAMP | DATE | YEAR | MONTH | DAY | HOUR | MINUTE | SECOND
    | SOME | ANY | ARRAY | MAP | STRUCT | NAMED_STRUCT
    | INT | INTEGER | TINYINT | SMALLINT | BIGINT
    | FLOAT | DOUBLE | DECIMAL | NUMERIC | REAL
    | BOOLEAN | BOOL | STRING | VARCHAR | CHAR | TEXT
    | BINARY | VARBINARY | BLOB | CLOB | JSON | XML | BYTES | RAW | MULTISET
    | COMMENT | TEMPORARY | TEMP | EXTERNAL | LOCATION | STORED | FORMAT
    | TBLPROPERTIES | PARTITIONED | CLUSTERED | SORTED | BUCKETS
    | SERDE | SERDEPROPERTIES | ROW_FORMAT
    | EXCLUDE | TIES | NO | OTHERS
    | WATERMARK | FOR | SYSTEM_TIME | SYSTEM | ENFORCED | METADATA | VIRTUAL | WITHOUT
    | GENERATED | ALWAYS | IDENTITY | START | INCREMENT | MINVALUE | MAXVALUE | CYCLE | CACHE
    | DISTRIBUTED | HASH | RANDOM | BROADCAST | REPLICATED | PROPERTIES
    | ENGINE | CHARSET | CHARACTER | COLLATE | TABLESPACE | INHERITS | FILEGROUP
    | GLOBAL | LOCAL | UNLOGGED | TTL | LIFECYCLE | AUTO | RESTRICT | CASCADE | ACTION
    ;

// ============================================
// Literals
// ============================================
literal
    : STRING_LITERAL
    | NUMBER
    | HEX_NUMBER
    | BIT_STRING
    | TRUE
    | FALSE
    | NULL
    | INTERVAL STRING_LITERAL identifier?
    ;

