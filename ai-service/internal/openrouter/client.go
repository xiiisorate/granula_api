// Package openrouter provides a client for OpenRouter API.
//
// OpenRouter is a unified API for accessing various LLM providers.
// This client supports:
// - Chat completions (sync and streaming)
// - Retry logic with exponential backoff
// - Rate limiting
// - Token counting
//
// Example:
//
//	client := openrouter.NewClient(cfg)
//	resp, err := client.ChatCompletion(ctx, messages)
//	// or streaming:
//	stream, err := client.ChatCompletionStream(ctx, messages)
package openrouter

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/xiiisorate/granula_api/shared/pkg/config"
	apperrors "github.com/xiiisorate/granula_api/shared/pkg/errors"
	"github.com/xiiisorate/granula_api/shared/pkg/logger"
)

// Client is the OpenRouter API client.
type Client struct {
	cfg        config.OpenRouterConfig
	httpClient *http.Client
	log        *logger.Logger

	// Rate limiting
	mu           sync.Mutex
	requestTimes []time.Time
}

// NewClient creates a new OpenRouter client.
func NewClient(cfg config.OpenRouterConfig, log *logger.Logger) *Client {
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		log:          log,
		requestTimes: make([]time.Time, 0),
	}
}

// Message represents a chat message.
type Message struct {
	Role    string `json:"role"` // "system", "user", "assistant"
	Content string `json:"content"`
}

// ImageContent for multimodal messages.
type ImageContent struct {
	Type     string    `json:"type"` // "text" or "image_url"
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

// ImageURL for image content.
type ImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"` // "low", "high", "auto"
}

// ChatRequest is the request body for chat completions.
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
	TopP        float64   `json:"top_p,omitempty"`
	Stop        []string  `json:"stop,omitempty"`
}

// ChatResponse is the response from chat completions.
type ChatResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice represents a completion choice.
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage represents token usage.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// StreamChunk represents a streaming response chunk.
type StreamChunk struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []StreamChoice `json:"choices"`
}

// StreamChoice represents a streaming choice.
type StreamChoice struct {
	Index        int          `json:"index"`
	Delta        MessageDelta `json:"delta"`
	FinishReason string       `json:"finish_reason"`
}

// MessageDelta represents incremental message content.
type MessageDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// ChatCompletion performs a synchronous chat completion.
func (c *Client) ChatCompletion(ctx context.Context, messages []Message) (*ChatResponse, error) {
	return c.ChatCompletionWithOptions(ctx, messages, ChatOptions{})
}

// ChatOptions for customizing chat requests.
type ChatOptions struct {
	Model        string
	MaxTokens    int
	Temperature  float64
	SystemPrompt string
}

// ChatCompletionWithOptions performs a chat completion with custom options.
func (c *Client) ChatCompletionWithOptions(ctx context.Context, messages []Message, opts ChatOptions) (*ChatResponse, error) {
	// Wait for rate limit
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, err
	}

	// Prepend system prompt if provided
	if opts.SystemPrompt != "" {
		messages = append([]Message{{Role: "system", Content: opts.SystemPrompt}}, messages...)
	}

	// Build request
	model := c.cfg.Model
	if opts.Model != "" {
		model = opts.Model
	}

	maxTokens := c.cfg.MaxTokens
	if opts.MaxTokens > 0 {
		maxTokens = opts.MaxTokens
	}

	temperature := c.cfg.Temperature
	if opts.Temperature > 0 {
		temperature = opts.Temperature
	}

	req := ChatRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: temperature,
		Stream:      false,
	}

	// Execute with retries
	var lastErr error
	for attempt := 0; attempt <= c.cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := time.Duration(1<<uint(attempt-1)) * 100 * time.Millisecond
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		resp, err := c.doRequest(ctx, req)
		if err == nil {
			return resp, nil
		}

		lastErr = err
		c.log.Warn("OpenRouter request failed, retrying",
			logger.Int("attempt", attempt+1),
			logger.Err(err),
		)
	}

	return nil, apperrors.Wrap(lastErr, "all retries exhausted")
}

// ChatCompletionStream performs a streaming chat completion.
func (c *Client) ChatCompletionStream(ctx context.Context, messages []Message, opts ChatOptions) (<-chan StreamEvent, error) {
	// Wait for rate limit
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, err
	}

	// Prepend system prompt if provided
	if opts.SystemPrompt != "" {
		messages = append([]Message{{Role: "system", Content: opts.SystemPrompt}}, messages...)
	}

	// Build request
	model := c.cfg.Model
	if opts.Model != "" {
		model = opts.Model
	}

	req := ChatRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   c.cfg.MaxTokens,
		Temperature: c.cfg.Temperature,
		Stream:      true,
	}

	// Create event channel
	events := make(chan StreamEvent, 100)

	go func() {
		defer close(events)
		c.doStreamRequest(ctx, req, events)
	}()

	return events, nil
}

