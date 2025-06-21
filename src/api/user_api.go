package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"ai-server-go/src/core/auth"
	"ai-server-go/src/core/pool"
	"ai-server-go/src/core/utils"
	"ai-server-go/src/database"

	"github.com/gin-gonic/gin"
)

// UserAPI 用户管理API
type UserAPI struct {
	userService    *database.UserService
	deviceService  *database.DeviceService
	configService  *database.ConfigService
	authMiddleware *auth.AuthMiddleware
	logger         *utils.Logger
	poolManager    *pool.PoolManager
}

// NewUserAPI 创建用户管理API
func NewUserAPI(
	userService *database.UserService,
	deviceService *database.DeviceService,
	configService *database.ConfigService,
	authMiddleware *auth.AuthMiddleware,
	logger *utils.Logger,
	poolManager *pool.PoolManager,
) *UserAPI {
	return &UserAPI{
		userService:    userService,
		deviceService:  deviceService,
		configService:  configService,
		authMiddleware: authMiddleware,
		logger:         logger,
		poolManager:    poolManager,
	}
}

// RegisterRoutes 注册路由
func (userApi *UserAPI) RegisterRoutes(r *gin.Engine) {
	// 认证相关路由
	auth := r.Group("/api/auth")
	{
		auth.POST("/login", userApi.authMiddleware.Login)
		auth.POST("/logout", userApi.authMiddleware.AuthRequired(), userApi.authMiddleware.Logout)
		auth.GET("/me", userApi.authMiddleware.AuthRequired(), userApi.authMiddleware.GetCurrentUser)
	}

	// 用户管理路由
	users := r.Group("/api/users")
	users.Use(userApi.authMiddleware.AuthRequired())
	{
		// 用户CRUD
		users.GET("", userApi.ListUsers)
		users.POST("", userApi.authMiddleware.AdminRequired(), userApi.CreateUser)
		users.GET("/:id", userApi.GetUser)
		users.PUT("/:id", userApi.UpdateUser)
		users.DELETE("/:id", userApi.authMiddleware.AdminRequired(), userApi.DeleteUser)
		users.PUT("/:id/password", userApi.UpdatePassword)

		// 用户设备管理
		users.GET("/:id/devices", userApi.GetUserDevices)
		users.POST("/:id/devices", userApi.BindUserDevice)
		users.DELETE("/:id/devices/:deviceUUID", userApi.UnbindUserDevice)

		// 用户AI能力管理
		users.GET("/:id/capabilities", userApi.GetUserCapabilities)
		users.POST("/:id/capabilities", userApi.SetUserCapability)
		users.DELETE("/:id/capabilities/:capabilityName/:capabilityType", userApi.RemoveUserCapability)
	}

	// 设备管理路由
	devices := r.Group("/api/devices")
	devices.Use(userApi.authMiddleware.AuthRequired())
	{
		devices.GET("", userApi.ListDevices)
		devices.POST("", userApi.CreateDevice)
		devices.GET("/:id", userApi.GetDevice)
		devices.GET("/oui/:oui/sn/:sn", userApi.GetDeviceByOUISN)
		devices.PUT("/:id", userApi.UpdateDevice)
		devices.DELETE("/:id", userApi.DeleteDevice)

		// 设备AI能力配置
		devices.GET("/:id/capabilities", userApi.GetDeviceCapabilities)
		devices.POST("/:id/capabilities", userApi.SetDeviceCapability)
		devices.DELETE("/:id/capabilities/:capabilityName/:capabilityType", userApi.RemoveDeviceCapability)

		// 设备AI能力配置（带回退逻辑）
		devices.GET("/:id/capabilities/with-fallback", userApi.GetDeviceCapabilitiesWithFallback)
	}

	// AI能力管理路由
	capabilities := r.Group("/api/capabilities")
	capabilities.Use(userApi.authMiddleware.AuthRequired())
	{
		capabilities.GET("", userApi.ListCapabilities)
		capabilities.GET("/:name/:type", userApi.GetCapability)
		capabilities.POST("", userApi.authMiddleware.AdminRequired(), userApi.CreateCapability)
		capabilities.PUT("/:name/:type", userApi.authMiddleware.AdminRequired(), userApi.UpdateCapability)
		capabilities.DELETE("/:name/:type", userApi.authMiddleware.AdminRequired(), userApi.DeleteCapability)

		// 默认能力类型管理
		capabilities.GET("/defaults", userApi.GetDefaultCapabilities)
		capabilities.POST("/defaults", userApi.authMiddleware.AdminRequired(), userApi.SetDefaultCapability)
		capabilities.DELETE("/defaults/:capabilityName", userApi.authMiddleware.AdminRequired(), userApi.RemoveDefaultCapability)
	}

	// 系统配置管理路由（仅管理员）
	systemConfigs := r.Group("/api/system-configs")
	systemConfigs.Use(userApi.authMiddleware.AuthRequired(), userApi.authMiddleware.AdminRequired())
	{
		systemConfigs.GET("", userApi.GetSystemConfigs)
		systemConfigs.GET("/:category/:key", userApi.GetSystemConfig)
		systemConfigs.POST("", userApi.SetSystemConfig)
		systemConfigs.DELETE("/:category/:key", userApi.DeleteSystemConfig)
		systemConfigs.GET("/:category", userApi.GetSystemConfigCategory)
		systemConfigs.POST("/initialize", userApi.InitializeSystemConfigs)

		// Provider配置管理API
		systemConfigs.GET("/provider", userApi.ListProviderConfigs)
		systemConfigs.GET("/provider/:category/:name", userApi.GetProviderConfig)
		systemConfigs.POST("/provider", userApi.CreateProviderConfig)
		systemConfigs.PUT("/provider/:category/:name", userApi.UpdateProviderConfig)
		systemConfigs.DELETE("/provider/:category/:name", userApi.DeleteProviderConfig)

		// 灰度发布管理API
		systemConfigs.GET("/provider/:category/:name/versions", userApi.ListProviderVersions)
		systemConfigs.GET("/provider/:category/:name/grayscale", userApi.GetGrayscaleStatus)
		systemConfigs.PUT("/provider/:category/:name/weight", userApi.UpdateProviderWeight)
		systemConfigs.PUT("/provider/:category/:name/default", userApi.SetDefaultProviderVersion)
		systemConfigs.POST("/provider/:category/:name/refresh", userApi.RefreshGrayscaleConfig)
	}

	// 用户统计路由
	stats := r.Group("/api/stats")
	stats.Use(userApi.authMiddleware.AuthRequired())
	{
		stats.GET("/user", userApi.GetUserStats)
	}
}

