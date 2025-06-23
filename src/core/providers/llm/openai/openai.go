package openai

import (
	"ai-server-go/src/core/providers/llm"
	"ai-server-go/src/core/types"
	"context"
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

// Provider OpenAI LLM提供者
type Provider struct {
	*llm.BaseProvider
	client    *openai.Client
	maxTokens int
}

// 配置结构体
type OpenAILLMConfig struct {
	APIKey      string  `json:"api_key"`
	BaseURL     string  `json:"base_url"`
	ModelName   string  `json:"model_name"`
	Temperature float64 `json:"temperature"`
	MaxTokens   int     `json:"max_tokens"`
	TopP        float64 `json:"top_p"`
}

// 通用配置解析
func parseProps(props map[string]interface{}, out interface{}) error {
	b, err := json.Marshal(props)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}

// 注册提供者
func init() {
	llm.Register("openai", NewProvider)
}

// NewProvider 创建OpenAI提供者
func NewProvider(config *llm.Config) (llm.Provider, error) {
	var cfg OpenAILLMConfig
	if err := parseProps(config.Extra, &cfg); err != nil {
		return nil, fmt.Errorf("配置解析失败: %v", err)
	}
	base := llm.NewBaseProvider(config)
	provider := &Provider{
		BaseProvider: base,
		maxTokens:    cfg.MaxTokens,
	}
	if provider.maxTokens <= 0 {
		provider.maxTokens = 500
	}
	// 将解析到的配置写回 config 以兼容后续逻辑
	config.APIKey = cfg.APIKey
	config.BaseURL = cfg.BaseURL
	config.ModelName = cfg.ModelName
	config.Temperature = cfg.Temperature
	config.TopP = cfg.TopP
	return provider, nil
}

// Initialize 初始化提供者
func (p *Provider) Initialize() error {
	config := p.Config()
	if config.APIKey == "" {
		return fmt.Errorf("missing OpenAI API key")
	}

	clientConfig := openai.DefaultConfig(config.APIKey)
	if config.BaseURL != "" {
		clientConfig.BaseURL = config.BaseURL
	}

	p.client = openai.NewClientWithConfig(clientConfig)
	return nil
}

// Cleanup 清理资源
func (p *Provider) Cleanup() error {
	return nil
}

// Response types.LLMProvider接口实现
func (p *Provider) Response(ctx context.Context, sessionID string, messages []types.Message) (<-chan string, error) {
	responseChan := make(chan string, 10)

	go func() {
		defer close(responseChan)

		// 转换消息格式
		chatMessages := make([]openai.ChatCompletionMessage, len(messages))
		for i, msg := range messages {
			chatMessages[i] = openai.ChatCompletionMessage{
				Role:    msg.Role,
				Content: msg.Content,
			}
		}

		stream, err := p.client.CreateChatCompletionStream(
			ctx,
			openai.ChatCompletionRequest{
				Model:     p.Config().ModelName,
				Messages:  chatMessages,
				Stream:    true,
				MaxTokens: p.maxTokens,
			},
		)
		if err != nil {
			responseChan <- fmt.Sprintf("【OpenAI服务响应异常: %v】", err)
			return
		}
		defer stream.Close()

		isActive := true
		for {
			response, err := stream.Recv()
			if err != nil {
				break
			}

			if len(response.Choices) > 0 {
				content := response.Choices[0].Delta.Content
				if content != "" {
					// 处理思考标签
					if content, isActive = handleThinkTags(content, isActive); content != "" {
						responseChan <- content
					}
				}
			}
		}
	}()

	return responseChan, nil
}

// ResponseWithFunctions types.LLMProvider接口实现
func (p *Provider) ResponseWithFunctions(ctx context.Context, sessionID string, messages []types.Message, tools []openai.Tool) (<-chan types.Response, error) {
	responseChan := make(chan types.Response, 10)

	go func() {
		defer close(responseChan)

		// 转换消息格式
		chatMessages := make([]openai.ChatCompletionMessage, len(messages))
		for i, msg := range messages {
			chatMessage := openai.ChatCompletionMessage{
				Role:    msg.Role,
				Content: msg.Content,
			}

			// 处理tool_call_id字段（tool消息必需）
			if msg.ToolCallID != "" {
				chatMessage.ToolCallID = msg.ToolCallID
			}

			// 处理tool_calls字段（assistant消息中的工具调用）
			if len(msg.ToolCalls) > 0 {
				openaiToolCalls := make([]openai.ToolCall, len(msg.ToolCalls))
				for j, tc := range msg.ToolCalls {
					openaiToolCalls[j] = openai.ToolCall{
						ID:   tc.ID,
						Type: openai.ToolType(tc.Type),
						Function: openai.FunctionCall{
							Name:      tc.Function.Name,
							Arguments: tc.Function.Arguments,
						},
					}
				}
				chatMessage.ToolCalls = openaiToolCalls
			}

			chatMessages[i] = chatMessage
		}

		stream, err := p.client.CreateChatCompletionStream(
			ctx,
			openai.ChatCompletionRequest{
				Model:    p.Config().ModelName,
				Messages: chatMessages,
				Tools:    tools,
				Stream:   true,
			},
		)
		if err != nil {
			responseChan <- types.Response{
				Content: fmt.Sprintf("【OpenAI服务响应异常: %v】", err),
				Error:   err.Error(),
			}
			return
		}
		defer stream.Close()

		for {
			response, err := stream.Recv()
			if err != nil {
				break
			}

			if len(response.Choices) > 0 {
				delta := response.Choices[0].Delta
				chunk := types.Response{
					Content: delta.Content,
				}
				//fmt.Println("openai delta:", delta)

				if len(delta.ToolCalls) > 0 {
					toolCalls := make([]types.ToolCall, len(delta.ToolCalls))
					for i, tc := range delta.ToolCalls {
						toolCalls[i] = types.ToolCall{
							ID:   tc.ID,
							Type: string(tc.Type),
							Function: types.FunctionCall{
								Name:      tc.Function.Name,
								Arguments: tc.Function.Arguments,
							},
						}
					}
					chunk.ToolCalls = toolCalls
				}

				responseChan <- chunk
			}
		}
	}()

	return responseChan, nil
}

// handleThinkTags 处理思考标签
func handleThinkTags(content string, isActive bool) (string, bool) {
	if content == "" {
		return "", isActive
	}

	if content == "<think>" {
		return "", false
	}
	if content == "</think>" {
		return "", true
	}

	if !isActive {
		return "", isActive
	}

	return content, isActive
}
