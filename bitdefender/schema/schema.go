package schema

import "encoding/json"

type BodyEvents struct {
	CEF    string   `json:"cef"`
	Events []string `json:"events"`
}

type TemplateConfigSetPush struct {
	PARAMS  Params `json:"params"`
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	ID      string `json:"id"`
}

type Params struct {
	Status                int             `json:"status"`
	ServiceType           string          `json:"serviceType"`
	Servicesettings       ServiceSettings `json:"serviceSettings"`
	SubscribeToEventTypes json.RawMessage `json:"subscribeToEventTypes"`
}

type ServiceSettings struct {
	Url                        string `json:"url"`
	Authorization              string `json:"authorization"`
	RequireValidSslCertificate bool   `json:"requireValidSslCertificate"`
}

type ParamsTest struct {
	EventType string `json:"eventType"`
}

type TemplateConfigGetPush struct {
	PARAMS  json.RawMessage `json:"params"`
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	ID      string          `json:"id"`
}

type TemplateTestPush struct {
	PARAMS  ParamsTest `json:"params"`
	JSONRPC string     `json:"jsonrpc"`
	Method  string     `json:"method"`
	ID      string     `json:"id"`
}
