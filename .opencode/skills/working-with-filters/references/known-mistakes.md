# Known filter mistakes (do not repeat)

Distilled from the O365 validation campaign and `docs/rvald26-draft-prs-review.md` (review of another engineer's 45 filter/rule PRs). Each entry: the mistake, the symptom, the fix.

## 1. Keying a `contains` on an array field
- **Mistake:** `where: contains("log.Parameters","ForwardTo")`.
- **Why dead:** `log.Parameters` is an array → `contains` returns false (scalar-only CEL). 11+ O365 rules shipped this way and could never fire.
- **Fix:** filter `cast: {fields:[log.Parameters, log.Members], to: string}` so the array is stringified before any rule `contains` runs. Verified live (843/824/846 all fired after the cast).

## 2. Assuming a normalized field that isn't there
- **Mistake:** referencing `log.clientIP` after the filter renamed `log.ClientIP → origin.ip` (rename deletes the source), or `MailboxOwnerUPN` (never emitted by O365).
- **Symptom:** rule never matches; no error.
- **Fix:** query the live index first to confirm the exact field name + JSON type before writing any step. `MailItemsAccessed` has `MailboxOwnerSid`/`LogonType`, not a UPN.

## 3. IP-guard that empties `origin.ip` and silently kills `{{.origin.ip}}` correlation  🔴 (rvald26 Issue 1)
- **Mistake:** appending a "keep only real IPs" guard:
  ```yaml
  - rename: {from:[origin.ip], to: log.unparsedOriginIp,
             where: exists("origin.ip") && !(inCIDR(...))}
  ```
  Any non-IP source (`"-"`, hostname — very common in Windows Kerberos 4768/4769) gets `origin.ip` emptied.
- **Why it breaks:** go-sdk `rules.go` bails when the `{{.origin.ip}}` afterEvents placeholder resolves to nil → the correlation **never runs → rule stops firing, silently**. Copy-pasted across 22 filters; 135 rules repo-wide depend on `origin.ip`.
- **Fix:** per-vendor — if the source can be a non-IP, do NOT move it out of `origin.ip`; or correlate on `origin.user`; or add an `or` branch for absent IP. Windows: keep the base `"-"` behavior.

## 4. Invalid proto paths (silently dropped)
- `origin.hostname` → `Side` has **`host`, not `hostname`** (VMware: 3 rules never fired).
- `system.hostname` → `Event` has **no `system`** (macOS: the real lever is `dataSource`, which the *agent* sets, not the filter).
- `command` at top level → must be `origin.command` (deceptive-bytes).
- `field_name:` → `fieldName:` in a `reformat` step (generic-input, syslog-generic).
- `csv from:` → `csv source:` (paloalto ×7).
- **Symptom:** proto conversion drops the unknown path; nothing errors.
- **Fix:** validate every written path against the go-sdk proto (`plugins/plugins.pb.go` `Event`/`Side`).

## 5. Wrong CEL function / typo
- `oneof(...)` → `oneOf(...)` (gcp: never compiled).
- `lgreaterOrEqual(...)` → `greaterOrEqual(...) && lessOrEqual(...)` (cisco-firepower).
- `fildName:` → `fieldName:` (ibm-aix).
- `int64` in a `cast` → invalid, use `int` (paloalto).
- **Fix:** compile-check CEL; cross-reference `references/cel-semantics.md`.

## 6. `actionResult` vocabulary drift
- ~12 filters re-spell `failed→failure`, `blocked→denied`, `pass→success`. Not a break by itself, but cross-vendor dashboards/queries filtering `actionResult=="failed"` silently stop matching those vendors.
- **Fix:** make one explicit cross-vendor vocabulary decision and migrate the queries — don't let it drift per-PR.
- **O365-specific:** the value must be lowercase `success`/`failed`/`blocked`; ~35 rules used capitalized values → dead. AAD failed-logins must be force-overridden to `failed` (see #7).

## 7. Trusting `ResultStatus` blindly (AAD quirk)
- AAD reports a *failed* sign-in as `ResultStatus: Success` (the API call succeeded; `LogonError`/`FailureReason` carries the real outcome). Generic `ResultStatus→actionResult` mapping then labels failures as `success`.
- **Fix:** append an override: `add: {key: actionResult, value: failed, where: equals("action","UserLoginFailed")}` **after** the generic mapping.

## 8. Duplicate/overlapping conditions
- `PartiallySucceeded` was mapped to `success` in two `add` steps (one `oneOf`, one `equals`) — the first ran, the second was dead code. Evaluate each value exactly once.

## 9. Delete list clobbers fields rules need
- A filter's `delete` step removed `log.bytes`/`log.packets` that 4 netflow rules referenced → dead. (rvald26 #2612.)
- **Fix:** before adding to `delete`/`drop`, grep all rules for the field.

## 10. `lastEvent.*` groupBy/dedup was a silent no-op (foundation bug, #2590)
- The wire `Alert` serializes events as `events[]`; there is **no** top-level `lastEvent` key. Grouping code did `gjson.Get(wireAlert,"lastEvent.origin.user")` → always Null → dedup/grouping silently never applied, for **231 rules**.
- **Fix (correct place):** a plugin helper that rewrites `lastEvent.<f>` → `events.<lastIndex>.<f>` for the *value* while keeping `lastEvent.<f>` as the *query* field. **Do NOT** edit the 231 rule files to `events.<last>` — `<last>` is a runtime property and would corrupt the indexed query field.

## 11. Test harness false-green
- The shared normalization harness **deliberately skips raw-log extraction** (grok/json/kv/csv/xml/dynamic) and, for most vendors, does **not assert a rule fires** (fixtures had `"rules": {}`). A green `go test` here is a *schema/normalization* claim only — it does NOT prove the filter parses real logs or the rule detects.
- **Fix:** for any vendor where a rule keys on IP or a renamed field, add at least one fixture that replays a rule's `afterEvents` against a *nil/`"-"`* IP and one that asserts the rule fires.
