import {
  ProFormText,
  ProFormTextArea,
  ProFormDigit,
  ProFormSwitch,
  ProFormSelect,
  ProFormInstance,
} from '@ant-design/pro-components';
import { Modal, message, Button, Space, Alert, Tabs, Tag } from 'antd';
import React, { useState, useEffect, useRef } from 'react';
import { updateDataSource, testConnection } from '@/services/datasource/api';
import { ApiOutlined, CheckCircleOutlined, CloseCircleOutlined } from '@ant-design/icons';
import DynamicConfigForm from './DynamicConfigForm';
import { getStrategy, getAllStrategies } from './strategies';
import { ProForm } from '@ant-design/pro-components';

interface UpdateFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  values?: DataSourceAPI.DataSource;
  onSuccess: () => void;
}

const getTypeInfo = (type: string) => {
  const strategies = getAllStrategies();
  const strategy = strategies.find((s) => s.type === type);
  return strategy ? { text: strategy.label, color: 'blue' } : { text: type, color: 'default' };
};

const UpdateForm: React.FC<UpdateFormProps> = ({ open, onOpenChange, values, onSuccess }) => {
  const [messageApi, contextHolder] = message.useMessage();
  const [testResult, setTestResult] = useState<{ success: boolean; message: string } | null>(null);
  const [testing, setTesting] = useState(false);
  const formRef = useRef<ProFormInstance>(null);
  const [activeTab, setActiveTab] = useState('basic');

  useEffect(() => {
    if (open && values && formRef.current) {
      const strategy = getStrategy(values.type);
      const configValues = strategy ? strategy.fromApiConfig(values.config || {}) : values.config;
      formRef.current.setFieldsValue({
        name: values.name,
        description: values.description,
        tags: values.tags?.join(', '),
        config: configValues,
      });
    }
  }, [open, values]);

  const handleTestConnection = async () => {
    if (!formRef.current || !values) return;
    const strategy = getStrategy(values.type);
    if (!strategy) {
      messageApi.error('未知的数据源类型');
      return;
    }
    try {
      const formValues = await formRef.current.validateFields();
      setTesting(true);
      setTestResult(null);
      const config = strategy.toApiConfig(formValues.config || {});
      const result = await testConnection({
        type: values.type as DataSourceAPI.DataSourceType,
        config,
      });
      setTestResult({ success: result.success, message: result.message });
    } catch (error: any) {
      if (error.errorFields) {
        messageApi.error('请先填写必填字段');
      } else {
        setTestResult({ success: false, message: error.message || '连接测试失败' });
      }
    } finally {
      setTesting(false);
    }
  };

  const handleSubmit = async () => {
    if (!formRef.current || !values) return;
    const strategy = getStrategy(values.type);
    if (!strategy) {
      messageApi.error('未知的数据源类型');
      return;
    }
    try {
      const formValues = await formRef.current.validateFields();
      const config = strategy.toApiConfig(formValues.config || {});
      await updateDataSource(values.id, {
        name: formValues.name,
        description: formValues.description,
        tags: formValues.tags?.split(',').map((t: string) => t.trim()).filter(Boolean),
        config,
      });
      messageApi.success('更新成功');
      onSuccess();
      onOpenChange(false);
    } catch (error: any) {
      if (!error.errorFields) {
        messageApi.error('更新失败');
      }
    }
  };

  if (!values) return null;

  const typeInfo = getTypeInfo(values.type);

  const tabItems = [
    {
      key: 'basic',
      label: '基本信息',
      children: (
        <>
          <ProFormText
            name="name"
            label="数据源名称"
            width="lg"
            rules={[
              { required: true, message: '请输入数据源名称' },
              { pattern: /^[a-zA-Z0-9_\u4e00-\u9fa5-]+$/, message: '名称只能包含字母、数字、中文、下划线和横线' },
            ]}
            placeholder="请输入数据源名称"
          />
          <ProFormTextArea
            name="description"
            label="描述"
            width="lg"
            placeholder="请输入数据源描述"
            fieldProps={{ rows: 3 }}
          />
          <ProFormText
            name="tags"
            label="标签"
            width="lg"
            placeholder="多个标签用逗号分隔"
            tooltip="标签用于分类和筛选"
          />
        </>
      ),
    },
    {
      key: 'connection',
      label: '数据源配置',
      children: (
        <>
          <DynamicConfigForm dataSourceType={values.type} />
          {testResult && (
            <Alert
              message={testResult.success ? '连接成功' : '连接失败'}
              description={testResult.message}
              type={testResult.success ? 'success' : 'error'}
              showIcon
              icon={testResult.success ? <CheckCircleOutlined /> : <CloseCircleOutlined />}
              style={{ marginBottom: 16, marginTop: 16 }}
            />
          )}
          <Space style={{ marginTop: 16 }}>
            <Button icon={<ApiOutlined />} loading={testing} onClick={handleTestConnection}>
              测试连接
            </Button>
          </Space>
        </>
      ),
    },
    {
      key: 'schedule',
      label: '调度配置',
      children: (
        <>
          <ProFormSwitch
            name={['schedule', 'enabled']}
            label="启用定时采集"
            tooltip="开启后将按照设定的时间自动采集元数据"
          />
          <ProFormSelect
            name={['schedule', 'type']}
            label="调度类型"
            width="md"
            options={[
              { label: '间隔执行', value: 'interval' },
              { label: 'Cron 表达式', value: 'cron' },
            ]}
            initialValue="interval"
          />
          <ProFormDigit
            name={['schedule', 'interval']}
            label="执行间隔(分钟)"
            width="md"
            min={1}
            max={10080}
            initialValue={60}
            tooltip="最小1分钟，最大7天(10080分钟)"
          />
          <ProFormText
            name={['schedule', 'cron']}
            label="Cron 表达式"
            width="md"
            placeholder="0 0 * * * (每小时执行)"
            tooltip="标准 Cron 表达式，如: 0 0 * * * 表示每小时执行"
          />
          <ProFormDigit
            name={['schedule', 'timeout']}
            label="超时时间(秒)"
            width="md"
            min={30}
            max={3600}
            initialValue={300}
            tooltip="单次采集任务的最大执行时间"
          />
          <ProFormDigit
            name={['schedule', 'retries']}
            label="重试次数"
            width="md"
            min={0}
            max={5}
            initialValue={3}
            tooltip="采集失败后的重试次数"
          />
        </>
      ),
    },
  ];

  return (
    <>
      {contextHolder}
      <Modal
        title={
          <Space>
            <span>编辑数据源</span>
            <Tag color={typeInfo.color}>{typeInfo.text}</Tag>
          </Space>
        }
        width={800}
        open={open}
        onCancel={() => {
          setTestResult(null);
          setActiveTab('basic');
          onOpenChange(false);
        }}
        onOk={handleSubmit}
        okText="保存"
        cancelText="取消"
        destroyOnClose
      >
        <ProForm
          formRef={formRef}
          submitter={false}
          layout="vertical"
        >
          <Tabs activeKey={activeTab} onChange={setActiveTab} items={tabItems} />
        </ProForm>
      </Modal>
    </>
  );
};

export default UpdateForm;
