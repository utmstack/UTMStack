/**
 * Mock assistant replies. No backend yet — these canned, SIEM-flavored answers
 * exercise every rich block the real SOC-AI will render (callouts, tables, code,
 * and mermaid diagrams) so the UI can be built and reviewed first.
 */

const INCIDENT_FLOW = `Here's how an alert becomes an incident in UTMStack.

> [!TIP]
> Start from **Threat Management → Alerts**, filter by severity, then convert the noisy clusters into a single incident.

## Triage flow

\`\`\`mermaid
flowchart TD
  A[Log ingested] --> B{Correlation rule match?}
  B -- yes --> C[Alert created]
  B -- no --> D[Stored in data lake]
  C --> E{Analyst review}
  E -- false positive --> F[Tune rule]
  E -- real --> G[Convert to incident]
  G --> H[SOAR playbook]
\`\`\`

## Severity guide

| Severity | SLA       | Who responds        |
|----------|-----------|---------------------|
| Critical | 15 min    | On-call + lead      |
| High     | 1 hour    | Tier-2 analyst      |
| Medium   | 8 hours   | Tier-1 analyst      |
| Low      | Best effort | Automated tuning  |

> [!WARNING]
> Converting an alert to an incident **does not** suppress the underlying rule. Tune the rule separately or it will keep firing.

You can script the conversion with the API:

\`\`\`bash
curl -X POST /api/v1/incidents \\
  -H "Authorization: Bearer $TOKEN" \\
  -d '{ "alert_ids": [1024, 1025], "severity": "high" }'
\`\`\``

const AD_AUDIT = `The **User Auditor** tracks Active Directory accounts seen in your Windows security logs (events 4720 / 4726 / 4624).

> [!NOTE]
> Accounts are discovered passively from logs — there's no agent on the DC beyond the log shipper.

| Field      | Source event | Meaning                    |
|------------|--------------|----------------------------|
| Created    | 4720         | Account creation           |
| Last logon | 4624         | Most recent successful auth |
| Deleted    | 4726         | Account deletion           |

\`\`\`mermaid
sequenceDiagram
  participant DC as Domain Controller
  participant Agent as Log Shipper
  participant UTM as UTMStack
  DC->>Agent: Security event 4624
  Agent->>UTM: Forward log
  UTM->>UTM: Upsert ad_user (last_seen)
\`\`\`

> [!DANGER]
> A burst of 4624 failures followed by a success is a classic password-spray signal — check the **With failures** view.`

const DEFAULT = `I'm the SOC assistant (mock). I can explain detections, walk you through the console, and render **diagrams**, \`code\`, tables and callouts.

> [!TIP]
> Try asking *"how does an alert become an incident?"* or *"what does the user auditor track?"*

\`\`\`mermaid
flowchart LR
  Q[Your question] --> R[Retrieval over docs + telemetry]
  R --> A[Grounded answer]
  A --> U[You]
\`\`\``

/** Pick a canned answer based on a few keywords in the question. */
export function mockReply(question: string): string {
  const q = question.toLowerCase()
  if (/incident|alert|triage|severity|soar/.test(q)) return INCIDENT_FLOW
  if (/auditor|ad |active directory|account|4624|logon|user audit/.test(q)) return AD_AUDIT
  return DEFAULT
}
