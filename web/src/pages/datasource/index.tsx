import type { ActionType, ProColumns } from '@ant-design/pro-components';
import {
  FooterToolbar,
  PageContainer,
  ProTable,
} from '@ant-design/pro-components';
import { useIntl } from '@umijs/max';
import { Badge, Button, Drawer, message, Modal, Space, Tag, Tooltip } from 'antd';
import React, { useRef, useState } from 'react';
import {
  listDataSources,
  deleteDataSource,
  testDataSourceConnection,
  canDeleteDataSource,
  batchUpdateStatus,
  batchDelete,
} from '@/services/datasource/api';
import CreateForm from './components/CreateForm';
import UpdateForm from './components/UpdateForm';
import DetailDrawer from './components/DetailDrawer';
import BatchOperations from './components/BatchOperations';
import {
  PlusOutlined,
  DeleteOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  ApiOutlined,
  ExclamationCircleOutlined,
} from '@ant-design/icons';

// 数据源类型映射
const dataSourceTypeMap: Record<string, { text: string; color: string }> = {
  mysql: { text: 'MySQL', color: 'blue' },
  postgresql: { text: 'PostgreSQL', color: 'cyan' },
  oracle: { text: 'Oracle', color: 'red' },
  sqlserver: { text: 'SQL Server', color: 'purple' },
  mongodb: { text: 'MongoDB', color: 'green' },
  redis: { text: 'Redis', color: 'volcano' },
  kafka: { text: 'Kafka', color: 'orange' },
  rabbitmq: { text: 'RabbitMQ', color: 'gold' },
  minio: { text: 'MinIO', color: 'lime' },
  clickhouse: { text: 'ClickHouse', color: 'geekblue' },
  doris: { text: 'Doris', color: 'magenta' },
  hive: { text: 'Hive', color: 'yellow' },
  elasticsearch: { text: 'Elasticsearch', color: 'cyan' },
};

// 状态映射
const statusMap: Record<string, { text: string; status: 'success' | 'error' | 'default' | 'processing' }> = {
  active: { text: '在线', status: 'success' },
  inactive: { text: '离线', status: 'default' },
  error: { text: '错误', status: 'error' },
  testing: { text: '测试中', status: 'processing' },
};

