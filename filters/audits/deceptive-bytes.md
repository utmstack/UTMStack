# Deceptive Bytes v11 filter review

The filter previously extracted a command into a root `command` field. The v11 alert
module pins go-sdk v1.1.36 (v1.1.33 when this was first reviewed); neither version's
`Event` has such a field. Finalization discards it.
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

These four input pairs first produced eight events with the public EventProcessor
playground at commit `497bf53dbd1ae096f7b2dbc7bce77a6bf9f22ce1`, whose parser and
writer link go-sdk v1.1.26. On 2026-09-24 they were run again with EventProcessor
`main` at `8a3ade72bd9d12db21f6b273200588fb49540f14`, whose playground and plugins all
link go-sdk v1.1.36, the version the v11 alerts module now pins: every original and
corrected output in the table is unchanged. The common regex patterns
were read from a deployed UTMStack configuration. The complete staged inputs,
binary hashes and resulting events are retained in the private Data Engine review
run. SDK CEL evaluated the diagnostic predicate
`equals("dataType", "deceptive-bytes") && equals("origin.command", "synthetic --flag")`
on all eight resulting events: only the corrected command case matched. The focused
contract test and full `plugins/alerts` Go suite pass with go-sdk v1.1.36 (48 tests
pass, 11 skip because they need other technologies' private evidence, none fail).
The builds are not asserted to be identical to a customer deployment.

The newest published engine image, `ghcr.io/utmstack/utmstack/eventprocessor:v11.2.14`
(built 2026-09-24 19:13 UTC on base image `eventprocessor/base:1.1.7`), embeds Go build
information showing that its playground and plugin binaries come from the same
EventProcessor revision `8a3ade7` with go-sdk v1.1.36, built with go1.26.8 for
linux/amd64. The local build used here is that source revision compiled natively for
darwin/arm64 with go1.25.7; only the Go toolchain and platform differ.

All 16 shipped Deceptive Bytes rules were checked for consumers of `command`,
`origin.command` and the three temporary KV inputs. None reads them, so the
command fix itself needs no rule rewrite.

Six shipped rules read 18 vendor keys that contain an underscore and reach the
event through KV, for example `log.event_type`, `log.decoy_sensitivity` and
`log.source_ip`, in predicates, history fields and placeholders, and grouping
paths. The KV parser passes every key through go-sdk `utils.SanitizeField`. Up to
v1.1.34 that function removed underscores, so KV stored `log.eventtype` and these
rules could not match; an earlier revision of this draft therefore respelled the 18
names without underscores. Since v1.1.35 the function keeps underscores, and `v11`
now pins v1.1.36, so KV stores the vendor's own spelling and those respelled names
would never match. This revision restores the original names. Four of the six rules
are again identical to `v11`; `data_theft_attempt_indicators` keeps only its
`origin.ip` guard and `nation_state_tactic_detection` only its `"true"` comparisons.
KV also stores values as strings, so ten boolean comparisons across five rules use
`"true"`; go-sdk v1.1.36 CEL still accepts a native boolean `true` for those
comparisons. Literal event labels, thresholds and source-IP requirements are
unchanged. Six source rules needed neither field nor boolean changes.

Three rules run a history search on `{{.origin.ip}}` (and `{{.log.tacticName}}` or
`{{.log.processName}}`), but this filter never writes `origin.ip`. A missing
placeholder makes the search fail, and five failures switch a rule off with a
Circuit Breaker alert, so the data theft, advanced threat tactic and zero-day
conditions now also require those fields, as the other seven `origin.ip` rules of
this source already do. For the same reason `ransomware_behavior_patterns`, which
searches on `{{.log.process}}` and `{{.log.source_ip}}`, now requires both fields. The
history-guard test checks each of the four rules: no match without the fields it needs or
without any one of them, a match with them, and every placeholder resolved. Without the
ransomware guard it fails; the go-sdk v1.1.36 replay over the playground events plus two
copies of the ransomware line that each lack one of those fields then matched both copies
with unresolved placeholders. With the guard the same replay passes 65 of 65 checks: each
rule matches only its intended case, and the ransomware rule only the line that carries
both fields.

The committed fabricated lines carry all 18 keys. On EventProcessor `8a3ade7` the KV
plugin stored every one with its underscore (for example `log.event_type`,
`log.process_name` and `log.deceptive_target`) and none without it. With the
respelled rules the Living Off The Land positive yielded no alert. With the restored
rules the playground raised exactly one alert for each of the Living Off The Land,
nation-state and privilege-escalation positives, containing only that line's event
ID, and none for the near-miss negative or any other line; these three rules need no
history search. The go-sdk v1.1.36 rule replay over the same events, plus copies
given an `origin.ip` and synthetic events for the boolean rules, matched each of the
six restored rules and the four other rules this draft edits only on its intended
case, resolved every history placeholder on those matches, and matched nothing with
the six other rules or with the respelled names (89 of 89 checks). This proves the
local parser and rule contract for these fabricated cases, not the frequency or
semantics of real vendor detections.

No retained Deceptive Bytes documents were found in 29 successful source-index
discovery queries across the accessible v11 estate; two discovery attempts failed.
Thus raw vendor syntax, deployed parser version, production rule coverage, alert
grouping and notification remain unverified. The available official product pages
do not define the severity-letter crosswalk, actor roles, or the vendor event
labels assumed by these rules. Those semantic mappings are left for
a separate review with relevant source logs or a technical export specification.

Sources: [SDK Event schema](https://github.com/threatwinds/go-sdk/blob/v1.1.36/plugins/plugins.proto),
[standard field meanings](https://github.com/threatwinds/go-sdk/wiki/Standard-Event-Schema),
[filter steps](https://github.com/threatwinds/go-sdk/wiki/Filter-Steps-Reference),
[public KV parser](https://github.com/utmstack/EventProcessor/blob/8a3ade72bd9d12db21f6b273200588fb49540f14/plugins/kv/main.go),
[SDK field sanitizer](https://github.com/threatwinds/go-sdk/blob/v1.1.36/utils/fields.go) (keeps letters,
digits, dots and underscores since v1.1.35).

## Reproduce the raw parser and local alert checks

`plugins/alerts/testdata/deceptive-bytes/` contains eleven fabricated raw inputs,
the twelve common regex definitions needed by this filter, and `replay.py`.
It stages the current filter and the three shipped rules that need no history search
(Living Off The Land, nation-state, privilege escalation), runs the actual
playground, and requires all eleven finalized events, zero parser errors, every
vendor key stored with its underscore, and exactly one alert from each rule, on its
positive line. The unrelated, near-miss and other cases must not alert. It records
the binary build information, input/configuration hashes and outputs in a fresh
local directory. It uses no customer connection or index writer. On EventProcessor
`8a3ade7` it passes: 11 events and 3 alerts.

Build a separately checked-out EventProcessor at `8a3ade72bd9d12db21f6b273200588fb49540f14`,
preserving each module's checked-in dependencies. With `EP` set to that checkout's
absolute path:

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
execute raw extraction. Playground startup can take several minutes. At `8a3ade7`
the CEL plugin reads its OpenSearch address from separate `host`, `port`, `user` and
`password` settings, so the single loopback URL in `replay.py` leaves it an empty
address; the staged rules have no history request. This does not validate any other
rule's history, production grouping or notification.
