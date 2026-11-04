package main

import (
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/plugins"
)

var (
	azureCategorySignIn    = "SignInLogs"
	azureCategoryDirectory = "DirectoryAuditLogs"
)

func handleAzure(event *plugins.Event) error {
	var rec map[string]any
	if err := json.Unmarshal([]byte(event.GetRaw()), &rec); err != nil {
		return io.EOF
	}
	category, _ := rec["category"].(string)
	switch category {
	case azureCategorySignIn:
		upn := azureUPN(rec)
		if upn != "" {
			applyCloudEvent(event.GetTenantId(), "azure", upn, upn, "", "login", azureTime(rec))
		}
		return io.EOF
	case azureCategoryDirectory:
		operation := azureOperation(rec)
		upn := azureUPN(rec)
		if upn == "" {
			if target, ok := rec["targetObjectDisplayName"].(string); ok {
				upn = strings.ToLower(strings.TrimSpace(target))
			}
		}
		if upn == "" {
			return io.EOF
		}
		et := azureTime(rec)
		switch {
		case userOpCreate[operation]:
			applyCloudEvent(event.GetTenantId(), "azure", upn, upn, "", "create", et)
		case userOpDelete[operation]:
			applyCloudEvent(event.GetTenantId(), "azure", upn, upn, "", "delete", et)
		case userOpDisabled[operation]:
			applyCloudEvent(event.GetTenantId(), "azure", upn, upn, "", "disabled", et)
		case userOpEnabled[operation]:
			applyCloudEvent(event.GetTenantId(), "azure", upn, upn, "", "enabled", et)
		}
		return io.EOF
	}
	return io.EOF
}

func azureUPN(rec map[string]any) string {
	if props, ok := rec["properties"].(map[string]any); ok {
		if upn, _ := props["userPrincipalName"].(string); upn != "" {
			return strings.ToLower(strings.TrimSpace(upn))
		}
	}
	if upn, _ := rec["userPrincipalName"].(string); upn != "" {
		return strings.ToLower(strings.TrimSpace(upn))
	}
	if upn, _ := rec["UserPrincipalName"].(string); upn != "" {
		return strings.ToLower(strings.TrimSpace(upn))
	}
	if target, ok := rec["targetObjects"].([]any); ok && len(target) > 0 {
		if first, ok := target[0].(map[string]any); ok {
			if upn, _ := first["userPrincipalName"].(string); upn != "" {
				return strings.ToLower(strings.TrimSpace(upn))
			}
		}
	}
	return ""
}

func azureOperation(rec map[string]any) string {
	if props, ok := rec["properties"].(map[string]any); ok {
		if op, _ := props["operationName"].(string); op != "" {
			return op
		}
	}
	if op, _ := rec["operationName"].(string); op != "" {
		return op
	}
	op, _ := rec["OperationName"].(string)
	return op
}

func azureTime(rec map[string]any) *time.Time {
	if props, ok := rec["properties"].(map[string]any); ok {
		if et := graphTime(props["createdDateTime"]); et != nil {
			return et
		}
	}
	if et := graphTime(rec["createdDateTime"]); et != nil {
		return et
	}
	return graphTime(rec["time"])
}
