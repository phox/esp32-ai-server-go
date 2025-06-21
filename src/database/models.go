package database

import (
	"encoding/json"
	"time"
)

// User 用户模型
type User struct {
	ID            int64      `json:"id" db:"id"`
	Username      string     `json:"username" db:"username"`
	Email         string     `json:"email" db:"email"`
	Phone         string     `json:"phone" db:"phone"`
	PasswordHash  string     `json:"-" db:"password_hash"`
	Salt          string     `json:"-" db:"salt"`
	Nickname      string     `json:"nickname" db:"nickname"`
	Avatar        string     `json:"avatar" db:"avatar"`
	Status        string     `json:"status" db:"status"`
	Role          string     `json:"role" db:"role"`
	LastLoginTime *time.Time `json:"last_login_time" db:"last_login_time"`
	LastLoginIP   string     `json:"last_login_ip" db:"last_login_ip"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

// UserAuth 用户认证模型
type UserAuth struct {
	ID         int64      `json:"id" db:"id"`
	UserID     int64      `json:"user_id" db:"user_id"`
	AuthType   string     `json:"auth_type" db:"auth_type"`
	AuthKey    string     `json:"auth_key" db:"auth_key"`
	AuthSecret string     `json:"auth_secret" db:"auth_secret"`
	IsActive   bool       `json:"is_active" db:"is_active"`
	ExpiresAt  *time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
}

// UserDevice 用户设备绑定模型
type UserDevice struct {
	ID          int64           `json:"id" db:"id"`
	UserID      int64           `json:"user_id" db:"user_id"`
	DeviceID    int64           `json:"device_id" db:"device_id"`
	DeviceAlias string          `json:"device_alias" db:"device_alias"`
	IsOwner     bool            `json:"is_owner" db:"is_owner"`
	Permissions json.RawMessage `json:"permissions" db:"permissions"`
	IsActive    bool            `json:"is_active" db:"is_active"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`

	// 关联查询字段
	Device *Device `json:"device,omitempty"`
	User   *User   `json:"user,omitempty"`
}

// UserCapability 用户AI能力模型
type UserCapability struct {
	ID           int64           `json:"id" db:"id"`
	UserID       int64           `json:"user_id" db:"user_id"`
	CapabilityID int64           `json:"capability_id" db:"capability_id"`
	ConfigData   json.RawMessage `json:"config_data" db:"config_data"`
	IsActive     bool            `json:"is_active" db:"is_active"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at" db:"updated_at"`

	// 关联查询字段
	Capability *AICapability `json:"capability,omitempty"`
}

// Device 设备模型
type Device struct {
	ID              int64      `json:"id" db:"id"`
	DeviceUUID      string     `json:"device_uuid" db:"device_uuid"`
	OUI             string     `json:"oui" db:"oui"`
	SN              string     `json:"sn" db:"sn"`
	DeviceName      string     `json:"device_name" db:"device_name"`
	DeviceType      string     `json:"device_type" db:"device_type"`
	DeviceModel     string     `json:"device_model" db:"device_model"`
	FirmwareVersion string     `json:"firmware_version" db:"firmware_version"`
	HardwareVersion string     `json:"hardware_version" db:"hardware_version"`
	Status          string     `json:"status" db:"status"`
	LastOnlineTime  *time.Time `json:"last_online_time" db:"last_online_time"`
	LastIPAddress   string     `json:"last_ip_address" db:"last_ip_address"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// DeviceAuth 设备认证模型
type DeviceAuth struct {
	ID         int64      `json:"id" db:"id"`
	DeviceID   int64      `json:"device_id" db:"device_id"`
	AuthType   string     `json:"auth_type" db:"auth_type"`
	AuthKey    string     `json:"auth_key" db:"auth_key"`
	AuthSecret string     `json:"auth_secret" db:"auth_secret"`
	IsActive   bool       `json:"is_active" db:"is_active"`
	ExpiresAt  *time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
}

// AICapability AI能力模型
type AICapability struct {
	ID             int64           `json:"id" db:"id"`
	CapabilityName string          `json:"capability_name" db:"capability_name"`
	CapabilityType string          `json:"capability_type" db:"capability_type"`
	DisplayName    string          `json:"display_name" db:"display_name"`
	Description    string          `json:"description" db:"description"`
	ConfigSchema   json.RawMessage `json:"config_schema" db:"config_schema"`
	IsGlobal       bool            `json:"is_global" db:"is_global"`
	IsActive       bool            `json:"is_active" db:"is_active"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
}

// DeviceCapability 设备AI能力关联模型
type DeviceCapability struct {
	ID           int64           `json:"id" db:"id"`
	DeviceID     int64           `json:"device_id" db:"device_id"`
	CapabilityID int64           `json:"capability_id" db:"capability_id"`
	Priority     int             `json:"priority" db:"priority"`
	ConfigData   json.RawMessage `json:"config_data" db:"config_data"`
	IsEnabled    bool            `json:"is_enabled" db:"is_enabled"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at" db:"updated_at"`

	// 关联查询字段
	Capability *AICapability `json:"capability,omitempty"`
}

