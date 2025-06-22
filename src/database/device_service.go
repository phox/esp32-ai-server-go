package database

import (
	"encoding/json"
	"fmt"
	"time"

	"ai-server-go/src/core/utils"

	"gorm.io/gorm"
)

// DeviceService 设备管理服务
type DeviceService struct {
	db     *Database
	logger *utils.Logger
}

// NewDeviceService 创建设备管理服务
func NewDeviceService(db *Database, logger *utils.Logger) *DeviceService {
	return &DeviceService{
		db:     db,
		logger: logger,
	}
}

// GetDB 获取数据库连接
func (s *DeviceService) GetDB() *Database {
	return s.db
}

// CreateDevice 创建设备
func (s *DeviceService) CreateDevice(device *Device) error {
	device.Status = "offline"
	if err := s.db.DB.Create(device).Error; err != nil {
		return fmt.Errorf("创建设备失败: %v", err)
	}

	s.logger.Info("设备创建成功: %s (UUID: %s)", device.DeviceName, device.DeviceUUID)
	return nil
}

// GetDeviceByID 根据ID获取设备
func (s *DeviceService) GetDeviceByID(id uint) (*Device, error) {
	var device Device
	if err := s.db.DB.First(&device, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询设备失败: %v", err)
	}
	return &device, nil
}

// GetDeviceByUUID 根据UUID获取设备
func (s *DeviceService) GetDeviceByUUID(deviceUUID string) (*Device, error) {
	var device Device
	if err := s.db.DB.Where("device_uuid = ?", deviceUUID).First(&device).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询设备失败: %v", err)
	}
	return &device, nil
}

// UpdateDevice 更新设备
func (s *DeviceService) UpdateDevice(device *Device) error {
	if err := s.db.DB.Save(device).Error; err != nil {
		return fmt.Errorf("更新设备失败: %v", err)
	}

	s.logger.Info("设备更新成功: %s", device.DeviceName)
	return nil
}

// DeleteDevice 删除设备
func (s *DeviceService) DeleteDevice(id uint) error {
	if err := s.db.DB.Delete(&Device{}, id).Error; err != nil {
		return fmt.Errorf("删除设备失败: %v", err)
	}

	s.logger.Info("设备删除成功: ID %d", id)
	return nil
}

// ListDevices 获取设备列表
func (s *DeviceService) ListDevices(offset, limit int, status, deviceType string) ([]*Device, error) {
	var devices []*Device
	query := s.db.DB

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if deviceType != "" {
		query = query.Where("device_type = ?", deviceType)
	}

	if err := query.Offset(offset).Limit(limit).Find(&devices).Error; err != nil {
		return nil, fmt.Errorf("查询设备列表失败: %v", err)
	}

	return devices, nil
}

// CountDevices 统计设备数量
func (s *DeviceService) CountDevices(status, deviceType string) (int64, error) {
	var count int64
	query := s.db.DB.Model(&Device{})

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if deviceType != "" {
		query = query.Where("device_type = ?", deviceType)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("统计设备数量失败: %v", err)
	}

	return count, nil
}

// UpdateDeviceStatus 更新设备状态
func (s *DeviceService) UpdateDeviceStatus(deviceUUID string, status string, ipAddress string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if ipAddress != "" {
		updates["last_ip_address"] = ipAddress
	}
	if status == "online" {
		now := time.Now()
		updates["last_online_time"] = &now
	}

	if err := s.db.DB.Model(&Device{}).Where("device_uuid = ?", deviceUUID).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新设备状态失败: %v", err)
	}

	s.logger.Info("设备状态更新成功: UUID %s, 状态 %s", deviceUUID, status)
	return nil
}

// GetDeviceCapabilities 获取设备AI能力列表
func (s *DeviceService) GetDeviceCapabilities(deviceID uint) ([]*DeviceCapability, error) {
	var deviceCapabilities []*DeviceCapability
	if err := s.db.DB.Where("device_id = ? AND is_enabled = ?", deviceID, true).
		Preload("Capability").
		Order("priority DESC").
		Find(&deviceCapabilities).Error; err != nil {
		return nil, fmt.Errorf("查询设备AI能力失败: %v", err)
	}
	return deviceCapabilities, nil
}

