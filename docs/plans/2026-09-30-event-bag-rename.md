# Event Bag Rename (`log`→`event` flattened) + `controls` — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use subagent-driven-development (recommended) or executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the heterogeneous `log` map with a new `event` bag (OpenSearch type `flattened`) so new logs can never hit dynamic-mapping type conflicts, add a `controls []string` field for future compliance tags, and migrate all UI/backend consumers off `log.*` (and dead `logx.*`).

**Architecture:** The `log` JSON key IS the protobuf field name for `Event.Log` (field 9, `go-sdk/plugins/plugins.proto`). Renaming that one field cascades the new `event` key through the draft JSON, the OpenSearch document, and CEL rule evaluation in a single move — no dual namespace. Parsers that hardcode the prefix (`json`/`xml`/`kv`) switch to `event.%s`; every other parser takes field paths from filter YAML, which gets a scripted `log.`→`event.` rename. The installer's `v11-log-*` template pins `event: {type:flattened}` + `controls: {type:keyword}` + typed top-level fields. The Java `SearchUtil` becomes flattened-aware (translates UI operators to the query types `flattened` supports). The file-management module moves to canonical `origin.*`/`target.*` fields. **Do NOT touch** `plugins.Log` (gRPC input message) or `Draft.log` (a JSON string) — only `Event.log`.

**Tech Stack:** Go (protobuf + gRPC + tidwall/gjson/sjson), Java 17 (Spring Boot + OpenSearch Java client), Angular 7 / TypeScript, OpenSearch (composable index templates, `flattened` field type), PowerShell for scripted renames.

**Supersedes:** `docs/plans/event-rename-plan.md` and the earlier `log`-pinning plan — this file is canonical.

---

## Cross-repo change order (strict — do not reorder)

Each repo is its own module; the rename only works when they deploy together. Order by dependency:

| Task | Repo | Work | Depends on |
|---|---|---|---|
| 1–2 | `threatwinds/go-sdk` | proto field 9 `log`→`event` + field 20 `controls`; regen `plugins.pb.go`; tests; push+tag | — |
| 3–5 | `utmstack/EventProcessor` | bump go-sdk (16 modules); fix 3 hardcoded `log.%s` parsers; rebuild base image | 1–2 |
| 6 | `utmstack/UTMStack` `filters/` | scripted `log.`→`event.` | 3–5 (for E2E) |
| 7 | `utmstack/UTMStack` `filters/windows` | promote object-access → origin/target; `eventName`→`action` grok-copy | 6 |
| 8 | `utmstack/UTMStack` `rules/` | scripted `log.`→`event.` | 6 |
| 9 | `utmstack/UTMStack` `installer/` | template `event:flattened` + `controls:keyword` + typed top-levels | — |
| 10 | `utmstack/UTMStack` `backend/` | `SearchUtil` flattened-aware (incl. sort/range 400s) | 6,8 |
| 11 | `utmstack/UTMStack` `backend/`+`user-auditor/` | `OverviewService` topWindowsEvents→event+action; drop `logx`; alert CSV + user-auditor read `event`; searchBySid `matchPhrase`→`term`; delete dead `LogType` | 10 |
| 12–13 | `utmstack/UTMStack` `frontend/` | `flattened` type; operator handling; file-mgmt→origin/target; dashboard→`action`; AD `item.event.*`; cache-bust | 9,10 |
| 14 | `utmstack/UTMStack` `plugins/soc-ai` | `LastEvent.Log`→`.Event` struct access + go-sdk bump | 1 |
| 15 | `utmstack/UTMStack` `plugins/alerts` | contract-test fixtures `log`→`event` + go-sdk bump | 1 |
| 16 | all | E2E verification, coordinated rollout, rollback | 1–15 |

**No liquibase rename is required.** `DefinitionSyncService` (a `CommandLineRunner` in the backend) resyncs Postgres from the `filters/`+`rules/` filesystem at every startup — it updates/creates filters and updates rules by `rule_name`, deleting orphans. Renaming the repo YAML therefore self-heals the DB on next backend boot. See Task 11 (verification only).

**Deploy as one coordinated release.** go-sdk tag + EventProcessor base image + UTMStack (filters/rules/installer/backend/frontend) must ship together; deploying only some leaves mixed `log`/`event` docs.

**Build-environment constraint:** this machine has Java 25 (Temurin) — the backend (Java 17 / JHipster) and user-auditor (Java 11) do NOT build here. Tasks 10–11 (Maven `compile`/`test`) must be run on a box with the matching JDK or in CI. Tasks 1–9, 12–16 (Go + Node 14) build fine locally.

---

## Worktree setup

The main checkouts carry other sessions' work. Create isolated worktrees per repo before editing:

```powershell
$base = "$env:LOCALAPPDATA\Temp\opencode"
git -C D:\Projects\github.com\threatwinds\go-sdk      worktree add "$base\go-sdk-event"      -b feat/event-bag-rename main
git -C D:\Projects\github.com\utmstack\EventProcessor worktree add "$base\eventproc-event"   -b feat/event-bag-rename main
git -C D:\Projects\github.com\utmstack\UTMStack       worktree add "$base\utmstack-event"    -b feat/event-bag-rename v11
```

All paths below are relative to that repo's worktree root.

---

## File structure map

### threatwinds/go-sdk
- `plugins/plugins.proto` — rename field 9, add field 20. **Source of truth.**
- `plugins/plugins.pb.go` / `plugins_grpc.pb.go` — regenerated (never hand-edit).
- `plugins/cel_event_field_test.go` — new tests.

### utmstack/EventProcessor
- `plugins/json/main.go`, `plugins/xml/main.go`, `plugins/kv/main.go` — `log.%s`→`event.%s`.
- 16 `go.mod` + root `go.mod` — bump `github.com/threatwinds/go-sdk`.

### utmstack/UTMStack
- `filters/**/*.yml` — scripted `log.`→`event.`.
- `filters/windows/windows-events.yml` — object-access promotion to origin/target.
- `rules/**/*.yml` — scripted `log.`→`event.`.
- `installer/services/search.go`, `installer/setup/apply.go` — template + upgrade lock.
- `backend/.../service/elasticsearch/SearchUtil.java` — flattened operator translation.
- `backend/.../config/Constants.java:72-73` — remove dead `logx*`.
- `user-auditor/.../model/event/Event.java:21`, `service/UserService.java`, `service/elasticsearch/Constants.java:14`.
- `frontend/.../enums/elastic-data-types.enum.ts`, `.../utm-elastic-filter/shared/util/operator.service.ts`, `.../elastic-filter-add/elastic-filter-add.component.ts`.
- `frontend/.../file-management/shared/enum/file-field.enum.ts`, `.../shared/const/file-field.constant.ts`, `file-view/file-view.component.ts`.
- `frontend/.../dashboard/dashboard-overview/dashboard-overview.component.ts`.
- `frontend/.../active-directory/.../event-timeline.component.ts`, `.../active-directory-event.component.ts`.
- `frontend/.../services/elasticsearch/local-field.service.ts` — cache-bust.
- `plugins/alerts/*_test.go` — fixtures.

---

## Task 1: go-sdk — rename `Event.log`→`Event.event`, add `controls`, regen, tests

**Repo:** `threatwinds/go-sdk` (worktree `feat/event-bag-rename`)
**Files:**
- Modify: `plugins/plugins.proto`
- Regenerate: `plugins/plugins.pb.go`, `plugins/plugins_grpc.pb.go`
- Test: `plugins/cel_event_field_test.go` (new)

**Background for the engineer:** The only proto field being renamed is `Event.log` (number 9). THREE different things are named "log" — do not confuse them:
- `message Event { map<string, google.protobuf.Value> log = 9; }` — **THIS ONE. Rename to `event`.**
- `message Log { string raw = 6; ... }` — a gRPC *input* message (the raw incoming log line). Do NOT rename.
- `message Draft { string log = 1; }` — a string field holding the draft JSON. Do NOT rename.

The Go field `Event.Log` (type `map[string]*structpb.Value`) becomes `Event.Event`; getter `GetLog()`→`GetEvent()`.

- [ ] **Step 1: Write the failing test**

Create `plugins/cel_event_field_test.go`:

```go
package plugins

import (
	"testing"

	sdkutils "github.com/threatwinds/go-sdk/utils"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/types/known/structpb"
)

func valOf(s string) *structpb.Value {
	v, _ := structpb.NewValue(s)
	return v
}

// TestEventMarshalUsesEventKey confirms Event.Event serializes under "event"
// (not "log") in the JSON handed to CEL and OpenSearch.
func TestEventMarshalUsesEventKey(t *testing.T) {
	e := &Event{
		Id:       "1",
		Timestamp: "2026-09-30T00:00:00Z",
		DataType: "wineventlog",
		Event:    map[string]*structpb.Value{"eventCode": valOf("4624"), "severity": valOf("high")},
		Controls: []string{"NIST-800-53-AC-3"},
	}

	s, err := sdkutils.ProtoMessageToString(e)
	if err != nil {
		t.Fatalf("ProtoMessageToString: %v", err)
	}
	if !gjson.Get(*s, "event.eventCode").Exists() {
		t.Fatalf("expected event.eventCode in %s", *s)
	}
	if gjson.Get(*s, "log").Exists() {
		t.Fatalf("event should NOT still serialize under 'log': %s", *s)
	}
	if len(e.Controls) != 1 || e.Controls[0] != "NIST-800-53-AC-3" {
		t.Fatalf("controls not populated: %v", e.Controls)
	}
}

// TestEventRoundTrip confirms a doc serialized with "event" reads back into an
// Event (the correlation path: OS _source -> ParseSourceToProtoMessage).
func TestEventRoundTrip(t *testing.T) {
	s, _ := sdkutils.ProtoMessageToString(&Event{
		Id: "2", DataType: "wineventlog",
		Event: map[string]*structpb.Value{"eventCode": valOf("4624")},
	})
	var back Event
	if err := sdkutils.StringToProtoMessage(s, &back); err != nil {
		t.Fatalf("round-trip: %v", err)
	}
	if !gjson.Get(*s, "event.eventCode").Exists() || back.Event == nil {
		t.Fatalf("round-trip lost event: %s / %+v", *s, back.Event)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run (in worktree root): `go test ./plugins/ -run TestEventMarshal -v`
Expected: FAIL — compile error `e.Event undefined (type *Event has no field or method Event)`.

- [ ] **Step 3: Edit the proto**

In `plugins/plugins.proto`, `message Event` (lines 52–72), change line 61 and add field 20:

```proto
  string raw = 8;
  // RENAMED from: map<string, google.protobuf.Value> log = 9;
  map<string, google.protobuf.Value> event = 9;
  Side target = 10;
  ...
  map<string, ComplianceValues> compliance = 19;
  // Compliance control tags (populated by the compliance orchestrator later).
  repeated string controls = 20;