const DataSourceList: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const intl = useIntl();
  const [messageApi, contextHolder] = message.useMessage();
  const [selectedRows, setSelectedRows] = useState<DataSourceAPI.DataSource[]>([]);
  const [createModalOpen, setCreateModalOpen] = useState(false);
  const [updateModalOpen, setUpdateModalOpen] = useState(false);
  const [detailDrawerOpen, setDetailDrawerOpen] = useState(false);
  const [currentRow, setCurrentRow] = useState<DataSourceAPI.DataSource>();
  const [testingId, setTestingId] = useState<string>();

  // 测试连接
  const handleTestConnection = async (id: string) => {
    setTestingId(id);
    try {
      const result = await testDataSourceConnection(id);
      if (result.success) {
        messageApi.success(`连接成功: ${result.message}`);
      } else {
        messageApi.error(`连接失败: ${result.message}`);
      }
      actionRef.current?.reload();
    } catch (error) {
      messageApi.error('测试连接失败');
    } finally {
      setTestingId(undefined);
    }
  };

  // 删除数据源
  const handleDelete = async (record: DataSourceAPI.DataSource) => {
    try {
      const canDelete = await canDeleteDataSource(record.id);
      if (!canDelete.can_delete) {
        Modal.warning({
          title: '无法删除',
          content: canDelete.reason || `该数据源有 ${canDelete.associated_tasks} 个关联任务，请先删除关联任务`,
        });
        return;
      }

      Modal.confirm({
        title: '确认删除',
        icon: <ExclamationCircleOutlined />,
        content: `确定要删除数据源 "${record.name}" 吗？`,
        okText: '确认',
        cancelText: '取消',
        onOk: async () => {
          await deleteDataSource(record.id);
          messageApi.success('删除成功');
          actionRef.current?.reload();
        },
      });
    } catch (error) {
      messageApi.error('删除失败');
    }
  };

  // 批量启用
  const handleBatchEnable = async () => {
    if (selectedRows.length === 0) return;
    try {
      const result = await batchUpdateStatus({
        ids: selectedRows.map((row) => row.id),
        status: 'active',
      });
      messageApi.success(`成功启用 ${result.success} 个数据源`);
      setSelectedRows([]);
      actionRef.current?.reload();
    } catch (error) {
      messageApi.error('批量启用失败');
    }
  };

  // 批量禁用
  const handleBatchDisable = async () => {
    if (selectedRows.length === 0) return;
    try {
      const result = await batchUpdateStatus({
        ids: selectedRows.map((row) => row.id),
        status: 'inactive',
      });
      messageApi.success(`成功禁用 ${result.success} 个数据源`);
      setSelectedRows([]);
      actionRef.current?.reload();
    } catch (error) {
      messageApi.error('批量禁用失败');
    }
  };

  // 批量删除
  const handleBatchDelete = async () => {
    if (selectedRows.length === 0) return;
    Modal.confirm({
      title: '确认批量删除',
      icon: <ExclamationCircleOutlined />,
      content: `确定要删除选中的 ${selectedRows.length} 个数据源吗？`,
      okText: '确认',
      cancelText: '取消',
      onOk: async () => {
        try {
          const result = await batchDelete(selectedRows.map((row) => row.id));
          messageApi.success(`成功删除 ${result.success} 个数据源`);
          if (result.failed > 0) {
            messageApi.warning(`${result.failed} 个数据源删除失败`);
          }
          setSelectedRows([]);
          actionRef.current?.reload();
        } catch (error) {
          messageApi.error('批量删除失败');
        }
      },
    });
  };

  // 表格列定义
  const columns: ProColumns<DataSourceAPI.DataSource>[] = [
    {
      title: '名称',
      dataIndex: 'name',
      render: (_, record) => (
        <a onClick={() => { setCurrentRow(record); setDetailDrawerOpen(true); }}>
          {record.name}
        </a>
      ),
    },
    {
      title: '类型',
      dataIndex: 'type',
      valueType: 'select',
      valueEnum: Object.fromEntries(
        Object.entries(dataSourceTypeMap).map(([key, value]) => [key, { text: value.text }])
      ),
      render: (_, record) => {
        const typeInfo = dataSourceTypeMap[record.type] || { text: record.type, color: 'default' };
        return <Tag color={typeInfo.color}>{typeInfo.text}</Tag>;
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      valueType: 'select',
      valueEnum: {
        active: { text: '在线', status: 'Success' },
        inactive: { text: '离线', status: 'Default' },
        error: { text: '错误', status: 'Error' },
        testing: { text: '测试中', status: 'Processing' },
      },
      render: (_, record) => {
        const statusInfo = statusMap[record.status] || { text: record.status, status: 'default' };
        return <Badge status={statusInfo.status} text={statusInfo.text} />;
      },
    },
    {
      title: '描述',
      dataIndex: 'description',
      ellipsis: true,
      search: false,
    },
    {
      title: '标签',
      dataIndex: 'tags',
      search: false,
      render: (_, record) => (
        <Space size={[0, 4]} wrap>
          {record.tags?.map((tag) => <Tag key={tag}>{tag}</Tag>)}
        </Space>
      ),
    },
    {
      title: '最后测试时间',
      dataIndex: 'last_test_at',
      valueType: 'dateTime',
      search: false,
    },
    {
      title: '更新时间',
      dataIndex: 'updated_at',
      valueType: 'dateTime',
      search: false,
      sorter: true,
    },
    {
      title: '操作',
      dataIndex: 'option',
      valueType: 'option',
      width: 200,
      render: (_, record) => (
        <Space size="small">
          <Tooltip title="测试连接">
            <Button
              type="link"
              size="small"
              icon={<ApiOutlined />}
              loading={testingId === record.id}
              onClick={() => handleTestConnection(record.id)}
            />
          </Tooltip>
          <Tooltip title="编辑">
            <Button
              type="link"
              size="small"
              onClick={() => { setCurrentRow(record); setUpdateModalOpen(true); }}
            >
              编辑
            </Button>
          </Tooltip>
          <Tooltip title="删除">
            <Button
              type="link"
              size="small"
              danger
              icon={<DeleteOutlined />}
              onClick={() => handleDelete(record)}
            />
          </Tooltip>
        </Space>
      ),
    },
  ];

  return (
    <PageContainer>
      {contextHolder}
      <ProTable<DataSourceAPI.DataSource, DataSourceAPI.ListDataSourcesParams>
        headerTitle="数据源列表"
        actionRef={actionRef}
        rowKey="id"
        search={{ labelWidth: 120 }}
        toolBarRender={() => [
          <BatchOperations
            key="batch"
            selectedIds={selectedRows.map((row) => row.id)}
            onSuccess={() => actionRef.current?.reload()}
          />,
          <Button
            type="primary"
            key="create"
            icon={<PlusOutlined />}
            onClick={() => setCreateModalOpen(true)}
          >
            新建数据源
          </Button>,
        ]}
        request={listDataSources}
        columns={columns}
        rowSelection={{
          onChange: (_, rows) => setSelectedRows(rows),
        }}
      />

      {selectedRows?.length > 0 && (
        <FooterToolbar
          extra={
            <div>
              已选择 <a style={{ fontWeight: 600 }}>{selectedRows.length}</a> 项
            </div>
          }
        >
          <Button icon={<PlayCircleOutlined />} onClick={handleBatchEnable}>
            批量启用
          </Button>
          <Button icon={<PauseCircleOutlined />} onClick={handleBatchDisable}>
            批量禁用
          </Button>
          <Button danger icon={<DeleteOutlined />} onClick={handleBatchDelete}>
            批量删除
          </Button>
        </FooterToolbar>
      )}

      <CreateForm
        open={createModalOpen}
        onOpenChange={setCreateModalOpen}
        onSuccess={() => {
          setCreateModalOpen(false);
          actionRef.current?.reload();
        }}
      />

      <UpdateForm
        open={updateModalOpen}
        onOpenChange={setUpdateModalOpen}
        values={currentRow}
        onSuccess={() => {
          setUpdateModalOpen(false);
          setCurrentRow(undefined);
          actionRef.current?.reload();
        }}
      />

      <DetailDrawer
        open={detailDrawerOpen}
        onClose={() => {
          setDetailDrawerOpen(false);
          setCurrentRow(undefined);
        }}
        dataSource={currentRow}
        onTestConnection={handleTestConnection}
      />
    </PageContainer>
  );
};

export default DataSourceList;
