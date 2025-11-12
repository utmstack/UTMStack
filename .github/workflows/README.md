# 🛠️ GitHub Actions Workflows – UTMStack

> This repository uses streamlined CI/CD workflows for building and deploying UTMStack v10 and v11 across different environments.

## 📋 Table of Contents

- [Workflows Overview](#workflows-overview)
- [V10 Deployment Pipeline](#v10-deployment-pipeline)
- [V11 Deployment Pipeline](#v11-deployment-pipeline)
- [Installer Release](#installer-release)
- [Required Secrets and Variables](#required-secrets-and-variables)

---

## 🔄 Workflows Overview

### 1. **installer-release.yml**
Automatically builds and publishes installers when a GitHub release is created.

**Trigger:** Release created (types: `released`)

**Behavior:**
- Detects version (v10 or v11) from release tag
- Builds installer for the detected version
- Uploads installer binary to the GitHub release

### 2. **v10-deployment-pipeline.yml**
Automated CI/CD pipeline for v10 builds and deployments.

**Triggers:**
- Push to `v10` branch → Deploys to **v10-rc**
- Push to `release/v10**` branches → Deploys to **v10-dev**
- Tags `v10.*` → Production build

**Environments:**
- `v10-dev` - Development environment (from release branches)
- `v10-rc` - Release candidate environment (from v10 branch)
- Production (from tags)

### 3. **v11-deployment-pipeline.yml**
Manual deployment pipeline for v11 with version control.

**Trigger:** Manual (`workflow_dispatch`)

**Required Inputs:**
- `version_tag`: Version to deploy (e.g., `v11.0.0-dev.1` or `v11.1.0`)
- `event_processor_tag`: Event processor version (e.g., `1.0.0-beta`)

**Version Formats:**
- **Dev:** `v11.x.x-dev.N` (e.g., `v11.0.0-dev.1`)
- **Production:** `v11.x.x` (e.g., `v11.1.0`)

---

## 🚀 V10 Deployment Pipeline

### Flow

```
┌─────────────────────┐
│  Push to Branch     │
└──────────┬──────────┘
           │
           ├─── release/v10** ──→ Build & Deploy to v10-dev
           ├─── v10 ──────────→ Build & Deploy to v10-rc
           └─── tag v10.* ────→ Build for Production
```

### Jobs

1. **setup_deployment** - Determines environment based on trigger
2. **validations** - Validates user permissions
3. **build_agent** - Builds and signs Windows/Linux agents
4. **build_agent_manager** - Builds agent-manager Docker image
5. **build_*** - Builds all microservices (aws, backend, correlation, frontend, etc.)
6. **all_builds_complete** - Checkpoint for all builds
7. **deploy_dev / deploy_rc** - Deploys to respective environments

### Permissions

- **Dev deployments**: `integration-developers` or `core-developers` teams
- **RC/Prod deployments**: Same as dev

---

## 🎯 V11 Deployment Pipeline

### Flow

```
┌─────────────────────────────┐
│  Manual Workflow Dispatch   │
│  with version_tag input     │
└──────────────┬──────────────┘
               │
               ├─── v11.x.x-dev.N ──→ DEV Environment
               └─── v11.x.x ────────→ PROD Environment
```

### Jobs

1. **validations** - Validates user permissions and version format
2. **build_agent** - Builds and signs Windows/Linux agents
3. **build_utmstack_collector** - Builds UTMStack Collector
4. **build_agent_manager** - Builds agent-manager Docker image
5. **build_event_processor** - Builds event processor with plugins
6. **build_backend** - Builds backend microservice (Java 17)
7. **build_frontend** - Builds frontend microservice
8. **build_user_auditor** - Builds user-auditor microservice
9. **build_web_pdf** - Builds web-pdf microservice
10. **all_builds_complete** - Checkpoint for all builds
11. **publish_new_version** - Publishes version to Customer Manager
12. **schedule** - Schedules release to configured instances

### Permissions

- **Dev versions** (`v11.x.x-dev.N`):
  - Must run from `release/` or `feature/` branches
  - Requires: `administrators`, `integration-developers`, or `core-developers` team membership

- **Production versions** (`v11.x.x`):
  - Requires: `administrators` team membership only

### Environment Detection

The pipeline automatically detects the environment based on version format:

| Version Format | Environment | CM Auth Secret | CM URL | Schedule Instances Var | Schedule Token Secret |
|----------------|-------------|----------------|--------|------------------------|----------------------|
| `v11.x.x-dev.N` | dev | `CM_AUTH_DEV` | `https://cm.dev.utmstack.com` | `SCHEDULE_INSTANCES_DEV` | `CM_SCHEDULE_TOKEN_DEV` |
| `v11.x.x` | prod | `CM_AUTH` | `https://cm.utmstack.com` | `SCHEDULE_INSTANCES_PROD` | `CM_SCHEDULE_TOKEN_PROD` |

---

## 📦 Installer Release

### Flow

```
┌─────────────────────┐
│  GitHub Release     │
│  Created & Published│
└──────────┬──────────┘
           │
           ├─── Tag v10.x.x ──→ Build v10 Installer
           └─── Tag v11.x.x ──→ Build v11 Installer
```

### Behavior

- Validates release tag format
- Builds installer with correct configuration:
  - **V10:** Basic build
  - **V11:** Build with ldflags (version, branch, encryption keys)
- Uploads installer to GitHub release assets

---

## 🔐 Required Secrets and Variables

### Secrets

| Secret Name | Used In | Description |
|-------------|---------|-------------|
| `API_SECRET` | All | GitHub API token for team membership validation |
| `AGENT_SECRET_PREFIX` | v10, v11 | Agent encryption key |
| `SIGN_CERT` | v10, v11 | Code signing certificate path (var) |
| `SIGN_KEY` | v10, v11 | Code signing key |
| `SIGN_CONTAINER` | v10, v11 | Code signing container name |
| `CM_AUTH` | v11 | Customer Manager auth credentials (prod) |
| `CM_AUTH_DEV` | v11 | Customer Manager auth credentials (dev) |
| `CM_ENCRYPT_SALT` | installer | Encryption salt for installer |
| `CM_SIGN_PUBLIC_KEY` | installer | Public key for installer verification |
| `CM_SCHEDULE_TOKEN_PROD` | v11 | Auth token for cm-version-publisher (prod) |
| `CM_SCHEDULE_TOKEN_DEV` | v11 | Auth token for cm-version-publisher (dev) |
| `GITHUB_TOKEN` | All | Auto-provided by GitHub Actions |

### Variables

| Variable Name | Used In | Description | Format |
|---------------|---------|-------------|--------|
| `SCHEDULE_INSTANCES_PROD` | v11 | Instance IDs for prod scheduling | Comma-separated UUIDs |
| `SCHEDULE_INSTANCES_DEV` | v11 | Instance IDs for dev scheduling | Comma-separated UUIDs |

**Example Variable Values:**
```
SCHEDULE_INSTANCES_PROD=uuid1,uuid2,uuid3
SCHEDULE_INSTANCES_DEV=uuid-dev1
```

---

## 🎮 How to Deploy

### V10 Deployment

**Dev Environment:**
```bash
git checkout release/v10.x.x
git push origin release/v10.x.x
# Automatically deploys to v10-dev
```

**RC Environment:**
```bash
git checkout v10
git merge release/v10.x.x
git push origin v10
# Automatically deploys to v10-rc
```

**Production Release:**
```bash
git tag v10.5.0
git push origin v10.5.0
# Builds production artifacts
```

### V11 Deployment

**Dev Environment:**
1. Navigate to Actions tab
2. Select "v11 - Build & Deploy Pipeline"
3. Click "Run workflow"
4. Fill in:
   - **version_tag:** `v11.0.0-dev.1`
   - **event_processor_tag:** `1.0.0-beta`
5. Click "Run workflow"

**Production Release:**
1. Navigate to Actions tab
2. Select "v11 - Build & Deploy Pipeline"
3. Click "Run workflow"
4. Fill in:
   - **version_tag:** `v11.1.0`
   - **event_processor_tag:** `1.0.0`
5. Click "Run workflow"

---

## 🏗️ Reusable Workflows

The following reusable workflows are called by the main pipelines:

- `reusable-basic.yml` - Basic Docker builds
- `reusable-golang.yml` - Golang microservice builds
- `reusable-java.yml` - Java microservice builds
- `reusable-node.yml` - Node.js/Frontend builds

---

## 📝 Notes

- All Docker images are pushed to `ghcr.io/utmstack/utmstack/*`
- V11 uses `-community` suffix for all image tags
- Agent signing requires `utmstack-signer` runner
- Artifacts (agents, collector) have 1-day retention
- Failed deployments will stop the pipeline and report errors

---

## 🆘 Troubleshooting

**Permission Denied:**
- Verify you're a member of the required team
- For v11 prod: Must be in `administrators` team
- For v11 dev: Can be in `administrators`, `integration-developers`, or `core-developers`

**Build Failures:**
- Check that all required secrets are configured
- Verify runner availability (especially `utmstack-signer` for agent builds)
- Review build logs for specific errors

**Version Format Errors:**
- Dev: Must match `v11.x.x-dev.N` (e.g., `v11.0.0-dev.1`)
- Prod: Must match `v11.x.x` (e.g., `v11.1.0`)

---

**For questions or issues, please contact the DevOps team.**
