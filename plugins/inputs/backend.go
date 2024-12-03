package main

import (
	"fmt"
	"io"
	"net/http"

	go_sdk "github.com/threatwinds/go-sdk"
)

func createPanelRequest(method string, endpoint string) (*http.Request, error) {
	pConfig := go_sdk.PluginCfg("com.utmstack", false)
	backend := pConfig.Get("backend").String()
	internalKey := pConfig.Get("internalKey").String()

	url := fmt.Sprintf(endpoint, backend)

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add(panelAPIKeyHeader, internalKey)

	return req, nil
}

func GetConnectionKey() ([]byte, error) {
	client := &http.Client{}

	req, err := createPanelRequest("GET", panelConnectionKeyEndpoint)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			go_sdk.Logger().ErrorF(err.Error())
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get connection key, received status code: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
