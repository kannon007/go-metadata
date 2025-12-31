import React, { useMemo } from 'react';
import {
  ProFormText,
  ProFormDigit,
  ProFormSwitch,
  ProFormSelect,
  ProFormTextArea,
} from '@ant-design/pro-components';
import { Row, Col, Card, Collapse } from 'antd';
import { ConfigField, getStrategy } from './strategies';

interface DynamicConfigFormProps {
  dataSourceType: string | undefined;
  namePrefix?: string[];
}

const renderField = (field: ConfigField, namePrefix: string[] = []) => {
  const name = namePrefix.length > 0 ? [...namePrefix, field.name] : field.name;
  const commonProps = {
    key: field.name,
    name,
    label: field.label,
    tooltip: field.tooltip,
    placeholder: field.placeholder,
    rules: field.required ? [{ required: true, message: `请输入${field.label}` }] : undefined,
    initialValue: field.defaultValue,
  };

  switch (field.type) {
    case 'text':
      return <ProFormText {...commonProps} />;
    case 'password':
      return <ProFormText.Password {...commonProps} />;
    case 'number':
      return (
        <ProFormDigit
          {...commonProps}
          fieldProps={{ precision: 0, min: field.min, max: field.max }}
        />
      );
    case 'switch':
      return <ProFormSwitch {...commonProps} />;
    case 'select':
      return <ProFormSelect {...commonProps} options={field.options} />;
    case 'textarea':
      return <ProFormTextArea {...commonProps} />;
    default:
      return <ProFormText {...commonProps} />;
  }
};

const getColSpan = (field: ConfigField) => {
  if (field.width === 'lg') return 24;
  if (field.type === 'switch') return 8;
  return 12;
};

const groupFields = (fields: ConfigField[]): Map<string, ConfigField[]> => {
  const groups = new Map<string, ConfigField[]>();
  fields.forEach((field) => {
    const groupName = field.group || '基础配置';
    if (!groups.has(groupName)) {
      groups.set(groupName, []);
    }
    groups.get(groupName)!.push(field);
  });
  return groups;
};

const DynamicConfigForm: React.FC<DynamicConfigFormProps> = ({
  dataSourceType,
  namePrefix = ['config'],
}) => {
  const strategy = useMemo(() => {
    if (!dataSourceType) return null;
    return getStrategy(dataSourceType);
  }, [dataSourceType]);

  if (!strategy) {
    return (
      <Card size="small">
        <div style={{ color: '#999', padding: '20px 0', textAlign: 'center' }}>
          请先选择数据源类型
        </div>
      </Card>
    );
  }

  const groupedFields = groupFields(strategy.fields);
  const basicFields = groupedFields.get('基础配置') || [];
  const advancedFields = groupedFields.get('高级配置') || [];

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      <Card title="基础配置" size="small">
        <Row gutter={16}>
          {basicFields.map((field) => (
            <Col key={field.name} span={getColSpan(field)}>
              {renderField(field, namePrefix)}
            </Col>
          ))}
        </Row>
      </Card>
      {advancedFields.length > 0 && (
        <Collapse
          ghost
          items={[{
            key: 'advanced',
            label: '高级配置',
            children: (
              <Row gutter={16}>
                {advancedFields.map((field) => (
                  <Col key={field.name} span={getColSpan(field)}>
                    {renderField(field, namePrefix)}
                  </Col>
                ))}
              </Row>
            ),
          }]}
        />
      )}
    </div>
  );
};

export default DynamicConfigForm;
