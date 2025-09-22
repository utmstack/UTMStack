package config

import (
	"bytes"
	"net/http"

	"github.com/threatwinds/go-sdk/catcher"

	"github.com/utmstack/UTMStack/plugins/bitdefender/utils"
)

func sendRequest(body []byte, config BDGZModuleConfig) (*http.Response, error) {
	r, err := http.NewRequest("POST", config.AccessUrl+EndpointPush, bytes.NewBuffer(body))
	if err != nil {
		return nil, catcher.Error("cannot create request", err, nil)
	}

	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Authorization", utils.GenerateAuthCode(config.ConnectionKey))

	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		return nil, catcher.Error("cannot send request", err, nil)
	}
	return resp, nil
}
