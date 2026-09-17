# Bitdefender GravityZone review

This replacement draft reviews one CEF filter and all 21 source rules against
ThreatWinds go-sdk v1.1.31, the official filter/rule wiki, vendor documentation and
35 bounded raw/indexed records from three instances. The previous PR is historical
review input. No customer configuration, production deployment or merge is included.

## Confirmed producer corrections

- All 35 raw records are native Bitdefender GravityZone CEF. The module is the first
  extension key immediately after the header pipe and matches all 35 indexed values.
  It is produced by extension parsing; the plugin does not need to inject it.
- Parse only recognized CEF envelopes at the start of a record. Reconstruct consumed
  fields at escaped CEF key boundaries so spaces are preserved and escaped key-like
  text cannot invent addresses, actions or classifications. Re-read the protected raw
  header after generic KV parsing so extension keys cannot overwrite its event/severity.
- Validate original IP fields before promotion and reject equivalent zero-address forms.
  `dvc` is the managed endpoint. Explicit Network Attack Defense attacker/victim fields
  and firewall source fields retain their documented roles. Three sampled incident
  `src` values had displaced the managed endpoint. Their exact CEF roles, and those of
  `spt`, are not established by the available mapping documentation; retain them under
  `log` instead of attributing them to the wrong endpoint.
- Prefer the consistently present `deviceExternalId` for correlation. The alternative
  product endpoint identifier differs in all 15 sampled records containing both values.
  Fallback namespaces distinguish computer ID, product endpoint ID, hostname and IP;
  they cannot accidentally count as different endpoints within the same namespace.
- Explicit block/quarantine actions produce `denied`. Deleted/disinfected/restored mean
  successful remediation only in antimalware/behavioral modules; `still present` means
  remediation failure. Report-only, ignored and no-action records do not assert failure
  or a successful connection. Task success/failure is scoped to the task's own fields.
- Keep infected-object categories (`file`, `process`, `boot`, etc.) in
  `log.BitdefenderGZMalwareType`; they are not SDK malware types such as trojan/ransomware.
  Preserve full original values while mapping process and file basenames, safe commands,
  account identifiers, hashes, domains, valid ports and firewall protocols. Scheme-less
  request hosts map to `origin.domain`; a URL is populated only when complete and safe.
- Preserve original timestamps. Promote calendar-valid RFC3339 values without replacing
  an existing ingress `deviceTime`. Five sampled task/inventory records have no raw time
  field; their stored time cannot be credited to this filter's timestamp extraction.
- The intentional CEF severity migration uses the SDK wiki vocabulary. CEF priority is
  retained in `log.cefSeverity`; high vendor priority alone is not proof of compromise.

## Rule contracts

All 21 consumers have positive and negative raw-model assertions. Five history rules
use exact candidate populations and collector/company/identity scopes. Input-supplied
markers are cleared before derivation. Malware histories exclude inventory and task
records; network history requires explicit denied outcomes and a valid source address.
Cross-endpoint counts are event counts on other endpoints, not distinct-host counts.

Sensitive task activity requires successful task status, a relevant task label, creator
and task type. Normal scan/update events no longer satisfy it. Task labels remain leads,
not proof that a console was compromised or a configuration was changed. Exclusion
alerts likewise require a successful exclusion-related task, not generic policy text.
Ordinary device-control blocks do not count as USB malware. Explicit fileless flags and
behavioral command evidence replace the assumption that every suspicious file is fileless.

The phishing rule describes report-only detection without claiming a page loaded or
credentials were submitted. Other descriptions distinguish detection, blocking,
remediation and confirmed compromise. Mining URL rules consume the domain, so a mining
keyword only in a URL path does not qualify. Grouping includes source and managed-endpoint
identity, with vendor/indicator details where available.

## Validation and limits

- 77 public synthetic CEF fixtures exercise native and bounded syslog envelopes, all
  consumers, escaping, header/marker injection controls, endpoint roles, original IP
  guards, namespaces, hashes, ports, timestamps and benign controls.
- Actual SDK CEL, configuration/Event/Alert serialization and placeholder handling are
  used. Five SDK history queries run against a local mock with below/at-threshold,
  expiry, scope-isolation, missing-placeholder and identity-fallback checks.
- The private 35-record replay retains all module values, changes three target IPs back
  to the managed endpoint, maps 24 explicit blocks to denied, adds seven file basenames
  and fixes/adds 15 process basenames. The seven matching rule predicates across 14
  records are candidates, not observed alerts or a population-rate estimate.
- Shared schema/manifest checks are run with the reviewed alerts foundation. The single
  manifest fixture is a normalization-stage negative control; the separate Go suite
  supplies the raw extraction and positive-consumer tests. These layers are not additive
  counts of unique scenarios. External geolocation and the closed executor are not run.

CEF backslash/equals escapes are retained exactly. The documented pipeline has no general
CEF unescape step: encoded Windows directories and commands are not silently published as
decoded standard values. Safe basenames and request domains still map; the original full
values remain in `log` and protected `raw`. Full decoded Windows paths, commands and URLs
need a supported decoder and staging verification. No scheme is invented for partial URLs.

Live evidence covers ten event classes. Other module/alias variants are compatibility
fixtures based on existing consumers and vendor field semantics, not observed CEF traffic.
The vendor syslog page documents JSON semantics rather than a complete CEF mapping table.
Legacy wrappers, parser throughput, closed-runtime behavior and alert-volume changes require
staging. New candidate markers require history warm-up (up to 24 hours); optional grouping
fields can still coalesce when absent. Filter and rules must ship together. The shared
alert-grouping fix has a separate fleet-wide rollout and is not included here.

Private record IDs, deployed configuration and replay results stay in the local evidence
pack. All three sampled deployed filters had SHA-256
`730a7caea1cad8f31f503ab3af775f2121691db9fdd811fa595ac219ad5a553b`.

## Sources

- [Versioned SDK schema](https://github.com/threatwinds/go-sdk/blob/v1.1.31/plugins/plugins.proto)
- [Standard fields and semantics](https://github.com/threatwinds/go-sdk/wiki/Standard-Event-Schema)
- [Filter operations](https://github.com/threatwinds/go-sdk/wiki/Filter-Steps-Reference)
- [Correlation semantics](https://github.com/threatwinds/go-sdk/wiki/Implementing-Rules)
- [Bitdefender event types](https://www.bitdefender.com/business/support/en/77212-237089-event-types.html)
- [Bitdefender syslog event fields](https://www.bitdefender.com/business/support/en/77212-237090-syslog-events.html)
