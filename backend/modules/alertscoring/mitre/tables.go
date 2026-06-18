package mitre

import (
	"regexp"
	"strings"
)

var TacticScores = map[string]int{
	// Active exploitation / damage (Critical tier)
	"Impact":            30,
	"Exfiltration":      28,
	"Data Exfiltration": 28,
	"Lateral Movement":  27,
	"Execution":         26,
	// Establishing foothold (High tier)
	"Privilege Escalation": 23,
	"Persistence":          22,
	"Defense Evasion":      21,
	"Credential Access":    20,
	// Pre-exploitation (Medium tier)
	"Command and Control": 17,
	"Collection":          15,
	"Initial Access":      14,
	// Reconnaissance (Low tier)
	"Discovery":            10,
	"Reconnaissance":       8,
	"Resource Development": 6,
}

var HighRiskTechniques = map[string]int{
	// Ransomware / destruction indicators
	"T1486": 20, // Data Encrypted for Impact
	"T1485": 20, // Data Destruction
	"T1490": 18, // Inhibit System Recovery
	// Active exploitation
	"T1059": 16, // Command and Scripting Interpreter
	"T1055": 16, // Process Injection
	"T1003": 16, // OS Credential Dumping
	"T1190": 15, // Exploit Public-Facing Application
	// Lateral movement / remote access
	"T1021": 15, // Remote Services
	"T1210": 15, // Exploitation of Remote Services
	"T1572": 14, // Protocol Tunneling
	// Persistence / evasion
	"T1078": 13, // Valid Accounts
	"T1098": 13, // Account Manipulation
	"T1562": 13, // Impair Defenses
	"T1110": 12, // Brute Force
	"T1566": 12, // Phishing
	"T1505": 12, // Server Software Component
	"T1048": 12, // Exfiltration Over Alternative Protocol
	"T1071": 11, // Application Layer Protocol
	"T1556": 11, // Modify Authentication Process
	"T1537": 11, // Transfer Data to Cloud Account
	"T1068": 11, // Exploitation for Privilege Escalation
	// Network recon
	"T1016": 8, // System Network Configuration Discovery
	"T1046": 8, // Network Service Discovery
	"T1018": 8, // Remote System Discovery
}

const DefaultTechniqueScore = 10

var SeverityMultiplier = map[int]float64{
	1: 0.5,  // Low
	2: 0.75, // Medium
	3: 1.0,  // High
	4: 1.2,  // Critical
}

const DefaultSeverityMultiplier = 0.75

var KillChainOrder = map[string]int{
	"Reconnaissance":       1,
	"Resource Development": 1,
	"Initial Access":       2,
	"Execution":            3,
	"Persistence":          4,
	"Privilege Escalation": 5,
	"Defense Evasion":      6,
	"Credential Access":    7,
	"Discovery":            8,
	"Lateral Movement":     9,
	"Collection":           10,
	"Command and Control":  11,
	"Exfiltration":         12,
	"Data Exfiltration":    12,
	"Impact":               13,
}

var techniqueIDPattern = regexp.MustCompile(`T\d{4}(?:\.\d{3})?`)

func TacticScore(category string) (int, string) {
	if strings.TrimSpace(category) == "" {
		return 0, ""
	}
	best := 0
	bestName := ""
	for _, raw := range strings.Split(category, ",") {
		t := strings.TrimSpace(raw)
		if s, ok := TacticScores[t]; ok && s > best {
			best = s
			bestName = t
		}
	}
	if best > 30 {
		best = 30
	}
	return best, bestName
}

func TechniqueScore(techniqueID string) int {
	if techniqueID == "" {
		return 0
	}
	if s, ok := HighRiskTechniques[techniqueID]; ok {
		return min(s, 20)
	}
	base := techniqueID
	if i := strings.Index(techniqueID, "."); i >= 0 {
		base = techniqueID[:i]
	}
	if s, ok := HighRiskTechniques[base]; ok {
		return min(s, 20)
	}
	return DefaultTechniqueScore
}

func ExtractTechniqueID(technique string, fallbacks ...string) string {
	if m := techniqueIDPattern.FindString(technique); m != "" {
		return m
	}
	for _, f := range fallbacks {
		if m := techniqueIDPattern.FindString(f); m != "" {
			return m
		}
	}
	return ""
}

func Stages(category string) []int {
	seen := map[int]bool{}
	for _, raw := range strings.Split(category, ",") {
		t := strings.TrimSpace(raw)
		if n, ok := KillChainOrder[t]; ok {
			seen[n] = true
		}
	}
	out := make([]int, 0, len(seen))
	for n := range seen {
		out = append(out, n)
	}
	return out
}
