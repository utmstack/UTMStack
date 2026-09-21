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

The initial bounded sample had correct source/destination fields. A later independent
raw-source search identified payload `src=` text contaminating indexed origin fields.
The strict authoritative extraction recovers the original header source; synthetic
XSS payload controls explicitly exercise that correction. This is payload contamination,
not a systematic source/destination inversion. YAML newline escaping, `afterEvents`, custom
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
| Invalid/unspecified IP | Retained under canonical vendor `log.src` / `log.dst`; rejected at the original rename before geolocation |
| `sub_type`, `main_type`, `attack_type` | Full `log.subtype`, `log.maintype`, `log.attacktype` for detection |
| `signature_id`, `signature_subclass`, OWASP fields | Full existing sanitized vendor fields |
| `action` | Original decision in `log.action`; attack/traffic activity becomes `action=http_request` |
| Explicit deny/block variants | `actionResult=denied`; `Alert` and unknown decisions leave the outcome unset |
| `proto` | Recognized names/identifiers become lowercase `protocol`; original remains in `log.proto` |
| `severity_level` | Low → info, Medium → warning, High → error; critical/debug retain their standard names; original remains |
| `HTTP_agent` | Complete quoted user-agent retained in `log.httpagent`; no SDK standard user-agent field |
| `HTTP_url`, `HTTP_host` | Relative request URL stays in `log.httpurl`; full HTTP(S) URL goes to `target.url`; Host header stays in `log.httphost` |

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

## Meeting follow-up and current evidence

The follow-up uses the same SDK authority as the initial incident review:
`plugins/alerts/go.mod` pins go-sdk v1.1.31, and names/types come from its protobuf;
the official wiki supplies field semantics. In particular, `Side.path` is a
filesystem directory, not a relative HTTP request URL. The new draft removes that
promotion, retaining `log.httpurl` and existing rule consumers. No rule depends
on `target.path`. Consumers of the earlier draft's relative `target.path` should
use `log.httpurl`; no synthetic scheme or Host-header identity is invented.

IP checks now occur at the original `log.src` / `log.dst` promotion, after quote
cleanup. The existing first-matching-token precedence across `src`/`src_ip` and
`dst`/`dst_ip`/`dest_ip` is unchanged. Invalid selected values remain in those
canonical vendor fields, rather than being promoted and moved to an auxiliary
`log.unparsed*Ip` field afterward. Such auxiliary fields are supported by the SDK;
the change concerns validation order and preservation of vendor values. Semantic CIDR exclusions reject all unspecified
representations, including expanded IPv6 and IPv4-mapped zero. Valid IPv6 and
mapped nonzero IPv4 retain their identities. The existing anchored envelope and
quote-aware extraction need no parser rewrite; the added tests also exercise
residual fields created by naive KV parsing before authoritative recovery.

The bounded current sample contains 24 distinct records from two instances
(12 per instance), selected by recent/subtype strata. Both deployed filter
copies have SHA-256
`d88a8909e39f84c70050270f553310b6a9bcf8abfff804e613e69c5b235d8bb6`.
Private document IDs, raw payloads and instance provenance remain in the private
evidence pack. Public fixtures are fabricated; this stratified sample is not a
population-rate estimate.

- All 24 stored source/destination IP pairs and ports agree with the raw values;
  no role inversion was observed. The model preserves the stored network fields.
- 23 of 24 stored user-agent values were truncated. All 24 complete values are
  now recovered and compared directly with their raw quoted value. User-agent
  stays vendor-specific because the SDK has no equivalent standard field.
- Eight explicit blocking decisions lack a stored standardized outcome; the
  model produces `actionResult=denied` for those eight. Existing native decision
  classifications are preserved. `Alert` and other decisions remain unset without explicit outcome evidence;
  the sampled `Erase` value is not assigned an unverified success meaning.
- All 24 request URLs are relative. All observed usernames are placeholders and
  do not establish a real user identity. Signature CVE data is either a
  placeholder or a multi-value string; it remains under `log.signaturecveid`.
- Native event time is retained as a 19-digit Unix-nanosecond value, while indexed
  `deviceTime` equals ingestion `@timestamp`. A supported epoch-nanosecond
  conversion is not documented in the available filter wiki. Preserve the native
  field rather than putting that integer string into the timestamp schema.
- HTTP response codes and request/response byte fields were not present in this
  sample. Vendor documentation of such fields does not prove a mapping defect
  in these observed events.

## Quoted-URL and nonblocking detection follow-up

