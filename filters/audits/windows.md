# Windows normalization and correlation review

This draft targets UTMStack `v11`. The SDK Event/Side protobuf in the pinned
`go-sdk v1.1.31` is the schema authority. The local dictionary is supporting
material, not an alternative schema.

## Correction after review

The first draft removed placeholder IPs after promoting them to `origin.ip`,
while authentication rules still required `{{.origin.ip}}` in historical
searches. The SDK returns an error when a placeholder is missing; it returns
before evaluating an `or` fallback. Restoring `-` as a standardized IP would
also aggregate unrelated sources under one placeholder.

The filter now validates `log.data.IpAddress` at the original rename. Invalid,
missing and unspecified addresses stay in their original vendor field and
never enter `origin.ip`. The unspecified-address guard uses CIDR membership,
not literal strings: expanded/compressed IPv6 zero and IPv4-mapped zero spellings
are rejected while the exact original vendor value is preserved. Tests cover
account and workstation fallback, no qualified fallback, and a valid mapped-IPv4
control. These alternate-spelling regressions are synthetic; no additional
customer occurrence is claimed. Authentication correlation chooses a real source IP,
then workstation, then actor account, recorded in `log.authenticationSource`.
The account fallback identifies an account, not a network client. Every affected
history search pairs that identity with the agent's `dataSource`. The brute-force
rule counts failures from one source whatever account they target; success after
failures also requires the same account. Kerberos searches use their candidate
marker, which already fixes the event code and ticket encryption. No shared
placeholder is an identity. A username-only fallback requires its
domain or an already qualified UPN; an unqualified username without a domain
is not treated as a safe correlation identity.

The native Windows collector sets `dataSource` from its hostname for every
record. The rules reject its documented `unknown`/empty sentinel so account
fallback cannot pool unidentified collectors. They do not require a redundant
`exists(dataSource)` check. The filter and all seven correlation consumers
must be deployed together. Existing indexed records do not have the new
correlation fields, so the updated history windows warm up after deployment.

Golden Ticket's historical query previously depended on `origin.host`, not
`origin.ip`; native Kerberos records can lack that workstation too. Its
correlation now uses the same identity selection. Filter-derived `log.authenticationCandidate.*` markers repeat the
trigger predicates that a history search cannot express (Kerberoasting, AS-REP
roasting, Silver and Golden Ticket, AD FS), so benign events with the same event
code cannot satisfy the historical threshold. The tests assert marker/predicate
parity. The two logon rules need no marker: their exact terms (event 4625,
`dataSource`, source and, for success, the account) already are the predicate.
The update preserves each rule's existing count and time window. It does not
claim that the existing Golden/Silver Ticket heuristics prove forged tickets.

## Simplification after production use (2026-09-30)

The first version also stored the kind of source (`log.authenticationSourceType`)
and a domain scope (`log.authenticationSourceDomain`), and marked every failed
and successful logon. Read-only counts on 27 v11.2.15 servers, which run this
filter unchanged, showed that none of it changed which events were counted:

- Across 26,418 source values seen in two days, no value ever appeared with two
  kinds. The domain scope separated five values; four of them were one domain
  written two ways (short and full name), so it split one account in two.
- The failed-logon marker only repeated the event code and the presence of the
  source and account, which the search terms already require. The success
  marker was written on every successful logon (about three million a day) and
  no rule read it.

Both fields and both logon markers are removed. Counting failures per source
instead of per source and account follows the rule's description and also
catches password spraying. Replaying two days of production failures, the
per-account version would have raised 100 alerts, up to 26 in one hour on one
server; the per-source version raises 69, at most 5 in one hour. The success
rule no longer searches history for computer accounts, which are 44% of
successful logons and whose passwords are machine-generated.

## Standard field promotion

The native agent emits `computer`, `timestamp`/`timeCreated`, and `data.IpPort`.
The filter promotes the event-producing computer to `target.host`, device
time to `deviceTime`, and valid remote ports to `origin.port`. WorkstationName
and the native NTLM `Workstation` alias describe `origin.host`; the recording computer must not overwrite
the remote workstation. Kerberos account names are also promoted to
`origin.user` while the existing `target.user` and vendor aliases remain
available to consumers. Vendor data without a standard counterpart remains
under `log`. `SubjectDomainName` and the authenticated account's domain are
promoted without mixing actor and target roles. A non-placeholder `ProcessName`
is copied to `origin.path`, preserving the vendor alias used by current rules.

