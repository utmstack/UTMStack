# AGENTS.md — UTMStack Repository Guide

## Overview

UTMStack is a SIEM/XDR platform. Multi-language monorepo with three technology stacks:

| Directory | Language | Build | Notes |
|---|---|---|---|
| `backend/` | Java 17 (Spring Boot 3.1, JHipster 7.3) | Maven (`./mvnw`) | WAR packaging, Liquibase migrations |
| `frontend/` | Angular 7 (TypeScript 3.2, Node 14) | Angular CLI + npm | Very old Angular — no modern APIs |
| `agent/` | Go 1.25.5 | `go build` | Cross-compiled for Linux/Windows/macOS |
| `agent-manager/` | Go 1.25.5 | `go build` | Distributes agent binaries |
| `utmstack-collector/` | Go 1.25.5 | `go build` | Log collector service |
| `as400/` | Go 1.25.5 | `go build` | AS400 log collector |
| `plugins/*/` | Go 1.25.5 | `go build` | 16 independent plugin modules |
| `shared/` | Go 1.25.1 | — | Shared library (consumed via `replace`) |
| `installer/` | Go 1.25.1 | `go build` | Self-updating installer binary |
| `user-auditor/` | Java 11 | Maven | Standalone microservice |
| `web-pdf/` | Java 11 | Maven | Standalone microservice |

## Build Commands

### Backend (Java)
```bash
cd backend
./mvnw                              # Run Spring Boot dev server (port 8080)
./mvnw -Pprod clean package         # Production WAR build → target/utmstack.war
./mvnw clean test                   # Run tests
```
- Maven settings: `backend/settings.xml` authenticates to GitHub Packages via `$MAVEN_TK` env var (GitHub PAT).
- Spring profiles: `dev` (default), `prod`, `tls`. Config in `backend/src/main/resources/config/`.
- Docker dev database: `docker-compose -f backend/src/main/docker/mysql.yml up -d`
- **No `src/test/` directory exists** — tests are embedded in `src/main/java/`.

### Frontend (Angular 7)
```bash
cd frontend
npm install
npm start                           # ng serve --host 0.0.0.0
npm run build                       # ng build --prod --base-href /
npm test                            # ng test (Karma + Jasmine)
npm run lint                        # ng lint (TSLint, NOT ESLint)
```
- **Node 14.16.1 required** (pinned in CI). Newer Node will break `node-sass@4`.
- Linter is **TSLint** (`tslint.json`), not ESLint. Uses `codelyzer` rules.
- Output: `dist/utm-stack`
- Styles use SCSS (see `angular.json` schematic config).
- **CI sets `NODE_OPTIONS=--max_old_space_size=8192`** for builds — you may need this locally too.

### Go Components
Each Go module is independent. Build from its directory:
```bash
cd agent
go build -o utmstack_agent_service .
```

**`shared/` replace directives:** Only `agent/go.mod` and `agent/updater/go.mod` have `replace github.com/utmstack/UTMStack/shared => ../shared`. You cannot build those two modules outside the repo structure. All other Go modules (plugins, agent-manager, installer, etc.) are fully independent.

**Agent ldflags:** Agent binaries require `-ldflags "-X 'github.com/utmstack/UTMStack/agent/config.REPLACE_KEY=<secret>'"` to inject an auth prefix. Without it, agents won't authenticate. CI uses `$AGENT_SECRET_PREFIX` secret.

**Cross-compilation:** Set `GOOS`/`GOARCH`/`CGO_ENABLED=0` before `go build`. CI builds Linux (amd64/arm64), Windows (amd64/arm64), and macOS (arm64).

**Geolocation data:** The event processor needs external CSV files downloaded from `storage.googleapis.com/utmstack-updates/dependencies/geolocation/` at build time.

### Plugins
Each plugin under `plugins/*/` is a standalone Go module. Build output is a single binary named `com.utmstack.<name>.plugin`. The `event_processor.Dockerfile` copies all 16 plugin binaries into the event processor container.

### Installer
```bash
cd installer
bash build.sh                       # Uses ldflags for config injection
```
`build.sh` sets `GOPRIVATE=github.com/utmstack` and injects `DEFAULT_BRANCH`, `INSTALLER_VERSION`, `REPLACE` (encryption salt), and `PUBLIC_KEY` via `-ldflags`.

## CI / CD

Two active deployment pipelines:

- **v11** (`.github/workflows/v11-deployment-pipeline.yml`) — current active line. Triggers on push to `release/v11*` branches (dev) or prerelease events (rc).
- **v10** (`.github/workflows/v10-deployment-pipeline.yml`) — legacy. Triggers on push to `release/v10*` and prerelease/release events.

Reusable workflow building blocks:
- `reusable-java.yml` — builds Maven projects, pushes Docker image to `ghcr.io/utmstack/utmstack/<image>:<tag>`
- `reusable-golang.yml` — runs `go test ./...`, builds binary, pushes Docker image
- `reusable-node.yml` — Node 14.16.1, `npm install && npm run-script build`, pushes Docker image
- `reusable-sign-agent.yml` — signs Windows binaries (jsign + GCP KMS) and macOS binaries (codesign + notarytool)
- `reusable-basic.yml` — Docker build-only, no compile step

