package database

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	gorm.Model
	Username      string     `json:"username" gorm:"uniqueIndex;size:50;not null"`
	Email         string     `json:"email" gorm:"uniqueIndex;size:100"`
	Phone         string     `json:"phone" gorm:"size:20"`
	PasswordHash  string     `json:"-" gorm:"size:255;not null"`
	Salt          string     `json:"-" gorm:"size:50"`
	Nickname      string     `json:"nickname" gorm:"size:50"`
	Avatar        string     `json:"avatar" gorm:"size:255"`
	Status        string     `json:"status" gorm:"size:20;default:'active'"`
	Role          string     `json:"role" gorm:"size:20;default:'user'"`
	LastLoginTime *time.Time `json:"last_login_time"`
	LastLoginIP   string     `json:"last_login_ip" gorm:"size:45"`

	// 关联关系
	UserAuths        []UserAuth       `json:"user_auths,omitempty" gorm:"foreignKey:UserID"`
	UserDevices      []UserDevice     `json:"user_devices,omitempty" gorm:"foreignKey:UserID"`
	UserCapabilities []UserCapability `json:"user_capabilities,omitempty" gorm:"foreignKey:UserID"`
	Sessions         []Session        `json:"sessions,omitempty" gorm:"foreignKey:UserID"`
	UsageStats       []UsageStats     `json:"usage_stats,omitempty" gorm:"foreignKey:UserID"`
}

