package database

import (
	"fmt"
	"strings"

	"ai-server-go/src/core/utils"
)

// MigrationManager 数据库迁移管理器
type MigrationManager struct {
	db      *Database
	dialect Dialect
	logger  *utils.Logger
}

// NewMigrationManager 创建新的迁移管理器
func NewMigrationManager(db *Database, logger *utils.Logger) *MigrationManager {
	return &MigrationManager{
		db:      db,
		dialect: db.GetDialect(),
		logger:  logger,
	}
}

// CreateTables 创建所有必要的表
func (m *MigrationManager) CreateTables() error {
	tables := []struct {
		name    string
		columns []Column
	}{
		{
			name: "users",
			columns: []Column{
				{Name: "id", Type: "VARCHAR", Length: 36, Nullable: false, PrimaryKey: true},
				{Name: "username", Type: "VARCHAR", Length: 50, Nullable: false, Unique: true},
				{Name: "password", Type: "VARCHAR", Length: 255, Nullable: false},
				{Name: "email", Type: "VARCHAR", Length: 100, Nullable: true},
				{Name: "role", Type: "VARCHAR", Length: 20, Nullable: false, Default: "'user'"},
				{Name: "created_at", Type: "TIMESTAMP", Nullable: false, Default: m.dialect.GetCurrentTimeSQL()},
				{Name: "updated_at", Type: "TIMESTAMP", Nullable: false, Default: m.dialect.GetCurrentTimeSQL()},
			},
		},
		{
			name: "devices",
			columns: []Column{
				{Name: "id", Type: "VARCHAR", Length: 36, Nullable: false, PrimaryKey: true},
				{Name: "name", Type: "VARCHAR", Length: 100, Nullable: false},
				{Name: "type", Type: "VARCHAR", Length: 50, Nullable: false},
				{Name: "status", Type: "VARCHAR", Length: 20, Nullable: false, Default: "'active'"},
				{Name: "user_id", Type: "VARCHAR", Length: 36, Nullable: true},
				{Name: "created_at", Type: "TIMESTAMP", Nullable: false, Default: m.dialect.GetCurrentTimeSQL()},
				{Name: "updated_at", Type: "TIMESTAMP", Nullable: false, Default: m.dialect.GetCurrentTimeSQL()},
			},
		},
		{
			name: "device_capabilities",
			columns: []Column{
				{Name: "id", Type: "VARCHAR", Length: 36, Nullable: false, PrimaryKey: true},
				{Name: "device_id", Type: "VARCHAR", Length: 36, Nullable: false},
				{Name: "capability_name", Type: "VARCHAR", Length: 50, Nullable: false},
				{Name: "capability_type", Type: "VARCHAR", Length: 50, Nullable: false},
				{Name: "priority", Type: "INT", Nullable: false, Default: "0"},
				{Name: "config", Type: "TEXT", Nullable: true},
				{Name: "created_at", Type: "TIMESTAMP", Nullable: false, Default: m.dialect.GetCurrentTimeSQL()},
				{Name: "updated_at", Type: "TIMESTAMP", Nullable: false, Default: m.dialect.GetCurrentTimeSQL()},
			},
		},
		{
			name: "system_configs",
			columns: []Column{
				{Name: "id", Type: "VARCHAR", Length: 36, Nullable: false, PrimaryKey: true},
				{Name: "category", Type: "VARCHAR", Length: 50, Nullable: false},
				{Name: "key", Type: "VARCHAR", Length: 100, Nullable: false},
				{Name: "value", Type: "TEXT", Nullable: true},
				{Name: "type", Type: "VARCHAR", Length: 20, Nullable: false, Default: "'string'"},
				{Name: "description", Type: "TEXT", Nullable: true},
				{Name: "created_at", Type: "TIMESTAMP", Nullable: false, Default: m.dialect.GetCurrentTimeSQL()},
				{Name: "updated_at", Type: "TIMESTAMP", Nullable: false, Default: m.dialect.GetCurrentTimeSQL()},
			},
		},
		{
			name: "provider_configs",
			columns: []Column{
				{Name: "id", Type: "VARCHAR", Length: 36, Nullable: false, PrimaryKey: true},
				{Name: "category", Type: "VARCHAR", Length: 50, Nullable: false},
				{Name: "name", Type: "VARCHAR", Length: 100, Nullable: false},
				{Name: "type", Type: "VARCHAR", Length: 50, Nullable: false},
				{Name: "model_name", Type: "VARCHAR", Length: 100, Nullable: true},
				{Name: "base_url", Type: "VARCHAR", Length: 255, Nullable: true},
				{Name: "api_key", Type: "VARCHAR", Length: 255, Nullable: true},
				{Name: "app_id", Type: "VARCHAR", Length: 100, Nullable: true},
				{Name: "token", Type: "VARCHAR", Length: 255, Nullable: true},
				{Name: "voice", Type: "VARCHAR", Length: 100, Nullable: true},
				{Name: "format", Type: "VARCHAR", Length: 20, Nullable: true},
				{Name: "output_dir", Type: "VARCHAR", Length: 255, Nullable: true},
				{Name: "cluster", Type: "VARCHAR", Length: 255, Nullable: true},
				{Name: "temperature", Type: "FLOAT", Nullable: true},
				{Name: "max_tokens", Type: "INT", Nullable: true},
				{Name: "top_p", Type: "FLOAT", Nullable: true},
				{Name: "security", Type: "TEXT", Nullable: true},
				{Name: "extra", Type: "TEXT", Nullable: true},
				{Name: "is_active", Type: "BOOLEAN", Nullable: false, Default: "true"},
				{Name: "is_default", Type: "BOOLEAN", Nullable: false, Default: "false"},
				{Name: "weight", Type: "INT", Nullable: false, Default: "100"},
				{Name: "health_score", Type: "FLOAT", Nullable: false, Default: "100.0"},
				{Name: "version", Type: "VARCHAR", Length: 20, Nullable: false, Default: "'v1'"},
				{Name: "created_at", Type: "TIMESTAMP", Nullable: false, Default: m.dialect.GetCurrentTimeSQL()},
				{Name: "updated_at", Type: "TIMESTAMP", Nullable: false, Default: m.dialect.GetCurrentTimeSQL()},
			},
		},
	}

	for _, table := range tables {
		if err := m.createTable(table.name, table.columns); err != nil {
			return fmt.Errorf("创建表 %s 失败: %v", table.name, err)
		}
	}

	return nil
}

