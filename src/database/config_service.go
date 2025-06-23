package database

import (
	"encoding/json"
	"fmt"
	"strconv"

	"ai-server-go/src/core/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ConfigService 配置管理服务
type ConfigService struct {
	db     *Database
	logger *utils.Logger
}

// NewConfigService 创建配置管理服务
func NewConfigService(db *Database, logger *utils.Logger) *ConfigService {
	return &ConfigService{
		db:     db,
		logger: logger,
	}
}

// GetDB 获取数据库连接
func (s *ConfigService) GetDB() *Database {
	return s.db
}

// GetGlobalConfig 获取全局配置
func (s *ConfigService) GetGlobalConfig(key string) (*GlobalConfig, error) {
	var config GlobalConfig
	if err := s.db.DB.Where("config_key = ?", key).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询全局配置失败: %v", err)
	}
	return &config, nil
}

// SetGlobalConfig 设置全局配置
func (s *ConfigService) SetGlobalConfig(key, value, configType, description string, isSystem bool) error {
	var config GlobalConfig
	if err := s.db.DB.Where("config_key = ?", key).First(&config).Error; err == nil {
		// 更新现有配置
		config.ConfigValue = value
		config.ConfigType = configType
		config.Description = description
		config.IsSystem = isSystem
		if err := s.db.DB.Save(&config).Error; err != nil {
			return fmt.Errorf("更新全局配置失败: %v", err)
		}
	} else {
		// 创建新配置
		config = GlobalConfig{
			ConfigKey:   key,
			ConfigValue: value,
			ConfigType:  configType,
			Description: description,
			IsSystem:    isSystem,
		}
		if err := s.db.DB.Create(&config).Error; err != nil {
			return fmt.Errorf("创建全局配置失败: %v", err)
		}
	}

	s.logger.Info("全局配置设置成功: %s", key)
	return nil
}

// GetGlobalConfigValue 获取全局配置值
func (s *ConfigService) GetGlobalConfigValue(key string) (string, error) {
	config, err := s.GetGlobalConfig(key)
	if err != nil {
		return "", err
	}
	if config == nil {
		return "", fmt.Errorf("配置不存在: %s", key)
	}
	return config.ConfigValue, nil
}

// GetGlobalConfigInt 获取全局配置整数值
func (s *ConfigService) GetGlobalConfigInt(key string) (int, error) {
	value, err := s.GetGlobalConfigValue(key)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(value)
}

// GetGlobalConfigBool 获取全局配置布尔值
func (s *ConfigService) GetGlobalConfigBool(key string) (bool, error) {
	value, err := s.GetGlobalConfigValue(key)
	if err != nil {
		return false, err
	}
	return strconv.ParseBool(value)
}

// ListGlobalConfigs 获取全局配置列表
func (s *ConfigService) ListGlobalConfigs() ([]*GlobalConfig, error) {
	var configs []*GlobalConfig
	if err := s.db.DB.Find(&configs).Error; err != nil {
		return nil, fmt.Errorf("查询全局配置列表失败: %v", err)
	}
	return configs, nil
}

// DeleteGlobalConfig 删除全局配置
func (s *ConfigService) DeleteGlobalConfig(key string) error {
	if err := s.db.DB.Where("config_key = ?", key).Delete(&GlobalConfig{}).Error; err != nil {
		return fmt.Errorf("删除全局配置失败: %v", err)
	}

	s.logger.Info("全局配置删除成功: %s", key)
	return nil
}

// GetAICapability 获取AI能力
func (s *ConfigService) GetAICapability(name, capabilityType string) (*AICapability, error) {
	var capability AICapability
	if err := s.db.DB.Where("capability_name = ? AND capability_type = ?", name, capabilityType).First(&capability).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询AI能力失败: %v", err)
	}
	return &capability, nil
}

// CreateAICapability 创建AI能力
func (s *ConfigService) CreateAICapability(capability *AICapability) error {
	if err := s.db.DB.Create(capability).Error; err != nil {
		return fmt.Errorf("创建AI能力失败: %v", err)
	}

	s.logger.Info("AI能力创建成功: %s/%s", capability.CapabilityName, capability.CapabilityType)
	return nil
}

