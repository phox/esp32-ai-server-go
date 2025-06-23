package pool

import (
	"ai-server-go/src/configs"
	"ai-server-go/src/core/mcp"
	"ai-server-go/src/core/providers"
	"ai-server-go/src/core/providers/asr"
	"ai-server-go/src/core/providers/llm"
	"ai-server-go/src/core/providers/tts"
	"ai-server-go/src/core/providers/vlllm"
	"ai-server-go/src/core/utils"
	"ai-server-go/src/database"
	"encoding/json"
	"fmt"
)

/*
* 工厂类，用于创建不同类型的资源池工厂。
* 通过数据库配置和提供者类型，动态创建资源池工厂。
* 支持ASR、LLM、TTS和VLLLM等多种提供者类型。
* 每个工厂实现了ResourceFactory接口，提供Create和Destroy方法。
 */

// ProviderFactory 简化的提供者工厂
type ProviderFactory struct {
	providerType     string
	config           interface{}
	logger           *utils.Logger
	params           map[string]interface{}  // 可选参数
	configService    *database.ConfigService // 数据库配置服务
	grayscaleManager *GrayscaleManager       // 灰度发布管理器
}

func (f *ProviderFactory) Create() (interface{}, error) {
	// provider初始化前输出配置
	if f.logger != nil {
		configJson, _ := json.MarshalIndent(f.config, "", "  ")
		f.logger.Debug("[ProviderFactory] 初始化provider，类型: %s，配置: %s", f.providerType, string(configJson))
	}
	provider, err := f.createProvider()
	// provider初始化后输出结果
	if f.logger != nil {
		if err != nil {
			f.logger.Error("[ProviderFactory] provider初始化失败，类型: %s，err: %v", f.providerType, err)
		} else {
			f.logger.Debug("[ProviderFactory] provider初始化成功，类型: %s", f.providerType)
		}
	}
	return provider, err
}

func (f *ProviderFactory) Destroy(resource interface{}) error {
	if provider, ok := resource.(providers.Provider); ok {
		return provider.Cleanup()
	}
	// 对于VLLLM，我们尝试调用Cleanup方法（如果存在）
	if resource != nil {
		// 使用反射或类型断言来调用Cleanup方法
		if cleaner, ok := resource.(interface{ Cleanup() error }); ok {
			return cleaner.Cleanup()
		}
	}
	return nil
}

// VLLLMConfig VLLLM配置结构（本地定义，避免循环依赖）
type VLLLMConfig struct {
	Type        string                 `json:"type"`
	ModelName   string                 `json:"model_name"`
	BaseURL     string                 `json:"url"`
	APIKey      string                 `json:"api_key"`
	Temperature float64                `json:"temperature"`
	MaxTokens   int                    `json:"max_tokens"`
	TopP        float64                `json:"top_p"`
	Security    map[string]interface{} `json:"security"`
	Extra       map[string]interface{} `json:"extra"`
}

func (f *ProviderFactory) createProvider() (interface{}, error) {
	switch f.providerType {
	case "asr":
		cfg := f.config.(*asr.Config)
		params := f.params
		delete_audio, _ := params["delete_audio"].(bool)
		asrType, _ := params["type"].(string)
		return asr.Create(asrType, cfg, delete_audio, f.logger)
	case "llm":
		cfg := f.config.(*llm.Config)
		return llm.Create(cfg.Type, cfg)
	case "tts":
		cfg := f.config.(*tts.Config)
		params := f.params
		delete_audio, _ := params["delete_audio"].(bool)
		return tts.Create(cfg.Type, cfg, delete_audio)
	case "vlllm":
		cfg := f.config.(*vlllm.VLLLMConfig)
		return vlllm.Create(cfg.Type, cfg, f.logger)
	case "mcp":
		_ = f.config.(*configs.Config)
		logger := f.logger
		return mcp.NewManagerForPool(logger), nil
	default:
		return nil, fmt.Errorf("未知的提供者类型: %s", f.providerType)
	}
}

// 创建各类型工厂的便利函数
func NewASRFactory(asrType string, configService *database.ConfigService, logger *utils.Logger, deleteAudio bool, grayscaleManager *GrayscaleManager) ResourceFactory {
	// 从灰度管理器获取ASR配置
	var providerConfig *database.ProviderConfig
	var err error

	if grayscaleManager != nil {
		providerConfig, err = grayscaleManager.GetProviderConfig("ASR", asrType)
	} else {
		providerConfig, err = configService.GetProviderConfigByCategoryAndName("ASR", asrType)
	}

	if err != nil {
		logger.Error("获取ASR配置失败: %v", err)
		return nil
	}

	return &ProviderFactory{
		providerType: "asr",
		config: &asr.Config{
			Type: providerConfig.Type,
		},
		logger: logger,
		params: map[string]interface{}{
			"type":         providerConfig.Type,
			"delete_audio": deleteAudio,
		},
		configService:    configService,
		grayscaleManager: grayscaleManager,
	}
}

