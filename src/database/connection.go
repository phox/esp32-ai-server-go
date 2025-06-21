package database

import (
	"database/sql"
	"fmt"
	"time"

	"ai-server-go/src/configs"
	"ai-server-go/src/core/utils"

	_ "github.com/go-sql-driver/mysql"
)

// Database 数据库连接管理器
type Database struct {
	DB     *sql.DB
	config *configs.DatabaseConfig
	logger *utils.Logger
}

// NewDatabase 创建新的数据库连接
func NewDatabase(config *configs.DatabaseConfig, logger *utils.Logger) (*Database, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true&loc=Local",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Name,
		config.Charset,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}

	// 配置连接池（使用默认值）
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Duration(3600) * time.Second)

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("数据库连接测试失败: %v", err)
	}

	database := &Database{
		DB:     db,
		config: config,
		logger: logger,
	}

	logger.Info("数据库连接成功: %s:%d/%s", config.Host, config.Port, config.Name)
	return database, nil
}

// Close 关闭数据库连接
func (d *Database) Close() error {
	if d.DB != nil {
		return d.DB.Close()
	}
	return nil
}

// GetDB 获取数据库连接
func (d *Database) GetDB() *sql.DB {
	return d.DB
}

// Begin 开始事务
func (d *Database) Begin() (*sql.Tx, error) {
	return d.DB.Begin()
}

// Exec 执行SQL语句
func (d *Database) Exec(query string, args ...interface{}) (sql.Result, error) {
	return d.DB.Exec(query, args...)
}

// Query 查询SQL语句
func (d *Database) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return d.DB.Query(query, args...)
}

// QueryRow 查询单行
func (d *Database) QueryRow(query string, args ...interface{}) *sql.Row {
	return d.DB.QueryRow(query, args...)
}

// Prepare 预处理SQL语句
func (d *Database) Prepare(query string) (*sql.Stmt, error) {
	return d.DB.Prepare(query)
}

// Ping 测试连接
func (d *Database) Ping() error {
	return d.DB.Ping()
}

// Stats 获取连接池统计信息
func (d *Database) Stats() sql.DBStats {
	return d.DB.Stats()
}