// ListUsers 获取用户列表
func (userApi *UserAPI) ListUsers(c *gin.Context) {
	// 获取查询参数
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")
	role := c.Query("role")

	if limit > 100 {
		limit = 100
	}

	users, err := userApi.userService.ListUsers(offset, limit, status, role)
	if err != nil {
		userApi.logger.Error("获取用户列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取用户列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": users,
		"pagination": gin.H{
			"offset": offset,
			"limit":  limit,
			"total":  len(users),
		},
	})
}

// CreateUser 创建用户
func (userApi *UserAPI) CreateUser(c *gin.Context) {
	var req database.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误",
		})
		return
	}

	// 检查用户名是否已存在
	existingUser, err := userApi.userService.GetUserByUsername(req.Username)
	if err != nil {
		userApi.logger.Error("检查用户名失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "检查用户名失败",
		})
		return
	}

	if existingUser != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "用户名已存在",
		})
		return
	}

	// 创建用户
	user := &database.User{
		Username: req.Username,
		Email:    req.Email,
		Nickname: req.Nickname,
		Status:   "active",
		Role:     "user",
	}

	err = userApi.userService.CreateUser(user, req.Password)
	if err != nil {
		userApi.logger.Error("创建用户失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "创建用户失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "用户创建成功",
		"data":    user,
	})
}

// GetUser 获取用户信息
func (userApi *UserAPI) GetUser(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的用户ID",
		})
		return
	}

	user, err := userApi.userService.GetUserByID(userID)
	if err != nil {
		userApi.logger.Error("获取用户信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取用户信息失败",
		})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "用户不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": user,
	})
}

// UpdateUser 更新用户信息
func (userApi *UserAPI) UpdateUser(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的用户ID",
		})
		return
	}

	var req database.User
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误",
		})
		return
	}

	// 获取用户信息
	user, err := userApi.userService.GetUserByID(userID)
	if err != nil {
		userApi.logger.Error("获取用户信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取用户信息失败",
		})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "用户不存在",
		})
		return
	}

	// 更新用户信息
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.Status != "" {
		user.Status = req.Status
	}
	if req.Role != "" {
		user.Role = req.Role
	}

	err = userApi.userService.UpdateUser(user)
	if err != nil {
		userApi.logger.Error("更新用户信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "更新用户信息失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "用户信息更新成功",
		"data":    user,
	})
}

