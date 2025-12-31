# Deployment Guide

## 部署指南

本文档介绍如何部署元数据管理系统。

## 系统要求

### 硬件要求

| 组件 | 最小配置 | 推荐配置 |
|------|----------|----------|
| CPU | 2 核 | 4 核+ |
| 内存 | 4 GB | 8 GB+ |
| 磁盘 | 20 GB | 100 GB+ |

### 软件要求

- Docker 20.10+
- Docker Compose 2.0+
- Go 1.21+ (仅开发环境)

## 快速开始

### 1. 克隆项目

```bash
git clone https://github.com/your-org/go-metadata.git
cd go-metadata
```

### 2. 配置环境变量

```bash
cp configs/config.yaml.example configs/config.yaml
```

编辑 `configs/config.yaml` 配置数据库连接等信息。

### 3. 启动服务

```bash
# 使用部署脚本
./scripts/deploy.sh up

# 或直接使用 docker-compose
docker-compose -f deployments/docker/docker-compose.yaml up -d
```

### 4. 验证部署

```bash
# 检查服务状态
./scripts/deploy.sh status

# 访问健康检查端点
curl http://localhost:8080/health
```

## 详细部署步骤

### Docker 部署

#### 构建镜像

```bash
# 构建所有镜像
./scripts/deploy.sh build --version v1.0.0

# 或手动构建
docker build -t go-metadata:latest -f deployments/docker/Dockerfile .
```

#### 启动服务

```bash
# 启动基础服务（MySQL + NebulaGraph）
./scripts/deploy.sh up

# 启动带 Redis 缓存
./scripts/deploy.sh up --profile cache

# 启动带 Neo4j
./scripts/deploy.sh up --profile neo4j

# 启动带 PostgreSQL
./scripts/deploy.sh up --profile postgres
```

#### 查看日志

```bash
# 查看所有服务日志
./scripts/deploy.sh logs

# 查看特定服务日志
./scripts/deploy.sh logs metadata-server
```

### Kubernetes 部署

#### 创建命名空间

```bash
kubectl create namespace metadata
```

#### 部署配置

```yaml
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: metadata-config
  namespace: metadata
data:
  config.yaml: |
    server:
      http:
        addr: 0.0.0.0:8080
        timeout: 30s
      grpc:
        addr: 0.0.0.0:9090
        timeout: 30s
    data:
      database:
        driver: mysql
        source: metadata:metadata123@tcp(mysql:3306)/metadata?parseTime=True
```

#### 部署服务

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: metadata-server
  namespace: metadata
spec:
  replicas: 2
  selector:
    matchLabels:
      app: metadata-server
  template:
    metadata:
      labels:
        app: metadata-server
    spec:
      containers:
      - name: metadata-server
        image: go-metadata:latest
        ports:
        - containerPort: 8080
        - containerPort: 9090
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
        volumeMounts:
        - name: config
          mountPath: /app/configs
      volumes:
      - name: config
        configMap:
          name: metadata-config
```

#### 部署服务

```yaml
# k8s/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: metadata-server
  namespace: metadata
spec:
  selector:
    app: metadata-server
  ports:
  - name: http
    port: 8080
    targetPort: 8080
  - name: grpc
    port: 9090
    targetPort: 9090
  type: ClusterIP
```

## 配置说明

### 数据库配置

```yaml
data:
  database:
    driver: mysql  # mysql, postgres
    source: user:password@tcp(host:port)/database?parseTime=True
    max_open_conns: 100
    max_idle_conns: 10
    conn_max_lifetime: 3600s
```

### 图数据库配置

```yaml
data:
  graph:
    type: nebula  # nebula, neo4j
    nebula:
      address: localhost:9669
      username: root
      password: nebula
      space: metadata
    neo4j:
      uri: bolt://localhost:7687
      username: neo4j
      password: neo4j123
```

### 调度器配置

```yaml
scheduler:
  type: builtin  # builtin, dolphinscheduler
  builtin:
    worker_count: 10
    queue_size: 1000
  dolphinscheduler:
    api_url: http://localhost:12345/dolphinscheduler
    token: your-token
    project_code: 123456
