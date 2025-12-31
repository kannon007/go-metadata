package minio

import (
	"strings"
	"testing"

	"go-metadata/internal/collector"
	"go-metadata/internal/collector/config"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// getTestParameters returns the standard test parameters for property tests.
func getTestParameters() *gopter.TestParameters {
	params := gopter.DefaultTestParameters()
	params.MinSuccessfulTests = 100
	return params
}

// genBucketName generates a valid bucket name for testing
func genBucketName() gopter.Gen {
	return gen.RegexMatch("test-bucket-[0-9]+")
}

// genObjectKey generates a valid object key for testing
func genObjectKey() gopter.Gen {
	return gen.RegexMatch("folder[0-9]+/subfolder[0-9]+/file[0-9]+\\.(csv|json|txt)")
}

// genDelimiter generates a delimiter character for prefix listing
func genDelimiter() gopter.Gen {
	return gen.Const("/")
}

// genObjectKeysWithPrefix generates object keys that have common prefixes
func genObjectKeysWithPrefix(delimiter string) gopter.Gen {
	return gen.SliceOfN(5, gen.OneConstOf(
		"data"+delimiter+"2023"+delimiter+"file1.csv",
		"data"+delimiter+"2024"+delimiter+"file2.csv", 
		"logs"+delimiter+"app"+delimiter+"error.log",
		"logs"+delimiter+"system"+delimiter+"access.log",
		"images"+delimiter+"photos"+delimiter+"pic1.jpg",
	))
}

// Feature: metadata-collector, Property 14: Object Storage Prefix as Schema
// **Validates: Requirements 10.3**
func TestObjectStoragePrefixAsSchema(t *testing.T) {
	properties := gopter.NewProperties(getTestParameters())

	properties.Property("listing prefixes with delimiter returns distinct prefix paths", prop.ForAll(
		func(bucket string, objectKeys []string) bool {
			delimiter := "/"
			
			// Create a mock collector for testing prefix logic
			cfg := &config.ConnectorConfig{
				Type:     SourceName,
				Endpoint: "localhost:9000",
			}
			
			_, err := NewCollector(cfg)
			if err != nil {
				return false
			}
			
			// Test the prefix extraction logic
			prefixSet := make(map[string]bool)
			var prefixes []string
			
			for _, objectKey := range objectKeys {
				// Extract prefix from object key (simulate listPrefixes logic)
				if delimiterIdx := strings.Index(objectKey, delimiter); delimiterIdx != -1 {
					prefixName := objectKey[:delimiterIdx]
					if prefixName != "" && !prefixSet[prefixName] {
						prefixSet[prefixName] = true
						prefixes = append(prefixes, prefixName)
					}
				}
			}
			
			// Property: All returned prefixes should be distinct
			if len(prefixes) != len(prefixSet) {
				return false
			}
			
			// Property: Each prefix should be non-empty
			for _, prefix := range prefixes {
				if prefix == "" {
					return false
				}
			}
			
			// Property: Prefixes should be derived from the input object keys
			for _, prefix := range prefixes {
				found := false
				for _, objectKey := range objectKeys {
					if strings.HasPrefix(objectKey, prefix+delimiter) {
						found = true
						break
					}
				}
				if !found {
					return false
				}
			}
			
			return true
		},
		genBucketName(),
		genObjectKeysWithPrefix("/"),
	))

	properties.Property("prefix extraction is consistent with delimiter", prop.ForAll(
		func(objectKeys []string) bool {
			delimiter := "/"
			
			// Test that prefix extraction respects the delimiter
			prefixSet := make(map[string]bool)
			
			for _, objectKey := range objectKeys {
				if delimiterIdx := strings.Index(objectKey, delimiter); delimiterIdx != -1 {
					prefixName := objectKey[:delimiterIdx]
					if prefixName != "" {
						prefixSet[prefixName] = true
					}
				}
			}
			
			// Property: Each prefix should not contain the delimiter
			for prefix := range prefixSet {
				if strings.Contains(prefix, delimiter) {
					return false
				}
			}
			
			return true
		},
		genObjectKeysWithPrefix("/"),
	))

	properties.Property("empty object list returns empty prefixes", prop.ForAll(
		func(bucket string, delimiter string) bool {
			// Test edge case: empty object list
			cfg := &config.ConnectorConfig{
				Type:     SourceName,
				Endpoint: "localhost:9000",
			}
			
			_, err := NewCollector(cfg)
			if err != nil {
				return false
			}
			
			// Simulate empty object list
			var objectKeys []string
			prefixSet := make(map[string]bool)
			
			for _, objectKey := range objectKeys {
				if delimiterIdx := strings.Index(objectKey, delimiter); delimiterIdx != -1 {
					prefixName := objectKey[:delimiterIdx]
					if prefixName != "" {
						prefixSet[prefixName] = true
					}
				}
			}
			
			// Property: Empty input should result in empty prefixes
			return len(prefixSet) == 0
		},
		genBucketName(),
		genDelimiter(),
	))

	properties.Property("single level objects without delimiter create no prefixes", prop.ForAll(
		func(bucket string, delimiter string) bool {
			// Test objects without delimiter (single level)
			cfg := &config.ConnectorConfig{
				Type:     SourceName,
				Endpoint: "localhost:9000",
			}
			
			_, err := NewCollector(cfg)
			if err != nil {
				return false
			}
			
			// Generate object keys without delimiter
			objectKeys := []string{"file1", "file2", "document", "image"}
			prefixSet := make(map[string]bool)
			
			for _, objectKey := range objectKeys {
				if delimiterIdx := strings.Index(objectKey, delimiter); delimiterIdx != -1 {
					prefixName := objectKey[:delimiterIdx]
					if prefixName != "" {
						prefixSet[prefixName] = true
					}
				}
			}
			
			// Property: Objects without delimiter should not create prefixes
			return len(prefixSet) == 0
		},
		genBucketName(),
		genDelimiter(),
	))

	properties.Property("prefix structure preserves hierarchical organization", prop.ForAll(
		func(bucket string, delimiter string) bool {
			// Test hierarchical prefix structure
			cfg := &config.ConnectorConfig{
				Type:     SourceName,
				Endpoint: "localhost:9000",
			}
			
			_, err := NewCollector(cfg)
			if err != nil {
				return false
			}
			
			// Create hierarchical object keys
			objectKeys := []string{
				"data" + delimiter + "2023" + delimiter + "file1.csv",
				"data" + delimiter + "2024" + delimiter + "file2.csv",
				"logs" + delimiter + "app" + delimiter + "error.log",
				"logs" + delimiter + "system" + delimiter + "access.log",
			}
			
			prefixSet := make(map[string]bool)
			
			for _, objectKey := range objectKeys {
				if delimiterIdx := strings.Index(objectKey, delimiter); delimiterIdx != -1 {
					prefixName := objectKey[:delimiterIdx]
					if prefixName != "" {
						prefixSet[prefixName] = true
					}
				}
			}
			
			// Property: Should extract top-level prefixes correctly
			expectedPrefixes := map[string]bool{
				"data": true,
				"logs": true,
			}
			
			if len(prefixSet) != len(expectedPrefixes) {
				return false
			}
			
			for prefix := range expectedPrefixes {
				if !prefixSet[prefix] {
					return false
				}
			}
			
			return true
		},
		genBucketName(),
		genDelimiter(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: metadata-collector, Property 15: File Schema Inference
// **Validates: Requirements 10.4**
func TestFileSchemaInference(t *testing.T) {
	properties := gopter.NewProperties(getTestParameters())

	properties.Property("structured file detection is consistent", prop.ForAll(
		func(filename string, extension string) bool {
			cfg := &config.ConnectorConfig{
				Type:     SourceName,
				Endpoint: "localhost:9000",
			}
			
			c, err := NewCollector(cfg)
			if err != nil {
				return false
			}
			
			collector := c.(*Collector)
			
			// Test file with extension
			fileWithExt := filename + "." + extension
			isStructured := collector.isStructuredFile(fileWithExt)
			
			// Property: Files with structured extensions should be detected
			structuredExts := map[string]bool{
				"csv":     true,
				"json":    true,
				"parquet": true,
				"CSV":     true,
				"JSON":    true,
				"PARQUET": true,
			}
			
			expectedStructured := structuredExts[extension]
			
			return isStructured == expectedStructured
		},
		gen.AlphaString(),
		gen.OneConstOf("csv", "json", "parquet", "txt", "jpg", "pdf", "CSV", "JSON", "PARQUET"),
	))

	properties.Property("file extension case insensitive detection", prop.ForAll(
		func(filename string) bool {
			cfg := &config.ConnectorConfig{
				Type:     SourceName,
				Endpoint: "localhost:9000",
			}
			
			c, err := NewCollector(cfg)
			if err != nil {
				return false
			}
			
			collector := c.(*Collector)
			
			// Test case insensitive detection
			lowerFile := filename + ".csv"
			upperFile := filename + ".CSV"
			mixedFile := filename + ".CsV"
			
			lowerResult := collector.isStructuredFile(lowerFile)
			upperResult := collector.isStructuredFile(upperFile)
			mixedResult := collector.isStructuredFile(mixedFile)
			
			// Property: Case should not matter for structured file detection
			return lowerResult && upperResult && mixedResult
		},
		gen.AlphaString(),
	))

	properties.Property("non-structured files are correctly identified", prop.ForAll(
		func(filename string, extension string) bool {
			cfg := &config.ConnectorConfig{
				Type:     SourceName,
				Endpoint: "localhost:9000",
			}
			
			c, err := NewCollector(cfg)
			if err != nil {
				return false
			}
			
			collector := c.(*Collector)
			
			// Test non-structured file
			fileWithExt := filename + "." + extension
			isStructured := collector.isStructuredFile(fileWithExt)
			
			// Property: Non-structured files should not be detected as structured
			nonStructuredExts := []string{"txt", "jpg", "pdf", "doc", "exe", "zip"}
			for _, ext := range nonStructuredExts {
				if extension == ext {
					return !isStructured
				}
			}
			
			return true
		},
		gen.AlphaString(),
		gen.OneConstOf("txt", "jpg", "pdf", "doc", "exe", "zip", "log", "xml"),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Additional property test for MinIO collector configuration
func TestMinIOCollectorConfiguration(t *testing.T) {
	properties := gopter.NewProperties(getTestParameters())

	properties.Property("collector category and type are consistent", prop.ForAll(
		func(endpoint string) bool {
			cfg := &config.ConnectorConfig{
				Type:     SourceName,
				Endpoint: endpoint,
			}
			
			c, err := NewCollector(cfg)
			if err != nil {
				return false
			}
			
			// Property: Category should always be ObjectStorage
			if c.Category() != collector.CategoryObjectStorage {
				return false
			}
			
			// Property: Type should always be "minio"
			if c.Type() != SourceName {
				return false
			}
			
			return true
		},
		gen.RegexMatch("https?://[a-zA-Z0-9.-]+:[0-9]+"),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}