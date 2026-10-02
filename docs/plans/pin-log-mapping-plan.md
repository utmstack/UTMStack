# Plan — Pin `log` / `logx` OpenSearch mapping to stop dynamic-mapping conflicts

## Problem

`Event.Log` is a heterogeneous map (`map[string]*structpb.Value` → `map[string]any` at the JSON level).
Today there is **no mapping** for `v11-log-*` (`InitOpenSearch` only sets `total_fields.limit` + `max_shards`),
so OpenSearch **dynamically infers** a type for every `log.*` key on first sight and never changes it.
The first value wins: if `log.port` is `443` (→ `long`) and a later vendor emits `log.port = "N/A"`
(→ string), the doc is rejected with a `mapper_parsing_exception` and the log is lost.
This is the class of failure the user is hitting.

## Decision (what we are NOT doing, and why)

| Option | Verdict | Why |
|---|---|---|
| **A. `log: {"type":"flattened"}`** | ❌ | The UI field picker (connector `IndexUtils.propertiesFromMapping`) only recurses into `object`-kind fields. `flattened` is its own kind → `log` collapses to a single picker entry and every `log.x` subfield vanishes from filters/columns. Also not in `ElasticDataTypesEnum` → wrong operators offered, and the backend search translator has no flattened support. |
| **B. Go model → `map[string]string`** | ❌ | Over-reach. The conflict is purely an **OpenSearch mapping** problem; the Go type is correct. Changing the model risks nested-traversal in the CEL engine, protobuf compat, and the alerts/stats/indexers that serialize it — for a fix that belongs in one template. |
| **C. `log: {"type":"object","dynamic":"false"}` + explicitly declared, UI-referenced `log.*`** | ✅ | Conflicts can only occur on *declared, indexed* fields. Undeclared keys are stored in `_source` but never mapped → no type can ever be locked in → no conflict. Declared fields are pinned to one stable type → no conflict. This is the smallest change that preserves the UI 100%. |
| **D. `log: {"enabled":false}`** | ❌ | Kills *all* `log.*` from the index → UI filters/columns/aggregations on `log.*` (file-management module, dashboard "top events") break. |

**C is the plan.** It touches only the **installer** (template producer) + a tiny **frontend** cache-bust.
No go-sdk / EventProcessor / rules changes, and **no backend/connector code change** (verified: the
connector already walks `object` children and never inspects `dynamic`).

### Type choice for declared `log.*` fields: `text` + `keyword` subfield (the status-quo shape)

OpenSearch's *default dynamic template* maps any JSON **string** to
`{"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}}`. That is exactly what the
UI already sees today for every string `log.*` field, and the code already appends `.keyword` for
`IS_ONE_OF`/distinct-values (e.g. `file-generic-filter.component.ts:108`). Pinning declared fields to
**the same shape** is therefore behavior-preserving:

- The field picker (`IndexUtils.propertiesFromMapping`) emits `log.x → "text"` **and** `log.x.keyword → "keyword"` — identical to today.
- All string `log.*` values that arrive in `_source` as *numbers/booleans* (e.g. `eventCode: 4624`) are
  coerced to the text field on write (no mapping conflict); they display and exact-match fine.
- **No declared field is ever `long`**, so there is no path by which a string can conflict with a number —
  the original bug is eliminated at the type level, not just by `dynamic:false`.
- Numeric **range** filtering on `log.*` is not used by the UI (the UI never ranges `log.*`; numeric
  thresholds are handled by the CEL rule engine against `_source`, which keeps real types). So we lose
  nothing in the product by declaring them text.

This is the key insight that makes C safe: **declare-as-text removes the type dimension from the
conflict entirely**, so even the declared fields cannot conflict, and `dynamic:false` only needs to
protect against *new* keys.

---

## Work items

### W1 — Installer: pin the `v11-log-*` mapping (fresh install)

**Repo:** `UTMStack` · **File:** `installer/services/search.go` (add a new constant; no other file changes)

Extract a reusable JSON constant so fresh + upgrade share one source of truth (avoids drift):

