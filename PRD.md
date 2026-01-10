# LiteCore 应用核心框架（GO版)

## 功能组件

应用核心启动器 bootstrap

配置管理器 configMgr
观测管理器 telemetryMgr
日志管理器 loggerMgr
数据库管理器 databaseMgr
缓存管理器 cacheMgr

存储层容器 repositoryCntr
服务层容器 serviceCntr
控制层容器 controllerCntr
中间件容器 middlewareCntr

## 配置文件示例

```yaml
# 应用基础信息
application:
  name: lite-demo-api
  version: 1.0.0

# HTTP服务配置
server:
  port: 6080
  gin_mode: debug

# 观测管理配置
telemetry:
  driver: otel # 可选：none / otel
  otel_config: # driver=otel 驱动配置
    endpoint: http://localhost:4317
    resource_attributes:
      - key: environment
        value: dev
      - key: region
        value: us-east-1
    traces:
      enabled: true
    metrics:
      enabled: true
    logs:
      enabled: true

# 日志管理配置
logger:
  output_telemetry:
    enabled: true
    level: info
  output_console:
    enabled: true
    level: info
  output_file:
    enabled: true
    level: info
    path: ./applogs/
    rotation:
      max_size: 100MB
      max_age: 30d
      max_backups: 10
      compress: true

# 数据库管理配置
database:
  driver: mysql # 可选：mysql / postgresql / sqlite
  sqlite_config: # driver=sqlite 的配置
    path: ./data.db
  postgresql_config: # driver=postgresql 的配置
    host: localhost
    port: 5432
    user: postgres
    password: password
    database: lite_demo
    max_open_conns: 10
    max_idle_conns: 5
    conn_max_lifetime: 30s
  mysql_config: # driver=mysql 的配置
    host: localhost
    port: 3306
    user: root
    password: password
    database: lite_demo
    max_open_conns: 10
    max_idle_conns: 5
    conn_max_lifetime: 30s

# 缓存管理配置
cache:
  driver: redis # 可选：redis / memory / none
  redis_config: # driver=redis 的配置
    host: localhost
    port: 6379
    password: password
    db: 0
    max_idle_conns: 10
    max_open_conns: 100
    conn_max_lifetime: 30s
  memory_config: # driver=memory 的配置
    max_size: 100MB
    max_age: 30d
    max_backups: 10
    compress: true
```
