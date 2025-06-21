package database

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"ai-server-go/src/core/utils"

	"golang.org/x/crypto/bcrypt"
)

// UserService 用户管理服务
type UserService struct {
	db     *Database
	logger *utils.Logger
}

// NewUserService 创建用户管理服务
func NewUserService(db *Database, logger *utils.Logger) *UserService {
	return &UserService{
		db:     db,
		logger: logger,
	}
}

// GetDB 获取数据库连接
func (s *UserService) GetDB() *Database {
	return s.db
}

// CreateUser 创建用户
func (s *UserService) CreateUser(user *User, password string) error {
	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %v", err)
	}

	query := `
		INSERT INTO users (username, email, phone, nickname, avatar, password_hash, status, role)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := s.db.Exec(query,
		user.Username,
		user.Email,
		user.Phone,
		user.Nickname,
		user.Avatar,
		hashedPassword,
		user.Status,
		user.Role,
	)

	if err != nil {
		return fmt.Errorf("创建用户失败: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取用户ID失败: %v", err)
	}

	user.ID = id
	s.logger.Info("用户创建成功: %s (ID: %d)", user.Username, id)
	return nil
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(id int64) (*User, error) {
	query := `SELECT * FROM users WHERE id = ?`

	var user User
	err := s.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Phone,
		&user.Nickname,
		&user.Avatar,
		&user.PasswordHash,
		&user.Status,
		&user.Role,
		&user.LastLoginTime,
		&user.LastLoginIP,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询用户失败: %v", err)
	}

	return &user, nil
}

// GetUserByUsername 根据用户名获取用户
func (s *UserService) GetUserByUsername(username string) (*User, error) {
	query := `SELECT * FROM users WHERE username = ?`

	var user User
	err := s.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Phone,
		&user.Nickname,
		&user.Avatar,
		&user.PasswordHash,
		&user.Status,
		&user.Role,
		&user.LastLoginTime,
		&user.LastLoginIP,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询用户失败: %v", err)
	}

	return &user, nil
}

// UpdateUser 更新用户
func (s *UserService) UpdateUser(user *User) error {
	query := `
		UPDATE users 
		SET email = ?, phone = ?, nickname = ?, avatar = ?, status = ?, role = ?
		WHERE id = ?
	`

	_, err := s.db.Exec(query,
		user.Email,
		user.Phone,
		user.Nickname,
		user.Avatar,
		user.Status,
		user.Role,
		user.ID,
	)

	if err != nil {
		return fmt.Errorf("更新用户失败: %v", err)
	}

	s.logger.Info("用户更新成功: %s", user.Username)
	return nil
}

// UpdatePassword 更新用户密码
func (s *UserService) UpdatePassword(userID int64, newPassword string) error {
	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %v", err)
	}

	query := `UPDATE users SET password_hash = ? WHERE id = ?`

	_, err = s.db.Exec(query, hashedPassword, userID)
	if err != nil {
		return fmt.Errorf("更新密码失败: %v", err)
	}

	s.logger.Info("用户密码更新成功: ID %d", userID)
	return nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(id int64) error {
	query := `DELETE FROM users WHERE id = ?`

	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("删除用户失败: %v", err)
	}

	s.logger.Info("用户删除成功: ID %d", id)
	return nil
}

// ListUsers 获取用户列表
func (s *UserService) ListUsers(offset, limit int, status, role string) ([]*User, error) {
	query := `SELECT * FROM users`
	args := []interface{}{}

	conditions := []string{}
	if status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, status)
	}
	if role != "" {
		conditions = append(conditions, "role = ?")
		args = append(args, role)
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
		return nil, fmt.Errorf("查询用户列表失败: %v", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Phone,
			&user.Nickname,
			&user.Avatar,
			&user.PasswordHash,
			&user.Status,
			&user.Role,
			&user.LastLoginTime,
			&user.LastLoginIP,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描用户数据失败: %v", err)
		}
		users = append(users, &user)
	}

	return users, nil
}

// AuthenticateUser 用户认证
func (s *UserService) AuthenticateUser(username, password string) (*User, error) {
	user, err := s.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("用户不存在")
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("密码错误")
	}

	// 检查用户状态
	if user.Status != "active" {
		return nil, fmt.Errorf("用户状态异常: %s", user.Status)
	}

	return user, nil
}

// UpdateLoginInfo 更新登录信息
func (s *UserService) UpdateLoginInfo(userID int64, ipAddress string) error {
	query := `
		UPDATE users 
		SET last_login_time = ?, last_login_ip = ?
		WHERE id = ?
	`

	now := time.Now()
	_, err := s.db.Exec(query, now, ipAddress, userID)

	if err != nil {
		return fmt.Errorf("更新登录信息失败: %v", err)
	}

	return nil
}

// generateToken 生成认证令牌
func (s *UserService) generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CreateUserAuth 创建用户认证令牌
func (s *UserService) CreateUserAuth(userID int64, expiresAt *time.Time) (*UserAuth, error) {
	token, err := s.generateToken()
	if err != nil {
		return nil, fmt.Errorf("生成令牌失败: %v", err)
	}

	auth := &UserAuth{
		UserID:    userID,
		AuthType:  "token",
		AuthKey:   token,
		IsActive:  true,
		ExpiresAt: expiresAt,
	}

	query := `
		INSERT INTO user_auth (user_id, auth_type, auth_key, is_active, expires_at)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := s.db.Exec(query,
		auth.UserID,
		auth.AuthType,
		auth.AuthKey,
		auth.IsActive,
		auth.ExpiresAt,
	)

	if err != nil {
		return nil, fmt.Errorf("创建用户认证失败: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("获取认证ID失败: %v", err)
	}

	auth.ID = id
	s.logger.Info("用户认证创建成功: UserID %d, Token %s", auth.UserID, auth.AuthKey)
	return auth, nil
}