All Docker images push to **ghcr.io/utmstack/utmstack/**. Agent binary signing requires GCP KMS credentials (Windows) or Apple Developer certificates (macOS).

## Key Architecture Notes

- **Event processor** is the core log correlation engine. It's a standalone Go-based service that loads compiled plugin binaries at runtime. The `event_processor.Dockerfile` expects all plugins pre-built alongside it.
- **Backend** serves the REST API and the Angular frontend. The frontend is a separate directory but gets built and deployed as its own container.
- **Agent** runs on endpoints (Windows/Linux/macOS). Communicates with `agent-manager` via gRPC. The agent-manager container bundles all agent binaries for distribution.
- **Filters** (`filters/`) and **rules** (`rules/`) are YAML files defining log parsing filters and detection rules. The backend Dockerfile copies these into the container at `/utmstack/filters` and `/utmstack/rules`.
- **UTMStack Collector** (`utmstack-collector/`) and **AS400** (`as400/`) are log collectors — separate from the endpoint agent.

## Event Processor Plugin Categories & Config (things that bite often)

The event processor loads plugin binaries from `/workdir/plugins/utmstack/`. There are two registration regimes — mixing them up is a classic trap:

| Category | Examples | How it runs |
|---|---|---|
| **Input plugins** (vendor collectors) | `com.utmstack.aws`, `com.utmstack.azure`, `com.utmstack.o365`, `com.utmstack.crowdstrike`, `com.utmstack.gcp`, `com.utmstack.sophos`, `com.utmstack.bitdefender` | **Execute because they exist.** No execution order, no registration anywhere. Adding/removing them from any `order` list does nothing. |
| **Analysis / correlation / notification plugins** | `com.utmstack.events`, `cel`, `com.utmstack.feeds`, `com.utmstack.alerts`, … | Must be listed in `plugins.<category>.order` in `/workdir/pipeline/system_plugins_analysis.yaml`, `system_plugins_correlation.yaml`, `system_plugins_notification.yaml` respectively. **Missing from the list = never runs, silently.** |

Do not "enable" an input plugin by editing an order list, and do not debug a dead analysis plugin by looking for a config file it never had.

**Vendor input plugin config** — each vendor plugin reads its own file `/workdir/pipeline/system_plugins_<vendor>.yaml` (e.g. `com.utmstack.aws` reads `system_plugins_aws.yaml`; structure `plugins: <vendor>: tenants: [{id, groups: [{name, config: {...}}]}]` — see each plugin's `config.go` `pluginFile` const). The file is generated from Postgres `utm_module` / `utm_module_group` / `utm_module_group_configuration` tables (modules-config path; `utmstack_plugins.yaml` points at the service via `modulesConfig`). Gotchas:

- **`com.utmstack.aws` is a UTMStack vendor plugin, NOT a system plugin.** `utmstack_plugins.yaml` only holds the shared `com.utmstack` base config (opensearch, postgresql, internalKey, certsFolder, …) — vendor collectors are not in it.
- **Missing `system_plugins_<vendor>.yaml` = plugin idles** (`ModuleActive=false`) even when the module group is configured and active in the UI. Check the file on the worker container before debugging credentials.
- **Module-group secrets are encrypted at rest.** Values in a plugin's `sensitiveKeys` (e.g. `aws_secret_access_key`) are encrypted in Postgres and decrypted with the `com.utmstack` base config `encryptionKey`. Without that key the plugin fails to decrypt and cannot authenticate. Symptom: the raw stored value tested outside the plugin gives `SignatureDoesNotMatch` — it's a ciphertext, not the real secret.
- **`plugins/modules-config/` in this repo contains only the built `.plugin` binary** — there is no Go source to read; the generator logic only exists in the binary.

## Gotchas

- **No `src/test/` in backend** — Maven test phase runs but the directory doesn't exist yet. Tests are embedded in `src/main/java/`.
- **Agent ldflags required** — `config.REPLACE_KEY` is injected at build time. Without it, agents won't authenticate.
- **Backend uses GitHub Packages** — `settings.xml` references `maven.pkg.github.com/utmstack/**`. You need a GitHub PAT with `read:packages` scope in `$MAVEN_TK`.
- **Frontend is Angular 7** — many modern Angular patterns (standalone components, signals, etc.) do not exist. CLI is v7.3.6.
- **Node 14 required for frontend** — CI pins 14.16.1. `npm install` on Node 16+ will fail on `node-sass@4`.
- **`npm run build` needs 8 GB heap** — CI sets `NODE_OPTIONS=--max_old_space_size=8192`.
- **Installer build requires ldflags** — `installer/build.sh` shows the required `-ldflags` for `DEFAULT_BRANCH`, `INSTALLER_VERSION`, `REPLACE`, and `PUBLIC_KEY`.
- **Geolocation data must be downloaded** — event processor needs CSV files from GCS at build time.
- **`geolocation/` directory is gitignored** — must be populated at build time from GCS.
- **`.plugin` files are gitignored** — plugin binaries are build artifacts, not committed.
