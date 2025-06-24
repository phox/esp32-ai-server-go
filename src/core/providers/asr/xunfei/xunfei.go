package xunfei

import (
	"ai-server-go/src/core/providers/asr"
	"ai-server-go/src/core/utils"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

type XunfeiASRConfig struct {
	AppID     string `json:"app_id"`
	APIKey    string `json:"api_key"`
	APISecret string `json:"api_secret"`
	Engine    string `json:"engine"`
	Language  string `json:"language"`
	Mode      string `json:"mode"` // ws
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
	config   XunfeiASRConfig
	listener asrEventListener
}

func NewProvider(config *asr.Config, deleteFile bool, logger *utils.Logger) (*Provider, error) {
	var cfg XunfeiASRConfig
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
	return p.transcribeWS(ctx, audioData)
}

func (p *Provider) transcribeWS(ctx context.Context, audioData []byte) (string, error) {
	urlStr, err := p.genWSURL()
	if err != nil {
		return "", err
	}
	conn, _, err := websocket.DefaultDialer.Dial(urlStr, nil)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	frameSize := 1280 // 40ms
	totalLen := len(audioData)
	var finalResult string
	// 1. 发送首包
	firstFrame := audioData[:min(frameSize, totalLen)]
	param := map[string]interface{}{
		"common": map[string]interface{}{
			"app_id": p.config.AppID,
		},
		"business": map[string]interface{}{
			"language":    p.config.Language,
			"domain":      "iat",
			"accent":      "mandarin",
			"engine_type": p.config.Engine,
		},
		"data": map[string]interface{}{
			"status":   0,
			"format":   "audio/L16;rate=16000",
			"encoding": "raw",
			"audio":    base64.StdEncoding.EncodeToString(firstFrame),
		},
	}
	if err := conn.WriteJSON(param); err != nil {
		return "", err
	}

	// 2. 分包发送中间包和尾包
	go func() {
		for i := frameSize; i < totalLen; i += frameSize {
			status := 1 // 中间包
			if i+frameSize >= totalLen {
				status = 2 // 尾包
			}
			data := map[string]interface{}{
				"data": map[string]interface{}{
					"status": status,
					"audio":  base64.StdEncoding.EncodeToString(audioData[i:min(i+frameSize, totalLen)]),
				},
			}
			conn.WriteJSON(data)
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
			Code int `json:"code"`
			Data struct {
				Result string `json:"result"`
			} `json:"data"`
			Desc string `json:"desc"`
		}
		if err := json.Unmarshal(msg, &resp); err != nil {
			continue
		}
		if resp.Code != 0 {
			return "", fmt.Errorf("ASR error: %s", resp.Desc)
		}
		if resp.Data.Result != "" {
			if p.listener != nil {
				p.listener.OnAsrPartialResult(resp.Data.Result)
			}
			finalResult = resp.Data.Result
		}
		if isFinal(msg) {
			if p.listener != nil {
				p.listener.OnAsrFinalResult(finalResult)
			}
			break
		}
	}
	return finalResult, nil
}

func (p *Provider) genWSURL() (string, error) {
	host := "ws-api.xfyun.cn"
	path := "/v2/iat"
	date := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
	signatureOrigin := fmt.Sprintf("host: %s\ndate: %s\nGET %s HTTP/1.1", host, date, path)
	h := hmac.New(sha256.New, []byte(p.config.APISecret))
	h.Write([]byte(signatureOrigin))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	authorizationOrigin := fmt.Sprintf("api_key=\"%s\", algorithm=\"hmac-sha256\", headers=\"host date request-line\", signature=\"%s\"", p.config.APIKey, signature)
	authorization := base64.StdEncoding.EncodeToString([]byte(authorizationOrigin))
	v := url.Values{}
	v.Add("authorization", authorization)
	v.Add("date", date)
	v.Add("host", host)
	return fmt.Sprintf("wss://%s%s?%s", host, path, v.Encode()), nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func isFinal(msg []byte) bool {
	// 需根据讯飞协议判断是否为最终包，可根据resp.data.status==2等实现
	return false // TODO: 实现协议判断
}

func init() {
	asr.Register("xunfei", func(config *asr.Config, deleteFile bool, logger *utils.Logger) (asr.Provider, error) {
		return NewProvider(config, deleteFile, logger)
	})
}
