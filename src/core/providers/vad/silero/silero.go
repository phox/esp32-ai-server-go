package silero

import (
	"ai-server-go/src/core/providers/vad"
	"ai-server-go/src/core/utils"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// SileroModel 实现 VadModel
// 通过python脚本/onnxruntime获取语音概率
// 支持初始化、概率、重置、关闭

type SileroModel struct {
	config      *vad.Config
	logger      *utils.Logger
	initialized bool
}

func (m *SileroModel) Initialize() error {
	m.initialized = true
	return nil // 本地python脚本无需特殊初始化
}

func (m *SileroModel) GetSpeechProbability(samples []float32) (float32, error) {
	// 保存为临时wav，调用python脚本返回概率
	tmpWav := filepath.Join(os.TempDir(), fmt.Sprintf("vad_prob_%d.wav", time.Now().UnixNano()))
	if err := utils.SaveFloat32PCM(samples, tmpWav); err != nil {
		return 0, fmt.Errorf("保存临时音频失败: %v", err)
	}
	defer os.Remove(tmpWav)

	sileroBin := "python"
	if val, ok := m.config.Extra["silero_bin"]; ok && val != nil {
		sileroBin = fmt.Sprintf("%v", val)
	}
	scriptPath := "src/core/providers/vad/silero/run_silero_vad.py"
	if val, ok := m.config.Extra["silero_script"]; ok && val != nil {
		scriptPath = fmt.Sprintf("%v", val)
	}
	modelPath := m.config.ModelDir
	if modelPath == "" {
		modelPath = "models/silero_vad.onnx"
	}
	threshold := m.config.Threshold
	if threshold == 0 {
		threshold = 0.5
	}

	cmd := exec.Command(sileroBin, scriptPath,
		"--input", tmpWav,
		"--model", modelPath,
		"--threshold", fmt.Sprintf("%f", threshold),
		"--sample-rate", "16000",
		"--prob-only", // 需在python脚本实现该参数，返回概率
	)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("silero vad概率执行失败: %v, 输出: %s", err, string(output))
	}
	var prob float32
	if err := json.Unmarshal(output, &prob); err != nil {
		return 0, fmt.Errorf("概率输出解析失败: %v, 输出: %s", err, string(output))
	}
	return prob, nil
}

func (m *SileroModel) Reset() {
	// 无状态，无需实现
}

func (m *SileroModel) Close() error {
	return nil
}

// SileroDetector 实现 VadDetector
// 支持多会话，管理session状态，流式概率、是否说话等

type sessionState struct {
	buffer     []byte
	isSpeaking bool
	prob       float32
}

type SileroDetector struct {
	model    *SileroModel
	sessions map[string]*sessionState
	mu       sync.Mutex
}

func NewSileroDetector(model *SileroModel) *SileroDetector {
	return &SileroDetector{
		model:    model,
		sessions: make(map[string]*sessionState),
	}
}

func (d *SileroDetector) ProcessAudio(sessionId string, pcmData []byte) ([]byte, error) {
	d.mu.Lock()
	state, ok := d.sessions[sessionId]
	if !ok {
		state = &sessionState{}
		d.sessions[sessionId] = state
	}
	// 简单拼接音频流
	state.buffer = append(state.buffer, pcmData...)
	d.mu.Unlock()

	// 仅示例：每次都计算概率，实际可按帧/窗口优化
	prob, err := d.model.GetSpeechProbability(utils.BytesToFloat32(state.buffer))
	if err != nil {
		return nil, err
	}
	state.prob = prob
	state.isSpeaking = prob > 0.5 // 阈值可配置

	// 检测到语音结束（概率降为0）时返回完整音频
	if !state.isSpeaking && len(state.buffer) > 0 {
		buf := state.buffer
		state.buffer = nil
		return buf, nil
	}
	return nil, nil
}

func (d *SileroDetector) ResetSession(sessionId string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.sessions, sessionId)
}

func (d *SileroDetector) IsSpeaking(sessionId string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if state, ok := d.sessions[sessionId]; ok {
		return state.isSpeaking
	}
	return false
}

func (d *SileroDetector) GetSpeechProbability(sessionId string) (float32, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if state, ok := d.sessions[sessionId]; ok {
		return state.prob, nil
	}
	return 0, fmt.Errorf("session not found")
}

// 能力注册
func RegisterVadModel() {
	// 可扩展注册到全局工厂
}

func RegisterVadDetector() {
	// 可扩展注册到全局工厂
}

type SileroProvider struct {
	config *vad.Config
	logger *utils.Logger
}

func (p *SileroProvider) Config() *vad.Config {
	return p.config
}

// Detect 调用本地python脚本运行silero_vad.onnx模型
func (p *SileroProvider) Detect(ctx context.Context, audio []byte, sampleRate int) ([][2]int, error) {
	tmpWav := filepath.Join(os.TempDir(), fmt.Sprintf("vad_%d.wav", time.Now().UnixNano()))
	if err := utils.SaveAudioFile(audio, tmpWav); err != nil {
		return nil, fmt.Errorf("保存临时音频失败: %v", err)
	}
	defer os.Remove(tmpWav)

	// 默认python
	sileroBin := "python"
	if val, ok := p.config.Extra["silero_bin"]; ok && val != nil {
		sileroBin = fmt.Sprintf("%v", val)
	}
	// 默认脚本路径
	scriptPath := "src/core/providers/vad/silero/run_silero_vad.py"
	if val, ok := p.config.Extra["silero_script"]; ok && val != nil {
		scriptPath = fmt.Sprintf("%v", val)
	}
	// 默认模型路径
	modelPath := p.config.ModelDir
	if modelPath == "" {
		modelPath = "models/silero_vad.onnx"
	}
	threshold := p.config.Threshold
	if threshold == 0 {
		threshold = 0.5
	}

	cmd := exec.CommandContext(ctx, sileroBin, scriptPath,
		"--input", tmpWav,
		"--model", modelPath,
		"--threshold", fmt.Sprintf("%f", threshold),
		"--sample-rate", fmt.Sprintf("%d", sampleRate),
	)
	output, err := cmd.Output()
	if err != nil {
		p.logger.Error("silero vad执行失败: %v, 输出: %s", err, string(output))
		return nil, fmt.Errorf("silero vad执行失败: %v, 输出: %s", err, string(output))
	}

	// 兼容脚本异常输出
	if len(output) > 0 && output[0] == '{' {
		var errObj map[string]interface{}
		if err := json.Unmarshal(output, &errObj); err == nil {
			if msg, ok := errObj["error"]; ok {
				return nil, fmt.Errorf("silero vad脚本错误: %v", msg)
			}
		}
	}

	// 假设输出为JSON: [[start1, end1], ...]
	var segments [][2]int
	if err := json.Unmarshal(output, &segments); err != nil {
		p.logger.Error("silero vad输出解析失败: %v, 输出: %s", err, string(output))
		return nil, fmt.Errorf("silero vad输出解析失败: %v", err)
	}
	return segments, nil
}

// 工厂注册
func init() {
	vad.Register("silero", NewProvider)
}

func NewProvider(config *vad.Config, logger *utils.Logger) (vad.Provider, error) {
	if config.ModelDir == "" {
		config.ModelDir = "models/silero_vad.onnx"
	}
	if config.Threshold == 0 {
		config.Threshold = 0.5
	}
	return &SileroProvider{config: config, logger: logger}, nil
}