// UserAuth 用户认证模型
type UserAuth struct {
	gorm.Model
	UserID     uint       `json:"user_id" gorm:"not null;index"`
	AuthType   string     `json:"auth_type" gorm:"size:20;not null"`
	AuthKey    string     `json:"auth_key" gorm:"size:255;not null;uniqueIndex"`
	AuthSecret string     `json:"auth_secret" gorm:"size:255"`
	IsActive   bool       `json:"is_active" gorm:"default:true"`
	ExpiresAt  *time.Time `json:"expires_at"`

	// 关联关系
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// UserDevice 用户设备绑定模型
type UserDevice struct {
	gorm.Model
	UserID      uint            `json:"user_id" gorm:"not null;index"`
	DeviceID    uint            `json:"device_id" gorm:"not null;index"`
	DeviceAlias string          `json:"device_alias" gorm:"size:100"`
	IsOwner     bool            `json:"is_owner" gorm:"default:false"`
	Permissions json.RawMessage `json:"permissions" gorm:"type:json"`
	IsActive    bool            `json:"is_active" gorm:"default:true"`

	// 关联关系
	Device Device `json:"device,omitempty" gorm:"foreignKey:DeviceID"`
	User   User   `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// UserCapability 用户AI能力模型
type UserCapability struct {
	gorm.Model
	UserID       uint            `json:"user_id" gorm:"not null;index"`
	CapabilityID uint            `json:"capability_id" gorm:"not null;index"`
	ConfigData   json.RawMessage `json:"config_data" gorm:"type:json"`
	IsActive     bool            `json:"is_active" gorm:"default:true"`

	// 关联关系
	Capability AICapability `json:"capability,omitempty" gorm:"foreignKey:CapabilityID"`
	User       User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// Device 设备模型
type Device struct {
	gorm.Model
	DeviceUUID      string     `json:"device_uuid" gorm:"uniqueIndex;size:100;not null"`
	OUI             string     `json:"oui" gorm:"size:8;not null"`
	SN              string     `json:"sn" gorm:"size:50;not null"`
	DeviceName      string     `json:"device_name" gorm:"size:100;not null"`
	DeviceType      string     `json:"device_type" gorm:"size:50"`
	DeviceModel     string     `json:"device_model" gorm:"size:50"`
	FirmwareVersion string     `json:"firmware_version" gorm:"size:20"`
	HardwareVersion string     `json:"hardware_version" gorm:"size:20"`
	Status          string     `json:"status" gorm:"size:20;default:'offline'"`
	LastOnlineTime  *time.Time `json:"last_online_time"`
	LastIPAddress   string     `json:"last_ip_address" gorm:"size:45"`

	// 关联关系
	DeviceAuths        []DeviceAuth       `json:"device_auths,omitempty" gorm:"foreignKey:DeviceID"`
	UserDevices        []UserDevice       `json:"user_devices,omitempty" gorm:"foreignKey:DeviceID"`
	DeviceCapabilities []DeviceCapability `json:"device_capabilities,omitempty" gorm:"foreignKey:DeviceID"`
	Sessions           []Session          `json:"sessions,omitempty" gorm:"foreignKey:DeviceID"`
	UsageStats         []UsageStats       `json:"usage_stats,omitempty" gorm:"foreignKey:DeviceID"`
}

// DeviceAuth 设备认证模型
type DeviceAuth struct {
	gorm.Model
	DeviceID   uint       `json:"device_id" gorm:"not null;index"`
	AuthType   string     `json:"auth_type" gorm:"size:20;not null"`
	AuthKey    string     `json:"auth_key" gorm:"size:255;not null;uniqueIndex"`
	AuthSecret string     `json:"auth_secret" gorm:"size:255"`
	IsActive   bool       `json:"is_active" gorm:"default:true"`
	ExpiresAt  *time.Time `json:"expires_at"`

	// 关联关系
	Device Device `json:"device,omitempty" gorm:"foreignKey:DeviceID"`
}

// AICapability AI能力模型
type AICapability struct {
	gorm.Model
	CapabilityName string          `json:"capability_name" gorm:"size:50;not null;uniqueIndex"`
	CapabilityType string          `json:"capability_type" gorm:"size:20;not null"`
	DisplayName    string          `json:"display_name" gorm:"size:100;not null"`
	Description    string          `json:"description" gorm:"size:500"`
	ConfigSchema   json.RawMessage `json:"config_schema" gorm:"type:json"`
	IsGlobal       bool            `json:"is_global" gorm:"default:false"`
	IsActive       bool            `json:"is_active" gorm:"default:true"`

	// 关联关系
	UserCapabilities   []UserCapability   `json:"user_capabilities,omitempty" gorm:"foreignKey:CapabilityID"`
	DeviceCapabilities []DeviceCapability `json:"device_capabilities,omitempty" gorm:"foreignKey:CapabilityID"`
}

// DeviceCapability 设备AI能力关联模型
type DeviceCapability struct {
	gorm.Model
	DeviceID     uint            `json:"device_id" gorm:"not null;index"`
	CapabilityID uint            `json:"capability_id" gorm:"not null;index"`
	Priority     int             `json:"priority" gorm:"default:0"`
	ConfigData   json.RawMessage `json:"config_data" gorm:"type:json"`
	IsEnabled    bool            `json:"is_enabled" gorm:"default:true"`

	// 关联关系
	Capability AICapability `json:"capability,omitempty" gorm:"foreignKey:CapabilityID"`
	Device     Device       `json:"device,omitempty" gorm:"foreignKey:DeviceID"`
}

// GlobalConfig 全局配置模型
type GlobalConfig struct {
	gorm.Model
	ConfigKey   string `json:"config_key" gorm:"size:100;not null;uniqueIndex"`
	ConfigValue string `json:"config_value" gorm:"type:text"`
	ConfigType  string `json:"config_type" gorm:"size:20;default:'string'"`
	Description string `json:"description" gorm:"size:500"`
	IsSystem    bool   `json:"is_system" gorm:"default:false"`
}

// SystemConfig 系统配置模型
type SystemConfig struct {
	gorm.Model
	ConfigCategory string `json:"config_category" gorm:"size:50;not null;index"`
	ConfigKey      string `json:"config_key" gorm:"size:100;not null;index"`
	ConfigValue    string `json:"config_value" gorm:"type:text"`
	ConfigType     string `json:"config_type" gorm:"size:20;default:'string'"`
	Description    string `json:"description" gorm:"size:500"`
	IsDefault      bool   `json:"is_default" gorm:"default:false"`
	CreatedBy      *uint  `json:"created_by"`
	UpdatedBy      *uint  `json:"updated_by"`

	// 关联关系
	Creator User `json:"creator,omitempty" gorm:"foreignKey:CreatedBy"`
	Updater User `json:"updater,omitempty" gorm:"foreignKey:UpdatedBy"`
}

// Session 会话模型
type Session struct {
	gorm.Model
	UserID       *uint      `json:"user_id" gorm:"index"`
	DeviceID     uint       `json:"device_id" gorm:"not null;index"`
	SessionID    string     `json:"session_id" gorm:"size:100;not null;uniqueIndex"`
	ClientID     string     `json:"client_id" gorm:"size:100;not null"`
	StartTime    time.Time  `json:"start_time" gorm:"not null"`
	EndTime      *time.Time `json:"end_time"`
	Status       string     `json:"status" gorm:"size:20;default:'active'"`
	MessageCount int        `json:"message_count" gorm:"default:0"`

	// 关联关系
	User   *User  `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Device Device `json:"device,omitempty" gorm:"foreignKey:DeviceID"`
}

// UsageStats 使用统计模型
type UsageStats struct {
	gorm.Model
	UserID         *uint     `json:"user_id" gorm:"index"`
	DeviceID       uint      `json:"device_id" gorm:"not null;index"`
	CapabilityName string    `json:"capability_name" gorm:"size:50;not null"`
	UsageDate      time.Time `json:"usage_date" gorm:"not null;index"`
	RequestCount   int       `json:"request_count" gorm:"default:0"`
	SuccessCount   int       `json:"success_count" gorm:"default:0"`
	ErrorCount     int       `json:"error_count" gorm:"default:0"`
	TotalDuration  int       `json:"total_duration" gorm:"default:0"`

	// 关联关系
	User   *User  `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Device Device `json:"device,omitempty" gorm:"foreignKey:DeviceID"`
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
	DeviceID      uint               `json:"device_id"`
	Capabilities  []CapabilityConfig `json:"capabilities"`
	GlobalConfigs map[string]string  `json:"global_configs"`
}

// UserCapabilityConfig 用户AI能力配置
type UserCapabilityConfig struct {
	UserID        uint               `json:"user_id"`
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
	Role     string `json:"role" binding:"required"`
}

// UserDeviceRequest 用户设备请求
type UserDeviceRequest struct {
	DeviceUUID  string                 `json:"device_uuid" binding:"required"`
	DeviceAlias string                 `json:"device_alias"`
	IsOwner     bool                   `json:"is_owner"`
	Permissions map[string]interface{} `json:"permissions"`
}

// DefaultAICapabilityRequest 默认AI能力请求
type DefaultAICapabilityRequest struct {
	CapabilityName string `json:"capability_name" binding:"required"`
	CapabilityType string `json:"capability_type" binding:"required"`
}

// UserCapabilityRequest 用户AI能力请求
type UserCapabilityRequest struct {
	CapabilityName string                 `json:"capability_name" binding:"required"`
	CapabilityType string                 `json:"capability_type" binding:"required"`
	Config         map[string]interface{} `json:"config"`
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

// DeviceCapabilityRequest 设备AI能力请求
type DeviceCapabilityRequest struct {
	CapabilityName string                 `json:"capability_name" binding:"required"`
	CapabilityType string                 `json:"capability_type" binding:"required"`
	Priority       int                    `json:"priority"`
	Config         map[string]interface{} `json:"config"`
	IsEnabled      bool                   `json:"is_enabled"`
}

// AICapabilityRequest AI能力请求
type AICapabilityRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Type        string                 `json:"type" binding:"required"`
	DisplayName string                 `json:"display_name" binding:"required"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config"`
	IsGlobal    bool                   `json:"is_global"`
}

// ProviderConfig AI能力配置
type ProviderConfig struct {
	gorm.Model
	Category  string          `json:"category" gorm:"size:20;not null;index"` // 类别（ASR/TTS/LLM/VLLLM）
	Name      string          `json:"name" gorm:"size:50;not null"`           // provider名称（如 EdgeTTS、OllamaLLM）
	Type      string          `json:"type" gorm:"size:20;not null"`           // provider类型（如 edge、ollama、openai等）
	Version   string          `json:"version" gorm:"size:20;default:'v1'"`    // 版本号（如 v1、v2）
	Weight    int             `json:"weight" gorm:"default:100"`              // 流量权重（0-100）
	IsActive  bool            `json:"is_active" gorm:"default:true"`          // 是否启用
	IsDefault bool            `json:"is_default" gorm:"default:false"`        // 是否为默认版本
	Props     json.RawMessage `json:"props" gorm:"type:json"`                 // 其他扩展参数
}

// ProviderVersion 封装了ProviderConfig部分字段，用于接口返回
type ProviderVersion struct {
	Category  string    `json:"category"`   // 类别（ASR/TTS/LLM/VLLLM）
	Name      string    `json:"name"`       // provider名称
	Version   string    `json:"version"`    // 版本号
	Weight    int       `json:"weight"`     // 流量权重
	IsActive  bool      `json:"is_active"`  // 是否启用
	IsDefault bool      `json:"is_default"` // 是否为默认版本
	CreatedAt time.Time `json:"created_at"` // 创建时间
	UpdatedAt time.Time `json:"updated_at"` // 更新时间
}

// UpdateProviderVersionRequest 更新Provider版本请求
type UpdateProviderVersionRequest struct {
	Weight    *int  `json:"weight"`
	IsActive  *bool `json:"is_active"`
	IsDefault *bool `json:"is_default"`
}

// 用户-Provider绑定
// category: TTS/ASR/LLM/VLLLM
// is_active: 是否启用
// provider_id: 关联ProviderConfig

type UserProvider struct {
	ID         uint   `gorm:"primaryKey"`
	UserID     uint   `gorm:"index"`
	ProviderID uint   `gorm:"index"`
	Category   string `gorm:"size:20"`
	IsActive   bool   `gorm:"default:true"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type DeviceProvider struct {
	ID         uint   `gorm:"primaryKey"`
	DeviceID   uint   `gorm:"index"`
	ProviderID uint   `gorm:"index"`
	Category   string `gorm:"size:20"`
	IsActive   bool   `gorm:"default:true"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// 保证GORM表名与schema一致
func (DeviceProvider) TableName() string {
	return "device_provider"
}
func (UserProvider) TableName() string {
	return "user_provider"
}
