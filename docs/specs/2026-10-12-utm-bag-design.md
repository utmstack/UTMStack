# Design: `utm.*` — a typed, aggregatable event-field bag

**Status:** Approved (Osmany, 2026-10-12)
**Scope:** v11.3.0 — full implementation (proto → filters → template → UI → liquibase → E2E)
**Supersedes:** The flat_object-only event-bag approach for the 178 viz that bucket on bag fields.

## 1. Problem

The `event` bag was renamed to `flat_object` in v11.3.0 for conflict-proofing. That is correct
for search, but **`flat_object` is not aggregatable** — OpenSearch rejects terms/date_histogram on
it (verified live: 68 docs contain `event.Workload`, yet a terms agg returns 0 buckets).

178 of 368 stored visualizations bucket their charts on bag fields (`event.eventCode` ×51,
`event.providerName` ×24, `event.eventDataProcessId` ×18, … 193 distinct tokens). On any fresh
install these charts are **structurally empty**.

Pre-rename, `log` was dynamically typed as an `object` — its sub-fields *were* aggregatable,
which is why the charts worked. The v10 `log` → v11 `event` rename silently moved those
aggregations onto a non-aggregatable type.

## 2. Design

### Two bags, distinct roles

| Bag | Type | Purpose | Aggregatable? |
|-----|------|---------|---------------|
| `event` | `flat_object` | Raw vendor payload, search-only (Log Explorer, rule `where`), conflict-proof | **No** |
| `utm` | `object` with explicit per-field types | Curated, typed, aggregatable | **Yes** |

### `utm` bag — typing (verified live on OpenSearch 2.19.1)

**Every `utm` field is typed `keyword`.** This is the one type that is safe for all
value shapes and aggregatable:

| write to a `keyword` utm field | result (tested) |
|---|---|
| string `"Exchange"` | stored `"Exchange"` |
| number `4672` | coerced `"4672"` |
| number-as-string `"4672"` | stored `"4672"` |
| bool `true` | coerced `"true"` |
| array `["a","b"]` | joined/coerced, doc kept |
| null | field absent, doc kept |

**No doc is ever rejected** for a `utm` type mismatch (unlike `long`, which 400'd the
whole doc on a string value in the scratch test). `terms` aggregation on `utm.X`
returns real buckets. This satisfies the "skip on mismatch, keep it in the bag"
requirement by construction: there is no mismatch to skip — the value always fits,
and if it's absent the field just isn't indexed.

Template: `"utm": {"type":"object","dynamic":false,"properties":{ "<f>":{"type":"keyword"}, ... }}`
for every registry field. `dynamic:false` means only registered fields are indexed;
stray `utm.Z` writes are silently dropped (the raw value stays in `event`).

Filters populate `utm.X` via the **grok-copy idiom** (non-lossy):
`grok: {source: event.X, patterns: [{fieldName: utm.X, pattern: (?s)^.+$}], where: exists("event.X")}`.
grok always emits strings, which is exactly what `keyword` wants.

### Field registry (per-vendor)

A static list of `utm.X` fields per vendor. The seed is the 193 distinct `event.*` tokens
found in the current viz `aggregation` column, grouped by vendor (windows, o365, aws,
azure, google, cisco, …). Additional fields can be added later by appending to the
registry + filter step + template.

The registry is the **single source of truth** consumed by:
1. **Vendor filters** — emit `utm.X` via `copy`/`add` steps for each registry field.
2. **Index template** (`installer/services/search.go`) — pins `utm` as an `object`
   with each registry field's explicit type.
3. **Frontend chart builder** — the aggregation dimension picker lists `utm.*` fields
   as eligible (they're real typed fields, so `getIndexProperties` already returns them
   once the template is live).
4. **Liquibase remap** — rewrites the 178 viz `aggregation` configs from `event.X`
   (and the 24 stale `log.X`) to `utm.X`.

### Why not promote to top-level?

Promoting 193 fields to the Event proto (one field each) would bloat the schema and the
proto. The `utm` bag groups them under one namespace, keeps the Event message lean, and
makes the "these are UTMStack-curated, aggregatable" intent explicit.

## 3. Layer-by-layer changes

### 3.1 go-sdk (threatwinds/go-sdk)

- `plugins/plugins.proto`: add `map<string, google.protobuf.Value> utm = 21;` to `Event`.
- Regenerate `plugins.pb.go` (protoc).
- Tag `v1.1.38` (next after v1.1.37 which has the `event` rename).
- **No EventProcessor or plugin code changes** — `sjson.Set` on the JSON draft is
  generic, and `protojson.Unmarshal` maps `{"utm": {...}}` onto the new field at the
  end of parsing.

### 3.2 EventProcessor

- Bump `go.mod` to go-sdk `v1.1.38` (lockstep with the UTMStack release).
- No code changes.

### 3.3 UTMStack vendor filters

For each vendor, add a final `copy` step block that writes each registry field from the
`event` bag to `utm.X`. Use `copy` (non-lossy, preserves the source in `event`) —
**not** `rename` (lossy, would remove the field from `event`).

Example (Windows, `filters/windows/windows.yml`):

```yaml
  # --- UTM aggregatable bag (v1.1.38+) ---
  - copy:
      source: event.eventCode
      destination: utm.eventCode
    where: exists("event.eventCode")
  - copy:
      source: event.providerName
      destination: utm.providerName
    where: exists("event.providerName")
  - copy:
      source: event.eventDataProcessId
      destination: utm.eventDataProcessId
    where: exists("event.eventDataProcessId")
  # ... one per registry field
```

