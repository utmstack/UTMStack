# Palo Alto PAN-OS filter and correlation review

Replacement for historical #2614. The standard is the ThreatWinds go-sdk protobuf that v11
pins, v1.1.36 since official `v11` (`d2479c1a`) was merged into this branch (the review used
v1.1.33, whose `plugins.proto` is identical), and its official wiki. This proposal changes
the PAN-OS filter and all seven existing detection definitions together; four
direction-specific counterparts bring the resulting rule set to eleven. Existing impact
ratings and thresholds are retained.

## Parsing and standard fields

- Accept bare/PRI, RFC3164 and RFC5424 envelopes. Select CSV layouts by the actual type
  column, without subtype whitelists or matching type names inside a description. Include
  the leading FUTURE_USE column, use the SDK's `csv.source` property, and correct the
  missing traffic virtual-system column and GTP protocol/action offset.
- Preserve all prior CEF extension keys and the official template's case-sensitive `Pan*`
  spellings. Extract at complete field boundaries, including the last extension and escaped
  equals signs. Generic THREAT CEF uses Signature ID as threat ID and `cat` as subtype;
  URL/Data/WildFire templates reverse those meanings. CONFIG's signature carries its result.
  SYSTEM's `cat` carries event ID. Reporter identity stays separate from session endpoints.
  This covers the tested seven-field CEF layouts based on the PAN-OS 10.0 guide; some printed
  templates have malformed header shapes and are not claimed as supported without profile evidence.
- Preserve the prior seven-field LEEF-labelled compatibility dialect. This is not a claim
  of verified standard LEEF 1.0/2.0 support; unmatched formats retain the original raw record.
- Keep source/destination session addresses distinct from NAT, XFF, reporting-firewall and
  GlobalProtect private addresses. Validate original IPs before promotion or geolocation,
  including semantic unspecified IPv6/IPv4-mapped zero addresses. Preserve invalid vendor
  values. Validate port ranges, MAC syntax and nonnegative counters before standard mapping.
- Map users, endpoint hosts/MAC/OS, explicit offset-bearing device time, URL resource, and
  client-relative CSV byte/packet counters. CEF packet fields have explicit directions.
  Only absolute HTTP(S) requests are promoted to `target.url`; scheme-less requests remain
  in `log.misc` because the SDK wiki defines the standard field as a full URL. Common
  hostname guards also apply to CONFIG, so missing/empty/`-`/`N/A` names are not promoted.
  Only valid hashes are promoted, to the threatened side identified by attack direction.
  WildFire `filePath` is the analysis cloud location; it is not a filesystem path.
- Recover user and source address only from explicit SYSTEM auth-fail/auth-success message
  shapes. Do not promote an arbitrary address mentioned in a description. CONFIG's admin
  client/user belong to the origin, not the reporting firewall or destination user.
- Preserve newer documented optional tail fields. Distinguish old/new URL tails with explicit
  timestamp and HTTP-status anchors, and default/custom CONFIG tails with timestamp anchors.
  Current URL and Data Filtering tails place cluster name before flow type; generic THREAT
  uses the reverse order. Legacy URL's unverified final two fields remain numbered vendor data.
  Ambiguous optional tails remain numbered vendor fields rather than invented host/time fields.
  This also avoids treating a configuration XPath as a file path.

## Outcomes and detections

- TRAFFIC `allow` means the policy allowed traffic. Explicit threat/policy/decryption blocking
  session-end reasons override that outcome to denied; resource exhaustion is failure.
  THREAT `allow` means a flood-detection alert and never creates success. Alert-only detection,
  syncookie and CONFIG Submitted do not prove success. No connectionStatus is inferred.
- Normalize explicit denial and authentication/configuration failures to SDK-wiki values.
  Textual informational/low, medium, high and critical become info, warning, error and critical.
  Retain original actions, results, severities and vendor fields for review and custom consumers.
