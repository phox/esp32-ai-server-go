package database

import (
	"fmt"
	"strings"
)

// Dialect 数据库方言接口
type Dialect interface {
	// 获取数据库类型
	GetType() string
	
	// 获取占位符
	GetPlaceholder(index int) string
	
	// 获取自增主键语法
	GetAutoIncrement() string
	
	// 获取创建表的语法
	GetCreateTableSQL(tableName string, columns []Column) string
	
	// 获取添加列的语法
	GetAddColumnSQL(tableName string, column Column) string
	
	// 获取修改列的语法
	GetModifyColumnSQL(tableName string, column Column) string
	
	// 获取删除列的语法
	GetDropColumnSQL(tableName string, columnName string) string
	
	// 获取添加索引的语法
	GetAddIndexSQL(tableName string, indexName string, columns []string, unique bool) string
	
	// 获取删除索引的语法
	GetDropIndexSQL(tableName string, indexName string) string
	
	// 获取限制结果数量的语法
	GetLimitSQL(limit, offset int) string
	
	// 获取当前时间函数
	GetCurrentTimeSQL() string
	
	// 获取字符串连接函数
	GetConcatSQL(values ...string) string
	
	// 获取IFNULL函数
	GetIfNullSQL(expr, defaultValue string) string
}

// Column 列定义
type Column struct {
	Name       string
	Type       string
	Length     int
	Nullable   bool
	Default    string
	PrimaryKey bool
	AutoIncr   bool
	Unique     bool
	Index      bool
}

// MySQLDialect MySQL方言
type MySQLDialect struct{}

func (d *MySQLDialect) GetType() string {
	return "mysql"
}

func (d *MySQLDialect) GetPlaceholder(index int) string {
	return "?"
}

func (d *MySQLDialect) GetAutoIncrement() string {
	return "AUTO_INCREMENT"
}

func (d *MySQLDialect) GetCreateTableSQL(tableName string, columns []Column) string {
	var columnDefs []string
	var primaryKeys []string
	
	for _, col := range columns {
		def := fmt.Sprintf("`%s` %s", col.Name, col.Type)
		
		if col.Length > 0 {
			def = fmt.Sprintf("%s(%d)", def, col.Length)
		}
		
		if !col.Nullable {
			def += " NOT NULL"
		}
		
		if col.Default != "" {
			def += fmt.Sprintf(" DEFAULT %s", col.Default)
		}
		
		if col.AutoIncr {
			def += " AUTO_INCREMENT"
		}
		
		if col.PrimaryKey {
			primaryKeys = append(primaryKeys, fmt.Sprintf("`%s`", col.Name))
		}
		
		columnDefs = append(columnDefs, def)
	}
	
	if len(primaryKeys) > 0 {
		columnDefs = append(columnDefs, fmt.Sprintf("PRIMARY KEY (%s)", strings.Join(primaryKeys, ", ")))
	}
	
	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (\n  %s\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4", 
		tableName, strings.Join(columnDefs, ",\n  "))
}

func (d *MySQLDialect) GetAddColumnSQL(tableName string, column Column) string {
	def := fmt.Sprintf("`%s` %s", column.Name, column.Type)
	if column.Length > 0 {
		def = fmt.Sprintf("%s(%d)", def, column.Length)
	}
	if !column.Nullable {
		def += " NOT NULL"
	}
	if column.Default != "" {
		def += fmt.Sprintf(" DEFAULT %s", column.Default)
	}
	return fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN %s", tableName, def)
}

func (d *MySQLDialect) GetModifyColumnSQL(tableName string, column Column) string {
	def := fmt.Sprintf("`%s` %s", column.Name, column.Type)
	if column.Length > 0 {
		def = fmt.Sprintf("%s(%d)", def, column.Length)
	}
	if !column.Nullable {
		def += " NOT NULL"
	}
	if column.Default != "" {
		def += fmt.Sprintf(" DEFAULT %s", column.Default)
	}
	return fmt.Sprintf("ALTER TABLE `%s` MODIFY COLUMN %s", tableName, def)
}

func (d *MySQLDialect) GetDropColumnSQL(tableName string, columnName string) string {
	return fmt.Sprintf("ALTER TABLE `%s` DROP COLUMN `%s`", tableName, columnName)
}

func (d *MySQLDialect) GetAddIndexSQL(tableName string, indexName string, columns []string, unique bool) string {
	uniqueStr := ""
	if unique {
		uniqueStr = "UNIQUE "
	}
	columnList := make([]string, len(columns))
	for i, col := range columns {
		columnList[i] = fmt.Sprintf("`%s`", col)
	}
	return fmt.Sprintf("CREATE %sINDEX `%s` ON `%s` (%s)", 
		uniqueStr, indexName, tableName, strings.Join(columnList, ", "))
}

