# Filter and correlation-rule dictionary audit

Draft review only. No production deployment or merge is part of this change.

## Scope and authority

Reviewed all **36 filter configurations and 635 rule files**, including `.yaml` as well as `.yml`, from UTMStack v11 commit `6c3af7eba9c8ec9b5ede5101feb308b2fca661d0`. Compared them with both supplied data dictionaries, the ThreatWinds SDK schema, the filter/rule wiki, and vendor documentation for the disputed outcomes.

The checked-in plugins pin **go-sdk v1.1.31**. Tests use the alerts plugin's existing module and dependency versions. The current SDK snapshot (`36461913c164f40176a7da881706970e9fd7c84d`) has the same `plugins.proto`. Wiki snapshot: `c18b54bd5ea5a34abb0e690458d73f89835edd29`.

The SDK determines names and types; the wiki supplies conventions and transformation semantics. In particular: matching pipeline stages run sequentially; `rename` moves a field; grok/CSV targets are literal; JSON/KV extraction lives under `log`; event paths and alert paths differ. `afterEvents` is a supported alias for `correlation`. These were considered before identifying defects.

## Confirmed defects and changes

| Area | Before | Draft behavior |
|---|---|---|
| Filter configuration | Azure supplies a scalar `dataTypes`; generic/syslog/IBM use misspelled grok keys; Palo Alto has misnested grok steps and CSV `from` keys | Configurations decode against the real SDK, including strict unknown-key validation |
| Lost standard fields | Top-level `command`, `from.host`, `origin.hostname`, and incorrectly cased Palo Alto counters disappear during Event conversion | Use `origin.command`, `origin.host`, and the exact counter names; Palo Alto client/server counters belong to the originating side |
| Outcome vocabulary | Filters emit `accepted`, `failed`, `blocked`, `Succeeded` and other aliases | Emit `success`, `failure`, or `denied` for known outcomes; dependent rules retain legacy aliases where needed |
| Cisco ASA/FTD | Invalid IPsec receives are accepted; a scan report is treated as success; an ACL result is overwritten after its source value is mutated | IPsec failures remain failures, scan reports have no invented outcome, and ACL results are classified from preserved vendor values. Correct the invalid `lgreaterOrEqual` condition |
| FortiGate | Quoted denial can miss classification; resets are treated as policy blocks; security-profile blocks can coexist with policy acceptance | Classify after trimming, give explicit UTM denial precedence, and record closure separately from outcome |
| Sophos XG | A blocked request with a block-page HTTP 200 can become successful | Explicit denial wins; HTTP status is a fallback, and non-policy HTTP failures are failures |
| Palo Alto | Traffic outcome is conditional on unrelated status fields; `Submitted` becomes successful | Honor explicit network denial, preserve unknown/pending outcomes, and classify administrative completion separately |
| Suricata | Per-signature `allowed` is treated as success even when final verdict drops traffic; protocol presence can create `action=success`; priorities 1 and 3 are inverted | Final packet/flow denial wins; `allowed` alone proves no successful connection; explicit pass/established flow evidence is required. Priority 1 is highest |
| ESET | Event names are placed in `actionResult` | Keep event type separately and classify actual vendor result/action values |
| Windows/Linux | Account disable/delete/lock success is labeled blocked; NTLM zero hex status is failure; negative audit exit is written into unsigned `statusCode` | Correct event outcomes; retain signed exit in `log.exitCode`; put executable path in `origin.path` and working directory under `log` |
| Address/actor mapping | Kaspersky reporting-agent metadata overwrites source identity; CrowdStrike LocalIP overwrites an explicit SourceIp; Bitdefender source fallback overwrites attacker IP | Preserve reporting metadata separately, retain explicit source/attacker addresses, and map standard CEF side identities |
| IP values | Placeholders, hostnames and compound endpoints are stored as IPs | Preserve unparseable values under `log.unparsed*Ip`; split O365 IPv4-with-port and bracketed IPv6 endpoints without truncating bare IPv6; preserve the reported endpoint |
| Cloud standard fields | Azure region/account name/status text are treated as country/hostname/connection status; AWS and GitHub actors remain only in vendor fields | Keep Azure metadata under appropriate `log` fields; promote known AWS/GitHub/Windows actor fields without consuming fields used by rules; handle Azure's common top-level `resultType` |
| Severity/protocol | Multiple severity scales and numeric JSON protocol values enter standard string fields | Map known priorities to wiki severity values; retain Bitdefender CEF priority for threshold rules; translate known numeric IP protocols and preserve unknown numbers separately |
| Rule grouping | `origin.*` and `*.hostname` are used on Alerts; raw grouping fields lack `lastEvent.` | Use mapped `adversary.*`/`target.*`, exact `host` names, and the documented `lastEvent.*` prefix. Respect each rule's adversary setting |
| Alert plugin | Documented `lastEvent.*` cannot resolve on the wire Alert; non-scalar values can enable a name-only search | Resolve `lastEvent` to the same final event indexed by `newAlert`; skip arrays/maps unless a scalar member is selected |
| Rule predicates | Two GCP rules call `oneof`; FortiWeb has invalid CEL regex escapes; Suricata uses unsupported integer `safe` defaults | Correct spelling/escaping/overloads; use actual EVE flow age and support raw/sanitized counter names |
| O365 correlation | A rule titled successful password guessing triggers on failure and references absent `log.clientIP`; password-spray history counts unrelated activity | Trigger successful guessing on `UserLoggedIn` success, correlate prior `UserLoginFailed` by standard user/IP, and constrain spray history to the triggering failure action/result |
| Other rule consumers | Anti-phish policy rule reads renamed O365 fields; Bitdefender phishing rule matches blocked pages; protocol comparisons disagree with source output | Use surviving standard fields, match Bitdefender `reportOnly` for the rule explicitly describing an unblocked page, and fix protocol comparisons |

