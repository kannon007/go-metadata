import type { ActionType, ProColumns } from '@ant-design/pro-components';
import {
  FooterToolbar,
  PageContainer,
  ProTable,
} from '@ant-design/pro-components';
import { Badge, Button, message, Modal, Space, Tag, Tooltip, Dropdown } from 'antd';
import React, { useRef, useState } from 'react';
import {
  listTasks,
  deleteTask,
  startTask,
  stopTask,
  pauseTask,
  resumeTask,
  executeTask,
  batchStartTasks,
  batchStopTasks,
  batchDeleteTasks,
} from '@/services/task/api';
import { listDataSources } from '@/services/datasource/api';
import CreateForm from './components/CreateForm';
import UpdateForm from './components/UpdateForm';
import ExecutionDrawer from './components/ExecutionDrawer';
import BatchTaskOperations from './components/BatchTaskOperations';
import {
  PlusOutlined,
  DeleteOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  StopOutlined,
  CaretRightOutlined,
  ExclamationCircleOutlined,
  HistoryOutlined,
  MoreOutlined,
} from '@ant-design/icons';

// 任务类型映射
const taskTypeMap: Record<string, { text: string; color: string }> = {
  full_collection: { text: '全量采集', color: 'blue' },
  incremental_collection: { text: '增量采集', color: 'green' },
  schema_only: { text: '仅Schema', color: 'orange' },
  data_profile: { text: '数据画像', color: 'purple' },
};

// 状态映射
const statusMap: Record<string, { text: string; status: 'success' | 'error' | 'default' | 'processing' | 'warning' }> = {
  active: { text: '已启用', status: 'success' },
  inactive: { text: '已禁用', status: 'default' },
  running: { text: '运行中', status: 'processing' },
  completed: { text: '已完成', status: 'success' },
  failed: { text: '失败', status: 'error' },
  paused: { text: '已暂停', status: 'warning' },
};

// 调度器类型映射
const schedulerTypeMap: Record<string, string> = {
  dolphinscheduler: 'DolphinScheduler',
  builtin: '内置调度器',
};

