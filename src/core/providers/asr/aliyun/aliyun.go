package aliyun

import (
	"ai-server-go/src/core/providers/asr"
	"context"
	"encoding/json"
	
	"time"
	"github.com/aliyun/alibabacloud-nls-go-sdk"
)

type AliyunASRConfig struct {
	AppKey    string `json:"app_key"`
	AccessKey string `json:"access_key"`
	Secret    string `json:"secret"`
	Region    string `json:"region"`
	Mode      string `json:"mode"`      // rest/ws
	Language  string `json:"language"`
	Model     string `json:"model"`
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
	config   AliyunASRConfig
	listener asrEventListener
}

func NewProvider(config *asr.Config, deleteFile bool, logger interface{}) (*Provider, error) {
	var cfg AliyunASRConfig
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
	return p.transcribeSDK(ctx, audioData)
}

func (p *Provider) transcribeSDK(ctx context.Context, audioData []byte) (string, error) {
	config := nls.NewConnectionConfigWithAKInfoDefault(
		nls.DEFAULT_URL,
		p.config.AppKey,
		p.config.AccessKey,
		p.config.Secret,
	)
	logger := nls.DefaultNlsLog()
	var finalResult string
	sr, err := nls.NewSpeechRecognition(
		config, logger,
		func(text string, param interface{}) {}, // onTaskFailed
		func(text string, param interface{}) {}, // onStarted
		func(text string, param interface{}) { // onResultChanged
			finalResult = text
			if p.listener != nil {
				p.listener.OnAsrPartialResult(text)
			}
		},
		func(text string, param interface{}) { // onCompleted
			finalResult = text
			if p.listener != nil {
				p.listener.OnAsrFinalResult(text)
			}
		},
		func(param interface{}) {}, // onClose
		nil,
	)
	if err != nil {
		return "", err
	}
	param := nls.DefaultSpeechRecognitionParam()
	ready, err := sr.Start(param, nil)
	if err != nil {
		return "", err
	}
	<-ready
	frameSize := 3200
	for i := 0; i < len(audioData); i += frameSize {
		end := i + frameSize
		if end > len(audioData) {
			end = len(audioData)
		}
		sr.SendAudioData(audioData[i:end])
		time.Sleep(10 * time.Millisecond)
	}
	ready, err = sr.Stop()
	if err != nil {
		return "", err
	}
	<-ready
	sr.Shutdown()
	return finalResult, nil
}

func init() {
	asr.Register("aliyun", func(config *asr.Config, deleteFile bool, logger interface{}) (asr.Provider, error) {
		return NewProvider(config, deleteFile, logger)
	})
}