NTLM status must also accept the agent's numeric zero: CEL `regexMatch` only
matches strings. The first draft's string-only check would have changed a
successful native 4776 event to failure. The revised check covers numeric zero
and hexadecimal zero strings, with nonzero numeric/string failure regressions.

The original fixes for successful account administration, NTLM status,
placeholder cleanup, and event-versus-alert grouping remain included.

## Validation and limits

- `windows_contract_test.go` is standalone and runs with `go test ./...` in
  `plugins/alerts`, without the shared test-runner PR.
- 77 sanitized raw JSON fixtures exercise valid IPv4/IPv6, missing/placeholder
  addresses, host/account fallback, valid/invalid ports, host roles and time,
  and the LSASS, certificate, AdminSDHolder, SMBv1, ransomware and loopback
  Remote Desktop rules.
- Positive predicate cases cover every changed correlation consumer. Negative
  identity cases also compile/evaluate all 48 shipped Windows rules.
- The real SDK executes historical requests against a local mock OpenSearch
  server, including mapping resolution, placeholder expansion, query creation,
  time/count boundaries and separation by every exact search term. A spray of
  failures against different accounts fills the brute-force threshold but not
  the success-after-failures threshold. The old missing-IP regression is
  reproduced with the SDK, without a customer connection.
- The shared manifest adds seven nonempty rule assertions to its normalization
  cases. Its runner is supplied by draft #2590.

JSON extraction and filter transformations are an offline model based on the
SDK sanitization utility and documented filter operations. These tests do not
execute the closed EventProcessor, a live OpenSearch cluster, alert creation
or delivery. Staging must compare real raw events, normalized results and
created alerts before rollout. No customer configuration was changed.

## Bounded live evidence

The deployed Windows filter was also read without modification. Its relevant
mappings matched the repository baseline: unguarded IpAddress promotion,
WorkstationName alone, no computer/port/process-path promotion, and numeric-zero
NTLM status handled through `equals`. Configuration SHA-256:
`c3a781cc3f2015c3b9e1cb66773da3463186ccec68518051fcf3be533f77cc1c`.

A read-only seven-day sample from three instances yielded 28 distinct records
for the requested authentication event codes. Fifteen had no usable source IP.
All 28 retained the native computer only under `log`, and all 28 had deviceTime
defaulted to ingestion time instead of the raw vendor timestamp. Three NTLM
records supplied `data.Workstation` without a standardized origin host; nine
records supplied a process path without `origin.path`. The private evidence
pack keeps the instance/document anchors without publishing customer payloads.

No missing-IP 4768/4769/4771 events were observed in this seven-day aggregate.
The report's broad claim that Kerberos placeholder IPs were observed is not
supported by this sample. The SDK's missing-placeholder behavior is reproduced
with synthetic Kerberos records; actual missing/placeholder IPs are confirmed
for local logon and credential-validation event types.

Replaying those 28 raw records through the offline filter model and actual CEL
produced two failed-logon candidates and five successful-logon candidates with
resolved history placeholders. This is not proof of historical thresholds or
created alerts: the sample is intentionally bounded and the live filter/rules
were not replaced.

## Sources

- [SDK schema](https://github.com/threatwinds/go-sdk/blob/v1.1.31/plugins/plugins.proto)
- [SDK correlation execution](https://github.com/threatwinds/go-sdk/blob/v1.1.31/plugins/rules.go)
- [Native Windows collector](https://github.com/utmstack/UTMStack/blob/v11/agent/collector/platform/windows_amd64.go)
- [Filter operations](https://github.com/threatwinds/go-sdk/wiki/Filter-Steps-Reference)
- [Standard event schema](https://github.com/threatwinds/go-sdk/wiki/Standard-Event-Schema)

Alert grouping for `lastEvent.*` still depends on the shared alert-grouping
fix in #2590, which requires its own staging comparison of alert counts.
