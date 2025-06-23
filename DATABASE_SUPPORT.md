# 数据库支持说明

本项目已全面采用 [GORM](https://gorm.io/) 作为 Go 语言的 ORM（对象关系映射）组件，支持多种主流数据库，极大提升了开发效率和数据库兼容性。

## 1. 支持的数据库类型
- MySQL
- PostgreSQL
- SQLite

## 2. 主要数据模型与表结构
所有数据表均通过 GORM 的模型自动生成，主要模型包括：
- **User**：用户表，含用户名、邮箱、密码、角色、状态等字段，支持唯一索引。
- **Device**：设备表，含设备UUID、OUI、SN、名称、型号、状态等。
- **UserDevice**：用户与设备的绑定关系，支持多用户多设备，含权限、别名、是否所有者等。
- **AICapability**：AI能力表，描述能力类型、名称、配置Schema等。
- **UserCapability/DeviceCapability**：用户/设备与AI能力的绑定及配置。
- **ProviderConfig**：AI Provider 配置表，支持多类型Provider及版本、权重、扩展参数。
- **UserProvider/DeviceProvider**：用户/设备与Provider的绑定表。
- **Session/UsageStats**：会话与统计信息。
- **GlobalConfig/SystemConfig**：全局与系统配置表。

所有模型均支持 GORM 的自动迁移、外键、索引、JSON字段等高级特性。详细结构见 `src/database/models.go`。

## 3. GORM 主要特性
- 自动建表与结构迁移（AutoMigrate）
- 事务支持
- 连接池管理
- 丰富的模型定义与关联关系
- 支持多数据库切换
- 代码即模型，易于维护

## 4. 配置方法
数据库相关配置在 `config.yaml` 文件中：

## 4.1 MySQL 配置示例
```yaml
Database:
  type: mysql
  host: 127.0.0.1
  port: 3306
  user: root
  password: yourpassword
  name: ai_server
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 3600s
```

## 4.2 SQLite 配置示例
```yaml
Database:
  type: sqlite
  name: ./data/ai_server.db   # SQLite数据库文件路径
  max_open_conns: 1           # 建议单连接
  max_idle_conns: 1
  conn_max_lifetime: 3600s
```

## 4.3 PostgreSQL 配置示例
```yaml
Database:
  type: postgresql
  host: 127.0.0.1
  port: 5432
  user: postgres
  password: yourpassword
  name: ai_server
  ssl_mode: disable           # 可选: disable, require, verify-ca, verify-full
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 3600s
```

## 5. 连接池与性能调优
- 连接池参数（最大连接数、最大空闲数、最大生命周期）可在配置文件中调整。
- 支持高并发场景下的连接池自动管理。
- 日志级别可调，便于开发和生产环境调试。

## 6. 自动迁移与升级
- 系统启动时自动执行所有模型的 `AutoMigrate`，无需手动建表。
- 支持平滑升级表结构，字段变更自动同步到数据库。
- 如需自定义表名、索引、外键等，可在模型结构体中通过 GORM Tag 配置。

## 7. 事务支持
- 支持显式事务（`Begin`/`Commit`/`Rollback`）和函数式事务（`db.Transaction(func(tx *gorm.DB) error { ... })`）。
- 推荐在涉及多表写操作时使用事务，保证数据一致性。

**事务代码示例：**
```go
err := db.Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&user).Error; err != nil {
        return err
    }
    if err := tx.Create(&device).Error; err != nil {
        return err
    }
    return nil
})
```

## 8. 关联关系与高级用法
- 支持一对多、多对多、外键、级联删除等复杂关系。
- 通过 `gorm:"foreignKey:XXX"`、`gorm:"uniqueIndex"` 等Tag实现约束。
- JSON字段（如权限、配置等）自动序列化/反序列化。

**模型定义示例：**
```go
type User struct {
    gorm.Model
    Username string `gorm:"uniqueIndex;size:50;not null"`
    // ...
}
```

## 9. 多数据库兼容
- 支持 MySQL、PostgreSQL、SQLite，切换只需修改配置文件。
- 各数据库的方言和特性由 GORM 自动适配。

## 10. 常见开发场景举例
- **新增模型**：在 `models.go` 中定义结构体，重启服务自动建表。
- **数据查询**：`db.Where(...).Find(&objs)`、`db.First(&obj, id)` 等。
- **复杂查询**：支持联表、聚合、原生SQL等。
- **数据迁移**：升级模型结构体，重启服务自动迁移。

## 11. 进阶参考
- [GORM 关联关系文档](https://gorm.io/zh_CN/docs/associations.html)
- [GORM 事务文档](https://gorm.io/zh_CN/docs/transactions.html)
- [GORM 迁移文档](https://gorm.io/zh_CN/docs/migration.html)
- [GORM 官方文档](https://gorm.io/zh_CN/)
- [GORM GitHub](https://github.com/go-gorm/gorm)

如需更详细的表结构、字段说明或高级用法，请查阅 `src/database/models.go` 或 GORM 官方文档。 