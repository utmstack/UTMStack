# Sophos filter and rule contract review

Replacement review for #2617, covering Sophos XG Firewall and Sophos Central together.
The SDK version pinned by v11 (`github.com/threatwinds/go-sdk` v1.1.31), its protobuf schema,
and its official wiki define the standard. Vendor documentation defines source semantics.
This draft proposes code changes; it does not deploy them or establish production alert rates.

## XG Firewall

- Parse type, component and subtype independently. WAF records can omit subtype without
  losing their client/backend addresses, HTTP status or remaining fields.
- Recover consumed values at quote-aware key boundaries. Support plain, PRI, RFC3164 and
  RFC5424 envelopes. Preserve vendor aliases and the complete raw message.
- Validate original addresses before standard-field promotion and enrichment, excluding
  unspecified IPv4/IPv6 addresses. Keep NAT fields separate. Preserve documented local/remote
  SSL VPN roles rather than inferring an unreported public peer.
- Map WAF client/backend addresses, HTTP status, host and source-relative counters; users,
  groups, heartbeat endpoints, mail attributes, file information and explicit device time.
  Guard numeric types/ranges before strict Event decoding. Preserve textual protocol case.
- Explicit rejection, drop and failed authentication take precedence over allowed statuses
  and HTTP success responses. Detect/Alert alone do not prove success. Only an explicit
  established connection state sets `connectionStatus=established`.
- Align IDP and ATP predicates with documented components, priorities and outcomes. Exclude
  routine updates, successful authentication and IKE parser errors from attack histories.
  Match IP Spoof without an incidental case mismatch.
- Scope four histories to collector, firewall identity and source address; IPS also retains
  destination. Count only relevant recomputed candidates. Thresholds stay 2/30m, 3/15m,
  10/15m and 10/15m. Device ID/serial takes precedence over the ingress-peer fallback.
- Keep existing impact ratings. Generic critical events do not establish exploitation.
  Message 17925 remains a literal notification because the available guide does not explain
  its body. Message 17913 alone no longer proves an administrator login failure; the reviewed
  authentication contract requires failed sign-in and message 17507. Review legacy coverage
  before rollout rather than assuming every firmware version has the same event meanings.

## Central

The integration collector consumes `/siem/v1/events` JSON, not Firewall syslog or a separate
MDR/alerts API. Sanitized key spelling and existing camel-case aliases are preserved.

- Associate validated IP, location and user with the managed computer/server as `target.*`.
  Endpoint identity remains usable without an IP. For explicit CryptoGuard `SMBOrigin`, the
  endpoint is the remote-encryption initiator and maps to `origin.*`; do not invent a peer.
- Vendor `origin` names the detection engine; `group` classifies the event. Neither names
  an attacker or security group. Keep these fields as vendor context. Sophos customer ID
  scopes source histories without overwriting the SIEM tenant metadata.
- Retain vendor fields while adding event type as `action`, explicit `when` as `deviceTime`,
  and low/info, medium/warning, high/error, critical/critical severity mappings. Validate
  artifact hashes. Promote malware labels only for the malware category, not PUA/blocklist
  policy labels.
- Promote a single remedy file to path and basename only when both array length and total
  count equal one. Preserve collections and archive-member locators as vendor context;
  choosing the first item of a multi-file event would falsely attribute the other items.
- Normalize explicit blocked/prevented events to `denied` and exact authentication failures
  to `failure`. Detection, cleanup/update success and VPN status do not establish successful
  network connections. Unknown outcome variants remain unknown.
- Successful cleanup and dismissed/resolved notifications do not create fresh malware alerts.
  Unsuccessful cleanup remains covered. Peripheral blocking is not device compromise.
  Cover behavioral detections, PUA events and account-level tamper-protection Raise warnings,
  while excluding Resolve. Add a device-health rule for high/critical endpoint protection
  unavailability without claiming an attack.
- Preserve all five history thresholds: behavioral 3/30m, exploit 2/30m, MTR 3/1h, ZTNA 10/5m
  and 10/1m. Require collector, account and endpoint ID or validated vendor IP. Count relevant
  derived candidates rather than arbitrary endpoint events. MTR keeps critical-only history.
  Input-supplied scope and candidate markers are cleared before recomputation.
- The one-minute ZTNA rule counts failed attempts, not distinct users. Rename it rapid
  authentication failures without changing its threshold. Both ZTNA rules can overlap.
