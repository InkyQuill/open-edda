package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Provider interface {
	Complete(ctx context.Context, request CompletionRequest) (CompletionResponse, error)
}

type CompletionRequest struct {
	Model           string
	Messages        []CompletionMessage
	Tools           []CompletionTool
	ToolChoice      any
	Temperature     *float64
	MaxOutputTokens int64
}

type CompletionMessage struct {
	Role       MessageRole          `json:"role"`
	Content    string               `json:"content,omitempty"`
	ToolCallID string               `json:"tool_call_id,omitempty"`
	ToolCalls  []CompletionToolCall `json:"tool_calls,omitempty"`
}

type CompletionTool struct {
	Type     string                 `json:"type"`
	Function CompletionToolFunction `json:"function"`
}

type CompletionToolFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

type CompletionToolCall struct {
	ID       string                     `json:"id"`
	Type     string                     `json:"type"`
	Function CompletionToolCallFunction `json:"function"`
}

type CompletionToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type CompletionResponse struct {
	ID             string            `json:"id"`
	Model          string            `json:"model"`
	Message        CompletionMessage `json:"message"`
	FinishReason   string            `json:"finishReason"`
	Usage          Usage             `json:"usage"`
	UsageAvailable bool              `json:"usageAvailable"`
}

type OpenAICompatibleClient struct {
	baseURL      string
	apiKey       string
	modelVariant ModelVariant
	httpClient   *http.Client
}

func NewOpenAICompatibleClient(baseURL, apiKey string, modelVariant ModelVariant) *OpenAICompatibleClient {
	return &OpenAICompatibleClient{
		baseURL:      normalizeBaseURL(baseURL),
		apiKey:       apiKey,
		modelVariant: modelVariant,
		httpClient:   http.DefaultClient,
	}
}

func (c *OpenAICompatibleClient) Complete(ctx context.Context, request CompletionRequest) (CompletionResponse, error) {
	body, err := c.requestBody(request)
	if err != nil {
		return CompletionResponse{}, err
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return CompletionResponse{}, fmt.Errorf("marshal completion request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return CompletionResponse{}, fmt.Errorf("create completion request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return CompletionResponse{}, fmt.Errorf("send completion request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errorBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return CompletionResponse{}, fmt.Errorf("completion request failed: status %d: %s", resp.StatusCode, strings.TrimSpace(string(errorBody)))
	}

	var decoded openAIChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return CompletionResponse{}, fmt.Errorf("decode completion response: %w", err)
	}
	if len(decoded.Choices) == 0 {
		return CompletionResponse{}, fmt.Errorf("decode completion response: missing choices")
	}

	result := CompletionResponse{
		ID:           decoded.ID,
		Model:        decoded.Model,
		Message:      decoded.Choices[0].Message,
		FinishReason: decoded.Choices[0].FinishReason,
	}
	if decoded.Usage != nil {
		result.Usage = normalizeUsage(*decoded.Usage, c.modelVariant)
		result.UsageAvailable = true
	}
	return result, nil
}

func (c *OpenAICompatibleClient) requestBody(request CompletionRequest) (map[string]any, error) {
	model := request.Model
	if model == "" {
		model = c.modelVariant.Model
	}
	if model == "" {
		return nil, fmt.Errorf("completion model is required")
	}

	temperature := request.Temperature
	if temperature == nil {
		temperature = &c.modelVariant.Temperature
	}
	maxOutputTokens := request.MaxOutputTokens
	if maxOutputTokens == 0 {
		maxOutputTokens = c.modelVariant.MaxOutputTokens
	}

	body := map[string]any{
		"model":       model,
		"messages":    request.Messages,
		"temperature": *temperature,
	}
	if len(request.Tools) > 0 {
		body["tools"] = request.Tools
	}
	if request.ToolChoice != nil {
		body["tool_choice"] = request.ToolChoice
	}
	if maxOutputTokens > 0 {
		body[defaultRequestTokenField(c.modelVariant.RequestTokenField)] = maxOutputTokens
	}

	return body, nil
}

func normalizeBaseURL(rawBaseURL string) string {
	trimmed := strings.TrimRight(rawBaseURL, "/")
	parsed, err := url.Parse(trimmed)
	if err == nil && strings.HasSuffix(parsed.Path, "/chat/completions") {
		parsed.Path = strings.TrimSuffix(parsed.Path, "/chat/completions")
		return strings.TrimRight(parsed.String(), "/")
	}
	if err == nil && !strings.HasSuffix(parsed.Path, "/v1") {
		parsed.Path = strings.TrimRight(parsed.Path, "/") + "/v1"
		return strings.TrimRight(parsed.String(), "/")
	}
	return trimmed
}

func normalizeUsage(usage openAIUsage, model ModelVariant) Usage {
	cacheReadTokens := usage.PromptTokensDetails.CachedTokens
	if usage.PromptCacheHitTokens > cacheReadTokens {
		cacheReadTokens = usage.PromptCacheHitTokens
	}
	cacheWriteTokens := usage.PromptTokensDetails.CacheWriteTokens
	inputTokens := usage.PromptTokens - cacheReadTokens - cacheWriteTokens
	if inputTokens < 0 {
		inputTokens = 0
	}

	result := Usage{
		InputTokens:      inputTokens,
		OutputTokens:     usage.CompletionTokens,
		CacheReadTokens:  cacheReadTokens,
		CacheWriteTokens: cacheWriteTokens,
		TotalTokens:      usage.PromptTokens + usage.CompletionTokens,
	}
	result.InputCost = tokenCost(result.InputTokens, model.InputPricePerMillion)
	result.OutputCost = tokenCost(result.OutputTokens, model.OutputPricePerMillion)
	result.CacheReadCost = tokenCost(result.CacheReadTokens, model.CacheReadPricePerMillion)
	result.CacheWriteCost = tokenCost(result.CacheWriteTokens, model.CacheWritePricePerMillion)
	result.TotalCost = result.InputCost + result.OutputCost + result.CacheReadCost + result.CacheWriteCost
	return result
}

func tokenCost(tokens int64, pricePerMillion float64) float64 {
	return float64(tokens) * pricePerMillion / 1_000_000
}

type openAIChatCompletionResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Choices []struct {
		Message      CompletionMessage `json:"message"`
		FinishReason string            `json:"finish_reason"`
	} `json:"choices"`
	Usage *openAIUsage `json:"usage"`
}

type openAIUsage struct {
	PromptTokens         int64 `json:"prompt_tokens"`
	CompletionTokens     int64 `json:"completion_tokens"`
	PromptCacheHitTokens int64 `json:"prompt_cache_hit_tokens"`
	PromptTokensDetails  struct {
		CachedTokens     int64 `json:"cached_tokens"`
		CacheWriteTokens int64 `json:"cache_write_tokens"`
	} `json:"prompt_tokens_details"`
}
