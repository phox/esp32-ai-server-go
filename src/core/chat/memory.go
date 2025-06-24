package chat

import (
	"context"
	"fmt"
	"strings"

	"ai-server-go/src/core/utils"
)

// MemoryInterface 定义对话记忆管理接口
type MemoryInterface interface {
	// QueryMemory 查询相关记忆
	QueryMemory(query string) (string, error)

	// SaveMemory 保存对话记忆
	SaveMemory(dialogue []Message) error

	// ClearMemory 清空记忆
	ClearMemory() error
}

// DatabaseMemory 数据库记忆实现
type DatabaseMemory struct {
	userID        *uint
	deviceID      uint
	sessionID     string
	memoryService interface {
		QueryMemory(userID *uint, deviceID uint, query string, limit int) (string, error)
		SaveMemory(userID *uint, deviceID uint, sessionID, memoryType, content string, importance int, tags []string) error
		GenerateMemoryFromDialogue(ctx context.Context, userID *uint, deviceID uint, sessionID string, dialogue []Message) error
		ClearMemory(sessionID string) error
	}
	logger *utils.Logger
}

// NewDatabaseMemory 创建数据库记忆实例
func NewDatabaseMemory(userID *uint, deviceID uint, sessionID string, memoryService interface {
	QueryMemory(userID *uint, deviceID uint, query string, limit int) (string, error)
	SaveMemory(userID *uint, deviceID uint, sessionID, memoryType, content string, importance int, tags []string) error
	GenerateMemoryFromDialogue(ctx context.Context, userID *uint, deviceID uint, sessionID string, dialogue []Message) error
	ClearMemory(sessionID string) error
}, logger *utils.Logger) *DatabaseMemory {
	return &DatabaseMemory{
		userID:        userID,
		deviceID:      deviceID,
		sessionID:     sessionID,
		memoryService: memoryService,
		logger:        logger,
	}
}

// QueryMemory 查询相关记忆
func (m *DatabaseMemory) QueryMemory(query string) (string, error) {
	if m.memoryService == nil {
		return "", fmt.Errorf("记忆服务未初始化")
	}

	// 从查询中提取关键词
	keywords := m.extractKeywords(query)
	if len(keywords) == 0 {
		return "", nil
	}

	// 使用关键词查询记忆
	memory, err := m.memoryService.QueryMemory(m.userID, m.deviceID, strings.Join(keywords, " "), 5)
	if err != nil {
		return "", fmt.Errorf("查询记忆失败: %v", err)
	}

	return memory, nil
}

// SaveMemory 保存对话记忆
func (m *DatabaseMemory) SaveMemory(dialogue []Message) error {
	if m.memoryService == nil {
		return fmt.Errorf("记忆服务未初始化")
	}

	// 异步生成记忆，避免阻塞对话流程
	go func() {
		ctx := context.Background()
		if err := m.memoryService.GenerateMemoryFromDialogue(ctx, m.userID, m.deviceID, m.sessionID, dialogue); err != nil {
			m.logger.Warn("生成对话记忆失败: %v", err)
		}
	}()

	return nil
}

// ClearMemory 清空记忆
func (m *DatabaseMemory) ClearMemory() error {
	if m.memoryService == nil {
		return fmt.Errorf("记忆服务未初始化")
	}

	return m.memoryService.ClearMemory(m.sessionID)
}

// extractKeywords 从查询中提取关键词
func (m *DatabaseMemory) extractKeywords(query string) []string {
	if query == "" {
		return nil
	}

	// 简单的关键词提取逻辑
	keywords := []string{}

	// 预定义的重要关键词
	importantWords := []string{"天气", "时间", "计算", "翻译", "编程", "代码", "文件", "图片", "音乐", "新闻", "帮助", "问题", "解决"}

	queryLower := strings.ToLower(query)
	for _, word := range importantWords {
		if strings.Contains(queryLower, strings.ToLower(word)) {
			keywords = append(keywords, word)
		}
	}

	// 如果没有找到预定义关键词，使用查询的前几个词
	if len(keywords) == 0 {
		words := strings.Fields(query)
		if len(words) > 0 {
			keywords = words[:min(3, len(words))]
		}
	}

	return keywords
}

// min 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// SimpleMemory 简单内存记忆实现（用于测试）
type SimpleMemory struct {
	memories map[string]string
	logger   *utils.Logger
}

// NewSimpleMemory 创建简单记忆实例
func NewSimpleMemory(logger *utils.Logger) *SimpleMemory {
	return &SimpleMemory{
		memories: make(map[string]string),
		logger:   logger,
	}
}

// QueryMemory 查询记忆
func (m *SimpleMemory) QueryMemory(query string) (string, error) {
	if memory, exists := m.memories[query]; exists {
		return memory, nil
	}
	return "", nil
}

// SaveMemory 保存记忆
func (m *SimpleMemory) SaveMemory(dialogue []Message) error {
	if len(dialogue) == 0 {
		return nil
	}

	// 简单的记忆保存逻辑
	lastMessage := dialogue[len(dialogue)-1]
	if lastMessage.Role == "user" {
		key := lastMessage.Content[:min(20, len(lastMessage.Content))]
		m.memories[key] = lastMessage.Content
		m.logger.Debug("保存简单记忆: %s", key)
	}

	return nil
}

// ClearMemory 清空记忆
func (m *SimpleMemory) ClearMemory() error {
	m.memories = make(map[string]string)
	m.logger.Debug("清空简单记忆")
	return nil
}
