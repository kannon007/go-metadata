// Package metrics provides Prometheus metrics for the datasource management system.
package metrics

import (
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all Prometheus metrics for the system
type Metrics struct {
	// DataSource metrics
	DataSourcesTotal       *prometheus.GaugeVec
	DataSourceOperations   *prometheus.CounterVec
	DataSourceConnections  *prometheus.GaugeVec
	ConnectionTestDuration *prometheus.HistogramVec
	ConnectionTestResults  *prometheus.CounterVec

	// Task metrics
	TasksTotal           *prometheus.GaugeVec
	TaskOperations       *prometheus.CounterVec
	TaskExecutions       *prometheus.CounterVec
	TaskExecutionDuration *prometheus.HistogramVec
	TasksRunning         prometheus.Gauge

	// Connection Pool metrics
	PoolConnectionsTotal *prometheus.GaugeVec
	PoolConnectionsIdle  *prometheus.GaugeVec
	PoolConnectionsInUse *prometheus.GaugeVec
	PoolWaitCount        *prometheus.CounterVec
	PoolWaitDuration     *prometheus.HistogramVec

	// API metrics
	HTTPRequestsTotal    *prometheus.CounterVec
	HTTPRequestDuration  *prometheus.HistogramVec
	HTTPRequestsInFlight prometheus.Gauge

	// System metrics
	SystemUptime    prometheus.Gauge
	SystemStartTime prometheus.Gauge

	// Registry
	registry *prometheus.Registry
	mu       sync.RWMutex
}

var (
	instance *Metrics
	once     sync.Once
)

// GetMetrics returns the singleton metrics instance
func GetMetrics() *Metrics {
	once.Do(func() {
		instance = newMetrics()
	})
	return instance
}


// newMetrics creates and registers all Prometheus metrics
func newMetrics() *Metrics {
	m := &Metrics{
		registry: prometheus.NewRegistry(),
	}

	// DataSource metrics
	m.DataSourcesTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "metadata",
			Subsystem: "datasource",
			Name:      "total",
			Help:      "Total number of datasources by type and status",
		},
		[]string{"type", "status"},
	)

	m.DataSourceOperations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "metadata",
			Subsystem: "datasource",
			Name:      "operations_total",
			Help:      "Total number of datasource operations",
		},
		[]string{"operation", "status"},
	)

	m.DataSourceConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "metadata",
			Subsystem: "datasource",
			Name:      "connections",
			Help:      "Current number of connections by datasource",
		},
		[]string{"datasource_id", "datasource_type"},
	)

	m.ConnectionTestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "metadata",
			Subsystem: "datasource",
			Name:      "connection_test_duration_seconds",
			Help:      "Duration of connection tests in seconds",
			Buckets:   []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"datasource_type"},
	)

	m.ConnectionTestResults = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "metadata",
			Subsystem: "datasource",
			Name:      "connection_test_results_total",
			Help:      "Total number of connection test results",
		},
		[]string{"datasource_type", "result"},
	)

	// Task metrics
	m.TasksTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "metadata",
			Subsystem: "task",
			Name:      "total",
			Help:      "Total number of tasks by type and status",
		},
		[]string{"type", "status"},
	)

	m.TaskOperations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "metadata",
			Subsystem: "task",
			Name:      "operations_total",
			Help:      "Total number of task operations",
		},
		[]string{"operation", "status"},
	)

	m.TaskExecutions = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "metadata",
			Subsystem: "task",
			Name:      "executions_total",
			Help:      "Total number of task executions",
		},
		[]string{"task_type", "status"},
	)

	m.TaskExecutionDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "metadata",
			Subsystem: "task",
			Name:      "execution_duration_seconds",
			Help:      "Duration of task executions in seconds",
			Buckets:   []float64{1, 5, 10, 30, 60, 120, 300, 600, 1800, 3600},
		},
		[]string{"task_type"},
	)

	m.TasksRunning = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "metadata",
			Subsystem: "task",
			Name:      "running",
			Help:      "Number of currently running tasks",
		},
	)

	// Connection Pool metrics
	m.PoolConnectionsTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "metadata",
			Subsystem: "pool",
			Name:      "connections_total",
			Help:      "Total number of connections in pool",
		},
		[]string{"datasource_id"},
	)

	m.PoolConnectionsIdle = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "metadata",
			Subsystem: "pool",
			Name:      "connections_idle",
			Help:      "Number of idle connections in pool",
		},
		[]string{"datasource_id"},
	)

	m.PoolConnectionsInUse = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "metadata",
			Subsystem: "pool",
			Name:      "connections_in_use",
			Help:      "Number of connections in use",
		},
		[]string{"datasource_id"},
	)

	m.PoolWaitCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "metadata",
			Subsystem: "pool",
			Name:      "wait_count_total",
			Help:      "Total number of times waited for a connection",
		},
		[]string{"datasource_id"},
	)

	m.PoolWaitDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "metadata",
			Subsystem: "pool",
			Name:      "wait_duration_seconds",
			Help:      "Duration of connection wait time in seconds",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"datasource_id"},
	)

	// API metrics
	m.HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "metadata",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	m.HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "metadata",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "Duration of HTTP requests in seconds",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	m.HTTPRequestsInFlight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "metadata",
			Subsystem: "http",
			Name:      "requests_in_flight",
			Help:      "Number of HTTP requests currently being processed",
		},
	)

	// System metrics
	m.SystemUptime = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "metadata",
			Subsystem: "system",
			Name:      "uptime_seconds",
			Help:      "System uptime in seconds",
		},
	)

	m.SystemStartTime = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "metadata",
			Subsystem: "system",
			Name:      "start_time_seconds",
			Help:      "System start time in Unix timestamp",
		},
	)

	// Register all metrics
	m.registry.MustRegister(
		m.DataSourcesTotal,
		m.DataSourceOperations,
		m.DataSourceConnections,
		m.ConnectionTestDuration,
		m.ConnectionTestResults,
		m.TasksTotal,
		m.TaskOperations,
		m.TaskExecutions,
		m.TaskExecutionDuration,
		m.TasksRunning,
		m.PoolConnectionsTotal,
		m.PoolConnectionsIdle,
		m.PoolConnectionsInUse,
		m.PoolWaitCount,
		m.PoolWaitDuration,
		m.HTTPRequestsTotal,
		m.HTTPRequestDuration,
		m.HTTPRequestsInFlight,
		m.SystemUptime,
		m.SystemStartTime,
	)

	// Also register with default registry for compatibility
	prometheus.MustRegister(
		m.DataSourcesTotal,
		m.DataSourceOperations,
		m.DataSourceConnections,
		m.ConnectionTestDuration,
		m.ConnectionTestResults,
		m.TasksTotal,
		m.TaskOperations,
		m.TaskExecutions,
		m.TaskExecutionDuration,
		m.TasksRunning,
		m.PoolConnectionsTotal,
		m.PoolConnectionsIdle,
		m.PoolConnectionsInUse,
		m.PoolWaitCount,
		m.PoolWaitDuration,
		m.HTTPRequestsTotal,
		m.HTTPRequestDuration,
		m.HTTPRequestsInFlight,
		m.SystemUptime,
		m.SystemStartTime,
	)

	// Set system start time
	m.SystemStartTime.Set(float64(time.Now().Unix()))

	return m
}