- Managed endpoints remain alert targets, including the credential rule. Group by available
  collector/account/endpoint context. Keep ZTNA's IP in vendor context until its endpoint role
  is established. All existing impact ratings remain unchanged; the new health rule is 1/2/2.
  Generic behavioral/PUA/malware notices do not assert an unsupported specific ATT&CK technique.

## Tests and deployment limits

Committed tests use **160 fabricated raw fixtures** (73 XG, 87 Central), positive/negative
assertions for all **29 rules**, strict final Event decoding, Alert-side checks and **nine SDK
history-query tests** with loopback mocks. History tests exercise threshold boundaries,
expiration, wrong/missing identities, unmarked records, unrelated events and MTR severity.
Fixtures contain example identities and documentation address ranges.

Extraction uses a declared offline model of YAML steps, observed JSON key behavior and the
versioned SDK CEL evaluator. It is **not the closed EventProcessor**. Geolocation and production
alert creation are not executed. No production false-positive-rate reduction is claimed.

Follow-up (2026-09-24): the EventProcessor rename step copies a value with go-sdk
`utils.GetValueOf`, which returns an object or list as its JSON text. Renaming
`log.core_remedy_items` to `log.coreremedyItems` therefore stored the object as a string, and
the remedy path, file and count were never produced. The remedy object now keeps its native
name and is read in place. Existing indices map `log.coreremedyItems` as text, so the object is
not written under that name. The Central model now follows the executor's JSON, grok and rename
steps, and a fixture covers the usual empty (`null`) remedy value.

Shared alerts draft #2627 supplies the indexed `lastEvent.*` grouping implementation. Its
rollout is separate. New candidate histories require up to one hour of warm-up; historical
unmarked events do not count. Reconcile renamed rules with installed definitions and review
outcome/severity searches before rollout. Vendor fields remain available; historical SQL
seeds and customer-specific saved searches are not rewritten by these changes.

Unexercised legacy XG appliance/antivirus variants and Central ZTNA/MTR/mobile/remote-CryptoGuard
variants remain explicit coverage gaps. In particular, broad XG antivirus adversary direction
is not declared verified. The legacy MTR predicate assumes a schema not proven to arrive via
the SIEM collector. Do not infer operational coverage from synthetic predicate tests alone.
Staging must verify closed-executor behavior, XG parsing cost and the complete rollout.

## Sources

- [SDK v1.1.31 schema](https://github.com/threatwinds/go-sdk/blob/v1.1.31/plugins/plugins.proto)
- [Standard event schema](https://github.com/threatwinds/go-sdk/wiki/Standard-Event-Schema)
- [Filter lifecycle](https://github.com/threatwinds/go-sdk/wiki/Implementing-Filters)
- [Rule implementation](https://github.com/threatwinds/go-sdk/wiki/Implementing-Rules)
- [SFOS 20.0 syslog guide](https://docs.sophos.com/nsg/sophos-firewall/20.0/syslog/index.html)
- [Central events](https://docs.sophos.com/central/customer/help/en-us/ManageYourProducts/LogsReports/Logs/Events/)
- [Malware, PUA and runtime events](https://docs.sophos.com/central/customer/help/en-us/ManageYourProducts/LogsReports/Logs/Events/EventTypes/)
- [Behavioral detections](https://docs.sophos.com/central/customer/help/en-us/ManageYourProducts/LogsReports/Logs/Events/MaliciousBehaviorTypes/)
- [Peripheral/network events](https://docs.sophos.com/central/customer/help/en-us/ManageYourProducts/LogsReports/Logs/Events/NetworkAccessEventTypes/)
- [Management/protection events](https://docs.sophos.com/central/customer/help/en-us/ManageYourProducts/LogsReports/Logs/Events/ManagementEventTypes/)
- [Firewall ATP/status alerts](https://docs.sophos.com/central/customer/help/en-us/ManageYourProducts/Alerts/FirewallAlerts/)
- [Account tamper-protection warnings](https://docs.sophos.com/central/customer/help/en-us/ManageYourProducts/AccountHealthCheck/FixEndpointTamperProtect/)

The older PDF syslog URL redirects to HTML. Off-allowlist API documentation was not fetched;
these sources do not establish a complete versioned enum contract for unobserved variants.
