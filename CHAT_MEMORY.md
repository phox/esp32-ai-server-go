# 聊天记忆功能

## 概述

聊天记忆功能为LLM提供了持久化的对话记忆能力，使AI助手能够记住用户的偏好、重要信息和历史对话内容，提供更加个性化和连贯的对话体验。

## 功能特性

### 1. 多层级记忆管理
- **会话级记忆**: 每个对话会话独立管理记忆
- **设备级记忆**: 同一设备的所有会话共享记忆
- **用户级记忆**: 登录用户的所有设备共享记忆

### 2. 智能记忆类型
- **conversation**: 重要对话片段
- **summary**: 会话摘要
- **key_points**: 关键信息点（用户偏好、重要信息等）

### 3. 自动记忆生成
- 自动提取用户偏好（"我喜欢..."、"我讨厌..."）
- 自动识别用户身份信息（"我叫..."、"我的名字是..."）
- 自动生成会话摘要

### 4. 智能记忆查询
- 基于关键词的语义查询
- 按重要性排序的记忆检索
- 使用频率统计和更新

## 架构设计

### 核心组件

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   DialogueManager│    │   MemoryInterface│    │  ChatMemoryService│
│                 │    │                 │    │                 │
│ - 对话管理      │◄──►│ - 记忆查询      │◄──►│ - 数据库操作    │
│ - 记忆集成      │    │ - 记忆保存      │    │ - 会话管理      │
│ - 上下文构建    │    │ - 记忆清空      │    │ - 统计信息      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### 记忆实现

1. **DatabaseMemory**: 基于数据库的持久化记忆
2. **SimpleMemory**: 基于内存的简单记忆（用于测试）

## 数据库设计

### 主要表结构

#### chat_sessions (聊天会话表)
```sql
CREATE TABLE chat_sessions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NULL,                    -- 用户ID（可选）
    device_id BIGINT NOT NULL,              -- 设备ID
    session_id VARCHAR(100) NOT NULL UNIQUE, -- 会话标识符
    title VARCHAR(200),                     -- 会话标题
    summary TEXT,                           -- 会话摘要
    message_count INT DEFAULT 0,            -- 消息数量
    start_time TIMESTAMP NOT NULL,          -- 开始时间
    end_time TIMESTAMP NULL,                -- 结束时间
    status ENUM('active','archived','deleted'), -- 状态
    tags VARCHAR(500)                       -- 标签
);
```

#### chat_messages (聊天消息表)
```sql
CREATE TABLE chat_messages (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    session_id VARCHAR(100) NOT NULL,       -- 会话ID
    user_id BIGINT NULL,                    -- 用户ID
    device_id BIGINT NOT NULL,              -- 设备ID
    role ENUM('user','assistant','system'), -- 消息角色
    content TEXT NOT NULL,                  -- 消息内容
    message_type VARCHAR(20) DEFAULT 'text', -- 消息类型
    metadata TEXT,                          -- 元数据
    timestamp TIMESTAMP NOT NULL,           -- 时间戳
    is_processed BOOLEAN DEFAULT FALSE      -- 是否已处理
);
```

#### chat_memories (聊天记忆表)
```sql
CREATE TABLE chat_memories (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NULL,                    -- 用户ID
    device_id BIGINT NOT NULL,              -- 设备ID
    session_id VARCHAR(100) NOT NULL,       -- 会话ID
    memory_type VARCHAR(20) NOT NULL,       -- 记忆类型
    content TEXT NOT NULL,                  -- 记忆内容
    importance INT DEFAULT 1,               -- 重要性评分
    tags VARCHAR(500),                      -- 标签
    last_used TIMESTAMP NULL,               -- 最后使用时间
    use_count INT DEFAULT 0,                -- 使用次数
    is_active BOOLEAN DEFAULT TRUE          -- 是否激活
);
```

## 使用方法

### 1. 基本使用

