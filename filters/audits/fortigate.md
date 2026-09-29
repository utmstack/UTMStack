# FortiGate filter and correlation review

This draft updates one FortiGate filter and six of its seven rules on UTMStack
`v11`. The sandbox rule is reviewed and covered by regression fixtures without a
rule edit. No customer configuration was changed, no live alert was generated,
and no reduction in customer false positives has been measured.

## Source and evidence

The data model authority is the [go-sdk protobuf at v1.1.31](https://github.com/threatwinds/go-sdk/blob/v1.1.31/plugins/plugins.proto),
the version pinned by `plugins/alerts/go.mod` in the reviewed v11 source. SDK main
at `92f4588df51441ae7f273e0e1eac457c9b20f0b7` has the same protobuf for this
review. The filter and rule wiki was read at commit
`c18b54bd5ea5a34abb0e690458d73f89835edd29`. The v11 comparison point is
`660c796f168670fd6ccb079995c09c46d4f0068d`; the prior source draft is
`31b82b551f816c229f983c72ba5164bea0991f14`.

Read-only sampling collected 63 distinct native key/value records from three
instances, using recent records and bounded representatives of action and event
classes in the preceding seven days. All three deployed filter copies have
SHA-256 `820b44dc36135080131c537fb44736dec2c22915f69cc4d551cadf78ab651328`.
Raw records, document IDs, instance provenance and deployed copies stay in the
private evidence pack; only aggregate observations and fabricated fixtures are
published here. The sample is stratified, not a prevalence estimate.

All 63 records have ingress `dataSource`, appliance `log.devid` and VDOM `log.vd`.
None has an indexed `actionResult`. Existing populated IPs, ports, source MAC,
byte counters and packet counters have the correct physical source/destination
orientation in this sample. Their absence in other event classes is not, by
itself, a defect. No malformed IP, alias-form input, CEF record, DLP event or
malicious sandbox result was observed in this bounded sample.

## Filter corrections

- Recover complete consumed key/value strings, including a quoted final field.
  The deployed output loses or truncates relevant `msg`, `logdesc`, user and
  multiword OS values. The updated recovery preserves all 40 sampled messages
  and 13 descriptions exactly, apart from surrounding quotes.
- Restrict syslog priority and CEF header recognition to their actual envelopes.
  Quoted message text containing HTML or a CEF-looking string cannot become a
  header. Consumed keys are reconstructed only at quote-aware field boundaries,
  so text such as a fake `srcip=` inside a message cannot supply an identity.
  Unterminated quoted values cannot fall back to an unquoted identity.
- Validate an IP at its original vendor-field rename before geolocation. Native
  aliases keep precedence; an invalid preferred value remains under its vendor
  key instead of silently falling back to a second alias. IPv4, IPv6 and quoted
  addresses are covered, including alternate unspecified representations.
  Invalid values do not enter `origin.ip` or `target.ip`.
- Preserve physical traffic roles. Promote `remip` only for explicit SSL-VPN
  authentication events, including the exact failed-login ID. All six sampled
  SSL login failures had `remip` but no `origin.ip`; ordinary tunnel shutdown and
  IPsec negotiation errors do not acquire authentication or adversary semantics.
- Copy supported endpoint fields into `origin.host`, `target.host`,
  `origin.operatingSystem`, `target.operatingSystem`, `target.mac`,
  `origin.user`, `origin.group` and `target.user` when present and meaningful.
  Vendor originals remain available. In the modelled sample this adds 13 source
  hosts, 16 source OS values, three destination OS values and eight destination
  MACs. Appliance identity and ingress `dataSource` remain unchanged.
- Copy six absolute HTTP/FTP URLs to `target.url` and seven valid UTM destination
  domain names to `target.domain`. Relative URLs remain in `log.url`; an embedded
  `://` does not make a relative path absolute. SDK `path` denotes a filesystem
  directory, so it is not used for a URL path.
- Derive outcome after quote cleanup. Policy `accept` maps to `success`; explicit
  policy or UTM denial maps to `denied` and wins over an accept. Exact failed
  connection/admin/SSL authentication classes map to `failure`. Bare `dns` or
  `ip-conn` action text is insufficient evidence of failure. The observed
  traffic `dns` records carry failed-connection log ID 11, which independently
  justifies their failure result. A normal `close` ends an allowed, established
  session and maps to `success`; resets and timeouts describe a closed
  connection without manufacturing a successful action.
- Preserve existing numeric protocol conversions: SDK v1.1.31 supports numeric
  comparisons against the parsed numeric strings. Map vendor logging levels to
  standard severity while retaining separate IPS attack severity under `log`.

The protobuf defines `actionResult` as a string, not an enforced enum. This patch
uses the wiki's `success`/`failure`/`denied` convention consistently with its
consumers; it does not claim that alternative strings are rejected by the SDK.
A permitted policy action is not proof of a completed TCP handshake or successful
application operation.

## Consumers and correlation

| Rule | Corrected contract |
|---|---|
| Admin account compromise | Recover full success description and `origin.user`; failed-login history requires the same source IP, user, ingress, appliance and VDOM, with the exact failed-admin event class. Group user identity on the adversary side. |
| Admin session anomaly | Preserve the existing heuristic but exclude IPv6 loopback, link-local and unique-local addresses from its external-source branch. Public IPv6 remains eligible. This edge is a synthetic predicate finding, not observed live incidence. |
| SSL-VPN brute force | Count explicit SSL authentication failures from one remote source and device/VDOM/ingress. Routine tunnel-down and generic IPsec negotiation messages are not authentication failures. |
| Critical IPS activity | Consume actual sibling type/subtype/severity fields and a denied outcome; history counts matching high/critical denied IPS candidates rather than unrelated traffic from that source. |
| DLP exfiltration | Require UTM/DLP context, consume the produced action/profile fields, and restrict history to matching DLP candidates and source/device/VDOM/ingress. Generic administrative messages mentioning DLP do not qualify. |
| Antivirus outbreak | Require the infected-file event class and denied outcome; count repeated blocked infected files from one physical source and appliance/VDOM. Analytics submissions, file-policy blocks and scan errors are not malware proof. The rule does not establish endpoint infection or successful delivery. |
| Sandbox malicious verdict | Retain the rule; trim consumed verdict/risk fields and verify positive and clean synthetic cases. No live malicious sandbox example was available. |

The five existing history counts and windows are preserved. Four candidate
markers under `log.correlationCandidate` have conditions exactly matching their
consuming trigger predicates. Input-supplied markers are cleared first. Every
history requires the identity needed to resolve its placeholders; missing
appliance/VDOM/actor data skips that correlation instead of pooling unknown
identities. Grouping includes appliance and VDOM so overlapping private IPs on
different devices do not collapse into one alert group. The antivirus history
uses physical appliance/VDOM/source identity without an additional ingress term.

SDK history counts are at-least thresholds over a processing-time `@timestamp`
lower bound. They do not establish strict event-time ordering, unique sources or
distinct destinations. Existing records without the new markers or promoted
identities will not satisfy the updated histories; allow the retained 15-minute
and one-hour windows to warm up after staged rollout.

## Validation and boundaries

- `go test ./... -count=1 -v` passes in `plugins/alerts` with cached dependencies
  and the optional private-evidence input enabled. The history test uses a local
  loopback HTTP mock, not a customer OpenSearch cluster.
- 72 fabricated raw fixtures exercise extraction, alias precedence, invalid and
  quoted IPs, quoted-key injection, malformed quotes, header ambiguity, new
  standard mappings, outcomes and all seven rule predicates. Every fixture
  asserts both matching and nonmatching rules, and candidate-marker parity.
- Actual SDK v1.1.31 CEL, protobuf conversion, placeholder expansion, generated
  search requests, mapping resolution and threshold decisions are exercised.
  All five history rules have count-minus-one/count, inside/expired-window,
  wrong-identity, wrong-event-class and missing-placeholder checks. Benign raw
  histories are reparsed to verify that they do not gain candidate markers.
- All 63 private raw records pass the same offline extraction model and actual
  SDK predicates, preserving their existing populated standard network fields.
  The model produces 35 outcomes, six additional SSL peer IPs, seven users and
  four groups. Predicate matches are one admin-success candidate, six SSL login
  failures and one critical denied IPS event; these are not generated alerts or
  proof that live history thresholds were met. The prior predicates matched no
  sampled records, which alone would not prove all seven rules defective.
- The shared contract manifest lists all seven consumers and contains nine
  isolated normalization checks for outcomes and numeric protocol strings. It
  requires the separate shared contract runner and does not test raw extraction.
  The standalone raw and SDK-history tests in this draft run without that runner.
  An overlay of the current shared runner passed the schema/CEL checks and all
  nine normalization cases alongside this draft's standalone tests.

The filter executor is closed. Raw parsing here is an explicitly limited offline
model of the YAML grok/rename/trim/cast/add/delete steps and observed KV splitting;
it is not an execution of the closed EventProcessor. Geolocation lookups, live
OpenSearch behaviour, alert creation and end-to-end delivery were not executed.
CEF compatibility is documented and synthetic, including a bounded vendor-shaped
example; arbitrary CEF escaping and unobserved vendor variants remain unverified.
The recovery adds gated scans for 129 consumed/mapped keys and aliases, not every
vendor field. Its CPU cost and actual parser behaviour require staging validation
at realistic message sizes and event rates.

Follow-up (2026-09-24): the EventProcessor grok step trims the text before each
pattern, rejects an empty match and writes fields only when every pattern
matched. The recovery extractors' trailing boundary pattern therefore never
matched, and none of the 129 wrote a field. Each extractor now reads a quoted
value, or unquoted words up to the next `key=`, and checks the boundary within
the same pattern. An empty value, or a quoted value followed directly by text,
is left unset instead of taking the next pair. The raw contract test
consumes grok patterns in order, as the executor does. Fortinet defines every
traffic action other than `deny` as allowed by policy. On two deployments, every
`close` record that carried a received-packet count had at least one received
packet, so `close` now maps to `success`.

`lastEvent.*` grouping requires the separate alert-foundation correction that
resolves indexed event aliases against the event carried by the Alert. This
source draft does not duplicate that runtime fix; deploy the foundation before
depending on these grouping scopes. Verify resulting alerts and history windows
in staging before production approval.

Outcome normalization changes queries that depend on prior literal values or
missing outcomes. Maintained source consumers were reviewed, but private saved
queries and customer dashboards were not enumerated. Review that compatibility
and the history warm-up during rollout; retained vendor fields support migration.

## Deliberately unresolved mappings

All sampled `eventtime` values are 19-digit Unix nanoseconds and agree with vendor
date/time/timezone; indexed `deviceTime` equals ingestion time. The available
filter wiki documents Go-layout time parsing, not a verified Unix-nanosecond
conversion. Preserve the original time under `log` until a supported conversion
can be proven; do not put an epoch-nanosecond string directly into `deviceTime`.
NAT pre/post-translation fields remain vendor fields because this SDK has no
dedicated equivalent. Vendor file-type labels are not MIME types, and the sample
does not establish a safe file-side role for incoming content, so those fields
are not promoted speculatively.

## References

- [Filter steps](https://github.com/threatwinds/go-sdk/wiki/Filter-Steps-Reference), [standard event schema](https://github.com/threatwinds/go-sdk/wiki/Standard-Event-Schema), [rule implementation](https://github.com/threatwinds/go-sdk/wiki/Implementing-Rules).
- [FortiOS log fields](https://docs.fortinet.com/document/fortigate/7.0.4/fortios-log-message-reference/357866/log-message-fields), [failed connection ID 11](https://docs.fortinet.com/document/fortigate/7.0.4/fortios-log-message-reference/11/11-log-id-traffic-fail-conn), [DNS query event](https://docs.fortinet.com/document/fortigate/7.0.4/fortios-log-message-reference/54000/54000-log-id-dns-query).
- [Admin login success](https://docs.fortinet.com/document/fortigate/7.0.4/fortios-log-message-reference/32001/32001-log-id-admin-login-succ), [admin login failure](https://docs.fortinet.com/document/fortigate/7.0.4/fortios-log-message-reference/32002/32002-log-id-admin-login-fail), [SSL login failure](https://docs.fortinet.com/document/fortigate/7.0.4/fortios-log-message-reference/39426/39426-log-id-event-ssl-vpn-user-ssl-login-fail).
- [Traffic CEF examples](https://docs.fortinet.com/document/fortigate/7.0.4/fortios-log-message-reference/949981/traffic-log-support-for-cef), [IPS CEF examples](https://docs.fortinet.com/document/fortigate/7.0.4/fortios-log-message-reference/311596/ips-log-support-for-cef), [DLP CEF examples](https://docs.fortinet.com/document/fortigate/7.0.4/fortios-log-message-reference/223332/dlp-log-support-for-cef), [antivirus CEF examples](https://docs.fortinet.com/document/fortigate/7.0.4/fortios-log-message-reference/807724/antivirus-log-support-for-cef), [infected-file event](https://docs.fortinet.com/document/fortigate/7.0.4/fortios-log-message-reference/8192/8192-mesgid-infect-warning).
