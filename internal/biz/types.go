package biz

import "time"

// DataSourceType represents the type of data source.
type DataSourceType = string

// DataSourceStatus represents the status of data source.
type DataSourceStatus = string

// TaskType represents the type of task.
type TaskType = string

// TaskStatus represents the status of task.
type TaskStatus = string

// ScheduleType represents the type of schedule.
type ScheduleType = string

// SchedulerType represents the type of scheduler.
type SchedulerType = string

// ExecutionStatus represents the status of execution.
type ExecutionStatus = string

// ConnectionConfig represents connection configuration.
type ConnectionConfig struct {
	Host         string
	Port         int32
	Database     string
	Username     string
	Password     string
	SSL          bool
	SSLMode      string
	Timeout      int32
	MaxConns     int32
	MaxIdleConns int32
	Charset      string
	Extra        map[string]string
}

// DataSource represents a data source entity.
type DataSource struct {
	ID          string
	Name        string
	Type        DataSourceType
	Description string
	Config      *ConnectionConfig
	Status      DataSourceStatus
	Tags        []string
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	LastTestAt  time.Time
}

// ConnectionTestResult represents the result of connection test.
type ConnectionTestResult struct {
	Success      bool
	Message      string
	Latency      int64
	ServerInfo   string
	DatabaseInfo string
	Version      string
}

// DataSourceStats represents data source statistics.
type DataSourceStats struct {
	TotalCount    int64
	ActiveCount   int64
	InactiveCount int64
	ErrorCount    int64
}

// CanDeleteResult represents the result of can delete check.
type CanDeleteResult struct {
	CanDelete       bool
	Reason          string
	AssociatedTasks int32
}

// BatchOperationResult represents the result of batch operation.
type BatchOperationResult struct {
	Total   int32
	Success int32
	Failed  int32
	Errors  []string
}

// TaskConfig represents task configuration.
type TaskConfig struct {
	IncludeSchemas    []string
	ExcludeSchemas    []string
	IncludeTables     []string
	ExcludeTables     []string
	BatchSize         int32
	Timeout           int32
	RetryCount        int32
	RetryInterval     int32
	IncrementalColumn string
	IncrementalValue  string
	Extra             map[string]string
}

// ScheduleConfig represents schedule configuration.
type ScheduleConfig struct {
	Type      ScheduleType
	CronExpr  string
	Interval  int32
	StartTime time.Time
	EndTime   time.Time
	Timezone  string
}

// CollectionTask represents a collection task entity.
type CollectionTask struct {
	ID             string
	Name           string
	DataSourceID   string
	Type           TaskType
	Config         *TaskConfig
	Schedule       *ScheduleConfig
	Status         TaskStatus
	SchedulerType  SchedulerType
	ExternalID     string
	CreatedBy      string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	LastExecutedAt time.Time
	NextExecuteAt  time.Time
}

// TaskExecution represents a task execution record.
type TaskExecution struct {
	ID           string
	TaskID       string
	Status       ExecutionStatus
	StartTime    time.Time
	EndTime      time.Time
	Duration     int64
	ErrorMessage string
	Logs         string
	ExternalID   string
	CreatedAt    time.Time
}

// TaskStatusInfo represents task status information.
type TaskStatusInfo struct {
	TaskID         string
	Status         TaskStatus
	LastExecution  *TaskExecution
	NextExecuteAt  time.Time
	AllowedActions []string
}

// DataSourceTemplate represents a data source template.
type DataSourceTemplate struct {
	ID             string
	Name           string
	Type           DataSourceType
	Description    string
	ConfigTemplate *ConnectionConfig
	IsSystem       bool
	CreatedBy      string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