The filter now recovers HTTP URL, Host and user-agent values when a native quoted
request URL contains bare quotes or ends in a backslash. The fallback requires a
strict prefix through the actual `http_url` key, the final complete native HTTP
field sequence and a strictly parsed remainder through the end. It runs only if
strict extraction did not obtain a user agent. It changes no IP, action, severity
or classification extraction. Fake `src`, decisions and metadata blocks inside
payloads cannot replace the authoritative security header in the regression cases.

The fallback deliberately does not recover message, signature, attack-type or
OWASP fields through this relaxed URL boundary. Such fields after a malformed URL
can remain absent. The inspected affected Generic/XSS records retain their header
subtype and match the intended predicates, but this does not establish coverage
for other rules that require tail-only evidence. Keep the original raw event for
investigation; complete recovery of malformed vendor tails is not claimed.

The new fixtures explicitly cover Medium Generic Attacks(Extended) with `Alert`,
`Alert_Deny` and case variants, each using the actual SDK history executor below
and at the three-event threshold and outside the 15-minute window. The Generic
rule has no severity gate; both monitoring and blocking decisions are eligible.
Additional controls cover malformed URL delimiters, false native-field blocks,
missing boundaries, multi-line URLs, missing genuine source IPs and five fabricated
XSS request shapes containing HTML `src=` attributes.

Final Event validation now uses strict protobuf JSON decoding. Private replay
supports explicit raw-authoritative network expectations for records whose indexed
source was itself corrupted; it no longer assumes such corruption must be preserved.
All private evidence stays outside this branch. Public fixtures contain invented
addresses, hosts, payloads and identifiers.

## Verification and limits

- The standalone suite now has **125 fabricated raw cases**, each evaluated against
  all ten actual SDK rule predicates, with strict configuration decoding and
  final SDK Event conversion. Original incident fixtures remain covered.
- Raw tests now include observed space-delimited KV behavior, so fake fields
  inside quoted messages are created before the filter must clear/recover them.
  They test malformed quotes, absent original IPs, alias precedence, unspecified
  representations, quoted IPs, embedded headers, user-agent recovery and relative
  URLs. External geolocation is not called; every address reaching that step is
  checked as valid and non-unspecified.
- All three threshold rules execute actual SDK history requests against a local
  loopback HTTP mock: count-minus-one/count, inside/expired windows, identity and
  classification mismatches, benign raw histories and all 19 required placeholder
  deletions are covered. Existing counts/windows remain 3/15 minutes for Generic
  Attacks, 3/30 minutes for upload violations and 5/15 minutes for OWASP. Missing
  required values are rejected by the trigger and by direct SDK placeholder
  resolution. The seven discrete detections need no history queries.
- The optional 24-document private replay preserves source/destination IPs,
  ports and ingress source; all ten predicates execute without CEL errors.
  Modeled predicate candidates are one Known Exploits, one SQLi, one OWASP and
  two Generic Attacks events. These are neither observed alerts nor proof that
  live history thresholds were met; no matches for another class is not a defect.
- The shared manifest lists the filter and all ten consumers, with one explicit
  negative control showing that absent extracted identities match none of the ten
  predicates. This normalization-only control does not prove raw extraction or
  IP validation; the standalone raw suite supplies that model coverage.
- The earlier shared-runner overlay passed 133 test/subtest records with private
  replay skipped. The updated overlay with shared grouping PR #2627 at `7010b8b5`
  passes **175 test records with no failures or skips**, including the private replay.
- The updated standalone `go test ./...` in `plugins/alerts` passes 140 test records
  with the bounded private replay enabled, including complete user-agent comparisons
  and corrected raw-authoritative source identities. `git diff --check` passes. SDK history tests
  use localhost only; no customer endpoint is contacted by the test suite.

The raw harness is an explicit offline model of documented grok concatenation,
Go RE2, observed KV splitting and filter transforms. It does not run the closed
EventProcessor, external enrichment, live OpenSearch, deduplication or alert
publication. Actual history windows use a processing-time lower bound on
`@timestamp`, not strict event-time sequencing. Classification markers on older
indexed documents are not backfilled; stage the filter and rule contracts
together and allow applicable windows to warm up. The standalone tests need no
shared runner. Review shared alert grouping fix #2627 before rollout, since it
affects actual grouping/deduplication behavior. Added user-agent recovery is one
bounded field; assess parser throughput with representative message sizes in
staging. Cross-appliance history/grouping isolation is not established by this
sample; the current rules retain their existing source/target/class scope.

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