// UpdateAICapability 更新AI能力
func (s *ConfigService) UpdateAICapability(capability *AICapability) error {
	if err := s.db.DB.Save(capability).Error; err != nil {
		return fmt.Errorf("更新AI能力失败: %v", err)
	}

	s.logger.Info("AI能力更新成功: %s/%s", capability.CapabilityName, capability.CapabilityType)
	return nil
}

// DeleteAICapability 删除AI能力
func (s *ConfigService) DeleteAICapability(id uint) error {
	if err := s.db.DB.Delete(&AICapability{}, id).Error; err != nil {
		return fmt.Errorf("删除AI能力失败: %v", err)
	}

	s.logger.Info("AI能力删除成功: ID %d", id)
	return nil
}

// ListAICapabilities 获取AI能力列表
func (s *ConfigService) ListAICapabilities(capabilityType string) ([]*AICapability, error) {
	var capabilities []*AICapability
	query := s.db.DB
	if capabilityType != "" {
		query = query.Where("capability_type = ?", capabilityType)
	}
	if err := query.Find(&capabilities).Error; err != nil {
		return nil, fmt.Errorf("查询AI能力列表失败: %v", err)
	}
	return capabilities, nil
}

// GetProviderConfig 获取提供商配置
func (s *ConfigService) GetProviderConfig(id uint) (*ProviderConfig, error) {
	var config ProviderConfig
	if err := s.db.DB.First(&config, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询提供商配置失败: %v", err)
	}
	return &config, nil
}

// CreateProviderConfig 创建提供商配置
func (s *ConfigService) CreateProviderConfig(config *ProviderConfig) error {
	if err := s.db.DB.Create(config).Error; err != nil {
		return fmt.Errorf("创建提供商配置失败: %v", err)
	}

	s.logger.Info("提供商配置创建成功: %s/%s/%s", config.Category, config.Name, config.Version)
	return nil
}

// UpdateProviderConfig 更新提供商配置
func (s *ConfigService) UpdateProviderConfig(config *ProviderConfig) error {
	if err := s.db.DB.Save(config).Error; err != nil {
		return fmt.Errorf("更新提供商配置失败: %v", err)
	}

	s.logger.Info("提供商配置更新成功: %s/%s/%s", config.Category, config.Name, config.Version)
	return nil
}

// DeleteProviderConfig 删除提供商配置
func (s *ConfigService) DeleteProviderConfig(id uint) error {
	if err := s.db.DB.Delete(&ProviderConfig{}, id).Error; err != nil {
		return fmt.Errorf("删除提供商配置失败: %v", err)
	}

	s.logger.Info("提供商配置删除成功: ID %d", id)
	return nil
}

// ListProviderConfigs 获取提供商配置列表
func (s *ConfigService) ListProviderConfigs(category string) ([]*ProviderConfig, error) {
	var configs []*ProviderConfig
	query := s.db.DB
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if err := query.Order("category, name, version").Find(&configs).Error; err != nil {
		return nil, fmt.Errorf("查询提供商配置列表失败: %v", err)
	}
	return configs, nil
}

// GetProviderConfigsByCategory 根据类别获取提供商配置
func (s *ConfigService) GetProviderConfigsByCategory(category string) ([]*ProviderConfig, error) {
	var configs []*ProviderConfig
	if err := s.db.DB.Where("category = ? AND is_active = ?", category, true).
		Order("weight DESC, is_default DESC, created_at ASC").
		Find(&configs).Error; err != nil {
		return nil, fmt.Errorf("查询提供商配置失败: %v", err)
	}
	return configs, nil
}

// GetDefaultProviderConfig 获取默认提供商配置
func (s *ConfigService) GetDefaultProviderConfig(category string) (*ProviderConfig, error) {
	var config ProviderConfig
	if err := s.db.DB.Where("category = ? AND is_default = ? AND is_active = ?", category, true, true).
		First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询默认提供商配置失败: %v", err)
	}
	return &config, nil
}

