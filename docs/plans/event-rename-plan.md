# Plan — Rename `log` → `event` (flattened) + add `controls` to the Event schema

## Summary

Replace the heterogeneous `log` bag (`map<string, structpb.Value>`) with a new `event` bag
(`flattened` in OpenSearch) to permanently eliminate dynamic-mapping conflicts. Add a new
`controls` field (`repeated string`) for future compliance-control tags. New documents in
**all** indices (existing and new) carry `event` + `controls`; old documents retain `log`
unchanged. The rename is a coordinated, breaking change across go-sdk, EventProcessor,
UTMStack (filters, rules, UI, backend, installer), and the opensearch-connector.

## Why this design

| Question | Answer |
|---|---|
| Why rename the proto field (not just the OS key)? | The proto field name IS the JSON key (protojson). The draft JSON, CEL evaluation, correlation queries, and OS document all derive from it. Renaming the proto field is the single source of truth; renaming only at the OS boundary creates a dual-namespace mismatch (CEL sees `log.*`, OS has `event.*`, correlation breaks). |
| Why `flattened`? | `flattened` never infers a type — it stores every sub-key as a keyword. No two documents can conflict. It's the purpose-built type for "arbitrary vendor fields I want to store and term-query without a schema." |
| Why keep `log` for old data? | Existing indices have `log` dynamically mapped. Adding `event: {type:flattened}` to those indices is a non-destructive mapping addition. Old docs keep `log`, new docs use `event`. ILM rotation eventually retires old data. No reindex needed. |
| What about `logx`? | Dead v10 legacy. Not live on v11. Removed from all references as part of this change. |
| What breaks on old data? | Correlation/afterEvents on `event.*` won't match old `log.*` docs. UI filters on `event.*` won't match old `log.*` docs. Accepted — the user confirmed new data uses `event`. |

## Data model (new Event proto)

```proto
message Event {
  string id = 1;
  string timestamp = 2 [json_name = "@timestamp"];
  string deviceTime = 3;
  string dataType = 4;
  string dataSource = 5;
  string tenantId = 6;
  string tenantName = 7;
  string raw = 8;
  // RENAMED from: map<string, google.protobuf.Value> log = 9;
  map<string, google.protobuf.Value> event = 9;
  Side target = 10;
  Side origin = 11;
  string protocol = 12;
  string connectionStatus = 13;
  uint32 statusCode = 14;
  string actionResult = 15;
  string action = 16;
  string severity = 17;
  repeated string errors = 18;
  map<string, ComplianceValues> compliance = 19;
  // NEW: compliance control tags
  repeated string controls = 20;
}
```

- Field 9 keeps its wire number (9) — only the name changes. Proto binary wire format is
  field-number-based, so gRPC messages in flight during a rolling upgrade remain compatible
  (old nodes send field 9, new nodes read field 9). The JSON name changes from `log` to `event`.
- Field 20 (`controls`) is new. `repeated string` → `["control-1", "control-2"]` in JSON.

### OpenSearch document shape (new)

```json
{
  "id": "...",
  "@timestamp": "...",
  "dataType": "firewall-cisco-asa",
  "raw": "...",
  "event": { "eventName": "SYSLOG", "localIp": "10.0.0.1", "severity": 5 },
  "origin": { "ip": "10.0.0.1", "user": "admin" },
  "target": { "ip": "10.0.0.2" },
  "controls": ["NIST-800-53-AC-3", "ISO27001-A.9.1.1"],
  "action": "deny",
  "actionResult": "blocked",
  "severity": "high"
}
```

### OpenSearch template (new)

```json
{
  "index_patterns": ["v11-log-*"],
  "template": {
    "settings": {
      "index.max_shards": 30000,
      "index.number_of_shards": 1,
      "index.number_of_replicas": 0,
      "index.mapping.total_fields.limit": 50000
    },
    "mappings": {
      "properties": {
        "event":    { "type": "flattened" },
        "controls": { "type": "keyword" }
      }
    }
  }
}
```

