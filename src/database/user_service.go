package database

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"ai-server-go/src/core/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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

	user.PasswordHash = string(hashedPassword)
	user.Status = "active"
	if user.Role == "" {
		user.Role = "user"
	}

	if err := s.db.DB.Create(user).Error; err != nil {
		return fmt.Errorf("创建用户失败: %v", err)
	}

	s.logger.Info("用户创建成功: %s (ID: %d)", user.Username, user.ID)
	return nil
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(id uint) (*User, error) {
	var user User
	if err := s.db.DB.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询用户失败: %v", err)
	}
	return &user, nil
}

// GetUserByUsername 根据用户名获取用户
func (s *UserService) GetUserByUsername(username string) (*User, error) {
	var user User
	if err := s.db.DB.Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询用户失败: %v", err)
	}
	return &user, nil
}

// UpdateUser 更新用户
func (s *UserService) UpdateUser(user *User) error {
	if err := s.db.DB.Save(user).Error; err != nil {
		return fmt.Errorf("更新用户失败: %v", err)
	}

	s.logger.Info("用户更新成功: %s", user.Username)
	return nil
}

// UpdatePassword 更新用户密码
func (s *UserService) UpdatePassword(userID uint, newPassword string) error {
	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %v", err)
	}

	if err := s.db.DB.Model(&User{}).Where("id = ?", userID).Update("password_hash", string(hashedPassword)).Error; err != nil {
		return fmt.Errorf("更新密码失败: %v", err)
	}

	s.logger.Info("用户密码更新成功: ID %d", userID)
	return nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(id uint) error {
	if err := s.db.DB.Delete(&User{}, id).Error; err != nil {
		return fmt.Errorf("删除用户失败: %v", err)
	}

	s.logger.Info("用户删除成功: ID %d", id)
	return nil
}

// ListUsers 获取用户列表
func (s *UserService) ListUsers(offset, limit int, status, role string) ([]*User, error) {
	var users []*User
	query := s.db.DB

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if role != "" {
		query = query.Where("role = ?", role)
	}

	if err := query.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("查询用户列表失败: %v", err)
	}

	return users, nil
}

// CountUsers 统计用户数量
func (s *UserService) CountUsers(status, role string) (int64, error) {
	var count int64
	query := s.db.DB.Model(&User{})

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if role != "" {
		query = query.Where("role = ?", role)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("统计用户数量失败: %v", err)
	}

	return count, nil
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

	s.logger.Info("开始认证用户: %s", username)
	s.logger.Info("传入的密码: '%s'", password)
	s.logger.Info("数据库中的哈希: '%s'", user.PasswordHash)

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		s.logger.Error("密码验证失败: %v", err)
		return nil, fmt.Errorf("密码错误")
	}

	s.logger.Info("密码验证成功 for user: %s", username)

	if user.Status != "active" {
		return nil, fmt.Errorf("用户状态异常")
	}

	return user, nil
}

// UpdateLoginInfo 更新登录信息
func (s *UserService) UpdateLoginInfo(userID uint, ipAddress string) error {
	now := time.Now()
	if err := s.db.DB.Model(&User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"last_login_time": &now,
		"last_login_ip":   ipAddress,
	}).Error; err != nil {
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

// CreateUserAuth 创建用户认证记录
func (s *UserService) CreateUserAuth(userID uint, expiresAt *time.Time) (*UserAuth, error) {
	token, err := s.generateToken()
	if err != nil {
		return nil, fmt.Errorf("生成令牌失败: %v", err)
	}

	auth := &UserAuth{
		UserID:     userID,
		AuthType:   "token",
		AuthKey:    token,
		AuthSecret: "",
		IsActive:   true,
		ExpiresAt:  expiresAt,
	}

	if err := s.db.DB.Create(auth).Error; err != nil {
		return nil, fmt.Errorf("创建认证记录失败: %v", err)
	}

	return auth, nil
}

// GetUserAuthByToken 根据令牌获取用户认证记录
func (s *UserService) GetUserAuthByToken(token string) (*UserAuth, error) {
	var auth UserAuth
	if err := s.db.DB.Where("auth_key = ? AND is_active = ?", token, true).First(&auth).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询认证记录失败: %v", err)
	}

	// 检查是否过期
	if auth.ExpiresAt != nil && time.Now().After(*auth.ExpiresAt) {
		// 标记为过期
		s.db.DB.Model(&auth).Update("is_active", false)
		return nil, nil
	}

	return &auth, nil
}

// GetUserAuthByKey 根据认证密钥获取用户认证记录
func (s *UserService) GetUserAuthByKey(authKey string) (*UserAuth, error) {
	var auth UserAuth
	if err := s.db.DB.Where("auth_key = ? AND is_active = ?", authKey, true).First(&auth).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询认证记录失败: %v", err)
	}

	// 检查是否过期
	if auth.ExpiresAt != nil && time.Now().After(*auth.ExpiresAt) {
		// 标记为过期
		s.db.DB.Model(&auth).Update("is_active", false)
		return nil, nil
	}

	return &auth, nil
}

// UpdateUserAuth 更新用户认证记录
func (s *UserService) UpdateUserAuth(auth *UserAuth) error {
	if err := s.db.DB.Save(auth).Error; err != nil {
		return fmt.Errorf("更新认证记录失败: %v", err)
	}
	return nil
}

// DeleteUserAuth 删除用户认证记录
func (s *UserService) DeleteUserAuth(id uint) error {
	if err := s.db.DB.Delete(&UserAuth{}, id).Error; err != nil {
		return fmt.Errorf("删除认证记录失败: %v", err)
	}
	return nil
}

