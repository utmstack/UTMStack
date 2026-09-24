# Cisco ASA v11 filter and rule review

The Cisco ASA filter (`filters/cisco/asa.yml`) had defects that its own patterns, the event
engine and the SDK make visible without any Cisco record. A line that no header pattern
accepted was stored with 519 CEL errors, geolocation replaced every non-private address in
16 `log.*` fields with an object, permitted hits of 106102/106103 were stored as denied, and
steps for 302013, 302304, 305011/305012, 302017 and 113009/113011 never matched or damaged
their own output. On the rule side, the
intrusion prevention rule disabled itself after five 108003 events, and the VPN rule read a
field nothing writes. This revision fixes those points and nothing else. The schema is
ThreatWinds go-sdk **v1.1.36**, as pinned by `plugins/alerts/go.mod` since official `v11`
(`d2479c1a`) was merged into this branch. The review itself used v1.1.33, whose
`plugins.proto`, `plugins/cel.go` and `plugins/rules.go` are identical. The draft was checked
again on the latest versions; see [Re-validation on the latest versions](#re-validation-on-the-latest-versions).

## Evidence basis

- **No Cisco ASA records.** None of the 29 v11 instances that could be searched holds a
  `firewall-cisco-asa` record in any retained index; two instances could not be searched. A
  full-text search of the v11 log indices for the words `asa` and `ftd` found no ASA syslog
  under another data type either. Cisco Firepower Threat Defense records stored under
  another data type use a different message catalog and were not used here.
- **Cisco documentation was not available.** Cisco's documentation site refused every
  automated request (HTTP 403) from the allowed fetch tool, so no fixture is taken from
  Cisco's documentation, and nothing that depends on Cisco's wording, message layouts, which
  address is the source, or what a result means is changed.
- **What the corrections rest on.** Each change below names its basis:
  - *the filter's own patterns*: what a step can receive follows from the earlier steps of
    this filter;
  - *the engine*: the public EventProcessor at commit
    `497bf53dbd1ae096f7b2dbc7bce77a6bf9f22ce1`. Its
    [parser](https://github.com/utmstack/EventProcessor/blob/497bf53dbd1ae096f7b2dbc7bce77a6bf9f22ce1/pkg/parsing/parsing.go)
    stores an error and skips the step when a `where` clause fails; its
    [grok plugin](https://github.com/utmstack/EventProcessor/blob/497bf53dbd1ae096f7b2dbc7bce77a6bf9f22ce1/plugins/grok/main.go)
    writes nothing unless every pattern matched, and stops when the text runs out; its
    [CEL plugin](https://github.com/utmstack/EventProcessor/blob/497bf53dbd1ae096f7b2dbc7bce77a6bf9f22ce1/plugins/cel/main.go)
    counts rule errors and, at the fifth, disables the rule and raises a
    `Circuit Breaker: <rule name>` alert. The latest EventProcessor, `main` at
    `8a3ade72bd9d12db21f6b273200588fb49540f14`, changes only these plugins' go-sdk version
    (to v1.1.36) and how the CEL plugin reads its OpenSearch address, so this behaviour is the
    same there. This repository's `plugins/geolocation/main.go`
    writes its result at the destination path with `sjson.Set`, which turns a string at that
    path, or above it, into an object;
  - *the SDK*: go-sdk v1.1.33 (both files below are identical in v1.1.36)
    [`plugins/cel.go`](https://github.com/threatwinds/go-sdk/blob/v1.1.33/plugins/cel.go)
    declares a CEL variable only for the top-level keys an event has, so `log.messageId==N`
    fails to compile when there is no `log` object, while `equals`, `greaterOrEqual` and
    `lessOrEqual` return false for a missing path;
    [`plugins/rules.go`](https://github.com/threatwinds/go-sdk/blob/v1.1.33/plugins/rules.go)
    returns an error when a `{{.field}}` history placeholder cannot be resolved.

## Filter changes (version 3.1.0)

| Change | Why | Basis | Proof |
|---|---|---|---|
| Rewrite the 519 `where` clauses that compared `log.messageId` or `log.severity` directly (for example `log.messageId==106001`) with `equals`, `greaterOrEqual` and `lessOrEqual`, and write the step added below the same way (520 clauses). | A line that no header pattern accepts has no `log` object, so every one of those clauses failed to compile, its step was skipped and an error was stored: 519 errors, about 145 KB, per event. | Engine and SDK | Playground: 12 such lines went from 519 errors each to none; every parsed line is unchanged. Go test: the 524 clauses that use these helpers on these fields (the 520 rewritten and 4 that already did) give the same result as the direct comparison on 2,510 events that have a `log` object, and false without an error when there is none. Earlier measurement: +0.3 ms per parsed event. |
| Write the 16 `log.*` geolocation results to `log.<field>Geolocation` (for example `log.localIpGeolocation`) instead of `log.<field>.geolocation`. `origin.geolocation` and `target.geolocation` are unchanged. | The result replaced the address with an object for every non-private address, so one field was text in some events and an object in others. In a local OpenSearch 2.13.0 test with default dynamic mapping, the old output lost one of two documents in each of three scenarios; the new output stored 4 of 4 in both orders. | Engine (this repository's geolocation plugin) | Playground: 42 address fields replaced before, none after; the sibling keys hold the geolocation. Go test: no geolocation destination lies under a field that holds a value. |
| 106102/106103: run the `denied` add before the `accepted` add. The values are unchanged. | The second add's condition was also true on the `accepted` the first add had just written, so every permitted hit was stored as `denied`. | The filter's own conditions, evaluated with the SDK | Playground: 3 permitted lines change from `denied` to `accepted`; denied lines stay `denied`. |
| 302013: drop the direction grok's last pattern. | Its source holds only `Built <direction>` (or `Built <direction> Probe`), so unless the message said `Probe` the last pattern met empty text and nothing was written. | Filter patterns and engine | `log.direction` is set on every 302013 line; the `Probe` form is unchanged. |
| 302304: the protocol grok reads `Teardown`. | The only step that produces `action` for 302304 starts it with `Teardown`; the grok expected `Built` and never matched. 302306 already does this. | Filter patterns | `protocol` is `TCP` on 302304 lines. |
| 305011/305012: the protocol grok writes its first match to `log.irrelevant` instead of `action`. | It overwrote the full phrase the main grok had captured (`Built dynamic TCP translation`) with its first words (`Built dynamic`). | Filter patterns | `action` keeps the full phrase. |
| 302017: drop the firewall-user grok's last pattern and remove the leading `(` from `target.user`. | The grok's last pattern always met empty text, so `log.firewallUserTo` was never set; the existing trims removed `)` and a trailing `(`, leaving `(name`. | Filter patterns and engine | `target.user` is `erin` and `log.firewallUserTo` is `dave` on the fabricated line. Which side that user belongs to is unchanged. |
| 113009/113011: run the second grok variant only when the first set no user (`&& !exists("origin.user")`). | Every text the first variant matches also matches the second, which then stored `origin.user` as `= alice` and put the parentheses back into `log.policy`. | Filter patterns | `origin.user` is `alice` and `log.policy` is `DfltGrpPolicy`; the form without `=` still works. |

The `actionResult` values are unchanged: a separate correction owns them (see Deferred).
`filters/cisco/firepower.yml` has the same 16 geolocation steps and belongs to its own review.

## Rule changes

Names, thresholds, windows, impact, adversary side, grouping and MITRE labels are unchanged.
`botnet_traffic_detection` is unchanged.

| Rule | Change | Why | Basis |
|---|---|---|---|
| `ips_signature_matches` (v1.0.1) | Wrap the condition in `exists("origin.ip") && (...)`; remove the `log.action` branch. The `log.message` text branches stay as they are. | The only branch that can match today is message 108003, which this filter does not parse, so the event has no `origin.ip`. The history placeholder `{{.origin.ip}}` then failed on every such event, and after five the rule was disabled with a Circuit Breaker alert. No step writes `log.action`. | SDK, engine and filter |
| `multiple_failed_vpn_attempts` (v1.0.1) | Read `log.msg` instead of `log.message`. | The header patterns write the message body to `log.msg`; nothing writes `log.message`. This has no effect until 113015 puts its source address in `origin.ip` (Deferred D01). | Filter |

## Re-validation on the latest versions

On 2026-09-24 official `v11` moved to `d2479c1a3705eec6a00016689c2bf5fbcc1814f2`, whose
`plugins/alerts` pins go-sdk v1.1.36, and EventProcessor `main` moved to
`8a3ade72bd9d12db21f6b273200588fb49540f14`, whose playground and parser, writer and CEL
plugins all link go-sdk v1.1.36. `v11` was merged into this branch. No file overlaps this
draft, so nothing conflicted. The geolocation plugin was built from `v11` `d2479c1a`, which
this branch now carries unchanged; its own `go.mod` pins go-sdk v1.1.34.

What changed in the SDK, and what it means here:

- Since v1.1.35, `utils.SanitizeField` keeps `_` in the field names that the `json`
  (top-level keys), `kv`, `grok`, `csv`, `xml`, `add` and `rename` plugins write. Other
  characters are still removed. This filter has no `json`, `kv` or `csv` step, and none of
  the 129 names it writes contains `_` or another removed character. So every stored name
  stays the same; the per-line comparison below confirms it.
- v1.1.36 makes `regexMatch` match string values only again. Since v1.1.34, `contains`,
  `containsAll`, `startsWith` and `endsWith` also search the JSON text of objects and lists.
  Every such call in this filter and its rules reads a text field (`log.message`, `log.msg`,
  `log.reason`, `action`), so no result changes. `plugins.proto`, `plugins/cel.go` and
  `plugins/rules.go` are identical in v1.1.33 and v1.1.36.
- No filter, rule or fixture needed a change.

| Check on the latest versions | Result |
|---|---|
| Full `plugins/alerts` suite, go-sdk v1.1.36 | 50 tests pass, 11 skip, none fail (2,266 passing results with subtests). The seven Cisco ASA tests pass; the Go model of the step plugins uses the SDK's own `SanitizeField`, so it follows v1.1.36. The skipped tests need other technologies' private evidence and skip on the base commit too. |
| `replay.py` on EventProcessor 8a3ade7 | 41 events without errors, every stored field as in `expected.json`. Two alerts, both from the botnet rule, on 338001 and 338002; no Circuit Breaker alert and no history search attempted. |
| The 80 private fabricated lines, original and corrected filter | Original: 80 events, 6,228 errors (519 on each of 12 lines), 42 address fields replaced by an object. Corrected: 80 events, no errors, none replaced. 84 of 84 planned field assertions pass for both. Every line's stored fields, types and error counts equal the original review's results, for both filters. |
| Corrected filter and all three rules, the same 12 lines | 12 events without errors; two botnet alerts, on 338001 and 338002; no intrusion prevention, VPN or Circuit Breaker alert; no compile, rule or history search error. 24 of 24 checks pass. |
| Original filter and rules, the same 12 lines | Six failed history searches (`expression value cannot be nil after placeholder resolution`) and one `Circuit Breaker` alert for the intrusion prevention rule, as in the original review. |
| go-sdk v1.1.36 rule replay | The same predicate checks as the original review, now with the latest playground output: 90 of 90 and 65 of 65 pass. |

At 8a3ade7 the CEL plugin reads its OpenSearch address from separate `host`, `port`, `user`
and `password` settings. `replay.py` still gives one URL, so the client gets an empty host
and connects to port 443 on the test computer, where nothing listened. Any history search
therefore still fails and is reported; none was attempted with the corrected rules.

## Validation

These are the original review's results, on EventProcessor `497bf53` and go-sdk v1.1.33.
The section above repeats them on the latest versions.

**Fabricated regression, committed.** `plugins/alerts/testdata/cisco-asa/` holds 41 invented
raw lines (`raw.json`), their expected fields and alerts (`expected.json`), the 13 shared
grok definitions this filter uses (`patterns.yaml`, copied from the repository changelog
`20250616001_insert_utm_regex_pattern.xml`), invented geolocation data (`geolocation-data/`)
and `replay.py`. The lines follow the filter's own patterns; they are not claimed to be Cisco's
format. Addresses are RFC 5737 and RFC 3849 documentation addresses, and device and user
names are examples. The geolocation data covers two IPv4 documentation ranges and the IPv6
documentation range with documentation AS numbers and invented places; the third IPv4 range
is left out on purpose.

**Playground.** A clean build of the EventProcessor commit above ran the original and the
corrected filter on the same 80 fabricated lines, a larger private set from which the
committed lines were derived.
Its parser and writer plugins link go-sdk v1.1.26 and its CEL plugin v1.1.34; the
geolocation plugin was built from this repository.

| Run | Result |
|---|---|
| Original filter, 80 lines | 80 events. The 12 lines no header pattern accepts carry 519 errors each (6,228 in all); 42 address fields hold an object instead of the address. |
| Corrected filter, 80 lines | 80 events, no errors. 84 of 84 planned field assertions pass for both runs; 36 events are identical and every other difference is one of the changes above. |
| Corrected filter and all three rules, 12 lines (338001, 338002, six 108003, two 113015, one 302013, one non-ASA line) | 12 events without errors. Two botnet alerts, on 338001 and 338002, without addresses; no intrusion prevention, VPN or Circuit Breaker alert; no compile, rule or history search error. No OpenSearch was running, and no line reached a history search. |
| Committed `replay.py`, 41 lines | 41 events without errors, every stored field as recorded in `expected.json`. Two alerts, both from the botnet rule, on 338001 and 338002; no Circuit Breaker alert and no history search attempted. |

With the original filter and rules and the same 12 lines, an earlier run of the same build had
produced 6 failed history searches and one Circuit Breaker alert for the intrusion
prevention rule.

**History searches.** The history blocks of both edited rules are unchanged. Their queries
were exercised only in an earlier local test with go-sdk v1.1.33 against a disposable
OpenSearch 2.13.0 (36 of 36 checks: thresholds, windows, other addresses, the OR branches and
the 108003 case without an address). No production cluster was queried for that.

**SDK predicate checks.** The go-sdk v1.1.33 rule replay evaluated the original and edited
conditions of both rules over the playground events above and 18 fabricated normalized
events. The edited intrusion prevention condition matches only the normalized 108003 event
that has `origin.ip`, so its history placeholder always resolves; the original condition also
matched all 13 108003 events from the playground runs, none of which has an address, so each
would have failed its history search. The edited VPN condition matches the three normalized
failures that have an address; on the playground events neither version matches. 90 of 90
and 65 of 65 checks passed.

**Go tests.** `cisco_asa_filter_test.go` has seven tests. They check that no `where` clause
compares `log.*` directly and that each helper clause keeps its old results; that no
geolocation destination lies under a field that holds a value; that a model of the engine's
step plugins, with every `where` clause evaluated by go-sdk v1.1.33, gives the playground's
result for every stored field of the 41 lines except the geolocation ones, including a
positive and a near-miss line for each step change; the reordered and guarded `where` clauses
in filter order; the rules' names, metadata, impact, grouping and history searches; 14
synthetic rule cases; and that whenever a rule with a history search matches any event, its
placeholders resolve. All seven fail against the original filter and rules and pass against
this revision. The full `plugins/alerts` suite passes: 50 tests pass, and the same 11 tests
that need other technologies' private evidence skip, as they do on the base commit.

## Deferred

Each of these needs Cisco's documentation or real records, or an owner decision, and is
unchanged here.

| Id | What | What would unblock it |
|---|---|---|
| D01 | 113015/113017: the `user IP` address goes to `target.ip`; `origin.ip` is never set (113005 and 113016 put the same token in `origin.ip`). Also the quotes around user names and the reason words in the VPN rule. This revives the VPN rule. | Cisco's explanation of 113015 and 113017, or one real 113015 record from a device whose login source is known. Ship this first. |
| D02 | Parse 108003, 338001, 338002, 113021, 109034 and 611102, which the rules name but the filter does not parse. | The layout and field meaning of each of the six messages (for 338001/338002, which address is the listed one), plus at least one real record of each. |
| D03 | The botnet and intrusion prevention text branches (`log.message`), their case and the `IPS` inside `IPSEC`. They cannot match and cost nothing until changed. | Cisco's wording for the Dynamic Filter and IPS messages, or real records with those phrases. |
| D04 | The intrusion prevention history counts any event from the address, not repeated matches. | An owner decision; it has no effect until D02. |
| D05 | Header forms the filter rejects: no timestamp, no device-id, no year, RFC 3339/5424 time, a space-padded day, `%FTD-`, an empty body. They are no longer stored with 519 errors. | Cisco's documentation of the timestamp, device-id and RFC 5424 logging options and their defaults, or real records. |
| D06 | Format variants that make a whole message pattern fail: one-digit or over-24-hour durations, 302014 without a reason, 106001 with several TCP flags, hexadecimal sequence numbers in 402114-402120, an IPv6 AAA server in 113004/113005/113016, the 302003/302004 port form, the `(user= name)` label in 402116/402118/402119, the 302305 trailing user, 113009/113011 without `=`. | Cisco's message formats or real records of each. |
| D07 | Which address is `origin` and which is `target` (connection messages, 109101-109103, 611307-611315, the 305010-305012 mapped address, trailing users) and identity values left under `log`. All three rules report the `origin` side as the adversary. | Cisco's field definitions, or real records from a device with a known layout. |
| D08 | `actionResult` values: 109102/109103 store `accepted` although the filter's own text says they failed; `accepted` on events that are not successes; `failure` is never written. Owned by the separate action-result correction, which must not map 109102/109103 to success. | Cisco's meaning of each message's outcome. |
| D09 | Byte counters: the teardown byte count in `origin.bytesSent`, 113019 `Bytes xmt`/`Bytes rcv`, fragment sizes of 106020 and 402118. | Cisco's definition of each counter. |
| D10 | `deviceTime` from the header time (`log.ciscoTime`). | The time zone of the header timestamp, from Cisco's documentation or a device with a known zone. |
| D11 | Standard values: `severity` high/medium/low and no case for level 0, `protocol` `NAT` for 611301/611303/611304 and mixed protocol case, the `action` wording. | A platform decision with a review of dashboards, saved searches and rules, and Cisco's level definitions. |
| D12 | Redundant or overlapping steps, the reasons and 113034-113039 text deleted with `log.rest`, the parentheses kept by the second 113009/113011 variant. | Nothing external; can ship later with fabricated lines (naming an action for 113034-113039 needs Cisco's meaning). |
| D13 | New rules for parsed attack-type messages that no rule reads (733100-733103, 400000-400050, 106017, 106018, 106020, 106021, 201003, 407002, 209003, 405001, 405002, 322001-322003, 406001, 406002, 605004, 710003, 113005/113016/113017, 316001, 719024). | Cisco's message explanations, and real records to size the noise and check attribution. |
| D14 | Address-only patterns for `origin.ip`/`target.ip`: 104 patterns also accept a host name and 19 accept any text. | Cisco's documentation of name substitution in syslog, or real records. |

MITRE ATT&CK v19.2 still lists T1071, T1190 and T1110 with the tactics the three rules use,
so no label changes.

## Known limits

- Every input is fabricated. No real Cisco ASA record and no Cisco documentation were
  available, so the fixtures prove the filter's behaviour on lines shaped by its own
  patterns, not what Cisco devices send.
- The latest check used EventProcessor `8a3ade7`, whose playground and plugins link go-sdk
  v1.1.36, the version the alerts module now pins; the predicates were also checked with
  v1.1.36. The original review used `497bf53` (parser and writer plugins v1.1.26, CEL plugin
  v1.1.34) and v1.1.33 predicates. Neither build is asserted to match a customer deployment.
- The playground's alert writer only records alerts. Indexing, grouping, deduplication,
  notifications and production alerts were not tested.
- The index rejection behind the geolocation change was measured on a local OpenSearch 2.13.0
  with default dynamic mapping. Production clusters report 7.10.2 compatibility and were only
  read, never written.
- The playground run has no positive control for the intrusion prevention and VPN rules,
  because this filter cannot give their messages an `origin.ip` yet (D01, D02). Their positive
  cases are the SDK predicate checks and the earlier local history test.
- The Go extraction test is a model of the engine's step plugins. It agreed with the
  playground on every non-geolocation field of the 41 lines, but `replay.py` is the check
  that runs the engine.
- `equals("log.severity", "4")` compares numbers, like the neighbouring `oneOf` severity
  clauses, so severity text such as `04` or `+4` now counts as 4; the old clause accepted only
  `4`. No fixture uses such text.
- When a 305012 duration has a one-digit hour (D06), `action` is now absent instead of the
  partial `Teardown dynamic`.
- Lines with a rejected header form stay unparsed (D05); they only lose the 519 errors.
- The 16 `log.*` geolocation fields change name. No rule or dashboard reads them, and none of
  the searched production indices held the old nested fields.
- The separate action-result correction edits the same filter. This change moves the two
  106102/106103 `actionResult` blocks without changing their values; combining the two needs
  a rebase at those blocks and at the version line.

## Reproduce

Build EventProcessor `8a3ade72bd9d12db21f6b273200588fb49540f14` (the latest check) without
changing its dependencies. With `EP` set to that checkout's absolute path:

```sh
mkdir -p "$EP/test-bin" "$EP/test-plugins"
(cd "$EP" && go build -mod=readonly -o "$EP/test-bin/playground" ./cmd/playground)
for plugin in add cast cel delete grok rename saw sew trim; do
  (cd "$EP/plugins/$plugin" && go build -mod=readonly -o "$EP/test-plugins/$plugin.plugin" .)
done
```

From this UTMStack checkout, with PyYAML installed:

```sh
(cd plugins/geolocation && go build -mod=readonly -o "$EP/test-plugins/com.utmstack.geolocation.plugin" .)
python3 plugins/alerts/testdata/cisco-asa/replay.py --playground "$EP/test-bin/playground" \
  --plugins "$EP/test-plugins" --geolocation-plugin "$EP/test-plugins/com.utmstack.geolocation.plugin"
(cd plugins/alerts && go test ./... -count=1)
```

The playground run takes about two minutes. The Go suite alone does not run the
engine.
