package main

import (
	"fmt"
	"io"
	"net/http"

	go_sdk "github.com/threatwinds/go-sdk"
)

func createPanelRequest(method string, endpoint string) (*http.Request, error) {
	pCfg, err := go_sdk.PluginCfg[PluginConfig]("com.utmstack")
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf(endpoint, pCfg.Backend)

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add(panelAPIKeyHeader, pCfg.InternalKey)

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
