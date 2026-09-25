# Generic input filter review (v11)

The generic filter (`filters/generic/generic.yml`) handles events of data type `generic`: logs that
reach the input without a data type (`plugins/inputs/handlers.go:110-111`, `:259-260`) and agent
Filebeat lines that match no known module (`agent/collector/platform/filebeat_amd64.go:41-47`,
`:174-179`). Its grok pattern named its target `field_name`. The SDK only knows `fieldName`, and
the engine's loader drops unknown keys without an error, so the grok step wrote nothing and the
json step failed on every generic event. No generic event kept `log.message` or any key parsed
from its text. This revision fixes the key, runs the json step only on text that starts like a
JSON object, and bumps the version comment. The only rule for this data type is unchanged: it
cannot fire on generic input, before or after the fix.

The schema is ThreatWinds go-sdk **v1.1.36**, as pinned by `plugins/alerts/go.mod` at official
`v11` (`d2479c1a3705eec6a00016689c2bf5fbcc1814f2`). The engine checks used EventProcessor `main`
at `8a3ade72bd9d12db21f6b273200588fb49540f14`, whose playground and grok, json and event-writer
plugins all link go-sdk v1.1.36.

## Evidence basis

- **No generic data exists to observe.** 29 of the 31 v11 instances were searched read-only over
  all retained time. None holds a generic document or an index for this data type, including
  closed and hidden ones. On one instance the search service was restarting, and one gave no SSH
  access. So no real record, sender or log shape was seen, and nothing below is an observation of
  production traffic. Nothing was written to any instance.
- **Production runs the unfixed file.** Two instances (engine images `v11.2.14` and `v11.2.13`)
  run this file as it was before this change, byte for byte, `field_name` included (SHA-256
  `9df08f41...`), and this rule unchanged apart from its database id. The key has been in the
  file since it was added in 2024 (`42daff37`), so every v11 release ships it.
