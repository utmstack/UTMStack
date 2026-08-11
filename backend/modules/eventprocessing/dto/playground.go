package dto

import "encoding/json"

type TestPipelineRequest struct {
	Log      json.RawMessage `json:"log"      binding:"required"`
	Pipeline *CustomContent  `json:"pipeline"`
}

type TestRuleRequest struct {
	Log  json.RawMessage `json:"log"  binding:"required"`
	Rule *CustomContent  `json:"rule"`
}

type CustomContent struct {
	Content string `json:"content" binding:"required"`
}

type PlaygroundResponse struct {
	UUID       string            `json:"uuid"`
	Event      json.RawMessage   `json:"event,omitempty"`
	Alerts     []json.RawMessage `json:"alerts"`
	StopReason string            `json:"stopReason,omitempty"`
	TimedOut   bool              `json:"timedOut"`
}
