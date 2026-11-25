package schema

import "encoding/json"

type BodyEvents struct {
	Cef    string   `json:"cef"`
	Events []string `json:"events"`
}

type TemplateConfigSetPush struct {
	Params  Params `json:"params"`
	JsonRpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Id      string `json:"id"`
}

type Params struct {
	Status                int             `json:"status"`
	ServiceType           string          `json:"serviceType"`
	ServiceSettings       ServiceSettings `json:"serviceSettings"`
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
	Params  json.RawMessage `json:"params"`
	JsonRpc string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Id      string          `json:"id"`
}

type TemplateTestPush struct {
	Params  ParamsTest `json:"params"`
	JsonRpc string     `json:"jsonrpc"`
	Method  string     `json:"method"`
	Id      string     `json:"id"`
}

type ModuleConfig struct {
	CompaniesIDs  string
	ConnectionKey string
	AccessUrl     string
	MasterIp      string
}
