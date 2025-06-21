package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"ai-server-go/src/core/utils"
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

// GetGlobalConfig 获取全局配置
func (s *ConfigService) GetGlobalConfig(key string) (*GlobalConfig, error) {
	query := `SELECT * FROM global_configs WHERE config_key = ?`

	var config GlobalConfig
	err := s.db.QueryRow(query, key).Scan(
		&config.ID,
		&config.ConfigKey,
		&config.ConfigValue,
		&config.ConfigType,
		&config.Description,
		&config.IsSystem,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询全局配置失败: %v", err)
	}

	return &config, nil
}

// SetGlobalConfig 设置全局配置
func (s *ConfigService) SetGlobalConfig(key, value, configType, description string, isSystem bool) error {
	query := `
		INSERT INTO global_configs (config_key, config_value, config_type, description, is_system)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE 
		config_value = VALUES(config_value),
		config_type = VALUES(config_type),
		description = VALUES(description),
		is_system = VALUES(is_system)
	`

	_, err := s.db.Exec(query, key, value, configType, description, isSystem)
	if err != nil {
		return fmt.Errorf("设置全局配置失败: %v", err)
	}

	s.logger.Info("全局配置设置成功: %s = %s", key, value)
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
func (s *ConfigService) ListGlobalConfigs(isSystem *bool) ([]*GlobalConfig, error) {
	query := `SELECT * FROM global_configs`
	args := []interface{}{}

	if isSystem != nil {
		query += ` WHERE is_system = ?`
		args = append(args, *isSystem)
	}

	query += ` ORDER BY config_key`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询全局配置列表失败: %v", err)
	}
	defer rows.Close()

	var configs []*GlobalConfig
	for rows.Next() {
		var config GlobalConfig
		err := rows.Scan(
			&config.ID,
			&config.ConfigKey,
			&config.ConfigValue,
			&config.ConfigType,
			&config.Description,
			&config.IsSystem,
			&config.CreatedAt,
			&config.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描全局配置数据失败: %v", err)
		}
		configs = append(configs, &config)
	}

	return configs, nil
}

// GetAICapability 获取AI能力配置
func (s *ConfigService) GetAICapability(capabilityName, capabilityType string) (*AICapability, error) {
	query := `SELECT * FROM ai_capabilities WHERE capability_name = ? AND capability_type = ?`

	var capability AICapability
	err := s.db.QueryRow(query, capabilityName, capabilityType).Scan(
		&capability.ID,
		&capability.CapabilityName,
		&capability.CapabilityType,
		&capability.DisplayName,
		&capability.Description,
		&capability.ConfigSchema,
		&capability.IsGlobal,
		&capability.IsActive,
		&capability.CreatedAt,
		&capability.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询AI能力配置失败: %v", err)
	}

	return &capability, nil
}

// ListAICapabilities 获取AI能力配置列表
func (s *ConfigService) ListAICapabilities(capabilityType string, isActive *bool) ([]*AICapability, error) {
	query := `SELECT * FROM ai_capabilities`
	args := []interface{}{}

	conditions := []string{}
	if capabilityType != "" {
		conditions = append(conditions, "capability_type = ?")
		args = append(args, capabilityType)
	}
	if isActive != nil {
		conditions = append(conditions, "is_active = ?")
		args = append(args, *isActive)
	}

	if len(conditions) > 0 {
		query += ` WHERE ` + conditions[0]
		for i := 1; i < len(conditions); i++ {
			query += ` AND ` + conditions[i]
		}
	}

	query += ` ORDER BY capability_name, capability_type`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询AI能力配置列表失败: %v", err)
	}
	defer rows.Close()

	var capabilities []*AICapability
	for rows.Next() {
		var capability AICapability
		err := rows.Scan(
			&capability.ID,
			&capability.CapabilityName,
			&capability.CapabilityType,
			&capability.DisplayName,
			&capability.Description,
			&capability.ConfigSchema,
			&capability.IsGlobal,
			&capability.IsActive,
			&capability.CreatedAt,
			&capability.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描AI能力配置数据失败: %v", err)
		}
		capabilities = append(capabilities, &capability)
	}

	return capabilities, nil
}

