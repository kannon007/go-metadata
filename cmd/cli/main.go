// Package main provides the entry point for the metadata management CLI tool.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	lineageService "go-metadata/internal/service/lineage"
	metadataService "go-metadata/internal/service/metadata"
)

const (
	appName    = "metadata-cli"
	appVersion = "0.1.0"
)

func main() {
	// Define subcommands
	analyzeCmd := flag.NewFlagSet("analyze", flag.ExitOnError)
	analyzeSQL := analyzeCmd.String("sql", "", "SQL statement to analyze")
	analyzeFile := analyzeCmd.String("file", "", "SQL file to analyze")

	syncCmd := flag.NewFlagSet("sync", flag.ExitOnError)
	syncSource := syncCmd.String("source", "", "Data source name to sync")

	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	listDatabase := listCmd.String("database", "", "Database name")

	// Check for subcommand
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Initialize services
	metaSvc := metadataService.NewService(nil)
	lineageSvc := lineageService.NewService(nil, nil)

	ctx := context.Background()

	switch os.Args[1] {
	case "analyze":
		analyzeCmd.Parse(os.Args[2:])
		runAnalyze(ctx, lineageSvc, *analyzeSQL, *analyzeFile)

	case "sync":
		syncCmd.Parse(os.Args[2:])
		runSync(ctx, metaSvc, *syncSource)

	case "list":
		listCmd.Parse(os.Args[2:])
		runList(ctx, metaSvc, *listDatabase)

	case "version":
		fmt.Printf("%s version %s\n", appName, appVersion)

	case "help":
		printUsage()

	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Printf(`%s - Metadata Management CLI Tool

Usage:
  %s <command> [options]

Commands:
  analyze   Analyze SQL statement for lineage
  sync      Synchronize metadata from data source
  list      List tables in a database
  version   Show version information
  help      Show this help message

Examples:
  %s analyze -sql "SELECT a.id, b.name FROM table_a a JOIN table_b b ON a.id = b.id"
  %s analyze -file query.sql
  %s sync -source mysql_prod
  %s list -database mydb

`, appName, appName, appName, appName, appName, appName)
}

func runAnalyze(ctx context.Context, svc *lineageService.Service, sql, file string) {
	if sql == "" && file == "" {
		fmt.Println("Error: either -sql or -file must be provided")
		os.Exit(1)
	}

	sqlContent := sql
	if file != "" {
		content, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			os.Exit(1)
		}
		sqlContent = string(content)
	}

	result, err := svc.AnalyzeSQL(ctx, sqlContent)
	if err != nil {
		fmt.Printf("Error analyzing SQL: %v\n", err)
		os.Exit(1)
	}

	if result == nil {
		fmt.Println("No lineage information extracted (analyzer not configured)")
		return
	}

	fmt.Println("Lineage analysis completed successfully")
	// TODO: Format and print lineage result
}

func runSync(ctx context.Context, svc *metadataService.Service, source string) {
	if source == "" {
		fmt.Println("Error: -source must be provided")
		os.Exit(1)
	}

	err := svc.SyncMetadata(ctx, source)
	if err != nil {
		fmt.Printf("Error syncing metadata: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Metadata synchronized from source: %s\n", source)
}

func runList(ctx context.Context, svc *metadataService.Service, database string) {
	if database == "" {
		fmt.Println("Error: -database must be provided")
		os.Exit(1)
	}

	tables, err := svc.ListTables(ctx, database)
	if err != nil {
		fmt.Printf("Error listing tables: %v\n", err)
		os.Exit(1)
	}

	if len(tables) == 0 {
		fmt.Printf("No tables found in database: %s\n", database)
		return
	}

	fmt.Printf("Tables in database %s:\n", database)
	for _, t := range tables {
		fmt.Printf("  - %s.%s\n", t.Database, t.Table)
	}
}
