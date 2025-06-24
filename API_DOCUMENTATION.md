# AI Server Go - 用户管理API文档

## 概述

AI Server Go 提供了完整的用户管理系统，包括用户认证、设备绑定、AI能力配置等功能。所有API都需要通过RESTful接口进行访问，支持Token认证。

## 认证机制

### 登录获取Token
```http
POST /api/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123"
}
```

**响应示例：**
```json
{
  "message": "登录成功",
  "data": {
    "token": "a1b2c3d4e5f6...",
    "user": {
      "id": 1,
      "username": "admin",
      "nickname": "系统管理员",
      "role": "admin"
    },
    "expires_at": "2024-01-02T12:00:00Z"
  }
}
```

### 使用Token认证
在后续请求中，需要在Header中包含Token：
```http
Authorization: Bearer a1b2c3d4e5f6...
```

## 用户管理API

### 1. 获取用户列表
```http
GET /api/users?offset=0&limit=20&status=active&role=user
Authorization: Bearer <token>
```

**查询参数：**
- `offset`: 偏移量（默认0）
- `limit`: 限制数量（默认20，最大100）
- `status`: 用户状态过滤（active/inactive/blocked）
- `role`: 用户角色过滤（admin/user/guest）

### 2. 创建用户
```http
POST /api/users
Authorization: Bearer <token>
Content-Type: application/json

{
  "username": "newuser",
  "email": "user@example.com",
  "password": "password123",
  "nickname": "新用户"
}
```

**权限要求：** 管理员权限

### 3. 获取用户信息
```http
GET /api/users/{id}
Authorization: Bearer <token>
```

### 4. 更新用户信息
```http
PUT /api/users/{id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "email": "newemail@example.com",
  "phone": "13800138000",
  "nickname": "新昵称",
  "avatar": "https://example.com/avatar.jpg"
}
```

**权限要求：** 只能更新自己的信息，或管理员可以更新任何用户

### 5. 更新用户密码
```http
PUT /api/users/{id}/password
Authorization: Bearer <token>
Content-Type: application/json

{
  "old_password": "oldpassword",
  "new_password": "newpassword123"
}
```

**权限要求：** 只能更新自己的密码，或管理员可以更新任何用户的密码

### 6. 删除用户
```http
DELETE /api/users/{id}
Authorization: Bearer <token>
```

**权限要求：** 管理员权限

## 用户设备管理API

### 1. 获取用户设备列表
```http
GET /api/users/{id}/devices
Authorization: Bearer <token>
```

### 2. 绑定用户设备
```http
POST /api/users/{id}/devices
Authorization: Bearer <token>
Content-Type: application/json

{
  "device_uuid": "550e8400-e29b-41d4-a716-446655440000",
  "device_alias": "客厅设备",
  "is_owner": true,
  "permissions": {
    "read": true,
    "write": true,
    "control": true
  }
}
```

### 3. 解绑用户设备
```http
DELETE /api/users/{id}/devices/{deviceUUID}
Authorization: Bearer <token>
```

## 用户AI能力管理API

### 1. 获取用户AI能力配置
```http
GET /api/users/{id}/capabilities
Authorization: Bearer <token>
```

### 2. 设置用户AI能力配置
```http
POST /api/users/{id}/capabilities
Authorization: Bearer <token>
Content-Type: application/json

{
  "capability_name": "LLM",
  "capability_type": "openai",
  "config": {
    "api_key": "sk-...",
    "model_name": "gpt-3.5-turbo",
    "max_tokens": 1000
  }
}
```

### 3. 移除用户AI能力配置
```http
DELETE /api/users/{id}/capabilities/{capabilityName}/{capabilityType}
Authorization: Bearer <token>
```

## 设备管理

