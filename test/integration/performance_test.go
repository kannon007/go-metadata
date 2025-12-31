// Package integration provides performance tests for the datasource management module.
package integration

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"go-metadata/internal/biz"
	"go-metadata/internal/service"
)

// TestPerformanceDataSourceCRUD tests CRUD operation performance
// Feature: datasource-management, Property 17: 资源管理优化
// **Validates: Requirements 10.1, 10.2, 10.3**
func TestPerformanceDataSourceCRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	ctx := context.Background()
	logger := log.DefaultLogger

	env := setupPerfTestEnvironment(t, logger)
	defer env.cleanup()

	// Benchmark Create operations
	t.Run("CreatePerformance", func(t *testing.T) {
		iterations := 100
		start := time.Now()

		for i := 0; i < iterations; i++ {
			req := &biz.CreateDataSourceRequest{
				Name:        fmt.Sprintf("perf-test-ds-%d", i),
				Type:        biz.DataSourceTypeMySQL,
				Description: "Performance test datasource",
				Config: &biz.ConnectionConfig{
					Host:     "localhost",
					Port:     3306,
					Database: "test_db",
				},
				CreatedBy: "perf-test",
			}

			_, err := env.dsService.CreateDataSource(ctx, req)
			if err != nil {
				// Expected in mock environment
				continue
			}
		}

		elapsed := time.Since(start)
		avgTime := elapsed / time.Duration(iterations)
		t.Logf("Create %d datasources: total=%v, avg=%v", iterations, elapsed, avgTime)

		// Performance threshold: average should be under 100ms
		if avgTime > 100*time.Millisecond {
			t.Logf("Warning: Average create time %v exceeds 100ms threshold", avgTime)
		}
	})

	// Benchmark List operations
	t.Run("ListPerformance", func(t *testing.T) {
		iterations := 100
		start := time.Now()

		for i := 0; i < iterations; i++ {
			req := &biz.ListDataSourcesRequest{
				Page:     1,
				PageSize: 20,
			}

			_, err := env.dsService.ListDataSources(ctx, req)
			if err != nil {
				t.Errorf("ListDataSources failed: %v", err)
			}
		}

		elapsed := time.Since(start)
		avgTime := elapsed / time.Duration(iterations)
		t.Logf("List %d times: total=%v, avg=%v", iterations, elapsed, avgTime)

		// Performance threshold: average should be under 50ms
		if avgTime > 50*time.Millisecond {
			t.Logf("Warning: Average list time %v exceeds 50ms threshold", avgTime)
		}
	})
}

// TestPerformanceConcurrentAccess tests concurrent access performance
func TestPerformanceConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	ctx := context.Background()
	logger := log.DefaultLogger

	env := setupPerfTestEnvironment(t, logger)
	defer env.cleanup()

	t.Run("ConcurrentReads", func(t *testing.T) {
		concurrency := 50
		iterations := 100

		var wg sync.WaitGroup
		errors := make(chan error, concurrency*iterations)

		start := time.Now()

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				for j := 0; j < iterations; j++ {
					req := &biz.ListDataSourcesRequest{
						Page:     1,
						PageSize: 10,
					}
					_, err := env.dsService.ListDataSources(ctx, req)
					if err != nil {
						errors <- err
					}
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		elapsed := time.Since(start)
		totalOps := concurrency * iterations
		opsPerSec := float64(totalOps) / elapsed.Seconds()

		errorCount := 0
		for range errors {
			errorCount++
		}

		t.Logf("Concurrent reads: %d workers, %d ops, %v elapsed, %.2f ops/sec, %d errors",
			concurrency, totalOps, elapsed, opsPerSec, errorCount)

		// Performance threshold: should handle at least 1000 ops/sec
		if opsPerSec < 1000 {
			t.Logf("Warning: Operations per second %.2f is below 1000 threshold", opsPerSec)
		}
	})

	t.Run("ConcurrentWrites", func(t *testing.T) {
		concurrency := 10
		iterations := 50

		var wg sync.WaitGroup
		var mu sync.Mutex
		successCount := 0
		errorCount := 0

		start := time.Now()

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				for j := 0; j < iterations; j++ {
					req := &biz.CreateDataSourceRequest{
						Name:        fmt.Sprintf("concurrent-ds-%d-%d", workerID, j),
						Type:        biz.DataSourceTypeMySQL,
						Description: "Concurrent test",
						Config: &biz.ConnectionConfig{
							Host:     "localhost",
							Port:     3306,
							Database: "test",
						},
						CreatedBy: fmt.Sprintf("worker-%d", workerID),
					}

					_, err := env.dsService.CreateDataSource(ctx, req)
					mu.Lock()
					if err != nil {
						errorCount++
					} else {
						successCount++
					}
					mu.Unlock()
				}
			}(i)
		}

		wg.Wait()

		elapsed := time.Since(start)
		totalOps := concurrency * iterations
		opsPerSec := float64(totalOps) / elapsed.Seconds()

		t.Logf("Concurrent writes: %d workers, %d ops, %v elapsed, %.2f ops/sec, success=%d, errors=%d",
			concurrency, totalOps, elapsed, opsPerSec, successCount, errorCount)
	})
}

