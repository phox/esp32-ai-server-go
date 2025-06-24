package database

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"ai-server-go/src/core/chat"
	"ai-server-go/src/core/utils"

	"gorm.io/gorm"
)

// ChatMemoryService 聊天记忆服务
type ChatMemoryService struct {
	db     *gorm.DB
	logger *utils.Logger
}

// NewChatMemoryService 创建聊天记忆服务实例
func NewChatMemoryService(db *gorm.DB, logger *utils.Logger) *ChatMemoryService {
	return &ChatMemoryService{
		db:     db,
		logger: logger,
	}
}

// CreateSession 创建聊天会话
func (s *ChatMemoryService) CreateSession(userID *uint, deviceID uint, sessionID string, title string) (*ChatSession, error) {
	session := &ChatSession{
		UserID:       userID,
		DeviceID:     deviceID,
		SessionID:    sessionID,
		Title:        title,
		StartTime:    time.Now(),
		Status:       "active",
		MessageCount: 0,
	}

	if err := s.db.Create(session).Error; err != nil {
		return nil, fmt.Errorf("创建会话失败: %v", err)
	}

	s.logger.Info("创建聊天会话成功 %v", map[string]interface{}{
		"session_id": sessionID,
		"user_id":    userID,
		"device_id":  deviceID,
		"title":      title,
	})

	return session, nil
}

// GetSession 获取会话信息
func (s *ChatMemoryService) GetSession(sessionID string) (*ChatSession, error) {
	var session ChatSession
	if err := s.db.Where("session_id = ?", sessionID).First(&session).Error; err != nil {
		return nil, fmt.Errorf("获取会话失败: %v", err)
	}
	return &session, nil
}

// UpdateSession 更新会话信息
func (s *ChatMemoryService) UpdateSession(sessionID string, updates map[string]interface{}) error {
	if err := s.db.Model(&ChatSession{}).Where("session_id = ?", sessionID).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新会话失败: %v", err)
	}
	return nil
}

// SaveMessage 保存聊天消息
func (s *ChatMemoryService) SaveMessage(sessionID string, userID *uint, deviceID uint, role, content, messageType string, metadata map[string]interface{}) error {
	metadataJSON := ""
	if metadata != nil {
		if data, err := json.Marshal(metadata); err == nil {
			metadataJSON = string(data)
		}
	}

	message := &ChatMessage{
		SessionID:   sessionID,
		UserID:      userID,
		DeviceID:    deviceID,
		Role:        role,
		Content:     content,
		MessageType: messageType,
		Metadata:    metadataJSON,
		Timestamp:   time.Now(),
		IsProcessed: false,
	}

	if err := s.db.Create(message).Error; err != nil {
		return fmt.Errorf("保存消息失败: %v", err)
	}

	// 更新会话消息计数
	s.db.Model(&ChatSession{}).Where("session_id = ?", sessionID).UpdateColumn("message_count", gorm.Expr("message_count + 1"))

	return nil
}

// GetSessionMessages 获取会话消息历史
func (s *ChatMemoryService) GetSessionMessages(sessionID string, limit int) ([]ChatMessage, error) {
	var messages []ChatMessage
	query := s.db.Where("session_id = ?", sessionID).Order("timestamp ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("获取会话消息失败: %v", err)
	}

	return messages, nil
}

// SaveMemory 保存聊天记忆
func (s *ChatMemoryService) SaveMemory(userID *uint, deviceID uint, sessionID, memoryType, content string, importance int, tags []string) error {
	memory := &ChatMemory{
		UserID:     userID,
		DeviceID:   deviceID,
		SessionID:  sessionID,
		MemoryType: memoryType,
		Content:    content,
		Importance: importance,
		Tags:       strings.Join(tags, ","),
		LastUsed:   &[]time.Time{time.Now()}[0],
		UseCount:   0,
		IsActive:   true,
	}

	if err := s.db.Create(memory).Error; err != nil {
		return fmt.Errorf("保存记忆失败: %v", err)
	}

	s.logger.Info("保存聊天记忆成功 %v", map[string]interface{}{
		"session_id":  sessionID,
		"memory_type": memoryType,
		"importance":  importance,
		"tags":        tags,
	})

	return nil
}

