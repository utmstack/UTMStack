# AWS filter and correlation review — 2026-09-17

Replacement for historical PR #2596 (`cd4034746a957520acad719929678029d6a2a36f`).
This review changes the AWS filter and all 73 consumers discovered recursively in
both `rules/cloud/aws/` and `rules/cloud/aws/aws/`. It is a draft for review, not a
production deployment. The shared alert grouping support is in draft #2627.

## Evidence and limits

Fresh read-only retained-index counts returned **zero AWS records on 30 reachable
v11 instances**. One other instance was unavailable. No live AWS filter output,
AWS customer document ID, resulting alert, or false-positive reduction is claimed.

The contract is go-sdk **v1.1.31**, pinned by `plugins/alerts/go.mod`; its protobuf
and official wiki take precedence over local dictionaries. The reviewed wiki
revision is `c18b54bd5ea5a34abb0e690458d73f89835edd29`. The AWS collector forwards
individual CloudWatch `GetLogEvents` messages unchanged. It does not unwrap S3
CloudTrail archive `Records` arrays. This filter handles individual CloudTrail
JSON records and explicitly identified GuardDuty EventBridge finding envelopes;
other AWS message formats are not silently treated as CloudTrail.

The private documentation replay contains 31 complete examples extracted from
AWS's record, sign-in, log-file and Route 53 documentation. Some are illustrative
examples with placeholder values. Additional official API references establish
operation semantics; they do not prove every service's deployed CloudTrail
request-parameter serialization. Nested variants in the synthetic fixtures must
still be checked against actual customer records when AWS telemetry is available.

## Producer and consumer corrections

- Retain `userIdentity`, `requestParameters`, `responseElements`,
  `additionalEventData` and `tlsDetails`. The previous filter moved selected
  children and deleted their parents, although many rules consumed those parents.
  Existing flattened aliases remain for compatibility. A second JSON pass restores
  the original sanitized structures after the alias moves; this has a processing
  and storage cost that has not been measured with the closed EventProcessor.
- Promote event name, time, caller ARN/user, literal source IP, service hostname and
  target AWS service domain into supported standard fields. AWS service DNS names
  and `AWS Internal/#` are not IP addresses. Unspecified IPv4/IPv6 addresses and
  non-string user values are not promoted. Geolocation receives only a usable IP.
  AWS accounts, reporting resources and GuardDuty findings are not invented actors.
- Derive `failure` from top-level or nested errors and explicit failed sign-in
  results, with recognized authorization errors classified `denied`. Error-message-
  only console failures cannot become `success`. Successful API requests normally
  omit `errorCode`; SDK `equals(path, "")` does not match an absent field. Rules now
  consume the derived outcome. API success means request success, not completion
  of an asynchronous export, deletion, instance launch or command execution.
- Preserve `Root`, `MFAUsed: Yes/No` and session MFA fields. Missing MFA evidence is
  not equivalent to an explicit negative value. Root login success, root login
  failure and successful root login without MFA have distinct predicates.
- Replace string searches on JSON objects with field, boolean and nested-array
  queries. SDK `contains` only handles string fields. Covered consumers include
  startup data, RDS public restore, S3 versioning/ACL/block controls, snapshot
  sharing, EC2 metadata, ECS task definitions and security-group permissions.
  SDK field sanitization means `x-amz-acl` is read as `xamzacl` after JSON parsing.
- Restrict exposure predicates to permission additions or relaxed controls;
  removing snapshot or security-group permissions does not establish exposure.
  Correct cross-account ARN comparison to exact account boundaries and quote the
  field paths passed to `safe`. Use the Route 53 Domains service namespace and its
  documented event-name capitalization variants.
- Consume GuardDuty EventBridge `detail.severity`, not the severity of an unrelated
  GuardDuty management API call. High/critical findings use the documented 7–10
  range and group by collector/account/region/finding ID. Ingestion of that feed is
  not proven by this review. No finding action or attacker IP is fabricated.
- Correct claims that exceeded the events: task management calls do not prove ECS
  credential endpoint access; rejected trust-policy updates are not password brute
  force; a SAML provider-change sequence does not prove forged SAML; country changes
  do not prove impossible travel; an instance type does not prove cryptomining.
  Existing administrative activity heuristics still require operational tuning.
  Correct unrelated ATT&CK labels for secret-store access, cloud discovery and IAM
  grants; omit specific techniques where the broad signal cannot establish one.

## Historical correlation and grouping

All 22 history-bearing rules now restrict the population to the relevant filter-
produced candidate, collector and namespaced account. Original IP-based populations
remain IP-scoped; actor-based populations use a namespaced ARN/principal/IP identity.
Missing required identities prevent the trigger instead of producing a broad query.
Input-supplied candidate and identity markers are cleared before derivation.

Existing thresholds and windows are retained, including 3/30m and 100/15m and windows up to 24h,
with these explicit sequence corrections:

- SAML role assumption requires a successful CreateSAMLProvider/UpdateSAMLProvider
  in the same account/collector within 24h. The administrator can differ. Ordinary
  account activity and the login itself cannot satisfy that prerequisite. This is
  account-level context, not proof that the changed provider issued the assertion.