const TaskList: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const [messageApi, contextHolder] = message.useMessage();
  const [selectedRows, setSelectedRows] = useState<TaskAPI.CollectionTask[]>([]);
  const [createModalOpen, setCreateModalOpen] = useState(false);
  const [updateModalOpen, setUpdateModalOpen] = useState(false);
  const [executionDrawerOpen, setExecutionDrawerOpen] = useState(false);
  const [currentRow, setCurrentRow] = useState<TaskAPI.CollectionTask>();
  const [dataSources, setDataSources] = useState<DataSourceAPI.DataSource[]>([]);

  // 加载数据源列表
  const loadDataSources = async () => {
    try {
      const result = await listDataSources({ current: 1, pageSize: 1000 });
      setDataSources(result.data || []);
    } catch (error) {
      console.error('Failed to load datasources:', error);
    }
  };

  // 启动任务
  const handleStart = async (id: string) => {
    try {
      await startTask(id);
      messageApi.success('任务已启动');
      actionRef.current?.reload();
    } catch (error) {
      messageApi.error('启动失败');
    }
  };

  // 停止任务
  const handleStop = async (id: string) => {
    try {
      await stopTask(id);
      messageApi.success('任务已停止');
      actionRef.current?.reload();
    } catch (error) {
      messageApi.error('停止失败');
    }
  };

  // 暂停任务
  const handlePause = async (id: string) => {
    try {
      await pauseTask(id);
      messageApi.success('任务已暂停');
      actionRef.current?.reload();
    } catch (error) {
      messageApi.error('暂停失败');
    }
  };

  // 恢复任务
  const handleResume = async (id: string) => {
    try {
      await resumeTask(id);
      messageApi.success('任务已恢复');
      actionRef.current?.reload();
    } catch (error) {
      messageApi.error('恢复失败');
    }
  };

  // 立即执行
  const handleExecute = async (id: string) => {
    try {
      await executeTask(id);
      messageApi.success('任务已开始执行');
      actionRef.current?.reload();
    } catch (error) {
      messageApi.error('执行失败');
    }
  };

  // 删除任务
  const handleDelete = async (record: TaskAPI.CollectionTask) => {
    Modal.confirm({
      title: '确认删除',
      icon: <ExclamationCircleOutlined />,
      content: `确定要删除任务 "${record.name}" 吗？`,
      okText: '确认',
      cancelText: '取消',
      onOk: async () => {
        try {
          await deleteTask(record.id);
          messageApi.success('删除成功');
          actionRef.current?.reload();
        } catch (error) {
          messageApi.error('删除失败');
        }
      },
    });
  };

  // 批量启动
  const handleBatchStart = async () => {
    if (selectedRows.length === 0) return;
    try {
      const result = await batchStartTasks(selectedRows.map((row) => row.id));
      messageApi.success(`成功启动 ${result.success} 个任务`);
      setSelectedRows([]);
      actionRef.current?.reload();
    } catch (error) {
      messageApi.error('批量启动失败');
    }
  };

  // 批量停止
  const handleBatchStop = async () => {
    if (selectedRows.length === 0) return;
    try {
      const result = await batchStopTasks(selectedRows.map((row) => row.id));
      messageApi.success(`成功停止 ${result.success} 个任务`);
      setSelectedRows([]);
      actionRef.current?.reload();
    } catch (error) {
      messageApi.error('批量停止失败');
    }
  };

  // 批量删除
  const handleBatchDelete = async () => {
    if (selectedRows.length === 0) return;
    Modal.confirm({
      title: '确认批量删除',
      icon: <ExclamationCircleOutlined />,
      content: `确定要删除选中的 ${selectedRows.length} 个任务吗？`,
      okText: '确认',
      cancelText: '取消',
      onOk: async () => {
        try {
          const result = await batchDeleteTasks(selectedRows.map((row) => row.id));
          messageApi.success(`成功删除 ${result.success} 个任务`);
          setSelectedRows([]);
          actionRef.current?.reload();
        } catch (error) {
          messageApi.error('批量删除失败');
        }
      },
    });
  };

  // 获取操作菜单
  const getActionMenu = (record: TaskAPI.CollectionTask) => {
    const items = [];
    if (record.status === 'inactive' || record.status === 'completed' || record.status === 'failed') {
      items.push({ key: 'start', label: '启动', icon: <PlayCircleOutlined />, onClick: () => handleStart(record.id) });
    }
    if (record.status === 'active' || record.status === 'running') {
      items.push({ key: 'pause', label: '暂停', icon: <PauseCircleOutlined />, onClick: () => handlePause(record.id) });
      items.push({ key: 'stop', label: '停止', icon: <StopOutlined />, onClick: () => handleStop(record.id) });
    }
    if (record.status === 'paused') {
      items.push({ key: 'resume', label: '恢复', icon: <CaretRightOutlined />, onClick: () => handleResume(record.id) });
    }
    if (record.status !== 'running') {
      items.push({ key: 'execute', label: '立即执行', icon: <CaretRightOutlined />, onClick: () => handleExecute(record.id) });
    }
    items.push({ key: 'history', label: '执行历史', icon: <HistoryOutlined />, onClick: () => { setCurrentRow(record); setExecutionDrawerOpen(true); } });
    items.push({ key: 'edit', label: '编辑', onClick: () => { setCurrentRow(record); setUpdateModalOpen(true); } });
    items.push({ key: 'delete', label: '删除', danger: true, icon: <DeleteOutlined />, onClick: () => handleDelete(record) });
    return items;
  };

  // 表格列定义
  const columns: ProColumns<TaskAPI.CollectionTask>[] = [
    {
      title: '任务名称',
      dataIndex: 'name',
      render: (_, record) => (
        <a onClick={() => { setCurrentRow(record); setExecutionDrawerOpen(true); }}>
          {record.name}
        </a>
      ),
    },
    {
      title: '数据源',
      dataIndex: 'datasource_id',
      valueType: 'select',
      request: async () => {
        await loadDataSources();
        return dataSources.map((ds) => ({ label: ds.name, value: ds.id }));
      },
      render: (_, record) => {
        const ds = dataSources.find((d) => d.id === record.datasource_id);
        return ds?.name || record.datasource_id;
      },
    },
    {
      title: '任务类型',
      dataIndex: 'type',
      valueType: 'select',
      valueEnum: Object.fromEntries(
        Object.entries(taskTypeMap).map(([key, value]) => [key, { text: value.text }])
      ),
      render: (_, record) => {
        const typeInfo = taskTypeMap[record.type] || { text: record.type, color: 'default' };
        return <Tag color={typeInfo.color}>{typeInfo.text}</Tag>;
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      valueType: 'select',
      valueEnum: {
        active: { text: '已启用', status: 'Success' },
        inactive: { text: '已禁用', status: 'Default' },
        running: { text: '运行中', status: 'Processing' },
        completed: { text: '已完成', status: 'Success' },
        failed: { text: '失败', status: 'Error' },
        paused: { text: '已暂停', status: 'Warning' },
      },
      render: (_, record) => {
        const statusInfo = statusMap[record.status] || { text: record.status, status: 'default' };
        return <Badge status={statusInfo.status} text={statusInfo.text} />;
      },
    },
    {
      title: '调度器',
      dataIndex: 'scheduler_type',
      search: false,
      render: (_, record) => schedulerTypeMap[record.scheduler_type] || record.scheduler_type,
    },
    {
      title: '最后执行时间',
      dataIndex: 'last_executed_at',
      valueType: 'dateTime',
      search: false,
    },
    {
      title: '下次执行时间',
      dataIndex: 'next_execute_at',
      valueType: 'dateTime',
      search: false,
    },
    {
      title: '操作',
      dataIndex: 'option',
      valueType: 'option',
      width: 120,
      render: (_, record) => (
        <Dropdown menu={{ items: getActionMenu(record) }} trigger={['click']}>
          <Button type="link" icon={<MoreOutlined />}>
            操作
          </Button>
        </Dropdown>
      ),
    },
  ];

  return (
    <PageContainer>
      {contextHolder}
      <ProTable<TaskAPI.CollectionTask, TaskAPI.ListTasksParams>
        headerTitle="任务列表"
        actionRef={actionRef}
        rowKey="id"
        search={{ labelWidth: 120 }}
        toolBarRender={() => [
          <BatchTaskOperations
            key="batch"
            selectedIds={selectedRows.map((row) => row.id)}
            onSuccess={() => { setSelectedRows([]); actionRef.current?.reload(); }}
          />,
          <Button
            type="primary"
            key="create"
            icon={<PlusOutlined />}
            onClick={() => { loadDataSources(); setCreateModalOpen(true); }}
          >
            新建任务
          </Button>,
        ]}
        request={async (params) => {
          await loadDataSources();
          return listTasks(params);
        }}
        columns={columns}
        rowSelection={{
          onChange: (_, rows) => setSelectedRows(rows),
        }}
      />

      {selectedRows?.length > 0 && (
        <FooterToolbar
          extra={<div>已选择 <a style={{ fontWeight: 600 }}>{selectedRows.length}</a> 项</div>}
        >
          <Button icon={<PlayCircleOutlined />} onClick={handleBatchStart}>批量启动</Button>
          <Button icon={<StopOutlined />} onClick={handleBatchStop}>批量停止</Button>
          <Button danger icon={<DeleteOutlined />} onClick={handleBatchDelete}>批量删除</Button>
        </FooterToolbar>
      )}

      <CreateForm
        open={createModalOpen}
        onOpenChange={setCreateModalOpen}
        dataSources={dataSources}
        onSuccess={() => { setCreateModalOpen(false); actionRef.current?.reload(); }}
      />

      <UpdateForm
        open={updateModalOpen}
        onOpenChange={setUpdateModalOpen}
        values={currentRow}
        onSuccess={() => { setUpdateModalOpen(false); setCurrentRow(undefined); actionRef.current?.reload(); }}
      />

      <ExecutionDrawer
        open={executionDrawerOpen}
        onClose={() => { setExecutionDrawerOpen(false); setCurrentRow(undefined); }}
        task={currentRow}
      />
    </PageContainer>
  );
};

export default TaskList;
