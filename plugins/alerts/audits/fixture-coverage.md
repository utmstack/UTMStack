# What filter fixtures prove

The shared runner uses the pinned SDK for configuration/schema decoding, CEL and
Event conversion. The closed filter executor is not invoked. Report each level
of coverage separately; a green module run does not establish end-to-end alerts.

## Synthetic normalization input (existing default)

```json
{
  "name": "normalization-only",
  "filter": "source/filter.yml",
  "input": {"log": {"VendorField": "synthetic"}},
  "expected": {"origin.user": "synthetic"},
  "rules": {}
}
```

`input` supplies fields as though extraction already ran. JSON/KV/CSV/XML,
complex grok, dynamic and reformat operations are skipped. This mode establishes
modeled normalization only. `rules: {}` or an absent `rules` member means **no
rule predicate was tested**, which the verbose test output states explicitly.

## Opt-in raw JSON model

```json
{
  "name": "raw-json-model",
  "filter": "utmstack/utmstack.yml",
  "raw": "{\"msg\":\"synthetic\",\"args\":{\"event-id\":17}}",
  "dataType": "utmstack",
  "dataSource": "synthetic-host",
  "expected": {"log.msg": "synthetic", "log.args.eventid": 17},
  "rules": {}
}
```

`raw` and `input` are mutually exclusive. Raw mode requires explicit ingress
`dataType` and `dataSource` and selects matching pipeline stages. It decodes JSON
objects into `log`, recursively uses `utils.SanitizeField` on JSON keys, then
models rename/delete/cast/trim, literal `add.function: string` with SDK string
conversion, and single-line whole-field copy captures. Multiline copies, custom
`greedy` patterns and non-literal add functions fail explicitly. Executed
unsupported operations fail the fixture rather than being silently skipped.
Empty/dotted sanitized keys and sanitization collisions also fail: the closed
executor's behavior for those ambiguous inputs is not established here.

This is a **bounded parser model**, not proof of the closed runtime's extraction.
A fixture demonstrates that its source keys reach the selected destinations
under that model; real retained raw/output pairs or separately authorized staging
execution are still needed to verify deployed behavior. No existing synthetic
technology fixture becomes a raw-parser test merely by loading this runner.

## Rule and history assertions

`rules` maps rule paths to expected `where` booleans; it does not assert generated
alerts. A fixture may additionally enable `checkHistoryPlaceholders: true`.
For each expected-positive, actually-positive rule, the runner checks all
`correlation`/legacy `afterEvents` placeholders, including recursive `or`
branches, against the normalized event. Numeric zero and boolean false resolve;
missing and null values do not.

This optional check intentionally requires **every branch** to resolve. The SDK
may skip an OR branch when its parent meets the threshold, so enabling the flag
requires a fixture intended to cover the whole graph. The flag performs no
history search, count comparison or alert generation. It cannot replace a
technology test using the real `SearchRequest.Execute` and a controlled history
backend. The SDK's process-wide OpenSearch connection also means independent
mock servers must not be assumed to isolate tests within one process.

The runner's output distinguishes synthetic input, raw JSON model, no rule
assertions and placeholder-only preflight. Store source document provenance and
real-vs-synthetic status in the source's audit report; do not publish customer
payloads in fixture JSON.