// Handler returns the HTTP handler for Prometheus metrics
func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}

// DefaultHandler returns the default Prometheus HTTP handler
func (m *Metrics) DefaultHandler() http.Handler {
	return promhttp.Handler()
}

// DataSource metric helpers

// RecordDataSourceOperation records a datasource operation
func (m *Metrics) RecordDataSourceOperation(operation, status string) {
	m.DataSourceOperations.WithLabelValues(operation, status).Inc()
}

// SetDataSourceCount sets the count of datasources by type and status
func (m *Metrics) SetDataSourceCount(dsType, status string, count float64) {
	m.DataSourcesTotal.WithLabelValues(dsType, status).Set(count)
}

// RecordConnectionTest records a connection test result
func (m *Metrics) RecordConnectionTest(dsType string, success bool, duration time.Duration) {
	result := "success"
	if !success {
		result = "failure"
	}
	m.ConnectionTestResults.WithLabelValues(dsType, result).Inc()
	m.ConnectionTestDuration.WithLabelValues(dsType).Observe(duration.Seconds())
}

// SetDataSourceConnections sets the number of connections for a datasource
func (m *Metrics) SetDataSourceConnections(dsID, dsType string, count float64) {
	m.DataSourceConnections.WithLabelValues(dsID, dsType).Set(count)
}