The `where: exists(...)` guard means the `utm` copy only happens when the source
is present (avoids writing empty `null` values that could confuse aggregations).

**Vendor priority:** windows (51 of the 178 viz), o365 (12), aws (12), azure (8),
google (5), cisco (4), then the rest. All 8–10 vendors in scope.

### 3.4 Index template (`installer/services/search.go`)

- `logIndexMappings`: add `"utm": {"type":"object","dynamic":false,"properties":{ ...
  per-vendor registry with explicit types ...}}`.
- `newBagMappings` (retro-PUT for existing indices): same `utm` block.
- `alertIndexMappings`: `lastEvent.utm` and `events[*].utm` as `object` with the same
  properties (alerts embed full Event copies).

**`dynamic: false`** on `utm` means only the registry fields are mapped — any stray
`utm.Z` not in the registry is rejected at write time (not silently indexed), which is
the desired behavior.

### 3.5 Backend

- `SearchUtil.isFlattenedField` — **no change**. `utm.` doesn't start with `event.`,
  so it falls into the normal (aggregatable) branch automatically.
- `getIndexProperties` — **no change**. The opensearch-java 2.10.2 client already
  handles `object` type and returns sub-fields; once the template pins `utm.eventCode:
  keyword`, the field picker will list it.

### 3.6 Frontend

- `elastic-filter-add.component.ts`: add a `utm.` synthesis branch mirroring the
  `event.` one (lines 201-203), so typed `utm.<leaf>` paths resolve to their
  server-reported type and get the full operator set.
- `chart-builder/util-field.ts`: ensure `utm.*` fields are **not** excluded from chart
  dimensions (they're typed, not FLATTENED, so they should already pass — verify).
- `ElasticDataTypesEnum` — no new value needed; `object`/`keyword`/`long`/`date` all
  exist.

### 3.7 Liquibase remap (new changeset, `20261012001_remap_viz_agg_to_utm.xml`)

Remap the 178 viz `aggregation` configs:
- `event.X` → `utm.X` (the 178 currently broken by flat_object)
- `log.X` / `logx.X` → `utm.X` (the 24 stale spaced-format no-ops)

This is idempotent (LIKE-guarded REPLACE) and runs on boot. On a fresh install, the
liquibase-seeded viz already use `utm.X` (the seed data is regenerated from the remapped
source), so the changeset is a no-op.

### 3.8 Rules

Rules can reference `utm.X` freely (CEL resolves it generically). **No rule changes
required** for this release — existing rules that use `event.X` in `where:` clauses
continue to work (the `event` bag is unchanged). Future rules can prefer `utm.X` for
clarity.

## 4. Coordination & rollout

**Lockstep (one release):**
1. go-sdk tag `v1.1.38` (proto field 21).
2. EventProcessor base bumped to `v1.1.38`.
3. UTMStack v11.3.1 (next patch) ships: filters + template + frontend + liquibase.

**Rollout order on upgrade:**
- Backend first (liquibase + template + field picker).
- EventProcessor next (starts writing `utm.X`).
- Frontend last (shows `utm.X` in pickers).

**Data:** Old indices keep `event.*` as `flat_object` (unchanged). New docs written by
the new EP carry both `event.X` (bag) and `utm.X` (typed). Existing docs don't have
`utm.X` — charts on historical data will show fewer buckets until ILM rotates, which is
acceptable (the `event` bag is still searchable for Log Explorer).

## 5. Out of scope

- Reindexing old data to add `utm.X` (rides out ILM).
- Modifying rules to use `utm.X` (optional follow-up).
- Adding new `utm` fields beyond the 193-registry seed (incremental, per-vendor).
- Changing the `event` bag's type (stays `flat_object`).

## 6. Risks

| Risk | Mitigation |
|------|------------|
| Proto field 21 collision (if go-sdk adds 21 elsewhere) | Field 21 is currently free (last used = 20 `controls`). Coordinate with go-sdk maintainer. |
| Type mismatch on a registry field (vendor writes `string`, template pins `long`) | OpenSearch rejects the write for that doc; `event` bag is intact. Add a filter-level cast before the `utm` copy if needed. |
| `utm` bag grows large (190+ fields) and bloats docs | `dynamic:false` means only registry fields are stored. Most docs only populate a handful (vendor-specific). |
| Liquibase remap misses a viz (new viz added after the changeset) | The changeset is idempotent; re-running on boot picks up new rows. |
| Frontend chart-builder excludes `utm.*` from dims | Verify in E2E; if excluded, add a whitelist. |

## 7. Open questions (for Osmany)

1. **Registry file format:** should the per-vendor `utm` field list live in a
   machine-readable file (e.g. `registry/utm-fields-windows.json`) that the filter
   generator and template generator both read, or hardcoded in each vendor filter +
   the Go template const? A registry file makes adding fields a one-line change but
   adds a build step.
2. **Field naming:** keep the vendor's original casing (`eventCode`, `Workload`,
   `P2Sender`) or normalize to snake_case (`event_code`, `workload`, `p2_sender`)?
   The viz configs use the original casing; normalizing would require a case-insensitive
   remap.
3. **`eventDataProcessId` type:** Windows event IDs are `long` in most cases, but
   some vendors emit them as strings. Pin as `long` with a filter-level cast, or
   `keyword` (safe but loses numeric sorting)?
