package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
)

const (
	graphAPIVersion        = "/v1.0"
	o365GraphGrantType     = "client_credentials"
	o365GraphEndPointLogin = "/oauth2/v2.0/token"

	maxRetries = 3
	retryDelay = 2 * time.Second
	reqTimeout = 60 * time.Second
	topParam   = 500

	// Late-arriving records: re-read the last 30s each tick; duplicates are
	// dropped by the deterministic identity downstream.
	overlap     = 30 * time.Second
	maxLookback = 24 * time.Hour
)

type microsoftLoginResponse struct {
	TokenType   string `json:"token_type,omitempty"`
	Expires     int    `json:"expires_in,omitempty"`
	AccessToken string `json:"access_token,omitempty"`
	Error       string `json:"error,omitempty"`
	ErrorDesc   string `json:"error_description,omitempty"`
}

type graphHTTPError struct {
	status     int
	body       []byte
	retryAfter time.Duration
}

func (e *graphHTTPError) Error() string {
	return fmt.Sprintf("graph http %d: %s", e.status, string(e.body))
}

// isPermanent is a license or permission error: retrying will not fix it, so
// the source backs off instead of failing the tick forever.
func isPermanent(err error) bool {
	var he *graphHTTPError
	if !errors.As(err, &he) {
		return false
	}
	if he.status == http.StatusForbidden || he.status == http.StatusNotFound {
		return true
	}
	return he.status == http.StatusBadRequest && strings.Contains(he.Error(), "AADSTS700082")
}

func joinPath(base, path string) string {
	if base != "" && !strings.HasSuffix(base, "/") {
		base += "/"
	}
	if path != "" && strings.HasPrefix(path, "/") {
		path = strings.TrimPrefix(path, "/")
	}
	return base + path
}

// parseRetryAfter accepts the two RFC 7231 forms: delta-seconds or HTTP-date.
func parseRetryAfter(v string) time.Duration {
	if v == "" {
		return 0
	}
	if sec, err := strconv.Atoi(v); err == nil && sec >= 0 {
		return time.Duration(sec) * time.Second
	}
	if t, err := http.ParseTime(v); err == nil {
		if d := time.Until(t); d > 0 {
			return d
		}
	}
	return 0
}

type GraphProcessor struct {
	TenantId         string
	UtmTenantId      string
	ClientId         string
	ClientSecret     string
	CloudEnvironment CloudEnvironment
	CloudConfig      CloudConfig

	httpClient  *http.Client
	accessToken string
	tokenType   string
	tokenExpiry time.Time
	// availableSources holds a source's backoff deadline after a permanent
	// failure, so the error is logged once and the tick is cheap in between.
	availableSources map[string]time.Time
}

func newGraphProcessor(group *ModuleGroup) *GraphProcessor {
	p := &GraphProcessor{
		UtmTenantId:      group.UtmTenantId,
		httpClient:       &http.Client{Timeout: reqTimeout},
		availableSources: make(map[string]time.Time),
	}
	for _, cnf := range group.ModuleGroupConfigurations {
		switch cnf.ConfKey {
		case "o365_graph_client_id":
			p.ClientId = cnf.ConfValue
		case "o365_graph_client_secret":
			p.ClientSecret = cnf.ConfValue
		case "o365_graph_tenant_id":
			p.TenantId = cnf.ConfValue
		case "o365_graph_cloud_environment":
			if cnf.ConfValue != "" {
				p.CloudEnvironment = CloudEnvironment(cnf.ConfValue)
			}
		}
	}
	if p.CloudEnvironment == "" {
		p.CloudEnvironment = CloudCommercial
	}
	p.CloudConfig = GetCloudConfig(p.CloudEnvironment)
	return p
}

func (p *GraphProcessor) sourceUnavailable(source string) bool {
	t, ok := p.availableSources[source]
	return ok && time.Now().UTC().Before(t)
}

func (p *GraphProcessor) markSourceUnavailable(source string) {
	p.availableSources[source] = time.Now().UTC().Add(time.Hour)
}