// GetUserAuthByToken 根据令牌获取认证信息
func (s *UserService) GetUserAuthByToken(token string) (*UserAuth, error) {
	query := `
		SELECT ua.*, u.username, u.status as user_status
		FROM user_auth ua
		JOIN users u ON ua.user_id = u.id
		WHERE ua.auth_key = ? AND ua.is_active = true
	`

	var auth UserAuth
	var username, userStatus string
	err := s.db.QueryRow(query, token).Scan(
		&auth.ID,
		&auth.UserID,
		&auth.AuthType,
		&auth.AuthKey,
		&auth.AuthSecret,
		&auth.IsActive,
		&auth.ExpiresAt,
		&auth.CreatedAt,
		&auth.UpdatedAt,
		&username,
		&userStatus,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询用户认证失败: %v", err)
	}

	// 检查用户状态
	if userStatus != "active" {
		return nil, fmt.Errorf("用户状态异常: %s", userStatus)
	}

	// 检查认证是否过期
	if auth.ExpiresAt != nil && auth.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("认证已过期")
	}

	return &auth, nil
}

// UpdateUserAuth 更新用户认证
func (s *UserService) UpdateUserAuth(auth *UserAuth) error {
	query := `
		UPDATE user_auth 
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
		return fmt.Errorf("更新用户认证失败: %v", err)
	}

	s.logger.Info("用户认证更新成功: ID %d", auth.ID)
	return nil
}

// DeleteUserAuth 删除用户认证
func (s *UserService) DeleteUserAuth(id int64) error {
	query := `DELETE FROM user_auth WHERE id = ?`

	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("删除用户认证失败: %v", err)
	}

	s.logger.Info("用户认证删除成功: ID %d", id)
	return nil
}

// GetUserDeviceBinding 获取用户设备绑定关系
func (s *UserService) GetUserDeviceBinding(userID int64, deviceID int64) (*UserDevice, error) {
	query := `SELECT * FROM user_devices WHERE user_id = ? AND device_id = ?`

	var userDevice UserDevice
	err := s.db.QueryRow(query, userID, deviceID).Scan(
		&userDevice.ID,
		&userDevice.UserID,
		&userDevice.DeviceID,
		&userDevice.DeviceAlias,
		&userDevice.IsOwner,
		&userDevice.Permissions,
		&userDevice.IsActive,
		&userDevice.CreatedAt,
		&userDevice.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询用户设备绑定失败: %v", err)
	}

	return &userDevice, nil
}

// BindUserDevice 绑定用户设备
func (s *UserService) BindUserDevice(userID int64, deviceUUID string, deviceAlias string, isOwner bool, permissions map[string]interface{}) error {
	// 检查设备是否存在
	deviceService := NewDeviceService(s.db, s.logger)
	device, err := deviceService.GetDeviceByUUID(deviceUUID)
	if err != nil {
		return fmt.Errorf("查询设备失败: %v", err)
	}
	if device == nil {
		return fmt.Errorf("设备不存在: %s", deviceUUID)
	}

	// 检查是否已绑定
	existingBinding, err := s.GetUserDeviceBinding(userID, device.ID)
	if err != nil {
		return fmt.Errorf("查询设备绑定失败: %v", err)
	}
	if existingBinding != nil {
		return fmt.Errorf("设备已绑定给用户")
	}

	// 序列化权限配置
	permissionsJSON, err := json.Marshal(permissions)
	if err != nil {
		return fmt.Errorf("序列化权限配置失败: %v", err)
	}

	query := `
		INSERT INTO user_devices (user_id, device_id, device_alias, is_owner, permissions, is_active)
		VALUES (?, ?, ?, ?, ?, true)
	`

	_, err = s.db.Exec(query, userID, device.ID, deviceAlias, isOwner, permissionsJSON)
	if err != nil {
		return fmt.Errorf("绑定用户设备失败: %v", err)
	}

	s.logger.Info("用户设备绑定成功: UserID %d, DeviceUUID %s", userID, deviceUUID)
	return nil
}

// UnbindUserDevice 解绑用户设备
func (s *UserService) UnbindUserDevice(userID int64, deviceUUID string) error {
	// 检查设备是否存在
	deviceService := NewDeviceService(s.db, s.logger)
	device, err := deviceService.GetDeviceByUUID(deviceUUID)
	if err != nil {
		return fmt.Errorf("查询设备失败: %v", err)
	}
	if device == nil {
		return fmt.Errorf("设备不存在: %s", deviceUUID)
	}

	query := `DELETE FROM user_devices WHERE user_id = ? AND device_id = ?`

	_, err = s.db.Exec(query, userID, device.ID)
	if err != nil {
		return fmt.Errorf("解绑用户设备失败: %v", err)
	}

	s.logger.Info("用户设备解绑成功: UserID %d, DeviceUUID %s", userID, deviceUUID)
	return nil
}

// GetUserDevices 获取用户设备列表
func (s *UserService) GetUserDevices(userID int64) ([]*UserDevice, error) {
	query := `
		SELECT ud.*, d.device_uuid, d.oui, d.sn, d.device_name, d.device_type, d.device_model, 
		       d.firmware_version, d.hardware_version, d.status, d.last_online_time, d.last_ip_address,
		       d.created_at as device_created_at, d.updated_at as device_updated_at
		FROM user_devices ud
		JOIN devices d ON ud.device_id = d.id
		WHERE ud.user_id = ? AND ud.is_active = true
		ORDER BY ud.created_at DESC
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("查询用户设备失败: %v", err)
	}
	defer rows.Close()

	var userDevices []*UserDevice
	for rows.Next() {
		var ud UserDevice
		var device Device
		var deviceCreatedAt, deviceUpdatedAt time.Time
		err := rows.Scan(
			&ud.ID,
			&ud.UserID,
			&ud.DeviceID,
			&ud.DeviceAlias,
			&ud.IsOwner,
			&ud.Permissions,
			&ud.IsActive,
			&ud.CreatedAt,
			&ud.UpdatedAt,
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
			&deviceCreatedAt,
			&deviceUpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描用户设备数据失败: %v", err)
		}
		device.ID = ud.DeviceID
		device.CreatedAt = deviceCreatedAt
		device.UpdatedAt = deviceUpdatedAt
		ud.Device = &device
		userDevices = append(userDevices, &ud)
	}

	return userDevices, nil
}

