# GitHub Actions Workflows — UTMStack

CI/CD for UTMStack v10 and v11. This folder contains two workflow families:

- **PR checks** (`pr-checks.yml`) — a single job that reviews every Pull
  Request targeting `release/**` with AI and posts the findings as a sticky
  comment. Purely informational — see
  [AI review policy](#ai-review-policy-informational-only-no-auto-merge).
  Does **not** run for PRs into `v10`/`v11`/`v12` — those are gated natively
  by a GitHub ruleset, not by this workflow (see below).
- **Deployment pipelines** (`v10-deployment-pipeline.yml`,
  `v11-deployment-pipeline.yml`, `installer-release.yml`) — build, publish
  and deploy artifacts once code is merged.

## Table of contents

- [Release policy](#release-policy)
- [PR Checks](#pr-checks)
- [V10 Deployment Pipeline](#v10-deployment-pipeline)
- [V11 Deployment Pipeline](#v11-deployment-pipeline)
- [Installer Release](#installer-release)
- [Secrets and variables](#secrets-and-variables)
- [Reusable workflows](#reusable-workflows)
- [How to deploy](#how-to-deploy)
- [Troubleshooting](#troubleshooting)

---

## Release policy

Hard rules:

- **Direct push is forbidden** on `release/**`, `v10`, `v11` and `v12`. PR only.
- **Branch protection** — two independent GitHub rulesets, enforced natively
  (not by any Action):
  - **`release/v11*` / `release/v12*`** ("Branch protection — release/v11.x.x
    and v12.x.x"): 0 approvals required, but the `All checks passed` status
    check (produced by `pr-checks.yml`) must be green. This is the only gate
    on these branches — see [AI review policy](#ai-review-policy-informational-only-no-auto-merge).
  - **`v10` / `v11` / `v12`** ("Protect Main"): 1 approval required from
    `@utmstack/administrators` specifically — no status check involved.
    `pr-checks.yml` doesn't even run for PRs into these branches; they're
    fully gated by this ruleset. `@utmstack/administrators`, org admins, and
    a repo admin role can bypass (push directly, no PR).
  - Both rulesets: no force push, no direct push outside the bypass list.
- **Roll-forward only.** No rollbacks. If a release breaks something, ship
  a hotfix that bumps the version (e.g. `v11.2.9` breaks → `v11.2.10`
  fixes it). Feature flags / kill switches are fine for turning features
  off without a redeploy.

### AI review policy (informational only, no auto-merge)

`pr-checks.yml` only triggers on PRs targeting `release/**` — where devs push
and iterate. By the time a PR promotes code to `v10`/`v11`/`v12` it was
already reviewed at the `release/**` stage, and merges there are gated by the
native "Protect Main" GitHub ruleset (1 approval from `@utmstack/administrators`
— see [Release policy](#release-policy)), independent of this workflow.

On `release/**`, the AI **never blocks the merge and never @mentions
anyone**. It always posts a single sticky comment with what it found — good
or bad — and the developer is responsible for reading it, fixing what's
real, and merging it themselves (whoever has repo write access can merge —
see [Release policy](#release-policy); there is no team-membership check in
this workflow). `tier` and finding `severity` only drive the wording/icon of
the comment, never the outcome:

| Signal | Comment |
|--------|---------|
| Tier 1, no findings | ✅ "Clean" |
| Medium/low findings only | ✅ "Minor findings" |
| Any high/critical finding | 🛑 "High/critical findings" |
| Tier 3 | 🛑 "Sensitive area, extra care recommended" |

The same comment also carries a **Go dependencies** section (🟢 up to date /
🔴 pending updates), from a plain `go-deps.sh --check --discover` run — not
an AI call, just the same informational treatment: it's reported, never
blocks.

Whatever the signal, the job always succeeds — the status check stays green.
When the author pushes new commits, the sticky comment is **updated
in-place** (same comment, no stacking) and the workflow re-runs
automatically.

Sensitive paths for Tier 3 are identified by each prompt's own rules (see
`.github/ai-prompts/*.md`) — it's still useful context for the author even
without a mention.

This is a deliberate policy: the team decided AI findings should never gate
a merge or require team-membership approval — only inform the developer, who
merges manually. If that changes again, update
`.github/scripts/post-ai-review-comment.sh` and this section together.

### Auto-merge

Disabled. Every merge — on `release/**` and on `v10`/`v11`/`v12` — is
triggered manually by a human. `pr-checks.yml` has no merge-related logic at
all; it only posts a comment. (An earlier version of this pipeline had a
separate `approver` job that auto-merged on `release/**` when Tier 1 +
author authorized; see git history before this policy changed if that
behavior is ever needed again.)

### Dependabot

Disabled. `.github/dependabot.yml` keeps `updates: []` so Dependabot
reads the file but creates no PRs. Dependency freshness is surfaced instead
by the "Check Go dependencies" step in `pr-checks.yml`, which runs
`bash .github/scripts/go-deps.sh --check --discover` and reports the result
as a section in the sticky comment (informational — see
[AI review policy](#ai-review-policy-informational-only-no-auto-merge)). To
re-enable Dependabot, restore the
previous `updates:` list (see git history of that file).

### Hotfixes

- `hotfix/x` branch from `v11` → PR to `v11` → same checks.
- `urgent` label allows fast-track: if checks pass and the AI approves,
  it merges without waiting for human review even when touching sensitive
  paths.
- **Recommended (not strictly required):** after the hotfix merges to
  `v11`, pull `v11` into the active `release/v11.x.x+1` branch (merge or
  cherry-pick — either works). The fix is **not** lost if you skip this
  step: git already has the hotfix in `v11`'s history, so when
  `release/v11.x.x+1` later merges back, git combines both lines and
  the fix lands automatically. Syncing early is good hygiene because it
  surfaces conflicts in your release branch rather than at the final
  merge, and it lets dev builds include the patched code immediately.

**Version derivation is automatic.** When a hotfix merges to `v11`, the
deployment pipeline compares the candidate BASE (from CM DEV) against
the latest version in CM PROD:

- If BASE > PROD → use BASE as RC tag (normal flow).
- If BASE ≤ PROD → the BASE was already shipped; bump the patch of PROD
  to get the next tag (hotfix flow).

Concrete example: PROD is on `v11.2.9`, dev is still on
`v11.2.9-dev.5` from the cycle that produced it. A hotfix lands on
`v11`. The pipeline sees BASE=`v11.2.9` collides with PROD=`v11.2.9`,
auto-bumps to `v11.2.10`, and the rest of the run (build, installer,
prerelease, CM register) proceeds with that tag. No manual rename, no
config change.

---

## PR Checks

`pr-checks.yml` triggers on any Pull Request targeting `release/**` only.
It does **not** trigger for `v10`/`v11`/`v12` — those are gated natively by
the "Protect Main" GitHub ruleset (see [Release policy](#release-policy)),
so there's nothing for this workflow to add there.

### Architecture

```
PR opened / updated (release/** only)
        │
        ▼
┌───────────────────────────────────────┐
│  review  (single job, "All checks     │
│           passed" — the exact status  │
│           check the ruleset requires) │
│                                        │
│  1. Fetch PR diff (gh pr diff)         │
│  2. Drop rules/filters/definitions     │
│     paths from the diff                │
│  3. Loop: ai-review.sh per prompt in   │
│     .github/ai-prompts/*.md            │
│  4. post-ai-review-comment.sh — build  │
│     + upsert the sticky comment        │
└───────────────────────────────────────┘
        │
        ▼
   always exits 0 — informational only
```

**Key decision:** everything runs in **one job, one runner** — no matrix,
no separate approver, no artifacts uploaded/downloaded between jobs. The
job always succeeds; nothing in it gates the merge. The only thing that
decides who can actually merge into `release/**` is repo write access (see
[Release policy](#release-policy)) — this workflow doesn't check identity
at all.

### Steps

1. **Check Go dependencies** — `actions/setup-go@v5`, then (if `API_SECRET`
   is set) configures git for private `utmstack/*` modules, then runs
   `bash .github/scripts/go-deps.sh --check --discover`. Captures stdout,
   stderr and the exit code to `/tmp/go-deps/`; **never fails the job** —
   `set +e` around the call, same informational treatment as the AI review.
2. **Fetch the diff** — `gh pr diff` (same unified diff the GitHub UI
   shows — no need for `fetch-depth: 0`).
3. **Filter the diff** — drops any file under a `rules/`, `filters/` or
   `definitions/` folder (detection rules / correlation filters / content,
   not code) before the AI ever sees it.
4. **Run AI review, once per prompt** — for each `.md` under
   `.github/ai-prompts/` (except `README.md`), calls
   `.github/scripts/ai-review.sh`, which:
   - Calls the **ThreatWinds AI** `/chat/completions` endpoint with the
     prompt + diff.
   - Validates the response against the `{tier, summary, findings}` schema.
   - Writes the JSON result to `/tmp/ai-results/<prompt-name>.json`.
   - If the model's response isn't valid JSON or the tier isn't 1/2/3,
     writes a fallback with `tier: 2` and a "Manual review recommended"
     finding (fail-safe) — and **always exits 0**.

   **Prompts today:**
   - `security.md` — vulnerabilities introduced in the diff (injection, XSS,
     SSRF, hardcoded secrets, weak crypto, insecure deserialization).
   - `bugs.md` — concrete bugs: nil derefs, races, off-by-one, unhandled
     errors, unclosed resources, inverted logic, out-of-context code.
   - `architecture.md` — architectural deviations: new couplings, logic in
     the wrong layer, broken contracts, unsafe migrations.

   Each prompt declares its own tier policy (Tier 3 covers paths critical
   to that dimension). See `.github/ai-prompts/README.md` for the full
   schema and tier semantics. **To scale:** drop a new `.md` into
   `.github/ai-prompts/` — the loop in `pr-checks.yml` discovers it at
   runtime, no YAML changes needed. **Default model:** `gemini-3-flash-lite`;
   each prompt can pin its own in frontmatter (`model: gemini-3-pro`, etc.).
5. **Post the comment** — `.github/scripts/post-ai-review-comment.sh` reads
   every JSON file plus the Go deps output/exit code, builds one combined
   markdown comment (severity/tier and the deps exit code only drive
   wording/icon), and upserts a single sticky comment (marker
   `<!-- approver:ai -->`, kept from the previous design so in-flight PRs
   keep updating the same comment). Never fails, never blocks, never
   @mentions anyone.

When the author pushes new commits the workflow re-runs (`pull_request:
synchronize`) and the comment is **updated in place** (no stacking) — see
`find_sticky_comment`/`upsert_sticky_comment` in
`post-ai-review-comment.sh`. The PR is never auto-closed.

### Adding a new check

To add another automated review dimension:

- **New AI prompt** — drop a `.md` into `.github/ai-prompts/`. No YAML
  changes needed; the loop in `pr-checks.yml` picks it up automatically.
- **Something that isn't an AI prompt** (e.g. a linter, a test run) — add
  a step to the `review` job in `pr-checks.yml` and, if you want it
  reflected in the sticky comment, extend `post-ai-review-comment.sh` to
  read its output. Since there's no gating logic left, exit codes generally
  shouldn't fail the job unless you specifically want that check to be a
  hard blocker (unlike AI findings, which are deliberately never blocking).

---

## V10 Deployment Pipeline

Triggers:

- Push to `v10` → deploy to **v10-rc**
- Push to `release/v10**` → deploy to **v10-dev**
- Tag `v10.*` → production build

Main jobs:

1. `setup_deployment` — determines environment from the trigger.
2. `validations` — checks permissions (team membership).
3. `build_agent` — Windows/Linux signed agents.
4. `build_agent_manager` — Docker image.
5. `build_*` — microservices (aws, backend, correlation, frontend, etc).
6. `all_builds_complete` — checkpoint.
7. `deploy_dev` / `deploy_rc` — deploy to the corresponding environment.

Permissions: `integration-developers` or `core-developers`.

---

## V11 Deployment Pipeline

Triggers:

- Push to `release/v11**` → deploy to **dev** (auto-incremented version
  `v11.x.x-dev.N`).
- Prerelease created → deploy to **rc** (version `v11.x.x` from the tag).

### Flow

```
Push to release/v11.x.x
        │
        ▼
Auto-increment version (v11.x.x-dev.N)
        │
        ▼
Build & Deploy to DEV
        │
        ▼
Publish to CM Dev → schedule to dev instances


Create Prerelease (tag v11.x.x)
        │
        ▼
Build & Deploy to RC
        │
        ▼
Generate Changelog (AI)
        │
        ▼
Build & Upload Installer
        │
        ▼
Publish to CM Prod → schedule to prod instances
```

Jobs: `setup_deployment`, `validations`, `build_agent`,
`build_utmstack_collector`, `build_agent_manager`, `build_event_processor`,
`build_backend` (Java 17), `build_frontend`, `build_user_auditor`,
`build_web_pdf`, `all_builds_complete`, `generate_changelog` (RC),
`build_installer_rc` (RC), `deploy_installer_dev` (Dev),
`publish_new_version`, `schedule`.

### Environment detection

| Trigger | Environment | CM URL | Service Account | Schedule Var |
|---------|-------------|--------|------------------|--------------|
| Push to `release/v11**` | dev | `https://cm.dev.utmstack.com` | `CM_SERVICE_ACCOUNT_DEV` | `SCHEDULE_INSTANCES_DEV` |
| Prerelease | rc | `https://cm.utmstack.com` | `CM_SERVICE_ACCOUNT_PROD` | `SCHEDULE_INSTANCES_PROD` |

### Version auto-increment (dev)

1. Extracts the base version from the branch (`release/v11.2.1` →
   `v11.2.1`).
2. Queries CM for the latest version.
3. If the base matches, bumps the dev suffix (`-dev.9` → `-dev.10`).
4. If the base differs, starts at `-dev.1`.

### Promotion to Community / Enterprise

- **Community:** manual — promoting the prerelease to `latest` on GitHub
  triggers the auto-deploy.
- **Enterprise:** manual with a checklist (zero crashes for 48h, no open
  P0 issues). The last safety net before touching large customers.

---

## Installer Release

Trigger: GitHub Release published (type `released`).

```
Tag v10.x.x → build v10 installer
Tag v11.x.x → build v11 installer (with ldflags: version, branch, encryption keys)
```

The installer is uploaded as a release asset.

---

## Secrets and variables

### Secrets

| Secret | Used in | Description |
|--------|---------|-------------|
| `API_SECRET` | v10, v11 deploy, installer, pr-checks | GitHub PAT with `read:org` scope. Used by deployment workflows for team-membership validation, and by `pr-checks.yml`'s "Check Go dependencies" step to fetch private `utmstack/*` Go modules via `go list`. **Not used for any team-membership check in `pr-checks.yml`** — that workflow has none. |
| `AGENT_SECRET_PREFIX` | v10, v11 | Agent encryption key |
| `SIGN_CERT` | v10, v11 | Code signing certificate path (it's a `var`) |
| `SIGN_KEY` | v10, v11 | Code signing key |
| `SIGN_CONTAINER` | v10, v11 | Code signing container name |
| `CM_SERVICE_ACCOUNT_PROD` | v11 | Customer Manager service account (prod/rc), JSON `{"id":"...","key":"..."}` |
| `CM_SERVICE_ACCOUNT_DEV` | v11 | Customer Manager service account (dev), JSON `{"id":"...","key":"..."}` |
| `CM_ENCRYPT_SALT` | installer | Installer encryption salt |
| `CM_SIGN_PUBLIC_KEY` | installer | Public key for verification |
| `THREATWINDS_API_KEY` | pr-checks, v11 changelog | ThreatWinds API key for the AI review step and `generate-changelog` |
| `THREATWINDS_API_SECRET` | pr-checks, v11 changelog | ThreatWinds API secret for the AI review step and `generate-changelog` |
| `GITHUB_TOKEN` | All | Provided automatically |

### Variables

| Variable | Used in | Description | Format |
|----------|---------|-------------|--------|
| `SCHEDULE_INSTANCES_PROD` | v11 | Instance IDs for prod/rc scheduling | Comma-separated UUIDs |
| `SCHEDULE_INSTANCES_DEV` | v11 | Instance IDs for dev scheduling | Comma-separated UUIDs |
| `TW_EVENT_PROCESSOR_VERSION_PROD` | v11 | ThreatWinds Event Processor version (prod/rc) | Semver (`1.0.0`) |
| `TW_EVENT_PROCESSOR_VERSION_DEV` | v11 | ThreatWinds Event Processor version (dev) | Semver (`1.0.0-beta`) |

---

## Reusable workflows

**PR checks:**

- `_pr-reusable-go-deps.yml` — a leftover, artifact-based version of the Go
  deps check (uploads `go-deps-result` instead of reporting inline). **Not
  called by `pr-checks.yml`** — `pr-checks.yml` runs
  `bash .github/scripts/go-deps.sh --check --discover` directly as a step
  instead (see [PR Checks](#pr-checks)). Kept only for manually invoking via
  `workflow_call` if you ever need the artifact form again.
- The AI review + comment logic used to be two more reusable workflows
  (`_pr-reusable-ai-review.yml`, `_pr-reusable-approver.yml`); both were
  removed and folded into steps of the single `review` job in
  `pr-checks.yml` — see [PR Checks](#pr-checks).

**Deployment pipelines:**

- `reusable-basic.yml` — generic Docker builds.
- `reusable-golang.yml` — Go microservices.
- `reusable-java.yml` — Java microservices.
- `reusable-node.yml` — frontend / node.
- `reusable-sign-agent.yml` — agent signing.

---

## How to deploy

### V10

**Dev:**

```bash
git checkout release/v10.x.x
# Make changes via PR → merge → auto-deploy to v10-dev
```

**RC:**

```bash
# PR from release/v10.x.x → v10 → merge → auto-deploy to v10-rc
```

**Production:**

```bash
git tag v10.5.0
git push origin v10.5.0
```

### V11

**Dev:**

```bash
# Open a PR against release/v11.2.1 → checks → merge → auto-deploy
# Version auto-incremented (v11.2.1-dev.1, v11.2.1-dev.2, ...)
```

**RC:**

1. GitHub Releases → "Draft a new release".
2. New tag (e.g. `v11.2.1`).
3. Mark as pre-release.
4. Publish.
5. The pipeline builds microservices, generates the AI changelog, uploads
   the installer, publishes to CM, and schedules updates to RC instances.

**Hotfix:**

```bash
git checkout v11
git checkout -b hotfix/auth-bug
# fix → PR to v11 (label `urgent` if applicable) → checks → merge
# Recommended after merge: sync v11 into release/v11.x.x+1
#   git checkout release/v11.x.x+1
#   git merge origin/v11      # or cherry-pick the specific commits
#   git push
```

---

## Troubleshooting

**Permission denied (deployment pipelines, not PR checks):**
- Verify membership in `integration-developers` or `core-developers`. This
  applies to `v10-deployment-pipeline.yml`/`v11-deployment-pipeline.yml`'s
  own `validations` job — `pr-checks.yml` has no team-membership check at
  all (see [AI review policy](#ai-review-policy-informational-only-no-auto-merge)).

**AI review result with tier 2 fallback "Manual review recommended":**
- The model didn't return valid JSON or returned an invalid tier
  (`ai-review.sh`'s fail-safe). It's surfaced prominently in the sticky
  comment but does not block — refine the prompt `.md` or re-run the
  workflow if it was transient.

**My PR against `v10`/`v11`/`v12` doesn't get an AI review comment:**
- Expected. `pr-checks.yml` only triggers for `release/**` PRs — it doesn't
  run at all for `v10`/`v11`/`v12` — see
  [AI review policy](#ai-review-policy-informational-only-no-auto-merge).
  Those branches are gated by the "Protect Main" GitHub ruleset instead.

**Go dependencies section shows 🔴 "Could not inspect ... run 'go mod tidy' there":**
- `go.sum` is out of sync, typically due to local `replace` directives in
  `packages/`. Run `go mod tidy` in the affected module and commit. This
  never fails the job — it's reported in the comment, same as any other
  finding.

**Build failures:**
- Check that all required secrets are configured.
- Verify availability of the `utmstack-signer` runner (required for
  agent signing).

**Version not incrementing:**
- Check that `CM_SERVICE_ACCOUNT_DEV` / `CM_SERVICE_ACCOUNT_PROD` are
  configured and that the CM API is reachable.
- The branch name must follow `release/v11.x.x`.

**Changelog not generated:**
- Only applies to RC (prereleases).
- Verify `THREATWINDS_API_KEY` and `THREATWINDS_API_SECRET` are configured.
- To test locally: export the same secrets and run
  `./scripts/test-generate-changelog.sh v11.2.8` from the repo root
  (auto-detects the previous tag; the wrapper also loads them from a
  local `.env` if present).

---

## Notes

- Docker images are published to `ghcr.io/utmstack/utmstack/*`.
- Agent signing requires the `utmstack-signer` runner.
- Artifacts (agents, collector) have a 1-day retention.
- Dev versions: `v11.x.x-dev.N` (auto-incremented).
- RC versions: the prerelease tag (e.g. `v11.2.1`).
