# JSON Input normalization and rule review

Correct the injection rule target.host grouping name to the SDK schema.

This draft targets UTMStack `v11`. It contains 0 filter changes
and 1 rule changes for this technology only. Review covered
1 filter configurations and 10 matching shipped rule files.
Unchanged rules are listed in the regression manifest; they are not duplicated in the diff.

## Contract and validation

- Compared exact standard names/types with go-sdk v1.1.31 and the supplied UTMStack dictionaries.
- Checked documented pipeline ordering, rename/move behavior, open vendor log fields,
  event-side versus alert-side fields, and surviving fields used by affected rule predicates/history/grouping.
- Strict SDK configuration decoding and actual CEL compilation pass for this scope.
- 1 synthetic normalization cases pass, including SDK Event conversion and any
  trigger predicate assertions recorded in the manifest.
- The scoped alerts module tests and `git diff --check` pass with the shared contract runner applied.

The shared alert-contract PR supplies the reusable Go runner for the manifest in
`plugins/alerts/testdata/filter-contracts/json-input.json`. Apply that support before running `go test ./...` in `plugins/alerts`.

The model starts from synthetic extraction results. It does not run complex grok,
JSON/KV/XML/CSV extraction, time conversion, dynamic plugins, historical OpenSearch
queries, or the closed EventProcessor. Raw vendor logs and resulting alerts must
still be checked in staging before rollout. No customer false-positive reduction
has been measured and no production rollout is included.

The generic JSON parser places payload fields under log. This change only corrects an invalid target.hostname grouping path. The fixture explicitly supplies standard side identity; deployments still need upstream/custom normalization to supply origin.ip for this rule to trigger. This PR does not claim that arbitrary JSON automatically gains standard side fields.

## References

- [SDK schema](https://github.com/threatwinds/go-sdk/blob/v1.1.31/plugins/plugins.proto)
- [Filter steps](https://github.com/threatwinds/go-sdk/wiki/Filter-Steps-Reference)
- [Standard event schema](https://github.com/threatwinds/go-sdk/wiki/Standard-Event-Schema)
- [Rule implementation](https://github.com/threatwinds/go-sdk/wiki/Implementing-Rules)

`afterEvents`, empty noncapturing grok names, supported numeric strings, and custom
`log.*` fields are accepted. Existing textual protocol casing and vendor action names
are preserved unless a concrete consumer mismatch requires correction.
