package main

// Fabricated AKS audit records run through the offline Azure parser model
// (azureParse) and the pinned SDK CEL. They pin the de-duplication keys that
// keep the admission webhook rule from alerting on every reconciliation patch
// the AKS control plane makes. The EventProcessor playground separately runs
// the real parser and alert plugins.
import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/tidwall/gjson"
)

func azureVolumeAudit(t *testing.T, user, verb, stage, resource, name string, code int) string {
	t.Helper()
	audit, err := json.Marshal(map[string]any{
		"kind": "Event", "apiVersion": "audit.k8s.io/v1", "stage": stage, "verb": verb,
		"user": map[string]any{"username": user}, "sourceIPs": []string{"198.51.100.4"},
		"responseStatus": map[string]any{"code": code},
		"objectRef":      map[string]any{"resource": resource, "name": name, "apiGroup": "admissionregistration.k8s.io"},
		"requestURI":     "/apis/admissionregistration.k8s.io/v1/" + resource + "/" + name + "?fieldManager=example",
	})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(map[string]any{
		"time": "2026-09-29T10:00:00Z", "tenantId": "directory-test", "category": "kube-audit-admin",
		"resourceId":    "/SUBSCRIPTIONS/00000000-0000-4000-8000-000000000000/RESOURCEGROUPS/RG-TEST/PROVIDERS/MICROSOFT.CONTAINERSERVICE/MANAGEDCLUSTERS/AKS-TEST",
		"operationName": "Microsoft.ContainerService/managedClusters/diagnosticLogs/Read",
		"properties":    map[string]any{"log": string(audit)},
	})
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func TestAzureKubernetesWebhookAlertVolume(t *testing.T) {
	cfg, cache := azureConfig(t), plugins.NewCELCache("azure-alert-volume")
	r := azureRules(t)["azure_kubernetes_admission_controller"]
	if r == nil {
		t.Fatal("missing azure_kubernetes_admission_controller")
	}
	want := []string{"dataSource", "lastEvent.event.azureScope", "adversary.user"}
	if len(r.GroupBy) != 0 || !reflect.DeepEqual(r.DeduplicateBy, want) {
		t.Fatalf("grouping got groupBy %v deduplicateBy %v, want deduplicateBy %v", r.GroupBy, r.DeduplicateBy, want)
	}
	for _, tc := range []struct {
		name string
		raw  string
		want bool
	}{
		{"control plane patch", azureVolumeAudit(t, "aksService", "patch", "ResponseComplete", "validatingwebhookconfigurations", "aks-node-validating-webhook", 200), true},
		{"administrator create", azureVolumeAudit(t, "admin@example.test", "create", "ResponseComplete", "mutatingwebhookconfigurations", "inject-sidecar", 201), true},
		{"read", azureVolumeAudit(t, "aksService", "get", "ResponseComplete", "validatingwebhookconfigurations", "aks-node-validating-webhook", 200), false},
		{"request received", azureVolumeAudit(t, "admin@example.test", "create", "RequestReceived", "mutatingwebhookconfigurations", "inject-sidecar", 200), false},
		{"denied", azureVolumeAudit(t, "admin@example.test", "patch", "ResponseComplete", "mutatingwebhookconfigurations", "inject-sidecar", 403), false},
		{"other resource", azureVolumeAudit(t, "admin@example.test", "patch", "ResponseComplete", "configmaps", "settings", 200), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := azureParse(t, cfg, tc.raw, "EventHub (test)", cache)
			got, err := cache.Eval(r.Where, out)
			if err != nil || got != tc.want {
				t.Fatalf("where got %v (%v), want %v for %s", got, err, tc.want, out)
			}
			if !tc.want {
				return
			}
			// adversary: origin, so adversary.user is the event's origin.user.
			for key, path := range map[string]string{"dataSource": "dataSource", "lastEvent.event.azureScope": "event.azureScope", "adversary.user": "origin.user"} {
				if v := gjson.Get(out, path); v.Type != gjson.String || v.String() == "" {
					t.Fatalf("de-duplication key %s (%s) does not resolve to text in %s", key, path, out)
				}
			}
		})
	}
}