// getToken refreshes 60s before expiry to avoid using a token at the edge.
func (p *GraphProcessor) getToken() (string, error) {
	if p.accessToken != "" && time.Now().UTC().Before(p.tokenExpiry) {
		return p.accessToken, nil
	}

	requestUrl := joinPath(p.CloudConfig.LoginAuthority, p.TenantId+o365GraphEndPointLogin)
	data := url.Values{}
	data.Set("grant_type", o365GraphGrantType)
	data.Set("client_id", p.ClientId)
	data.Set("client_secret", p.ClientSecret)
	data.Set("scope", p.CloudConfig.GraphScope)

	var result microsoftLoginResponse
	var lastErr error

	for retry := 0; retry < maxRetries; retry++ {
		req, err := http.NewRequest(http.MethodPost, requestUrl, strings.NewReader(data.Encode()))
		if err != nil {
			return "", err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := p.httpClient.Do(req)
		if err != nil {
			lastErr = err
		} else {
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
			_ = resp.Body.Close()
			if err := json.Unmarshal(body, &result); err != nil {
				lastErr = fmt.Errorf("graph token decode: %w", err)
			} else if result.Error != "" {
				// A protocol error (bad secret, no consent, wrong scope) is not
				// retryable: stop at once.
				lastErr = &graphHTTPError{status: resp.StatusCode, body: body}
				break
			} else if result.AccessToken == "" {
				lastErr = errors.New("graph token response missing access_token")
			} else {
				p.tokenType = result.TokenType
				if p.tokenType == "" {
					p.tokenType = "Bearer"
				}
				p.accessToken = result.AccessToken
				p.tokenExpiry = time.Now().UTC().Add(time.Duration(result.Expires) * time.Second).Add(-60 * time.Second)
				return p.accessToken, nil
			}
		}
		_ = catcher.Error("error getting graph auth, retrying", lastErr, map[string]any{
			"process": processName, "retry": retry + 1,
		})
		if retry < maxRetries-1 {
			time.Sleep(retryDelay * time.Duration(1<<uint(retry)))
		}
	}
	return "", fmt.Errorf("graph auth failed after %d retries: %w", maxRetries, lastErr)
}

// doGraphRequest retries 429/5xx (honoring Retry-After) and re-auths once on
// 401; any other 4xx is terminal and returned immediately.
func (p *GraphProcessor) doGraphRequest(method, path string, body []byte) ([]byte, error) {
	token, err := p.getToken()
	if err != nil {
		return nil, err
	}

	requestUrl := joinPath(p.CloudConfig.GraphEndpoint, path)
	var lastErr error

	for retry := 0; retry < maxRetries; retry++ {
		req, err := http.NewRequest(method, requestUrl, strings.NewReader(string(body)))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", fmt.Sprintf("%s %s", p.tokenType, token))
		req.Header.Set("Content-Type", "application/json")

		resp, err := p.httpClient.Do(req)
		if err != nil {
			lastErr = err
		} else {
			respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
			_ = resp.Body.Close()
			if resp.StatusCode >= 400 {
				lastErr = &graphHTTPError{status: resp.StatusCode, body: respBody, retryAfter: parseRetryAfter(resp.Header.Get("Retry-After"))}
				if resp.StatusCode == http.StatusUnauthorized {
					p.accessToken = ""
					token, err = p.getToken()
					if err != nil {
						return nil, err
					}
					continue
				} else if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
					d := lastErr.(*graphHTTPError).retryAfter
					if d <= 0 {
						d = retryDelay * time.Duration(1<<uint(retry))
					}
					time.Sleep(d)
					continue
				} else {
					return respBody, lastErr
				}
			} else {
				return respBody, nil
			}
		}
		_ = catcher.Error("graph request failed, retrying", lastErr, map[string]any{
			"process": processName, "retry": retry + 1, "path": path,
		})
	}
	return nil, fmt.Errorf("graph request %s failed after %d retries: %w", path, maxRetries, lastErr)
}

// resolveNextLink turns an @odata.nextLink — absolute, or "?query"-relative
// to the request that produced it — into the path doGraphRequest expects.
func (p *GraphProcessor) resolveNextLink(base, next string) string {
	if next == "" {
		return ""
	}
	b, err := url.Parse(base)
	if err != nil {
		return next
	}
	n, err := url.Parse(next)
	if err != nil {
		return next
	}
	uri := b.ResolveReference(n).RequestURI()
	if uri == "" || uri == "/" {
		return ""
	}
	return uri
}

func (p *GraphProcessor) getJSONPages(path string) ([]map[string]any, error) {
	items := make([]map[string]any, 0, 100)
	next := path
	for next != "" {
		base := joinPath(p.CloudConfig.GraphEndpoint, next)
		body, err := p.doGraphRequest(http.MethodGet, next, nil)
		if err != nil {
			return nil, err
		}
		var page struct {
			NextLink string           `json:"@odata.nextLink"`
			Value    []map[string]any `json:"value"`
		}
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, fmt.Errorf("graph response decode: %w", err)
		}
		items = append(items, page.Value...)
		next = p.resolveNextLink(base, page.NextLink)
	}
	return items, nil
}

func sourceWindow(position, windowEnd time.Time) (time.Time, time.Time) {
	end := windowEnd
	start := end.Add(-maxLookback)
	if !position.IsZero() {
		s := position.Add(-overlap)
		if s.After(start) {
			start = s
		}
	}
	return start, end
}

func odata(t time.Time) string { return t.UTC().Format("2006-01-02T15:04:05Z") }
