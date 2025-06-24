-- AI Server Go 数据库表结构设计
-- 创建数据库
CREATE DATABASE IF NOT EXISTS ai_server_go CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE ai_server_go;

-- 1. 用户表 (users)
CREATE TABLE users (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '用户ID',
    username VARCHAR(64) UNIQUE NOT NULL COMMENT '用户名',
    email VARCHAR(128) UNIQUE COMMENT '邮箱',
    phone VARCHAR(20) COMMENT '手机号',
    password_hash VARCHAR(255) NOT NULL COMMENT '密码哈希',
    salt VARCHAR(64) NOT NULL COMMENT '密码盐值',
    nickname VARCHAR(64) COMMENT '昵称',
    avatar VARCHAR(255) COMMENT '头像URL',
    status ENUM('active', 'inactive', 'blocked') NOT NULL DEFAULT 'active' COMMENT '用户状态',
    role ENUM('admin', 'user', 'guest') NOT NULL DEFAULT 'user' COMMENT '用户角色',
    last_login_time TIMESTAMP NULL COMMENT '最后登录时间',
    last_login_ip VARCHAR(45) COMMENT '最后登录IP',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_username (username),
    INDEX idx_email (email),
    INDEX idx_status (status),
    INDEX idx_role (role)
) COMMENT '用户表';

-- 2. 用户认证表 (user_auth)
CREATE TABLE user_auth (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '认证ID',
    user_id BIGINT NOT NULL COMMENT '用户ID',
    auth_type ENUM('password', 'token', 'oauth', 'api_key') NOT NULL COMMENT '认证类型',
    auth_key VARCHAR(255) NOT NULL COMMENT '认证密钥',
    auth_secret VARCHAR(255) COMMENT '认证密钥(可选)',
    is_active BOOLEAN NOT NULL DEFAULT TRUE COMMENT '是否激活',
    expires_at TIMESTAMP NULL COMMENT '过期时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_auth_key (auth_key),
    INDEX idx_is_active (is_active)
) COMMENT '用户认证表';

-- 3. 设备表 (devices)
CREATE TABLE devices (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '设备ID',
    device_uuid VARCHAR(36) UNIQUE NOT NULL COMMENT '设备UUID标识',
    oui VARCHAR(8) NOT NULL COMMENT '厂商标识(OUI)',
    sn VARCHAR(32) NOT NULL COMMENT '设备序列号',
    device_name VARCHAR(128) NOT NULL COMMENT '设备名称',
    device_type VARCHAR(32) NOT NULL DEFAULT 'esp32' COMMENT '设备类型',
    device_model VARCHAR(64) COMMENT '设备型号',
    firmware_version VARCHAR(32) COMMENT '固件版本',
    hardware_version VARCHAR(32) COMMENT '硬件版本',
    status ENUM('active', 'inactive', 'blocked') NOT NULL DEFAULT 'active' COMMENT '设备状态',
    last_online_time TIMESTAMP NULL COMMENT '最后在线时间',
    last_ip_address VARCHAR(45) COMMENT '最后IP地址',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_device_uuid (device_uuid),
    INDEX idx_oui (oui),
    INDEX idx_sn (sn),
    UNIQUE KEY uk_oui_sn (oui, sn),
    INDEX idx_status (status),
    INDEX idx_last_online (last_online_time)
) COMMENT '设备信息表';

-- 4. 用户设备绑定表 (user_devices)
CREATE TABLE user_devices (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '绑定ID',
    user_id BIGINT NOT NULL COMMENT '用户ID',
    device_id BIGINT NOT NULL COMMENT '设备ID',
    device_alias VARCHAR(64) COMMENT '设备别名',
    is_owner BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否为设备所有者',
    permissions JSON COMMENT '设备权限配置',
    is_active BOOLEAN NOT NULL DEFAULT TRUE COMMENT '是否激活',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE,
    UNIQUE KEY uk_user_device (user_id, device_id),
    INDEX idx_user_id (user_id),
    INDEX idx_device_id (device_id),
    INDEX idx_is_owner (is_owner),
    INDEX idx_is_active (is_active)
) COMMENT '用户设备绑定表';