### 获取设备列表
- **GET** `/api/devices`
- **描述**: 获取所有设备列表
- **权限**: 需要认证
- **响应**:
```json
{
  "data": [
    {
      "id": 1,
      "device_uuid": "550e8400-e29b-41d4-a716-446655440000",
      "oui": "12345678",
      "sn": "SN001",
      "device_name": "ESP32设备1",
      "device_type": "esp32",
      "device_model": "ESP32-WROOM-32",
      "firmware_version": "1.0.0",
      "hardware_version": "1.0",
      "status": "online",
      "last_online_time": "2024-01-01T12:00:00Z",
      "last_ip_address": "192.168.1.100",
      "created_at": "2024-01-01T10:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  ]
}
```

### 创建设备
- **POST** `/api/devices`
- **描述**: 创建新设备
- **权限**: 需要认证
- **请求体**:
```json
{
  "oui": "12345678",
  "sn": "SN001",
  "device_name": "ESP32设备1",
  "device_type": "esp32",
  "device_model": "ESP32-WROOM-32",
  "firmware_version": "1.0.0",
  "hardware_version": "1.0"
}
```

### 获取设备详情
- **GET** `/api/devices/:id`
- **描述**: 根据设备UUID获取设备详情
- **权限**: 需要认证

### 根据OUI和SN获取设备
- **GET** `/api/devices/oui/:oui/sn/:sn`
- **描述**: 根据OUI和SN获取设备详情
- **权限**: 需要认证

### 更新设备
- **PUT** `/api/devices/:id`
- **描述**: 更新设备信息
- **权限**: 需要认证
- **请求体**:
```json
{
  "device_name": "ESP32设备1-更新",
  "device_type": "esp32",
  "device_model": "ESP32-WROOM-32",
  "firmware_version": "1.1.0",
  "hardware_version": "1.0",
  "status": "online"
}
```

### 删除设备
- **DELETE** `/api/devices/:id`
- **描述**: 删除设备
- **权限**: 需要认证

### 获取设备AI能力配置
- **GET** `/api/devices/:id/capabilities`
- **描述**: 获取设备的AI能力配置
- **权限**: 需要认证
- **响应**:
```json
{
  "data": [
    {
      "id": 1,
      "device_id": 1,
      "capability_id": 1,
      "priority": 1,
      "config_data": {
        "api_key": "sk-xxx",
        "model": "gpt-3.5-turbo"
      },
      "is_enabled": true,
      "capability": {
        "capability_name": "llm",
        "capability_type": "openai",
        "display_name": "OpenAI LLM",
        "description": "OpenAI语言模型"
      }
    }
  ]
}
```

### 获取设备AI能力配置（带回退逻辑）
- **GET** `/api/devices/:id/capabilities/with-fallback`
- **描述**: 获取设备的AI能力配置，如果设备没有配置则回退到用户自定义或系统默认配置
- **权限**: 需要认证
- **优先级**: 设备专属配置 > 用户自定义配置 > 系统默认配置
- **响应**:
```json
{
  "data": {
    "device_id": 1,
    "capabilities": [
      {
        "capability_name": "llm",
        "capability_type": "openai",
        "config": {
          "api_key": "sk-xxx",
          "model": "gpt-3.5-turbo"
        },
        "priority": 1,
        "is_enabled": true,
        "priority_source": "device"
      },
      {
        "capability_name": "tts",
        "capability_type": "edge",
        "config": {
          "voice": "zh-CN-XiaoxiaoNeural"
        },
        "priority": 100,
        "is_enabled": true,
        "priority_source": "user"
      }
    ],
    "global_configs": {
      "default.asr": "gosherpa",
      "default.llm": "openai"
    }
  }
}
```

### 设置设备AI能力配置
- **POST** `/api/devices/:id/capabilities`
- **描述**: 设置设备的AI能力配置
- **权限**: 需要认证
- **请求体**:
```json
{
  "capability_name": "llm",
  "capability_type": "openai",
  "priority": 1,
  "config": {
    "api_key": "sk-xxx",
    "model": "gpt-3.5-turbo",
    "temperature": 0.7
  },
  "is_enabled": true
}
```

### 移除设备AI能力配置
- **DELETE** `/api/devices/:id/capabilities/:capabilityName/:capabilityType`
- **描述**: 移除设备的AI能力配置
- **权限**: 需要认证

