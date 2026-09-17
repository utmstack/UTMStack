# ESET JSON export and detection contracts

This draft targets the official UTMStack v11 repository. It reviews the ESET
filter and all thirteen existing consumers, with separate inbound/outbound botnet
handling. The target alerts module pins go-sdk v1.1.31; its protobuf and official
wiki define the standard. No customer configuration was changed and no alert was
created during this review.

## Evidence and limits

Fresh read-only counts over the retained ESET source index pattern returned zero
records on thirty v11 instances. One instance remained unreachable. The query
failures were not counted as zero; one stale container cache was refreshed before
its successful retry. No ESET customer raw/normalized pair or deployed filter
could be compared. This is documented-format and offline contract evidence,
not a measured recovery of live detections or reduction in false positives.

The substantive primary reference is ESET PROTECT On-Prem 13.1's JSON-export
specification. The prior 7.0 and 11.0 URLs now redirect to end-of-life notices and
cannot substantiate historical event variants. ESET explicitly says values vary
by endpoint application/version and its lists are not exhaustive. The public
fixtures use fabricated identities and documented field shapes; detection-label
examples are labeled synthetic, not claimed as observed vendor values.

## Producer corrections

- Recognize bounded bare JSON, the documented ERAServer RFC3164 wrapper and the
  existing RFC5424 family. Message text containing an embedded header is not an
  outer envelope. The syslog version is not the ESET event class.
- Extract the vendor JSON with SDK key sanitization, then expose `log.eventType`
  from `event_type`. Retain vendor severity, event detail and protected `raw` for
  investigation; remove the temporary JSON string after extraction. Consumers use structured detection attributes rather than
  searching a serialized object for attack words.
- Validate original address fields before standard promotion and geolocation.
  Invalid and semantically unspecified addresses remain vendor data. IPv6-only
  reporting endpoints and integer transport-port bounds are covered.
- Keep the management server's header identity separate from the managed endpoint.
  Firewall source/destination fields retain their physical roles; the `inbound`
  boolean determines which side receives the reporting endpoint's host, account
  and process. Missing direction does not justify pairing a local hostname with
  a remote IP. Filtered website records describe a local client and remote target.
- Map documented user, executable, operating-system, object URL and valid SHA-1
  fields where their roles are established. An Inspect alarm hash is not assumed
  to identify a file. Static management groups remain vendor fields rather than
  being mislabeled as security roles.
- Map unambiguous absolute executable/object paths into directory and basename
  fields. Keep registry targets, percent-encoded file URIs, remote authorities and
  URI query/fragment variants under vendor fields. Threat detection name/type map
  to malware metadata; a bounded firewall CVE signature maps to the destination
  vulnerability field. `occurred` and the examples' `occured` alias become UTC
  device time through a guarded temporary field, without reformatting an existing
  ingress timestamp as vendor text.
- Preserve native action detail. Explicit blocking means `denied`; audit results
  describe the audited operation; remediation errors mean `failure`. A handled
  detection, deleted file or event class does not establish successful malicious
  execution or a network connection. Vendor severity is retained while standard
  severity follows Information/Notice, Warning, Error and Critical/Fatal.

## Consumer corrections

The original consumers mixed a numeric header field with event classes, read a
removed JSON body or severity, and used management-relay identity for endpoint
history. The revised consumers use the produced canonical class, structured
vendor detection attributes and standard action/severity fields. All three
history rules count their own eligible candidates, scoped by collector and a
namespaced managed-endpoint identity; console login history also scopes the
account. Ordinary traffic or unrelated audit events cannot satisfy those counts.

Botnet consumers choose the remote side as adversary separately for inbound and
outbound events without reversing the filter's network fields. Missing direction
is an explicit coverage gap. The generic network rule excludes that detector
family to avoid duplicate alerts with contradictory adversary attribution.
Blocked tampering or exploitation is described as an
attempt, not proof that protection was disabled or compromise succeeded. Normal
policy changes and successful console tasks are not evidence of console abuse;
that consumer detects repeated console authentication failures.

The generic HIPS rule requires a security-specific operation or detector label;
an ordinary blocked operation under a restrictive policy is insufficient. Specific
registry and PowerShell rules describe blocked activity, which still requires
context and policy tuning rather than proving malicious intent.

The heuristic, machine-learning, botnet and behavior labels remain detection
heuristics over documented fields. Their completeness across product versions is
not established by this review. An unknown event class remains available under
vendor fields; unsupported historical classes are not declared impossible.

## Verification

- 155 fabricated raw cases exercise all fourteen consumers with explicit positive
  and negative expectations, SDK Event conversion and exact history-marker parity.
  Header, JSON, address, type, calendar, URI and direction boundaries are included.
  Malformed JSON is a model error control, not proof of the closed engine's error
  handling. The temporary JSON body is absent and protected raw remains unchanged.
- Independent SDK/model retests cover opposite-side botnet duplication, an ordinary
  HIPS policy block, an unrelated agent label, and file-URI query/fragment boundaries.
- The shared manifest contains two isolated empty/unknown-class negative controls.
  Its limited normalization model does not execute the anchored extraction used
  here; raw and positive-consumer proof comes from the standalone source suite.

- Actual SDK history queries pass against an isolated loopback mock for all three
  consumers: below/at threshold, inside/expired window, unrelated history, cleared
  forged markers, collector/endpoint/account separation, missing placeholders and
  UUID-to-host-to-IP fallback. Explicit denied/rejected login results are included.
- SDK Event/Alert wire-contract checks cover all fourteen positive consumers,
  produced grouping fields and physical/adversary direction. They assign sides
  according to rule metadata and do not execute the closed correlation service.
- The final shared-runner overlay passes 182 test/subtest records, with no failures
  or skips. `git diff --check` passes. These counts include parent test records and
  are not unique events or generated alerts.

Stage the filter and consumers together. New history markers require up to one
hour to populate the longest window. Thresholds count matching indexed documents,
not underlying occurrences represented by vendor aggregation counts. Endpoint
identity uses namespaced UUID, hostname, then valid reported IP; the relay hostname
is never its fallback. Optional detector/file grouping terms can be missing, in
which case the known endpoint forms a broader group. Shared grouping #2590 remains
a separate runtime dependency and rollout.

Before production approval, replay native records through the closed collector
and EventProcessor, verify actual stored field names, history and alert output,
measure parsing cost and alert volume, and review saved searches that consume
legacy fields/outcomes. Dynamic geolocation, production history and publication
were not exercised here. No source-only offline result establishes that an alert
fired or that every ESET application/version is covered.

## References

- [ESET JSON export fields, semantics and examples](https://help.eset.com/protect_admin/13.1/en-US/events-exported-to-json-format.html)
- [SDK v1.1.31 protobuf](https://github.com/threatwinds/go-sdk/blob/v1.1.31/plugins/plugins.proto)
- [SDK field sanitization](https://github.com/threatwinds/go-sdk/blob/v1.1.31/utils/fields.go)
- [Standard event schema](https://github.com/threatwinds/go-sdk/wiki/Standard-Event-Schema)
- [Filter steps](https://github.com/threatwinds/go-sdk/wiki/Filter-Steps-Reference)
- [Rule implementation](https://github.com/threatwinds/go-sdk/wiki/Implementing-Rules)