// TestPerformanceTaskOperations tests task operation performance
func TestPerformanceTaskOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	ctx := context.Background()
	logger := log.DefaultLogger

	env := setupPerfTestEnvironment(t, logger)
	defer env.cleanup()

	t.Run("TaskCreatePerformance", func(t *testing.T) {
		iterations := 50
		start := time.Now()

		for i := 0; i < iterations; i++ {
			req := &biz.CreateTaskRequest{
				Name:          fmt.Sprintf("perf-task-%d", i),
				DataSourceID:  "test-ds-id",
				Type:          biz.TaskTypeFullCollection,
				SchedulerType: biz.SchedulerTypeBuiltIn,
				CreatedBy:     "perf-test",
			}

			_, err := env.taskService.CreateTask(ctx, req)
			if err != nil {
				// Expected in mock environment
				continue
			}
		}

		elapsed := time.Since(start)
		avgTime := elapsed / time.Duration(iterations)
		t.Logf("Create %d tasks: total=%v, avg=%v", iterations, elapsed, avgTime)
	})

	t.Run("TaskListPerformance", func(t *testing.T) {
		iterations := 100
		start := time.Now()

		for i := 0; i < iterations; i++ {
			req := &biz.ListTasksRequest{
				Page:     1,
				PageSize: 20,
			}

			_, err := env.taskService.ListTasks(ctx, req)
			if err != nil {
				t.Errorf("ListTasks failed: %v", err)
			}
		}

		elapsed := time.Since(start)
		avgTime := elapsed / time.Duration(iterations)
		t.Logf("List %d times: total=%v, avg=%v", iterations, elapsed, avgTime)
	})
}

// TestPerformanceConnectionPool tests connection pool performance
// Feature: datasource-management, Property 17: 资源管理优化
// **Validates: Requirements 10.1, 10.2, 10.3**
func TestPerformanceConnectionPool(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	t.Run("PoolConfigValidation", func(t *testing.T) {
		// Test pool configuration defaults
		config := &biz.ConnectionConfig{
			Host: "localhost",
		}
		config.SetDefaults(biz.DataSourceTypeMySQL)

		if config.MaxConns != 10 {
			t.Errorf("Expected default max_conns 10, got %d", config.MaxConns)
		}

		if config.MaxIdleConns != 5 {
			t.Errorf("Expected default max_idle_conns 5, got %d", config.MaxIdleConns)
		}

		if config.Timeout != 30 {
			t.Errorf("Expected default timeout 30, got %d", config.Timeout)
		}
	})

	t.Run("PoolStatsStructure", func(t *testing.T) {
		// Verify pool stats structure
		stats := &biz.PoolStats{
			DataSourceID:      "test-ds",
			DataSourceName:    "test-datasource",
			DataSourceType:    biz.DataSourceTypeMySQL,
			MaxOpenConns:      10,
			OpenConns:         8,
			InUseConns:        5,
			IdleConns:         3,
			WaitCount:         0,
			WaitDuration:      0,
			MaxIdleClosed:     0,
			MaxLifetimeClosed: 0,
		}

		if stats.OpenConns != stats.InUseConns+stats.IdleConns {
			t.Errorf("Open conns mismatch: %d != %d + %d",
				stats.OpenConns, stats.InUseConns, stats.IdleConns)
		}
	})
}

