package metadata

import (
	"encoding/json"
	"testing"
)

func TestMemoryProvider_AddTable(t *testing.T) {
	provider := NewMemoryProvider()

	err := provider.AddTable("", "users", []string{"id", "name", "email"})
	if err != nil {
		t.Fatalf("AddTable failed: %v", err)
	}

	schema, err := provider.GetTableSchema("", "users")
	if err != nil {
		t.Fatalf("GetTableSchema failed: %v", err)
	}

	if schema.Table != "users" {
		t.Errorf("Expected table name 'users', got '%s'", schema.Table)
	}

	if len(schema.Columns) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(schema.Columns))
	}

	names := schema.GetColumnNames()
	expected := []string{"id", "name", "email"}
	for i, name := range names {
		if name != expected[i] {
			t.Errorf("Expected column '%s', got '%s'", expected[i], name)
		}
	}
}

func TestMemoryProvider_LoadFromJSON(t *testing.T) {
	jsonData := `[
		{
			"database": "test_db",
			"table": "orders",
			"columns": [
				{"name": "id", "data_type": "INT", "primary_key": true},
				{"name": "user_id", "data_type": "INT", "nullable": false},
				{"name": "amount", "data_type": "DECIMAL(10,2)"},
				{"name": "status", "data_type": "VARCHAR(50)"}
			]
		},
		{
			"database": "test_db",
			"table": "users",
			"columns": [
				{"name": "id", "data_type": "INT", "primary_key": true},
				{"name": "name", "data_type": "VARCHAR(100)"},
				{"name": "email", "data_type": "VARCHAR(255)"}
			]
		}
	]`

	provider := NewMemoryProvider()
	err := provider.LoadFromJSONBytes([]byte(jsonData))
	if err != nil {
		t.Fatalf("LoadFromJSONBytes failed: %v", err)
	}

	// Check orders table
	orders, err := provider.GetTableSchema("test_db", "orders")
	if err != nil {
		t.Fatalf("GetTableSchema(orders) failed: %v", err)
	}

	if len(orders.Columns) != 4 {
		t.Errorf("Expected 4 columns in orders, got %d", len(orders.Columns))
	}

	idCol := orders.GetColumn("id")
	if idCol == nil {
		t.Fatal("Column 'id' not found")
	}
	if !idCol.PrimaryKey {
		t.Error("Expected 'id' to be primary key")
	}

	// Check users table
	users, err := provider.GetTableSchema("test_db", "users")
	if err != nil {
		t.Fatalf("GetTableSchema(users) failed: %v", err)
	}

	if len(users.Columns) != 3 {
		t.Errorf("Expected 3 columns in users, got %d", len(users.Columns))
	}
}

func TestDDLParser_CreateTable(t *testing.T) {
	ddl := `
		CREATE TABLE users (
			id INT PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(255),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`

	parser := NewDDLParser()
	schema, err := parser.ParseDDL(ddl)
	if err != nil {
		t.Fatalf("ParseDDL failed: %v", err)
	}

	if schema == nil {
		t.Fatal("Expected schema, got nil")
	}

	if schema.Table != "users" {
		t.Errorf("Expected table name 'users', got '%s'", schema.Table)
	}

	if len(schema.Columns) != 4 {
		t.Errorf("Expected 4 columns, got %d", len(schema.Columns))
	}

	// Check id column
	idCol := schema.GetColumn("id")
	if idCol == nil {
		t.Fatal("Column 'id' not found")
	}
	if !idCol.PrimaryKey {
		t.Error("Expected 'id' to be primary key")
	}

	// Check name column
	nameCol := schema.GetColumn("name")
	if nameCol == nil {
		t.Fatal("Column 'name' not found")
	}
	if nameCol.Nullable {
		t.Error("Expected 'name' to be NOT NULL")
	}

	// Check email column
	emailCol := schema.GetColumn("email")
	if emailCol == nil {
		t.Fatal("Column 'email' not found")
	}
	if !emailCol.Nullable {
		t.Error("Expected 'email' to be nullable")
	}
}