// GetUserDeviceBinding 获取用户设备绑定关系
func (s *UserService) GetUserDeviceBinding(userID, deviceID uint) (*UserDevice, error) {
	var userDevice UserDevice
	if err := s.db.DB.Where("user_id = ? AND device_id = ?", userID, deviceID).First(&userDevice).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询设备绑定关系失败: %v", err)
	}
	return &userDevice, nil
}

// BindUserDevice 绑定用户设备
func (s *UserService) BindUserDevice(userID uint, deviceUUID string, deviceAlias string, isOwner bool, permissions map[string]interface{}) error {
	// 查找设备
	var device Device
	if err := s.db.DB.Where("device_uuid = ?", deviceUUID).First(&device).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("设备不存在")
		}
		return fmt.Errorf("查询设备失败: %v", err)
	}

	// 检查是否已绑定
	var existingBinding UserDevice
	if err := s.db.DB.Where("user_id = ? AND device_id = ?", userID, device.ID).First(&existingBinding).Error; err == nil {
		return fmt.Errorf("设备已绑定")
	}

	// 序列化权限
	permissionsJSON, err := json.Marshal(permissions)
	if err != nil {
		return fmt.Errorf("序列化权限失败: %v", err)
	}

	userDevice := &UserDevice{
		UserID:      userID,
		DeviceID:    device.ID,
		DeviceAlias: deviceAlias,
		IsOwner:     isOwner,
		Permissions: permissionsJSON,
		IsActive:    true,
	}

	if err := s.db.DB.Create(userDevice).Error; err != nil {
		return fmt.Errorf("绑定设备失败: %v", err)
	}

	s.logger.Info("设备绑定成功: 用户ID %d, 设备UUID %s", userID, deviceUUID)
	return nil
}

// UnbindUserDevice 解绑用户设备
func (s *UserService) UnbindUserDevice(userID uint, deviceUUID string) error {
	// 查找设备
	var device Device
	if err := s.db.DB.Where("device_uuid = ?", deviceUUID).First(&device).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("设备不存在")
		}
		return fmt.Errorf("查询设备失败: %v", err)
	}

	// 删除绑定关系
	if err := s.db.DB.Where("user_id = ? AND device_id = ?", userID, device.ID).Delete(&UserDevice{}).Error; err != nil {
		return fmt.Errorf("解绑设备失败: %v", err)
	}

	s.logger.Info("设备解绑成功: 用户ID %d, 设备UUID %s", userID, deviceUUID)
	return nil
}

// GetUserDevices 获取用户设备列表
func (s *UserService) GetUserDevices(userID uint) ([]*UserDevice, error) {
	var userDevices []*UserDevice
	if err := s.db.DB.Where("user_id = ? AND is_active = ?", userID, true).
		Preload("Device").
		Find(&userDevices).Error; err != nil {
		return nil, fmt.Errorf("查询用户设备失败: %v", err)
	}
	return userDevices, nil
}

// GetUserCapabilities 获取用户AI能力列表
func (s *UserService) GetUserCapabilities(userID uint) ([]*UserCapability, error) {
	var userCapabilities []*UserCapability
	if err := s.db.DB.Where("user_id = ? AND is_active = ?", userID, true).
		Preload("Capability").
		Find(&userCapabilities).Error; err != nil {
		return nil, fmt.Errorf("查询用户AI能力失败: %v", err)
	}
	return userCapabilities, nil
}

// SetUserCapability 设置用户AI能力
func (s *UserService) SetUserCapability(userID uint, capabilityName, capabilityType string, config map[string]interface{}) error {
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
	var userCapability UserCapability
	if err := s.db.DB.Where("user_id = ? AND capability_id = ?", userID, capability.ID).First(&userCapability).Error; err == nil {
		// 更新现有配置
		userCapability.ConfigData = configJSON
		userCapability.IsActive = true
		if err := s.db.DB.Save(&userCapability).Error; err != nil {
			return fmt.Errorf("更新用户AI能力失败: %v", err)
		}
	} else {
		// 创建新配置
		userCapability = UserCapability{
			UserID:       userID,
			CapabilityID: capability.ID,
			ConfigData:   configJSON,
			IsActive:     true,
		}
		if err := s.db.DB.Create(&userCapability).Error; err != nil {
			return fmt.Errorf("创建用户AI能力失败: %v", err)
		}
	}

	s.logger.Info("用户AI能力设置成功: 用户ID %d, 能力 %s", userID, capabilityName)
	return nil
}

// GetUserWithDevices 获取用户及其设备
func (s *UserService) GetUserWithDevices(userID uint) (*UserWithDevices, error) {
	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("用户不存在")
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
func (s *UserService) GetUserWithCapabilities(userID uint) (*UserWithCapabilities, error) {
	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("用户不存在")
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

// GetUserStats 获取用户统计信息
func (s *UserService) GetUserStats() (map[string]interface{}, error) {
	var totalUsers int64
	var activeUsers int64
	var inactiveUsers int64

	if err := s.db.DB.Model(&User{}).Count(&totalUsers).Error; err != nil {
		return nil, err
	}
	if err := s.db.DB.Model(&User{}).Where("status = ?", "active").Count(&activeUsers).Error; err != nil {
		return nil, err
	}
	if err := s.db.DB.Model(&User{}).Where("status = ?", "inactive").Count(&inactiveUsers).Error; err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_users":    totalUsers,
		"active_users":   activeUsers,
		"inactive_users": inactiveUsers,
	}

	return stats, nil
}

// ResetPassword 重置用户密码
func (s *UserService) ResetPassword(id int, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	result := s.db.DB.Model(&User{}).Where("id = ?", id).Update("password", hashedPassword)
	return result.Error
}