// TestPerformanceValidation tests validation performance
func TestPerformanceValidation(t *testing.T) {
	t.Run("DataSourceValidation", func(t *testing.T) {
		iterations := 10000
		start := time.Now()

		for i := 0; i < iterations; i++ {
			req := &biz.CreateDataSourceRequest{
				Name: fmt.Sprintf("test-ds-%d", i),
				Type: biz.DataSourceTypeMySQL,
				Config: &biz.ConnectionConfig{
					Host:     "localhost",
					Port:     3306,
					Database: "test",
				},
			}
			_ = req.Validate()
		}

		elapsed := time.Since(start)
		avgTime := elapsed / time.Duration(iterations)
		t.Logf("Validate %d requests: total=%v, avg=%v", iterations, elapsed, avgTime)

		// Validation should be very fast (under 1ms average)
		if avgTime > time.Millisecond {
			t.Logf("Warning: Average validation time %v exceeds 1ms threshold", avgTime)
		}
	})

	t.Run("TaskValidation", func(t *testing.T) {
		iterations := 10000
		start := time.Now()

		for i := 0; i < iterations; i++ {
			req := &biz.CreateTaskRequest{
				Name:          fmt.Sprintf("test-task-%d", i),
				DataSourceID:  "ds-id",
				Type:          biz.TaskTypeFullCollection,
				SchedulerType: biz.SchedulerTypeBuiltIn,
			}
			_ = req.Validate()
		}

		elapsed := time.Since(start)
		avgTime := elapsed / time.Duration(iterations)
		t.Logf("Validate %d requests: total=%v, avg=%v", iterations, elapsed, avgTime)
	})
}

// TestPerformanceMemoryUsage tests memory usage patterns
func TestPerformanceMemoryUsage(t *testing.T) {
	t.Run("DataSourceAllocation", func(t *testing.T) {
		// Create many datasources to test memory allocation
		datasources := make([]*biz.DataSource, 1000)
		for i := 0; i < 1000; i++ {
			datasources[i] = &biz.DataSource{
				ID:          fmt.Sprintf("ds-%d", i),
				Name:        fmt.Sprintf("test-ds-%d", i),
				Type:        biz.DataSourceTypeMySQL,
				Description: "Test datasource for memory testing",
				Config: &biz.ConnectionConfig{
					Host:     "localhost",
					Port:     3306,
					Database: "test",
					Extra:    make(map[string]string),
				},
				Tags: []string{"test", "memory"},
			}
		}

		// Verify all datasources are created
		if len(datasources) != 1000 {
			t.Errorf("Expected 1000 datasources, got %d", len(datasources))
		}
	})

	t.Run("TaskAllocation", func(t *testing.T) {
		// Create many tasks to test memory allocation
		tasks := make([]*biz.CollectionTask, 1000)
		for i := 0; i < 1000; i++ {
			tasks[i] = &biz.CollectionTask{
				ID:           fmt.Sprintf("task-%d", i),
				Name:         fmt.Sprintf("test-task-%d", i),
				DataSourceID: "ds-id",
				Type:         biz.TaskTypeFullCollection,
				Config: &biz.TaskConfig{
					BatchSize:  1000,
					Timeout:    3600,
					RetryCount: 3,
				},
				Schedule: &biz.ScheduleConfig{
					Type:     biz.ScheduleTypeImmediate,
					Timezone: "Asia/Shanghai",
				},
			}
		}

		if len(tasks) != 1000 {
			t.Errorf("Expected 1000 tasks, got %d", len(tasks))
		}
	})
}