-- 5. 设备认证表 (device_auth)
CREATE TABLE device_auth (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '认证ID',
    device_id BIGINT NOT NULL COMMENT '设备ID',
    auth_type ENUM('token', 'api_key', 'oauth') NOT NULL DEFAULT 'token' COMMENT '认证类型',
    auth_key VARCHAR(255) NOT NULL COMMENT '认证密钥',
    auth_secret VARCHAR(255) COMMENT '认证密钥(可选)',
    is_active BOOLEAN NOT NULL DEFAULT TRUE COMMENT '是否激活',
    expires_at TIMESTAMP NULL COMMENT '过期时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE,
    INDEX idx_device_id (device_id),
    INDEX idx_auth_key (auth_key),
    INDEX idx_is_active (is_active)
) COMMENT '设备认证表';

-- 6. AI能力配置表 (ai_capabilities)
CREATE TABLE ai_capabilities (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '能力ID',
    capability_name VARCHAR(32) NOT NULL COMMENT '能力名称(ASR/LLM/TTS/VLLLM/MCP)',
    capability_type VARCHAR(64) NOT NULL COMMENT '能力类型(如openai/doubao/edge等)',
    display_name VARCHAR(128) NOT NULL COMMENT '显示名称',
    description TEXT COMMENT '能力描述',
    config_schema JSON COMMENT '配置模式定义',
    is_global BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否为全局能力',
    is_active BOOLEAN NOT NULL DEFAULT TRUE COMMENT '是否激活',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY uk_capability (capability_name, capability_type),
    INDEX idx_capability_name (capability_name),
    INDEX idx_is_global (is_global),
    INDEX idx_is_active (is_active)
) COMMENT 'AI能力配置表';

-- 7. 用户AI能力表 (user_capabilities)
CREATE TABLE user_capabilities (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '用户能力ID',
    user_id BIGINT NOT NULL COMMENT '用户ID',
    capability_id BIGINT NOT NULL COMMENT '能力ID',
    config_data JSON COMMENT '用户特定的配置数据',
    is_active BOOLEAN NOT NULL DEFAULT TRUE COMMENT '是否激活',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (capability_id) REFERENCES ai_capabilities(id) ON DELETE CASCADE,
    UNIQUE KEY uk_user_capability (user_id, capability_id),
    INDEX idx_user_id (user_id),
    INDEX idx_capability_id (capability_id),
    INDEX idx_is_active (is_active)
) COMMENT '用户AI能力表';

-- 8. 设备AI能力关联表 (device_capabilities)
CREATE TABLE device_capabilities (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '关联ID',
    device_id BIGINT NOT NULL COMMENT '设备ID',
    capability_id BIGINT NOT NULL COMMENT '能力ID',
    priority INT NOT NULL DEFAULT 0 COMMENT '优先级(数字越小优先级越高)',
    config_data JSON COMMENT '设备特定的配置数据',
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE,
    FOREIGN KEY (capability_id) REFERENCES ai_capabilities(id) ON DELETE CASCADE,
    UNIQUE KEY uk_device_capability (device_id, capability_id),
    INDEX idx_device_id (device_id),
    INDEX idx_capability_id (capability_id),
    INDEX idx_priority (priority),
    INDEX idx_is_enabled (is_enabled)
) COMMENT '设备AI能力关联表';

-- 9. 全局配置表 (global_configs)
CREATE TABLE global_configs (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '配置ID',
    config_key VARCHAR(128) UNIQUE NOT NULL COMMENT '配置键',
    config_value TEXT COMMENT '配置值',
    config_type ENUM('string', 'int', 'float', 'bool', 'json') NOT NULL DEFAULT 'string' COMMENT '配置类型',
    description TEXT COMMENT '配置描述',
    is_system BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否为系统配置',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_config_key (config_key),
    INDEX idx_is_system (is_system)
) COMMENT '全局配置表';

