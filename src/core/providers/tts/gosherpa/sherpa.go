package gosherpa

import (
	"ai-server-go/src/core/providers/tts"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/websocket"
)

// Provider Sherpa TTS提供者实现
type Provider struct {
	*tts.BaseProvider
	conn *websocket.Conn
}

// 配置结构体
type GoSherpaTTSConfig struct {
	Cluster   string `json:"cluster"`
	OutputDir string `json:"output_dir"`
}

// 通用配置解析
func parseProps(props map[string]interface{}, out interface{}) error {
	b, err := json.Marshal(props)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}

// NewProvider 创建Sherpa TTS提供者
func NewProvider(config *tts.Config, deleteFile bool) (*Provider, error) {
	var cfg GoSherpaTTSConfig
	if err := parseProps(config.Props, &cfg); err != nil {
		return nil, fmt.Errorf("配置解析失败: %v", err)
	}
	if cfg.Cluster == "" {
		return nil, fmt.Errorf("缺少cluster配置")
	}
	base := tts.NewBaseProvider(config, deleteFile)
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
	conn, _, err := dialer.DialContext(context.Background(), cfg.Cluster, map[string][]string{})
	if err != nil {
		return nil, err
	}
	return &Provider{
		BaseProvider: base,
		conn:         conn,
	}, nil
}

// ToTTS 将文本转换为音频文件，并返回文件路径
func (p *Provider) ToTTS(text string) (string, error) {
	// 获取配置的声音，如果未配置则使用默认值
	SherpaTTSStartTime := time.Now()

	// 创建临时文件路径用于保存 SherpaTTS 生成的 MP3
	outputDir := p.BaseProvider.Config().OutputDir
	if outputDir == "" {
		outputDir = os.TempDir() // Use system temp dir if not configured
	}
	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("创建输出目录失败04 '%s': %v", outputDir, err)
	}
	// Use a unique filename
	tempFile := filepath.Join(outputDir, fmt.Sprintf("go_sherpa_tts_%d.wav", time.Now().UnixNano()))

	p.conn.WriteMessage(websocket.TextMessage, []byte(text))
	_, bytes, err := p.conn.ReadMessage()

	if err != nil {
		return "", fmt.Errorf("go-sherpa-tts 获取音频流失败: %v", err)
	}

	ttsDuration := time.Since(SherpaTTSStartTime)
	fmt.Println(fmt.Sprintf("go-sherpa-tts 语音合成完成，耗时: %s", ttsDuration))

	// 将音频数据写入临时文件
	err = os.WriteFile(tempFile, bytes, 0644)
	if err != nil {
		return "", fmt.Errorf("写入音频文件 '%s' 失败: %v", tempFile, err)
	}

	// 检查文件是否成功创建
	if _, err := os.Stat(tempFile); os.IsNotExist(err) {
		return "", fmt.Errorf("go-sherpa-tts 未能创建音频文件: %s", tempFile)
	}

	// Return the path to the generated audio file
	return tempFile, nil
}

func init() {
	// 注册Sherpa TTS提供者
	tts.Register("gosherpa", func(config *tts.Config, deleteFile bool) (tts.Provider, error) {
		return NewProvider(config, deleteFile)
	})
}
