package providers

import (
	"ai-server-go/src/core/types"
	"context"
)

// Provider 所有提供者的基础接口
type Provider interface {
	Initialize() error
	Cleanup() error
}

type AsrEventListener interface {
	OnAsrResult(result string) bool
}

// ASRProvider 语音识别能力接口
// 兼容现有ASR实现，便于统一管理和能力回退
// 典型方法：Transcribe、Reset、SetListener等
type ASRProvider interface {
	Provider
	// 直接识别音频数据
	Transcribe(ctx context.Context, audioData []byte) (string, error)
	// 添加音频数据到缓冲区
	AddAudio(data []byte) error

	SetListener(listener AsrEventListener)
	// 复位ASR状态
	Reset() error

	// 获取当前静音计数
	GetSilenceCount() int

	ResetStartListenTime()
}

// TTSProvider 语音合成提供者接口
type TTSProvider interface {
	Provider

	// 合成音频并返回文件路径
	ToTTS(text string) (string, error)
}

// LLMProvider 大语言模型提供者接口
type LLMProvider interface {
	types.LLMProvider
}

// Message 对话消息
type Message = types.Message

// VADProvider 语音活动检测能力接口（预留）
type VADProvider interface {
	Detect(audio []byte, sampleRate int) ([][2]int, error)
}