-- 10. 系统配置表 (system_configs)
CREATE TABLE system_configs (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '配置ID',
    config_category VARCHAR(32) NOT NULL COMMENT '配置分类(prompt/audio/ai_providers/connectivity)',
    config_key VARCHAR(128) NOT NULL COMMENT '配置键',
    config_value TEXT COMMENT '配置值',
    config_type ENUM('string', 'int', 'float', 'bool', 'json', 'array') NOT NULL DEFAULT 'string' COMMENT '配置类型',
    description TEXT COMMENT '配置描述',
    is_default BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否为默认配置',
    created_by BIGINT NULL COMMENT '创建者用户ID',
    updated_by BIGINT NULL COMMENT '更新者用户ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE SET NULL,
    UNIQUE KEY uk_category_key (config_category, config_key),
    INDEX idx_config_category (config_category),
    INDEX idx_config_key (config_key),
    INDEX idx_is_default (is_default)
) COMMENT '系统配置表';

-- 11. 会话记录表 (sessions)
CREATE TABLE sessions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '会话ID',
    user_id BIGINT NULL COMMENT '用户ID(可为空，支持匿名会话)',
    device_id BIGINT NOT NULL COMMENT '设备ID',
    session_id VARCHAR(64) UNIQUE NOT NULL COMMENT '会话唯一标识',
    client_id VARCHAR(64) COMMENT '客户端ID',
    start_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '开始时间',
    end_time TIMESTAMP NULL COMMENT '结束时间',
    status ENUM('active', 'closed', 'timeout') NOT NULL DEFAULT 'active' COMMENT '会话状态',
    message_count INT NOT NULL DEFAULT 0 COMMENT '消息数量',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE,
    INDEX idx_session_id (session_id),
    INDEX idx_user_id (user_id),
    INDEX idx_device_id (device_id),
    INDEX idx_status (status),
    INDEX idx_start_time (start_time)
) COMMENT '会话记录表';

-- 12. 使用统计表 (usage_stats)
CREATE TABLE usage_stats (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '统计ID',
    user_id BIGINT NULL COMMENT '用户ID(可为空，支持匿名统计)',
    device_id BIGINT NOT NULL COMMENT '设备ID',
    capability_name VARCHAR(32) NOT NULL COMMENT '能力名称',
    usage_date DATE NOT NULL COMMENT '使用日期',
    request_count INT NOT NULL DEFAULT 0 COMMENT '请求次数',
    success_count INT NOT NULL DEFAULT 0 COMMENT '成功次数',
    error_count INT NOT NULL DEFAULT 0 COMMENT '错误次数',
    total_duration INT NOT NULL DEFAULT 0 COMMENT '总耗时(毫秒)',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE,
    UNIQUE KEY uk_user_device_capability_date (user_id, device_id, capability_name, usage_date),
    INDEX idx_user_id (user_id),
    INDEX idx_device_id (device_id),
    INDEX idx_capability_name (capability_name),
    INDEX idx_usage_date (usage_date)
) COMMENT '使用统计表';

-- 插入默认的AI能力配置
INSERT INTO ai_capabilities (capability_name, capability_type, display_name, description, is_global, config_schema) VALUES
-- ASR能力
('ASR', 'doubao', '豆包语音识别', '豆包流式语音识别服务', TRUE, '{"appid": "string", "access_token": "string", "output_dir": "string"}'),
('ASR', 'gosherpa', 'GoSherpa语音识别', 'GoSherpa本地语音识别服务', TRUE, '{"addr": "string"}'),

-- LLM能力
('LLM', 'openai', 'OpenAI大语言模型', 'OpenAI GPT系列模型', TRUE, '{"api_key": "string", "model_name": "string", "base_url": "string", "max_tokens": "int"}'),
('LLM', 'ollama', 'Ollama本地模型', 'Ollama本地大语言模型', TRUE, '{"url": "string", "model_name": "string"}'),

-- TTS能力
('TTS', 'edge', 'Edge TTS', '微软Edge TTS语音合成', TRUE, '{"voice": "string", "output_dir": "string"}'),
('TTS', 'doubao', '豆包语音合成', '豆包语音合成服务', TRUE, '{"voice": "string", "output_dir": "string", "appid": "string", "token": "string", "cluster": "string"}'),
('TTS', 'gosherpa', 'GoSherpa语音合成', 'GoSherpa本地语音合成', TRUE, '{"cluster": "string", "output_dir": "string"}'),