## AI能力管理

### 获取AI能力列表
- **GET** `/api/capabilities?type=llm&enabled=true`
- **描述**: 获取AI能力列表
- **权限**: 需要认证
- **查询参数**:
  - `type`: 能力类型（可选）
  - `enabled`: 是否启用（可选，true/false）

### 获取AI能力详情
- **GET** `/api/capabilities/:name/:type`
- **描述**: 获取AI能力详情
- **权限**: 需要认证

### 创建AI能力
- **POST** `/api/capabilities`
- **描述**: 创建新的AI能力
- **权限**: 需要管理员权限
- **请求体**:
```json
{
  "name": "llm",
  "type": "openai",
  "config": {
    "api_key": "sk-xxx",
    "model": "gpt-3.5-turbo"
  }
}
```

### 更新AI能力
- **PUT** `/api/capabilities/:name/:type`
- **描述**: 更新AI能力配置
- **权限**: 需要管理员权限
- **请求体**:
```json
{
  "name": "llm",
  "type": "openai",
  "config": {
    "api_key": "sk-xxx",
    "model": "gpt-4"
  }
}
```

### 删除AI能力
- **DELETE** `/api/capabilities/:name/:type`
- **描述**: 删除AI能力（软删除，设置为非活跃状态）
- **权限**: 需要管理员权限

### 获取默认AI能力列表
- **GET** `/api/capabilities/defaults`
- **描述**: 获取系统默认的AI能力类型配置
- **权限**: 需要认证
- **响应**:
```json
{
  "data": {
    "asr": "gosherpa",
    "llm": "openai",
    "tts": "edge"
  }
}
```

### 设置默认AI能力
- **POST** `/api/capabilities/defaults`
- **描述**: 设置系统默认的AI能力类型
- **权限**: 需要管理员权限
- **请求体**:
```json
{
  "capability_name": "llm",
  "capability_type": "openai"
}
```

### 移除默认AI能力
- **DELETE** `/api/capabilities/defaults/:capabilityName`
- **描述**: 移除系统默认的AI能力类型
- **权限**: 需要管理员权限

## 认证相关API

### 1. 用户登录
```http
POST /api/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123"
}
```

### 2. 用户登出
```http
POST /api/auth/logout
Authorization: Bearer <token>
```

### 3. 获取当前用户信息
```http
GET /api/auth/me
Authorization: Bearer <token>
```

## 错误响应格式

所有API在发生错误时都会返回统一的错误格式：

```json
{
  "error": "错误描述信息"
}
```

常见HTTP状态码：
- `200`: 请求成功
- `201`: 创建成功
- `400`: 请求参数错误
- `401`: 未认证或Token无效
- `403`: 权限不足
- `404`: 资源不存在
- `500`: 服务器内部错误

## 数据模型

