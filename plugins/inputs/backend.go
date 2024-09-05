package main

import (
	"fmt"
	"io"
	"net/http"

	go_sdk "github.com/threatwinds/go-sdk"
	"github.com/threatwinds/logger"
)

func createPanelRequest(method string, endpoint string) (*http.Request, *logger.Error) {
	pCfg, e := go_sdk.PluginCfg[PluginConfig]("com.utmstack")
	if e != nil {
		return nil, e
	}

	url := fmt.Sprintf(endpoint, pCfg.Backend)

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, go_sdk.Logger().ErrorF(err.Error())
	}

	req.Header.Add(panelAPIKeyHeader, pCfg.InternalKey)

	return req, nil
}

func GetConnectionKey() ([]byte, *logger.Error) {
	client := &http.Client{}

	req, e := createPanelRequest("GET", panelConnectionKeyEndpoint)
	if e != nil {
		return nil, e
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, go_sdk.Logger().ErrorF(err.Error())
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			go_sdk.Logger().ErrorF(err.Error())
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, go_sdk.Logger().ErrorF("failed to get connection key, received status code: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, go_sdk.Logger().ErrorF(err.Error())
	}

	return body, nil
}