// UpdatePassword 更新用户密码
func (userApi *UserAPI) UpdatePassword(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的用户ID",
		})
		return
	}

	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误",
		})
		return
	}

	// 检查权限
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "用户未登录",
		})
		return
	}

	currentUserObj := currentUser.(*database.User)
	if currentUserObj.ID != userID && currentUserObj.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "权限不足",
		})
		return
	}

	// 如果不是管理员，需要验证旧密码
	if currentUserObj.Role != "admin" {
		user, err := userApi.userService.GetUserByID(userID)
		if err != nil {
			userApi.logger.Error("获取用户信息失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "获取用户信息失败",
			})
			return
		}

		// 验证旧密码
		_, err = userApi.userService.AuthenticateUser(user.Username, req.OldPassword)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "旧密码错误",
			})
			return
		}
	}

	// 更新密码
	err = userApi.userService.UpdatePassword(userID, req.NewPassword)
	if err != nil {
		userApi.logger.Error("更新密码失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "更新密码失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "密码更新成功",
	})
}

// DeleteUser 删除用户
func (userApi *UserAPI) DeleteUser(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的用户ID",
		})
		return
	}

	err = userApi.userService.DeleteUser(userID)
	if err != nil {
		userApi.logger.Error("删除用户失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "删除用户失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "用户删除成功",
	})
}

// GetUserDevices 获取用户设备列表
func (userApi *UserAPI) GetUserDevices(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的用户ID",
		})
		return
	}

	devices, err := userApi.userService.GetUserDevices(userID)
	if err != nil {
		userApi.logger.Error("获取用户设备失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取用户设备失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": devices,
	})
}

// BindUserDevice 绑定用户设备
func (userApi *UserAPI) BindUserDevice(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req database.UserDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误",
		})
		return
	}

	err := userApi.userService.BindUserDevice(userID, req.DeviceUUID, req.DeviceAlias, req.IsOwner, req.Permissions)
	if err != nil {
		userApi.logger.Error("绑定用户设备失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "绑定用户设备失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "设备绑定成功",
	})
}

// UnbindUserDevice 解绑用户设备
func (userApi *UserAPI) UnbindUserDevice(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	deviceUUID := c.Param("deviceUUID")

	err := userApi.userService.UnbindUserDevice(userID, deviceUUID)
	if err != nil {
		userApi.logger.Error("解绑用户设备失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "解绑用户设备失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "设备解绑成功",
	})
}

// GetUserCapabilities 获取用户AI能力配置
func (userApi *UserAPI) GetUserCapabilities(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的用户ID",
		})
		return
	}

	capabilities, err := userApi.userService.GetUserCapabilities(userID)
	if err != nil {
		userApi.logger.Error("获取用户AI能力失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取用户AI能力失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": capabilities,
	})
}

// SetUserCapability 设置用户AI能力配置
func (userApi *UserAPI) SetUserCapability(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的用户ID",
		})
		return
	}

	var req database.UserCapabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误",
		})
		return
	}

	err = userApi.userService.SetUserCapability(userID, req.CapabilityName, req.CapabilityType, req.Config)
	if err != nil {
		userApi.logger.Error("设置用户AI能力失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "设置用户AI能力失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "AI能力配置设置成功",
	})
}

// RemoveUserCapability 移除用户AI能力配置
func (userApi *UserAPI) RemoveUserCapability(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的用户ID",
		})
		return
	}

	capabilityName := c.Param("capabilityName")
	capabilityType := c.Param("capabilityType")

	// 查找AI能力
	capability, err := userApi.configService.GetAICapability(capabilityName, capabilityType)
	if err != nil {
		userApi.logger.Error("获取AI能力失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取AI能力失败",
		})
		return
	}

	if capability == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "AI能力不存在",
		})
		return
	}

	// 禁用用户AI能力
	query := `UPDATE user_capabilities SET is_active = false WHERE user_id = ? AND capability_id = ?`
	_, err = userApi.userService.GetDB().Exec(query, userID, capability.ID)
	if err != nil {
		userApi.logger.Error("移除用户AI能力失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "移除用户AI能力失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "AI能力移除成功",
	})
}

// ListDevices 获取设备列表
func (userApi *UserAPI) ListDevices(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")
	oui := c.Query("oui")

	if limit > 100 {
		limit = 100
	}

	devices, err := userApi.deviceService.ListDevices(offset, limit, status, oui)
	if err != nil {
		userApi.logger.Error("获取设备列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取设备列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": devices,
		"pagination": gin.H{
			"offset": offset,
			"limit":  limit,
			"total":  len(devices),
		},
	})
}

