package main

import (
	"fmt"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"io"
	"net/http"
)

func createPanelRequest(method string, endpoint string) (*http.Request, error) {
	pConfig := plugins.PluginCfg("com.utmstack", false)
	backend := pConfig.Get("backend").String()
	internalKey := pConfig.Get("internalKey").String()

	url := fmt.Sprintf(endpoint, backend)

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, catcher.Error("cannot create request", err, nil)
	}

	req.Header.Add(panelAPIKeyHeader, internalKey)

	return req, nil
}

func GetConnectionKey() ([]byte, error) {
	client := &http.Client{}

	req, err := createPanelRequest("GET", panelConnectionKeyEndpoint)
	if err != nil {
		return nil, catcher.Error("cannot create request", err, nil)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, catcher.Error("cannot send request", err, nil)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, catcher.Error("cannot get connection key", nil, map[string]any{
			"status": resp.StatusCode,
		})
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