func TestDDLParser_MultipleTables(t *testing.T) {
	ddl := `
		CREATE TABLE users (
			id INT PRIMARY KEY,
			name VARCHAR(100)
		);
		
		CREATE TABLE orders (
			id INT PRIMARY KEY,
			user_id INT NOT NULL,
			amount DECIMAL(10,2)
		);
	`

	parser := NewDDLParser()
	schemas, err := parser.ParseMultipleDDL(ddl)
	if err != nil {
		t.Fatalf("ParseMultipleDDL failed: %v", err)
	}

	if len(schemas) != 2 {
		t.Fatalf("Expected 2 schemas, got %d", len(schemas))
	}

	// Check first table
	if schemas[0].Table != "users" {
		t.Errorf("Expected first table 'users', got '%s'", schemas[0].Table)
	}

	// Check second table
	if schemas[1].Table != "orders" {
		t.Errorf("Expected second table 'orders', got '%s'", schemas[1].Table)
	}
}

func TestMetadataBuilder_Fluent(t *testing.T) {
	ddl := `
		CREATE TABLE products (
			id INT PRIMARY KEY,
			name VARCHAR(100),
			price DECIMAL(10,2)
		)
	`

	jsonData := `{
		"table": "categories",
		"columns": [
			{"name": "id", "data_type": "INT"},
			{"name": "name", "data_type": "VARCHAR(100)"}
		]
	}`

	analyzer := NewMetadataBuilder().
		WithDefaultDatabase("shop").
		AddTable("", "users", []string{"id", "name", "email"}).
		LoadFromDDL(ddl).
		LoadFromJSON([]byte(jsonData)).
		BuildAnalyzer()

	// Test simple query
	sql := `SELECT id, name FROM users`
	result, err := analyzer.Analyze(sql)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if len(result.Columns) != 2 {
		t.Errorf("Expected 2 columns, got %d", len(result.Columns))
	}
}

func TestMetadataBuilder_ExportJSON(t *testing.T) {
	provider := NewMetadataBuilder().
		AddTable("db1", "users", []string{"id", "name"}).
		AddTable("db1", "orders", []string{"id", "user_id", "amount"}).
		Build()

	jsonBytes, err := provider.ExportToJSON()
	if err != nil {
		t.Fatalf("ExportToJSON failed: %v", err)
	}

	var schemas []TableSchema
	if err := json.Unmarshal(jsonBytes, &schemas); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(schemas) != 2 {
		t.Errorf("Expected 2 schemas, got %d", len(schemas))
	}
}

func TestCatalogAdapter(t *testing.T) {
	provider := NewMemoryProvider()
	_ = provider.AddTable("", "users", []string{"id", "name", "email"})

	adapter := NewCatalogAdapter(provider)

	schema, err := adapter.GetTableSchema("", "users")
	if err != nil {
		t.Fatalf("GetTableSchema failed: %v", err)
	}

	if schema.Table != "users" {
		t.Errorf("Expected table 'users', got '%s'", schema.Table)
	}

	if len(schema.Columns) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(schema.Columns))
	}
}

func TestDDLParser_CreateView(t *testing.T) {
	ddl := `
		CREATE VIEW user_orders AS
		SELECT u.id, u.name, o.amount
		FROM users u
		JOIN orders o ON u.id = o.user_id
	`

	parser := NewDDLParser()
	schema, err := parser.ParseDDL(ddl)
	if err != nil {
		t.Fatalf("ParseDDL failed: %v", err)
	}

	if schema == nil {
		t.Fatal("Expected schema, got nil")
	}

	if schema.Table != "user_orders" {
		t.Errorf("Expected view name 'user_orders', got '%s'", schema.Table)
	}

	if schema.TableType != "VIEW" {
		t.Errorf("Expected table type 'VIEW', got '%s'", schema.TableType)
	}
}

