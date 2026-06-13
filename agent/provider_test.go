package agent

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
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
