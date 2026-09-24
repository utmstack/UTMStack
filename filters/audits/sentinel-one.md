# SentinelOne v11 filter and rule review

The SentinelOne filter did not parse the management-console records observed in the field.
Every record lost the CEF event name, values containing spaces kept only their first word,
header text was glued into field names, the CEF time was cut to a fragment and no standard
field was set. Twelve of the 19 rules require the event name, so none of them could match a
real record. This revision parses the CEF header by position, keeps whole
values, converts the CEF time, maps the console actor to `origin.user` for console events,
and changes 14 rules so they consume what the filter now produces without firing on
ordinary console administration. The schema is ThreatWinds go-sdk **v1.1.36**, as pinned
by `plugins/alerts/go.mod` since official `v11` (`d2479c1a`) was merged into this branch.
The review itself used v1.1.33, whose `plugins.proto` is identical. The draft was checked
again on the latest versions; see [Re-validation on the latest versions](#re-validation-on-the-latest-versions).

## Evidence basis

- **Real records, described without identifying data.** Only five genuine SentinelOne
  records were found on the 28 v11 instances whose indices could be searched (three could
  not be searched). All five are on one instance, from one console build and one day:
  CEF `SentinelOne|Mgmt` records for a console user added, a console user deleted and a
  role assigned (activity types 23, 25 and 37). All five are benign administration. They
  were examined and replayed privately and are not reproduced here. The same data type also
  held hand-typed test lines without CEF and synthetic load records, used only as controls.
  No SentinelOne threat, mitigation, Deep Visibility or STAR record was found on any
  searchable instance.
- **What those records showed.** The header is `CEF:0|SentinelOne|Mgmt|<console
  build>|<signature id>|<event name>|<severity>|`. The version slot holds a console build
  string such as `S-<major>.<minor>.<patch>#<build>`, never an IPv4 address or a
  "word number", so none of the three header parsers that captured the event name matched.
  Only the general parser matched, and it left the signature id, event name and severity in
  the text given to `kv`. The signature id always equals the `activityType` extension key.
  `rt` is `#arcsightDate(<Day>, <DD> <Mon> <YYYY>, <hh:mm:ss> UTC)` and appears after
  other keys. In these records `suser` is the console account that made the change and
  `cat` is `SystemEvent`.
- **Engine behaviour.** The public EventProcessor at commit
  `497bf53dbd1ae096f7b2dbc7bce77a6bf9f22ce1`: the
  [grok plugin](https://github.com/utmstack/EventProcessor/blob/497bf53dbd1ae096f7b2dbc7bce77a6bf9f22ce1/plugins/grok/main.go)
  matches each pattern only at the start of the remaining text and writes nothing unless
  every pattern matches; the
  [kv plugin](https://github.com/utmstack/EventProcessor/blob/497bf53dbd1ae096f7b2dbc7bce77a6bf9f22ce1/plugins/kv/main.go)
  splits on every space, keeps the first word of a value and fails when its source is
  missing; the
  [reformat plugin](https://github.com/utmstack/EventProcessor/blob/497bf53dbd1ae096f7b2dbc7bce77a6bf9f22ce1/plugins/reformat/main.go)
  records an error when a time does not parse. The latest EventProcessor, `main` at
  `8a3ade72bd9d12db21f6b273200588fb49540f14`, changes only these plugins' go-sdk version
  (to v1.1.36), so this behaviour is the same there.
- **Schema and semantics.** [Event and Side fields](https://github.com/threatwinds/go-sdk/blob/v1.1.36/plugins/plugins.proto)
  (`origin.user`, `target.host`, `deviceTime`, `severity` are strings) and the
  [standard field meanings](https://github.com/threatwinds/go-sdk/wiki/Standard-Event-Schema)
  (`deviceTime` is the source's original time; origin is the actor).
- **Vendor documentation.** SentinelOne's syslog/CEF reference could not be read from the
  allowed documentation hosts, so the header layout, extension key catalog, activity codes
  and CEF severity scale are **not** documented here. Two public SentinelOne pages were
  used: the [Static AI engine case study](https://www.sentinelone.com/blog/decrypting-sentinelones-detection-an-in-depth-look-at-our-real-time-cwpp-static-ai-engine/)
  (section "Case Study": there are two confidence levels, `SUSPICIOUS` and `MALICIOUS`,
  and the agent policy per confidence level is Detect or Protect) together with the
  [threat intelligence engine example](https://www.sentinelone.com/blog/decrypting-sentinelone-cloud-detection-the-threat-intelligence-engine-in-real-time-cwpp/)
  (section "Example: Shellshock Detection": `MALICIOUS` is the highest level), and the
  [rollback demo](https://www.sentinelone.com/blog/ransomware-mitigation-sentinelones-rollback-demo-rsac-2018/)
  (section "Demo": policy written as a pair such as `Detect/Detect` or `Protect/Detect`).

## Filter changes (version 3.0.1)

| Change | Why | Proof |
|---|---|---|
| Parse the CEF header by position after `CEF:`, whatever syslog header precedes it. Vendor, product, version and signature id go to `log.cefDeviceVendor`, `log.cefDeviceProduct`, `log.cefDeviceVersion`, `log.cefSignatureId`; the name to `log.eventDescription`, which rules already read; the header severity to `log.cefSeverity`; only the extension to `log.restData`. | The event name was lost on every real record and header text reached `kv`. The standard `severity` is **not** set: the CEF severity scale is not documented. | Real records (5 of 5) and playground-validated for the recorded build |
| Remove the two header parsers that stored the version slot as `log.syslogHost`. | The slot is the console build, not a host; real records never matched these parsers. Their only consumers were rules, changed below. | Real records; playground |
| Run `kv` only when `log.restData` exists. | Lines without CEF recorded a `source does not exists` error carrying the serialized pipeline. | Playground (zero errors afterwards) |
| Take the whole `rt=#arcsightDate(...)` value into `log.ruleTime`; convert it to RFC 3339 `deviceTime` only when it has exactly the observed UTC layout. | `log.ruleTime` held `#arcsightDate(Mon,`; `deviceTime` was the ingestion time. Other formats are left alone and keep the default `deviceTime`. | Real records (5 of 5 converted); playground |
| Re-extract `accountName`, `eventDesc`, `suser`, `duser`, `endpointDeviceControlDeviceName`, `sourceGroupName`, `sourceIpAddresses`, `sourceMacAddresses` and `siteName` up to the next key or to the end, then strip the next key. Output names are unchanged. | `siteName` kept its first word; the old two-step extraction lost any of these values when its key was last. | Real records (`siteName`, `suser`, `accountName`); fabricated last-key and all-keys lines in the playground |
| Copy `log.sourceUser` to `origin.user` only when `cat` is `SystemEvent` and the value is not empty or a placeholder. `log.sourceUser` is kept. | In console events `suser` is the account that made the change. Its meaning in other event classes is not established, so they are unchanged. | Real records (5 of 5); playground |

The `actionResult` steps are unchanged; a separate correction owns them. No rule reads
`actionResult`.

## Rule changes

Names, ids, thresholds, windows, impact, adversary side and MITRE labels are unchanged.

| Rule | Change | Why |
|---|---|---|
| `memory_injection_detection` | Remove the branch reading `log.eventDescToParse`; group by `target.host` and `adversary.user` instead of `adversary.host`. | The filter deletes that scratch field before rules run; event `origin.host` is never set. |
| `behavioral_threat_detection`, `custom_detection_rule_triggers`, `deep_visibility_threat_indicators`, `endpoint_detection_response_alerts`, `suspicious_process_tree` | Require `target.host` instead of `log.syslogHost`; group by `target.host`. | `log.syslogHost` was the console build and is no longer produced. The endpoint is the affected side of a detection. |
| `kernel_level_threat`, `rollback_operation_patterns`, `threat_intelligence_matches` | Group by `target.host` instead of `log.syslogHost`. | Same producer change. |
| `kernel_level_threat` and the reputation branch of `threat_intelligence_matches` | Also require `target.host`. | With the event name parsed, console text such as a new user named "Kernel Team" or "Reputation Desk" would fire them. |
| `s1_policy_downgrade` | Fire only on Protect to Detect: the word `downgrade` (any form), "from Protect ... to Detect", "Protect to Detect", or the vendor's paired wording where either mode goes from Protect to Detect (for example `Protect/Detect` to `Detect/Detect`). `policy` is matched as a whole word, in any case. The unchanged activity-name branch now also needs this direction. | The old condition also fired on upgrades (Detect to Protect) and on a role named "Protect and Detect Reviewers". |
| `s1_exclusion_abuse`, `s1_policy_downgrade` | Group by `adversary.user` instead of `adversary.host`. | These are console changes; the console account now reaches `origin.user`. |
| `agent_tampering_attempts`, `threat_mitigation_failures`, `iot_device_compromise_indicators` | Match the word lists as whole words (with the usual inflections). Case is unchanged. | "nonstop" contains "stop", "agentur" contains "agent", "failover" contains "fail", "analytics" contains "ics". |
| `threat_intelligence_matches` | Replace `greaterOrEqual("log.confidencelevel", 90)` with `equalsIgnoreCase("log.confidencelevel", "MALICIOUS")`; the key name is unchanged. | SentinelOne documents two confidence levels, `SUSPICIOUS` and `MALICIOUS`, not a 0-100 score. The syslog spelling of the key is not established. |

No SentinelOne rule has a history query, so no history search is affected. Filter and
rules must ship together: with the new filter and the old rules, benign console text fires
seven old rules (measured below).

## Re-validation on the latest versions

On 2026-09-24 official `v11` moved to `d2479c1a3705eec6a00016689c2bf5fbcc1814f2`, whose
`plugins/alerts` pins go-sdk v1.1.36, and EventProcessor `main` moved to
`8a3ade72bd9d12db21f6b273200588fb49540f14`, whose playground and parser, writer and CEL
plugins all link go-sdk v1.1.36. `v11` was merged into this branch. No file overlaps this
draft, so nothing conflicted.

The newest published engine image, `ghcr.io/utmstack/utmstack/eventprocessor:v11.2.14`
(built 2026-09-24 19:13 UTC on base image `eventprocessor/base:1.1.7`), embeds Go build
information showing that its playground and plugin binaries come from the same
EventProcessor revision `8a3ade7` with go-sdk v1.1.36, built with go1.26.8 for linux/amd64.
The local build used here is that source revision compiled natively for darwin/arm64 with
go1.25.7; only the Go toolchain and platform differ.

What changed in the SDK, and what it means here:

- Since v1.1.35, `utils.SanitizeField` keeps `_` in the field names that the `json`
  (top-level keys), `kv`, `grok`, `csv`, `xml`, `add` and `rename` plugins write. Other
  characters are still removed. This filter only writes names made of letters and digits,
  and no `kv` key in the 52 fabricated lines or the five genuine records contains `_`. So
  every stored name stays the same.
- v1.1.36 makes `regexMatch` match string values only again. Since v1.1.34, `contains`,
  `containsAll`, `startsWith` and `endsWith` also search the JSON text of objects and lists.
  Every such call in this filter and its rules reads a text field, so no result changes.
  `plugins.proto`, `plugins/cel.go` and `plugins/rules.go` are identical in v1.1.33 and v1.1.36.
- No filter, rule or fixture needed a change.

| Check on the latest versions | Result |
|---|---|
| Full `plugins/alerts` suite, go-sdk v1.1.36 | 48 tests pass, 11 skip, none fail (2,342 passing results with subtests). The five SentinelOne tests pass. The skipped tests need other technologies' private evidence and skip on the base commit too. |
| `replay.py` on EventProcessor 8a3ade7 | 52 events, zero parser errors, every key set and value as in `expected.json`, 10 alerts, each from its intended rule. With `--endpoint-harness`: 59 events and 17 alerts. |
| go-sdk v1.1.36 rule replay | All 19 rules over those 52 and 59 events: no compile or evaluation error. The matches are exactly the 10 and 17 playground alerts, and exactly the alerts in `expected.json`. |
| The four private playground runs described below, same 64 and 71 inputs | Same results as before. Original filter and rules: 4 alerts. Corrected filter with the original rules: 22. Corrected filter and rules: 12, none on a genuine record. With the test-only step: 19. All 4,304 assertions pass. |

At 8a3ade7 the CEL plugin reads its OpenSearch address from separate `host`, `port`, `user`
and `password` settings. `replay.py` still gives one URL, so the client gets an empty
address. No SentinelOne rule has a history search, so no result depends on it.

## Validation

These are the original review's results, on EventProcessor `497bf53` and go-sdk v1.1.33.
The section above repeats them on the latest versions.

**Fabricated regression, committed.** `plugins/alerts/testdata/sentinel-one/` holds 52
invented raw lines (`raw.json`), their expected fields and alerts (`expected.json`), the
eleven shared grok definitions this filter uses (`patterns.yaml`, copied from the
repository changelog `20250616001_insert_utm_regex_pattern.xml`) and `replay.py`. The lines
reuse the observed record shape with made-up names, `example.com` addresses, RFC 5737
addresses and invented ids.

**Playground.** A clean build of the EventProcessor commit above ran the original and the
corrected filter and rules on the same 64 inputs: the 52 fabricated lines, two more
fabricated console lines for the deferred false positives, and, privately, the five
genuine records and five non-SentinelOne control records. Its parser and writer plugins
link go-sdk v1.1.26 and its CEL plugin v1.1.34.

| Run | Result |
|---|---|
| Original filter and rules | 64 events, 4 alerts. Each genuine record lost its event name, kept one word of its site name and a fragment of `rt`, carried header text as field names and kept the ingestion time as `deviceTime`. The IPv4 and "word number" version slots became `log.syslogHost`, and one such line fired the behavioral rule on that value. Six lines without a recognised header each carried a `kv` error. |
| Corrected filter, original rules | Normalized events identical to the corrected run; 22 alerts. Benign console text fired seven rules, and the policy rule also fired on both upgrades. |
| Corrected filter and rules | 64 events, zero parser errors, exact field sets and values on every fabricated and genuine record, no alert on any genuine record. 12 alerts: the ten fabricated positives and the two deferred false positives, each exactly once. |
| Same, with the test-only `target.host` step | 71 events, 19 alerts: each of the seven endpoint-gated rules alerted exactly once, on its marked copy only, with that host as the alert target. |

All 4,304 assertions of these four runs passed. The committed `replay.py` repeats the last
two runs with the fabricated lines only: 52 events and 10 alerts, then 59 events and 17
alerts with `--endpoint-harness`, zero parser errors and every alert from its intended rule.

**SDK predicate checks.** The go-sdk v1.1.33 replay evaluated the 14 changed rules, in
original and corrected form, and nine single-branch probes over 94 documents: the corrected
run's 64 events, the seven marked copies and the 23 documents previously stored for this
data type. Each corrected rule matched exactly its intended fabricated positives and no
genuine or stored document; each original rule matched exactly what it alerted on in the
playground; the removed memory-injection branch matched nothing; no corrected event carries
`log.syslogHost` or `severity`; `origin.user` is set on exactly the 36 console events that
name an actor. All 68 checks passed.

**Go tests.** `sentinel_one_filter_test.go` checks the filter structure, compiles every grok,
trim and `regexMatch` pattern, runs the header, value and `rt` patterns over the 52 fabricated
lines with a model of the grok step, checks rule grouping paths, and evaluates 41 synthetic
normalized cases against the shipped rules with go-sdk v1.1.33. Against the original filter
and rules, four of its five tests fail. The full `plugins/alerts` suite passes; the same
private-evidence tests of other technologies skip as they do on the base commit.

## Deferred

These need SentinelOne's syslog/CEF reference or real threat records, and are unchanged:

- Activity-code name lists in five rules (`policy_updated`, `exclusion_created`,
  `agent_uninstall`, `rollback`, `mitigation`). SentinelOne sends numeric codes; only 23,
  25 and 37 are known. Until the codes are documented, a console user named "Exception
  Queue" still fires `s1_exclusion_abuse` and a role named "Rollback Approvers" still
  fires `rollback_operation_patterns` (both reproduced with fabricated lines).
- `duser` to `target.user`, the endpoint key to `target.host`, and any address mapping.
  `endpointDeviceControlDeviceName` most likely names a USB or Bluetooth device, not the
  endpoint. Until an endpoint key is mapped, the rules that require `target.host` cannot
  fire; before this change they required a field real records never had.
- The standard `severity` (CEF scale not documented) and `action` (activity catalog not
  available).
- Threat extension keys and value formats (`threatName`, `engines`, message key `msg` or
  `message`, `confidencelevel` spelling), and case: most rule word lists still match only
  lower-case wording.
- Retiring or redefining the IoT and Storyline rules, whose premises the vendor pages do
  not support.
- MITRE relabelling: ATT&CK v19 revoked T1562 and T1562.001 in favour of T1685; this is a
  repository-wide change.
- New rules for console users added and roles assigned (the only SentinelOne activity
  observed so far).
- The other Stage 3 grouping drafts (`agent_tampering_attempts`, `container_security_alerts`,
  `storyline_correlation`) and the `log.syslogHost` mentions left in four rule descriptions.

## Known limits

- No threat record was available, so threat-event parsing, keys and rule wording are
  verified only with invented lines. The syslog wording of a policy-mode change is not
  documented either; the direction test covers the wordings listed above.
- The latest check used EventProcessor `8a3ade7`, whose playground and plugins link go-sdk
  v1.1.36, the version the alerts module now pins; predicates were also checked with
  v1.1.36. The original review used `497bf53` (parser and writer plugins v1.1.26, CEL plugin
  v1.1.34) and v1.1.33 predicates. Neither build is asserted to match a customer deployment.
  `reformat` is already used by the ESET and Sophos XG filters.
- The playground `saw` writer only records alerts. Grouping, deduplication, indexing,
  notifications and production alerts were not tested.
- A header missing one of the seven CEF fields, or with an empty extension, falls back to
  the old header parsers. `kv` still mishandles quoted values and an escaped `=` inside a
  value (it can manufacture a key); only the nine re-extracted keys are kept whole.
- A separate `actionResult` correction edits the same file. This change keeps that block
  byte-identical; combining the two needs a rebase at the version comment line.

## Reproduce

Build EventProcessor `8a3ade72bd9d12db21f6b273200588fb49540f14` (the latest check) without
changing its dependencies. With `EP` set to that checkout's absolute path:

```sh
mkdir -p "$EP/test-bin" "$EP/test-plugins"
(cd "$EP" && go build -mod=readonly -o "$EP/test-bin/playground" ./cmd/playground)
for plugin in add cel delete grok kv reformat rename saw sew trim; do
  (cd "$EP/plugins/$plugin" && go build -mod=readonly -o "$EP/test-plugins/$plugin.plugin" .)
done
```

From this UTMStack checkout, with PyYAML installed:

```sh
python3 plugins/alerts/testdata/sentinel-one/replay.py \
  --playground "$EP/test-bin/playground" --plugins "$EP/test-plugins"
python3 plugins/alerts/testdata/sentinel-one/replay.py --endpoint-harness \
  --playground "$EP/test-bin/playground" --plugins "$EP/test-plugins"
(cd plugins/alerts && go test ./... -count=1)
```

Each playground run takes about three minutes. The Go suite alone does not execute raw
extraction.