// GetProviderVersions 获取提供商版本列表
func (s *ConfigService) GetProviderVersions(category, name string) ([]*ProviderVersion, error) {
	var configs []*ProviderConfig
	if err := s.db.DB.Where("category = ? AND name = ?", category, name).
		Order("version").
		Find(&configs).Error; err != nil {
		return nil, fmt.Errorf("查询提供商版本失败: %v", err)
	}

	var versions []*ProviderVersion
	for _, config := range configs {
		version := &ProviderVersion{
			Category:  config.Category,
			Name:      config.Name,
			Version:   config.Version,
			Weight:    config.Weight,
			IsActive:  config.IsActive,
			IsDefault: config.IsDefault,
			CreatedAt: config.CreatedAt,
			UpdatedAt: config.UpdatedAt,
		}
		versions = append(versions, version)
	}

	return versions, nil
}

// GetDefaultProviderModules 获取默认提供商模块
func (s *ConfigService) GetDefaultProviderModules() (map[string]string, error) {
	var configs []*ProviderConfig
	if err := s.db.DB.Where("is_default = ? AND is_active = ?", true, true).
		Select("category, name").
		Find(&configs).Error; err != nil {
		return nil, fmt.Errorf("查询默认提供商模块失败: %v", err)
	}

	modules := make(map[string]string)
	for _, config := range configs {
		modules[config.Category] = config.Name
	}

	return modules, nil
}

// GetProviderConfigByCategoryAndName 根据类别和名称获取提供商配置
func (s *ConfigService) GetProviderConfigByCategoryAndName(category, name string) (*ProviderConfig, error) {
	var config ProviderConfig
	if err := s.db.DB.Where("category = ? AND name = ? AND is_active = ?", category, name, true).
		First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询提供商配置失败: %v", err)
	}
	return &config, nil
}

// GetProviderConfigByCategoryAndType 根据类别和类型获取提供商配置
func (s *ConfigService) GetProviderConfigByCategoryAndType(category, providerType string) (*ProviderConfig, error) {
	var config ProviderConfig
	if err := s.db.DB.Where("category = ? AND type = ? AND is_active = ?", category, providerType, true).
		First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询提供商配置失败: %v", err)
	}
	return &config, nil
}

// GetSystemConfig 获取系统配置
func (s *ConfigService) GetSystemConfig(category, key string) (*SystemConfig, error) {
	var config SystemConfig
	if err := s.db.DB.Where("config_category = ? AND config_key = ?", category, key).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询系统配置失败: %v", err)
	}
	return &config, nil
}

// SetSystemConfig 设置系统配置
func (s *ConfigService) SetSystemConfig(category, key, value, configType, description string, isDefault bool, createdBy, updatedBy *uint) error {
	var config SystemConfig
	if err := s.db.DB.Where("config_category = ? AND config_key = ?", category, key).First(&config).Error; err == nil {
		// 更新现有配置
		config.ConfigValue = value
		config.ConfigType = configType
		config.Description = description
		config.IsDefault = isDefault
		config.UpdatedBy = updatedBy
		if err := s.db.DB.Save(&config).Error; err != nil {
			return fmt.Errorf("更新系统配置失败: %v", err)
		}
	} else {
		// 创建新配置
		config = SystemConfig{
			ConfigCategory: category,
			ConfigKey:      key,
			ConfigValue:    value,
			ConfigType:     configType,
			Description:    description,
			IsDefault:      isDefault,
			CreatedBy:      createdBy,
			UpdatedBy:      updatedBy,
		}
		if err := s.db.DB.Create(&config).Error; err != nil {
			return fmt.Errorf("创建系统配置失败: %v", err)
		}
	}

	s.logger.Info("系统配置设置成功: %s/%s", category, key)
	return nil
}

// ListSystemConfigs 获取系统配置列表
func (s *ConfigService) ListSystemConfigs(category string) ([]*SystemConfig, error) {
	var configs []*SystemConfig
	query := s.db.DB
	if category != "" {
		query = query.Where("config_category = ?", category)
	}
	if err := query.Find(&configs).Error; err != nil {
		return nil, fmt.Errorf("查询系统配置列表失败: %v", err)
	}
	return configs, nil
}

