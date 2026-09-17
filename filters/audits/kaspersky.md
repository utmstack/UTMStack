# Kaspersky native parsing and CEF consumer review

Security Center native RFC5424 records were reaching the Kaspersky index with only
`log.priority` parsed. The CEF-only filter did not recognize their `event@23668`
structured data. This revision adds that observed format and repairs the legacy CEF
producer/consumer contract. It replaces the review submission in historical PR #2594.
The authoritative schema is ThreatWinds go-sdk **v1.1.31**, as pinned by v11.

## Evidence and source identity

A completed fleet index inventory identified one carrier, `isspol.utmstack.com`.
Its fresh retained count was 2,981; a separate 30-day raw-token aggregation found
2,738 records containing `KLSRV_HOST_STATUS_CRITICAL` and 126 without that token.
These are different query scopes and captures, not a class-complete population estimate.
Ten deduplicated representatives were selected using latest records, dataSource strata,
and a supplementary noncritical-token stratum. Six describe critical device status;
four describe an unmanaged device. All ten contain native `event@23668` structured
data and neither a CEF nor a LEEF header.

Examples used privately:

- `152e263c-ab27-46b1-9e46-b661e7fec8d9`, `v11-log-antivirus-kaspersky-2026-09-17`.
- `f062b03a-09d1-4e66-9aa9-925846217bae`, `v11-log-antivirus-kaspersky-2026-09-02`.

Both are on `isspol.utmstack.com`. Customer payloads and device identities are excluded
from this PR. The deployed filter SHA-256 was
`4200885fb3762a61bea1b8fd78d9625037e55dbdf0eadbfa5857b535acda1bb9`.
Field-cap inspection covered 31 retained indices; the requested parsed fields exposed
only priority. The shipped integration guide requests CEF. The evidence establishes an
input-format mismatch at this instance, not a universal vendor format change.

## Mappings and outcomes

Native parsing checks the complete envelope and quoted structured-data syntax before
promoting values. `hip` and `hdn` describe the managed device and populate `target.ip`
and `target.host`; reporting-host and management-server fields stay under `log.ksc`.
No actor is inferred from the relay. Original time populates `deviceTime` when it is
RFC3339, and syslog priority supplies the documented SDK severity values. Event type,
display name, reason and administrative group remain vendor data. An administrative
device group is not assumed to be a security group. Native health events establish
neither an action result nor a successful connection.

CEF headers must start at the beginning of a bare record or supported syslog envelope;
message text cannot become a second header. Vendor/product header strings remain open
for compatibility. After generic KV, 63 consumed vendor keys are reconstructed at CEF
boundaries and protected header/derived fields are restored or cleared. Escaped equals
signs cannot manufacture source IPs. Original CEF values remain available under `log`.
Source/destination IPs are validated before promotion and geolocation; unspecified IPv4,
expanded IPv6 and IPv4-mapped zero addresses stay vendor data. CEF hosts, users, MACs,
ports and protocol map to their standard side when usable. Protocol spelling is retained.
General CEF/structured-data unescaping is not implemented by an invented filter step;
escaped identity text remains vendor data.

Explicit allow decisions map to `success`, explicit block decisions to `denied`,
case-insensitively. Redirect, deletion, termination, quarantine and unknown CEF action
names alone no longer imply a connection outcome. Their original action remains visible.
These changes do not assert that an allowed request established a network session.

## Consumers and history

All 19 shipped rules now explicitly consume the CEF format they interpret. Native health
messages containing malware, task, WMI or connection-related words cannot become attack
candidates. No new health or native-malware rule was invented from this sample.

CEF consumer repairs include standard source/destination fields, quoted-path `safe`
lookups, normalized block outcomes, `log.msg` rather than an unproduced `log.message`,
and the actual standard `actionResult`. The C2 rule requires a C2 indicator; generic
`NetworkThreat` alone is insufficient. Explicit blocks are excluded for all tested
spellings, while unknown outcomes are described as unknown, not successful C2.

Four histories preserve their existing thresholds/windows: lateral movement 3/2h,
suspicious network activity 5/30m, ransomware 3/10m, exfiltration 5/30m. Each counts
only candidates for that rule, from the same collector and source-identity namespace.
Source identity prefers a valid IP, then a real source hostname; it never fabricates an
IP. Network history also matches the destination. Exfiltration retains its existing
`NetworkThreat` history restriction. Ordinary endpoint activity cannot fill a threshold.
Input-supplied candidate markers are cleared before deriving them.

Filter and rules must ship together. New history markers need up to **2 hours** of
warm-up; old events do not contain them. Saved searches depending on remediation actions
being `denied` must use the retained vendor action or explicitly review that migration.
Native IP/host/time/severity coverage will increase. All 19 YAML consumers and the shipped
integration guide were inspected; customer-created dashboards/searches were not exported
or changed. Original CEF fields are retained to reduce compatibility losses.

## Verification and limits

- **82 synthetic raw cases** cover native quoting, malformed input, embedded CEF,
  reporting/managed-device roles, missing identities, semantic zero IPs, ports, severity,
  outcomes, CEF compatibility and positive/negative predicates for every rule.
- Actual SDK v1.1.31 performs CEL evaluation, Event conversion and all four historical
  query/threshold suites against a local isolated mock. Tests cover threshold boundaries,
  time windows, collector/source/destination scope, unrelated populations and missing keys.
- Shared contract overlay: **112 passing test/subtest records**, no failures or skips,
  including the separate private ten-record replay.
- Private raw replay recovers target IP, target host and device time in **10/10** records;
  all 19 attack predicates remain false. No actor or action result is invented.
- The extraction tests model explicit YAML operations. They are not the closed
  EventProcessor, external geolocation, a live history query or observed alert creation.

The sample contains no CEF attack telemetry. CEF custom-slot meanings, attack-side
identity semantics and the precision of legacy text heuristics therefore remain
unverified against real vendor attacks. Synthetic compatibility is not proof of those
meanings or of reduced customer alert volume. Official support links redirected to
country/general support pages, so their HTTP 200 responses were not accepted as format
documentation. Other native event classes, other export envelopes, closed-runtime
behavior, parsing cost and alert volume require staging and vendor-format evidence.
No customer configuration was modified; no deployment or merge is authorized here.

Indexed `lastEvent.*` grouping depends separately on draft
[#2627](https://github.com/utmstack/UTMStack/pull/2627). Its fleet-wide deduplication
rollout and volume measurement must remain separate from this source review.

## Authority and attempted vendor reference

- [Pinned SDK schema](https://github.com/threatwinds/go-sdk/blob/v1.1.31/plugins/plugins.proto)
- [SDK field semantics](https://github.com/threatwinds/go-sdk/wiki/Standard-Event-Schema)
- [Filter steps](https://github.com/threatwinds/go-sdk/wiki/Filter-Steps-Reference)
- [Correlation rules](https://github.com/threatwinds/go-sdk/wiki/Implementing-Rules)
- [Kaspersky configured documentation entry](https://support.kaspersky.com/KSC/14/en-US/)
  (redirected; log-format content unavailable during this review).
