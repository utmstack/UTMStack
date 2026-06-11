package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/utmstack/UTMStack/plugins/soc-ai/utils"
)

const anthropicVersion = "2023-06-01"

type anthropicClient struct {
	url            string
	model          string
	apiKey         string
	authType       string
	authHeaderName string
	customHeaders  map[string]string
}

func (c *anthropicClient) Provider() string { return "anthropic" }

type anResponse struct {
	Content []struct {
		Type  string          `json:"type"`
		Text  string          `json:"text"`
		ID    string          `json:"id"`
		Name  string          `json:"name"`
		Input json.RawMessage `json:"input"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
	Error      *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (c *anthropicClient) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	msgs := c.buildMessages(req.Messages)

	tools := make([]map[string]any, 0, len(req.Tools))
	for _, t := range req.Tools {
		schema := any(t.InputSchema)
		if t.InputSchema == nil {
			schema = map[string]any{"type": "object", "properties": map[string]any{}}
		}
		tools = append(tools, map[string]any{"name": t.Name, "description": t.Description, "input_schema": schema})
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}
	payload := map[string]any{
		"model":      req.Model,
		"max_tokens": maxTokens,
		"messages":   msgs,
	}
	if req.System != "" {
		payload["system"] = req.System
	}
	if len(tools) > 0 {
		payload["tools"] = tools
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return CompletionResponse{}, fmt.Errorf("marshal request: %w", err)
	}

	respBody, status, err := utils.DoReq(c.url, body, http.MethodPost, c.headers(), llmTimeoutSec)
	if err != nil {
		return CompletionResponse{}, err
	}
	if status == http.StatusUnauthorized {
		return CompletionResponse{}, fmt.Errorf("LLM auth failed (401): check api_key")
	}

	var parsed anResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return CompletionResponse{}, fmt.Errorf("decode response (status %d): %w", status, err)
	}
	if parsed.Error != nil {
		return CompletionResponse{}, fmt.Errorf("LLM error: %s", parsed.Error.Message)
	}

	out := CompletionResponse{Stop: parsed.StopReason}
	var text strings.Builder
	for _, blk := range parsed.Content {
		switch blk.Type {
		case "text":
			text.WriteString(blk.Text)
		case "tool_use":
			out.ToolCalls = append(out.ToolCalls, ToolCall{ID: blk.ID, Name: blk.Name, Args: blk.Input})
		}
	}
	out.Content = text.String()
	return out, nil
}

func (c *anthropicClient) buildMessages(in []Message) []map[string]any {
	var msgs []map[string]any
	for _, m := range in {
		switch m.Role {
		case RoleUser:
			msgs = append(msgs, map[string]any{
				"role":    "user",
				"content": []map[string]any{{"type": "text", "text": m.Content}},
			})
		case RoleAssistant:
			blocks := make([]map[string]any, 0, len(m.ToolCalls)+1)
			if m.Content != "" {
				blocks = append(blocks, map[string]any{"type": "text", "text": m.Content})
			}
			for _, tc := range m.ToolCalls {
				input := any(tc.Args)
				if len(tc.Args) == 0 {
					input = map[string]any{}
				}
				blocks = append(blocks, map[string]any{"type": "tool_use", "id": tc.ID, "name": tc.Name, "input": input})
			}
			msgs = append(msgs, map[string]any{"role": "assistant", "content": blocks})
		case RoleTool:
			if m.ToolResult == nil {
				continue
			}
			block := map[string]any{"type": "tool_result", "tool_use_id": m.ToolResult.ID, "content": m.ToolResult.Content}
			if m.ToolResult.IsError {
				block["is_error"] = true
			}
			if n := len(msgs); n > 0 && msgs[n-1]["role"] == "user" && isToolResultTurn(msgs[n-1]) {
				msgs[n-1]["content"] = append(msgs[n-1]["content"].([]map[string]any), block)
			} else {
				msgs = append(msgs, map[string]any{"role": "user", "content": []map[string]any{block}})
			}
		}
	}
	return msgs
}

func isToolResultTurn(msg map[string]any) bool {
	blocks, ok := msg["content"].([]map[string]any)
	if !ok || len(blocks) == 0 {
		return false
	}
	return blocks[0]["type"] == "tool_result"
}

func (c *anthropicClient) headers() map[string]string {
	h := map[string]string{
		"Content-Type":      "application/json",
		"anthropic-version": anthropicVersion,
	}
	if c.authType == "header" && c.authHeaderName != "" {
		h[c.authHeaderName] = c.apiKey
	} else {
		h["x-api-key"] = c.apiKey
	}
	for k, v := range c.customHeaders {
		h[k] = v
	}
	return h
}
