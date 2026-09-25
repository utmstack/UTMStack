# CEL function semantics (from go-sdk@v1.1.31/plugins/cel_overloads.go)

The `where:` clauses in filters and rules are CEL expressions evaluated against a gjson view of the normalized event. Source of truth: `go-sdk .../plugins/cel_overloads.go`. **All are scalar-oriented** — this is the #1 cause of "my rule never fires."

## Type requirements (the trap)

| function | arg type | matches if | returns false when the gjson value is… |
|---|---|---|---|
| `contains(key, sub)` | value must be `String` | substring in the string | an **array** or **object** (e.g. `log.Parameters`, `log.Members`) |
| `containsAll(key, [subs])` | value must be `String` | all substrings present | array/object |
| `startsWith(key, pref)` / `endsWith` | value must be `String` | prefix/suffix match | array/object |
| `regexMatch(key, re)` | value must be `String` | regex match | array/object |
| `equalsIgnoreCase(key, val)` | value must be `String` | case-insensitive eq | array/object/non-string |
| `equals(key, val)` | any scalar | numeric / string / bool / int / float (via `flexibleMatch`) | — |
| `oneOf(key, [vals])` | any scalar | value equals any in list (flexibleMatch) | **null** (not exists → false) |
| `inCIDR(key, cidr)` | value must be `String` (IP) | IP in subnet | non-string |
| `lessThan`/`greaterThan`/`*OrEqual` | numeric-coercible | numeric compare | non-numeric |
| `exists(key)` | any | `gjson.Get(...).Exists()` | **returns TRUE for arrays and objects too** — it only tests presence, not stringness |
| `safe(key, def)` / `safe(key, num)` / `safe(key, bool)` | value must be `String`/numeric/bool else `def` | value else default | array/object → default |
| `isHour`/`isMinute`/`isDayOfWeek`/`isWeekend`/`isWorkDay`/`isBetweenTime` | `String` RFC3339 or `HH:MM` | time logic | non-time → false |

`flexibleMatch` (`equals`/`oneOf`): numeric if both parse as float; else strict string eq if both strings; else bool eq if both bool; else string-representation eq if target is a string and field is non-object/array.

## Consequences (why rules go dead)

1. **`contains` on an array is always false.** O365 `log.Parameters` is `[{Name,Value}]`, `log.Members` is `[...]`. A rule doing `contains("log.Parameters","ForwardTo")` never fires. Two fixes:
   - **Preferred (no rule change per-rule):** add a filter step `cast: {fields:[log.Parameters], to: string}` — `cast` stringifies the array, after which `contains` works. This is what the O365 filter v1.2.x does for both `log.Parameters` and `log.Members`.
   - Or use the gjson query form in the rule: `log.Parameters.#(Name=ForwardTo).Value` (bare `Name=`, no inner quotes, resolves to a String).
2. **`exists` ≠ "is a usable string."** `exists("log.Members")` is true, but `contains("log.Members","#EXT#")` is false unless it's been cast. Pair `exists` (presence) with a `cast` (usability) in the filter.
3. **`actionResult` is not raw — it's synthesized.** The filter `add`-s it from `log.ResultStatus`. `equals("actionResult","success")` only matches events whose `ResultStatus` mapped to success. Check the actual events: AAD **failed** logins carry `ResultStatus: Success` (so they'd wrongly map to success unless overridden); `TeamDeleted`/`MailboxLogin` carry **no** `ResultStatus` at all (so `actionResult` is null and `equals(actionResult,"success")` is false — drop that clause for those ops).
4. **`oneOf` on a null field is false.** If the field isn't populated for that event, no match.
5. **`rename` is lossy** and order-dependent. `rename: {from: [log.Operation], to: action}` deletes `log.Operation`. After it, `where` must use `action`; before it, use `log.Operation`.

## O365 normalized vocabulary (verified against live index, Sept 2026)
- Workload field: `log.Workload` (`Exchange`, `MicrosoftTeams`, `OneDrive`, `SharePoint`, `AzureActiveDirectory`, `MicrosoftDefenderForCloudApps`…).
- `action` ← `log.Operation`. Common ops (keep-list, i.e. NOT dropped): `UserLoggedIn`, `UserLoginFailed`, `Add service principal.`, `Add member to role.`, `Remove member from role.`, `MailboxLogin`, `MailItemsAccessed`, `TeamDeleted`, `MemberAdded`, `ChatCreated`, `MessageSent`, `FileAccessed`, `FileDownloaded`, `FilePreviewed`, `FileShared`, `FileRecycled`, `New-InboxRule`, `Set-InboxRule`, `Set-Mailbox`, `New-TransportRule`.
- Dropped (never ingested — do NOT key on these): `TeamCreated`, `FileUploaded`, `AccessedOdataLink`, `ChatRetrieved`, `ChatUpdated`, `MessageDeleted`, `Copy`, `Create`, `Update`, `ViewDocument`, and ~300 more. (Full list: the `oneOf("log.Operation", [...])` drop step in the filter — 323 entries.)
- File ops carry `target.filename` (full name) + `log.SourceFileExtension` (real last ext) — NOT `log.SourceFileName`.
- `MailItemsAccessed` carries `MailboxOwnerSid`, `LogonType` (int), `ClientInfoString`, `Folders` (array) — **no `MailboxOwnerUPN`** (so owner-vs-accessor comparisons need the owner SID/UPN from elsewhere).
- Teams ops: `TeamDeleted` has `log.TeamName`, `log.TeamGuid`, `log.AADGroupId`; `MemberAdded`/`ChatCreated` carry `log.Members[].UPN` (cast to string to `contains "#EXT#"`) and `log.ParticipantInfo`.
