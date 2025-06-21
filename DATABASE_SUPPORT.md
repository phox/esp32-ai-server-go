# 多数据库支持

本项目现在支持多种数据库，包括 MySQL、PostgreSQL 和 SQLite。

## 支持的数据库

### 1. MySQL
- **驱动**: `github.com/go-sql-driver/mysql`
- **版本**: 5.7+, 8.0+
- **特点**: 高性能、功能丰富、广泛使用

### 2. PostgreSQL
- **驱动**: `github.com/lib/pq`
- **版本**: 10+
- **特点**: 功能强大、支持复杂查询、ACID 事务

### 3. SQLite
- **驱动**: `github.com/mattn/go-sqlite3`
- **版本**: 3.x
- **特点**: 轻量级、文件数据库、无需服务器

## 配置说明

### 基础配置

在 `config.yaml` 中配置数据库连接：

```yaml
database:
  # 数据库类型：mysql, postgresql, sqlite
  type: mysql
  
  # 连接池配置
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 3600s
```

### MySQL 配置

```yaml
database:
  type: mysql
  host: localhost
  port: 3306
  user: root
  password: "your_password"
  name: ai_server_go
  charset: utf8mb4
  parse_time: true
  loc: Local
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 3600s
```

### PostgreSQL 配置

```yaml
database:
  type: postgresql
  host: localhost
  port: 5432
  user: postgres
  password: "your_password"
  name: ai_server_go
  ssl_mode: disable  # disable, require, verify-ca, verify-full
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 3600s
```

### SQLite 配置

```yaml
database:
  type: sqlite
  name: ./data/ai_server.db
  file_path: ./data/ai_server.db
  max_open_conns: 1
  max_idle_conns: 1
  conn_max_lifetime: 3600s
```

## 数据库方言支持

系统自动根据数据库类型使用相应的 SQL 方言：

### 占位符
- **MySQL**: `?`
- **PostgreSQL**: `$1`, `$2`, `$3`...
- **SQLite**: `?`

### 自增主键
- **MySQL**: `AUTO_INCREMENT`
- **PostgreSQL**: `SERIAL`
- **SQLite**: `AUTOINCREMENT`

### 数据类型映射
- **BOOLEAN**: MySQL/PostgreSQL 使用 `BOOLEAN`，SQLite 使用 `INTEGER`
- **TEXT**: 所有数据库都支持
- **TIMESTAMP**: 所有数据库都支持

## 数据库迁移

系统启动时会自动执行数据库迁移，创建必要的表和索引：

### 表结构
1. **users** - 用户表
2. **devices** - 设备表
3. **device_capabilities** - 设备能力配置表
4. **system_configs** - 系统配置表
5. **provider_configs** - AI 提供商配置表

### 索引
- 用户名唯一索引
- 设备用户关联索引
- 配置分类键值索引
- 提供商分类名称索引

## 使用建议

### 开发环境
推荐使用 **SQLite**，配置简单，无需安装数据库服务器：

```yaml
database:
  type: sqlite
  name: ./data/dev.db
  max_open_conns: 1
  max_idle_conns: 1
```

### 生产环境
推荐使用 **MySQL** 或 **PostgreSQL**：

- **MySQL**: 适合大多数场景，性能好，社区支持广泛
- **PostgreSQL**: 适合需要复杂查询和高级功能的场景

### 性能调优

#### MySQL
```yaml
database:
  type: mysql
  max_open_conns: 200
  max_idle_conns: 20
  conn_max_lifetime: 3600s
```

#### PostgreSQL
```yaml
database:
  type: postgresql
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 3600s
  ssl_mode: require  # 生产环境建议启用 SSL
```

#### SQLite
```yaml
database:
  type: sqlite
  max_open_conns: 1  # SQLite 建议使用单连接
  max_idle_conns: 1
  conn_max_lifetime: 3600s
```

## 故障排除

### 常见问题

1. **连接失败**
   - 检查数据库服务是否启动
   - 验证连接参数是否正确
   - 确认网络连接是否正常

2. **权限问题**
   - MySQL/PostgreSQL: 确认用户有相应权限
   - SQLite: 确认文件路径可写

3. **字符集问题**
   - MySQL: 使用 `utf8mb4` 字符集
   - PostgreSQL: 使用 `UTF8` 编码

4. **SSL 连接问题**
   - PostgreSQL: 根据服务器配置调整 `ssl_mode`

### 日志调试

启动时查看日志输出：
```
INFO 连接数据库: mysql
INFO 数据库连接成功: mysql
INFO 开始数据库迁移...
INFO 创建表: users
INFO 创建表: devices
INFO 数据库迁移完成
```

## 扩展支持

如需支持其他数据库，可以：

1. 在 `dialect.go` 中实现新的方言接口
2. 在 `connection.go` 中添加新的 DSN 构建函数
3. 更新配置结构体支持新数据库的特定参数

## 依赖管理

确保 `go.mod` 包含必要的数据库驱动：

```go
require (
    github.com/go-sql-driver/mysql v1.8.1  // MySQL
    github.com/lib/pq v1.10.9              // PostgreSQL
    github.com/mattn/go-sqlite3 v1.14.28   // SQLite
)
``` 