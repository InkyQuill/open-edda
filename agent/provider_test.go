package agent

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestOpenAICompatibleClientSendsChatCompletionRequest(t *testing.T) {
	var captured struct {
		Method        string
		Path          string
		Authorization string
		Body          map[string]any
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured.Method = r.Method
		captured.Path = r.URL.Path
		captured.Authorization = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&captured.Body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "chatcmpl-test",
			"choices": [
				{
					"message": {
						"role": "assistant",
						"content": "A clean opening line."
					},
					"finish_reason": "stop"
				}
			],
			"usage": {
				"prompt_tokens": 10,
				"completion_tokens": 3
			}
		}`))
	}))
	t.Cleanup(server.Close)

	variant := ModelVariant{
		Name:                  "DeepSeek Chat",
		Model:                 "deepseek-chat",
		Temperature:           0.4,
		MaxOutputTokens:       512,
		RequestTokenField:     "max_tokens",
		InputPricePerMillion:  0.27,
		OutputPricePerMillion: 1.10,
	}
	client := NewOpenAICompatibleClient(server.URL+"/", "test-key", variant)

	response, err := client.Complete(context.Background(), CompletionRequest{
		Messages: []CompletionMessage{
			{Role: "system", Content: "Write concise prose."},
			{Role: "user", Content: "Continue the scene."},
		},
		Tools: []CompletionTool{
			{
				Type: "function",
				Function: CompletionToolFunction{
					Name:        "read_context",
					Description: "Read project context.",
					Parameters:  map[string]any{"type": "object"},
				},
			},
		},
		ToolChoice: "auto",
	})
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}

	if captured.Method != http.MethodPost {
		t.Fatalf("method = %q, want POST", captured.Method)
	}
	if captured.Path != "/v1/chat/completions" {
		t.Fatalf("path = %q, want /v1/chat/completions", captured.Path)
	}
	if captured.Authorization != "Bearer test-key" {
		t.Fatalf("authorization = %q, want bearer key", captured.Authorization)
	}
	assertJSONField(t, captured.Body, "model", "deepseek-chat")
	assertJSONField(t, captured.Body, "tool_choice", "auto")
	assertJSONField(t, captured.Body, "temperature", 0.4)
	assertJSONField(t, captured.Body, "max_tokens", float64(512))
	if got := len(captured.Body["messages"].([]any)); got != 2 {
		t.Fatalf("messages count = %d, want 2", got)
	}
	if got := len(captured.Body["tools"].([]any)); got != 1 {
		t.Fatalf("tools count = %d, want 1", got)
	}

	if response.Message.Role != MessageRoleAssistant {
		t.Fatalf("response role = %q, want assistant", response.Message.Role)
	}
	if response.Message.Content != "A clean opening line." {
		t.Fatalf("response content = %q", response.Message.Content)
	}
	if response.FinishReason != "stop" {
		t.Fatalf("finish reason = %q, want stop", response.FinishReason)
	}
}

func TestOpenAICompatibleClientSendsExplicitZeroTemperature(t *testing.T) {
	var capturedBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&capturedBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"choices": [
				{"message": {"role": "assistant", "content": "Done."}}
			]
		}`))
	}))
	t.Cleanup(server.Close)

	temperature := 0.0
	client := NewOpenAICompatibleClient(server.URL, "test-key", ModelVariant{
		Model:           "deepseek-chat",
		Temperature:     0.7,
		MaxOutputTokens: 128,
	})

	_, err := client.Complete(context.Background(), CompletionRequest{
		Messages:    []CompletionMessage{{Role: "user", Content: "Use exact wording."}},
		Temperature: &temperature,
	})
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}

	assertJSONField(t, capturedBody, "temperature", 0.0)
}

func TestOpenAICompatibleClientRejectsResponseWithoutChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[]}`))
	}))
	t.Cleanup(server.Close)

	client := NewOpenAICompatibleClient(server.URL, "test-key", ModelVariant{Model: "deepseek-chat"})
	_, err := client.Complete(context.Background(), CompletionRequest{
		Messages: []CompletionMessage{{Role: "user", Content: "Continue."}},
	})
	if err == nil {
		t.Fatal("Complete() error = nil, want missing choices error")
	}
	if !strings.Contains(err.Error(), "choices") {
		t.Fatalf("Complete() error = %q, want choices validation error", err)
	}
}

func TestOpenAICompatibleClientNormalizesDeepSeekUsageAndCosts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"choices": [
				{
					"message": {
						"role": "assistant",
						"content": "Done."
					}
				}
			],
			"usage": {
				"prompt_tokens": 1000,
				"completion_tokens": 200,
				"prompt_tokens_details": {
					"cached_tokens": 700,
					"cache_write_tokens": 0
				}
			}
		}`))
	}))
	t.Cleanup(server.Close)

	variant := ModelVariant{
		Model:                     "deepseek-chat",
		MaxOutputTokens:           128,
		InputPricePerMillion:      0.27,
		OutputPricePerMillion:     1.10,
		CacheReadPricePerMillion:  0.07,
		CacheWritePricePerMillion: 0.27,
	}
	client := NewOpenAICompatibleClient(server.URL, "test-key", variant)

	response, err := client.Complete(context.Background(), CompletionRequest{
		Messages: []CompletionMessage{{Role: "user", Content: "Summarize."}},
	})
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}

	if response.Usage.InputTokens != 300 ||
		response.Usage.OutputTokens != 200 ||
		response.Usage.CacheReadTokens != 700 ||
		response.Usage.CacheWriteTokens != 0 ||
		response.Usage.TotalTokens != 1200 {
		t.Fatalf("token usage = %#v, want input=300 output=200 cacheRead=700 cacheWrite=0 total=1200", response.Usage)
	}
	assertFloatNear(t, response.Usage.InputCost, 300*0.27/1_000_000)
	assertFloatNear(t, response.Usage.OutputCost, 200*1.10/1_000_000)
	assertFloatNear(t, response.Usage.CacheReadCost, 700*0.07/1_000_000)
	assertFloatNear(t, response.Usage.CacheWriteCost, 0)
	assertFloatNear(t, response.Usage.TotalCost, (300*0.27+200*1.10+700*0.07)/1_000_000)
}

