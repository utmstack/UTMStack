# Why agent commands throw gRPC errors while the agent still sends logs and pings

Research report — instance `10.11.11.18` (v12 dev deployment, built 2026-08-31).

---

## TL;DR

The deployed stack is **missing the entire Sep 1–3 gRPC keepalive/timeout fix chain**
(#2523, #2532, #2534, #2535). On that build, `AgentStream` — the **only long-lived
idle gRPC stream** — is torn down by `grpc_read_timeout 900` in nginx every 15 minutes
of inactivity, and the pre-#2535 **unbuffered, blocking result channel** turns one
unclaimed command result into a permanent command freeze. Ping and logs keep working
because they ride different, never-idle streams. Updating the deployment to a build
at or after `6d3ca69f1` (#2535) fixes the known failure modes.

---

## 1. What is deployed (evidence from the live box)

The frontend asset `index.html` was served with
`last-modified: Mon, 31 Aug 2026 04:19:56 GMT`. The matching `release/v12.0.0` commit
(`git rev-list -1 --before="2026-08-31 04:19:56 +0000"`) is **`f5cb983c8`**
(2026-08-31 04:18:55 UTC, one minute before the build). All "deployed build" code-state
claims below are read with `git show f5cb983c8:<file>`.

| Fact | Evidence |
| --- | --- |
| Frontend built **2026-08-31 04:19 GMT** | `curl -kI https://10.11.11.18/` → `last-modified: Mon, 31 Aug 2026 04:19:56` |
| **Deployed commit = `f5cb983c8`** (#2518) | `git rev-list -1 --before=…` on `release/v12.0.0` |
| Stack reinstalled/cert issued **2026-08-22** | TLS cert `notBefore=Aug 22 22:53:51 2026` |
| Front nginx is **1.24.0 (Ubuntu)**, HTTP/2, gRPC passthrough live | `grpc-status: 16` (Unauthenticated) on `/agent.AgentService/AgentStream` and `/agent.PingService/Ping` — the proxy and agent-manager gRPC plane are up |
| Backend healthy | `GET /api/ping` → `{"status":"ok"}` |

**Code state at the deployed commit `f5cb983c8`** (verified with `git show`):

| Component | State at `f5cb983c8` | Fixed in |
| --- | --- | --- |
| agent `agent/agent/conn.go` | `KeepaliveParams` count = **0** | #2534 `760c9afd1` (Sep 2) |
| agent-manager `agent/agent/utmgrpc.go` | `KeepaliveParams` count = **0** | #2534 `760c9afd1` (Sep 2) |
| backend `pkg/agentmanager/client.go` | `KeepaliveParams` count = **0** | #2532 `34bf5ef2b` (Sep 2) |
| nginx `installer/templates/front-end.go` | `grpc_ssl_verify off;` + `grpc_read_timeout 900;` + `grpc_send_timeout 900;` on **all five** gRPC locations; `grpc_socket_keepalive` count = **0** | #2523 `79ff39b9b` (Sep 1, `grpcs`+ssl+keepalive) and #2535 `6d3ca69f1` (Sep 3, `→ 0`) |
| agent-manager `agent/agent/agent_imp.go` | `s.CommandResultChannel[cmdID] = make(chan *CommandResult)` **unbuffered** (line 413); no `reclaimResultSlot`/`tryDeliverResult` | #2535 `6d3ca69f1` (Sep 3) |

So the deployed build is **before every command-path fix** (#2515 `evictIfOwner`/no-`CloseSend`
is the only relevant one already in it — it landed Aug 27, *before* `f5cb983c8`). The
fix commits that specifically address this symptom landed Sep 1–3, days after the build.

## 2. The three channels are independent — that is why logs/ping survive

All agent egress is on **port 443** (`agent/config/const.go`: `AgentManagerPort =
"443"`, `LogAuthProxyPort = "443"`); nginx routes by gRPC service path.

| Channel | gRPC stream | nginx location | upstream | Traffic pattern |
| --- | --- | --- | --- | --- |
| Logs | `plugins.Integration/ProcessLog` on the **`correlationEntry`** connection | `/plugins.Integration/` | `log-input:50051` | continuous — never idle |
| Ping | `agent.PingService/Ping` on the **`agentManagerEntry`** connection | `/agent.PingService/` | `agentmanager:9000` | client heartbeat every **15 s** |
| Commands | `agent.AgentService/AgentStream` on the **same `agentManagerEntry`** connection | `/agent.AgentService/` | `agentmanager:9000` | **idle until the console sends a command** |

Key architectural facts:

1. **Logs are a different TCP connection to a different service**
   (`agent/agent/logprocessor.go` → `GetCorrelationConnection`). A failure on the
   agent-manager connection cannot affect them.
2. **Ping self-refreshes nginx's read timer.** Every 15 s the agent sends a
   `PingRequest`; nginx's `grpc_read_timeout 900` counts *inactivity*, so the ping
   hop is renewed forever. Even when `AgentStream` is broken and redialing, ping
   keeps proving liveness → the UI shows the agent **online**.
3. **AgentStream has no application-level traffic while idle.** With zero keepalive
   pings in the deployed build, nothing resets nginx's 900 s timer on that stream.

## 3. The failure sequence (deployed build)

```
t=0          Agent connects. AgentStream opens, registers in AgentStreamMap[agentID].
t=0..15min   No command sent. Stream is idle. Ping + logs flow on their own hops.
t≈15min      nginx grpc_read_timeout 900 fires on the AgentStream hop
             → RST/FIN to agent-manager (upstream side of that stream dies)
t=15m+       Server: Recv() → io.EOF → WaitForReconnect → goroutine returns
             → AgentStreamMap[agentID] evicted
             Agent: Recv() → "error receiving command from server: EOF"
             → logged once (dedup, agent/agent/grpc_errors.go) → redial
t=15m+x      Console sends a command during the reconnect window:
             handlePanelCommand → AgentStreamMap lookup misses
             → codes.NotFound "agent not found or is disconnected"   ← THE ERROR
             SOAR: mapped to ErrAgentOffline (backend/modules/soar/executor/shell.go)
             Console WS: "error" frame + close 1011
                   (backend/modules/soar/handler/command_ws.go)
```

`grpc_send_timeout 900` cuts the **other direction**: if the agent's connection to
nginx goes quiet outbound (e.g. agent-side NAT/LB aging), the command the server
tries to `Send` fails with
`codes.Internal "failed to send command to agent: ..."`.

### Secondary bug present in the deployed build: the command freeze

Pre-#2535 `agent_imp.go`:

```go
s.CommandResultChannel[cmdID] = make(chan *CommandResult)      // unbuffered
...
resultChan, ok := s.CommandResultChannel[cmdID]
if ok { resultChan <- &CommandResult{...} }                    // blocks with no defer-release
```

If a command is dispatched and the console closes the websocket before the result
arrives, the `AgentStream` goroutine **blocks forever on that send** — because the
channel has no buffer and nothing deletes the slot. Every later command to that
agent then hangs until the 5-minute `codes.DeadlineExceeded` ("agent did not respond
within 5 minutes"), and `AgentStreamMap` still holds the agent, so the agent keeps
pinging fine. One abandoned command = permanently frozen command channel.

# 2535's `tryDeliverResult` (non-blocking select + drop) and `reclaimResultSlot`

(defer on every exit path) exist precisely for this.

Note the asymmetry this creates: **ping and logs are fire-and-forget with their own
streams; only commands need a live registered stream *and* a working result
round-trip *and* a non-frozen AgentStream goroutine.** That is why the failure is
exclusive to commands.

## 4. How to read the actual error string (disambiguation table)

| Console/SOAR message | gRPC code | Meaning |
| --- | --- | --- |
| `agent not found or is disconnected` | `NotFound` | Stream evicted; command hit the reconnect window (primary symptom here) |
| `failed to send command to agent: rpc error: ... EOF / Internal / Unavailable` | `Internal` | Map entry existed but `Send` on a half-dead stream failed |
| `agent did not respond within 5 minutes` | `DeadlineExceeded` | Command sent, result never returned — stream broken one-way or AgentStream goroutine frozen by the unbuffered-channel bug |
| `agentmanager: ProcessCommand open stream: ... Unavailable` | `Unavailable` | Backend→agent-manager network/conn down — **would also kill ping**, so this is *not* the reported symptom |
| `agent offline` (SOAR only) | mapped from `NotFound` | `shell.go:isOfflineError` |

The agent-side smoking gun (one-time dedup'd log, then debug-level):
`error receiving command from server: rpc error: code = Unavailable desc = ...` or
`... EOF`.

## 5. Fix status on the branch

| # | Commit (date) | What it removes |
| --- | --- | --- |
| #2523 | `79ff39b9b` (Sep 1) | `grpcs://` + `grpc_socket_keepalive` on AgentService |
| #2532 | `34bf5ef2b` (Sep 2) | idle backend→agent-manager client conn resets (keepalive 30 s / 10 s, `PermitWithoutStream`) |
| #2534 | `760c9afd1` (Sep 2) | **agent + collector dials and agent-manager + log-input servers** get keepalive — the idle `AgentStream` no longer dies at 900 s |
| #2535 | `6d3ca69f1` (Sep 3) | nginx `grpc_read/send_timeout 900 → 0` on all four agent locations; `tryDeliverResult` + `reclaimResultSlot` (command-freeze fix); socket keepalive on Panel/Collector/Ping |
| #2515 | `161020652` (Aug 27) | `evictIfOwner` (stale-stream eviction race), no more `CloseSend` race — **already in the deployed build** |

Residual (by design, not a bug): if the agent process itself restarts, its
`AgentStreamMap` entry is stale until the new stream registers (seconds). Any
command in that window still returns `NotFound`. The frontend has no
retry-on-NotFound in the console path; SOAR maps it to `ErrAgentOffline` for its
dispatcher to decide.

## 6. Recommended actions

1. **Rebuild/redeploy the affected instance from `release/v12.0.0` ≥ `6d3ca69f1`**
   (i.e. #2534 + #2535 must be in *all four* artifacts: agent binary, agent-manager
   image, backend image, and the nginx template written by the installer). Partial
   upgrades (e.g. new agent but old nginx, or new nginx but old agent-manager)
   leave a gap: keepalive only helps if the *peer answering PING frames* and the
   proxy timeouts are consistent.
2. **Update the endpoint agents** — `dependency.Reconcile` fetches new binaries
   from `/private/dependencies` on restart; the deployed agent predates #2534 and
   has no client keepalive.
3. Verify on the box after deploy:
   - `docker exec <nginx> nginx -T | grep -A4 'agent.AgentService'` → expect
     `grpc_read_timeout 0;` and `grpc_socket_keepalive on;`
   - run a command, wait 16 min of no commands, run another → should succeed.
   - agent log should no longer show the periodic `error receiving command from
     server: ... EOF` every ~15 minutes.
4. If the instance cannot be upgraded soon: a hotfix of *only* the nginx template
   (`grpc_read_timeout 0; grpc_send_timeout 0; grpc_socket_keepalive on;` on all
   four locations) stops the 15-minute teardown, but the command-freeze bug
   (#2535's channel fix) remains in that build.

## 7. Files referenced

- `agent/agent/incident_response.go` — `IncidentResponseStream` recv loop, `commandProcessor`
- `agent/agent/ping_imp.go` — `StartPing`, 15 s ticker
- `agent/agent/logprocessor.go` — `ProcessLogs` on the separate correlation conn
- `agent/agent/conn.go` — `agentManagerEntry` / `correlationEntry` (keepalive added in #2534)
- `agent/agent/grpc_errors.go` — `HandleGRPCStreamError` (EOF → reconnect, dedup logging)
- `agent/config/const.go` — both ports = `443`
- `agent-manager/agent/agent_imp.go` — `AgentStream`, `handlePanelCommand`, `tryDeliverResult`, `reclaimResultSlot` (last two: #2535)
- `agent-manager/agent/utmgrpc.go` — server keepalive (#2534)
- `agent-manager/utils/reconnect.go` — `WaitForReconnect`
- `agent-manager/config/global_const.go` — key/connection-key/internal-key route tables
- `backend/pkg/agentmanager/client.go` — `ProcessCommand`, `ProcessCommandStream` (keepalive: #2532; no `CloseSend`: #2515)
- `backend/modules/soar/handler/command_ws.go` — console command websocket; writes `grpc error` frames (line ~250)
- `backend/modules/soar/executor/shell.go` — SOAR path; `isOfflineError`
- `installer/templates/front-end.go` — nginx gRPC locations (900 s vs 0)
