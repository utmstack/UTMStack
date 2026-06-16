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
      "severity": "critical" | "high" | "medium" | "low",
      "file": "<path>",
      "line": <int>,
      "message": "<description and mitigation>"
    }
  ]
}
```

### Severity drives the merge gate

The approver blocks the merge based on **severity**, not on how many findings
there are. Pick the lowest severity that honestly fits — don't inflate a nit.

- **`critical` / `high` → BLOCKING.** Something that can break: crashes, nil
  dereferences, data loss/corruption, races/deadlocks, broken or unsafe DB
  migrations, security holes, breaking API/proto/contract changes. These stop
  auto-merge.
- **`medium` / `low` → non-blocking WARNING.** Real but contained: missing
  user feedback, inconsistent patterns, naming, typos in docs/strings, style.
  Reported as warnings; the PR can still merge.

### Tier semantics

`tier` is a coarse signal. The gate uses severity for blocking, **plus** Tier 3:

- **Tier 1** — fine to merge; no high/critical issues (minor warnings allowed).
- **Tier 2** — at least one high-severity bug that should be fixed.
- **Tier 3** — engineer review required / could break. Critical paths (crypto,
  auth, DB migrations, installer, gRPC contracts, CI/CD, secret handling) or
  changes the model can't judge confidently. Always blocks and @mentions the
  team.

**The merge is blocked if** any finding is `high`/`critical`, **or** any prompt
returns Tier 3, **or** no review ran. Otherwise the approver approves the PR
(any medium/low findings ride along as warnings).

### Routine dependency bumps

A separate required check (`go_deps`) already enforces that Go modules are on
their latest version, so mass `go.mod` / `go.sum` bumps are routine and
expected. The `architecture` and `security` prompts treat a version bump of
existing modules as **Tier 1** — not an architectural/agent-breaking change
and not a vulnerability — and only flag genuine anomalies (new deps, major
breaking jumps, downgrades, known-vulnerable pins, suspicious `replace`
directives). Don't add prompts that re-block on routine bumps.

### When there's nothing to report

Tier 1, a brief `summary` ("No security concerns detected.") and
`findings: []`. Don't invent findings to seem useful.

### Unparseable responses

If the model returns something that isn't valid JSON matching the schema, the
approver treats it as a blocking `high` finding. Fail-safe behaviour — we'd
rather hold for a human than let something pass without understanding it.

## Picking a model

- `gemini-3-flash-lite` — fast/cheap, default for broad passes.
- `gemini-3-pro` — better reasoning, for prompts needing deeper analysis
  (architecture, complex logic).
- `claude-sonnet-4-6` / `claude-opus-4-6` — top quality, higher latency
  and cost.
