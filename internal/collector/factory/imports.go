// Package factory provides a factory pattern implementation for creating
// metadata collectors based on configuration.
//
// IMPORTANT: To register all collector types with the DefaultFactory,
// import the drivers package in your application's main package:
//
//	import _ "go-metadata/internal/collector/drivers"
//
// The drivers package exists separately to avoid import cycles.
// Collector packages (mysql, postgres, hive) import this factory package
// to register themselves, so this package cannot import them directly.
//
// See internal/collector/drivers/drivers.go for the list of registered collectors.
package factory
