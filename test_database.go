package main

import (
	"fmt"
	"os"

	"ai-server-go/src/configs"
	"ai-server-go/src/core/utils"
	"ai-server-go/src/database"
)

func main() {
	// 创建日志器
	config := &configs.Config{}
	logger, _ := utils.NewLogger(config)

	// 测试不同的数据库配置
	testConfigs := []struct {
		name   string
		config *configs.DatabaseConfig
	}{
		{
			name: "MySQL",
			config: &configs.DatabaseConfig{
				Type:            "mysql",
				Host:            "localhost",
				Port:            3306,
				User:            "root",
				Password:        "",
				Name:            "ai_server_go_test",
				Charset:         "utf8mb4",
				ParseTime:       true,
				Loc:             "Local",
				MaxOpenConns:    10,
				MaxIdleConns:    5,
				ConnMaxLifetime: 3600,
			},
		},
		{
			name: "PostgreSQL",
			config: &configs.DatabaseConfig{
				Type:            "postgresql",
				Host:            "localhost",
				Port:            5432,
				User:            "postgres",
				Password:        "password",
				Name:            "ai_server_go_test",
				SSLMode:         "disable",
				MaxOpenConns:    10,
				MaxIdleConns:    5,
				ConnMaxLifetime: 3600,
			},
		},
		{
			name: "SQLite",
			config: &configs.DatabaseConfig{
				Type:            "sqlite",
				Name:            "./test.db",
				FilePath:        "./test.db",
				MaxOpenConns:    1,
				MaxIdleConns:    1,
				ConnMaxLifetime: 3600,
			},
		},
	}

	for _, test := range testConfigs {
		fmt.Printf("\n=== 测试 %s ===\n", test.name)

		// 尝试连接数据库
		db, err := database.NewDatabase(test.config, logger)
		if err != nil {
			fmt.Printf("❌ %s 连接失败: %v\n", test.name, err)
			continue
		}
		defer db.Close()

		fmt.Printf("✅ %s 连接成功\n", test.name)
		fmt.Printf("   驱动: %s\n", db.GetDriver())
		fmt.Printf("   方言: %s\n", db.GetDialect().GetType())

		// 测试迁移
		migrationManager := database.NewMigrationManager(db, logger)
		if err := migrationManager.Migrate(); err != nil {
			fmt.Printf("❌ %s 迁移失败: %v\n", test.name, err)
			continue
		}

		fmt.Printf("✅ %s 迁移成功\n", test.name)

		// 测试基本操作
		if err := testBasicOperations(db, logger); err != nil {
			fmt.Printf("❌ %s 基本操作测试失败: %v\n", test.name, err)
			continue
		}

		fmt.Printf("✅ %s 基本操作测试成功\n", test.name)
	}

	// 清理测试文件
	if err := os.Remove("./test.db"); err == nil {
		fmt.Println("\n🧹 清理测试文件完成")
	}
}

func testBasicOperations(db *database.Database, logger *utils.Logger) error {
	// 测试查询
	rows, err := db.Query("SELECT 1")
	if err != nil {
		return fmt.Errorf("查询测试失败: %v", err)
	}
	defer rows.Close()

	// 测试事务
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("事务开始失败: %v", err)
	}
	defer tx.Rollback()

	// 测试预处理语句
	stmt, err := db.Prepare("SELECT 1")
	if err != nil {
		return fmt.Errorf("预处理语句失败: %v", err)
	}
	defer stmt.Close()

	return nil
}
