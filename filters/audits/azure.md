# Azure Event Hub review — 17 September 2026

Replacement for historical draft #2597. This draft changes the Azure filter and 40
surviving rule predicates, consolidates five redundant definitions, and proposes retiring
one unsupported managed-identity heuristic. All 46 original rules were inspected; 40
remain. Some families have only official-schema/synthetic proof, not native matches.

## Evidence and deployment

The contract is ThreatWinds go-sdk **v1.1.31**, including the protobuf and actual CEL
and SearchRequest implementations, plus the official wiki at commit
`c18b54bd5ea5a34abb0e690458d73f89835edd29`. Local dictionaries are secondary.
The reviewed Azure collector already unwraps JSON arrays and `records` envelopes
before enqueueing each record. No collector change is needed for those wrappers.

Read-only SSH sampling used port **2135**. Thirty v11 instances answered retained
Azure counts; one was unavailable. Azure data was retained on casbio.utmstack.com
and veritas.utmstack.com. The private raw/normalized review covers 21 records:

| Instance | Native class | Representative document IDs |
| --- | --- | --- |
| casbio.utmstack.com | Event Grid resource actions and write | `7f121466-deb9-4311-b444-2ee9ce414123`, `361a46c0-25e1-4418-b710-9579bc424ea0`, `457d5649-d56f-4b70-bd60-5f5d0f7ad32f`, `32511d9a-8211-4102-8dca-e9b67097b3e2` |
| veritas.utmstack.com | App Service console | `af081b1e-04d4-44e2-90ac-e91c026cc756`, `db0b49ac-cbee-4c8e-a8af-6551504eea8f`, `9d556100-2a81-464f-8f7c-79ead36edca5`, `4359695f-123c-49a6-9d76-8762412ada6b` |
| veritas.utmstack.com | App Service HTTP | `7f9d8696-d3e1-472d-9f6f-4889ae926507`, `c6b3ce1c-cc03-4aaf-acf7-90f55691e93f` |
| veritas.utmstack.com | Key Vault Authentication | `cdb26069-04d7-4ac7-8e54-6b24d87a5bae`, `3656b759-9533-447a-ac3c-0643aa1ed7c8` |
| veritas.utmstack.com | Entra audit | `7716268a-dddd-4494-ab94-b6ff779ec7cf`, `6b2784f1-c3df-482b-abdb-89841e441286` |
| veritas.utmstack.com | Entra sign-in | `036f1731-fa1b-40d1-8715-7210f5ef7a68`, `7e03c45e-e8ff-4f0f-bd40-deafc4f2785f` |
| veritas.utmstack.com | AKS audit | `b9f4fe58-21c0-4693-9c40-fc7fab300857`, `f38ea089-c82f-4b03-afc6-cd2947d82c85` |
| veritas.utmstack.com | Defender endpoint process/event export | `114226c3-7b92-40be-a8f8-06c701a57800`, `9ead9961-1769-4800-bd31-5986f4a16219` |
| veritas.utmstack.com | Event Grid role assignment | `a0d8710d-49ed-4891-a985-80d75111aa03` |

**Separate deployment gap:** neither inspected active pipeline directory contains an
Azure filter. Each has 35 numeric YAML files; both manager and worker mount that
same directory. Case-insensitive inspection included both YAML suffixes. This is
not an access failure, and the repository filter alone cannot repair that current
configuration. It does not establish the configuration in July or why the definition
is absent. No customer settings or files were changed.

The four July Event Grid documents retain `log.data`; the September representatives
from veritas have no parsed `log` or standard endpoints/results. Consequently these
samples establish incoming formats and a current configuration gap, not execution
of the proposed filter. Raw text searches were used only to select strata: incidental
matches (for example “Administrative” in endpoint telemetry) are not category counts.

## Corrections

- Preserve vendor objects and legacy aliases. Decode the observed JSON strings in
  App Service `properties` and AKS `properties.log` without letting nested keys replace
  the enclosing event. Temporary Draft namespaces are unconditionally deleted before
  finalization; tests also decode the final Draft with strict protobuf validation.
- Promote Event Grid `eventTime`; Entra sign-in and audit actor identities; Event Grid
  and documented Key Vault token principals; App Service client/server endpoints,
  port and byte directions; and observed endpoint process context. Keep Azure directory,
  subscription and resource identifiers separate from the SIEM tenant identity.
