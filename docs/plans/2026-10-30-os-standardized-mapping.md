# OpenSearch standardized-field mapping (log + alert indices)

## Objective (user)
Predefine a manual mapping for ALL known standardized fields in OpenSearch so a field
cannot be inferred as `keyword` in one index and `ip` in another. The ONLY field group
allowed to be inferred is `utm.*`; `event.*` stays `flat_object`. Applies to log AND alert
indices.

## Root cause of the visible symptom
`logIndexMappings` in `installer/services/search.go` maps `origin`/`target` as
`{"type":"object","dynamic":true,"properties":{ip,port,user,host,file,bytesSent,bytesReceived}}`.
So unlisted sub-fields — most importantly `origin.geolocation.*` (used by the
coordinate-map charts) — are inferred per-index. On the lab, `origin.ip.keyword` terms-agg
returns 0 buckets → every COORDINATE_MAP_CHART shows "No data found".

## Canonical schema source
`go-sdk@v1.1.38/plugins/plugins.proto`:
- `Event` (10 fields: id,@timestamp,deviceTime,dataType,dataSource,tenantId,tenantName,raw,event(flat),target,origin,protocol,connectionStatus,statusCode,actionResult,action,severity,errors,compliance,controls,utm)
- `Side` (origin/target): bytesSent,bytesReceived,packagesSent,packagesReceived,ip,host,user,group,port,domain,mac,geolocation{country,city,latitude,longitude,asn,aso,countryCode,accuracy},url,cidr,certificateFingerprint,ja3Fingerprint,jarmFingerprint,sshBanner,sshFingerprint,cookie,jabberId,email,dkim,dkimSignature,emailAddress,emailBody,emailDisplayName,emailSubject,emailThreadIndex,emailXMailer,whoisRegistrant,whoisRegistrar,process,processState,command,windowsScheduledTask,windowsServiceDisplayName,windowsServiceName,file,path,filename,sizeInBytes,mimeType,hash,authentihash,cdhash,md5,sha1,sha224,sha256,sha384,sha3224,sha3256,sha3384,sha3512,sha512,sha512224,sha512256,hex,base64,operatingSystem,chromeExtension,mobileAppId,cpe,cve,malware,malwareFamily,malwareType,pgpPrivateKey,pgpPublicKey,connections,usedCpuPercent,usedMemPercent,totalCpuUnits,totalMem,disks[{name,totalSpace,usedPercent}]
- `Alert` (id,@timestamp,lastUpdate,name,tenantId,tenantName,dataSource,dataType,category,technique,description,references[],impact{confidentiality,integrity,availability},impactScore,severity,adversary(Side),target(Side),events[Event],deduplicatedBy[],errors[],parentId,groupBy[])
- JSON naming: proto3 json_name = lowerCamelCase (verified on the lab: `deduplicatedBy`, `lastUpdate`, `emailXMailer`).

## Decisions (locked)
1. **Log index mapping**: full explicit properties for Event + origin/target as `object` with `dynamic:false`, ALL Side fields + geolocation + disks typed. `origin.ip`/`target.ip`/`adversary.ip` = `ip` + `.keyword` multi-field (maps bucket on `origin.ip` work; aggs on `.keyword` also work — cross-generation compatible per the v11.3.0 finding).
2. **Alert index mapping**: full explicit properties for Alert top-level (severity INT, impactScore long, impact{...} int, deduplicatedBy/groupBy/errors/references keyword arrays, adversary/target Side objects) AND full Event-shaped mapping under `lastEvent` and `events` (embedded event copies must have the same deterministic types). `event` bags stay `flat_object`.
3. **utm.***: `object` with `dynamic:true` + explicit `keyword` pin for every known key (the user-approved "known pinned, rest inferred" pattern).
4. **event.***: `flat_object` (unchanged).
5. **Existing indices**: templates only apply to new indices. Add an `UpdateOpenSearch` retro-PUT of the FULL mapping (not just newBagMappings) — a full-mapping PUT is idempotent for fields that already match and fixes inferred ones where they don't conflict. Verify empirically on the lab which fields currently conflict; if a field already has a different type in an existing index, the PUT for THAT field 400s — handle by field-level pre-check (only PUT fields whose current mapping is absent or identical).

