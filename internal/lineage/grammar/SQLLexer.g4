lexer grammar SQLLexer;

// ============================================
// DML Keywords
// ============================================
SELECT: S E L E C T;
FROM: F R O M;
WHERE: W H E R E;
AND: A N D;
OR: O R;
NOT: N O T;
AS: A S;
ON: O N;
JOIN: J O I N;
INNER: I N N E R;
LEFT: L E F T;
RIGHT: R I G H T;
OUTER: O U T E R;
CROSS: C R O S S;
FULL: F U L L;
NATURAL: N A T U R A L;
SEMI_JOIN: S E M I;
ANTI: A N T I;
INSERT: I N S E R T;
INTO: I N T O;
VALUES: V A L U E S;
UPDATE: U P D A T E;
SET: S E T;
DELETE: D E L E T E;
TRUNCATE: T R U N C A T E;

// ============================================
// DDL Keywords
// ============================================
CREATE: C R E A T E;
TABLE: T A B L E;
VIEW: V I E W;
DATABASE: D A T A B A S E;
SCHEMA: S C H E M A;
DROP: D R O P;
ALTER: A L T E R;
INDEX: I N D E X;
ADD: A D D;
COLUMN: C O L U M N;
RENAME: R E N A M E;
TO: T O;
MODIFY: M O D I F Y;
CHANGE: C H A N G E;
CONSTRAINT: C O N S T R A I N T;
PRIMARY: P R I M A R Y;
KEY: K E Y;
FOREIGN: F O R E I G N;
REFERENCES: R E F E R E N C E S;
UNIQUE: U N I Q U E;
CHECK: C H E C K;
DEFAULT: D E F A U L T;
AUTO_INCREMENT: A U T O '_' I N C R E M E N T;
COMMENT: C O M M E N T;
IF_P: I F;
TEMPORARY: T E M P O R A R Y;
TEMP: T E M P;
EXTERNAL: E X T E R N A L;
LOCATION: L O C A T I O N;
STORED: S T O R E D;
FORMAT: F O R M A T;
TBLPROPERTIES: T B L P R O P E R T I E S;
ROW_FORMAT: R O W F O R M A T;
FIELDS: F I E L D S;
TERMINATED: T E R M I N A T E D;
LINES: L I N E S;
COLLECTION: C O L L E C T I O N;
ITEMS: I T E M S;
KEYS: K E Y S;
ESCAPED: E S C A P E D;
SERDE: S E R D E;
SERDEPROPERTIES: S E R D E P R O P E R T I E S;
CLUSTERED: C L U S T E R E D;
SORTED: S O R T E D;
BUCKETS: B U C K E T S;
SKEWED: S K E W E D;
PARTITIONED: P A R T I T I O N E D;

// ============================================
// Flink / Streaming Keywords
// ============================================
WATERMARK: W A T E R M A R K;
FOR: F O R;
SYSTEM_TIME: S Y S T E M '_' T I M E;
SYSTEM: S Y S T E M;
ENFORCED: E N F O R C E D;
METADATA: M E T A D A T A;
VIRTUAL: V I R T U A L;
CONNECTOR: C O N N E C T O R;
OPTIONS: O P T I O N S;
WITHOUT: W I T H O U T;
PERIOD: P E R I O D;
VERSIONING: V E R S I O N I N G;
GENERATED: G E N E R A T E D;
ALWAYS: A L W A Y S;
IDENTITY: I D E N T I T Y;
START: S T A R T;
INCREMENT: I N C R E M E N T;
MINVALUE: M I N V A L U E;
MAXVALUE: M A X V A L U E;
CYCLE: C Y C L E;
CACHE: C A C H E;

