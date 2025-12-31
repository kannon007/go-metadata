package lineage

// Parser parses SQL statements and extracts column-level lineage.
type Parser interface {
	// Parse analyzes the given SQL and returns the lineage result.
	Parse(sql string) (*LineageResult, error)
}

// parserImpl is the default implementation of Parser.
type parserImpl struct {
	catalog Catalog
}

// NewParser creates a new lineage parser with the given catalog.
func NewParser(catalog Catalog) Parser {
	return &parserImpl{
		catalog: catalog,
	}
}

// Parse analyzes the given SQL and returns the lineage result.
func (p *parserImpl) Parse(sql string) (*LineageResult, error) {
	resolver := newResolver(p.catalog)
	return resolver.Resolve(sql)
}
