package ai

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"

	"github.com/ymhhh/goblocks/resilience"
)

// OpenAIConfig holds OpenAI-compatible client configuration.
type OpenAIConfig struct {
	BaseURL string
	APIKey  string
	Model   string
	Policy  *resilience.Policy
}

// OpenAIClient implements Client using OpenAI-compatible APIs.
type OpenAIClient struct {
	client *openai.Client
	model  string
	policy *resilience.Policy
}

// NewOpenAIClient creates an OpenAI-compatible client.
func NewOpenAIClient(cfg OpenAIConfig) *OpenAIClient {
	config := openai.DefaultConfig(cfg.APIKey)
	if cfg.BaseURL != "" {
		config.BaseURL = cfg.BaseURL
	}

	model := cfg.Model
	if model == "" {
		model = openai.GPT4oMini
	}

	return &OpenAIClient{
		client: openai.NewClientWithConfig(config),
		model:  model,
		policy: cfg.Policy,
	}
}

// Chat sends a chat completion request.
func (c *OpenAIClient) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	model := req.Model
	if model == "" {
		model = c.model
	}

	messages := make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, m := range req.Messages {
		messages[i] = openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	doChat := func() (*ChatResponse, error) {
		resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:    model,
			Messages: messages,
		})
		if err != nil {
			return nil, fmt.Errorf("chat completion: %w", err)
		}
		if len(resp.Choices) == 0 {
			return nil, fmt.Errorf("empty response from AI")
		}
		return &ChatResponse{
			Content: resp.Choices[0].Message.Content,
			Model:   resp.Model,
		}, nil
	}

	if c.policy == nil {
		return doChat()
	}

	if err := c.policy.Allow(); err != nil {
		return nil, err
	}

	result, err := c.policy.Execute(func() (any, error) {
		return doChat()
	})
	if err != nil {
		return nil, err
	}
	return result.(*ChatResponse), nil
}