### 用户模型 (User)
```json
{
  "id": 1,
  "username": "admin",
  "email": "admin@example.com",
  "phone": "13800138000",
  "nickname": "系统管理员",
  "avatar": "https://example.com/avatar.jpg",
  "status": "active",
  "role": "admin",
  "last_login_time": "2024-01-01T12:00:00Z",
  "last_login_ip": "192.168.1.100",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

### 设备模型 (Device)
```json
{
  "id": 1,
  "device_uuid": "550e8400-e29b-41d4-a716-446655440000",
  "oui": "12345678",
  "sn": "ESP32WROOM32001",
  "device_name": "ESP32设备",
  "device_type": "esp32",
  "device_model": "ESP32-WROOM-32",
  "firmware_version": "1.0.0",
  "hardware_version": "1.0",
  "status": "active",
  "last_online_time": "2024-01-01T12:00:00Z",
  "last_ip_address": "192.168.1.101",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

### AI能力模型 (AICapability)
```json
{
  "id": 1,
  "capability_name": "LLM",
  "capability_type": "openai",
  "display_name": "OpenAI大语言模型",
  "description": "OpenAI GPT系列模型",
  "config_schema": {
    "api_key": "string",
    "model_name": "string",
    "max_tokens": "int"
  },
  "is_global": true,
  "is_active": true,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

## 使用示例

### 完整的工作流程

1. **用户登录**
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}'
```

2. **创建新用户**
```bash
curl -X POST http://localhost:8080/api/users \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123",
    "nickname": "测试用户"
  }'
```

3. **绑定设备**
```bash
curl -X POST http://localhost:8080/api/users/2/devices \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "device_uuid": "550e8400-e29b-41d4-a716-446655440000",
    "device_alias": "我的设备",
    "oui": "phox",
    "sn": "100101"
    "is_owner": true,
    "permissions": {"read": true, "write": true}
  }'
```

4. **配置AI能力**
```bash
curl -X POST http://localhost:8080/api/users/2/capabilities \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "capability_name": "LLM",
    "capability_type": "openai",
    "config": {
      "api_key": "sk-your-api-key",
      "model_name": "gpt-3.5-turbo"
    }
  }'
```

## 注意事项

1. **Token有效期**: Token默认24小时有效，过期后需要重新登录
2. **权限控制**: 用户只能操作自己的资源，管理员可以操作所有资源
3. **设备绑定**: 一个设备可以绑定给多个用户，但每个用户只能绑定一次
4. **AI能力配置**: 支持用户级别和设备级别的AI能力配置，设备级别优先级更高
5. **数据库**: 确保MySQL数据库已正确配置并运行
6. **默认管理员**: 系统初始化时会创建默认管理员账户（用户名：admin，密码：admin123）

## 测试说明

### 1. 用户管理测试
```bash
# 1. 创建管理员用户
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "email": "admin@example.com",
    "password": "admin123",
    "nickname": "管理员"
  }'

# 2. 登录获取token
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'

# 3. 使用token访问API
curl -X GET http://localhost:8080/api/users \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

### 2. 设备管理测试
```bash
# 创建设备
curl -X POST http://localhost:8080/api/devices \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "oui": "12345678",
    "sn": "ESP32_001",
    "device_name": "测试设备1",
    "device_type": "esp32",
    "device_model": "ESP32-WROOM-32"
  }'

# 绑定设备到用户
curl -X POST http://localhost:8080/api/users/1/devices \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "device_uuid": "DEVICE_UUID_HERE",
    "device_alias": "我的设备",
    "is_owner": true
  }'
```

### 3. AI能力配置测试
```bash
# 设置设备AI能力
curl -X POST http://localhost:8080/api/devices/1/capabilities \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "capability_name": "ASR",
    "capability_type": "DoubaoASR",
    "priority": 1,
    "config": {
      "appid": "your_appid",
      "access_token": "your_token"
    },
    "is_enabled": true
  }'

# 获取设备AI能力（带回退逻辑）
curl -X GET http://localhost:8080/api/devices/1/capabilities/with-fallback \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

## 系统配置管理API

### 概述
系统配置管理API允许管理员管理全局的AI配置，包括提示词、音频处理、AI提供商默认设置等。这些配置会作为系统默认值，当设备或用户没有特定配置时使用。

### 权限要求
- 所有系统配置管理API都需要管理员权限
- 需要在请求头中包含有效的管理员token

### API端点

#### 1. 获取系统配置列表
```http
GET /api/system-configs?category={category}&is_default={is_default}
```

**参数：**
- `category` (可选): 配置分类 (prompt/audio/ai_providers/connectivity)
- `is_default` (可选): 是否为默认配置 (true/false)

**响应：**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "config_category": "prompt",
      "config_key": "default_prompt",
      "config_value": "你是小智/小志，来自中国台湾省的00后女生...",
      "config_type": "string",
      "description": "默认AI提示词",
      "is_default": true,
      "created_by": 1,
      "updated_by": 1,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

#### 2. 获取单个系统配置
```http
GET /api/system-configs/{category}/{key}
```

