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
never enter `origin.ip`. Authentication correlation chooses a real source IP,
then workstation, then actor account, recorded in `log.authenticationSource`
and `log.authenticationSourceType`. The account fallback identifies an account,
not a network client. Every affected history search scopes that identity by
its kind, domain scope, the agent's `dataSource`, and event code. Brute-force rules also
require the target account; Kerberos searches constrain ticket encryption and,
for AS-REP, preauthentication type. No shared placeholder is an identity. A username-only fallback requires its
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
correlation now uses the same identity selection. Filter-derived `log.authenticationCandidate.*` markers repeat each exact
trigger predicate so benign events with the same event code cannot satisfy the
historical threshold. The tests assert marker/predicate parity. Success after
failures searches the failed-logon marker. The update preserves each rule's
existing count and time window. It does not claim that the existing
Golden/Silver Ticket heuristics prove forged tickets.

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
- 53 sanitized raw JSON fixtures exercise valid IPv4/IPv6, missing/placeholder
  addresses, host/account fallback, valid/invalid ports, host roles and time.
- 44 positive predicate cases cover all seven changed correlation consumers.
  Negative identity cases also compile/evaluate all 38 shipped Windows rules.
- The real SDK executes historical requests against a local mock OpenSearch
  server, including mapping resolution, placeholder expansion, query creation,
  time/count boundaries and separation of different sources, identity kinds,
  collectors, account domains, event types and non-candidate history. The old missing-IP regression is reproduced with
  the SDK, without a customer connection.
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