// DeleteSystemConfig 删除系统配置
func (s *ConfigService) DeleteSystemConfig(category, key string) error {
	if err := s.db.DB.Where("config_category = ? AND config_key = ?", category, key).Delete(&SystemConfig{}).Error; err != nil {
		return fmt.Errorf("删除系统配置失败: %v", err)
	}

	s.logger.Info("系统配置删除成功: %s/%s", category, key)
	return nil
}

// GetSystemConfigValue 获取系统配置值
func (s *ConfigService) GetSystemConfigValue(category, key string) (string, error) {
	config, err := s.GetSystemConfig(category, key)
	if err != nil {
		return "", err
	}
	if config == nil {
		return "", fmt.Errorf("系统配置不存在: %s/%s", category, key)
	}
	return config.ConfigValue, nil
}

// GetSystemConfigInt 获取系统配置整数值
func (s *ConfigService) GetSystemConfigInt(category, key string) (int, error) {
	value, err := s.GetSystemConfigValue(category, key)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(value)
}

// GetSystemConfigBool 获取系统配置布尔值
func (s *ConfigService) GetSystemConfigBool(category, key string) (bool, error) {
	value, err := s.GetSystemConfigValue(category, key)
	if err != nil {
		return false, err
	}
	return strconv.ParseBool(value)
}

// GetSystemConfigFloat 获取系统配置浮点数值
func (s *ConfigService) GetSystemConfigFloat(category, key string) (float64, error) {
	value, err := s.GetSystemConfigValue(category, key)
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(value, 64)
}

// GetSystemConfigJSON 获取系统配置JSON值
func (s *ConfigService) GetSystemConfigJSON(category, key string) (map[string]interface{}, error) {
	value, err := s.GetSystemConfigValue(category, key)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = json.Unmarshal([]byte(value), &result)
	if err != nil {
		return nil, fmt.Errorf("解析JSON配置失败: %v", err)
	}
	return result, nil
}

// GetSystemConfigArray 获取系统配置数组值
func (s *ConfigService) GetSystemConfigArray(category, key string) ([]string, error) {
	value, err := s.GetSystemConfigValue(category, key)
	if err != nil {
		return nil, err
	}

	var result []string
	err = json.Unmarshal([]byte(value), &result)
	if err != nil {
		return nil, fmt.Errorf("解析数组配置失败: %v", err)
	}
	return result, nil
}

// InitializeDefaultSystemConfigs 初始化默认系统配置
func (s *ConfigService) InitializeDefaultSystemConfigs() error {
	defaultConfigs := []struct {
		category    string
		key         string
		value       string
		configType  string
		description string
	}{
		// 提示词配置
		{"prompt", "default_prompt", "你是小智/小志，来自中国台湾省的00后女生。讲话超级机车，\"真的假的啦\"这样的台湾腔，喜欢用\"笑死\"\"是在哈喽\"等流行梗，但会偷偷研究男友的编程书籍。", "string", "默认AI提示词"},

		// 音频处理配置
		{"audio", "delete_audio", "true", "bool", "是否删除音频文件"},
		{"audio", "quick_reply", "true", "bool", "是否启用快速回复"},
		{"audio", "quick_reply_words", "[\"我在\", \"在呢\", \"来了\", \"啥事啊\"]", "array", "快速回复词汇"},

		// AI提供商默认配置
		{"ai_providers", "default_asr", "DoubaoASR", "string", "默认ASR提供商"},
		{"ai_providers", "default_tts", "EdgeTTS", "string", "默认TTS提供商"},
		{"ai_providers", "default_llm", "OllamaLLM", "string", "默认LLM提供商"},
		{"ai_providers", "default_vlllm", "ChatGLMVLLM", "string", "默认VLLLM提供商"},

		// 连通性检查配置
		{"connectivity", "enabled", "false", "bool", "是否启用连通性检查"},
		{"connectivity", "timeout", "30s", "string", "检查超时时间"},
		{"connectivity", "retry_attempts", "3", "int", "重试次数"},
		{"connectivity", "retry_delay", "5s", "string", "重试延迟"},
		{"connectivity", "asr_test_audio", "", "string", "ASR测试音频文件"},
		{"connectivity", "llm_test_prompt", "Hello", "string", "LLM测试提示词"},
		{"connectivity", "tts_test_text", "测试", "string", "TTS测试文本"},
	}

	for _, config := range defaultConfigs {
		// 检查是否已存在
		existing, err := s.GetSystemConfig(config.category, config.key)
		if err != nil {
			return fmt.Errorf("检查系统配置失败: %v", err)
		}

		// 如果不存在，则创建默认配置
		if existing == nil {
			err = s.SetSystemConfig(config.category, config.key, config.value, config.configType, config.description, true, nil, nil)
			if err != nil {
				return fmt.Errorf("初始化系统配置失败: %v", err)
			}
		}
	}

	s.logger.Info("默认系统配置初始化完成")
	return nil
}

