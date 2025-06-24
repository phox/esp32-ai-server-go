package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"ai-server-go/src/core/types"
	"ai-server-go/src/core/utils"
)

type Message = types.Message

// DialogueManager 管理对话上下文和历史
type DialogueManager struct {
	logger   *utils.Logger
	dialogue []Message
	memory   MemoryInterface
	// 记忆相关配置
	memoryEnabled bool
	memoryLimit   int
}

// NewDialogueManager 创建对话管理器实例
func NewDialogueManager(logger *utils.Logger, memory MemoryInterface) *DialogueManager {
	return &DialogueManager{
		logger:        logger,
		dialogue:      make([]Message, 0),
		memory:        memory,
		memoryEnabled: true,
		memoryLimit:   5, // 默认返回5条记忆
	}
}

// SetMemoryEnabled 设置是否启用记忆功能
func (dm *DialogueManager) SetMemoryEnabled(enabled bool) {
	dm.memoryEnabled = enabled
}

// SetMemoryLimit 设置记忆查询限制
func (dm *DialogueManager) SetMemoryLimit(limit int) {
	if limit > 0 {
		dm.memoryLimit = limit
	}
}

func (dm *DialogueManager) SetSystemMessage(systemMessage string) {
	if systemMessage == "" {
		return
	}

	// 如果对话中已经有系统消息，则不再添加
	if len(dm.dialogue) > 0 && dm.dialogue[0].Role == "system" {
		dm.dialogue[0].Content = systemMessage
		return
	}

	// 添加新的系统消息到对话开头
	dm.dialogue = append([]Message{
		{Role: "system", Content: systemMessage},
	}, dm.dialogue...)
}

// Put 添加新消息到对话
func (dm *DialogueManager) Put(message Message) {
	dm.dialogue = append(dm.dialogue, message)
	
	// 如果启用了记忆功能，异步保存记忆
	if dm.memoryEnabled && dm.memory != nil {
		go func() {
			if err := dm.memory.SaveMemory(dm.dialogue); err != nil {
				dm.logger.Warn("保存对话记忆失败: %v", err)
			}
		}()
	}
}

// GetLLMDialogue 获取完整对话历史
func (dm *DialogueManager) GetLLMDialogue() []Message {
	return dm.dialogue
}

// GetLLMDialogueWithMemory 获取带记忆的对话
func (dm *DialogueManager) GetLLMDialogueWithMemory(memoryStr string) []Message {
	if memoryStr == "" {
		return dm.GetLLMDialogue()
	}

	memoryMsg := Message{
		Role:    "system",
		Content: memoryStr,
	}

	dialogue := make([]Message, 0, len(dm.dialogue)+1)
	dialogue = append(dialogue, memoryMsg)
	dialogue = append(dialogue, dm.dialogue...)

	return dialogue
}

// GetLLMDialogueWithAutoMemory 获取带自动记忆查询的对话
func (dm *DialogueManager) GetLLMDialogueWithAutoMemory(query string) []Message {
	if !dm.memoryEnabled || dm.memory == nil {
		return dm.GetLLMDialogue()
	}

	// 如果没有提供查询，使用最后一条用户消息
	if query == "" && len(dm.dialogue) > 0 {
		for i := len(dm.dialogue) - 1; i >= 0; i-- {
			if dm.dialogue[i].Role == "user" {
				query = dm.dialogue[i].Content
				break
			}
		}
	}

	// 查询相关记忆
	memoryStr, err := dm.memory.QueryMemory(query)
	if err != nil {
		dm.logger.Warn("查询记忆失败: %v", err)
		return dm.GetLLMDialogue()
	}

	if memoryStr == "" {
		return dm.GetLLMDialogue()
	}

	dm.logger.Debug("查询到相关记忆: %s", memoryStr)
	return dm.GetLLMDialogueWithMemory(memoryStr)
}

