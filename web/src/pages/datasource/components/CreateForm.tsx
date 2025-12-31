import {
  ModalForm,
  ProFormSelect,
  ProFormText,
  ProFormTextArea,
  ProFormDigit,
  ProFormSwitch,
  StepsForm,
} from '@ant-design/pro-components';
import { message, Button, Space, Alert } from 'antd';
import React, { useState } from 'react';
import { createDataSource, testConnection } from '@/services/datasource/api';
import { ApiOutlined, CheckCircleOutlined, CloseCircleOutlined } from '@ant-design/icons';
import DynamicConfigForm from './DynamicConfigForm';
import { getStrategy, getDataSourceTypeOptions } from './strategies';

interface CreateFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess: () => void;
}

const CreateForm: React.FC<CreateFormProps> = ({ open, onOpenChange, onSuccess }) => {
  const [messageApi, contextHolder] = message.useMessage();
  const [testResult, setTestResult] = useState<{ success: boolean; message: string } | null>(null);
  const [testing, setTesting] = useState(false);
  const [selectedType, setSelectedType] = useState<string | undefined>();
  const [formData, setFormData] = useState<any>({});

  const handleTestConnection = async () => {
    if (!selectedType) {
      messageApi.error('请先选择数据源类型');
      return;
    }
    const strategy = getStrategy(selectedType);
    if (!strategy) {
      messageApi.error('未知的数据源类型');
      return;
    }
    try {
      setTesting(true);
      setTestResult(null);
      const config = strategy.toApiConfig(formData.config || {});
      const result = await testConnection({
        type: selectedType as DataSourceAPI.DataSourceType,
        config,
      });
      setTestResult({ success: result.success, message: result.message });
    } catch (error: any) {
      setTestResult({ success: false, message: error.message || '连接测试失败' });
    } finally {
      setTesting(false);
    }
  };

  const handleTypeChange = (value: string) => {
    setSelectedType(value);
    setTestResult(null);
  };

  return (
    <>
      {contextHolder}
      <StepsForm
        stepsFormRender={(dom, submitter) => (
          <ModalForm
            title="新建数据源"
            width={800}
            open={open}
            onOpenChange={(visible) => {
              if (!visible) {
                setTestResult(null);
                setSelectedType(undefined);
                setFormData({});
              }
              onOpenChange(visible);
            }}
            submitter={false}
            modalProps={{ destroyOnClose: true }}
          >
            {dom}
            <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: 24 }}>
              {submitter}
            </div>
          </ModalForm>
        )}
        onFinish={async (values) => {
          if (!selectedType) {
            messageApi.error('请选择数据源类型');
            return false;
          }
          const strategy = getStrategy(selectedType);
          if (!strategy) {
            messageApi.error('未知的数据源类型');
            return false;
          }
          try {
            const allData = { ...formData, ...values };
            const config = strategy.toApiConfig(allData.config || {});
            await createDataSource({
              name: allData.name,
              type: selectedType as DataSourceAPI.DataSourceType,
              description: allData.description,
              tags: allData.tags?.split(',').map((t: string) => t.trim()).filter(Boolean),
              config,
            });
            messageApi.success('创建成功');
            onSuccess();
            onOpenChange(false);
            return true;
          } catch (error) {
            messageApi.error('创建失败');
            return false;
          }
        }}
      >
        {/* 步骤1: 基本信息 */}
        <StepsForm.StepForm
          name="basic"
          title="基本信息"
          onFinish={async (values) => {
            setFormData((prev: any) => ({ ...prev, ...values }));
            setSelectedType(values.type);
            return true;
          }}
        >
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
          <ProFormSelect
            name="type"
            label="数据源类型"
            width="lg"
            options={getDataSourceTypeOptions()}
            rules={[{ required: true, message: '请选择数据源类型' }]}
            fieldProps={{ onChange: handleTypeChange, showSearch: true }}
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
        </StepsForm.StepForm>

        {/* 步骤2: 数据源配置 */}
        <StepsForm.StepForm
          name="connection"
          title="数据源配置"
          onFinish={async (values) => {
            setFormData((prev: any) => ({ ...prev, ...values }));
            return true;
          }}
        >
          <DynamicConfigForm dataSourceType={selectedType} />
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
            <Button
              icon={<ApiOutlined />}
              loading={testing}
              onClick={handleTestConnection}
              disabled={!selectedType}
            >
              测试连接
            </Button>
          </Space>
        </StepsForm.StepForm>

        {/* 步骤3: 调度配置 */}
        <StepsForm.StepForm
          name="schedule"
          title="调度配置"
          onFinish={async (values) => {
            setFormData((prev: any) => ({ ...prev, ...values }));
            return true;
          }}
        >
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
        </StepsForm.StepForm>
      </StepsForm>
    </>
  );
};

export default CreateForm;
