package ai

import "context"

// Message represents a chat message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest holds a chat completion request.
type ChatRequest struct {
	Messages []Message
	Model    string
}

// ChatResponse holds a chat completion response.
type ChatResponse struct {
	Content string
	Model   string
}

// Client defines the AI chat client interface.
type Client interface {
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
}
