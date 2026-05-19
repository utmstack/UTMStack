# AI review prompts

Each `*.md` (except this `README.md`) defines a **prompt** that the
`AI review` job runs in parallel against the PR diff. Discovery is by glob:
to add a new review dimension just drop another `.md` here — no YAML
changes needed.

## File format

```markdown
---
name: short-name              # optional, defaults to filename without extension
model: gemini-3-flash-lite    # optional, defaults to workflow's AI_REVIEW_MODEL
---

<instructions for the model>
```

## Output contract

The prompt **must** instruct the model to respond with a JSON object of
this exact shape (no markdown, no code fences, no extra text):

```json
{
  "tier": 1 | 2 | 3,
  "summary": "<one line, max 200 chars>",
  "findings": [
    {
      "severity": "high" | "medium" | "low",
      "file": "<path>",
      "line": <int>,
      "message": "<description and mitigation>"
    }
  ]
}
```

### Tier semantics

- **Tier 1 — Approve.** The change is simple, doesn't touch critical logic,
  no issues detected. The approver aggregates all tiers and, if every
  prompt returns Tier 1, approves the PR.
- **Tier 2 — Changes requested.** Minor issues the author must fix before
  merging: typos, small bugs, out-of-context code, noticeable style
  problems, incomplete mocks or tests.
- **Tier 3 — Engineer review required.** The diff touches critical paths
  (crypto, auth, DB migrations, installer, gRPC contracts, CI/CD, secret
  handling) or introduces changes the model can't judge with sufficient
  confidence. The approver blocks the merge and @mentions the senior
  engineering team.

The approver takes the **maximum tier** across all prompts: if security
returns Tier 1 but architecture returns Tier 3, the final verdict is Tier 3.

### When there's nothing to report

Tier 1, a brief `summary` ("No security concerns detected.") and
`findings: []`. Don't invent findings to seem useful.

### Unparseable responses

If the model returns something that isn't valid JSON matching the schema,
the approver treats it as **Tier 2** with a generic finding asking for
manual review. Fail-safe behaviour — we'd rather block and ask for human
review than let something pass without understanding it.

## Picking a model

- `gemini-3-flash-lite` — fast/cheap, default for broad passes.
- `gemini-3-pro` — better reasoning, for prompts needing deeper analysis
  (architecture, complex logic).
- `claude-sonnet-4-6` / `claude-opus-4-6` — top quality, higher latency
  and cost.
