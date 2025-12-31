import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { ProDescriptions, ProTable } from '@ant-design/pro-components';
import { Badge, Button, Drawer, Modal, Space, Tag, Divider } from 'antd';
import React, { useRef, useState } from 'react';
import { listTaskExecutions, getExecutionLogs } from '@/services/task/api';
import { FileTextOutlined } from '@ant-design/icons';

interface ExecutionDrawerProps {
  open: boolean;
  onClose: () => void;
  task?: TaskAPI.CollectionTask;
}

// 任务类型映射
const taskTypeMap: Record<string, { text: string; color: string }> = {
  full_collection: { text: '全量采集', color: 'blue' },
  incremental_collection: { text: '增量采集', color: 'green' },
  schema_only: { text: '仅Schema', color: 'orange' },
  data_profile: { text: '数据画像', color: 'purple' },
};

// 任务状态映射
const taskStatusMap: Record<string, { text: string; status: 'success' | 'error' | 'default' | 'processing' | 'warning' }> = {
  active: { text: '已启用', status: 'success' },
  inactive: { text: '已禁用', status: 'default' },
  running: { text: '运行中', status: 'processing' },
  completed: { text: '已完成', status: 'success' },
  failed: { text: '失败', status: 'error' },
  paused: { text: '已暂停', status: 'warning' },
};

// 执行状态映射
const executionStatusMap: Record<string, { text: string; status: 'success' | 'error' | 'default' | 'processing' }> = {
  pending: { text: '等待中', status: 'default' },
  running: { text: '运行中', status: 'processing' },
  completed: { text: '已完成', status: 'success' },
  failed: { text: '失败', status: 'error' },
  cancelled: { text: '已取消', status: 'default' },
};

const ExecutionDrawer: React.FC<ExecutionDrawerProps> = ({ open, onClose, task }) => {
  const actionRef = useRef<ActionType>();
  const [logsModalOpen, setLogsModalOpen] = useState(false);
  const [currentLogs, setCurrentLogs] = useState<string>('');
  const [loadingLogs, setLoadingLogs] = useState(false);

  // 查看日志
  const handleViewLogs = async (execution: TaskAPI.TaskExecution) => {
    setLoadingLogs(true);
    try {
      const result = await getExecutionLogs(execution.id);
      setCurrentLogs(result.logs || execution.logs || '暂无日志');
      setLogsModalOpen(true);
    } catch (error) {
      setCurrentLogs(execution.logs || '暂无日志');
      setLogsModalOpen(true);
    } finally {
      setLoadingLogs(false);
    }
  };

  // 格式化持续时间
  const formatDuration = (ms: number) => {
    if (ms < 1000) return `${ms}ms`;
    if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
    if (ms < 3600000) return `${Math.floor(ms / 60000)}m ${Math.floor((ms % 60000) / 1000)}s`;
    return `${Math.floor(ms / 3600000)}h ${Math.floor((ms % 3600000) / 60000)}m`;
  };

  // 执行记录列定义
  const executionColumns: ProColumns<TaskAPI.TaskExecution>[] = [
    {
      title: '执行ID',
      dataIndex: 'id',
      width: 200,
      ellipsis: true,
      copyable: true,
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      render: (_, record) => {
        const statusInfo = executionStatusMap[record.status] || { text: record.status, status: 'default' };
        return <Badge status={statusInfo.status} text={statusInfo.text} />;
      },
    },
    {
      title: '开始时间',
      dataIndex: 'start_time',
      valueType: 'dateTime',
      width: 160,
    },
    {
      title: '结束时间',
      dataIndex: 'end_time',
      valueType: 'dateTime',
      width: 160,
    },
    {
      title: '耗时',
      dataIndex: 'duration',
      width: 100,
      render: (_, record) => formatDuration(record.duration),
    },
    {
      title: '处理表数',
      dataIndex: ['result', 'tables_processed'],
      width: 100,
      render: (_, record) => record.result?.tables_processed ?? '-',
    },
    {
      title: '处理记录数',
      dataIndex: ['result', 'records_processed'],
      width: 120,
      render: (_, record) => record.result?.records_processed?.toLocaleString() ?? '-',
    },
    {
      title: '错误信息',
      dataIndex: 'error_message',
      ellipsis: true,
      render: (_, record) => record.error_message || '-',
    },
    {
      title: '操作',
      dataIndex: 'option',
      valueType: 'option',
      width: 80,
      render: (_, record) => (
        <Button
          type="link"
          size="small"
          icon={<FileTextOutlined />}
          loading={loadingLogs}
          onClick={() => handleViewLogs(record)}
        >
          日志
        </Button>
      ),
    },
  ];

  if (!task) return null;

  const typeInfo = taskTypeMap[task.type] || { text: task.type, color: 'default' };
  const statusInfo = taskStatusMap[task.status] || { text: task.status, status: 'default' };

  return (
    <>
      <Drawer
        title={
          <Space>
            <span>{task.name}</span>
            <Tag color={typeInfo.color}>{typeInfo.text}</Tag>
          </Space>
        }
        width={900}
        open={open}
        onClose={onClose}
      >
        <ProDescriptions column={2} title="任务信息">
          <ProDescriptions.Item label="任务ID" copyable>{task.id}</ProDescriptions.Item>
          <ProDescriptions.Item label="状态">
            <Badge status={statusInfo.status} text={statusInfo.text} />
          </ProDescriptions.Item>
          <ProDescriptions.Item label="调度器">{task.scheduler_type === 'builtin' ? '内置调度器' : 'DolphinScheduler'}</ProDescriptions.Item>
          <ProDescriptions.Item label="创建者">{task.created_by || '-'}</ProDescriptions.Item>
          <ProDescriptions.Item label="最后执行时间" valueType="dateTime">{task.last_executed_at || '-'}</ProDescriptions.Item>
          <ProDescriptions.Item label="下次执行时间" valueType="dateTime">{task.next_execute_at || '-'}</ProDescriptions.Item>
        </ProDescriptions>

        {task.schedule && (
          <>
            <Divider />
            <ProDescriptions column={2} title="调度配置">
              <ProDescriptions.Item label="调度类型">
                {task.schedule.type === 'once' ? '单次执行' : task.schedule.type === 'cron' ? 'Cron表达式' : '固定间隔'}
              </ProDescriptions.Item>
              {task.schedule.cron_expr && <ProDescriptions.Item label="Cron表达式">{task.schedule.cron_expr}</ProDescriptions.Item>}
              {task.schedule.interval && <ProDescriptions.Item label="执行间隔">{task.schedule.interval}秒</ProDescriptions.Item>}
              <ProDescriptions.Item label="时区">{task.schedule.timezone}</ProDescriptions.Item>
            </ProDescriptions>
          </>
        )}

        <Divider />

        <ProTable<TaskAPI.TaskExecution>
          headerTitle="执行历史"
          actionRef={actionRef}
          rowKey="id"
          search={false}
          options={false}
          pagination={{ pageSize: 10 }}
          request={async (params) => listTaskExecutions(task.id, { current: params.current, pageSize: params.pageSize })}
          columns={executionColumns}
        />
      </Drawer>

      <Modal
        title="执行日志"
        open={logsModalOpen}
        onCancel={() => setLogsModalOpen(false)}
        footer={null}
        width={800}
      >
        <pre style={{ maxHeight: 500, overflow: 'auto', background: '#f5f5f5', padding: 16, borderRadius: 4 }}>
          {currentLogs}
        </pre>
      </Modal>
    </>
  );
};

export default ExecutionDrawer;