// QueryMemory 查询相关记忆
func (s *ChatMemoryService) QueryMemory(userID *uint, deviceID uint, query string, limit int) (string, error) {
	if limit <= 0 {
		limit = 5 // 默认返回5条最相关的记忆
	}

	// 构建查询条件
	var memories []ChatMemory
	dbQuery := s.db.Where("is_active = ?", true)

	// 按用户和设备过滤
	if userID != nil {
		dbQuery = dbQuery.Where("(user_id = ? OR user_id IS NULL) AND device_id = ?", *userID, deviceID)
	} else {
		dbQuery = dbQuery.Where("device_id = ?", deviceID)
	}

	// 按重要性排序，优先返回重要的记忆
	dbQuery = dbQuery.Order("importance DESC, last_used DESC, use_count DESC").Limit(limit)

	if err := dbQuery.Find(&memories).Error; err != nil {
		return "", fmt.Errorf("查询记忆失败: %v", err)
	}

	if len(memories) == 0 {
		return "", nil
	}

	// 构建记忆摘要
	var memoryParts []string
	for _, memory := range memories {
		memoryParts = append(memoryParts, fmt.Sprintf("[%s] %s", memory.MemoryType, memory.Content))

		// 更新使用统计
		s.db.Model(&memory).Updates(map[string]interface{}{
			"use_count": gorm.Expr("use_count + 1"),
			"last_used": time.Now(),
		})
	}

	memorySummary := strings.Join(memoryParts, "\n")

	s.logger.Debug("查询到相关记忆 %v", map[string]interface{}{
		"query":        query,
		"memory_count": len(memories),
		"user_id":      userID,
		"device_id":    deviceID,
	})

	return memorySummary, nil
}

// GenerateMemoryFromDialogue 从对话历史生成记忆
func (s *ChatMemoryService) GenerateMemoryFromDialogue(ctx context.Context, userID *uint, deviceID uint, sessionID string, dialogue []chat.Message) error {
	if len(dialogue) == 0 {
		return nil
	}

	// 生成会话摘要
	summary := s.generateSummary(dialogue)
	if summary != "" {
		if err := s.SaveMemory(userID, deviceID, sessionID, "summary", summary, 8, []string{"auto_generated"}); err != nil {
			s.logger.Warn("保存会话摘要失败: %v", err)
		}
	}

	// 提取关键信息
	keyPoints := s.extractKeyPoints(dialogue)
	if len(keyPoints) > 0 {
		for _, point := range keyPoints {
			if err := s.SaveMemory(userID, deviceID, sessionID, "key_points", point, 6, []string{"auto_generated"}); err != nil {
				s.logger.Warn("保存关键信息失败: %v", err)
			}
		}
	}

	// 保存重要对话片段
	importantConversations := s.extractImportantConversations(dialogue)
	for _, conv := range importantConversations {
		if err := s.SaveMemory(userID, deviceID, sessionID, "conversation", conv, 5, []string{"auto_generated"}); err != nil {
			s.logger.Warn("保存重要对话失败: %v", err)
		}
	}

	return nil
}

// generateSummary 生成对话摘要
func (s *ChatMemoryService) generateSummary(dialogue []chat.Message) string {
	if len(dialogue) < 4 {
		return ""
	}

	// 简单的摘要生成逻辑
	var userMessages []string
	var assistantMessages []string

	for _, msg := range dialogue {
		if msg.Role == "user" {
			userMessages = append(userMessages, msg.Content)
		} else if msg.Role == "assistant" {
			assistantMessages = append(assistantMessages, msg.Content)
		}
	}

	if len(userMessages) == 0 || len(assistantMessages) == 0 {
		return ""
	}

	// 提取用户的主要问题或需求
	mainTopic := s.extractMainTopic(userMessages)
	if mainTopic == "" {
		return ""
	}

	return fmt.Sprintf("用户询问了关于%s的问题，进行了%d轮对话", mainTopic, len(userMessages))
}