```go
// 创建记忆服务
memoryService := database.NewChatMemoryService(db, logger)

// 创建记忆实例
memory := chat.NewDatabaseMemory(userID, deviceID, sessionID, memoryService, logger)

// 创建对话管理器
dialogueManager := chat.NewDialogueManager(logger, memory)

// 添加消息（自动保存记忆）
dialogueManager.Put(chat.Message{
    Role:    "user",
    Content: "我叫张三，我喜欢吃苹果",
})

// 使用带记忆的对话生成回复
messages := dialogueManager.GetLLMDialogueWithAutoMemory("你还记得我喜欢什么吗？")
```

### 2. 记忆查询

```go
// 查询相关记忆
memoryStr, err := memory.QueryMemory("张三")
if err != nil {
    log.Printf("查询记忆失败: %v", err)
} else {
    fmt.Printf("查询到的记忆: %s\n", memoryStr)
}
```

### 3. 记忆管理

```go
// 清空记忆
err := memory.ClearMemory()

// 获取记忆统计
stats := dialogueManager.GetMemoryStats()
```

## API接口

### 记忆统计
```
GET /api/memory/stats?device_id=123
```

### 会话管理
```
GET /api/memory/sessions?device_id=123&limit=20&offset=0
GET /api/memory/sessions/{sessionID}
DELETE /api/memory/sessions/{sessionID}
```

### 消息查询
```
GET /api/memory/sessions/{sessionID}/messages?limit=50
```

### 记忆管理
```
GET /api/memory/sessions/{sessionID}/memories?type=summary&limit=10
DELETE /api/memory/sessions/{sessionID}/memories
```

## 配置选项

### 系统配置

```sql
-- 记忆功能开关
INSERT INTO global_configs (config_key, config_value, config_type, description, is_system) VALUES
('memory.enabled', 'true', 'boolean', '是否启用聊天记忆功能', TRUE),
('memory.limit', '5', 'int', '记忆查询限制数量', TRUE),
('memory.auto_generate', 'true', 'boolean', '是否自动生成记忆', TRUE),
('memory.importance_threshold', '5', 'int', '记忆重要性阈值', TRUE);
```

### 代码配置

```go
// 设置记忆功能开关
dialogueManager.SetMemoryEnabled(true)

// 设置记忆查询限制
dialogueManager.SetMemoryLimit(5)
```

## 测试

运行记忆功能测试：

```bash
go run test_memory.go
```

测试将验证：
- 记忆的保存和查询
- 关键词提取
- 对话上下文构建
- 记忆清空功能

## 性能优化

### 1. 异步处理
- 记忆生成采用异步处理，避免阻塞对话流程
- 使用goroutine进行后台记忆处理

### 2. 缓存策略
- 重要记忆优先缓存
- 按使用频率进行记忆排序

### 3. 数据库优化
- 合理的索引设计
- 定期清理过期记忆
- 分页查询避免大量数据加载

## 扩展功能

### 1. 记忆分析
- 记忆使用模式分析
- 用户偏好趋势分析
- 记忆效果评估

### 2. 智能记忆
- 基于语义的记忆检索
- 记忆重要性自动评估
- 记忆关联性分析

### 3. 多模态记忆
- 图片记忆
- 音频记忆
- 视频记忆

## 注意事项

1. **隐私保护**: 记忆数据包含用户敏感信息，需要妥善保护
2. **存储限制**: 定期清理过期和低重要性记忆
3. **性能影响**: 大量记忆可能影响查询性能，需要合理限制
4. **数据一致性**: 确保记忆数据与对话历史的一致性

## 故障排除

### 常见问题

1. **记忆查询失败**
   - 检查数据库连接
   - 验证用户权限
   - 检查记忆服务初始化

2. **记忆不生效**
   - 确认记忆功能已启用
   - 检查记忆查询逻辑
   - 验证对话管理器配置

3. **性能问题**
   - 优化数据库查询
   - 减少记忆查询频率
   - 使用缓存机制 