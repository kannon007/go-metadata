import { ProDescriptions } from '@ant-design/pro-components';
import { Badge, Button, Drawer, Space, Tag, Divider } from 'antd';
import React from 'react';
import { ApiOutlined } from '@ant-design/icons';

interface DetailDrawerProps {
  open: boolean;
  onClose: () => void;
  dataSource?: DataSourceAPI.DataSource;
  onTestConnection: (id: string) => void;
}

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

const DetailDrawer: React.FC<DetailDrawerProps> = ({
  open,
  onClose,
  dataSource,
  onTestConnection,
}) => {
  if (!dataSource) return null;

  const typeInfo = dataSourceTypeMap[dataSource.type] || { text: dataSource.type, color: 'default' };
  const statusInfo = statusMap[dataSource.status] || { text: dataSource.status, status: 'default' };

  return (
    <Drawer
      title={
        <Space>
          <span>{dataSource.name}</span>
          <Tag color={typeInfo.color}>{typeInfo.text}</Tag>
        </Space>
      }
      width={600}
      open={open}
      onClose={onClose}
      extra={
        <Button
          type="primary"
          icon={<ApiOutlined />}
          onClick={() => onTestConnection(dataSource.id)}
        >
          测试连接
        </Button>
      }
    >
      <ProDescriptions column={2} title="基本信息">
        <ProDescriptions.Item label="ID" copyable>
          {dataSource.id}
        </ProDescriptions.Item>
        <ProDescriptions.Item label="名称">{dataSource.name}</ProDescriptions.Item>
        <ProDescriptions.Item label="类型">
          <Tag color={typeInfo.color}>{typeInfo.text}</Tag>
        </ProDescriptions.Item>
        <ProDescriptions.Item label="状态">
          <Badge status={statusInfo.status} text={statusInfo.text} />
        </ProDescriptions.Item>
        <ProDescriptions.Item label="描述" span={2}>
          {dataSource.description || '-'}
        </ProDescriptions.Item>
        <ProDescriptions.Item label="标签" span={2}>
          <Space size={[0, 4]} wrap>
            {dataSource.tags?.length > 0
              ? dataSource.tags.map((tag) => <Tag key={tag}>{tag}</Tag>)
              : '-'}
          </Space>
        </ProDescriptions.Item>
        <ProDescriptions.Item label="创建者">{dataSource.created_by || '-'}</ProDescriptions.Item>
        <ProDescriptions.Item label="创建时间" valueType="dateTime">
          {dataSource.created_at}
        </ProDescriptions.Item>
        <ProDescriptions.Item label="更新时间" valueType="dateTime">
          {dataSource.updated_at}
        </ProDescriptions.Item>
        <ProDescriptions.Item label="最后测试时间" valueType="dateTime">
          {dataSource.last_test_at || '-'}
        </ProDescriptions.Item>
      </ProDescriptions>

      <Divider />

      <ProDescriptions column={2} title="连接配置">
        <ProDescriptions.Item label="主机地址">{dataSource.config?.host}</ProDescriptions.Item>
        <ProDescriptions.Item label="端口">{dataSource.config?.port}</ProDescriptions.Item>
        <ProDescriptions.Item label="数据库">{dataSource.config?.database || '-'}</ProDescriptions.Item>
        <ProDescriptions.Item label="用户名">{dataSource.config?.username || '-'}</ProDescriptions.Item>
        <ProDescriptions.Item label="SSL">
          {dataSource.config?.ssl ? <Tag color="green">启用</Tag> : <Tag>禁用</Tag>}
        </ProDescriptions.Item>
        <ProDescriptions.Item label="SSL模式">{dataSource.config?.ssl_mode || '-'}</ProDescriptions.Item>
        <ProDescriptions.Item label="字符集">{dataSource.config?.charset || '-'}</ProDescriptions.Item>
        <ProDescriptions.Item label="超时时间">{dataSource.config?.timeout}秒</ProDescriptions.Item>
      </ProDescriptions>

      <Divider />

      <ProDescriptions column={2} title="连接池配置">
        <ProDescriptions.Item label="最大连接数">{dataSource.config?.max_conns}</ProDescriptions.Item>
        <ProDescriptions.Item label="最大空闲连接">{dataSource.config?.max_idle_conns}</ProDescriptions.Item>
      </ProDescriptions>

      {dataSource.config?.extra && Object.keys(dataSource.config.extra).length > 0 && (
        <>
          <Divider />
          <ProDescriptions column={2} title="扩展配置">
            {Object.entries(dataSource.config.extra).map(([key, value]) => (
              <ProDescriptions.Item key={key} label={key}>
                {value}
              </ProDescriptions.Item>
            ))}
          </ProDescriptions>
        </>
      )}
    </Drawer>
  );
};

export default DetailDrawer;
