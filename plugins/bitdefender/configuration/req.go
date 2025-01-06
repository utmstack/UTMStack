package configuration

import (
	"bytes"
	"github.com/threatwinds/go-sdk/catcher"
	"net/http"

	"github.com/utmstack/UTMStack/plugins/bitdefender/utils"
	"github.com/utmstack/config-client-go/types"
)

func sendRequest(body []byte, config types.ModuleGroup) (*http.Response, error) {
	r, err := http.NewRequest("POST", config.Configurations[1].ConfValue+EndpointPush, bytes.NewBuffer(body))
	if err != nil {
		return nil, catcher.Error("cannot create request", err, map[string]any{})
	}

	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Authorization", utils.GenerateAuthCode(config.Configurations[0].ConfValue))

	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		return nil, catcher.Error("cannot send request", err, map[string]any{})
	}
	return resp, nil
}
