package tencent

import (
	"ai-server-go/src/core/providers/asr"
	"ai-server-go/src/core/utils"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/asr/v20190614"
	"github.com/gorilla/websocket"
)

type TencentASRConfig struct {
	AppID     string `json:"app_id"`
	SecretID  string `json:"secret_id"`
	SecretKey string `json:"secret_key"`
	Region    string `json:"region"`
	Mode      string `json:"mode"`      // rest/ws
	Engine    string `json:"engine"`    // 16k_zh, 16k_en, etc.
}

type asrEventListener interface {
	OnAsrPartialResult(result string)
	OnAsrFinalResult(result string)
}

func parseProps(props map[string]interface{}, out interface{}) error {
	b, err := json.Marshal(props)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}

type Provider struct {
	*asr.BaseProvider
	config   TencentASRConfig
	listener asrEventListener
}

func NewProvider(config *asr.Config, deleteFile bool, logger *utils.Logger) (*Provider, error) {
	var cfg TencentASRConfig
	if err := parseProps(config.Data, &cfg); err != nil {
		return nil, err
	}
	return &Provider{
		BaseProvider: asr.NewBaseProvider(config, deleteFile),
		config:       cfg,
	}, nil
}

func (p *Provider) SetListener(listener asrEventListener) {
	p.listener = listener
}

func (p *Provider) Transcribe(ctx context.Context, audioData []byte) (string, error) {
	if p.config.Mode == "ws" {
		return p.transcribeWS(ctx, audioData)
	}
	return p.transcribeSDK(ctx, audioData)
}

func (p *Provider) transcribeSDK(ctx context.Context, audioData []byte) (string, error) {
	credential := common.NewCredential(p.config.SecretID, p.config.SecretKey)
	cpf := profile.NewClientProfile()
	client, err := asr.NewClient(credential, p.config.Region, cpf)
	if err != nil {
		return "", err
	}
	tmpFile := fmt.Sprintf("/tmp/tencent_asr_%d.wav", time.Now().UnixNano())
	if err := os.WriteFile(tmpFile, audioData, 0644); err != nil {
		return "", err
	}
	defer os.Remove(tmpFile)
	request := asr.NewSentenceRecognitionRequest()
	request.ProjectId = common.Uint64Ptr(0)
	request.SubServiceType = common.Uint64Ptr(2) // 2: 一句话识别
	request.EngSerViceType = common.StringPtr(p.config.Engine)
	request.SourceType = common.Uint64Ptr(1) // 1: 语音数据
	request.Data = common.StringPtr(utils.Base64EncodeFile(tmpFile))
	request.VoiceFormat = common.StringPtr("wav")
	response, err := client.SentenceRecognition(request)
	if err != nil {
		return "", err
	}
	return *response.Result, nil
}

func (p *Provider) transcribeWS(ctx context.Context, audioData []byte) (string, error) {
	// 伪代码，需根据腾讯云WebSocket协议实现分包、鉴权、异步接收
	wsURL := fmt.Sprintf("wss://asr.cloud.tencent.com/asr/v1/%s?engine_model_type=%s", p.config.AppID, p.config.Engine)
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	frameSize := 8000
	totalLen := len(audioData)
	var finalResult string
	// 1. 发送首包
	firstFrame := audioData[:min(frameSize, totalLen)]
	if err := conn.WriteMessage(websocket.BinaryMessage, firstFrame); err != nil {
		return "", err
	}
	// 2. 分包发送中间包和尾包
	go func() {
		for i := frameSize; i < totalLen; i += frameSize {
			if err := conn.WriteMessage(websocket.BinaryMessage, audioData[i:min(i+frameSize, totalLen)]); err != nil {
				break
			}
			time.Sleep(40 * time.Millisecond)
		}
	}()
	// 3. 异步接收识别结果
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		var resp struct {
			Result string `json:"result"`
			Status int    `json:"status"`
		}
		if err := json.Unmarshal(msg, &resp); err != nil {
			continue
		}
		if resp.Result != "" {
			if p.listener != nil {
				p.listener.OnAsrPartialResult(resp.Result)
			}
			finalResult = resp.Result
		}
		if resp.Status == 2 {
			if p.listener != nil {
				p.listener.OnAsrFinalResult(finalResult)
			}
			break
		}
	}
	return finalResult, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func init() {
	asr.Register("tencent", func(config *asr.Config, deleteFile bool, logger *utils.Logger) (asr.Provider, error) {
		return NewProvider(config, deleteFile, logger)
	})
}
