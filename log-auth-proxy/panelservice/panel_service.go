package panelservice

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/utmstack/UTMStack/log-auth-proxy/config"
	"github.com/utmstack/UTMStack/log-auth-proxy/model"
)

var PanelUrl = os.Getenv(config.UTMHostEnv)
var InternalKey = os.Getenv(config.UTMSharedKeyEnv)

func createPanelRequest(method string, endpoint string) (*http.Request, error) {
	url := fmt.Sprintf(endpoint, PanelUrl)
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add(config.PanelAPIKeyHeader, InternalKey)
	return req, nil
}

func GetConnectionKey() ([]byte, error) {
	client := &http.Client{}
	req, err := createPanelRequest("GET", config.PanelConnectionKeyEndpoint)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		return body, err
	}
	return nil, err
}

func GetPipelines() ([]model.PipelinePortConfiguration, error) {
	var pipelines []model.PipelinePortConfiguration
	client := &http.Client{}
	req, err := createPanelRequest("GET", config.PanelPipelinesEndpoint)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Add("page", "0")
	q.Add("size", "1000")
	q.Add("isInternal", "false")
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &pipelines)
	if err != nil {
		return nil, err
	}

	return pipelines, err
}
