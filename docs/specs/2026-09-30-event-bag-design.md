# Design: replace the `log`/`logx` bag with a flattened `event` + `controls` field

**Date:** 2026-09-30
**Status:** Approved for planning
**Issue:** [utmstack/UTMStack#2791](https://github.com/utmstack/UTMStack/issues/2791) (UTMStack OSS project, v11.2.x, feature, @osmontero)

## 1. Problem

`Event.Log` is a heterogeneous map (`map<string, google.protobuf.Value>` in `go-sdk/plugins/plugins.proto`, field 9 of `Event`). Documents land in `v11-log-*` indices with **no mapping**, so OpenSearch dynamically infers a hard type for every `log.*` key on first sight and never changes it. A later document with a different type for the same key (`log.port=443` then `log.port="N/A"`) is rejected with `mapper_parsing_exception` and the log is lost.

`logx.*` is dead v10 legacy and not live on v11.

The UI field picker is 100% driven by the live mapping (`opensearch-connector` `IndexUtils.propertiesFromMapping` walks `object` children + `text`→`.keyword` subfields), so any fix must keep the mapped surface the UI depends on.

## 2. Goals

- New log documents can never be rejected for `log`/bag field type conflicts.
- The UI keeps working with no field-pickup changes beyond declaring new types.
- Numeric/semantic fields that need range/sort/aggregate keep working via the canonical `origin`/`target`/top-level schema.
- A `controls` field is available for future compliance tags, at zero cost when empty.

## 3. Non-goals

- Reindexing or migrating existing `log`-keyed documents (ILM rotation handles it).
- Correlation across the old/new data boundary (new `event.*` afterEvents won't match old `log.*` docs — accepted).
- Populating `controls` (compliance-orchestrator work, later).
- Changes to the `Side` (origin/target) schema itself.

## 4. Core design decisions

### 4.1 Rename the bag at its single source of truth

The `log` JSON key is the protobuf field name for `Event.Log`. Renaming the proto field to `event` cascades the new key through the draft JSON, the OpenSearch document, and CEL rule evaluation in one move — no dual namespace, no translation layer:

```proto
message Event {
  ...
  map<string, google.protobuf.Value> event = 9;   // was: log
  ...
  repeated string controls = 20;                  // NEW
}
```

Field number 9 is unchanged (wire format is field-number-based, so in-flight gRPC stays compatible); only the protojson key changes. `controls` is a new field 20; empty lists serialize to nothing (proto3), so the field costs zero until populated.

**Do NOT rename** `message Log` (gRPC input message) or `Draft.log` (a JSON string field) — only `Event.log`.

### 4.2 Map `event` as OpenSearch `flat_object`

OpenSearch's `flat_object` (Elasticsearch's `flattened` under a different name — UTMStack ships OpenSearch, so the mapping type string must be `flat_object`) stores every sub-key as a keyword and never infers types → **conflicts are impossible by construction** (this is what cures the bug). Trade-offs, accepted:

- No aggregations, sorting, or numeric range on `event.*` (OpenSearch `flat_object` limitation). Fields that need those belong in the canonical schema and must be promoted by filters.
- Supported queries: `term`, `terms`, `terms_set`, `prefix`, `range` (lexical), `match`, `multi_match`, `query_string`, `simple_query_string`, `exists`, `wildcard`. **Not supported: `match_phrase`.**

### 4.3 Keep `origin`/`target` typed

`Side` is the controlled canonical schema (filters already guard types before writing). It stays `dynamic` with typed properties pinned where the UI ranges/sorts/aggregates (`origin.port`, `target.port` long; `origin.ip`/`target.ip` ip; bytes double; `origin.file`/`target.path`/`target.file` text+keyword). This is the "promote what you need to sort on" rule.

### 4.4 UI/connector requires no code change to enumerate

The connector's `IndexUtils.propertiesFromMapping` only checks `isObject()`/`isText()` and never reads `dynamic`, so it already handles the pinned top-level fields. A `flat_object` field appears as one field of type `flat_object` — which the frontend must learn about (Task: add `FLATTENED = 'flat_object'` to `ElasticDataTypesEnum` and restrict its operators).

### 4.5 The backend query builder becomes flattened-aware

`SearchUtil.toQuery` is operator-blind: `IS`/`IS_ONE_OF`/`CONTAIN_ONE_OF` emit `match_phrase`, which `flattened` rejects. For any filter field starting with `event.` (the only flattened field shipped), operators translate to the supported set:

| Operator | Current | Flattened |
|---|---|---|
| `IS` | `match_phrase` | `term` |
| `IS_NOT` | `must_not match_phrase` | `must_not term` |
| `IS_ONE_OF` / `CONTAIN_ONE_OF` | `bool.should(match_phrase)` | `terms` |
| `IS_NOT_ONE_OF` / `DOES_NOT_CONTAIN_ONE_OF` | `must_not bool.should(match_phrase)` | `must_not terms` |
| `START_WITH` | `wildcard v*` | `prefix` |
| `NOT_START_WITH` | `must_not wildcard v*` | `must_not prefix` |
| `ENDS_WITH`/`NOT_ENDS_WITH` | `wildcard *v` | unchanged (supported) |
| `CONTAIN`/`DOES_NOT_CONTAIN` | `query_string *v*` | unchanged (supported) |
| `EXIST`/`DOES_NOT_EXIST` | `exists` | unchanged |
| `IS_BETWEEN` | `range` (no format) | unchanged — lexical range works on flattened |
| `IS_GREATER_THAN` / `IS_LESS_THAN_OR_EQUALS` | `range` + `.format(date)` | **rejected with 400** — the date `format` clause is invalid on flattened keyword values; the UI never offers them on `event.*` (section 4.6) |

Genuinely-unsupported combinations throw `ApiException(BAD_REQUEST)` — a loud 400, not a silent wrong result.

### 4.6 Sorting and aggregation on `event.*`

OpenSearch `flat_object` **cannot be sorted or aggregated** (it is "not indexed for fast lookup"). Both surfaces are handled:

- **UI:** `dynamic-table.component.ts isSortableColumn()` returns `false` for `FLATTENED` columns — the header is never offered as sortable. Chart dimensions come from the same field picker, and `flattened` columns are excluded from bucket/metric aggregation options (operator-restriction set, Task 12).
- **Backend (defense-in-depth for API callers):** `SearchUtil.applySort`/`applyPaginationAndSort` guard sort fields starting with `event.` and throw `ApiException("Sorting is not supported on flattened field [x]. Promote the field to a canonical field if you need to sort by it.", BAD_REQUEST)` — mapped to a clean **400** by the existing `GlobalExceptionHandler`. Today this path returns **500** (`RuntimeException` → catch-all `INTERNAL_SERVER_ERROR` in `ElasticsearchResource`), which is wrong status and a false application-error event.
- **SQL path:** free-text SQL that sorts on `event.*` surfaces OpenSearch SQL's own 400; accepted as a documented limitation (we do not parse/rewrite arbitrary SQL).

### 4.7 Migration: no reindex

New documents carry `event` + `controls`; old documents keep `log`. The installer:
- Fresh install: `v11-log-*` composable template gets the pinned mappings (`event` flattened, `controls` keyword, typed top-levels + origin/target).
- Upgrade: re-PUT the template (new indices) **and** `PUT v11-log-*/_mapping` with only the *new* keys `event`/`controls` — adding them is non-destructive and cannot conflict with existing dynamic mappings. Typed top-levels are NOT retro-applied to existing indices (adding typed subfields to already-dynamic `origin`/`target` can 400 on a type mismatch); existing indices keep their current mappings until ILM rotation.

Old `log.*` data remains queryable on `log.*` (its mapping is untouched).

### 4.8 File-management module → canonical fields

The module must not filter/sort on flattened bag fields. Decision: the Windows filter promotes the two fields with clean canonical slots, and the module uses those:

| Bag field | Promoted to | Why |
|---|---|---|
| `event.eventDataObjectName` | `target.path` | accessed object's full path; `Side.path` exists |
| `event.eventDataProcessName` | `origin.file` | acting process; `Side.file` exists |
| `event.computer` | *(already `target.host`)* | existing promotion |
| everything else (`eventCode`, `accessMask`, share fields, SDDL, host.os.*, …) | stays in `event.*` | no `Side` slot; filtered via flattened dot-path (`term`/`terms`), displayed from `_source` |

Dashboard "Top events" chart (was dead `logx.wineventlog.event_name.keyword`) moves to **`action`**; the Windows filter promotes `event.eventName`→`action` (same idiom the AWS filter already uses: `log.eventName→action`), so the chart is populated for wineventlog.

## 5. Architecture / change surface

Single-source cascade, three layers:

```
go-sdk (plugins.proto)          Event.log -> Event.event (+controls)  [source of truth]
   │  (protojson key)
   ├─ EventProcessor            json/xml/kv parsers write "event.%s"; 16 modules bump go-sdk; base image
   ├─ UTMStack filters/*.yml    log. -> event.  (scripted; ~19.7k refs)
   ├─ UTMStack rules/*.yml      log. -> event.  (scripted; ~4.2k refs)
   ├─ UTMStack installer        v11-log-* template: event flattened, controls keyword, typed top-levels
   ├─ UTMStack backend          SearchUtil flattened-aware; sort guard -> 400; alert CSV reads event
   ├─ UTMStack user-auditor     reads event only; dead logx constants dropped
   ├─ UTMStack frontend         FLATTENED type + operators; file-mgmt -> origin/target; dashboard -> action; AD item.event.*; cache-bust
   └─ UTMStack plugins/alerts   contract-test fixtures Event{Event: ...}
```

No DB migration of stored definitions: `DefinitionSyncService` (backend `CommandLineRunner`) resyncs `utm_logstash_filter` / `utm_correlation_rules` from the repo YAML on every backend start (filters by content hash, rules by `rule_name`). Caveat: renamed filters are deleted + re-created with **new IDs** — verified nothing references filter IDs by value.

## 6. Components and responsibilities

| Unit | Responsibility | Depends on |
|---|---|---|
| go-sdk `Event` proto | canonical field names + types | — |
| EventProcessor parsers | write bag keys under `event.` | go-sdk |
| filters | map vendor fields → canonical/event paths | EventProcessor |
| rules | CEL/afterEvents on `event.*`/canonical | filters |
| installer template | pin OS mappings | — |
| `SearchUtil` | operator→DSL, flattened-aware, 400 on bad combos | installer types |
| `dynamic-table`/`operator.service` (FE) | what's offered per type | installer types |
| file-management module | columns/filters on canonical + event dot-paths | filters, FE |
| user-auditor | `event` reads for AD audit | filters |

## 7. Error handling

- Unsupported operator/sort/agg on `event.*` → **400 `ApiException`** with a message naming the field and advising promotion to a canonical field (never 500, never silent no-match).
- `flattened` query support enforced at build time: only `term`/`terms`/`prefix`/`match`/`query_string`/`exists`/`wildcard` are ever emitted for `event.*`.
- user-auditor null-guards `getEvent()` (old docs have no `event`).
- Frontend field-list cache busted (`_fields` → `_fields_v2`) so stale pre-rename field lists don't survive a deploy.

## 8. Testing strategy (TDD)

| Layer | Test |
|---|---|
| go-sdk | `Event` marshals to `event.*`/no `log`; round-trips back; `controls` omitempty |
| EventProcessor | builds; smoke doc lands with `event`, not `log` |
| filters/rules | scripted-rename verification (0 `log.` field refs remain; 0 `logx.`); YAML validity |
| installer | JSON validity of mapping constant; manual conflict-proof index (two conflicting `event` shapes both 201) |
| backend `SearchUtil` | `IS`→term, `IS_ONE_OF`→terms, `START_WITH`→prefix on `event.*`; text fields unchanged; **sort on `event.*` → `ApiException` 400** |
| frontend | build; `FLATTENED` columns not sortable in dynamic table; operator list restricted |
| alerts plugin | contract fixtures compile/pass on `event.*` |
| E2E (RC) | ingest→no mapping error→CEL fires→correlation→UI filter/sort behavior→file module→dashboard `action` chart→user-auditor→`controls` term query |

## 9. Rollout and rollback

Coordinated single release: go-sdk tag → EventProcessor base → UTMStack (filters/rules/installer/backend/frontend). `TW_EVENT_PROCESSOR_VERSION_PROD` flips **only after** the UTMStack release with renamed YAML is cut (engine + filters ship in separate pipelines; desync = silent breakage). Rollback: revert both; the added OS mappings (`event`/`controls`) are harmless no-ops when unused; no data loss (`log` and `event` docs coexist).

## 10. Risks

| Risk | Mitigation |
|---|---|
| Missed `log.` reference in a 24k-line scripted rename | grep-verification steps per dir + contract tests + E2E; per-vendor commits |
| Renamed filter gets a new DB ID; something references it | Task verification grep (none found to date) |
| `event.*` numeric range via UI | UI restricts operators; backend 400s; numeric thresholds remain CEL-on-`_source` (real types) |
| wineventlog `action` is a sentence, not a verb | accepted; consistent with AWS idiom |
| rolling upgrade mixes `log`/`event` docs | deployment is atomic (one release); old/new coexist safely in OS |
