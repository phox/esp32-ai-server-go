package gosherpa

import (
	"ai-server-go/src/core/providers/asr"
	"ai-server-go/src/core/utils"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

type Provider struct {
	*asr.BaseProvider
	conn *websocket.Conn
}

// 配置结构体
type GoSherpaASRConfig struct {
	Addr      string `json:"addr"`
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

func NewProvider(config *asr.Config, deleteFile bool, logger *utils.Logger) (*Provider, error) {
	var cfg GoSherpaASRConfig
	if err := parseProps(config.Data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Addr == "" {
		return nil, fmt.Errorf("缺少addr配置")
	}
	base := asr.NewBaseProvider(config, deleteFile)

	provider := &Provider{
		BaseProvider: base,
	}
	// 初始化音频处理
	provider.InitAudioProcessing()
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second, // 设置握手超时
	}
	conn, _, err := dialer.DialContext(context.Background(), cfg.Addr, map[string][]string{})
	if err != nil {
		return nil, err
	}
	provider.conn = conn
	go func() {
		defer func() {
			if err := recover(); err != nil {
			}
		}()
		for {
			messageType, p, _ := conn.ReadMessage()
			if messageType == websocket.TextMessage {
				if listener := provider.GetListener(); listener != nil {
					if finished := listener.OnAsrResult(string(p)); finished {

					}
				}
			}
		}

	}()

	return provider, nil
}

func (p *Provider) Transcribe(ctx context.Context, audioData []byte) (string, error) {
	return "", nil
}

// 添加音频数据到缓冲区
func (p *Provider) AddAudio(data []byte) error {
	p.conn.WriteMessage(websocket.BinaryMessage, data)

	return nil
}

// 复位ASR状态
func (p *Provider) Reset() error {
	return nil
}

func init() {
	asr.Register("gosherpa", func(config *asr.Config, deleteFile bool, logger *utils.Logger) (asr.Provider, error) {
		return NewProvider(config, deleteFile, logger)
	})
}
