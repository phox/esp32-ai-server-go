package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"ai-server-go/src/configs"
	"ai-server-go/src/core/utils"

	_ "github.com/go-sql-driver/mysql" // MySQL驱动
	_ "github.com/lib/pq"              // PostgreSQL驱动
	_ "github.com/mattn/go-sqlite3"    // SQLite驱动
)

// Database 数据库连接管理器
type Database struct {
	DB      *sql.DB
	config  *configs.DatabaseConfig
	logger  *utils.Logger
	driver  string
	dialect Dialect
}

// NewDatabase 创建新的数据库连接
func NewDatabase(config *configs.DatabaseConfig, logger *utils.Logger) (*Database, error) {
	// 设置默认值
	if config.Type == "" {
		config.Type = "mysql" // 默认使用MySQL
	}
	if config.MaxOpenConns == 0 {
		config.MaxOpenConns = 100
	}
	if config.MaxIdleConns == 0 {
		config.MaxIdleConns = 10
	}
	if config.ConnMaxLifetime == 0 {
		config.ConnMaxLifetime = time.Duration(3600) * time.Second
	}

	var dsn string
	var driver string

	switch strings.ToLower(config.Type) {
	case "mysql":
		driver = "mysql"
		dsn = buildMySQLDSN(config)
	case "postgresql", "postgres":
		driver = "postgres"
		dsn = buildPostgreSQLDSN(config)
	case "sqlite":
		driver = "sqlite3"
		dsn = buildSQLiteDSN(config)
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", config.Type)
	}

	// 获取对应的方言
	dialect, err := GetDialect(config.Type)
	if err != nil {
		return nil, fmt.Errorf("获取数据库方言失败: %v", err)
	}

	logger.Info("连接数据库: %s", config.Type)
	logger.Debug("DSN: %s", dsn)

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}

	// 配置连接池
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("数据库连接测试失败: %v", err)
	}

	database := &Database{
		DB:      db,
		config:  config,
		logger:  logger,
		driver:  driver,
		dialect: dialect,
	}

	logger.Info("数据库连接成功: %s", config.Type)
	return database, nil
}

// buildMySQLDSN 构建MySQL连接字符串
func buildMySQLDSN(config *configs.DatabaseConfig) string {
	params := []string{}

	if config.Charset != "" {
		params = append(params, "charset="+config.Charset)
	} else {
		params = append(params, "charset=utf8mb4")
	}

	if config.ParseTime {
		params = append(params, "parseTime=true")
	}

	if config.Loc != "" {
		params = append(params, "loc="+config.Loc)
	} else {
		params = append(params, "loc=Local")
	}

	paramStr := strings.Join(params, "&")
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Name,
		paramStr,
	)
}

// buildPostgreSQLDSN 构建PostgreSQL连接字符串
func buildPostgreSQLDSN(config *configs.DatabaseConfig) string {
	params := []string{}

	if config.SSLMode != "" {
		params = append(params, "sslmode="+config.SSLMode)
	} else {
		params = append(params, "sslmode=disable")
	}

	paramStr := strings.Join(params, " ")
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s %s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.Name,
		paramStr,
	)
}

// buildSQLiteDSN 构建SQLite连接字符串
func buildSQLiteDSN(config *configs.DatabaseConfig) string {
	filePath := config.Name
	if config.FilePath != "" {
		filePath = config.FilePath
	}
	return filePath
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

// GetDriver 获取数据库驱动类型
func (d *Database) GetDriver() string {
	return d.driver
}

// GetDialect 获取数据库方言
func (d *Database) GetDialect() Dialect {
	return d.dialect
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
