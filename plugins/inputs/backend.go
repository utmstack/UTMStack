package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/threatwinds/go-sdk/helpers"
	"github.com/threatwinds/logger"
)

func createPanelRequest(method string, endpoint string) (*http.Request, *logger.Error) {
	pCfg, e := helpers.PluginCfg[PluginConfig]("com.utmstack")
	if e != nil {
		return nil, e
	}

	url := fmt.Sprintf(endpoint, pCfg.Backend)

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, helpers.Logger().ErrorF(err.Error())
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
		return nil, helpers.Logger().ErrorF(err.Error())
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			helpers.Logger().ErrorF(err.Error())
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, helpers.Logger().ErrorF("failed to get connection key, received status code: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, helpers.Logger().ErrorF(err.Error())
	}

	return body, nil
}