- Secrets Manager uses **10 GetSecretValue OR 5 BatchGetSecretValue within 10m**;
  the earlier two top-level blocks required both populations. Mixed counts below
  both thresholds do not qualify.
- Different-country login history also requires successful console authentication
  and a country value. Failed logins and missing geolocation cannot satisfy it.
- Administrative policy-attachment history counts the same reviewed role/user
  attachment population as its trigger, rather than unrelated IAM requests.

Grouping/deduplication retain the intended field type and remain mutually exclusive.
CloudTrail groups include collector and account scope; the snapshot identifier is
read from the request. Indexed `lastEvent.*` names remain valid and depend on shared
PR #2627 to resolve values from the alert's wire `events[]`. This review does not
claim to have run live grouping or alert creation.

## Validation

- **190 synthetic raw JSON cases**, each evaluated against all **73 predicates**,
  with positive and negative coverage for every rule. Cases cover omissions, nested
  errors, denials, MFA states, arrays, removal-only changes, exact ARN boundaries,
  unknown records, marker spoofing, invalid identity/numeric values, DNS callers,
  unspecified IPs, GuardDuty boundaries and archive wrappers.
- **31 official AWS examples** replayed privately with independently specified
  field, outcome and predicate assertions; no evaluation errors. They include
  error-message-only failed root logins and successful calls with absent errorCode.
- **22 actual SDK history suites**, against an isolated loopback OpenSearch mock,
  assert exact query terms, account/collector/actor/IP scopes, threshold/window
  boundaries, missing placeholders, unrelated populations and sequence/OR behavior.
- Strict SDK YAML decoding, field-name/type checks, final Event serialization and
  the shared contract suite pass: **275 pass records, zero skips/failures** in the
  source run with official examples supplied. The history subprocess asserts its
  22 cases internally; the outer pass count is not a count of live detections.

`aws_contract_test.go` models explicit YAML JSON sanitization, grok, rename, add and
delete operations and uses the real SDK CEL evaluator and Event conversion.
`aws_history_test.go` uses the SDK SearchRequest implementation with a mock backend.
Neither executes the closed EventProcessor, external geolocation service, real
OpenSearch mappings nor production alerts. Country tests explicitly supply mocked
geolocation. Without the private official example file, that test reports a skip.

Run the source suite with the shared test support from #2627, then run the combined
source overlay before rollout. The filter and rules must ship together. Existing
indexed events do not acquire retained nested objects or candidate markers; allow
up to **24 hours** of new history before the longest rules reach full coverage.
Review dashboards/custom searches for changed standard outcomes, corrected aliases,
new nested objects, group scope and revised titles. Vendor action names are retained.

## Explicit unresolved mapping

The legacy `bytesTransferredIn -> origin.bytesReceived` and
`bytesTransferredOut -> origin.bytesSent` directions are **not verified**. The
fetched references and retained telemetry did not establish their viewpoint, so
this change does not reverse them speculatively. Finite nonnegative numeric guards
prevent invalid values entering the SDK double fields, and original vendor byte
fields are retained. None of these 73 AWS rules reads the byte fields. Do not use
this review as validation of AWS byte-based dashboards or external rules.

## Primary references

- [SDK schema](https://github.com/threatwinds/go-sdk/blob/v1.1.31/plugins/plugins.proto)
- [Filter steps](https://github.com/threatwinds/go-sdk/wiki/Filter-Steps-Reference)
- [Standard event schema](https://github.com/threatwinds/go-sdk/wiki/Standard-Event-Schema)
- [Rules and correlation](https://github.com/threatwinds/go-sdk/wiki/Implementing-Rules)
- [CloudTrail record fields](https://docs.aws.amazon.com/awscloudtrail/latest/userguide/cloudtrail-event-reference-record-contents.html)
- [CloudTrail identity](https://docs.aws.amazon.com/awscloudtrail/latest/userguide/cloudtrail-event-reference-user-identity.html)
- [Console sign-in examples](https://docs.aws.amazon.com/awscloudtrail/latest/userguide/cloudtrail-event-reference-aws-console-sign-in-events.html)
- [CloudTrail log examples](https://docs.aws.amazon.com/awscloudtrail/latest/userguide/cloudtrail-log-file-examples.html)
- [Route 53 CloudTrail](https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/logging-using-cloudtrail.html)
- [S3 CloudTrail event names](https://docs.aws.amazon.com/AmazonS3/latest/userguide/cloudtrail-logging-s3-info.html)
- [IAM trust-policy update errors](https://docs.aws.amazon.com/IAM/latest/APIReference/API_UpdateAssumeRolePolicy.html)
- [EC2 snapshot permission changes](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_ModifySnapshotAttribute.html)
- [GuardDuty EventBridge shape](https://docs.aws.amazon.com/guardduty/latest/ug/guardduty_findings_eventbridge.html)
- [GuardDuty severity ranges](https://docs.aws.amazon.com/guardduty/latest/ug/guardduty_findings-severity.html)

Legacy third-party rule references are retained as contextual citations, not
verified format evidence. Vendor evidence in this review was fetched only from
allowlisted official AWS hosts. No customer payloads are committed.