```

- [ ] **Step 4: Regenerate Go code**

```bash
cd plugins
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       plugins.proto
```
(Use the repo's documented codegen entrypoint if the plugins aren't on PATH; the exact flag set is recoverable from git history of `plugins.pb.go`. The requirement: `plugins.pb.go` is regenerated from the edited `plugins.proto`, not hand-edited.)

Confirm the regen produced:
```bash
grep -n "Event map\[string\]\*structpb.Value" plugins/plugins.pb.go
grep -n "Controls \[\]string" plugins/plugins.pb.go
grep -n "func (x \*Event) GetEvent()\|func (x \*Event) GetControls()" plugins/plugins.pb.go
```

- [ ] **Step 5: Fix go-sdk internal references to the Event.Log field**

Find references to the Event struct's field/getter (NOT the `plugins.Log` message, NOT `Draft.log`, NOT logger calls):
```bash
grep -rn "\.Log\b\|GetLog()" plugins/ os/ utils/ | grep -v "plugins\.Log\|Draft\|t\.Log\|\.Log()\|catcher\|log\."
```
Rename each genuine `Event.Log`/`GetLog()` to `Event.Event`/`GetEvent()`. CEL marshals via protojson, so the JSON key changes automatically — no CEL logic change.

- [ ] **Step 6: Run the tests**

Run: `go test ./plugins/ -run 'TestEventMarshal|TestEventRoundTrip' -v`
Expected: PASS.

- [ ] **Step 7: Build the whole SDK**

Run: `go build ./... && go test ./plugins/ ./os/ -count=1`
Expected: PASS. If a pre-existing test constructs `Event{Log: ...}`, update it to `Event{Event: ...}`.

- [ ] **Step 7b: Verify `controls` omitempty behavior**

In `plugins/cel_event_field_test.go`, add:
```go
// TestControlsOmittedWhenEmpty confirms an empty controls list is absent from
// the JSON so the field costs nothing today.
func TestControlsOmittedWhenEmpty(t *testing.T) {
	s, _ := sdkutils.ProtoMessageToString(&Event{Id: "3", DataType: "x"})
	if gjson.Get(*s, "controls").Exists() {
		t.Fatalf("empty controls should be omitted: %s", *s)
	}
}
```
Run: `go test ./plugins/ -run TestControlsOmittedWhenEmpty -v`
Expected: PASS (proto3 omits empty repeated fields by default).

- [ ] **Step 8: Commit**

```bash
git add plugins/plugins.proto plugins/plugins.pb.go plugins/plugins_grpc.pb.go plugins/cel_event_field_test.go
git commit -m "feat(plugins): rename Event.log->event, add Event.controls"
```

---

## Task 2: go-sdk — push + tag

**Repo:** `threatwinds/go-sdk`

- [ ] **Step 1: Pick the next tag**

```bash
git tag --sort=-v:refname | head -5
```
Note the highest tag; use the next patch version (e.g. `v1.1.37`).

- [ ] **Step 2: Push branch + tag**

```bash
git push -u origin feat/event-bag-rename
git tag <NEXT_TAG>
git push origin <NEXT_TAG>
```

**Note to orchestrator:** this tag feeds the EventProcessor base image (Task 5) and then `TW_EVENT_PROCESSOR_VERSION_PROD`. Do not flip that org variable until Tasks 3–14 are merged and E2E passes (Task 15).

---

## Task 3: EventProcessor — bump go-sdk across all modules

**Repo:** `utmstack/EventProcessor` (worktree `feat/event-bag-rename`)
**Files:** root `go.mod` + each `plugins/*/go.mod` (16 modules) + `go.sum`

**Background:** The EventProcessor is a multi-module Go repo. Every module that depends on `github.com/threatwinds/go-sdk` must move to the new tag together, or the build breaks (mixed `Event.Log` vs `Event.Event`).

- [ ] **Step 1: Find every go.mod referencing go-sdk**

```powershell
Get-ChildItem -Recurse -Filter go.mod | Where-Object { (Get-Content $_.FullName -Raw) -match 'threatwinds/go-sdk' } | Select-Object FullName
```
Expected: root `go.mod` + the plugin modules (root `cmd/...` may or may not reference it — the list is authoritative).

- [ ] **Step 2: Bump each to the new tag**

```powershell
Get-ChildItem -Recurse -Filter go.mod | ForEach-Object {
  $c = Get-Content $_.FullName -Raw
  if ($c -match 'threatwinds/go-sdk') {
    $c = [regex]::Replace($c, 'github\.com/threatwinds/go-sdk v[\d.]+', 'github.com/threatwinds/go-sdk <NEXT_TAG>')
    Set-Content $_.FullName $c
  }
}
```

- [ ] **Step 3: Tidy + build every module**

```powershell
Get-ChildItem -Recurse -Filter go.mod | ForEach-Object {
  Push-Location $_.Directory.FullName
  go mod tidy; go build ./...
  if ($LASTEXITCODE -ne 0) { Write-Error "BUILD FAILED in $($_.Directory.FullName)" }
  Pop-Location
}
```
Expected: all build. A module referencing `Event.Log` on the struct will fail here — fix it to `Event.Event` (only `cel`/`feeds`/`stats` build Events/Alerts; check those).

- [ ] **Step 4: Commit**

```bash
git add go.mod go.sum plugins/*/go.mod plugins/*/go.sum
git commit -m "chore: bump go-sdk to <NEXT_TAG> (event bag rename)"
```

---

## Task 4: EventProcessor — fix the 3 parsers that hardcode `log.`

**Repo:** `utmstack/EventProcessor`
**Files:**
- Modify: `plugins/json/main.go:54`
- Modify: `plugins/xml/main.go:55`
- Modify: `plugins/kv/main.go:58`

**Background:** These three parsing plugins explode a raw blob into keys under a hardcoded prefix. That prefix is the bag root. Today they write `log.<key>`; they must write `event.<key>`. The other parsers (grok, csv, add, cast, rename, delete, reformat, trim) take full field paths from filter YAML — no code change (their YAML changes are Tasks 6/7).

- [ ] **Step 1: json plugin**

`plugins/json/main.go` line 54 — change:
```go
transform.Draft.Log, err = sjson.Set(transform.Draft.Log, fmt.Sprintf("log.%s", k), v)
```
to:
```go
transform.Draft.Log, err = sjson.Set(transform.Draft.Log, fmt.Sprintf("event.%s", k), v)
```
(`transform.Draft.Log` stays — it is the `Draft.log` JSON string field, NOT the bag.)

- [ ] **Step 2: xml plugin**

`plugins/xml/main.go` line 55 — same change: `fmt.Sprintf("log.%s", k)` → `fmt.Sprintf("event.%s", k)`.

- [ ] **Step 3: kv plugin**

`plugins/kv/main.go` line 58 — same change: `fmt.Sprintf("log.%s", k)` → `fmt.Sprintf("event.%s", k)`.

- [ ] **Step 4: Verify no other hardcoded `log.` prefix remains in parsing plugins**

```powershell
Get-ChildItem plugins -Recurse -Filter *.go | Select-String -Pattern '"log\.' | Select-Object Path,Line
```
Expected: zero hits under `plugins/`. (Hits under `cmd/log-pusher` or `cmd/opensearch-fetcher` are test fixtures that push raw docs into OpenSearch — update any fixture JSON key `"log":` → `"event":` if present, but don't let it block this task.)

- [ ] **Step 5: Build + commit**

```powershell
go build ./plugins/json/... ./plugins/xml/... ./plugins/kv/...
git add plugins/json/main.go plugins/xml/main.go plugins/kv/main.go
git commit -m "fix(parsing): write json/xml/kv fields under event. not log."
```

---

## Task 5: EventProcessor — build + push the base image

**Repo:** `utmstack/EventProcessor`
**Files:** none (build/publish only)

**Background:** The base image bakes the engine + all plugin binaries. Filters/rules YAML are NOT baked — they ship separately via the UTMStack `release/v11.x` pipeline, so the engine and the renamed YAML must be released in the same window (else new engine + old filters = silent breakage; see repo memory). The UTMStack release pins the base image via `TW_EVENT_PROCESSOR_VERSION_PROD`.

- [ ] **Step 1: Push the branch and run the base build workflow**

```bash
git push -u origin feat/event-bag-rename
gh workflow run build.yaml -f version=<NEW_BASE>   # e.g. 1.1.7
gh run watch
```
Expected: `ghcr.io/utmstack/eventprocessor/base:<NEW_BASE>` published with the renamed parsers.

- [ ] **Step 2: Smoke-test the image against a dev OpenSearch**

Run the built image, feed one raw syslog line through the input, then:
```bash
curl -sk -u admin:PASS https://localhost:9200/v11-log-*/_search -H 'Content-Type: application/json' -d '{"query":{"match_all":{}},"size":1}'
```
Expected: the hit `_source` contains `"event":{...}` and **no** `"log"` key. Also confirm `cel_analysis.sock` exists in the worker (a dead cel plugin = no rules fire; see repo memory).

- [ ] **Step 3: DO NOT flip the org variable yet**

Hold `TW_EVENT_PROCESSOR_VERSION_PROD` at its current value until Tasks 6–14 land and the coordinated release is ready (Task 15 Step 1).

---

## Task 6: UTMStack filters — scripted `log.`→`event.` rename

**Repo:** `utmstack/UTMStack` (worktree `feat/event-bag-rename`, off `v11`)
**Files:** every `filters/**/*.yml`

**Background:** Filter steps reference the bag by full path (`log.localIp`, `to: log.eventName`, `where: exists("log.localIp")`). These become `event.localIp` etc. The rename is mechanical but must ONLY touch bag-prefixed field paths — not the word "log" in comments, vendor names (`winlogbeat`, `Syslog`, `syslog`), or top-level fields.

- [ ] **Step 1: Baseline count (before)**

```powershell
cd filters
$before = (Get-ChildItem -Recurse -Filter *.yml | Select-String -Pattern '"log\.|\bto: log\.|\bfrom: log\.|- log\.' | Measure-Object).Count
Write-Output "before: $before"
```
Record the number (should be in the thousands; ~19.7k total `log.` occurrences).

- [ ] **Step 2: Apply the rename (bag-prefix positions only)**

```powershell
cd filters
Get-ChildItem -Recurse -Filter *.yml | ForEach-Object {
  $p = $_.FullName
  $c = Get-Content $p -Raw
  # (1) quoted CEL/gjson paths:  "log.foo" -> "event.foo"
  $c = [regex]::Replace($c, '"log\.', '"event.')
  # (2) YAML field values after a step key:  fieldName: log.foo / to: log.foo / from: log.foo / source: log.foo / destination: log.foo / key: log.foo
  $c = [regex]::Replace($c, '(?m)^(\s*(?:- )?(?:fieldName|to|source|destination|key|substring):\s*)log\.', '${1}event.')
  # (3) `from:` list items that are bare field names:  - log.foo
  $c = [regex]::Replace($c, '(?m)^(\s*- )log\.', '${1}event.')
  Set-Content $p $c
}
```
Note: `from:` is sometimes a single value (`from: log.foo`) and sometimes a list of `- log.foo`. Pattern (2) covers the single-value form via the key list; pattern (3) covers list items. `fields:` lists are also `- log.foo` items, covered by (3).

- [ ] **Step 3: Verify (after) — no `log.` field refs remain**

```powershell
cd filters
$after = (Get-ChildItem -Recurse -Filter *.yml | Select-String -Pattern '"log\.|\bto: log\.|\bfrom: log\.|- log\.|\bfields:\s*log\.' | Measure-Object).Count
Write-Output "after: $after"   # must be 0
```

- [ ] **Step 4: Sanity — comments/vendor names untouched**

```powershell
cd filters
Get-ChildItem -Recurse -Filter *.yml | Select-String -Pattern "winlogbeat|syslog|Syslog" | Select-Object -First 5   # still present
Select-String -Path windows/windows-events.yml -Pattern "event\." | Select-Object -First 5                          # renamed refs present
```

- [ ] **Step 5: Verify no filter writes `logx`**

```powershell
cd filters
$lx = (Get-ChildItem -Recurse -Filter *.yml | Select-String -Pattern '\blogx\.' | Measure-Object).Count
Write-Output "logx refs: $lx"   # must be 0 (user-confirmed logx is dead on v11)
```
If non-zero, those refs were missed by the rename (they don't match `log.`) — delete the steps that set `logx.*` (dead output) and re-verify.

- [ ] **Step 6: YAML validity check (catch a broken edit)**

```powershell
cd filters
Get-ChildItem -Recurse -Filter *.yml | ForEach-Object {
  try { python -c "import yaml,sys; yaml.safe_load(open(sys.argv[1]))" $_.FullName } catch { Write-Error "BAD YAML: $($_.Exception.Message)" }
}
```
Expected: no BAD YAML.

- [ ] **Step 7: Commit (per-vendor chunks recommended)**

```bash
git add filters/
git commit -m "refactor(filters): rename log.* bag to event.*"
```
This is a very large diff (~19.7k lines) and one logical change. If the reviewer wants smaller chunks, commit per `filters/<vendor>/`.

---

## Task 7: UTMStack Windows filter — promote object-access fields to origin/target

**Repo:** `utmstack/UTMStack`
**Files:**
- Modify: `filters/windows/windows-events.yml`

**Background (your Q4 decision):** the file-management module must use canonical `origin.*`/`target.*` fields, not bag fields, because `flattened` can't range/sort/aggregate. The `Side` proto already has `file`/`path`/`filename`/`sizeInBytes` plus `host`/`ip`/`user`. The Windows filter already promotes `computer→target.host`, `WorkstationName→origin.host`, `IpAddress→origin.ip`, `SubjectUserName→target.user` — but NOT the object-access fields the file module filters on (`eventCode`, `eventName`, `accessMask`, `processName`, `objectName`). Promote the ones with a clean canonical slot; leave the rest in the flattened bag.

**Promotion map (bag → canonical), applied in this task:**

| Bag field (post-Task-6) | Canonical target | How | Why |
|---|---|---|---|
| `event.eventDataObjectName` | `target.path` | `rename` | accessed object's full path; `Side.path` exists |
| `event.eventDataProcessName` | `origin.file` | `rename` | acting process; `Side.file` exists |
| `event.eventName` | `action` | **grok-copy** (AWS idiom, `aws.yml:207-212`) | dashboard "Top events" charts on `action` (user decision); the copy — NOT a rename — keeps `event.eventName` for the file module's "Event name" column |
| `event.eventCode`, `event.eventDataAccessMask`, `event.computer`, `event.host.os.*`, … | *(keep in bag)* | — | no `Side`/top-level slot; UI uses `event.*` dot-paths (Task 10 makes `term`/`terms` work) |

**Note:** `action` and `target.path`/`origin.file` are top-level/canonical — the scripted rename (Task 6) never touches them, and the promotions must be inserted AFTER Task 6's rename so their sources are already `event.*`.

- [ ] **Step 1: Add the two object-access promotions**

In `filters/windows/windows-events.yml`, after the `event.eventDataProcessName` rename block (search for `to: event.eventDataProcessName`), insert:

```yaml
      # Object-access fields promoted to canonical origin/target so the
      # file-management module can filter on typed (non-flattened) fields.
      - rename:
          from: [event.eventDataObjectName]
          to: target.path
          where: exists("event.eventDataObjectName")
      - rename:
          from: [event.eventDataProcessName]
          to: origin.file
          where: exists("event.eventDataProcessName")
