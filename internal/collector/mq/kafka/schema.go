// Package kafka provides Schema Registry client implementation for Kafka metadata collection.
package kafka

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// SchemaRegistryClient Schema Registry 客户端
type SchemaRegistryClient struct {
	baseURL    string
	httpClient *http.Client
	username   string
	password   string
}

// Schema represents a schema from Schema Registry
type Schema struct {
	ID         int    `json:"id"`
	Subject    string `json:"subject"`
	Version    int    `json:"version"`
	Schema     string `json:"schema"`
	SchemaType string `json:"schemaType"`
}

// Subject represents a subject in Schema Registry
type Subject struct {
	Name     string `json:"subject"`
	Versions []int  `json:"versions"`
}

// SchemaVersion represents a specific version of a schema
type SchemaVersion struct {
	Subject    string `json:"subject"`
	ID         int    `json:"id"`
	Version    int    `json:"version"`
	Schema     string `json:"schema"`
	SchemaType string `json:"schemaType"`
}

// AvroSchema represents an Avro schema structure
type AvroSchema struct {
	Type      string        `json:"type"`
	Name      string        `json:"name,omitempty"`
	Namespace string        `json:"namespace,omitempty"`
	Fields    []AvroField   `json:"fields,omitempty"`
	Items     *AvroSchema   `json:"items,omitempty"`
	Values    *AvroSchema   `json:"values,omitempty"`
	Symbols   []string      `json:"symbols,omitempty"`
	Union     []AvroSchema  `json:"-"` // For union types
}

// AvroField represents a field in an Avro record
type AvroField struct {
	Name    string      `json:"name"`
	Type    interface{} `json:"type"` // Can be string or AvroSchema or []interface{} for unions
	Default interface{} `json:"default,omitempty"`
	Doc     string      `json:"doc,omitempty"`
}

// NewSchemaRegistryClient creates a new Schema Registry client
func NewSchemaRegistryClient(baseURL, username, password string) (*SchemaRegistryClient, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("baseURL is required")
	}

	// Validate URL
	if _, err := url.Parse(baseURL); err != nil {
		return nil, fmt.Errorf("invalid baseURL: %v", err)
	}

	client := &SchemaRegistryClient{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		username: username,
		password: password,
	}

	return client, nil
}

// GetSubjects returns all subjects in the Schema Registry
func (c *SchemaRegistryClient) GetSubjects() ([]string, error) {
	url := fmt.Sprintf("%s/subjects", c.baseURL)
	
	resp, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get subjects: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get subjects: status %d", resp.StatusCode)
	}

	var subjects []string
	if err := json.NewDecoder(resp.Body).Decode(&subjects); err != nil {
		return nil, fmt.Errorf("failed to decode subjects response: %v", err)
	}

	return subjects, nil
}

// GetSubjectVersions returns all versions for a given subject
func (c *SchemaRegistryClient) GetSubjectVersions(subject string) ([]int, error) {
	url := fmt.Sprintf("%s/subjects/%s/versions", c.baseURL, url.PathEscape(subject))
	
	resp, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get subject versions: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("subject not found: %s", subject)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get subject versions: status %d", resp.StatusCode)
	}

	var versions []int
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return nil, fmt.Errorf("failed to decode versions response: %v", err)
	}

	return versions, nil
}

// GetLatestSchema returns the latest schema for a given subject
func (c *SchemaRegistryClient) GetLatestSchema(subject string) (*Schema, error) {
	return c.GetSchemaByVersion(subject, "latest")
}

// GetSchemaByVersion returns a specific version of a schema for a given subject
func (c *SchemaRegistryClient) GetSchemaByVersion(subject, version string) (*Schema, error) {
	url := fmt.Sprintf("%s/subjects/%s/versions/%s", c.baseURL, url.PathEscape(subject), version)
	
	resp, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("schema not found: subject=%s, version=%s", subject, version)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get schema: status %d", resp.StatusCode)
	}

	var schemaVersion SchemaVersion
	if err := json.NewDecoder(resp.Body).Decode(&schemaVersion); err != nil {
		return nil, fmt.Errorf("failed to decode schema response: %v", err)
	}

	schema := &Schema{
		ID:         schemaVersion.ID,
		Subject:    schemaVersion.Subject,
		Version:    schemaVersion.Version,
		Schema:     schemaVersion.Schema,
		SchemaType: schemaVersion.SchemaType,
	}

	// Default to AVRO if schema type is not specified
	if schema.SchemaType == "" {
		schema.SchemaType = "AVRO"
	}

	return schema, nil
}

