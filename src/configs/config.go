package configs

import (
	"os"

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
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	Charset  string `yaml:"charset"`
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
