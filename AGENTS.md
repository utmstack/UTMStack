# AGENTS.md — UTMStack Repository Guide

## Overview

SIEM/XDR platform. Multi-language monorepo:

| Directory | Language | Build |
|---|---|---|
| `backend/` | Java 17 (Spring Boot 3.1, JHipster 7.3) | Maven (`mvn`) |
| `frontend/` | Angular 7 (TypeScript 3.2, Node 14) | Angular CLI + npm |
| `agent/` | Go 1.25.5 | `go build` (+ ldflags) |
| `agent-manager/` | Go 1.25.5 | `go build` |
| `utmstack-collector/` | Go 1.25.5 | `go build` (+ ldflags) |
| `as400/` | Go 1.25.5 | `go build` (+ ldflags) |
| `plugins/*/` | Go 1.25.5 | `go build` (17 modules) |
| `shared/` | Go 1.25.1 | — (shared library) |
| `installer/` | Go 1.25.1 | `go build` (+ ldflags) |
| `user-auditor/` | Java 11 | Maven |
| `web-pdf/` | Java 11 | Maven |

## Build Commands

### Backend (Java)
```bash
cd backend
mvn -s settings.xml -B                        # Run Spring Boot dev server (port 8080)
mvn -B -Pprod clean package -s settings.xml   # Production WAR → target/utmstack.war
```
- Maven settings: `backend/settings.xml` authenticates to GitHub Packages via `MAVEN_TK` env var (GitHub PAT with `read:packages`).
- Spring profiles: `dev` (default), `prod`, `tls`. Config in `backend/src/main/resources/config/`.
- Docker dev database: `docker-compose -f backend/src/main/docker/mysql.yml up -d`
- **No `src/test/` directory** — tests are embedded in `src/main/java/`.

### Frontend (Angular 7)
```bash
cd frontend
npm install
npm start                           # ng serve --host 0.0.0.0
NODE_OPTIONS=--max_old_space_size=8192 npm run build   # ng build --prod
npm test                            # ng test (Karma + Jasmine)
npm run lint                        # ng lint (TSLint, NOT ESLint)
```
- **Node 14.16.1 required**. Newer Node breaks `node-sass@4`.
- Linter is **TSLint** (`tslint.json`), not ESLint. Uses `codelyzer` rules.
- Output: `dist/utm-stack`. Styles use SCSS (`angular.json`).
- **Builds need 8 GB heap**: `NODE_OPTIONS=--max_old_space_size=8192`.

### Go Components
Each module is independent. Build from its directory:
```bash
cd agent
go build -o utmstack_agent_service .
```

**`shared/` replace directives:** `agent/go.mod` and `agent/updater/go.mod` have `replace github.com/utmstack/UTMStack/shared => ../shared` (or `../../shared`). These two modules cannot be built outside the repo.

**ldflags required for agent, collector, and as400:**
```bash
# Agent
go build -ldflags "-X 'github.com/utmstack/UTMStack/agent/config.REPLACE_KEY=<secret>'" .

# UTMStack Collector
go build -ldflags "-X 'github.com/utmstack/UTMStack/utmstack-collector/config.REPLACE_KEY=<secret>'" .

# AS400 Collector
go build -ldflags "-X 'github.com/utmstack/UTMStack/as400/config.REPLACE_KEY=<secret>'" .
```
CI injects `$AGENT_SECRET_PREFIX` for all three. Without it, these services cannot authenticate.

**Cross-compilation:** Set `GOOS`/`GOARCH`/`CGO_ENABLED=0` before `go build`. CI builds Linux (amd64/arm64), Windows (amd64/arm64), macOS (arm64).

### Plugins
Each plugin under `plugins/*/` is a standalone Go module. Build binary named `com.utmstack.<name>.plugin`.

**16 plugins** are copied into `event_processor.Dockerfile`. The `plugins/` directory currently has 17 modules — `compliance-orchestrator` exists but is not yet in the Dockerfile.

### Installer
```bash
cd installer
bash build.sh                       # Uses ldflags for config injection
```
`build.sh` injects `DEFAULT_BRANCH`, `INSTALLER_VERSION`, `REPLACE` (encryption salt), and `PUBLIC_KEY` via `-ldflags`.

### Geolocation Data
Event processor needs CSV files downloaded at build time from:
```
https://storage.googleapis.com/utmstack-updates/dependencies/geolocation/
```
`geolocation/` is gitignored — must be populated from GCS before Docker build.

## CI / CD

### PR Checks
`.github/workflows/pr-checks.yml` — runs on PRs to `release/**`, `v10`, `v11`. Triggers Go dependency checks, AI review, and approver job.

### Deployment Pipelines
- **v11** (`.github/workflows/v11-deployment-pipeline.yml`) — active line. Triggers: push to `release/v11**` (dev), push to `v11` (rc), `release.released` (prod).
- **v10** (`.github/workflows/v10-deployment-pipeline.yml`) — legacy. Triggers: push to `release/v10**` (dev), push to `v10` (rc), `release.released` with `v10.*` tag (prod).

### Reusable Workflows
- `reusable-java.yml` — Maven build + Docker push to `ghcr.io/utmstack/utmstack/<image>:<tag>`
- `reusable-golang.yml` — `go test ./...`, `go build`, Docker push
- `reusable-node.yml` — Node 14.16.1, `npm install && npm run-script build`, Docker push
- `reusable-sign-agent.yml` — Windows (jsign + GCP KMS) and macOS (codesign + notarytool) signing
- `reusable-basic.yml` — Docker build-only

### Agent Signing Workflow
`installer-release.yml` builds the installer binary with ldflags using `CM_ENCRYPT_SALT` and `CM_SIGN_PUBLIC_KEY` secrets.

## Key Architecture Notes

- **Event processor** is the core Go-based log correlation engine. Loads compiled plugin binaries at runtime. `event_processor.Dockerfile` expects all plugins pre-built alongside it.
- **Backend** serves the REST API. WAR packaging. `filters/` and `rules/` are YAML files copied into the container at `/utmstack/filters` and `/utmstack/rules`.
- **Frontend** is a standalone Angular app in a separate container.
- **Agent** runs on endpoints (Windows/Linux/macOS). Communicates with `agent-manager` via gRPC.
- **Collector** (`utmstack-collector/`) and **AS400** (`as400/`) are separate log collection services.

## Gotchas

- **ldflags are mandatory** for `agent`, `utmstack-collector`, and `as400` — `REPLACE_KEY` is injected at build time. Without it, services cannot authenticate.
- **Backend uses GitHub Packages** — `settings.xml` references `maven.pkg.github.com/utmstack/**`. Requires GitHub PAT in `$MAVEN_TK`.
- **Frontend is Angular 7** — no standalone components, signals, or modern APIs. CLI v7.3.6.
- **Node 14.16.1 required for frontend** — `npm install` on Node 16+ fails on `node-sass@4`.
- **Frontend build needs 8 GB heap** — set `NODE_OPTIONS=--max_old_space_size=8192`.
- **Installer build requires ldflags** — see `installer/build.sh` for `DEFAULT_BRANCH`, `INSTALLER_VERSION`, `REPLACE`, `PUBLIC_KEY`.
- **Geolocation data must be downloaded** — event processor Docker build fails without `./geolocation/` CSV files from GCS.
- **`.plugin` binaries are gitignored** — build artifacts, not committed.
- **`jib-maven-plugin` image mismatch** — pom.xml references `eclipse-temurin:11-jre-focal` but the backend requires Java 17.
