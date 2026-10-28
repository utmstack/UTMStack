package main

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/threatwinds/go-sdk/plugins"
)

func handleO365Graph(event *plugins.Event) error {
	var rec map[string]any
	if err := json.Unmarshal([]byte(event.GetRaw()), &rec); err != nil {
		return io.EOF
	}
	tenantID := event.GetTenantId()
	if activity, _ := rec["activityDisplayName"].(string); activity != "" {
		return o365GraphDirectoryAudit(tenantID, rec, activity)
	}
	upn := o365GraphUPN(rec)
	if upn == "" {
		return io.EOF
	}
	if status, ok := rec["status"].(map[string]any); ok {
		if code, _ := status["errorCode"].(float64); code == 0 {
			applyCloudEvent(tenantID, "o365", upn, upn, "", "login", graphTime(rec["createdDateTime"]))
		}
		return io.EOF
	}
	return io.EOF
}

func o365GraphUPN(rec map[string]any) string {
	upn, _ := rec["userPrincipalName"].(string)
	return strings.ToLower(strings.TrimSpace(upn))
}

func o365GraphDirectoryAudit(tenantID string, rec map[string]any, activity string) error {
	targets, ok := rec["targetResources"].([]any)
	if !ok || len(targets) == 0 {
		return io.EOF
	}
	first, ok := targets[0].(map[string]any)
	if !ok {
		return io.EOF
	}
	upn := strings.ToLower(strings.TrimSpace(strField(first, "userPrincipalName")))
	if upn == "" {
		return io.EOF
	}
	display := strField(first, "displayName")
	if display == "" {
		display = upn
	}
	et := graphTime(rec["activityDateTime"])
	switch {
	case userOpCreate[activity]:
		applyCloudEvent(tenantID, "o365", upn, display, "", "create", et)
	case userOpDelete[activity]:
		applyCloudEvent(tenantID, "o365", upn, display, "", "delete", et)
	case userOpDisabled[activity]:
		applyCloudEvent(tenantID, "o365", upn, display, "", "disabled", et)
	case userOpEnabled[activity]:
		applyCloudEvent(tenantID, "o365", upn, display, "", "enabled", et)
	}
	return io.EOF
}
