# utm.* viz migration to .keyword (fixes dash 13 500 + the whole bare-utm class)

## Root cause (verified)
- 192 distinct bare `utm.X` field values in `utm_visualization.aggregation` (any JSON depth:
  bucket + subBucket chains). With `utm.*` inferred, every `utm.X` string is `text`+`.keyword`
  (grok plugin `sjson.Set`s a string -> `text`; confirmed `EventProcessor/plugins/grok/main.go:93,105`).
- A `terms` aggregation on a `text` field with no fielddata = "all shards failed" -> backend 500.
  Dash 13 viz 468 buckets bare `utm.Workload` (o365 = text+kw) -> 500.
- The stock v11.3.0 `newBagMappings` retro pins 176 utm keys BARE keyword on existing indices,
  so a wildcard spanning pinned+inferred indices can't use ONE field string -> the product ships
  this drift. Migrating every bare `utm.X` to `utm.X.keyword` makes ALL indices return data
  (inferred have .keyword; legacy bare-keyword contribute empty, not error) and fresh indices
  (plain-inferred) have it too.

## Locked decision (user-confirmed)
Migrate vizs to `utm.X.keyword`; `utm.*` is a plain inferred object (pin stripped from
logIndexMappings + alertIndexMappings, uncommitted in installer/services/search.go).

## Scope
- EXCLUDE from migration (present in the 192, numeric/date, no .keyword, bare works fine):
  `utm.responseElements.credentials.expiration` (date) — 2 vizs (747, 753) stay bare.
- Every other bare `utm.X` -> `utm.X.keyword` (both `"field":"utm.X"` compact and
  `"field": "utm.X"` spaced, since stored JSON whitespace varies).
- Non-COUNT metrics on utm.*: NONE (verified) — no metric-field migration needed.
- sqlQuery vizs referencing utm.*: NONE (verified).

## Artifacts
1. `installer/services/search.go` — utm pin already stripped (verify + commit).
2. `installer/services/search_test.go` — canonical test updated for inferred utm (verify + commit).
3. `backend/.../liquibase/changelog/20261030003_utm_viz_keyword.xml` — new changeset.
4. `backend/.../liquibase/master.xml` — include 20261030003.

## Lab actions
- Deploy stripped templates (log+alert) via PUT.
- Reindex contaminated small indices (aws, o365) so utm.* = text+kw (remove my pin artifact).
- Apply 20261030003 SQL to lab DB.
- Verify: dash 13 0 x 500; viz 468 returns rows; o365 + aws utm.* = text+kw uniform.
