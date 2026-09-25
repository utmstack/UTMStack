# Deploying & verifying rules (end-to-end)

A rule is only "done" when it has **fired** on a real event with correct attribution. The loop: deploy → engine confirms loaded → generate the event → confirm the alert.

## 1) Deploy to Postgres (live) + reseed dir (durable)

Postgres is the live source of truth; the backend reseed dir is what a restart re-seeds from. Update **both**, or a restart reverts you (see trap below).

Container names change on restart — re-lookup each time:
```
docker ps --format "{{.Names}} | {{.Status}}" | grep -E "backend|postgres"
```

Insert (new rule) — pipe SQL via stdin to dodge shell quoting. `data_type_id=4` is o365 (re-query per tenant). JSON text cols = the YAML lists serialized:
```python
import subprocess
def sh_exec(c, sh, data=None):
    p=subprocess.run(['ssh','-o','BatchMode=yes','osmany@192.168.99.75','sudo docker exec -i '+c+' '+sh],
                     input=(data.encode() if data else None),capture_output=True,timeout=120)
    return p.returncode,p.stdout.decode('utf-8','replace'),p.stderr.decode('utf-8','replace')
PG='<postgres name>'
q=lambda s:"'"+s.replace("'","''")+"'"
import yaml,json
d=yaml.safe_load(open('rules/office365/foo.yml',encoding='utf-8'))
where=d['where'].strip()
ae=json.dumps(d.get('afterEvents')) if d.get('afterEvents') else None
gb=json.dumps(d.get('groupBy')) if d.get('groupBy') else None
refs=json.dumps(d.get('references',[]))
sql=("INSERT INTO utm_correlation_rules (rule_name,rule_confidentiality,rule_integrity,rule_availability,"
 "rule_category,rule_technique,rule_description,rule_references_def,rule_definition_def,rule_last_update,"
 "rule_active,system_owner,rule_adversary,rule_after_events_def,rule_group_by_def) "
 "VALUES ({},{},{},{},{},{},{},{},{},NOW(),true,true,{},{},{});"
 .format(q(d['name']), d['impact']['confidentiality'], d['impact']['integrity'], d['impact']['availability'],
         q(d.get('category','Collection')), q(d.get('technique','')), q(d.get('description','')),
         q(refs), q(where), q(d.get('adversary','origin')), q(ae) if ae else 'NULL', q(gb) if gb else 'NULL'))
sql+="\nINSERT INTO utm_group_rules_data_type (rule_id,data_type_id,last_update) VALUES (currval((select setval(pg_get_serial_sequence('utm_correlation_rules','id'))),4,NOW());\n"
rc,o,e=sh_exec(PG,'psql -U postgres -d utmstack -v ON_ERROR_STOP=1',data=sql)
```
For an **update** (existing id), `UPDATE utm_correlation_rules SET rule_definition_def=..., rule_after_events_def=..., rule_group_by_def=..., rule_last_update=NOW() WHERE id=<id>;`

Reseed: `docker exec <backend> sh -c "cat > /utmstack/rules/office365/<file>.yml"` (pipe the YAML via stdin).

## 2) Confirm the engine loaded it (~30s config-plugin poll)
The worker holds `/workdir/rules/<vendor>/<id>.yaml`. (Rule files are named by numeric id, filter by `<vendor>` subdir.)
```
sudo docker exec <worker> sh -c "cat /workdir/rules/office365/<id>.yaml"
```
If it's absent, the config plugin hasn't synced yet (wait) or the DB write didn't land (re-check step 1).

