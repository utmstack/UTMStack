# soc-ai

The SOC agent of UTMStack, as a service of its own (it used to be a plugin of the event processor, in
`plugins/soc-ai`, which is kept only as a reference until this one replaces it).

It does two jobs:

1. **Triage of new alerts.** The backend publishes `{"tenantId", "alertId"}` to the NATS subject
   `socai.alerts` when an alert is created. This service takes them from a durable consumer, gets the
   facts in one call, asks the model for a verdict, investigates once if the model says it cannot tell,
   and writes the `[AI SOC Agent] ...` note the alerts page parses.
2. **The operations chat.** `POST /api/v1/agent/task` streams every step (server-sent events, same
   `tool_call` / `tool_result` / `final` / `error` events as before). A supervisor decides which
   specialist (alerts, incidents, logs, dashboards, compliance, soar, datasources, detection) gets each
   narrow task, and answers from what they report.

## Why it is built this way

The plugin it replaces gave one model every backend tool (close to a hundred small CRUD calls), let it
repeat calls, and fed it the raw documents. It made many calls and concluded nothing. Here:

- **Tools answer questions.** `get_alert_dossier` is one call that returns the alert, its score, the
  related alerts and the recurrence, instead of five calls the model had to join. Results are shortened
  before the model sees them (`shaping.py`).
- **Each specialist sees only its area's tools.** The supervisor has no tools; it only delegates or
  finishes.
- **The loop has limits** (`graph.py`): a repeated call is not run again, the number of calls and steps
  is capped, and near the cap the model is told to answer with what it has.
- **Triage is a pipeline, not a free agent** (`triage.py`): four steps, one model call in the usual case.
- **Changes are opt-in.** Reading and writing the assessment note are the core. Setting a status,
  tagging, opening an incident or touching dashboards, SOAR and the rest needs the capability group on:
  `TRIAGE_GROUPS` for the automatic triage, `CHAT_GROUPS` (or the chat request's `enabled_groups`) for the
  chat.
- **People are hidden from the model** in triage (`anonymize.py`).

## What the backend does for it

- **It queues new alerts with their dossier.** When the event processor tells the backend an alert was
  raised, the backend (`backend/modules/socai/triage.go`) checks that the tenant turned automatic analysis
  on and has quota left, works out the dossier next to the data (the alert's fields, its score, the other
  alerts from the same adversary and target in 7 days, how often the name fired in 30 days:
  `dossier.go`) and publishes `{"tenantId", "alertId", "facts"}` to NATS (`socai.alerts`, stream
  `SOCAI_ALERTS`). Done there it is function calls against the database; from this service it was one
  round trip for each piece. If the dossier cannot be built, the message goes without `facts` and this
  service builds it itself, with the tools.
- **It shows this service a short list of tools** (`backend/modules/mcp/agent_profile.go`): a core of nine
  tools (what investigating an alert takes, and writing the assessment note) plus the capability groups
  the request asks for in the `X-Agent-Groups` header. What is not asked for does not exist for the agent:
  it is neither listed nor callable. Users, tenants, billing, settings, bulk operations, raw SQL and
  anything that deletes are not on it at all. The choice is made from who calls, so it cannot be widened
  from here.

## Configuration (environment)

| Variable | Meaning |
|---|---|
| `BACKEND_URL`, `INTERNAL_KEY` | The backend and the key it accepts on `/api/v1/mcp` |
| `NATS_URL` | Where `socai.alerts` is. Without it only the HTTP API runs |
| `ENCRYPTION_KEY` | The backend's key: it decrypts the model configuration the backend publishes in NATS |
| `LLM_PROVIDER`, `LLM_MODEL`, `LLM_API_KEY`, `LLM_BASE_URL` | A model for development only, used when the backend has published none |
| `TRIAGE_GROUPS` | Groups the automatic triage may use besides the core: `alert_actions` (close and tag alerts), `incident_actions` (open incidents). Empty by default: it only writes the note |
| `CHAT_GROUPS` | Groups the chat gets. All of them by default: `alert_actions`, `incident_actions`, `dashboards`, `compliance`, `soar`, `detection`, `datasources` |
| `MAX_STEPS`, `MAX_TOOL_CALLS`, `TRIAGE_CONCURRENCY` | Limits |

The old names of the settings page (`alerts`, `incidents`, `correlation`) are still accepted for the groups.

## Which model it uses

Each tenant runs on its own model when it has set one on the settings page, and on the instance default
(ThreatWinds' proxy, unless an administrator changed it) otherwise. The backend keeps these in its store
and publishes all of them, encrypted, in the NATS bucket `soc-ai-config` (key `config`) every time they
change and every few minutes (`backend/modules/socai/config_mirror.go`); this service watches the key
(`config_source.py`), so a change on the settings page applies without a restart. Until something is
published there is no model: an alert is retried later and the chat says it is not configured. The
endpoint in the configuration is a full one (`.../chat/completions`); `llm.py` turns it into what the
client libraries want and sends the key as a bearer token, in a header of its own, or not at all, as the
configuration says. The limits of the configuration apply to that tenant's work (steps per task), and
turning on "allow incidents" lets its triage open incidents.

## Not done yet

- Not tried against a real backend, NATS or model: the tests use a scripted model, a fake backend and a
  fake NATS bucket (the backend's side has tests against a real NATS server). The filters the tools send
  (`store.search` on the alerts and logs datasets) follow the alert document as the backend defines it and
  need checking against live data.
- The old plugin's `plugins/soc-ai` is still in the repository as a reference; nothing builds or runs it.

## Tests

```
pip install -e '.[test]' && pytest
```
