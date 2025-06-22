package database

import (
	"fmt"
	"log"
	"time"

	"ai-server-go/src/configs"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database 数据库管理器
type Database struct {
	DB *gorm.DB
}

// NewDatabase 创建数据库连接
func NewDatabase(config *configs.DatabaseConfig) (*Database, error) {
	var db *gorm.DB
	var err error

	// 配置GORM日志
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// 根据数据库类型创建连接
	switch config.Type {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.User, config.Password, config.Host, config.Port, config.Name)
		db, err = gorm.Open(mysql.Open(dsn), gormConfig)
		if err != nil {
			return nil, fmt.Errorf("MySQL连接失败: %v", err)
		}

	case "postgres":
		sslMode := "disable"
		if config.SSLMode != "" {
			sslMode = config.SSLMode
		}
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=Asia/Shanghai",
			config.Host, config.User, config.Password, config.Name, config.Port, sslMode)
		db, err = gorm.Open(postgres.Open(dsn), gormConfig)
		if err != nil {
			return nil, fmt.Errorf("PostgreSQL连接失败: %v", err)
		}

	case "sqlite":
		db, err = gorm.Open(sqlite.Open(config.Name), gormConfig)
		if err != nil {
			return nil, fmt.Errorf("SQLite连接失败: %v", err)
		}

	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", config.Type)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取底层数据库连接失败: %v", err)
	}

	// 设置连接池参数（使用配置文件中的值）
	if config.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	} else {
		sqlDB.SetMaxOpenConns(100) // 默认值
	}
	if config.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	} else {
		sqlDB.SetMaxIdleConns(10) // 默认值
	}
	if config.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	} else {
		sqlDB.SetConnMaxLifetime(time.Hour) // 默认值
	}

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("数据库连接测试失败: %v", err)
	}

	log.Printf("数据库连接成功: %s", config.Type)

	return &Database{DB: db}, nil
}

// AutoMigrate 自动迁移数据库表结构
func (d *Database) AutoMigrate() error {
	log.Println("开始自动迁移数据库表结构...")

	// 定义所有模型
	models := []interface{}{
		&User{},
		&UserAuth{},
		&Device{},
		&DeviceAuth{},
		&UserDevice{},
		&AICapability{},
		&UserCapability{},
		&DeviceCapability{},
		&GlobalConfig{},
		&SystemConfig{},
		&Session{},
		&UsageStats{},
		&ProviderConfig{},
	}

	// 执行自动迁移
	if err := d.DB.AutoMigrate(models...); err != nil {
		return fmt.Errorf("数据库迁移失败: %v", err)
	}

	log.Println("数据库表结构迁移完成")
	return nil
}

// Close 关闭数据库连接
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// GetDB 获取GORM数据库实例
func (d *Database) GetDB() *gorm.DB {
	return d.DB
}

// Transaction 执行数据库事务
func (d *Database) Transaction(fc func(tx *gorm.DB) error) error {
	return d.DB.Transaction(fc)
}

// Begin 开始事务
func (d *Database) Begin() *gorm.DB {
	return d.DB.Begin()
}

// Commit 提交事务
func (d *Database) Commit() *gorm.DB {
	return d.DB.Commit()
}

// Rollback 回滚事务
func (d *Database) Rollback() *gorm.DB {
	return d.DB.Rollback()
}