The outcome convention describes the action being logged. A firewall policy allowing a packet is **not proof of a completed TCP handshake**. Successful account disabling is the success of an administrative operation. An absent outcome must not be interpreted as success by downstream detection.

## Validation

- Strict SDK decoding passes for all **36 filters and 635 rules**. Baseline had five configurations with schema/unknown-key diagnostics; Azure's scalar `dataTypes` is a hard decoding failure.
- All filter/rule conditions compile using the SDK's actual CEL implementation. Five baseline compilation defects are corrected. Missing nested fields in a deliberately empty compile-check sample are not treated as parser defects.
- **124 synthetic normalization cases pass across 17 filters**, with positive and negative detection-predicate assertions, final SDK Event conversion, and before/after outputs retained in the local audit report.
- Alerts plugin unit tests cover standard side fields, `lastEvent` selection, nested array members, missing/empty events, and rejecting non-scalar grouping terms.
- `go test ./...` in `plugins/alerts` and `git diff --check` pass. Regression fixtures and tests are included in this draft.

Run from the repository root:

```sh
cd plugins/alerts
go test ./...
```

The normalization test is explicitly a **model of documented normalization steps**, supplied with synthetic extraction results. It uses the real SDK YAML, CEL, casting helpers and protobuf finalization, but skips JSON/KV/XML/CSV extraction, complex grok, timestamp reformatting, and dynamic plugins. The single greedy copy pattern is modeled. It does not exercise the closed parsing engine, OpenSearch historical searches, the threat-intelligence feed engine, or customer environments. Predicate assertions test trigger conditions, not full historical alert generation.

## Choices preserved and remaining validation

- `log.*` is open. The absence of an explicit writer is not proof that JSON/KV data can never supply a field. Empty grok field names used for separators are valid and are not flagged.
- Numeric strings accepted by SDK comparisons/protobuf are not reported as type defects. Integral floating-point JSON values are not automatically invalid unsigned integers.
- The dictionaries disagree about protocol casing and some file/email meanings. Existing textual protocol casing and vendor `action` vocabulary are preserved, except concrete consumer mismatches. A platform-wide action migration would require coordinated changes to many more rules.
- No blanket `adversary: origin`/`target` reversal was applied to endpoint detections. An agent, a compromised endpoint, and a remote attacker are different roles; event-specific evidence is needed for further changes.
- The macOS rule's impossible `system.hostname` path is corrected to `origin.host`. The shipped macOS filters do not themselves supply a host identity, so its availability from upstream metadata still needs a representative event. Missing identity does not justify broadening the search across hosts.
- Review custom customer rules that match legacy `actionResult` or numeric `severity` values before any eventual rollout. Included rules are updated, and Bitdefender retains its CEF priority under `log.cefSeverity`. Short correlation windows may span historical/new outcome spellings during migration.
- Confirm filter ordering/extraction with representative raw logs and verify the resulting alerts in a staging engine. The reported customer threat-intelligence noise has **not** been reproduced against a live customer instance; these fixes address demonstrated causes, not a measured reduction in that customer's alert volume.

## Coverage by filter

All entries received schema, literal-value and condition review. “No edit” means no confirmed change in this draft, not certification against every vendor log variant.

