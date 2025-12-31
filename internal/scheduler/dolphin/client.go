package dolphin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client DolphinScheduler API客户端
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
	projectID  int64
}

// ClientConfig 客户端配置
type ClientConfig struct {
	Endpoint string
	Token    string
	Timeout  time.Duration
}

// NewClient 创建DolphinScheduler客户端
func NewClient(config *ClientConfig) *Client {
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &Client{
		baseURL: strings.TrimSuffix(config.Endpoint, "/"),
		token:   config.Token,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// SetProjectID 设置项目ID
func (c *Client) SetProjectID(projectID int64) {
	c.projectID = projectID
}

// APIResponse API响应
type APIResponse struct {
	Code    int             `json:"code"`
	Msg     string          `json:"msg"`
	Data    json.RawMessage `json:"data"`
	Success bool            `json:"success"`
}

// doRequest 执行HTTP请求
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) (*APIResponse, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	reqURL := fmt.Sprintf("%s%s", c.baseURL, path)
	req, err := http.NewRequestWithContext(ctx, method, reqURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("token", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if apiResp.Code != 0 {
		return nil, fmt.Errorf("API error: %s (code: %d)", apiResp.Msg, apiResp.Code)
	}

	return &apiResp, nil
}

// doFormRequest 执行表单请求
func (c *Client) doFormRequest(ctx context.Context, method, path string, params url.Values) (*APIResponse, error) {
	reqURL := fmt.Sprintf("%s%s", c.baseURL, path)
	req, err := http.NewRequestWithContext(ctx, method, reqURL, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("token", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if apiResp.Code != 0 {
		return nil, fmt.Errorf("API error: %s (code: %d)", apiResp.Msg, apiResp.Code)
	}

	return &apiResp, nil
}

// Project 项目
type Project struct {
	ID          int64     `json:"id"`
	Code        int64     `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreateTime  time.Time `json:"createTime"`
	UpdateTime  time.Time `json:"updateTime"`
}

// CreateProject 创建项目
func (c *Client) CreateProject(ctx context.Context, name, description string) (*Project, error) {
	params := url.Values{}
	params.Set("projectName", name)
	params.Set("description", description)

	resp, err := c.doFormRequest(ctx, http.MethodPost, "/projects", params)
	if err != nil {
		return nil, err
	}

	var project Project
	if err := json.Unmarshal(resp.Data, &project); err != nil {
		return nil, fmt.Errorf("failed to unmarshal project: %w", err)
	}

	return &project, nil
}

// GetProject 获取项目
func (c *Client) GetProject(ctx context.Context, projectCode int64) (*Project, error) {
	path := fmt.Sprintf("/projects/%d", projectCode)
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var project Project
	if err := json.Unmarshal(resp.Data, &project); err != nil {
		return nil, fmt.Errorf("failed to unmarshal project: %w", err)
	}

	return &project, nil
}

// ListProjects 列出项目
func (c *Client) ListProjects(ctx context.Context, pageNo, pageSize int) ([]*Project, error) {
	path := fmt.Sprintf("/projects?pageNo=%d&pageSize=%d", pageNo, pageSize)
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		TotalList []*Project `json:"totalList"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal projects: %w", err)
	}

	return result.TotalList, nil
}


// ProcessDefinition 工作流定义
type ProcessDefinition struct {
	ID          int64     `json:"id"`
	Code        int64     `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ProjectCode int64     `json:"projectCode"`
	ReleaseState string   `json:"releaseState"` // ONLINE, OFFLINE
	CreateTime  time.Time `json:"createTime"`
	UpdateTime  time.Time `json:"updateTime"`
}

// TaskDefinition 任务定义
type TaskDefinition struct {
	Code        int64             `json:"code"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	TaskType    string            `json:"taskType"`
	TaskParams  map[string]interface{} `json:"taskParams"`
	Flag        string            `json:"flag"` // YES, NO
	Priority    string            `json:"taskPriority"` // HIGHEST, HIGH, MEDIUM, LOW, LOWEST
	FailRetryTimes int            `json:"failRetryTimes"`
	FailRetryInterval int         `json:"failRetryInterval"`
	TimeoutFlag string            `json:"timeoutFlag"` // OPEN, CLOSE
	Timeout     int               `json:"timeout"`
}

// CreateProcessDefinitionRequest 创建工作流定义请求
type CreateProcessDefinitionRequest struct {
	Name            string            `json:"name"`
	Description     string            `json:"description"`
	Locations       string            `json:"locations"`
	TaskDefinitionJson string         `json:"taskDefinitionJson"`
	TaskRelationJson string           `json:"taskRelationJson"`
	Timeout         int               `json:"timeout"`
	GlobalParams    string            `json:"globalParams"`
	ExecutionType   string            `json:"executionType"` // PARALLEL, SERIAL_WAIT, SERIAL_DISCARD, SERIAL_PRIORITY
}

// CreateProcessDefinition 创建工作流定义
func (c *Client) CreateProcessDefinition(ctx context.Context, projectCode int64, req *CreateProcessDefinitionRequest) (*ProcessDefinition, error) {
	params := url.Values{}
	params.Set("name", req.Name)
	params.Set("description", req.Description)
	params.Set("locations", req.Locations)
	params.Set("taskDefinitionJson", req.TaskDefinitionJson)
	params.Set("taskRelationJson", req.TaskRelationJson)
	params.Set("timeout", fmt.Sprintf("%d", req.Timeout))
	params.Set("globalParams", req.GlobalParams)
	params.Set("executionType", req.ExecutionType)

	path := fmt.Sprintf("/projects/%d/process-definition", projectCode)
	resp, err := c.doFormRequest(ctx, http.MethodPost, path, params)
	if err != nil {
		return nil, err
	}

	var pd ProcessDefinition
	if err := json.Unmarshal(resp.Data, &pd); err != nil {
		return nil, fmt.Errorf("failed to unmarshal process definition: %w", err)
	}

	return &pd, nil
}

// UpdateProcessDefinition 更新工作流定义
func (c *Client) UpdateProcessDefinition(ctx context.Context, projectCode, processCode int64, req *CreateProcessDefinitionRequest) (*ProcessDefinition, error) {
	params := url.Values{}
	params.Set("name", req.Name)
	params.Set("description", req.Description)
	params.Set("locations", req.Locations)
	params.Set("taskDefinitionJson", req.TaskDefinitionJson)
	params.Set("taskRelationJson", req.TaskRelationJson)
	params.Set("timeout", fmt.Sprintf("%d", req.Timeout))
	params.Set("globalParams", req.GlobalParams)
	params.Set("executionType", req.ExecutionType)

	path := fmt.Sprintf("/projects/%d/process-definition/%d", projectCode, processCode)
	resp, err := c.doFormRequest(ctx, http.MethodPut, path, params)
	if err != nil {
		return nil, err
	}

	var pd ProcessDefinition
	if err := json.Unmarshal(resp.Data, &pd); err != nil {
		return nil, fmt.Errorf("failed to unmarshal process definition: %w", err)
	}

	return &pd, nil
}

// DeleteProcessDefinition 删除工作流定义
func (c *Client) DeleteProcessDefinition(ctx context.Context, projectCode, processCode int64) error {
	path := fmt.Sprintf("/projects/%d/process-definition/%d", projectCode, processCode)
	_, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	return err
}

// GetProcessDefinition 获取工作流定义
func (c *Client) GetProcessDefinition(ctx context.Context, projectCode, processCode int64) (*ProcessDefinition, error) {
	path := fmt.Sprintf("/projects/%d/process-definition/%d", projectCode, processCode)
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var pd ProcessDefinition
	if err := json.Unmarshal(resp.Data, &pd); err != nil {
		return nil, fmt.Errorf("failed to unmarshal process definition: %w", err)
	}

	return &pd, nil
}

// ReleaseProcessDefinition 上线/下线工作流定义
func (c *Client) ReleaseProcessDefinition(ctx context.Context, projectCode, processCode int64, releaseState string) error {
	params := url.Values{}
	params.Set("releaseState", releaseState)

	path := fmt.Sprintf("/projects/%d/process-definition/%d/release", projectCode, processCode)
	_, err := c.doFormRequest(ctx, http.MethodPost, path, params)
	return err
}

// ProcessInstance 工作流实例
type ProcessInstance struct {
	ID              int64     `json:"id"`
	ProcessDefinitionCode int64 `json:"processDefinitionCode"`
	Name            string    `json:"name"`
	State           string    `json:"state"` // SUBMITTED_SUCCESS, RUNNING_EXECUTION, READY_PAUSE, PAUSE, READY_STOP, STOP, FAILURE, SUCCESS, NEED_FAULT_TOLERANCE, KILL, WAITING_THREAD, WAITING_DEPEND, DELAY_EXECUTION, FORCED_SUCCESS, SERIAL_WAIT
	StartTime       time.Time `json:"startTime"`
	EndTime         *time.Time `json:"endTime"`
	Duration        string    `json:"duration"`
	RunTimes        int       `json:"runTimes"`
}

// StartProcessInstanceRequest 启动工作流实例请求
type StartProcessInstanceRequest struct {
	ProcessDefinitionCode int64             `json:"processDefinitionCode"`
	ScheduleTime          string            `json:"scheduleTime"`
	FailureStrategy       string            `json:"failureStrategy"` // CONTINUE, END
	WarningType           string            `json:"warningType"`     // NONE, SUCCESS, FAILURE, ALL
	WarningGroupId        int64             `json:"warningGroupId"`
	ExecType              string            `json:"execType"`        // NONE, REPEAT_RUNNING, RECOVER_SUSPENDED_PROCESS, START_FAILURE_TASK_PROCESS, COMPLEMENT_DATA, SCHEDULER, RECOVER_WAITING_THREAD, RECOVER_SERIAL_WAIT
	StartParams           map[string]string `json:"startParams"`
	RunMode               string            `json:"runMode"`         // RUN_MODE_SERIAL, RUN_MODE_PARALLEL
	DryRun                int               `json:"dryRun"`          // 0: normal, 1: dry run
}

// StartProcessInstance 启动工作流实例
func (c *Client) StartProcessInstance(ctx context.Context, projectCode int64, req *StartProcessInstanceRequest) (*ProcessInstance, error) {
	params := url.Values{}
	params.Set("processDefinitionCode", fmt.Sprintf("%d", req.ProcessDefinitionCode))
	params.Set("failureStrategy", req.FailureStrategy)
	params.Set("warningType", req.WarningType)
	if req.ScheduleTime != "" {
		params.Set("scheduleTime", req.ScheduleTime)
	}
	if req.ExecType != "" {
		params.Set("execType", req.ExecType)
	}
	if req.RunMode != "" {
		params.Set("runMode", req.RunMode)
	}
	params.Set("dryRun", fmt.Sprintf("%d", req.DryRun))

	// 序列化启动参数
	if len(req.StartParams) > 0 {
		startParamsJSON, _ := json.Marshal(req.StartParams)
		params.Set("startParams", string(startParamsJSON))
	}

	path := fmt.Sprintf("/projects/%d/executors/start-process-instance", projectCode)
	resp, err := c.doFormRequest(ctx, http.MethodPost, path, params)
	if err != nil {
		return nil, err
	}

	var pi ProcessInstance
	if err := json.Unmarshal(resp.Data, &pi); err != nil {
		return nil, fmt.Errorf("failed to unmarshal process instance: %w", err)
	}

	return &pi, nil
}

// StopProcessInstance 停止工作流实例
func (c *Client) StopProcessInstance(ctx context.Context, projectCode int64, processInstanceID int64) error {
	params := url.Values{}
	params.Set("processInstanceId", fmt.Sprintf("%d", processInstanceID))
	params.Set("executeType", "STOP")

	path := fmt.Sprintf("/projects/%d/executors/execute", projectCode)
	_, err := c.doFormRequest(ctx, http.MethodPost, path, params)
	return err
}

// PauseProcessInstance 暂停工作流实例
func (c *Client) PauseProcessInstance(ctx context.Context, projectCode int64, processInstanceID int64) error {
	params := url.Values{}
	params.Set("processInstanceId", fmt.Sprintf("%d", processInstanceID))
	params.Set("executeType", "PAUSE")

	path := fmt.Sprintf("/projects/%d/executors/execute", projectCode)
	_, err := c.doFormRequest(ctx, http.MethodPost, path, params)
	return err
}

// GetProcessInstance 获取工作流实例
func (c *Client) GetProcessInstance(ctx context.Context, projectCode int64, processInstanceID int64) (*ProcessInstance, error) {
	path := fmt.Sprintf("/projects/%d/process-instances/%d", projectCode, processInstanceID)
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var pi ProcessInstance
	if err := json.Unmarshal(resp.Data, &pi); err != nil {
		return nil, fmt.Errorf("failed to unmarshal process instance: %w", err)
	}

	return &pi, nil
}

// TaskInstance 任务实例
type TaskInstance struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	TaskType        string    `json:"taskType"`
	State           string    `json:"state"`
	StartTime       time.Time `json:"startTime"`
	EndTime         *time.Time `json:"endTime"`
	Duration        string    `json:"duration"`
	RetryTimes      int       `json:"retryTimes"`
}

// GetTaskInstanceLog 获取任务实例日志
func (c *Client) GetTaskInstanceLog(ctx context.Context, projectCode int64, taskInstanceID int64, skipLineNum, limit int) (string, error) {
	path := fmt.Sprintf("/projects/%d/task-instances/%d/log-detail?skipLineNum=%d&limit=%d",
		projectCode, taskInstanceID, skipLineNum, limit)
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return "", err
	}

	var result struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return "", fmt.Errorf("failed to unmarshal log: %w", err)
	}

	return result.Message, nil
}

// Schedule 调度配置
type Schedule struct {
	ID                    int64  `json:"id"`
	ProcessDefinitionCode int64  `json:"processDefinitionCode"`
	StartTime             string `json:"startTime"`
	EndTime               string `json:"endTime"`
	Crontab               string `json:"crontab"`
	TimezoneId            string `json:"timezoneId"`
	ReleaseState          string `json:"releaseState"` // ONLINE, OFFLINE
}

// CreateScheduleRequest 创建调度请求
type CreateScheduleRequest struct {
	ProcessDefinitionCode int64  `json:"processDefinitionCode"`
	Schedule              string `json:"schedule"` // JSON格式的调度配置
	WarningType           string `json:"warningType"`
	WarningGroupId        int64  `json:"warningGroupId"`
	FailureStrategy       string `json:"failureStrategy"`
	WorkerGroup           string `json:"workerGroup"`
	EnvironmentCode       int64  `json:"environmentCode"`
}

// CreateSchedule 创建调度
func (c *Client) CreateSchedule(ctx context.Context, projectCode int64, req *CreateScheduleRequest) (*Schedule, error) {
	params := url.Values{}
	params.Set("processDefinitionCode", fmt.Sprintf("%d", req.ProcessDefinitionCode))
	params.Set("schedule", req.Schedule)
	params.Set("warningType", req.WarningType)
	params.Set("failureStrategy", req.FailureStrategy)
	if req.WorkerGroup != "" {
		params.Set("workerGroup", req.WorkerGroup)
	}

	path := fmt.Sprintf("/projects/%d/schedules", projectCode)
	resp, err := c.doFormRequest(ctx, http.MethodPost, path, params)
	if err != nil {
		return nil, err
	}

	var schedule Schedule
	if err := json.Unmarshal(resp.Data, &schedule); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schedule: %w", err)
	}

	return &schedule, nil
}