**响应：**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "config_category": "prompt",
    "config_key": "default_prompt",
    "config_value": "你是小智/小志，来自中国台湾省的00后女生...",
    "config_type": "string",
    "description": "默认AI提示词",
    "is_default": true,
    "created_by": 1,
    "updated_by": 1,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

#### 3. 设置系统配置
```http
POST /api/system-configs
```

**请求体：**
```json
{
  "category": "prompt",
  "key": "default_prompt",
  "value": "你是小智/小志，来自中国台湾省的00后女生...",
  "config_type": "string",
  "description": "默认AI提示词",
  "is_default": true
}
```

**配置类型说明：**
- `string`: 字符串类型
- `int`: 整数类型
- `float`: 浮点数类型
- `bool`: 布尔类型
- `json`: JSON对象类型
- `array`: 数组类型

#### 4. 删除系统配置
```http
DELETE /api/system-configs/{category}/{key}
```

#### 5. 获取指定分类的所有系统配置
```http
GET /api/system-configs/{category}
```

**响应：**
```json
{
  "success": true,
  "data": {
    "default_prompt": "你是小智/小志，来自中国台湾省的00后女生...",
    "delete_audio": true,
    "quick_reply": true,
    "quick_reply_words": ["我在", "在呢", "来了", "啥事啊"]
  }
}
```

#### 6. 初始化默认系统配置
```http
POST /api/system-configs/initialize
```

**说明：** 此API会初始化所有默认的系统配置，包括：
- 提示词配置
- 音频处理配置
- AI提供商默认配置
- 连通性检查配置

### 配置分类说明

#### 1. prompt (提示词配置)
- `default_prompt`: 默认AI提示词

#### 2. audio (音频处理配置)
- `delete_audio`: 是否删除音频文件 (bool)
- `quick_reply`: 是否启用快速回复 (bool)
- `quick_reply_words`: 快速回复词汇 (array)

#### 3. ai_providers (AI提供商默认配置)
- `default_asr`: 默认ASR提供商 (string)
- `default_tts`: 默认TTS提供商 (string)
- `default_llm`: 默认LLM提供商 (string)
- `default_vlllm`: 默认VLLLM提供商 (string)

#### 4. connectivity (连通性检查配置)
- `enabled`: 是否启用连通性检查 (bool)
- `timeout`: 检查超时时间 (string)
- `retry_attempts`: 重试次数 (int)
- `retry_delay`: 重试延迟 (string)
- `asr_test_audio`: ASR测试音频文件 (string)
- `llm_test_prompt`: LLM测试提示词 (string)
- `tts_test_text`: TTS测试文本 (string)

### 使用示例

#### 1. 修改默认AI提示词
```bash
curl -X POST http://localhost:8080/api/system-configs \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "category": "prompt",
    "key": "default_prompt",
    "value": "你是一个友好的AI助手，请用简洁的语言回答问题。",
    "config_type": "string",
    "description": "默认AI提示词",
    "is_default": true
  }'
```

#### 2. 修改音频处理配置
```bash
curl -X POST http://localhost:8080/api/system-configs \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "category": "audio",
    "key": "quick_reply_words",
    "value": "[\"好的\", \"没问题\", \"收到\"]",
    "config_type": "array",
    "description": "快速回复词汇",
    "is_default": true
  }'
```

#### 3. 修改默认AI提供商
```bash
curl -X POST http://localhost:8080/api/system-configs \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "category": "ai_providers",
    "key": "default_llm",
    "value": "OpenAILLM",
    "config_type": "string",
    "description": "默认LLM提供商",
    "is_default": true
  }'
```

### 注意事项

1. **权限控制**: 只有管理员用户可以访问系统配置管理API
2. **配置优先级**: 系统配置作为最后的默认值，优先级低于设备特定配置和用户特定配置
3. **配置类型**: 设置配置时必须指定正确的配置类型，否则可能导致解析错误
4. **初始化**: 系统启动时会自动初始化默认配置，也可以通过API手动初始化
5. **配置验证**: 建议在修改配置前先获取当前配置，确保修改的正确性