// Task metric helpers

// RecordTaskOperation records a task operation
func (m *Metrics) RecordTaskOperation(operation, status string) {
	m.TaskOperations.WithLabelValues(operation, status).Inc()
}

// SetTaskCount sets the count of tasks by type and status
func (m *Metrics) SetTaskCount(taskType, status string, count float64) {
	m.TasksTotal.WithLabelValues(taskType, status).Set(count)
}

// RecordTaskExecution records a task execution
func (m *Metrics) RecordTaskExecution(taskType, status string, duration time.Duration) {
	m.TaskExecutions.WithLabelValues(taskType, status).Inc()
	m.TaskExecutionDuration.WithLabelValues(taskType).Observe(duration.Seconds())
}

// IncRunningTasks increments the running tasks counter
func (m *Metrics) IncRunningTasks() {
	m.TasksRunning.Inc()
}

// DecRunningTasks decrements the running tasks counter
func (m *Metrics) DecRunningTasks() {
	m.TasksRunning.Dec()
}

// Connection Pool metric helpers

// UpdatePoolStats updates connection pool statistics
func (m *Metrics) UpdatePoolStats(dsID string, total, idle, inUse int) {
	m.PoolConnectionsTotal.WithLabelValues(dsID).Set(float64(total))
	m.PoolConnectionsIdle.WithLabelValues(dsID).Set(float64(idle))
	m.PoolConnectionsInUse.WithLabelValues(dsID).Set(float64(inUse))
}

// RecordPoolWait records a pool wait event
func (m *Metrics) RecordPoolWait(dsID string, duration time.Duration) {
	m.PoolWaitCount.WithLabelValues(dsID).Inc()
	m.PoolWaitDuration.WithLabelValues(dsID).Observe(duration.Seconds())
}

// HTTP metric helpers

// RecordHTTPRequest records an HTTP request
func (m *Metrics) RecordHTTPRequest(method, path, status string, duration time.Duration) {
	m.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}

// IncHTTPRequestsInFlight increments the in-flight requests counter
func (m *Metrics) IncHTTPRequestsInFlight() {
	m.HTTPRequestsInFlight.Inc()
}

// DecHTTPRequestsInFlight decrements the in-flight requests counter
func (m *Metrics) DecHTTPRequestsInFlight() {
	m.HTTPRequestsInFlight.Dec()
}

// System metric helpers

// UpdateUptime updates the system uptime
func (m *Metrics) UpdateUptime(startTime time.Time) {
	m.SystemUptime.Set(time.Since(startTime).Seconds())
}

// Reset resets all metrics (useful for testing)
func (m *Metrics) Reset() {
	m.DataSourcesTotal.Reset()
	m.DataSourceOperations.Reset()
	m.DataSourceConnections.Reset()
	m.ConnectionTestDuration.Reset()
	m.ConnectionTestResults.Reset()
	m.TasksTotal.Reset()
	m.TaskOperations.Reset()
	m.TaskExecutions.Reset()
	m.TaskExecutionDuration.Reset()
	m.TasksRunning.Set(0)
	m.PoolConnectionsTotal.Reset()
	m.PoolConnectionsIdle.Reset()
	m.PoolConnectionsInUse.Reset()
	m.PoolWaitCount.Reset()
	m.PoolWaitDuration.Reset()
	m.HTTPRequestsTotal.Reset()
	m.HTTPRequestDuration.Reset()
	m.HTTPRequestsInFlight.Set(0)
}