// extractMainTopic 提取主要话题
func (s *ChatMemoryService) extractMainTopic(messages []string) string {
	if len(messages) == 0 {
		return ""
	}

	// 简单的关键词提取
	keywords := []string{"天气", "时间", "计算", "翻译", "编程", "代码", "文件", "图片", "音乐", "新闻"}

	for _, msg := range messages {
		for _, keyword := range keywords {
			if strings.Contains(msg, keyword) {
				return keyword
			}
		}
	}

	// 如果没有找到预定义关键词，返回第一个消息的前20个字符
	if len(messages[0]) > 20 {
		return messages[0][:20] + "..."
	}
	return messages[0]
}

// extractKeyPoints 提取关键信息
func (s *ChatMemoryService) extractKeyPoints(dialogue []chat.Message) []string {
	var keyPoints []string

	for _, msg := range dialogue {
		if msg.Role == "user" {
			// 提取用户的重要信息
			if strings.Contains(msg.Content, "我叫") || strings.Contains(msg.Content, "我的名字是") {
				keyPoints = append(keyPoints, "用户姓名信息: "+msg.Content)
			}
			if strings.Contains(msg.Content, "我喜欢") || strings.Contains(msg.Content, "我讨厌") {
				keyPoints = append(keyPoints, "用户偏好: "+msg.Content)
			}
		}
	}

	return keyPoints
}

// extractImportantConversations 提取重要对话片段
func (s *ChatMemoryService) extractImportantConversations(dialogue []chat.Message) []string {
	var conversations []string

	// 提取包含重要信息的对话片段
	for i := 0; i < len(dialogue)-1; i++ {
		if dialogue[i].Role == "user" && i+1 < len(dialogue) && dialogue[i+1].Role == "assistant" {
			userMsg := dialogue[i].Content
			assistantMsg := dialogue[i+1].Content

			// 判断是否为重要对话
			if len(userMsg) > 20 && len(assistantMsg) > 30 {
				conversation := fmt.Sprintf("用户: %s\n助手: %s", userMsg, assistantMsg)
				conversations = append(conversations, conversation)
			}
		}
	}

	return conversations
}

// ClearMemory 清空指定会话的记忆
func (s *ChatMemoryService) ClearMemory(sessionID string) error {
	if err := s.db.Where("session_id = ?", sessionID).Delete(&ChatMemory{}).Error; err != nil {
		return fmt.Errorf("清空记忆失败: %v", err)
	}

	s.logger.Info("清空会话记忆成功 %v", map[string]interface{}{
		"session_id": sessionID,
	})

	return nil
}

// GetMemoryStats 获取记忆统计信息
func (s *ChatMemoryService) GetMemoryStats(userID *uint, deviceID uint) (map[string]interface{}, error) {
	var stats map[string]interface{}

	// 统计不同类型的记忆数量
	var typeStats []struct {
		MemoryType string `json:"memory_type"`
		Count      int    `json:"count"`
	}

	query := s.db.Model(&ChatMemory{}).Where("is_active = ?", true)
	if userID != nil {
		query = query.Where("(user_id = ? OR user_id IS NULL) AND device_id = ?", *userID, deviceID)
	} else {
		query = query.Where("device_id = ?", deviceID)
	}

	if err := query.Select("memory_type, count(*) as count").Group("memory_type").Find(&typeStats).Error; err != nil {
		return nil, fmt.Errorf("获取记忆统计失败: %v", err)
	}

	stats = make(map[string]interface{})
	stats["type_stats"] = typeStats

	// 统计总记忆数
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("统计总记忆数失败: %v", err)
	}
	stats["total_count"] = totalCount

	return stats, nil
}