func (d *MySQLDialect) GetDropIndexSQL(tableName string, indexName string) string {
	return fmt.Sprintf("DROP INDEX `%s` ON `%s`", indexName, tableName)
}

func (d *MySQLDialect) GetLimitSQL(limit, offset int) string {
	if offset > 0 {
		return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
	}
	return fmt.Sprintf("LIMIT %d", limit)
}

func (d *MySQLDialect) GetCurrentTimeSQL() string {
	return "NOW()"
}

func (d *MySQLDialect) GetConcatSQL(values ...string) string {
	return fmt.Sprintf("CONCAT(%s)", strings.Join(values, ", "))
}

func (d *MySQLDialect) GetIfNullSQL(expr, defaultValue string) string {
	return fmt.Sprintf("IFNULL(%s, %s)", expr, defaultValue)
}

// PostgreSQLDialect PostgreSQL方言
type PostgreSQLDialect struct{}

func (d *PostgreSQLDialect) GetType() string {
	return "postgresql"
}

func (d *PostgreSQLDialect) GetPlaceholder(index int) string {
	return fmt.Sprintf("$%d", index+1)
}

func (d *PostgreSQLDialect) GetAutoIncrement() string {
	return "SERIAL"
}

func (d *PostgreSQLDialect) GetCreateTableSQL(tableName string, columns []Column) string {
	var columnDefs []string
	var primaryKeys []string
	
	for _, col := range columns {
		def := fmt.Sprintf("\"%s\" %s", col.Name, col.Type)
		
		if col.Length > 0 {
			def = fmt.Sprintf("%s(%d)", def, col.Length)
		}
		
		if !col.Nullable {
			def += " NOT NULL"
		}
		
		if col.Default != "" {
			def += fmt.Sprintf(" DEFAULT %s", col.Default)
		}
		
		if col.PrimaryKey {
			primaryKeys = append(primaryKeys, fmt.Sprintf("\"%s\"", col.Name))
		}
		
		columnDefs = append(columnDefs, def)
	}
	
	if len(primaryKeys) > 0 {
		columnDefs = append(columnDefs, fmt.Sprintf("PRIMARY KEY (%s)", strings.Join(primaryKeys, ", ")))
	}
	
	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS \"%s\" (\n  %s\n)", 
		tableName, strings.Join(columnDefs, ",\n  "))
}

func (d *PostgreSQLDialect) GetAddColumnSQL(tableName string, column Column) string {
	def := fmt.Sprintf("\"%s\" %s", column.Name, column.Type)
	if column.Length > 0 {
		def = fmt.Sprintf("%s(%d)", def, column.Length)
	}
	if !column.Nullable {
		def += " NOT NULL"
	}
	if column.Default != "" {
		def += fmt.Sprintf(" DEFAULT %s", column.Default)
	}
	return fmt.Sprintf("ALTER TABLE \"%s\" ADD COLUMN %s", tableName, def)
}

func (d *PostgreSQLDialect) GetModifyColumnSQL(tableName string, column Column) string {
	def := fmt.Sprintf("\"%s\" %s", column.Name, column.Type)
	if column.Length > 0 {
		def = fmt.Sprintf("%s(%d)", def, column.Length)
	}
	if !column.Nullable {
		def += " NOT NULL"
	}
	if column.Default != "" {
		def += fmt.Sprintf(" DEFAULT %s", column.Default)
	}
	return fmt.Sprintf("ALTER TABLE \"%s\" ALTER COLUMN %s", tableName, def)
}

func (d *PostgreSQLDialect) GetDropColumnSQL(tableName string, columnName string) string {
	return fmt.Sprintf("ALTER TABLE \"%s\" DROP COLUMN \"%s\"", tableName, columnName)
}

func (d *PostgreSQLDialect) GetAddIndexSQL(tableName string, indexName string, columns []string, unique bool) string {
	uniqueStr := ""
	if unique {
		uniqueStr = "UNIQUE "
	}
	columnList := make([]string, len(columns))
	for i, col := range columns {
		columnList[i] = fmt.Sprintf("\"%s\"", col)
	}
	return fmt.Sprintf("CREATE %sINDEX \"%s\" ON \"%s\" (%s)", 
		uniqueStr, indexName, tableName, strings.Join(columnList, ", "))
}

func (d *PostgreSQLDialect) GetDropIndexSQL(tableName string, indexName string) string {
	return fmt.Sprintf("DROP INDEX \"%s\"", indexName)
}

func (d *PostgreSQLDialect) GetLimitSQL(limit, offset int) string {
	if offset > 0 {
		return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
	}
	return fmt.Sprintf("LIMIT %d", limit)
}

func (d *PostgreSQLDialect) GetCurrentTimeSQL() string {
	return "CURRENT_TIMESTAMP"
}

