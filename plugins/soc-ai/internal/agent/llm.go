package agent

import (
	"context"
	"encoding/json"
	"errors"
)

var ErrLLMRateLimited = errors.New("rate limit reached — upgrade to Enterprise for higher usage limits")

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

type ToolSpec struct {
	Name        string
	Description string
	InputSchema map[string]any
	ReadOnly    bool
}

type ToolCall struct {
	ID   string
	Name string
	Args json.RawMessage
}

type ToolResult struct {
	ID      string // matches ToolCall.ID
	Name    string
	Content string
	IsError bool
}

type Message struct {
	Role       Role
	Content    string
	ToolCalls  []ToolCall  // set on assistant turns that call tools
	ToolResult *ToolResult // set on tool turns (one per message)
}

type CompletionRequest struct {
	System    string
	Messages  []Message
	Tools     []ToolSpec
	Model     string
	MaxTokens int
}

type CompletionResponse struct {
	Content   string
	ToolCalls []ToolCall
	Stop      string // provider stop/finish reason
}

type LLMClient interface {
	Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error)
	Provider() string
}

const llmTimeoutSec = 120 // per LLM HTTP call

type LLMConfig struct {
	Provider       string // "anthropic" → Anthropic wire format; anything else → OpenAI-compatible
	URL            string // full endpoint
	Model          string
	APIKey         string
	AuthType       string // bearer | header | none
	AuthHeaderName string // header name when AuthType == "header"
	CustomHeaders  map[string]string
}

func NewLLMClient(cfg LLMConfig) LLMClient {
	if cfg.Provider == "anthropic" {
		return &anthropicClient{url: cfg.URL, model: cfg.Model, apiKey: cfg.APIKey, authType: cfg.AuthType, authHeaderName: cfg.AuthHeaderName, customHeaders: cfg.CustomHeaders}
	}
	return &openaiClient{provider: cfg.Provider, url: cfg.URL, model: cfg.Model, apiKey: cfg.APIKey, authType: cfg.AuthType, authHeaderName: cfg.AuthHeaderName, customHeaders: cfg.CustomHeaders}
}
