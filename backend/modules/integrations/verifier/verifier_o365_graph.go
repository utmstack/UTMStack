package verifier

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
)

func verifyO365Graph(c map[string]string) error {
	clientId := c["o365_graph_client_id"]
	clientSecret := c["o365_graph_client_secret"]
	tenantId := c["o365_graph_tenant_id"]
	cloudEnv := c["o365_graph_cloud_environment"]

	if strings.TrimSpace(clientId) == "" || strings.TrimSpace(clientSecret) == "" || strings.TrimSpace(tenantId) == "" {
		return fmt.Errorf("o365_graph: client ID, client secret and tenant ID are required")
	}

	loginAuthority, graphEndpoint := o365GraphEndpoints(cloudEnv)

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", clientId)
	data.Set("client_secret", clientSecret)
	data.Set("scope", graphEndpoint+"/.default")

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s/oauth2/v2.0/token", loginAuthority, tenantId), strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("o365_graph: failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("o365_graph: token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	var tokenResp struct {
		TokenType   string `json:"token_type"`
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}
	_ = json.Unmarshal(body, &tokenResp)

	if resp.StatusCode != http.StatusOK || tokenResp.AccessToken == "" {
		return fmt.Errorf("o365_graph: token request failed with status %d: %s", resp.StatusCode, tokenResp.ErrorDesc)
	}

	// Probe alerts_v2. A 401/403/400 means the token works but the
	// application lacks the permission or the tenant lacks the license —
	// surfaced as a warning, not a failure, because the tenant may not have
	// M365 E5 / EDR. Any other non-2xx is a hard failure.
	alertsReq, err := http.NewRequest("GET", graphEndpoint+"/v1.0/security/alerts_v2?$top=1", nil)
	if err != nil {
		return fmt.Errorf("o365_graph: failed to create probe request: %w", err)
	}
	alertsReq.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
	alertsResp, err := client.Do(alertsReq)
	if err != nil {
		return fmt.Errorf("o365_graph: alerts_v2 probe failed: %w", err)
	}
	defer alertsResp.Body.Close()

	switch {
	case alertsResp.StatusCode == http.StatusOK:
		return nil
	case alertsResp.StatusCode == http.StatusForbidden || alertsResp.StatusCode == http.StatusUnauthorized || alertsResp.StatusCode == http.StatusBadRequest:
		catcher.Warn("o365_graph verifier: token ok but alerts_v2 probe returned a permission/license status", map[string]any{
			"status": alertsResp.StatusCode,
		})
		return nil
	default:
		return fmt.Errorf("o365_graph: alerts_v2 probe failed with status %d", alertsResp.StatusCode)
	}
}

// o365GraphEndpoints returns the login authority and graph endpoint for a
// cloud environment. The plugin keeps the full table in its own config; the
// verifier only needs the two fields above, so the map is duplicated here to
// keep the verifier free of a plugin dependency.
func o365GraphEndpoints(cloudEnv string) (login, graph string) {
	switch cloudEnv {
	case "GCC":
		return "https://login.microsoftonline.com", "https://graph.microsoft.com"
	case "GCCHigh":
		return "https://login.microsoftonline.us", "https://graph.microsoft.us"
	case "DoD":
		return "https://login.microsoftonline.us", "https://dod-graph.microsoft.us"
	default:
		return "https://login.microsoftonline.com", "https://graph.microsoft.com"
	}
}
