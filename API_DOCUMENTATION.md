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

## 设备管理API

### 1. 获取设备列表
```http
GET /api/devices?offset=0&limit=20&status=active
Authorization: Bearer <token>
```

### 2. 创建设备
```http
POST /api/devices
Authorization: Bearer <token>
Content-Type: application/json

{
  "oui": "12345678",
  "sn": "ESP32WROOM32001",
  "device_name": "ESP32设备",
  "device_type": "esp32",
  "device_model": "ESP32-WROOM-32",
  "firmware_version": "1.0.0",
  "hardware_version": "1.0"
}
```

### 3. 获取设备信息
```http
GET /api/devices/{deviceUUID}
Authorization: Bearer <token>
```

### 4. 根据OUI和SN获取设备信息
```http
GET /api/devices/oui/{oui}/sn/{sn}
Authorization: Bearer <token>
```

### 5. 更新设备信息
```http
PUT /api/devices/{deviceUUID}
Authorization: Bearer <token>
Content-Type: application/json

{
  "device_name": "新设备名称",
  "status": "active"
}
```

### 6. 删除设备
```http
DELETE /api/devices/{deviceUUID}
Authorization: Bearer <token>
```

## 设备AI能力配置API

### 1. 获取设备AI能力配置
```http
GET /api/devices/{deviceUUID}/capabilities
Authorization: Bearer <token>
```

### 2. 设置设备AI能力配置
```http
POST /api/devices/{deviceUUID}/capabilities
Authorization: Bearer <token>
Content-Type: application/json

{
  "capability_name": "ASR",
  "capability_type": "doubao",
  "priority": 1,
  "config": {
    "appid": "your_appid",
    "access_token": "your_token"
  },
  "is_enabled": true
}
```

### 3. 移除设备AI能力配置
```http
DELETE /api/devices/{deviceUUID}/capabilities/{capabilityName}/{capabilityType}
Authorization: Bearer <token>
```

## AI能力管理API

### 1. 获取AI能力列表
```http
GET /api/capabilities
Authorization: Bearer <token>
```

### 2. 获取AI能力详情
```http
GET /api/capabilities/{name}/{type}
Authorization: Bearer <token>
```

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