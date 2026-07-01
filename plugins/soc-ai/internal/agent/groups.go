package agent

import "strings"

type CapabilityGroup struct {
	ID       string
	Label    string // English description, used in the agent's permission prompt
	Prefixes []string
}

var capabilityGroups = []CapabilityGroup{
	{"alerts", "manage alerts (status, tags, notes, tag rules)", []string{"alerts", "alert_tags", "alert_tag_rules", "adversary"}},
	{"incidents", "manage incidents", []string{"incidents", "incident_notes", "incident_alerts", "incident_history", "incident_histories"}},
	{"dashboards", "create and manage dashboards", []string{"dashboards", "visualizations", "dashboard_layouts"}},
	{"compliance", "manage compliance frameworks and generate reports", []string{"compliance"}},
	{"correlation", "manage correlation and event-processing rules", []string{"correlation_rule", "regex_pattern", "tenant_config", "filter"}},
	{"datasources", "manage data sources", []string{"datasources", "datasource_groups"}},
}

var prefixToGroup = func() map[string]string {
	m := make(map[string]string)
	for _, g := range capabilityGroups {
		for _, p := range g.Prefixes {
			m[p] = g.ID
		}
	}
	return m
}()

func groupOf(toolName string) string {
	prefix := toolName
	if i := strings.IndexByte(toolName, '.'); i >= 0 {
		prefix = toolName[:i]
	}
	return prefixToGroup[strings.ToLower(prefix)]
}

func splitGroups(enabled []string) (allowed, locked []string) {
	set := make(map[string]bool, len(enabled))
	for _, g := range enabled {
		set[g] = true
	}
	for _, g := range capabilityGroups {
		if set[g.ID] {
			allowed = append(allowed, g.Label)
		} else {
			locked = append(locked, g.Label)
		}
	}
	return allowed, locked
}