## Files to change
- `installer/services/search.go` — `logIndexMappings`, `alertIndexMappings`, `newBagMappings` (retro), `UpdateOpenSearch` (retro strategy)
- `installer/services/search_test.go` — extend validation for the new structure
- `docs/specs/` — short spec doc for the canonical mapping (field → type table)

## Verification — DONE (2026-10-30)
- **Unit (6 tests, all green):** `TestLogIndexMappingsValidJSON`, `TestNewBagMappingsValidJSON`, `TestAlertIndexMappingsValidJSON` (existing) + `TestLogIndexMappingsCanonical`, `TestAlertIndexMappingsCanonical`, `TestNewBagRetroIsAdditive` (new — pin dynamic:false on Sides, ip+.keyword, geo types, event flat_object, utm dynamic:true with 176 pinned keys, alert severity long, embedded event bags).
- **Live OpenSearch validation (lab 192.168.99.75):** all four installer operations PUT clean (200): log template, alert template (carries `total_fields.limit:50000` because the 1180-field alert mapping exceeds the 1000-field template-validation default), additive log retro (ip .keyword + bags), additive alert retro (bags only — the utm pin was stripped because existing alert indices already inferred `events.utm.*` as text and pinning keyword 400s; utm.* is the one inferred bag by design).
- **Fresh canonical index proof:** `v11-log-e2e-canonical` created under the template then a Side-rich doc indexed → mapping came out exactly canonical: root `dynamic:false`, `origin`/`target` `dynamic:false`, `origin.ip`=ip+.keyword, `geolocation.asn`=long/`lat`=double, `event`=flat_object, `controls`=keyword, `utm.dynamic:true` (unknown `utm.NewField` inferred), `severity`=text (log) vs long (alert).
- **Maps render again:** `/dashboard/render/227/aws-cloud` → leaflet `object`, 8 tiles loaded, **2 markers**, 0 page errors (was 0 leaflet/0 tiles + `Unexpected token export` before the bundle fix).

## Design decisions locked
1. **Templates = full canonical schema** (the deliverable): every standardized field explicitly typed; nested typed objects `dynamic:false`; `origin.ip`/`target.ip`/`adversary.ip` = `ip` + `.keyword` sub-field (fixes coordinate-map terms-agg); `utm.*` = the ONLY inferred bag (`dynamic:true`, 176 known keys pinned keyword); `compliance` stays dynamic (control keys vary per standard, values always string arrays); `event.*` = `flat_object`.
2. **severity is family-scoped:** `text` in `v11-log-*` (Event.severity is a proto string), `long` in `v11-alert-*` (alerts plugin writes an int) — separate index families, both pinned.
3. **Retro = additive only** (`newBagMappings`, `alertRetroMappings`): adds the ip `.keyword` sub-field + flat_object bags; never retypes an existing leaf (a base-type change 400s and needs a reindex). Old indices fully converge to the canonical template on their next reindex/ILM roll-off.
4. **Templates carry `index.mapping.total_fields.limit:50000` in their own settings** so composable-template validation passes even before the settings-only `utmstack_indexes` template composes in.

## Root cause found (why maps were empty)
`logIndexMappings` mapped `origin`/`target` as `{"type":"object","dynamic":true,"properties":{...7 fields...}}`, so unlisted Side sub-fields (most importantly `origin.geolocation.*`) were **inferred per-index**. The coordinate-map charts bucket on `origin.ip.keyword`; on indices where `origin.ip` was inferred as bare `ip` (no `.keyword` sub-field), the terms-agg silently returns 0 buckets → "No data found". This mapping pins the type once, for every index, forever.
