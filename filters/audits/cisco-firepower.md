# Cisco Firepower v11 filter and rule review

The Cisco Firepower filter (`filters/cisco/firepower.yml`) could not read the records a
Firepower Threat Defense device actually sends. Its header patterns required a timestamp and a
device name that the real records do not carry, so every real record was stored without any
parsed field and with 98 CEL errors. It had no parsing at all for the Firepower security events
(430001 intrusion, 430002/430003 connection start and end, 430007), which are the only events the
five Firepower rules can use, and none of those rules could fire. It also shared six step defects,
the misplaced geolocation results and the direct `log.*` comparisons with the Cisco ASA filter.
This revision parses the observed security events, fixes the engine-visible defects, and
corrects two rule conditions. The schema is ThreatWinds go-sdk **v1.1.36**, as pinned by
`plugins/alerts/go.mod`.

## Evidence basis

- **Genuine device records, reviewed privately.** None of the 29 v11 instances that could be
  searched holds a `firewall-cisco-firepower` record: the index pattern
  `v11-log-firewall-cisco-firepower-*` was empty on all of them (two could not be searched). One
  instance holds about 1.25 million genuine Firepower Threat Defense syslog records from one
  device over about a month, stored under another Cisco integration's data type because the
  device sends to that integration's listener (see the routing note below). They were read only,
  never written. A bounded private sample of 83 of them (20 intrusion, 20 connection start, 19
  connection end, 4 elephant-flow, 20 with two syslog messages joined by the collector) was
  replayed locally as `firewall-cisco-firepower`. None of these records, their addresses or
  their names is in this repository: the committed fixtures are fabricated and only follow the
  key layout.
- **Cisco documentation was not available.** Cisco's documentation site refused every
  automated request (HTTP 403) from the allowed fetch tool. Nothing here depends on Cisco's
  wording; changes that need Cisco's definition of a key, a value or a side are deferred.