// UpdateSchedule 更新调度
func (c *Client) UpdateSchedule(ctx context.Context, projectCode int64, scheduleID int64, req *CreateScheduleRequest) (*Schedule, error) {
	params := url.Values{}
	params.Set("schedule", req.Schedule)
	params.Set("warningType", req.WarningType)
	params.Set("failureStrategy", req.FailureStrategy)
	if req.WorkerGroup != "" {
		params.Set("workerGroup", req.WorkerGroup)
	}

	path := fmt.Sprintf("/projects/%d/schedules/%d", projectCode, scheduleID)
	resp, err := c.doFormRequest(ctx, http.MethodPut, path, params)
	if err != nil {
		return nil, err
	}

	var schedule Schedule
	if err := json.Unmarshal(resp.Data, &schedule); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schedule: %w", err)
	}

	return &schedule, nil
}

// OnlineSchedule 上线调度
func (c *Client) OnlineSchedule(ctx context.Context, projectCode int64, scheduleID int64) error {
	path := fmt.Sprintf("/projects/%d/schedules/%d/online", projectCode, scheduleID)
	_, err := c.doFormRequest(ctx, http.MethodPost, path, nil)
	return err
}

// OfflineSchedule 下线调度
func (c *Client) OfflineSchedule(ctx context.Context, projectCode int64, scheduleID int64) error {
	path := fmt.Sprintf("/projects/%d/schedules/%d/offline", projectCode, scheduleID)
	_, err := c.doFormRequest(ctx, http.MethodPost, path, nil)
	return err
}

// DeleteSchedule 删除调度
func (c *Client) DeleteSchedule(ctx context.Context, projectCode int64, scheduleID int64) error {
	path := fmt.Sprintf("/projects/%d/schedules/%d", projectCode, scheduleID)
	_, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	return err
}