func TestOpenAICompatibleClientNormalizesUsageEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		usageJSON string
		want      Usage
	}{
		{
			name: "prompt cache hit tokens map to cache read",
			usageJSON: `{
				"prompt_tokens": 100,
				"completion_tokens": 20,
				"prompt_cache_hit_tokens": 30
			}`,
			want: Usage{
				InputTokens:     70,
				OutputTokens:    20,
				CacheReadTokens: 30,
				TotalTokens:     120,
			},
		},
		{
			name: "cache write tokens reduce input",
			usageJSON: `{
				"prompt_tokens": 100,
				"completion_tokens": 20,
				"prompt_tokens_details": {
					"cache_write_tokens": 40
				}
			}`,
			want: Usage{
				InputTokens:      60,
				OutputTokens:     20,
				CacheWriteTokens: 40,
				TotalTokens:      120,
			},
		},
		{
			name: "cache tokens greater than prompt clamp input",
			usageJSON: `{
				"prompt_tokens": 100,
				"completion_tokens": 20,
				"prompt_cache_hit_tokens": 80,
				"prompt_tokens_details": {
					"cache_write_tokens": 50
				}
			}`,
			want: Usage{
				InputTokens:      0,
				OutputTokens:     20,
				CacheReadTokens:  80,
				CacheWriteTokens: 50,
				TotalTokens:      120,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{
					"choices": [
						{"message": {"role": "assistant", "content": "Done."}}
					],
					"usage": ` + tt.usageJSON + `
				}`))
			}))
			t.Cleanup(server.Close)

			client := NewOpenAICompatibleClient(server.URL, "test-key", ModelVariant{Model: "deepseek-chat"})
			response, err := client.Complete(context.Background(), CompletionRequest{
				Messages: []CompletionMessage{{Role: "user", Content: "Continue."}},
			})
			if err != nil {
				t.Fatalf("Complete() error = %v", err)
			}
			if response.Usage.InputTokens != tt.want.InputTokens ||
				response.Usage.OutputTokens != tt.want.OutputTokens ||
				response.Usage.CacheReadTokens != tt.want.CacheReadTokens ||
				response.Usage.CacheWriteTokens != tt.want.CacheWriteTokens ||
				response.Usage.TotalTokens != tt.want.TotalTokens {
				t.Fatalf("usage = %#v, want %#v", response.Usage, tt.want)
			}
		})
	}
}

func TestOpenAICompatibleClientReturnsNon2xxStatusAndBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "provider unavailable", http.StatusBadGateway)
	}))
	t.Cleanup(server.Close)

	client := NewOpenAICompatibleClient(server.URL, "test-key", ModelVariant{Model: "deepseek-chat"})
	_, err := client.Complete(context.Background(), CompletionRequest{
		Messages: []CompletionMessage{{Role: "user", Content: "Continue."}},
	})
	if err == nil {
		t.Fatal("Complete() error = nil, want non-2xx error")
	}
	if !strings.Contains(err.Error(), "status 502") || !strings.Contains(err.Error(), "provider unavailable") {
		t.Fatalf("Complete() error = %q, want status and body", err)
	}
}

func TestOpenAICompatibleClientSurfacesContextCancellation(t *testing.T) {
	requestStarted := make(chan struct{})
	releaseHandler := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		close(requestStarted)
		select {
		case <-r.Context().Done():
		case <-releaseHandler:
		}
	}))
	t.Cleanup(func() {
		close(releaseHandler)
		server.Close()
	})

	ctx, cancel := context.WithCancel(context.Background())
	client := NewOpenAICompatibleClient(server.URL, "test-key", ModelVariant{Model: "deepseek-chat"})
	errc := make(chan error, 1)
	go func() {
		_, err := client.Complete(ctx, CompletionRequest{
			Messages: []CompletionMessage{{Role: "user", Content: "Continue."}},
		})
		errc <- err
	}()

	select {
	case <-requestStarted:
	case <-time.After(time.Second):
		t.Fatal("request did not reach test server")
	}
	cancel()

	var err error
	select {
	case err = <-errc:
	case <-time.After(time.Second):
		t.Fatal("Complete() did not return after context cancellation")
	}
	if err == nil {
		t.Fatal("Complete() error = nil, want context cancellation error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Complete() error = %v, want context.Canceled", err)
	}
}

func assertFloatNear(t *testing.T, got, want float64) {
	t.Helper()

	if math.Abs(got-want) > 0.000000000001 {
		t.Fatalf("float = %.18f, want %.18f", got, want)
	}
}

func assertJSONField(t *testing.T, body map[string]any, field string, want any) {
	t.Helper()

	if got := body[field]; got != want {
		t.Fatalf("%s = %#v, want %#v", field, got, want)
	}
}