```

### 安全配置

```yaml
auth:
  jwt:
    secret: your-secret-key
    expire: 3600s
    refresh_expire: 86400s
  rate_limit:
    enabled: true
    requests_per_ip: 100
    burst_size: 200
    window: 1s
```

## 数据库迁移

### 自动迁移

```bash
./scripts/deploy.sh migrate
```

### 手动迁移

```bash
# MySQL
mysql -u metadata -p metadata < migrations/001_create_metadata_tables.sql
mysql -u metadata -p metadata < migrations/002_create_indexes.sql
mysql -u metadata -p metadata < migrations/003_create_datasource_tables.sql
```

## 监控和告警

### Prometheus 指标

服务暴露 Prometheus 指标端点：

```
GET /metrics
```

主要指标：
- `metadata_http_requests_total` - HTTP 请求总数
- `metadata_http_request_duration_seconds` - HTTP 请求延迟
- `metadata_datasource_connections_active` - 活跃连接数
- `metadata_task_executions_total` - 任务执行总数

### Grafana 仪表板

导入预配置的 Grafana 仪表板：

```bash
# 仪表板 JSON 文件位于
deployments/grafana/dashboards/
```

### 告警规则

```yaml
# prometheus/alerts.yaml
groups:
- name: metadata-alerts
  rules:
  - alert: HighErrorRate
    expr: rate(metadata_http_requests_total{status=~"5.."}[5m]) > 0.1
    for: 5m
    labels:
      severity: critical
    annotations:
      summary: High error rate detected
      
  - alert: SlowRequests
    expr: histogram_quantile(0.95, rate(metadata_http_request_duration_seconds_bucket[5m])) > 1
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: Slow request latency detected
```

## 备份和恢复

### 数据库备份

```bash
# MySQL 备份
mysqldump -u metadata -p metadata > backup_$(date +%Y%m%d).sql

# PostgreSQL 备份
pg_dump -U metadata metadata > backup_$(date +%Y%m%d).sql
```

### 恢复

```bash
# MySQL 恢复
mysql -u metadata -p metadata < backup_20240101.sql

# PostgreSQL 恢复
psql -U metadata metadata < backup_20240101.sql
```

## 故障排除

### 常见问题

#### 1. 服务无法启动

检查日志：
```bash
./scripts/deploy.sh logs metadata-server
```

常见原因：
- 数据库连接失败
- 端口被占用
- 配置文件错误

#### 2. 数据库连接失败

```bash
# 检查数据库状态
docker-compose -f deployments/docker/docker-compose.yaml exec mysql mysqladmin ping

# 检查网络连接
docker-compose -f deployments/docker/docker-compose.yaml exec metadata-server ping mysql
```

#### 3. 内存不足

调整容器资源限制：
```yaml
services:
  metadata-server:
    deploy:
      resources:
        limits:
          memory: 1G
```

### 日志级别

调整日志级别进行调试：

```yaml
log:
  level: debug  # debug, info, warn, error
  format: json
```

## 升级指南

### 版本升级步骤

1. 备份数据库
2. 停止服务
3. 拉取新版本镜像
4. 运行数据库迁移
5. 启动服务
6. 验证功能

```bash
# 1. 备份
mysqldump -u metadata -p metadata > backup_before_upgrade.sql

# 2. 停止服务
./scripts/deploy.sh down

# 3. 拉取新版本
docker pull go-metadata:v2.0.0

# 4. 运行迁移
./scripts/deploy.sh migrate

# 5. 启动服务
./scripts/deploy.sh up

# 6. 验证
curl http://localhost:8080/health
```

## 安全建议

1. **使用强密码** - 所有数据库和服务使用强密码
2. **启用 TLS** - 生产环境启用 HTTPS
3. **网络隔离** - 使用私有网络隔离服务
4. **定期更新** - 及时更新依赖和镜像
5. **审计日志** - 启用并定期审查审计日志
6. **备份策略** - 实施定期备份策略
