# Deceptive Bytes v11 filter review

The filter previously extracted a command into a root `command` field. The v11 alert
module pins go-sdk v1.1.33, whose `Event` has no such field. Finalization discards it.
The corrected grok and both quote trims use `origin.command`, the documented
`Side.command` field. Three KV inputs are optional products of different grok
branches; the public KV plugin returns an error when a source is absent. Each KV
step now runs only when its source exists. Present-source separators and outputs
are unchanged. The filter's explicit `log.action=blocked` or `prevented` case
now yields the documented `actionResult=denied` value. It retains the vendor
action in `log.action`, and no shipped Deceptive Bytes rule reads root
`actionResult`.

The following raw inputs are **fabricated parser fixtures**, not captured Deceptive
Bytes records:

| Case | Raw input | Original output | Corrected output |
|---|---|---|---|
| Command | `<14>2026-09-23T12:00:00Z,123,-,45,source,67,path,platform,/tmp/fake.bin,"synthetic --flag"` | `origin.path=/tmp/fake.bin`; command absent; three missing-source KV errors | same path; `origin.command=synthetic --flag`; zero errors |
| Present KV source | `<14>1 2026-09-23T12:00:00Z host 2 foo:1 sampleKey=sampleValue` | `log.sampleKey=sampleValue`; two missing-source KV errors | same vendor fields; zero errors |
| Blocked action | `<14>1 2026-09-23T12:00:00Z host 2 foo:1 action=blocked` | `actionResult=blocked`; two missing-source KV errors | `actionResult=denied`, `log.action=blocked`; zero errors |
| Unrelated raw | `not a Deceptive Bytes message` | three missing-source KV errors | zero errors; no command |

These four input pairs produced eight events with the public EventProcessor playground
at commit `497bf53dbd1ae096f7b2dbc7bce77a6bf9f22ce1`. The parser and writer
link go-sdk v1.1.26; the v11 alerts module pins v1.1.33. The common regex patterns
were read from a deployed UTMStack configuration. The complete staged inputs,
binary hashes and resulting events are retained in the private Data Engine review
run. SDK v1.1.33 CEL evaluated the diagnostic predicate
`equals("dataType", "deceptive-bytes") && equals("origin.command", "synthetic --flag")`
on all eight resulting events: only the corrected command case matched. The focused
contract test and full `plugins/alerts` Go suite pass with go-sdk v1.1.33.
The playground CEL plugin links v1.1.34, so its local alert result is additionally
checked with the reviewed v1.1.33 predicate tests; the builds are not asserted to
be identical to a customer deployment.

All 16 shipped Deceptive Bytes rules were checked for consumers of `command`,
`origin.command` and the three temporary KV inputs. None reads them, so the
command fix itself needs no rule rewrite. Separately, the KV parser calls
`utils.SanitizeField`, which removes underscores from vendor keys before placing
them under `log`. Six shipped rules named 18 underscored `log` fields that this
filter cannot produce through KV. Their predicates, history field/placeholder
paths and alert grouping paths now use the exact sanitized spellings. KV also
stores values as strings, so ten boolean comparisons across five rules now use
`"true"`; the pinned SDK CEL still accepts a native boolean `true` for those
comparisons. Literal event labels, thresholds and source-IP requirements are
unchanged. Six source rules needed neither field nor boolean changes.

For the shipped Living Off The Land rule, a fabricated raw positive containing
`event_type=lolbin_trap process_name=cmd.exe deceptive_target=decoy` and a
near-miss negative with `event_type=ordinary` both parsed to events. The KV
plugin wrote `log.eventtype`, `log.processname` and `log.deceptivetarget`.
With the old rule, neither yielded a local alert. With the corrected rule and
identical candidate filter, the positive yielded exactly one playground alert
containing its event ID, while the negative yielded none. No history query was
needed by this rule. This proves the local parser and rule contract for these
fabricated cases, not the frequency or semantics of real vendor detections.

No retained Deceptive Bytes documents were found in 29 successful source-index
discovery queries across the accessible v11 estate; two discovery attempts failed.
Thus raw vendor syntax, deployed parser version, production rule coverage, alert
grouping and notification remain unverified. The available official product pages
do not define the severity-letter crosswalk, actor roles, or the vendor event
labels assumed by these rules. Those semantic mappings are left for
a separate review with relevant source logs or a technical export specification.

Sources: [SDK Event schema](https://github.com/threatwinds/go-sdk/blob/v1.1.33/plugins/plugins.proto),
[standard field meanings](https://github.com/threatwinds/go-sdk/wiki/Standard-Event-Schema),
[filter steps](https://github.com/threatwinds/go-sdk/wiki/Filter-Steps-Reference),
[public KV parser](https://github.com/utmstack/EventProcessor/blob/497bf53dbd1ae096f7b2dbc7bce77a6bf9f22ce1/plugins/kv/main.go),
[SDK field sanitizer](https://github.com/threatwinds/go-sdk/blob/v1.1.33/utils/fields.go).

## Reproduce the raw parser and local alert checks

`plugins/alerts/testdata/deceptive-bytes/` contains six fabricated raw inputs,
the twelve common regex definitions needed by this filter, and `replay.py`.
It stages the current filter and shipped Living Off The Land rule, runs the actual
playground, and requires all six finalized events, zero parser errors and exactly
one alert containing the positive event ID. The unrelated and near-miss cases must
not alert. It records the binary build information, input/configuration hashes and
outputs in a fresh local directory. It uses no customer connection or index writer.

Build a separately checked-out EventProcessor at the commit above, preserving each
module's checked-in dependencies. With `EP` set to that checkout's absolute path:

```sh
mkdir -p "$EP/test-bin" "$EP/test-plugins"
(cd "$EP" && go build -mod=readonly -o "$EP/test-bin/playground" ./cmd/playground)
for plugin in add grok delete sew kv trim cel saw; do
  (cd "$EP/plugins/$plugin" && go build -mod=readonly -o "$EP/test-plugins/$plugin.plugin" .)
done
```

From this UTMStack checkout, use a Python environment with PyYAML installed:

```sh
python3 plugins/alerts/testdata/deceptive-bytes/replay.py \
  --playground "$EP/test-bin/playground" --plugins "$EP/test-plugins"
(cd plugins/alerts && go test ./... -count=1)
```

The Python replay and Go suite are separate checks. The Go suite alone does not
execute raw extraction. Playground startup can take several minutes. Its CEL
client points only to loopback; the selected rule has no history request. This
does not validate any other rule's history, production grouping or notification.