```
(`rename` with `from`/`to` + `where` is the existing idiom in this file — e.g. the `origin.ip` promotion at ~line 142.)

- [ ] **Step 2: Promote `event.eventName` → `action` at the END of the pipeline**

`event.eventName` is set by ~100 `add` steps scattered through the pipeline, so the promotion must run last to catch every computed value. Append at the very end of the `wineventlog` pipeline (after the final `log.authenticationCandidate.*` `add` block, ~line 3410):

```yaml
       # Copy event.eventName to the canonical action field (AWS idiom:
       # aws.yml grok-copy). The copy — not a rename — preserves
       # event.eventName for the file-management module's "Event name" column.
       # The dashboard "Top events" chart aggregates on action.
       - grok:
           source: event.eventName
           patterns:
             - fieldName: action
               pattern: '{{.greedy}}'
           where: exists("event.eventName")
```
**Before appending, confirm nothing in this pipeline deletes or overwrites `action` after this point:**
```powershell
Select-String -Path filters/windows/windows-events.yml -Pattern 'action' | Select-Object -Last 8
```
The only existing `action*` writes in the Windows pipeline are `actionResult` `add` steps (~lines 3100-3137, which run BEFORE the append point) — confirm none target bare `action` and none `delete` it after the insert. If any does, insert the grok-copy after the last `action`-touching step instead of the file end.

- [ ] **Step 3: Verify the promotions landed**

```powershell
Select-String -Path filters/windows/windows-events.yml -Pattern "to: target\.path|to: origin\.file|fieldName: action$"
```
Expected: the two `to:` lines plus exactly one `fieldName: action` (the grok-copy; confirm it is the LAST step in the file).

- [ ] **Step 4: YAML validity**

```powershell
python -c "import yaml,sys; yaml.safe_load(open('filters/windows/windows-events.yml'))"
```
Expected: no error.

- [ ] **Step 5: Commit**

```bash
git add filters/windows/windows-events.yml
git commit -m "feat(windows): promote object-access to origin/target, eventName to action"
```

---

## Task 8: UTMStack rules — scripted `log.`→`event.` rename

**Repo:** `utmstack/UTMStack`
**Files:** every `rules/**/*.yml`

**Background:** Rules reference the bag in three positions: (a) quoted CEL args in `where:` (`equals("log.eventCode","4624")`), (b) `afterEvents`/`correlation` `field:` values and bare list items (`field: log.eventCode`, `- log.user`), (c) inline JSON lists in `groupBy`/`deduplicateBy` (`["log.user"]`). All become `event.*`. Note: `afterEvents.field` is an OpenSearch query field; on a `flattened` field the dotted path `event.eventCode` resolves for `term`/`terms` (the `filter_term`/`filter_not_match` operators the rules use), so the rename is correct.

- [ ] **Step 1: Baseline count**

```powershell
cd rules
$before = (Get-ChildItem -Recurse -Filter *.yml | Select-String -Pattern '"log\.|\bfield:\s*log\.|^(\s*- )log\.|\[\s*"log\.|,\s*"log\.' | Measure-Object).Count
Write-Output "before: $before"
```

- [ ] **Step 2: Apply the rename (one pass, four patterns)**

```powershell
cd rules
Get-ChildItem -Recurse -Filter *.yml | ForEach-Object {
  $p = $_.FullName
  $c = Get-Content $p -Raw
  # (a) quoted CEL/gjson args:  "log.foo" -> "event.foo"
  $c = [regex]::Replace($c, '"log\.', '"event.')
  # (b1) afterEvents/correlation field: values:  field: log.foo
  $c = [regex]::Replace($c, '(?m)^(\s*(?:- )?(?:field):\s*)log\.', '${1}event.')
  # (b2) bare list items:  - log.foo
  $c = [regex]::Replace($c, '(?m)^(\s*- )log\.', '${1}event.')
  # (c) inline JSON list entries:  ["log.user"]  /  ,"log.user"
  $c = [regex]::Replace($c, '\[\s*"log\.', '[ "event.')
  $c = [regex]::Replace($c, ',\s*"log\.', ', "event.')
  Set-Content $p $c
}
```
The `{{.log.foo}}` placeholders inside `value:` (e.g. `value: '{{.log.eventDataProcessID}}'`) are ALSO renamed by pattern (a) — which is correct, because the placeholder is resolved against the previous event's JSON, which now uses `event.`.

- [ ] **Step 3: Verify — no `log.` field refs remain**

```powershell
cd rules
$after = (Get-ChildItem -Recurse -Filter *.yml | Select-String -Pattern '"log\.|\bfield:\s*log\.|^(\s*- )log\.|\[\s*"log\.|,\s*"log\.|\{\{\.log\.' | Measure-Object).Count
Write-Output "after: $after"   # must be 0
```

- [ ] **Step 4: Spot-check known rules**

```powershell
Select-String -Path rules/windows/ntds_extraction_attempts.yml -Pattern "event\.eventCode|event\.eventDataObjectName"
Select-String -Path rules/netflow/beaconing_behavior_detection.yml -Pattern "origin\.bytesSent"
```
Expected: ntds uses `event.*`; netflow still uses `origin.bytesSent` (untouched — it never used the bag).

- [ ] **Step 5: Commit**

```bash
git add rules/
git commit -m "refactor(rules): rename log.* to event.* in CEL/afterEvents/groupBy"
```

---

## Task 9: UTMStack installer — `event:flattened` + `controls:keyword` + typed top-levels

**Repo:** `utmstack/UTMStack`
**Files:**
- Modify: `installer/services/search.go`
- Modify: `installer/setup/apply.go`
- Modify: `installer/samples.go:93` (dormant `logx` emitter)
- Test: `installer/services/search_test.go` (new)

**Background:** The installer is the only code that creates the `v11-log-*` index template. Fresh installs run `InitOpenSearch()` (lock 7); upgrades run `UpdateOpenSearch()` (lock-gated). A composable template only affects **newly created** indices — existing indices need an explicit `_mapping` PUT (non-destructive: adding `event`/`controls`/typed top-levels doesn't touch the legacy dynamic `log` mapping, which old docs still use).

The template pins: `event` as `flattened` (the conflict-proof bag), `controls` as `keyword`, and the top-level Event fields the UI/SQL/sort depend on with real types (`@timestamp` date, `dataType`/`dataSource`/`action`/`protocol`/`actionResult`/`severity`/`connectionStatus` text+keyword, `statusCode` long, `origin.ip`/`target.ip` ip, `origin.port`/`target.port` long, `origin.user`/`target.user`/`origin.host`/`target.host` text+keyword, `origin.file`/`target.path`/`target.file` text+keyword — the last three are the canonical file slots from Task 7 + syslog/o365 filters). `origin`/`target` stay `dynamic:true` beyond those so new canonical fields map naturally (their values are schema-controlled, not vendor-arbitrary).

- [ ] **Step 1: Add the mappings constant to `search.go`**

After the imports in `installer/services/search.go`:

```go
// logIndexMappings pins the canonical fields for v11-log-* documents.
// "event" is flattened: OpenSearch stores every sub-key as a keyword and never
// infers types, so heterogeneous vendor values can no longer produce
// mapper_parsing_exception. "controls" holds compliance control tags.
// Top-level Event fields keep real types so the UI/SQL/sort work on them.
const logIndexMappings = `
{
  "@timestamp": {"type":"date"},
  "dataType": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
  "dataSource": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
  "action": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
  "protocol": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
  "actionResult": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
  "severity": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
  "connectionStatus": {"type":"keyword"},
  "statusCode": {"type":"long"},
  "origin": {"type":"object","dynamic":true,"properties":{
    "ip": {"type":"ip"},
    "port": {"type":"long"},
    "user": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
    "host": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
    "file": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
    "bytesSent": {"type":"double"},
    "bytesReceived": {"type":"double"}
  }},
  "target": {"type":"object","dynamic":true,"properties":{
    "ip": {"type":"ip"},
    "port": {"type":"long"},
    "user": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
    "host": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
    "path": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
    "file": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
    "bytesSent": {"type":"double"},
    "bytesReceived": {"type":"double"}
  }},
  "event": {"type":"flattened"},
  "controls": {"type":"keyword"}
}`
```

- [ ] **Step 2: Wire it into `InitOpenSearch`**

Replace the two template PUTs (currently lines 60–69) with:

```go
	// Create index templates
	templateData := `{"index_patterns":["v11-alert-*","v11-log-*",".utm-*",".utmstack-*"],"template":{"settings":{"index.number_of_shards":1,"index.number_of_replicas":0,"index.mapping.total_fields.limit":50000}}}`
	if err := execCurl(containerID, "PUT", "https://localhost:9200/_index_template/utmstack_indexes", templateData); err != nil {
		return err
	}

	logTemplateData := `{"index_patterns":["v11-log-*"],"template":{"settings":{"index.max_shards":30000},"mappings":{"properties":` + logIndexMappings + `}}}`
	if err := execCurl(containerID, "PUT", "https://localhost:9200/_index_template/utmstack_log_indexes", logTemplateData); err != nil {
		return err
	}