-- VLLLM能力
('VLLLM', 'openai', 'OpenAI视觉模型', 'OpenAI GPT-4V等视觉模型', TRUE, '{"api_key": "string", "model_name": "string", "base_url": "string", "max_tokens": "int"}'),
('VLLLM', 'ollama', 'Ollama视觉模型', 'Ollama本地视觉模型', TRUE, '{"url": "string", "model_name": "string", "max_tokens": "int"}'),

-- MCP能力
('MCP', 'xiaozhi', '小智MCP工具', '小智内置MCP工具', TRUE, '{}'),
('MCP', 'external', '外部MCP服务', '外部MCP服务器', TRUE, '{"command": "string", "args": "array", "env": "object"}');

-- 插入默认全局配置
INSERT INTO global_configs (config_key, config_value, config_type, description, is_system) VALUES
('server.port', '8000', 'int', 'WebSocket服务端口', TRUE),
('web.port', '8080', 'int', 'HTTP服务端口', TRUE),
('log.level', 'INFO', 'string', '日志级别', TRUE),
('log.dir', 'logs', 'string', '日志目录', TRUE),
('default.asr', 'doubao', 'string', '默认ASR服务', TRUE),
('default.llm', 'openai', 'string', '默认LLM服务', TRUE),
('default.tts', 'edge', 'string', '默认TTS服务', TRUE),
('default.vlllm', 'openai', 'string', '默认VLLLM服务', TRUE);

-- 插入默认管理员用户 (密码: admin123)
INSERT INTO users (username, email, password_hash, salt, nickname, role) VALUES
('admin', 'admin@example.com', '8c6976e5b5410415bde908bd4dee15dfb167a9c873fc4bb8a81f6f2ab448a918', 'admin_salt', '系统管理员', 'admin');

