---
name: working-with-filters
description: Use when creating, editing, debugging, or deploying UTMStack event-processor filters (the YAML pipelines under filters/ that normalize vendor logs into the go-sdk Event/Side schema), or when a filter field is missing, a rule cannot match a value, actionResult is wrong, or a dropped op needs to be kept. Typical prompts are edit the o365 filter, add a rename or cast step, why is log.Parameters not matching, deploy the filter, or the filter reverted after a restart. Not for correlation rules, use working-with-rules for those.
---

# Working with UTMStack Filters

Filters are YAML in `filters/<vendor>/<name>.yml` — a `pipeline` of ordered `steps` that turn a vendor's raw log into the canonical `Event` + `Side` (origin/target) shape. Rules then match the **normalized** field, never the raw one.

**The single most important mental model:** a filter field only exists if (1) it is a valid go-sdk proto path AND (2) a preceding step actually populates it. Unknown top-level paths are **silently dropped** at proto conversion — no error, the field just never appears.

## Core workflow (always this order)

1. **Read the raw shape first.** Before writing a step, query the live index to see the exact fields a real event carries and their JSON types (string vs array vs int). See `references/deploy-and-durability.md` for the no-JWT query snippet. Guessing field names is how filters go dead.
2. **Edit** `filters/<vendor>/<name>.yml`.
3. **Validate with PyYAML, never the LSP.** The LSP throws false `All sequence items must start at the same column` / `Implicit map key` errors on these files. The source of truth:
   `python -c "import yaml; d=yaml.safe_load(open('filters/office365/o365.yml',encoding='utf-8')); print(len(d['pipeline'][0]['steps']))"` (use the utmstack venv python, set `PYTHONIOENCODING=utf-8`).
4. **Deploy to all three layers** or it reverts on restart. See `references/deploy-and-durability.md`.
5. **Verify the engine** actually reloaded it (`/workdir/pipeline/filters/<id>.yaml`), not just the DB.

## Critical gotchas (the ones that cause dead detections)

- **CEL is scalar-only.** `contains()` / `startsWith` / `endsWith` require the gjson value to be a *String*; `equals`/`oneOf` compare scalars. Array/object fields (`log.Parameters`, `log.Members`, `Target`) return **false**. If a rule must `contains` one of them, add a `cast: {fields:[...], to: string}` step — but `cast` on an array stringifies it, so `contains` then works. Confirm in `go-sdk .../plugins/cel_overloads.go`.
- **`action` is a RENAME, not a raw field.** In O365, `log.Operation → action`. A `drop`/`rename` keys on `action` only work *after* that rename; key on `log.Operation` if you move them before it. Renames are lossy — the source is deleted.
- **`actionResult` is ADDED by the filter**, derived from `log.ResultStatus`. Order of `add` steps matters (later `add` wins for a matching event), and some vendors lie (AAD reports `ResultStatus: Success` on a *failed* login → force-override to `failed`).
- **`drop` is a performance lever.** Put the `drop` step as early as possible (right after `json`) keyed on the raw op field, so dropped events skip all renames/`add`s and the geolocation `dynamic` plugin call. Reorder is safe iff the match set is identical.
- **Schema is proto-gated.** `Side` has `host`, not `hostname`; `Event` has no `system.*`. Writing `origin.hostname` silently produces nothing.

## Read these when you need them
- `references/cel-semantics.md` — every CEL function + its exact type requirement (from go-sdk).
- `references/deploy-and-durability.md` — the 3-layer deploy, container re-lookup, restart-revert trap, filter-ID renumbering, no-JWT queries.
- `references/known-mistakes.md` — error classes from past campaigns + the `docs/rvald26-draft-prs-review.md` findings, so you don't repeat them.

**Cross-reference:** the rule that consumes a field is the one to check when a filter change lands — see the `working-with-rules` skill.
