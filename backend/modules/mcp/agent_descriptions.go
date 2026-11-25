package mcp

import "github.com/modelcontextprotocol/go-sdk/mcp"

// A tool with no description leaves the model to guess what it is for and what its parameters mean. These
// are the ones the SOC agent sees that were registered without one; a description given where the tool is
// registered wins over what is here.
var agentDescriptions = map[string]string{
	"compliance.control.list": "List the controls of the compliance library (the framework-independent controls that frameworks refer to).",
	"compliance.framework.list": "List the compliance frameworks available. Each has a key; use it with compliance.framework.get " +
		"and compliance.report.evaluate.",
	"compliance.framework.get": "One compliance framework by key (from compliance.framework.list), with the controls it asks for.",
	"compliance.report.evaluate": "ACTION. Evaluate a framework against the current data and store the report. framework_key comes from " +
		"compliance.framework.list; window_days is how many days back to look. It writes a report and takes a while: only run it when asked.",

	"correlation_rule.list": "List correlation rules. Filters: rule_name, active (true/false), categories, adversaries, system_owner, " +
		"search. Paged: page starts at 0, size is the page size. Each rule has a rel_path, which the other correlation_rule tools take.",
	"correlation_rule.get": "One correlation rule by rel_path (from correlation_rule.list), with its full definition. Read an existing " +
		"rule with this before writing a new one: its definition shows the shape the rule language expects.",
	"correlation_rule.create": "ACTION. Create a correlation rule. confidentiality, integrity and availability are impact scores (0 to 3). " +
		"definition is the rule's logic in the platform's rule language: copy the shape of a rule from correlation_rule.get. " +
		"Create it with ruleActive=false; turning it on is a separate step the user must ask for.",
	"correlation_rule.update": "ACTION. Change a correlation rule. relPath says which one (from correlation_rule.list); the other fields " +
		"are the whole rule, as in correlation_rule.create.",
	"correlation_rule.set_active": "ACTION. Turn a correlation rule on or off: rel_path (from correlation_rule.list) and active. " +
		"A rule that is on raises alerts for real, so only do it when the user asks.",

	"dashboards.list": "List dashboards; name filters by name. Paged: page starts at 0, size is the page size. Each has an id (uuid).",
	"dashboards.get":  "One dashboard by id (the uuid from dashboards.list). Its widgets are listed with visualizations.list and its dashboard_id.",
	"dashboards.update": "ACTION. Change a dashboard: id (uuid) and name (required: the new name, or the same one), with description or " +
		"config if they change. It does not touch the widgets.",
	"visualizations.list": "The widgets of a dashboard: pass dashboard_id (uuid). Without it, every widget. Check it before creating a widget " +
		"again after a call timed out: the first one may have been saved.",

	"datasources.list": "List the connected data sources with their state and last activity. search_query filters by name; page_number " +
		"and page_size page the list; sort_by orders it. Each has an id (uuid).",
	"datasources.get": "One data source by id (the uuid from datasources.list).",
	"ingestion_stats.totals": "How much was ingested in a window, broken down by data source or data type, in events and bytes. " +
		"group_by is REQUIRED: dataSource or dataType. status is received (default), parsing_dropped, analysis_dropped, " +
		"correlation_dropped or all. from and to are RFC3339 or relative (now, now-24h, now-7d); the default window is the last 24 hours. " +
		"top is how many rows to return.",

	"incidents.list": "List incidents. Filters: name (contains), status, assigned_to, from and to (the creation time). Paged: page starts " +
		"at 0, size is the page size. Each incident has an id that is a UUID, not a number.",
	"incidents.get": "One incident by id. The id is the UUID from incidents.list, never a number or a name.",
	"incidents.change_status": "ACTION. Change an incident's status: incident_id (uuid), name (its current name, required), status, and " +
		"optionally solution or description. Only when the user asks.",
	"incident_alerts.list":  "The alerts attached to an incident: incident_id is its UUID. Optional alert_id, alert_status, page, size, sort.",
	"incident_notes.list":   "The notes of an incident: incident_id is its UUID. Paged with page and size.",
	"incident_notes.create": "ACTION. Add a note to an incident: incident_id (its UUID) and note_text.",

	"soar.node_types": "The kinds of node a SOAR flow can be built from (shell, http, conditional, notify, incident, mail, ...) and what each " +
		"needs. Call it before building a flow, and use only the executors it lists.",
	"soar.rule.list": "List SOAR rules. Filters: name, active (true/false), created_by, system_owner. Paged with page and size. " +
		"Each has a rel_path, which the other soar.rule tools take.",
	"soar.rule.get": "One SOAR rule by rel_path (from soar.rule.list), with its trigger conditions and nodes.",
	"soar.rule.set_enabled": "ACTION. Enable or disable a SOAR rule: rel_path and enabled. An enabled rule runs on real alerts, so only " +
		"when the user explicitly asks for it to be switched on.",
	"soar.execution.list": "Runs of SOAR rules, newest first. Filters: rule_path, alert_id, agent, execution_status, date_gte and date_lte " +
		"(RFC3339). Paged with page and size.",
}

// describe fills a tool's description from the table above when it was registered without one.
func describe(t *mcp.Tool) {
	if t.Description != "" {
		return
	}

	if d, ok := agentDescriptions[t.Name]; ok {
		t.Description = d
	}
}