// CreateAICapability 创建AI能力配置
func (s *ConfigService) CreateAICapability(capability *AICapability) error {
	query := `
		INSERT INTO ai_capabilities (capability_name, capability_type, display_name, description, config_schema, is_global, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	result, err := s.db.Exec(query,
		capability.CapabilityName,
		capability.CapabilityType,
		capability.DisplayName,
		capability.Description,
		capability.ConfigSchema,
		capability.IsGlobal,
		capability.IsActive,
	)

	if err != nil {
		return fmt.Errorf("创建AI能力配置失败: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取AI能力配置ID失败: %v", err)
	}

	capability.ID = id
	s.logger.Info("AI能力配置创建成功: %s/%s (ID: %d)", capability.CapabilityName, capability.CapabilityType, id)
	return nil
}

// UpdateAICapability 更新AI能力配置
func (s *ConfigService) UpdateAICapability(capability *AICapability) error {
	query := `
		UPDATE ai_capabilities 
		SET display_name = ?, description = ?, config_schema = ?, is_global = ?, is_active = ?
		WHERE id = ?
	`

	_, err := s.db.Exec(query,
		capability.DisplayName,
		capability.Description,
		capability.ConfigSchema,
		capability.IsGlobal,
		capability.IsActive,
		capability.ID,
	)

	if err != nil {
		return fmt.Errorf("更新AI能力配置失败: %v", err)
	}

	s.logger.Info("AI能力配置更新成功: %s/%s", capability.CapabilityName, capability.CapabilityType)
	return nil
}

// SetDeviceCapability 设置设备AI能力配置
func (s *ConfigService) SetDeviceCapability(deviceID int64, capabilityID int64, priority int, configData map[string]interface{}, isEnabled bool) error {
	configJSON, err := json.Marshal(configData)
	if err != nil {
		return fmt.Errorf("序列化配置数据失败: %v", err)
	}

	query := `
		INSERT INTO device_capabilities (device_id, capability_id, priority, config_data, is_enabled)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE 
		priority = VALUES(priority),
		config_data = VALUES(config_data),
		is_enabled = VALUES(is_enabled)
	`

	_, err = s.db.Exec(query, deviceID, capabilityID, priority, configJSON, isEnabled)
	if err != nil {
		return fmt.Errorf("设置设备AI能力配置失败: %v", err)
	}

	s.logger.Info("设备AI能力配置设置成功: DeviceID %d, CapabilityID %d", deviceID, capabilityID)
	return nil
}

// GetDeviceCapabilityConfig 获取设备AI能力配置
func (s *ConfigService) GetDeviceCapabilityConfig(deviceID string) (*DeviceCapabilityConfig, error) {
	// 获取设备信息
	deviceService := NewDeviceService(s.db, s.logger)
	deviceWithCapabilities, err := deviceService.GetDeviceWithCapabilities(deviceID)
	if err != nil {
		return nil, err
	}

	// 获取全局配置
	globalConfigs, err := s.ListGlobalConfigs(nil)
	if err != nil {
		return nil, err
	}

	// 构建配置
	config := &DeviceCapabilityConfig{
		DeviceID:      deviceWithCapabilities.Device.ID,
		Capabilities:  make([]CapabilityConfig, 0),
		GlobalConfigs: make(map[string]string),
	}

	// 处理设备AI能力配置
	for _, dc := range deviceWithCapabilities.Capabilities {
		var configData map[string]interface{}
		if dc.ConfigData != nil {
			if err := json.Unmarshal(dc.ConfigData, &configData); err != nil {
				s.logger.Error("解析设备AI能力配置失败: %v", err)
				continue
			}
		}

		capabilityConfig := CapabilityConfig{
			CapabilityName: dc.Capability.CapabilityName,
			CapabilityType: dc.Capability.CapabilityType,
			Config:         configData,
			Priority:       dc.Priority,
			IsEnabled:      dc.IsEnabled,
		}
		config.Capabilities = append(config.Capabilities, capabilityConfig)
	}

	// 处理全局配置
	for _, gc := range globalConfigs {
		config.GlobalConfigs[gc.ConfigKey] = gc.ConfigValue
	}

	return config, nil
}

// GetDefaultCapabilityType 获取默认能力类型
func (s *ConfigService) GetDefaultCapabilityType(capabilityName string) (string, error) {
	key := fmt.Sprintf("default.%s", capabilityName)
	return s.GetGlobalConfigValue(key)
}

// GetDeviceCapabilityByType 根据类型获取设备AI能力配置
func (s *ConfigService) GetDeviceCapabilityByType(deviceID int64, capabilityName string) (*DeviceCapability, error) {
	query := `
		SELECT dc.*, ac.capability_name, ac.capability_type, ac.display_name, ac.description, ac.config_schema
		FROM device_capabilities dc
		JOIN ai_capabilities ac ON dc.capability_id = ac.id
		WHERE dc.device_id = ? AND ac.capability_name = ? AND dc.is_enabled = true
		ORDER BY dc.priority ASC
		LIMIT 1
	`

	var dc DeviceCapability
	var ac AICapability
	err := s.db.QueryRow(query, deviceID, capabilityName).Scan(
		&dc.ID,
		&dc.DeviceID,
		&dc.CapabilityID,
		&dc.Priority,
		&dc.ConfigData,
		&dc.IsEnabled,
		&dc.CreatedAt,
		&dc.UpdatedAt,
		&ac.CapabilityName,
		&ac.CapabilityType,
		&ac.DisplayName,
		&ac.Description,
		&ac.ConfigSchema,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询设备AI能力配置失败: %v", err)
	}

	dc.Capability = &ac
	return &dc, nil
}

// GetDeviceCapabilities 获取设备AI能力配置列表
func (s *ConfigService) GetDeviceCapabilities(deviceID int64) ([]*DeviceCapability, error) {
	query := `
		SELECT dc.*, ac.capability_name, ac.capability_type, ac.display_name, ac.description, ac.config_schema
		FROM device_capabilities dc
		JOIN ai_capabilities ac ON dc.capability_id = ac.id
		WHERE dc.device_id = ?
		ORDER BY dc.priority ASC, ac.capability_name, ac.capability_type
	`

	rows, err := s.db.Query(query, deviceID)
	if err != nil {
		return nil, fmt.Errorf("查询设备AI能力配置失败: %v", err)
	}
	defer rows.Close()

	var capabilities []*DeviceCapability
	for rows.Next() {
		var dc DeviceCapability
		var ac AICapability
		err := rows.Scan(
			&dc.ID,
			&dc.DeviceID,
			&dc.CapabilityID,
			&dc.Priority,
			&dc.ConfigData,
			&dc.IsEnabled,
			&dc.CreatedAt,
			&dc.UpdatedAt,
			&ac.CapabilityName,
			&ac.CapabilityType,
			&ac.DisplayName,
			&ac.Description,
			&ac.ConfigSchema,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描设备AI能力配置数据失败: %v", err)
		}
		dc.Capability = &ac
		capabilities = append(capabilities, &dc)
	}

	return capabilities, nil
}

// GetAICapabilityByName 根据名称和类型获取AI能力配置
func (s *ConfigService) GetAICapabilityByName(capabilityName, capabilityType string) (*AICapability, error) {
	return s.GetAICapability(capabilityName, capabilityType)
}

// RemoveDeviceCapability 移除设备AI能力配置
func (s *ConfigService) RemoveDeviceCapability(deviceID int64, capabilityID int64) error {
	query := `DELETE FROM device_capabilities WHERE device_id = ? AND capability_id = ?`

	_, err := s.db.Exec(query, deviceID, capabilityID)
	if err != nil {
		return fmt.Errorf("移除设备AI能力配置失败: %v", err)
	}

	s.logger.Info("设备AI能力配置移除成功: DeviceID %d, CapabilityID %d", deviceID, capabilityID)
	return nil
}

// GetDeviceCapabilityWithFallback 获取设备AI能力配置（带回退逻辑）
// 优先级：设备专属配置 > 用户自定义配置 > 系统默认配置
func (s *ConfigService) GetDeviceCapabilityWithFallback(deviceID int64, userID *int64, capabilityName string) (*DeviceCapability, error) {
	// 1. 首先尝试获取设备专属配置
	deviceCapability, err := s.GetDeviceCapabilityByType(deviceID, capabilityName)
	if err != nil {
		return nil, fmt.Errorf("查询设备专属能力配置失败: %v", err)
	}
	if deviceCapability != nil && deviceCapability.IsEnabled {
		s.logger.Info("使用设备专属能力配置: DeviceID %d, Capability %s", deviceID, capabilityName)
		return deviceCapability, nil
	}

	// 2. 如果设备没有配置，尝试获取用户自定义配置
	if userID != nil {
		userCapability, err := s.GetUserCapabilityByType(*userID, capabilityName)
		if err != nil {
			s.logger.Error("查询用户自定义能力配置失败: %v", err)
		} else if userCapability != nil && userCapability.IsActive {
			// 将用户配置转换为设备配置格式
			deviceCap := &DeviceCapability{
				DeviceID:     deviceID,
				CapabilityID: userCapability.CapabilityID,
				Priority:     100, // 用户配置优先级较低
				ConfigData:   userCapability.ConfigData,
				IsEnabled:    true,
				Capability:   userCapability.Capability,
			}
			s.logger.Info("使用用户自定义能力配置: UserID %d, DeviceID %d, Capability %s", *userID, deviceID, capabilityName)
			return deviceCap, nil
		}
	}

	// 3. 最后尝试获取系统默认配置
	defaultCapability, err := s.GetDefaultCapabilityByName(capabilityName)
	if err != nil {
		s.logger.Error("查询系统默认能力配置失败: %v", err)
	} else if defaultCapability != nil && defaultCapability.IsActive {
		// 将默认配置转换为设备配置格式
		deviceCap := &DeviceCapability{
			DeviceID:     deviceID,
			CapabilityID: defaultCapability.ID,
			Priority:     200, // 默认配置优先级最低
			ConfigData:   nil, // 使用默认配置
			IsEnabled:    true,
			Capability:   defaultCapability,
		}
		s.logger.Info("使用系统默认能力配置: DeviceID %d, Capability %s", deviceID, capabilityName)
		return deviceCap, nil
	}

	// 4. 所有配置都没有找到
	s.logger.Warn("未找到任何能力配置: DeviceID %d, UserID %v, Capability %s", deviceID, userID, capabilityName)
	return nil, nil
}

// GetUserCapabilityByType 根据类型获取用户AI能力配置
func (s *ConfigService) GetUserCapabilityByType(userID int64, capabilityName string) (*UserCapability, error) {
	query := `
		SELECT uc.*, ac.capability_name, ac.capability_type, ac.display_name, ac.description, ac.config_schema
		FROM user_capabilities uc
		JOIN ai_capabilities ac ON uc.capability_id = ac.id
		WHERE uc.user_id = ? AND ac.capability_name = ? AND uc.is_active = true
		ORDER BY uc.created_at DESC
		LIMIT 1
	`

	var uc UserCapability
	var ac AICapability
	err := s.db.QueryRow(query, userID, capabilityName).Scan(
		&uc.ID,
		&uc.UserID,
		&uc.CapabilityID,
		&uc.ConfigData,
		&uc.IsActive,
		&uc.CreatedAt,
		&uc.UpdatedAt,
		&ac.CapabilityName,
		&ac.CapabilityType,
		&ac.DisplayName,
		&ac.Description,
		&ac.ConfigSchema,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询用户AI能力配置失败: %v", err)
	}

	uc.Capability = &ac
	return &uc, nil
}

// GetDefaultCapabilityByName 根据名称获取系统默认AI能力配置
func (s *ConfigService) GetDefaultCapabilityByName(capabilityName string) (*AICapability, error) {
	// 首先获取默认能力类型
	defaultType, err := s.GetDefaultCapabilityType(capabilityName)
	if err != nil {
		return nil, fmt.Errorf("获取默认能力类型失败: %v", err)
	}
	if defaultType == "" {
		return nil, nil
	}

	// 然后获取该类型的默认配置
	query := `
		SELECT * FROM ai_capabilities 
		WHERE capability_name = ? AND capability_type = ? AND is_global = true AND is_active = true
		ORDER BY created_at ASC
		LIMIT 1
	`

	var capability AICapability
	err = s.db.QueryRow(query, capabilityName, defaultType).Scan(
		&capability.ID,
		&capability.CapabilityName,
		&capability.CapabilityType,
		&capability.DisplayName,
		&capability.Description,
		&capability.ConfigSchema,
		&capability.IsGlobal,
		&capability.IsActive,
		&capability.CreatedAt,
		&capability.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询默认AI能力配置失败: %v", err)
	}

	return &capability, nil
}

// GetDeviceCapabilityConfigWithFallback 获取设备AI能力配置（带回退逻辑）
func (s *ConfigService) GetDeviceCapabilityConfigWithFallback(deviceID string, userID *int64) (*DeviceCapabilityConfig, error) {
	// 获取设备信息
	deviceService := NewDeviceService(s.db, s.logger)
	device, err := deviceService.GetDeviceByUUID(deviceID)
	if err != nil {
		return nil, err
	}
	if device == nil {
		return nil, fmt.Errorf("设备不存在: %s", deviceID)
	}

	// 获取所有可用的AI能力类型
	capabilities, err := s.ListAICapabilities("", nil)
	if err != nil {
		return nil, err
	}

	// 构建配置
	config := &DeviceCapabilityConfig{
		DeviceID:      device.ID,
		Capabilities:  make([]CapabilityConfig, 0),
		GlobalConfigs: make(map[string]string),
	}

	// 为每种能力类型获取配置（带回退逻辑）
	for _, capability := range capabilities {
		if !capability.IsActive {
			continue
		}

		deviceCap, err := s.GetDeviceCapabilityWithFallback(device.ID, userID, capability.CapabilityName)
		if err != nil {
			s.logger.Error("获取设备能力配置失败: %v", err)
			continue
		}

		if deviceCap != nil {
			var configData map[string]interface{}
			if deviceCap.ConfigData != nil {
				if err := json.Unmarshal(deviceCap.ConfigData, &configData); err != nil {
					s.logger.Error("解析设备能力配置失败: %v", err)
					continue
				}
			}

			capabilityConfig := CapabilityConfig{
				CapabilityName: deviceCap.Capability.CapabilityName,
				CapabilityType: deviceCap.Capability.CapabilityType,
				Config:         configData,
				Priority:       deviceCap.Priority,
				IsEnabled:      deviceCap.IsEnabled,
			}
			config.Capabilities = append(config.Capabilities, capabilityConfig)
		}
	}

	// 获取全局配置
	globalConfigs, err := s.ListGlobalConfigs(nil)
	if err != nil {
		return nil, err
	}

	// 处理全局配置
	for _, gc := range globalConfigs {
		config.GlobalConfigs[gc.ConfigKey] = gc.ConfigValue
	}

	return config, nil
}

// DeleteAICapability 删除AI能力配置
func (s *ConfigService) DeleteAICapability(capabilityName, capabilityType string) error {
	// 先获取AI能力
	capability, err := s.GetAICapability(capabilityName, capabilityType)
	if err != nil {
		return fmt.Errorf("获取AI能力失败: %v", err)
	}
	if capability == nil {
		return fmt.Errorf("AI能力不存在: %s/%s", capabilityName, capabilityType)
	}

	// 检查是否有关联的设备或用户配置
	deviceCount, err := s.getCapabilityUsageCount(capability.ID, "device")
	if err != nil {
		return fmt.Errorf("检查设备使用情况失败: %v", err)
	}

	userCount, err := s.getCapabilityUsageCount(capability.ID, "user")
	if err != nil {
		return fmt.Errorf("检查用户使用情况失败: %v", err)
	}

	if deviceCount > 0 || userCount > 0 {
		return fmt.Errorf("AI能力正在被使用，无法删除: 设备使用 %d, 用户使用 %d", deviceCount, userCount)
	}

	// 软删除：设置为非活跃状态
	query := `UPDATE ai_capabilities SET is_active = false WHERE id = ?`
	_, err = s.db.Exec(query, capability.ID)
	if err != nil {
		return fmt.Errorf("删除AI能力失败: %v", err)
	}

	s.logger.Info("AI能力删除成功: %s/%s", capabilityName, capabilityType)
	return nil
}

// getCapabilityUsageCount 获取能力使用次数
func (s *ConfigService) getCapabilityUsageCount(capabilityID int64, usageType string) (int, error) {
	var query string
	switch usageType {
	case "device":
		query = `SELECT COUNT(*) FROM device_capabilities WHERE capability_id = ? AND is_enabled = true`
	case "user":
		query = `SELECT COUNT(*) FROM user_capabilities WHERE capability_id = ? AND is_active = true`
	default:
		return 0, fmt.Errorf("未知的使用类型: %s", usageType)
	}

	var count int
	err := s.db.QueryRow(query, capabilityID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// ListDefaultAICapabilities 获取默认AI能力列表
func (s *ConfigService) ListDefaultAICapabilities() (map[string]string, error) {
	query := `SELECT config_key, config_value FROM global_configs WHERE config_key LIKE 'default.%'`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询默认AI能力失败: %v", err)
	}
	defer rows.Close()

	defaults := make(map[string]string)
	for rows.Next() {
		var key, value string
		err := rows.Scan(&key, &value)
		if err != nil {
			return nil, fmt.Errorf("扫描默认AI能力数据失败: %v", err)
		}
		// 去掉 "default." 前缀
		capabilityName := strings.TrimPrefix(key, "default.")
		defaults[capabilityName] = value
	}

	return defaults, nil
}

// SetDefaultAICapability 设置默认AI能力
func (s *ConfigService) SetDefaultAICapability(capabilityName, capabilityType string) error {
	// 验证AI能力是否存在
	capability, err := s.GetAICapability(capabilityName, capabilityType)
	if err != nil {
		return fmt.Errorf("验证AI能力失败: %v", err)
	}
	if capability == nil {
		return fmt.Errorf("AI能力不存在: %s/%s", capabilityName, capabilityType)
	}

	// 设置全局配置
	configKey := fmt.Sprintf("default.%s", capabilityName)
	configValue := capabilityType

	query := `
		INSERT INTO global_configs (config_key, config_value, config_type, description, is_system)
		VALUES (?, ?, 'string', ?, true)
		ON DUPLICATE KEY UPDATE 
		config_value = VALUES(config_value),
		description = VALUES(description)
	`

	description := fmt.Sprintf("默认%s能力类型", capabilityName)
	_, err = s.db.Exec(query, configKey, configValue, description)
	if err != nil {
		return fmt.Errorf("设置默认AI能力失败: %v", err)
	}

	s.logger.Info("默认AI能力设置成功: %s -> %s", capabilityName, capabilityType)
	return nil
}

// RemoveDefaultAICapability 移除默认AI能力
func (s *ConfigService) RemoveDefaultAICapability(capabilityName string) error {
	query := `DELETE FROM global_configs WHERE config_key = ?`

	_, err := s.db.Exec(query, "default_"+capabilityName)
	if err != nil {
		return fmt.Errorf("移除默认AI能力失败: %v", err)
	}

	s.logger.Info("默认AI能力移除成功: %s", capabilityName)
	return nil
}

// ==================== 系统配置管理 ====================

// GetSystemConfig 获取系统配置
func (s *ConfigService) GetSystemConfig(category, key string) (*SystemConfig, error) {
	query := `SELECT * FROM system_configs WHERE config_category = ? AND config_key = ?`

	var config SystemConfig
	err := s.db.QueryRow(query, category, key).Scan(
		&config.ID,
		&config.ConfigCategory,
		&config.ConfigKey,
		&config.ConfigValue,
		&config.ConfigType,
		&config.Description,
		&config.IsDefault,
		&config.CreatedBy,
		&config.UpdatedBy,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询系统配置失败: %v", err)
	}

	return &config, nil
}

// SetSystemConfig 设置系统配置
func (s *ConfigService) SetSystemConfig(category, key, value, configType, description string, isDefault bool, userID *int64) error {
	query := `
		INSERT INTO system_configs (config_category, config_key, config_value, config_type, description, is_default, created_by, updated_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE 
		config_value = VALUES(config_value),
		config_type = VALUES(config_type),
		description = VALUES(description),
		is_default = VALUES(is_default),
		updated_by = VALUES(updated_by)
	`

	_, err := s.db.Exec(query, category, key, value, configType, description, isDefault, userID, userID)
	if err != nil {
		return fmt.Errorf("设置系统配置失败: %v", err)
	}

	s.logger.Info("系统配置设置成功: %s.%s = %s", category, key, value)
	return nil
}

// ListSystemConfigs 获取系统配置列表
func (s *ConfigService) ListSystemConfigs(category string, isDefault *bool) ([]*SystemConfig, error) {
	query := `SELECT * FROM system_configs`
	args := []interface{}{}

	conditions := []string{}
	if category != "" {
		conditions = append(conditions, "config_category = ?")
		args = append(args, category)
	}
	if isDefault != nil {
		conditions = append(conditions, "is_default = ?")
		args = append(args, *isDefault)
	}

	if len(conditions) > 0 {
		query += ` WHERE ` + conditions[0]
		for i := 1; i < len(conditions); i++ {
			query += ` AND ` + conditions[i]
		}
	}

	query += ` ORDER BY config_category, config_key`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询系统配置列表失败: %v", err)
	}
	defer rows.Close()

	var configs []*SystemConfig
	for rows.Next() {
		var config SystemConfig
		err := rows.Scan(
			&config.ID,
			&config.ConfigCategory,
			&config.ConfigKey,
			&config.ConfigValue,
			&config.ConfigType,
			&config.Description,
			&config.IsDefault,
			&config.CreatedBy,
			&config.UpdatedBy,
			&config.CreatedAt,
			&config.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描系统配置数据失败: %v", err)
		}
		configs = append(configs, &config)
	}

	return configs, nil
}

// DeleteSystemConfig 删除系统配置
func (s *ConfigService) DeleteSystemConfig(category, key string) error {
	query := `DELETE FROM system_configs WHERE config_category = ? AND config_key = ?`

	_, err := s.db.Exec(query, category, key)
	if err != nil {
		return fmt.Errorf("删除系统配置失败: %v", err)
	}

	s.logger.Info("系统配置删除成功: %s.%s", category, key)
	return nil
}

// GetSystemConfigValue 获取系统配置值
func (s *ConfigService) GetSystemConfigValue(category, key string) (string, error) {
	config, err := s.GetSystemConfig(category, key)
	if err != nil {
		return "", err
	}
	if config == nil {
		return "", fmt.Errorf("系统配置不存在: %s.%s", category, key)
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
			err = s.SetSystemConfig(config.category, config.key, config.value, config.configType, config.description, true, nil)
			if err != nil {
				return fmt.Errorf("初始化系统配置失败: %v", err)
			}
		}
	}

	s.logger.Info("默认系统配置初始化完成")
	return nil
}

// GetSystemConfigCategory 获取指定分类的所有系统配置
func (s *ConfigService) GetSystemConfigCategory(category string) (map[string]interface{}, error) {
	configs, err := s.ListSystemConfigs(category, nil)
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

// ==================== Provider配置管理 ====================

// SetProviderConfig 设置/更新AI Provider配置
func (s *ConfigService) SetProviderConfig(category, name string, config *ProviderConfig) error {
	valueBytes, err := json.Marshal(config)
	if err != nil {
		return err
	}
	query := `REPLACE INTO system_configs (config_category, config_key, config_value, config_type, description, is_default, created_by, updated_by) VALUES (?, ?, ?, 'json', ?, 0, 'system', 'system')`
	_, err = s.db.Exec(query, category, name, string(valueBytes), "AI Provider配置")
	return err
}

// GetProviderConfig 获取AI Provider配置
func (s *ConfigService) GetProviderConfig(category, name string) (*ProviderConfig, error) {
	query := `SELECT config_value FROM system_configs WHERE config_category = ? AND config_key = ?`
	row := s.db.QueryRow(query, category, name)
	var value string
	if err := row.Scan(&value); err != nil {
		return nil, err
	}
	var config ProviderConfig
	if err := json.Unmarshal([]byte(value), &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// ListProviderConfigs 列出某类所有Provider配置
func (s *ConfigService) ListProviderConfigs(category string) ([]ProviderConfig, error) {
	query := `SELECT config_value FROM system_configs WHERE config_category = ?`
	rows, err := s.db.Query(query, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var configs []ProviderConfig
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		var config ProviderConfig
		if err := json.Unmarshal([]byte(value), &config); err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}
	return configs, nil
}

// DeleteProviderConfig 删除Provider配置
func (s *ConfigService) DeleteProviderConfig(category, name string) error {
	query := `DELETE FROM system_configs WHERE config_category = ? AND config_key = ?`
	_, err := s.db.Exec(query, category, name)
	return err
}

// GetProviderConfigWithVersion 获取指定版本的provider配置
func (s *ConfigService) GetProviderConfigWithVersion(category, name, version string) (*ProviderConfig, error) {
	query := `
		SELECT * FROM provider_configs 
		WHERE category = ? AND name = ? AND version = ? AND is_active = true
	`

	var config ProviderConfig
	err := s.db.QueryRow(query, category, name, version).Scan(
		&config.ID, &config.Category, &config.Name, &config.Version, &config.Weight,
		&config.IsActive, &config.IsDefault, &config.Type, &config.ModelName,
		&config.BaseURL, &config.APIKey, &config.Voice, &config.Format,
		&config.OutputDir, &config.AppID, &config.Token, &config.Cluster,
		&config.Temperature, &config.MaxTokens, &config.TopP, &config.Security,
		&config.Extra, &config.HealthScore, &config.CreatedAt, &config.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询provider配置失败: %v", err)
	}

	return &config, nil
}

// ListProviderVersions 获取指定provider的所有版本
func (s *ConfigService) ListProviderVersions(category, name string) ([]*ProviderVersion, error) {
	query := `
		SELECT category, name, version, weight, is_active, is_default, 
		       health_score, created_at, updated_at
		FROM provider_configs 
		WHERE category = ? AND name = ?
		ORDER BY version ASC, created_at DESC
	`

	rows, err := s.db.Query(query, category, name)
	if err != nil {
		return nil, fmt.Errorf("查询provider版本列表失败: %v", err)
	}
	defer rows.Close()

	var versions []*ProviderVersion
	for rows.Next() {
		var version ProviderVersion
		err := rows.Scan(
			&version.Category, &version.Name, &version.Version, &version.Weight,
			&version.IsActive, &version.IsDefault, &version.HealthScore,
			&version.CreatedAt, &version.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描provider版本数据失败: %v", err)
		}
		versions = append(versions, &version)
	}

	return versions, nil
}

// GetActiveProviderConfigs 获取指定provider的所有活跃配置（用于灰度发布）
func (s *ConfigService) GetActiveProviderConfigs(category, name string) ([]*ProviderConfig, error) {
	query := `
		SELECT * FROM provider_configs 
		WHERE category = ? AND name = ? AND is_active = true
		ORDER BY weight DESC, version ASC
	`

	rows, err := s.db.Query(query, category, name)
	if err != nil {
		return nil, fmt.Errorf("查询活跃provider配置失败: %v", err)
	}
	defer rows.Close()

	var configs []*ProviderConfig
	for rows.Next() {
		var config ProviderConfig
		err := rows.Scan(
			&config.ID, &config.Category, &config.Name, &config.Version, &config.Weight,
			&config.IsActive, &config.IsDefault, &config.Type, &config.ModelName,
			&config.BaseURL, &config.APIKey, &config.Voice, &config.Format,
			&config.OutputDir, &config.AppID, &config.Token, &config.Cluster,
			&config.Temperature, &config.MaxTokens, &config.TopP, &config.Security,
			&config.Extra, &config.HealthScore, &config.CreatedAt, &config.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描provider配置数据失败: %v", err)
		}
		configs = append(configs, &config)
	}

	return configs, nil
}

// UpdateProviderWeight 更新provider版本的流量权重
func (s *ConfigService) UpdateProviderWeight(category, name, version string, weight int) error {
	query := `UPDATE provider_configs SET weight = ?, updated_at = CURRENT_TIMESTAMP WHERE category = ? AND name = ? AND version = ?`

	_, err := s.db.Exec(query, weight, category, name, version)
	if err != nil {
		return fmt.Errorf("更新provider权重失败: %v", err)
	}

	s.logger.Info("Provider权重更新成功: %s/%s/%s -> %d", category, name, version, weight)
	return nil
}

// UpdateProviderHealthScore 更新provider版本的健康评分
func (s *ConfigService) UpdateProviderHealthScore(category, name, version string, healthScore float64) error {
	query := `UPDATE provider_configs SET health_score = ?, updated_at = CURRENT_TIMESTAMP WHERE category = ? AND name = ? AND version = ?`

	_, err := s.db.Exec(query, healthScore, category, name, version)
	if err != nil {
		return fmt.Errorf("更新provider健康评分失败: %v", err)
	}

	s.logger.Info("Provider健康评分更新成功: %s/%s/%s -> %.2f", category, name, version, healthScore)
	return nil
}

// SetDefaultProviderVersion 设置默认provider版本
func (s *ConfigService) SetDefaultProviderVersion(category, name, version string) error {
	// 先取消所有版本的默认状态
	query1 := `UPDATE provider_configs SET is_default = FALSE WHERE category = ? AND name = ?`
	_, err := s.db.Exec(query1, category, name)
	if err != nil {
		return fmt.Errorf("取消默认版本失败: %v", err)
	}

	// 设置指定版本为默认
	query2 := `UPDATE provider_configs SET is_default = TRUE WHERE category = ? AND name = ? AND version = ?`
	_, err = s.db.Exec(query2, category, name, version)
	if err != nil {
		return fmt.Errorf("设置默认版本失败: %v", err)
	}

	s.logger.Info("默认provider版本设置成功: %s/%s/%s", category, name, version)
	return nil
}

// GetDefaultProviderModules 返回所有类别的默认provider name
func (s *ConfigService) GetDefaultProviderModules() (map[string]string, error) {
	query := `SELECT category, name FROM provider_configs WHERE is_default = true AND is_active = true`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make(map[string]string)
	for rows.Next() {
		var category, name string
		if err := rows.Scan(&category, &name); err != nil {
			return nil, err
		}
		result[category] = name
	}
	return result, nil
}
