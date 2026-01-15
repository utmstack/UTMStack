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
Automated CI/CD pipeline for v11 builds and deployments.

**Triggers:**
- Push to `release/v11**` branches → Deploys to **dev** environment
- Prerelease created → Deploys to **rc** environment

**Version Formats:**
- **Dev:** `v11.x.x-dev.N` (e.g., `v11.2.1-dev.1`) - Auto-incremented
- **RC:** `v11.x.x` (e.g., `v11.2.1`) - From prerelease tag

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
│  Push to release/v11.x.x    │
│  branch                     │
└──────────────┬──────────────┘
               │
               ▼
       Auto-increment version
       (v11.x.x-dev.N)
               │
               ▼
       Build & Deploy to DEV
               │
               ▼
       Publish to CM Dev
               │
               ▼
       Schedule to Dev Instances


┌─────────────────────────────┐
│  Create Prerelease          │
│  (tag: v11.x.x)             │
└──────────────┬──────────────┘
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
       Publish to CM Prod
               │
               ▼
       Schedule to Prod Instances
```

### Jobs

1. **setup_deployment** - Determines environment and version based on trigger
2. **validations** - Validates user permissions (team membership)
3. **build_agent** - Builds and signs Windows/Linux agents
4. **build_utmstack_collector** - Builds UTMStack Collector
5. **build_agent_manager** - Builds agent-manager Docker image
6. **build_event_processor** - Builds event processor with plugins
7. **build_backend** - Builds backend microservice (Java 17)
8. **build_frontend** - Builds frontend microservice
9. **build_user_auditor** - Builds user-auditor microservice
10. **build_web_pdf** - Builds web-pdf microservice
11. **all_builds_complete** - Checkpoint for all builds
12. **generate_changelog** - Generates AI-powered changelog (RC only)
13. **build_installer_rc** - Builds and uploads installer (RC only)
14. **deploy_installer_dev** - Deploys installer (Dev only)
15. **publish_new_version** - Publishes version to Customer Manager
16. **schedule** - Schedules release to configured instances

### Permissions

- Requires: `integration-developers` or `core-developers` team membership

### Environment Detection

The pipeline automatically detects the environment based on trigger:

| Trigger | Environment | CM URL | Service Account | Schedule Instances Var |
|---------|-------------|--------|-----------------|------------------------|
| Push to `release/v11**` | dev | `https://cm.dev.utmstack.com` | `CM_SERVICE_ACCOUNT_DEV` | `SCHEDULE_INSTANCES_DEV` |
| Prerelease created | rc | `https://cm.utmstack.com` | `CM_SERVICE_ACCOUNT_PROD` | `SCHEDULE_INSTANCES_PROD` |

### Version Auto-Increment (Dev)

For dev deployments, the version is automatically calculated:
1. Extracts base version from branch name (e.g., `release/v11.2.1` → `v11.2.1`)
2. Queries CM for latest version
3. If base versions match, increments dev number (e.g., `v11.2.1-dev.9` → `v11.2.1-dev.10`)
4. If base versions differ, starts fresh (e.g., `v11.2.1-dev.1`)

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
| `CM_SERVICE_ACCOUNT_PROD` | v11 | Customer Manager service account credentials (prod/rc) - JSON format `{"id": "...", "key": "..."}` |
| `CM_SERVICE_ACCOUNT_DEV` | v11 | Customer Manager service account credentials (dev) - JSON format `{"id": "...", "key": "..."}` |
| `CM_ENCRYPT_SALT` | installer | Encryption salt for installer |
| `CM_SIGN_PUBLIC_KEY` | installer | Public key for installer verification |
| `OPENAI_API_KEY` | v11 | OpenAI API key for changelog generation |
| `GITHUB_TOKEN` | All | Auto-provided by GitHub Actions |

### Variables

| Variable Name | Used In | Description | Format |
|---------------|---------|-------------|--------|
| `SCHEDULE_INSTANCES_PROD` | v11 | Instance IDs for prod/rc scheduling | Comma-separated UUIDs |
| `SCHEDULE_INSTANCES_DEV` | v11 | Instance IDs for dev scheduling | Comma-separated UUIDs |
| `TW_EVENT_PROCESSOR_VERSION_PROD` | v11 | ThreatWinds Event Processor version (prod/rc) | Semver (e.g., `1.0.0`) |
| `TW_EVENT_PROCESSOR_VERSION_DEV` | v11 | ThreatWinds Event Processor version (dev) | Semver (e.g., `1.0.0-beta`) |

**Example Variable Values:**
```
SCHEDULE_INSTANCES_PROD=uuid1,uuid2,uuid3
SCHEDULE_INSTANCES_DEV=uuid-dev1
TW_EVENT_PROCESSOR_VERSION_PROD=1.0.0
TW_EVENT_PROCESSOR_VERSION_DEV=1.0.0-beta
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
```bash
git checkout release/v11.2.1
# Make your changes
git add .
git commit -m "Your changes"
git push origin release/v11.2.1
# Automatically builds and deploys to dev
# Version is auto-incremented (e.g., v11.2.1-dev.1, v11.2.1-dev.2, ...)
```

**RC Release:**
1. Navigate to GitHub Releases
2. Click "Draft a new release"
3. Create a new tag (e.g., `v11.2.1`)
4. Select "Set as a pre-release"
5. Click "Publish release"
6. Pipeline automatically:
   - Builds all microservices
   - Generates AI-powered changelog
   - Builds and uploads installer
   - Publishes version to CM
   - Schedules updates to RC instances

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
- Agent signing requires `utmstack-signer` runner
- Artifacts (agents, collector) have 1-day retention
- Failed deployments will stop the pipeline and report errors
- Dev versions follow the format `v11.x.x-dev.N` (auto-incremented)
- RC versions use the prerelease tag directly (e.g., `v11.2.1`)

---

## 🆘 Troubleshooting

**Permission Denied:**
- Verify you're a member of the required team
- For v11: Must be in `integration-developers` or `core-developers` team

**Build Failures:**
- Check that all required secrets are configured
- Verify runner availability (especially `utmstack-signer` for agent builds)
- Review build logs for specific errors

**Version Not Incrementing:**
- Check that the CM API is accessible
- Verify `CM_SERVICE_ACCOUNT_DEV` or `CM_SERVICE_ACCOUNT_PROD` secrets are correctly configured
- Ensure the branch name follows the format `release/v11.x.x`

**Changelog Not Generated:**
- Verify `OPENAI_API_KEY` secret is configured
- Only applies to RC releases (prereleases)

---

**For questions or issues, please contact the DevOps team.**
