package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	appconfig_connectors "github.com/utmstack/utmstack/backend/modules/appconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/network_scan/connectors"
	"github.com/utmstack/utmstack/backend/modules/network_scan/domain"
)

const probeAPIURLPrefix = "utmstack.networkScan.apiUrl."

type httpProbeClient struct {
	store  appconfig_connectors.Store
	client *http.Client
}

func NewHTTPProbeClient(store appconfig_connectors.Store, timeout time.Duration) connectors.ProbeClient {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &httpProbeClient{
		store:  store,
		client: &http.Client{Timeout: timeout},
	}
}

func (h *httpProbeClient) baseURL(ctx context.Context, probeID uint64) (string, error) {
	key := fmt.Sprintf("%s%d", probeAPIURLPrefix, probeID)
	v, ok, err := h.store.GetString(ctx, key)
	if err != nil {
		return "", err
	}
	if !ok || strings.TrimSpace(v) == "" {
		return "", domain.ErrProbeNotConfigured
	}
	return strings.TrimRight(v, "/"), nil
}

func (h *httpProbeClient) Scan(
	ctx context.Context,
	probeID uint64,
	iface string,
	networkRange string,
) ([]domain.NetworkScan, error) {
	base, err := h.baseURL(ctx, probeID)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("%s/api/v1/scanner/scan?interface=%s&network=%s",
		base, url.QueryEscape(iface), url.QueryEscape(networkRange))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", domain.ErrProbeUnavailable, err.Error())
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("%w: probe returned %d: %s", domain.ErrProbeUnavailable, resp.StatusCode, string(body))
	}
	var out []domain.NetworkScan
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("probe scan: parse response: %w", err)
	}
	return out, nil
}

func (h *httpProbeClient) Ping(ctx context.Context, probeID uint64) error {
	base, err := h.baseURL(ctx, probeID)
	if err != nil {
		return err
	}
	u := base + "/api/v1/checks/ping"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %s", domain.ErrProbeUnavailable, err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("%w: ping returned %d", domain.ErrProbeUnavailable, resp.StatusCode)
	}
	return nil
}

func (h *httpProbeClient) CheckInterface(ctx context.Context, probeID uint64, iface string) (bool, error) {
	base, err := h.baseURL(ctx, probeID)
	if err != nil {
		return false, err
	}
	u := fmt.Sprintf("%s/api/v1/checks/check_interface?interface=%s", base, url.QueryEscape(iface))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return false, err
	}
	resp, err := h.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("%w: %s", domain.ErrProbeUnavailable, err.Error())
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return false, fmt.Errorf("%w: check_interface returned %d", domain.ErrProbeUnavailable, resp.StatusCode)
	}
	s := strings.TrimSpace(string(body))
	switch strings.ToLower(s) {
	case "true":
		return true, nil
	case "false":
		return false, nil
	}
	var b bool
	if err := json.Unmarshal([]byte(s), &b); err != nil {
		return false, errors.New("probe check_interface: unexpected response: " + s)
	}
	return b, nil
}
