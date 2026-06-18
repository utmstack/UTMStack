# soc-ai — agentic SOC plugin

An autonomous SOC agent for UTMStack. It connects to **any LLM** and executes
operations against the SIEM through the backend **MCP** server (50+ tools),
running a real tool-calling loop: *plan → call tool → observe → repeat → conclude*.

## Two surfaces

1. **Real-time alert triage** — alerts arrive via the correlation gRPC stream
   (`InitCorrelationPlugin`), are anonymized, and handed to the agent. The agent
   investigates with read-only tools, classifies the alert, and records an
   `[AI SOC Agent] …` assessment note (the Alerts UI parses this format).
   Mutating actions (status change, incident creation) are **gated by config**.

2. **Live operations endpoint** — `POST /api/v1/agent/task` (SSE). Give it a
   free-form task; it operates the SIEM end to end and streams every step:
   ```
   curl -N -X POST localhost:8090/api/v1/agent/task \
     -H "X-Internal-Key: $KEY" -d '{"task":"how many open alerts, grouped by adversary?"}'
   → event: tool_call   data: {...}
   → event: tool_result data: {...}
   → event: final       data: {...}
   ```

## Architecture

```
correlate(alert) ─► queue ─┐
                           ├─► agent.Run ─► LLMClient (openai-compat | anthropic)
POST /agent/task (SSE) ────┘        │   tools
                                    └─► ToolBroker ─► backend /api/v1/mcp (X-Internal-Key)
```

- `internal/agent/` — the engine.
  - `llm.go` — provider-neutral `LLMClient` interface + factory.
  - `openai.go` / `anthropic.go` — the two provider implementations.
  - `mcp.go` — `ToolBroker`, the MCP client to the backend (auth via
    `X-Internal-Key`, which maps to an internal actor that bypasses all gates).
  - `loop.go` — the agent loop, tool-gating, and the live-agent holder.
  - `prompt.go` — triage and ops system prompts.
- `config/` — YAML config (`system_plugins_soc_ai.yaml`) with fsnotify hot-reload
  and AES-CBC/PBKDF2 secret decryption (same scheme as aws/azure/gcp).
- `internal/queue/` — bounded worker queue feeding the triage agent.
- `internal/api/` — HTTP server: `/health`, `/api/v1/analyze` (manual submit),
  `/api/v1/metrics`, `/api/v1/agent/task` (SSE).

## Configuration

See `system_plugins_soc_ai.example.yaml`. Provider/model/secret + behavior flags
come from the YAML; backend URL, internal key and encryption key come from the
platform plugin config.

## Authentication

All SIEM operations go through the backend MCP server. The plugin authenticates
with the platform internal key (`X-Internal-Key`); the backend resolves it to an
internal actor (`backend/pkg/http/middleware/auth.go`) that skips every
permission/role/license gate (`backend/modules/mcp/server.go`).