// GlobalConfig 全局配置模型
type GlobalConfig struct {
	ID          int64     `json:"id" db:"id"`
	ConfigKey   string    `json:"config_key" db:"config_key"`
	ConfigValue string    `json:"config_value" db:"config_value"`
	ConfigType  string    `json:"config_type" db:"config_type"`
	Description string    `json:"description" db:"description"`
	IsSystem    bool      `json:"is_system" db:"is_system"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Session 会话模型
type Session struct {
	ID           int64      `json:"id" db:"id"`
	UserID       *int64     `json:"user_id" db:"user_id"`
	DeviceID     int64      `json:"device_id" db:"device_id"`
	SessionID    string     `json:"session_id" db:"session_id"`
	ClientID     string     `json:"client_id" db:"client_id"`
	StartTime    time.Time  `json:"start_time" db:"start_time"`
	EndTime      *time.Time `json:"end_time" db:"end_time"`
	Status       string     `json:"status" db:"status"`
	MessageCount int        `json:"message_count" db:"message_count"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// UsageStats 使用统计模型
type UsageStats struct {
	ID             int64     `json:"id" db:"id"`
	UserID         *int64    `json:"user_id" db:"user_id"`
	DeviceID       int64     `json:"device_id" db:"device_id"`
	CapabilityName string    `json:"capability_name" db:"capability_name"`
	UsageDate      time.Time `json:"usage_date" db:"usage_date"`
	RequestCount   int       `json:"request_count" db:"request_count"`
	SuccessCount   int       `json:"success_count" db:"success_count"`
	ErrorCount     int       `json:"error_count" db:"error_count"`
	TotalDuration  int       `json:"total_duration" db:"total_duration"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// DeviceWithCapabilities 设备及其AI能力
type DeviceWithCapabilities struct {
	Device       *Device             `json:"device"`
	Capabilities []*DeviceCapability `json:"capabilities"`
}

// UserWithDevices 用户及其设备
type UserWithDevices struct {
	User    *User         `json:"user"`
	Devices []*UserDevice `json:"devices"`
}

// UserWithCapabilities 用户及其AI能力
type UserWithCapabilities struct {
	User         *User             `json:"user"`
	Capabilities []*UserCapability `json:"capabilities"`
}

// CapabilityConfig AI能力配置
type CapabilityConfig struct {
	CapabilityName string                 `json:"capability_name"`
	CapabilityType string                 `json:"capability_type"`
	Config         map[string]interface{} `json:"config"`
	Priority       int                    `json:"priority"`
	IsEnabled      bool                   `json:"is_enabled"`
}

// DeviceCapabilityConfig 设备AI能力配置
type DeviceCapabilityConfig struct {
	DeviceID      int64              `json:"device_id"`
	Capabilities  []CapabilityConfig `json:"capabilities"`
	GlobalConfigs map[string]string  `json:"global_configs"`
}

// UserCapabilityConfig 用户AI能力配置
type UserCapabilityConfig struct {
	UserID        int64              `json:"user_id"`
	Capabilities  []CapabilityConfig `json:"capabilities"`
	GlobalConfigs map[string]string  `json:"global_configs"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Nickname string `json:"nickname"`
}

// UserDeviceRequest 用户设备绑定请求
type UserDeviceRequest struct {
	DeviceUUID  string                 `json:"device_uuid" binding:"required"`
	DeviceAlias string                 `json:"device_alias"`
	IsOwner     bool                   `json:"is_owner"`
	Permissions map[string]interface{} `json:"permissions"`
}

// UserCapabilityRequest 用户AI能力配置请求
type UserCapabilityRequest struct {
	CapabilityName string                 `json:"capability_name" binding:"required"`
	CapabilityType string                 `json:"capability_type" binding:"required"`
	Config         map[string]interface{} `json:"config"`
}

// DeviceCapabilityRequest 设备AI能力配置请求
type DeviceCapabilityRequest struct {
	CapabilityName string                 `json:"capability_name" binding:"required"`
	CapabilityType string                 `json:"capability_type" binding:"required"`
	Priority       int                    `json:"priority"`
	Config         map[string]interface{} `json:"config"`
	IsEnabled      bool                   `json:"is_enabled"`
}

// CreateDeviceRequest 创建设备请求
type CreateDeviceRequest struct {
	OUI             string `json:"oui" binding:"required,len=8"`
	SN              string `json:"sn" binding:"required"`
	DeviceName      string `json:"device_name" binding:"required"`
	DeviceType      string `json:"device_type"`
	DeviceModel     string `json:"device_model"`
	FirmwareVersion string `json:"firmware_version"`
	HardwareVersion string `json:"hardware_version"`
}

// UpdateDeviceRequest 更新设备请求
type UpdateDeviceRequest struct {
	DeviceName      string `json:"device_name"`
	DeviceType      string `json:"device_type"`
	DeviceModel     string `json:"device_model"`
	FirmwareVersion string `json:"firmware_version"`
	HardwareVersion string `json:"hardware_version"`
	Status          string `json:"status"`
}

// AICapabilityRequest AI能力请求
type AICapabilityRequest struct {
	Name   string                 `json:"name" binding:"required"`
	Type   string                 `json:"type" binding:"required"`
	Config map[string]interface{} `json:"config"`
}

// DefaultAICapabilityRequest 默认AI能力请求
type DefaultAICapabilityRequest struct {
	CapabilityName string `json:"capability_name" binding:"required"`
	CapabilityType string `json:"capability_type" binding:"required"`
}
