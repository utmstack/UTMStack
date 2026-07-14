package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/utmstack/UTMStack/plugins/soc-ai/utils"
)

type openaiClient struct {
	provider       string
	url            string
	model          string
	apiKey         string
	authType       string // bearer | header | none
	authHeaderName string
	customHeaders  map[string]string
}

func (c *openaiClient) Provider() string { return c.provider }

type oaFunc struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON encoded as a string
}

type oaToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function oaFunc `json:"function"`
}

type oaMessage struct {
	Role       string       `json:"role"`
	Content    any          `json:"content,omitempty"` // string, or nil for tool-call-only turns
	ToolCalls  []oaToolCall `json:"tool_calls,omitempty"`
	ToolCallID string       `json:"tool_call_id,omitempty"`
}

type oaToolDef struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Parameters  any    `json:"parameters"`
}

type oaTool struct {
	Type     string    `json:"type"`
	Function oaToolDef `json:"function"`
}

type oaRequest struct {
	Model     string      `json:"model"`
	Messages  []oaMessage `json:"messages"`
	Tools     []oaTool    `json:"tools,omitempty"`
	MaxTokens int         `json:"max_tokens,omitempty"`
}

type oaResponse struct {
	Choices []struct {
		Message struct {
			Content   string       `json:"content"`
			ToolCalls []oaToolCall `json:"tool_calls"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (c *openaiClient) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	msgs := make([]oaMessage, 0, len(req.Messages)+1)
	if req.System != "" {
		msgs = append(msgs, oaMessage{Role: "system", Content: req.System})
	}
	for _, m := range req.Messages {
		switch m.Role {
		case RoleUser:
			msgs = append(msgs, oaMessage{Role: "user", Content: m.Content})
		case RoleAssistant:
			om := oaMessage{Role: "assistant"}
			if m.Content != "" {
				om.Content = m.Content
			}
			for _, tc := range m.ToolCalls {
				args := string(tc.Args)
				if args == "" {
					args = "{}"
				}
				om.ToolCalls = append(om.ToolCalls, oaToolCall{
					ID:       tc.ID,
					Type:     "function",
					Function: oaFunc{Name: tc.Name, Arguments: args},
				})
			}
			msgs = append(msgs, om)
		case RoleTool:
			if m.ToolResult != nil {
				msgs = append(msgs, oaMessage{Role: "tool", ToolCallID: m.ToolResult.ID, Content: m.ToolResult.Content})
			}
		}
	}

	tools := make([]oaTool, 0, len(req.Tools))
	for _, t := range req.Tools {
		params := any(t.InputSchema)
		if t.InputSchema == nil {
			params = map[string]any{"type": "object", "properties": map[string]any{}}
		}
		tools = append(tools, oaTool{Type: "function", Function: oaToolDef{Name: t.Name, Description: t.Description, Parameters: params}})
	}

	body, err := json.Marshal(oaRequest{Model: req.Model, Messages: msgs, Tools: tools, MaxTokens: req.MaxTokens})
	if err != nil {
		return CompletionResponse{}, fmt.Errorf("marshal request: %w", err)
	}

	respBody, status, err := utils.DoReq(c.url, body, http.MethodPost, c.headers(), llmTimeoutSec)
	if err != nil {
		return CompletionResponse{}, err
	}
	if status == http.StatusUnauthorized {
		return CompletionResponse{}, fmt.Errorf("LLM auth failed (401): check api_key/auth_type")
	}
	if status == http.StatusTooManyRequests {
		return CompletionResponse{}, ErrLLMRateLimited
	}

	var parsed oaResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return CompletionResponse{}, fmt.Errorf("decode response (status %d): %w", status, err)
	}
	if parsed.Error != nil {
		return CompletionResponse{}, fmt.Errorf("LLM error: %s", parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 {
		return CompletionResponse{}, fmt.Errorf("LLM returned no choices (status %d): %s", status, truncate(string(respBody), 300))
	}

	ch := parsed.Choices[0]
	out := CompletionResponse{Content: ch.Message.Content, Stop: ch.FinishReason}
	for _, tc := range ch.Message.ToolCalls {
		out.ToolCalls = append(out.ToolCalls, ToolCall{
			ID:   tc.ID,
			Name: tc.Function.Name,
			Args: json.RawMessage(tc.Function.Arguments),
		})
	}
	return out, nil
}

func (c *openaiClient) headers() map[string]string {
	h := map[string]string{"Content-Type": "application/json"}
	switch c.authType {
	case "none":
		// no auth header
	case "header":
		name := c.authHeaderName
		if name == "" {
			name = "Authorization"
		}
		h[name] = c.apiKey
	default: // bearer
		h["Authorization"] = "Bearer " + c.apiKey
	}
	for k, v := range c.customHeaders {
		h[k] = v
	}
	return h
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