// ============================================
// Big Data / Analytics Keywords
// ============================================
MERGE: M E R G E;
USING: U S I N G;
MATCHED: M A T C H E D;
UPSERT: U P S E R T;
OVERWRITE: O V E R W R I T E;
REPLACE: R E P L A C E;
IGNORE: I G N O R E;
DUPLICATE: D U P L I C A T E;
LATERAL: L A T E R A L;
UNNEST: U N N E S T;
EXPLODE: E X P L O D E;
POSEXPLODE: P O S E X P L O D E;
INLINE: I N L I N E;
STACK: S T A C K;
TABLESAMPLE: T A B L E S A M P L E;
PERCENT: P E R C E N T;
BUCKET: B U C K E T;
OUT: O U T;
OF: O F;
DISTRIBUTED: D I S T R I B U T E D;
HASH: H A S H;
RANDOM: R A N D O M;
BROADCAST: B R O A D C A S T;
REPLICATED: R E P L I C A T E D;
PROPERTIES: P R O P E R T I E S;
ENGINE: E N G I N E;
CHARSET: C H A R S E T;
CHARACTER: C H A R A C T E R;
COLLATE: C O L L A T E;
TABLESPACE: T A B L E S P A C E;
INHERITS: I N H E R I T S;
FILEGROUP: F I L E G R O U P;
CLUSTERED_INDEX: N O N C L U S T E R E D;
ON_COMMIT: O N '_' C O M M I T;
PRESERVE: P R E S E R V E;
GLOBAL: G L O B A L;
LOCAL: L O C A L;
UNLOGGED: U N L O G G E D;
TTL: T T L;
LIFECYCLE: L I F E C Y C L E;
AUTO: A U T O;
INCR: I N C R;
RESTRICT: R E S T R I C T;
CASCADE: C A S C A D E;
ACTION: A C T I O N;

// ============================================
// Window / Analytic Keywords
// ============================================
OVER: O V E R;
PARTITION: P A R T I T I O N;
ROWS: R O W S;
RANGE: R A N G E;
GROUPS: G R O U P S;
UNBOUNDED: U N B O U N D E D;
PRECEDING: P R E C E D I N G;
FOLLOWING: F O L L O W I N G;
CURRENT: C U R R E N T;
ROW: R O W;
FIRST: F I R S T;
LAST: L A S T;
NULLS: N U L L S;
EXCLUDE: E X C L U D E;
TIES: T I E S;
NO: N O;
OTHERS: O T H E R S;

// ============================================
// Query Structure Keywords
// ============================================
GROUP: G R O U P;
BY: B Y;
ORDER: O R D E R;
ASC: A S C;
DESC: D E S C;
HAVING: H A V I N G;
LIMIT: L I M I T;
OFFSET: O F F S E T;
FETCH: F E T C H;
NEXT: N E X T;
ONLY: O N L Y;
TOP: T O P;
UNION: U N I O N;
INTERSECT: I N T E R S E C T;
EXCEPT: E X C E P T;
MINUS_SET: M I N U S;
ALL: A L L;
DISTINCT: D I S T I N C T;
WITH: W I T H;
WITHIN: W I T H I N;
RECURSIVE: R E C U R S I V E;

// ============================================
// Expression Keywords
// ============================================
CASE: C A S E;
WHEN: W H E N;
THEN: T H E N;
ELSE: E L S E;
END: E N D;
NULL: N U L L;
IS: I S;
IN: I N;
BETWEEN: B E T W E E N;
LIKE: L I K E;
ILIKE: I L I K E;
RLIKE: R L I K E;
REGEXP: R E G E X P;
SIMILAR: S I M I L A R;
ESCAPE: E S C A P E;
EXISTS: E X I S T S;
TRUE: T R U E;
FALSE: F A L S E;
UNKNOWN: U N K N O W N;
CAST: C A S T;
CONVERT: C O N V E R T;
TRY_CAST: T R Y '_' C A S T;
EXTRACT: E X T R A C T;
INTERVAL: I N T E R V A L;
AT: A T;
ZONE: Z O N E;
TIME: T I M E;
TIMESTAMP: T I M E S T A M P;
DATE: D A T E;
YEAR: Y E A R;
MONTH: M O N T H;
DAY: D A Y;
HOUR: H O U R;
MINUTE: M I N U T E;
SECOND: S E C O N D;
SOME: S O M E;
ANY: A N Y;
ARRAY: A R R A Y;
MAP: M A P;
STRUCT: S T R U C T;
NAMED_STRUCT: N A M E D '_' S T R U C T;