// GetLLMDialogueWithContext 获取带上下文的对话（智能记忆）
func (dm *DialogueManager) GetLLMDialogueWithContext(ctx context.Context, query string) []Message {
	if !dm.memoryEnabled || dm.memory == nil {
		return dm.GetLLMDialogue()
	}

	// 构建上下文查询
	contextQuery := dm.buildContextQuery(query)
	
	// 查询相关记忆
	memoryStr, err := dm.memory.QueryMemory(contextQuery)
	if err != nil {
		dm.logger.Warn("查询上下文记忆失败: %v", err)
		return dm.GetLLMDialogue()
	}

	if memoryStr == "" {
		return dm.GetLLMDialogue()
	}

	// 构建带记忆的对话
	memoryMsg := Message{
		Role:    "system",
		Content: fmt.Sprintf("基于之前的对话记忆，请参考以下信息：\n%s\n\n请根据这些记忆信息来回答用户的问题。", memoryStr),
	}

	dialogue := make([]Message, 0, len(dm.dialogue)+1)
	dialogue = append(dialogue, memoryMsg)
	dialogue = append(dialogue, dm.dialogue...)

	dm.logger.Debug("使用上下文记忆构建对话 %v", map[string]interface{}{
		"query":        query,
		"context_query": contextQuery,
		"memory_length": len(memoryStr),
		"dialogue_size": len(dialogue),
	})

	return dialogue
}

// buildContextQuery 构建上下文查询
func (dm *DialogueManager) buildContextQuery(query string) string {
	if query == "" {
		return ""
	}

	// 提取查询中的关键词
	keywords := dm.extractKeywords(query)
	if len(keywords) == 0 {
		return query
	}

	// 构建上下文查询
	contextQuery := fmt.Sprintf("用户询问关于 %s 的问题", strings.Join(keywords, " "))
	
	// 如果有历史对话，添加一些上下文
	if len(dm.dialogue) > 0 {
		contextQuery += "，参考之前的对话历史"
	}

	return contextQuery
}

// extractKeywords 提取关键词
func (dm *DialogueManager) extractKeywords(text string) []string {
	if text == "" {
		return nil
	}

	// 简单的关键词提取
	keywords := []string{}
	
	// 预定义的重要关键词
	importantWords := []string{"天气", "时间", "计算", "翻译", "编程", "代码", "文件", "图片", "音乐", "新闻", "帮助", "问题", "解决", "名字", "喜欢", "讨厌"}
	
	textLower := strings.ToLower(text)
	for _, word := range importantWords {
		if strings.Contains(textLower, strings.ToLower(word)) {
			keywords = append(keywords, word)
		}
	}

	// 如果没有找到预定义关键词，使用前几个词
	if len(keywords) == 0 {
		words := strings.Fields(text)
		if len(words) > 0 {
			keywords = words[:min(3, len(words))]
		}
	}

	return keywords
}

// Clear 清空对话历史
func (dm *DialogueManager) Clear() {
	dm.dialogue = make([]Message, 0)
	
	// 如果启用了记忆功能，清空记忆
	if dm.memoryEnabled && dm.memory != nil {
		go func() {
			if err := dm.memory.ClearMemory(); err != nil {
				dm.logger.Warn("清空记忆失败: %v", err)
			}
		}()
	}
}

// ToJSON 将对话历史转换为JSON字符串
func (dm *DialogueManager) ToJSON() (string, error) {
	bytes, err := json.Marshal(dm.dialogue)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// LoadFromJSON 从JSON字符串加载对话历史
func (dm *DialogueManager) LoadFromJSON(jsonStr string) error {
	return json.Unmarshal([]byte(jsonStr), &dm.dialogue)
}

// GetMemoryStats 获取记忆统计信息
func (dm *DialogueManager) GetMemoryStats() map[string]interface{} {
	stats := map[string]interface{}{
		"memory_enabled": dm.memoryEnabled,
		"memory_limit":   dm.memoryLimit,
		"dialogue_count": len(dm.dialogue),
		"has_memory":     dm.memory != nil,
	}
	return stats
}
