package configs

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// TokenConfig Token配置
type TokenConfig struct {
	Token string `yaml:"token"`
}

// Config 主配置结构
type Config struct {
	Server struct {
		IP    string `yaml:"ip"`
		Port  int    `yaml:"port"`
		Token string
		Auth  struct {
			Enabled        bool          `yaml:"enabled"`
			AllowedDevices []string      `yaml:"allowed_devices"`
			Tokens         []TokenConfig `yaml:"tokens"`
		} `yaml:"auth"`
	} `yaml:"server"`

	Log struct {
		LogFormat string `yaml:"log_format"`
		LogLevel  string `yaml:"log_level"`
		LogDir    string `yaml:"log_dir"`
		LogFile   string `yaml:"log_file"`
	} `yaml:"log"`

	Web struct {
		Enabled   bool   `yaml:"enabled"`
		Port      int    `yaml:"port"`
		StaticDir string `yaml:"static_dir"`
		Websocket string `yaml:"websocket"`
		VisionURL string `yaml:"vision"`
	} `yaml:"web"`

	VAD      map[string]VADConfig `yaml:"VAD"`
	Database DatabaseConfig       `yaml:"database"`
}

// VADConfig VAD配置结构
type VADConfig struct {
	Type               string                 `yaml:"type"`
	ModelDir           string                 `yaml:"model_dir"`
	Threshold          float64                `yaml:"threshold"`
	MinSilenceDuration int                    `yaml:"min_silence_duration_ms"`
	Extra              map[string]interface{} `yaml:",inline"`
}

// DatabaseConfig 数据库配置结构
type DatabaseConfig struct {
	Type     string `yaml:"type"`     // 数据库类型：mysql, postgresql, sqlite
	Host     string `yaml:"host"`     // 主机地址（SQLite不需要）
	Port     int    `yaml:"port"`     // 端口（SQLite不需要）
	User     string `yaml:"user"`     // 用户名（SQLite不需要）
	Password string `yaml:"password"` // 密码（SQLite不需要）
	Name     string `yaml:"name"`     // 数据库名/文件路径
	Charset  string `yaml:"charset"`  // 字符集（MySQL专用）

	// PostgreSQL特定配置
	SSLMode string `yaml:"ssl_mode"` // SSL模式：disable, require, verify-ca, verify-full

	// SQLite特定配置
	FilePath string `yaml:"file_path"` // 文件路径（SQLite专用）

	// 连接池配置
	MaxOpenConns    int           `yaml:"max_open_conns"`    // 最大打开连接数
	MaxIdleConns    int           `yaml:"max_idle_conns"`    // 最大空闲连接数
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"` // 连接最大生命周期

	// 其他配置
	ParseTime bool   `yaml:"parse_time"` // 是否解析时间（MySQL专用）
	Loc       string `yaml:"loc"`        // 时区（MySQL专用）
}

// LoadConfig 从文件加载配置
func LoadConfig() (*Config, string, error) {
	path := ".config.yaml"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		path = "config.yaml"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, path, err
	}

	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, path, err
	}

	return config, path, nil
}