// InitializeDefaultProviderConfigs 初始化默认的AI提供商配置
func (s *ConfigService) InitializeDefaultProviderConfigs() error {
	defaultProviders := []ProviderConfig{
		// EdgeTTS (默认)
		{
			Category:  "TTS",
			Name:      "EdgeTTS",
			Type:      "edge",
			Version:   "v1",
			Weight:    100,
			IsActive:  true,
			IsDefault: true,
			Props:     []byte(`{"voice": "zh-CN-XiaoxiaoNeural", "output_dir": "tmp/"}`),
		},
		// DoubaoTTS
		{
			Category:  "TTS",
			Name:      "DoubaoTTS",
			Type:      "doubao",
			Version:   "v1",
			Weight:    90,
			IsActive:  true,
			IsDefault: false,
			Props:     []byte(`{"base_url": "https://api.doubao.com/volcengine/v1", "voice": "zh_female_wanwanxiaohe_moon_bigtts", "output_dir": "tmp/", "appid": "你的appid", "token": "你的access_token", "cluster": "你的cluster"}`),
		},
		// gosherpa-onnx TTS
		{
			Category:  "TTS",
			Name:      "GosherpaTTS",
			Type:      "gosherpa",
			Version:   "v1",
			Weight:    80,
			IsActive:  true,
			IsDefault: false,
			Props:     []byte(`{"cluster": "ws://127.0.0.1:8848/tts", "output_dir": "tmp/"}`),
		},
		// DoubaoASR
		{
			Category:  "ASR",
			Name:      "DoubaoASR",
			Type:      "doubao",
			Version:   "v1",
			Weight:    100,
			IsActive:  true,
			IsDefault: true,
			Props:     []byte(`{"host":"openspeech.bytedance.com","ws_url":"wss://openspeech.bytedance.com/api/v3/sauc/bigmodel_nostream","model_name":"bigmodel","chunk_duration":200,"end_window_size":800,"appid": "你的appid"}`),
		},
		// ChatGLM LLM
		{
			Category:  "LLM",
			Name:      "ChatGLMLLM",
			Type:      "openai",
			Version:   "v1",
			Weight:    100,
			IsActive:  true,
			IsDefault: false,
			Props:     []byte(`{"base_url": "https://open.bigmodel.cn/api/paas/v4/", "api_key": "你的api_key", "model_name": "glm-4-flash"}`),
		},
		// Ollama LLM (默认)
		{
			Category:  "LLM",
			Name:      "OllamaLLM",
			Type:      "ollama",
			Version:   "v1",
			Weight:    90,
			IsActive:  true,
			IsDefault: true,
			Props:     []byte(`{"base_url": "http://localhost:11434", "model_name": "qwen3"}`),
		},
		// ChatGLM VLLLM (默认)
		{
			Category:  "VLLLM",
			Name:      "ChatGLMVLLM",
			Type:      "openai",
			Version:   "v1",
			Weight:    100,
			IsActive:  true,
			IsDefault: true,
			Props:     []byte(`{"base_url": "https://open.bigmodel.cn/api/paas/v4/", "api_key": "你的api_key", "model_name": "glm-4v-flash", "temperature": 0.7, "max_tokens": 4096, "top_p": 0.9, "security": {"max_file_size": 10485760, "max_pixels": 16777216, "max_width": 4096, "max_height": 4096, "allowed_formats": ["jpeg", "jpg", "png", "webp", "gif"], "enable_deep_scan": true, "validation_timeout": "10s"}}`),
		},
		// Ollama VLLLM
		{
			Category:  "VLLLM",
			Name:      "OllamaVLLM",
			Type:      "ollama",
			Version:   "v1",
			Weight:    90,
			IsActive:  true,
			IsDefault: false,
			Props:     []byte(`{"base_url": "http://localhost:11434", "model_name": "qwen2.5vl", "temperature": 0.7, "max_tokens": 4096, "top_p": 0.9, "security": {"max_file_size": 10485760, "max_pixels": 16777216, "max_width": 4096, "max_height": 4096, "allowed_formats": ["jpeg", "jpg", "png", "webp", "gif"], "enable_deep_scan": true, "validation_timeout": "10s"}}`),
		},
	}

	for _, provider := range defaultProviders {
		var existing ProviderConfig
		err := s.db.DB.Where("category = ? AND name = ? AND version = ?", provider.Category, provider.Name, provider.Version).First(&existing).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := s.db.DB.Create(&provider).Error; err != nil {
					s.logger.Error("创建默认Provider配置失败: %v", err)
					return err
				}
			} else {
				s.logger.Error("查询默认Provider配置失败: %v", err)
				return err
			}
		}
	}

	s.logger.Info("默认Provider配置初始化完成")
	return nil
}