- **What the corrections rest on.** Each change below names its basis:
  - *real records*: the header shape, the key names, which keys occur on which message, the
    separators, and values such as the intrusion events' `Priority` and `Classification`;
  - *the filter's own patterns*: what a step can receive follows from the earlier steps;
  - *the engine*: the public EventProcessor at commit
    `8a3ade72bd9d12db21f6b273200588fb49540f14`. Its
    [parser](https://github.com/utmstack/EventProcessor/blob/8a3ade72bd9d12db21f6b273200588fb49540f14/pkg/parsing/parsing.go)
    stores an error and skips the step when a `where` clause fails; its
    [grok plugin](https://github.com/utmstack/EventProcessor/blob/8a3ade72bd9d12db21f6b273200588fb49540f14/plugins/grok/main.go)
    writes nothing unless every pattern matched; its
    [kv plugin](https://github.com/utmstack/EventProcessor/blob/8a3ade72bd9d12db21f6b273200588fb49540f14/plugins/kv/main.go)
    splits on the field separator, cuts each pair at the first value separator, stores every
    value as text under `log.<key>` and keeps the last copy of a repeated key; its
    [CEL plugin](https://github.com/utmstack/EventProcessor/blob/8a3ade72bd9d12db21f6b273200588fb49540f14/plugins/cel/main.go)
    counts rule errors and, at the fifth, disables the rule and raises a
    `Circuit Breaker: <rule name>` alert. This repository's `plugins/geolocation/main.go`
    writes its result at the destination path with `sjson.Set`, which turns a string at that
    path, or above it, into an object;
  - *the SDK*: go-sdk v1.1.36
    [`utils/fields.go`](https://github.com/threatwinds/go-sdk/blob/v1.1.36/utils/fields.go)
    keeps letters, digits, dots and underscores in field names and removes everything else, so
    the key `DNS_TTL` is stored as `log.DNS_TTL` and `Prefilter Policy` as `log.PrefilterPolicy`;
    [`plugins/cel.go`](https://github.com/threatwinds/go-sdk/blob/v1.1.36/plugins/cel.go)
    declares a CEL variable only for the top-level keys an event has, so `log.messageId==N`
    fails to compile when there is no `log` object, while `equals`, `oneOf`, `greaterOrEqual`
    and `lessOrEqual` return false for a missing path;
    [`plugins/rules.go`](https://github.com/threatwinds/go-sdk/blob/v1.1.36/plugins/rules.go)
    returns an error when a `{{.field}}` history placeholder cannot be resolved.

## Filter changes (version 3.1.0)

| Change | Why | Basis | Proof |
|---|---|---|---|
| F-H1: a third header pattern for the real shape `<PRI>%FTD-<level>-<id>: <text>`, with no timestamp and no device name. It runs only when the two existing header patterns set nothing (`where: '!exists("log.messageId")'`). | The two existing patterns require a day, date, time and device name before `%FTD-`; the device sends none of them, so no real record was parsed. | Real records, engine | Playground: all 70 real records with an intact header are accepted; the 13 whose header was cut by the collector stay rejected, without errors. A LINA message sent in this shape is now parsed by the existing LINA steps (fabricated line only; no real one was seen). |
| F-K1: split the text of 430001, 430002, 430003 and 430007 (`oneOf("log.messageId", [430001, 430002, 430003, 430007])`) at `, ` and `: ` into `log.<Key>`. Map only `SrcIP`, `DstIP`, `SrcPort`, `DstPort` and `Protocol` to `origin.ip`, `target.ip`, `origin.port`, `target.port` and `protocol`, each from the first copy of its key, and delete their `log` copies. Read `UserAgent` whole, and split the text before it again. Skip a text that holds a second `%FTD-` header. | These are the device's security events and the only input of the Firepower rules; the filter ignored them. A `UserAgent` value can contain `, ` and even `, Key: value` text, which would cut the value or replace an earlier key. The collector sometimes joins two messages; their values must not be mixed. Other 4300xx IDs (for example 430005) have no real record to show their layout, so they keep only the header fields. | Real records, engine, SDK | Playground: all 63 real single-message records match an independent model of the text key for key; all 63 get `origin.ip`, `target.ip` and `protocol`, and all 59 TCP/UDP ones get both ports. Key names are stored as the SDK writes them, for example `log.DNS_TTL` and `log.PrefilterPolicy`. The 7 real records with a second message inside the text keep only `log.messageId`, `log.severity` and `log.msg`. A fabricated `UserAgent` carrying `, SrcIP: ...` text changes no standard field and no earlier key. |
| F-W1: rewrite the 98 `where` clauses that compared `log.messageId` or `log.severity` directly (for example `log.messageId==113032`) with `equals`, `greaterOrEqual` and `lessOrEqual`. | Without a `log` object every one of them failed to compile, so each unparsed record was stored with 98 errors (about 35 KB). | Engine and SDK | Playground: 83 real records went from 98 errors each to none. Go test: the 468 clauses that use these helpers on these fields keep the direct comparison's result on 2,903 events that have a `log` object, and are false without an error when there is none. |
| F-W2: `lgreaterOrEqual(...)` in the 109201-109213 trim becomes `greaterOrEqual(...) && lessOrEqual(...)`. | The function does not exist, so the clause failed on every event. | SDK | The trim now runs on 109201 (`origin.user` `alice`, `log.session` `0x1a2b` on the fabricated line). |
| F-A2: write the 16 `log.*` geolocation results to `log.<field>Geolocation` (for example `log.localIpGeolocation`) instead of `log.<field>.geolocation`. `origin.geolocation` and `target.geolocation` are unchanged. | The result replaced the address with an object, as in the Cisco ASA filter. | Engine | Playground: the address stays text next to the sibling object. Go test: no geolocation destination lies under a field that holds a value. |
| F-A3 to F-A8: the six LINA step fixes of the Cisco ASA review: 302013 direction, 302304 protocol, 305011/305012 action, 302017 users, the 106102/106103 `denied`/`accepted` order (values unchanged) and the second 113009/113011 variant only when no user was set. | Same steps and same defects as the Cisco ASA filter. | Filter patterns and engine | Playground and Go test, fabricated lines: each fix changes its field as intended and its near miss is unchanged. |

With the new header every 4300xx record gets `severity` `high`, because the device sends all of
them at syslog level 1. The same records are already stored as `high` today under the other
data type. Choosing another source for severity needs Cisco's definitions (D03).

## Rule changes

Names, impact, adversary side, grouping, history search and MITRE labels are unchanged. The
malware, indicator-of-compromise and threat-intelligence rules are unchanged (D02).

| Rule | Change | Why | Basis |
|---|---|---|---|
| `intrusion_prevention_high_priority_events` (v1.0.1) | Read message 430001 and the device's own `log.Priority` (1) and `log.Classification` (four descriptions). Remove the `log.eventType`, `log.impact` and syslog-level branches. | The rule read keys and values that no Firepower syslog record carries (`IPS_EVENT`, short class names such as `attempted-admin`), and the syslog-level branch would have been true for every intrusion event, because the device sends them all at level 1. | Real records, filter |
| `c2_nonstandard_port` (v1.0.1) | Read `log.ApplicationProtocol` and `log.InitiatorPackets` (above zero). Require a destination port and a destination outside 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16 and fc00::/7. Remove the `unknown-tcp` branch. The history search (5 events for the same source and destination within 1 hour) is unchanged. | The rule read keys that do not exist (`appProto`, `initiatorPackets` compared to `true`). With only the key names and the packet test fixed, internal web services on unlisted ports (for example 8181) would have raised about 53,000 to 58,000 alert documents a month on the one device. | Real records, filter |

**Assumptions about Cisco's meaning.** These follow the real records but could not be checked
against Cisco's documentation:

- intrusion rule: `Priority` is the device's rating of the signature; the three classification
  descriptions that no real record carried (`Attempted Administrator Privilege Gain`,
  `Web Application Attack`, `Exploit Kit Activity Detected`) are spelled as the observed one
  (`Attempted User Privilege Gain`); `SrcIP` is the attacking side for every signature;
- non-standard-port rule: `ApplicationProtocol` is the device's content-based identification;
  `DstIP`/`DstPort` is the service side; `InitiatorPackets` counts the initiator's packets (only
  "above zero" is used).

**Expected volume on the one device (30 days).** Intrusion rule: 32 alert documents, grouped
under 7 top-level alerts; all 32 events were already blocked by the device and look like one
internal scanner. Non-standard-port rule: 122 to 188 alert documents under 7 to 13 top-level
alerts, mostly SSL to outside hosts on port 9001. The second figure is a model of the history
threshold over the device's records, not an executed search.

## Validation

**Engine and versions.** EventProcessor `8a3ade72bd9d12db21f6b273200588fb49540f14`: every
parser, writer and CEL plugin links go-sdk v1.1.36; the geolocation plugin was built from this
repository (go-sdk v1.1.34). Rule predicates were also replayed with go-sdk v1.1.36.

**Genuine records (private, not committed).** The unchanged and the corrected filter ran on the
same 148 lines: the 83 genuine records above and 65 fabricated lines.

| Run | Result |
|---|---|
| Unchanged filter, 148 lines | 148 events, 13,049 errors. Every genuine record: no `log` object, 98 errors, no field. |
| Corrected filter, 148 lines | 148 events, no errors. 70 of 70 intact headers accepted, 13 of 13 cut headers rejected; standard fields only the five listed above; 30 of 30 fabricated checks pass (1 of 30 before). |
| The 144 of these lines that also ran on the previous engine (EventProcessor 497bf53, go-sdk v1.1.26) | The unchanged filter's events are identical. The corrected filter's events differ only in `log.DNSTTL` becoming `log.DNS_TTL` (two genuine records) and in two fabricated 430005 lines, which an earlier draft parsed and this revision leaves unparsed. |
| go-sdk v1.1.36 predicate replay of the rules on those events | Intrusion rule: 27 matches, all 20 genuine intrusion records and 7 fabricated positives; no near miss. Non-standard-port rule: the 8 fabricated positives, none of the 83 genuine records and no near miss; its placeholders resolve on every match. The three unchanged rules match only fabricated lines carrying crafted keys. The original two conditions match none of the genuine records. |
| Playground with all five rules and no OpenSearch, 140 intrusion, connection and near-miss lines | 30 local alerts: 27 intrusion alerts (the replay's matches) and one for each crafted-key line of the three unchanged rules; each alert's adversary and target are the event's origin and target. No Circuit Breaker, no compile or rule error, and no history search: no line met the non-standard-port condition. |
| Same, the 8 positives and 2 near misses of the non-standard-port rule | The rule's condition matched the 8 positives and each history search failed (connection refused: no OpenSearch). After the fifth failure the engine raised one Circuit Breaker alert for the rule. A test-only copy without the history search raised the 8 expected alerts. This shows the search is reached; it does not test it. |

In a separate 30-day count on the instance, all 32 genuine intrusion events carry `Priority` 1
and `Attempted User Privilege Gain`, so the corrected intrusion condition matches all of them.

**Fabricated regression, committed.** `plugins/alerts/testdata/cisco-firepower/` holds 58
invented raw lines (`raw.json`), their expected fields and alerts (`expected.json`), the 13
shared grok definitions this filter uses (`patterns.yaml`, copied from the repository changelog
`20250616001_insert_utm_regex_pattern.xml`), invented geolocation data (`geolocation-data/`)
and `replay.py`. Addresses are RFC 5737 and RFC 3849 documentation addresses, except the
RFC 1918 and RFC 4193 private addresses of the private-destination near misses; device, user,
zone, rule and policy names are examples and the device UUID is made up. `replay.py` stages the
non-standard-port rule as a condition-only test copy because no OpenSearch runs; it passed: 58
events without errors, every stored field as recorded, 10 alerts, each from its intended rule
with the event's origin as adversary, and no history search.

**Go tests.** `cisco_firepower_filter_test.go` has seven tests. They check that no `where`
clause compares `log.*` directly, that every clause runs without a `log` object and that each
helper clause keeps the direct comparison's results; that no geolocation destination lies under
a field that holds a value; that a model of the engine's step plugins, with every `where` clause
evaluated by go-sdk v1.1.36, gives the playground's result for every stored field of the 58
lines except the geolocation ones, including a positive and a near-miss line for each change; the
new, reordered and guarded `where` clauses in filter order; the rules' names, metadata, impact,
grouping, MITRE labels and history search; 27 synthetic rule cases plus the rule results on all
58 lines; and that whenever the rule with a history search matches, its placeholders resolve.
Six of the seven fail against the original filter and rules; the placeholder test passes on both,
because the original condition also required both addresses. The full `plugins/alerts` suite
passes: 50 tests pass and the same 11 tests that need other technologies' private evidence skip,
as they do on the base commit (43 pass, 11 skip).

## Deferred

Each of these needs Cisco's documentation, more real records or an owner decision, and is
unchanged here.

| Id | What | What would unblock it |
|---|---|---|
| D01 | Header shapes other than the real one: no day of week, no `<PRI>`, RFC 5424, other timestamp and device-id options. | Cisco's logging options and defaults, or real records in those shapes. |
| D02 | Parse 430004 to 430006 (file, malware and other unseen events); rewrite the malware, indicator-of-compromise and threat-intelligence rules with that parsing, and fix their labels (T1566 on the malware and indicator rules, the TA0040 reference, the parent T1071 link). Keys keep underscores, so a rewritten rule must read, for example, `log.SHA_Disposition`. | Cisco's field list for these events, plus one real record of each from a device with those features. |
| D03 | Severity of 4300xx events: the syslog level (today `high` on all), `EventPriority` or `Priority`. | Cisco's definitions and a review of severity consumers. |
| D04 | `deviceTime` from `FirstPacketSecond` (a connection's first packet, not its end) and the time zone of the older header time. | Cisco's definition and a platform decision. |
| D05 | `User`, byte and packet counters, `URL`, DNS names and the original client address to standard fields. | Cisco's definition of the side each belongs to. |
| D06 | Outcome from `AccessControlRuleAction` and `InlineResult`; action names; the LINA outcome meanings. | The separate action-result correction and Cisco's value definitions. |
| D07 | Intrusion rule technique T1203 against the device's own tag T1190 on the real events. | An owner decision; no detection effect. |
| D08 | Non-standard-port options: `Unknown` on TCP, deduplication instead of grouping, more TLS ports (for example 5222 and 50051). | Cisco's meaning of `Unknown` and an owner decision after some weeks of data. |
| D09 | The LINA items shared with the Cisco ASA review (sides, formats, counters, vocabulary, cleanups). | Cisco's syslog guide or real LINA records from this device type (none seen). |
| D10 | Records with two syslog messages joined (about 1.25 %) and the routing below. | A sender or collector change, outside this filter. |
| D11 | New rules for real data no rule reads (blocked connections, intrusion results, DNS answers, remote-access VPN users, long or large connections). | Cisco's meaning of each key and a volume check. |

MITRE ATT&CK v19.2 still lists T1571, T1203, T1566 and T1071 with the tactics the rules use;
no label changes here.

## Routing note for customers

The only device seen sends to the syslog listener of another Cisco integration, so its records
are stored under that integration's data type and this filter never sees them. The collector
assigns the data type of the listener that received a record, and the Cisco integrations share
the same default ports (UDP 514 and TCP 1470). Until such a device is pointed at the Firepower
integration's own listener (or the collector distinguishes the senders), none of these
corrections apply to its records.

## Known limits

- History searches were not executed. No OpenSearch was available for this review; the
  non-standard-port rule's threshold, window and address terms are unchanged and untested, and
  its volume figures are a model.
- Production indexing of the new `log.*` fields was not tested. Each genuine record adds
  22 to 31 text fields under `log`; the Firepower index is empty everywhere today, so no
  stored field changes type, but mapping growth was not measured.
- One device, one software version and one configuration. Keys or layouts that device did not
  send (for example the file and malware events) are not parsed.
- Every committed input is fabricated. The genuine records prove the layout of one device;
  they stay private and are described here only in aggregate.
- Key-value values are text, including counters such as `log.InitiatorPackets`; the rule's
  `greaterThan` compares numeric text as a number.
- A record whose collector framing joined two messages keeps only its header fields.
- The playground's alert writer only records alerts. Indexing, grouping, deduplication,
  notifications and production alerts were not tested.
- The Go extraction test is a model of the engine's step plugins; it agreed with the playground
  on every non-geolocation field of the 58 lines, but `replay.py` is the check that runs the
  engine.
- The separate action-result correction edits the same filter. This change moves the two
  106102/106103 `actionResult` blocks without changing their values; combining the two needs a
  rebase at those blocks and at the version line.

## Reproduce

Build the EventProcessor commit above without changing its dependencies. With `EP` set to
that checkout's absolute path:

```sh
mkdir -p "$EP/test-bin" "$EP/test-plugins"
(cd "$EP" && go build -mod=readonly -o "$EP/test-bin/playground" ./cmd/playground)
for plugin in add cast cel delete grok kv rename saw sew trim; do
  (cd "$EP/plugins/$plugin" && go build -mod=readonly -o "$EP/test-plugins/$plugin.plugin" .)
done
```

From this UTMStack checkout, with PyYAML installed:

```sh
(cd plugins/geolocation && go build -mod=readonly -o "$EP/test-plugins/com.utmstack.geolocation.plugin" .)
python3 plugins/alerts/testdata/cisco-firepower/replay.py --playground "$EP/test-bin/playground" \
  --plugins "$EP/test-plugins" --geolocation-plugin "$EP/test-plugins/com.utmstack.geolocation.plugin"
(cd plugins/alerts && go test ./... -count=1)
```

The playground run takes about two and a half minutes. The Go suite alone does not run the
engine.