// GetUserCapabilities 获取用户AI能力配置
func (s *UserService) GetUserCapabilities(userID int64) ([]*UserCapability, error) {
	query := `
		SELECT uc.*, ac.capability_name, ac.capability_type, ac.display_name, ac.description
		FROM user_capabilities uc
		JOIN ai_capabilities ac ON uc.capability_id = ac.id
		WHERE uc.user_id = ? AND uc.is_active = true
		ORDER BY ac.capability_name, ac.capability_type
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("查询用户AI能力失败: %v", err)
	}
	defer rows.Close()

	var userCapabilities []*UserCapability
	for rows.Next() {
		var uc UserCapability
		var ac AICapability
		err := rows.Scan(
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
		)
		if err != nil {
			return nil, fmt.Errorf("扫描用户AI能力数据失败: %v", err)
		}
		uc.Capability = &ac
		userCapabilities = append(userCapabilities, &uc)
	}

	return userCapabilities, nil
}

// SetUserCapability 设置用户AI能力配置
func (s *UserService) SetUserCapability(userID int64, capabilityName, capabilityType string, config map[string]interface{}) error {
	// 获取AI能力ID
	configService := NewConfigService(s.db, s.logger)
	capability, err := configService.GetAICapability(capabilityName, capabilityType)
	if err != nil {
		return fmt.Errorf("获取AI能力失败: %v", err)
	}
	if capability == nil {
		return fmt.Errorf("AI能力不存在: %s/%s", capabilityName, capabilityType)
	}

	// 序列化配置数据
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置数据失败: %v", err)
	}

	query := `
		INSERT INTO user_capabilities (user_id, capability_id, config_data, is_active)
		VALUES (?, ?, ?, true)
		ON DUPLICATE KEY UPDATE 
		config_data = VALUES(config_data),
		is_active = VALUES(is_active)
	`

	_, err = s.db.Exec(query, userID, capability.ID, configJSON)
	if err != nil {
		return fmt.Errorf("设置用户AI能力失败: %v", err)
	}

	s.logger.Info("用户AI能力设置成功: UserID %d, Capability %s/%s", userID, capabilityName, capabilityType)
	return nil
}

// GetUserWithDevices 获取用户及其设备
func (s *UserService) GetUserWithDevices(userID int64) (*UserWithDevices, error) {
	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("用户不存在: %d", userID)
	}

	devices, err := s.GetUserDevices(userID)
	if err != nil {
		return nil, err
	}

	return &UserWithDevices{
		User:    user,
		Devices: devices,
	}, nil
}

// GetUserWithCapabilities 获取用户及其AI能力
func (s *UserService) GetUserWithCapabilities(userID int64) (*UserWithCapabilities, error) {
	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("用户不存在: %d", userID)
	}

	capabilities, err := s.GetUserCapabilities(userID)
	if err != nil {
		return nil, err
	}

	return &UserWithCapabilities{
		User:         user,
		Capabilities: capabilities,
	}, nil
}

// GetUserAuthByKey 根据认证密钥获取用户认证信息
func (s *UserService) GetUserAuthByKey(authKey string) (*UserAuth, error) {
	query := `SELECT * FROM user_auth WHERE auth_key = ? AND is_active = 1`

	var auth UserAuth
	err := s.db.QueryRow(query, authKey).Scan(
		&auth.ID,
		&auth.UserID,
		&auth.AuthType,
		&auth.AuthKey,
		&auth.AuthSecret,
		&auth.IsActive,
		&auth.ExpiresAt,
		&auth.CreatedAt,
		&auth.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询用户认证失败: %v", err)
	}

	return &auth, nil
}

// GetUserStats 获取用户统计信息
func (s *UserService) GetUserStats(userID int64) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 获取设备数量
	var deviceCount int
	query := `SELECT COUNT(*) FROM user_devices WHERE user_id = ? AND is_active = 1`
	err := s.db.QueryRow(query, userID).Scan(&deviceCount)
	if err != nil {
		return nil, fmt.Errorf("获取设备数量失败: %v", err)
	}
	stats["device_count"] = deviceCount

	// 获取AI能力数量
	var capabilityCount int
	query = `SELECT COUNT(*) FROM user_capabilities WHERE user_id = ? AND is_active = 1`
	err = s.db.QueryRow(query, userID).Scan(&capabilityCount)
	if err != nil {
		return nil, fmt.Errorf("获取AI能力数量失败: %v", err)
	}
	stats["capability_count"] = capabilityCount

	// 获取会话数量（最近30天）
	var sessionCount int
	query = `SELECT COUNT(*) FROM sessions WHERE user_id = ? AND start_time >= DATE_SUB(NOW(), INTERVAL 30 DAY)`
	err = s.db.QueryRow(query, userID).Scan(&sessionCount)
	if err != nil {
		return nil, fmt.Errorf("获取会话数量失败: %v", err)
	}
	stats["session_count_30d"] = sessionCount

	// 获取使用统计（最近30天）
	var totalRequests, successRequests, errorRequests int
	query = `SELECT 
		SUM(request_count) as total_requests,
		SUM(success_count) as success_requests,
		SUM(error_count) as error_requests
		FROM usage_stats 
		WHERE user_id = ? AND usage_date >= DATE_SUB(NOW(), INTERVAL 30 DAY)`

	err = s.db.QueryRow(query, userID).Scan(&totalRequests, &successRequests, &errorRequests)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("获取使用统计失败: %v", err)
	}

	stats["total_requests_30d"] = totalRequests
	stats["success_requests_30d"] = successRequests
	stats["error_requests_30d"] = errorRequests

	// 计算成功率
	if totalRequests > 0 {
		stats["success_rate_30d"] = float64(successRequests) / float64(totalRequests) * 100
	} else {
		stats["success_rate_30d"] = 0.0
	}

	return stats, nil
}