- Consume the produced action, category, subtype, threat ID and authentication fields. A blocked
  vendor category does not prove an external intelligence-feed match; a known CVE or unknown
  category does not prove a zero-day; a WildFire verdict does not prove execution. Rename and
  describe these rules accordingly, removing unsupported ATT&CK claims.
- WildFire malware detection requires the malware verdict; benign, grayware and phishing
  verdicts are not labeled malware. URL coverage includes alert-only and blocked requests, so
  its title no longer claims every match was blocked or every risky category was malware.
- Preserve physical event origin/target. For directional threats, choose the alert adversary
  from documented client-to-server/server-to-client direction. For risky URL categories, the
  remote resource is the adversary and the requester is the alert target. Missing threat
  direction is an attribution gap, not permission to guess a malicious endpoint.
- DNS histories cover the legacy dns subtype plus spyware sinkhole/dns-c2/dns-malware events.
  Alert-only ddns, parked and grayware categories are not added to the DNS history. A source
  address may identify a forwarding resolver; trace the client before claiming compromise.
  DNS and URL union predicates use the general Intrusion Detection category without an
  ATT&CK technique: their matched populations do not uniformly establish command and control.
- Count recomputed candidates in five SDK histories: authentication 10/15m, DNS 3/30m, URL
  5/30m, and vulnerability 2/30m for each direction. Require collector, firewall serial and
  virtual system; vulnerability also retains both endpoints and direction. Successful or
  unrelated activity cannot satisfy the counts. Clear input-supplied candidates and scope.

## Re-validation on the latest versions

On 2026-09-24 official `v11` moved to `d2479c1a3705eec6a00016689c2bf5fbcc1814f2`, whose
`plugins/alerts` pins go-sdk v1.1.36, and EventProcessor `main` moved to
`8a3ade72bd9d12db21f6b273200588fb49540f14`, whose playground and parser, writer and CEL
plugins all link go-sdk v1.1.36. `v11` was merged into this branch; no file overlaps this
draft, so nothing conflicted. The newest published engine image,
`ghcr.io/utmstack/utmstack/eventprocessor:v11.2.14` (built 2026-09-24 19:13 UTC on base image
`eventprocessor/base:1.1.7`), embeds Go build information showing that its playground and
plugin binaries come from the same revision `8a3ade7` with go-sdk v1.1.36, built with
go1.26.8 for linux/amd64. The local build used here is that source revision compiled
natively for darwin/arm64 with go1.25.7; only the Go toolchain and platform differ. Its
`csv` plugin was built from the same clean source.

What changed in the SDK, and what it means here:

- Since v1.1.35, `utils.SanitizeField` keeps `_` in the field names that the `json`
  (top-level keys), `kv`, `grok`, `csv`, `xml`, `add` and `rename` plugins write; other
  characters are still removed. This filter writes its 784 names, including 368 CSV
  headers, in camelCase with letters and digits only (for example `log.paType`), so every
  version stores them unchanged and the conditions and rules find them. No name needed a
  change.
- v1.1.36 makes `regexMatch` match string values only again. Every `regexMatch` in this
  filter reads a text field written by `grok` or `csv`, so no result changes.
  `plugins.proto`, `plugins/cel.go` and `plugins/rules.go` are identical in v1.1.33 and
  v1.1.36.
- The only file changed for the new SDK is a test comment that still said the sanitizer
  removes underscores.