```

- [ ] **Step 3: Wire it into `UpdateOpenSearch` (existing indices get ONLY `event`+`controls`)**

**Why the split:** existing indices already have `origin`/`target` (and some top-levels) dynamically mapped — adding pinned subfield types to them can 400 on a type mismatch (e.g. existing `origin.port` inferred as `integer` vs pinned `long`, or OpenSearch dynamic-strict rejecting new fields under an object with `dynamic` set). `event` and `controls` are NEW keys with no existing mapping, so they're safe to add in-place. Typed top-levels are therefore fresh-install-only (the template); old indices keep their existing dynamic mapping until ILM rotation.

Replace `UpdateOpenSearch` (lines 81–97) with:

```go
func UpdateOpenSearch() error {

	containerID, err := getOpenSearchContainerID()
	if err != nil {
		return err
	}

	// (Re)create the log template with the pinned mappings (affects NEW indices).
	logTemplateData := `{"index_patterns":["v11-log-*"],"template":{"settings":{"index.max_shards":30000},"mappings":{"properties":` + logIndexMappings + `}}}`
	if err := execCurl(containerID, "PUT", "https://localhost:9200/_index_template/utmstack_log_indexes", logTemplateData); err != nil {
		return err
	}

	// Add ONLY the new bag mappings to existing v11-log-* indices (non-destructive:
	// event/controls are new keys — no type clash with the legacy dynamic log mapping).
	// Do NOT add the typed top-levels here: they can conflict with existing dynamic
	// mappings (e.g. origin.port already inferred) and would 400 the whole upgrade.
	if err := execCurl(containerID, "PUT", "https://localhost:9200/v11-log-*/_mapping?allow_no_indices=true",
		`{"properties":{"event":{"type":"flattened"},"controls":{"type":"keyword"}}}`); err != nil {
		return err
	}

	if err := execCurl(containerID, "PUT", "https://localhost:9200/v11-log-*/_settings?allow_no_indices=true", `{"index.max_shards":30000}`); err != nil {
		return err
	}

	if err := execCurl(containerID, "PUT", "https://localhost:9200/v11-alert-*,v11-log-*,.utm-*,.utmstack-*/_settings?allow_no_indices=true", `{"index.mapping.total_fields.limit":50000}`); err != nil {
		return err
	}
	return nil

}
```

- [ ] **Step 4: Add the upgrade lock in `apply.go`**

In `installer/setup/apply.go`, after the existing `} else if utils.GetLock(20260926001, ...)` block (line ~267), add a second lock so the mapping update runs once on next upgrade:

```go
	if utils.GetLock(20261001001, stack.LocksDir) {
		fmt.Print("Pinning event/controls OpenSearch mappings.")
		if err := services.UpdateOpenSearch(); err != nil {
			return "", err
		}

		if err := utils.SetLock(20261001001, stack.LocksDir); err != nil {
			return "", err
		}
		fmt.Println(" [OK]")
	}
```
(`UpdateOpenSearch` is idempotent — both the template PUT and the mapping add are safe to run twice.)

- [ ] **Step 5: Add a JSON-validity test for the constant**

Create `installer/services/search_test.go`:

```go
package services

import (
	"encoding/json"
	"testing"
)

func TestLogIndexMappingsValidJSON(t *testing.T) {
	var m map[string]any
	if err := json.Unmarshal([]byte(logIndexMappings), &m); err != nil {
		t.Fatalf("logIndexMappings is not valid JSON: %v", err)
	}
	ev, _ := m["event"].(map[string]any)
	if ev["type"] != "flattened" {
		t.Fatalf("event should be flattened: %v", m["event"])
	}
	ct, _ := m["controls"].(map[string]any)
	if ct["type"] != "keyword" {
		t.Fatalf("controls should be keyword: %v", m["controls"])
	}
}
```

- [ ] **Step 6: Build + run the test**

```bash
cd installer
go build ./... && go test ./services/ -run TestLogIndexMappingsValidJSON -v
```
Expected: PASS.

- [ ] **Step 7: Manual OpenSearch verification (dev instance)**

```bash
# apply the template as the code does, create a fresh index, index two conflicting docs:
curl -sk -u admin:PASS -XPUT https://localhost:9200/v11-log-testpin/_doc/1 -d '{"event":{"port":443}}'
curl -sk -u admin:PASS -XPUT https://localhost:9200/v11-log-testpin/_doc/2 -d '{"event":{"port":"N/A"},"controls":["NIST-800-53-AC-3"]}'
```
Expected: both return 201 (previously the second would 400 `mapper_parsing_exception`). Then:
```bash
curl -sk -u admin:PASS https://localhost:9200/v11-log-testpin/_search -d '{"query":{"term":{"event.port":"443"}}}'
```
Expected: 1 hit (the first doc).

- [ ] **Step 8: Fix the dormant sample-doc emitter**

`installer/samples.go:93` — its `Log` struct emits the bag under `logx` and `SendSampleData` is currently commented out at `apply.go:305`, but it would produce wrong-shaped docs if re-enabled:
```go
	Log        map[string]any `json:"logx"`
```
→
```go
	Event      map[string]any `json:"event"`
```
(update the `generateSample` body's `Log:` literal key to `Event:` in the same edit).

- [ ] **Step 9: Commit**

```bash
git add installer/services/search.go installer/services/search_test.go installer/setup/apply.go installer/samples.go
git commit -m "feat(installer): pin v11-log-* event bag as flattened + controls keyword"
```

---

## Task 10: UTMStack backend — `SearchUtil` flattened-aware query building

**Repo:** `utmstack/UTMStack`
**Files:**
- Modify: `backend/src/main/java/com/park/utmstack/service/elasticsearch/SearchUtil.java`
- Test: `backend/src/main/java/com/park/utmstack/service/elasticsearch/SearchUtilFlattenedTest.java` (new; tests live in `src/main/java` per repo convention — there is no `src/test/` tree)

**Background:** `SearchUtil.toQuery` is operator→DSL but not type-aware. OpenSearch `flattened` supports: **term, terms, terms_set, prefix, range, match, multi_match, query_string, simple_query_string, exists, wildcard** — but **NOT** `match_phrase`. Today's builders emit `match_phrase` for `IS`/`IS_ONE_OF`/`CONTAIN_ONE_OF` (lines 126, 216, 164) and `query_string` for `CONTAIN` (line 147). So `IS`/`IS_ONE_OF` on `event.*` would fail. Translation for flattened fields:

| Operator | Current DSL | Flattened DSL |
|---|---|---|
| `IS` | `match_phrase` | `term` |
| `IS_NOT` | `must_not match_phrase` | `must_not term` |
| `IS_ONE_OF` | `bool.should(match_phrase)` | `terms` |
| `IS_NOT_ONE_OF` | `must_not bool.should(match_phrase)` | `must_not terms` |
| `CONTAIN_ONE_OF` / `DOES_NOT_CONTAIN_ONE_OF` | `bool.should(match_phrase)` | `terms` / `must_not terms` |
| `START_WITH` | `wildcard v*` | `prefix` |
| `NOT_START_WITH` | `must_not wildcard v*` | `must_not prefix` |
| `ENDS_WITH` | `wildcard *v` | `wildcard` (unchanged — supported on flattened) |
| `NOT_ENDS_WITH` | `must_not wildcard *v` | `must_not wildcard` (unchanged) |
| `CONTAIN` / `DOES_NOT_CONTAIN` | `query_string *v*` | `match` (analyzed-ish on keyword; acceptable) / `must_not match` |
| `EXIST` / `DOES_NOT_EXIST` | `exists` | `exists` (unchanged, already works) |
| `IS_BETWEEN` | `range` (no format) | unchanged — lexical range works on flattened |
| `IS_GREATER_THAN` / `IS_LESS_THAN_OR_EQUALS` | `range` + `.format(date)` | **rejected with 400** — the date `format` clause is invalid on flattened keyword values; the UI never offers them on `event.*` (Task 12) |
| `IS_IN_FIELDS` / `IS_NOT_IN_FIELDS` | `query_string defaultField:"*"` | unchanged (searches all fields incl. flattened subfields) |

Detection is simple: the only flattened field we ship is named `event`, so any filter field starting with `event.` is flattened.

- [ ] **Step 1: Write the failing test**

Create `backend/src/main/java/com/park/utmstack/service/elasticsearch/SearchUtilFlattenedTest.java`:

```java
package com.park.utmstack.service.elasticsearch;

import com.park.utmstack.domain.chart_builder.types.query.FilterType;
import com.park.utmstack.domain.chart_builder.types.query.OperatorType;
import org.junit.jupiter.api.Test;
import org.opensearch.client.opensearch._types.query_dsl.BoolQuery;
import org.opensearch.client.opensearch._types.query_dsl.Query;

import java.util.List;

import static org.junit.jupiter.api.Assertions.*;

public class SearchUtilFlattenedTest {

    private static FilterType f(OperatorType op, String field, Object value) {
        return new FilterType(field, op, value);
    }

    private static BoolQuery bool(Query q) {
        return q.bool();
    }

    @Test
    public void isOnFlattenedFieldUsesTerm() {
        Query q = SearchUtil.toQuery(List.of(f(OperatorType.IS, "event.eventCode", "4624")));
        assertTrue(bool(q).must().stream().anyMatch(c -> c.term() != null
            && c.term().field().equals("event.eventCode")));
        assertTrue(bool(q).must().stream().noneMatch(c -> c.matchPhrase() != null));
    }

    @Test
    public void isOneOfOnFlattenedFieldUsesTerms() {
        Query q = SearchUtil.toQuery(List.of(f(OperatorType.IS_ONE_OF, "event.eventCode", List.of("4624", "4670"))));
        assertTrue(bool(q).must().stream().anyMatch(c -> c.terms() != null
            && c.terms().field().equals("event.eventCode")));
    }

    @Test
    public void startWithOnFlattenedFieldUsesPrefix() {
        Query q = SearchUtil.toQuery(List.of(f(OperatorType.START_WITH, "event.message", "SYS")));
        assertTrue(bool(q).must().stream().anyMatch(c -> c.prefix() != null
            && c.prefix().field().equals("event.message")));
    }

    @Test
    public void textFieldsStillUseMatchPhrase() {
        Query q = SearchUtil.toQuery(List.of(f(OperatorType.IS, "action", "deny")));
        assertTrue(bool(q).must().stream().anyMatch(c -> c.matchPhrase() != null
            && c.matchPhrase().field().equals("action")));
    }

    @Test
    public void endsWithOnFlattenedFieldUsesWildcard() {
        Query q = SearchUtil.toQuery(List.of(f(OperatorType.ENDS_WITH, "event.message", "SYS")));
        assertTrue(bool(q).must().stream().anyMatch(c -> c.wildcard() != null
            && c.wildcard().field().equals("event.message")));
    }

    // --- sort guard (spec 4.6): sorting/aggregating on flattened must 400, not 500 ---

