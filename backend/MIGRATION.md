# v12 backend — where every module stands

Five things are being done to this backend at once. This is the state of each
module against each of them, and what is left.

The dimensions:

| | What "done" means |
|---|---|
| **Multitenancy** | Every read and write is scoped to a tenant, and it is not possible to forget. Batches that span tenants say so explicitly. |
| **ClickHouse** | Nothing reaches OpenSearch. Queries go through the store driver, which takes a scope. |
| **Scalability** | Correct with N replicas: no in-process state that matters, periodic work claimed, writes off the request path. |
| **Schema** | Table and column names are ours, not inherited from the Java ORM. No dead columns. |
| **Frontend** | The UI talks to the shape the backend now has. |

Legend: ✅ done · **—** not started · **n/a** does not apply · **required** the UI must follow a breaking change

---

## Status

| Module | Multitenancy | ClickHouse | Scalability | Schema | Frontend |
|---|---|---|---|---|---|
| iam | ✅ | n/a | ✅ | ✅ | ✅ |
| tenant | ✅ | n/a | ✅ | ✅ | ✅ |
| audit | ✅ | n/a | ✅ | ✅ | ✅ |
| billing | ✅ | n/a | ✅ | ✅ | ✅ |
| notifications | ✅ | n/a | ✅ | ✅ | ✅ |
| appconfig | ✅ | n/a | ✅ | ✅ | ✅ |
| socai | ✅ | n/a | ✅ | ✅ | ✅ |
| threatintel | ✅ | n/a | ✅ | n/a | ✅ |
| soar | — | — | — | — | — |
| adaudit | — | — | — | — | — |
| datasources | — | — | — | — | — |
| loganalyzer | — | — | — | — | — |
| dashboards | — | — | — | — | — |
| alerts | — | — | — | — | — |
| alertscoring | — | — | — | — | — |
| incidents | — | — | — | — | — |
| compliance | — | — | — | — | — |
| eventprocessing | — | — | — | — | — |
| integrations | — | — | — | — | — |
| mcp | — | — | — | — | — |
| opensearch | — | **retire** | — | — | — |

---

## What is left, in the order it has to happen

**1. alerts** — the root. alertscoring, incidents, compliance and the MCP tools
all read alerts through it, so nothing below it can be finished first.

Its port (`AlertRepository`) is clean, so a ClickHouse implementation slots in.
The hard part is not the reads: it is `UpdateStatus`, `UpdateNotes`,
`UpdateTags`, `ConvertToIncident` — seven mutation methods. The alerts table is
a plain `MergeTree`, which does not update cheaply. That decision has to be made
before any code is written:

- `ReplacingMergeTree` keyed on the alert id, and every change writes a new
  version. Reads need `FINAL` or a de-duplicating view.
- Mutations (`ALTER TABLE … UPDATE`), which are asynchronous and heavy.
- Keep the mutable part in Postgres — status, notes, tags, assignee — and join.
  The alert body stays immutable in ClickHouse.

The third is the least ClickHouse-shaped and the most likely to be right: what
changes about an alert is small, bounded, and low-volume compared to the alert
itself.

**2. alertscoring, incidents, compliance** — all downstream of alerts. Nothing
in them needs its own decision; they follow whatever alerts does.

**3. eventprocessing, integrations** — the last two with their own OpenSearch
reads (`ingestion_stats_os.go`; the index-pattern registry).

**4. mcp** — 59 references. It exposes OpenSearch's shape to the model: the
filter DSL is documented as "a single OpenSearch bool query", `opensearch.search`
is a tool, and the audit event types are `OPENSEARCH_INDEX_*`. Renaming is not
free — those names are the contract the model was prompted against.

**5. opensearch** — retired last, with `utm_index_pattern`, its seed, the model
in AutoMigrate, and the `/opensearch/index*` routes.

**6. frontend** — one pass, at the end, against a backend that has stopped
moving. Two things are already broken and waiting: dashboards no longer accepts
SQL (visualizations need to send a spec), and the log explorer must ask
`/log-analyzer/datasets` instead of the OpenSearch endpoints.

---

## Schemas

Renaming was blocked by the in-place upgrade path. That path was deleted, so
every install is now clean and a rename is a tag in a model — no data to
preserve, no backfill.

What still carries the Java ORM:

| Table | Problem |
|---|---|
| ~~`jhi_user`, `jhi_authority`, `jhi_user_authority`~~ | Gone: RBAC rebuilt as `user`, `role`, `user_role`, `role_permission` |
| `utm_*` (17 tables) | Prefix that means nothing here |
| `app_config` (renamed) | Still a key-value with `conf_param_short/large/value/datatype`, holding SMTP, date format, language *and* branding-as-JSON. The name is clean; the shape is not. |
| `utm_log_analyzer_query` | `la_name`, `la_filters`, `la_data_origin` |
| `utm_incident_actions` + `utm_incident_action_command` | Two tables for one concept |
| `utm_alert_response_rule`, `utm_alert_response_rule_template` | Dead: SOAR rules are YAML now |
| `utm_data_types`, `utm_correlation_rules`, `utm_logstash_filter`, `utm_regex_pattern`, `utm_tenant_config` | Dead: all moved to YAML |

The dead ones cost nothing to drop. The renames are mechanical. The two that are
actual redesign are the config table and the incident-actions pair.

---

## What only appeared when it was run

Twenty modules were changed before anything was executed against a real
database. The first run found three faults that no test could see, and all three
would have stopped an install:

1. **The migration did not apply.** The `utm_alert_tag` seed used
   `ON CONFLICT (tag_name)` while the unique index is `(tenant_id, tag_name)` —
   Postgres refuses with 42P10. Same fault as `datasources(source_ref)` and the
   `ad_user` indexes, found the same way. Three times in one day.

2. **The SOAR bootstrap queried a table called `tenants`.** It is `tenant`. The
   error was swallowed and it fell back to the default tenant, so on a managed
   install no customer would have received the flows the product ships.

3. **A clean install cannot be logged into.** `APP_TFA_ENABLED` defaults to
   true, login mails a code, a fresh install has no SMTP, and login returns 500.
   The administrator cannot log in, so cannot configure SMTP. This matters more
   now that every install is clean.

The lesson is the method, not the code: **a module is not finished until it has
run against a real database.**

---

## Still unverified

Postgres and ClickHouse are up locally and the migration applies. Verified
against them: every unique index exists including the partial ones on `ad_user`;
the upserts for `ad_user`, `socai_ai_usage` and `job_leases` all work; a live
lease is not stolen; appconfig's inheritance returns the tenant's own value and
falls back to the default's.

Not yet verified:

- The whole ClickHouse query path. It has never returned a row. Data is seeded
  for two tenants and the check was interrupted at login.
- Anything in the modules not yet migrated.

---

## Carried debt

- **`backend/go.mod` has a `replace` pointing at a local checkout of the SDK.**
  The eight filter operators it needs are not published. This blocks merging.
- The generic list repository counts with `SELECT count(*)` and pages with
  `OFFSET` on every request. Shared by every module.
- SOAR's flow bootstrap and the audit purge still run in every replica.
  `pkg/joblease` exists and is used by the datasources reconciler; those two are
  the remaining candidates.
- A new tenant does not receive SOAR's seeded response actions, which are seeded
  for the default tenant only.
