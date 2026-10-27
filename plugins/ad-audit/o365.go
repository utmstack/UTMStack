package main

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/threatwinds/go-sdk/plugins"
)

var (
	o365AadWorkload         = "AzureActiveDirectory"
	o365AadRecordTypeLegacy = "Audit.AzureActiveDirectory"

	o365MgmtOpCreate = map[string]bool{"Add user.": true, "Add user": true}
	o365MgmtOpDelete = map[string]bool{"Delete user.": true, "Delete user": true}
	o365MgmtOpLogin  = map[string]bool{"UserLoggedIn": true}
)

func handleO365API(event *plugins.Event) error {
	var rec map[string]any
	if err := json.Unmarshal([]byte(event.GetRaw()), &rec); err != nil {
		return io.EOF
	}
	if !isO365AADRecord(rec) {
		return io.EOF
	}
	op, _ := rec["Operation"].(string)
	et := graphTime(rec["CreationTime"])
	tenant := event.GetTenantId()

	switch {
	case o365MgmtOpCreate[op]:
		upn := o365MgmtTargetUPN(rec)
		if upn == "" {
			return io.EOF
		}
		applyCloudEvent(tenant, "o365", upn, upn, "", "create", et)
	case o365MgmtOpDelete[op]:
		upn := o365MgmtTargetUPN(rec)
		if upn == "" {
			return io.EOF
		}
		applyCloudEvent(tenant, "o365", upn, upn, "", "delete", et)
	case o365MgmtOpLogin[op]:
		upn := strings.ToLower(strings.TrimSpace(strField(rec, "UserId")))
		if upn == "" || !strings.Contains(upn, "@") {
			return io.EOF
		}
		applyCloudEvent(tenant, "o365", upn, upn, "", "login", et)
	}
	return io.EOF
}

func isO365AADRecord(rec map[string]any) bool {
	if wl, _ := rec["Workload"].(string); wl == o365AadWorkload {
		return true
	}
	if rt, _ := rec["RecordType"].(string); rt == o365AadRecordTypeLegacy {
		return true
	}
	return false
}

func o365MgmtTargetUPN(rec map[string]any) string {
	target, ok := rec["Target"].([]any)
	if !ok {
		return ""
	}
	var objectIDs []string
	for _, item := range target {
		t, ok := item.(map[string]any)
		if !ok {
			continue
		}
		typ, _ := t["Type"].(float64)
		id, _ := t["ID"].(string)
		if typ == 2 {
			guid := strings.TrimPrefix(id, "User_")
			if isGUID(guid) {
				objectIDs = append(objectIDs, strings.ToLower(strings.ReplaceAll(guid, "-", "")))
			}
		} else if typ == 5 {
			upn := strings.ToLower(strings.TrimSpace(id))
			if upn == "" || !strings.Contains(upn, "@") {
				continue
			}
			for _, oid := range objectIDs {
				if strings.HasPrefix(upn, oid) {
					return upn[len(oid):]
				}
			}
			return upn
		}
	}
	return ""
}

func isGUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	parts := strings.Split(s, "-")
	if len(parts) != 5 {
		return false
	}
	for i, want := range []int{8, 4, 4, 4, 12} {
		if len(parts[i]) != want || !isAllHex(parts[i]) {
			return false
		}
	}
	return true
}

func isAllHex(s string) bool {
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
			return false
		}
	}
	return len(s) > 0
}
