package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
)

const internalKeyHeader = "X-Internal-Key"

var (
	backendURL  string
	internalKey string
	httpClient  = &http.Client{Timeout: 30 * time.Second}
)

type flexString string

func (f *flexString) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		*f = flexString(s)
		return nil
	}
	var n json.Number
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	}
	*f = flexString(n.String())
	return nil
}

type ingestUser struct {
	TenantID         string      `json:"tenantId"`
	Source           string      `json:"source,omitempty"`
	SID              string      `json:"sid,omitempty"`
	SamAccountName   string      `json:"samAccountName,omitempty"`
	Domain           string      `json:"domain,omitempty"`
	MachineID        *string     `json:"machineId,omitempty"`
	UIDNumber        *flexString `json:"uidNumber,omitempty"`
	Hostname         *string     `json:"hostname,omitempty"`
	Username         *string     `json:"username,omitempty"`
	Active           *bool       `json:"active"`
	AccountCreatedAt *time.Time  `json:"accountCreatedAt"`
	LastLogon        *time.Time  `json:"lastLogon"`
	AccountDeletedAt *time.Time  `json:"accountDeletedAt"`
	LastSeen         *time.Time  `json:"lastSeen"`
}

func loadBackendConfig() {
	cfg := plugins.PluginCfg("com.utmstack")
	backendURL = strings.TrimRight(cfg.Get("backend").String(), "/")
	internalKey = cfg.Get("internalKey").String()
}

func request(method, path string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest(method, backendURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set(internalKeyHeader, internalKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return httpClient.Do(req)
}

func seedCacheFromBackend() {
	seedWindows()
	seedLinux()
	seedCloud()
}

type resolveIntent struct {
	TenantID  string
	Hostname  string
	MachineID string
}

func flushLoop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			flush()
			return
		case <-ticker.C:
			flush()
		}
	}
}

func flush() {
	cacheMu.Lock()
	batch := make([]ingestUser, 0, len(windowsDirty)+len(linuxDirty)+len(cloudDirty))
	windowsIDs := make([]string, 0, len(windowsDirty))
	linuxIDs := make([]string, 0, len(linuxDirty))
	cloudIDs := make([]string, 0, len(cloudDirty))

	for id := range windowsDirty {
		if cu := windowsCache[id]; cu != nil {
			batch = append(batch, cu.toIngest())
			windowsIDs = append(windowsIDs, id)
		}
	}
	for id := range linuxDirty {
		if cu := linuxCache[id]; cu != nil {
			batch = append(batch, cu.toIngest())
			linuxIDs = append(linuxIDs, id)
		}
	}
	for id := range cloudDirty {
		if cu := cloudCache[id]; cu != nil {
			batch = append(batch, cu.toIngest())
			cloudIDs = append(cloudIDs, id)
		}
	}
	windowsDirty = map[string]struct{}{}
	linuxDirty = map[string]struct{}{}
	cloudDirty = map[string]struct{}{}
	cacheMu.Unlock()

	if len(batch) > 0 {
		if err := postBatch(batch); err != nil {
			_ = catcher.Error("ad-audit: flush failed; re-queuing", err, map[string]any{"count": len(batch)})
			cacheMu.Lock()
			for _, id := range windowsIDs {
				windowsDirty[id] = struct{}{}
			}
			for _, id := range linuxIDs {
				linuxDirty[id] = struct{}{}
			}
			for _, id := range cloudIDs {
				cloudDirty[id] = struct{}{}
			}
			cacheMu.Unlock()
		}
	}

	resolveMu.Lock()
	intents := resolveQueue
	resolveQueue = nil
	resolveMu.Unlock()

	for _, intent := range intents {
		if err := resolveLinuxIdentity(intent.TenantID, intent.Hostname, intent.MachineID); err != nil {
			_ = catcher.Error("ad-audit: resolve failed", err, map[string]any{
				"hostname":  intent.Hostname,
				"machineID": intent.MachineID,
			})
			continue
		}
		cacheMu.Lock()
		provisionalPrefix := intent.TenantID + ":linux:provisional:" + intent.Hostname + ":"
		for k, cu := range linuxCache {
			if !strings.HasPrefix(k, provisionalPrefix) {
				continue
			}
			if cu.UIDNumber == "" {
				continue
			}
			newKey := intent.TenantID + ":linux:resolved:" + intent.MachineID + ":" + cu.UIDNumber
			if _, exists := linuxCache[newKey]; exists {
				delete(linuxCache, k)
				continue
			}
			cu.MachineID = intent.MachineID
			linuxCache[newKey] = cu
			delete(linuxCache, k)
			linuxDirty[newKey] = struct{}{}
		}
		cacheMu.Unlock()
	}
}

func postBatch(users []ingestUser) error {
	body, err := json.Marshal(map[string]any{"users": users})
	if err != nil {
		return err
	}
	resp, err := request(http.MethodPost, "/api/v1/ad-audit/users", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("ingest returned status %d", resp.StatusCode)
	}
	return nil
}

func enqueueResolveIntent(tenantID, hostname, machineID string) {
	resolveMu.Lock()
	resolveQueue = append(resolveQueue, resolveIntent{
		TenantID:  tenantID,
		Hostname:  hostname,
		MachineID: machineID,
	})
	resolveMu.Unlock()
}

var resolveLinuxIdentityFn = resolveLinuxIdentityDefault

func resolveLinuxIdentity(tenantID, hostname, machineID string) error {
	return resolveLinuxIdentityFn(tenantID, hostname, machineID)
}

func resolveLinuxIdentityDefault(tenantID, hostname, machineID string) error {
	if backendURL == "" || tenantID == "" || hostname == "" || machineID == "" {
		return nil
	}
	body, err := json.Marshal(map[string]string{
		"tenant_id":  tenantID,
		"hostname":   hostname,
		"machine_id": machineID,
	})
	if err != nil {
		return err
	}
	resp, err := request(http.MethodPost, "/api/v1/ad-audit/users/resolve", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("resolve returned status %d", resp.StatusCode)
	}
	return nil
}

func decodeSeed(r io.Reader, source string) ([]ingestUser, error) {
	var users []ingestUser

	dec := json.NewDecoder(r)
	for {
		var u ingestUser
		err := dec.Decode(&u)
		if errors.Is(err, io.EOF) {
			return users, nil
		}
		if err != nil {
			_ = catcher.Error("ad-audit: decoding the seed stream failed", err, map[string]any{"source": source})
			return nil, err
		}
		users = append(users, u)
	}
}