// ============================================
// Data Types
// ============================================
INT: I N T;
INTEGER: I N T E G E R;
TINYINT: T I N Y I N T;
SMALLINT: S M A L L I N T;
BIGINT: B I G I N T;
FLOAT: F L O A T;
DOUBLE: D O U B L E;
DECIMAL: D E C I M A L;
NUMERIC: N U M E R I C;
REAL: R E A L;
BOOLEAN: B O O L E A N;
BOOL: B O O L;
STRING: S T R I N G;
VARCHAR: V A R C H A R;
CHAR: C H A R;
TEXT: T E X T;
BINARY: B I N A R Y;
VARBINARY: V A R B I N A R Y;
BLOB: B L O B;
CLOB: C L O B;
JSON: J S O N;
XML: X M L;
BYTES: B Y T E S;
MULTISET: M U L T I S E T;
RAW: R A W;

// ============================================
// Operators
// ============================================
EQ: '=';
NEQ: '<>' | '!=';
LT: '<';
LTE: '<=';
GT: '>';
GTE: '>=';
PLUS: '+';
MINUS: '-';
STAR: '*';
DIV: '/';
MOD: '%';
CONCAT: '||';
AMPERSAND: '&';
PIPE: '|';
CARET: '^';
TILDE: '~';
DOUBLE_COLON: '::';
ARROW: '->';
DOUBLE_ARROW: '->>';
AT_SIGN: '@';
HASH_SIGN: '#';
DOLLAR: '$';

// ============================================
// Symbols
// ============================================
LPAREN: '(';
RPAREN: ')';
LBRACKET: '[';
RBRACKET: ']';
LBRACE: '{';
RBRACE: '}';
COMMA: ',';
DOT: '.';
SEMI: ';';
COLON: ':';
QUESTION: '?';

// ============================================
// Literals
// ============================================
STRING_LITERAL: '\'' (~'\'' | '\'\'')* '\'';
DOUBLE_QUOTED_STRING: '"' (~'"' | '""')* '"';
NUMBER: DIGIT+ ('.' DIGIT+)? (E (PLUS | MINUS)? DIGIT+)?;
HEX_NUMBER: '0' X HEX_DIGIT+;
BIT_STRING: B '\'' [01]+ '\'';

// ============================================
// Identifiers
// ============================================
IDENTIFIER: (LETTER | '_') (LETTER | DIGIT | '_')*;
BACKTICK_IDENTIFIER: '`' (~'`')+ '`';
BRACKET_IDENTIFIER: '[' (~']')+ ']';
VARIABLE: '@' (LETTER | '_') (LETTER | DIGIT | '_')*;
SYSTEM_VARIABLE: '@@' (LETTER | '_') (LETTER | DIGIT | '_')*;

// ============================================
// Whitespace and comments
// ============================================
WS: [ \t\r\n]+ -> skip;
LINE_COMMENT: '--' ~[\r\n]* -> skip;
BLOCK_COMMENT: '/*' .*? '*/' -> skip;
HINT_COMMENT: '/*+' .*? '*/';

// Fragments
fragment LETTER: [a-zA-Z];
fragment DIGIT: [0-9];
fragment HEX_DIGIT: [0-9a-fA-F];
fragment A: [aA];
fragment B: [bB];
fragment C: [cC];
fragment D: [dD];
fragment E: [eE];
fragment F: [fF];
fragment G: [gG];
fragment H: [hH];
fragment I: [iI];
fragment J: [jJ];
fragment K: [kK];
fragment L: [lL];
fragment M: [mM];
fragment N: [nN];
fragment O: [oO];
fragment P: [pP];
fragment Q: [qQ];
fragment R: [rR];
fragment S: [sS];
fragment T: [tT];
fragment U: [uU];
fragment V: [vV];
fragment W: [wW];
fragment X: [xX];
fragment Y: [yY];
fragment Z: [zZ];
