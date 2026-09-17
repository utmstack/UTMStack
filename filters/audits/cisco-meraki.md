# Cisco Meraki parsing and rule contracts

This source draft targets UTMStack v11 and revises one filter and its seven rules.
The meeting follow-up is based on SDK v1.1.31, pinned by `plugins/alerts/go.mod`,
and the official SDK field and filter/rule documentation. It preserves ingress
`dataSource` and the physical source/destination roles in network events. It does
not deploy configuration, generate customer telemetry or claim that an alert fired.

## Evidence and its limits

The current carrier investigation found nine distinct records without recognizable
Meraki envelopes under `firewall-meraki`: five ISO syslog records from unrelated
programs and four other unrecognized network/syslog records. They lack normalized Meraki fields. The deployed
filter SHA-256 is
`b39216313c9e141d7bb9aef93aafc97f044e482ea7a013bc62ea4bb4ecf64e18`.
Bounded searches over the retained 30-day window found no strong documented
Meraki class signatures. That search is not an exhaustive proof that Meraki
telemetry never arrived. The routing mismatch is a separate ingress-classification issue with its upstream
producer unresolved;
this patch does not reinterpret unrelated syslog as Meraki data.

Private raw records, document IDs and instance provenance remain outside the
repository. All public positive fixtures are fabricated from official Cisco
examples, not customer records. Positive parsing and detection results below are
therefore documented-format compatibility checks, not measured recovery of
customer Meraki events. The reviewed prior source draft is
`db4ae93a194009ca05631245a6316b6c87134f2c`.

## Producer corrections

The old raw header paths require a Cisco calendar-time wrapper. The documented
Meraki native envelope instead starts with fractional epoch, device name and
an event group. The new anchored paths accept that envelope, a bounded RFC3164
wrapper, and the documented standalone Air Marshal form. Group boundaries are
explicit: device or message text cannot become part of a native event category.
The old greedy group copy was not overwriting an existing subtype; it was the
sole producer of `log.eventType`.

`log.merakiGroup` retains the outer event group and `log.merakiType` retains the
header device name. `log.eventType` exposes the canonical category; nested Air
Marshal messages use `airmarshal_events`. `log.type` holds the events/Air Marshal
classification, while `log.alertType` holds the security-event class. These fields
now match the rule consumers. Class-prefix recognizers are anchored, and quoted
field recovery cannot treat message text as an independent identity or decision.
The full body remains available as `log.message`.

Legacy body handlers remain for event families outside the bounded new handlers;
legacy header/control-field rewrites cannot overwrite a recognized native header.
Representative DHCP, VPN connectivity, cellular status, STP, disassociation and
URL formats have regression fixtures. Not every historical vendor variant has
been exercised, and the model's built-in legacy regex aliases are approximations
of the closed executor's patterns. New native patterns are explicit in YAML.

Three review controls reproduced false fields before correction: a quoted VPN
fragment produced a peer IP and success, quoted disassociation field names
produced success, and an embedded legacy date header reclassified application
text as a flow. Native VPN/disassociation now consume only their top-level
tokens, and both legacy header patterns require the start of the raw record.
Genuine VPN connectivity fields, full disassociation metadata and both leading
legacy header forms have positive controls. Disassociation timing/radio fields
remain metadata; their presence alone does not establish authentication success.

| Vendor information | Result |
|---|---|
| Source/destination addresses | Captured and cleaned under vendor fields, validated at promotion, then written to `origin.ip`/`target.ip`. Invalid/unspecified values stay under `log.sourceIp`/`log.destinationIp`; original native `log.src`/`log.dst` also remain. |
| IPv6 and endpoint ports | Valid IPv6 ending in `::` is not truncated by colon cleanup. Semantic zero CIDRs cover alternate and IPv4-mapped unspecified forms. Ports retain numeric bounds before SDK conversion. |
| Flow/firewall decision | Documented `flows`, `firewall`, `cellular_firewall`, `vpn_firewall` and `l7_firewall` groups are recognized. Explicit allow/deny/block becomes success/denied as a policy outcome, not proof a connection completed. |
| IDS / AMP enforcement | IDS `decision=blocked` and AMP scan `action=block` produce denied. Mere IDS observation and retrospective AMP `action=allow` do not manufacture execution success. |
| AnyConnect authentication | Only the documented auth-success/failure classes extract `Peer IP` from their structured message. Generic VPN negotiation or text mentioning failure does not become authentication evidence. |
| AMP hash and URL | Valid 64-hex `log.sha256` is copied to `target.sha256`, describing the file offered by the remote download endpoint; vendor hash is retained. A full URL maps to `target.url`. Threat names such as EICAR classifications are not filenames. |
| 802.1X identity | Documented identity maps to `origin.user`; physical switch interface numbers stay in `log.switchPort`, not transport `origin.port`. |
| Wireless device | Packet-flood device and Air Marshal source/destination MACs retain their documented roles. The observing appliance name is not an attacker hostname. |
| IDS `dhost` | Retained as `log.dhost`; it is no longer unconditionally assigned to the source MAC. Its all-direction endpoint semantics are not established by the available examples. |
| Auxiliary geolocation | Hostnames and zero addresses never enter geolocation. Results use distinct `log.serverGeolocation` / `log.localGeolocation` fields rather than children of scalar address fields. `ip_resp` is response timing and is not geolocated. |

