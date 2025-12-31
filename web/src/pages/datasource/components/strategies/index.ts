// 数据源配置策略模式
// 每种数据源类型有不同的配置字段和验证规则

// 导出类型
export * from './types';

// 导入所有策略
import { mysqlStrategy } from './mysql';
import { postgresqlStrategy } from './postgresql';
import { mongodbStrategy } from './mongodb';
import { redisStrategy } from './redis';
import { kafkaStrategy } from './kafka';
import { elasticsearchStrategy } from './elasticsearch';
import { clickhouseStrategy } from './clickhouse';
import { hiveStrategy } from './hive';
import { DataSourceStrategy } from './types';

// 策略注册表
const strategies: Map<string, DataSourceStrategy> = new Map();

// 注册所有策略
[
  mysqlStrategy,
  postgresqlStrategy,
  mongodbStrategy,
  redisStrategy,
  kafkaStrategy,
  elasticsearchStrategy,
  clickhouseStrategy,
  hiveStrategy,
].forEach((strategy) => {
  strategies.set(strategy.type, strategy);
});

// 获取策略
export function getStrategy(type: string): DataSourceStrategy | undefined {
  return strategies.get(type);
}

// 获取所有策略
export function getAllStrategies(): DataSourceStrategy[] {
  return Array.from(strategies.values());
}

// 获取数据源类型选项
export function getDataSourceTypeOptions() {
  return getAllStrategies().map((s) => ({
    label: s.label,
    value: s.type,
  }));
}

// 导出所有策略
export {
  mysqlStrategy,
  postgresqlStrategy,
  mongodbStrategy,
  redisStrategy,
  kafkaStrategy,
  elasticsearchStrategy,
  clickhouseStrategy,
  hiveStrategy,
};
