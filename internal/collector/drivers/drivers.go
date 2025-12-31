// Package drivers imports all collector packages to trigger their init() functions,
// which register each collector type with the factory.DefaultFactory.
//
// This package exists to avoid import cycles. The collector packages import
// the factory package to register themselves, so the factory package cannot
// import the collector packages directly.
//
// Import this package in your application's main package to ensure all
// collectors are registered:
//
//	import _ "go-metadata/internal/collector/drivers"
//
// When adding a new collector, add an import line for its package here.
// For example, to add a ClickHouse collector:
//
//	_ "go-metadata/internal/collector/clickhouse"
package drivers

import (
	// Import collector packages to trigger init() registration
	
	// RDBMS collectors
	_ "go-metadata/internal/collector/rdbms/mysql"
	_ "go-metadata/internal/collector/rdbms/oracle"
	_ "go-metadata/internal/collector/rdbms/postgres"
	_ "go-metadata/internal/collector/rdbms/sqlserver"
	
	// DataWarehouse collectors
	_ "go-metadata/internal/collector/warehouse/clickhouse"
	_ "go-metadata/internal/collector/warehouse/doris"
	_ "go-metadata/internal/collector/warehouse/hive"
	
	// DocumentDB collectors
	_ "go-metadata/internal/collector/docdb/elasticsearch"
	_ "go-metadata/internal/collector/docdb/mongodb"
	
	// KeyValue collectors
	_ "go-metadata/internal/collector/kv/redis"
	
	// MessageQueue collectors
	_ "go-metadata/internal/collector/mq/kafka"
	_ "go-metadata/internal/collector/mq/rabbitmq"
	
	// ObjectStorage collectors
	_ "go-metadata/internal/collector/oss/minio"
)
