package api

import (
	"net/http"
	"strconv"

	"ai-server-go/src/core/utils"
	"ai-server-go/src/database"

	"github.com/gin-gonic/gin"
)

// MemoryAPI 聊天记忆API
type MemoryAPI struct {
	memoryService *database.ChatMemoryService
	logger        *utils.Logger
}

// NewMemoryAPI 创建记忆API实例
func NewMemoryAPI(memoryService *database.ChatMemoryService, logger *utils.Logger) *MemoryAPI {
	return &MemoryAPI{
		memoryService: memoryService,
		logger:        logger,
	}
}

// RegisterRoutes 注册路由
func (api *MemoryAPI) RegisterRoutes(router *gin.Engine) {
	memoryGroup := router.Group("/api/memory")
	{
		memoryGroup.GET("/stats", api.GetMemoryStats)
		memoryGroup.GET("/sessions", api.GetSessions)
		memoryGroup.GET("/sessions/:sessionID", api.GetSession)
		memoryGroup.GET("/sessions/:sessionID/messages", api.GetSessionMessages)
		memoryGroup.GET("/sessions/:sessionID/memories", api.GetSessionMemories)
		memoryGroup.DELETE("/sessions/:sessionID", api.DeleteSession)
		memoryGroup.DELETE("/sessions/:sessionID/memories", api.ClearSessionMemories)
	}
}

// GetMemoryStats 获取记忆统计信息
func (api *MemoryAPI) GetMemoryStats(c *gin.Context) {
	userID := api.getUserID(c)
	deviceID := api.getDeviceID(c)

	if deviceID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "设备ID不能为空"})
		return
	}

	stats, err := api.memoryService.GetMemoryStats(userID, deviceID)
	if err != nil {
		api.logger.Error("获取记忆统计失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取记忆统计失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetSessions 获取会话列表
func (api *MemoryAPI) GetSessions(c *gin.Context) {
	userID := api.getUserID(c)
	deviceID := api.getDeviceID(c)
	limit := api.getIntParam(c, "limit", 20)
	offset := api.getIntParam(c, "offset", 0)

	if deviceID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "设备ID不能为空"})
		return
	}

	// 这里需要实现获取会话列表的方法
	// 暂时返回空列表，使用参数避免编译警告
	_ = userID
	_ = offset

	sessions := []gin.H{}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"sessions": sessions,
			"total":    len(sessions),
			"limit":    limit,
			"offset":   offset,
		},
	})
}

// GetSession 获取会话详情
func (api *MemoryAPI) GetSession(c *gin.Context) {
	sessionID := c.Param("sessionID")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "会话ID不能为空"})
		return
	}

	session, err := api.memoryService.GetSession(sessionID)
	if err != nil {
		api.logger.Error("获取会话失败: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "会话不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    session,
	})
}

// GetSessionMessages 获取会话消息
func (api *MemoryAPI) GetSessionMessages(c *gin.Context) {
	sessionID := c.Param("sessionID")
	limit := api.getIntParam(c, "limit", 50)

	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "会话ID不能为空"})
		return
	}

	messages, err := api.memoryService.GetSessionMessages(sessionID, limit)
	if err != nil {
		api.logger.Error("获取会话消息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取会话消息失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"messages": messages,
			"total":    len(messages),
		},
	})
}

// GetSessionMemories 获取会话记忆
func (api *MemoryAPI) GetSessionMemories(c *gin.Context) {
	sessionID := c.Param("sessionID")
	memoryType := c.Query("type")
	limit := api.getIntParam(c, "limit", 10)

	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "会话ID不能为空"})
		return
	}

	// 这里需要实现获取会话记忆的方法
	// 暂时返回空列表，使用参数避免编译警告
	_ = memoryType
	_ = limit

	memories := []gin.H{}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"memories": memories,
			"total":    len(memories),
		},
	})
}

// DeleteSession 删除会话
func (api *MemoryAPI) DeleteSession(c *gin.Context) {
	sessionID := c.Param("sessionID")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "会话ID不能为空"})
		return
	}

	// 更新会话状态为deleted
	updates := map[string]interface{}{
		"status":   "deleted",
		"end_time": "CURRENT_TIMESTAMP",
	}

	err := api.memoryService.UpdateSession(sessionID, updates)
	if err != nil {
		api.logger.Error("删除会话失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除会话失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "会话删除成功",
	})
}

// ClearSessionMemories 清空会话记忆
func (api *MemoryAPI) ClearSessionMemories(c *gin.Context) {
	sessionID := c.Param("sessionID")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "会话ID不能为空"})
		return
	}

	err := api.memoryService.ClearMemory(sessionID)
	if err != nil {
		api.logger.Error("清空会话记忆失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "清空会话记忆失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "会话记忆清空成功",
	})
}

// 辅助方法

// getUserID 从请求中获取用户ID
func (api *MemoryAPI) getUserID(c *gin.Context) *uint {
	// 从JWT token或请求头中获取用户ID
	// 暂时返回nil，表示匿名用户
	return nil
}

// getDeviceID 从请求中获取设备ID
func (api *MemoryAPI) getDeviceID(c *gin.Context) uint {
	deviceIDStr := c.Query("device_id")
	if deviceIDStr == "" {
		deviceIDStr = c.GetHeader("X-Device-ID")
	}

	if deviceIDStr == "" {
		return 0
	}

	deviceID, err := strconv.ParseUint(deviceIDStr, 10, 32)
	if err != nil {
		return 0
	}

	return uint(deviceID)
}

// getIntParam 获取整数参数
func (api *MemoryAPI) getIntParam(c *gin.Context, key string, defaultValue int) int {
	valueStr := c.Query(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}
