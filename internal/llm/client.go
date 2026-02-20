package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	defaultClaudeBaseURL    = "https://api.anthropic.com/v1/messages"
	defaultClaudeModel      = "claude-3-5-haiku-latest"
	defaultAnthropicVersion = "2023-06-01"
	defaultMaxTokens        = 512
	defaultRetryCount       = 3
	defaultMinInterval      = 200 * time.Millisecond
)

// ClientOptions configures a Claude API client.
type ClientOptions struct {
	APIKey             string
	BaseURL            string
	Model              string
	AnthropicVersion   string
	MaxTokens          int
	MaxRetries         int
	MinRequestInterval time.Duration
	HTTPClient         *http.Client
}

// Completion is a normalized LLM completion result.
type Completion struct {
	Text         string
	InputTokens  int
	OutputTokens int
	StopReason   string
	ID           string
}

// Client wraps Claude Messages API with retries and rate limiting.
type Client struct {
	apiKey           string
	baseURL          string
	model            string
	anthropicVersion string
	maxTokens        int
	maxRetries       int
	minInterval      time.Duration
	httpClient       *http.Client

	clockNow func() time.Time
	sleep    func(time.Duration)

	rateMu      sync.Mutex
	lastRequest time.Time
}

// NewClaudeClient creates a configured Claude API client.
func NewClaudeClient(opts ClientOptions) (*Client, error) {
	if strings.TrimSpace(opts.APIKey) == "" {
		return nil, errors.New("llm: API key is required")
	}
	if opts.BaseURL == "" {
		opts.BaseURL = defaultClaudeBaseURL
	}
	if opts.Model == "" {
		opts.Model = defaultClaudeModel
	}
	if opts.AnthropicVersion == "" {
		opts.AnthropicVersion = defaultAnthropicVersion
	}
	if opts.MaxTokens <= 0 {
		opts.MaxTokens = defaultMaxTokens
	}
	if opts.MaxRetries < 0 {
		opts.MaxRetries = defaultRetryCount
	}
	if opts.MinRequestInterval <= 0 {
		opts.MinRequestInterval = defaultMinInterval
	}
	if opts.HTTPClient == nil {
		opts.HTTPClient = &http.Client{Timeout: 15 * time.Second}
	}

	return &Client{
		apiKey:           opts.APIKey,
		baseURL:          opts.BaseURL,
		model:            opts.Model,
		anthropicVersion: opts.AnthropicVersion,
		maxTokens:        opts.MaxTokens,
		maxRetries:       opts.MaxRetries,
		minInterval:      opts.MinRequestInterval,
		httpClient:       opts.HTTPClient,
		clockNow:         time.Now,
		sleep:            time.Sleep,
	}, nil
}

type claudeRequest struct {
	Model     string                 `json:"model"`
	MaxTokens int                    `json:"max_tokens"`
	System    string                 `json:"system,omitempty"`
	Messages  []claudeRequestMessage `json:"messages"`
}

type claudeRequestMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeResponse struct {
	ID         string `json:"id"`
	StopReason string `json:"stop_reason"`
	Content    []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

type claudeErrorResponse struct {
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// Complete sends a prompt to Claude and returns its text response.
func (c *Client) Complete(ctx context.Context, systemPrompt, userPrompt string) (Completion, error) {
	reqBody := claudeRequest{
		Model:     c.model,
		MaxTokens: c.maxTokens,
		System:    strings.TrimSpace(systemPrompt),
		Messages: []claudeRequestMessage{
			{Role: "user", Content: userPrompt},
		},
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return Completion{}, fmt.Errorf("llm: marshal request: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		c.waitForRateLimit()

		comp, shouldRetry, err := c.doRequest(ctx, payload)
		if err == nil {
			return comp, nil
		}
		lastErr = err
		if !shouldRetry || attempt == c.maxRetries {
			break
		}
		c.sleep(backoff(attempt))
	}
	return Completion{}, lastErr
}

func (c *Client) waitForRateLimit() {
	c.rateMu.Lock()
	defer c.rateMu.Unlock()

	now := c.clockNow()
	if !c.lastRequest.IsZero() {
		elapsed := now.Sub(c.lastRequest)
		if elapsed < c.minInterval {
			c.sleep(c.minInterval - elapsed)
		}
	}
	c.lastRequest = c.clockNow()
}

func (c *Client) doRequest(ctx context.Context, payload []byte) (Completion, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(payload))
	if err != nil {
		return Completion{}, false, fmt.Errorf("llm: build request: %w", err)
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", c.anthropicVersion)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Completion{}, true, fmt.Errorf("llm: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return Completion{}, shouldRetryStatus(resp.StatusCode), fmt.Errorf("llm: read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := strings.TrimSpace(string(body))
		var e claudeErrorResponse
		if json.Unmarshal(body, &e) == nil && strings.TrimSpace(e.Error.Message) != "" {
			msg = e.Error.Message
		}
		return Completion{}, shouldRetryStatus(resp.StatusCode), fmt.Errorf("llm: api error status=%d: %s", resp.StatusCode, msg)
	}

	var out claudeResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return Completion{}, false, fmt.Errorf("llm: decode response: %w", err)
	}
	text := extractText(out.Content)
	if strings.TrimSpace(text) == "" {
		return Completion{}, false, errors.New("llm: empty text response")
	}

	return Completion{
		Text:         text,
		InputTokens:  out.Usage.InputTokens,
		OutputTokens: out.Usage.OutputTokens,
		StopReason:   out.StopReason,
		ID:           out.ID,
	}, false, nil
}

func extractText(content []struct {
	Type string `json:"type"`
	Text string `json:"text"`
}) string {
	var parts []string
	for _, c := range content {
		if strings.EqualFold(c.Type, "text") && strings.TrimSpace(c.Text) != "" {
			parts = append(parts, c.Text)
		}
	}
	return strings.Join(parts, "\n")
}

func shouldRetryStatus(status int) bool {
	return status == http.StatusTooManyRequests ||
		status == http.StatusInternalServerError ||
		status == http.StatusBadGateway ||
		status == http.StatusServiceUnavailable ||
		status == http.StatusGatewayTimeout
}

func backoff(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	d := 200 * time.Millisecond
	for i := 0; i < attempt; i++ {
		d *= 2
		if d > 2*time.Second {
			return 2 * time.Second
		}
	}
	if d > 2*time.Second {
		return 2 * time.Second
	}
	return d
}
