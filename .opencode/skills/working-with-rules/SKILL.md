---
name: working-with-rules
description: Use when creating, editing, debugging, or validating UTMStack correlation rules (the YAML detections under rules/ with where/afterEvents/groupBy), or when a rule never fires, fires on the wrong event, or needs to be deployed and verified end to end. Typical prompts are the rule is not firing, add an afterEvents correlation, deploy this rule, why did this alert attribute to the wrong user, validate the rule against live events. Not for the log-normalization pipeline, use working-with-filters for that.
---

# Working with UTMStack Rules

Rules are YAML in `rules/<vendor>/<name>.yml`. A rule = `where` (CEL on the normalized event) + optional `afterEvents` (correlation: N more matching events within a window) + `groupBy`/`deduplicateBy`. It fires only when the event shape the `where` expects actually exists in the index.

**The #1 failure mode is not logic, it's data shape.** A rule "works" only if every field/value it references is present and typed as expected in the *ingested* event. Before writing or trusting a rule, query the live index for a real sample of that `action` and read its exact fields and JSON types. A rule that references `actionResult=="success"` on an op that never emits `actionResult` is dead, and nothing errors.

## Rule schema
```yaml
dataTypes: [o365]                 # must match the filter's dataTypes
name: "Human Readable Name"
impact: {confidentiality: 3, integrity: 2, availability: 1}
category: "Credential Access"     # MITRE tactic
technique: "T1110 - Brute Force"
adversary: origin                 # origin | target (drives which Side attributes the alert to)
references: [ ...urls... ]
description: "..."
where: |                          # CEL, evaluated against the normalized event
  equals("action","X") && exists("origin.user")
afterEvents:                      # optional; fires when count more matching events in within
  - indexPattern: v11-log-o365-*
    with:
      - field: origin.user
        operator: filter_term
        value: '{{.origin.user}}' # placeholder resolves from the trigger event
      - field: action
        operator: filter_term
        value: 'X'
    within: 24h
    count: 4
groupBy:                          # OR deduplicateBy, not both
  - adversary.user
  - origin.ip
```
See `references/rule-schema.md` for every field, the `afterEvents` operators, and `{{.field}}` placeholder rules.

## Core workflow (always this order)
1. **Ground it in live data.** Query `v11-log-<vendor>-*` for a real sample of the target `action`. Confirm: the op is **ingested** (not in the filter's drop list), the fields you reference **exist**, and their **values/types** match (`actionResult` present? `origin.user` populated? `target.filename` vs `log.SourceFileName`?).
2. **Write** the rule YAML.
3. **Validate** with PyYAML (LSP is unreliable on these files).
4. **Deploy** to Postgres (`utm_correlation_rules` + `utm_group_rules_data_type`) and the backend reseed dir. See `references/deploy-and-verify.md`.
5. **Verify the engine** loaded it (`/workdir/rules/<vendor>/<id>.yaml`), then **generate the real event** and confirm an alert fires with correct attribution. A rule is only "done" when it has fired.

## Critical gotchas
- **`actionResult` is filter-synthesized, not raw.** If the op emits no `log.ResultStatus`, `actionResult` is null → `equals("actionResult","success")` never matches. Check the live sample; drop the clause for such ops (TeamDeleted, MailboxLogin).
- **`contains` needs a string** (scalar-only CEL). Array fields (`log.Parameters`, `log.Members`) require a filter `cast` to string first — otherwise dead. (See `working-with-filters` → `cel-semantics.md`.)
- **Field name drift.** Rules reference the *normalized* field. If a rule says `log.clientIP` but the filter renamed it to `origin.ip`, it's dead. O365: `action` (not `log.Operation`), `origin.ip`, `origin.user`, `target.filename`.
- **`{{.field}}` placeholder must be non-nil** or the whole afterEvents correlation is skipped silently (go-sdk `rules.go` nil-bails). Correlate on a field guaranteed populated (e.g. `origin.user`), not one that can be absent (e.g. `origin.ip` on IP-less events).
- **`groupBy` vs `deduplicateBy` — set one, not both.**
- **`adversary: origin`** attributes the alert to the acting user/IP; `target` to the affected asset. Wrong choice = alert points at the wrong entity.

## Read these when you need them
- `references/rule-schema.md` — full field reference, afterEvents operators, placeholder rules, deploy DDL.
- `references/deploy-and-verify.md` — Postgres schema/insert, engine verify, generating events, the restart-revert trap, ID renumbering.
- `references/known-mistakes.md` — the error classes from past campaigns + `docs/rvald26-draft-prs-review.md`.

**Cross-reference:** the field your rule matches is produced by a filter — when a rule's field is missing or mistyped, fix it there. See the `working-with-filters` skill.