### 错误处理

```json
{
  "error": "需要管理员权限"
}
```

```json
{
  "error": "系统配置不存在"
}
```

```json
{
  "error": "设置系统配置失败: 配置类型错误"
}
```

## 1. AI Provider 配置与绑定 API（后端实现版）

## 1.1 Provider 配置管理

### 1.1.1 获取 Provider 配置列表
- **GET** `/api/configs/provider?category={category}`
- **描述**: 获取所有或指定类型（ASR/TTS/LLM/VLLLM）的 Provider 配置列表
- **权限**: 认证用户
- **参数**:
  - `category`（可选）: Provider类别
- **响应**:
  ```json
  {
    "success": true,
    "data": [
      {
        "id": 1,
        "category": "llm",
        "name": "openai",
        "type": "openai",
        "version": "v1",
        "weight": 100,
        "is_active": true,
        "is_default": false,
        "props": { ... },
        "created_at": "...",
        "updated_at": "..."
      }
    ]
  }
  ```

### 1.1.2 获取单个 Provider 配置
- **GET** `/api/configs/provider/{id}`
- **描述**: 获取指定ID的 Provider 配置

### 1.1.3 创建 Provider 配置
- **POST** `/api/configs/provider`
- **权限**: 管理员
- **请求体**:
  ```json
  {
    "category": "llm",
    "name": "openai",
    "type": "openai",
    "version": "v1",
    "weight": 100,
    "is_active": true,
    "is_default": false,
    "props": { ... }
  }
  ```

### 1.1.4 更新 Provider 配置
- **PUT** `/api/configs/provider/{id}`
- **权限**: 管理员
- **请求体**: 同上

### 1.1.5 删除 Provider 配置
- **DELETE** `/api/configs/provider/{id}`
- **权限**: 管理员

## 1.2 Provider 绑定与优先级

### 1.2.1 用户绑定 Provider
- **POST** `/api/user/provider/bind`
- **权限**: 认证用户（仅本人或管理员）
- **请求体**:
  ```json
  {
    "user_id": 1,
    "provider_id": 2,
    "category": "llm"
  }
  ```
- **说明**: 绑定后该用户在该类别下优先使用此 Provider

### 1.2.2 用户解绑 Provider
- **POST** `/api/user/provider/unbind`
- **权限**: 认证用户（仅本人或管理员）
- **请求体**:
  ```json
  {
    "user_id": 1,
    "category": "llm"
  }
  ```

### 1.2.3 查询用户绑定 Provider
- **GET** `/api/user/provider/list?user_id=1`
- **权限**: 认证用户
- **响应**: Provider 绑定关系列表

### 1.2.4 设备绑定 Provider
- **POST** `/api/device/provider/bind`
- **权限**: 认证用户（仅设备所有者或管理员）
- **请求体**:
  ```json
  {
    "device_id": 1,
    "provider_id": 2,
    "category": "llm"
  }
  ```

### 1.2.5 设备解绑 Provider
- **POST** `/api/device/provider/unbind`
- **权限**: 认证用户（仅设备所有者或管理员）
- **请求体**:
  ```json
  {
    "device_id": 1,
    "category": "llm"
  }
  ```

### 1.2.6 查询设备绑定 Provider
- **GET** `/api/device/provider/list?device_id=1`
- **权限**: 认证用户

### 1.2.7 Provider 下拉列表
- **GET** `/api/provider/list?category=llm`
- **描述**: 获取指定类型 Provider 列表（只读，供前端下拉选择）
- **参数**:
  - `category`（可选）: Provider类别（如 TTS/ASR/LLM/VLLLM）
- **返回**:
  ```json
  [
    {
      "id": 1,
      "category": "llm",
      "name": "openai",
      "type": "openai",
      "version": "v1",
      "weight": 100,
      "is_active": true,
      "is_default": false,
      "props": { ... },
      "created_at": "...",
      "updated_at": "..."
    }
  ]
  ```

