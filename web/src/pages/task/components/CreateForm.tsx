import {
  ModalForm,
  ProFormSelect,
  ProFormText,
  ProFormDigit,
  ProFormGroup,
  ProFormTextArea,
} from '@ant-design/pro-components';
import { message } from 'antd';
import React from 'react';
import { createTask } from '@/services/task/api';

interface CreateFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  dataSources: DataSourceAPI.DataSource[];
  onSuccess: () => void;
}

// 任务类型选项
const taskTypeOptions = [
  { label: '全量采集', value: 'full_collection' },
  { label: '增量采集', value: 'incremental_collection' },
  { label: '仅Schema', value: 'schema_only' },
  { label: '数据画像', value: 'data_profile' },
];

// 调度器类型选项
const schedulerTypeOptions = [
  { label: '内置调度器', value: 'builtin' },
  { label: 'DolphinScheduler', value: 'dolphinscheduler' },
];

// 调度类型选项
const scheduleTypeOptions = [
  { label: '单次执行', value: 'once' },
  { label: 'Cron表达式', value: 'cron' },
  { label: '固定间隔', value: 'interval' },
];

const CreateForm: React.FC<CreateFormProps> = ({ open, onOpenChange, dataSources, onSuccess }) => {
  const [messageApi, contextHolder] = message.useMessage();

  return (
    <>
      {contextHolder}
      <ModalForm
        title="新建采集任务"
        width={700}
        open={open}
        onOpenChange={onOpenChange}
        onFinish={async (values) => {
          try {
            await createTask({
              name: values.name,
              datasource_id: values.datasource_id,
              type: values.type,
              scheduler_type: values.scheduler_type,
              config: {
                include_schemas: values.include_schemas?.split(',').map((s: string) => s.trim()).filter(Boolean),
                exclude_schemas: values.exclude_schemas?.split(',').map((s: string) => s.trim()).filter(Boolean),
                include_tables: values.include_tables?.split(',').map((s: string) => s.trim()).filter(Boolean),
                exclude_tables: values.exclude_tables?.split(',').map((s: string) => s.trim()).filter(Boolean),
                batch_size: values.batch_size || 1000,
                timeout: values.timeout || 3600,
                retry_count: values.retry_count || 3,
                retry_interval: values.retry_interval || 60,
              },
              schedule: values.schedule_type ? {
                type: values.schedule_type,
                cron_expr: values.cron_expr,
                interval: values.interval,
                timezone: values.timezone || 'Asia/Shanghai',
              } : undefined,
            });
            messageApi.success('创建成功');
            onSuccess();
            return true;
          } catch (error) {
            messageApi.error('创建失败');
            return false;
          }
        }}
      >
        <ProFormGroup title="基本信息">
          <ProFormText
            name="name"
            label="任务名称"
            width="md"
            rules={[{ required: true, message: '请输入任务名称' }]}
            placeholder="请输入任务名称"
          />
          <ProFormSelect
            name="datasource_id"
            label="数据源"
            width="md"
            options={dataSources.map((ds) => ({ label: ds.name, value: ds.id }))}
            rules={[{ required: true, message: '请选择数据源' }]}
          />
        </ProFormGroup>

        <ProFormGroup>
          <ProFormSelect
            name="type"
            label="任务类型"
            width="md"
            options={taskTypeOptions}
            rules={[{ required: true, message: '请选择任务类型' }]}
          />
          <ProFormSelect
            name="scheduler_type"
            label="调度器"
            width="md"
            options={schedulerTypeOptions}
            rules={[{ required: true, message: '请选择调度器' }]}
            initialValue="builtin"
          />
        </ProFormGroup>

        <ProFormGroup title="采集范围">
          <ProFormTextArea
            name="include_schemas"
            label="包含Schema"
            width="md"
            placeholder="多个用逗号分隔，留空表示全部"
            tooltip="指定要采集的Schema，多个用逗号分隔"
          />
          <ProFormTextArea
            name="exclude_schemas"
            label="排除Schema"
            width="md"
            placeholder="多个用逗号分隔"
            tooltip="指定要排除的Schema，多个用逗号分隔"
          />
        </ProFormGroup>

        <ProFormGroup>
          <ProFormTextArea
            name="include_tables"
            label="包含表"
            width="md"
            placeholder="多个用逗号分隔，支持通配符*"
          />
          <ProFormTextArea
            name="exclude_tables"
            label="排除表"
            width="md"
            placeholder="多个用逗号分隔，支持通配符*"
          />
        </ProFormGroup>

        <ProFormGroup title="执行配置">
          <ProFormDigit
            name="batch_size"
            label="批次大小"
            width="sm"
            initialValue={1000}
            fieldProps={{ precision: 0, min: 1 }}
          />
          <ProFormDigit
            name="timeout"
            label="超时时间(秒)"
            width="sm"
            initialValue={3600}
            fieldProps={{ precision: 0, min: 60 }}
          />
          <ProFormDigit
            name="retry_count"
            label="重试次数"
            width="sm"
            initialValue={3}
            fieldProps={{ precision: 0, min: 0 }}
          />
          <ProFormDigit
            name="retry_interval"
            label="重试间隔(秒)"
            width="sm"
            initialValue={60}
            fieldProps={{ precision: 0, min: 1 }}
          />
        </ProFormGroup>

        <ProFormGroup title="调度配置">
          <ProFormSelect
            name="schedule_type"
            label="调度类型"
            width="md"
            options={scheduleTypeOptions}
            placeholder="不选择则为手动执行"
          />
          <ProFormText
            name="cron_expr"
            label="Cron表达式"
            width="md"
            placeholder="如: 0 0 * * *"
            tooltip="标准Cron表达式，如: 0 0 * * * 表示每天0点执行"
          />
        </ProFormGroup>

        <ProFormGroup>
          <ProFormDigit
            name="interval"
            label="执行间隔(秒)"
            width="sm"
            fieldProps={{ precision: 0, min: 60 }}
            tooltip="固定间隔执行时使用"
          />
          <ProFormText
            name="timezone"
            label="时区"
            width="sm"
            initialValue="Asia/Shanghai"
          />
        </ProFormGroup>
      </ModalForm>
    </>
  );
};

export default CreateForm;