    @Test
    public void sortByFlattenedFieldThrowsApiExceptionBadRequest() {
        var srb = new org.opensearch.client.opensearch.core.SearchRequest.Builder();
        org.springframework.data.domain.Sort sort = org.springframework.data.domain.Sort.by(
            org.springframework.data.domain.Sort.Order.asc("event.eventCode"));
        var ex = assertThrows(com.park.utmstack.util.exceptions.ApiException.class,
            () -> SearchUtil.applySort(srb, sort));
        assertEquals(org.springframework.http.HttpStatus.BAD_REQUEST, ex.getStatus());
    }

    @Test
    public void sortByNonFlattenedFieldDoesNotThrow() {
        var srb = new org.opensearch.client.opensearch.core.SearchRequest.Builder();
        org.springframework.data.domain.Sort sort = org.springframework.data.domain.Sort.by(
            org.springframework.data.domain.Sort.Order.asc("action"));
        assertDoesNotThrow(() -> SearchUtil.applySort(srb, sort));
    }

    @Test
    public void dateFormattedRangeOnFlattenedFieldThrowsBadRequest() {
        var ex = assertThrows(com.park.utmstack.util.exceptions.ApiException.class,
            () -> SearchUtil.toQuery(List.of(f(OperatorType.IS_GREATER_THAN, "event.timestamp", "2026-01-01"))));
        assertEquals(org.springframework.http.HttpStatus.BAD_REQUEST, ex.getStatus());
    }
}
```

**`ApiException` import note:** `com.park.utmstack.util.exceptions.ApiException` (exists — see `backend/.../util/exceptions/ApiException.java`); it carries `getStatus()`, and `GlobalExceptionHandler.handleApiException` maps it to the carried status. `applySort` takes a Spring `Sort` (same type the resource passes in).

- [ ] **Step 2: Run it and confirm it fails**

Run: `cd backend && mvn -s settings.xml -B test -Dtest=SearchUtilFlattenedTest`
Expected: FAIL (`isOnFlattenedFieldUsesTerm`: no `term` clause — it's `match_phrase`).

- [ ] **Step 3: Implement the flattened branches in `SearchUtil.java`**

Add a helper after the `toQuery` method (~line 120):

```java
    /** Flattened fields (event.*) only support term/terms/prefix/match/query_string/range/exists/wildcard. */
    private static boolean isFlattenedField(String field) {
        return field != null && field.startsWith("event.");
    }
```

Then add the flattened short-circuit at the top of each affected builder, after `filter.validate()`:

`buildIsOperator` (line 122):
```java
        if (isFlattenedField(filter.getField())) {
            bool.filter(f -> f.term(t -> t.field(filter.getField())
                .value(org.opensearch.client.opensearch._types.FieldValue.of(String.valueOf(filter.getValue())))));
            return;
        }
```
`buildIsNotOperator` (line 132):
```java
        if (isFlattenedField(filter.getField())) {
            bool.mustNot(n -> n.term(t -> t.field(filter.getField())
                .value(org.opensearch.client.opensearch._types.FieldValue.of(String.valueOf(filter.getValue())))));
            return;
        }
```
`buildIsOneOfOperator` (line 209) and `buildContainOnOneOfOperator` (line 157):
```java
        if (isFlattenedField(filter.getField())) {
            List<FieldValue> values = ((List<?>) filter.getValue()).stream()
                .map(v -> FieldValue.of(String.valueOf(v))).collect(Collectors.toList());
            bool.filter(f -> f.terms(t -> t.field(filter.getField()).terms(q -> q.value(values))));
            return;
        }
```
`buildIsNotOneOfOperator` (line 223) and `buildDoesNotContainOnOneOfOperator` (line 176):
```java
        if (isFlattenedField(filter.getField())) {
            List<FieldValue> values = ((List<?>) filter.getValue()).stream()
                .map(v -> FieldValue.of(String.valueOf(v))).collect(Collectors.toList());
            bool.mustNot(n -> n.terms(t -> t.field(filter.getField()).terms(q -> q.value(values))));
            return;
        }
```
`buildStartWith` (line 331):
```java
        if (isFlattenedField(filter.getField())) {
            bool.filter(f -> f.prefix(p -> p.field(filter.getField())
                .value(String.valueOf(filter.getValue()))));
            return;
        }
```
`buildNotStartWith` (line 341):
```java
        if (isFlattenedField(filter.getField())) {
            bool.mustNot(n -> n.prefix(p -> p.field(filter.getField())
                .value(String.valueOf(filter.getValue()))));
            return;
        }
```
`buildIsGreaterThan` (line 366) and `buildIsLessThanOrEquals` (line 376) emit a `range` with a **date** `.format(Constants.INDEX_TIMESTAMP_FORMAT)` — invalid on flattened keyword values. Reject with a 400:
```java
        if (isFlattenedField(filter.getField())) {
            throw new ApiException("Range comparison is not supported on flattened field [" + filter.getField()
                + "]. Promote the field to a canonical typed field if you need numeric/date ranges.",
                HttpStatus.BAD_REQUEST);
        }
```
(Add the same two-line guard at the top of both methods — neither calls `filter.validate()`, so put the guard as the first statement.)

`buildEndsWith` (line 309) / `buildNotEndsWith` (line 319): no change needed — `wildcard` is supported on flattened fields.

`buildContainOperator` (line 142) / `buildDoesNotContainOperator` (line 194): no change needed — `query_string` is supported on flattened fields (it searches across the flattened sub-values).

**`FieldValue` import:** `SearchUtil.java` already imports `org.opensearch.client.opensearch._types.FieldValue` (line 10). Use it unqualified.

- [ ] **Step 4: Add the sort guard (spec §4.6)**

`flattened` cannot be sorted or aggregated. Guard both sort entry points so an API caller sorting on `event.*` gets a clean **400**, not the current **500** (`RuntimeException` → `ElasticsearchResource` catch-all `INTERNAL_SERVER_ERROR`, which also logs a false application-error event).

Add after the `isFlattenedField` helper:
```java
    private static void ensureNotFlattenedForSort(String field) {
        if (isFlattenedField(field)) {
            throw new ApiException("Sorting is not supported on flattened field [" + field
                + "]. Promote the field to a canonical field (origin./target./top-level) if you need to sort by it.",
                HttpStatus.BAD_REQUEST);
        }
    }
```
(`ApiException` = `com.park.utmstack.util.exceptions.ApiException`, imported at top; `HttpStatus` = `org.springframework.http.HttpStatus`.)

In `applySort` (line 402), inside the `sort.forEach` lambda, call the guard before building the sort:
```java
            else
                sort.forEach(order -> {
                    ensureNotFlattenedForSort(order.getProperty());
                    srb.sort(s -> s.field(f -> f.field(order.getProperty())
                        .order(order.isAscending() ? SortOrder.Asc : SortOrder.Desc)));
                });
