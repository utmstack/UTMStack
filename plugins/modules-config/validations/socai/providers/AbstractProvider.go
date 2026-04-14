package providers

import (
	"encoding/json"
	"fmt"
	"maps"

	"github.com/threatwinds/go-sdk/utils"
)

type AbstractProvider struct {
	URL           string
	Model         string
	AuthType      string
	CustomHeaders map[string]string
}

func NewAbstractProvider(
	URL string,
	Model string,
	AuthType string,
	CustomHeaders map[string]string,
) *AbstractProvider {
	return &AbstractProvider{
		URL:           URL,
		Model:         Model,
		AuthType:      AuthType,
		CustomHeaders: CustomHeaders,
	}
}

func (p *AbstractProvider) Validate() error {
	if p.AuthType == "" {
		p.AuthType = "none"
	}
	if p.AuthType != "custom-headers" && p.AuthType != "none" {
		return fmt.Errorf("Invalid authentication type '%s'. Must be 'custom-headers' or 'none'.", p.AuthType)
	}

	return nil
}

func (p *AbstractProvider) PerformTestRequest() (int, error) {
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	if p.AuthType == "custom-headers" {
		maps.Copy(headers, p.CustomHeaders)
	}

	body,_ := json.Marshal(map[string]any{
		"model": p.Model,
        "messages": []map[string]string{
			 {
				"role": "user",
				"content": "only say ok",
			  },
		},
	})

	_, status, err := utils.DoReq[map[string]any](p.URL,body , "POST", headers, false)
	return status, err
}
