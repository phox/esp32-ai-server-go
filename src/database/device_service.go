package database

import (
	"database/sql"
	"fmt"
	"time"

	"ai-server-go/src/core/utils"

	"github.com/google/uuid"
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

// generateDeviceUUID 生成设备UUID
func (s *DeviceService) generateDeviceUUID() string {
	return uuid.New().String()
}

// CreateDevice 创建设备
func (s *DeviceService) CreateDevice(device *Device) error {
	// 生成设备UUID
	device.DeviceUUID = s.generateDeviceUUID()

	query := `
		INSERT INTO devices (device_uuid, oui, sn, device_name, device_type, device_model, firmware_version, hardware_version, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := s.db.Exec(query,
		device.DeviceUUID,
		device.OUI,
		device.SN,
		device.DeviceName,
		device.DeviceType,
		device.DeviceModel,
		device.FirmwareVersion,
		device.HardwareVersion,
		device.Status,
	)

	if err != nil {
		return fmt.Errorf("创建设备失败: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取设备ID失败: %v", err)
	}

	device.ID = id
	s.logger.Info("设备创建成功: %s (OUI: %s, SN: %s, UUID: %s, ID: %d)", device.DeviceName, device.OUI, device.SN, device.DeviceUUID, id)
	return nil
}

// GetDeviceByID 根据ID获取设备
func (s *DeviceService) GetDeviceByID(id int64) (*Device, error) {
	query := `SELECT * FROM devices WHERE id = ?`

	var device Device
	err := s.db.QueryRow(query, id).Scan(
		&device.ID,
		&device.DeviceUUID,
		&device.OUI,
		&device.SN,
		&device.DeviceName,
		&device.DeviceType,
		&device.DeviceModel,
		&device.FirmwareVersion,
		&device.HardwareVersion,
		&device.Status,
		&device.LastOnlineTime,
		&device.LastIPAddress,
		&device.CreatedAt,
		&device.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询设备失败: %v", err)
	}

	return &device, nil
}

// GetDeviceByUUID 根据设备UUID获取设备
func (s *DeviceService) GetDeviceByUUID(deviceUUID string) (*Device, error) {
	query := `SELECT * FROM devices WHERE device_uuid = ?`

	var device Device
	err := s.db.QueryRow(query, deviceUUID).Scan(
		&device.ID,
		&device.DeviceUUID,
		&device.OUI,
		&device.SN,
		&device.DeviceName,
		&device.DeviceType,
		&device.DeviceModel,
		&device.FirmwareVersion,
		&device.HardwareVersion,
		&device.Status,
		&device.LastOnlineTime,
		&device.LastIPAddress,
		&device.CreatedAt,
		&device.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询设备失败: %v", err)
	}

	return &device, nil
}

// GetDeviceByOUISN 根据OUI和SN获取设备
func (s *DeviceService) GetDeviceByOUISN(oui, sn string) (*Device, error) {
	query := `SELECT * FROM devices WHERE oui = ? AND sn = ?`

	var device Device
	err := s.db.QueryRow(query, oui, sn).Scan(
		&device.ID,
		&device.DeviceUUID,
		&device.OUI,
		&device.SN,
		&device.DeviceName,
		&device.DeviceType,
		&device.DeviceModel,
		&device.FirmwareVersion,
		&device.HardwareVersion,
		&device.Status,
		&device.LastOnlineTime,
		&device.LastIPAddress,
		&device.CreatedAt,
		&device.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询设备失败: %v", err)
	}

	return &device, nil
}

// UpdateDevice 更新设备
func (s *DeviceService) UpdateDevice(device *Device) error {
	query := `
		UPDATE devices 
		SET device_name = ?, device_type = ?, device_model = ?, firmware_version = ?, 
		    hardware_version = ?, status = ?, last_online_time = ?, last_ip_address = ?
		WHERE id = ?
	`

	_, err := s.db.Exec(query,
		device.DeviceName,
		device.DeviceType,
		device.DeviceModel,
		device.FirmwareVersion,
		device.HardwareVersion,
		device.Status,
		device.LastOnlineTime,
		device.LastIPAddress,
		device.ID,
	)

	if err != nil {
		return fmt.Errorf("更新设备失败: %v", err)
	}

	s.logger.Info("设备更新成功: %s (OUI: %s, SN: %s, UUID: %s)", device.DeviceName, device.OUI, device.SN, device.DeviceUUID)
	return nil
}

// UpdateDeviceOnlineStatus 更新设备在线状态
func (s *DeviceService) UpdateDeviceOnlineStatus(deviceUUID string, ipAddress string) error {
	query := `
		UPDATE devices 
		SET last_online_time = ?, last_ip_address = ?
		WHERE device_uuid = ?
	`

	now := time.Now()
	_, err := s.db.Exec(query, now, ipAddress, deviceUUID)

	if err != nil {
		return fmt.Errorf("更新设备在线状态失败: %v", err)
	}

	return nil
}

// DeleteDevice 删除设备
func (s *DeviceService) DeleteDevice(id int64) error {
	query := `DELETE FROM devices WHERE id = ?`

	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("删除设备失败: %v", err)
	}

	s.logger.Info("设备删除成功: ID %d", id)
	return nil
}

// ListDevices 获取设备列表
func (s *DeviceService) ListDevices(offset, limit int, status, oui string) ([]*Device, error) {
	query := `SELECT * FROM devices`
	args := []interface{}{}

	conditions := []string{}
	if status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, status)
	}
	if oui != "" {
		conditions = append(conditions, "oui = ?")
		args = append(args, oui)
	}

	if len(conditions) > 0 {
		query += ` WHERE ` + conditions[0]
		for i := 1; i < len(conditions); i++ {
			query += ` AND ` + conditions[i]
		}
	}

	query += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询设备列表失败: %v", err)
	}
	defer rows.Close()

	var devices []*Device
	for rows.Next() {
		var device Device
		err := rows.Scan(
			&device.ID,
			&device.DeviceUUID,
			&device.OUI,
			&device.SN,
			&device.DeviceName,
			&device.DeviceType,
			&device.DeviceModel,
			&device.FirmwareVersion,
			&device.HardwareVersion,
			&device.Status,
			&device.LastOnlineTime,
			&device.LastIPAddress,
			&device.CreatedAt,
			&device.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描设备数据失败: %v", err)
		}
		devices = append(devices, &device)
	}

	return devices, nil
}

// CreateDeviceAuth 创建设备认证
func (s *DeviceService) CreateDeviceAuth(auth *DeviceAuth) error {
	query := `
		INSERT INTO device_auth (device_id, auth_type, auth_key, auth_secret, is_active, expires_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := s.db.Exec(query,
		auth.DeviceID,
		auth.AuthType,
		auth.AuthKey,
		auth.AuthSecret,
		auth.IsActive,
		auth.ExpiresAt,
	)

	if err != nil {
		return fmt.Errorf("创建设备认证失败: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取认证ID失败: %v", err)
	}

	auth.ID = id
	s.logger.Info("设备认证创建成功: DeviceID %d, AuthKey %s", auth.DeviceID, auth.AuthKey)
	return nil
}

// GetDeviceAuthByKey 根据认证密钥获取认证信息
func (s *DeviceService) GetDeviceAuthByKey(authKey string) (*DeviceAuth, error) {
	query := `
		SELECT da.*, d.device_id as device_device_id, d.status as device_status
		FROM device_auth da
		JOIN devices d ON da.device_id = d.id
		WHERE da.auth_key = ? AND da.is_active = true
	`

	var auth DeviceAuth
	var deviceDeviceID, deviceStatus string
	err := s.db.QueryRow(query, authKey).Scan(
		&auth.ID,
		&auth.DeviceID,
		&auth.AuthType,
		&auth.AuthKey,
		&auth.AuthSecret,
		&auth.IsActive,
		&auth.ExpiresAt,
		&auth.CreatedAt,
		&auth.UpdatedAt,
		&deviceDeviceID,
		&deviceStatus,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询设备认证失败: %v", err)
	}

	// 检查设备状态
	if deviceStatus != "active" {
		return nil, fmt.Errorf("设备状态异常: %s", deviceStatus)
	}

	// 检查认证是否过期
	if auth.ExpiresAt != nil && auth.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("认证已过期")
	}

	return &auth, nil
}

// UpdateDeviceAuth 更新设备认证
func (s *DeviceService) UpdateDeviceAuth(auth *DeviceAuth) error {
	query := `
		UPDATE device_auth 
		SET auth_type = ?, auth_key = ?, auth_secret = ?, is_active = ?, expires_at = ?
		WHERE id = ?
	`

	_, err := s.db.Exec(query,
		auth.AuthType,
		auth.AuthKey,
		auth.AuthSecret,
		auth.IsActive,
		auth.ExpiresAt,
		auth.ID,
	)

	if err != nil {
		return fmt.Errorf("更新设备认证失败: %v", err)
	}

	s.logger.Info("设备认证更新成功: ID %d", auth.ID)
	return nil
}

// DeleteDeviceAuth 删除设备认证
func (s *DeviceService) DeleteDeviceAuth(id int64) error {
	query := `DELETE FROM device_auth WHERE id = ?`

	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("删除设备认证失败: %v", err)
	}

	s.logger.Info("设备认证删除成功: ID %d", id)
	return nil
}

// GetDeviceWithCapabilities 获取设备及其AI能力配置
func (s *DeviceService) GetDeviceWithCapabilities(deviceUUID string) (*DeviceWithCapabilities, error) {
	// 获取设备信息
	device, err := s.GetDeviceByUUID(deviceUUID)
	if err != nil {
		return nil, err
	}
	if device == nil {
		return nil, fmt.Errorf("设备不存在: %s", deviceUUID)
	}

	// 获取设备AI能力配置
	query := `
		SELECT dc.*, ac.capability_name, ac.capability_type, ac.display_name, ac.description, ac.config_schema
		FROM device_capabilities dc
		JOIN ai_capabilities ac ON dc.capability_id = ac.id
		WHERE dc.device_id = ? AND dc.is_enabled = true
		ORDER BY dc.priority ASC
	`

	rows, err := s.db.Query(query, device.ID)
	if err != nil {
		return nil, fmt.Errorf("查询设备AI能力失败: %v", err)
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
			return nil, fmt.Errorf("扫描设备AI能力数据失败: %v", err)
		}
		dc.Capability = &ac
		capabilities = append(capabilities, &dc)
	}

	return &DeviceWithCapabilities{
		Device:       device,
		Capabilities: capabilities,
	}, nil
}
