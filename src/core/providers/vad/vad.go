package vad

import (
	"ai-server-go/src/core/utils"
	"context"
	"fmt"
)

// Config VAD配置结构
// Type: "silero"（本地），"thirdparty"（第三方API）等
// ModelDir: 本地模型目录
// Threshold: 检测阈值
// Extra: 其他扩展参数
// ...
type Config struct {
	Type      string                 // vad类型
	ModelDir  string                 // 本地模型目录
	Threshold float64                // 检测阈值
	Extra     map[string]interface{} // 其他扩展参数
}

// Provider VAD接口（区间检测，兼容原有能力体系）
type Provider interface {
	Detect(ctx context.Context, audio []byte, sampleRate int) ([][2]int, error) // 返回语音区间（起止帧）
	Config() *Config
}

// VadModel VAD模型接口 - 定义VAD模型的基本功能（对标Java接口）
type VadModel interface {
	// 初始化VAD模型
	Initialize() error
	// 获取语音概率（0.0-1.0）
	GetSpeechProbability(samples []float32) (float32, error)
	// 重置模型状态
	Reset()
	// 关闭模型资源
	Close() error
}

// VadDetector 语音活动检测器接口（对标Java接口）
type VadDetector interface {
	// 处理音频数据，返回完整音频或nil
	ProcessAudio(sessionId string, pcmData []byte) ([]byte, error)
	// 重置会话状态
	ResetSession(sessionId string)
	// 检查当前是否正在说话
	IsSpeaking(sessionId string) bool
	// 获取当前语音概率（0.0-1.0）
	GetSpeechProbability(sessionId string) (float32, error)
}

// Factory VAD工厂函数类型
// deleteFile: 是否自动删除临时文件
// logger: 日志
// ...
type Factory func(config *Config, logger *utils.Logger) (Provider, error)

var (
	factories = make(map[string]Factory)
)

// Register 注册VAD工厂
func Register(name string, factory Factory) {
	factories[name] = factory
}

// Create 创建VAD实例
func Create(name string, config *Config, logger *utils.Logger) (Provider, error) {
	factory, ok := factories[name]
	if !ok {
		return nil, fmt.Errorf("未知的VAD类型: %s", name)
	}
	return factory(config, logger)
}