- Validate original IP values and numeric fields before promotion/enrichment. Unspecified
  addresses, invalid ports/status values and nonfinite/negative bytes stay vendor detail.
  An App Service worker's `EventIpAddress` is not assigned to an attacker. Gateway host
  headers describe the target. Relative URIs remain vendor detail rather than invented URLs.
- Both sampled Key Vault Authentication records report `resultType: Success` with
  `resultSignature: Unauthorized`. Explicit rejection/HTTP failure overrides a success
  token. A private before/after replay of historical head `ef7e737a` confirms both
  records modeled as `success`/401 before and `denied`/401 after (two native comparisons).
  Numeric zero means success only in the sign-in schema. Intermediate Accepted,
  Started and In Progress states, WAF Matched/Detected/Allowed, console messages and
  security-alert issuance do not establish a completed attacker connection.
- The wiki expressly permits a role in `origin.group`; that Event Grid mapping is retained.
  Authorization evidence describes the caller's permission. It is not the granted role:
  the sampled role-assignment envelope lacks target-role/request-body details and is
  correctly insufficient to establish subscription Owner assignment.
- Rules consume structured audit arrays, sign-in risk fields/arrays, class-specific results,
  and parsed AKS verbs/resources/stages. Failed, intermediate and unrelated operations have
  negative fixtures. PIM selectors bind the role property name and value to the same item.
  No general `Update user` event is equated with MFA disablement, and risk/token labels no
  longer imply impossible travel, credential theft, Golden SAML or completed exfiltration
  without evidence. Existing rule impact scores, numeric protocol aliases and HTTP action vocabulary are retained.

## Correlation and duplicate alerts

Ten corrected histories require collector, namespaced directory/resource scope and
an actor identity; WAF and invalid-password bursts require a real client IP. Vault
and AKS histories additionally stay within the actual vault/cluster resource. Candidate
markers are cleared from input and recomputed from the qualifying event condition.
Unrelated activity, denied vault access, missing identities, other scopes, expired
windows and below-threshold populations cannot satisfy these histories.

The sign-in burst is **15 invalid-credential failures in 15 minutes**, not proof of
password spraying across distinct accounts. The Secret rule counts **writes/deletes**,
not secret reads. One sampled routine Secret deletion matches that corrected predicate;
no malicious classification, history threshold or live alert is inferred from it.
Routine controllers and administration need environment-specific tuning.

Retire these duplicate rule definitions when reviewing the combined filter/rule rollout:

| Retired file stem | Surviving file stem |
| --- | --- |
| diagnostic_settings_tampering | defense_evasion_azure_diagnostic_settings_deletion |
| defense_evasion_azure_application_credential_modification | azure_app_credential_added |
| impact_azure_service_principal_credentials_added | azure_app_credential_added |
| golden_saml_federation_abuse | azure_federation_modified |
| mfa_disabled_privileged_users | persistence_mfa_disabled_for_azure_user |

The generic risky-sign-in and successful high-risk-sign-in predicates are disjoint;
application consent and application role-assignment predicates are separate. Grouping
and deduplication remain mutually exclusive. Indexed `lastEvent.*` resolution depends
on shared draft #2627. Removing a YAML file is not proof an existing database-loaded
rule has been retired: the team must include that retirement in its reviewed import.

The separate `managed_identity_abuse` retirement is a coverage decision for review,
not a duplicate consolidation. Its control-plane `message` substring “token” and a
history of any events from the same IP do not establish token issuance or abuse.
A replacement requires a documented relevant producer and defensible signal; none
was demonstrated here. Existing database-loaded copies must be retired explicitly
if the team approves this proposal.

The remaining families use explicit semantics: certificate/CA creation audit operation
names; successful TAP registration's `AuthenticationMethod` detail; the SecurityAlert
schema and its own ID; actual requested Owner role ID at subscription scope; and
explicit public-access settings in a present request body. Account permission does
not prove every container is public, and granting Owner does not transfer billing ownership.
Absent bodies, false settings, private containers, failed/intermediate requests, other
role IDs and non-subscription scopes have negative fixtures. The caller's authorization
role is never substituted for the requested role.

## Validation and limits

- **185 synthetic raw fixtures**, each replayed under both preserved and sanitized nested-key
  models; explicit positive and negative assertions cover every changed surviving predicate.
  All 40 remaining predicates are evaluated on every fixture.