| Filter | Disposition | Synthetic normalization cases |
|---|---|---:|
| `antivirus/bitdefender_gz.yml` | Corrected | 8 |
| `antivirus/deceptive-bytes.yml` | Corrected | 0 |
| `antivirus/esmc-eset.yml` | Corrected | 2 |
| `antivirus/kaspersky.yml` | Corrected | 2 |
| `antivirus/sentinel-one.yml` | Corrected | 0 |
| `aws/aws.yml` | Corrected | 2 |
| `azure/azure-eventhub.yml` | Corrected | 11 |
| `cisco/asa.yml` | Corrected | 12 |
| `cisco/cs_switch.yml` | Corrected | 0 |
| `cisco/firepower.yml` | Corrected | 12 |
| `cisco/meraki.yml` | Corrected | 0 |
| `crowdstrike/crowdstrike.yml` | Corrected | 2 |
| `fortinet/fortinet.yml` | Corrected | 6 |
| `fortinet/fortiweb.yml` | Corrected | 0 |
| `generic/generic.yml` | Corrected | 0 |
| `github/github.yml` | Corrected | 0 |
| `google/gcp.yml` | Corrected | 13 |
| `ibm/ibm_aix.yml` | Corrected | 0 |
| `ibm/ibm_as_400.yml` | Corrected | 0 |
| `json/json-input.yml` | No edit | 0 |
| `linux/linux.yml` | Corrected | 3 |
| `macos/macos-syslog.yml` | No edit | 0 |
| `macos/macos.yml` | No edit | 0 |
| `mikrotik/mikrotik-fw.yml` | Corrected | 0 |
| `netflow/netflow.yml` | Corrected | 0 |
| `office365/o365.yml` | Corrected | 14 |
| `paloalto/pa_firewall.yml` | Corrected | 4 |
| `pfsense/pfsense_fw.yml` | Corrected | 2 |
| `sonicwall/sonic_wall.yml` | Corrected | 0 |
| `sophos/sophos_central.yml` | Corrected | 0 |
| `sophos/sophos_xg_firewall.yml` | Corrected | 5 |
| `suricata/suricata.yml` | Corrected | 18 |
| `syslog/syslog-generic.yml` | Corrected | 0 |
| `utmstack/utmstack.yml` | No edit | 0 |
| `vmware/vmware-esxi.yml` | Corrected | 0 |
| `windows/windows-events.yml` | Corrected | 8 |

## References used

- [SDK v1.1.31 schema](https://github.com/threatwinds/go-sdk/blob/v1.1.31/plugins/plugins.proto), [CEL implementation](https://github.com/threatwinds/go-sdk/blob/v1.1.31/plugins/cel.go), [correlation implementation and alias normalization](https://github.com/threatwinds/go-sdk/blob/v1.1.31/plugins/rules.go).
- [Filter implementation](https://github.com/threatwinds/go-sdk/wiki/Implementing-Filters), [step reference](https://github.com/threatwinds/go-sdk/wiki/Filter-Steps-Reference), [standard schema](https://github.com/threatwinds/go-sdk/wiki/Standard-Event-Schema), [rule implementation](https://github.com/threatwinds/go-sdk/wiki/Implementing-Rules), [advanced features](https://github.com/threatwinds/go-sdk/wiki/Advanced-Features), and [CEL overloads](https://github.com/threatwinds/go-sdk/wiki/CEL-Overloads). Architecture, components, examples, best-practice and troubleshooting pages were also reviewed.
- [Cisco invalid IPsec messages](https://www.cisco.com/c/en/us/td/docs/security/asa/syslog/asa-syslog/syslog-messages-400000-to-450001.html), [Cisco scan/shun messages](https://www.cisco.com/c/en/us/td/docs/security/asa/syslog/asa-syslog/syslog-messages-722001-to-776020.html).
- [FortiOS 7.4.9 log reference](https://fortinetweb.s3.amazonaws.com/docs.fortinet.com/v2/attachments/514718ad-8f65-11f0-9bfd-6af4c3636dc7/FortiOS_7.4.9_Log_Reference.pdf), [Palo Alto traffic fields](https://docs.paloaltonetworks.com/ngfw/administration/monitoring/use-syslog-for-monitoring/syslog-field-descriptions/traffic-log-fields), [Sophos syslog reference](https://docs.sophos.com/nsg/sophos-firewall/19.0/syslog/index.html).
- [Suricata EVE format, verdicts and flows](https://docs.suricata.io/en/suricata-8.0.3/output/eve/eve-json-format.html), [Bitdefender syslog events](https://www.bitdefender.com/business/support/en/77212-237090-syslog-events.html).
- [Windows account-disabled event](https://learn.microsoft.com/en-us/previous-versions/windows/it-pro/windows-10/security/threat-protection/auditing/event-4725), [Azure resource-log schema](https://learn.microsoft.com/en-us/azure/azure-monitor/platform/resource-logs-schema), [O365 audit properties](https://learn.microsoft.com/en-us/purview/audit-log-detailed-properties).
