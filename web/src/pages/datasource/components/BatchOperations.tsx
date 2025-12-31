import {
  ModalForm,
  ProFormSelect,
  ProFormTextArea,
} from '@ant-design/pro-components';
import { Button, message, Space, Upload, Dropdown } from 'antd';
import type { UploadProps } from 'antd';
import React, { useState } from 'react';
import { batchImport, batchExport } from '@/services/datasource/api';
import {
  UploadOutlined,
  DownloadOutlined,
  ImportOutlined,
  ExportOutlined,
} from '@ant-design/icons';

interface BatchOperationsProps {
  selectedIds: string[];
  onSuccess: () => void;
}

const BatchOperations: React.FC<BatchOperationsProps> = ({ selectedIds, onSuccess }) => {
  const [messageApi, contextHolder] = message.useMessage();
  const [importModalOpen, setImportModalOpen] = useState(false);
  const [exportModalOpen, setExportModalOpen] = useState(false);
  const [importData, setImportData] = useState<string>('');
  const [exporting, setExporting] = useState(false);

  // 处理文件上传
  const handleFileUpload: UploadProps['beforeUpload'] = (file) => {
    const reader = new FileReader();
    reader.onload = (e) => {
      const content = e.target?.result as string;
      setImportData(content);
    };
    reader.readAsText(file);
    return false; // 阻止自动上传
  };

  // 导入数据源
  const handleImport = async (values: { format: 'json' | 'yaml'; data: string }) => {
    try {
      const result = await batchImport({
        format: values.format,
        data: values.data,
      });
      messageApi.success(`成功导入 ${result.success} 个数据源`);
      if (result.failed > 0) {
        messageApi.warning(`${result.failed} 个数据源导入失败`);
      }
      onSuccess();
      return true;
    } catch (error) {
      messageApi.error('导入失败');
      return false;
    }
  };

  // 导出数据源
  const handleExport = async (values: { format: 'json' | 'yaml' }) => {
    setExporting(true);
    try {
      const result = await batchExport({
        ids: selectedIds.length > 0 ? selectedIds : undefined,
        format: values.format,
      });

      // 创建下载链接
      const blob = new Blob([result.data], { type: 'text/plain;charset=utf-8' });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `datasources_export_${new Date().toISOString().slice(0, 10)}.${values.format}`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);

      messageApi.success(`成功导出 ${result.count} 个数据源`);
      return true;
    } catch (error) {
      messageApi.error('导出失败');
      return false;
    } finally {
      setExporting(false);
    }
  };

  // 下拉菜单项
  const menuItems = [
    {
      key: 'import',
      label: '导入配置',
      icon: <ImportOutlined />,
      onClick: () => setImportModalOpen(true),
    },
    {
      key: 'export',
      label: selectedIds.length > 0 ? `导出选中(${selectedIds.length})` : '导出全部',
      icon: <ExportOutlined />,
      onClick: () => setExportModalOpen(true),
    },
  ];

  return (
    <>
      {contextHolder}
      <Dropdown menu={{ items: menuItems }} trigger={['click']}>
        <Button>
          <Space>
            批量操作
            <DownloadOutlined />
          </Space>
        </Button>
      </Dropdown>

      {/* 导入弹窗 */}
      <ModalForm
        title="导入数据源配置"
        width={600}
        open={importModalOpen}
        onOpenChange={setImportModalOpen}
        onFinish={handleImport}
      >
        <ProFormSelect
          name="format"
          label="配置格式"
          options={[
            { label: 'JSON', value: 'json' },
            { label: 'YAML', value: 'yaml' },
          ]}
          rules={[{ required: true, message: '请选择配置格式' }]}
          initialValue="json"
        />

        <Upload
          accept=".json,.yaml,.yml"
          beforeUpload={handleFileUpload}
          maxCount={1}
          showUploadList={false}
        >
          <Button icon={<UploadOutlined />} style={{ marginBottom: 16 }}>
            选择文件
          </Button>
        </Upload>

        <ProFormTextArea
          name="data"
          label="配置内容"
          fieldProps={{
            rows: 15,
            value: importData,
            onChange: (e) => setImportData(e.target.value),
          }}
          rules={[{ required: true, message: '请输入或上传配置内容' }]}
          placeholder="请粘贴或上传JSON/YAML格式的数据源配置"
        />
      </ModalForm>

      {/* 导出弹窗 */}
      <ModalForm
        title="导出数据源配置"
        width={400}
        open={exportModalOpen}
        onOpenChange={setExportModalOpen}
        onFinish={handleExport}
        submitter={{
          submitButtonProps: {
            loading: exporting,
          },
        }}
      >
        <ProFormSelect
          name="format"
          label="导出格式"
          options={[
            { label: 'JSON', value: 'json' },
            { label: 'YAML', value: 'yaml' },
          ]}
          rules={[{ required: true, message: '请选择导出格式' }]}
          initialValue="json"
        />

        <div style={{ color: '#666', marginTop: 8 }}>
          {selectedIds.length > 0
            ? `将导出选中的 ${selectedIds.length} 个数据源配置`
            : '将导出所有数据源配置'}
        </div>
      </ModalForm>
    </>
  );
};

export default BatchOperations;
