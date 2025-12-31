// Package metrics provides a metrics collector that periodically updates system metrics.
package metrics

import (
	"context"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// StatsProvider defines the interface for providing statistics
type StatsProvider interface {
	// GetDataSourceStats returns datasource statistics
	GetDataSourceStats(ctx context.Context) (*DataSourceStats, error)
	// GetTaskStats returns task statistics
	GetTaskStats(ctx context.Context) (*TaskStats, error)
	// GetPoolStats returns connection pool statistics
	GetPoolStats(ctx context.Context) ([]*PoolStatsInfo, error)
}

// DataSourceStats holds datasource statistics
type DataSourceStats struct {
	ByTypeAndStatus map[string]map[string]int64 // type -> status -> count
}

// TaskStats holds task statistics
type TaskStats struct {
	ByTypeAndStatus map[string]map[string]int64 // type -> status -> count
	RunningCount    int64
}

// PoolStatsInfo holds connection pool statistics for a single datasource
type PoolStatsInfo struct {
	DataSourceID string
	TotalConns   int
	IdleConns    int
	InUseConns   int
}

// MetricsCollector periodically collects and updates metrics
type MetricsCollector struct {
	metrics       *Metrics
	statsProvider StatsProvider
	log           *log.Helper
	interval      time.Duration
	startTime     time.Time

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mu     sync.RWMutex
	running bool
}

// MetricsCollectorConfig holds configuration for the metrics collector
type MetricsCollectorConfig struct {
	CollectionInterval time.Duration // How often to collect metrics
}

// DefaultMetricsCollectorConfig returns default configuration
func DefaultMetricsCollectorConfig() *MetricsCollectorConfig {
	return &MetricsCollectorConfig{
		CollectionInterval: 30 * time.Second,
	}
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(
	statsProvider StatsProvider,
	config *MetricsCollectorConfig,
	logger log.Logger,
) *MetricsCollector {
	if config == nil {
		config = DefaultMetricsCollectorConfig()
	}

	return &MetricsCollector{
		metrics:       GetMetrics(),
		statsProvider: statsProvider,
		log:           log.NewHelper(logger),
		interval:      config.CollectionInterval,
		startTime:     time.Now(),
	}
}

// Start starts the metrics collector
func (c *MetricsCollector) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		return nil
	}

	c.ctx, c.cancel = context.WithCancel(ctx)
	c.running = true
	c.startTime = time.Now()

	c.wg.Add(1)
	go c.run()

	c.log.Info("Metrics collector started")
	return nil
}

// Stop stops the metrics collector
func (c *MetricsCollector) Stop() error {
	c.mu.Lock()
	if !c.running {
		c.mu.Unlock()
		return nil
	}
	c.cancel()
	c.running = false
	c.mu.Unlock()

	c.wg.Wait()
	c.log.Info("Metrics collector stopped")
	return nil
}

// IsRunning returns whether the collector is running
func (c *MetricsCollector) IsRunning() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.running
}

// run is the main collection loop
func (c *MetricsCollector) run() {
	defer c.wg.Done()

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	// Collect immediately on start
	c.collect()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.collect()
		}
	}
}

// collect performs a single collection cycle
func (c *MetricsCollector) collect() {
	// Update system uptime
	c.metrics.UpdateUptime(c.startTime)

	// Collect datasource stats
	if c.statsProvider != nil {
		c.collectDataSourceStats()
		c.collectTaskStats()
		c.collectPoolStats()
	}
}

// collectDataSourceStats collects datasource statistics
func (c *MetricsCollector) collectDataSourceStats() {
	ctx, cancel := context.WithTimeout(c.ctx, 10*time.Second)
	defer cancel()

	stats, err := c.statsProvider.GetDataSourceStats(ctx)
	if err != nil {
		c.log.Warnf("Failed to collect datasource stats: %v", err)
		return
	}

	if stats == nil || stats.ByTypeAndStatus == nil {
		return
	}

	for dsType, statusMap := range stats.ByTypeAndStatus {
		for status, count := range statusMap {
			c.metrics.SetDataSourceCount(dsType, status, float64(count))
		}
	}
}

// collectTaskStats collects task statistics
func (c *MetricsCollector) collectTaskStats() {
	ctx, cancel := context.WithTimeout(c.ctx, 10*time.Second)
	defer cancel()

	stats, err := c.statsProvider.GetTaskStats(ctx)
	if err != nil {
		c.log.Warnf("Failed to collect task stats: %v", err)
		return
	}

	if stats == nil {
		return
	}

	if stats.ByTypeAndStatus != nil {
		for taskType, statusMap := range stats.ByTypeAndStatus {
			for status, count := range statusMap {
				c.metrics.SetTaskCount(taskType, status, float64(count))
			}
		}
	}

	c.metrics.TasksRunning.Set(float64(stats.RunningCount))
}

// collectPoolStats collects connection pool statistics
func (c *MetricsCollector) collectPoolStats() {
	ctx, cancel := context.WithTimeout(c.ctx, 10*time.Second)
	defer cancel()

	stats, err := c.statsProvider.GetPoolStats(ctx)
	if err != nil {
		c.log.Warnf("Failed to collect pool stats: %v", err)
		return
	}

	for _, poolStats := range stats {
		c.metrics.UpdatePoolStats(
			poolStats.DataSourceID,
			poolStats.TotalConns,
			poolStats.IdleConns,
			poolStats.InUseConns,
		)
	}
}

// CollectNow triggers an immediate collection
func (c *MetricsCollector) CollectNow() {
	c.collect()
}
