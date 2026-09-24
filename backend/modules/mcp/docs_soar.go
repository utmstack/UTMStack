package mcp

// soarFlowGuideDoc is the static half of mcp://utmstack/docs/soar-flow-guide.
// The live half (registered executor types) is appended at read time in
// resources.go so the model always sees what THIS instance can run.
const soarFlowGuideDoc = `# SOAR flow guide — building alert-response DAGs

A SOAR rule ("flow") is a DAG of nodes that runs when an alert matches its
trigger conditions. Build flows with soar.rule.create / soar.rule.update:
pass conditions + roots + nodes in one call. Read an existing flow first
with soar.rule.get (rel_path from soar.rule.list).

## Node kinds — executor vs enrichment

Every node has kind: "executor" or "enrichment". The difference is what
happens to the node's output:

- **executor** — side effect (block IP, send mail, open incident). Its output
  is DISCARDED; downstream nodes cannot read it.
- **enrichment** — data lookup (query API, run query on agent, LLM analysis).
  Its JSON output is stored under the node's id in every descendant's context
  bag, so children can interpolate it: $(<nodeId>.<path>).

Enrichment output rules per executor:
- http: response body must be valid JSON (non-2xx = node fails).
- shell: agent stdout must parse as JSON.
- llm_enrich: normalized to {"result": ...} — downstream may rely on $(<nodeId>.result).

## Executor types and their params

The live list for this instance is appended below this doc. Params shapes:

| type | params | notes |
|---|---|---|
| shell | (none) | uses node-level command/shell/agent fields; runs on endpoint agent |
| http | {"method","url","headers","body","timeoutSec"} | method defaults GET (POST if body set) |
| conditional | {"conditions":[{operator,field,value}]} | AND of predicates over context bag; see Branching |
| llm_enrich / llm_action | {"prompt","page","lang","history"} | SOC-AI; llm_enrich returns JSON result |
| notify | {"message","type"} | in-app notification |
| incident | {"name","description"} | opens incident linked to the alert |
| mail | {"to","cc","subject","body"} | tenant SMTP |

Node-level fields (all kinds): kind, executor, command (shell), shell
("powershell"|"cmd"|"bash"), platform ("windows"|"linux"|"macos"), agent
(hostname; empty = auto-resolve from alert source), excludedAgents (only in
auto-resolve mode), params (JSON), onSuccess[], onError[].

## DAG wiring

- roots: node ids that start at depth 0 when the flow fires. Multiple roots
  run in parallel.
- Each node lists child ids in onSuccess (run when node SUCCEEDS) and onError
  (run when node FAILS). A node with neither is a leaf.
- AND-join: if N parents point at one child, the child waits until ALL N have
  resolved. If ANY parent arrived via its "wrong" branch (e.g. parent failed
  but edge was onSuccess), the child AND its whole subtree are marked DEAD —
  dead-if-any-dead wins.
- maxDepth caps recursion (default 50); cycles are safe but pointless — depth
  guard stops them.

## Interpolation — parent responses into children

Two independent syntaxes, applied to command, params, shell and agent before
execution:

1. $(a.b.c) — gjson path into the merged context bag:
   - alert.* — frozen alert payload (e.g. $(alert.name), $(alert.adversary.ip))
   - <nodeId>.* — output of an ANCESTOR ENRICHMENT node only (e.g. $(lookup.ip2country))
   - Unresolved placeholders are left verbatim.
2. $[variables.NAME] — SOAR variable value; secrets are masked in stored results.

The context bag merges every ancestor's bag plus each ancestor enrichment's
output under its node id; later parents overwrite earlier ones on collision.

## Trigger setup (alerting)

conditions: array of {operator, field, value} evaluated against the incoming
alert by the event processor's active-response plugin; ALL must match (AND).
On match it posts {rulePath, alert} to /soar/rule-executions which starts a
flow run — you do not wire triggers yourself, only declare conditions. Typical:
[{"operator":"IS","field":"name","value":"<exact alert name>"}]. Operators:
IS, IS_NOT, CONTAINS, NOT_CONTAINS, EXISTS, NOT_EXISTS, START_WITH, NOT_START_WITH,
ENDS_WITH, NOT_ENDS_WITH, IS_ONE_OF (value = string[]), IS_NOT_ONE_OF. Use
soar.rule.resolve_filter_values for field suggestions. A rule only fires when
enabled — create with active:true or flip later with soar.rule.set_enabled.

## Branching on success/failure of a step

Use a conditional node as a router: its params.conditions test the context bag
(e.g. an enrichment's verdict). All true → onSuccess children; any false → node
fails → onError children. This is how one trigger fans out into "success path"
and "failure path" subtrees.

## Editing an existing flow

soar.rule.update is FULL REPLACE, not a patch — it overwrites conditions, roots
and the entire nodes map. Always: soar.rule.get(rel_path) → modify the complete
payload in memory → send everything back.

1. **Add a node** — new id in nodes; wire it by appending that id to an
   existing node's onSuccess or onError (or to roots if it starts at depth 0).
2. **Change a node** — edit its entry in place; keep its id stable (other
   nodes' edges and $(<nodeId>...) interpolations reference it).
3. **Remove/replace a node** — delete its entry AND scrub its id from every
   other node's onSuccess/onError and from roots. To replace, keep the id and
   swap kind/executor/params, or delete + add under a new id and rewire edges.
   The server does NOT validate edge consistency at write time — dangling refs
   only surface at run time as dead branches, so you own referential integrity.
4. **Enable/disable** — soar.rule.set_enabled(rel_path, enabled); no payload
   needed. System-owned flows are read-only (update/delete rejected).

## Worked example — block brute-force source IP unless internal

{
  "name": "Block brute force source",
  "conditions": [{"operator":"IS","field":"name","value":"Windows: Multiple remote access login failures"}],
  "roots": ["lookup"],
  "nodes": {
    "lookup": {  // enrichment: child can read $(lookup.*) after this runs
      "kind": "enrichment", "executor": "http",
      "params": {"method":"GET","url":"https://geo.example.com/lookup/$(alert.adversary.ip)"},
      "onSuccess": ["is_internal"], "onError": ["notify_fail"]
    },
    "is_internal": {  // router: true → block path, false → notify path
      "kind": "executor", "executor": "conditional",
      "params": {"conditions":[{"operator":"IS_NOT","field":"lookup.internal","value":"true"}]},
      "onSuccess": ["block"], "onError": ["notify_only"]
    },
    "block": {  // side effect; output discarded
      "kind": "executor", "executor": "shell", "shell": "powershell", "platform": "windows",
      "command": "New-NetFirewallRule -DisplayName 'Blocked_$(alert.adversary.ip)' -Direction Inbound -RemoteAddress '$(alert.adversary.ip)' -Action Block -Protocol Any -Profile Any -Enabled True",
      "onSuccess": ["done"]
    },
    "notify_only": {
      "kind": "executor", "executor": "notify",
      "params": {"message":"Internal IP $(alert.adversary.ip) flagged by $(alert.name) — no block applied"},
      "onSuccess": ["done"]
    },
    "notify_fail": {  // lookup itself failed → onError branch of root
      "kind": "executor", "executor": "notify",
      "params": {"message":"Geo lookup failed for $(alert.adversary.ip)"}
    },
    "done": {  // AND-join: runs only after both block and notify_only succeeded
      "kind": "executor", "executor": "incident",
      "params": {"name":"Blocked $(alert.adversary.ip)","description":"Auto-blocked after $(alert.name)"} 
    } 
  } 
}

Rules of thumb: keep side effects in executor nodes; put every data lookup you
want children to see in an enrichment node; give router nodes both onSuccess and
onError children so both outcomes are handled; end paths in notify/incident so a
human always hears about what happened.`