func NewLLMFactory(llmType string, configService *database.ConfigService, logger *utils.Logger, grayscaleManager *GrayscaleManager) ResourceFactory {
	var providerConfig *database.ProviderConfig
	var err error

	if grayscaleManager != nil {
		providerConfig, err = grayscaleManager.GetProviderConfig("LLM", llmType)
	} else {
		providerConfig, err = configService.GetProviderConfigByCategoryAndName("LLM", llmType)
	}

	if err != nil {
		logger.Error("获取LLM配置失败: %v", err)
		return nil
	}

	var props map[string]interface{}
	if len(providerConfig.Props) > 0 {
		err := json.Unmarshal(providerConfig.Props, &props)
		if err != nil {
			logger.Error("解析LLM Props失败: %v", err)
			props = map[string]interface{}{}
		}
	} else {
		props = map[string]interface{}{}
	}

	return &ProviderFactory{
		providerType: "llm",
		config: &llm.Config{
			Type:  providerConfig.Type,
			Extra: props,
		},
		logger:           logger,
		configService:    configService,
		grayscaleManager: grayscaleManager,
	}
}

func NewTTSFactory(ttsType string, configService *database.ConfigService, logger *utils.Logger, deleteAudio bool, grayscaleManager *GrayscaleManager) ResourceFactory {
	// 从灰度管理器获取TTS配置
	var providerConfig *database.ProviderConfig
	var err error

	if grayscaleManager != nil {
		providerConfig, err = grayscaleManager.GetProviderConfig("TTS", ttsType)
	} else {
		providerConfig, err = configService.GetProviderConfigByCategoryAndName("TTS", ttsType)
	}

	if err != nil {
		logger.Error("获取TTS配置失败: %v", err)
		return nil
	}

	// 反序列化Props
	var props map[string]interface{}
	if len(providerConfig.Props) > 0 {
		err := json.Unmarshal(providerConfig.Props, &props)
		if err != nil {
			logger.Error("解析TTS Props失败: %v", err)
			props = map[string]interface{}{}
		}
	} else {
		props = map[string]interface{}{}
	}

	return &ProviderFactory{
		providerType: "tts",
		config: &tts.Config{
			Type:  providerConfig.Type,
			Props: props,
		},
		logger: logger,
		params: map[string]interface{}{
			"type":         providerConfig.Type,
			"delete_audio": deleteAudio,
		},
		configService:    configService,
		grayscaleManager: grayscaleManager,
	}
}

func NewVLLLMFactory(vlllmType string, configService *database.ConfigService, logger *utils.Logger, grayscaleManager *GrayscaleManager) ResourceFactory {
	var providerConfig *database.ProviderConfig
	var err error

	if grayscaleManager != nil {
		providerConfig, err = grayscaleManager.GetProviderConfig("VLLLM", vlllmType)
	} else {
		providerConfig, err = configService.GetProviderConfigByCategoryAndName("VLLLM", vlllmType)
	}

	if err != nil {
		logger.Error("获取VLLLM配置失败: %v", err)
		return nil
	}

	var props map[string]interface{}
	if len(providerConfig.Props) > 0 {
		err := json.Unmarshal(providerConfig.Props, &props)
		if err != nil {
			logger.Error("解析VLLLM Props失败: %v", err)
			props = map[string]interface{}{}
		}
	} else {
		props = map[string]interface{}{}
	}

	// 补充type字段
	if _, ok := props["type"]; !ok || props["type"] == "" {
		props["type"] = providerConfig.Type
	}
	typeVal, _ := props["type"].(string)

	return &ProviderFactory{
		providerType: "vlllm",
		config: &vlllm.VLLLMConfig{
			Type:  typeVal,
			Extra: props,
		},
		logger:           logger,
		configService:    configService,
		grayscaleManager: grayscaleManager,
	}
}

func NewMCPFactory(config *configs.Config, logger *utils.Logger) ResourceFactory {
	return &ProviderFactory{
		providerType: "mcp",
		config:       config,
		logger:       logger,
		params:       map[string]interface{}{},
	}
}