// CreateDevice 创建设备
func (userApi *UserAPI) CreateDevice(c *gin.Context) {
	var req database.CreateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误",
		})
		return
	}

	// 检查OUI和SN组合是否已存在
	existingDeviceByOUISN, err := userApi.deviceService.GetDeviceByOUISN(req.OUI, req.SN)
	if err != nil {
		userApi.logger.Error("检查OUI和SN失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "检查OUI和SN失败",
		})
		return
	}

	if existingDeviceByOUISN != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "OUI和SN组合已存在",
		})
		return
	}

	// 创建设备
	device := &database.Device{
		OUI:             req.OUI,
		SN:              req.SN,
		DeviceName:      req.DeviceName,
		DeviceType:      req.DeviceType,
		DeviceModel:     req.DeviceModel,
		FirmwareVersion: req.FirmwareVersion,
		HardwareVersion: req.HardwareVersion,
		Status:          "active",
	}

	err = userApi.deviceService.CreateDevice(device)
	if err != nil {
		userApi.logger.Error("创建设备失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "创建设备失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "设备创建成功",
		"data":    device,
	})
}

// GetDevice 获取设备信息
func (userApi *UserAPI) GetDevice(c *gin.Context) {
	deviceUUID := c.Param("id")

	device, err := userApi.deviceService.GetDeviceByUUID(deviceUUID)
	if err != nil {
		userApi.logger.Error("获取设备信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取设备信息失败",
		})
		return
	}

	if device == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "设备不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": device,
	})
}

// GetDeviceByOUISN 根据OUI和SN获取设备信息
func (userApi *UserAPI) GetDeviceByOUISN(c *gin.Context) {
	oui := c.Param("oui")
	sn := c.Param("sn")

	device, err := userApi.deviceService.GetDeviceByOUISN(oui, sn)
	if err != nil {
		userApi.logger.Error("获取设备信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取设备信息失败",
		})
		return
	}

	if device == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "设备不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": device,
	})
}

// UpdateDevice 更新设备信息
func (userApi *UserAPI) UpdateDevice(c *gin.Context) {
	deviceUUID := c.Param("id")

	var req database.UpdateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误",
		})
		return
	}

	// 获取现有设备信息
	device, err := userApi.deviceService.GetDeviceByUUID(deviceUUID)
	if err != nil {
		userApi.logger.Error("获取设备信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取设备信息失败",
		})
		return
	}

	if device == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "设备不存在",
		})
		return
	}

	// 更新设备信息
	if req.DeviceName != "" {
		device.DeviceName = req.DeviceName
	}
	if req.DeviceType != "" {
		device.DeviceType = req.DeviceType
	}
	if req.DeviceModel != "" {
		device.DeviceModel = req.DeviceModel
	}
	if req.FirmwareVersion != "" {
		device.FirmwareVersion = req.FirmwareVersion
	}
	if req.HardwareVersion != "" {
		device.HardwareVersion = req.HardwareVersion
	}
	if req.Status != "" {
		device.Status = req.Status
	}

	err = userApi.deviceService.UpdateDevice(device)
	if err != nil {
		userApi.logger.Error("更新设备信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "更新设备信息失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "设备信息更新成功",
		"data":    device,
	})
}

// DeleteDevice 删除设备
func (userApi *UserAPI) DeleteDevice(c *gin.Context) {
	deviceUUID := c.Param("id")

	// 先获取设备信息以获取数据库ID
	device, err := userApi.deviceService.GetDeviceByUUID(deviceUUID)
	if err != nil {
		userApi.logger.Error("获取设备信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取设备信息失败",
		})
		return
	}

	if device == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "设备不存在",
		})
		return
	}

	err = userApi.deviceService.DeleteDevice(device.ID)
	if err != nil {
		userApi.logger.Error("删除设备失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "删除设备失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "设备删除成功",
	})
}

// GetDeviceCapabilities 获取设备AI能力配置
func (userApi *UserAPI) GetDeviceCapabilities(c *gin.Context) {
	deviceUUID := c.Param("id")

	// 先获取设备信息以获取数据库ID
	device, err := userApi.deviceService.GetDeviceByUUID(deviceUUID)
	if err != nil {
		userApi.logger.Error("获取设备信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取设备信息失败",
		})
		return
	}

	if device == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "设备不存在",
		})
		return
	}

	// 获取设备能力配置
	capabilities, err := userApi.configService.GetDeviceCapabilities(device.ID)
	if err != nil {
		userApi.logger.Error("获取设备AI能力失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取设备AI能力失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": capabilities,
	})
}

