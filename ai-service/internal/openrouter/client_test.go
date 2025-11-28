// Package openrouter provides tests for the OpenRouter client.
//
// These tests cover:
// - Client initialization
// - Message formatting
// - Rate limiting logic
// - Token estimation
//
// Note: Integration tests with actual OpenRouter API should be run separately
// with valid API credentials and are not included in unit tests.
package openrouter

import (
	"context"
	"testing"
	"time"

	"github.com/xiiisorate/granula_api/shared/pkg/config"
	"github.com/xiiisorate/granula_api/shared/pkg/logger"
)

// createTestConfig creates a test configuration.
func createTestConfig() config.OpenRouterConfig {
	return config.OpenRouterConfig{
		APIKey:          "test-api-key",
		BaseURL:         "https://openrouter.ai/api/v1",
		Model:           "anthropic/claude-sonnet-4",
		Timeout:         30 * time.Second,
		MaxTokens:       4096,
		Temperature:     0.7,
		MaxRetries:      3,
		RateLimitPerMin: 60,
	}
}

// createTestLogger creates a test logger.
func createTestLogger(t *testing.T) *logger.Logger {
	log, err := logger.New(logger.Config{
		Level:       "debug",
		Format:      "json",
		ServiceName: "test",
	})
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	return log
}

// TestNewClient verifies client initialization.
func TestNewClient(t *testing.T) {
	t.Parallel()

	cfg := createTestConfig()
	log := createTestLogger(t)

	client := NewClient(cfg, log)

	if client == nil {
		t.Fatal("expected non-nil client")
	}

	if client.httpClient == nil {
		t.Error("expected non-nil http client")
	}

	if client.cfg.APIKey != cfg.APIKey {
		t.Errorf("expected API key %s, got %s", cfg.APIKey, client.cfg.APIKey)
	}

	if client.cfg.Model != cfg.Model {
		t.Errorf("expected model %s, got %s", cfg.Model, client.cfg.Model)
	}
}

// TestEstimateTokens verifies token estimation.
func TestEstimateTokens(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{
			name:     "empty_string",
			text:     "",
			expected: 0,
		},
		{
			name:     "short_text",
			text:     "Hello",
			expected: 1, // 5/3 = 1
		},
		{
			name:     "medium_text",
			text:     "This is a test message for token estimation",
			expected: 14, // 44/3 = 14
		},
		{
			name:     "russian_text",
			text:     "Привет мир",
			expected: 6, // ~19 bytes / 3 = 6
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EstimateTokens(tt.text)
			// Allow some margin for estimation
			if result != tt.expected && result != tt.expected+1 && result != tt.expected-1 {
				t.Errorf("expected ~%d tokens, got %d for text '%s'", tt.expected, result, tt.text)
			}
		})
	}
}

// TestRateLimiting verifies rate limiting behavior.
func TestRateLimiting(t *testing.T) {
	t.Parallel()

	cfg := createTestConfig()
	cfg.RateLimitPerMin = 2 // Low limit for testing
	log := createTestLogger(t)

	client := NewClient(cfg, log)

	// Simulate recording requests
	client.recordRequest()
	client.recordRequest()

	// Check that we're at the limit
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// This should timeout waiting for rate limit
	err := client.waitForRateLimit(ctx)
	if err != context.DeadlineExceeded {
		t.Errorf("expected deadline exceeded when at rate limit, got: %v", err)
	}
}

// TestRateLimitCleanup verifies old requests are cleaned up.
func TestRateLimitCleanup(t *testing.T) {
	t.Parallel()

	cfg := createTestConfig()
	log := createTestLogger(t)

	client := NewClient(cfg, log)

	// Add some old request times (more than 1 minute ago)
	// These will be cleaned up on next waitForRateLimit call
	oldTimes := []time.Time{
		time.Now().Add(-2 * time.Minute),
		time.Now().Add(-90 * time.Second),
	}
	client.requestTimes = oldTimes

	// Wait for rate limit should clean up old requests
	ctx := context.Background()
	err := client.waitForRateLimit(ctx)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// After cleanup, old requests should be removed
	// Use len() safely as waitForRateLimit already released the lock
	if len(client.requestTimes) != 0 {
		t.Errorf("expected old requests to be cleaned up, still have %d", len(client.requestTimes))
	}
}

// TestMessageStructure verifies message structure.
func TestMessageStructure(t *testing.T) {
	t.Parallel()

	msg := Message{
		Role:    "user",
		Content: "Hello, world!",
	}

	if msg.Role != "user" {
		t.Errorf("expected role 'user', got '%s'", msg.Role)
	}

	if msg.Content != "Hello, world!" {
		t.Errorf("expected content 'Hello, world!', got '%s'", msg.Content)
	}
}

// TestChatOptions verifies chat options defaults.
func TestChatOptions(t *testing.T) {
	t.Parallel()

	opts := ChatOptions{}

	if opts.Model != "" {
		t.Errorf("expected empty default model, got '%s'", opts.Model)
	}

	if opts.MaxTokens != 0 {
		t.Errorf("expected 0 default max tokens, got %d", opts.MaxTokens)
	}

	if opts.Temperature != 0 {
		t.Errorf("expected 0 default temperature, got %f", opts.Temperature)
	}
}

// TestChatRequest verifies chat request structure.
func TestChatRequest(t *testing.T) {
	t.Parallel()

	req := ChatRequest{
		Model: "test-model",
		Messages: []Message{
			{Role: "system", Content: "You are a helpful assistant"},
			{Role: "user", Content: "Hello"},
		},
		MaxTokens:   1000,
		Temperature: 0.7,
		Stream:      false,
	}

	if len(req.Messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(req.Messages))
	}

	if req.Messages[0].Role != "system" {
		t.Errorf("expected first message role 'system', got '%s'", req.Messages[0].Role)
	}
}

// BenchmarkEstimateTokens benchmarks token estimation.
func BenchmarkEstimateTokens(b *testing.B) {
	text := "This is a sample text for benchmarking the token estimation function. " +
		"It contains multiple sentences to simulate a realistic message."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EstimateTokens(text)
	}
}

// BenchmarkRateLimitCheck benchmarks rate limit checking.
func BenchmarkRateLimitCheck(b *testing.B) {
	cfg := createTestConfig()
	log, _ := logger.New(logger.Config{
		Level:       "error",
		Format:      "json",
		ServiceName: "bench",
	})

	client := NewClient(cfg, log)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.waitForRateLimit(ctx)
	}
}