func (d *PostgreSQLDialect) GetConcatSQL(values ...string) string {
	return fmt.Sprintf("CONCAT(%s)", strings.Join(values, ", "))
}

func (d *PostgreSQLDialect) GetIfNullSQL(expr, defaultValue string) string {
	return fmt.Sprintf("COALESCE(%s, %s)", expr, defaultValue)
}

// SQLiteDialect SQLite方言
type SQLiteDialect struct{}

func (d *SQLiteDialect) GetType() string {
	return "sqlite"
}

func (d *SQLiteDialect) GetPlaceholder(index int) string {
	return "?"
}

func (d *SQLiteDialect) GetAutoIncrement() string {
	return "AUTOINCREMENT"
}

func (d *SQLiteDialect) GetCreateTableSQL(tableName string, columns []Column) string {
	var columnDefs []string
	var primaryKeys []string
	
	for _, col := range columns {
		def := fmt.Sprintf("\"%s\" %s", col.Name, col.Type)
		
		if col.Length > 0 {
			def = fmt.Sprintf("%s(%d)", def, col.Length)
		}
		
		if !col.Nullable {
			def += " NOT NULL"
		}
		
		if col.Default != "" {
			def += fmt.Sprintf(" DEFAULT %s", col.Default)
		}
		
		if col.AutoIncr {
			def += " AUTOINCREMENT"
		}
		
		if col.PrimaryKey {
			primaryKeys = append(primaryKeys, fmt.Sprintf("\"%s\"", col.Name))
		}
		
		columnDefs = append(columnDefs, def)
	}
	
	if len(primaryKeys) > 0 {
		columnDefs = append(columnDefs, fmt.Sprintf("PRIMARY KEY (%s)", strings.Join(primaryKeys, ", ")))
	}
	
	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS \"%s\" (\n  %s\n)", 
		tableName, strings.Join(columnDefs, ",\n  "))
}

func (d *SQLiteDialect) GetAddColumnSQL(tableName string, column Column) string {
	def := fmt.Sprintf("\"%s\" %s", column.Name, column.Type)
	if column.Length > 0 {
		def = fmt.Sprintf("%s(%d)", def, column.Length)
	}
	if !column.Nullable {
		def += " NOT NULL"
	}
	if column.Default != "" {
		def += fmt.Sprintf(" DEFAULT %s", column.Default)
	}
	return fmt.Sprintf("ALTER TABLE \"%s\" ADD COLUMN %s", tableName, def)
}

func (d *SQLiteDialect) GetModifyColumnSQL(tableName string, column Column) string {
	// SQLite不支持直接修改列，需要重建表
	return fmt.Sprintf("-- SQLite不支持直接修改列，需要重建表: %s", tableName)
}

func (d *SQLiteDialect) GetDropColumnSQL(tableName string, columnName string) string {
	// SQLite不支持直接删除列，需要重建表
	return fmt.Sprintf("-- SQLite不支持直接删除列，需要重建表: %s", tableName)
}

func (d *SQLiteDialect) GetAddIndexSQL(tableName string, indexName string, columns []string, unique bool) string {
	uniqueStr := ""
	if unique {
		uniqueStr = "UNIQUE "
	}
	columnList := make([]string, len(columns))
	for i, col := range columns {
		columnList[i] = fmt.Sprintf("\"%s\"", col)
	}
	return fmt.Sprintf("CREATE %sINDEX \"%s\" ON \"%s\" (%s)", 
		uniqueStr, indexName, tableName, strings.Join(columnList, ", "))
}

func (d *SQLiteDialect) GetDropIndexSQL(tableName string, indexName string) string {
	return fmt.Sprintf("DROP INDEX \"%s\"", indexName)
}

func (d *SQLiteDialect) GetLimitSQL(limit, offset int) string {
	if offset > 0 {
		return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
	}
	return fmt.Sprintf("LIMIT %d", limit)
}

func (d *SQLiteDialect) GetCurrentTimeSQL() string {
	return "datetime('now')"
}

func (d *SQLiteDialect) GetConcatSQL(values ...string) string {
	return fmt.Sprintf("(%s)", strings.Join(values, " || "))
}

func (d *SQLiteDialect) GetIfNullSQL(expr, defaultValue string) string {
	return fmt.Sprintf("COALESCE(%s, %s)", expr, defaultValue)
}

// GetDialect 根据数据库类型获取对应的方言
func GetDialect(dbType string) (Dialect, error) {
	switch strings.ToLower(dbType) {
	case "mysql":
		return &MySQLDialect{}, nil
	case "postgresql", "postgres":
		return &PostgreSQLDialect{}, nil
	case "sqlite":
		return &SQLiteDialect{}, nil
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", dbType)
	}
} 