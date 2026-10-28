package main

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/threatwinds/go-sdk/plugins"
)

var (
	awsOpCreate   = map[string]bool{"CreateUser": true, "ProvisionAccount": true}
	awsOpDelete   = map[string]bool{"DeleteUser": true}
	awsOpDisabled = map[string]bool{"DeactivateUser": true}
	awsOpEnabled  = map[string]bool{"ActivateUser": true}
)

func handleAWS(event *plugins.Event) error {
	var rec map[string]any
	if err := json.Unmarshal([]byte(event.GetRaw()), &rec); err != nil {
		return io.EOF
	}
	op, _ := rec["eventName"].(string)
	et := graphTime(rec["eventTime"])
	id, _ := rec["userIdentity"].(map[string]any)

	if op == "ConsoleLogin" {
		upn := awsUPN(id)
		if upn == "" {
			return io.EOF
		}
		if msg, _ := rec["errorMessage"].(string); strings.TrimSpace(msg) != "" {
			return io.EOF
		}
		if resp, ok := rec["responseElements"].(map[string]any); ok {
			if v, _ := resp["ConsoleLogin"].(string); strings.EqualFold(v, "Failure") {
				return io.EOF
			}
		}
		applyCloudEvent(event.GetTenantId(), "aws", upn, upn, "", "login", et)
		return io.EOF
	}

	upn := awsLifecycleUPN(rec)
	if upn == "" {
		return io.EOF
	}
	switch {
	case awsOpCreate[op]:
		applyCloudEvent(event.GetTenantId(), "aws", upn, upn, "", "create", et)
	case awsOpDelete[op]:
		applyCloudEvent(event.GetTenantId(), "aws", upn, upn, "", "delete", et)
	case awsOpDisabled[op]:
		applyCloudEvent(event.GetTenantId(), "aws", upn, upn, "", "disabled", et)
	case awsOpEnabled[op]:
		applyCloudEvent(event.GetTenantId(), "aws", upn, upn, "", "enabled", et)
	}
	return io.EOF
}

func awsLifecycleUPN(rec map[string]any) string {
	params, ok := rec["requestParameters"].(map[string]any)
	if !ok {
		return ""
	}
	if name, _ := params["userName"].(string); name != "" {
		return strings.ToLower(strings.TrimSpace(name))
	}
	if name, _ := params["userLoginName"].(string); name != "" {
		return strings.ToLower(strings.TrimSpace(name))
	}
	return ""
}

func awsUPN(id map[string]any) string {
	if id == nil {
		return ""
	}
	if name, _ := id["userName"].(string); name != "" {
		return strings.ToLower(strings.TrimSpace(name))
	}
	return ""
}