// SetDeviceCapability 设置设备AI能力配置
func (userApi *UserAPI) SetDeviceCapability(c *gin.Context) {
	deviceUUID := c.Param("id")

	var req database.DeviceCapabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误",
		})
		return
	}

	// 先获取设备信息以获取数据库ID
	device, err := userApi.deviceService.GetDeviceByUUID(deviceUUID)
	if err != nil {
		userApi.logger.Error("获取设备信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取设备信息失败",
		})
		return
	}

	if device == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "设备不存在",
		})
		return
	}

	// 获取能力ID
	capability, err := userApi.configService.GetAICapabilityByName(req.CapabilityName, req.CapabilityType)
	if err != nil {
		userApi.logger.Error("获取AI能力信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取AI能力信息失败",
		})
		return
	}

	if capability == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "AI能力不存在",
		})
		return
	}

	err = userApi.configService.SetDeviceCapability(device.ID, capability.ID, req.Priority, req.Config, req.IsEnabled)
	if err != nil {
		userApi.logger.Error("设置设备AI能力失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "设置设备AI能力失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "设备AI能力配置设置成功",
	})
}

// RemoveDeviceCapability 移除设备AI能力配置
func (userApi *UserAPI) RemoveDeviceCapability(c *gin.Context) {
	deviceUUID := c.Param("id")
	capabilityName := c.Param("capabilityName")
	capabilityType := c.Param("capabilityType")

	// 先获取设备信息以获取数据库ID
	device, err := userApi.deviceService.GetDeviceByUUID(deviceUUID)
	if err != nil {
		userApi.logger.Error("获取设备信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取设备信息失败",
		})
		return
	}

	if device == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "设备不存在",
		})
		return
	}

	// 获取能力ID
	capability, err := userApi.configService.GetAICapabilityByName(capabilityName, capabilityType)
	if err != nil {
		userApi.logger.Error("获取AI能力信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取AI能力信息失败",
		})
		return
	}

	if capability == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "AI能力不存在",
		})
		return
	}

	err = userApi.configService.RemoveDeviceCapability(device.ID, capability.ID)
	if err != nil {
		userApi.logger.Error("移除设备AI能力失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "移除设备AI能力失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "设备AI能力移除成功",
	})
}

// GetDeviceCapabilitiesWithFallback 获取设备AI能力配置（带回退逻辑）
// 优先级：设备专属配置 > 用户自定义配置 > 系统默认配置
func (userApi *UserAPI) GetDeviceCapabilitiesWithFallback(c *gin.Context) {
	deviceUUID := c.Param("id")

	// 获取当前用户ID（从认证中间件中获取）
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "用户未认证",
		})
		return
	}
	userObj := user.(*database.User)
	userID := &userObj.ID

	// 获取设备能力配置（带回退逻辑）
	config, err := userApi.configService.GetDeviceCapabilityConfigWithFallback(deviceUUID, userID)
	if err != nil {
		userApi.logger.Error("获取设备能力配置失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取设备能力配置失败: " + err.Error(),
		})
		return
	}

	// 构建响应数据，包含优先级信息
	type CapabilityWithPriority struct {
		database.CapabilityConfig
		PrioritySource string `json:"priority_source"` // "device", "user", "system"
	}

	capabilitiesWithPriority := make([]CapabilityWithPriority, 0)
	for _, cap := range config.Capabilities {
		prioritySource := "device"
		if cap.Priority == 100 {
			prioritySource = "user"
		} else if cap.Priority == 200 {
			prioritySource = "system"
		}

		capWithPriority := CapabilityWithPriority{
			CapabilityConfig: cap,
			PrioritySource:   prioritySource,
		}
		capabilitiesWithPriority = append(capabilitiesWithPriority, capWithPriority)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"device_id":      config.DeviceID,
			"capabilities":   capabilitiesWithPriority,
			"global_configs": config.GlobalConfigs,
		},
	})
}

// ListCapabilities 获取AI能力列表
func (userApi *UserAPI) ListCapabilities(c *gin.Context) {
	capabilityType := c.Query("type")
	var enabled *bool
	if enabledStr := c.Query("enabled"); enabledStr != "" {
		if enabledStr == "true" {
			enabled = &[]bool{true}[0]
		} else if enabledStr == "false" {
			enabled = &[]bool{false}[0]
		}
	}

	capabilities, err := userApi.configService.ListAICapabilities(capabilityType, enabled)
	if err != nil {
		userApi.logger.Error("获取AI能力列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取AI能力列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": capabilities,
	})
}

// GetCapability 获取AI能力详情
func (userApi *UserAPI) GetCapability(c *gin.Context) {
	capabilityName := c.Param("name")
	capabilityType := c.Param("type")

	capability, err := userApi.configService.GetAICapability(capabilityName, capabilityType)
	if err != nil {
		userApi.logger.Error("获取AI能力详情失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取AI能力详情失败",
		})
		return
	}

	if capability == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "AI能力不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": capability,
	})
}