- **21 private native records** have explicit field/result/match expectations. No raw customer
  payloads, addresses or identities are committed. All expected mappings pass; the modeled
  Secret deletion is the sole native predicate match. These are offline extraction models,
  not execution of the closed EventProcessor.
- **449 passing source test records, zero failures/skips** with the shared contract support
  overlaid. The history child process separately asserts ten rule scenarios.
- Actual SDK YAML, CEL, strict final Event conversion, Alert endpoint direction, required
  placeholders and **10 SearchRequest histories** are exercised. Histories use an isolated
  localhost mock and wall-clock-anchored timestamps, not customer OpenSearch.
- JSON key sanitization placement remains unverified in the closed engine. Both key layouts
  are tested; native nested claims retain URI punctuation. External geolocation is not executed.
- No new live alert, customer false-positive reduction, pipeline performance or rollout is
  claimed. Profile the additional JSON parsing and alias preservation in staging.
- Existing indexed records do not gain new fields/markers. Deploy filter and consumers together;
  the longest corrected history needs **four hours** of new matching records. Review dashboards,
  saved searches and custom rules for changed outcome semantics and corrected endpoint roles.

Certificate/TAP, forwarded SecurityAlert, public-access changes and target Owner grants
were not observed as qualifying native events. Their proof uses official schemas and
explicit synthetic raw positives/negatives. A bounded September 17 count-only search
found no selected CA/TAP/Sentinel/managed-identity/public-access signatures; this is
not exhaustive absence. The single role-assignment hit lacks the target role.
Management messages containing “token” remain insufficient evidence of token abuse.


## Primary references

- [SDK schema](https://github.com/threatwinds/go-sdk/blob/v1.1.31/plugins/plugins.proto),
  [Draft lifecycle](https://github.com/threatwinds/go-sdk/wiki/Implementing-Filters),
  [filter steps](https://github.com/threatwinds/go-sdk/wiki/Filter-Steps-Reference),
  [event schema](https://github.com/threatwinds/go-sdk/wiki/Standard-Event-Schema).
- [Azure resource-log schema](https://learn.microsoft.com/en-us/azure/azure-monitor/platform/resource-logs-schema),
  [Event Grid events](https://learn.microsoft.com/en-us/azure/event-grid/event-schema-subscriptions),
  [Key Vault log example and meanings](https://learn.microsoft.com/en-us/azure/key-vault/general/logging).
- [Entra activity schema](https://learn.microsoft.com/en-us/entra/identity/monitoring-health/concept-activity-log-schemas),
  [audit activities](https://learn.microsoft.com/en-us/entra/identity/monitoring-health/reference-audit-activities),
  [sign-in errors](https://learn.microsoft.com/en-us/entra/identity-platform/reference-error-codes),
  [risk detections](https://learn.microsoft.com/en-us/graph/api/resources/riskdetection?view=graph-rest-1.0),
  [authentication protocol/token fields](https://learn.microsoft.com/en-us/graph/api/resources/signin?view=graph-rest-beta).
- [App Service HTTP fields](https://learn.microsoft.com/en-us/azure/azure-monitor/reference/tables/appservicehttplogs),
  [console fields](https://learn.microsoft.com/en-us/azure/azure-monitor/reference/tables/appserviceconsolelogs),
  [Gateway access/WAF schemas](https://learn.microsoft.com/en-us/azure/application-gateway/monitor-application-gateway-reference),
  [AKS audit schema](https://learn.microsoft.com/en-us/azure/azure-monitor/reference/tables/aksaudit),
  [Defender process fields](https://learn.microsoft.com/en-us/defender-xdr/advanced-hunting-deviceprocessevents-table).

- [TAP registration](https://learn.microsoft.com/en-us/entra/identity/authentication/howto-authentication-temporary-access-pass),
  [certificate trust-store audit events](https://learn.microsoft.com/en-us/entra/identity/authentication/how-to-certificate-based-authentication),
  [SecurityAlert schema](https://learn.microsoft.com/en-us/azure/sentinel/security-alert-schema),
  [role-assignment request](https://learn.microsoft.com/en-us/rest/api/authorization/role-assignments/create?view=rest-authorization-2022-04-01),
  [built-in Owner role](https://learn.microsoft.com/en-us/azure/role-based-access-control/built-in-roles/privileged),
  [storage account public-access setting](https://learn.microsoft.com/en-us/azure/storage/blobs/anonymous-read-access-configure).
