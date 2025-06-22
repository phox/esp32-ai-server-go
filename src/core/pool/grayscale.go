package pool

import (
	"ai-server-go/src/core/utils"
	"ai-server-go/src/database"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// GrayscaleManager 灰度发布管理器
type GrayscaleManager struct {
	configService *database.ConfigService
	logger        *utils.Logger
	mu            sync.RWMutex
	cache         map[string]*GrayscaleConfig // key: category/name
	healthChecker *HealthChecker
}

// GrayscaleConfig 灰度发布配置
type GrayscaleConfig struct {
	Category string              `json:"category"`
	Name     string              `json:"name"`
	Versions []*GrayscaleVersion `json:"versions"`
	Strategy string              `json:"strategy"` // "weight", "health", "round_robin"
	mu       sync.RWMutex
}

// GrayscaleVersion 灰度版本信息
type GrayscaleVersion struct {
	Version     string                   `json:"version"`
	Weight      int                      `json:"weight"`
	IsActive    bool                     `json:"is_active"`
	IsDefault   bool                     `json:"is_default"`
	HealthScore float64                  `json:"health_score"`
	Config      *database.ProviderConfig `json:"config"`
}

// NewGrayscaleManager 创建灰度发布管理器
func NewGrayscaleManager(configService *database.ConfigService, logger *utils.Logger) *GrayscaleManager {
	gm := &GrayscaleManager{
		configService: configService,
		logger:        logger,
		cache:         make(map[string]*GrayscaleConfig),
	}

	// 启动健康检查协程
	go gm.startHealthCheck()

	return gm
}

// GetProviderConfig 根据灰度策略获取provider配置
func (gm *GrayscaleManager) GetProviderConfig(category, name string) (*database.ProviderConfig, error) {
	gm.mu.RLock()
	config, exists := gm.cache[fmt.Sprintf("%s/%s", category, name)]
	gm.mu.RUnlock()

	if !exists {
		// 缓存未命中，从数据库加载
		if err := gm.loadGrayscaleConfig(category, name); err != nil {
			return nil, err
		}
		gm.mu.RLock()
		config = gm.cache[fmt.Sprintf("%s/%s", category, name)]
		gm.mu.RUnlock()
	}

	if config == nil || len(config.Versions) == 0 {
		return nil, fmt.Errorf("没有可用的provider配置: %s/%s", category, name)
	}

	// 根据策略选择版本
	var selectedVersion *GrayscaleVersion
	switch config.Strategy {
	case "weight":
		selectedVersion = gm.selectByWeight(config)
	case "health":
		selectedVersion = gm.selectByHealth(config)
	case "round_robin":
		selectedVersion = gm.selectByRoundRobin(config)
	default:
		selectedVersion = gm.selectByWeight(config) // 默认使用权重策略
	}

	if selectedVersion == nil {
		return nil, fmt.Errorf("无法选择合适的provider版本: %s/%s", category, name)
	}

	return selectedVersion.Config, nil
}

// selectByWeight 根据权重选择版本
func (gm *GrayscaleManager) selectByWeight(config *GrayscaleConfig) *GrayscaleVersion {
	config.mu.RLock()
	defer config.mu.RUnlock()

	// 计算总权重
	totalWeight := 0
	activeVersions := make([]*GrayscaleVersion, 0)

	for _, version := range config.Versions {
		if version.IsActive {
			totalWeight += version.Weight
			activeVersions = append(activeVersions, version)
		}
	}

	if totalWeight == 0 || len(activeVersions) == 0 {
		return nil
	}

	// 随机选择
	rand.Seed(time.Now().UnixNano())
	r := rand.Intn(totalWeight)

	currentWeight := 0
	for _, version := range activeVersions {
		currentWeight += version.Weight
		if r < currentWeight {
			return version
		}
	}

	// 兜底返回第一个活跃版本
	return activeVersions[0]
}

// selectByHealth 根据健康评分选择版本
func (gm *GrayscaleManager) selectByHealth(config *GrayscaleConfig) *GrayscaleVersion {
	config.mu.RLock()
	defer config.mu.RUnlock()

	var bestVersion *GrayscaleVersion
	bestScore := -1.0

	for _, version := range config.Versions {
		if version.IsActive && version.HealthScore > bestScore {
			bestScore = version.HealthScore
			bestVersion = version
		}
	}

	return bestVersion
}

// selectByRoundRobin 轮询选择版本
func (gm *GrayscaleManager) selectByRoundRobin(config *GrayscaleConfig) *GrayscaleVersion {
	config.mu.Lock()
	defer config.mu.Unlock()

	activeVersions := make([]*GrayscaleVersion, 0)
	for _, version := range config.Versions {
		if version.IsActive {
			activeVersions = append(activeVersions, version)
		}
	}

	if len(activeVersions) == 0 {
		return nil
	}

	// 简单的轮询实现
	// 这里可以使用更复杂的轮询算法
	return activeVersions[time.Now().Unix()%int64(len(activeVersions))]
}

// loadGrayscaleConfig 从数据库加载灰度配置
func (gm *GrayscaleManager) loadGrayscaleConfig(category, name string) error {
	configs, err := gm.configService.GetActiveProviderConfigs(category, name)
	if err != nil {
		return fmt.Errorf("加载灰度配置失败: %v", err)
	}

	grayscaleConfig := &GrayscaleConfig{
		Category: category,
		Name:     name,
		Strategy: "weight", // 默认使用权重策略
		Versions: make([]*GrayscaleVersion, 0),
	}

	for _, config := range configs {
		version := &GrayscaleVersion{
			Version:     config.Version,
			Weight:      config.Weight,
			IsActive:    config.IsActive,
			IsDefault:   config.IsDefault,
			HealthScore: config.HealthScore,
			Config:      config,
		}
		grayscaleConfig.Versions = append(grayscaleConfig.Versions, version)
	}

	gm.mu.Lock()
	gm.cache[fmt.Sprintf("%s/%s", category, name)] = grayscaleConfig
	gm.mu.Unlock()

	return nil
}

// RefreshConfig 刷新指定provider的灰度配置
func (gm *GrayscaleManager) RefreshConfig(category, name string) error {
	gm.mu.Lock()
	delete(gm.cache, fmt.Sprintf("%s/%s", category, name))
	gm.mu.Unlock()

	return gm.loadGrayscaleConfig(category, name)
}

// UpdateWeight 更新版本权重
func (gm *GrayscaleManager) UpdateWeight(category, name, version string, weight int) error {
	err := gm.configService.UpdateProviderWeight(category, name, version, weight)
	if err != nil {
		return err
	}

	// 刷新缓存
	return gm.RefreshConfig(category, name)
}

// UpdateHealthScore 更新健康评分
func (gm *GrayscaleManager) UpdateHealthScore(category, name, version string, healthScore float64) error {
	err := gm.configService.UpdateProviderHealthScoreByKey(category, name, version, healthScore)
	if err != nil {
		return err
	}

	// 刷新缓存
	return gm.RefreshConfig(category, name)
}

// startHealthCheck 启动健康检查协程
func (gm *GrayscaleManager) startHealthCheck() {
	ticker := time.NewTicker(30 * time.Second) // 每30秒检查一次
	defer ticker.Stop()

	for range ticker.C {
		gm.performHealthCheck()
	}
}

// performHealthCheck 执行健康检查
func (gm *GrayscaleManager) performHealthCheck() {
	gm.mu.RLock()
	configs := make([]*GrayscaleConfig, 0, len(gm.cache))
	for _, config := range gm.cache {
		configs = append(configs, config)
	}
	gm.mu.RUnlock()

	for _, config := range configs {
		for _, version := range config.Versions {
			if version.IsActive {
				// 这里可以调用实际的健康检查逻辑
				// 暂时使用模拟的健康评分
				healthScore := gm.simulateHealthCheck(version.Config)
				if healthScore != version.HealthScore {
					_ = gm.UpdateHealthScore(config.Category, config.Name, version.Version, healthScore)
				}
			}
		}
	}
}

// simulateHealthCheck 模拟健康检查（实际项目中应该调用真实的健康检查）
func (gm *GrayscaleManager) simulateHealthCheck(config *database.ProviderConfig) float64 {
	// 这里应该实现真实的健康检查逻辑
	// 暂时返回一个随机值作为示例
	rand.Seed(time.Now().UnixNano())
	return 80 + rand.Float64()*20 // 80-100之间的随机值
}

// GetGrayscaleStatus 获取灰度发布状态
func (gm *GrayscaleManager) GetGrayscaleStatus(category, name string) (*GrayscaleConfig, error) {
	gm.mu.RLock()
	config, exists := gm.cache[fmt.Sprintf("%s/%s", category, name)]
	gm.mu.RUnlock()

	if !exists {
		if err := gm.loadGrayscaleConfig(category, name); err != nil {
			return nil, err
		}
		gm.mu.RLock()
		config = gm.cache[fmt.Sprintf("%s/%s", category, name)]
		gm.mu.RUnlock()
	}

	return config, nil
}