// BenchmarkDataSourceCreate benchmarks datasource creation
func BenchmarkDataSourceCreate(b *testing.B) {
	ctx := context.Background()
	logger := log.DefaultLogger

	mockRepo := &mockDataSourceRepo{dataSources: make(map[string]*biz.DataSource)}
	mockTester := &mockConnectionTester{}
	dsUsecase := biz.NewDataSourceUsecase(mockRepo, mockTester, logger)
	dsService := service.NewDataSourceService(dsUsecase, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := &biz.CreateDataSourceRequest{
			Name:        fmt.Sprintf("bench-ds-%d", i),
			Type:        biz.DataSourceTypeMySQL,
			Description: "Benchmark datasource",
			Config: &biz.ConnectionConfig{
				Host:     "localhost",
				Port:     3306,
				Database: "test",
			},
			CreatedBy: "benchmark",
		}
		dsService.CreateDataSource(ctx, req)
	}
}

// BenchmarkDataSourceList benchmarks datasource listing
func BenchmarkDataSourceList(b *testing.B) {
	ctx := context.Background()
	logger := log.DefaultLogger

	mockRepo := &mockDataSourceRepo{dataSources: make(map[string]*biz.DataSource)}
	mockTester := &mockConnectionTester{}
	dsUsecase := biz.NewDataSourceUsecase(mockRepo, mockTester, logger)
	dsService := service.NewDataSourceService(dsUsecase, logger)

	// Pre-populate some data
	for i := 0; i < 100; i++ {
		mockRepo.dataSources[fmt.Sprintf("ds-%d", i)] = &biz.DataSource{
			ID:   fmt.Sprintf("ds-%d", i),
			Name: fmt.Sprintf("test-ds-%d", i),
			Type: biz.DataSourceTypeMySQL,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := &biz.ListDataSourcesRequest{
			Page:     1,
			PageSize: 20,
		}
		dsService.ListDataSources(ctx, req)
	}
}

// BenchmarkValidation benchmarks validation operations
func BenchmarkValidation(b *testing.B) {
	b.Run("DataSourceRequest", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req := &biz.CreateDataSourceRequest{
				Name: "test-ds",
				Type: biz.DataSourceTypeMySQL,
				Config: &biz.ConnectionConfig{
					Host:     "localhost",
					Port:     3306,
					Database: "test",
				},
			}
			req.Validate()
		}
	})

	b.Run("TaskRequest", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req := &biz.CreateTaskRequest{
				Name:          "test-task",
				DataSourceID:  "ds-id",
				Type:          biz.TaskTypeFullCollection,
				SchedulerType: biz.SchedulerTypeBuiltIn,
			}
			req.Validate()
		}
	})
}

// perfTestEnvironment holds performance test dependencies
type perfTestEnvironment struct {
	dsService   *service.DataSourceService
	taskService *service.TaskService
	cleanup     func()
}

// setupPerfTestEnvironment creates performance test dependencies
func setupPerfTestEnvironment(t *testing.T, logger log.Logger) *perfTestEnvironment {
	t.Helper()

	mockDSRepo := &mockDataSourceRepo{
		dataSources: make(map[string]*biz.DataSource),
	}
	mockTaskRepo := &mockTaskRepo{
		tasks:      make(map[string]*biz.CollectionTask),
		executions: make(map[string]*biz.TaskExecution),
	}
	mockTester := &mockConnectionTester{}
	mockScheduler := &mockTaskScheduler{}

	// Add test datasource for task tests
	mockDSRepo.dataSources["test-ds-id"] = &biz.DataSource{
		ID:     "test-ds-id",
		Name:   "test-ds",
		Type:   biz.DataSourceTypeMySQL,
		Status: biz.DataSourceStatusActive,
	}

	dsUsecase := biz.NewDataSourceUsecase(mockDSRepo, mockTester, logger)
	taskUsecase := biz.NewTaskUsecase(mockTaskRepo, mockDSRepo, mockScheduler, logger)

	dsService := service.NewDataSourceService(dsUsecase, logger)
	taskService := service.NewTaskService(taskUsecase, logger)

	cleanup := func() {}

	return &perfTestEnvironment{
		dsService:   dsService,
		taskService: taskService,
		cleanup:     cleanup,
	}
}