The SDK uses string fields for outcomes and protocol. This change preserves the
existing protocol conversion policy and uses explicit outcome evidence. Fractional
epoch values remain vendor time metadata; a verified conversion into `deviceTime`
was not established here. Auxiliary geolocation field names and new consumer
aliases require review of saved queries during staging; private dashboards were
not enumerated.

## Consumers and identities

- AMP consumes exact documented scan/disposition-change classes and malicious
  disposition. Network source remains the downloading client; the rule's
  adversary side is the remote target serving the file. This is not proof that
  the client executed malware. Retrospective events remain eligible without IPs.
- AMP groups valid hashes using a `hash` namespace. Missing or malformed hashes
  retain the vendor value and use the actual ingress `Event.id` in an `event`
  namespace, keeping distinct records separate. The filter does not create an
  event ID; the input producer supplies it. Neither a usable hash nor an ingress
  ID is an explicit completeness negative, not a fabricated file identity.
- IDS accepts high/medium priorities 1 and 2, including documented legacy
  `ids-alerts` records containing only a source endpoint. A target IP is not
  invented to satisfy a rule.
- Rogue SSID and Air Marshal rules consume the actual Air Marshal class. The
  Evil Twin rule requires `ssid_spoofing_detected`, not a generic rogue SSID.
  The two rogue rules retain overlapping coverage and the existing RSSI threshold
  for team review; primary syslog RSSI units were not established. Cisco documents
  removal of these legacy rogue/spoofing messages in MR29 and later.
- Wireless intrusion consumes `device_packet_flood` with `state=start`; an end
  event or ordinary message containing attack words is not a new attack trigger.
- VPN uses exact AnyConnect auth failure, valid produced origin IP, and meaningful
  ingress/device identities. Its 10-in-15-minute history counts only the filter's
  matching failure marker for that source, collector and device. Generic traffic
  from the same IP no longer satisfies the history. Input markers are cleared.

History windows use the SDK's processing-time lower bound on `@timestamp`, not
strict event-time sequencing. Older indexed records have no new candidate marker;
allow the 15-minute history window to warm up when staging the filter and rules
together. Indexed `lastEvent.*` grouping requires the separate alert-foundation
fix; this source draft does not duplicate it.

## Verification

`go test ./... -count=1 -v` passes in `plugins/alerts` with the optional private
routing evidence enabled. `git diff --check` also passes.

- 85 fabricated raw cases exercise all seven rule predicates, native envelopes,
  quoted field/class injection, valid and unspecified addresses, documented
  legacy event families, decision semantics, hash/ID fallback and new mappings.
- Fourteen isolated auxiliary-geolocation controls test both address fields with
  valid IPs, hostnames and alternate unspecified representations. These are
  source-step controls, not legacy-envelope parsing proofs.
- Actual SDK CEL and Event/Alert serialization validate consumer identities,
  physical endpoint preservation, AMP actor direction and grouping separation.
- Actual SDK history executes against a loopback HTTP mock for the VPN threshold,
  count-minus-one/count, inside/expired windows, source/device/collector isolation,
  benign raw classes and missing placeholders. It does not query a customer.
- All nine private unrelated/unrecognized records pass as negative routing controls without
  invented Meraki identities, outcomes or candidates.
- The shared manifest contains three explicitly isolated normalization/empty
  identity controls. It does not prove raw parsing; the standalone raw suite does
  that within its stated model boundary.
- The final shared-runner overlay passes 121 test/subtest records, with the
  optional private replay skipped there and run separately with its evidence.

The closed EventProcessor, live OpenSearch, dynamic geolocation lookup and alert
publication are not run by this harness. No production false-positive reduction
has been measured. Before production approval, stage actual native Meraki payloads
through the collector/engine, inspect produced alerts and grouping, check parsing
cost at realistic event rates, and investigate the separate carrier-routing issue.

## References

- [Official Meraki syslog formats and examples](https://documentation.meraki.com/Platform_Management/Dashboard_Administration/Operate_and_Maintain/Monitoring_and_Reporting/Syslog_Event_Types_and_Log_Samples)
- [SDK v1.1.31 schema](https://github.com/threatwinds/go-sdk/blob/v1.1.31/plugins/plugins.proto)
- [Standard event semantics](https://github.com/threatwinds/go-sdk/wiki/Standard-Event-Schema)
- [Filter steps](https://github.com/threatwinds/go-sdk/wiki/Filter-Steps-Reference)
- [Rule implementation](https://github.com/threatwinds/go-sdk/wiki/Implementing-Rules)