// GetSystemConfigCategory 获取指定分类的所有系统配置
func (s *ConfigService) GetSystemConfigCategory(category string) (map[string]interface{}, error) {
	configs, err := s.ListSystemConfigs(category)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for _, config := range configs {
		switch config.ConfigType {
		case "int":
			if val, err := strconv.Atoi(config.ConfigValue); err == nil {
				result[config.ConfigKey] = val
			} else {
				result[config.ConfigKey] = config.ConfigValue
			}
		case "float":
			if val, err := strconv.ParseFloat(config.ConfigValue, 64); err == nil {
				result[config.ConfigKey] = val
			} else {
				result[config.ConfigKey] = config.ConfigValue
			}
		case "bool":
			if val, err := strconv.ParseBool(config.ConfigValue); err == nil {
				result[config.ConfigKey] = val
			} else {
				result[config.ConfigKey] = config.ConfigValue
			}
		case "json":
			var val interface{}
			if err := json.Unmarshal([]byte(config.ConfigValue), &val); err == nil {
				result[config.ConfigKey] = val
			} else {
				result[config.ConfigKey] = config.ConfigValue
			}
		case "array":
			var val []string
			if err := json.Unmarshal([]byte(config.ConfigValue), &val); err == nil {
				result[config.ConfigKey] = val
			} else {
				result[config.ConfigKey] = config.ConfigValue
			}
		default:
			result[config.ConfigKey] = config.ConfigValue
		}
	}

	return result, nil
}

// GetActiveProviderConfigs 获取指定category和name下所有激活的ProviderConfig
func (s *ConfigService) GetActiveProviderConfigs(category, name string) ([]*ProviderConfig, error) {
	var configs []*ProviderConfig
	if err := s.db.DB.Where("category = ? AND name = ? AND is_active = ?", category, name, true).
		Order("version").
		Find(&configs).Error; err != nil {
		return nil, fmt.Errorf("查询激活的提供商配置失败: %v", err)
	}
	return configs, nil
}

// UpdateProviderWeight 更新提供商权重
func (s *ConfigService) UpdateProviderWeight(category, name, version string, weight int) error {
	if err := s.db.DB.Model(&ProviderConfig{}).
		Where("category = ? AND name = ? AND version = ?", category, name, version).
		Update("weight", weight).Error; err != nil {
		return fmt.Errorf("更新提供商权重失败: %v", err)
	}
	return nil
}