// CreateCapability 创建AI能力
func (userApi *UserAPI) CreateCapability(c *gin.Context) {
	var req database.AICapabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误",
		})
		return
	}

	// 创建AI能力对象
	capability := &database.AICapability{
		CapabilityName: req.Name,
		CapabilityType: req.Type,
		DisplayName:    req.Name,
		Description:    fmt.Sprintf("%s %s 能力", req.Name, req.Type),
		IsGlobal:       false,
		IsActive:       true,
	}

	// 如果有配置数据，序列化为JSON
	if req.Config != nil {
		configJSON, err := json.Marshal(req.Config)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "配置数据格式错误",
			})
			return
		}
		capability.ConfigSchema = configJSON
	}

	err := userApi.configService.CreateAICapability(capability)
	if err != nil {
		userApi.logger.Error("创建AI能力失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "创建AI能力失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "AI能力创建成功",
	})
}

// UpdateCapability 更新AI能力
func (userApi *UserAPI) UpdateCapability(c *gin.Context) {
	capabilityName := c.Param("name")
	capabilityType := c.Param("type")

	var req database.AICapabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误",
		})
		return
	}

	// 获取现有能力
	capability, err := userApi.configService.GetAICapability(capabilityName, capabilityType)
	if err != nil {
		userApi.logger.Error("获取AI能力失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取AI能力失败",
		})
		return
	}

	if capability == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "AI能力不存在",
		})
		return
	}

	// 更新字段
	if req.Name != "" {
		capability.DisplayName = req.Name
	}
	if req.Config != nil {
		configJSON, err := json.Marshal(req.Config)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "配置数据格式错误",
			})
			return
		}
		capability.ConfigSchema = configJSON
	}

	err = userApi.configService.UpdateAICapability(capability)
	if err != nil {
		userApi.logger.Error("更新AI能力失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "更新AI能力失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "AI能力更新成功",
	})
}

// DeleteCapability 删除AI能力
func (userApi *UserAPI) DeleteCapability(c *gin.Context) {
	capabilityName := c.Param("name")
	capabilityType := c.Param("type")

	err := userApi.configService.DeleteAICapability(capabilityName, capabilityType)
	if err != nil {
		userApi.logger.Error("删除AI能力失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "删除AI能力失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "AI能力删除成功",
	})
}

// GetDefaultCapabilities 获取默认AI能力列表
func (userApi *UserAPI) GetDefaultCapabilities(c *gin.Context) {
	capabilities, err := userApi.configService.ListDefaultAICapabilities()
	if err != nil {
		userApi.logger.Error("获取默认AI能力列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取默认AI能力列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": capabilities,
	})
}

// SetDefaultCapability 设置默认AI能力
func (userApi *UserAPI) SetDefaultCapability(c *gin.Context) {
	var req database.DefaultAICapabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误",
		})
		return
	}

	err := userApi.configService.SetDefaultAICapability(req.CapabilityName, req.CapabilityType)
	if err != nil {
		userApi.logger.Error("设置默认AI能力失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "设置默认AI能力失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "默认AI能力设置成功",
	})
}

// RemoveDefaultCapability 移除默认AI能力
func (userApi *UserAPI) RemoveDefaultCapability(c *gin.Context) {
	capabilityName := c.Param("capabilityName")

	err := userApi.configService.RemoveDefaultAICapability(capabilityName)
	if err != nil {
		userApi.logger.Error("移除默认AI能力失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "移除默认AI能力失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "默认AI能力移除成功",
	})
}

// GetUserStats 获取用户统计信息
func (api *UserAPI) GetUserStats(c *gin.Context) {
	userID := api.getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	stats, err := api.userService.GetUserStats(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("获取用户统计失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// ==================== 系统配置管理API ====================

// GetSystemConfigs 获取系统配置列表
func (api *UserAPI) GetSystemConfigs(c *gin.Context) {
	// 检查管理员权限
	if !api.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	category := c.Query("category")
	isDefaultStr := c.Query("is_default")

	var isDefault *bool
	if isDefaultStr != "" {
		if val, err := strconv.ParseBool(isDefaultStr); err == nil {
			isDefault = &val
		}
	}

	configs, err := api.configService.ListSystemConfigs(category, isDefault)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("获取系统配置失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    configs,
	})
}

// GetSystemConfig 获取单个系统配置
func (api *UserAPI) GetSystemConfig(c *gin.Context) {
	// 检查管理员权限
	if !api.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	category := c.Param("category")
	key := c.Param("key")

	config, err := api.configService.GetSystemConfig(category, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("获取系统配置失败: %v", err)})
		return
	}

	if config == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "系统配置不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    config,
	})
}

