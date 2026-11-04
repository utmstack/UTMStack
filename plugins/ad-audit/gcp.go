package main

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"strings"

	"github.com/threatwinds/go-sdk/plugins"
)

var gcpOpCreate = map[string]bool{
	"cloudidentity#users.create": true,
}
var gcpOpDelete = map[string]bool{
	"cloudidentity#users.delete": true,
}
var gcpOpDisable = map[string]bool{
	"cloudidentity#users.deactivate": true,
}
var gcpOpEnable = map[string]bool{
	"cloudidentity#users.activate": true,
}

func handleGCP(event *plugins.Event) error {
	var envelope struct {
		Message struct {
			Data string `json:"data"`
		} `json:"message"`
	}
	if err := json.Unmarshal([]byte(event.GetRaw()), &envelope); err != nil || envelope.Message.Data == "" {
		envelope.Message.Data = event.GetRaw()
	}

	data := envelope.Message.Data
	if decoded, err := base64.StdEncoding.DecodeString(data); err == nil && json.Valid(decoded) {
		data = string(decoded)
	}

	var log map[string]any
	if err := json.Unmarshal([]byte(data), &log); err != nil {
		return io.EOF
	}

	principal := gcpprincipal(log)
	if principal == "" {
		return io.EOF
	}

	method := gcpMethod(log)
	et := graphTime(log["timestamp"])

	switch {
	case gcpOpCreate[method]:
		applyCloudEvent(event.GetTenantId(), "gcp", principal, principal, "", "create", et)
	case gcpOpDelete[method]:
		applyCloudEvent(event.GetTenantId(), "gcp", principal, principal, "", "delete", et)
	case gcpOpDisable[method]:
		applyCloudEvent(event.GetTenantId(), "gcp", principal, principal, "", "disabled", et)
	case gcpOpEnable[method]:
		applyCloudEvent(event.GetTenantId(), "gcp", principal, principal, "", "enabled", et)
	}
	return io.EOF
}

func gcpprincipal(log map[string]any) string {
	if proto, ok := log["protoPayload"].(map[string]any); ok {
		if auth, ok := proto["authenticationInfo"].(map[string]any); ok {
			if email, _ := auth["principalEmail"].(string); email != "" {
				return strings.ToLower(strings.TrimSpace(email))
			}
		}
	}
	if email, _ := log["principalEmail"].(string); email != "" {
		return strings.ToLower(strings.TrimSpace(email))
	}
	return ""
}

func gcpMethod(log map[string]any) string {
	if proto, ok := log["protoPayload"].(map[string]any); ok {
		if method, _ := proto["methodName"].(string); method != "" {
			return method
		}
	}
	if method, _ := log["methodName"].(string); method != "" {
		return method
	}
	return ""
}