- **How the key is lost, from source.** The backend stores the filter text unchanged
  (`backend/src/main/java/com/park/utmstack/service/DefinitionSyncService.java:80`, `:114-130`),
  and the config plugin writes it unchanged to `<WORK_DIR>/pipeline/filters/<id>.yaml`
  (`plugins/config/main.go:789-815`). The engine reads its pipelines only through go-sdk
  `plugins.GetCfg`
  ([parsing.go:326](https://github.com/utmstack/EventProcessor/blob/8a3ade72bd9d12db21f6b273200588fb49540f14/pkg/parsing/parsing.go#L326)).
  That loader turns each `*.yaml` file into JSON without renaming keys and decodes it with
  `protojson.UnmarshalOptions{DiscardUnknown: true}`
  ([config.go:162-174](https://github.com/threatwinds/go-sdk/blob/v1.1.36/plugins/config.go#L162-L174)),
  and the SDK's [`Pattern`](https://github.com/threatwinds/go-sdk/blob/v1.1.36/plugins/plugins.proto#L275-L278)
  message has only `fieldName` and `pattern`. So `field_name` is dropped without an error or a log
  line; go-sdk v1.1.26 to v1.1.36 all load definitions this way. The grok plugin then skips the
  nameless pattern and writes nothing
  ([grok/main.go:90-94](https://github.com/utmstack/EventProcessor/blob/8a3ade72bd9d12db21f6b273200588fb49540f14/plugins/grok/main.go#L90-L94)),
  the json plugin returns "source was not found"
  ([json/main.go:33-39](https://github.com/utmstack/EventProcessor/blob/8a3ade72bd9d12db21f6b273200588fb49540f14/plugins/json/main.go#L33-L39)),
  and the parser stores that error on the event and keeps the event
  ([parsing.go:469-479](https://github.com/utmstack/EventProcessor/blob/8a3ade72bd9d12db21f6b273200588fb49540f14/pkg/parsing/parsing.go#L469-L479)).
- **Proven by running it.** `plugins/alerts/generic_filter_test.go` loads this file with the SDK's
  own `plugins.GetCfg`, and the EventProcessor playground runs it on 29 fabricated inputs. Both
  results are under [Validation](#validation).
- **No vendor documentation.** Generic input has no vendor, so severity, event meaning and
  coverage are not judged. Everything that would need documentation or real data is deferred.

## Filter changes (version 2.0.1)

| Id | Change | Why | Proof |
|---|---|---|---|
| C-1 | Line 9: `field_name: log.message` becomes `fieldName: log.message`. | The SDK's key name. The loader dropped the old one, so grok wrote nothing and every generic event carried a json error. | Playground, 29 inputs: `log.message` on 0 before and on 27 after (every input that is not blank); errors on 29 before. Go test: the SDK loader keeps `fieldName` `log.message`; the check fails on the old file. |
| C-2 | Line 14: the json step gets `where: 'startsWith("log.message", "{")'`. | The filter's own comment names syslog as expected input. With C-1 alone, every line that is not a JSON object is stored with a JSON parse error, and a blank input with "source was not found": 16 of the 29 inputs in the review's earlier run, each adding two error lines to the engine log. | Playground: errors on 2 of 29, both on text that starts with `{` but is not a JSON object on its first line; apart from the errors field, every stored event equals C-1 alone on all 29 inputs; no where-clause error. Go test: the clause is true for JSON objects, false for text, and false without an error when `log.message` is missing. |
| C-3 | Line 1: version 2.0 becomes 2.0.1. | House style for a fix; nothing reads the comment. | Not needed: the loader drops comments. |

C-2 reads `log.message` rather than `raw`, unlike the same guard in `filters/linux/linux.yml:20`
and `filters/ibm/ibm_as_400.yml:19`, because grok has already trimmed it: JSON after leading
spaces still parses. The go-sdk `startsWith` helper returns false, not an error, when the field is
missing
([cel_overloads.go:47-58, 353-368](https://github.com/threatwinds/go-sdk/blob/v1.1.36/plugins/cel_overloads.go#L353-L368)).

C-2 must never ship without C-1. On the old key it would hide the defect: the review's earlier run
of the old filter with only the guard stored 27 of 27 events with no content and no error. Its
cost is that two input shapes lose their only visible sign of not being parsed, JSON that starts
with a byte-order mark and JSON behind a syslog header; they stay unparsed with or without the
guard. If maintainers want the smallest possible change, drop C-2 together with the json-guard
check in `generic_filter_test.go`; C-1 alone fixes the defect.

## The rule is unchanged

`rules/generic/generic/cross_source_lateral_movement.yml` requires top-level `origin.ip` and
`origin.user` on every match (lines 26-27), and its second branch reads top-level `action` (lines
38-42). Nothing that handles generic events writes those fields. An incoming log carries only its
id, data type, source, time stamp, tenant and raw text; this filter is the only one for the data
type; and its json step puts a sender's own keys under `log.` (a sender's `origin.ip` becomes
`log.origin.ip`). So the rule cannot match a generic event with the old or the new filter. C-1
restores `log.message`, which the rule already reads under that name, and renames nothing the rule
reads. No rule change has to ship with it, and the rule's alert volume stays at zero.

Whether to retire the rule, or re-scope it to data types that do write `origin.ip` and
`origin.user`, is a maintainer decision (D-1). Its name and description promise detection across
several log sources "followed by execution indicators", which the condition does not check, and
its history value `remote login OR authenticated OR session opened` goes to a match query, in
which OR is an ordinary word. Removing the file makes each instance delete the rule at its next
definition sync (`DefinitionSyncService.java:315-328`). Re-scoping would be a new detection with
its own review. The rule's label T1021 Remote Services and its category Lateral Movement match the
current ATT&CK page.

## Validation

**Playground.** EventProcessor `8a3ade7` playground with its grok, json and event-writer plugins,
whose checksums match the verified build; file input, a filter-only profile, one fresh private
working directory per filter and identical input bytes. The 29 inputs are fabricated Log
envelopes with RFC 5737 addresses and example names; they are not committed.

| Input class | Inputs | Unchanged filter (`d2479c1a`) | Committed filter (`34a0f9c6`) |
|---|---|---|---|
| One-line JSON object: plain; nested values and types; a sender `message` key holding text, a number or an object; envelope-like keys; leading spaces; surrounding newlines; two inputs shaped like the agent's Filebeat output | 10 | no `log` object; error "source was not found" | `log.message` and the sender's keys under `log.`; no error |
| Two-line input whose first line is a JSON object (LF and CR LF line breaks) | 2 | the same | the first line parsed, the rest only in `raw`; no error |
| Text that does not start with `{`: syslog (RFC 3164, RFC 5424, JSON behind a syslog header), key=value, CEF, plain and two-line text, a JSON array, number, string, `true`, `null`, JSON after a byte-order mark | 13 | the same | `log.message` (the first line); no parsed keys; no error |
| Text that starts with `{` but is not a JSON object on its first line: brace text, pretty-printed JSON | 2 | the same | `log.message`; the JSON parse error is kept |
| Blank: spaces only, empty | 2 | the same | nothing written; no error (the production input rejects an empty raw text) |
| **Total** | **29** | **0 with `log.message`, 29 with an error** | **27 with `log.message`, 12 with parsed keys, 2 with an error** |

Both runs enqueued and stored 29 of 29 events and finished. `raw` is unchanged on every event,
and nothing is written outside the envelope and `log`. Engine log: 29 json-plugin and 29 pipeline
error lines before, 2 and 2 after; no where-clause error and no mention of the dropped key in
either run. Each run is identical, event by event, to the review's earlier run of the same inputs
(the unchanged filter, and the candidate with the guard).

**Go test.** `plugins/alerts/generic_filter_test.go` checks, with go-sdk v1.1.36:

- that a strict protojson decode of the file finds no unknown key;
- that `plugins.GetCfg`, run in a child process with `WORK_DIR` set to a private temporary folder
  and the filter staged where the config plugin writes it, keeps the grok target `log.message`
  and the pattern `(.*)`;
- that the json step's `where` from that loaded definition, evaluated with the SDK CEL cache the
  engine uses for step conditions, is true for a JSON object (also after leading spaces) and for
  text that starts with a brace; false for syslog text, JSON behind a syslog header, key=value
  text, a JSON array and a number; and false, without an error, when `log.message` is missing.

All three checks fail against the original file: unknown field "field_name", an empty
`fieldName`, and no `where`. In a throwaway copy, a guard on `raw`, an `exists` guard, and the
guard without C-1 each fail the matching check. The full `plugins/alerts` suite passes: 44 tests
pass, 11 skip and none fail (2,205 passing subtests); the base commit gives 43 pass and 11 skip.
The skipped tests need other technologies' private evidence. With `UTMSTACK_CONTRACT_ALL=1`,
`TestFilterAndRuleContracts` also passes for this filter and the rule.

**Rule replay.** The go-sdk v1.1.36 replay of the unchanged rule matches 0 of the 29
committed-filter events and 0 of the 29 unchanged-filter events, with no compile or evaluation
error. Positive control: 2 of 6 fabricated normalized events match, the two that carry an invented
top-level `origin.ip` and `origin.user` (one through the text branch, one through the action
branch), and their history value resolves; the 4 negatives do not match. Without its two `origin`
tests the condition would match 8 of the 29 committed-filter events; none of the 58 playground
events has a top-level `origin` or `action`.

## Deferred

| Id | What | What would unblock it |
|---|---|---|
| D-1 | Retire the rule, or re-scope it to data types that write `origin.ip` and `origin.user`. | A maintainer decision. A re-scoped rule needs its own review, fixtures and history tests. |
| D-2 | `filters/syslog/syslog-generic.yml:9` uses the same `field_name` key, so its grok step writes nothing either, and silently, because it has no json step. | The syslog source's own review. |
| D-3 | Multi-line input: `(.*)` keeps only the first line, and pretty-printed JSON is not parsed. The filter's comment scopes it to one-line logs, and the agent's Filebeat path sends one line per event (`agent/utils/watcher.go:60-64`). | Real multi-line generic records. `(?s:.*)` would keep whole values. |
| D-4 | OpenSearch indexing: a sender's own `message` key can make `log.message` a number or an object, and a key made only of removed characters becomes an empty field name. | An indexing test on a disposable OpenSearch. |
| D-5 | Promoting sender keys that already use standard names, for example `log.origin.ip` to `origin.ip`. | Documentation or observed senders. A new convention, not proposed. |
| D-6 | Coverage and new rules. | Generic data on some instance. |
| D-7 | A shared contract manifest (`plugins/alerts/testdata/filter-contracts/generic.json`). | Optional. The new test, the shared checker in all-files mode and the playground cover this change. |

## Known limits

- No generic event exists on any searched instance, so production behavior before and after the
  change is not observed. The playground results hold for the recorded engine build and the
  fabricated inputs.
- The two instances whose deployed filter was read run older engine builds, whose exact
  EventProcessor revision was not read. The fix rests on the SDK loader, which is the same in
  go-sdk v1.1.26 to v1.1.36, and on the grok and json plugins exercised here.
- No OpenSearch was used: indexing, history searches, alert creation and notifications are not
  tested. The rule cannot reach them on generic input.
- The two Filebeat-shaped inputs are modeled on the agent's code, not on observed records.

## Reproduce

```sh
(cd plugins/alerts && go test ./... -count=1)
```

The Go suite does not run the engine. The playground runs used the file input of an
EventProcessor `8a3ade7` build without dependency changes, with only the grok, json and sew
plugins staged.
