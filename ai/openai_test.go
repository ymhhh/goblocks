package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ymhhh/goblocks/resilience"
)

func TestOpenAIClientChat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":    "chatcmpl-test",
			"model": "gpt-4o-mini",
			"choices": []map[string]any{
				{
					"index": 0,
					"message": map[string]string{
						"role":    "assistant",
						"content": "Hello!",
					},
					"finish_reason": "stop",
				},
			},
		})
	}))
	defer server.Close()

	client := NewOpenAIClient(OpenAIConfig{
		BaseURL: server.URL + "/",
		APIKey:  "test-key",
		Model:   "gpt-4o-mini",
	})

	resp, err := client.Chat(context.Background(), ChatRequest{
		Messages: []Message{
			{Role: "user", Content: "Hi"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Content != "Hello!" {
		t.Fatalf("expected Hello!, got %s", resp.Content)
	}
}

func TestOpenAIClientWithPolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"model": "gpt-4o-mini",
			"choices": []map[string]any{
				{
					"message": map[string]string{
						"role":    "assistant",
						"content": "OK",
					},
				},
			},
		})
	}))
	defer server.Close()

	policy := &resilience.Policy{
		RateLimits: resilience.RateLimits{
			Backend:    resilience.NewMemoryRateLimiter(),
			GlobalRule: resilience.LimitRule{RPS: 100, Burst: 200},
			GlobalKey:  resilience.GlobalKey(""),
		},
	}
	client := NewOpenAIClient(OpenAIConfig{
		BaseURL: server.URL + "/",
		APIKey:  "test-key",
		Policy:  policy,
	})

	resp, err := client.Chat(context.Background(), ChatRequest{
		Messages: []Message{{Role: "user", Content: "test"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Content != "OK" {
		t.Fatalf("expected OK, got %s", resp.Content)
	}
}

func TestOpenAIClientRateLimited(t *testing.T) {
	policy := &resilience.Policy{
		RateLimits: resilience.RateLimits{
			Backend:    resilience.NewMemoryRateLimiter(),
			GlobalRule: resilience.LimitRule{RPS: 0.001, Burst: 1},
			GlobalKey:  resilience.GlobalKey(""),
		},
	}
	client := NewOpenAIClient(OpenAIConfig{
		BaseURL: "http://localhost:9999/",
		APIKey:  "test",
		Policy:  policy,
	})

	_, err := client.Chat(context.Background(), ChatRequest{
		Messages: []Message{{Role: "user", Content: "test"}},
	})
	if err == nil {
		t.Fatal("expected rate limit error")
	}
}
