package api

import (
	"net/http"
	"strconv"

	"ai-server-go/src/core/auth"
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
}

// NewUserAPI 创建用户管理API
func NewUserAPI(
	userService *database.UserService,
	deviceService *database.DeviceService,
	configService *database.ConfigService,
	authMiddleware *auth.AuthMiddleware,
	logger *utils.Logger,
) *UserAPI {
	return &UserAPI{
		userService:    userService,
		deviceService:  deviceService,
		configService:  configService,
		authMiddleware: authMiddleware,
		logger:         logger,
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
	}

	// AI能力管理路由
	capabilities := r.Group("/api/capabilities")
	capabilities.Use(userApi.authMiddleware.AuthRequired())
	{
		capabilities.GET("", userApi.ListCapabilities)
		capabilities.GET("/:name/:type", userApi.GetCapability)
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