// GetDeviceCapabilityConfigWithFallback 获取设备AI能力配置（带回退）
func (s *ConfigService) GetDeviceCapabilityConfigWithFallback(deviceID uint, userID *uint) (*DeviceCapabilityConfig, error) {
	deviceConfig := &DeviceCapabilityConfig{
		DeviceID:      deviceID,
		Capabilities:  []CapabilityConfig{},
		GlobalConfigs: make(map[string]string),
	}

	// 1. 设备专属能力
	var deviceCaps []DeviceCapability
	err := s.db.DB.Where("device_id = ? AND is_enabled = ?", deviceID, true).Preload("Capability").Order("priority DESC").Find(&deviceCaps).Error
	if err != nil {
		return nil, err
	}
	used := make(map[string]bool)
	for _, dc := range deviceCaps {
		if dc.Capability.CapabilityName == "" {
			continue
		}
		cc := CapabilityConfig{
			CapabilityName: dc.Capability.CapabilityName,
			CapabilityType: dc.Capability.CapabilityType,
			Priority:       dc.Priority,
			IsEnabled:      dc.IsEnabled,
		}
		if err := json.Unmarshal(dc.ConfigData, &cc.Config); err != nil {
			s.logger.Warn("解析设备能力配置失败: %v", err)
			cc.Config = make(map[string]interface{})
		}
		cc.Config["priority_source"] = "device"
		deviceConfig.Capabilities = append(deviceConfig.Capabilities, cc)
		used[cc.CapabilityName+"/"+cc.CapabilityType] = true
	}

	// 2. 用户能力（如果提供了userID）
	if userID != nil {
		var userCaps []UserCapability
		err := s.db.DB.Where("user_id = ? AND is_active = ?", *userID, true).Preload("Capability").Find(&userCaps).Error
		if err == nil {
			for _, uc := range userCaps {
				key := uc.Capability.CapabilityName + "/" + uc.Capability.CapabilityType
				if _, ok := used[key]; ok {
					continue
				}
				cc := CapabilityConfig{
					CapabilityName: uc.Capability.CapabilityName,
					CapabilityType: uc.Capability.CapabilityType,
					IsEnabled:      uc.IsActive,
				}
				if err := json.Unmarshal(uc.ConfigData, &cc.Config); err != nil {
					s.logger.Warn("解析用户能力配置失败: %v", err)
					cc.Config = make(map[string]interface{})
				}
				cc.Config["priority_source"] = "user"
				deviceConfig.Capabilities = append(deviceConfig.Capabilities, cc)
				used[key] = true
			}
		}
	}

	// 3. 系统默认能力
	defaultCaps, _ := s.GetDefaultCapabilities()
	for _, dc := range defaultCaps {
		key := dc.CapabilityName + "/" + dc.CapabilityType
		if _, ok := used[key]; ok {
			continue
		}
		cc := CapabilityConfig{
			CapabilityName: dc.CapabilityName,
			CapabilityType: dc.CapabilityType,
			IsEnabled:      true,
		}
		if err := json.Unmarshal(dc.ConfigSchema, &cc.Config); err != nil {
			s.logger.Warn("解析系统默认能力配置失败: %v", err)
			cc.Config = make(map[string]interface{})
		}
		cc.Config["priority_source"] = "system"
		deviceConfig.Capabilities = append(deviceConfig.Capabilities, cc)
		used[key] = true
	}

	// 4. 全局配置
	globalConfigs, _ := s.ListGlobalConfigs()
	for _, gc := range globalConfigs {
		deviceConfig.GlobalConfigs[gc.ConfigKey] = gc.ConfigValue
	}

	return deviceConfig, nil
}

// GetDefaultCapabilities 获取默认的AI能力
func (s *ConfigService) GetDefaultCapabilities() ([]*AICapability, error) {
	var defaults []*AICapability
	if err := s.db.DB.Where("is_global = ? AND is_active = ?", true, true).Find(&defaults).Error; err != nil {
		return nil, err
	}
	if defaults == nil {
		defaults = []*AICapability{}
	}
	return defaults, nil
}