```go
// logIndexMappings is the fixed mapping for v11-log-* documents.
//
// `log` and `logx` are pinned to dynamic:false so OpenSearch never locks a
// type in from an arbitrary vendor value (the cause of mapper_parsing_exception
// on first sight). Only the fields the product UI actually filters, columns,
// and aggregates on are declared — everything else is stored in _source (used
// by the CEL rule engine) but not indexed.
//
// Declared fields are text+keyword (the shape OpenSearch's default dynamic
// template already produces for strings), which is what the UI expects.
const logIndexMappings = `
{
  "@timestamp":   {"type":"date"},
  "dataType":     {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
  "dataSource":   {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
  "protocol":     {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
  "statusCode":   {"type":"long"},
  "action":       {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
  "actionResult": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
  "severity":     {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
  "origin": {"type":"object","dynamic":true,"properties":{
      "ip":   {"type":"ip"},
      "port": {"type":"long"},
      "user": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
      "host": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
      "bytesSent":     {"type":"long"},
      "bytesReceived": {"type":"long"}
  }},
  "target": {"type":"object","dynamic":true,"properties":{
      "ip":   {"type":"ip"},
      "port": {"type":"long"},
      "user": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
      "host": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
      "bytesSent":     {"type":"long"},
      "bytesReceived": {"type":"long"}
  }},
  "log": {"type":"object","dynamic":false,"properties":{
      "eventCode":    {"type":"keyword"},
      "eventName":    {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
      "message":      {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
      "computer":     {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
      "cpuArchitecture":{"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
      "keywords":     {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
      "opcode":       {"type":"keyword"},
      "providerGuid": {"type":"keyword"},
      "timestamp":    {"type":"date"},
      "eventDataAccessList":      {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
      "eventDataAccessMask":      {"type":"keyword"},
      "eventDataHandleId":        {"type":"keyword"},
      "eventDataObjectName":      {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
      "eventDataObjectServer":    {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
      "eventDataObjectType":      {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
      "eventDataProcessId":       {"type":"keyword"},
      "eventDataProcessName":     {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
      "eventDataResourceAttributes":{"type":"keyword"},
      "eventDataSubjectDomainName":{"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
      "eventDataSubjectLogonId":  {"type":"keyword"},
      "eventDataSubjectUserName": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
      "eventDataSubjectUserSid":  {"type":"keyword"},
      "eventDataNewSd":           {"type":"keyword"},
      "eventDataOldSd":           {"type":"keyword"},
      "eventDataShareName":       {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
      "eventDataShareLocalPath":  {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
      "host": {"type":"object","dynamic":false,"properties":{
          "os": {"type":"object","dynamic":false,"properties":{
              "build":    {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
              "family":   {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
              "platform": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
              "version":  {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256"}}}
          }}
      }}
  }},
  "logx": {"type":"object","dynamic":false,"properties":{
      "type": {"type":"keyword"},
      "wineventlog": {"type":"object","dynamic":false,"properties":{
          "event_name": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}}
      }}
  }}
}`
```

Wiring (both in `search.go`):
- `InitOpenSearch()`: change the existing `logTemplateData` (currently settings-only) so its
  `template` object is `{"settings":{"index.max_shards":30000},"mappings":{"properties": <logIndexMappings>}}`.
  The general `utmstack_indexes` template (`total_fields.limit`, shard/replica settings) stays as-is —
  composable templates merge, so mappings + settings coexist.
- `UpdateOpenSearch()`: add a `PUT _index_template/utmstack_log_indexes` with the **same** constant
  so upgrades (lock `20260926001`) also get the pinned mapping. Keep the existing `max_shards` settings PUT.

**Why `origin`/`target` are `dynamic:true`:** their keys come from the *canonical schema* (filters
already guard types before writing), and they include geo/bytes the UI ranges/sorts on, so we let them
map naturally. The conflict-prone surface is `log`/`logx` (raw vendor bags) — that is what we pin.

**Verify:** `go build ./installer/...`; on a throwaway OpenSearch, `PUT` the template, index a doc with
`log.eventCode=4624` then `log.eventCode="4624-abc"` → **both accepted** (previously the 2nd would 400).
Confirm `GET v11-log-*/_mapping` shows `log` as `object/dynamic:false` with the declared children.

---

### W2 — Installer: upgrade path for *existing* (already-broken) indices

**Repo:** `UTMStack` · **Files:** `installer/utils/opensearch.go` (new helper), `installer/services/search.go`

A new template only fixes **newly created** indices. Indices that already locked a conflicting type keep
rejecting docs until reindexed. Add a best-effort repair (idempotent, non-fatal):

- New helper `ReindexLogIndices(containerID, password string)`:
  - `GET /v11-log-*/_stats` (or iterate `GET /_cat/indices/v11-log-*`) for indices with docs.
  - For each, `POST _reindex` to `v11-log-<ts>-reindexed` (the new index inherits the pinned template),
    then `POST /v11-log-<ts>/_alias` swap (atomic: add new + remove old alias under one request) and
    delete the old index.
  - Wrap per-index in `recover`/error-logged; **never fail the install** if one index errors (log and move on).
- Call it from `UpdateOpenSearch()` **after** the template PUT.

> **Default is "let ILM rotate, reindex on demand."** Most customers will not have widespread conflicts;
> for those who do, this is an opt-in repair. The alias-swap is the safe atomic step — no downtime window.

**Verify:** reproduce a broken index (lock `log.port` to `long`, push a string), run the reindex, confirm
the new index accepts both and the alias serves it.

---

### W3 — Frontend: bust the cached field list

**Repo:** `UTMStack` · **File:** `frontend/src/app/shared/services/elasticsearch/local-field.service.ts`

The picker caches the mapping-derived field list in `localStorage` under `<pattern>_fields`
(`local-field.service.ts:17`). Without a bust, users keep a *stale* list (pre-pinning, with `log.*`
fields now undeclared) and may build filters on fields that no longer resolve (silent no-match per the
backend audit). Bump the suffix:

```ts
export const INDEX_PATTERN_FIELD = '_fields_v2';   // was '_fields'
```

First load after deploy re-fetches (`field-data.service.ts:20` re-queries when the stored list is empty)
and stores under the new key; the old key becomes dead weight (harmless).

**Verify:** bump, clear nothing manually, load Log Analyzer → field picker shows the *pinned* `log.*`
set (and no `log.*` that was previously dynamic-only).

---

### W4 — Backend: **no code change** (verification only)

**Repo:** `UTMStack` / `opensearch-connector` · (no edits)

The audit confirmed the connector (`IndexUtils.propertiesFromMapping`) only uses `isObject()`/`isText()`
and the `.keyword` subfield presence — it **ignores `dynamic`** — so it already enumerates the pinned
`log.*` children exactly as before. No backend/connector edit is required. **Do** add a regression test:

- In `opensearch-connector` (or backend `src/main/java` test tree), spin an embedded/in-docker OpenSearch,
  `PUT` the pinned template, index a doc, assert `getIndexProperties("v11-log-*")` returns
  `log.eventCode`, `log.eventName`, `log.eventName.keyword`, `log.host.os.build`, … and that a
  **non-declared** `log.*` key is *absent* from the list but *present* in `_source`.
- Assert the two 500-prone paths degrade gracefully: a `match_phrase` filter on a non-declared `log.*`
  returns 0 rows (not 500), and a sort on one returns the mapping error (documented, expected).

**Verify:** test passes; confirms the template is sufficient without touching Java.

---

### W5 — Frontend field-coverage audit (gate for W1 completeness)

**Repo:** `UTMStack` · (read-only)

Run before freezing the W1 constant: grep all of `frontend/src` for `log\.[A-Za-z0-9_.]+` and
`logx\.[A-Za-z0-9_.]+` (string literals, constants, templates), and union with the file-management
`FILE_*` enum. Every distinct field the UI references **must** be in the W1 declared set; if the audit
surfaces additional ones (e.g. another module's `log.<x>` column), add them to the constant. This is the
single biggest risk (a missing field = silently removed from the picker), so the audit gates the merge.

**Verify:** produce the list; diff against W1; add any gaps; re-run W1 verify.

---

## Rollout

- **New installs:** get the pinned template from first boot (lock 7 → `InitOpenSearch`).
- **Upgrades:** lock `20260926001` → `UpdateOpenSearch` PUTs the same template (W1) + optional reindex (W2).
- **Existing broken indices:** W2 reindex, or let ILM rotate; new data lands in pinned indices regardless.
- **Ordering:** land W5 (audit) → freeze W1 constant → W1 + W3 + W4 in one PR (installer + frontend, no
  cross-repo dep) → W2 in a follow-up once reindex is proven on a live box.

## Non-goals
- No change to the canonical schema, filters, rules, go-sdk, or EventProcessor.
- No numeric `log.*` range UI — not a product feature today.
- No `flattened`/`enabled:false` migration.

## Open questions (need your call before I start)
1. **Declared `log.*` field set:** I've pinned the Windows file-management set + `logx.type`/`logx.wineventlog.event_name`. Does the W5 audit reveal modules I'm missing (e.g. any `log.*` in dashboard/IR/alert-detail defaults)? Confirm the final list.
2. **`origin`/`target` `dynamic:true`:** acceptable, or do you want them fully pinned too (stricter, but requires enumerating the whole canonical `Side` + geolocation set)?
3. **W2 default:** auto-reindex on upgrade, or only on-demand? (I'd default on-demand to avoid surprise reindex cost.)
4. **Frontend key bump** (`_fields` → `_fields_v2`): fine, or do you prefer a versioned constant in one place instead?
