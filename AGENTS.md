# AGENTS.md — UTMStack Repository Guide

See [org-level AGENTS.md](../AGENTS.md) for branch strategy and CI/CD overview.

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
| `user-auditor/` | Java 11 (Spring Boot 2.7) | Maven |
| `web-pdf/` | Java 11 (Spring Boot 2.7) | Maven |

## Build Commands

### Backend (Java 17)
```bash
cd backend
mvn -s settings.xml -B                                # Spring Boot dev server (port 8080)
mvn -B -Pprod clean package -s settings.xml           # Production WAR → target/utmstack.war
```
- `settings.xml` authenticates to GitHub Packages via `MAVEN_TK` env var (GitHub PAT with `read:packages`).
- Spring profiles: `dev` (default), `prod`, `tls`. Config in `backend/src/main/resources/config/`.
- Docker dev database: `docker-compose -f backend/src/main/docker/mysql.yml up -d`
- **No `src/test/` directory** — tests are embedded in `src/main/java/`.
- **`jib-maven-plugin` image mismatch** — `pom.xml` references `eclipse-temurin:11-jre-focal` but the backend requires Java 17. The `Dockerfile` correctly uses `eclipse-temurin:17`. Jib builds will fail or produce broken images.
- **`proto-command.txt`** contains the protoc command for gRPC code generation.

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
- Frontend Dockerfile serves via nginx (see `frontend/nginx/`).

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

**16 plugins** are in `event_processor.Dockerfile`. The `plugins/` directory has 17 modules — `compliance-orchestrator` is in the directory but not yet in the Dockerfile.

Plugin list in Dockerfile: `alerts`, `aws`, `azure`, `bitdefender`, `config`, `events`, `gcp`, `geolocation`, `inputs`, `o365`, `sophos`, `stats`, `soc-ai`, `modules-config`, `crowdstrike`, `feeds`.

### user-auditor and web-pdf
Both are small Spring Boot 2.7.14 microservices (Java 11). Each has its own `pom.xml`, `Dockerfile`, and `compose.yml`. Built independently of the main backend.

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

## Architecture Overview

```
endpoint agent ──gRPC──▶ agent-manager ──▶ backend (Java REST API) ↔ Angular frontend
                                      │
                               event processor (Go, plugin-loaded)
                               ├── plugins (17 Go modules)
                               └── filters/ + rules/ (YAML)
                                      │
                               utmstack-collector / as400 (log ingestion)
```

- **Backend** serves the REST API. WAR packaging. `filters/` and `rules/` are YAML files copied into the container at `/utmstack/filters` and `/utmstack/rules`.
- **Event processor** is the core Go-based log correlation engine. Loads compiled plugin binaries at runtime. `event_processor.Dockerfile` expects all plugins pre-built alongside it.
- **Frontend** is a standalone Angular app served by nginx in a separate container.
- **Agent** runs on endpoints (Windows/Linux/macOS). Communicates with `agent-manager` via gRPC. Has Windows resource files (`rsrc_windows_*.syso`) for embedded icons.
- **Collector** (`utmstack-collector/`) and **AS400** (`as400/`) are separate log collection services. `as400/` is a near-identical copy of `utmstack-collector/` (same structure, different parser).
- **etc/** contains ISO build configs and OpenSearch configuration.

## Testing

- **Go**: `go test -v ./...` in each module directory. CI runs this via `reusable-golang.yml` before build.
- **Frontend**: `npm test` (Karma + Jasmine). Note: `e2e` script exists but uses Protractor (deprecated).
- **Backend**: No separate `src/test/` tree. Tests live in `src/main/java/` alongside production code.
- **Go deps check**: `bash .github/scripts/go-deps.sh --check --discover` (discovers all `go.mod` files, checks for outdated direct deps and out-of-sync `go.sum`).

## CI / CD

### PR Checks
`.github/workflows/pr-checks.yml` — runs on PRs to `release/**`, `v10`, `v11`. Three jobs:
1. `go_deps` — dependency freshness check across all Go modules
2. `ai_review` — matrix of ThreatWinds AI prompts (`.github/ai-prompts/`)
3. `approver` — consolidates results, sticky comments, optional formal review + auto-merge

**AI prompts:** `security.md`, `bugs.md`, `architecture.md`. Default model: `gemini-3-flash-lite`. Add new prompts by dropping `.md` into `.github/ai-prompts/`.

**Tiered approval:**
- Tier 1: AI approves, auto-merge on `release/**`
- Tier 2: AI flags issues, `REQUEST_CHANGES`
- Tier 3: Critical path (crypto, auth, migrations, gRPC, CI/CD) → human review

See `.github/workflows/README.md` for full CI/CD documentation (secrets, approver setup, deployment flows, hotfix procedure).

### Deployment Pipelines
- **v11** (`.github/workflows/v11-deployment-pipeline.yml`) — active line
- **v10** (`.github/workflows/v10-deployment-pipeline.yml`) — legacy (EOL Dec 2026)

Images published to `ghcr.io/utmstack/utmstack/<image>:<tag>`.

### Reusable Workflows
- `reusable-java.yml` — Maven build + Docker push
- `reusable-golang.yml` — `go test ./...` + `go build` + Docker push
- `reusable-node.yml` — Node 14.16.1, `npm install && npm run-script build`, Docker push
- `reusable-sign-agent.yml` — Windows (jsign + GCP KMS) and macOS (codesign + notarytool) signing
- `reusable-basic.yml` — Docker build only

## Filters, Rules, and Plugins

- **filters/** and **rules/** — YAML files, organized by vendor/domain (25+ categories each: antivirus, aws, azure, cisco, crowdstrike, fortinet, generic, windows, etc.)
- Plugin/filerule/rule documentation: [UTMStack Wiki](https://github.com/utmstack/UTMStack/wiki)
- Custom plugin development guide: `UTMStack.wiki/Custom-Plugin-Development.md`

## Gotchas

- **ldflags are mandatory** for `agent`, `utmstack-collector`, and `as400` — `REPLACE_KEY` is injected at build time. Without it, services cannot authenticate.
- **Backend uses GitHub Packages** — `settings.xml` references `maven.pkg.github.com/utmstack/**`. Requires GitHub PAT in `$MAVEN_TK`.
- **Frontend is Angular 7** — no standalone components, signals, or modern APIs. CLI v7.3.6.
- **Node 14.16.1 required for frontend** — `npm install` on Node 16+ fails on `node-sass@4`.
- **Frontend build needs 8 GB heap** — set `NODE_OPTIONS=--max_old_space_size=8192`.
- **Installer build requires ldflags** — see `installer/build.sh` for `DEFAULT_BRANCH`, `INSTALLER_VERSION`, `REPLACE`, `PUBLIC_KEY`.
- **Geolocation data must be downloaded** — event processor Docker build fails without `./geolocation/` CSV files from GCS.
- **`.plugin` binaries are gitignored** — build artifacts, not committed.
- **`jib-maven-plugin` image mismatch** — `pom.xml` references `eclipse-temurin:11-jre-focal` but the backend requires Java 17.
- **`as400/` is a copy of `utmstack-collector/`** — nearly identical structure (agent, collector, config, conn, database, logservice, models, serv, utils). Changes to one usually need mirroring in the other.
- **`CONTRIBUTING.md` is stale** — mentions PEP 8 (Python) but there is no Python code in the repo.
- **`shared/` is a library, not buildable standalone** — consumed via `replace` directives. Don't try to `go build` it directly.
