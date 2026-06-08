# SOAR Plugin

Reacts to every alert pushed through the SDK pipeline, matches it against the
active rules served by the backend (`GET /api/soar/rules?active.equals=true`),
records each match via `POST /api/soar/rule-executions`, and dispatches every
PENDING execution to its target agent via the `agent-manager`
`PanelService.ProcessCommand` bidi stream — updating execution state via
`PATCH /api/soar/rule-executions/:id`.

The plugin holds no DB connection. All SOAR state lives in the backend; the
plugin reads/writes it over the internal HTTP API authenticated with
`X-Internal-Key`. Agent resolution by OS platform is served by
`GET /api/soar/agents?platform=...`, which queries the new `datasources` table
(replacing the legacy `utm_network_scan` schema).

## Configuration

Reads from the shared `com.utmstack` config namespace:

- `backend` — backend base URL (host:port or full URL)
- `internalKey` — internal-key shared secret for backend auth (sent as `X-Internal-Key`)
- `com.utmstack.soar.agentManager.host`, `com.utmstack.soar.agentManager.port` — `agent-manager` gRPC endpoint

Reads `INTERNAL_KEY` from the environment for `agent-manager` gRPC auth.

## Build

```bash
go build -o com.utmstack.soar.plugin -v .
```
