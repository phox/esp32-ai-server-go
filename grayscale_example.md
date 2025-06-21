# 灰度发布功能使用指南

## 概述

灰度发布功能允许您为同一个 AI Provider 配置多个版本，并通过不同的策略（权重、健康评分、轮询）来分配流量，实现平滑的版本升级和回滚。

## 功能特性

### 1. 多版本支持
- 同一 provider 可以有多个版本（如 v1、v2、v3）
- 每个版本可以有不同的配置参数
- 支持版本权重设置

### 2. 流量分配策略
- **权重策略（weight）**：根据权重比例分配流量
- **健康策略（health）**：根据健康评分选择最佳版本
- **轮询策略（round_robin）**：轮流使用各个版本

### 3. 健康检查
- 自动检测各版本的健康状态
- 动态调整健康评分
- 支持故障自动切换

### 4. 动态管理
- 实时调整流量权重
- 热更新配置
- 支持版本切换

## API 接口

### 1. 获取 Provider 版本列表
```bash
GET /api/system-configs/provider/{category}/{name}/versions
```

**响应示例：**
```json
{
  "data": [
    {
      "category": "LLM",
      "name": "OllamaLLM",
      "version": "v1",
      "weight": 70,
      "is_active": true,
      "is_default": true,
      "health_score": 95.5,
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
      "health_score": 88.2,
      "created_at": "2024-01-02T00:00:00Z",
      "updated_at": "2024-01-02T00:00:00Z"
    }
  ],
  "total": 2
}
```

### 2. 获取灰度发布状态
```bash
GET /api/system-configs/provider/{category}/{name}/grayscale
```

**响应示例：**
```json
{
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
        "health_score": 95.5,
        "config": {
          "id": 1,
          "category": "LLM",
          "name": "OllamaLLM",
          "version": "v1",
          "type": "ollama",
          "model_name": "llama2",
          "url": "http://localhost:11434",
          "temperature": 0.7,
          "max_tokens": 2048,
          "top_p": 0.9
        }
      },
      {
        "version": "v2",
        "weight": 30,
        "is_active": true,
        "is_default": false,
        "health_score": 88.2,
        "config": {
          "id": 2,
          "category": "LLM",
          "name": "OllamaLLM",
          "version": "v2",
          "type": "ollama",
          "model_name": "llama2:13b",
          "url": "http://localhost:11434",
          "temperature": 0.8,
          "max_tokens": 4096,
          "top_p": 0.95
        }
      }
    ]
  }
}
```

### 3. 更新版本权重
```bash
PUT /api/system-configs/provider/{category}/{name}/weight
Content-Type: application/json

{
  "version": "v2",
  "weight": 50
}
```

### 4. 设置默认版本
```bash
PUT /api/system-configs/provider/{category}/{name}/default
Content-Type: application/json

{
  "version": "v2"
}
```

### 5. 刷新灰度配置
```bash
POST /api/system-configs/provider/{category}/{name}/refresh
```

## 使用场景

### 1. 新版本发布
1. 创建新版本配置（v2）
2. 设置较小的权重（如 10%）
3. 监控新版本表现
4. 逐步增加权重
5. 完全切换到新版本

### 2. A/B 测试
1. 创建两个版本配置
2. 设置相同的权重（50%/50%）
3. 收集性能数据
4. 根据结果调整权重

### 3. 故障回滚
1. 监控健康评分
2. 发现故障版本
3. 降低故障版本权重
4. 或直接切换到备用版本

## 配置示例

### 创建多版本 LLM 配置

```bash
# 创建 v1 版本
POST /api/system-configs/provider
Content-Type: application/json

{
  "category": "LLM",
  "name": "OllamaLLM",
  "version": "v1",
  "weight": 70,
  "is_active": true,
  "is_default": true,
  "type": "ollama",
  "model_name": "llama2",
  "url": "http://localhost:11434",
  "temperature": 0.7,
  "max_tokens": 2048,
  "top_p": 0.9
}

# 创建 v2 版本
POST /api/system-configs/provider
Content-Type: application/json

{
  "category": "LLM",
  "name": "OllamaLLM",
  "version": "v2",
  "weight": 30,
  "is_active": true,
  "is_default": false,
  "type": "ollama",
  "model_name": "llama2:13b",
  "url": "http://localhost:11434",
  "temperature": 0.8,
  "max_tokens": 4096,
  "top_p": 0.95
}
```

### 创建多版本 TTS 配置

```bash
# 创建 v1 版本（Edge TTS）
POST /api/system-configs/provider
Content-Type: application/json

{
  "category": "TTS",
  "name": "EdgeTTS",
  "version": "v1",
  "weight": 60,
  "is_active": true,
  "is_default": true,
  "type": "edge",
  "voice": "zh-CN-XiaoxiaoNeural",
  "format": "mp3",
  "output_dir": "./audio"
}

# 创建 v2 版本（Doubao TTS）
POST /api/system-configs/provider
Content-Type: application/json

{
  "category": "TTS",
  "name": "EdgeTTS",
  "version": "v2",
  "weight": 40,
  "is_active": true,
  "is_default": false,
  "type": "doubao",
  "voice": "xiaoyun",
  "format": "mp3",
  "output_dir": "./audio",
  "appid": "your_appid",
  "token": "your_token"
}
```

## 监控和运维

### 1. 健康检查
系统会每 30 秒自动检查各版本的健康状态，并更新健康评分。

### 2. 日志监控
```bash
# 查看灰度发布相关日志
grep "grayscale\|weight\|health" logs/app.log
```

### 3. 性能指标
- 各版本的响应时间
- 成功率
- 错误率
- 流量分布

## 最佳实践

### 1. 版本命名
- 使用语义化版本号（v1.0.0、v1.1.0）
- 或使用时间戳版本（v20240101）

### 2. 权重设置
- 新版本初始权重建议 5-10%
- 逐步增加权重，每次不超过 20%
- 监控指标稳定后再继续增加

### 3. 健康检查
- 设置合理的健康检查间隔
- 配置告警阈值
- 准备回滚方案

### 4. 配置管理
- 版本配置变更要记录
- 重要变更需要审批
- 保持配置的版本控制

## 故障处理

### 1. 版本故障
```bash
# 快速禁用故障版本
PUT /api/system-configs/provider/LLM/OllamaLLM/weight
{
  "version": "v2",
  "weight": 0
}
```

### 2. 配置错误
```bash
# 刷新配置缓存
POST /api/system-configs/provider/LLM/OllamaLLM/refresh
```

### 3. 回滚操作
```bash
# 设置旧版本为默认
PUT /api/system-configs/provider/LLM/OllamaLLM/default
{
  "version": "v1"
}
```

通过灰度发布功能，您可以安全地进行 AI Provider 的版本升级和配置变更，确保服务的稳定性和可用性。 