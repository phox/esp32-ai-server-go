package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

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
