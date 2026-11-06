package main

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"strings"

	"github.com/threatwinds/go-sdk/plugins"
)

var (
	gcpUserCreate   = map[string]bool{"google.admin.AdminService.createUser": true}
	gcpUserDelete   = map[string]bool{"google.admin.AdminService.deleteUser": true}
	gcpUserDisable  = map[string]bool{"google.admin.AdminService.suspendUser": true}
	gcpUserEnable   = map[string]bool{"google.admin.AdminService.unsuspendUser": true}
	gcpLoginSuccess = map[string]bool{"google.login.LoginService.loginSuccess": true}
)

var gcpSARequest = map[string]string{
	"google.iam.admin.v1.CreateServiceAccount":                        "create",
	"google.iam.admin.v1.DeleteServiceAccount":                        "delete",
	"google.iam.admin.v1.UndeleteServiceAccount":                      "enabled",
	"google.iam.admin.v1.DisableServiceAccount":                       "disabled",
	"google.iam.admin.v1.EnableServiceAccount":                        "enabled",
	"google.iam.admin.v1.WorkforcePools.CreateWorkforcePoolSubject":   "create",
	"google.iam.admin.v1.WorkforcePools.DeleteWorkforcePoolSubject":   "delete",
	"google.iam.admin.v1.WorkforcePools.UndeleteWorkforcePoolSubject": "enabled",
}

func handleGCP(event *plugins.Event) error {
	log, ok := gcpDecode(event)
	if !ok {
		return io.EOF
	}

	method := gcpMethod(log)
	et := graphTime(log["timestamp"])

	if gcpLoginSuccess[method] {
		upn := gcpprincipal(log)
		if upn == "" {
			return io.EOF
		}
		applyCloudEvent(event.GetTenantId(), "gcp", upn, upn, "", "login", et)
		return io.EOF
	}

	if eventType, isTarget := gcpSARequest[method]; isTarget {
		upn := gcpRequestTarget(log, method)
		if upn == "" {
			return io.EOF
		}
		applyCloudEvent(event.GetTenantId(), "gcp", upn, upn, "", eventType, et)
		return io.EOF
	}

	var eventType string
	switch {
	case gcpUserCreate[method]:
		eventType = "create"
	case gcpUserDelete[method]:
		eventType = "delete"
	case gcpUserDisable[method]:
		eventType = "disabled"
	case gcpUserEnable[method]:
		eventType = "enabled"
	default:
		return io.EOF
	}
	upn := gcpMetadataUserEmail(log)
	if upn == "" {
		return io.EOF
	}
	applyCloudEvent(event.GetTenantId(), "gcp", upn, upn, "", eventType, et)
	return io.EOF
}

func gcpDecode(event *plugins.Event) (map[string]any, bool) {
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
		return nil, false
	}
	return log, true
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

func gcpMetadataUserEmail(log map[string]any) string {
	proto, ok := log["protoPayload"].(map[string]any)
	if !ok {
		return ""
	}
	meta, ok := proto["metadata"].(map[string]any)
	if !ok {
		return ""
	}
	events, ok := meta["event"].([]any)
	if !ok {
		return ""
	}
	for _, e := range events {
		ev, ok := e.(map[string]any)
		if !ok {
			continue
		}
		params, ok := ev["parameter"].([]any)
		if !ok {
			continue
		}
		for _, p := range params {
			pm, ok := p.(map[string]any)
			if !ok {
				continue
			}
			if name, _ := pm["name"].(string); name == "USER_EMAIL" {
				if v, _ := pm["value"].(string); v != "" {
					return strings.ToLower(strings.TrimSpace(v))
				}
			}
		}
	}
	return ""
}

func gcpRequestTarget(log map[string]any, method string) string {
	proto, ok := log["protoPayload"].(map[string]any)
	if !ok {
		return ""
	}
	req := proto["request"]
	if s, ok := req.(string); ok {
		if sa, ok := gcpServiceAccountEmail(s); ok {
			return sa
		}
		if wp, ok := gcpWorkforceSubjectEmail(s); ok {
			return wp
		}
		return ""
	}
	rm, ok := req.(map[string]any)
	if !ok {
		return ""
	}
	if name, _ := rm["name"].(string); name != "" {
		if wp, ok := gcpWorkforceSubjectEmail(name); ok {
			return wp
		}
	}
	parent, _ := rm["parent"].(string)
	subj, _ := rm["workforcePoolSubjectId"].(string)
	if parent != "" && subj != "" {
		return gcpWorkforceSubjectFromParent(parent, subj)
	}
	return ""
}

func gcpServiceAccountEmail(s string) (string, bool) {
	idx := strings.LastIndex(s, "/serviceAccounts/")
	if idx < 0 {
		return "", false
	}
	projectIdx := strings.Index(s, "projects/")
	if projectIdx < 0 {
		return "", false
	}
	project := s[projectIdx+9 : idx]
	id := strings.TrimSpace(s[idx+len("/serviceAccounts/"):])
	if id == "" || strings.Contains(id, "/") {
		return "", false
	}
	return strings.ToLower(id + "@" + strings.ToLower(project) + ".iam.gserviceaccount.com"), true
}

func gcpWorkforceSubjectEmail(resource string) (string, bool) {
	wpIdx := strings.Index(resource, "/workforcePools/")
	wppIdx := strings.Index(resource, "/workforcePoolProviders/")
	subjIdx := strings.Index(resource, "/workforcePoolSubjects/")
	if wpIdx < 0 || wppIdx < 0 || subjIdx < 0 || wppIdx < wpIdx || subjIdx < wppIdx {
		return "", false
	}
	pIdx := strings.Index(resource, "projects/")
	if pIdx < 0 {
		return "", false
	}
	project := strings.TrimSpace(resource[pIdx+9 : wpIdx])
	pool := strings.TrimSpace(resource[wpIdx+len("/workforcePools/") : wppIdx])
	provider := strings.TrimSpace(resource[wppIdx+len("/workforcePoolProviders/") : subjIdx])
	subject := strings.TrimSpace(resource[subjIdx+len("/workforcePoolSubjects/"):])
	if project == "" || pool == "" || provider == "" || subject == "" {
		return "", false
	}
	return strings.ToLower(subject) + "@" + strings.ToLower(pool) + "." + strings.ToLower(provider) + "." + strings.ToLower(project) + ".identitytoolkit.com", true
}

func gcpWorkforceSubjectFromParent(parent, subject string) string {
	rest := strings.TrimPrefix(parent, "projects/")
	if rest == parent {
		return ""
	}
	project := rest
	if i := strings.Index(rest, "/"); i >= 0 {
		project = rest[:i]
	}
	wpIdx := strings.Index(parent, "/workforcePools/")
	wppIdx := strings.Index(parent, "/workforcePoolProviders/")
	if wpIdx < 0 || wppIdx < 0 {
		return ""
	}
	pool := strings.TrimSpace(parent[wpIdx+len("/workforcePools/") : wppIdx])
	provider := strings.TrimSpace(parent[wppIdx+len("/workforcePoolProviders/"):])
	if project == "" || pool == "" || provider == "" || subject == "" {
		return ""
	}
	return strings.ToLower(subject) + "@" + strings.ToLower(pool) + "." + strings.ToLower(provider) + "." + strings.ToLower(project) + ".identitytoolkit.com"
}