// createTable 创建单个表
func (m *MigrationManager) createTable(tableName string, columns []Column) error {
	sql := m.dialect.GetCreateTableSQL(tableName, columns)

	// 根据数据库类型调整SQL
	switch m.dialect.GetType() {
	case "sqlite":
		// SQLite需要特殊处理BOOLEAN类型
		sql = strings.ReplaceAll(sql, "BOOLEAN", "INTEGER")
		sql = strings.ReplaceAll(sql, "true", "1")
		sql = strings.ReplaceAll(sql, "false", "0")
	case "postgresql":
		// PostgreSQL需要特殊处理TEXT类型
		sql = strings.ReplaceAll(sql, "TEXT", "TEXT")
	}

	m.logger.Info("创建表: %s", tableName)
	m.logger.Debug("SQL: %s", sql)

	_, err := m.db.Exec(sql)
	return err
}

// CreateIndexes 创建索引
func (m *MigrationManager) CreateIndexes() error {
	indexes := []struct {
		tableName string
		indexName string
		columns   []string
		unique    bool
	}{
		// users表索引
		{tableName: "users", indexName: "idx_users_username", columns: []string{"username"}, unique: true},
		{tableName: "users", indexName: "idx_users_email", columns: []string{"email"}, unique: false},

		// devices表索引
		{tableName: "devices", indexName: "idx_devices_user_id", columns: []string{"user_id"}, unique: false},
		{tableName: "devices", indexName: "idx_devices_status", columns: []string{"status"}, unique: false},

		// device_capabilities表索引
		{tableName: "device_capabilities", indexName: "idx_device_capabilities_device_id", columns: []string{"device_id"}, unique: false},
		{tableName: "device_capabilities", indexName: "idx_device_capabilities_name_type", columns: []string{"capability_name", "capability_type"}, unique: false},

		// system_configs表索引
		{tableName: "system_configs", indexName: "idx_system_configs_category_key", columns: []string{"category", "key"}, unique: true},

		// provider_configs表索引
		{tableName: "provider_configs", indexName: "idx_provider_configs_category_name", columns: []string{"category", "name"}, unique: true},
		{tableName: "provider_configs", indexName: "idx_provider_configs_category_active", columns: []string{"category", "is_active"}, unique: false},
		{tableName: "provider_configs", indexName: "idx_provider_configs_category_default", columns: []string{"category", "is_default"}, unique: false},
	}

	for _, idx := range indexes {
		if err := m.createIndex(idx.tableName, idx.indexName, idx.columns, idx.unique); err != nil {
			return fmt.Errorf("创建索引 %s 失败: %v", idx.indexName, err)
		}
	}

	return nil
}

// createIndex 创建单个索引
func (m *MigrationManager) createIndex(tableName, indexName string, columns []string, unique bool) error {
	sql := m.dialect.GetAddIndexSQL(tableName, indexName, columns, unique)

	m.logger.Info("创建索引: %s", indexName)
	m.logger.Debug("SQL: %s", sql)

	_, err := m.db.Exec(sql)
	return err
}

// Migrate 执行完整的数据库迁移
func (m *MigrationManager) Migrate() error {
	m.logger.Info("开始数据库迁移...")

	// 创建表
	if err := m.CreateTables(); err != nil {
		return fmt.Errorf("创建表失败: %v", err)
	}

	// 创建索引
	if err := m.CreateIndexes(); err != nil {
		return fmt.Errorf("创建索引失败: %v", err)
	}

	m.logger.Info("数据库迁移完成")
	return nil
}