func TestDDLParser_ExternalTable(t *testing.T) {
	ddl := `
		CREATE EXTERNAL TABLE logs (
			timestamp TIMESTAMP,
			level VARCHAR(10),
			message TEXT
		)
		LOCATION 's3://bucket/logs/'
	`

	parser := NewDDLParser()
	schema, err := parser.ParseDDL(ddl)
	if err != nil {
		t.Fatalf("ParseDDL failed: %v", err)
	}

	if schema == nil {
		t.Fatal("Expected schema, got nil")
	}

	if schema.TableType != "EXTERNAL" {
		t.Errorf("Expected table type 'EXTERNAL', got '%s'", schema.TableType)
	}

	if len(schema.Columns) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(schema.Columns))
	}
}

func TestDDLParser_FlinkWatermark(t *testing.T) {
	ddl := `
		CREATE TABLE user_events (
			user_id BIGINT,
			event_type STRING,
			event_time TIMESTAMP(3),
			page_url STRING,
			WATERMARK FOR event_time AS event_time - INTERVAL '5' SECOND
		) WITH (
			'connector' = 'kafka',
			'topic' = 'user_events',
			'format' = 'json'
		)
	`

	parser := NewDDLParser()
	schema, err := parser.ParseDDL(ddl)
	if err != nil {
		t.Fatalf("ParseDDL failed: %v", err)
	}

	if schema == nil {
		t.Fatal("Expected schema, got nil")
	}

	if schema.Table != "user_events" {
		t.Errorf("Expected table name 'user_events', got '%s'", schema.Table)
	}

	// Should have 4 columns (WATERMARK is not a column)
	if len(schema.Columns) != 4 {
		t.Errorf("Expected 4 columns, got %d", len(schema.Columns))
		for _, col := range schema.Columns {
			t.Logf("  Column: %s (%s)", col.Name, col.DataType)
		}
	}
}

func TestDDLParser_FlinkPrimaryKeyNotEnforced(t *testing.T) {
	ddl := `
		CREATE TABLE user_info (
			user_id BIGINT,
			user_name STRING,
			PRIMARY KEY (user_id) NOT ENFORCED
		) WITH (
			'connector' = 'jdbc',
			'url' = 'jdbc:mysql://localhost:3306/db'
		)
	`

	parser := NewDDLParser()
	schema, err := parser.ParseDDL(ddl)
	if err != nil {
		t.Fatalf("ParseDDL failed: %v", err)
	}

	if schema == nil {
		t.Fatal("Expected schema, got nil")
	}

	if len(schema.Columns) != 2 {
		t.Errorf("Expected 2 columns, got %d", len(schema.Columns))
	}
}

func TestDDLParser_SparkHiveTable(t *testing.T) {
	ddl := `
		CREATE EXTERNAL TABLE IF NOT EXISTS ods.user_behavior_log (
			user_id BIGINT COMMENT 'user id',
			event_type STRING COMMENT 'event type',
			event_time TIMESTAMP COMMENT 'event time'
		)
		PARTITIONED BY (dt STRING COMMENT 'date partition')
		STORED AS PARQUET
		LOCATION 'hdfs:///data/ods/user_behavior_log'
		TBLPROPERTIES ('parquet.compression' = 'SNAPPY')
	`

	parser := NewDDLParser()
	schema, err := parser.ParseDDL(ddl)
	if err != nil {
		t.Fatalf("ParseDDL failed: %v", err)
	}

	if schema == nil {
		t.Fatal("Expected schema, got nil")
	}

	if schema.Database != "ods" {
		t.Errorf("Expected database 'ods', got '%s'", schema.Database)
	}

	if schema.Table != "user_behavior_log" {
		t.Errorf("Expected table 'user_behavior_log', got '%s'", schema.Table)
	}

	if schema.TableType != "EXTERNAL" {
		t.Errorf("Expected table type 'EXTERNAL', got '%s'", schema.TableType)
	}
}
