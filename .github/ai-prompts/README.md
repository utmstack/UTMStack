# AI review prompts

Each `*.md` (except this `README.md`) defines a **prompt** that `pr-checks.yml`
runs against the PR diff, one at a time in a loop (see the "Run AI review"
step). Discovery is by glob: to add a new review dimension just drop another
`.md` here — no YAML changes needed.

## File format

```markdown
---
name: short-name              # optional, defaults to filename without extension
model: gemini-3-flash-lite    # optional, defaults to ai-review.sh's built-in default
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

### Severity drives the comment, not the merge

The AI review is **informational only** — severity/tier decide how a finding
is presented in the sticky comment, never whether the PR can merge. Only
`release/**` targets get this comment (see the main
`.github/workflows/README.md` → "AI review policy"); `pr-checks.yml` doesn't
even trigger for PRs into `v10`/`v11`/`v12`. Pick the lowest severity that
honestly fits — don't inflate a nit.

- **`critical` / `high`** — Something that can break: crashes, nil
  dereferences, data loss/corruption, races/deadlocks, broken or unsafe DB
  migrations, security holes, breaking API/proto/contract changes. Flagged
  prominently (🛑) — the author decides whether to fix before merging.
- **`medium` / `low`** — Real but contained: missing user feedback,
  inconsistent patterns, naming, typos in docs/strings, style. Reported as
  minor findings.

### Tier semantics

`tier` is a coarse signal that only affects the comment's wording/urgency —
it never blocks:

- **Tier 1** — no high/critical issues (minor findings allowed).
- **Tier 2** — at least one high-severity bug worth a look.
- **Tier 3** — sensitive area, extra care recommended. Critical paths (crypto,
  auth, DB migrations, installer, gRPC contracts, CI/CD, secret handling) or
  changes the model can't judge confidently. Flagged in the comment only —
  nobody is @mentioned.

**Nothing here blocks the merge or @mentions anyone.** The comment exists
purely so the author knows what to fix. `pr-checks.yml` has no
team-membership check at all — whoever has write access to the repo can
merge `release/**` PRs; see `.github/workflows/README.md` → "Release policy".

### Routine dependency bumps

`pr-checks.yml` already reports outdated Go modules as its own "Go
dependencies" section in the sticky comment (see
`.github/workflows/README.md` → "AI review policy") — that's informational
too, so a dependency bump is expected to still show up there, not something
these prompts need to flag separately. The `architecture` and `security`
prompts treat a version bump of
existing modules as **Tier 1** — not an architectural/agent-breaking change
and not a vulnerability — and only flag genuine anomalies (new deps, major
breaking jumps, downgrades, known-vulnerable pins, suspicious `replace`
directives). Don't add prompts that re-block on routine bumps.

### When there's nothing to report

Tier 1, a brief `summary` ("No security concerns detected.") and
`findings: []`. Don't invent findings to seem useful.

### Unparseable responses

If the model returns something that isn't valid JSON matching the schema,
`ai-review.sh` writes a fallback: Tier 2 with a `high` finding saying "manual
review recommended". Still informational only — flagged prominently in the
comment, but it does not block the merge.

## Picking a model

- `gemini-3-flash-lite` — fast/cheap, default for broad passes.
- `gemini-3-pro` — better reasoning, for prompts needing deeper analysis
  (architecture, complex logic).
- `claude-sonnet-4-6` / `claude-opus-4-6` — top quality, higher latency
  and cost.
