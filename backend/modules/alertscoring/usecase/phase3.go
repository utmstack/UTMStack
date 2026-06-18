package usecase

import (
	"strings"

	"github.com/utmstack/utmstack/backend/modules/alertscoring/connectors"
	"github.com/utmstack/utmstack/backend/modules/alertscoring/domain"
)

var dataTypeCriticality = map[string]int{
	"firewall-fortigate-traffic": 40,
	"firewall-cisco-asa":         40,
	"firewall-cisco-firepower":   40,
	"firewall-paloalto":          40,
	"firewall-sophos-xg":         40,
	"firewall-sonicwall":         40,
	"firewall-pfsense":           40,
	"firewall-mikrotik":          38,
	"firewall-fortiweb":          38,
	"firewall":                   38,
	"suricata":                   38,
	"nids":                       38,
	"vmware-esxi":                36,
	"aws":                        34,
	"azure":                      34,
	"google":                     34,
	"ibm-as400":                  32,
	"o365":                       30,
	"postgresql":                 30,
	"mysql":                      30,
	"mongodb":                    30,
	"linux":                      28,
	"apache":                     28,
	"nginx":                      28,
	"iis":                        28,
	"wineventlog":                25,
	"antivirus-sentinel-one":     22,
	"antivirus-kaspersky":        22,
	"antivirus-esmc-eset":        22,
	"antivirus-bitdefender-gz":   22,
	"sophos-central":             22,
	"macos":                      20,
	"syslog":                     18,
	"generic":                    15,
	"netflow":                    15,
	"json-input":                 12,
}

const defaultDataTypeScore = 20

func phase3Asset(dataType string, asset connectors.AssetInfo) domain.PhaseResult {
	dt := strings.ToLower(strings.TrimSpace(dataType))
	dtScore := dataTypeScore(dt)
	platformScore, platformLabel := platformScore(dt, asset)
	netScore := networkPositionScore(asset.IPs)
	agentScore, agentLabel := agentStatusScore(asset.Status)

	final := clampScore(float64(dtScore + platformScore + netScore + agentScore))

	return domain.PhaseResult{
		Score: final,
		Breakdown: map[string]any{
			"dataType":       dt,
			"dataTypePoints": dtScore,
			"hostname":       asset.Hostname,
			"platformLabel":  platformLabel,
			"platformPoints": platformScore,
			"assetCIA":       ciaTriplet(asset),
			"ipAddresses":    asset.IPs,
			"networkPoints":  netScore,
			"agentStatus":    asset.Status,
			"agentPoints":    agentScore,
			"agentLabel":     agentLabel,
		},
	}
}

func dataTypeScore(dt string) int {
	if dt == "" {
		return defaultDataTypeScore
	}
	if v, ok := dataTypeCriticality[dt]; ok {
		return v
	}
	for key, v := range dataTypeCriticality {
		if strings.Contains(dt, key) || strings.Contains(key, dt) {
			return v
		}
	}
	return defaultDataTypeScore
}

func platformScore(dt string, asset connectors.AssetInfo) (int, string) {
	if asset.HasSensitivity {
		maxCIA := asset.Confidentiality
		if asset.Integrity > maxCIA {
			maxCIA = asset.Integrity
		}
		if asset.Availability > maxCIA {
			maxCIA = asset.Availability
		}
		score := []int{6, 14, 22, 30}[clampInt(maxCIA, 0, 3)]
		return score, "asset CIA " + ciaTriplet(asset)
	}

	hostname := strings.ToLower(asset.Hostname)
	osStr := strings.ToLower(asset.OS + " " + asset.OSVersion)
	isWindows := dt == "wineventlog" || strings.Contains(osStr, "windows")

	if isWindows {
		return classifyWindows(hostname, osStr), "Windows"
	}
	return classifyNonWindows(hostname, dt), nonWindowsLabel(dt)
}

func classifyWindows(hostname, osStr string) int {
	serverKW := []string{"srv", "server", "dc", "ad", "dns", "exchange", "sql", "db", "web", "app", "mail", "ca", "pki", "hyperv", "vcenter", "esxi", "fw", "gw", "proxy"}
	isServer := strings.Contains(osStr, "server") || containsAny(hostname, serverKW)
	if isServer {
		if strings.Contains(osStr, "domain controller") || strings.Contains(hostname, "dc") || strings.Contains(hostname, "ad") {
			return 30
		}
		if containsAny(hostname, []string{"sql", "db", "exchange", "mail"}) {
			return 28
		}
		return 24
	}
	if containsAny(hostname, []string{"laptop", "desktop", "ws", "pc", "win10", "win11"}) {
		return 12
	}
	return 18
}

func classifyNonWindows(hostname, dt string) int {
	if strings.Contains(hostname, "utm") || dt == "utmstack" {
		return 30
	}
	if containsAny(hostname, []string{"fw", "firewall", "gateway", "proxy", "router", "switch", "vpn"}) {
		return 30
	}
	if containsAny(hostname, []string{"srv", "server", "db", "sql", "web", "app", "api", "mail", "dns", "ldap", "nfs", "san"}) {
		return 26
	}
	if containsAny(hostname, []string{"dev", "staging", "test", "sandbox"}) {
		return 14
	}
	return 20
}

func nonWindowsLabel(dt string) string {
	if dt == "" {
		return "Non-Windows"
	}
	return dt
}

// networkPositionScore (0–15): public/external IP scores highest.
func networkPositionScore(ips []string) int {
	if len(ips) == 0 {
		return 8
	}
	for _, raw := range ips {
		ip := strings.TrimSpace(raw)
		switch {
		case strings.HasPrefix(ip, "10.") || strings.HasPrefix(ip, "172.") || strings.HasPrefix(ip, "192.168."):
			return 10
		case ip != "" && !strings.HasPrefix(ip, "127.") && !strings.HasPrefix(ip, "::"):
			return 15
		}
	}
	return 8
}

// agentStatusScore (0–15): offline agents score higher — they can't be actively
// investigated and may be compromised or unreachable.
func agentStatusScore(status string) (int, string) {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "ONLINE":
		return 8, "Online"
	case "OFFLINE":
		return 15, "Offline"
	case "":
		return 10, "Unknown"
	default:
		return 10, "Unknown"
	}
}

func ciaTriplet(a connectors.AssetInfo) string {
	if !a.HasSensitivity {
		return "n/a"
	}
	return itoa(a.Confidentiality) + "/" + itoa(a.Integrity) + "/" + itoa(a.Availability)
}

func containsAny(s string, kws []string) bool {
	for _, kw := range kws {
		if strings.Contains(s, kw) {
			return true
		}
	}
	return false
}
