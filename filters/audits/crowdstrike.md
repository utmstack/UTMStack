# CrowdStrike v11 filter and rule review

The CrowdStrike filter (`filters/crowdstrike/crowdstrike.yml`) damaged quoted command lines and
carried three renames that could never run. Five of the seventeen CrowdStrike rules tested a
field that this integration can never deliver, so they could never fire, and eight rules grouped
or deduplicated their alerts by fields that do not exist on an alert. This change removes the
two damaging trim steps and the three dead renames, drops the impossible test from the five
rules, and moves the eight grouping and deduplication keys to the alert's `adversary` fields.
Everything else is unchanged and listed under "Deferred". The schema is ThreatWinds go-sdk
**v1.1.36**, as pinned by `plugins/alerts/go.mod`.

## Evidence basis

- **Genuine records, reviewed privately.** Of the 31 v11 instances, 29 could be searched and
  one, a pre-release instance, holds CrowdStrike data: the Falcon event stream of one Falcon
  account, retained since 2026-09-21 (about 1,170 records when read on 2026-09-24). It holds
  four event types: `APIActivityAuditEvent` (API calls), `AuditLogV3Event` (a second, nested copy
  of each audit record), `UserActivityAuditEvent` and `AuthActivityAuditEvent`. It holds no
  detection summary, incident, real-time response, custom indicator or process record, so the
  detection half of the filter and eight of the rules have never received real input. An older
  index of the same instance, since removed, held two console sign-in records that the related
  outcome review (draft #2680) read. The records were only read. A private set of
  63 of them was replayed locally; none of them, their addresses, identifiers or names is in
  this repository. Every committed input is fabricated.
- **Vendor documentation was not available.** The field reference for the event stream is in the
  Falcon console documentation, which needs a login on a host outside the allowed list, and the
  public pages that used to describe the SIEM integration now redirect to a general landing page.
  No field name, value, operation name or severity scale could be checked against CrowdStrike.
  Every change below rests on code and records; everything that needs CrowdStrike's definitions
  is deferred.
- **What the corrections rest on.**
  - *the producer*: this repository's `plugins/crowdstrike/main.go` (`processEvent`) forwards
    each event-stream record unchanged as `raw`, so the filter's `json` step only ever sees the
    top-level keys `metadata` and `event`;
  - *the engine*: the public EventProcessor at commit `8a3ade72bd9d12db21f6b273200588fb49540f14`.
    Its [trim plugin](https://github.com/utmstack/EventProcessor/blob/8a3ade72bd9d12db21f6b273200588fb49540f14/plugins/trim/main.go)
    removes surrounding spaces and then one copy of the prefix or suffix;
    its [rename plugin](https://github.com/utmstack/EventProcessor/blob/8a3ade72bd9d12db21f6b273200588fb49540f14/plugins/rename/main.go)
    moves a value and deletes the source;
    its [CEL plugin](https://github.com/utmstack/EventProcessor/blob/8a3ade72bd9d12db21f6b273200588fb49540f14/plugins/cel/main.go)
    (`generateAlert`) builds the alert with the event's `origin` as `adversary` for rules with
    `adversary: origin`, and an alert has no `origin` field;
  - *the alerts plugin*: `plugins/alerts/grouping.go` resolves `groupBy` and `deduplicateBy` keys
    on the alert and skips keys it cannot resolve; `main.go` uses the same keys for grouping and
    for its 7-day duplicate search;
  - *the SDK*: go-sdk v1.1.36 `plugins/plugins.proto` (the `Alert` and `Side` messages) and its
    CEL helpers, which return false for a missing field.

## Filter changes

The version comment on line 1 stays at **1.2.0** on purpose. Draft #2680 changes the same line
to 1.3.0; leaving it alone keeps the two drafts free of conflicts. Whichever draft merges second
should take the next version.

| Change | Why | Basis | Proof |
|---|---|---|---|
| F1: delete the two `trim` steps that remove a double quote from the start and from the end of `log.eventCommandLine` (lines 620-631 before this change). The `[{` and `}]` trims stay. | Each step removes one quote after trimming spaces, so a Windows command line that starts with a quoted program path or ends with a quoted argument lost one quote of a pair: `"C:\Tools\procdump64.exe" -accepteula -ma lsass.exe "C:\Temp\lsass.dmp"` was stored as `C:\Tools\procdump64.exe" -accepteula -ma lsass.exe "C:\Temp\lsass.dmp`. The stored text no longer matched what ran, and three rules group alerts by it. The parent and grandparent command lines were never trimmed. | Engine | Playground: the original filter altered 7 of 8 fabricated text command lines, 5 of them left with an odd number of quotes; the corrected filter keeps all 8 as sent, apart from surrounding spaces, which the two remaining trims still remove. A command line sent as a list is stored exactly as before. 63 of 63 genuine records are identical before and after. |
| F2: delete the second rename of `log.event.Attributes.trace_id`, `log.event.ServiceName` and `log.event.Message` (lines 112-116, 217-221 and 267-271 before this change). | Each repeats an earlier rename with the same source and target. The first one moves the value and nothing in between recreates the source, so the second one never finds anything. | Filter order, engine | Playground: 63 of 63 genuine records identical; the trace ID (26 genuine records), service name (38) and message (30) are unchanged. |

## Rule changes

Names, descriptions, conditions other than the removed test, impact, thresholds, windows, history
searches and MITRE labels are unchanged. None of the 17 rule files carries a version comment, so
no rule version changes.

| Rules | Change | Why | Basis |
|---|---|---|---|
| `inhibit_system_recovery`, `os_credential_dumping_activity`, `suspicious_encoded_powershell_execution`, `suspicious_native_downloaders`, `windows_event_log_clearing` | Delete `equals("log.event_simpleName", "ProcessRollup2") &&` (line 15). | `event_simpleName` is a Falcon Data Replicator field. This integration forwards only the event stream, whose records have the top-level keys `metadata` and `event`, and no filter step writes that field, so the condition could never be true and the five rules never fired. They now fire on a detection summary whose command line matches their pattern. | Producer, filter, genuine records (the field exists in none of them) |
| The three rules above that grouped by `origin.host` and `origin.user` (`inhibit_system_recovery`, `os_credential_dumping_activity`, `windows_event_log_clearing`), and `suspicious_encoded_powershell_execution`, `suspicious_native_downloaders`, `suspicious_downloader_execution_linux_macos`, `security_defenses_impaired_or_policy_disabled` | `groupBy`: `origin.host` becomes `adversary.host`, `origin.user` becomes `adversary.user`. | An alert has no `origin`, so these keys were always skipped. The first three never grouped at all; the other four grouped only by their `lastEvent` key, which merged alerts from different hosts that shared a command line or a disposition text. | Engine, alerts plugin, SDK |
| `multiple_authentication_failures_(possible_brute_force_attack)` | `deduplicateBy`: `origin.ip` becomes `adversary.ip` (line 30). | For the same reason no alert was ever deduplicated: once an address reached five failures in 15 minutes, every further failure raised another alert. Now repeats from the same address within the plugin's 7-day search are dropped. The condition and history search are unchanged. | Engine, alerts plugin, SDK |

A detection summary can now raise the general detection alert and one or more of these specific
alerts at the same time (D11).

## Validation

**Engine and versions.** EventProcessor `8a3ade72bd9d12db21f6b273200588fb49540f14`, the same
source revision as the newest published engine image at the time of testing
(`ghcr.io/utmstack/utmstack/eventprocessor:v11.2.14` on `ghcr.io/utmstack/eventprocessor/base:1.1.7`).
The local binaries were compiled from that revision for darwin/arm64 with go1.25.7; every
parser, writer and CEL plugin links go-sdk v1.1.36, and the geolocation plugin, built from this
repository, links v1.1.34 and read fabricated location data. Rule conditions were also replayed
with go-sdk v1.1.36.

**Filter runs** (playground, file input, no rules): 79 records, the 63 genuine ones and 16
fabricated, each run in a fresh working directory.

| Run | Result |
|---|---|
| Original and corrected filter | 79 of 79 events each, no errors. 63 of 63 genuine records identical. The only differing field is `log.eventCommandLine`, on 7 fabricated detection summaries. |
| Both filters with #2680 applied | The three-way merge of this filter with #2680's has no conflict. 79 of 79 events each, no errors. The two differ only in `log.eventCommandLine` of the same 7 fabricated records; no genuine record differs. |

**Rule runs** (playground with the CEL plugin and the alert writer, and an OpenSearch address
where nothing listens): 116 records, the 63 genuine ones and 53 fabricated positives and near
misses, under the original, the corrected and the corrected-plus-#2680 filter.

| Check | Result |
|---|---|
| Original rules | 42 local alerts under each filter; the five command-line rules raise none. |
| Corrected rules | 56 local alerts under each filter. The five command-line rules alert on exactly the 14 fabricated positives (4 shadow-copy or recovery, 2 credential dumping, 1 encoded PowerShell, 5 downloaders, 2 log clearing), on none of the 6 near misses and on no genuine record. The other 12 rules raise the same alerts as with the original rules. |
| Errors | No condition error, no missing history value and no Circuit Breaker. The only rule errors are brute-force history searches, one per run (three with #2680's filter, which gives two more fabricated sign-in failures an address), each refused because no OpenSearch runs. The history search is reached but not tested here. |
| go-sdk v1.1.36 replay of all 17 rules on every event | Original rules: 0 matches for the five command-line rules. Corrected rules: 4, 2, 1, 5 and 2 matches. No genuine record matches any rule, in the rule runs or in the four filter runs. |
| `grouping.go` on every local alert | The corrected keys resolve on all 65 alerts of the seven regrouped rules. One fabricated record has no user name and groups by host only, as intended. The original keys resolve on none of those alerts for the three host-and-user rules and reduce to the `lastEvent` key for the other four. |
| Brute-force history, go-sdk v1.1.36 `SearchRequest.Execute` against a local mock | The same for the original and the corrected rule: 4 earlier failures give no alert, 5 give an alert, 5 older than 15 minutes or from another address give none; a record without `origin.ip` stops with a placeholder error before any query. Only `adversary.ip` gives a usable deduplication key. |

**Fabricated regression, committed.** `plugins/alerts/testdata/crowdstrike-review/raw.json` holds
18 invented event-stream records: 15 detection summaries with quoted, spaced and list-shaped
command lines (a positive and a near miss for each command-line rule), an API record and two
sign-in failures. Addresses are RFC 5737 documentation addresses, MAC addresses RFC 7042
documentation addresses, names are examples and every customer and client identifier is zero.
The records carry no `UserIp` and no case checks `actionResult`, because both belong to #2680.

**Go tests.** `plugins/alerts/crowdstrike_review_test.go` runs the records through a model of the
engine's step plugins, with every `where` clause evaluated by go-sdk v1.1.36, and checks that
command lines are stored as sent; that no unconditional rename repeats an earlier one; that
each record matches exactly its listed rules and every field a rule reads is one the filter
writes; and that every `groupBy` and `deduplicateBy` key is an alert field that resolves through
`grouping.go` on the alert the CEL plugin would build. All four tests fail on the original filter
and rules. The model gave the same fields as the playground on every record tried: the 18
committed records, and the 116 and 79 private records under each filter (timestamps, tenant,
data source, raw text and geolocation excluded). On the 18 committed records the playground raised
exactly the 26 alerts the test expects, plus one refused brute-force history search. The full
`plugins/alerts` suite passes: 47 tests pass and the same 11 tests that need other
technologies' private evidence skip, as on the base commit (43 pass, 11 skip). With #2680
merged in a separate checkout, 49 pass and the same 11 skip: both CrowdStrike test files
compile together and both of #2680's tests pass with the corrected brute-force rule.

**Static checks.** The repository's contract test over all CrowdStrike files fails 8 of 18 on the
base commit (the rules with `origin.*` keys) and passes 18 of 18 here. The review tooling's
contract check goes from 13 errors to 2; the two left are the `failed` outcome values that
#2680 replaces.

## Relationship with draft #2680

Draft #2680 corrects the outcome values (`actionResult`) and maps the console sign-in address
(`UserIp`) to `origin.ip`. This change touches none of its lines: not the version comment, not
the new rename after the `LocalIP` step, not the outcome block and not the delete list. #2680
changes no rule file. The filter changes merge without conflict, the two test files use
different names, and the combined result is covered by the runs above. With the unchanged filter
a sign-in failure that carries its address only in `UserIp` gets no `origin.ip`, so the
brute-force rule can only start from such records once #2680 is merged; its deduplication key
here is correct either way.

## Deferred

Each item needs CrowdStrike's documentation, records that have not arrived or an owner decision,
and is unchanged here.

| Id | What | What would unblock it |
|---|---|---|
| D1 | `AuditLogV3Event`: every audit record arrives twice, once in the legacy shape and once as a nested V3 copy that no step maps, so the copies get no standard field. Choose one copy per event: drop, pass through or map. | A CrowdStrike statement on retiring the legacy audit events, or records showing they stop; decide with #2680 and D12. |
| D2 | Identities: `UserId` to `origin.user`, `target_name`/`action_target_name` to `target.user`, and the API client identity. | CrowdStrike's field list or real user-management records; ship with the six rules that group by `lastEvent.log.eventUserId`. |
| D3 | Detection summaries: `origin.command`, `severity` and `origin.process`. | CrowdStrike's severity scale and detection field list, or real detections; ship `origin.command` with the six command-line rules. |
| D4 | The `[{` and `}]` command-line trims. | A real detection or documentation showing whether `CommandLine` can be a list of objects. |
| D5 | The `scope(s)` rename never matches (`scopes` in API records, `scope` in V3 copies). | A decision; no rule reads it. |
| D6 | `SourceIp` and `LocalIP` are both renamed unconditionally into `origin.ip`. | A real record with both, or documentation. |
| D7 | The delete of `log.statusCode` does nothing. | Nothing; left beside #2680's delete entry. |
| D8 | `Lin` or `Linux` as the Linux platform value in `suspicious_downloader_execution_linux_macos`. | Documentation or a real Linux detection; the native-downloader rule already covers curl and wget over HTTP on any platform. |
| D9 | The meaning of `PatternDispositionFlags.PolicyDisabled` and of the disposition value list in `security_defenses_impaired_or_policy_disabled`. | CrowdStrike's definitions: a detect-only detection might otherwise be raised as high-impact tampering. |
| D10 | The log-clearing pattern misses `wevtutil.exe cl`. | A real command line, or a small follow-up. |
| D11 | The general detection rule fires on every severity although it says "critical", and overlaps with the specific rules. | CrowdStrike's severity scale or real detections. |
| D12 | The brute-force history counts any CrowdStrike record from the address with `Success=false`, and it would count each failure twice if the V3 copies were ever normalized: three real failures would then reach the threshold of five. No effect today, because the copies stay unparsed. | Change together with D1 and #2680's history test, which checks the current search terms. |
| D13 | MITRE ATT&CK v19: four rules cite revoked techniques (T1070.001 and T1562.001/007) and the removed "Defense Evasion" tactic name, and six labels do not describe their conditions. | One fleet-wide relabel: 66 rule files cite a T1562 sub-technique, 3 cite T1070.001, 132 use "Defense Evasion", and the backend seed data repeats the labels. |
| D14 | Console operation names and event types that the rules expect but no record has shown. | Documentation or real records. |
| D15 | The CrowdStrike plugin skips events created while it is down or reconnecting (offsets are kept in memory only), and `deviceTime` is the receive time. | The plugin owners: persist stream offsets and resume from them. |
| D16 | A failing API client (see the operational note). | Operations. |
| D17 | The filter header cites a third-party page, not CrowdStrike. | A reachable vendor reference. |
| D18 | The log-clearing rule's name and description speak of raw process telemetry, which this integration never delivers. | A text follow-up; keep the alert name unless saved searches are migrated. |
| D19 | `action` for records without an HTTP method; `request_host` and `request_path` left nested. | A naming decision; no rule reads them. |

## Operational note

The audit stream shows a second API client, not the collector of the instance that receives the
data, that has failed every stream-refresh call in the retained data since 2026-09-21 (189 calls,
all answered with HTTP 404) while still obtaining access tokens. That is a configuration fault
at its source, not a filter problem. Its owner should be identified and the client fixed or
revoked. A rule for it would need CrowdStrike's meaning of those answers and grouping by client.

## Known limits

- No detection, incident, real-time response, custom indicator or failed sign-in record has
  reached any instance. The command-line, grouping and deduplication behavior rests on fabricated
  records.
- The playground's alert writer only records alerts. Production indexing, grouping,
  deduplication and notifications were not exercised; the grouping and deduplication keys were
  checked with this plugin's code on the local alerts.
- History searches were not executed against OpenSearch. The brute-force threshold, window and
  terms are unchanged and were checked only against a local mock.
- CrowdStrike's documentation could not be read, so field names and values are those of the
  code and of one customer account's records.
- The Go test models the engine's step plugins and skips the geolocation step. It agreed with
  the playground on every record tried, but only the playground runs the engine.

## Reproduce

From this checkout: `(cd plugins/alerts && go test ./... -count=1)`, or
`go test -run TestCrowdStrikeReview -v .` inside `plugins/alerts` for the four tests above.
The engine runs used private records and are not reproducible from the repository; the 18
committed records can be staged in the EventProcessor playground with this filter and these
rules to repeat the committed part.
