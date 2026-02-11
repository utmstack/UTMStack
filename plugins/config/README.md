# UTMStack Config Plugin

A lightweight service that keeps UTMStack detection content (rules, filters, patterns, and tenants) in sync from PostgreSQL to the local plugin work directory used by the ThreatWinds SDK.

It periodically checks the database for updates and, when changes are detected, rewrites the YAML files that power the correlation and pipeline layers.

## What it does
- Connects to PostgreSQL using settings under the `com.utmstack` configuration group.
- Detects changes in these tables (by their last-update timestamps):
  - `utm_tenant_config` (assets/tenants)
  - `utm_correlation_rules` (rules)
  - `utm_logstash_filter` (pipeline filters)
  - `utm_regex_pattern` (Grok/regex patterns)
- Writes files into the plugin work directory (`plugins.WorkDir`):
  - `pipeline/tenants.yaml` — tenant and assets configuration
  - `pipeline/patterns.yaml` — regex/grok patterns
  - `pipeline/filters/<id>.yaml` — Logstash/Log pipeline snippets, one file per active filter
  - `rules/utmstack/<id>.yaml` — correlation rules, one file per active rule
- Removes files that no longer exist in the database (safe cleanup per folder).
- Uses an inter-process lock so only one writer runs at a time.

The loop runs every 30 seconds. If the environment mode is set to `playground`, the program exits immediately (disabled mode).

## Configuration
This plugin relies on ThreatWinds Go SDK configuration. Two config namespaces are used:

1) Database credentials: `com.utmstack`
```
com:
  utmstack:
    postgresql:
      server: 127.0.0.1
      port: 5432
      database: utmstack
      user: utmstack
      password: change_me
```

2) Plugin runtime options: `plugin_com.utmstack.config`
```
plugin_com:
  utmstack:
    config:
      env:
        mode: prod  # set to "playground" to disable the loop
```

Notes:
- The exact loading mechanism (file path, env) is provided by the ThreatWinds SDK; ensure the process can read these settings.
- `plugins.WorkDir` is also provided by the SDK. By default it points to the plugin’s writable work directory.

## Database expectations
- Filters are read when `utm_logstash_filter.is_active = true` and written as `<id>.yaml`.
- Rules are read when `utm_correlation_rules.rule_active = true` and written as `<id>.yaml`.
- Rule data types are resolved from `utm_group_rules_data_type` and `utm_data_types`.
- Assets come from `utm_tenant_config` and are exported into `pipeline/tenants.yaml` as a single default tenant.
- Patterns come from `utm_regex_pattern` and are exported into `pipeline/patterns.yaml`.

## Build
Requires Go (version as specified in `go.mod`).
```
go build -o utmstack-config-plugin .
```

## Run
Ensure configuration is available to the process (see Configuration section), then run:
```
./utmstack-config-plugin
```
Behavior:
- On startup, the service connects to PostgreSQL and checks for changes.
- On change, it regenerates the output files under `plugins.WorkDir`.
- It sleeps for ~30s between checks.

To temporarily disable processing (e.g., during local development), set the plugin env mode to `playground`.

## Output layout (relative to plugins.WorkDir)
```
./pipeline/
  tenants.yaml
  patterns.yaml
  /filters/
    <filter-id>.yaml
./rules/
  /utmstack/
    <rule-id>.yaml
```

## Error handling & logging
- Database, I/O, and serialization errors are wrapped with context via the `catcher` package.
- Locking is handled by the SDK (`plugins.AcquireLock`/`plugins.ReleaseLock`).

## Development notes
Key functions are implemented in `main.go`:
- Database access: `connect`, `getFilters`, `getRules`, `getPatterns`, `getAssets`, `getRuleDataTypes`.
- Change detection: `hasChanges` (uses `MAX(updated_at)`/`MAX(last_update)` timestamps).
- Writers: `writeFilters`, `writeRules`, `writePatterns`, `writeTenant`.
- Cleanup: `cleanUpFilters`, `cleanUpRules`.

## Troubleshooting
- Verify that the `com.utmstack` database credentials are correct and reachable.
- Ensure the process has write permissions to the SDK work directory (`plugins.WorkDir`).
- Confirm the plugin mode is not set to `playground`.
- Check logs for messages emitted by the `catcher` error wrapper.