- `event` is `flattened`: every sub-key stored as a keyword, no type inference, no conflicts.
  Supports `term`, `terms`, `prefix`, `exists` queries. Does NOT support `match`,
  `match_phrase`, `range`, or aggregations.
- `controls` is `keyword`: term queries, terms queries, aggregations all work.
- `log` is NOT in the template (old field, already dynamically mapped in existing indices).
  New documents don't have a `log` key, so no new `log` mappings are created.

---

## Phases (ordered by dependency)

### Phase 1 — go-sdk: proto rename + regen + controls

**Repo:** `threatwinds/go-sdk`
**Files:**
- `plugins/plugins.proto`
- `plugins/plugins.pb.go` (regenerated)

**Changes:**
1. `plugins.proto:61`: `map<string, google.protobuf.Value> log = 9;` →
   `map<string, google.protobuf.Value> event = 9;`
2. `plugins.proto` Event message: add `repeated string controls = 20;`
3. Regenerate: `protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative plugins/plugins.proto`
   (or the repo's existing codegen script / `proto-command.txt` equivalent)
4. `plugins.pb.go`: Go field `Log` → `Event`, getter `GetLog()` → `GetEvent()`,
   add `Controls []string` + `GetControls()`.
5. **Audit all go-sdk references** to `Event.Log` / `GetLog()` / `.Log` on Event and update:
   - `plugins/cel.go` — CEL evaluation marshals the Event proto via protojson; the JSON key
     changes from `log` to `event` automatically. No code change needed, but verify.
   - `os/source.go` — `ParseSourceToProtoMessage` reads hit `_source` into Event; the
     `event` key in the JSON will map to the `event` proto field. Verify.
   - Any test fixtures / contract tests that construct `Event{Log: ...}` → `Event{Event: ...}`.
6. **Tag the go-sdk** (e.g. `v1.1.37`) and push.

**Verify:**
- `go build ./...` passes.
- Unit test: marshal an Event with `event: {"foo": "bar"}` → JSON has `"event":{"foo":"bar"}`.
- Unit test: unmarshal `{"event":{"foo":"bar"},"controls":["a","b"]}` → Event.Event + Event.Controls populated.
- Unit test: `CELCache.Eval` against `{"event":{"severity":5}}` with rule `greaterThan("event.severity", 3)` → true.

---

### Phase 2 — EventProcessor: parsing plugins + engine

**Repo:** `utmstack/EventProcessor`
**Bump go-sdk dependency** to the new tag (all 16 go.mod files: root + plugins).

**Changes:**

**Hardcoded `log.` prefix in parsing plugins** (3 plugins):
- `plugins/json/main.go:54`: `fmt.Sprintf("log.%s", k)` → `fmt.Sprintf("event.%s", k)`
- `plugins/xml/main.go:55`: `fmt.Sprintf("log.%s", k)` → `fmt.Sprintf("event.%s", k)`
- `plugins/kv/main.go:58`: `fmt.Sprintf("log.%s", k)` → `fmt.Sprintf("event.%s", k)`

**No code change needed** (field names come from filter YAML, which is updated in Phase 3):
- `plugins/grok/main.go` — uses `fieldName` from filter YAML directly
- `plugins/csv/main.go` — uses `header` from filter YAML
- `plugins/add/main.go` — uses `fieldName` from filter YAML
- `plugins/trim/main.go`, `plugins/cast/main.go`, `plugins/rename/main.go`,
  `plugins/delete/main.go`, `plugins/reformat/main.go` — all use filter YAML field names

**Engine plumbing** (verify, likely no change):
- `pkg/parsing/parsing.go:493`: `StringToProtoMessage(&draft.Log, event)` — the draft JSON
  will have `event:{...}` (set by parsing plugins in Phase 3 filters), which maps to the
  `event` proto field. Verify the unmarshal works.
- `pkg/parsing/parsing.go:87-88,320-321`: `gjson.Get(item.Log, "dataType")` — reads
  top-level `dataType` from the draft, not under `log`/`event`. No change.
- `plugins/cel/main.go:72`: `ProtoMessageToString(event)` — protojson serializes with
  `event` key. No change.
- `plugins/cel/main.go:208-222`: `parseHitsToEvents` → `ParseSourceToProtoMessage` —
  reads OS hit `_source` (which has `event` key) into Event proto. Verify.
- `plugins/feeds/main.go:337`: constructs `Alert{Events: ...}` — the Event proto now has
  `event` field. No change.

**Rebuild all plugin binaries** (`com.utmstack.*`, `cel`, `json`, `xml`, `kv`, etc.) and
update the base image tag (`TW_EVENT_PROCESSOR_VERSION_PROD`).

**Verify:**
- `go build ./...` passes for all 16 modules.
- End-to-end: pipe a raw ASA syslog through the engine → resulting OS doc has `event:{...}`,
  not `log:{...}`. CEL rule `equals("event.messageId", "106001")` fires.
- Correlation: rule with `afterEvents.field: "event.eventCode"` finds prior docs with `event.eventCode`.

---

### Phase 3 — UTMStack filters: mechanical rename `log.` → `event.`

**Repo:** `UTMStack`
**Scope:** ~19,755 `log.` references across all `filters/**/*.yml`

**Changes (mechanical, safe with word-boundary regex):**
- `filters/**/*.yml`: `log.` → `event.` in all field paths:
  - `fieldName: log.syslogPri` → `fieldName: event.syslogPri`
  - `source: log.localIp` → `source: event.localIp`
  - `to: log.localIpGeolocation` → `to: event.localIpGeolocation`
  - `fields: [log.messageId]` → `fields: [event.messageId]`
  - `where: exists("log.localIp")` → `where: exists("event.localIp")`
  - `where: equals("log.severity", "5")` → `where: equals("event.severity", "5")`
  - `destination: log.localIpGeolocation` → `destination: event.localIpGeolocation`
- **Do NOT rename** top-level fields that happen to contain "log" as a substring:
  `logx.*` (dead, remove separately), `winlogEventData*` (these are under `log.`, so
  `log.winlogEventDataTargetUserSid` → `event.winlogEventDataTargetUserSid` — correct).
- **Watch for:** `log.` appearing in comments, descriptions, or non-field contexts.
  Use a targeted regex: `(?<=^|\s|:|"|\[|\()log\.` → `event.` (field-path positions only).

**Also:**
- Remove `logx.*` references (dead v10 legacy):
  - `frontend/src/app/shared/constants/main-index-pattern.constant.ts` — `LOG_INDEX_WINLOGBEAT` if it references `logx`
  - `backend/src/main/java/com/park/utmstack/config/Constants.java:72-73` — `logxWineventlog*` constants (dead)
  - Any filter YAML with `logx.` (grep and remove)

**Verify:**
- `grep -r "log\." filters/ | grep -v "event\." | grep -v "\.log\." | wc -l` → 0
  (no remaining `log.` field references; `.log.` matches are file paths in comments)
- Spot-check 5 vendor filters (ASA, FortiGate, AWS, O365, Windows) for correct rename.
- Run filter contract tests: `cd plugins/alerts && go test -run TestFilterAndRuleContracts -v`

---

### Phase 4 — UTMStack rules: mechanical rename `log.` → `event.`

**Repo:** `UTMStack`
**Scope:** ~4,207 `log.` references across all `rules/**/*.yml`

**Changes (mechanical):**
- `rules/**/*.yml`: `log.` → `event.` in:
  - CEL `where` expressions: `equals("log.eventCode", "4624")` → `equals("event.eventCode", "4624")`
  - `afterEvents.field`: `field: log.eventCode` → `field: event.eventCode`
  - `correlation.field`: same
  - `groupBy`: `["log.user"]` → `["event.user"]`
  - `deduplicateBy`: same
- **Watch for:** `log.` in rule names, descriptions, MITRE references (don't rename those).
  Target only CEL string args and YAML field values.

**Verify:**
- `grep -r '"log\.' rules/ | wc -l` → 0
- `grep -r 'field: log\.' rules/ | wc -l` → 0
- Run rule contract tests: `cd plugins/alerts && go test -run TestFilterAndRuleContracts -v`
- Spot-check: netflow rules use `origin.bytesSent` (not `log.*`), Windows rules use `event.eventCode`.

---

### Phase 5 — OpenSearch template: add `event` + `controls`

**Repo:** `UTMStack`
**Files:**
- `installer/services/search.go`

**Changes:**

1. **`InitOpenSearch()`** (fresh install, line 61-66):
   - Update `templateData` (general `utmstack_indexes`) to add mappings:
     ```json
     "template": {
       "settings": { "index.number_of_shards":1, "index.number_of_replicas":0, "index.mapping.total_fields.limit":50000 },
       "mappings": { "properties": {
         "event":    { "type": "flattened" },
         "controls": { "type": "keyword" }
       } }
     }
     ```
   - Update `logTemplateData` (log-specific `utmstack_log_indexes`, line 66-67):
     Keep settings, add the same mappings (or rely on the general template — composable
     templates merge, so the general template's mappings apply to `v11-log-*` too).
     Cleanest: put `event` + `controls` in the general template (applies to all patterns),
     keep `logTemplateData` settings-only.

2. **`UpdateOpenSearch()`** (upgrade, line 81-97):
   - Add `PUT _index_template/utmstack_indexes` with the updated mappings (same as Init).
   - Add `PUT v11-log-*/_mapping` to add `event: {type:flattened}` + `controls: {type:keyword}`
     to **existing** indices (non-destructive mapping addition). Use
     `?allow_no_indices=true` to skip if no indices match.
   - Keep existing `max_shards` + `total_fields.limit` settings PUTs.

3. **New lock** in `installer/setup/apply.go`: add a lock (e.g. `20261201001`) for the
   template update, so it runs once per upgrade.

**Verify:**
- Fresh install: `GET v11-log-*/_mapping` shows `event: {type:flattened}` + `controls: {type:keyword}`.
- Upgrade: existing index gets `event` + `controls` mapping added; old `log` mapping unchanged.
- Index a doc with `event: {"severity": 5, "name": "test"}` → accepted.
- Query: `GET v11-log-*/_search` with `{"term": {"event.severity": "5"}}` → hits.
- Query: `GET v11-log-*/_search` with `{"terms": {"controls": ["NIST-800-53-AC-3"]}}` → hits.

---

### Phase 6 — Frontend: update field references

**Repo:** `UTMStack`
**Files:**

1. **`elastic-data-types.enum.ts`**: add `FLATTENED = 'flattened'`.

2. **`operator.service.ts`**: add a branch for `FLATTENED`:
   ```ts
   } else if (field.type === ElasticDataTypesEnum.FLATTENED) {
     // flattened supports: term (IS, IS_NOT, IS_ONE_OF, IS_NOT_ONE_OF),
   //   exists (EXIST, DOES_NOT_EXIST), prefix (START_WITH, NOT_START_WITH)
     //   Does NOT support: contain, does_not_contain, is_between, start_with (match-based)
     operators = FILTER_OPERATORS.filter(value =>
       value.operator === ElasticOperatorsEnum.IS ||
       value.operator === ElasticOperatorsEnum.IS_NOT ||
       value.operator === ElasticOperatorsEnum.IS_ONE_OF ||
       value.operator === ElasticOperatorsEnum.IS_NOT_ONE_OF ||
       value.operator === ElasticOperatorsEnum.EXIST ||
       value.operator === ElasticOperatorsEnum.DOES_NOT_EXIST ||
       value.operator === ElasticOperatorsEnum.IS_ONE_OF_TERMS ||
       value.operator === ElasticOperatorsEnum.PREFIX
     );
   }
   ```
   (Exact operator set to be finalized during implementation — the point is to restrict
   to flattened-compatible operators only.)

3. **`file-field.enum.ts`**: rename all `log.*` → `event.*`:
   ```ts
   FILE_ACCESS_LIST_FIELD = 'event.eventDataAccessList',
   FILE_EVENT_ID_FIELD = 'event.eventCode',
   FILE_EVENT_NAME_FIELD = 'event.eventName',
   // ... etc for all 29 FILE_* constants
   ```

4. **`alert-field.constant.ts`**: check for `events.log.*` → `events.event.*`
   (likely none — the alert field constants reference top-level Event fields like
   `events.dataType`, `events.origin.ip`, not the bag).

5. **`active-directory/**`**: `event-timeline.component.ts:143`:
   `item.log.eventCode` → `item.event.eventCode`
   `active-directory-event.component.ts:34`:
   `this.event.log.message` → `this.event.event.message`

6. **`dashboard-overview.component.ts`**: `logx.wineventlog.event_name.keyword` →
   `event.eventName.keyword` (flattened doesn't have `.keyword` subfield — use
   `event.eventName` directly with `term`). **This is a behavior change**: the old
   `logx.wineventlog.event_name.keyword` was a keyword subfield; the new `event.eventName`
   is a flattened path. The chart dimension query changes from `terms agg on
   logx.wineventlog.event_name.keyword` to... **flattened doesn't support aggregations**.
   This chart breaks. Mitigation: keep the chart on `dataType` or `action` (top-level
   typed fields) instead, or accept the loss.

7. **`local-field.service.ts`**: bump cache key `_fields` → `_fields_v3` (cache bust).

8. **Any other hardcoded `log.` / `logx.` references**: grep `frontend/src` for
   `log\.` and `logx\.` and update.

**Verify:**
- Log Analyzer: field picker shows `event.eventCode`, `event.eventName`, etc. (from mapping).
- File management: filters work on `event.*` fields.
- AD module: `item.event.eventCode` renders correctly.
- Dashboard: "Top events" chart still works (on a non-flattened field).
- Build: `NODE_OPTIONS=--max_old_space_size=8192 npm run build` passes.

---

### Phase 7 — Backend (Java) + user-auditor + compliance-orchestrator

**Repo:** `UTMStack`

1. **`Constants.java:72-73`**: remove `logxWineventlog*` constants (dead v10).

2. **`user-auditor/UserService.java:113-169`**:
   - `eventLog.getLog().get("eventCode")` → `eventLog.getEvent().get("eventCode")`
   - `eventLog.getLog().get("winlogEventDataTargetUserSid")` → `eventLog.getEvent().get("winlogEventDataTargetUserSid")`
   - **BUT**: user-auditor reads from OpenSearch (not the go-sdk proto). The Java model
     `WinlogbeatInfo` has a `Map<String, Object> log` field. If the OS doc now has `event`
     instead of `log`, the Java model needs a `Map<String, Object> event` field, and the
     service code updates accordingly.
   - **Migration concern**: user-auditor queries old indices where the key is still `log`.
     During the transition, it may need to check both `log` and `event`.
     Simplest: update to `event` only; old data is being rotated out.

3. **`user-auditor/Constants.java:14`**:
   `LOG_WINLOG_EVENT_DATA_TARGET_USER_SID_KEYWORD = "log.winlogEventDataTargetUserSid.keyword"` →
   `"event.winlogEventDataTargetUserSid"` (flattened, no `.keyword` suffix).

4. **`compliance-orchestrator/`**: grep for `log.` / `.getLog()` / `"log"` and update.
   The `controls` field is the future integration point here — the compliance orchestrator
   will populate `event.controls` with matched control tags. For now, just ensure the
   struct exists and is serialized.

5. **`SearchUtil.java`**: the query builder maps operators to OS query DSL. For
   `flattened` fields, `match_phrase` (IS), `query_string` (CONTAIN), `range` (IS_BETWEEN)
   are **not supported**. Options:
   - (a) Add type-aware query building: if the field is `flattened`, use `term` instead of
     `match_phrase`, skip `range`/`match`.
   - (b) Restrict the UI (Phase 6) to only offer `flattened`-compatible operators, so the
     backend never receives a `match_phrase` on a `flattened` field.
   - **Recommendation: (b)** — the frontend operator filter (Phase 6) is the gate. The
     backend already handles unknown/unmapped fields gracefully (silent no-match per the
     audit). No backend code change needed for `flattened`-specific query building, as long
     as the UI doesn't offer unsupported operators.

6. **`ElasticsearchService.java`**: no change (it's field-agnostic, uses the mapping-driven
   field list from the connector).

**Verify:**
- User-auditor queries on `event.*` fields return results.
- Compliance-orchestrator builds with the new `controls` field.
- No `log.` / `logx.` references remain in backend Java (grep).

---

### Phase 8 — opensearch-connector: `flattened` support in field enumeration

**Repo:** `utmstack/opensearch-connector`
**File:** `util/IndexUtils.java:20-42`

**Problem:** `propertiesFromMapping` only recurses into `isObject()` properties and emits
`.keyword` subfields for `isText()`. A `flattened` property is its own kind (`isFlattened()`
or `flattened()` in the Java client) — it falls into the `else` branch and emits
`log → "flattened"` as a single leaf. No `log.x` children are emitted.

**This is actually correct for `flattened`:** the UI should see `event` as a single field
of type `flattened`, not enumerate its children (which are dynamic and unbounded). The
user picks `event` as a filter field, and the operator is restricted to
term/exists/prefix (Phase 6).

**However**, the UI currently builds filters on `log.eventCode` (a specific subfield).
With `flattened`, the field is `event` (the whole bag), and the filter value would be
a path like `event.eventCode`. This requires a **UI interaction change**: instead of
picking a field from a dropdown and entering a value, the user picks `event` and enters
`eventCode: 4624` as a path-value pair.

**Decision needed:** How should the UI expose `flattened` fields?
- Option A: Show `event` as a single field; user types `eventCode=4624` in the value box
  (the backend translates to `term: {"event.eventCode": "4624"}`).
- Option B: Keep the field picker showing specific `event.*` subfields (requires the
  connector to enumerate `flattened` subfields from `_source` of recent docs — expensive,
  non-deterministic).
- **Recommendation: Option A.** Simpler, honest about the `flattened` type, and the
  operator restriction (Phase 6) makes it clear what's supported.

**If Option A:** No connector code change. The `flattened` field appears as a single
entry in the picker. The frontend needs a small change to render `event` with a
"dot-notation" hint and build `term` queries with the full path.

**If Option B (future):** Add `flattened` enumeration to `IndexUtils` (fetch a sample doc,
extract `event` keys, emit as virtual fields). Defer.

**Verify:**
- `getIndexProperties("v11-log-*")` returns `event → "flattened"` (single entry).
- UI picker shows `event` with type icon for `flattened`.
- Filter on `event` with value `eventCode: 4624` → OS `term` query → results.

---

### Phase 9 — Verification & rollout

**Unit tests (per phase, already listed).**

**Integration test (end-to-end):**
1. Deploy new go-sdk + EventProcessor + UTMStack on a test instance.
2. Send a raw ASA syslog → verify OS doc has `event:{...}` (not `log:{...}`), no mapping error.
3. Send a doc with `event: {"severity": 5}` then `event: {"severity": "high"}` → both accepted (flattened).
4. Run a CEL rule: `equals("event.messageId", "106001")` → alert fires.
5. Run a correlation rule: `afterEvents.field: "event.eventCode"` → finds prior docs.
6. UI: Log Analyzer field picker shows `event` (flattened). Filter on `event` with `eventCode: 4624` → results.
7. UI: File management columns show `event.*` fields. Filter works.
8. User-auditor: queries on `event.*` return results.
9. Dashboard: charts on top-level fields (`dataType`, `action`) work.
10. Old data: query old index on `log.eventCode` → still works (old docs unchanged).

**Rollout order:**
1. go-sdk tag + push (Phase 1)
2. EventProcessor base image rebuild + push (Phase 2) — `TW_EVENT_PROCESSOR_VERSION_PROD` update
3. UTMStack PR: filters + rules rename (Phase 3+4) — large diff, review carefully
4. UTMStack PR: installer template (Phase 5) + frontend (Phase 6) + backend (Phase 7) + connector (Phase 8)
5. Deploy to RC instance → run integration test
6. Deploy to prod

**Rollback:**
- Revert the go-sdk tag → EventProcessor rebuilds with `log` → new docs use `log` again.
- The `event: {type:flattened}` mapping in OS is harmless (unused by `log`-keyed docs).
- Filters/rules revert to `log.` references.
- No data loss: old `log` docs and any `event` docs coexist.

---

## Risks & mitigations

| Risk | Impact | Mitigation |
|---|---|---|
| 24,000+ mechanical renames miss a reference | Dead rule, broken filter, silent no-match | Grep verification (Phase 3+4); contract tests; spot-check per vendor |
| `flattened` breaks UI interactions (dashboard, file mgmt) | Lost functionality | Phase 6 operator restriction; dashboard chart migration to top-level fields; file-mgmt `event.*` filters use `term` not `match` |
| Correlation across old/new data boundary | Rules that correlate on `event.*` miss `log.*` docs (and vice versa) | Accepted (user confirmed). Transition window = ILM rotation period. No mitigation needed. |
| go-sdk wire compat during rolling upgrade | gRPC messages in flight use field 9; old nodes expect `log`, new nodes expect `event` | Proto field number 9 is unchanged. Wire format is field-number-based, not name-based. Old nodes reading field 9 get the same bytes. JSON name only affects protojson (used for OS serialization + CEL), not gRPC binary. **No wire incompatibility.** |
| Frontend localStorage cache serves stale field list | Users see old `log.*` fields | Phase 6: bump `_fields` → `_fields_v3` |
| `controls` field unused (populated "later") | Empty field in every doc (zero bytes, `omitempty`) | `repeated string` with `omitempty` → absent from JSON when empty. No storage cost. |

## Out of scope
- Populating `controls` with actual compliance tags (future work, compliance-orchestrator).
- Migrating old `log` data to `event` (ILM rotation handles this).
- `logx` data (dead, not live on v11).
- Changes to `origin`/`target` (Side) fields (unaffected by this rename).
- Changes to the alert document structure beyond the `event` key rename inside `events[]`.

## Open questions
1. **Dashboard "Top events" chart**: currently on `logx.wineventlog.event_name.keyword`.
   `flattened` doesn't support aggregations. Migrate to `dataType` or `action` (top-level,
   typed), or accept the chart is lost?
2. **UI interaction for `flattened` fields**: Option A (single `event` field, dot-notation
   value) or Option B (enumerated subfields)? I recommend A.
3. **user-auditor transition**: query `event` only, or dual-read `event` + `log` during
   the rotation window?
4. **Filter contract tests**: the `plugins/alerts` test fixtures use `log.*` in expected
   JSON. They need updating to `event.*` as part of Phase 3. Confirm the test suite is
   comprehensive enough to catch missed renames.
