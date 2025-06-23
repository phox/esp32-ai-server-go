# 灰度发布功能使用指南

## 概述

灰度发布功能允许为同一个 AI Provider 配置多个版本，并通过不同的策略（权重、健康、轮询）分配流量，实现平滑升级、A/B 测试和回滚。

## 主要特性
- 支持 Provider 多版本（如 v1/v2/v3），每个版本可独立配置
- 支持权重分流、健康优选、轮询等灰度策略
- 支持动态调整权重、热更新、健康检查
- 支持一键切换默认版本、禁用/启用版本

## 1. API接口

### 1.1 获取 Provider 版本列表
```bash
GET /api/configs/provider/{category}/{name}/versions
```
**响应示例：**
```json
{
  "success": true,
  "data": [
    {
      "category": "LLM",
      "name": "OllamaLLM",
      "version": "v1",
      "weight": 70,
      "is_active": true,
      "is_default": true,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    },
    {
      "category": "LLM",
      "name": "OllamaLLM",
      "version": "v2",
      "weight": 30,
      "is_active": true,
      "is_default": false,
      "created_at": "2024-01-02T00:00:00Z",
      "updated_at": "2024-01-02T00:00:00Z"
    }
  ]
}
```

### 1.2 获取灰度发布状态
```bash
GET /api/configs/provider/{category}/{name}/grayscale
```
**响应示例：**
```json
{
  "success": true,
  "data": {
    "category": "LLM",
    "name": "OllamaLLM",
    "strategy": "weight",
    "versions": [
      {
        "version": "v1",
        "weight": 70,
        "is_active": true,
        "is_default": true,
        "config": { ...ProviderConfig... }
      },
      {
        "version": "v2",
        "weight": 30,
        "is_active": true,
        "is_default": false,
        "config": { ...ProviderConfig... }
      }
    ]
  }
}
```

### 1.3 更新版本权重
```bash
PUT /api/configs/provider/{category}/{name}/weight
Content-Type: application/json

{
  "version": "v2",
  "weight": 50
}
```

### 1.4 设置默认版本
```bash
PUT /api/configs/provider/{category}/{name}/default
Content-Type: application/json

{
  "version": "v2"
}
```

### 1.5 刷新灰度配置
```bash
POST /api/configs/provider/{category}/{name}/refresh
```

## 2. 使用场景

### 2.1 新版本灰度发布
1. 创建新版本 Provider 配置（如 v2）
2. 设置新版本较小权重（如 10%）
3. 监控新版本表现
4. 逐步提升权重
5. 最终切换为默认版本

### 2.2 A/B 测试
1. 创建两个或多个版本
2. 设置权重（如 50%/50%）
3. 收集数据，动态调整权重

### 2.3 故障回滚
1. 监控健康状态
2. 降低或禁用故障版本权重
3. 切换默认版本

## 3. ProviderConfig 结构说明
```json
{
  "id": 1,
  "category": "llm",
  "name": "openai",
  "type": "openai",
  "version": "v1",
  "weight": 100,
  "is_active": true,
  "is_default": false,
  "props": {
    "api_key": "...",
    "model_name": "...",
    "base_url": "...",
    "temperature": 0.7,
    "max_tokens": 1000
  },
  "created_at": "...",
  "updated_at": "..."
}
```

## 4. 配置与操作示例

### 4.1 创建多版本 Provider 配置
```bash
# 创建 v1 版本
POST /api/configs/provider
Content-Type: application/json
{
  "category": "LLM",
  "name": "OllamaLLM",
  "version": "v1",
  "weight": 70,
  "is_active": true,
  "is_default": true,
  "type": "ollama",
  "props": {
    "model_name": "llama2",
    "base_url": "http://localhost:11434",
    "temperature": 0.7,
    "max_tokens": 2048,
    "top_p": 0.9
  }
}

# 创建 v2 版本
POST /api/configs/provider
Content-Type: application/json
{
  "category": "LLM",
  "name": "OllamaLLM",
  "version": "v2",
  "weight": 30,
  "is_active": true,
  "is_default": false,
  "type": "ollama",
  "props": {
    "model_name": "llama2:13b",
    "base_url": "http://localhost:11434",
    "temperature": 0.8,
    "max_tokens": 4096,
    "top_p": 0.95
  }
}
```

### 4.2 快速禁用故障版本
```bash
PUT /api/configs/provider/LLM/OllamaLLM/weight
Content-Type: application/json
{
  "version": "v2",
  "weight": 0
}
```

### 4.3 设置旧版本为默认
```bash
PUT /api/configs/provider/LLM/OllamaLLM/default
Content-Type: application/json
{
  "version": "v1"
}
```

### 4.4 刷新灰度配置缓存
```bash
POST /api/configs/provider/LLM/OllamaLLM/refresh
```

## 5. 策略说明
- **weight**：按权重分流（默认）
- **health**：按健康优选（预留）
- **round_robin**：轮询分流（预留）

## 6. 最佳实践
- 新版本初始权重建议 5-10%，逐步提升
- 监控健康状态，及时回滚
- 重要变更建议审批和记录
- 版本命名建议语义化（如 v1.0.0）或时间戳

---
通过灰度发布功能，您可以安全地进行 AI Provider 的版本升级和配置变更，确保服务的稳定性和可用性。 