-- Provider配置表（支持灰度发布）
CREATE TABLE IF NOT EXISTS provider_configs (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    category VARCHAR(20) NOT NULL COMMENT '类别：ASR/TTS/LLM/VLLLM',
    name VARCHAR(100) NOT NULL COMMENT 'provider名称',
    version VARCHAR(20) DEFAULT 'v1' COMMENT '版本号',
    weight INT DEFAULT 100 COMMENT '流量权重（0-100）',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    is_default BOOLEAN DEFAULT FALSE COMMENT '是否为默认版本',
    type VARCHAR(50) NOT NULL COMMENT 'provider类型',
    model_name VARCHAR(100) COMMENT '模型名称',
    url VARCHAR(500) COMMENT 'API地址',
    api_key VARCHAR(500) COMMENT 'API密钥',
    voice VARCHAR(100) COMMENT '语音（TTS专用）',
    format VARCHAR(20) COMMENT '音频格式（TTS专用）',
    output_dir VARCHAR(500) COMMENT '输出目录（TTS专用）',
    appid VARCHAR(100) COMMENT '应用ID（TTS专用）',
    token VARCHAR(500) COMMENT 'Token（TTS专用）',
    cluster VARCHAR(500) COMMENT '集群（TTS专用）',
    temperature DECIMAL(3,2) DEFAULT 0.7 COMMENT '温度参数',
    max_tokens INT DEFAULT 2048 COMMENT '最大令牌数',
    top_p DECIMAL(3,2) DEFAULT 0.9 COMMENT 'TopP参数',
    security JSON COMMENT '安全配置（VLLLM专用）',
    extra JSON COMMENT '其他扩展参数',
    health_score DECIMAL(5,2) DEFAULT 100.00 COMMENT '健康评分（0-100）',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_category_name_version (category, name, version),
    INDEX idx_category_name (category, name),
    INDEX idx_is_active (is_active),
    INDEX idx_is_default (is_default)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='AI Provider配置表（支持灰度发布）'; 

-- 用户-Provider绑定表
CREATE TABLE IF NOT EXISTS user_provider (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    provider_id BIGINT NOT NULL,
    category VARCHAR(20) NOT NULL, -- 如 TTS/ASR/LLM/VLLLM
    is_active BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uniq_user_category (user_id, category)
);

-- 设备-Provider绑定表
CREATE TABLE IF NOT EXISTS device_provider (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    device_id BIGINT NOT NULL,
    provider_id BIGINT NOT NULL,
    category VARCHAR(20) NOT NULL, -- 如TTS/ASR/LLM/VLLLM
    is_active BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uniq_device_category (device_id, category)
);

-- 聊天会话表
CREATE TABLE IF NOT EXISTS chat_sessions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '会话ID',
    user_id BIGINT NULL COMMENT '用户ID（可选，支持匿名会话）',
    device_id BIGINT NOT NULL COMMENT '设备ID',
    session_id VARCHAR(100) NOT NULL UNIQUE COMMENT '会话标识符',
    title VARCHAR(200) COMMENT '会话标题',
    summary TEXT COMMENT '会话摘要',
    message_count INT DEFAULT 0 COMMENT '消息数量',
    start_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '开始时间',
    end_time TIMESTAMP NULL COMMENT '结束时间',
    status ENUM('active', 'archived', 'deleted') DEFAULT 'active' COMMENT '状态',
    tags VARCHAR(500) COMMENT '标签，逗号分隔',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_device_id (device_id),
    INDEX idx_session_id (session_id),
    INDEX idx_status (status),
    INDEX idx_start_time (start_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='聊天会话表';

-- 聊天消息表
CREATE TABLE IF NOT EXISTS chat_messages (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '消息ID',
    session_id VARCHAR(100) NOT NULL COMMENT '会话ID',
    user_id BIGINT NULL COMMENT '用户ID（可选）',
    device_id BIGINT NOT NULL COMMENT '设备ID',
    role ENUM('user', 'assistant', 'system') NOT NULL COMMENT '消息角色',
    content TEXT NOT NULL COMMENT '消息内容',
    message_type VARCHAR(20) DEFAULT 'text' COMMENT '消息类型：text, image, audio, function_call',
    metadata TEXT COMMENT '元数据（JSON格式）',
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '消息时间戳',
    is_processed BOOLEAN DEFAULT FALSE COMMENT '是否已处理（用于记忆生成）',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE,
    INDEX idx_session_id (session_id),
    INDEX idx_user_id (user_id),
    INDEX idx_device_id (device_id),
    INDEX idx_role (role),
    INDEX idx_timestamp (timestamp),
    INDEX idx_is_processed (is_processed)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='聊天消息表';

-- 聊天记忆表
CREATE TABLE IF NOT EXISTS chat_memories (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '记忆ID',
    user_id BIGINT NULL COMMENT '用户ID（可选，支持匿名记忆）',
    device_id BIGINT NOT NULL COMMENT '设备ID',
    session_id VARCHAR(100) NOT NULL COMMENT '会话ID',
    memory_type VARCHAR(20) NOT NULL COMMENT '记忆类型：conversation, summary, key_points',
    content TEXT NOT NULL COMMENT '记忆内容',
    importance INT DEFAULT 1 COMMENT '重要性评分（1-10）',
    tags VARCHAR(500) COMMENT '标签，逗号分隔',
    last_used TIMESTAMP NULL COMMENT '最后使用时间',
    use_count INT DEFAULT 0 COMMENT '使用次数',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否激活',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_device_id (device_id),
    INDEX idx_session_id (session_id),
    INDEX idx_memory_type (memory_type),
    INDEX idx_importance (importance),
    INDEX idx_last_used (last_used),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='聊天记忆表';

-- 插入默认系统配置
INSERT INTO global_configs (config_key, config_value, config_type, description, is_system) VALUES
('memory.enabled', 'true', 'boolean', '是否启用聊天记忆功能', TRUE),
('memory.limit', '5', 'int', '记忆查询限制数量', TRUE),
('memory.auto_generate', 'true', 'boolean', '是否自动生成记忆', TRUE),
('memory.importance_threshold', '5', 'int', '记忆重要性阈值', TRUE);