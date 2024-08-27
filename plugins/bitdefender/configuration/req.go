package configuration

import (
	"bytes"
	"net/http"

	"github.com/utmstack/UTMStack/plugins/bitdefender/utils"
	"github.com/utmstack/config-client-go/types"
)

func sendRequest(body []byte, config types.ModuleGroup) (*http.Response, error) {
	r, err := http.NewRequest("POST", config.Configurations[1].ConfValue+EndpointPush, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Authorization", utils.GenerateAuthCode(config.Configurations[0].ConfValue))

	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
