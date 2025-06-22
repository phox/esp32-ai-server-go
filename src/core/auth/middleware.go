package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"ai-server-go/src/core/utils"
	"ai-server-go/src/database"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 认证中间件
type AuthMiddleware struct {
	userService *database.UserService
	logger      *utils.Logger
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware(userService *database.UserService, logger *utils.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		userService: userService,
		logger:      logger,
	}
}

// generateToken 生成随机Token
func (m *AuthMiddleware) generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// AuthRequired 需要认证的中间件
func (m *AuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Header获取Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "缺少认证信息",
			})
			c.Abort()
			return
		}

		// 解析Bearer Token
		token := ""
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			token = authHeader
		}

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "无效的认证Token",
			})
			c.Abort()
			return
		}

		// 验证Token
		userAuth, err := m.userService.GetUserAuthByKey(token)
		if err != nil {
			m.logger.Error("验证Token失败: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token验证失败",
			})
			c.Abort()
			return
		}

		if userAuth == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token无效或已过期",
			})
			c.Abort()
			return
		}

		// 获取用户信息
		user, err := m.userService.GetUserByID(userAuth.UserID)
		if err != nil {
			m.logger.Error("获取用户信息失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "获取用户信息失败",
			})
			c.Abort()
			return
		}

		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "用户不存在",
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user", user)
		c.Set("userAuth", userAuth)
		c.Next()
	}
}

// RoleRequired 需要特定角色的中间件
func (m *AuthMiddleware) RoleRequired(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先执行认证
		m.AuthRequired()(c)
		if c.IsAborted() {
			return
		}

		// 获取用户信息
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "用户信息获取失败",
			})
			c.Abort()
			return
		}

		userObj := user.(*database.User)

		// 检查用户角色
		hasRole := false
		for _, role := range roles {
			if userObj.Role == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error": fmt.Sprintf("需要角色权限: %v", roles),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AdminRequired 需要管理员权限的中间件
func (m *AuthMiddleware) AdminRequired() gin.HandlerFunc {
	return m.RoleRequired("admin")
}

// OptionalAuth 可选认证中间件（不强制要求登录）
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Header获取Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// 解析Bearer Token
		token := ""
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			token = authHeader
		}

		if token == "" {
			c.Next()
			return
		}

		// 验证Token
		userAuth, err := m.userService.GetUserAuthByKey(token)
		if err != nil {
			m.logger.Error("验证Token失败: %v", err)
			c.Next()
			return
		}

		if userAuth == nil {
			c.Next()
			return
		}

		// 获取用户信息
		user, err := m.userService.GetUserByID(userAuth.UserID)
		if err != nil {
			m.logger.Error("获取用户信息失败: %v", err)
			c.Next()
			return
		}

		if user == nil {
			c.Next()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user", user)
		c.Set("userAuth", userAuth)
		c.Next()
	}
}

// Login 用户登录
func (m *AuthMiddleware) Login(c *gin.Context) {
	var loginReq database.LoginRequest
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误",
		})
		return
	}

	// 验证用户名密码
	user, err := m.userService.AuthenticateUser(loginReq.Username, loginReq.Password)
	if err != nil {
		m.logger.Error("用户认证失败: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "用户名或密码错误",
		})
		return
	}

	// 生成Token
	token, err := m.generateToken()
	if err != nil {
		m.logger.Error("生成Token失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "生成Token失败",
		})
		return
	}

	// 创建用户认证记录
	expiresAt := time.Now().Add(24 * time.Hour) // 24小时过期
	userAuth := &database.UserAuth{
		UserID:    user.ID,
		AuthType:  "token",
		AuthKey:   token,
		IsActive:  true,
		ExpiresAt: &expiresAt,
	}

	// 直接创建认证记录，使用生成的token
	if err := m.userService.GetDB().DB.Create(userAuth).Error; err != nil {
		m.logger.Error("创建用户认证失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "创建认证失败",
		})
		return
	}

	// 更新登录信息
	clientIP := c.ClientIP()
	err = m.userService.UpdateLoginInfo(user.ID, clientIP)
	if err != nil {
		m.logger.Error("更新登录信息失败: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "登录成功",
		"data": gin.H{
			"token": token,
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"nickname": user.Nickname,
				"role":     user.Role,
			},
			"expires_at": userAuth.ExpiresAt,
		},
	})
}

// Logout 用户登出
func (m *AuthMiddleware) Logout(c *gin.Context) {
	userAuth, exists := c.Get("userAuth")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "未找到认证信息",
		})
		return
	}

	auth := userAuth.(*database.UserAuth)

	// 禁用Token
	if err := m.userService.GetDB().DB.Model(&database.UserAuth{}).Where("id = ?", auth.ID).Update("is_active", false).Error; err != nil {
		m.logger.Error("禁用Token失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "登出失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "登出成功",
	})
}

// GetCurrentUser 获取当前用户信息
func (m *AuthMiddleware) GetCurrentUser(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "用户未登录",
		})
		return
	}

	userObj := user.(*database.User)
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"id":         userObj.ID,
			"username":   userObj.Username,
			"email":      userObj.Email,
			"phone":      userObj.Phone,
			"nickname":   userObj.Nickname,
			"avatar":     userObj.Avatar,
			"status":     userObj.Status,
			"role":       userObj.Role,
			"created_at": userObj.CreatedAt,
		},
	})
}
