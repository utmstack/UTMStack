# GitHub Actions Workflows — UTMStack

CI/CD for UTMStack v10 and v11. This folder contains two workflow families:

- **PR checks** (`pr-checks.yml` + `_pr-reusable-*.yml`) — validate every
  Pull Request before merge. The only gate into code on `release/**`,
  `v10` and `v11`.
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
- [Approver GitHub App setup](#approver-github-app-setup)
- [Reusable workflows](#reusable-workflows)
- [How to deploy](#how-to-deploy)
- [Troubleshooting](#troubleshooting)

---

## Release policy

Hard rules:

- **Direct push is forbidden** on `release/**`, `v10` and `v11`. PR only.
- **Branch protection** is enabled on those branches: PR required, status
  checks green (`All checks passed`), no force push.
- **Roll-forward only.** No rollbacks. If a release breaks something, ship
  a hotfix that bumps the version (e.g. `v11.2.9` breaks → `v11.2.10`
  fixes it). Feature flags / kill switches are fine for turning features
  off without a redeploy.

### Tiered approval model

The team is small (3 devs + 2 part-time seniors), so the AI can approve
and merge on its own in most cases. Seniors only get involved when the
cost of being wrong is high.

The **final tier** of a PR is decided by the approver, taking the maximum
across all AI prompts (see [PR Checks](#pr-checks)).

| Tier | Meaning | Approver action |
|------|---------|-----------------|
| **1** | Simple change, AI detects no issues, deps OK. | ✅ Sticky "Approved" comment + (when the approver App is configured) formal `APPROVE` review. Status check green. |
| **2** | Minor issues the author should fix before merging (typos, small bugs, out-of-context code). | ⚠️ Sticky comment with the findings list + formal `REQUEST_CHANGES` review. Status check red. |
| **3** | Touches critical paths (crypto, auth, migrations, installer, gRPC contracts, CI/CD) or the model can't judge with confidence. | 🛑 Sticky comment @mentioning the handles configured in `tier3_reviewers` + formal `REQUEST_CHANGES` review. Status check red. |

When the author pushes new commits, the sticky comments are **updated
in-place** (same comment, no stacking) and the workflow re-runs
automatically. A blocked PR is **never auto-closed** — it stays open
waiting for the fixes.

Sensitive paths for Tier 3 are identified by each prompt's own rules (see
`.github/ai-prompts/*.md`). In the future this could be reinforced with
`CODEOWNERS` for additional per-path gates.

### Auto-merge

The approver enables GitHub's native auto-merge **only** when **all** of
the following hold:

- Target branch matches `release/**` (PRs to `v10` / `v11` stay manual
  so production deploys are always intentional).
- `deps_failed == false`.
- `max_tier == 1` across every AI prompt.
- PR author is in `@utmstack/administrators` or `@utmstack/core-developers`.
- The approver GitHub App is configured (`APPROVER_APP_ID` + `APPROVER_PRIVATE_KEY` secrets present).

Auto-merge does NOT merge immediately — it queues the merge until every
branch-protection requirement is satisfied. If another check fails later
or a human leaves `REQUEST_CHANGES`, the merge stays pending.

### Dependabot

Disabled. `.github/dependabot.yml` keeps `updates: []` so Dependabot
reads the file but creates no PRs. Dependency freshness is enforced via
the `go_deps` check on every PR. To re-enable Dependabot, restore the
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

`pr-checks.yml` triggers on any Pull Request whose target is:

- `release/**` (any release branch, v10 or v11)
- `v10`
- `v11`

### Architecture

```
PR opened / updated
        │
        ├─────────────────┬─────────────────┐
        ▼                 ▼                 │
   ┌─────────┐      ┌─────────────┐         │
   │ go_deps │      │  ai_review  │         │
   │ (repo)  │      │  (matrix    │         │
   │         │      │   per       │         │
   │         │      │   prompt)   │         │
   └────┬────┘      └──────┬──────┘         │
        │                  │                 │
        │ artifact         │ artifacts       │
        │ go-deps-result   │ ai-review-*     │
        ▼                  ▼                 │
   ┌──────────────────────────────┐         │
   │           approver           │         │
   │  - reads artifacts           │         │
   │  - decides final tier        │         │
   │  - posts sticky comments     │         │
   │  - (optional) formal review  │         │
   └──────────────┬───────────────┘         │
                  │                          │
                  ▼                          ▼
          all_checks_passed   ← single status check branch protection requires
```

**Key decision:** the producers (`go_deps`, `ai_review`) **always exit
green**. They only upload artifacts. The `approver` is the single source
of truth — it consolidates results, decides the tier (maximum across all
AI prompts), posts comments, and passes or fails the final check.

### `go_deps`

Single job, no matrix, no change detection. Runs:

```bash
bash .github/scripts/go-deps.sh --check --discover
```

It discovers **every** `go.mod` in the repo (excluding `vendor/`,
dot-directories and `node_modules/`) and fails if any explicit **direct
dependency** has a newer version available. The script also detects
out-of-sync `go.sum` files (typically caused by local `replace` directives
in `packages/`) and reports them all at once.

The job uploads its stdout and exit code as the `go-deps-result` artifact.
The approver reads it and, if exit code != 0, posts the sticky comment
"Go dependencies check failed" with the script output embedded.

**Expected dev fix:** run
`bash .github/scripts/go-deps.sh --update --discover` locally, commit the
updated `go.mod` / `go.sum`, push.

### `ai_review`

Matrix with one job per `.md` under `.github/ai-prompts/` (except
`README.md`). Each job:

1. Fetches the diff via `gh pr diff` (same unified diff the GitHub UI
   shows — no need for `fetch-depth: 0`).
2. Calls the **ThreatWinds AI** `/chat/completions` endpoint with the
   prompt and the diff.
3. Validates the response against the `{tier, summary, findings}` schema.
4. Uploads the JSON as the `ai-review-<name>` artifact.

If the model's response isn't valid JSON or the tier isn't 1/2/3, the
script writes a fallback with `tier: 2` and a "Manual review recommended"
finding (fail-safe).

**Initial prompts:**

- `security.md` — vulnerabilities introduced in the diff (injection, XSS,
  SSRF, hardcoded secrets, weak crypto, insecure deserialization).
- `bugs.md` — concrete bugs: nil derefs, races, off-by-one, unhandled
  errors, unclosed resources, inverted logic, out-of-context code.
- `architecture.md` — architectural deviations: new couplings, logic in
  the wrong layer, broken contracts, unsafe migrations.

Each prompt declares its own tier policy (Tier 3 covers paths critical
to that dimension). See `.github/ai-prompts/README.md` for the full
schema and tier semantics.

**To scale:** drop a new `.md` into `.github/ai-prompts/`. Discovered at
runtime — no YAML changes needed.

**Default model:** `gemini-3-flash-lite`. Each prompt can pin its own
model in frontmatter (`model: gemini-3-pro`, etc.).

### `approver`

Single job that **depends on `go_deps` and `ai_review`**. Steps:

1. Downloads every PR-check artifact.
2. Reads `go-deps-result/exit_code.txt` → determines `deps_failed`.
3. Reads each `ai-review-*/result.json` → takes the **max tier** as the
   AI verdict.
4. **Sticky comments** with invisible HTML markers
   (`<!-- approver:deps -->`, `<!-- approver:ai -->`,
   `<!-- approver:permission -->`):
   - If deps failed → upsert "Go dependencies check failed" comment with
     the script output. Otherwise delete it if a previous run posted one.
   - Always upsert the "AI review" comment with the final tier + findings.
   - These two are posted **regardless of who opened the PR** — even
     unauthorized contributors get useful feedback.
5. **Permission check (LAST gate).** Looks up `github.actor` against the
   GitHub teams `administrators` and `core-developers` of the
   `utmstack` org via `API_SECRET`. Notably **does NOT** include
   `integration-developers`. If the author is in neither team:
   - Upsert "⛔ Permission denied" comment @mentioning
     `@utmstack/administrators`.
   - Skip the formal `APPROVE` review (always `REQUEST_CHANGES`).
   - Skip auto-merge.
   - Exit 1.
6. **(Optional) Formal PR review** when the approver App is installed
   (see [Approver GitHub App setup](#approver-github-app-setup)):
   - Tier 1 + deps OK + authorized → `APPROVE`.
   - Anything else → `REQUEST_CHANGES`.
7. **Auto-merge** — only when **all** of: deps OK, Tier 1, authorized,
   AND `BASE_REF` starts with `release/`. Calls
   `gh pr merge --auto --<method>` (default `squash`). This uses
   GitHub's native auto-merge, so the actual merge waits until **every**
   branch-protection requirement is satisfied (other checks green, no
   pending human reviews). PRs targeting `v10` / `v11` never auto-merge
   — those branches stay manually merged so deploys are intentional.
8. **Exit code:** 0 only if everything is OK; 1 if deps failed,
   tier ≥ 2, or author unauthorized.

When the author pushes new commits the workflow re-runs and the comments
are **updated in place** (no stacking). The PR is never auto-closed —
it stays open waiting for the author's fixes.

### Adding a new check

The architecture is designed to scale. To add, for example, a test check:

1. Create `.github/workflows/_pr-reusable-<name>.yml` that runs the check
   and uploads an artifact with the result (ideally JSON).
2. Call the reusable from `pr-checks.yml` as another job.
3. Add that job to the `approver`'s `needs:` (and to `all_checks_passed`).
4. Extend `approver.sh` to read the new artifact and factor it into the
   final verdict.

To add a new AI prompt **no YAML changes are needed** — just drop a `.md`
into `.github/ai-prompts/`.

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
| `API_SECRET` | All, pr-checks | GitHub PAT with `read:org` scope. Used by deployment workflows for team-membership validation and by the `approver` job to check that the PR author belongs to `administrators` or `core-developers`. |
| `AGENT_SECRET_PREFIX` | v10, v11 | Agent encryption key |
| `SIGN_CERT` | v10, v11 | Code signing certificate path (it's a `var`) |
| `SIGN_KEY` | v10, v11 | Code signing key |
| `SIGN_CONTAINER` | v10, v11 | Code signing container name |
| `CM_SERVICE_ACCOUNT_PROD` | v11 | Customer Manager service account (prod/rc), JSON `{"id":"...","key":"..."}` |
| `CM_SERVICE_ACCOUNT_DEV` | v11 | Customer Manager service account (dev), JSON `{"id":"...","key":"..."}` |
| `CM_ENCRYPT_SALT` | installer | Installer encryption salt |
| `CM_SIGN_PUBLIC_KEY` | installer | Public key for verification |
| `THREATWINDS_API_KEY` | pr-checks, v11 changelog | ThreatWinds API key for `ai_review` and `generate-changelog` |
| `THREATWINDS_API_SECRET` | pr-checks, v11 changelog | ThreatWinds API secret for `ai_review` and `generate-changelog` |
| `APPROVER_APP_ID` | pr-checks | GitHub App ID for the approver bot. See [Approver GitHub App setup](#approver-github-app-setup). Without this, the approver runs in comments-only mode (no formal review, no auto-merge). |
| `APPROVER_PRIVATE_KEY` | pr-checks | GitHub App private key (full `.pem` content, multi-line) paired with `APPROVER_APP_ID`. |
| `GITHUB_TOKEN` | All | Provided automatically |

### Variables

| Variable | Used in | Description | Format |
|----------|---------|-------------|--------|
| `SCHEDULE_INSTANCES_PROD` | v11 | Instance IDs for prod/rc scheduling | Comma-separated UUIDs |
| `SCHEDULE_INSTANCES_DEV` | v11 | Instance IDs for dev scheduling | Comma-separated UUIDs |
| `TW_EVENT_PROCESSOR_VERSION_PROD` | v11 | ThreatWinds Event Processor version (prod/rc) | Semver (`1.0.0`) |
| `TW_EVENT_PROCESSOR_VERSION_DEV` | v11 | ThreatWinds Event Processor version (dev) | Semver (`1.0.0-beta`) |

---

## Approver GitHub App setup

The `approver` job uses a GitHub App (instead of a personal PAT) to leave
formal PR reviews and enable auto-merge. Pros:

- Per-run installation token, valid for ~1 hour, auto-revoked when the
  job ends. No long-lived credential in the repo.
- The App acts as its own identity, so it can `APPROVE` PRs opened by any
  human contributor — including the workflow's own author (GitHub blocks
  self-approval when using a PAT).
- One place to audit who/what changed your branch protection state.

### One-time setup

**1. Create the App.**

Go to: `https://github.com/organizations/utmstack/settings/apps/new`

- **GitHub App name**: e.g. `UTMStack Approver`.
- **Homepage URL**: any (the UTMStack repo URL is fine).
- **Webhook**: untick **Active** — no callbacks needed.
- **Repository permissions:**
  - `Contents`: **Read-only**
  - `Pull requests`: **Read and write**
  - `Metadata`: Read-only (default, can't be removed).
- **Organization permissions:**
  - `Members`: **Read-only** — needed for the team-membership check.
- **Where can this GitHub App be installed?** Only on this account.

Click **Create GitHub App**.

**2. Get the App ID and a private key.**

On the App's settings page you'll see the **App ID** (numeric). Save it.

Scroll to **Private keys** → **Generate a private key**. A `.pem` file
downloads. Save the **full contents** (BEGIN/END lines included).

**3. Install the App on the UTMStack repo.**

On the App page → **Install App** → pick the `utmstack` org → choose
**Only select repositories** → select `UTMStack` → Install.

**4. Add the secrets to the repo.**

Settings → Secrets and variables → Actions → New repository secret.

- `APPROVER_APP_ID` = the numeric App ID.
- `APPROVER_PRIVATE_KEY` = the full PEM contents of the `.pem` file,
  including the `-----BEGIN/END PRIVATE KEY-----` lines. Paste as-is —
  GitHub preserves multi-line values.

**5. Optional: drop `API_SECRET`.**

If the App has `Members: Read` at org level, you can stop maintaining a
separate `API_SECRET` PAT for the permission check. The approver
workflow falls back to the App token when `API_SECRET` is not set
(`API_SECRET: ${{ secrets.API_SECRET || steps.app-token.outputs.token }}`
in `_pr-reusable-approver.yml`).

`API_SECRET` is still used by the deployment workflows
(`v10-deployment-pipeline.yml`, `v11-deployment-pipeline.yml`) for things
like fetching private Go modules during installer builds — don't delete
it from the repo until you confirm those workflows no longer need it.

### How it gets minted at runtime

In `_pr-reusable-approver.yml`:

```yaml
- name: Generate approver token from GitHub App
  id: app-token
  if: ${{ env.APP_ID != '' }}
  env:
    APP_ID: ${{ secrets.APPROVER_APP_ID }}
  uses: actions/create-github-app-token@v1
  with:
    app-id: ${{ secrets.APPROVER_APP_ID }}
    private-key: ${{ secrets.APPROVER_PRIVATE_KEY }}
```

If the secrets aren't configured, the step is skipped, the approver
runs in comments-only mode, and everything else still works (deps
comment, AI review comment, status check) — just no formal review and
no auto-merge.

### Verifying it works

1. Open a small, low-risk PR against `release/v11.x.x` (or push the
   workflow to a sandbox branch).
2. After the approver job runs, check the PR page:
   - The sticky `<!-- approver:ai -->` comment is signed by your bot
     account (e.g. `utmstack-approver[bot]`).
   - The "Files changed" tab shows a review by the same bot, marked
     `Approved` (Tier 1 + deps OK + authorized) or `Changes requested`.
3. If the target is `release/**` and Tier 1 → auto-merge is queued in
   the PR header ("Auto-merge enabled by … via GitHub Actions").

---

## Reusable workflows

**PR checks:**

- `_pr-reusable-go-deps.yml` — runs `go-deps.sh --check --discover` at
  repo level and uploads `go-deps-result` as an artifact.
- `_pr-reusable-ai-review.yml` — fan-out per prompt; each job uploads
  `ai-review-<name>` as an artifact.
- `_pr-reusable-approver.yml` — downloads artifacts, decides verdict,
  posts sticky comments, optionally leaves a formal PR review.

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

**Permission denied:**
- Verify membership in `integration-developers` or `core-developers`.

**`ai_review` artifact with tier 2 fallback "Manual review recommended":**
- The model didn't return valid JSON or returned an invalid tier. The
  approver treats it as Tier 2 (changes requested) fail-safe. Refine the
  prompt `.md` or re-run the workflow if it was transient.

**`go_deps` fails with "Could not inspect ... run 'go mod tidy' there":**
- `go.sum` is out of sync, typically due to local `replace` directives in
  `packages/`. Run `go mod tidy` in the affected module and commit.

**The approver posts two separate comments (deps + AI):**
- That's the expected behaviour when both dimensions fail. Each comment
  is independent and gets updated in place on subsequent runs.

**The approver doesn't leave a formal review (only comments):**
- The approver GitHub App is not configured. Add both `APPROVER_APP_ID`
  and `APPROVER_PRIVATE_KEY` secrets — see
  [Approver GitHub App setup](#approver-github-app-setup).

**Want a senior engineer @mentioned on Tier 3:**
- Edit `pr-checks.yml`, in the `approver` job set the `tier3_reviewers`
  input with comma-separated handles:
  ```yaml
  with:
    tier3_reviewers: 'Kbayero,osmontero'
  ```

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
