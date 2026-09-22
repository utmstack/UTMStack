# SOCAI pointing at prod CM despite dev URL in instance-config.yml

## Root cause: SOCAI URL is baked in at config-write time, not resolved per request

Key insight from `backend/modules/socai/usecase/config.go:106` and `:250`:

```go
URL: strings.TrimRight(inst.Server, "/") + threatWindsAIPath
```

That URL string is written **into the stored SOCAI config** (either by `EnsureDefault` on first boot or by `Update` when the tenant saves). After that, every AI request reads the URL from the stored config — it does **not** re-check `instance-config.yml`. So the CM URL the assistant hits is a **snapshot from whenever the SOCAI config was last written**, and it drifts from `instance-config.yml` easily.

## Concrete causes for prod URL leaking through

Ranked by how often each actually bites:

### 1. SOCAI config was written before instance-config.yml was flipped to dev
Most common. Timeline:
```
install with branch=prod  → instance-config.yml server = cm.utmstack.com
backend boots             → EnsureDefault runs → SOCAI URL saved with prod
someone edits instance-config.yml to cm.dev.utmstack.com
                          → SOCAI config still holds the old prod URL
```
The 30s `StartEnsureDefaultLoop` (line 37) only writes if **no config exists yet** (`if existing != nil { return true }` at line 83), so it never rewrites.

**Check**:
```bash
docker exec utmstack-backend cat /workdir/pipeline/socai/config.yml 2>/dev/null
# or via API:
curl -H "Authorization: Bearer $TOKEN" https://<instance>/api/v1/soc-ai/config | jq .url
```

**Fix**: reset the config, let `EnsureDefault` rebuild it from current `instance-config.yml`:
```bash
curl -X POST -H "Authorization: Bearer $TOKEN" https://<instance>/api/v1/soc-ai/config/reset
# or delete the stored config file directly then restart backend
```

### 2. A tenant saved a custom SOCAI config that overrides the default
`Update()` (line 206) writes a **per-tenant** config. `Get()` (line 153) returns the tenant's own config if present, and only falls back to the instance default otherwise. If someone hit "save" in the AI settings UI while on a tenant, that tenant now has its own URL baked in.

**Check**: same `/soc-ai/config` endpoint — the response has `"inherited": false` when it's a tenant-own config, `true` when it's the instance default. If `inherited=false`, that tenant is on a snapshot.

**Fix**: `ResetToDefault` (line 193) — the "reset to default" button in the UI, or:
```bash
curl -X DELETE -H "Authorization: Bearer $TOKEN" https://<instance>/api/v1/soc-ai/config
```

### 3. `instance-config.yml` says dev but was originally prod
If the file was manually edited but the backend still has the old value cached in `instanceconfig.Get()`, or if the config was re-registered under prod later. Verify what the backend actually sees:

```bash
docker exec utmstack-backend cat /updates/instance-config.yml
# server field — must match dev
docker restart utmstack-backend   # force re-read of instance-config.yml
```

Then reset SOCAI config so `EnsureDefault` rebuilds it from the fresh read.

### 4. Multiple replicas with different config-store views
`ensureDefaultIfMine` uses a distributed lease (line 66). If replicas have inconsistent Redis/lease state, one may have written a config with a stale value that others now inherit. Rare unless you're actually running HA.

### 5. Installer `Branch` config says prod
`GetCMServer()` (installer/config/const.go:47) returns prod unless `Branch == "dev"`. If `/root/utmstack.yml`'s `branch` field is empty/prod but you expected dev, every re-registration writes prod back into `instance-config.yml`:

```bash
grep -i branch /root/utmstack.yml
```

### 6. There's no auto-migration when you flip environments
This is the design gap, not really a bug: switching an instance from prod→dev CM requires manually:
1. Fix `/root/utmstack.yml` (`branch: dev`)
2. Delete `/updates/instance-config.yml` (so the installer re-registers against dev CM)
3. Restart `UTMStackComponentsUpdater` → gets new `instance_id`/`instance_key` from dev CM
4. Reset SOCAI config → `EnsureDefault` rebuilds URL from the new `instance-config.yml`

Skip any step and something in the chain will still point at prod.

---

## Fastest triage

```bash
# 1. What does SOCAI actually think its URL is?
curl -sH "Authorization: Bearer $TOKEN" https://<instance>/api/v1/soc-ai/config | jq '{url, inherited}'

# 2. What does instance-config.yml say right now?
docker exec utmstack-backend cat /updates/instance-config.yml | grep ^server

# 3. If they disagree → reset SOCAI config; EnsureDefault will re-derive.
```

If `inherited: false` → it's cause #2 (tenant override). If `inherited: true` but URL is prod → cause #1 (default was baked pre-flip). Both fix with a reset.
