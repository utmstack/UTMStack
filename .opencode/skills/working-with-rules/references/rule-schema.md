# Rule schema reference

## Top-level fields
| field | required | notes |
|---|---|---|
| `dataTypes` | yes | list; must equal the filter's `dataTypes` (e.g. `o365`, `aws`, `windows`). Index pattern is `v11-log-<dtype>-*`. |
| `name` | yes | human-readable; this is what appears on the alert and is how you look up alerts in the index. |
| `impact` | yes | `{confidentiality,integrity,availability}` each 1–3; drives alert severity. |
| `category` | yes | MITRE tactic (e.g. `Credential Access`, `Collection`, `Defense Evasion`, `Impact`, `Initial Access`, `Persistence`). |
| `technique` | yes | MITRE ATT&CK id, e.g. `T1110.001`. |
| `adversary` | yes | `origin` or `target`. Controls which `Side` the alert attributes (user/ip/host) to. |
| `references` | no | list of URLs. |
| `description` | no | prose. |
| `where` | yes | CEL, evaluated on each matching event. |
| `afterEvents` | no | list of correlation windows (see below). |
| `groupBy` OR `deduplicateBy` | no | **only one**, list of field paths. |

## `where` CEL
Evaluated against the normalized event (post-filter). Available functions and their scalar-only type rules are in `working-with-filters` → `references/cel-semantics.md`. The field you reference must be **populated by the filter** — if it's not in the ingested event, the clause is false (or the whole rule no-ops), with no error.
- `oneOf("action", ["A","B"])`, `equals("action","A")`, `exists("origin.user")`, `contains("log.X","sub")` (string only), `inCIDR("origin.ip","10.0.0.0/8")`, `isWeekend("deviceTime")`.

## `afterEvents` (correlation / threshold)
Fires the rule when, in addition to the `where` trigger, **`count` more events** matching `with` occur from the same key within `within`.
```yaml
afterEvents:
  - indexPattern: v11-log-o365-*
    with:
      - field: origin.user          # filter on this field
        operator: filter_term        # filter_term | filter_match
        value: '{{.origin.user}}'    # placeholder -> value from the TRIGGER event
      - field: action
        operator: filter_term
        value: 'UserLoginFailed'
    within: 1h                       # duration: 30s, 5m, 24h ...
    count: 5                         # number of matching events required
```
- **`{{.field}}` placeholder**: resolves from the trigger event. It must resolve to a **non-nil** value, or go-sdk `plugins/rules.go` bails and the correlation silently never runs. Prefer a field guaranteed present (e.g. `origin.user`); avoid `origin.ip` if the op can be IP-less.
- `with` can be a flat list (all AND) or an `or:` list (any).
- Multiple `afterEvents` entries are ANDed.

## `groupBy` / `deduplicateBy`
- List of field paths, e.g. `[adversary.user, origin.ip]` or `[lastEvent.log.appAccessContextClientAppId]`.
- `lastEvent.*` reads the last event of the correlated set (the wire alert stores events as `events[]`; the alert plugin maps `lastEvent.<f>` → `events[<last>].<f>` — see known-mistakes #10, do NOT hand-edit rule files to use `events.`).
- Setting **both** `groupBy` and `deduplicateBy` is invalid (proto asserts).

## `adversary: origin` vs `target`
- `origin` → alert's `adversary.user/.ip/.host` come from the acting side (who did it).
- `target` → from the affected side (what was hit).
Mismatch = the alert attributes the wrong entity. O365 user-behavior rules → `origin`.

## Deploy DDL (Postgres)
Tables: `utm_correlation_rules` (the rule) + `utm_group_rules_data_type` (links rule_id → data_type_id).
- `data_type_id` for `o365` is **4** (re-query per tenant: `select id from utm_data_types where data_type='o365'`).
- Rule id is auto-assigned on insert (max existing id + 1). New O365 rules landed at 1462–1466 after the restart renumbering; the O365 filter id also shifted 1590 → 1591.
- Columns: `rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def (json), rule_definition_def (the where), rule_after_events_def (json), rule_group_by_def (json), rule_deduplicate_by_def (json), rule_active (bool), system_owner (bool), rule_adversary`.
- `rule_references_def` / `rule_after_events_def` / `rule_group_by_def` / `rule_deduplicate_by_def` are JSON **text** columns (store the YAML lists as JSON).

Full insert + the live-query/verify loop are in `references/deploy-and-verify.md`.