// GetSchemaByID returns a schema by its ID
func (c *SchemaRegistryClient) GetSchemaByID(id int) (*Schema, error) {
	url := fmt.Sprintf("%s/schemas/ids/%d", c.baseURL, id)
	
	resp, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema by ID: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("schema not found: id=%d", id)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get schema by ID: status %d", resp.StatusCode)
	}

	var schema Schema
	if err := json.NewDecoder(resp.Body).Decode(&schema); err != nil {
		return nil, fmt.Errorf("failed to decode schema response: %v", err)
	}

	schema.ID = id

	// Default to AVRO if schema type is not specified
	if schema.SchemaType == "" {
		schema.SchemaType = "AVRO"
	}

	return &schema, nil
}

// ParseAvroSchema parses an Avro schema string into a structured format
func (c *SchemaRegistryClient) ParseAvroSchema(schemaStr string) (*AvroSchema, error) {
	var schema AvroSchema
	if err := json.Unmarshal([]byte(schemaStr), &schema); err != nil {
		return nil, fmt.Errorf("failed to parse Avro schema: %v", err)
	}

	return &schema, nil
}

// GetTopicSchemas returns schemas for a topic (both key and value schemas)
func (c *SchemaRegistryClient) GetTopicSchemas(topic string) (keySchema *Schema, valueSchema *Schema, err error) {
	// Try to get key schema
	keySubject := topic + "-key"
	if keySchema, err = c.GetLatestSchema(keySubject); err != nil {
		// Key schema is optional, so we don't fail if it's not found
		keySchema = nil
	}

	// Try to get value schema
	valueSubject := topic + "-value"
	if valueSchema, err = c.GetLatestSchema(valueSubject); err != nil {
		// Value schema is also optional in some cases
		valueSchema = nil
	}

	return keySchema, valueSchema, nil
}

// ListTopicSchemas returns all topics that have schemas in the registry
func (c *SchemaRegistryClient) ListTopicSchemas() (map[string][]string, error) {
	subjects, err := c.GetSubjects()
	if err != nil {
		return nil, err
	}

	topicSchemas := make(map[string][]string)
	
	for _, subject := range subjects {
		// Parse topic name from subject
		// Subjects typically follow the pattern: <topic>-key or <topic>-value
		var topic, schemaType string
		
		if strings.HasSuffix(subject, "-key") {
			topic = strings.TrimSuffix(subject, "-key")
			schemaType = "key"
		} else if strings.HasSuffix(subject, "-value") {
			topic = strings.TrimSuffix(subject, "-value")
			schemaType = "value"
		} else {
			// Some subjects might not follow the standard naming convention
			topic = subject
			schemaType = "unknown"
		}
		
		if topicSchemas[topic] == nil {
			topicSchemas[topic] = make([]string, 0)
		}
		topicSchemas[topic] = append(topicSchemas[topic], schemaType)
	}

	return topicSchemas, nil
}

// doRequest performs an HTTP request with authentication if configured
func (c *SchemaRegistryClient) doRequest(method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	// Add authentication if configured
	if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	// Set content type for POST/PUT requests
	if method == "POST" || method == "PUT" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Set Accept header
	req.Header.Set("Accept", "application/json")

	return c.httpClient.Do(req)
}

// HealthCheck checks if the Schema Registry is accessible
func (c *SchemaRegistryClient) HealthCheck() error {
	url := fmt.Sprintf("%s/subjects", c.baseURL)
	
	resp, err := c.doRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to Schema Registry: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Schema Registry returned error: status %d", resp.StatusCode)
	}

	return nil
}