// StreamEvent represents a streaming event.
type StreamEvent struct {
	Content      string
	Done         bool
	Error        error
	Usage        *Usage
	FinishReason string
}

// doRequest performs the actual HTTP request.
func (c *Client) doRequest(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, apperrors.Internal("failed to marshal request").WithCause(err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.cfg.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, apperrors.Internal("failed to create request").WithCause(err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	httpReq.Header.Set("HTTP-Referer", "https://granula.ru")
	httpReq.Header.Set("X-Title", "Granula")

	c.log.Debug("sending OpenRouter request",
		logger.String("model", req.Model),
		logger.Int("messages", len(req.Messages)),
	)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, apperrors.Unavailable("openrouter").WithCause(err)
	}
	defer resp.Body.Close()

	// Record request time for rate limiting
	c.recordRequest()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		c.log.Error("OpenRouter error response",
			logger.Int("status", resp.StatusCode),
			logger.String("body", string(bodyBytes)),
		)

		if resp.StatusCode == 429 {
			return nil, apperrors.RateLimited("OpenRouter rate limit exceeded")
		}
		return nil, apperrors.Internalf("OpenRouter error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, apperrors.Internal("failed to decode response").WithCause(err)
	}

	c.log.Debug("OpenRouter response received",
		logger.Int("prompt_tokens", chatResp.Usage.PromptTokens),
		logger.Int("completion_tokens", chatResp.Usage.CompletionTokens),
	)

	return &chatResp, nil
}

// doStreamRequest performs a streaming HTTP request.
func (c *Client) doStreamRequest(ctx context.Context, req ChatRequest, events chan<- StreamEvent) {
	body, err := json.Marshal(req)
	if err != nil {
		events <- StreamEvent{Error: apperrors.Internal("failed to marshal request").WithCause(err)}
		return
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.cfg.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		events <- StreamEvent{Error: apperrors.Internal("failed to create request").WithCause(err)}
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	httpReq.Header.Set("HTTP-Referer", "https://granula.ru")
	httpReq.Header.Set("X-Title", "Granula")
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		events <- StreamEvent{Error: apperrors.Unavailable("openrouter").WithCause(err)}
		return
	}
	defer resp.Body.Close()

	// Record request time
	c.recordRequest()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		events <- StreamEvent{Error: apperrors.Internalf("OpenRouter error: %d - %s", resp.StatusCode, string(bodyBytes))}
		return
	}

	// Parse SSE stream
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		// Parse "data: " prefix
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		// Check for stream end
		if data == "[DONE]" {
			events <- StreamEvent{Done: true}
			return
		}

		// Parse chunk
		var chunk StreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			c.log.Warn("failed to parse stream chunk", logger.String("data", data), logger.Err(err))
			continue
		}

		// Send content
		for _, choice := range chunk.Choices {
			if choice.Delta.Content != "" {
				events <- StreamEvent{Content: choice.Delta.Content}
			}
			if choice.FinishReason != "" {
				events <- StreamEvent{FinishReason: choice.FinishReason, Done: true}
				return
			}
		}
	}

	if err := scanner.Err(); err != nil {
		events <- StreamEvent{Error: apperrors.Internal("stream read error").WithCause(err)}
	}
}

// waitForRateLimit waits if rate limit is exceeded.
func (c *Client) waitForRateLimit(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Clean old requests (older than 1 minute)
	cutoff := time.Now().Add(-time.Minute)
	filtered := make([]time.Time, 0)
	for _, t := range c.requestTimes {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}
	c.requestTimes = filtered

	// Check if we're at the limit
	if len(c.requestTimes) >= c.cfg.RateLimitPerMin {
		// Wait until oldest request expires
		waitTime := c.requestTimes[0].Add(time.Minute).Sub(time.Now())
		if waitTime > 0 {
			c.log.Debug("rate limit reached, waiting", logger.Duration("wait", waitTime.Milliseconds()))
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(waitTime):
			}
		}
	}

	return nil
}

// recordRequest records a request time for rate limiting.
func (c *Client) recordRequest() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.requestTimes = append(c.requestTimes, time.Now())
}

// EstimateTokens provides a rough token estimate for a string.
// Actual tokenization depends on the model, this is an approximation.
func EstimateTokens(text string) int {
	// Rough estimate: ~4 characters per token for English
	// Russian text is typically ~2-3 characters per token
	return len(text) / 3
}
