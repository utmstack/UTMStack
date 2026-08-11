package mcp

import (
	_ "embed"

	alertdto "github.com/utmstack/utmstack/backend/modules/alerts/dto"
)

// catalogJSON is the tool/resource catalog reference file shipped alongside
// the module. Exposed via mcp://utmstack/docs/catalog so AI clients can
// introspect the full surface from a single resource read.
//
//go:embed catalog.json
var catalogJSON string

// alertTagListAll returns a filter that pulls a reasonable batch of all alert
// tags for the catalog resource. Callers wanting fine-grained pagination use
// the alert_tags.list tool directly.
func alertTagListAll() alertdto.AlertTagFilters {
	return alertdto.AlertTagFilters{
		Page: 0,
		Size: 500,
	}
}

const filterOperatorsDoc = `# FilterType DSL — operator reference

Most search/filter inputs accept a ` + "`[]FilterType`" + ` array. Each entry has:

| field | description |
|---|---|
| ` + "`field`" + ` | document property (e.g. ` + "`source.user.name`" + `, ` + "`@timestamp`" + `) |
| ` + "`operator`" + ` | one of the operators below |
| ` + "`value`" + ` | operator-dependent value (string, number, ` + "`[from,to]`" + ` tuple, or string array) |

## Operators

### Equality
- ` + "`IS`" + ` — exact match (match_phrase)
- ` + "`IS_NOT`" + ` — inverse exact match

### Containment
- ` + "`CONTAIN`" + ` — substring (query_string ` + "`*v*`" + `)
- ` + "`DOES_NOT_CONTAIN`" + ` — inverse substring
- ` + "`NOT_CONTAINS`" + ` — legacy alias of DOES_NOT_CONTAIN
- ` + "`CONTAIN_ONE_OF`" + ` — substring matches any of value[]
- ` + "`DOES_NOT_CONTAIN_ONE_OF`" + ` — none of value[] are substrings

### Sets
- ` + "`IS_ONE_OF_TERMS`" + ` — value[] terms filter
- ` + "`IS_ONE_OF_TERMS_OR`" + ` — value[] terms (OR'd)
- ` + "`IS_ONE_OF`" + ` — value[] match_phrase (any)
- ` + "`IS_NOT_ONE_OF`" + ` — none of value[] match

### Ranges
- ` + "`IS_BETWEEN`" + ` — value=[from, to], inclusive
- ` + "`IS_NOT_BETWEEN`" + ` — outside the inclusive range
- ` + "`IS_GREATER_THAN`" + ` — gt (timestamps are RFC3339)
- ` + "`IS_LESS_THAN_OR_EQUALS`" + ` — lte

### Existence
- ` + "`EXIST`" + ` — field is present
- ` + "`DOES_NOT_EXIST`" + ` — field is missing

### String boundaries
- ` + "`START_WITH`" + ` / ` + "`NOT_START_WITH`" + ` — prefix
- ` + "`ENDS_WITH`" + ` / ` + "`NOT_ENDS_WITH`" + ` — suffix

### Cross-field
- ` + "`IS_IN_FIELDS`" + ` — value matches any of the listed fields (query_string)
- ` + "`IS_NOT_IN_FIELDS`" + ` — value matches none of the listed fields

## Example

` + "```json" + `
[
  {"field": "@timestamp", "operator": "IS_GREATER_THAN", "value": "2026-06-01T00:00:00Z"},
  {"field": "status", "operator": "IS", "value": 0},
  {"field": "tags", "operator": "CONTAIN", "value": "lateral"}
]
` + "```" + `

The list is ANDed: every filter has to hold.
`

const eventTypesDoc = `# Audit event-type enum (ApplicationEventType)

Every audit row written by the backend tags its EventType with one of these
string constants. The MCP layer writes ` + "`MCP_TOOL_CALL`" + ` for each tool invocation
(ResourceType=<tool name>, Metadata.transport="mcp").

The full enumeration lives in ` + "`modules/audit/domain/event_type.go`" + ` — query
` + "`audit.list`" + ` with ` + "`search_query=event_type.equals.<NAME>`" + ` to filter on any of:

- Auth: AUTH_ATTEMPT, AUTH_SUCCESS, AUTH_LOGOUT, TFA_*, RESET_USER_PASSWORD_*, PASSWORD_CHANGE_*
- User management: USER_CREATION_*, USER_UPDATE_*, USER_DELETE_*, USER_MANAGEMENT
- API keys: API_KEY_CREATE_*, API_KEY_DELETE_*
- Alerts: ALERT_UPDATE_*, ALERT_NOTE_UPDATE_*, ALERT_TAG_UPDATE_*, ALERT_CONVERT_TO_INCIDENT_*
- Incidents: INCIDENT_CREATION_*, INCIDENT_CREATED, INCIDENT_ALERT_ADD_*, INCIDENT_UPDATE_*
- SOAR: SOAR_RULE_CREATE_*, SOAR_RULE_UPDATE_*
- Correlation: CORRELATION_RULE_CREATE_*, CORRELATION_RULE_UPDATE_*, CORRELATION_RULE_DELETE_*
- Logstash filters: LOGSTASH_FILTER_*_ATTEMPT/SUCCESS
- Regex / Tenant: REGEX_PATTERN_*, TENANT_CONFIG_*, DATA_TYPE_*
- Dashboards: DASHBOARD_*, VISUALIZATION_*, DASHBOARD_LAYOUT_*
- Log analyzer: LOG_ANALYZER_QUERY_*
- Datasources: DATASOURCE_*, DATASOURCE_GROUP_*
- License: LICENSE_UPLOAD_*
- IdP: IDP_CONFIG_*
- Config: CONFIG_CHANGED
- Compliance: COMPLIANCE_SCHEDULE_*, COMPLIANCE_CONTROL_*, COMPLIANCE_FRAMEWORK_*
- MCP: MCP_TOOL_CALL
`