// GetSchemaCompatibility returns the compatibility level for a subject
func (c *SchemaRegistryClient) GetSchemaCompatibility(subject string) (string, error) {
	url := fmt.Sprintf("%s/config/%s", c.baseURL, url.PathEscape(subject))
	
	resp, err := c.doRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get compatibility: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// Try global compatibility
		return c.GetGlobalCompatibility()
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get compatibility: status %d", resp.StatusCode)
	}

	var config struct {
		Compatibility string `json:"compatibilityLevel"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return "", fmt.Errorf("failed to decode compatibility response: %v", err)
	}

	return config.Compatibility, nil
}

// GetGlobalCompatibility returns the global compatibility level
func (c *SchemaRegistryClient) GetGlobalCompatibility() (string, error) {
	url := fmt.Sprintf("%s/config", c.baseURL)
	
	resp, err := c.doRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get global compatibility: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get global compatibility: status %d", resp.StatusCode)
	}

	var config struct {
		Compatibility string `json:"compatibilityLevel"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return "", fmt.Errorf("failed to decode compatibility response: %v", err)
	}

	return config.Compatibility, nil
}

// RegisterSchema registers a new schema for a subject
func (c *SchemaRegistryClient) RegisterSchema(subject, schemaStr, schemaType string) (*Schema, error) {
	url := fmt.Sprintf("%s/subjects/%s/versions", c.baseURL, url.PathEscape(subject))
	
	payload := map[string]interface{}{
		"schema":     schemaStr,
		"schemaType": schemaType,
	}
	
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema payload: %v", err)
	}
	
	resp, err := c.doRequest("POST", url, bytes.NewReader(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to register schema: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to register schema: status %d, body: %s", resp.StatusCode, string(body))
	}

	var result struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode register response: %v", err)
	}

	// Get the registered schema details
	return c.GetSchemaByID(result.ID)
}

// DeleteSubject deletes a subject and all its versions
func (c *SchemaRegistryClient) DeleteSubject(subject string) error {
	url := fmt.Sprintf("%s/subjects/%s", c.baseURL, url.PathEscape(subject))
	
	resp, err := c.doRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to delete subject: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete subject: status %d", resp.StatusCode)
	}

	return nil
}

// GetSchemaMetrics returns basic metrics about schemas in the registry
func (c *SchemaRegistryClient) GetSchemaMetrics() (map[string]interface{}, error) {
	subjects, err := c.GetSubjects()
	if err != nil {
		return nil, err
	}

	metrics := map[string]interface{}{
		"total_subjects": len(subjects),
		"total_schemas":  0,
		"schema_types":   make(map[string]int),
	}

	totalSchemas := 0
	schemaTypes := make(map[string]int)

	for _, subject := range subjects {
		versions, err := c.GetSubjectVersions(subject)
		if err != nil {
			continue // Skip subjects we can't access
		}
		
		totalSchemas += len(versions)
		
		// Get latest schema to determine type
		if len(versions) > 0 {
			schema, err := c.GetLatestSchema(subject)
			if err == nil {
				schemaTypes[schema.SchemaType]++
			}
		}
	}

	metrics["total_schemas"] = totalSchemas
	metrics["schema_types"] = schemaTypes

	return metrics, nil
}

// ExtractAvroFields extracts field information from an Avro schema
func (c *SchemaRegistryClient) ExtractAvroFields(schema *AvroSchema) []AvroField {
	if schema == nil || schema.Fields == nil {
		return nil
	}
	
	return schema.Fields
}

// ConvertAvroTypeToSQLType converts Avro types to SQL-like types for metadata
func (c *SchemaRegistryClient) ConvertAvroTypeToSQLType(avroType interface{}) string {
	switch t := avroType.(type) {
	case string:
		switch t {
		case "null":
			return "null"
		case "boolean":
			return "boolean"
		case "int":
			return "int"
		case "long":
			return "bigint"
		case "float":
			return "float"
		case "double":
			return "double"
		case "bytes":
			return "bytes"
		case "string":
			return "string"
		default:
			return t // Custom types
		}
	case []interface{}:
		// Union type - find the non-null type
		for _, unionType := range t {
			if unionType != "null" {
				return c.ConvertAvroTypeToSQLType(unionType)
			}
		}
		return "union"
	case map[string]interface{}:
		// Complex type
		if typeVal, ok := t["type"]; ok {
			switch typeVal {
			case "record":
				return "record"
			case "enum":
				return "enum"
			case "array":
				return "array"
			case "map":
				return "map"
			default:
				return "complex"
			}
		}
		return "object"
	default:
		return "unknown"
	}
}