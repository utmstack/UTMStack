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

## Verification — DONE (2026-10-31)
- **500 mechanism proven on a single index:** terms agg on BARE `utm.Workload` over a
  text-inferred o365 index → `illegal_argument_exception: "Text fields are not optimised
  for ... aggregations ... use a keyword field"` → the `/run` 500. `utm.Workload.keyword`
  on the same index returns buckets (Exchange 34, AAD 25, ...).
- **Pin proven shipped:** commit `0a6303a228` (176-key bare-keyword utm pin) IS in the
  v11.3.0 tag; v11.2.16 has 0 occurrences. So v11.3.0 deployments = bare keyword, older =
  inferred text+kw → the cross-generation drift.
- **Migration applied to lab DB (20261030003 SQL, extracted verbatim):** viz 468 bucket
  field `utm.Workload` → `utm.Workload.keyword`; 0 vizs left on bare utm.X; subBuckets
  migrated too; `utm.responseElements.credentials.expiration` (date) preserved bare;
  idempotent re-run = no-op; 0 customLabels touched.
- **Templates deployed stripped** (live log template utm node = `{"type":"object"}`).
- **Contaminated indices reindexed** to the canonical inferred shape (o365 10-30/10-31,
  aws 10-30/10-31): all doc counts preserved (o365 69/178/54/28/6; aws-10-31 restored
  after the in-flight reindex completed: 692,889), 0 failures, no `.new` scratch left.
- **Full sweep: 204/204 utm-bucketed vizs → HTTP 200 (0 500s)** across o365, aws,
  windows, azure, bitdefender, linux, vmware, meraki families, via the real backend
  `/api/utm-visualizations/run`.
- **Dash 13 browser render:** 0 x 500 over 3 attempts; "O365 Top 5 events by Source"
  shows OneDrive/Exchange/AAD/SharePoint/PowerBI; AD Login map shows markers.
- **Installer:** build green, all 6 mapping tests pass (canonical test now asserts utm
  is the inferred bag with NO pin).

## Residual (documented, not blocking)
- Legacy bare-keyword utm indices (any v11.3.0 customer deployment) contribute EMPTY
  buckets for `utm.X.keyword` until reindex/ILM roll-off — not an error. The reindex is
  the convergence path (proven on the lab).
- 36 of the 192 scanned fields are absent from every current lab index (no vendor data
  yet); under the corrected template they infer text+kw on first sight, so `.keyword`
  is correct for them too.