// SetDeviceCapability 设置设备AI能力
func (s *DeviceService) SetDeviceCapability(deviceID uint, capabilityName, capabilityType string, priority int, config map[string]interface{}, isEnabled bool) error {
	// 查找AI能力
	var capability AICapability
	if err := s.db.DB.Where("capability_name = ? AND capability_type = ?", capabilityName, capabilityType).First(&capability).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("AI能力不存在")
		}
		return fmt.Errorf("查询AI能力失败: %v", err)
	}

	// 序列化配置
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	// 查找现有配置
	var deviceCapability DeviceCapability
	if err := s.db.DB.Where("device_id = ? AND capability_id = ?", deviceID, capability.ID).First(&deviceCapability).Error; err == nil {
		// 更新现有配置
		deviceCapability.Priority = priority
		deviceCapability.ConfigData = configJSON
		deviceCapability.IsEnabled = isEnabled
		if err := s.db.DB.Save(&deviceCapability).Error; err != nil {
			return fmt.Errorf("更新设备AI能力失败: %v", err)
		}
	} else {
		// 创建新配置
		deviceCapability = DeviceCapability{
			DeviceID:     deviceID,
			CapabilityID: capability.ID,
			Priority:     priority,
			ConfigData:   configJSON,
			IsEnabled:    isEnabled,
		}
		if err := s.db.DB.Create(&deviceCapability).Error; err != nil {
			return fmt.Errorf("创建设备AI能力失败: %v", err)
		}
	}

	s.logger.Info("设备AI能力设置成功: 设备ID %d, 能力 %s", deviceID, capabilityName)
	return nil
}

// GetDeviceWithCapabilities 获取设备及其AI能力
func (s *DeviceService) GetDeviceWithCapabilities(deviceID uint) (*DeviceWithCapabilities, error) {
	device, err := s.GetDeviceByID(deviceID)
	if err != nil {
		return nil, err
	}
	if device == nil {
		return nil, fmt.Errorf("设备不存在")
	}

	capabilities, err := s.GetDeviceCapabilities(deviceID)
	if err != nil {
		return nil, err
	}

	return &DeviceWithCapabilities{
		Device:       device,
		Capabilities: capabilities,
	}, nil
}

// GetDeviceByOUIAndSN 根据OUI和SN获取设备
func (s *DeviceService) GetDeviceByOUIAndSN(oui, sn string) (*Device, error) {
	var device Device
	if err := s.db.DB.Where("oui = ? AND sn = ?", oui, sn).First(&device).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询设备失败: %v", err)
	}
	return &device, nil
}

// GetDeviceStats 获取设备统计信息
func (s *DeviceService) GetDeviceStats(deviceID uint) (map[string]interface{}, error) {
	var stats map[string]interface{}

	// 统计AI能力数量
	var capabilityCount int64
	if err := s.db.DB.Model(&DeviceCapability{}).Where("device_id = ? AND is_enabled = ?", deviceID, true).Count(&capabilityCount).Error; err != nil {
		return nil, fmt.Errorf("统计AI能力数量失败: %v", err)
	}

	// 统计会话数量
	var sessionCount int64
	if err := s.db.DB.Model(&Session{}).Where("device_id = ?", deviceID).Count(&sessionCount).Error; err != nil {
		return nil, fmt.Errorf("统计会话数量失败: %v", err)
	}

	// 统计使用次数
	var usageCount int64
	if err := s.db.DB.Model(&UsageStats{}).Where("device_id = ?", deviceID).Count(&usageCount).Error; err != nil {
		return nil, fmt.Errorf("统计使用次数失败: %v", err)
	}

	stats = map[string]interface{}{
		"capability_count": capabilityCount,
		"session_count":    sessionCount,
		"usage_count":      usageCount,
	}

	return stats, nil
}

// GetOnlineDevices 获取在线设备列表
func (s *DeviceService) GetOnlineDevices() ([]*Device, error) {
	var devices []*Device
	if err := s.db.DB.Where("status = ?", "online").Find(&devices).Error; err != nil {
		return nil, fmt.Errorf("查询在线设备失败: %v", err)
	}
	return devices, nil
}

// GetOfflineDevices 获取离线设备列表
func (s *DeviceService) GetOfflineDevices() ([]*Device, error) {
	var devices []*Device
	if err := s.db.DB.Where("status = ?", "offline").Find(&devices).Error; err != nil {
		return nil, fmt.Errorf("查询离线设备失败: %v", err)
	}
	return devices, nil
}

// GetUserDeviceBinding 获取用户设备绑定关系
func (s *DeviceService) GetUserDeviceBinding(userID, deviceID uint) (*UserDevice, error) {
	var userDevice UserDevice
	if err := s.db.DB.Where("user_id = ? AND device_id = ? AND is_active = ?", userID, deviceID, true).First(&userDevice).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询用户设备绑定失败: %v", err)
	}
	return &userDevice, nil
}
