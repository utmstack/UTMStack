# FortiWeb parsing and detection contract

This draft fixes the FortiWeb filter and all seven existing FortiWeb rules, and adds
separate XSS, Known Exploits and Trojans detections. It targets UTMStack `v11`.
No customer configuration or indexed data was changed. Public fixtures are synthetic.

## Confirmed defects

Read-only inspection of retained production events and the loaded configuration confirmed:

- Space-delimited KV truncates quoted classifications: `sub_type="SQL Injection"`
  becomes `log.subtype="\"SQL"`; `main_type="Signature Detection"` is also truncated.
  The vendor keys are sanitized to `subtype`, `maintype`, `attacktype`, etc.; rules
  querying `log.attack_type` do not read those fields.
- The old message parser stops at embedded `key=` text, including a policy name
  inside `msg`. A truncated message cannot reliably identify the attack.
- The rules compare lowercase actions against observed `Alert` and `Alert_Deny`.
  `Alert_Deny` and `Period_Block` were not normalized to `actionResult=denied`.
  A detection with `Alert` is not evidence of successful exploitation.
- The deployed web-shell expression fails SDK CEL compilation because of invalid
  string escapes. Its old generic upload-restriction alternative also does not
  establish web-shell activity.
- Several history queries counted every event from an address, including unrelated
  traffic and requests to another target. Dedicated XSS and Known Exploits rules
  were missing; the retained Trojans class also warrants a discrete detection.

Source/destination direction in the inspected events was already correct. This is
not evidence of an IP inversion. YAML newline escaping, `afterEvents`, custom
`log.*` names and empty noncapturing grok field names are supported, not defects.

## Producer and consumer changes

The filter keeps KV for other vendor metadata and re-extracts authoritative fields
with quote-aware RE2 token boundaries. Spaces and embedded assignments survive;
assignments inside quoted request/message values cannot supply IPs, decisions or
classifications. The existing sanitized `log.*` names remain available.

| Vendor value | Output and use |
| --- | --- |
| `src` / `src_ip`, `dst` / `dst_ip` / `dest_ip` | `origin.ip` / `target.ip`; rules map origin to adversary |
| Source/destination port aliases | SDK numeric ports; invalid values retained under `log.unparsed*Port` |
| Invalid/unspecified IP | Retained under `log.unparsed*Ip`, removed before geolocation |
| `sub_type`, `main_type`, `attack_type` | Full `log.subtype`, `log.maintype`, `log.attacktype` for detection |
| `signature_id`, `signature_subclass`, OWASP fields | Full existing sanitized vendor fields |
| `action` | Original decision in `log.action`; attack/traffic activity becomes `action=http_request` |
| Explicit deny/block variants | `actionResult=denied`; `Alert` and unknown decisions leave the outcome unset |
| `proto` | Recognized names/identifiers become lowercase `protocol`; original remains in `log.proto` |
| `severity_level` | Low → info, Medium → warning, High → error; critical/debug retain their standard names; original remains |
| `HTTP_url`, `HTTP_host` | Relative URL goes to `target.path`; full HTTP(S) URL to `target.url`; Host header stays in `log.httphost` |

All ten rules read the fields this filter produces and accept the observed decision
casing. SQLi, XSS, known exploits, malware, authentication bypass and SSRF/web-shell
signatures use discrete triggers, following the wiki's high-fidelity detection guidance.
SQLi/XSS prefer classification over message text. SSRF needs explicit SSRF evidence;
`localhost` alone does not qualify. Authentication notification policy names and
account lockouts do not establish bypass. Generic script uploads do not establish a web shell.

Generic attacks, upload-policy violations and selected medium-or-higher HTTP/OWASP
violations keep history thresholds. They constrain source, target, attack subtype,
event type and vendor decision, and use alert-side deduplication. The filter's
`log.fileUploadViolation` classification also constrains upload history to eligible
violations, without an exact-term dependency on potentially long message strings.
OWASP history additionally constrains severity, main type and OWASP category.
Routine GEO/IP-reputation blocks, low-severity missing Content-Type, duplicate
parameters and low-severity information disclosure do not match these rules.

## Verification and limits

- Read the filter/rule wiki and SDK v1.1.31 implementation; checked standard fields
  against the supplied dictionaries and SDK schema. The wiki's lowercase protocol
  convention resolves the older dictionary's conflicting uppercase recommendation.
- `go test ./...` in `plugins/alerts` runs the standalone FortiWeb tests: **74 synthetic
  raw-log cases, each evaluated against all 10 rule predicates**, plus history and
  alert-side grouping contracts. Strict SDK YAML decoding, real CEL evaluation and
  final SDK Event conversion are exercised.
- With the shared runner from PR #2590 temporarily applied, all **96 subtests** pass.
  The technology manifest supplies an additional invalid-address normalization case.
- Read-only category aggregates were projected through the new filter/rule contract
  to check coverage of retained classifications. These are predicate candidates,
  not observed alerts or a full historical replay.
- `git diff --check` passes.

The raw contract harness models documented grok concatenation and transforms with
Go RE2. It deliberately omits KV, dynamic plugins and the closed EventProcessor;
all asserted/detection fields must therefore come from the quote-aware extraction.
It does not execute live OpenSearch correlation, deduplication or alert publication.
The standalone tests need no shared runner. Review shared alert grouping fix #2590
before rollout, since it affects actual grouping/deduplication behavior.

Before rollout, stage the filter and rules together and replay sanitized representative
payloads through the actual collector/engine. Confirm resulting alert IDs, blocked
versus monitored outcomes, history thresholds, ingestion delay, and burst grouping.
Existing indexed truncated fields are not repaired automatically. Custom consumers
of the old root `action` decision must migrate to `log.action`. Severity mapping and
noise thresholds are explicit policy choices for team review. No production alert
reduction or end-to-end live validation is claimed.

## References

- [Fortinet header/body field semantics](https://docs.fortinet.com/document/fortiweb/7.2.2/log-message-reference/578387/header-body-fields)
- [Fortinet attack log semantics](https://docs.fortinet.com/document/fortiweb/7.2.2/log-message-reference/445549/attack)
- [SDK schema](https://github.com/threatwinds/go-sdk/blob/v1.1.31/plugins/plugins.proto)
- [Filter steps](https://github.com/threatwinds/go-sdk/wiki/Filter-Steps-Reference)
- [Standard event schema](https://github.com/threatwinds/go-sdk/wiki/Standard-Event-Schema)
- [Rule implementation and trigger guidance](https://github.com/threatwinds/go-sdk/wiki/Implementing-Rules)
