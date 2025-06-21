# 能力配置回退功能测试指南

## 功能概述

新增的能力配置回退功能实现了以下优先级逻辑：
1. **设备专属配置** (优先级最高)
2. **用户自定义配置** (优先级中等)
3. **系统默认配置** (优先级最低)

## 测试步骤

### 1. 准备测试数据

#### 1.1 创建用户
```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <admin_token>" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "123456",
    "nickname": "测试用户"
  }'
```

#### 1.2 用户登录获取token
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "123456"
  }'
```

#### 1.3 创建设备
```bash
curl -X POST http://localhost:8080/api/devices \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <user_token>" \
  -d '{
    "oui": "12345678",
    "sn": "TEST001",
    "device_name": "测试设备",
    "device_type": "esp32"
  }'
```

#### 1.4 创建AI能力
```bash
# 创建OpenAI LLM能力
curl -X POST http://localhost:8080/api/capabilities \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <admin_token>" \
  -d '{
    "name": "llm",
    "type": "openai",
    "config": {
      "api_key": "sk-test-openai",
      "model": "gpt-3.5-turbo"
    }
  }'

# 创建Edge TTS能力
curl -X POST http://localhost:8080/api/capabilities \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <admin_token>" \
  -d '{
    "name": "tts",
    "type": "edge",
    "config": {
      "voice": "zh-CN-XiaoxiaoNeural"
    }
  }'
```

#### 1.5 设置系统默认能力
```bash
# 设置默认LLM为OpenAI
curl -X POST http://localhost:8080/api/capabilities/defaults \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <admin_token>" \
  -d '{
    "capability_name": "llm",
    "capability_type": "openai"
  }'

# 设置默认TTS为Edge
curl -X POST http://localhost:8080/api/capabilities/defaults \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <admin_token>" \
  -d '{
    "capability_name": "tts",
    "capability_type": "edge"
  }'
```

### 2. 测试回退逻辑

#### 2.1 测试场景1：设备无配置，使用系统默认
```bash
# 获取设备能力配置（带回退逻辑）
curl -X GET http://localhost:8080/api/devices/{device_uuid}/capabilities/with-fallback \
  -H "Authorization: Bearer <user_token>"
```

**预期结果**：
- LLM和TTS都使用系统默认配置
- `priority_source` 为 "system"
- `priority` 为 200

#### 2.2 测试场景2：用户设置自定义配置
```bash
# 用户设置自定义LLM配置
curl -X POST http://localhost:8080/api/users/{user_id}/capabilities \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <user_token>" \
  -d '{
    "capability_name": "llm",
    "capability_type": "openai",
    "config": {
      "api_key": "sk-user-openai",
      "model": "gpt-4"
    }
  }'
```

再次获取设备能力配置：
```bash
curl -X GET http://localhost:8080/api/devices/{device_uuid}/capabilities/with-fallback \
  -H "Authorization: Bearer <user_token>"
```

**预期结果**：
- LLM使用用户自定义配置，`priority_source` 为 "user"，`priority` 为 100
- TTS仍使用系统默认配置，`priority_source` 为 "system"，`priority` 为 200

#### 2.3 测试场景3：设备设置专属配置
```bash
# 设备设置专属TTS配置
curl -X POST http://localhost:8080/api/devices/{device_uuid}/capabilities \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <user_token>" \
  -d '{
    "capability_name": "tts",
    "capability_type": "edge",
    "priority": 1,
    "config": {
      "voice": "zh-CN-YunxiNeural"
    },
    "is_enabled": true
  }'
```

再次获取设备能力配置：
```bash
curl -X GET http://localhost:8080/api/devices/{device_uuid}/capabilities/with-fallback \
  -H "Authorization: Bearer <user_token>"
```

**预期结果**：
- LLM使用用户自定义配置，`priority_source` 为 "user"，`priority` 为 100
- TTS使用设备专属配置，`priority_source` 为 "device"，`priority` 为 1

### 3. 验证优先级逻辑

#### 3.1 优先级验证
- **设备专属配置**：优先级 1-99
- **用户自定义配置**：优先级 100
- **系统默认配置**：优先级 200

#### 3.2 配置覆盖验证
- 设备配置会覆盖用户配置和系统默认配置
- 用户配置会覆盖系统默认配置
- 系统默认配置作为最后的兜底方案

### 4. 错误处理测试

#### 4.1 设备不存在
```bash
curl -X GET http://localhost:8080/api/devices/invalid-uuid/capabilities/with-fallback \
  -H "Authorization: Bearer <user_token>"
```

#### 4.2 用户未认证
```bash
curl -X GET http://localhost:8080/api/devices/{device_uuid}/capabilities/with-fallback
```

#### 4.3 能力不存在
- 测试获取不存在的AI能力类型
- 验证是否正确返回空结果

## 日志验证

检查服务器日志，确认以下信息：
1. 能力配置的优先级选择过程
2. 回退逻辑的执行情况
3. 错误处理和异常情况

## 性能测试

1. **并发测试**：多个用户同时获取设备能力配置
2. **缓存测试**：验证配置获取的性能
3. **数据库查询优化**：确保查询效率

## 注意事项

1. 确保数据库中有足够的测试数据
2. 测试前清理可能冲突的配置
3. 验证配置的JSON格式正确性
4. 检查权限控制是否正常工作 