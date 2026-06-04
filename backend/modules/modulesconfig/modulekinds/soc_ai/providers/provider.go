// Package providers ports the SOC AI provider verifiers from the
// modules-config plugin. Each provider mirrors a generative-AI vendor
// (OpenAI, Anthropic, Azure, Gemini, Ollama, Mistral, DeepSeek, Groq, plus a
// Custom catch-all) and runs a minimal "only say ok" completion request to
// confirm credentials and endpoint are reachable before the user saves a
// config.
package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"

	"github.com/utmstack/utmstack/backend/modules/modulesconfig/modulekinds/baseline"
)

// IProvider is implemented by each per-vendor verifier; returning a nil error
// means the credentials succeeded against the provider's chat endpoint.
type IProvider interface {
	Validate(ctx context.Context) error
}

// AbstractProvider holds the four fields every vendor verifier needs and
// implements the shared test-request shape (POST a 4-token "only say ok"
// completion). Per-vendor structs embed it and override Validate.
type AbstractProvider struct {
	URL           string
	Model         string
	AuthType      string
	CustomHeaders map[string]string
}

// NewAbstractProvider is the per-vendor convenience constructor; the URL is
// either hardcoded by the wrapping provider (OpenAI, Anthropic, …) or taken
// from the user-supplied config (Azure, Ollama, Custom).
func NewAbstractProvider(URL, Model, AuthType string, CustomHeaders map[string]string) *AbstractProvider {
	return &AbstractProvider{
		URL:           URL,
		Model:         Model,
		AuthType:      AuthType,
		CustomHeaders: CustomHeaders,
	}
}

// validateAuthType normalises blank to "none" and rejects any value the
// abstract layer doesn't understand. Per-vendor Validate methods call this
// first and short-circuit on error.
func (p *AbstractProvider) validateAuthType() error {
	if p.AuthType == "" {
		p.AuthType = "none"
	}
	if p.AuthType != "custom-headers" && p.AuthType != "none" {
		return fmt.Errorf("Invalid authentication type '%s'. Must be 'custom-headers' or 'none'.", p.AuthType)
	}
	return nil
}

// performTestRequest issues the shared "only say ok" completion POST and
// returns (status, body, err). Per-vendor testConnection methods inspect
// status plus err.Error() to map vendor-specific failure modes back to
// user-facing messages.
func (p *AbstractProvider) performTestRequest(ctx context.Context) (int, string, error) {
	headers := map[string]string{"Content-Type": "application/json"}
	if p.AuthType == "custom-headers" {
		maps.Copy(headers, p.CustomHeaders)
	}

	body, _ := json.Marshal(map[string]any{
		"model": p.Model,
		"messages": []map[string]string{
			{"role": "user", "content": "only say ok"},
		},
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.URL, bytes.NewReader(body))
	if err != nil {
		return 0, "", err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := baseline.ValidateHTTPClient.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(respBody), nil
}
