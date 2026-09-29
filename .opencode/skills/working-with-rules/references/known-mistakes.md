# Known rule mistakes (do not repeat)

Distilled from the O365 validation campaign (62 rules, ~36 fixed this session) and `docs/rvald26-draft-prs-review.md` (review of 45 filter/rule PRs by another engineer). The recurring theme: **a rule is dead when it references a field or value the ingested event doesn't actually have** — and nothing errors. The single best defense is to read a real sample event before writing/trusting the rule.

## The 7 error classes (in rough order of how often they bit)

### 1. `actionResult` value that never occurs  (dead on arrival)
- AAD **failed** logins are reported as `ResultStatus: Success` → filter maps them to `actionResult=success`, so `where: ...actionResult=failed` on `UserLoginFailed` is dead. Fix: filter force-overrides `UserLoginFailed → actionResult=failed` (append `add` after the generic mapping).
- Ops that emit **no** `ResultStatus` at all → `actionResult` is null. `equals("actionResult","success")` on `TeamDeleted`, `MailboxLogin` is dead. Fix: drop the clause; key on `action` + `log.Workload` only. (This is exactly why 1465 didn't fire until the clause was removed.)
- Capitalized values: ~35 O365 rules used `"Success"`/`"Blocked"` but the filter emits lowercase `success`/`blocked`. Engine is case-sensitive. Fix: lowercase.

### 2. Stale raw field name after a filter rename
- Rule says `log.clientIP` / `log.Operation` / `log.siteUrl` but the filter renamed those to `origin.ip` / `action` / `target.url` (renames delete the source). Dead. Fix: reference the normalized name.

### 3. `contains` on an array / field with the wrong JSON type
- `log.Parameters`, `log.Members`, `Target`, `Folder.Path` are arrays/objects. `contains()` is scalar-only → always false. Fix: filter `cast: {fields:[...], to: string}` (O365 does this for `log.Parameters` + `log.Members`), or a gjson query `log.Parameters.#(Name=X).Value`.

### 4. `{{.field}}` afterEvents placeholder that can be nil
- go-sdk `plugins/rules.go` bails when the placeholder resolves to nil → the whole correlation silently never runs. Correlating on `{{.origin.ip}}` breaks for IP-less events (Kerberos). This is the **rvald26 Issue 1** amplifier: a filter IP-guard that empties `origin.ip` on `"-"`/hostname sources, combined with a rule correlating on `{{.origin.ip}}`, silently kills that rule. Fix: correlate on a guaranteed field (`origin.user`), or the filter must not empty `origin.ip`, or add an `or` branch for absent IP. **Per-vendor check: 135 rules repo-wide depend on `origin.ip`.**

### 5. Invalid `lastEvent.*` / proto field in groupBy or dedup  (rvald26 Issue 2 + 5)
- `lastEvent.system.hostname` — `Event` has **no `system`** → silent no-op groupBy (macOS). The real host field is `dataSource` (set by the **agent**, not the filter — verify the producer exists before declaring a gate dead; that first-pass call was a false positive).
- `Side` has `host`, not `hostname` → `origin.hostname`/`adversary.hostname` writes dropped on proto conversion (VMware, 3 rules dead).
- The **foundation** bug (rvald26 #2590): grouping code read `lastEvent.<f>` from the *wire* alert, which only has `events[]` — so `lastEvent.*` groupBy/dedup was a no-op for **231 rules**. Fixed in the plugin helper (`lastEvent.<f>` → `events[<last>].<f>` for the value), NOT by editing rule files.

### 6. Referencing a field the filter deletes
- netflow filter's `delete` removed `log.bytes`/`log.packets` that 4 rules read → dead (rvald26 #2612). Before adding to a filter `delete`/`drop`, grep every rule for the field.

### 7. Wrong action spelling / a dropped op
- `NewInboxRule` vs the actual `New-InboxRule`; `actionResult` typos; keying on an op that's in the filter's **drop list** (O365 drops `TeamCreated`, `FileUploaded`, `AccessedOdataLink`, …) → the event never reaches the index, rule can't fire. Fix: confirm the op is in the keep-list AND present in the live index.

## Process mistakes (how the above slipped through)
- **Trust "the test passed."** The shared normalization harness **skips raw-log extraction** (grok/json/kv/csv/xml/dynamic) and, for most vendors, has `"rules": {}` (never asserts a rule fires). Green ≠ detects. The Windows false-green came straight from empty `rules`.
- **One-pass review.** The macOS #2622 "block" was a false positive on re-check (the agent, not a filter, produces `dataSource`). Always find the actual producer of a gated field before declaring it dead.
- **Scope creep in a "fix."** #2590 was labeled a test-harness PR but is a fleet-wide dedup activation (231 rules, alert volume shifts). A rule/grouping change that touches `lastEvent.*` or non-scalar groupBy has blast radius — call it out and roll out separately with a before/after alert-volume check.
- **Two streams editing the same files.** #2613 (O365) collided with the in-flight O365 campaign on the same 6 rules + filter. Reconcile into one change set; don't stack.

## The O365 field map (verified Sept 2026) — use to avoid guessing
- `action` ← `log.Operation`; `origin.user` ← `log.UserId`; `origin.ip` ← `log.ClientIP`/`log.ClientIPAddress`; `target.filename` ← `log.ItemName` (file ops) — NOT `log.SourceFileName`; `log.SourceFileExtension` = real last ext.
- `MailItemsAccessed`: has `MailboxOwnerSid`, `LogonType`(int), `ClientInfoString`, `Folders`(array); **no `MailboxOwnerUPN`**.
- Teams `MemberAdded`/`ChatCreated`: `log.Members[].UPN` (cast to string to `contains "#EXT#"`), `log.ParticipantInfo`. `TeamDeleted`: `log.TeamName`, `log.TeamGuid`, no `actionResult`.
- AAD admin ops (keep-list): `Add service principal.`, `Add member to role.`, `Remove member from role.`, `Update device.`.
- Drop-list (never ingested): `TeamCreated`, `FileUploaded`, `AccessedOdataLink`, `ChatRetrieved`, `ChatUpdated`, `MessageDeleted`, `Copy`, `Create`, `Update`, `ViewDocument` (+ ~300 more in the filter's `oneOf("log.Operation",[...])`).