// SetSystemConfig 设置系统配置
func (api *UserAPI) SetSystemConfig(c *gin.Context) {
	// 检查管理员权限
	if !api.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	var req struct {
		Category    string `json:"category" binding:"required"`
		Key         string `json:"key" binding:"required"`
		Value       string `json:"value" binding:"required"`
		ConfigType  string `json:"config_type" binding:"required"`
		Description string `json:"description"`
		IsDefault   bool   `json:"is_default"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("请求参数错误: %v", err)})
		return
	}

	userID := api.getUserIDFromContext(c)
	err := api.configService.SetSystemConfig(req.Category, req.Key, req.Value, req.ConfigType, req.Description, req.IsDefault, &userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("设置系统配置失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "系统配置设置成功",
	})
}

// DeleteSystemConfig 删除系统配置
func (api *UserAPI) DeleteSystemConfig(c *gin.Context) {
	// 检查管理员权限
	if !api.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	category := c.Param("category")
	key := c.Param("key")

	err := api.configService.DeleteSystemConfig(category, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("删除系统配置失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "系统配置删除成功",
	})
}

// GetSystemConfigCategory 获取指定分类的所有系统配置
func (api *UserAPI) GetSystemConfigCategory(c *gin.Context) {
	// 检查管理员权限
	if !api.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	category := c.Param("category")

	configs, err := api.configService.GetSystemConfigCategory(category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("获取系统配置分类失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    configs,
	})
}

// InitializeSystemConfigs 初始化默认系统配置
func (api *UserAPI) InitializeSystemConfigs(c *gin.Context) {
	// 检查管理员权限
	if !api.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	err := api.configService.InitializeDefaultSystemConfigs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("初始化系统配置失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "默认系统配置初始化成功",
	})
}

// getUserIDFromContext 从上下文获取用户ID
func (api *UserAPI) getUserIDFromContext(c *gin.Context) int64 {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(int64); ok {
			return id
		}
	}
	return 0
}

// isAdmin 检查当前用户是否为管理员
func (api *UserAPI) isAdmin(c *gin.Context) bool {
	if role, exists := c.Get("user_role"); exists {
		if userRole, ok := role.(string); ok {
			return userRole == "admin"
		}
	}
	return false
}

// ListProviderConfigs 获取提供者配置列表
func (userApi *UserAPI) ListProviderConfigs(c *gin.Context) {
	category := c.Query("category")

	var configs []database.ProviderConfig
	var err error

	if category != "" {
		configs, err = userApi.configService.ListProviderConfigs(category)
	} else {
		// 获取所有类别的配置
		allConfigs := make([]database.ProviderConfig, 0)
		categories := []string{"ASR", "TTS", "LLM", "VLLLM"}
		for _, cat := range categories {
			catConfigs, _ := userApi.configService.ListProviderConfigs(cat)
			allConfigs = append(allConfigs, catConfigs...)
		}
		configs = allConfigs
	}

	if err != nil {
		userApi.logger.Error("获取提供者配置列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取提供者配置列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  configs,
		"total": len(configs),
	})
}

// GetProviderConfig 获取单个提供者配置
func (userApi *UserAPI) GetProviderConfig(c *gin.Context) {
	category := c.Param("category")
	name := c.Param("name")

	config, err := userApi.configService.GetProviderConfig(category, name)
	if err != nil {
		userApi.logger.Error("获取提供者配置失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取提供者配置失败",
		})
		return
	}

	if config == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "提供者配置不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": config,
	})
}

// CreateProviderConfig 创建提供者配置
func (userApi *UserAPI) CreateProviderConfig(c *gin.Context) {
	var config database.ProviderConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误: " + err.Error(),
		})
		return
	}
	category := c.Query("category")
	if category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "类别参数为必需"})
		return
	}
	if config.Name == "" || config.Type == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "名称和类型为必需字段"})
		return
	}
	existing, _ := userApi.configService.GetProviderConfig(category, config.Name)
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "提供者配置已存在"})
		return
	}
	err := userApi.configService.SetProviderConfig(category, config.Name, &config)
	if err != nil {
		userApi.logger.Error("创建提供者配置失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建提供者配置失败"})
		return
	}
	if userApi.poolManager != nil {
		_ = userApi.poolManager.ReloadProviderConfig(category, config.Name)
	}
	c.JSON(http.StatusCreated, gin.H{"message": "提供者配置创建成功", "data": config})
}

// UpdateProviderConfig 更新提供者配置
func (userApi *UserAPI) UpdateProviderConfig(c *gin.Context) {
	category := c.Param("category")
	name := c.Param("name")
	var config database.ProviderConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}
	existing, err := userApi.configService.GetProviderConfig(category, name)
	if err != nil {
		userApi.logger.Error("检查提供者配置失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "检查提供者配置失败"})
		return
	}
	if existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "提供者配置不存在"})
		return
	}
	config.Name = name
	err = userApi.configService.SetProviderConfig(category, name, &config)
	if err != nil {
		userApi.logger.Error("更新提供者配置失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新提供者配置失败"})
		return
	}
	if userApi.poolManager != nil {
		_ = userApi.poolManager.ReloadProviderConfig(category, name)
	}
	c.JSON(http.StatusOK, gin.H{"message": "提供者配置更新成功", "data": config})
}

// DeleteProviderConfig 删除提供者配置
func (userApi *UserAPI) DeleteProviderConfig(c *gin.Context) {
	category := c.Param("category")
	name := c.Param("name")
	existing, err := userApi.configService.GetProviderConfig(category, name)
	if err != nil {
		userApi.logger.Error("检查提供者配置失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "检查提供者配置失败"})
		return
	}
	if existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "提供者配置不存在"})
		return
	}
	err = userApi.configService.DeleteProviderConfig(category, name)
	if err != nil {
		userApi.logger.Error("删除提供者配置失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除提供者配置失败"})
		return
	}
	if userApi.poolManager != nil {
		_ = userApi.poolManager.ReloadProviderConfig(category, name)
	}
	c.JSON(http.StatusOK, gin.H{"message": "提供者配置删除成功"})
}

// ListProviderVersions 获取provider的所有版本
func (userApi *UserAPI) ListProviderVersions(c *gin.Context) {
	category := c.Param("category")
	name := c.Param("name")

	versions, err := userApi.configService.ListProviderVersions(category, name)
	if err != nil {
		userApi.logger.Error("获取provider版本列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取provider版本列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  versions,
		"total": len(versions),
	})
}

// GetGrayscaleStatus 获取灰度发布状态
func (userApi *UserAPI) GetGrayscaleStatus(c *gin.Context) {
	category := c.Param("category")
	name := c.Param("name")

	if userApi.poolManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "灰度发布管理器未初始化",
		})
		return
	}

	grayscaleManager := userApi.poolManager.GetGrayscaleManager()
	if grayscaleManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "灰度发布管理器未初始化",
		})
		return
	}

	status, err := grayscaleManager.GetGrayscaleStatus(category, name)
	if err != nil {
		userApi.logger.Error("获取灰度发布状态失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取灰度发布状态失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": status,
	})
}

// UpdateProviderWeight 更新provider版本权重
func (userApi *UserAPI) UpdateProviderWeight(c *gin.Context) {
	category := c.Param("category")
	name := c.Param("name")

	var req struct {
		Version string `json:"version" binding:"required"`
		Weight  int    `json:"weight" binding:"required,min=0,max=100"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误: " + err.Error(),
		})
		return
	}

	if userApi.poolManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "灰度发布管理器未初始化",
		})
		return
	}

	grayscaleManager := userApi.poolManager.GetGrayscaleManager()
	if grayscaleManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "灰度发布管理器未初始化",
		})
		return
	}

	err := grayscaleManager.UpdateWeight(category, name, req.Version, req.Weight)
	if err != nil {
		userApi.logger.Error("更新provider权重失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "更新provider权重失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "provider权重更新成功",
	})
}