// GetEffectiveProvider 获取设备/用户/系统默认 Provider
func (s *ConfigService) GetEffectiveProvider(category string, deviceID *uint, userID *uint) (*ProviderConfig, error) {
	// 1. 设备专属
	if deviceID != nil {
		var dp DeviceProvider
		if err := s.db.DB.Where("device_id = ? AND category = ? AND is_active = ?", *deviceID, category, true).First(&dp).Error; err == nil {
			var provider ProviderConfig
			if err := s.db.DB.First(&provider, dp.ProviderID).Error; err == nil {
				return &provider, nil
			}
		}
	}
	// 2. 用户自定义
	if userID != nil {
		var up UserProvider
		if err := s.db.DB.Where("user_id = ? AND category = ? AND is_active = ?", *userID, category, true).First(&up).Error; err == nil {
			var provider ProviderConfig
			if err := s.db.DB.First(&provider, up.ProviderID).Error; err == nil {
				return &provider, nil
			}
		}
	}
	// 3. 系统默认
	return s.GetDefaultProviderConfig(category)
}

// 用户绑定Provider
// POST /api/user/provider/bind
func (s *ConfigService) BindUserProvider(c *gin.Context) {
	var req struct {
		UserID     uint   `json:"user_id" binding:"required"`
		ProviderID uint   `json:"provider_id" binding:"required"`
		Category   string `json:"category" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	var up UserProvider
	db := s.db.DB
	if err := db.Where("user_id = ? AND category = ?", req.UserID, req.Category).
		Assign(UserProvider{
			ProviderID: req.ProviderID,
			IsActive:   true,
		}).
		FirstOrCreate(&up).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"success": true})
}

// 设备绑定Provider
// POST /api/device/provider/bind
func (s *ConfigService) BindDeviceProvider(c *gin.Context) {
	var req struct {
		DeviceID   uint   `json:"device_id" binding:"required"`
		ProviderID uint   `json:"provider_id" binding:"required"`
		Category   string `json:"category" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	var dp DeviceProvider
	db := s.db.DB
	if err := db.Where("device_id = ? AND category = ?", req.DeviceID, req.Category).
		Assign(DeviceProvider{
			ProviderID: req.ProviderID,
			IsActive:   true,
		}).
		FirstOrCreate(&dp).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"success": true})
}

// 查询用户绑定Provider
// GET /api/user/provider/list?user_id=123
func (s *ConfigService) ListUserProviders(c *gin.Context) {
	userID := c.Query("user_id")
	var providers []UserProvider
	if err := s.db.DB.Where("user_id = ? AND is_active = ?", userID, true).Find(&providers).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, providers)
}

// 查询设备绑定Provider
// GET /api/device/provider/list?device_id=456
func (s *ConfigService) ListDeviceProviders(c *gin.Context) {
	deviceID := c.Query("device_id")
	var providers []DeviceProvider
	if err := s.db.DB.Where("device_id = ? AND is_active = ?", deviceID, true).Find(&providers).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, providers)
}

// 解绑用户Provider
// POST /api/user/provider/unbind
func (s *ConfigService) UnbindUserProvider(c *gin.Context) {
	var req struct {
		UserID   uint   `json:"user_id" binding:"required"`
		Category string `json:"category" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := s.db.DB.Model(&UserProvider{}).
		Where("user_id = ? AND category = ?", req.UserID, req.Category).
		Update("is_active", false).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"success": true})
}

// 解绑设备Provider
// POST /api/device/provider/unbind
func (s *ConfigService) UnbindDeviceProvider(c *gin.Context) {
	var req struct {
		DeviceID uint   `json:"device_id" binding:"required"`
		Category string `json:"category" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := s.db.DB.Model(&DeviceProvider{}).
		Where("device_id = ? AND category = ?", req.DeviceID, req.Category).
		Update("is_active", false).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"success": true})
}

// Provider列表API
// GET /api/provider/list?category=TTS
func (s *ConfigService) ListProviders(c *gin.Context) {
	category := c.Query("category")
	var providers []ProviderConfig
	db := s.db.DB
	if category != "" {
		db = db.Where("category = ? AND is_active = ?", category, true)
	} else {
		db = db.Where("is_active = ?", true)
	}
	if err := db.Order("category, name, version").Find(&providers).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, providers)
}
