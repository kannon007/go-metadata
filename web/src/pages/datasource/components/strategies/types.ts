// 数据源配置策略模式 - 类型定义

export interface DataSourceConfig {
  [key: string]: any;
}

export interface ConfigField {
  name: string;
  label: string;
  type: 'text' | 'password' | 'number' | 'switch' | 'select' | 'textarea';
  required?: boolean;
  placeholder?: string;
  tooltip?: string;
  defaultValue?: any;
  options?: { label: string; value: any }[];
  min?: number;
  max?: number;
  width?: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
  group?: string;
}

export interface DataSourceStrategy {
  type: string;
  label: string;
  icon?: string;
  defaultPort: number;
  fields: ConfigField[];
  toApiConfig: (values: any) => DataSourceConfig;
  fromApiConfig: (config: DataSourceConfig) => any;
}
