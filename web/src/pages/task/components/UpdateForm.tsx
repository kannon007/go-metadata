import {
  ModalForm,
  ProFormText,
  ProFormDigit,
  ProFormGroup,
  ProFormTextArea,
  ProFormSelect,
} from '@ant-design/pro-components';
import { message, Tag, Space } from 'antd';
import React, { useEffect, useState } from 'react';
import { updateTask } from '@/services/task/api';

interface UpdateFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  values?: TaskAPI.CollectionTask;
  onSuccess: () => void;
}

// 任务类型映射
const taskTypeMap: Record<string, { text: string; color: string }> = {
  full_collection: { text: '全量采集', color: 'blue' },
  incremental_collection: { text: '增量采集', color: 'green' },
  schema_only: { text: '仅Schema', color: 'orange' },
  data_profile: { text: '数据画像', color: 'purple' },
};

// 调度类型选项
const scheduleTypeOptions = [
  { label: '单次执行', value: 'once' },
  { label: 'Cron表达式', value: 'cron' },
  { label: '固定间隔', value: 'interval' },
];

const UpdateForm: React.FC<UpdateFormProps> = ({ open, onOpenChange, values, onSuccess }) => {
  const [messageApi, contextHolder] = message.useMessage();
  const [formRef, setFormRef] = useState<any>(null);

  useEffect(() => {
    if (open && values && formRef) {
      formRef.setFieldsValue({
        name: values.name,
        include_schemas: values.config?.include_schemas?.join(', '),
        exclude_schemas: values.config?.exclude_schemas?.join(', '),
        include_tables: values.config?.include_tables?.join(', '),
        exclude_tables: values.config?.exclude_tables?.join(', '),
        batch_size: values.config?.batch_size,
        timeout: values.config?.timeout,
        retry_count: values.config?.retry_count,
        retry_interval: values.config?.retry_interval,
        schedule_type: values.schedule?.type,
        cron_expr: values.schedule?.cron_expr,
        interval: values.schedule?.interval,
        timezone: values.schedule?.timezone,
      });
    }
  }, [open, values, formRef]);

  if (!values) return null;

  const typeInfo = taskTypeMap[values.type] || { text: values.type, color: 'default' };

  return (
    <>
      {contextHolder}
      <ModalForm
        title={
          <Space>
            <span>编辑任务</span>
            <Tag color={typeInfo.color}>{typeInfo.text}</Tag>
          </Space>
        }
        width={700}
        open={open}
        onOpenChange={onOpenChange}
        formRef={(ref) => setFormRef(ref)}
        initialValues={{
          name: values.name,
          include_schemas: values.config?.include_schemas?.join(', '),
          exclude_schemas: values.config?.exclude_schemas?.join(', '),
          include_tables: values.config?.include_tables?.join(', '),
          exclude_tables: values.config?.exclude_tables?.join(', '),
          batch_size: values.config?.batch_size || 1000,
          timeout: values.config?.timeout || 3600,
          retry_count: values.config?.retry_count || 3,
          retry_interval: values.config?.retry_interval || 60,
          schedule_type: values.schedule?.type,
          cron_expr: values.schedule?.cron_expr,
          interval: values.schedule?.interval,
          timezone: values.schedule?.timezone || 'Asia/Shanghai',
        }}
        onFinish={async (formValues) => {
          try {
            await updateTask(values.id, {
              name: formValues.name,
              config: {
                include_schemas: formValues.include_schemas?.split(',').map((s: string) => s.trim()).filter(Boolean),
                exclude_schemas: formValues.exclude_schemas?.split(',').map((s: string) => s.trim()).filter(Boolean),
                include_tables: formValues.include_tables?.split(',').map((s: string) => s.trim()).filter(Boolean),
                exclude_tables: formValues.exclude_tables?.split(',').map((s: string) => s.trim()).filter(Boolean),
                batch_size: formValues.batch_size || 1000,
                timeout: formValues.timeout || 3600,
                retry_count: formValues.retry_count || 3,
                retry_interval: formValues.retry_interval || 60,
              },
              schedule: formValues.schedule_type ? {
                type: formValues.schedule_type,
                cron_expr: formValues.cron_expr,
                interval: formValues.interval,
                timezone: formValues.timezone || 'Asia/Shanghai',
              } : undefined,
            });
            messageApi.success('更新成功');
            onSuccess();
            return true;
          } catch (error) {
            messageApi.error('更新失败');
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
          />
        </ProFormGroup>

        <ProFormGroup title="采集范围">
          <ProFormTextArea name="include_schemas" label="包含Schema" width="md" placeholder="多个用逗号分隔" />
          <ProFormTextArea name="exclude_schemas" label="排除Schema" width="md" placeholder="多个用逗号分隔" />
        </ProFormGroup>

        <ProFormGroup>
          <ProFormTextArea name="include_tables" label="包含表" width="md" placeholder="多个用逗号分隔" />
          <ProFormTextArea name="exclude_tables" label="排除表" width="md" placeholder="多个用逗号分隔" />
        </ProFormGroup>

        <ProFormGroup title="执行配置">
          <ProFormDigit name="batch_size" label="批次大小" width="sm" fieldProps={{ precision: 0, min: 1 }} />
          <ProFormDigit name="timeout" label="超时时间(秒)" width="sm" fieldProps={{ precision: 0, min: 60 }} />
          <ProFormDigit name="retry_count" label="重试次数" width="sm" fieldProps={{ precision: 0, min: 0 }} />
          <ProFormDigit name="retry_interval" label="重试间隔(秒)" width="sm" fieldProps={{ precision: 0, min: 1 }} />
        </ProFormGroup>

        <ProFormGroup title="调度配置">
          <ProFormSelect name="schedule_type" label="调度类型" width="md" options={scheduleTypeOptions} />
          <ProFormText name="cron_expr" label="Cron表达式" width="md" placeholder="如: 0 0 * * *" />
        </ProFormGroup>

        <ProFormGroup>
          <ProFormDigit name="interval" label="执行间隔(秒)" width="sm" fieldProps={{ precision: 0, min: 60 }} />
          <ProFormText name="timezone" label="时区" width="sm" />
        </ProFormGroup>
      </ModalForm>
    </>
  );
};

export default UpdateForm;
