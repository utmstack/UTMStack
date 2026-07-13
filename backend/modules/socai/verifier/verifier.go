package verifier

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"strings"
	"time"
)

var providerDefaultURLs = map[string]string{
	"openai":    "https://api.openai.com/v1/chat/completions",
	"anthropic": "https://api.anthropic.com/v1/messages",
	"azure":     "",
	"gemini":    "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions",
	"ollama":    "http://localhost:11434/v1/chat/completions",
	"mistral":   "https://api.mistral.ai/v1/chat/completions",
	"deepseek":  "https://api.deepseek.com/chat/completions",
	"groq":      "https://api.groq.com/openai/v1/chat/completions",
}

const (
	anthropicVersion = "2023-06-01"
	healthCheckName  = "health_check"
	toolProbePrompt  = "Health check: call the " + healthCheckName + " tool with ok set to true."
)

var healthCheckSchema = map[string]any{
	"type":       "object",
	"properties": map[string]any{"ok": map[string]any{"type": "boolean"}},
	"required":   []string{"ok"},
}

type Config struct {
	Provider       string
	URL            string
	Model          string
	APIKey         string
	AuthType       string // bearer | header | none
	AuthHeaderName string
	CustomHeaders  map[string]string
}

type Verifier struct {
	httpClient *http.Client
}

func New() *Verifier {
	return &Verifier{httpClient: &http.Client{Timeout: 90 * time.Second}}
}

func (v *Verifier) Verify(ctx context.Context, c Config) error {
	url := c.URL
	if url == "" {
		url = providerDefaultURLs[c.Provider]
	}
	if url == "" {
		return fmt.Errorf("url is required for provider %q", c.Provider)
	}
	if err := v.ping(ctx, c, url); err != nil {
		return err
	}
	return v.toolProbe(ctx, c, url)
}

func (v *Verifier) ping(ctx context.Context, c Config, url string) error {
	body, err := json.Marshal(map[string]any{
		"model":      c.Model,
		"max_tokens": 1,
		"messages":   []map[string]any{{"role": "user", "content": "ping"}},
	})
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	status, respBody, err := v.do(ctx, url, headersFor(c), body)
	if err != nil {
		return err
	}
	if status == http.StatusOK || status == http.StatusAccepted {
		return nil
	}
	if status == http.StatusUnauthorized || status == http.StatusForbidden {
		return fmt.Errorf("authentication failed (%d): check the API key / auth headers", status)
	}
	return fmt.Errorf("LLM check failed (%d): %s", status, snippet(respBody))
}

func (v *Verifier) toolProbe(ctx context.Context, c Config, url string) error {
	var (
		body []byte
		err  error
	)
	if c.Provider == "anthropic" {
		body, err = json.Marshal(map[string]any{
			"model":       c.Model,
			"max_tokens":  64,
			"messages":    []map[string]any{{"role": "user", "content": toolProbePrompt}},
			"tools":       []map[string]any{{"name": healthCheckName, "description": "Confirms tool-calling works.", "input_schema": healthCheckSchema}},
			"tool_choice": map[string]any{"type": "tool", "name": healthCheckName},
		})
	} else {
		body, err = json.Marshal(map[string]any{
			"model":      c.Model,
			"max_tokens": 64,
			"messages":   []map[string]any{{"role": "user", "content": toolProbePrompt}},
			"tools": []map[string]any{{"type": "function", "function": map[string]any{
				"name": healthCheckName, "description": "Confirms tool-calling works.", "parameters": healthCheckSchema,
			}}},
			"tool_choice": map[string]any{"type": "function", "function": map[string]any{"name": healthCheckName}},
		})
	}
	if err != nil {
		return fmt.Errorf("tool-calling check: build request: %w", err)
	}

	status, respBody, err := v.do(ctx, url, headersFor(c), body)
	if err != nil {
		return fmt.Errorf("tool-calling check: %w", err)
	}
	if status != http.StatusOK && status != http.StatusAccepted {
		return fmt.Errorf("tool-calling not supported (%d): %s", status, snippet(respBody))
	}
	if hasToolCall(c.Provider, respBody) {
		return nil
	}
	return errors.New("model did not return a tool call — tool-calling appears unsupported by this model/provider")
}

func hasToolCall(provider string, respBody []byte) bool {
	if provider == "anthropic" {
		var r struct {
			Content []struct {
				Type string `json:"type"`
			} `json:"content"`
		}
		_ = json.Unmarshal(respBody, &r)
		for _, b := range r.Content {
			if b.Type == "tool_use" {
				return true
			}
		}
		return false
	}
	var r struct {
		Choices []struct {
			Message struct {
				ToolCalls []struct {
					ID string `json:"id"`
				} `json:"tool_calls"`
			} `json:"message"`
		} `json:"choices"`
	}
	_ = json.Unmarshal(respBody, &r)
	return len(r.Choices) > 0 && len(r.Choices[0].Message.ToolCalls) > 0
}

func (v *Verifier) do(ctx context.Context, url string, headers map[string]string, body []byte) (int, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return 0, nil, fmt.Errorf("build request: %w", err)
	}
	for k, val := range headers {
		req.Header.Set(k, val)
	}
	resp, err := v.httpClient.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("endpoint unreachable: %w", err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 2000))
	return resp.StatusCode, b, nil
}

func headersFor(c Config) map[string]string {
	if c.Provider == "anthropic" {
		return anthropicHeaders(c)
	}
	return openaiHeaders(c)
}

func openaiHeaders(c Config) map[string]string {
	h := map[string]string{"Content-Type": "application/json"}
	switch c.AuthType {
	case "none":
		// no auth header
	case "header":
		name := c.AuthHeaderName
		if name == "" {
			name = "Authorization"
		}
		h[name] = c.APIKey
	default: // bearer
		h["Authorization"] = "Bearer " + c.APIKey
	}
	maps.Copy(h, c.CustomHeaders)
	return h
}

func anthropicHeaders(c Config) map[string]string {
	h := map[string]string{
		"Content-Type":      "application/json",
		"anthropic-version": anthropicVersion,
	}
	if c.AuthType == "header" && c.AuthHeaderName != "" {
		h[c.AuthHeaderName] = c.APIKey
	} else {
		h["x-api-key"] = c.APIKey
	}
	maps.Copy(h, c.CustomHeaders)
	return h
}

func snippet(b []byte) string {
	return strings.TrimSpace(string(b))
}
