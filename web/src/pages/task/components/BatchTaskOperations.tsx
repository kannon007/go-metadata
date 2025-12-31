import {
  ModalForm,
  ProFormSelect,
} from '@ant-design/pro-components';
import { Button, message, Space, Dropdown } from 'antd';
import React, { useState } from 'react';
import { batchStartTasks, batchStopTasks, batchDeleteTasks } from '@/services/task/api';
import {
  PlayCircleOutlined,
  StopOutlined,
  DeleteOutlined,
  MoreOutlined,
} from '@ant-design/icons';

interface BatchTaskOperationsProps {
  selectedIds: string[];
  onSuccess: () => void;
}

const BatchTaskOperations: React.FC<BatchTaskOperationsProps> = ({ selectedIds, onSuccess }) => {
  const [messageApi, contextHolder] = message.useMessage();
  const [batchActionModalOpen, setBatchActionModalOpen] = useState(false);
  const [loading, setLoading] = useState(false);

  // 批量启动
  const handleBatchStart = async () => {
    if (selectedIds.length === 0) {
      messageApi.warning('请先选择任务');
      return;
    }
    setLoading(true);
    try {
      const result = await batchStartTasks(selectedIds);
      messageApi.success(`成功启动 ${result.success} 个任务`);
      if (result.failed > 0) {
        messageApi.warning(`${result.failed} 个任务启动失败`);
      }
      onSuccess();
    } catch (error) {
      messageApi.error('批量启动失败');
    } finally {
      setLoading(false);
    }
  };

  // 批量停止
  const handleBatchStop = async () => {
    if (selectedIds.length === 0) {
      messageApi.warning('请先选择任务');
      return;
    }
    setLoading(true);
    try {
      const result = await batchStopTasks(selectedIds);
      messageApi.success(`成功停止 ${result.success} 个任务`);
      if (result.failed > 0) {
        messageApi.warning(`${result.failed} 个任务停止失败`);
      }
      onSuccess();
    } catch (error) {
      messageApi.error('批量停止失败');
    } finally {
      setLoading(false);
    }
  };

  // 批量删除
  const handleBatchDelete = async () => {
    if (selectedIds.length === 0) {
      messageApi.warning('请先选择任务');
      return;
    }
    setLoading(true);
    try {
      const result = await batchDeleteTasks(selectedIds);
      messageApi.success(`成功删除 ${result.success} 个任务`);
      if (result.failed > 0) {
        messageApi.warning(`${result.failed} 个任务删除失败`);
      }
      onSuccess();
    } catch (error) {
      messageApi.error('批量删除失败');
    } finally {
      setLoading(false);
    }
  };

  // 下拉菜单项
  const menuItems = [
    {
      key: 'start',
      label: `批量启动${selectedIds.length > 0 ? `(${selectedIds.length})` : ''}`,
      icon: <PlayCircleOutlined />,
      onClick: handleBatchStart,
      disabled: selectedIds.length === 0,
    },
    {
      key: 'stop',
      label: `批量停止${selectedIds.length > 0 ? `(${selectedIds.length})` : ''}`,
      icon: <StopOutlined />,
      onClick: handleBatchStop,
      disabled: selectedIds.length === 0,
    },
    {
      key: 'delete',
      label: `批量删除${selectedIds.length > 0 ? `(${selectedIds.length})` : ''}`,
      icon: <DeleteOutlined />,
      danger: true,
      onClick: handleBatchDelete,
      disabled: selectedIds.length === 0,
    },
  ];

  return (
    <>
      {contextHolder}
      <Dropdown menu={{ items: menuItems }} trigger={['click']} disabled={loading}>
        <Button loading={loading}>
          <Space>
            批量操作
            <MoreOutlined />
          </Space>
        </Button>
      </Dropdown>
    </>
  );
};

export default BatchTaskOperations;
