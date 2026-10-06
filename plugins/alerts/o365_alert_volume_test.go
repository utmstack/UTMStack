package main

// These fabricated raw records run through the checked-in O365 filter model and
// the pinned SDK CEL and history implementation. They pin the volume thresholds
// and de-duplication keys that keep these rules from flooding. The history
// transport is a loopback mock, not customer storage; the EventProcessor
// playground separately runs the real parser, history and alert plugins.
import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"

	sdkos "github.com/threatwinds/go-sdk/os"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/tidwall/gjson"
)

func o365VolumeRaw(t *testing.T, fields map[string]any) string {
	t.Helper()
	event := map[string]any{
		"CreationTime": "2026-09-29T10:00:00", "Id": "3c7c1f5e-9a55-4c1e-9f7b-000000000001",
		"OrganizationId": "11111111-2222-4333-8444-555555555555", "ResultStatus": "Succeeded",
		"UserKey": "10030000A0000001", "UserType": 0, "Version": 1,
	}
	for k, v := range fields {
		event[k] = v
	}
	b, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func o365VolumeFile(user, operation, workload string) map[string]any {
	return map[string]any{
		"Operation": operation, "RecordType": 6, "Workload": workload, "ClientIP": "198.51.100.22",
		"UserId": user, "ItemType": "File", "EventSource": "SharePoint",
		"ObjectId":          "https://contoso-my.sharepoint.com/personal/reader_example_test/Documents/plan.docx",
		"SiteUrl":           "https://contoso-my.sharepoint.com/personal/reader_example_test/",
		"SourceFileName":    "plan.docx",
		"SourceRelativeUrl": "Documents",
	}
}

func o365VolumeMailbox(user, operation string, recordType int) map[string]any {
	return map[string]any{
		"Operation": operation, "RecordType": recordType, "Workload": "Exchange", "ClientIP": "198.51.100.23",
		"ClientIPAddress": "198.51.100.23", "UserId": user, "MailboxOwnerUPN": user,
		"LogonType": 0, "OperationCount": 1,
		"OperationProperties": []map[string]any{{"Name": "MailAccessType", "Value": "Bind"}},
	}
}

func o365VolumeDirectory(operation string) map[string]any {
	return map[string]any{
		"Operation": operation, "RecordType": 8, "Workload": "AzureActiveDirectory", "ResultStatus": "Success",
		"UserId": "admin@example.test", "ObjectId": "target@example.test",
		"ModifiedProperties": []map[string]any{{"Name": "Role.DisplayName", "NewValue": "Global Administrator", "OldValue": ""}},
	}
}

func o365VolumeTeams() map[string]any {
	return map[string]any{
		"Operation": "MessagesListed", "RecordType": 25, "Workload": "MicrosoftTeams", "UserType": 5,
		"UserId":           "backup-app",
		"AppAccessContext": map[string]any{"ClientAppId": "0b3d7a52-8f4e-4d2b-9c61-000000000009"},
	}
}

func o365VolumeLogin() map[string]any {
	return map[string]any{
		"Operation": "UserLoginFailed", "RecordType": 15, "Workload": "AzureActiveDirectory",
		"ClientIP": "198.51.100.10", "UserId": "sprayed@example.test",
	}
}

func o365VolumeAudit(operation string) map[string]any {
	return map[string]any{
		"Operation": operation, "RecordType": 18, "Workload": "SecurityComplianceCenter", "UserType": 2,
		"UserId": "admin@example.test", "ClientIP": "198.51.100.77",
	}
}

// Each changed rule: the fields its alerts are de-duplicated by, the raw records
// its condition must accept, and the raw records it must now ignore.
var o365VolumeRules = []struct {
	file     string
	dedup    []string
	positive []map[string]any
	negative []map[string]any
}{
	{"o365_mailbox_mass_access", []string{"adversary.user"},
		[]map[string]any{o365VolumeMailbox("reader@example.test", "MailItemsAccessed", 50)},
		[]map[string]any{o365VolumeMailbox("reader@example.test", "MailboxLogin", 2)}},
	{"onedrive_mass_file_access", []string{"adversary.user"},
		[]map[string]any{o365VolumeFile("reader@example.test", "FileAccessed", "OneDrive"),
			o365VolumeFile("reader@example.test", "FileAccessedExtended", "OneDrive"),
			o365VolumeFile("reader@example.test", "FilePreviewed", "OneDrive")},
		[]map[string]any{o365VolumeFile("reader@example.test", "FileAccessed", "SharePoint"),
			o365VolumeFile("reader@example.test", "FileModified", "OneDrive")}},
	{"sharepoint_mass_downloads", []string{"adversary.user"},
		[]map[string]any{o365VolumeFile("reader@example.test", "FileDownloaded", "OneDrive"),
			o365VolumeFile("reader@example.test", "FileDownloaded", "SharePoint")},
		[]map[string]any{o365VolumeFile("reader@example.test", "FileAccessed", "SharePoint")}},
	{"teams_data_exfiltration", []string{"adversary.user", "lastEvent.log.appAccessContextClientAppId"},
		[]map[string]any{o365VolumeTeams()},
		[]map[string]any{{"Operation": "ChatCreated", "RecordType": 25, "Workload": "MicrosoftTeams", "UserId": "backup-app"}}},
	{"mass_email_deletion", []string{"adversary.user"},
		[]map[string]any{o365VolumeMailbox("cleaner@example.test", "SoftDelete", 3),
			o365VolumeMailbox("cleaner@example.test", "HardDelete", 3)},
		[]map[string]any{o365VolumeMailbox("cleaner@example.test", "MoveToDeletedItems", 3)}},
	{"o365-audit-log-purge", []string{"adversary.user", "lastEvent.action"},
		[]map[string]any{o365VolumeAudit("DSIPurgeStarted"), o365VolumeAudit("AuditSearchDeleted")},
		[]map[string]any{o365VolumeMailbox("cleaner@example.test", "HardDelete", 3),
			{"Operation": "Set-AdminAuditLogConfig", "RecordType": 1, "ResultStatus": "True", "Workload": "Exchange", "UserId": "admin@example.test"}}},
	{"credential_access_microsoft_365_potential_password_spraying_attack", []string{"adversary.ip"},
		[]map[string]any{o365VolumeLogin()},
		[]map[string]any{{"Operation": "UserLoggedIn", "RecordType": 15, "Workload": "AzureActiveDirectory", "ClientIP": "198.51.100.10", "UserId": "sprayed@example.test"}}},
	{"o365-admin-role-assignment", []string{"adversary.user", "lastEvent.log.ObjectId"},
		[]map[string]any{o365VolumeDirectory("Add member to role.")},
		[]map[string]any{o365VolumeDirectory("Add member to group."), o365VolumeDirectory("Add delegated permission grant."),
			o365VolumeDirectory("Update user.")}},
}

func TestO365AlertVolumePredicatesAndKeys(t *testing.T) {
	cache := plugins.NewCELCache("o365-alert-volume")
	for _, tc := range o365VolumeRules {
		t.Run(tc.file, func(t *testing.T) {
			r := o365OutcomeRule(t, tc.file)
			if len(r.GroupBy) != 0 || !reflect.DeepEqual(r.DeduplicateBy, tc.dedup) {
				t.Fatalf("grouping got groupBy %v deduplicateBy %v, want deduplicateBy %v", r.GroupBy, r.DeduplicateBy, tc.dedup)
			}
			for i, fields := range tc.positive {
				event := o365ActionResultNormalize(t, o365VolumeRaw(t, fields))
				if got, err := cache.Eval(r.Where, event); err != nil || !got {
					t.Fatalf("positive %d: got %v %v for %s", i, got, err, event)
				}
			}
			for i, fields := range tc.negative {
				event := o365ActionResultNormalize(t, o365VolumeRaw(t, fields))
				if got, err := cache.Eval(r.Where, event); err != nil || got {
					t.Fatalf("negative %d: got %v %v for %s", i, got, err, event)
				}
			}
		})
	}
	if _, err := os.Stat("../../rules/office365/o365-admin-role-granted.yml"); !os.IsNotExist(err) {
		t.Fatal("the Update user. proxy rule must stay removed; role assignments are O365 Admin Role Assignment")
	}
}

func TestO365AlertVolumeSDKHistory(t *testing.T) {
	// Isolate the SDK's process-global OpenSearch connection and field mapper.
	if os.Getenv("UTM_O365_VOLUME_HISTORY_CHILD") != "1" {
		c := exec.Command(os.Args[0], "-test.run=^TestO365AlertVolumeSDKHistory$")
		c.Env = append(os.Environ(), "UTM_O365_VOLUME_HISTORY_CHILD=1")
		if b, err := c.CombinedOutput(); err != nil {
			t.Fatalf("isolated history: %v\n%s", err, b)
		}
		return
	}
	var history []string
	var expectedTerms map[string]string
	var window time.Duration
	mapping := map[string]any{"properties": map[string]any{
		"@timestamp": map[string]any{"type": "date"}, "action": map[string]any{"type": "keyword"},
		"origin": map[string]any{"properties": map[string]any{"user": map[string]any{"type": "keyword"}, "ip": map[string]any{"type": "ip"}}},
		"log":    map[string]any{"properties": map[string]any{"Workload": map[string]any{"type": "keyword"}}},
	}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/_mapping") {
			_ = json.NewEncoder(w).Encode(map[string]any{"v11-log-o365-test": map[string]any{"mappings": mapping}})
			return
		}
		if r.URL.Path != "/v11-log-o365-*/_search" {
			t.Errorf("unexpected request %s", r.URL.Path)
			http.Error(w, "bad request", 400)
			return
		}
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			return
		}
		q := string(b)
		terms := map[string]string{}
		cutoff := time.Time{}
		if gjson.Get(q, "query.bool.must_not").Exists() {
			t.Error("unexpected negative history clauses")
		}
		for _, clause := range append(gjson.Get(q, "query.bool.filter").Array(), gjson.Get(q, "query.bool.must").Array()...) {
			if term := clause.Get("term"); term.Exists() {
				for field, value := range term.Map() {
					terms[strings.TrimSuffix(field, ".keyword")] = value.Get("value").String()
				}
			} else if span := clause.Get("range"); span.Exists() {
				if cutoff, err = time.Parse(time.RFC3339Nano, span.Get("@timestamp.gte").String()); err != nil {
					t.Error(err)
				}
			} else {
				t.Errorf("unsupported history clause %s", clause.Raw)
			}
		}
		if !reflect.DeepEqual(terms, expectedTerms) {
			t.Errorf("history scope got %v want %v", terms, expectedTerms)
		}
		if delta := time.Since(cutoff) - window; delta < -2*time.Second || delta > 2*time.Second {
			t.Errorf("wrong cutoff %v", delta)
		}
		hits := []map[string]any{}
		for _, doc := range history {
			match := true
			for field, value := range terms {
				if gjson.Get(doc, field).String() != value {
					match = false
				}
			}
			stamp, err := time.Parse(time.RFC3339Nano, gjson.Get(doc, "@timestamp").String())
			if err != nil || stamp.Before(cutoff) {
				match = false
			}
			if match {
				hits = append(hits, map[string]any{"_id": fmt.Sprint(len(hits)), "_index": "v11-log-o365-test", "_source": map[string]any{}})
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"took": 1, "hits": map[string]any{"total": map[string]any{"value": len(hits), "relation": "eq"}, "hits": hits}})
	}))
	defer server.Close()
	if err := sdkos.Connect([]string{server.URL}, "", ""); err != nil {
		t.Fatal(err)
	}
	// Every history search, parent first and then its "or" branches in order:
	// the raw record whose event is counted, the window, the threshold and the scope.
	type search struct {
		fields map[string]any
		within string
		count  uint64
		terms  map[string]string
	}
	user := "reader@example.test"
	for _, tc := range []struct {
		file     string
		searches []search
	}{
		{"o365_mailbox_mass_access", []search{
			{o365VolumeMailbox(user, "MailItemsAccessed", 50), "1h", 2000, map[string]string{"origin.user": user, "action": "MailItemsAccessed"}}}},
		{"onedrive_mass_file_access", []search{
			{o365VolumeFile(user, "FileAccessed", "OneDrive"), "1h", 1000, map[string]string{"origin.user": user, "action": "FileAccessed", "log.Workload": "OneDrive"}},
			{o365VolumeFile(user, "FileAccessedExtended", "OneDrive"), "1h", 1000, map[string]string{"origin.user": user, "action": "FileAccessedExtended", "log.Workload": "OneDrive"}},
			{o365VolumeFile(user, "FilePreviewed", "OneDrive"), "1h", 1000, map[string]string{"origin.user": user, "action": "FilePreviewed", "log.Workload": "OneDrive"}}}},
		{"sharepoint_mass_downloads", []search{
			{o365VolumeFile(user, "FileDownloaded", "SharePoint"), "1h", 500, map[string]string{"origin.user": user, "action": "FileDownloaded"}}}},
		{"teams_data_exfiltration", []search{
			{o365VolumeTeams(), "1h", 100, map[string]string{"origin.user": "backup-app", "log.Workload": "MicrosoftTeams"}}}},
		{"mass_email_deletion", []search{
			{o365VolumeMailbox("cleaner@example.test", "HardDelete", 3), "1h", 200, map[string]string{"origin.user": "cleaner@example.test", "action": "HardDelete"}},
			{o365VolumeMailbox("cleaner@example.test", "SoftDelete", 3), "1h", 5000, map[string]string{"origin.user": "cleaner@example.test", "action": "SoftDelete"}}}},
		{"credential_access_microsoft_365_potential_password_spraying_attack", []search{
			{o365VolumeLogin(), "10m", 50, map[string]string{"origin.ip": "198.51.100.10", "action": "UserLoginFailed"}}}},
	} {
		t.Run(tc.file, func(t *testing.T) {
			r := o365OutcomeRule(t, tc.file)
			if len(r.Correlation) != 1 || len(r.Correlation[0].Or) != len(tc.searches)-1 {
				t.Fatalf("unexpected history shape: %d searches", len(r.Correlation))
			}
			searches := append([]*plugins.SearchRequest{r.Correlation[0]}, r.Correlation[0].Or...)
			for i, want := range tc.searches {
				s := searches[i]
				if s.Count != want.count || s.Within != want.within || len(s.Or) != 0 && i > 0 {
					t.Fatalf("search %d: count %d within %s, want %d %s", i, s.Count, s.Within, want.count, want.within)
				}
				event := o365ActionResultNormalize(t, o365VolumeRaw(t, want.fields))
				expectedTerms = want.terms
				window, _ = time.ParseDuration(want.within)
				// The branch is executed alone; the parent's own "or" list is checked above.
				alone := &plugins.SearchRequest{IndexPattern: s.IndexPattern, With: s.With, Within: s.Within, Count: s.Count}
				prior := o365OutcomeSet(t, event, "@timestamp", time.Now().Add(-window/2).UTC().Format(time.RFC3339Nano))
				history = nil
				for n := uint64(0); n < want.count-1; n++ {
					history = append(history, prior)
				}
				if yes, _, err := alone.Execute(&event); err != nil || yes {
					t.Fatalf("search %d below threshold: %v %v", i, yes, err)
				}
				history = append(history, prior)
				if yes, _, err := alone.Execute(&event); err != nil || !yes {
					t.Fatalf("search %d at threshold: %v %v", i, yes, err)
				}
				history[len(history)-1] = o365OutcomeSet(t, prior, "@timestamp", time.Now().Add(-window-time.Minute).UTC().Format(time.RFC3339Nano))
				if yes, _, err := alone.Execute(&event); err != nil || yes {
					t.Fatalf("search %d counted expired history: %v %v", i, yes, err)
				}
				for field := range want.terms {
					history[len(history)-1] = o365OutcomeSet(t, prior, field, "different-value")
					if yes, _, err := alone.Execute(&event); err != nil || yes {
						t.Fatalf("search %d counted a different %s: %v %v", i, field, yes, err)
					}
				}
			}
		})
	}
}