// SetDefaultProviderVersion 设置默认provider版本
func (userApi *UserAPI) SetDefaultProviderVersion(c *gin.Context) {
	category := c.Param("category")
	name := c.Param("name")

	var req struct {
		Version string `json:"version" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误: " + err.Error(),
		})
		return
	}

	err := userApi.configService.SetDefaultProviderVersion(category, name, req.Version)
	if err != nil {
		userApi.logger.Error("设置默认provider版本失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "设置默认provider版本失败",
		})
		return
	}

	// 刷新灰度配置缓存
	if userApi.poolManager != nil {
		grayscaleManager := userApi.poolManager.GetGrayscaleManager()
		if grayscaleManager != nil {
			_ = grayscaleManager.RefreshConfig(category, name)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "默认provider版本设置成功",
	})
}

// RefreshGrayscaleConfig 刷新灰度配置缓存
func (userApi *UserAPI) RefreshGrayscaleConfig(c *gin.Context) {
	category := c.Param("category")
	name := c.Param("name")

	if userApi.poolManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "灰度发布管理器未初始化",
		})
		return
	}

	grayscaleManager := userApi.poolManager.GetGrayscaleManager()
	if grayscaleManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "灰度发布管理器未初始化",
		})
		return
	}

	err := grayscaleManager.RefreshConfig(category, name)
	if err != nil {
		userApi.logger.Error("刷新灰度配置失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "刷新灰度配置失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "灰度配置刷新成功",
	})
}