## 3) Generate the real event
You must actually produce the O365 action so it flows through the collector → filter → rule. O365 collector pulls 5 content types in 5-min windows, ~5–6 min end-to-end. Examples used in the campaign:
- **Service Principal** → register an app in Entra (fires `Add service principal.`); delete it after.
- **Double-ext / reserved filename** → upload a file to OneDrive (fires `FileAccessed`/`FilePreviewed` on open, `FileDownloaded` on download). Windows reserved names (`NUL`,`CON`…) can't be created on a Windows client — platform block.
- **Teams mass-delete** → create N Teams then delete (fires `TeamDeleted`, but that audit event lags 30–60+ min).
- **Teams external phishing (846)** → message an `#EXT#` guest (fires `MessageSent`/`ChatCreated` with `log.Members[].UPN` containing `#EXT#`; needs the filter `cast log.Members to string`).
- **MailboxLogin / PowerShell** → `Connect-ExchangeOnline` (device code) then a mailbox query. NOTE: modern EXO uses Graph auth; classic `MailboxLogin` audit op is often absent — verify the op exists in the index before relying on it.

## 4) Confirm the alert fired + attributed correctly
No-JWT read (allowlisted only on `/api/elasticsearch/search`):
```python
import json,ssl,urllib.request
ctx=ssl.create_default_context(); ctx.check_hostname=False; ctx.verify_mode=ssl.CERT_NONE
K='tJ2yzvMshr98qmXvYtHLuHNOqzNmIMJz'   # Utm-Internal-Key (this instance)
def es(idx,filt,sz=100):
    r=urllib.request.Request('https://192.168.99.75/api/elasticsearch/search?top=%d&indexPattern=%s&page=0&size=%d'%(sz,idx,sz),
        data=json.dumps(filt).encode(),headers={'Utm-Internal-Key':K,'Content-Type':'application/json'},method='POST')
    b=json.loads(urllib.request.urlopen(r,context=ctx,timeout=45).read()); return b if isinstance(b,list) else []
# event present?
ev=es('v11-log-o365-*',[{'field':'action.keyword','operator':'IS','value':'Add service principal.'},
                        {'field':'@timestamp','operator':'IS_BETWEEN','value':['now-1h','now']}])
# alert fired? (match on rule name)
al=es('v11-alert-*',[{'field':'dataType.keyword','operator':'IS','value':'o365'},
                     {'field':'@timestamp','operator':'IS_BETWEEN','value':['now-1h','now']}])
hit=[a for a in al if a.get('name')==d['name']]
for a in hit: print(a['@timestamp'], a.get('severity'), a.get('adversary'))
```
Correct = alert exists AND `adversary.user/.ip` is the expected actor AND severity matches impact. If the event is in the index but no alert → the rule's `where`/`afterEvents` doesn't match the real event shape; re-read the event's fields.

## The restart-revert trap (re-verify after ANY backend restart)
A backend (re)start runs `DefinitionSyncService`, re-seeding `utm_correlation_rules`/`utm_logstash_filter` from the **baked image** reseed dir (NOT a volume). Any rule/filter you only `INSERT`ed into Postgres — and any you `docker cp`'d into the old container's writable layer — is **reverted** to the image default. The O365 filter id even renumbered 1590→1591 on re-seed.
- Detect: after a restart, `select max(id) from utm_correlation_rules` and compare to what you deployed; check filter `filter_version`.
- Prevent: always also write to the **current** backend's `/utmstack/rules/...` + `/utmstack/filters/...` (so the next re-seed carries your changes), AND keep the repo as source of truth. Note the reseed dir itself is baked at image build, so a *new* image without your changes still reverts — for durable fixes, ship the repo change.
- The backend crash-looped once in this campaign (OpenSearch `Ping ... Read timed out`, exit 1) which is what triggered the re-seed.

## Windows/SSH gotchas
- `bash` unavailable natively (WSL removed); use MSYS2 bash or ephemeral Docker for bash scripts.
- Inline `python -c` with quotes/parens breaks in PowerShell — write a `.py` file and run it.
- `PYTHONIOENCODING=utf-8` before printing O365/emoji (cp1252 crashes).
- Use venv python `C:\Users\osmon\utmstack-creds\venv-utmctl\Scripts\python.exe` (has pyyaml).
