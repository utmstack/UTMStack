# Deploying & making filters durable

## The three layers — update ALL THREE or a restart reverts you

```
repo file  →  Postgres (LIVE source of truth)  →  engine /workdir  ←  config plugin polls Postgres every ~30s
   ↑                ↑                                                ↑
   └── reseed dir (backend container /utmstack/filters/...) is what a backend restart re-seeds Postgres FROM
```

1. **Postgres** `utm_logstash_filter` row — what the running engine loads. UPDATE it for a live change.
2. **Backend reseed dir** — `/utmstack/filters/office365/o365.yml` inside the **current** backend container. On any backend (re)start, `DefinitionSyncService` (a CommandLineRunner) copies this baked dir over the DB. If you only update Postgres, the next restart reverts you.
3. **Engine** `/workdir/pipeline/filters/<filterId>.yaml` inside the event-processor-**worker** — verify THIS to confirm the config plugin actually reloaded.

A "deploy" = write the YAML to all three and verify all three agree.

## Container names change after restarts — re-lookup every time

`docker ps --format "{{.Names}} | {{.Status}}" | grep <service>`. The current one is the row with `Up <recent>`; older ones are `Exited`. Never hardcode a container id across sessions. Known services (swarm tasks, `<svc>.1.<suffix>`):
- `utmstack_backend.1.*` — REST API + reseed source
- `utmstack_postgres.1.*` — Postgres (DB `utmstack`, user `postgres`)
- `utmstack_event-processor-manager.1.*` — runs the O365 collector (pulled from Graph; swarm `replicas 1/1`)
- `utmstack_event-processor-worker.u0g8inlzv4*.*` — runs the pipeline engine (`/workdir/`)

## The restart-revert trap (bit us in the O365 campaign)

The backend crash-looped once (OpenSearch `Ping ... Read timed out`, exit 1); swarm spun a **fresh** task whose image has **pristine v11.2.13** reseed files. On start it re-seeded Postgres, silently reverting ~60 rule fixes + the filter + **any `docker cp`'d files** (those live in the old container's writable layer, which is gone). Symptom: filter version back to 1.0.4, max rule id back to 1461, old container `Exited (1)`.
- **Fix:** after any backend restart, re-check the DB and re-deploy. To make work survive a restart, it must be in the NEW container's reseed dir — `docker cp` it in **and** it will persist only until the container is replaced again.
- **Also note:** the reseed re-assigns filter ids. The O365 filter was id **1590** in one backend and **1591** in the next. Re-query by name (`filter_name ilike '%o365%'`), not id.

## Deploy snippet (Postgres update, via stdin to avoid shell quoting)

```python
import subprocess
def sh_exec(c, sh, data=None):
    p=subprocess.run(['ssh','-o','BatchMode=yes','osmany@192.168.99.75','sudo docker exec -i '+c+' '+sh],
                     input=(data.encode() if data else None),capture_output=True,timeout=120)
    return p.returncode, p.stdout.decode('utf-8','replace'), p.stderr.decode('utf-8','replace')
PG='<current utmstack_postgres name>'
q=lambda s: "'"+s.replace("'","''")+"'"
yaml_text=open('filters/office365/o365.yml',encoding='utf-8').read()
# UPDATE by NAME-matched id:
fid = sh_exec(PG,'psql -U postgres -d utmstack -t -A -c "select id from utm_logstash_filter where filter_name ilike \x27%o365%\x27;"')[1].strip()
sh_exec(PG,'psql -U postgres -d utmstack -v ON_ERROR_STOP=1',
        data="UPDATE utm_logstash_filter SET logstash_filter=%s, filter_version='1.2.3', updated_at=NOW() WHERE id=%s;"%(q(yaml_text),fid))
```

Then `docker cp` the file into the backend reseed dir, then confirm the engine file header shows the new version within ~30–45s.

## Verifying the engine reloaded
```
ssh: sudo docker exec <worker> sh -c "head -1 /workdir/pipeline/filters/<fid>.yaml"
ssh: sudo docker exec <worker> sh -c "grep -c log.Members /workdir/pipeline/filters/<fid>.yaml"
```

## No-JWT query paths (read the index without a login token)
`Utm-Internal-Key: tJ2yzvMshr98qmXvYtHLuHNOqzNmIMJz` header is allowlisted **only** on `/api/elasticsearch/search`.
```python
import json, ssl, urllib.request
ctx=ssl.create_default_context(); ctx.check_hostname=False; ctx.verify_mode=ssl.CERT_NONE
K='tJ2yzvMshr98qmXvYtHLuHNOqzNmIMJz'
def es(idx, filt, sz=50):
    r=urllib.request.Request('https://192.168.99.75/api/elasticsearch/search?top=%d&indexPattern=%s&page=0&size=%d'
        %(sz,idx,sz), data=json.dumps(filt).encode(),
        headers={'Utm-Internal-Key':K,'Content-Type':'application/json'}, method='POST')
    b=json.loads(urllib.request.urlopen(r,context=ctx,timeout=45).read()); return b if isinstance(b,list) else []
docs = es('v11-log-o365-*',[{'field':'action.keyword','operator':'IS','value':'MailboxLogin'},
                            {'field':'@timestamp','operator':'IS_BETWEEN','value':['now-7d','now']}])
```
Filters use `{field, operator, value}` (operators: IS, IS_BETWEEN, IN, etc). Use `*.keyword` for exact term matches on text fields.

## Ingestion lag (don't mistake for a filter bug)
- O365 collector (in manager) pulls 5 content types in **5-min rolling windows**, ~5–6 min end-to-end. A fresh event may not be indexed for up to ~6 min.
- **Teams `TeamDeleted` audit events lag 30–60+ min** past the actual delete.
- Collector idling (no O365 queue writes for 10 min) is normal when there's no new data — check `docker logs` manager for recent `v11-log-o365` writes and absence of `status=503` before assuming it stalled.

## Windows shell gotchas (this box)
- `bash` unavailable natively (WSL removed); run bash in ephemeral Docker.
- Inline `python -c` with `<`, `|`, `$`, quotes breaks in PowerShell — **write a `.py` file and run it**.
- Set `$env:PYTHONIOENCODING='utf-8'` before running python that prints O365/emoji or `→` (U+2192) — cp1252 otherwise crashes on decode AND encode.
- Use the venv python: `C:\Users\osmon\utmstack-creds\venv-utmctl\Scripts\python.exe` (has pyyaml, msal, etc).