```
In `applyPaginationAndSort` (line 424), same guard in its `sort.forEach` (line 441):
```java
            if (sort.isSorted()) {
                sort.forEach(order -> {
                    ensureNotFlattenedForSort(order.getProperty());
                    srb.sort(s -> s.field(f -> f.field(order.getProperty())
                        .order(order.isAscending() ? SortOrder.Asc : SortOrder.Desc)));
                });
            } else {
```
Keep each method's existing try/catch — `ApiException` is a `RuntimeException` and propagates out of the lambda, is caught, and must **not** be re-wrapped. Change **three** catch blocks (both sort methods AND `toQuery` at line 117) from `catch (Exception e) { throw new RuntimeException(...) }` to:
```java
        } catch (ApiException e) {
            throw e;
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
```
The `toQuery` rethrow matters: the date-format range guards (Step 4) throw `ApiException` from inside `buildIsGreaterThan`/`buildIsLessThanOrEquals`, and without it `toQuery`'s catch-all would wrap it into a `RuntimeException` → 500 instead of the tested 400.

- [ ] **Step 5: Run the tests**

Run: `cd backend && mvn -s settings.xml -B test -Dtest=SearchUtilFlattenedTest`
Expected: PASS (all 7 tests, including the two sort-guard tests).

- [ ] **Step 6: Build the backend**

Run: `cd backend && mvn -s settings.xml -B compile`
Expected: BUILD SUCCESS.

- [ ] **Step 7: Commit**

```bash
git add backend/src/main/java/com/park/utmstack/service/elasticsearch/
git commit -m "feat(backend): SearchUtil flattened-aware queries + 400 on event.* sort"
```

---

## Task 11: UTMStack backend + user-auditor — drop `logx`, read `event`; verify DB self-heal

**Repo:** `utmstack/UTMStack`
**Files:**
- Modify: `backend/src/main/java/com/park/utmstack/service/overview/OverviewService.java:265-276`
- Modify: `backend/src/main/java/com/park/utmstack/config/Constants.java:72-73`
- Delete: `backend/src/main/java/com/park/utmstack/domain/shared_types/LogType.java` (dead)
- Modify: `backend/src/main/java/com/park/utmstack/service/impl/UtmAlertServiceImpl.java` (drop dead `getRelatedAlerts`)
- Modify: `backend/src/main/java/com/park/utmstack/domain/shared_types/alert/Event.java:36,51-53`
- Modify: `backend/src/main/java/com/park/utmstack/service/MailService.java:10,373,383`
- Modify: `user-auditor/src/main/java/com/utmstack/userauditor/model/event/Event.java:21`
- Modify: `user-auditor/src/main/java/com/utmstack/userauditor/service/UserService.java` (9 call sites)
- Modify: `user-auditor/src/main/java/com/utmstack/userauditor/service/elasticsearch/Constants.java:14`
- Modify: `user-auditor/src/main/java/com/utmstack/userauditor/service/Impl/ElasticsearchService.java:93`

**Background:** Five consumers read the bag from OpenSearch docs, all move to `event` (user decision: read `event` only, no dual-read):
1. **`OverviewService.topWindowsEvents()`** (the "Top Windows Events" dashboard table) — filters on `logx.wineventlog.log_name.keyword` IS "Security" and terms-aggregates `logx.wineventlog.event_name.keyword` over `v11-log-wineventlog-*`. `logx` is dead v10 (new docs have no `logx` at all), so this table is **already empty for new data**; after the rename it must use `event`/`action`. A terms agg on the flattened `event` is impossible, so the dimension becomes `action.keyword` (populated by Task 7's `eventName→action` promotion) and the filter becomes `event.eventCode` IN the file-event set (flat-keyword `term`, supported on flattened — Task 10).
2. The backend's alert related-logs CSV — `shared_types/alert/Event.java:36` maps `Map<String,Object> log` (used by `MailService.buildRelatedEventCsvAttachment` via `getLogxFlatted()`).
3. `LogType` model (`shared_types/LogType.java:23` maps `logx`) — its only user, `UtmAlertServiceImpl.getRelatedAlerts` (line 137), is a private method with **no callers** (verified); `MailService` imports `LogType` but its CSV path uses the alert `Event` model. Both are dead — delete.
4. user-auditor's `Event` model (`model/event/Event.java:21`), read in `UserService`.
5. user-auditor `ElasticsearchService.searchBySid` (line 93) queries `LOG_WINLOG_EVENT_DATA_TARGET_USER_SID_KEYWORD` (= `log.winlogEventDataTargetUserSid.keyword`) with **`matchPhrase`** — invalid on flattened (no `match_phrase`, no `.keyword` subfield) → `term` on `event.winlogEventDataTargetUserSid`.

Separately, `DefinitionSyncService` (a `CommandLineRunner`) resyncs Postgres from the `filters/`+`rules/` filesystem at every backend start — it updates/creates filters (matched by content hash) and updates rules (matched by `rule_name`), deleting orphans. So the repo YAML rename (Tasks 6/8) self-heals the DB on next backend boot; **no liquibase rename is needed.** The one caveat: a renamed filter no longer matches by content, so it's deleted + re-created with a **new ID** — we verify in Step 8 that nothing references filter IDs by value.

- [ ] **Step 1: Fix `OverviewService.topWindowsEvents` (the live `logx` consumer)**

`backend/.../service/overview/OverviewService.java`, in `topWindowsEvents` (~lines 265-276). The current body:
```java
        List<FilterType> filters = new ArrayList<>();
        filters.add(new FilterType(Constants.timestamp, OperatorType.IS_BETWEEN, List.of(from, to)));
        filters.add(new FilterType(Constants.logxWineventlogLogNameKeyword, OperatorType.IS, "Security"));

        SearchRequest rq = SearchRequest.of(s -> s.size(0).query(SearchUtil.toQuery(filters))
            .index(Constants.SYS_INDEX_PATTERN.get(SystemIndexPattern.LOGS_WINDOWS))
            .aggregations(AGG_NAME, agg -> agg.terms(t -> t.field(Constants.logxWineventlogEventNameKeyword)
                .size(top).order(List.of(Map.of("_count", SortOrder.Desc))))));
```
becomes (filter on the flattened `event.eventCode` for the file/Security object-access events — `term` via Task 10; aggregate on the typed `action.keyword`):
```java
        List<FilterType> filters = new ArrayList<>();
        filters.add(new FilterType(Constants.timestamp, OperatorType.IS_BETWEEN, List.of(from, to)));
        filters.add(new FilterType("event.eventCode", OperatorType.IS_ONE_OF_TERMS,
            List.of("4656", "4658", "4659", "4660", "4661", "4662", "4663", "4664", "4670")));

        SearchRequest rq = SearchRequest.of(s -> s.size(0).query(SearchUtil.toQuery(filters))
            .index(Constants.SYS_INDEX_PATTERN.get(SystemIndexPattern.LOGS_WINDOWS))
            .aggregations(AGG_NAME, agg -> agg.terms(t -> t.field("action.keyword")
                .size(top).order(List.of(Map.of("_count", SortOrder.Desc))))));
```
The event-ID list (4656/4658-4664/4670) mirrors the file-management module's `ALL_FILE_EVENT_ID_NUMBER` (`file-field.constant.ts:794`) so the table shows object-access events, matching the old "log_name = Security" intent. `action.keyword` exists for wineventlog only after Task 7 ships — coordinate (this method returns an empty table if the index predates it; that's the pre-rename state anyway).

- [ ] **Step 2: Remove the dead `logx` constants + dead `LogType`/`getRelatedAlerts`**

First confirm the constant's only consumer is now gone:
```powershell
Get-ChildItem backend -Recurse -Filter *.java | Select-String -Pattern "logxWineventlog" | Select-Object Path,Line
```
Expected: only `Constants.java:72-73`. Delete those two lines in `backend/.../config/Constants.java`.

Then delete the dead model + dead method:
```bash
rm backend/src/main/java/com/park/utmstack/domain/shared_types/LogType.java
```
In `backend/.../service/impl/UtmAlertServiceImpl.java`: delete the `getRelatedAlerts` method (lines 137-159) and its `import com.park.utmstack.domain.shared_types.LogType;` (line 13). In `backend/.../service/MailService.java`: delete `import com.park.utmstack.domain.shared_types.LogType;` (line 10).
Verify no remaining references:
```powershell
Get-ChildItem backend -Recurse -Filter *.java | Select-String -Pattern "\bLogType\b|getRelatedAlerts" | Select-Object Path,Line
```
Expected: zero hits.

- [ ] **Step 3: Rename the backend alert Event model field**

`backend/.../domain/shared_types/alert/Event.java:36` — change:
```java
    private Map<String, Object> log;
```
to:
```java
    private Map<String, Object> event;
```
and the flatten helper at lines 51–53:
```java
    public Map<String, String> getEventFlatted() {
        return MapUtil.flattenToStringMap(event, true);
    }
```
(Lombok `@Data`/`@Getter` regenerates `getEvent()`/`setEvent(...)` automatically; keep the method name change.)

- [ ] **Step 4: Update `MailService` callers of the flatten helper**

`backend/.../service/MailService.java:373,383` — replace `getLogxFlatted()` with `getEventFlatted()`:
```java
            v.forEach(value -> set.addAll(value.getEventFlatted().keySet()));
...
                        cells[i] = value.getEventFlatted().computeIfPresent(headers.get(i), (kk, vv) -> vv);
```

- [ ] **Step 5: Rename the user-auditor model field**

`user-auditor/.../model/event/Event.java:21` — change:
```java
    private Map<String, Object> log;
```
to:
```java
    private Map<String, Object> event;
```
(If the model has explicit getters/setters, rename `getLog()`/`setLog(...)` → `getEvent()`/`setEvent(...)`. Check the file for `@Data`/`@Getter`/`@Setter` first.)

- [ ] **Step 6: Update `UserService` call sites + null-guard**

In `user-auditor/.../service/UserService.java`, replace every `eventLog.getLog()` / `e.getLog()` / `s.getLog()` with `getEvent()` (lines 113, 117, 118, 141, 145, 146, 153, 160, 164, 169). E.g. line 113:
```java
        int eventId = Integer.parseInt(eventLog.getEvent().get("eventCode").toString());
```
**Null-safety:** old docs (pre-migration) have no `event` key, so `getEvent()` can be null on them. Guard at each method entry point that calls `getEvent().get(...)`:
```java
        if (eventLog.getEvent() == null) {
            return null;   // match the method's existing null-contract; read each method first
        }
```

- [ ] **Step 7: Update user-auditor OS constant + its `matchPhrase` query + verify no filter-ID references**

`user-auditor/.../service/elasticsearch/Constants.java:14`:
```java
    public static final String LOG_WINLOG_EVENT_DATA_TARGET_USER_SID_KEYWORD = "event.winlogEventDataTargetUserSid";
```
(flattened has no `.keyword` subfield — drop the suffix.)

**The constant's only consumer uses `matchPhrase`, which flattened rejects** — `user-auditor/.../service/Impl/ElasticsearchService.java:93` (in `searchBySid`):
```java
    shouldList.should(f -> f.matchPhrase(m -> m.field(Constants.LOG_WINLOG_EVENT_DATA_TARGET_USER_SID_KEYWORD).query(St...
```
change to a `term` query (exact keyword match — the correct semantics for a SID):
```java
    shouldList.should(f -> f.term(t -> t.field(Constants.LOG_WINLOG_EVENT_DATA_TARGET_USER_SID_KEYWORD)
        .value(org.opensearch.client.opensearch._types.FieldValue.of(St...
```
(match the existing value expression already passed to `.query(...)`; keep the rest of the lambda unchanged.)

**Note:** user-auditor has its own `SearchUtil` (not the backend's Task 10 one) — it does NOT need the flattened translation because its bag queries all go through this `term` fix + the `UserService` `_source` reads; its other filters target `@timestamp`/`target.user.keyword` (typed). Verify no user-auditor query path builds a `match_phrase` on an `event.*` field:
```powershell
Get-ChildItem user-auditor/src -Recurse -Filter *.java | Select-String -Pattern 'event\.' | Select-Object Path,Line
```
Expected: only `Constants.java:14` and the new `UserService.getEvent()` reads.

Verify no filter-ID-by-value references (the Task 6 rename re-creates filters with new IDs):
```powershell
Get-ChildItem backend,user-auditor,frontend -Recurse -Include *.java,*.ts,*.sql,*.xml | Select-String -Pattern "utm_logstash_filter.*id\s*=\s*\d|filterId\s*=\s*\d" | Select-Object -First 10 Path,Line
```
Expected: no hard-coded filter IDs (if found, note them — a re-created filter gets a new ID).

- [ ] **Step 8: Build both**

```powershell
cd backend; mvn -s settings.xml -B compile
cd ../user-auditor; mvn -B compile
```
Expected: both BUILD SUCCESS.

- [ ] **Step 8: Commit**

```bash
git add backend/ user-auditor/
git commit -m "feat(backend,user-auditor): read event bag, drop dead logx"
```

---

## Task 12: UTMStack frontend — `flattened` type + operator handling

**Repo:** `utmstack/UTMStack`
**Files:**
- Modify: `frontend/src/app/shared/enums/elastic-data-types.enum.ts`
- Modify: `frontend/src/app/shared/components/utm/filters/utm-elastic-filter/shared/util/operator.service.ts`
- Modify: `frontend/src/app/shared/components/utm/filters/utm-elastic-filter/elastic-filter-add/elastic-filter-add.component.ts`
- Modify: `frontend/src/app/shared/components/utm/table/utm-table/dynamic-table/dynamic-table.component.ts`
- Modify: `frontend/src/app/graphic-builder/chart-builder/chart-property-builder/shared/functions/util-field.ts`
- Modify: `backend/.../service/overview/OverviewService.java` (topWindowsEvents)
- Delete: `backend/.../domain/shared_types/LogType.java` (dead `logx` model)
- Modify: `backend/.../service/impl/UtmAlertServiceImpl.java` (drop dead `getRelatedAlerts`)
- Modify: `user-auditor/.../service/Impl/ElasticsearchService.java:93` (searchBySid query)
- Modify: `plugins/soc-ai/internal/alert/transform.go` (direct `LastEvent.Log` struct access)

**Background:** The connector returns OS field types as strings; a `flattened` field arrives as type `"flattened"`. The operator list is derived from the field type (`operator.service.ts`), and the value box is either an `ng-select` (multi-value operators, driven by `applySelectFilter()`) or a plain `<input>`. For a flattened field: add the type, restrict operators to the flattened-compatible set (the backend Task 10 translates `IS`/`IS_NOT`/`IS_ONE_OF`/`IS_NOT_ONE_OF`/`START_WITH`/`NOT_START_WITH`; `IS_ONE_OF_TERMS` already emits `terms` and needs no translation; everything else on `event.*` is unsupported), and force the plain input for single-value operators so the user types a dot-path value (e.g. `eventCode: 4624`).

- [ ] **Step 1: Add the enum value**

`frontend/src/app/shared/enums/elastic-data-types.enum.ts` — add after `KEYWORD = 'keyword'`:
```ts
  FLATTENED = 'flattened'
```

- [ ] **Step 2: Add the flattened operator branch**

`frontend/.../utm-elastic-filter/shared/util/operator.service.ts` — inside `getOperators`, insert before the final `else`:
```ts
      } else if (field.type === ElasticDataTypesEnum.FLATTENED) {
        operators = FILTER_OPERATORS.filter(value =>
          value.operator === ElasticOperatorsEnum.IS ||
          value.operator === ElasticOperatorsEnum.IS_NOT ||
          value.operator === ElasticOperatorsEnum.IS_ONE_OF ||
          value.operator === ElasticOperatorsEnum.IS_NOT_ONE_OF ||
          value.operator === ElasticOperatorsEnum.START_WITH ||
          value.operator === ElasticOperatorsEnum.NOT_START_WITH);
      }
```
**Do NOT include `IS_ONE_OF_TERMS` here**: it is only offered through the multi-value `selectableOperators` path in `applySelectFilter()` (Step 3). If it were available with the plain `<input>` (single string value), the backend's `buildIsOneOfTermsOperator` would cast the value to a `List` and throw a `ClassCastException` → 500. Programmatic filters (Task 13 file module) build `IS_ONE_OF_TERMS` with real array values directly — they don't go through this operator list, so they are unaffected.

- [ ] **Step 3: Use the plain value input for flattened single-value ops**

`frontend/.../elastic-filter-add/elastic-filter-add.component.ts` — at the top of `applySelectFilter()` (line 223), force the plain input for flattened fields:
```ts
  applySelectFilter(): boolean {
    if (this.field && this.field.type === ElasticDataTypesEnum.FLATTENED) {
      return this.selectableOperators.includes(this.formFilter.get('operator').value);
    }
    const fieldSelected = this.formFilter.get('field').value;
    // ...existing body unchanged...
```
This makes `IS`/`IS_NOT`/`START_WITH`/`NOT_START_WITH` (single-value) render the plain `<input>`, while `IS_ONE_OF`/`IS_NOT_ONE_OF` (in `selectableOperators`) render the tag-adding `ng-select` — both send the dot-path value the backend Task 10 turns into a `term`/`terms`.

- [ ] **Step 4: Don't offer sorting on flattened columns (spec §4.6)**

`frontend/.../utm-table/dynamic-table/dynamic-table.component.ts` — `isSortableColumn()` (line 150) returns `true` for every non-text column today, so a `flattened` column header would be clickable and a sort click would hit the backend 400 (Task 10) or, pre-fix, a 500. Make flattened columns non-sortable:

```ts
  isSortableColumn(column: UtmFieldType): boolean {
    if (column.type === ElasticDataTypesEnum.FLATTENED) {
      return false;
    }
    if (column.type === ElasticDataTypesEnum.TEXT || column.type === ElasticDataTypesEnum.STRING) {
      return column.field.includes('.keyword');
    } else {
      return true;
    }
  }
```
(`dynamic-table.component.ts` already imports `ElasticDataTypesEnum` — confirm at the top of the file; the `FLATTENED` member comes from Step 1.)

This is the UI-side gate; the backend 400 (Task 10) is defense-in-depth for direct API calls.

- [ ] **Step 5: Exclude flattened columns from chart dimensions**

`flattened` fields can't be aggregated, so they must not be offered as a chart-builder bucket/dimension. The single hook is `filterFieldAgg`, which feeds the bucket picker (`bucket-aggregation.component.ts:60`):

`frontend/src/app/graphic-builder/chart-builder/chart-property-builder/shared/functions/util-field.ts` — change:
```ts
export function filterFieldAgg(fields: ElasticSearchFieldInfoType[]): ElasticSearchFieldInfoType[] {
  return fields.filter(value => {
    if (value.type === ElasticDataTypesEnum.FLATTENED) {
      return false;
    }
    if (value.type !== ElasticDataTypesEnum.TEXT) {
      return true;
    } else {
      return value.name.includes('.keyword');
    }
  });
}
```
(The metric field picker is already numeric-only — `metric-aggregation.component.ts:137` filters to LONG/FLOAT — so `flattened` is already excluded there; no change needed.)

- [ ] **Step 6: Build the frontend**

```powershell
cd frontend
NODE_OPTIONS=--max_old_space_size=8192 npm run build
```
Expected: build succeeds (Node 14.16.1 per AGENTS.md).

- [ ] **Step 7: Commit**

```bash
git add frontend/
git commit -m "feat(frontend): flattened type, restricted operators, no sort on event.*"
```

---

## Task 13: UTMStack frontend — file-mgmt → origin/target+event, dashboard→`action`, AD, cache-bust

**Repo:** `utmstack/UTMStack`
**Files:**
- Modify: `frontend/src/app/data-management/file-management/shared/enum/file-field.enum.ts`
- Modify: `frontend/src/app/data-management/file-management/shared/const/file-field.constant.ts`
- Modify: `frontend/src/app/data-management/file-management/file-view/file-view.component.ts`
- Modify: `frontend/src/app/dashboard/dashboard-overview/dashboard-overview.component.ts`
- Modify: `frontend/src/app/active-directory/shared/components/event-timeline/event-timeline.component.ts`
- Modify: `frontend/src/app/active-directory/shared/components/active-directory-event/active-directory-event.component.ts`
- Modify: `frontend/src/app/shared/services/elasticsearch/local-field.service.ts`

**Background (your Q4 decision):** the file-management module uses canonical `origin.*`/`target.*` where a slot exists (Task 7 promoted `event.eventDataObjectName→target.path`, `event.eventDataProcessName→origin.file`) and the flattened `event.*` dot-path otherwise. The dashboard "Top events" chart (dead `logx.wineventlog.event_name.keyword`) moves to the top-level typed `action` field (your Q3 call) — but the Windows filter does NOT set `action` (verified: it only sets `actionResult`), so for wineventlog the chart will show empty `action` buckets. Accept that, or point it at `dataType` (always present); this plan uses `action` per your call and notes the empty case. The AD module reads `_source`, so `item.log`→`item.event` is a raw-read rename.

- [ ] **Step 1: Repoint the `FILE_*` enum**

`frontend/.../file-management/shared/enum/file-field.enum.ts`:
```ts
export enum FileFieldEnum {
  FILE_TIMESTAMP_FIELD = '@timestamp',
  FILE_OBJECT_NAME_FIELD = 'target.path',        // promoted (Task 7)
  FILE_PROCESS_NAME_FIELD = 'origin.file',       // promoted (Task 7)
  FILE_HOST_NAME_FIELD = 'origin.host',          // already canonical (WorkstationName)
  FILE_HOST_ID_FIELD = 'id',
  FILE_ACCESS_LIST_FIELD = 'event.eventDataAccessList',
  FILE_ACCESS_MASK_FIELD = 'event.eventDataAccessMask',
  FILE_HANDLE_ID_FIELD = 'event.eventDataHandleId',
  FILE_OBJECT_SERVER_FIELD = 'event.eventDataObjectServer',
  FILE_OBJECT_TYPE_FIELD = 'event.eventDataObjectType',
  FILE_PROCESS_ID_FIELD = 'event.eventDataProcessId',
  FILE_RESOURCE_ATT_FIELD = 'event.eventDataResourceAttributes',
  FILE_SUBJECT_DOMAIN_NAME_FIELD = 'event.eventDataSubjectDomainName',
  FILE_SUBJECT_LOGON_ID_FIELD = 'event.eventDataSubjectLogonId',
  FILE_SUBJECT_USER_NAME_FIELD = 'event.eventDataSubjectUserName',
  FILE_SUBJECT_USER_ID_FIELD = 'event.eventDataSubjectUserSid',
  FILE_EVENT_ID_FIELD = 'event.eventCode',
  FILE_EVENT_NAME_FIELD = 'event.eventName',
  FILE_HOST_ARCHITECTURE_FIELD = 'event.cpuArchitecture',
  FILE_HOST_OS_NAME_FIELD = 'target.host',       // NOTE: filter renames computer->target.host (windows-events.yml:14-16); there is NO event.computer
  FILE_MESSAGE_FIELD = 'event.eventName',
  FILE_NEW_SDDL_FIELD = 'event.eventDataNewSd',
  FILE_OLD_SDDL_FIELD = 'event.eventDataOldSd',
  FILE_HOTS_OS_BUILD_FIELD = 'event.host.os.build',
  FILE_HOST_OS_FAMILY_FIELD = 'event.host.os.family',
  FILE_HOST_OS_PLATFORM_FIELD = 'event.host.os.platform',
  FILE_HOST_OS_VERSION_FIELD = 'event.host.os.version',
  FILE_KEYWORD_FIELD = 'event.keywords',
  FILE_OPCODE_FIELD = 'event.opcode',
  FILE_PROVIDER_GUID_FIELD = 'event.providerGuid',
  FILE_SHARE_NAME_FIELD = 'event.eventDataShareName',
  FILE_SHARE_PATH_FIELD = 'event.eventDataShareLocalPath',
}
```

- [ ] **Step 2: Set the `type` on every `FILE_*` column entry in `file-field.constant.ts`**

The column `type` drives operator availability and the `.keyword` suffix logic (`file-generic-filter.component.ts:108` appends `.keyword` only when `type === 'string'` — which would produce the bogus path `event.foo.keyword` on flattened fields). Set:
- `FILE_OBJECT_NAME_FIELD` (`target.path`) and `FILE_PROCESS_NAME_FIELD` (`origin.file`) → `type: ElasticDataTypesEnum.TEXT` (typed text+keyword in OS; normal `IS`/`CONTAIN`/sort).
- `FILE_HOST_OS_NAME_FIELD` (`target.host`) → `type: ElasticDataTypesEnum.TEXT`.
- Every `event.*` column (`FILE_EVENT_ID_FIELD`, `FILE_ACCESS_MASK_FIELD`, `FILE_OBJECT_TYPE_FIELD`, `FILE_SUBJECT_*`, `FILE_NEW_SDDL_FIELD`, `FILE_OLD_SDDL_FIELD`, `FILE_HOTS_OS_*`, `FILE_KEYWORD_FIELD`, `FILE_OPCODE_FIELD`, `FILE_PROVIDER_GUID_FIELD`, `FILE_SHARE_*`, `FILE_HANDLE_ID_FIELD`, `FILE_RESOURCE_ATT_FIELD`, `FILE_OBJECT_SERVER_FIELD`, `FILE_PROCESS_ID_FIELD`, `FILE_ACCESS_LIST_FIELD`, `FILE_EVENT_NAME_FIELD`, `FILE_MESSAGE_FIELD`, `FILE_HOST_ARCHITECTURE_FIELD`) → `type: ElasticDataTypesEnum.FLATTENED`.
- `FILE_TIMESTAMP_FIELD` (`@timestamp`) → stays `DATE`; `FILE_HOST_ID_FIELD` (`id`) → stays `STRING`.

Audit the three column-list arrays (`FILE_FIELDS`, `FILE_PERMISSION_FIELDS`, `FILE_SHARED_FIELDS` — and `FILE_FILTER_FIELDS` if present) so every entry whose field starts with `event.` carries `FLATTENED`.

- [ ] **Step 3: Fix all flattened `IS_ONE_OF` filters in `file-view.component.ts`**

Every filter in `setFiltersByFileType` that targets a flattened `event.*` field with `IS_ONE_OF` must become `IS_ONE_OF_TERMS` with string values (flattened stores keyword strings; `match_phrase` — what `IS_ONE_OF` emits — fails on flattened). Affected lines: 80 (`FILE_OBJECT_TYPE_FIELD`, value `FILE_OBJECT_TYPE_VALUE`), 89-91 (`FILE_EVENT_ID_FIELD`, `ALL_FILE_EVENT_ID_NUMBER`), 97-99 (`FILE_EVENT_ID_FIELD`, `CREATED_FILE_EVENT_ID_NUMBER` — single value, wrap it), 100-104 (`FILE_ACCESS_MASK_FIELD`), 110-112 (`FILE_EVENT_ID_FIELD`, `DELETED_FILE_EVENT_ID_NUMBER`), 133-135 (`FILE_EVENT_ID_FIELD`, `PERMISSION_FILE_EVENT_ID_NUMBER`), 142-144 (`FILE_EVENT_ID_FIELD`, `SHARE_FILE_EVENT_ID_NUMBER`). Pattern:
```ts
{field: FileFieldEnum.FILE_EVENT_ID_FIELD, operator: ElasticOperatorsEnum.IS_ONE_OF_TERMS, value: ALL_FILE_EVENT_ID_NUMBER.map(String)}
{field: FileFieldEnum.FILE_OBJECT_TYPE_FIELD, operator: ElasticOperatorsEnum.IS_ONE_OF_TERMS, value: FILE_OBJECT_TYPE_VALUE.map(String)}
```
Single scalar values become `[String(x)]`. The `@timestamp` `IS_BETWEEN` is unchanged (date field).

- [ ] **Step 4: Fix the dashboard "Top events" click-through + stored chart**

Two things, because `paramsTopEvent`/`paramEvenTopCLick` (lines 61-69) are **not** the chart's dimension — they are the **click-through navigation params** (where clicking a chart bucket sends the user):

(a) Update the click-through target in `dashboard-overview.component.ts`:
```ts
  paramEvenTopCLick = 'action.keyword';   // was 'logx.wineventlog.event_name.keyword' (dead v10 field)
```
`paramsTopEvent` keeps its `dataType.keyword: 'wineventlog'` filter (still valid) and `action` is now populated by the Task 7 promotion.

(b) The **chart dimension itself** lives in the stored visualization (the `utm_visualization` DB row rendered by `runList`/`visualizationRender`). The seeded chart references the dead `logx.wineventlog.event_name.keyword` and won't show data for new docs. During E2E (Task 15 Step 2, item 7), **re-save the "Top events" chart with dimension `action`** (keyword) via the graphic builder — or, if a seed for that visualization exists in liquibase, update it in the same commit. Note which in the PR.

- [ ] **Step 5: AD raw-read fields `log`→`event`**

`event-timeline.component.ts:143`:
```ts
    const eventCode = item.event && item.event.eventCode;
```
`active-directory-event.component.ts:34`:
```ts
    this.message = this.event ? this.replaceDetail(this.event.event.message) : '';
```

- [ ] **Step 6: Cache-bust the field list**

`frontend/.../services/elasticsearch/local-field.service.ts:25`:
```ts
export const INDEX_PATTERN_FIELD = '_fields_v2';   // was '_fields'
```

- [ ] **Step 7: Build + commit**

```powershell
cd frontend
NODE_OPTIONS=--max_old_space_size=8192 npm run build
git add frontend/
git commit -m "feat(frontend): file module uses origin/target+event, dashboard on action, cache-bust"
```

---

## Task 14: UTMStack plugins — soc-ai direct `.Log` struct access + go-sdk bump

**Repo:** `utmstack/UTMStack`
**Files:**
- Modify: `plugins/soc-ai/internal/alert/transform.go:52,63`
- Modify: `plugins/soc-ai/go.mod` (+ `go.sum`)

**Background:** `soc-ai` accesses the bag **directly as a Go struct field** (`alert.LastEvent.Log`) — not via protojson — so the go-sdk regen (Task 1) alone doesn't rename it for this plugin: it won't compile against the new SDK until fixed. The sweep confirmed it is the ONLY plugin doing this (all other `.Log` hits in Go are the `plugins.Log` gRPC input message — do NOT touch those).

- [ ] **Step 1: Bump go-sdk in `plugins/soc-ai/go.mod`**

```powershell
cd plugins/soc-ai
(Get-Content go.mod -Raw).Replace('github.com/threatwinds/go-sdk v<OLD_TAG>', 'github.com/threatwinds/go-sdk <NEXT_TAG>') | Set-Content go.mod
go mod tidy
```
(`<OLD_TAG>` = whatever it currently pins; `<NEXT_TAG>` = the Task 2 tag.)
Expected: `go build ./...` FAILS at `transform.go` with `alert.LastEvent.Log undefined (type *plugins.Event has no field or method Log)` — this is the compile error that proves the dependency.

- [ ] **Step 2: Fix the struct access**

`plugins/soc-ai/internal/alert/transform.go` — change line 52:
```go
		if alert.LastEvent.Log != nil {
			for key, val := range alert.LastEvent.Log {
```
to:
```go
		if alert.LastEvent.Event != nil {
			for key, val := range alert.LastEvent.Event {
```
and line 63:
```go
						alert.LastEvent.Log[key] = structpb.NewStringValue(cleaned)
```
to:
```go
						alert.LastEvent.Event[key] = structpb.NewStringValue(cleaned)
```
Leave line 64 (`anonymized = append(anonymized, "lastEvent.log."+key)`) — it's an audit-report string, cosmetic only.

- [ ] **Step 3: Build**

```powershell
cd plugins/soc-ai
go build ./...
```
Expected: PASS.

- [ ] **Step 4: Sweep every UTMStack plugin module for the same pattern (safety net)**

```powershell
cd plugins
Get-ChildItem -Recurse -Filter *.go | Select-String -Pattern '\.LastEvent\.Log|\.Events\[\d+\]\.Log|Event\.Log\b' | Select-Object Path,LineNumber,Line
```
Expected: zero hits. If any appear, apply the same `.Log`→`.Event` fix within that module before proceeding.

- [ ] **Step 5: Commit**

```bash
git add plugins/soc-ai/
git commit -m "fix(soc-ai): access Event.Event bag after go-sdk rename"
```

---

## Task 15: UTMStack alerts plugin — update contract-test fixtures

**Repo:** `utmstack/UTMStack`
**Files:** `plugins/alerts/*_test.go` (any fixture building `Event{Log: ...}` or asserting a `log` JSON key) + `plugins/alerts/go.mod`

**Background:** The alerts plugin's contract tests build `plugins.Event` objects and assert on their JSON. After the go-sdk rename they must use `Event{Event: ...}` and assert `event.*`.

- [ ] **Step 1: Bump go-sdk in `plugins/alerts/go.mod`**

```powershell
cd plugins/alerts
(Get-Content go.mod -Raw).Replace('github.com/threatwinds/go-sdk v<OLD_TAG>', 'github.com/threatwinds/go-sdk <NEXT_TAG>') | Set-Content go.mod
go mod tidy
```
(`<OLD_TAG>` = whatever it currently pins; `<NEXT_TAG>` = the Task 2 tag.)

- [ ] **Step 2: Find the affected fixtures**

```powershell
cd plugins/alerts
Get-ChildItem -Recurse -Filter *_test.go | Select-String -Pattern "Log:\s*map\[|Event\{Log:|\"log\"" | Select-Object Path,LineNumber
```

- [ ] **Step 3: Rename `Event{Log:` → `Event{Event:` and `"log"` → `"event"` in fixtures**

Only where it refers to the Event bag (not the `plugins.Log` gRPC message, not logger calls). For each hit, change the struct literal and any JSON key assertion.

- [ ] **Step 4: Run the suite**

```powershell
cd plugins/alerts
go test ./... -count=1
```
Expected: PASS. Pre-existing v11 failures noted in repo memory (`TestWindowsAwarenessRules`, `TestBitdefenderActionResultRaw`, 3 O365 subtests) are NOT ours — confirm no NEW failures.

- [ ] **Step 5: Commit**

```bash
git add plugins/alerts/
git commit -m "test(alerts): update fixtures to event bag rename"
```

---

## Task 16: E2E verification + coordinated rollout + rollback

**No code** — verification + release coordination.

- [ ] **Step 1: Coordinated RC deploy**

Deploy together (one release): go-sdk `<NEXT_TAG>`, EventProcessor base `<NEW_BASE>` (Task 5), UTMStack filters/rules/installer/backend/frontend/plugins (Tasks 6–15). Set `TW_EVENT_PROCESSOR_VERSION_PROD` to `<NEW_BASE>` **only after** the UTMStack release with renamed filters/rules is cut (repo memory: engine + filters ship in separate pipelines and must be coordinated, or you ship new-engine/old-filters = silent breakage).

- [ ] **Step 2: End-to-end test on the RC instance**

```powershell
# 1) fresh ingest: raw ASA syslog -> doc has event, no log, no mapping error
curl -sk -u admin:PASS https://OS/v11-log-*/_search -H 'Content-Type: application/json' -d '{"query":{"match_all":{}},"size":1}' | Select-String -Pattern '"event"|"log"'
# 2) conflict proof: two docs, same event key, different shapes -> both 201
curl -sk -XPUT -u admin:PASS https://OS/v11-log-testconflict/_doc/1 -d '{"event":{"port":443}}'
curl -sk -XPUT -u admin:PASS https://OS/v11-log-testconflict/_doc/2 -d '{"event":{"port":"N/A"}}'
# 3) CEL rule fires on event.* (e.g. Windows ntds rule on event.eventCode/event.eventDataObjectName)
# 4) correlation afterEvents on event.* finds prior event.* docs
# 5) UI: Log Analyzer field picker lists "event" (flattened); filter event.eventCode=4624 returns rows
# 6) UI: file-management columns render origin.file / target.path (sortable); event.* dot-path filter works; event.* columns NOT sortable
# 7) UI: dashboard "Top events" chart renders wineventlog buckets on action (populated by the Task 7 eventName->action promotion)
# 8) user-auditor: /winlogbeat-info-by-filter returns rows (event.eventCode read)
# 9) controls: index a doc with controls:["NIST-800-53-AC-3"]; term query on controls returns it
# 10) confirm cel_analysis.sock exists in the worker (dead cel = no rules fire; repo memory)
# 11) backend "Top Windows Events" table (OverviewService.topWindowsEvents): returns action buckets, not empty
# 12) sort 400: API call sorting on a flattened field -> HTTP 400 (not 500)
curl -sk -o /dev/null -w "%{http_code}`n" -u admin:PASS -XPOST "https://UTM/api/elasticsearch/search?top=10&indexPattern=v11-log-wineventlog-*&pageable.sort=event.eventCode,asc"
# 13) typed sort still works: same call with sort=origin.port,desc -> 200
curl -sk -o /dev/null -w "%{http_code}`n" -u admin:PASS -XPOST "https://UTM/api/elasticsearch/search?top=10&indexPattern=v11-log-*&pageable.sort=origin.port,desc"
```
Expected: 1) `"event"` present, no `"log"` key; 2) both 201; 5-9) as described; 11) action buckets; 12) `400`; 13) `200`.

- [ ] **Step 3: Confirm old-data behavior**

Old `log`-keyed docs still queryable on `log.*` (legacy mapping untouched). New `event`-keyed docs queryable on `event.*`. No reindex (ILM rotates old data; correlation across the boundary is out of scope by design).

- [ ] **Step 4: Verify DB self-heal**

After the backend boots once with the renamed YAML on disk, confirm:
```sql
SELECT count(*) FROM utm_logstash_filter WHERE logstash_filter LIKE '%log.%';  -- 0
SELECT count(*) FROM utm_correlation_rules WHERE rule_definition_def LIKE '%log.%' OR rule_after_events_def LIKE '%log.%';  -- 0
```

- [ ] **Step 5: Rollback runbook (if anything breaks)**

1. Revert `TW_EVENT_PROCESSOR_VERSION_PROD` to the previous base (engine emits `log` again).
2. Revert UTMStack filters/rules/installer/backend/frontend to the pre-rename release.
3. The `event:flattened` + `controls:keyword` + typed top-level OS mappings are harmless no-ops if unused — no need to remove them.
4. No data loss: `log` docs and any `event` docs coexist.

- [ ] **Step 6: Close out**

Mark this plan complete; record in `.opencode/MEMORY.md`: `logx` is dead v10; the event bag is `flattened` under `event` (new data only); `controls` is reserved for compliance; file-management uses `origin.file`/`target.path` + `event.*` dot-paths; backend `SearchUtil` is flattened-aware (term/terms/prefix, 400 on sort/range). Missed-deps closed by the full sweep: `OverviewService.topWindowsEvents` (logx filter+agg → event.eventCode + action.keyword), user-auditor `searchBySid` (matchPhrase→term on event.winlogEventDataTargetUserSid), backend dead `LogType`/`getRelatedAlerts` deleted, `installer/samples.go` dormant `logx` emitter → `event`, `plugins/soc-ai` direct `LastEvent.Log` struct access → `.Event`.
