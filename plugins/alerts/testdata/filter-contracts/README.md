# Filter and rule contracts

Each technology draft adds one independent JSON manifest with the filter paths,
all matching shipped rule paths, and synthetic normalization cases. The shared
runner discovers manifests automatically, so technology PRs do not duplicate Go
helpers or unrelated changes. Merge this test support before relying on `go test`
to execute those manifests. Technology YAML changes remain separate.

Run `go test ./...` from `plugins/alerts`. To check every filter and rule after the
technology corrections are combined, run `UTMSTACK_CONTRACT_ALL=1 go test ./...`.
Before those corrections merge, the all-files mode intentionally reports existing
schema and CEL defects in uncorrected technologies.

Contract tests use the pinned go-sdk v1.1.31 protobuf descriptors and CEL functions.
They check config keys, explicit output paths, Event versus Alert namespaces,
correlation search/placeholder paths, supported operators, and CEL compilation.
Normalization fixtures supply **synthetic extraction results**. The model handles
rename, add, delete, selected casts, trim and one-field copy captures, followed by
real SDK Event conversion and actual rule trigger evaluation. It skips complex
grok, JSON/KV/XML/CSV extraction, timestamp reformatting and dynamic plugins.

These tests do not replace raw-log replay through the EventProcessor or historical
OpenSearch/alert testing. A passing case validates its stated input and assertions;
it does not certify all vendor log variants or live threat-intelligence behavior.

Authority: [filter steps](https://github.com/threatwinds/go-sdk/wiki/Filter-Steps-Reference),
[event schema](https://github.com/threatwinds/go-sdk/wiki/Standard-Event-Schema),
[rule implementation](https://github.com/threatwinds/go-sdk/wiki/Implementing-Rules),
[SDK schema](https://github.com/threatwinds/go-sdk/blob/v1.1.31/plugins/plugins.proto).
`afterEvents`, empty noncapturing grok names, open `log.*` fields and supported
numeric strings are valid and deliberately accepted.
