// Package lineage provides SQL data lineage parsing capabilities.
// It uses ANTLR4 for SQL parsing and extracts column-level lineage.
package lineage

// Analyzer is the main entry point for lineage analysis.
type Analyzer struct {
	catalog Catalog
}

// NewAnalyzer creates a new lineage analyzer.
func NewAnalyzer(catalog Catalog) *Analyzer {
	return &Analyzer{
		catalog: catalog,
	}
}

// Analyze parses the SQL and extracts column-level lineage.
func (a *Analyzer) Analyze(sql string) (*LineageResult, error) {
	// Parse SQL using ANTLR-generated parser
	stmt, err := ParseSQL(sql)
	if err != nil {
		return nil, err
	}

	extractor := NewExtractor(a.catalog)
	return extractor.Extract(stmt)
}