| Check on the latest versions | Result |
|---|---|
| Full `plugins/alerts` suite, go-sdk v1.1.36 | 46 tests pass, 12 skip, none fail. The skipped tests need private evidence (the 11 of other technologies and this draft's private records). |
| The 34 private records, same suite | 34 of 34 pass once their expected field names are written as the camelCase names this draft now uses; the private expectation file predates that rename. With them the suite gives 47 passes, 11 skips and no failures (2,418 passing results with subtests). |
| The 122 committed fabricated lines through the EventProcessor `8a3ade7` playground, with the filter and the eleven rules as committed | 122 events, but 75 carry the `csv` plugin error `number of headers should match the number of resulting columns`: all 28 TRAFFIC and 36 THREAT lines and 11 lines of other CSV types. The real plugin stops the pipeline when a line has fewer columns than the layout's headers (the TRAFFIC layout lists 130; the TRAFFIC line has 102), so those events keep only the envelope fields. The offline model skips missing columns instead. Of the 47 lines without errors, 16 carry every expected field and 31 do not (20 CEF, 8 SYSTEM and 3 others); for example, SYSTEM authentication lines lack `origin.ip` and `origin.user`, because the real `grok` plugin matches each pattern from the start of the remaining text, while the model searches the whole text with one expression. The same 120 lines present in an earlier run on EventProcessor `497bf53` gave identical events, so the SDK change did not cause this. |
| go-sdk v1.1.36 rule replay over those playground events | 2 of the 38 expected rule matches. On every event that carries its expected fields, each rule gives exactly its expected result and history placeholders resolve; the misses all come from the missing fields. |

So the committed Go suite passes, but the real parser plugins do not yet produce what it
expects for most native CSV lines, most CEF lines and part of the SYSTEM lines. This is
not a v1.1.36 change and is left to a separate correction of the CSV layouts, the SYSTEM
authentication extraction and the offline model; the rules cannot be relied on for native
CSV traffic until then.

## Validation and rollout limits

The committed suite has **122 fabricated raw fixtures**, positive/negative assertions for all
**eleven rules**, strict final Event decoding, independent alert-side expectations, and **five
real SDK history-request tests** against loopback mocks. History tests cover below/at threshold,
expiration, irrelevant/unmarked events, wrong identities and unresolved required placeholders.
The manifest also validates schema/conditions with the shared contract runner. These checks
first used the SDK **v1.1.33** pinned by the reviewed v11 snapshot and now pass with
**v1.1.36** (see above).

Read-only sampling on 23 September 2026 also retrieved **34 distinct stored records** and the
mounted filter/rules from two deployments. Twenty-two are native SYSTEM records: nine
authentication failures, three authentication successes and ten other SYSTEM events. Twelve
are Cortex XDR records arriving under this data type and are deliberately not treated as
PAN-OS CSV. Each stored record includes two `unknown operation` errors with an empty step;
the deployed filter contains the corresponding two empty steps. The sampled records were
indexed, but lack the expected normalized event fields. This establishes actual runtime
errors and missing mappings, not wholesale configuration rejection or event loss.

Private raw replay verifies independently derived SYSTEM columns, explicit authentication
user/address extraction, severity and time. Nine authentication failures match the revised
authentication predicate; successes and other sampled classes remain negative. Documented
traffic/threat/CONFIG CSV and CEF layouts are checked against the cited contracts and fabricated
regressions, not those SYSTEM samples. The legacy LEEF-labelled dialect has compatibility tests
only; neither it nor standard LEEF has vendor-semantic validation here. Ambiguous CEF
counters/severity remain unpromoted.

Raw extraction in the Go suite uses an explicitly declared offline YAML/Go CSV model. It is not
execution of the EventProcessor, geolocation service or live alert creation; the playground run
above shows where it and the real parser plugins differ. Native replay and deployment
identity evidence are maintained separately; no customer records or identifying metadata are
committed. No production alert-volume result is asserted.

Before rollout, correct the parser differences described above, exercise the parser with the
actual appliance export profiles, measure throughput, and verify the resulting indexed fields
and alerts. In particular:

- Palo Alto publishes no supported CEF guidance after PAN-OS 10.0. Its guide gives a numeric
  0–10 importance scale but no exact mapping to PAN-OS textual severity bins. Numeric-only CEF
  severity remains vendor data and does not satisfy textual high/critical predicates. Verify
  the export profile's mapping before enabling that detection coverage; do not silently guess.
- The same CEF guide contradicts the byte orientation of `in`/`out` between its template and
  extension dictionary. Both values are retained, but are not promoted until the actual
  configured profile establishes orientation. Escaped CEF strings are preserved; full escape
  decoding, arbitrary custom profiles and standard LEEF remain unverified.
- Current vendor pages redirect versioned URLs to the general NGFW guide. Tests cover the
  documented layouts, not every firmware/export variant. Unknown optional tails stay vendor
  data. Counters above the exact numeric comparison range are retained without promotion.
- Cortex XDR has a separate export contract. A Cortex payload on this dataType is not parsed
  as a PAN-OS firewall record; routing and Cortex-specific parsing require a separate review.
- Install matching filter/rules together. New candidate fields need up to thirty minutes of
  fresh history. Reconcile renamed rule definitions with installed definitions; this proposal
  does not rewrite existing indexed documents, custom searches or local rule modifications.
- Indexed `lastEvent.*` grouping requires the shared alert fix #2627, already merged in the
  reviewed v11 base.
  Validate its grouping/volume behavior independently. The single-event and repeated-signature
  detections, and the category and URL detections, can intentionally overlap.
  History rules require their scope identities; discrete rules can still match with missing
  collector or firewall identity, in which case grouping uses the available keys and is less
  specific. Validate the actual export profile and grouping behavior before rollout.

## References

- [SDK protobuf, pinned v1.1.36](https://github.com/threatwinds/go-sdk/blob/v1.1.36/plugins/plugins.proto) (identical to v1.1.33)
- [SDK field-name sanitizer, v1.1.36](https://github.com/threatwinds/go-sdk/blob/v1.1.36/utils/fields.go)
- [EventProcessor csv plugin at 8a3ade7](https://github.com/utmstack/EventProcessor/blob/8a3ade72bd9d12db21f6b273200588fb49540f14/plugins/csv/main.go)
- [Filter step semantics](https://github.com/threatwinds/go-sdk/wiki/Filter-Steps-Reference)
- [Rule evaluation and history](https://github.com/threatwinds/go-sdk/wiki/Implementing-Rules)
- [PAN-OS syslog field descriptions](https://docs.paloaltonetworks.com/pan-os/11-1/pan-os-admin/monitoring/use-syslog-for-monitoring/syslog-field-descriptions)
- [Traffic fields and session-end reasons](https://docs.paloaltonetworks.com/ngfw/administration/monitoring/use-syslog-for-monitoring/syslog-field-descriptions/traffic-log-fields)
- [Threat actions, verdicts and direction](https://docs.paloaltonetworks.com/ngfw/administration/monitoring/use-syslog-for-monitoring/syslog-field-descriptions/threat-log-fields)
- [System authentication context](https://docs.paloaltonetworks.com/ngfw/administration/monitoring/use-syslog-for-monitoring/syslog-field-descriptions/system-log-fields)
- [Configuration fields](https://docs.paloaltonetworks.com/ngfw/administration/monitoring/use-syslog-for-monitoring/syslog-field-descriptions/config-log-fields)
- [URL fields](https://docs.paloaltonetworks.com/ngfw/administration/monitoring/use-syslog-for-monitoring/syslog-field-descriptions/url-filtering-log-fields)
- [Data Filtering fields and subtypes](https://docs.paloaltonetworks.com/ngfw/administration/monitoring/use-syslog-for-monitoring/syslog-field-descriptions/data-filtering-log-fields)
- [GTP fields](https://docs.paloaltonetworks.com/ngfw/administration/monitoring/use-syslog-for-monitoring/syslog-field-descriptions/gtp-log-fields)
- [DNS Security logging](https://docs.paloaltonetworks.com/dns-security/administration/monitor-dns-security/view-dns-security-logs)
- [CEF support boundary](https://docs.paloaltonetworks.com/resources/cef)
- [PAN-OS 10.0 CEF guide](https://docs.paloaltonetworks.com/content/dam/techdocs/en_US/pdf/cef/pan-os-10-0-cef-configuration-guide.pdf)
