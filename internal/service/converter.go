package service

import (
	v1 "go-metadata/api/metadata/v1"
	"go-metadata/internal/biz"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// toBizConnectionConfig converts proto ConnectionConfig to biz.
func toBizConnectionConfig(cfg *v1.ConnectionConfig) *biz.ConnectionConfig {
	if cfg == nil {
		return nil
	}
	return &biz.ConnectionConfig{
		Host:         cfg.Host,
		Port:         cfg.Port,
		Database:     cfg.Database,
		Username:     cfg.Username,
		Password:     cfg.Password,
		SSL:          cfg.Ssl,
		SSLMode:      cfg.SslMode,
		Timeout:      cfg.Timeout,
		MaxConns:     cfg.MaxConns,
		MaxIdleConns: cfg.MaxIdleConns,
		Charset:      cfg.Charset,
		Extra:        cfg.Extra,
	}
}

// toProtoConnectionConfig converts biz ConnectionConfig to proto.
func toProtoConnectionConfig(cfg *biz.ConnectionConfig) *v1.ConnectionConfig {
	if cfg == nil {
		return nil
	}
	return &v1.ConnectionConfig{
		Host:         cfg.Host,
		Port:         cfg.Port,
		Database:     cfg.Database,
		Username:     cfg.Username,
		Password:     "******", // hide password
		Ssl:          cfg.SSL,
		SslMode:      cfg.SSLMode,
		Timeout:      cfg.Timeout,
		MaxConns:     cfg.MaxConns,
		MaxIdleConns: cfg.MaxIdleConns,
		Charset:      cfg.Charset,
		Extra:        cfg.Extra,
	}
}

// toProtoDataSource converts biz DataSource to proto.
func toProtoDataSource(ds *biz.DataSource) *v1.DataSource {
	if ds == nil {
		return nil
	}
	return &v1.DataSource{
		Id:          ds.ID,
		Name:        ds.Name,
		Type:        v1.DataSourceType(v1.DataSourceType_value[ds.Type]),
		Description: ds.Description,
		Config:      toProtoConnectionConfig(ds.Config),
		Status:      v1.DataSourceStatus(v1.DataSourceStatus_value[ds.Status]),
		Tags:        ds.Tags,
		CreatedBy:   ds.CreatedBy,
		CreatedAt:   timestamppb.New(ds.CreatedAt),
		UpdatedAt:   timestamppb.New(ds.UpdatedAt),
		LastTestAt:  timestamppb.New(ds.LastTestAt),
	}
}

// toProtoDataSources converts a list of biz DataSource to proto.
func toProtoDataSources(list []*biz.DataSource) []*v1.DataSource {
	result := make([]*v1.DataSource, 0, len(list))
	for _, ds := range list {
		result = append(result, toProtoDataSource(ds))
	}
	return result
}

// toProtoConnectionTestResult converts biz ConnectionTestResult to proto.
func toProtoConnectionTestResult(r *biz.ConnectionTestResult) *v1.ConnectionTestResult {
	if r == nil {
		return nil
	}
	return &v1.ConnectionTestResult{
		Success:      r.Success,
		Message:      r.Message,
		Latency:      r.Latency,
		ServerInfo:   r.ServerInfo,
		DatabaseInfo: r.DatabaseInfo,
		Version:      r.Version,
	}
}

// toBizTaskConfig converts proto TaskConfig to biz.
func toBizTaskConfig(cfg *v1.TaskConfig) *biz.TaskConfig {
	if cfg == nil {
		return nil
	}
	return &biz.TaskConfig{
		IncludeSchemas:    cfg.IncludeSchemas,
		ExcludeSchemas:    cfg.ExcludeSchemas,
		IncludeTables:     cfg.IncludeTables,
		ExcludeTables:     cfg.ExcludeTables,
		BatchSize:         cfg.BatchSize,
		Timeout:           cfg.Timeout,
		RetryCount:        cfg.RetryCount,
		RetryInterval:     cfg.RetryInterval,
		IncrementalColumn: cfg.IncrementalColumn,
		IncrementalValue:  cfg.IncrementalValue,
		Extra:             cfg.Extra,
	}
}

// toProtoTaskConfig converts biz TaskConfig to proto.
func toProtoTaskConfig(cfg *biz.TaskConfig) *v1.TaskConfig {
	if cfg == nil {
		return nil
	}
	return &v1.TaskConfig{
		IncludeSchemas:    cfg.IncludeSchemas,
		ExcludeSchemas:    cfg.ExcludeSchemas,
		IncludeTables:     cfg.IncludeTables,
		ExcludeTables:     cfg.ExcludeTables,
		BatchSize:         cfg.BatchSize,
		Timeout:           cfg.Timeout,
		RetryCount:        cfg.RetryCount,
		RetryInterval:     cfg.RetryInterval,
		IncrementalColumn: cfg.IncrementalColumn,
		IncrementalValue:  cfg.IncrementalValue,
		Extra:             cfg.Extra,
	}
}

// toBizScheduleConfig converts proto ScheduleConfig to biz.
func toBizScheduleConfig(cfg *v1.ScheduleConfig) *biz.ScheduleConfig {
	if cfg == nil {
		return nil
	}
	return &biz.ScheduleConfig{
		Type:     biz.ScheduleType(cfg.Type.String()),
		CronExpr: cfg.CronExpr,
		Interval: cfg.Interval,
		Timezone: cfg.Timezone,
	}
}

// toProtoScheduleConfig converts biz ScheduleConfig to proto.
func toProtoScheduleConfig(cfg *biz.ScheduleConfig) *v1.ScheduleConfig {
	if cfg == nil {
		return nil
	}
	return &v1.ScheduleConfig{
		Type:     v1.ScheduleType(v1.ScheduleType_value[cfg.Type]),
		CronExpr: cfg.CronExpr,
		Interval: cfg.Interval,
		Timezone: cfg.Timezone,
	}
}

// toProtoTask converts biz CollectionTask to proto.
func toProtoTask(task *biz.CollectionTask) *v1.CollectionTask {
	if task == nil {
		return nil
	}
	return &v1.CollectionTask{
		Id:             task.ID,
		Name:           task.Name,
		DatasourceId:   task.DataSourceID,
		Type:           v1.TaskType(v1.TaskType_value[task.Type]),
		Config:         toProtoTaskConfig(task.Config),
		Schedule:       toProtoScheduleConfig(task.Schedule),
		Status:         v1.TaskStatus(v1.TaskStatus_value[task.Status]),
		SchedulerType:  v1.SchedulerType(v1.SchedulerType_value[task.SchedulerType]),
		ExternalId:     task.ExternalID,
		CreatedBy:      task.CreatedBy,
		CreatedAt:      timestamppb.New(task.CreatedAt),
		UpdatedAt:      timestamppb.New(task.UpdatedAt),
		LastExecutedAt: timestamppb.New(task.LastExecutedAt),
		NextExecuteAt:  timestamppb.New(task.NextExecuteAt),
	}
}

// toProtoTasks converts a list of biz CollectionTask to proto.
func toProtoTasks(list []*biz.CollectionTask) []*v1.CollectionTask {
	result := make([]*v1.CollectionTask, 0, len(list))
	for _, task := range list {
		result = append(result, toProtoTask(task))
	}
	return result
}

// toProtoExecution converts biz TaskExecution to proto.
func toProtoExecution(exec *biz.TaskExecution) *v1.TaskExecution {
	if exec == nil {
		return nil
	}
	return &v1.TaskExecution{
		Id:           exec.ID,
		TaskId:       exec.TaskID,
		Status:       v1.ExecutionStatus(v1.ExecutionStatus_value[exec.Status]),
		StartTime:    timestamppb.New(exec.StartTime),
		EndTime:      timestamppb.New(exec.EndTime),
		Duration:     exec.Duration,
		ErrorMessage: exec.ErrorMessage,
		Logs:         exec.Logs,
		ExternalId:   exec.ExternalID,
		CreatedAt:    timestamppb.New(exec.CreatedAt),
	}
}

// toProtoExecutions converts a list of biz TaskExecution to proto.
func toProtoExecutions(list []*biz.TaskExecution) []*v1.TaskExecution {
	result := make([]*v1.TaskExecution, 0, len(list))
	for _, exec := range list {
		result = append(result, toProtoExecution(exec))
	}
	return result
}

// toProtoTemplate converts biz DataSourceTemplate to proto.
func toProtoTemplate(tpl *biz.DataSourceTemplate) *v1.DataSourceTemplate {
	if tpl == nil {
		return nil
	}
	return &v1.DataSourceTemplate{
		Id:             tpl.ID,
		Name:           tpl.Name,
		Type:           v1.DataSourceType(v1.DataSourceType_value[tpl.Type]),
		Description:    tpl.Description,
		ConfigTemplate: toProtoConnectionConfig(tpl.ConfigTemplate),
		IsSystem:       tpl.IsSystem,
		CreatedBy:      tpl.CreatedBy,
		CreatedAt:      timestamppb.New(tpl.CreatedAt),
		UpdatedAt:      timestamppb.New(tpl.UpdatedAt),
	}
}

// toProtoTemplates converts a list of biz DataSourceTemplate to proto.
func toProtoTemplates(list []*biz.DataSourceTemplate) []*v1.DataSourceTemplate {
	result := make([]*v1.DataSourceTemplate, 0, len(list))
	for _, tpl := range list {
		result = append(result, toProtoTemplate(tpl))
	}
	return result
}