### 1.2.8 获取生效 Provider（优先级查找）
- **后端服务方法**: `GetEffectiveProvider(category, deviceID, userID)`
- **优先级**: 设备专属 > 用户自定义 > 系统默认

## 1.3 Provider 灰度发布与版本管理

### 1.3.1 获取 Provider 版本列表
- **GET** `/api/configs/provider/{id}/versions`
- **描述**: 获取指定 Provider 的所有版本

### 1.3.2 设置默认 Provider 版本
- **PUT** `/api/configs/provider/{id}/default`
- **权限**: 管理员
- **请求体**:
  ```json
  {
    "version": "v2"
  }
  ```

### 1.3.3 获取灰度发布状态
- **GET** `/api/configs/provider/{id}/grayscale`
- **描述**: 获取指定 Provider 的灰度发布状态

### 1.3.4 更新 Provider 版本权重
- **PUT** `/api/configs/provider/{id}/weight`
- **权限**: 管理员
- **请求体**:
  ```json
  {
    "weight": 50
  }
  ```

### 1.3.5 刷新 Provider 灰度配置
- **POST** `/api/configs/provider/{id}/refresh`
- **权限**: 管理员

## 1.4 Provider 配置数据结构

### 1.4.1 ProviderConfig
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
- **props** 字段为 JSON，结构由各 Provider 自定义。

## 1.5 其他注意事项

- Provider 配置全部走 props（JSON），便于扩展和热更新。
- 绑定优先级：设备专属 > 用户自定义 > 系统默认。
- API 返回统一错误格式，注意处理权限和参数校验。
- 灰度发布相关接口仅管理员可用。

---

如需更详细的接口参数、响应示例、错误码等，请参考本文件其他章节或后端源码。
如需前端集成示例或后端结构说明，可进一步补充。

**建议：**
将上述内容补充/替换到 `API_DOCUMENTATION.md` Provider 相关章节，确保文档与后端实现一致。

### 1.1.6 获取VAD Provider配置列表
- **GET** `/api/configs/provider?category=vad`
- **描述**: 获取所有VAD能力Provider配置

### 1.1.7 创建/更新/删除VAD Provider配置
- **POST** `/api/configs/provider`
- **PUT** `/api/configs/provider/{id}`
- **DELETE** `/api/configs/provider/{id}`
- **请求体示例**（silero本地模型）:
```json
{
  "category": "vad",
  "name": "silero",
  "type": "silero",
  "version": "v1",
  "weight": 100,
  "is_active": true,
  "is_default": true,
  "props": {
    "model_dir": "models/silero_vad.onnx",
    "threshold": 0.5
  }
}
```
- **请求体示例**（第三方API）:
```json
{
  "category": "vad",
  "name": "thirdparty",
  "type": "thirdparty",
  "version": "v1",
  "weight": 100,
  "is_active": true,
  "is_default": false,
  "props": {
    "api_url": "https://api.example.com/vad",
    "api_key": "..."
  }
}
```

### 1.2.6 用户/设备绑定VAD Provider
- **POST** `/api/user/provider/bind`  `{ "category": "vad", "provider_id": 1 }`
- **POST** `/api/device/provider/bind`  `{ "category": "vad", "provider_id": 1 }`
- **GET** `/api/user/provider/list?category=vad`
- **GET** `/api/device/provider/list?category=vad`

### 1.3.6 灰度发布与版本管理（VAD）
- **GET** `/api/configs/provider/vad/{name}/versions`
- **PUT** `/api/configs/provider/vad/{name}/weight`
- **PUT** `/api/configs/provider/vad/{name}/default`
- **POST** `/api/configs/provider/vad/{name}/refresh`

### 1.4.2 ProviderConfig VAD props字段说明
- silero: `{ "model_dir": "models/silero_vad.onnx", "threshold": 0.5 }`
- thirdparty: `{ "api_url": "...", "api_key": "..." }`

---
VAD能力已支持Provider统一管理、绑定、优先级、灰度发布等机制，前后端可按上述API进行能力扩展和配置。 