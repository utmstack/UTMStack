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
| `integrations/*/` | Go 1.25.5 | `go build` | 8 pullers (aws, azure, bitdefender, crowdstrike, gcp, o365, o365_graph, sophos), one container each; `integrations/shared` holds their coordination, crypto and identity code |
| `soc-ai/` | Python 3.12 (LangGraph) | `pip install .` / `pytest` | The SOC agent as its own service (supervisor + specialist agents over the backend MCP tools); `plugins/soc-ai` is the old Go plugin, kept only as a reference |
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
Each integration under `integrations/*/` is a standalone Go module (the pullers; see the next section) built into its own image: the binary is compiled in the `build_integrations` matrix of the v12 pipeline (through `reusable-golang.yml` with `working_directory`) and the `Dockerfile` in each folder only copies it, like the backend's. The image is `ghcr.io/utmstack/utmstack/<name>:<tag>`. They publish logs straight to NATS (`LOGS_TRANSPORT=nats`, `NATS_URL`) and read their module configuration from the NATS key-value bucket `integration-config` (`integrations/shared/modulecfg`), decrypted with `ENCRYPTION_KEY`; nothing is mounted from the host except, for `bitdefender`, the certificate in `/cert` (`CERTS_FOLDER`), which it serves HTTPS with on 8000. The event processor image carries no plugin binaries: `event_processor.Dockerfile` only adds the geolocation data. `soc-ai/` is its own service, built by the `build_soc_ai` job (tests, then its image) and run as the `soc-ai` container: the backend queues each new alert for it in NATS (`socai.alerts`, with the alert's dossier already worked out), publishes the per-tenant model configuration in the bucket `soc-ai-config` (encrypted with `ENCRYPTION_KEY`), and shows it a short list of its MCP tools (`backend/modules/mcp/agent_profile.go`); `plugins/soc-ai` is no longer built or run.

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

**Built into the event processor (no binary here):** the parsing steps (including `com.utmstack.geolocation`), the rule engine `cel`, the event writer `com.utmstack.events`, the statistics collector `com.utmstack.stats` the alert correlator `com.utmstack.alerts` and the automation trigger `com.utmstack.soar` the user account inventory `com.utmstack.ad-audit` (`pkg/analysis/adaudit`) and the login anomaly detector `com.utmstack.user-behavioral` (`pkg/analysis/userbehavior`, worker mode only) run inside the event processor image (its repository, `pkg/analysis`, `pkg/notifications`, `pkg/parsing/builtin`). Their names stay in the `plugins.<category>.order` lists and in the filters (`dynamic: plugin: com.utmstack.geolocation`); the names resolve to the built-in code. The playground API (`:8091`, filters and rules testing) is also part of the event processor now (`pkg/playground`). There is no `plugins/events`, `plugins/stats`, `plugins/geolocation`, `plugins/playground`, `plugins/alerts`, `plugins/soar`, `plugins/user-behavioral`, `plugins/ad-audit` or `plugins/feeds` module in this repository, and the deployment pipeline does not build them. The rule flood guard (it disables a correlation rule that floods a tenant's alerts) is a periodic task of the backend alerts module (`backend/modules/alerts/usecase/flood_guard.go`), configured by `system_plugins_rule-flood-guard.yaml` in the pipeline folder. The ThreatWinds contribution (what used to be `plugins/feeds`: entities from incidents sent through the Customers Manager proxy) is a periodic task of the backend threatintel module (`backend/modules/threatintel/contribution`), single runner by the `threatintel-contribution` lease, with the already-sent entities remembered in Valkey. This needs an event processor base image that includes them (`TW_EVENT_PROCESSOR_VERSION_V12_*`).

Do not "enable" an input plugin by editing an order list, and do not debug a dead analysis plugin by looking for a config file it never had.

**Vendor integration config** (the pullers in `integrations/`, which run in their own containers and no longer as event processor plugins) — the backend keeps each integration's module groups in the NATS key-value bucket `integration-config`, one key per integration (`aws`, `azure`, `bitdefender`, `crowdstrike`, `gcp`, `o365`, `o365_graph`, `sophos`; see `backend/modules/integrations/repository/nats_config_store.go`). The value is the whole YAML document (`plugins: <vendor>: tenants: [{id, groups: [{name, config: {...}}]}]`) **encrypted** with `ENCRYPTION_KEY`, and the secrets inside it are encrypted on top of that. Each container watches its key (`modulecfg.Watch`), so it gets the current value when it starts and every change after, with no polling. Gotchas:

- **The backend is the only writer.** On start it moves any `system_plugins_<vendor>.yaml` an earlier version left in the pipeline folder into NATS (never overwriting a value NATS already has) and renames the file to `.migrated`; a file it cannot read or verify stays where it was. Without `NATS_URL` the backend keeps using the files (development).
- **No key in NATS = the integration idles** (`ModuleActive=false`), and the container reports nothing: check the bucket (`nats kv get integration-config <vendor>`), not a file.
- **`ENCRYPTION_KEY` must be the same in the backend and in every integration container** (the installer sets both to the internal key). With another key the container cannot read its configuration and logs an error; the old value stays in force.
- **A configuration is lost if the NATS data volume is.** It is now the only copy; back up `NATSData` with the rest.
- **`com.utmstack` base config is not read by the pullers any more**: what they needed comes from environment variables (`ENCRYPTION_KEY`, `NATS_URL`, `CERTS_FOLDER`).

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
