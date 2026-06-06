package ai

import (
	"context"
	"fmt"
	"time"

	openai "github.com/sashabaranov/go-openai"

	"github.com/ymhhh/goblocks/metrics"
	"github.com/ymhhh/goblocks/resilience"
)

// OpenAIConfig holds OpenAI-compatible client configuration.
type OpenAIConfig struct {
	BaseURL string
	APIKey  string
	Model   string
	Policy  *resilience.Policy
	Metrics *metrics.Registry
}

// OpenAIClient implements Client using OpenAI-compatible APIs.
type OpenAIClient struct {
	client  *openai.Client
	model   string
	policy  *resilience.Policy
	metrics *metrics.Registry
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
		client:  openai.NewClientWithConfig(config),
		model:   model,
		policy:  cfg.Policy,
		metrics: cfg.Metrics,
	}
}

// Chat sends a chat completion request.
func (c *OpenAIClient) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	model := req.Model
	if model == "" {
		model = c.model
	}

	start := time.Now()
	record := func(result string) {
		if c.metrics != nil {
			c.metrics.ObserveAIRequest(model, result, time.Since(start))
		}
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
		resp, err := doChat()
		if err != nil {
			record("error")
			return nil, err
		}
		record("success")
		return resp, nil
	}

	if err := c.policy.AllowGlobal(ctx); err != nil {
		record("rate_limited")
		return nil, err
	}

	result, err := c.policy.Execute(func() (any, error) {
		return doChat()
	})
	if err != nil {
		if err == resilience.ErrCircuitOpen {
			record("circuit_open")
		} else {
			record("error")
		}
		return nil, err
	}
	record("success")
	return result.(*ChatResponse), nil
}
