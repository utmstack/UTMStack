# Native Go analysis engine — implementation and migration status

The later fresh-inspection source is preserved through a seven-input bridge;
its two affected packages, races, whole-agent vet and rebuilt endpoints pass
17 additional checks. See the [publication bridge](../../docs/edr/native-engine/audits/2026-11-17-v281-inspection-bridge/REPORT.md).
Analysis inputs and the preceding metadata/image evidence are unchanged.

November 17 combined metadata checkpoint: `12.0.113-native` / `native-go-v281`
preserves published image112 and Windows cancellation work while correcting
scan-wide JSON root scope, legacy metadata ancestry/options, PRECLASS extraction
depth and path-aware cache bypass for all metadata-enabled scans. The 1,227-input
combined source passes 27 checks, including the complete host suite, races and
five-platform endpoint builds. All 46 harmless native metadata controls and 950 retained
feed outcomes agree with their expected results. Frozen three-platform image
proof is verified and reused; no new combined native acceptance is claimed.
Full replacement and production readiness remain open, with 118 strict-feed
conditions. See the [combined audit](../../docs/edr/native-engine/audits/2026-11-17-v281-combined-platform/REPORT.md).
The bzip2 boundary prototype is subsequent work and is not part of 113/281.

November 17 combined source checkpoint: `12.0.111-native` / `native-go-v279`
includes the published Windows/Linux source corrections through `6be0b5249`.
Fresh full-suite, vet, core/platform race and 29 validation checks pass;
102 scanner controls and 950 feed outcomes are unchanged. Frozen platform
proof is verified and reused, with no new native execution counted. The 575
core analysis inputs are identical to the standalone v279 candidate. See the
[combined audit](../../docs/edr/native-engine/audits/2026-11-17-v279-combined-platform/REPORT.md)
for exact source identity and limits. Native combined-source and production
qualification remain pending; the strict feed gate retains 118 conditions.

November 17 carried JavaScript continuation: `native-go-v279` /
`12.0.111-native` closes the older oracle's 115 carried-state incompletes.
All 2,846 new native byte controls and all 2,247 older pages agree. Actual full
scanners agree on 36 detections and 66 clean controls, correcting six prior
false cleans and completing 50 prior incompletes. The complete host suite,
vet, races, 100,000 carry fuzz executions and cross-builds pass; all 950 retained
feed outcomes are unchanged. Native platform acceptance and full replacement
remain open; the strict feed gate retains 118 conditions. See the
[v279 evidence and limits](../../docs/edr/native-engine/audits/2026-11-17-v279-js-carried-state/REPORT.md).

November 17 ordered JavaScript continuation: `native-go-v278` /
`12.0.110-native` processes packer, dF and eval transforms in one forward walk.
Its 363 new native byte controls and 136 actual full-scanner comparisons agree,
including correction of six prior false cleans and five positive incompletes.
The older HTML oracle retains the exact same 115 incomplete pages and zero
differences; all 950 retained feed outcomes are unchanged. Full host suite, vet,
races, 100,000 ordering fuzz executions and cross-builds pass. Native platform,
unproved carried-state, broader analysis and full replacement gates remain open;
the strict feed gate retains 118 conditions. See the
[v278 evidence and limits](../../docs/edr/native-engine/audits/2026-11-17-v278-js-decoder-order/REPORT.md).

November 17 JavaScript continuation: `native-go-v277` / `12.0.109-native`
adds bounded repeated named packer calls and decoded-function scope handling.
All 204 new native byte-output controls agree; the older 2,247-page oracle
retains zero differences with 115 still-incomplete pages. Actual full-scanner
controls agree on nine detections and 26 negatives, replacing 30 previous
incomplete outcomes. All 950 retained scan outcomes stay unchanged. The complete
host agent suite, vet, focused races, over 100,000 fuzz executions and cross-builds
pass. Native platform, broader JavaScript and full replacement qualification
remain open; the strict feed gate still has 118 conditions. See the
[v277 evidence and limits](../../docs/edr/native-engine/audits/2026-11-17-v277-js-named-calls/REPORT.md).

November 17 Go LZMA continuation: `native-go-v276` / `12.0.108-native` adds
bounded ClamBC decoding across input/output windows, native error cleanup and
separate shared dictionary/model storage ceilings. All 2,842,278 reference
decisions and 24 full-scanner controls agree; ten formerly incomplete positives
detect and all 950 retained outcomes are unchanged. Host tests, vet, races,
100,000 fuzz executions and cross-builds pass. These are development checks;
new native platform, efficacy and production qualification remain pending.
The strict gate still has 118 conditions. The user's November 17 instruction
resumes core work after publication. See the
[v276 evidence and limits](../../docs/edr/native-engine/audits/2026-11-17-v276-lzma-stream/REPORT.md).

November 14 Go bzip2 continuation: `native-go-v275` / `12.0.106-native`
implements bounded ClamBC decompression across input/output pipe windows.
The independent native adapter corpus and 16 full-scanner controls agree;
seven former positive incompletes detect, with all 950 retained outcomes
unchanged. Host tests, vet, races, fuzz and cross-builds pass. This is not new
native endpoint or production acceptance. Error re-entry, LZMA streaming and
the broader full-replacement gates remain open; the strict gate still has
118 conditions. See the [v275 evidence and limits](../../docs/edr/native-engine/audits/2026-11-14-v275-bzip2-stream/REPORT.md).

November 12 later local main: package `12.0.104-native` / matcher `native-go-v274`
adds reviewed retention-hold release. Both local mains include qualified audit
commit `8a5d17df2a`, following user-authorized Linux/Windows coordination. The
[release audit](../../docs/edr/native-engine/audits/2026-11-12-v274-recovery-release/REPORT.md)
records 128 native Windows AMD64 and 147 Linux ARM64 top-level Go passes, explicit
skips, real process exits, earlier-v102 API behavior and two bounded guest resets
with synced starting artifacts. The final Windows test-only correction rebuilds
to the exact already-executed Linux program; that proof is reused, not recounted.
All cloud resources are independently absent; no local Docker was invoked.

The earlier unflushed raw-source reset remains failed: the same raw file ID
contains seven NUL bytes while the separate durable snapshot retains payload.
Original context/journal bytes, diagnostic identities and hashes are sealed.
No production admission fix or arbitrary hardware power-loss qualification is
claimed. This is the next Windows W2 investigation. Release authorizes ordinary
recovery of the original operation; cancellation, archived-stage protection,
fresh restore, production delivery and full replacement remain open. See the
[operator procedure](quarantine/RECOVERY.md). No remote push or deployment occurred.

Earlier November 12 checkpoint: package `12.0.103-native` retains matcher
`native-go-v274` and includes the coordinated restore-stage safeguards
(`28a54ad850`) and explicit recovery retention holds (`8ed133f890`). Both source
clones were fast-forwarded after user-authorized coordination with the Linux
chat. Windows owns further shared quarantine/cache recovery-decision edits;
Linux's later service line and byte-safe admission prototype remain isolated
pending shared review and combined-source Windows qualification. No remote push,
signing or deployment occurred.

The [restore-stage audit](../../docs/edr/native-engine/audits/2026-11-12-v274-restore-stage/REPORT.md)
retains the original and legacy-stage admission failures, corrected native proof,
bounded guest resets and cleanup. Journal/root/live-stage identities validate
lookup hints; those hints never prove publication or completion. The
[retention-hold audit](../../docs/edr/native-engine/audits/2026-11-12-v274-recovery-hold/REPORT.md)
records 116 native Windows and 135 Linux ARM64 top-level Go passes, exact source
and program identities, process-exit boundaries and earlier-v102 refusal.
All campaign cloud resources were removed. A hold retains the original journal
and artifacts; release, cancellation, fresh retry and hold power-loss proof
remain pending. The matcher/conditional-feed gate and broader W2–W6 remain open.

November 12: package 12.0.101-native retains engine v274 and adds read-only
[quarantine recovery inspection](quarantine/RECOVERY.md). Its host/compile
evidence is retained separately from the preceding candidate's native campaign.
See the [inspection audit](../../docs/edr/native-engine/audits/2026-11-12-v274-recovery-inspection/REPORT.md).
The [GCP follow-up](../../docs/edr/native-engine/audits/2026-11-12-v274-recovery-gcp/REPORT.md)
executes the inspection/CLI checks on native Windows AMD64 and Linux ARM64, plus
the full Linux quarantine suite. It fixes test-only Windows permission setup;
production code is unchanged. Further runtime checks use disposable GCP VMs.
The user's clarified November 12 instruction permits Docker only inside GCP
VMs: no local Docker commands, including inventory or cleanup, and no request
to restart it on this Mac. Delete the disposable cloud resources after testing.

November 11/12 Windows continuation: candidate v274 / 12.0.100-native has bounded
Windows Server 2022 AMD64 acceptance with 37 checks and independently absent
lab resources. The [current audit](../../docs/edr/native-engine/audits/2026-11-11-v274-windows-native/REPORT.md)
records the exact composite proof, native fixes and remaining limits. Native
ARM64 testing is excluded by the user. Production readiness and full replacement
remain open; earlier entries below are dated checkpoints.

November 11 owned AV RSA and root-key binding: `native-go-v272` /
`12.0.98-native` matches all 423,974 primitive C decisions and validates 6,984
new signed controls. There are 150 intentional root-substitution protections:
the changed exponent lacks proof under the actual authorized key. Separately
signed nonstandard roots remain supported. Final policy corrects 1,358 v271
results, including 36 inappropriate trust exemptions. All 21,060 defined earlier
signed outcomes agree; 16 original C crashes remain undefined. Original public
integers and signed bytes stay authoritative; TLS/feed CMS are unchanged.
Complete agent tests, vet, race, five EDR and five integrated-agent builds,
Windows test builds and 400,000 fuzz executions pass. Native suites, 950 unchanged
replays, 296 quarantines and real restore/purge/recovery pass. Measured peak is
1.497 GiB under hard 2 GiB/no swap. All 12 local containers are independently
absent; no cloud resources or private uploads were created. Temporary builds and
caches are removed after evidence preservation (5.03 GiB logical).
See the [v272 audit](../../docs/edr/native-engine/audits/2026-11-11-v272-rsa-numerics/REPORT.md) for source identities, the explicit protective
difference from ClamAV and remaining limits. The strict gate still fails on
**118 conditional rules: 110 fuzzy-image and eight certificate-trust**.
Remaining certificate grammar/store profiles, fuzzy image behavior, unpackers,
documents and other bytecode hooks, native Windows/AMSI, compatible Linux
prevention, ambiguous/lossless-path and power-loss recovery, independent efficacy,
combined workloads and signed authenticated delivery remain open.
**Full replacement remains the active, unqualified goal.** Earlier entries are
dated checkpoints.

November 11 ordered certificate stores: `native-go-v271` / `12.0.97-native`
uses common ordered admission and authenticated purposes for PE, catalog and
legacy timestamp verification. All 4,536 new signed comparisons match, correcting
268 inappropriate v270 trust exemptions. All 16,524 defined retained outcomes
still agree; 16 original reference crashes remain undefined. Separate admissions
of identical raw certificates no longer acquire merged authority. Exhausted work
stays incomplete for PE and rejects catalog publication. Complete agent tests,
vet, race, five EDR and five integrated-agent builds, Windows test builds and
200,000 final-source fuzz executions pass. Native suites, 950 unchanged replays,
296 actual quarantines and real restore/purge/recovery pass. Measured peak is
1.481 GiB under hard 2 GiB/no swap. All 12 exact containers are independently
absent; no cloud resources or private uploads were created. Temporary builds and
caches are removed after evidence preservation (4.99 GiB logical).
See the [v271 audit](../../docs/edr/native-engine/audits/2026-11-11-v271-certificate-store/REPORT.md) for exact identities and limits.
The strict replacement gate still fails on **118 conditional rules: 110 fuzzy-image
and eight certificate-trust**. Numeric/wire and broader store profiles, fuzzy
image behavior, unpackers/documents/other bytecode hooks, Windows/AMSI native
runtime, compatible Linux prevention, ambiguous/lossless-path and power-loss
recovery, independent efficacy, combined workloads and signed authenticated
delivery remain open. **Full replacement remains the active, unqualified goal.**
Earlier entries are dated checkpoints.

November 11 certificate framing and ordered bags: `native-go-v270` /
`12.0.96-native` matches 71,380 independent object decisions, all 7,022 new
signed-file controls and 9,502 defined retained certificate outcomes. Sixteen
original reference crashes remain undefined. The sealed v269 rebuild reproduces
all 5,248 earlier differences, including 26 inappropriate trust exemptions.
Original signed bytes remain authoritative, while exhausted work stays incomplete.
Complete agent tests, vet, race, five scanner and five integrated-agent builds,
Windows test builds and 200,000 fuzz inputs pass. Native suites, 950 unchanged
replays, 296 actual quarantines and restore/purge/recovery pass. Peak measured use
is 1.472 GiB under hard 2 GiB/no swap. All 12 containers are independently
absent; no cloud resources or private uploads were created. Temporary builds and
caches are removed after evidence preservation (4.63 GiB logical).
See the [v270 audit](../../docs/edr/native-engine/audits/2026-11-11-v270-certificate-framing/REPORT.md) for identities, measurements and limitations.
The strict gate still fails on **118 conditional rules: 110 fuzzy-image and eight
certificate-trust**. Global certificate selection, numeric keys and remaining
wire profiles; fuzzy decoders/numerics, unpackers, document variants and other
bytecode hooks; current Windows/AMSI runtime and power loss; compatible Linux
prevention; recovery ambiguities/lossless paths; independent efficacy, combined
workloads and signed authenticated delivery remain open. **Full replacement
remains the active, unqualified goal.** Earlier entries are dated checkpoints.

November 11 Authenticode purpose verification: `native-go-v269` / `12.0.95-native`
uses one bounded, authenticated policy for PE, catalog and timestamp paths.
All 1,564 new reference decisions match, correcting 632 frozen-v268 differences:
50 inappropriate trust exemptions and 582 missed trust decisions. All 7,938
defined earlier certificate outcomes match; 16 original reference crashes remain
undefined. Original signed bytes remain authoritative, and general TLS policy
is unchanged. Complete agent tests, vet, race, five scanner and five integrated
agent builds, Windows test builds and 200,000 fuzz executions pass. Native
suites, 950 unchanged replays, 296 actual quarantines and real restore/purge/
recovery pass. Measured peak is 1.495 GiB under hard 2 GiB/no swap. All 13
containers, including one failed harness launch, are independently absent;
no cloud resources or private uploads were created. Temporary builds and caches
are removed after evidence preservation. See the [v269 audit](../../docs/edr/native-engine/audits/2026-11-11-v269-certificate-constraints/REPORT.md) for
source identities, independent controls, fixture diagnostics and limitations.
The strict gate still fails on **118 conditional rules: 110 fuzzy-image and eight
certificate-trust**. Certificate selection, numeric key and outer ASN.1 profiles;
fuzzy decoders/numerics, unpackers, document variants and other bytecode hooks;
current Windows/AMSI runtime and power loss; compatible Linux prevention;
recovery ambiguities/lossless paths; independent efficacy, combined workloads
and authenticated signed delivery remain open. **Full replacement remains the
active, unqualified goal.** Earlier entries below are dated checkpoints.

November 11 certificate-time compatibility: `native-go-v268` / `12.0.94-native`
passes 1,468 independent signed-file comparisons, correcting 580 frozen-v267
disagreements. All 49,342 time-value controls and 6,470 defined prior certificate
outcomes agree; 16 prior reference crashes remain undefined. Host tests, vet,
race, five executable builds, Windows test builds and 200,000 fuzz executions
pass. Native suites, 950 unchanged replay outcomes, 296 actual quarantines and
real restore/purge/recover pass. Peak measured use is 1.483 GiB under hard
2 GiB/no swap. All 14 owned containers are independently absent; no cloud
resources or private uploads were created. Temporary builds and caches are
removed after evidence preservation. The strict replacement gate still fails
on **118 conditional rules: 110 fuzzy-image and eight certificate-trust**.
Certificate profiles, fuzzy decoders/numerics, unpackers, document variants and
other bytecode hooks; current Windows/AMSI runtime and power loss; compatible
Linux prevention; recovery ambiguities/lossless paths; independent efficacy,
combined workloads and signed authenticated delivery remain open.
**Full ClamAV replacement remains unqualified.** See the [v268 audit](../../docs/edr/native-engine/audits/2026-11-08-v268-certificate-times/REPORT.md).

> November 8 icon-reader and comparison completion: `native-go-v267` /
> `12.0.93-native` matches all 3,030 reference hashes across 3,042 inert PEs,
> including every square width 16..256 at all six depths. Wrapping physical
> header offsets, palette alpha and scaled width buckets are implemented;
> icon conditions now gate logical bytecode callbacks. All 6,084 ClamAV decisions
> match the signature layer and both host/Linux full engines. Frozen v266 had
> 3,084 disagreements, including 48 missed detections. Four independent CBC
> controls agree. The signed pinned-feed audit removes the blanket unsupported
> classification from 81 icon rules: **118 conditional rules remain, 110 fuzzy-image
> and eight certificate-trust**. The strict replacement gate still fails.
> Tests, vet, race, five executable builds, four Windows test builds and 200,000
> fuzz inputs pass. All 950 retained outcomes are unchanged. The Linux service
> verifies 296 quarantines plus restore/purge/recover and both receipts. Peak
> measured memory is 1.464 GiB under hard 2 GiB/no swap, without OOM. All thirteen
> owned containers are removed; 3.69 GiB of unique logical temporary data is removed.
> No cloud resources or private uploads were created. See the [v267 audit](../../docs/edr/native-engine/audits/2026-11-08-v267-icon-offsets/REPORT.md).
> Remaining work includes fuzzy decoders/numerics, certificate trust, unpackers,
> document variants and other bytecode hooks; wider native Windows/AMSI and icon
> profiles; recovery ambiguity/lossless paths and power loss; compatible-kernel
> prevention; independent efficacy, combined workloads and signed authenticated
> production delivery. **Full ClamAV replacement remains unqualified.**

> November 8 icon-resource parity and PE scan integration: `native-go-v266` /
> `12.0.92-native` matches 384 ClamAV comparisons over 192 inert PEs. Frozen v265
> had two false detections, 38 missed detections and 262 incomplete nonmatches.
> The owned walker now preserves 16-bit counters, entry order and readable prefixes;
> icon traversal/comparison shares scan budgets and stops after a proven match.
> A separate PE scan view resolves 304 full-engine incomplete results. Signed-image
> trust retains complete-extent validation; truncated security-table declarations
> remain incomplete. Host and Linux full engines match every new oracle outcome.
> Tests, vet, race, five executable builds and Windows test builds pass. There are
> 300,000 verified fuzz executions across development, including 200,000 covering
> the final relevant implementations. All 950 retained replay outcomes are unchanged.
> The Linux service verifies 296 quarantines and real restore/purge/recover with
> both receipts. Peak measured cgroup use is 1.477 GiB under hard 2 GiB/no swap,
> without OOM. All 26 owned containers are removed, including superseded runs;
> 4.71 GiB of cumulative temporary logical data is removed. No cloud resources or
> private uploads were created. See the [v266 audit](../../docs/edr/native-engine/audits/2026-11-08-v266-icon-resources/REPORT.md).
> The strict pinned-feed gate still fails on 199 conditional rules: 110 fuzzy-image,
> 81 icon-group and eight certificate-trust. Wider image/icon/certificate profiles,
> unpackers/bytecode, Windows/AMSI runtime, recovery qualification, compatible-kernel
> prevention, independent efficacy, combined workloads and signed authenticated
> production delivery remain pending. Full ClamAV replacement is unqualified.

> November 8 durable restore recovery: `native-go-v265` / `12.0.91-native`
> adds private restore journals and an atomic SQLite publication receipt. Verified
> outputs are copied independently; the stored backup survives until database
> commit, then identity-checked cleanup retires it. Startup/recover preserves later
> output edits, replacements and removals. Unidentified creation or ambiguous
> publication remains visible for attention; recovery never guesses a new target.
> Twenty restore abrupt-exit checkpoints pass on Mac and Linux, together with the
> prior 36 purge crash cases per platform. Twelve restore and twelve purge real
> Linux CLI configurations pass. Tests, vet, race, five executable builds, Windows
> test builds and 100,000 restore-journal fuzz executions pass; Windows runtime
> and hardware power loss remain unqualified. All 950 replay outcomes are unchanged.
> The actual Linux service verifies 296 quarantines and real restore/purge/recover,
> including both receipts. Peak cgroup use is 1.480 GiB under hard 2 GiB/no swap,
> without OOM. All thirteen owned containers are removed; 3.53 GiB of unique logical
> temporary data is removed. No cloud resources or private uploads were created.
> See the [v265 audit](../../docs/edr/native-engine/audits/2026-11-08-v265-restore-recovery/REPORT.md).
> Remaining recovery work includes unidentified creation/ambiguous publication
> resolution, admission lossless-path migration and wider namespace/platform tests.
> The strict gate still fails on 199 conditional rules: 110 fuzzy-image, 81 icon-group
> and eight certificate-trust. Wider detection, unpackers/bytecode, Windows/AMSI,
> compatible-kernel prevention, independent efficacy, combined workloads and signed
> authenticated delivery remain pending. Full ClamAV replacement is unqualified.

> November 8 durable purge recovery: `native-go-v264` / `12.0.90-native`
> adds private write-ahead purge journals, lossless approval identities and atomic
> SQLite completion receipts. Startup/recover reconciles approved work; replacement
> names are preserved, restore history stays intact and retention errors surface.
> Twenty-one interruption boundaries, 36 actual process crashes and twelve real
> Linux CLI recovery configurations pass, including reused-name refusal/repair.
> Native cross-volume and foreign-owned unreadable-file controls pass. Tests, vet,
> race, five executable builds, Windows test builds and 200,000 verified fuzz
> executions pass; Windows runtime and power-loss behavior remain unqualified.
> All 950 replay outcomes are unchanged; the actual Linux service verifies 296
> quarantines and real restore/purge/recovery, including its completion receipt.
> Peak measured cgroup use is 1.493 GiB under hard 2 GiB/no swap, without OOM.
> All thirteen owned containers are removed, including a preserved CLI setup
> failure; 3.54 GiB of unique logical temporary data is removed. No cloud resources
> or private uploads were created. See the [v264 audit](../../docs/edr/native-engine/audits/2026-11-08-v264-purge-recovery/REPORT.md).
> Restore filesystem/SQLite crash journaling, wider path/namespace qualification,
> Windows/AMSI, independent efficacy, combined workloads and authenticated production
> delivery remain pending. The strict gate still fails on 199 conditional rules:
> 110 fuzzy-image, 81 icon-group and eight certificate-trust; wider detection,
> unpackers and bytecode work also remains. Full replacement is unqualified.

> November 8 owned resize math: `native-go-v263` / `12.0.89-native`
> uses Go-owned Mac/Linux sine and Lanczos3 profiles, with explicit f32 rounding.
> Each profile passes all 16,384 admitted dimensions, 1,535,181,860 sine evaluations
> and 768,074,104 normalized weights on Mac and native Linux ARM64. All fifteen
> previously partial controls in the 1,425-case image ledger now match definitely.
> The expanded 1,681-case ledger has 1,679 definite decisions and two correctly
> conditional native-profile disagreements; 421 reader refusals remain separate.
> Another 3,456 tiny WebP comparisons have zero mismatches or uncertain results.
> Tests, vet, race, five executable builds, Windows test builds and 220,000 verified
> fuzz executions pass. All 950 replays retain their outcomes; the actual Linux
> service verifies 296 quarantines and real CLI restore/purge/recovery. Peak measured
> cgroup use is 1.476 GiB under hard 2 GiB/no swap, without OOM. All eleven local
> test containers are removed and 4.09 GiB of owned logical temporary data is removed.
> No cloud resources or private uploads were created. See the [v263 audit](../../docs/edr/native-engine/audits/2026-11-08-v263-image-resize/REPORT.md).
> The strict gate still fails on 199 conditional rules: 110 fuzzy-image, 81 icon-group
> and eight certificate-trust. Wider math/decoder/icon/certificate compatibility,
> unpackers/bytecode, restore/purge crash journaling, Windows/AMSI runtime,
> compatible-kernel prevention, independent efficacy, combined workloads and
> signed authenticated production delivery remain pending. Full replacement is unqualified.

> November 8 CRB selection and issuer time policy: `native-go-v262` / `12.0.88-native`
> retains duplicate-rule precedence, reports unnamed blocked certificates and
> verifies local Authenticode issuer/TSA policy with original-byte signature proofs.
> Primary code-signer validity remains required; TLS/feed CMS policy is unchanged.
> All 1,144 final controls match ClamAV on macOS and Linux, including detection names.
> Frozen v261 differs on 268: 128 attribution errors, 108 false rejections and 32
> incomplete results where the reference detects a block. Another 5,326 defined
> prior decisions match; 16 prior reference crashes remain separate. Full tests,
> vet, race checks, five executable builds, Windows test builds, native Linux suites,
> 400,000 verified fuzz executions, pinned CRB retention and 27 static real-file
> controls pass. All 950 replay outcomes remain unchanged; the actual Linux service
> reconciles 296 quarantines and CLI restore/purge/recover. Maximum measured cgroup
> use is 1.515 GiB under hard 2 GiB/no swap, without OOM. See the
> [v262 audit](../../docs/edr/native-engine/audits/2026-11-08-v262-certificate-policy/REPORT.md). All seven local containers are removed;
> 3.50 GiB of task-owned temporary files/cache/downloads are removed. No cloud
> machines or private uploads were created. The strict gate still fails on 199
> conditional rules, including all eight certificate-trust rules. Broader time,
> constraint/selection/CRB numeric and RSA cases, image/icon compatibility,
> unpackers/bytecode, restore/purge crash journaling, Windows/AMSI, compatible-kernel
> prevention, independent efficacy, combined workloads and authenticated production
> delivery remain pending. Full replacement is unqualified.

> November 8 raw certificate identities and CRB state: `native-go-v261` / `12.0.87-native`
> preserves original name/serial identity bytes through metadata normalization,
> with parsing limits applied by grammar position. Legacy CRB wildcard/not-before
> state is confined to each selected source. All 1,128 final controls match ClamAV
> on macOS and Linux. Frozen v260 differs on 572: 36 inappropriate exemptions,
> 464 ordinary false rejections and 72 incomplete results (56 detected/16 clean
> in the reference). Another 4,198 defined prior decisions match; 16 prior reference
> crashes remain separate. Full tests, vet, race checks, five executable builds,
> Windows test builds, native Linux suites, 400,000 verified fuzz executions and
> 27 static real-file controls pass. All 950 replay outcomes remain unchanged;
> the actual Linux service reconciles 296 quarantines and CLI restore/purge/recover.
> Maximum measured cgroup use is 1.500 GiB under hard 2 GiB/no swap, without OOM.
> See the [v261 audit](../../docs/edr/native-engine/audits/2026-11-08-v261-certificate-identities/REPORT.md). All seven local containers are removed;
> 3.58 GiB of task-owned temporary files/cache/downloads are removed. No cloud
> machines or private uploads were created. The strict gate still fails on 199
> conditional rules, including all eight certificate-trust rules. Certificate
> validity/extensions/selection and duplicate/CRB issuer policy, numeric key ranges,
> image/icon compatibility, unpackers/bytecode, restore/purge crash journaling,
> Windows/AMSI, compatible-kernel prevention, independent efficacy, combined workloads
> and authenticated production delivery remain pending. Full replacement is unqualified.

> November 8 RSA public-key encodings: `native-go-v260` / `12.0.86-native`
> validates bounded RSA key framing and unsigned integer encodings before Go's
> metadata parser, retaining original certificate, TBS and SPKI bytes. The original
> 624 configurations expose 300 v259 differences. All 900 final configurations
> match ClamAV on macOS and Linux; frozen v259 reproduces 72 inappropriate marker
> exemptions and 467 false rejections there. Another 3,298 defined prior outcomes
> match; 16 prior reference crashes remain separate. Full tests, vet, race checks,
> five executable builds, Windows test builds, native Linux suites, 300,000 verified
> fuzz executions and 27 static real-file controls pass. All 950 replay outcomes
> remain unchanged; the actual Linux service reconciles 296 quarantines and real
> CLI restore/purge/recover. Maximum measured cgroup use is 1.506 GiB under hard
> 2 GiB/no swap, without OOM. See the [v260 audit](../../docs/edr/native-engine/audits/2026-11-08-v260-public-keys/REPORT.md).
> All eight local containers, including a preserved harness launch failure, are
> removed; 3.28 GiB of task-owned temporary files/cache/downloads are removed.
> No cloud machines or private uploads were created. The strict gate still fails
> on 199 conditional rules, including all eight certificate-trust rules. Broader
> certificate metadata/selection/numeric key ranges, image/icon compatibility,
> unpackers/bytecode, restore/purge crash journaling, Windows/AMSI runtime, Linux
> permission prevention, independent efficacy, sustained combined workloads and
> authenticated production delivery remain pending. Full replacement is unqualified.

> November 8 certificate encoding and recovery: `native-go-v259` / `12.0.85-native`
> validates original certificate algorithms and numeric RSA proofs while retaining
> the original signed bytes. It recovers known-invalid framed certificate entries
> without silently ignoring unsupported forms or bypassing the certificate budget.
> The original 380 configurations expose 132 v258 differences, all corrected on
> identical inputs. All 1,324 expanded configurations match ClamAV: the frozen old
> source reproduces 180 inappropriate marker exemptions and 304 rejected valid
> trust decisions there. Another 1,974 defined prior outcomes remain unchanged;
> 16 prior reference-crash configurations stay explicitly separate. Full tests,
> vet, focused race checks, 200,000 fuzz executions, five executable builds,
> Windows test builds, native Linux suites and 27 static real-file controls pass.
> All 950 replay outcomes remain unchanged. The actual Linux service reconciles
> 296 verified quarantines and real CLI restore/purge/recover. Maximum measured
> cgroup use is 1.479 GiB under hard 2 GiB/no swap, without OOM. See the
> [v259 audit](../../docs/edr/native-engine/audits/2026-11-08-v259-certificate-wire/REPORT.md). All seven local containers are removed,
> and 3.46 GiB of task-owned temporary files/cache/downloads are removed. No cloud
> machines or private uploads were created. The strict gate still fails on 199
> conditional rules, including all eight certificate-trust rules. Broader certificate
> forms/selection, image/icon compatibility, unpackers/bytecode, restore/purge crash
> journaling, Windows/AMSI runtime, Linux permission prevention, independent efficacy,
> sustained combined workloads and authenticated production delivery remain pending.

> November 8 verified quarantine admission: `native-go-v258` / `12.0.84-native`
> connects the scanned file identity to an independent verified snapshot, privately
> retains the original inode, and journals interrupted admission for idempotent
> recovery. Nine boundaries pass with both abrupt exit and forced termination.
> Source replacements, retained writers, invalid journals and recovery collisions
> have explicit controls. Native Linux exposed and verified a directory-cleanup
> fix. Full tests, vet, race checks, five executable builds and Windows test builds
> pass; Windows execution remains pending. Thirteen successful native acceptance
> runs include two actual service sweeps with 592 matching quarantine identities,
> real CLI restore/purge/recover, and zero drift across 950 replay fixtures.
> Maximum measured cgroup use is 1.497 GiB under hard 2 GiB/no swap, without OOM.
> See the [v258 audit](../../docs/edr/native-engine/audits/2026-11-08-v258-quarantine-admission/REPORT.md). All 20 local test containers
> are removed; final task-owned temporary/cache cleanup removes 3.87 GiB. No cloud
> resources, private uploads or new remote artifact downloads were created.
> The strict gate still fails on 199 conditional rules (110 fuzzy-image, 81 icon,
> eight certificate). Full replacement remains unqualified: restore/purge crash
> journaling, wider compatibility, Windows/AMSI, compatible-kernel prevention,
> independent efficacy, sustained combined workloads and production delivery remain
> pending. Retained raw source artifacts consume storage even after restore until
> purge or retention expiry; the CLI now displays that state.

> November 8 digest contexts: `native-go-v257` / `12.0.83-native`
> validates image/catalog digest parameters, keeps RSA and digest aliases separate,
> and rejects duplicate recognized unsigned attributes. The original 450 controls
> expose 61 defined ClamAV differences (52 inappropriate marker exemptions and nine
> false rejections); the current engine corrects all of them on identical inputs.
> Both official/pinned reference builds agree on 612 final verdicts. Both crash on
> eight valid CAB controls, which remain explicitly separate from clean verdicts.
> The Go engine handles those controls without crashing. Another 920 frozen chain
> controls match; all 950 full-feed outcomes remain unchanged on native Linux.
> Full tests, vet, race checks, 375,562 final fuzz executions, a large-OID allocation
> bound, five executable builds and native Linux suites pass. Maximum replay
> cgroup use is 1.259 GiB under hard 2 GiB/no swap, with no OOM events. See the
> [v257 audit](../../docs/edr/native-engine/audits/2026-11-08-v257-digest-contexts/REPORT.md). All nine local test containers are removed; no cloud
> resources or private uploads were created. 3.86 GiB of task-owned temporary files were removed. The strict gate still fails with
> 199 conditional rules. Wider compatibility, Windows/AMSI runtime, compatible-
> kernel prevention, response crash recovery, sustained combined workloads,
> independent efficacy and production delivery remain unqualified.
>

> November 8 primary signed data: `native-go-v256` / `12.0.82-native`
> replaces PE/catalog SignedData verification with owned Go parsing of the original
> signed bytes and cryptographic signer selection. The original 200 configurations
> had 120 ClamAV differences, including 64 inappropriate marker exemptions; all
> are corrected on identical bytes. The final 400 primary/counter configurations,
> 520 frozen chain configurations and 27 static Microsoft-file reference controls
> match ClamAV 1.5.4. Full tests, vet, focused race checks, 115,320 fuzz executions,
> five executable builds and native Linux engine/signature suites pass. All 950
> full-feed outcomes remain unchanged under hard 2 GiB/no swap, with a maximum
> 1.292 GiB cgroup peak and no OOM events. See the [v256 audit](../../docs/edr/native-engine/audits/2026-11-08-v256-primary-signed-data/REPORT.md).
> All six local test containers are removed; no cloud resources or private uploads
> were created. 3.01 GiB of task-owned temporary downloads/builds/cache were removed. The strict gate still fails with 199 conditional rules (eight
> certificate, 110 fuzzy-image and 81 icon-group). Wider certificate/format coverage,
> Windows/AMSI runtime, compatible-kernel prevention, response crash recovery,
> sustained combined workloads, independent efficacy and production delivery
> remain unqualified. The feed CMS dependency is unchanged.
>

> November 8 historical certificate chains: `native-go-v255` / `12.0.81-native`
> verifies bounded historical RSA paths for code, catalogs and legacy timestamps,
> resolves countersigner identifier collisions by signature verification, and
> preserves independently authenticated catalog certificates with their proven
> code/time permissions. An inappropriate timestamp exemption is reproduced and
> fixed. The final 522 synthetic configurations and 27 real signed-file reference
> observations match ClamAV 1.5.4. Full EDR/dependency tests, vet, focused race tests,
> five executable builds and native Linux suites pass; the 950-file full-feed
> replay has zero drift under hard 2 GiB/no swap, with a 1.296 GiB maximum cgroup
> peak and no OOM events. See the [v255 audit](../../docs/edr/native-engine/audits/2026-11-08-v255-legacy-cert-chains/REPORT.md). All twelve local containers
> are removed; no cloud resources or private uploads were created. 3.26 GiB of
> task-owned temporary builds/cache/downloads were removed after verification. The strict gate
> still fails with 199 conditional rules, including eight certificate-trust rules.
> Broader certificate/format semantics, latest Windows/AMSI, compatible-kernel
> prevention, response recovery, sustained combined loads, independent efficacy
> and production delivery remain unqualified.

> November 8 restore integrity: `native-go-v254` / `12.0.80-native`
> rejects corrupted or ineligible stored content and publishes verified snapshots
> without replacing occupied destinations. Pinned descriptors, private staging,
> identity-checked source retirement and OS-backed restore/purge locks protect
> against retained writers and pathname replacement. New relative paths are pinned
> at admission. All EDR/dependency tests, vet, targeted race checks, five executable
> builds, native macOS and root/unprivileged/cross-volume/ownership Linux controls
> pass. The real CLI verifies restore, tamper refusal and no-overwrite behavior.
> See the [v254 audit](../../docs/edr/native-engine/audits/2026-11-08-v254-restore-integrity/REPORT.md). Six disposable local containers were removed;
> no cloud resources or private uploads were created. 2.53 GiB of task-owned
> temporary build/cache files were removed after verification. The new Windows path is
> cross-built, with native runtime qualification still pending. Filesystem/SQLite
> crash recovery and descriptor-bound quarantine admission remain open. Prior
> v253 full-feed evidence is historical; the strict 199-rule gate, wider detection
> coverage, compatible-kernel enforcement, sustained combined workloads,
> independent efficacy and production delivery remain unqualified.

> **2026-11-08 native Linux:** the exact `native-go-v253` / `12.0.79-native`
> Linux ARM64 executable passes 16 selected package suites, 2,338 replay scans,
> two 694-file foreground service sweeps and full-feed changed signed-update/
> failure/restart controls within hard 2 GiB/no swap. Twelve retained envelopes
> and six committed generations verify independently. See the [Linux audit](../../docs/edr/native-engine/audits/2026-11-08-v253-linux-acceptance/REPORT.md).
> Permission prevention remains failed/unqualified because this guest disables
> fanotify access-permission support. The strict 199-rule gate, wider formats/
> predicates, recovery/restore integrity, sustained combined workloads, latest
> Windows/AMSI, independent efficacy and release gates remain open. All 30
> owned containers are removed; no cloud resources were created.

> **2026-11-08 independent JPEG planes:** `native-go-v253` / `12.0.79-native`
> adds unequal/split/non-first sampling and vertical factor four to TIFF JPEG.
> All 114 native-positive TIFF hashes and RGB buffers and 82 raw JPEG buffers
> match independently; 86 TIFF refusals and 40 pipeline checks pass. Full
> tests, vet, race, 383,355 fuzz executions, four builds and the zero-drift
> 950-file replay pass. See the [v253 audit](../../docs/edr/native-engine/audits/2026-11-07-v253-jpeg-planes/REPORT.md).
> Full replacement remains unqualified: 199 conditional rules, fifteen prior
> resize uncertainties, wider sampling/scan semantics and broader native/
> protection/memory/delivery gates. No cloud resources were created.

> **2026-11-07 JPEG-compressed TIFF:** `native-go-v252` / `12.0.78-native`
> adds shared-table/raw-channel TIFF decoding and fixes shared JPEG table/
> ending behavior. The 323 TIFF controls have 240 exact hashes/RGB matches
> and 83 native refusals; 262 component buffers and 34 pipeline checks pass.
> Full tests, vet, race, 135,918 fuzz executions, four builds and the zero-drift
> 950-file replay pass. See the [v252 audit](../../docs/edr/native-engine/audits/2026-11-07-v252-tiff-jpeg/REPORT.md).
> Full replacement remains unqualified: 199 conditional rules, fifteen prior
> resize uncertainties, unqualified layouts and broader native/protection/
> memory/delivery gates. The task-owned Rust oracle is not a runtime dependency.

> **2026-11-07 floating TIFF reader:** `native-go-v251` / `12.0.77-native`
> adds RGB/RGBA32F physical chunk decoding and predictors. All 210 accepted
> controls match native hashes and independent RGB bytes; 97 native refusals,
> 34,057 conversion values and 28 pipeline checks are retained. Full tests,
> vet, race, 34,415 fuzz executions, four builds and the zero-drift 950-file
> replay pass. See the [v251 audit](../../docs/edr/native-engine/audits/2026-11-07-v251-tiff-samples/REPORT.md).
> Full replacement remains unqualified: 199 conditional rules, fifteen prior
> resize uncertainties and broader format/native/protection/release gates.

> **2026-11-07 TIFF CMYK and admission:** `native-go-v250` / `12.0.76-native`
> adds CMYK8/16 f32 conversion with source-preserving IFD views and physical
> bounds. The 159 controls have 146 exact hashes/RGB comparisons and 13
> rejections; 27 pipeline checks include builtin classification. Full tests,
> vet, race, 34,602 fuzz executions, four builds and the zero-drift 950-file
> replay pass. See the [v250 audit](../../docs/edr/native-engine/audits/2026-11-07-v250-tiff-cmyk/REPORT.md).
> Full replacement remains unqualified: 199 conditional rules, fifteen prior
> resize-partial controls and the broader format/native/protection/release
> gates remain open.

> **2026-11-07 owned Go EXR reader:** `native-go-v249` / `12.0.75-native`
> adds the reference's RGB/layer/window and raw/RLE/ZIP/PIZ/PXR24/B44 image
> paths, with bounded metadata and scan work. The new set has 218 exact
> native hashes and 73 rejections across 291 controls; 173 accepted images
> also match independent C++ RGB8 bytes. Exhaustive binary16 conversion and
> 35 engine-pipeline checks pass. Full tests, vet, race, 196,009 final-source
> fuzz executions, four builds and the zero-drift 950-file replay pass.
> See the [v249 audit](../../docs/edr/native-engine/audits/2026-11-07-v249-exr/REPORT.md).
> Image-type overlays are explicit; deep/subsampled and compressed DWA
> inputs remain rejected as in the pinned reference. Full replacement
> remains unqualified: 199 conditional rules, resize uncertainty and the
> broader format/native/efficacy/memory/recovery/delivery gates remain open.

> **2026-11-07 Go HDR/DDS readers and resize safety:** `native-go-v248` /
> `12.0.74-native` adds bounded RGBE/RLE and DXT1/3/5/DX10 image readers.
> There are 178 new controls: 134 exact, 12 resize-partial and 32 rejected.
> A demonstrated shared mixed-axis rounding difference now stays incomplete;
> two prior JPEG controls are partial again. The final native ledger has
> 482 definite and 15 partial matching-rule controls, plus 69 rejections.
> Full tests, vet, race, 228,631 fuzz executions, four builds and the
> zero-drift 950-file replay pass. See the
> [v248 audit](../../docs/edr/native-engine/audits/2026-11-07-v248-hdr-dds/REPORT.md).
> HDR/DDS positives require explicit image typing, as in the native tests.
> Full replacement remains unqualified; 199 conditional rules and the
> image/format/native/efficacy/memory/recovery/delivery gates remain open.

> **2026-11-07 ICO reader and reference f32 transform:** `native-go-v247` /
> `12.0.73-native` reads bounded physical views independently of directory
> lengths and supports the newly qualified PNG, CORE/INFO, palette,
> orientation and RLE layouts. The Go transform follows the reference's
> fixed f32 plan, resolving low-color and selected JPEG ambiguities. There
> are 350 matching native hashes with positive/negative rule controls and
> 37 invalid-reader checks. Full tests, vet, race, 42,507 fuzz executions,
> four builds and the zero-drift 950-file replay pass. See the
> [v247 audit](../../docs/edr/native-engine/audits/2026-11-07-v247-ico-reader/REPORT.md).
> Full replacement remains unqualified; 199 conditional rules and the
> remaining image/format/native/efficacy/memory/delivery gates stay open.

> **2026-11-07 ICO directory and bitmap admission:** `native-go-v246` /
> `12.0.72-native` corrects a confident wrong-image hash on tied entries,
> admits supported selected images despite unused data, and supports wider
> RGB/bitfield headers, advisory fields and optional complete/absent masks.
> There are 201 matching native hashes with positive and different-hash
> rule controls, plus 18 invalid-reader controls and 118 static goldens.
> Cancellation checks now bracket image decoding/hashing. Full tests, vet,
> race, 20,806 fuzz executions and four builds pass; the 950-file replay has
> zero drift. See the
> [v246 audit](../../docs/edr/native-engine/audits/2026-11-07-v246-ico-directory/REPORT.md).
> Full replacement remains unqualified; the strict gate still has 199
> conditional unknowns and the format/release gates below stay open.

> **2026-11-07 CAB admission and local continuation:** `native-go-v245` /
> `12.0.71-native` fixes v244's four exact-child misses. Root physical-input
> bounds and embedded declared-size views now match the pinned reference;
> complete local continuation works across all four CAB codecs, while
> unavailable history and split data remain incomplete. All 144 new native
> comparisons agree (114 positive, 30 negative). Full tests, vet, focused
> race, 212,450 fuzz executions and four builds pass; the 950-file replay has
> zero drift. See the
> [v245 audit](../../docs/edr/native-engine/audits/2026-11-07-v245-cab-admission/REPORT.md).
> Full replacement remains unqualified. The strict feed gate still has 199
> conditional unknowns; remaining format and release gates below stay open.

> **2026-11-07 Go MSZIP core:** `native-go-v244` / `12.0.70-native`
> implements reference-compatible Huffman/repair behavior, CK search and
> dictionary streaming across CAB records. The original repaired-child miss
> is fixed. There are 170 new ClamAV comparisons, including native golden
> hashes for 44 independent zlib controls. Full tests, vet, focused race,
> fuzzing and four builds pass; the 950-file replay has zero drift. Its four
> CAB header/continuation misses are fixed in v245 above. See the
> [v244 audit](../../docs/edr/native-engine/audits/2026-11-07-v244-mszip/REPORT.md).
> Full replacement remains unqualified; the strict feed gate still has 199
> conditional unknowns and the release gates below remain open.

> **2026-11-07 CAB salvage:** `native-go-v243` / `12.0.69-native` fixes
> both v242 CAB extraction misses, preserves validated output after later
> LZX/Quantum errors and bounds retained Quantum allocations. Twenty-eight
> new reference comparisons agree. Full tests, vet, focused race, fuzzing,
> four cross-builds and the 950-file replay pass. That version's MSZIP
> repaired-child miss is fixed in v244 above. See the
> [v243 audit](../../docs/edr/native-engine/audits/2026-11-07-v243-cab-salvage/REPORT.md).
> The strict feed gate still has 199 conditional unknowns. Full replacement
> remains unqualified pending these and the release gates below.

> **2026-11-07 CAB member metadata:** `native-go-v242` /
> `12.0.68-native` fixes the confirmed CAB `.cdb` miss, preserving member
> order and checking metadata before extraction. Thirty-four pinned ClamAV
> comparisons agree. Full tests, vet, focused race, fuzzing and four service
> cross-builds pass; the 950-file replay has zero drift. Its two confirmed
> extraction misses are fixed in v243 above. See the
> [v242 audit](../../docs/edr/native-engine/audits/2026-11-07-v242-cab-metadata/REPORT.md).
> The strict feed gate still has 199 conditional unknowns. Full replacement
> remains unqualified pending these and the release gates below.

> **2026-11-07 Quantum CAB decoder:** `native-go-v241` /
> `12.0.67-native` adds bounded Go Quantum decompression. Twenty positive
> child-hash comparisons match pinned ClamAV, including independent
> multi-frame input and zero-history controls. The GandCrab feed program
> also passes all three ciphers with `.data` and `.rdata` keys. See the
> [v241 audit](../../docs/edr/native-engine/audits/2026-11-07-v241-quantum/REPORT.md).
> That version's CAB member `.cdb` miss is fixed in v242 above.
> The strict feed gate still has 199 conditional unknowns.
> Full replacement remains unqualified pending these and the gates below.

> **2026-11-07 Word 6 macro reader:** `native-go-v240` /
> `12.0.66-native` decodes legacy macro records and scans OLE2 buffers.
> Twenty-nine ClamAV 1.5.4 comparisons agree, including an independent
> upstream negative control. MPRESS and LizardNest official bytecode
> programs also produce exact child detections. Full tests, vet, focused
> race, fuzzing and four cross-builds pass; the 950-file replay has zero
> drift. See the [v240 audit](../../docs/edr/native-engine/audits/2026-11-07-v240-word6-macros/REPORT.md).
> The strict feed audit still reports 199 conditional unknown rules.
> Full replacement remains unqualified pending the gates below.

> **2026-11-06 Linux canary identity:** Distribution `12.0.65-native`
> keeps content engine `native-go-v239`. Linux canary responses now require
> the decoy's admitted file identity; bounded live descriptors prevent
> stale inode reuse, and aliases remain eligible when delivered in scope.
> Full Go tests, vet, race checks and four OS/architecture service and test
> cross-builds pass. Native fanotify execution, periodic reconciliation
> and safe post-incident replant remain open. See the
> [canary identity audit](../../docs/edr/native-engine/audits/2026-11-06-canary-file-identity/REPORT.md).
> The strict feed audit still reports 199 conditional unknown rules.
> Full replacement remains unqualified pending the gates below.

> **2026-11-03 UTC extended bitfield ICOs:** `native-go-v239` /
> `12.0.64-native` hashes bounded 16/32-bit V2/V3 BI_BITFIELDS ICOs.
> Twelve SHA-pinned ClamAV hashes and four custom-rule detections match;
> the 950-file replay has zero drift. Full Go tests, vet, focused race
> and four cross-builds pass. A local full-feed heap profile identifies
> the largest retained allocations but does not qualify Linux service
> memory. The strict official-feed gate still has 199 conditional
> unknowns. See the
> [v239 audit](../../docs/edr/native-engine/audits/2026-11-03-v239-ico-extended-bitfields/REPORT.md).
> Full replacement remains unqualified pending the gates below.

> **2026-11-03 UTC 16-bit ICO DIBs:** `native-go-v238` /
> `12.0.63-native` hashes bounded RGB555 and canonical RGB565/RGB888
> bitfield ICOs. Twelve pinned ClamAV reference hashes match exactly;
> custom graphics-typed ClamAV rules detect. The 950-file replay has
> zero drift; full Go tests, vet, focused race and Windows/Linux AMD64
> and ARM64 cross-builds pass. The strict official-feed gate remains at
> 199 conditional unknowns. See the
> [v238 audit](../../docs/edr/native-engine/audits/2026-11-02-v238-ico-bitfields/REPORT.md).
> Full replacement remains unqualified pending the gates below.

> **2026-11-03 UTC native Windows v207 acceptance:** The exact approved
> bundle ran on Windows Server 2022 x64: 24/24 acceptance groups, 9/9
> real-service checks, three non-elevated standard-user modes and six
> native ETW gates passed. The one-signature lab did not configure a
> production signed feed and does not execute current v237 bytes. All
> dedicated cloud resources were independently verified absent after
> teardown. See the [native Windows audit](../../docs/edr/native-engine/audits/2026-11-02-windows-v207-native/REPORT.md).

> **2026-11-03 UTC extended ICO DIB headers:** `native-go-v237` /
> `12.0.62-native` hashes bounded BI_RGB V2/V3/V4/V5 ICOs with zero or
> canonical masks. Twenty-four pinned ClamAV reference probes agree on
> all certain bits, the 950-file replay has zero drift, full Go tests,
> vet, focused race and four OS/architecture cross-builds pass. The
> strict official-feed gate still has 199 conditional unknowns. See the
> [v237 audit](../../docs/edr/native-engine/audits/2026-11-01-v237-ico-extended-dib/REPORT.md).
> Full replacement remains unqualified pending the gates below.

> **2026-11-01 UTC DIB and multi-frame ICO images:** `native-go-v236` /
> `12.0.61-native` hashes bounded indexed and 24/32-bit DIB ICOs and
> selects among validated PNG/DIB frames using the pinned reference's
> depth, area and directory order. A low-contrast BMP discrepancy now
> carries uncertainty instead of a confident wrong hash. Exact and
> randomized ClamAV probes, the 950-file replay, full Go tests, vet,
> focused race and Windows/Linux AMD64 cross-builds pass. The strict
> official-feed gate still has 199 conditional unknowns. See the
> [v236 audit](../../docs/edr/native-engine/audits/2026-11-01-v236-ico-dib-multiframe/REPORT.md).
> Full replacement remains unqualified pending the gates below.

> **2026-11-01 UTC literal compilation:** `native-go-v235` /
> `12.0.60-native` compiles pure hex bodies directly into the existing
> compact literal matcher. Two alternating full-feed profile pairs saved
> 175.9 MB median cumulative allocation and 38.2 MB median sampled load
> peak. The 950-file replay has zero verdict drift; the strict official-feed
> gate still has 199 conditional unknowns. See the
> [v235 audit](../../docs/edr/native-engine/audits/2026-11-01-v235-literal-compile/REPORT.md).
> Full replacement remains unqualified pending the gates below.

> **2026-11-01 UTC PNG-backed ICO images:** `native-go-v234` /
> `12.0.59-native` hashes bounded single-frame RGBA PNG ICOs; ten exact
> generated probes match pinned ClamAV 1.5.4. Custom graphics-typed
> fuzzy rules detect in both scanners, while unproved ICO layouts remain
> incomplete. The 950-file replay has zero verdict drift; the strict
> official-feed gate still has 199 conditional unknowns. See the
> [v234 audit](../../docs/edr/native-engine/audits/2026-11-01-v234-ico-png/REPORT.md).
> Full replacement remains unqualified pending the gates below.

> **2026-11-01 UTC hash-index memory:** `native-go-v233` /
> `12.0.58-native` reuses the hash compiler's name-offset array as its
> sorted lookup spans. Three full-feed loads saved 12.9 MB median
> cumulative allocation but did not improve sampled peak memory; a
> 950-file replay had zero verdict drift. The strict feed gate still
> has 199 conditional unknowns. See the
> [v233 audit](../../docs/edr/native-engine/audits/2026-11-01-v233-hash-offset-reuse/REPORT.md).
> Full replacement remains unqualified pending the gates below.

> **2026-11-01 UTC phishing URL compatibility:** `native-go-v229` /
> `12.0.54-native` follows ClamAV's bare displayed-host TLD truncation
> for a bounded protected-domain case. A pinned 1.5.4 MIME-email
> differential agrees on 17 GDB and 17 PDB controls, and end-to-end
> Go email tests verify both detection paths. See the
> [v229 audit](../../docs/edr/native-engine/audits/2026-11-01-v229-phishing-bare-tld/REPORT.md).
> Full replacement remains unqualified pending the gates below.

> **2026-11-01 UTC lossy WebP alpha:** `native-go-v228` /
> `12.0.53-native` accepts bounded VP8X/ALPH/VP8 still images after the
> complete alpha container decodes. It hashes extracted VP8 planes with
> the pure-Go fancy upscaler, ignoring transparency as the pinned
> reference does. Forty generated alpha-bearing inputs and forty XMP
> wrappers match ClamAV 1.5.4; a graphics-typed fixture detects in both.
> See the [v228 audit](../../docs/edr/native-engine/audits/2026-11-01-v228-webp-alpha-lossy/REPORT.md).
> Full replacement remains unqualified pending the gates below.

> **2026-11-01 UTC opaque lossy WebP:** `native-go-v227` /
> `12.0.52-native` uses pure-Go chroma upscaling and color conversion
> for bounded VP8 still images, both simple and VP8X with allowed
> metadata. Forty generated base and forty metadata-wrapped lossy hashes
> match pinned ClamAV 1.5.4 across ten sizes and four quality settings.
> Graphics-typed scanner controls detect in both engines. The exact
> profile excludes VP8+ALPH, animation, unsupported chunks and dimensions
> below 16 or above 512. See the
> [v227 audit](../../docs/edr/native-engine/audits/2026-11-01-v227-webp-lossy/REPORT.md).
> Full replacement remains unqualified pending the release gates below.

> **2026-11-01 UTC extended lossless WebP:** `native-go-v226` /
> `12.0.51-native` validates VP8X still-image containers with ICCP,
> VP8L, EXIF and XMP chunks. Sixty generated extended hashes match
> pinned ClamAV 1.5.4, and a graphics-typed control detects in both
> scanners. Malformed, animated, lossy and unknown-chunk layouts remain
> incomplete. Full tests, vet, focused race, Windows cross-builds and
> 950-file signed-feed replay pass with no per-file drift. See the
> [v226 audit](../../docs/edr/native-engine/audits/2026-11-01-v226-webp-extended/REPORT.md).
> Full replacement remains unqualified pending the gates below.

> **2026-11-01 UTC bounded WebP fuzzy images:** `native-go-v225` /
> `12.0.50-native` supports exact `fuzzy_img` decisions for simple
> lossless VP8L files, with or without alpha, typed as graphics by a
> custom `.ftm` feed. Sixty varied hashes and two scanner detections
> match pinned ClamAV 1.5.4.
> Lossy WebP stays incomplete because 11 of 20 raw hashes differed; the
> official feed's RIFF typing still does not run image rules on WebP.
> Full tests, vet, focused race, Windows cross-builds and the 950-file
> signed-feed replay pass with no verdict drift. See the
> [v225 audit](../../docs/edr/native-engine/audits/2026-11-01-v225-webp-typed/REPORT.md).
> Full replacement remains unqualified pending the release gates below.

> **2026-11-01 UTC second official ClamBC output control:** Authenticated
> hta2exe CBC `6754171` now has a harmless PE32 resource that makes both
> Go and pinned ClamAV 1.5.4 detect the same independently derived
> 256-byte child. Broken-resource and disabled-hook controls remain clean.
> See the [hta2exe audit](../../docs/edr/native-engine/audits/2026-11-01-official-hta2exe-output/REPORT.md).
> This is validation of `native-go-v224`, not a new scanner release.

> **2026-11-01 UTC official ClamBC output validation:** The authenticated
> script2exe CBC now has a harmless PE32 resource control that makes both
> Go and pinned ClamAV 1.5.4 detect the same extracted 64-byte child.
> A separate two-child synthetic CBC confirms repeated output scanning.
> This is new validation for `native-go-v224`, not a scanner release.
> See the [official-output audit](../../docs/edr/native-engine/audits/2026-11-01-official-script2exe-output/REPORT.md).
> Other official unpacker paths and the full-replacement gates remain open.

> **2026-11-01 UTC QOI and farbfeld fuzzy images:** `native-go-v224` /
> `12.0.49-native` adds bounded decoders for QOI chunk streams and farbfeld
> RGBA16 rasters. Thirty-six varied hashes match SHA-pinned ClamAV 1.5.4;
> three graphics-typed scanner controls detect in both implementations.
> Malformed or oversized inputs remain incomplete. See the
> [v224 audit](../../docs/edr/native-engine/audits/2026-11-01-v224-qoi-farbfeld/REPORT.md).
> Full replacement remains unqualified pending the release gates below.

> **2026-11-01 UTC PAM fuzzy images:** `native-go-v223` /
> `12.0.48-native` adds P7/PAM visual tuple and alpha-variant decoding.
> Nine graphics-typed rule controls detect in Go and pinned ClamAV 1.5.4;
> 336 varied hashes match exactly, including the reference's untyped
> 16-bit path. Full tests, vet, focused race, Windows cross-builds and
> 950-file signed-feed replay pass with zero per-file drift from v222.
> See the [v223 audit](../../docs/edr/native-engine/audits/2026-11-01-v223-pam-fuzzy/REPORT.md).
> Full replacement remains unqualified on the feed, bytecode, Windows,
> efficacy and ThreatWind gates.

> **2026-11-01 UTC Netpbm sample depths:** `native-go-v222` /
> `12.0.47-native` scales P2/P3/P5/P6 declared maxima 1–65,535 and reads
> 16-bit binary samples. All 224 varied max-value/size hashes match pinned
> ClamAV 1.5.4; two new typed scanner detections agree. Full tests, vet,
> focused race, Windows cross-builds and 950-file signed-feed replay pass
> with zero per-file drift from v221. See the
> [v222 audit](../../docs/edr/native-engine/audits/2026-11-01-v222-netpbm-depth/REPORT.md).
> Full replacement remains unqualified on the official-feed, bytecode,
> Windows, efficacy and ThreatWind gates.

> **2026-11-01 UTC Netpbm fuzzy images:** `native-go-v221` /
> `12.0.46-native` adds P1–P6 PBM/PGM/PPM decoding for reference-tested
> 1-bit and 8-bit profiles. Six graphics-typed rule controls detect in Go
> and pinned ClamAV 1.5.4; 54 varied image hashes match exactly. Full tests,
> vet, focused race, Windows cross-builds and 950-file signed-feed replay
> pass with zero per-file drift from v220. See the
> [v221 audit](../../docs/edr/native-engine/audits/2026-11-01-v221-netpbm-fuzzy/REPORT.md).
> Other PNM sample depths and the full-replacement release gates remain open.

> **2026-11-01 UTC partition scan windows:** `native-go-v220` /
> `12.0.45-native` enables embedded file-type analysis in MBR boot sectors
> and MBR/GPT partitions. Three harmless nested ZIP controls detect in Go
> and pinned ClamAV 1.5.4; the two partition controls were false clean
> before the fix. The full agent suite, vet, Windows cross-builds and
> 950-file signed-feed replay pass with zero per-file drift from v219. See
> the [v220 audit](../../docs/edr/native-engine/audits/2026-11-01-v220-partition-windows/REPORT.md).
> Full replacement still requires the remaining official-feed, bytecode,
> Windows, efficacy and deployed ThreatWind acceptance gates.

> **2026-11-01 UTC nested scan windows:** `native-go-v219` /
> `12.0.44-native` scans embedded file-type records in CPIO members and
> universal Mach-O slices. The old-CPIO compressed-child fixture was a
> confirmed false clean before the fix; Go and pinned ClamAV 1.5.4 now
> detect the same inert marker. Full tests, vet, Windows cross-builds and
> 950-file signed-feed replay pass with zero per-file verdict drift. See the
> [v219 audit](../../docs/edr/native-engine/audits/2026-11-01-v219-nested-windows/REPORT.md).
> Full replacement still needs the remaining feed, bytecode, Windows,
> efficacy and deployed ThreatWind acceptance gates.

> **2026-11-01 UTC expanded OneNote differential:** `native-go-v218` /
> `12.0.43-native` matches 218/218 ClamAV-positive and 0/182 ClamAV-clean
> expanded web mutations. The optional 400-case test passes after bounded
> BinaryItem framing and object-data BLOB reference corrections. The original
> 100 mutations still match 51/51 positive, 0/49 clean. Full tests, vet, focused
> race, Windows AMD64/ARM64 cross-builds and 950-file signed-feed replay pass
> with zero verdict drift from v217. A second whole-file 1,000-mutation
> sample matches 846/846 positive and 0/154 clean with typed metadata and
> declaration-count handling. See the
> [v218 audit](../../docs/edr/native-engine/audits/2026-11-01-v218-onenote-differential/REPORT.md).
> Full replacement still requires the remaining feed, bytecode, Windows,
> efficacy and deployed ThreatWind acceptance gates.

> **2026-11-01 UTC OneNote web-export detection:** `native-go-v217` /
> `12.0.42-native` follows active OneStore object spaces and page/title/body
> references to embedded-file containers before scanning their BLOBs. Go and
> pinned ClamAV 1.5.4 detect the same child in all three official OneNote
> fixtures. The pinned 100-mutation web differential now has 51/51 reference
> detections and 0/49 reference-clean detections at the final scanner level.
> Full tests, vet, focused race, Windows AMD64/ARM64 cross-builds and the
> 950-file signed-feed replay pass with zero per-file drift from v215. See the
> [v217 audit](../../docs/edr/native-engine/audits/2026-11-01-v217-onenote-web/REPORT.md).
> A later 400-mutation differential found 12 verdict mismatches (nine misses,
> two extra HDB detections and one extra broken-PE heuristic detection). Its
> pinned failing output is retained as v217 history; v218 resolves it.
> Full replacement is still unqualified pending the listed feed, bytecode,
> native Windows runtime, malware-efficacy and ThreatWind deployment gates.

> **2026-11-01 UTC OneNote graph validation:** `native-go-v216` /
> `12.0.41-native` parses FSSHTTPB elements and OneStore storage, cell,
> revision, object-group and BLOB links, then validates bounded object
> property streams. A 100-mutation pinned ClamAV differential retains all
> 51 reference positives at this prerequisite stage and rejects 39 of 49
> reference-clean packages. Ten clean packages remain structurally valid;
> all web exports remain incomplete until active page and embedded-file
> references prove the attachment. See the [v216 audit](../../docs/edr/native-engine/audits/2026-11-01-v216-onenote-graph/REPORT.md).
> Full replacement is still unqualified.

> **2026-11-01 UTC OneNote web-export safety correction:** `native-go-v215`
> with distribution `12.0.40-native` makes FSSHTTPB blob extraction
> incomplete until the OneStore revision graph proves the child is live.
> A 100-mutation pinned ClamAV differential found 39 v214 false detections;
> v215 has none on those mutations, while 51 reference detections remain
> incomplete. The original web-app fixture is also incomplete; desktop
> 2007/2010 embedded-child detection remains. See the
> [v215 audit](../../docs/edr/native-engine/audits/2026-11-01-v215-onenote-graph-guard/REPORT.md).
> Full agent checks and Windows cross-builds pass, and the 950-file
> signed-feed replay has zero per-file drift. Full ClamAV replacement is
> not qualified; graph validation and the other release gates remain open.

> **2026-11-01 UTC OneNote web exports:** `native-go-v214` scans validated
> FSSHTTPB Object Data BLOB children within a bounded OneStore package.
> Go and pinned ClamAV 1.5.4 now detect the same child in all three public
> OneNote fixtures, including the web-app export. A detached in-package
> blob also detects in both; a copy after the package terminator does not.
> Distribution `12.0.39-native` refreshes the core. See the
> [v214 audit](../../docs/edr/native-engine/audits/2026-11-01-v214-onenote-fsshttpb/REPORT.md).
> Full agent checks and Windows AMD64/ARM64 cross-builds pass; the
> hash-verified 950-file safe replay has zero per-file drift from v213.
> Full graph-aware OneNote parsing and the other replacement gates remain
> open.

> **2026-11-01 UTC OneNote attachments:** `native-go-v213` recognizes
> MS-ONESTORE `.one` files even with hash-only feeds and scans bounded
> `FileDataStoreObject` payloads as recursive children. Pinned ClamAV 1.5.4
> and Go both detect the 544-byte child in the public 2007 and 2010 test
> sections. The web-app export uses a different storage path: ClamAV detects
> the child; Go reports incomplete. Distribution `12.0.38-native` refreshes
> the core. See the
> [v213 audit](../../docs/edr/native-engine/audits/2026-11-01-v213-onenote-attachments/REPORT.md).
> Full agent tests, vet, targeted race, Windows AMD64/ARM64 cross-builds and
> the hash-verified 950-file safe replay passed with zero per-file drift.
> Native Windows runtime, full OneNote graph/FSSHTTPB parsing and the other
> full-replacement release gates remain open.

> **2026-11-01 UTC multiple PDF literal strings:** `native-go-v212` follows
> the reference's first-dictionary flag rule across successive flat literal
> strings in one indirect object. Six serialized-hook decisions match pinned
> ClamAV 1.5.4. Distribution `12.0.37-native` refreshes the core. A separate
> 144-image pinned GIF differential confirms the conservative >512-pixel
> hash uncertainty guard is still needed. See the
> [v212 audit](../../docs/edr/native-engine/audits/2026-11-01-v212-pdf-multiliteral/REPORT.md).
> Full agent checks and Windows AMD64 cross-build pass; a 950-file verified
> safe/assumed-benign replay has zero per-file verdict drift from v211.
> Full replacement remains unqualified pending the listed release gates.

> **2026-11-01 UTC leading-string PDF dictionaries:** `native-go-v211` models
> the reference's first-dictionary flags when an indirect object has a
> literal string before an ordinary dictionary. Six harmless serialized-hook
> decisions agree with ClamAV 1.5.4. Distribution `12.0.36-native` fetches
> the changed core. Full checks and Windows AMD64 cross-build pass; 950
> newer-feed safe controls have zero per-file drift from v210. See the
> [v211 audit](../../docs/edr/native-engine/audits/2026-11-01-v211-pdf-leading-string/REPORT.md).
> Full replacement remains unqualified pending the listed release gates.

> **2026-11-01 UTC official unpacker trigger controls:** Three authenticated,
> SHA-pinned ClamBC PE unpackers now have harmless Go/ClamAV hook-trigger
> tests. All three execute and return clean in pinned ClamAV 1.5.4; the Go
> pipeline also returns clean. Separate synthetic CBCs positively validate
> `write` and `extract_new` child-HDB detection in both engines. The official
> unpacker controls do not validate child extraction.
> The attempted synthetic script2exe resource never caused the reference to
> write a child. See the [unpacker-control audit](../../docs/edr/native-engine/audits/2026-11-01-official-bytecode-unpacker-controls/REPORT.md).

> **2026-11-01 UTC complex PDF flags:** `native-go-v210` models the reference's
> slash-name walk through bounded dictionary strings, comments, arrays and
> nested dictionaries and string objects. Seven new harmless hooks and three exact flag values
> agree with ClamAV 1.5.4. Distribution `12.0.35-native` expires caches.
> Full agent checks and Windows AMD64 cross-build pass; 950 newer-feed safe
> controls have zero per-file drift from v209. The strict feed still has 199
> conditional unknowns. See the [v210 audit](../../docs/edr/native-engine/audits/2026-11-01-v210-pdf-complex-flags/REPORT.md).
> Full replacement remains unqualified pending the listed release gates.


> **2026-11-01 UTC hidden PDF object headers:** `native-go-v209` models the
> reference's raw `obj`/`endobj` discovery order for bytecode object APIs.
> Eighteen harmless native/Go decisions agree across comment, string and
> stream controls for count, lookup, ID, size, offset and simple flags.
> Complex flags remain conditional. Distribution `12.0.34-native`
> fetches the changed core. Full checks and Windows AMD64 cross-build
> pass; both controlled and newer-feed 950-file replays have zero drift.
> The strict feed still has 199 conditional unknown rules. See the
> [v209 audit](../../docs/edr/native-engine/audits/2026-11-01-v209-pdf-hidden-objects/REPORT.md).
> Native Windows v209 and independent malware efficacy remain open.

> **2026-11-01 UTC PDF object-stream flags:** `native-go-v208` answers
> `pdf_getobjflags` for proved outer objects during PARSED. Plain and
> Flate one-member `/ObjStm` positive hooks match ClamAV 1.5.4; an inner
> `/Page` name does not cause an outer flag detection. The parser now
> owns reached PDF hook phases for recognized PDFs, so this negative
> control finishes clean in both engines. A PRE hook also detects before
> an unterminated object. Complex later phases remain conditional.
> Distribution `12.0.33-native` fetches the changed core.
> See the [v208 audit](../../docs/edr/native-engine/audits/2026-11-01-v208-pdf-objstm-flags/REPORT.md).
> Full agent checks and Windows AMD64 cross-build pass. Both controlled
> and newer-feed 950-file replays have zero per-file drift from v207;
> the strict signed-feed gate retains 199 conditional unknowns.
> The prepared Windows bundle below is v207, so native v208 runtime is
> still untested.

> **2026-11-01 UTC Windows v207 preflight:** A fresh 18-file Windows AMD64
> acceptance bundle was built and hash-verified from the current scanner.
> See the [preflight](../../docs/edr/native-engine/audits/2026-11-01-windows-v207-preflight/REPORT.md).
> The GCP token needs interactive reauthentication and the earlier lab was
> removed; this is not a native Windows v207 pass.

> **2026-11-01 UTC Authenticode digest parity:** `native-go-v207` checks
> that the PKCS#7 `SignedData` and `SignerInfo` digest algorithms agree by
> effective hash and that optional parameters are empty NULLs. Twenty-seven
> signed-PE outcomes across three pinned official Microsoft binaries agree
> with ClamAV 1.5.4, including the valid SHA-256 RSA OID alias. Distribution
> `12.0.32-native` fetches the changed core. See the [v207 audit](../../docs/edr/native-engine/audits/2026-11-01-v207-authenticode-digest/REPORT.md).
> Full agent checks and Linux/Windows cross-builds pass. The strict feed
> retains 199 conditional unknowns; a 950-file hash-verified safe replay
> has zero per-file drift from v206. Native Windows and independent malware
> efficacy remain unqualified.

> **2026-11-01 UTC Authenticode trust parity:** `native-go-v206` rejects a
> PKCS#7 `SignerInfo` whose version is not 1 before granting a CRB trust
> exemption. Fifteen original and mutated signed-PE decisions across three
> pinned Microsoft Sigcheck binaries agree with ClamAV 1.5.4. Distribution
> `12.0.31-native` fetches the changed core. See the [v206 audit](../../docs/edr/native-engine/audits/2026-11-01-v206-authenticode-signer-version/REPORT.md).
> Full agent checks and Linux/Windows cross-builds pass. The strict feed
> still has 199 conditional unknowns, and a 950-file hash-verified safe
> replay has zero per-file drift from v205. Native Windows and independent
> malware efficacy remain unqualified.

> **2026-11-01 UTC PDF object streams:** `native-go-v205` now exposes a
> proved outer object table to PARSED-phase bytecode hooks on PDFs with
> object streams. Twenty harmless native/Go hook decisions match across
> plain and Flate cases. Object-stream flags and hidden-header discovery
> remain undecided. Distribution `12.0.30-native` fetches the changed core.
> See the [v205 audit](../../docs/edr/native-engine/audits/2026-11-01-v205-pdf-objstm/REPORT.md).
> Full agent checks pass. A fresh signed-feed audit retains 199 conditional
> unknown rules, and a 950-file hash-verified replay has zero per-file drift
> from v203.


> **2026-11-01 UTC ClamBC stream tails:** `native-go-v204` preserves the
> source pipe's bytes after a completed bzip2 or LZMA stream. Harmless
> serialized bytecode tests match ClamAV 1.5.4 for trailing-byte detection
> and a no-tail clean control. Distribution `12.0.29-native` fetches the
> changed core. The [v204 audit](../../docs/edr/native-engine/audits/2026-11-01-v204-bytecode-stream-tail/REPORT.md)
> inventories the current authenticated bytecode feed and remaining gates.
> The same audit adds a hash-pinned nine-packer PE control differential:
> pinned ClamAV and Go both find the unpacked child in every sample.


> **2026-11-01 UTC larger GIF fuzzy hashes:** `native-go-v203` extends
> the tested exact-resampling profile to opaque and transparent GIFs up to
> 512 pixels per axis. All 112 generated native/Go hash probes agree, and a
> positive/negative fuzzy rule pair agrees with ClamAV 1.5.4. DCT-boundary
> uncertainty remains. Distribution `12.0.28-native` fetches the core.
> See the [v203 audit](../../docs/edr/native-engine/audits/2026-11-01-v203-large-gif/REPORT.md).
> Full agent checks pass, and a current-feed replay has zero per-file drift
> among 950 retained safe and benign inputs. The strict gate remains at 199
> conditional unknown rules.

> **2026-11-01 UTC PE icon header boundary:** `native-go-v202` follows
> ClamAV 1.5.4's definite skip for a non-wrapping bitmap header extending
> beyond the mapped file. A wrapping offset remains undecided. Native and
> Go agree on harmless malformed/valid controls; the strict signed-feed gate
> remains at 199 conditional unknown rules. Distribution
> `12.0.27-native` fetches the core. See the
> [v202 audit](../../docs/edr/native-engine/audits/2026-11-01-v202-icon-header/REPORT.md).
> A fresh 694-file current-feed replay and 256 benign-control scan show
> zero per-file drift from v178 and no missed retained native detections.

> **2026-11-01 UTC ClamBC API 101:** `native-go-v201` accepts the
> `engine_scan_options_ex` declaration and reproduces the pinned 1.5.4
> interpreter's observed zero result for documented option names. Native
> harmless fixtures and Go agree with `--gen-json` off/on. The upstream
> name-copy defect means this does not establish intended named-option
> semantics. Distribution `12.0.26-native` fetches the core. See the
> [v201 audit](../../docs/edr/native-engine/audits/2026-11-01-v201-clambc-options-ex/REPORT.md).

> **2026-11-01 UTC metadata activation:** `native-go-v200` adds default-off
> `collect_metadata`. Root PRECLASS and legacy target-13 now run only when
> enabled, matching native ClamAV 1.5.4 default/`--gen-json` behavior.
> Bytecode API 68 reflects the option and JSON API 92 sees active metadata
> in parser callbacks. Unknown metadata fields remain undecided.
> Distribution `12.0.25-native` fetches the core. See the
> [v200 audit](../../docs/edr/native-engine/audits/2026-11-01-v200-metadata-mode/REPORT.md).

> **2026-11-01 UTC document metadata:** `native-go-v199` adds a
> reference-compared HWPML property tree and makes it and the proved Office
> XML properties visible to root PRECLASS hooks. The tree remains partial;
> unmodeled fields are undecided. Distribution `12.0.24-native` fetches the
> core. See the
> [v199 audit](../../docs/edr/native-engine/audits/2026-11-01-v199-document-metadata/REPORT.md).

> **2026-11-01 UTC XML depth:** `native-go-v198` matches the shared ClamAV
> XML reader's level-20 subtree skip for HWPML and Office XML children and
> explicit Office metadata. Native/Go harmless level-19/20 fixtures agree;
> skipped subtrees remain cancellation-aware. Distribution
> `12.0.23-native` fetches this core. See the
> [v198 audit](../../docs/edr/native-engine/audits/2026-11-01-v198-xml-depth/REPORT.md).

> **2026-11-01 UTC HWPML:** `native-go-v197` scans HWPML binary and script
> text children, including Base64 and raw-DEFLATE content. The Go path
> detects decoded children that ClamAV 1.5.4 misses because its temporary
> file is not rewound in the tested case. Undetected HWPML remains incomplete
> pending full parser/metadata parity. Distribution `12.0.22-native` fetches
> the core. See the [v197 audit](../../docs/edr/native-engine/audits/2026-11-01-v197-hwpml/REPORT.md).

> **2026-11-01 UTC CryptFF:** `native-go-v196` decodes CryptFF bodies and
> scans them as children, with harmless direct and nested detections matching
> ClamAV 1.5.4. Distribution `12.0.21-native` fetches the core. See the
> [v196 audit](../../docs/edr/native-engine/audits/2026-11-01-v196-cryptff/REPORT.md).

> **2026-11-01 UTC Apple Partition Map:** `native-go-v195` scans APM
> partitions, legacy driver extents and optional overlap alerts. Harmless
> detecting/clean fixtures and the overlap heuristic agree with ClamAV 1.5.4.
> Distribution `12.0.20-native` fetches the core. See the
> [v195 audit](../../docs/edr/native-engine/audits/2026-11-01-v195-apm/REPORT.md).

> **2026-11-01 UTC ClamBC LZMA:** `native-go-v194` adds wire APIs
> 102–104 for complete LZMA-alone streams in one pair of pipe windows.
> Four harmless known-size/EOS native/Go differential cases agree; partial
> windows remain incomplete. Distribution `12.0.19-native` fetches the core.
> See the [v194 audit](../../docs/edr/native-engine/audits/2026-11-01-v194-clambc-lzma/REPORT.md).

> **2026-11-01 UTC ClamBC bzip2:** `native-go-v193` adds wire APIs
> 105–107 for complete streams inside one pair of pipe windows. A harmless
> decoded-child oracle agrees with ClamAV 1.5.4 on detection, clean control,
> stream-end status and consumed input. Partial-window paths remain
> incomplete. Distribution `12.0.18-native` fetches the core. See the
> [v193 audit](../../docs/edr/native-engine/audits/2026-11-01-v193-clambc-bzip2/REPORT.md).

> **2026-11-01 UTC ClamBC file reliability:** `native-go-v192` implements
> wire API 91. Root scans supply a proved normal-input value; unproved
> nested/reconstructed values remain incomplete. Four harmless native/Go
> differential cases pass. Distribution `12.0.17-native` fetches the core.
> See the [v192 audit](../../docs/edr/native-engine/audits/2026-11-01-v192-file-reliability/REPORT.md).

> **2026-11-01 UTC ClamBC extracted-container typing:** `native-go-v191`
> implements wire API 70 and passes a forced file type through bytecode output
> into the recursive child scanner. Harmless `CL_TYPE_OTHER` target-14 child
> scans agree with the pinned ClamAV 1.5.4 interpreter; unmodeled type/parser
> paths remain incomplete. Distribution `12.0.16-native` fetches the core.
> See the [v191 audit](../../docs/edr/native-engine/audits/2026-11-01-v191-container-type/REPORT.md).

> **2026-11-01 UTC ClamBC input switching:** `native-go-v190` implements
> wire API 71 with output snapshots, file-size updates, file-backed pipe
> redirection and extracted-pointer retirement. Harmless Go and pinned
> ClamAV 1.5.4 scans agree on a detection and clean control. Distribution
> `12.0.15-native` fetches the core. Mapped-output truncation remains an
> incomplete path. See the [v190 audit](../../docs/edr/native-engine/audits/2026-11-01-v190-input-switch/REPORT.md).

> **2026-11-01 UTC ClamBC startup controls:** `native-go-v189` implements
> wire APIs 73 and 74 and applies full bytecode disablement to one feed
> generation, leaving static signatures active. Six harmless native ClamAV
> 1.5.4 scans agree. Distribution `12.0.14-native` fetches the core. The
> [v189 audit](../../docs/edr/native-engine/audits/2026-11-01-v189-startup-control/REPORT.md)
> records validation and open release gates.

> **2026-10-31 ClamBC platform queries:** `native-go-v188` implements wire
> APIs 75, 76 and 90 for version, platform and JIT checks. Five harmless
> unsigned ClamBC fixtures match the retained native ClamAV 1.5.4 macOS
> interpreter. Distribution `12.0.13-native` fetches the core. The
> [v188 audit](../../docs/edr/native-engine/audits/2026-10-31-v188-clambc-platform/REPORT.md)
> retains 199 conditional feed unknowns; native Windows and independent
> malware efficacy gates remain open.

> **2026-10-31 ClamBC JavaScript normalization:** `native-go-v187` adds
> wire APIs 43–45 over bounded pipes. The shared tokenizer preserves state
> across process calls, and completed normalized output enters recursive
> child scanning after VM execution. Unmodeled transforms remain incomplete.
> Distribution `12.0.12-native` fetches the changed core. Native ClamBC
> differential and release gates remain open; see the
> [v187 audit](../../docs/edr/native-engine/audits/2026-10-31-v187-clambc-jsnorm/REPORT.md).

> **2026-10-31 ClamBC inflate recovery:** `native-go-v186` implements
> full-flush-marker `inflateSync` recovery and aligns per-call cursor,
> status and output behavior with a test-only C zlib adapter over valid
> and corrupt streams. Preset-dictionary streams return `Z_NEED_DICT` with
> matched cursors. Distribution `12.0.11-native` fetches this core.
> Full bytecode/feed/native acceptance remains open.
> See the [v186 audit](../../docs/edr/native-engine/audits/2026-10-31-v186-clambc-inflate-sync/REPORT.md).

> **2026-10-31 ClamBC stateful inflate:** `native-go-v185` supports wire
> APIs 39–41 over bounded pipes, with incremental raw/zlib/gzip decoding,
> checksums, optional gzip headers and recursive child scanning of emitted
> bytes. Trace APIs 9–14 use the reference's disabled profile with pointer
> checks. Distribution `12.0.10-native` fetches the changed core. `inflateSync`
> recovery, preset dictionaries and native differential validation remain
> open. See the [v185 audit](../../docs/edr/native-engine/audits/2026-10-31-v185-clambc-inflate/REPORT.md).

> **2026-10-31 ClamBC maps:** `native-go-v184` implements map APIs 57–64,
> with bounded fixed/dynamic value storage and VM pointer lifetime checks.
> It also rejects zero-size pipe creation as upstream does. Distribution
> `12.0.9-native` fetches the changed core. The
> [v184 audit](../../docs/edr/native-engine/audits/2026-10-31-v184-bytecode-maps/REPORT.md)
> records the remaining release gaps.

> **2026-10-31 ClamBC buffer pipes:** `native-go-v183` implements fixed and
> file-backed pipe APIs 30–38 with per-run bounds and pointer lifetime checks.
> The distribution advances to `12.0.8-native`. The
> [v183 audit](../../docs/edr/native-engine/audits/2026-10-31-v183-buffer-pipe/REPORT.md)
> retains 199 conditional unknowns. Pipe-consuming decompression hooks and
> native Windows acceptance still block full replacement.

> **2026-10-31 ClamBC input-window API:** `native-go-v182` executes bounded
> `fill_buffer` calls, including the reference's EOF order and ignored `fill`
> argument. The agent distribution advances to `12.0.7-native`. The
> [v182 audit](../../docs/edr/native-engine/audits/2026-10-31-v182-fill-buffer/REPORT.md)
> retains 199 conditional unknowns; full replacement remains unqualified.

> **2026-10-31 rollout identity:** The agent and scanner now share an exact
> `native-go-v181` build identity, and EDR distribution version
> `12.0.6-native` triggers a tracked update. A same-version tracked scanner
> is revalidated. This is not signed publisher verification or paired AMSI
> DLL delivery; see the [rollout audit](../../docs/edr/native-engine/audits/2026-10-31-v181-rollout-identity/REPORT.md).

> **2026-10-31 container-name ERE compatibility:** `native-go-v181`
> translates ClamAV's bundled `.cdb` escape rules, including the current
> signed feed's literal-backslash bracket class. Future directional word
> assertions load as conditional rather than failing a generation or
> silently deciding clean. The [v181 audit](../../docs/edr/native-engine/audits/2026-10-31-v181-container-regex/REPORT.md)
> still has 199 conditional unknowns, so full replacement remains unqualified.

> **2026-10-31 ClamBC APIs:** `native-go-v180` accepts and executes
> hash-set wire APIs 24–29 with private, bounded per-run state, plus hex
> conversion and buffer entropy. The [v180 audit](../../docs/edr/native-engine/audits/2026-10-31-v180-bytecode-hashset/REPORT.md)
> retains 199 conditional unknowns. Full replacement remains unqualified.

> **2026-10-31 traditional ZIP decryption:** `native-go-v179` loads `.pwdb`
> and decrypts PKWARE-encrypted ZIP members with applicable passwords. Nine
> harmless pinned ClamAV 1.5.4 oracle cases cover stored/deflated data,
> cleartext/hex passwords, version-20/21 headers and wrong/absent/disabled
> passwords. Go keeps unreadable encrypted content incomplete. The current
> official feed has no `.pwdb`, and its 199 conditional unknown rules remain.
> The verified 694-file safe and 256-control current-feed replay has zero
> per-file verdict or attribution changes from v178.
> See the [v179 audit](../../docs/edr/native-engine/audits/2026-10-31-v179-zip-passwords/REPORT.md).
> Full replacement remains unqualified.

> **2026-10-31 matcher order:** `native-go-v178` runs static body rules,
> whole-file hashes, then logical bytecode callbacks. Pinned ClamAV 1.5.4
> agrees on local EICAR attribution. The 694-file safe replay retained all
> verdicts and closed the 81 shared detection-name differences; 256 benign
> controls were unchanged. The strict feed gate still has 199 conditional
> unknowns, and native Windows plus malware efficacy acceptance remains open.
> See the [v178 audit](../../docs/edr/native-engine/audits/2026-10-31-v178-matcher-order/REPORT.md).
> Full replacement remains unqualified.

> **2026-10-31 current-feed replay:** The `native-go-v177` macOS scanner
> reproduced all 694 safe and 256 assumed-benign per-file `v169` verdicts.
> Pinned ClamAV 1.5.4 detected 259 of the 694; Go detected all 259, plus
> 37 synthetic positives, and marked 22 ClamAV-clean files incomplete.
> EICAR signature attribution differs on 81 shared detections. A fresh
> strict audit still fails on 199 conditional unknowns. The replay does not
> prove production malware efficacy or close the remaining release gates.
> See the
> [v177 replay audit](../../docs/edr/native-engine/audits/2026-10-31-v177-safe-replay/REPORT.md).
> Full replacement remains unqualified.

> **2026-10-31 icon-group languages:** `native-go-v177` scans all languages
> of the first PE group-icon ID, stops at a different ID, shares the
> 100-entry budget and sign-extends icon IDs (`0xffff` selects any name).
> Pinned ClamAV 1.5.4 and Go agree on two detections and two clean cases.
> The full EDR suite, vet, and Linux/Windows AMD64 builds passed. The last
> strict feed audit still has 199 conditional unknowns; the remaining
> release gates remain open. See the
> [v177 audit](../../docs/edr/native-engine/audits/2026-10-31-v177-icon-group-languages/REPORT.md).
> Full replacement remains unqualified.

> **2026-10-31 PE icon-resource traversal:** `native-go-v176` skips
> non-directory type and icon-name entries and nested language entries, and
> scans duplicate icon IDs in resource order. Pinned ClamAV 1.5.4 and Go
> agree on three clean cases and a second-duplicate detection. The full
> EDR suite, vet, and
> Linux/Windows AMD64 builds passed. The strict signed-feed gate still has
> 199 conditional unknowns, and native Windows and malware-corpus acceptance
> remain open. See the [v176 audit](../../docs/edr/native-engine/audits/2026-10-31-v176-icon-resource-walk/REPORT.md).
> Full replacement remains unqualified.

> **2026-10-31 bytecode environment:** `native-go-v175` aligns API-72's
> platform IDs and host strings with the pinned ClamAV 1.5.4 macOS ARM64
> reference. Linux and Windows compiler/build fields still need native
> validation; the 199 conditional feed rules and release gates remain open.
> See the [v175 audit](../../docs/edr/native-engine/audits/2026-10-31-v175-bytecode-environment/REPORT.md).
> Full replacement remains unqualified.

> **2026-10-31 pinned ClamAV 1.5.4 level:** `native-go-v174` uses functionality
> and dconf level 234 for feed selection and bytecode queries. A harmless
> native differential proves 234 selected and 235 skipped; Go matches. The
> retained authenticated current feed selects the same 3,628,730 rules as
> before, rejects zero, and still has 199 conditional unknowns. The strict
> gate fails. EDR package tests, vet and a Windows AMD64 cross-build passed.
> See the [v174 audit](../../docs/edr/native-engine/audits/2026-10-31-v174-level234/REPORT.md).
> Full replacement remains unqualified.

> **2026-10-31 plain-text root metadata:** `native-go-v173` admits complete
> scalar JSON for childless `CL_TYPE_TEXT_ASCII` roots into the active
> target-13 pass and aligns backslash filename separators and raw Unicode
> path escapes. Pinned ClamAV 1.5.4 detects a harmless target-13 rule on
> ASCII and short UTF-8 text when JSON collection is enabled, and is clean
> with its default metadata-off policy. The Go regression matches; affected
> tests, vet and Windows AMD64 cross-build passed. Parser-bearing metadata
> trees and the remaining release gates are still open. See the
> [v173 audit](../../docs/edr/native-engine/audits/2026-10-31-v173-text-metadata/REPORT.md).
> Full replacement remains unqualified.

> **2026-10-31 path-aware metadata and cache:** `native-go-v172` supplies
> `FileName`/`FilePath` to active preclassification and childless binary-root
> target-13 analysis. A native ClamAV 1.5.4 filename-rule oracle and Go tests
> distinguish identical bytes at two names. The production scanner bypasses
> hash-only verdict-cache reuse while metadata can be active, and fanotify
> carries the pinned event path. The remaining metadata tree and native
> release gates remain open. See the [v172 audit](../../docs/edr/native-engine/audits/2026-10-31-v172-path-metadata/REPORT.md).
> Full replacement remains unqualified.

> **2026-10-31 native level-234 and fallback differential:** The retained
> authenticated current feed selects the same 3,628,730 rules at reference
> levels 216 and 234, with zero rejected and 199 conditional unknowns. The
> portable ClamAV 1.5.4 binary agrees with Go on six no-`.ftm` HTML/Python
> cases; native `--gen-json` confirms target-13 fallback INTERNAL detection.
> Its file JSON includes `FileName` and `FilePath`, which the Go byte-only
> scan and content-hash verdict cache cannot yet handle safely. See the
> [v171 audit](../../docs/edr/native-engine/audits/2026-10-31-v171-fallback-level/REPORT.md).
> Full replacement remains unqualified.

> **2026-10-31 fallback types and audit level:** `native-go-v171` aligns the
> no-`.ftm` HTML markers, built-in `CLAMJSONv0` INTERNAL magic and a modern
> compiled-Python header pattern with the pinned ClamAV 1.5.4 table.
> Local-only target-13 and compiled-Python target-14 detections now work
> through the fallback. `sigaudit` defaults to reference level 234,
> while the runtime remains at 216; level-217–234 rules therefore need a
> fresh current-feed and runtime reconciliation. Tests, vet and Windows AMD64
> cross-build passed. See the [v171 audit](../../docs/edr/native-engine/audits/2026-10-31-v171-fallback-level/REPORT.md).
> Full replacement remains unqualified.

> **2026-10-31 type dispatch and legacy metadata:** `native-go-v170` adds
> `CL_TYPE_AI_MODEL` at its pinned 1.5.4 numeric position, restoring every
> subsequent type value, and applies target-14 `OTHER` rules to binary data,
> compiled Python and AI models. Active legacy target-13 rules now inspect
> a complete childless binary root's json-c-compatible metadata bytes;
> other partial trees are explicitly incomplete. Synthetic regression tests
> cover these passes and refresh. The full metadata graph and native
> acceptance remain open. See the
> [type-dispatch audit](../../docs/edr/native-engine/audits/2026-10-31-type-dispatch-target13/REPORT.md).
> Full replacement remains unqualified.

> **2026-10-30 root preclass metadata:** The kind-260 phase now receives
> `Magic`, `FileSize`, `ObjectID`, `FileType`, `RootFileType` and
> `sha2-256` from the root bytes and type. A synthetic CBC reads the type
> through JSON APIs and detects a harmless marker. The full engine suite,
> vet and Windows AMD64 cross-build passed. The metadata graph remains
> partial, and target-13 serialized-JSON scanning is still pending. See the
> [root metadata audit](../../docs/edr/native-engine/audits/2026-10-30-preclass-root-metadata/REPORT.md).
> Full replacement remains unqualified.

> **2026-10-30 ClamBC root preclassification:** Kind-260 hooks now run after
> the root scan; logical triggers from child files cannot activate them.
> Bare hooks, break behavior, recursive output and incremental CVD refresh
> are covered by synthetic tests. The metadata tree is still partial, and
> deprecated target-13 metadata scanning remains pending. The official v339
> bytecode feed has no kind-260 program, so native custom-feed differential
> testing also remains open. See the
> [preclass audit](../../docs/edr/native-engine/audits/2026-10-30-preclass-bytecode/REPORT.md).
> Full replacement remains unqualified.

> **2026-10-30 ClamBC startup preparation:** The signature compiler now
> validates and runs the first loaded kind-1 startup program before publishing
> a full or incremental feed generation. Failure rejects the new generation;
> a non-magic return follows ClamAV's normal-mode warning behavior. Synthetic
> CBC tests, signatures/bytecode/engine suites, vet and Windows AMD64
> cross-build passed. The official v339 feed contains no startup program, so
> native custom-feed differential validation remains open. See the
> [startup bytecode audit](../../docs/edr/native-engine/audits/2026-10-30-startup-bytecode/REPORT.md).
> Full replacement is still unqualified.

> **2026-10-30 ASCII logical byte comparisons:** The matcher now evaluates
> zero-offset and ASCII hex/decimal/auto forms, with exact-length and
> little-endian hex behavior matched to pinned ClamAV 1.5.4. A 600-case
> deterministic harmless-rule differential found zero verdict disagreements;
> signatures and engine suites, vet and Windows AMD64 build passed. Invalid
> and unmodeled forms remain incomplete. See the
> [ASCII byte-comparison audit](../../docs/edr/native-engine/audits/2026-10-30-bytecompare-ascii/REPORT.md).


> **2026-10-30 NSIS bzip2 extraction:** The Go engine now decodes NSIS's
> headerless bzip2 variant for separate and solid installer streams. A
> three-block 1,800,192-byte fixture decodes byte-exactly, and malformed and
> over-budget streams remain incomplete. A compressed-marker PE fixture was
> detected by both the Go scanner and pinned ClamAV 1.5.4 with a local NDB
> rule. This does not yet cover a corpus of installer binaries produced by
> NSIS bzip2 or close the other replacement gates. See the
> [NSIS bzip2 audit](../../docs/edr/native-engine/audits/2026-10-30-nsis-bzip2/REPORT.md).


> **2026-10-30 automatic Linux cgroup Go memory target:** The actual EDR
> executable now sets a 75%-of-cgroup Go soft target, capped at 1000 MiB,
> when the finite cgroup limit is 1–2 GiB and `GOMEMLIMIT` is unset. An
> explicit operator target takes precedence. No-override Linux ARM64 runs
> loaded all 3,628,228 official signatures and scanned a harmless file under
> hard 1, 1.5 and 2 GiB caps, peaking at 895,369,216, 1,136,193,536 and
> 1,135,943,680 bytes with no memory-limit/OOM events. Two no-override
> 694-file CLI scans with explicit 200/400 MiB input budgets matched every
> prior per-file verdict and peaked at 1,022,857,216 and 1,032,044,544
> bytes, without cgroup pressure. The override run, Linux policy tests,
> EDR host tests, vet, and Linux AMD64/Windows AMD64 cross-builds passed.
> The previous default 1 GiB feed-load OOM is resolved in these CLI probes;
> an isolated no-override 2 GiB service run also scheduled all 694 files,
> matched every scanner state, verified 296 quarantine hashes, and peaked
> at 1,229,758,464 bytes with no cgroup pressure. Live sensors,
> changed-feed service promotion, native Windows, independent malware
> efficacy and other ClamAV
> parity gates remain open. See the
> [memory-policy audit](../../docs/edr/native-engine/audits/2026-10-30-memory-policy/REPORT.md).
> Full replacement remains unqualified.

> **2026-10-30 signed CVD plus ThreatWind-format promotion:** A gated feed
> integration test promoted authenticated `main` 63, `daily` 28158→28167
> and `bytecode` 339 with a signed synthetic SHA-256 indicator list. Two
> isolated hard-2-GiB Linux ARM64 runs kept 889 and 863 EICAR scans
> detecting across both generations, activated a newly added daily rule,
> invalidated an old clean cache verdict and rejected a bad next artifact
> without losing the active generation. Cgroup peaks were 1,909,800,960
> and 1,913,438,208 bytes with zero limit/OOM events under
> `GOMEMLIMIT=1000MiB`; the full feed package suite, vet and focused race
> test passed. The
> indicator is synthetic and the mirror is local, so live ThreatWind
> acquisition and whole-service promotion remain unqualified. See the
> [signed promotion audit](../../docs/edr/native-engine/audits/2026-10-30-signed-promotion/REPORT.md).
> Full replacement remains unqualified.

> **2026-10-28 signed daily-CVD reload under live scans:** Two isolated
> Linux ARM64 hard-2-GiB runs moved an authenticated ClamAV daily CVD from
> version 28158 to 28167 while EICAR scans continued. All 823 and 773
> concurrent scans detected; both old and new engine generations served
> scans. The new feed added 64 signatures; a new exact NDB rule changed from
> clean to detected and remained detected after a corrupt follow-up update
> was rejected. Cgroup peaks were 1,469,173,760 and 1,463,480,320 bytes
> with zero memory-limit/OOM events under `GOMEMLIMIT=1000MiB`. This is
> direct-engine reload evidence,
> not an integrated-service promotion or full replacement qualification.
> See the [signed reload audit](../../docs/edr/native-engine/audits/2026-10-28-v169-daily-reload/REPORT.md).
> Full replacement remains unqualified.

> **2026-10-28 current-feed integrated service:** The cgroup-aware v169
> Linux ARM64 EDR service loaded 3,628,232 official-plus-local signatures
> and completed a 694-file scheduled scan under a hard 2 GiB cap with
> `GOMEMLIMIT=1000MiB`. All 296 detections, 376 clean and 22 incomplete
> results match the current scanner baseline; 296 quarantine records and
> stored files had matching paths, hashes and signatures. One inert marker
> restored with its exact SHA-256. The service stayed healthy and exited
> normally; cgroup peak was 1,207,848,960 bytes with zero memory-limit or
> OOM events. Live sensors and changed-feed reload were not exercised, and
> this does not qualify Windows, ThreatWind update promotion, real-malware
> efficacy or every conditional feed rule. See the
> [current service audit](../../docs/edr/native-engine/audits/2026-10-28-v169-service-current/REPORT.md).
> Full replacement remains unqualified.

> **2026-10-28 current-feed safe-corpus replay:** The cgroup-aware v169
> scanner completed two hard-1-GiB Linux ARM64, three-round replays of the
> 694-file generated/EICAR corpus with `GOMEMLIMIT=768MiB`. All 2,082
> verdict fields in each run match the pinned v163 baseline; peaks were
> 1,037,533,184 and 1,023,242,240 bytes with zero cgroup pressure/OOM
> events. A separate three-round run of 256 Apple-anchored benign controls
> returned 768 clean verdicts, peaking at 1,014,755,328 bytes without
> pressure. The explicit 200/400 MiB scan budgets reproduce the prior
> corpus policy. This is scanner-only evidence, with narrow memory margin,
> and does not establish service memory, independent malware efficacy or
> Windows runtime parity. See the
> [full safe-corpus audit](../../docs/edr/native-engine/audits/2026-10-28-v169-current-feed-694/REPORT.md).
> Full replacement remains unqualified.

> **2026-10-28 current-feed and Linux cgroup check:** CVDUpdate fetched
> authenticated `main` 63, `daily` 28167 and `bytecode` 339. The strict
> audit accepts 3,628,730 selected rules with zero rejects but still has
> **199 conditional unknowns**: 110 fuzzy-image, 81 icon-group, 8
> certificate-trust. A v169 Linux ARM64 scanner-only probe under a hard
> 1 GiB limit was OOM-killed during feed loading with the default Go
> runtime; `GOMEMLIMIT=768MiB` loaded the feed and scanned one benign file
> with an 875,495,424-byte cgroup peak. This is not service or corpus
> acceptance. Linux tuning now reads finite cgroup RAM limits instead of
> treating physical host memory as the available process budget; the
> 1 GiB native container test passes. A rebuilt scanner completed the
> same one-file probe with constrained tuning and an 873,312,256-byte
> cgroup peak. At that time, this tuning fix did not set a Go soft limit or
> resolve the default-memory load OOM; the October 30 policy above addresses
> that one-file probe. See the
> [current-feed audit](../../docs/edr/native-engine/audits/2026-10-28-v169-current-feed/REPORT.md).
> Full replacement remains unqualified.

> **2026-10-28 source candidate `native-go-v169`:** Anchor index construction
> now generates ASCII case variants in fixed-size packed keys. A controlled
> 20,000-anchor microbenchmark removed 20,000 allocations and about 160 KB
> per index build, with matching behavior preserved. The effect on the
> official full-feed peak was later measured on one benign scanner-only
> fixture; the 1 GiB endpoint gate is still open.
> See the [v169 anchor audit](../../docs/edr/native-engine/audits/2026-10-28-v169-anchor-keys/REPORT.md).

> **2026-10-28 source candidate `native-go-v168`:** GDB URL percent decoding
> now follows ClamAV 1.5.4's leading-escape handling and repeated decode
> loop, with no arbitrary eight-pass cutoff. A temporary source-loop C
> harness and Go agreed on 500 deterministic percent-encoded strings;
> permanent tests cover deep nesting and leading encoded host bytes. Full
> native differential and replacement acceptance remain open. See the
> [v168 GDB audit](../../docs/edr/native-engine/audits/2026-10-28-v168-gdb-percent/REPORT.md).

> **2026-10-28 source candidate `native-go-v167`:** GDB URL hashing now
> follows ClamAV 1.5.4's mutable buffer walk for chained `/./` and `/../`
> segments. A temporary C harness of the pinned source loop and the Go
> implementation agreed on 500 deterministic paths; permanent cases cover
> the counterintuitive chains that retain a later dot segment. Full native
> ClamAV differential and replacement acceptance remain open. See the
> [v167 GDB audit](../../docs/edr/native-engine/audits/2026-10-28-v167-gdb-dots/REPORT.md).

> **2026-10-28 source candidate `native-go-v166`:** GDB URL hashing now
> considers the reference's fourth directory prefix, handles repeated
> slashes, literal path spaces, non-ASCII hostname bytes, invalid percent
> escapes, `%00`, port/userinfo edge forms and one `/./` or `/../` segment.
> Protocol case and the six-path-candidate limit follow pinned ClamAV 1.5.4.
> Chained dot segments and remaining URL normalization edges stay incomplete;
> native differential and full replacement acceptance remain open. See the
> [v166 GDB audit](../../docs/edr/native-engine/audits/2026-10-28-v166-gdb-canon/REPORT.md).

> **2026-10-28 source candidate `native-go-v165`:** `.gdb` Safe Browsing
> S/S1/S2 full SHA-256 hashes and local W allow hashes now compile and match
> HTML destination URLs before displayed-link checks, with host suffixes and
> path/query prefixes. P records load but do not gate full hashes, matching
> pinned ClamAV 1.5.4. PDB/WDB flag text is ignored and Y uses X matching,
> as in that pinned parser. Ambiguous URL forms return incomplete analysis. The
> full EDR suite passes. This adds one phishing feed format; the remaining
> compatibility and native acceptance gates below still prevent full
> replacement qualification. See the
> [v165 GDB audit](../../docs/edr/native-engine/audits/2026-10-28-v165-gdb/REPORT.md).

> **2026-10-28 source candidate `native-go-v164`:** KWAJ methods 0–4 now
> decode in Go with bounded recursion and expansion. A malformed stream or
> false short header length does not suppress a detection in decoded bytes.
> Independent libmspack `msexpand` accepts all five method fixtures and
> yields the same payload; official valid/invalid header fixtures are tested.
> Full EDR tests, focused race, vet and Windows AMD64 engine cross-build pass.
> Remaining feed, format/bytecode, service sizing, Windows runtime and malware
> corpus gates still prevent full replacement qualification. See the
> [v164 KWAJ audit](../../docs/edr/native-engine/audits/2026-10-28-v164-kwaj/REPORT.md).

> **2026-10-28 response patch (scanner remains `native-go-v163`):** Windows WFP block events
> now require a classify-drop record, an EDR-owned filter ID and current
> non-exempt policy coverage; the event uses the record's actual direction.
> The dynamic WFP session closes on manager shutdown. Full EDR tests,
> netblock race, vet and Windows AMD64/ARM64 netblock cross-builds pass.
> Native Windows runtime proof remains open, as do the v163 scanner parity,
> resource and malware-efficacy gates. See the
> [WFP attribution audit](../../docs/edr/native-engine/audits/2026-10-28-wfp-attribution/REPORT.md).
> Full replacement remains unqualified.

> **2026-10-28 source candidate `native-go-v163`:** An unreadable icon-group
> data entry beneath a readable PE resource tree is now a definite icon
> miss, matching pinned ClamAV 1.5.4. The reference-backed test checks a
> valid detecting control and two damaged clean variants. Full EDR tests,
> focused race, vet and Linux AMD64/Windows AMD64/ARM64 builds pass;
> 2,082 constrained safe-corpus verdicts match v162. The strict feed still
> has 198 conditional unknowns. Native service, Windows runtime and malware
> efficacy gates remain open. See the
> [v163 audit](../../docs/edr/native-engine/audits/2026-10-28-v163-icon-group-data/REPORT.md).
> Full replacement remains unqualified.

> **2026-10-28 source candidate `native-go-v162`:** Compact hash slots
> deduplicate anchor literals during compilation; exact byte comparisons
> resolve collisions. Three Linux ARM64 full-feed profiles show 144.2 MB
> less median allocation and a 0.59-second faster load than v160, but an
> 8.1 MB higher median cgroup peak and essentially unchanged retained heap.
> Full EDR tests, focused race, vet, and Linux AMD64/Windows AMD64/ARM64
> builds pass. Under the hard 1 GiB cap, `GOMEMLIMIT=768MiB` produced 2,082
> exact v160 safe-corpus verdicts with no cgroup pressure. Default loading
> still OOM-kills; disposable traces locate the failure at anchor building.
> The strict feed still has 198 conditional unknowns, and native service,
> Windows runtime and malware efficacy gates remain open. See the
> [v162 audit](../../docs/edr/native-engine/audits/2026-10-28-v162-anchor-hash32/REPORT.md).
> Full replacement remains unqualified.

> **2026-10-28 source candidate `native-go-v160`:** Signature-body parsing
> now allocates each contiguous exact-hex literal once. Three Linux ARM64
> 2 GiB full-feed profiles show median cgroup peak 48.5 MB lower, retained
> heap 13.8 MB lower and cumulative allocation 59.9 MB lower than v159.
> Full EDR tests, focused race, vet and Linux AMD64/Windows AMD64/ARM64
> builds pass; 2,776 constrained safe-corpus verdicts match v159. Default
> 1 GiB loading still OOM-kills, the strict feed has 198 conditional unknowns,
> and native service/Windows/malware efficacy gates remain open. See the
> [v160 audit](../../docs/edr/native-engine/audits/2026-10-28-v160-literal-reserve/REPORT.md).
> Full replacement remains unqualified.

> **2026-10-28 source candidate `native-go-v159`:** Publisher-authenticated
> feed preflight supplies bounded capacity hints for packed hash indexes and
> compiled pattern arrays. Three Linux ARM64 2 GiB profiles show median
> cumulative allocation 1.14 GB and cgroup peak 45.6 MB lower than v157,
> with 22.9 MB more retained heap and about two seconds more load time. Full
> EDR tests, focused race, vet and Linux AMD64/Windows AMD64/ARM64 builds
> pass. The 2,776 one- and three-round safe-corpus verdicts match v157.
> A 768 MiB Go soft target keeps the scanner-only 1 GiB cgroup peak at
> 1,038,938,112 B without pressure events, but default loading still
> OOM-kills and endpoint service headroom is unproven. Strict audit retains
> 198 conditional unknowns. See the
> [v159 audit](../../docs/edr/native-engine/audits/2026-10-28-v159-pattern-reserve/REPORT.md).
> Full replacement remains unqualified.

> **2026-10-28 source candidate `native-go-v157`:** Full-feed CVD/CLD
> compilation now preflights member metadata and streams text members one
> at a time through a bounded reader, including `main.mdb`. Three paired
> Linux ARM64 2 GiB profiles show median cgroup peak 108.5 MB and cumulative
> allocation 309.4 MB lower than v155, with about 1.5 seconds more load
> time. Full EDR tests, focused race, vet and Linux AMD64/Windows AMD64/ARM64
> builds pass; 2,776 safe-corpus verdicts match v155. Default 1 GiB loading
> still OOM-kills, successful constrained runs touch the cap and the strict
> feed retains 198 conditional unknowns. See the
> [v157 audit](../../docs/edr/native-engine/audits/2026-10-28-v157-member-stream/REPORT.md).
> Full replacement remains unqualified.

> **2026-10-28 negative memory experiment:** Whole-CVD replay after a
> metadata/selection preflight still OOM-killed under 1 GiB with default GC
> and `GOGC=20`. The temporary v156 code was reverted to v155 at that
> point. See the
> [replay audit](../../docs/edr/native-engine/audits/2026-10-28-v156-replay-negative/REPORT.md).
> The later v157 compiler added individual-member streaming and direct
> large-member parsing.

> **2026-10-28 source candidate `native-go-v155`:** Common pattern
> compilation uses local temporary arrays and hashes keys through a reused
> buffer. Three paired Linux ARM64 full-feed profiles show median cumulative
> allocation 131.0 MB lower and cgroup peak 43.4 MB lower than v154, with
> retained heap essentially unchanged. Full EDR tests, focused race, vet
> and Linux AMD64/Windows AMD64/ARM64 builds pass; 2,776 safe-corpus
> verdicts match v154. Default 1 GiB loading still OOM-kills, successful
> constrained runs touch the cap and the strict feed retains 198 conditional
> unknowns. See the
> [v155 audit](../../docs/edr/native-engine/audits/2026-10-28-v155-pattern-temp/REPORT.md).
> Full replacement remains unqualified.

> **2026-10-28 source candidate `native-go-v154`:** Fixed-array hash-line
> parsing removes a per-rule split allocation. Three paired Linux ARM64
> full-feed profiles show median cumulative allocation 162.9 MB lower,
> sampled load heap 91.5 MB lower and cgroup peak 27.7 MB lower than v153.
> Full EDR tests, focused race, vet and Linux AMD64/Windows AMD64/ARM64
> builds pass; 2,776 safe-corpus verdicts match v153. Default 1 GiB loading
> still OOM-kills, successful constrained runs touch the cap, and the
> strict feed retains 198 conditional unknowns. See the
> [v154 audit](../../docs/edr/native-engine/audits/2026-10-28-v154-hash-parse/REPORT.md).
> Full replacement remains unqualified.

> **2026-10-28 source candidate `native-go-v153`:** Hash-rule names use a
> length-prefixed arena and four-byte sorted references. Three paired
> Linux ARM64 2 GiB full-feed profiles show median retained heap 12.8 MB
> lower and median cgroup peak 72.4 MB lower than v152. Full EDR tests,
> focused race, vet and Linux AMD64/Windows AMD64/ARM64 builds pass; 2,776
> safe-corpus verdicts match v152. Default 1 GiB loading still OOM-kills,
> successful constrained runs touch the cap, and the strict feed retains
> 198 conditional unknowns. See the
> [v153 audit](../../docs/edr/native-engine/audits/2026-10-28-v153-hash-names/REPORT.md).
> Full replacement remains unqualified.

> **2026-10-28 source candidate `native-go-v152`:** Packed hash-index size
> fields are four bytes instead of eight, matching the accepted feed size
> range. Large object sizes cannot alias a smaller exact rule or the
> wildcard marker. Full EDR tests, focused race, vet and Linux AMD64/Windows
> AMD64/ARM64 builds pass. Linux ARM64 one- and three-round full-feed safe
> corpora show 2,776 exact v151 verdicts and lower sampled load heap.
> Default 1 GiB loading still OOM-kills, successful diagnostic runs touch
> the cap, and the strict feed retains 198 conditional unknowns. See the
> [v152 audit](../../docs/edr/native-engine/audits/2026-10-28-v152-hash-width/REPORT.md).
> Full replacement remains unqualified.

> **2026-10-28 source candidate `native-go-v151`:** Optional rule descriptors,
> hooks and count caps moved off the common rule, shrinking it from 200 to
> 64 bytes. One- and three-round full-feed Linux ARM64 safe corpora show zero
> verdict drift and about 59–60 MB less sampled load heap. Full EDR tests,
> focused race, vet and Linux AMD64/Windows AMD64/ARM64 builds pass. Default
> 1 GiB loading still OOM-kills, successful diagnostic runs touch the cap,
> and the strict feed retains 198 conditional unknowns. See the
> [v151 audit](../../docs/edr/native-engine/audits/2026-10-28-v151-rule-cold/REPORT.md).
> Full replacement remains unqualified.

> **2026-10-28 source candidate `native-go-v150`:** HWPOLE2 typed objects
> now scan the nested document after their four-byte prefix, matching the
> reference's behavior on declared-size mismatches. Official ClamAV 1.5.4 and
> Go agree on six inert wrapper fixtures. Full EDR tests, focused race, vet
> and Linux AMD64/Windows AMD64/ARM64 builds pass; three Linux ARM64
> safe-corpus rounds show zero drift. Default 1 GiB loading still OOM-kills,
> while a 768 MiB soft target reaches the cap. The strict feed retains 198
> conditional unknowns. See the
> [v150 audit](../../docs/edr/native-engine/audits/2026-10-28-v150-hwpole2/REPORT.md).
> Full HWP and endpoint replacement remain unqualified.

> **2026-10-28 source candidate `native-go-v149`:** Anchor lookup buckets
> use immutable-index-owned slabs. Three paired full-feed loads show 5.1 MB
> less retained heap and 12.8 MB lower sampled load heap peak, while RSS
> remains variable. Full EDR tests, focused race, vet and Linux AMD64 and
> Windows AMD64/ARM64 builds pass. Three Linux ARM64 safe-corpus rounds show
> zero drift. Default 1 GiB loading still OOM-kills; a 768 MiB Go soft target
> succeeds but reaches the cap. The strict feed retains 198 conditional
> unknowns. See the
> [v149 audit](../../docs/edr/native-engine/audits/2026-10-28-v149-bucket-slabs/REPORT.md).
> Full replacement remains unqualified.

> **2026-10-28 source candidate `native-go-v148`:** Literal-only anchors are
> derived when the index is built, shrinking the common body-pattern record
> from 112 to 88 bytes. Three paired full-feed loads show about 25 MB less
> retained heap and lower sampled RSS, but endpoint peak memory is not yet
> qualified. Full EDR tests, focused race, vet, Linux AMD64 and Windows
> AMD64/ARM64 builds pass. Three Linux ARM64 cgroup rounds on the 694-file
> safe corpus have zero per-file drift. Default 1 GiB loading still OOM-kills;
> both 768 and 640 MiB Go soft targets reach the cap. The strict feed retains 198
> conditional unknowns. See the
> [v148 audit](../../docs/edr/native-engine/audits/2026-10-28-v148-anchor-storage/REPORT.md).
> Full replacement remains unqualified.

> **2026-10-28 source candidate `native-go-v147`:** The common body-pattern
> record shrinks from 136 to 112 bytes by keeping matching graphs only for
> non-literal patterns. Three paired full-feed loads show about 30 MB less
> retained heap, while sampled process peak RSS was higher. Full EDR tests,
> focused race, vet and Linux AMD64/Windows AMD64/ARM64 builds pass. Three
> 694-file Linux ARM64 cgroup rounds have zero per-file drift. Default
> 1 GiB loading still OOM-kills; a 768 MiB Go soft target reaches the cap.
> The strict feed retains 198 conditional unknowns. See the
> [v147 audit](../../docs/edr/native-engine/audits/2026-10-28-v147-graph-storage/REPORT.md).
> Full replacement remains unqualified.

> **2026-10-28 source candidate `native-go-v146`:** Signature offsets now
> use a 16-byte bounded representation instead of 40 bytes. Across
> three alternating full-feed diagnostics each, retained heap fell from
> about 832 MB to 808 MB; peak RSS varied. Full EDR tests, focused race,
> vet, Windows AMD64/ARM64 and Linux AMD64 builds pass. A 694-file macOS
> corpus and three Linux ARM64 cgroup rounds show zero per-file drift.
> Default 1 GiB load still OOM-kills; a 768 MiB Go soft target completes
> with no headroom. The strict feed retains 198 conditional unknowns.
> See the [v146 audit](../../docs/edr/native-engine/audits/2026-10-28-v146-compact-offset/REPORT.md).
> Full replacement remains unqualified.

> **2026-10-28 source candidate `native-go-v145`:** EGG block CRCs and
> decoded lengths, plus member total lengths, are checked before an
> undetected archive can be called clean. Decoded children still reach
> the scanner when these fields disagree. Pinned ClamAV 1.5.4 retains
> child-only detections on the three inert mutations; Go conservatively
> returns incomplete on their nonmatching controls. Full EDR tests,
> focused race, vet, Windows/Linux AMD64 builds and the 694-file
> official-feed corpus pass, with zero v144 per-file drift. The strict
> gate still has 198 conditional unknown rules. See the
> [integrity audit](../../docs/edr/native-engine/audits/2026-10-28-v145-egg-integrity/REPORT.md).
> Full replacement and endpoint release acceptance remain open.

> **2026-10-28 source candidate `native-go-v144`:** Top-level EGG archive
> comments now scan as separate children, including valid UTF-8 with the
> multibyte flag. Eight harmless reference fixtures match pinned ClamAV;
> full EDR tests and the 694-file official-feed corpus pass with zero
> drift. The [comment audit](../../docs/edr/native-engine/audits/2026-10-28-v144-egg-comments/REPORT.md)
> has source and oracle receipts. A separate [1 GiB Linux ARM64 cgroup
> probe](../../docs/edr/native-engine/audits/2026-10-28-v144-memory-cgroup/REPORT.md)
> completed 2,082 safe-corpus scans only with a 768 MiB Go soft memory
> target; the default run was OOM-killed. Native AMD64, full service,
> changed-generation and efficacy acceptance remain open.

> **2026-10-28 source candidate `native-go-v143`:** Ordinary non-solid
> EGG archives now scan stored, raw-deflate and bzip2 members, including
> mixed blocks. Seven inert archives and two EGG metadata rules have
> pinned ClamAV 1.5.4 reference receipts. Full EDR tests, focused race,
> vet and Linux/Windows AMD64 builds pass. The 694-file official-feed
> corpus has zero drift from v142; the strict feed still has 198
> conditional unknowns. See the
> [v143 EGG audit](../../docs/edr/native-engine/audits/2026-10-28-v143-egg/REPORT.md).

> **2026-10-28 source candidate `native-go-v142`:** ALZ stored, bzip2 and
> raw-deflate members now expand into the recursive Go scanner. Eight
> inert archives and two ALZ metadata rules have pinned ClamAV 1.5.4
> reference receipts. Full EDR tests, focused race, vet and Linux/Windows
> AMD64 builds pass; the 694-file official-feed corpus has zero verdict
> drift from v141. The strict feed remains at 198 conditional unknowns.
> See the [v142 ALZ audit](../../docs/edr/native-engine/audits/2026-10-28-v142-alz/REPORT.md).

> **2026-10-28 source candidate `native-go-v141`:** LHA `-lh1-`
> adaptive-Huffman members now expand into the recursive Go scanner,
> including tree reconstruction and an OSK level-2 header terminator.
> Eight benign historical archives detect child-only hashes and scan
> clean with nonmatching rules in both Go and pinned ClamAV 1.5.4.
> The 694-file official-feed corpus has zero verdict drift from v140;
> the strict gate still fails on 198 conditional unknowns. See the
> [v141 LH1 audit](../../docs/edr/native-engine/audits/2026-10-28-v141-lh1/REPORT.md).

> **2026-10-28 source candidate `native-go-v140`:** LHA level-3
> word-size-4 headers and four-byte extended-header chains now lead
> stored, LH5 and LZ5 members to the recursive scanner. Five harmless
> layouts detect child-only hashes in pinned ClamAV 1.5.4 and Go;
> malformed chains remain incomplete. See the
> [v140 level-3 audit](../../docs/edr/native-engine/audits/2026-10-28-v140-lha-level3/REPORT.md).

> **2026-10-28 source candidate `native-go-v139`:** LArc `-lz4-`,
> `-lzs-` and `-lz5-` members now reach the recursive Go scanner.
> Nine harmless reference-verified fixtures cover stored data,
> literals, back-references, dictionary wraparound, multiple flag
> groups and preset history. The 694-file
> official-feed safe corpus has zero verdict drift from v138; the
> strict feed still has 198 conditional unknowns. See the
> [v139 LArc audit](../../docs/edr/native-engine/audits/2026-10-28-v139-larc/REPORT.md).

> **2026-10-28 source candidate `native-go-v138`:** LHA level-1 and level-2
> headers, including extended-header chains, now lead stored and supported
> compressed members to the recursive scanner. Eight harmless layouts
> have child-only HDB detections in pinned ClamAV 1.5.4 and Go. Malformed
> extension lengths remain incomplete. Full EDR tests, focused race,
> vet, Windows/Linux AMD64 builds and the 694-file official-feed corpus
> pass with zero per-file drift from v137. The strict feed still has 198
> conditional unknowns. See the
> [v138 LHA header audit](../../docs/edr/native-engine/audits/2026-10-28-v138-lha-levels/REPORT.md).

> **2026-10-28 source candidate `native-go-v137`:** The LHA reader now
> decodes bounded `-lh4-` through `-lh7-` static-Huffman/LZ members and
> scans their expanded child. Six harmless fixtures, including a
> back-reference and a two-symbol Huffman tree, have pinned ClamAV 1.5.4
> child-hash detection and nonmatching-rule clean controls; Go detects the
> same child hashes. Full EDR tests, focused race, vet, Windows/Linux AMD64
> builds and the 694-file official-feed corpus pass with zero per-file
> drift from v136. The strict feed still has 198 conditional unknowns.
> See the [v137 compressed LHA audit](../../docs/edr/native-engine/audits/2026-10-28-v137-lha-huffman/REPORT.md).

> **2026-10-28 source candidate `native-go-v136`:** Bounded Go level-0
> LHA `-lh0-` parsing now scans stored child files, closing a
> reference-confirmed child-hash detection gap. Three harmless layouts
> detect in pinned ClamAV 1.5.4 and scan clean with a nonmatching rule;
> the Go test pins the same later-member fixture. Unsupported compressed
> methods remain incomplete. Focused race, full EDR package tests, vet,
> Windows/Linux AMD64 builds and the 694-file official-feed corpus pass
> with zero per-file drift from v135. The strict feed retains 198
> conditional unknowns. See the
> [v136 LHA audit](../../docs/edr/native-engine/audits/2026-10-28-v136-lha-stored/REPORT.md).

> **2026-10-28 source candidate `native-go-v135`:** CPIO archives no longer
> require a `TRAILER!!!` member after complete scanned members. Four
> harmless odc/newc/binary fixtures match pinned ClamAV 1.5.4: each is
> clean with a nonmatching rule and detects a child-only hash rule. A full
> invalid later header remains incomplete. Full EDR tests, focused race,
> vet and Windows/Linux AMD64 builds pass. The 694-file corpus now has
> 296 detected, 376 clean and 22 incomplete; strict feed still has 198
> conditional unknowns. See the
> [v135 CPIO EOF audit](../../docs/edr/native-engine/audits/2026-10-28-v135-cpio-eof/REPORT.md).

> **2026-10-28 source candidate `native-go-v134`:** Uncompressed BMP fuzzy
> hashing now uses a bounded Go decoder matching the pinned reference's
> core/Info/V2/V3/V4/V5 headers, 1/2/4/8-bit palettes, 24/32-bit RGB,
> top-down rows and pixel-offset gaps. Ten harmless BMPs have exact ClamAV
> 1.5.4 hashes and custom-rule detections. Full EDR tests, focused race,
> vet, Windows/Linux AMD64 builds and the 694-file corpus pass with zero
> per-file drift; the strict feed retains 198 conditional unknowns. See the
> [v134 plain BMP audit](../../docs/edr/native-engine/audits/2026-10-28-v134-bmp-plain/REPORT.md).

> **2026-10-28 source candidate `native-go-v133`:** The fuzzy-image matcher
> now decodes 16-bit RGB555 and 16/32-bit BI_BITFIELDS BMPs, including
> top-down rows and 4-, 5-, 6-, 8- and 10-bit channels. Five harmless images
> have exact pinned ClamAV 1.5.4 hashes and custom-rule detections.
> Encoded RLE8 runs crossing a row boundary now follow ClamAV's row clamp;
> a sixth inert fixture has an exact hash and custom-rule detection.
> Malformed bitfields remain incomplete. Full EDR tests, focused race, vet,
> Windows/Linux AMD64 builds, and the 694-file corpus pass with zero verdict
> drift; the strict feed retains 198 conditional unknowns. See the
> [v133 BMP bitfields audit](../../docs/edr/native-engine/audits/2026-10-28-v133-bmp-bitfields/REPORT.md).

> **2026-10-28 source candidate `native-go-v132`:** The fuzzy-image matcher
> now decodes bounded indexed BMP RLE4 and RLE8 streams, including encoded
> runs, literal spans and deltas. Eight inert images produce exact pinned
> ClamAV 1.5.4 fuzzy hashes and custom-rule detections. Unsupported or
> malformed layouts remain incomplete. Full EDR tests, focused race, vet
> and Windows/Linux AMD64 builds pass; the 694-file safe corpus has zero
> verdict drift. The strict feed still has 198 conditional unknowns. See the
> [v132 BMP RLE audit](../../docs/edr/native-engine/audits/2026-10-28-v132-bmp-rle/REPORT.md).

> **2026-10-28 source candidate `native-go-v131`:** A `.cat` source that
> fails signed-catalog verification is now skipped without adding hashes or
> certificate trust; other selected rules continue to load. This matches the
> pinned ClamAV 1.5.4 catalog loader and an inert direct CLI differential.
> The compiler and auditor count `invalid-catalog` skips. A valid catalog
> still requires CRB trust and counts against the signature limit. Full EDR
> tests, focused race checks, vet and Windows/Linux builds pass; the
> 694-file safe corpus has zero verdict drift. The strict feed gate still
> fails on 198 conditional unknowns. See the
> [v131 catalog-skip audit](../../docs/edr/native-engine/audits/2026-10-28-v131-catalog-skip/REPORT.md).

> **2026-10-28 source candidate `native-go-v130`:** The PE icon reader now
> reads image bytes from the mapped file rather than the resource entry's
> declared size, and cleanly skips unmapped or file-truncated images while
> allowing a later icon to detect. Eight inert PE fixtures agree with pinned
> ClamAV 1.5.4. The full EDR suite, focused race test, vet and Windows/Linux
> AMD64 builds pass; the 694-file safe corpus has zero verdict drift from
> v129. The strict feed still has 198 conditional unknowns, so replacement
> is not accepted. See the [v130 icon-bounds audit](../../docs/edr/native-engine/audits/2026-10-28-v130-icon-file-bounds/REPORT.md).

> **2026-10-28 source candidate `native-go-v129`:** The PE icon parser now
> follows pinned ClamAV 1.5.4 on group-dimension mismatches, extended bitmap
> headers, unused compression/palette-count fields, odd heights, rectangular
> scaling, and 192/256-pixel bitmaps. Seventeen inert PE fixtures agree on
> verdict; three wide or rectangular hashes match the reference debug output
> exactly. The full EDR suite, focused race test, vet and Windows/Linux AMD64
> builds pass. The 694-file safe corpus has zero verdict drift from v128. The
> strict feed still has 198 conditional unknowns, so full replacement is not
> accepted. See the [v129 icon-image audit](../../docs/edr/native-engine/audits/2026-10-28-v129-icon-image/REPORT.md).

> **2026-10-28 source candidate `native-go-v128`:** Icon-group parsing now
> visits complete entries in a group with a short tail, treats a too-short
> group as a definite nonmatch, and follows the default 100-entry limit.
> Six harmless fixtures agree with pinned ClamAV 1.5.4, including a
> child-icon detection before a missing second entry. See the
> [v128 icon-tail audit](../../docs/edr/native-engine/audits/2026-10-28-v128-icon-group-tail/REPORT.md).

> **2026-10-27 source candidate `native-go-v127`:** The PE-header parser now
> matches pinned ClamAV 1.5.4 on 13 harmless fixtures under both broken-PE
> settings. Ten malformed headers alert in the endpoint's strict profile and
> are clean in the permissive profile after raw and embedded analysis. A
> detected ZIP member takes precedence over its broken parent. Full EDR
> tests, focused race checks, vet and Windows/Linux AMD64 builds pass; the
> 694-file official-feed corpus has zero drift from v126. The strict feed
> still has 198 conditional unknowns. See the
> [v127 PE-header audit](../../docs/edr/native-engine/audits/2026-10-27-v127-broken-pe/REPORT.md).

> **2026-10-27 source candidate `native-go-v126`:** FSG 1.31 PE32 output now
> rebuilds byte-for-byte like pinned ClamAV 1.5.4 on three harmless layouts.
> The scanner detects a hash present only in the rebuilt child when the
> broken-executable setting is off; in the incumbent strict profile both
> engines report the broken-PE label on these overlapping-section cases.
> Full EDR tests, focused race tests, vet and Windows/Linux AMD64 builds
> pass. The 694-file safe corpus has zero verdict drift from v125, and the
> strict feed gate retains 198 conditional unknowns. See the
> [v126 FSG 1.31 audit](../../docs/edr/native-engine/audits/2026-10-27-fsg131/REPORT.md).

> **2026-10-27 v125 validation:** The 142 archived TIFF images now have exact,
> certain fuzzy hashes and decisive rule outcomes in Go. Pinned ClamAV
> 1.5.4 agrees on 142 matching detections and 142 one-bit nonmatches. The
> [TIFF rule audit](../../docs/edr/native-engine/audits/2026-10-27-tiff-fuzzy-rule-validation/REPORT.md)
> supersedes the earlier nine-incomplete result for those fixtures. The
> scanner code and version are unchanged; 198 conditional feed rules remain.

> **2026-10-27 source candidate `native-go-v125`:** Reached PDF PARSED hooks
> now receive proved stream-object flags for ordinary unfiltered streams and
> simple single ASCIIHex or Flate filters. Four custom ClamBC cases agree
> with pinned ClamAV 1.5.4. Full EDR tests, focused race checks, vet, and
> Windows/Linux AMD64 builds pass; the 694-file official-feed corpus has
> zero verdict drift from v124. The strict feed audit still lists 198
> unknown-when-reached rules. See the
> [v125 PDF stream-flag audit](../../docs/edr/native-engine/audits/2026-10-27-v125-pdf-stream-flags/REPORT.md).

> **2026-10-27 source candidate `native-go-v124`:** A reached PDF END hook
> now runs when a simple missing-EOF document's only final flag is the
> reference-verified `BAD_PDF_TRAILER` bit. Three harmless custom-hook
> cases agree with pinned ClamAV 1.5.4. Full EDR tests, focused race
> checks, vet and Windows/Linux AMD64 builds pass; the 694-file official
> feed has zero verdict drift from v123. The strict audit still lists
> 198 unknown-when-reached rules. See the
> [v124 PDF END audit](../../docs/edr/native-engine/audits/2026-10-27-v124-pdf-end-hook/REPORT.md).

> **2026-10-27 source candidate `native-go-v123`:** The PDF parser supplies
> exact object flags to ClamBC hooks for a bounded set of complete,
> independent objects. Six harmless differentials agree with pinned
> ClamAV 1.5.4; complex flags remain incomplete. Full EDR tests, focused
> race checks, vet and Windows/Linux AMD64 builds pass. The 694-file
> official-feed corpus has zero verdict drift from v122, while the strict
> feed audit still lists 198 unknown-when-reached rules. See the
> [v123 PDF object-flag audit](../../docs/edr/native-engine/audits/2026-10-27-v123-pdf-object-flags/REPORT.md).

> **2026-10-27 source candidate `native-go-v122`:** Reached PARSED and
> POSTDUMP PDF hooks now receive a proved snapshot of ordinary indirect
> objects; count, lookup, packed ID, offset, size and bounded raw reads
> agree with seven custom ClamBC cases in pinned ClamAV 1.5.4. A PARSED-only
> label stays incomplete until the parser reaches that phase. Flags and
> parser-mutating hook effects remain incomplete. Full EDR tests, focused
> race checks, vet, Windows/Linux AMD64 builds and a 694-file official-feed
> corpus pass with zero verdict drift from v121. The strict feed audit still
> has 198 unknown-when-reached rules. See the
> [v122 PDF object API audit](../../docs/edr/native-engine/audits/2026-10-27-v122-pdf-object-api/REPORT.md).

> **2026-10-27 source candidate `native-go-v121`:** Reached nonempty PDF
> stream extractions dispatch POSTDUMP hooks on the extracted bytes. Custom
> ClamBC fixtures agree with pinned ClamAV on the verified paths. See the
> [v121 PDF POSTDUMP audit](../../docs/edr/native-engine/audits/2026-10-27-v121-pdf-postdump/REPORT.md).

> **2026-10-27 source candidate `native-go-v120`:** PDF bytecode hooks now
> run in ClamAV's PRE(4), PARSED(1), POSTDUMP(2), END(3) order. The VM answers
> the empty PRE object-table APIs and exact header/trailer flags when EOF is
> missing. Custom ClamBC programs match pinned ClamAV on phase precedence,
> PRE lookup, and PRE flag values. A label in POSTDUMP or END remains
> incomplete until parser reachability is modeled; a canonical PDF no longer
> yields a false END-only detection. The 694-file official-feed corpus has
> zero verdict drift from v119. See the
> [v120 PDF hook audit](../../docs/edr/native-engine/audits/2026-10-27-v120-pdf-bytecode/REPORT.md).


> **2026-10-27 source candidate `native-go-v119`:** ClamAV's default scan
> disables metadata collection. The Go bytecode VM now returns inactive for
> `json_is_active` and -1 for JSON queries in that mode; Word/Workbook XML
> properties no longer activate JSON bytecode unconditionally. A custom
> ClamBC differential against pinned ClamAV confirms normal and `--gen-json`
> behavior. Full EDR tests and the 694-file official-feed corpus pass with
> zero verdict drift from v118. See the
> [v119 correction audit](../../docs/edr/native-engine/audits/2026-10-27-v119-json-default/REPORT.md).


> **2026-10-27 source candidate `native-go-v118`:** The ClamBC JSON query
> APIs now execute in a bounded VM, and Word/Workbook XML properties are
> collected for logical callbacks. Word/Workbook attributes, `binData` count,
> and typed document-property values match pinned ClamAV 1.5.4 on controlled
> fixtures. The Office XML conservative incomplete gate remains while general
> metadata, other hooks, and independent efficacy are unverified. See the
> [v118 JSON API audit](../../docs/edr/native-engine/audits/2026-10-27-v118-json-bytecode/REPORT.md).

> **2026-10-27 source candidate `native-go-v117`:** The Office XML parser
> repairs unterminated decimal and lowercase-x hexadecimal numeric entities
> before scanning `binData`. Both forms detect in pinned ClamAV and Go;
> [the audit](../../docs/edr/native-engine/audits/2026-10-27-v117-msxml-entities/REPORT.md)
> records the differential and the unchanged 694-file corpus.

> **2026-10-27 source candidate `native-go-v113`:** Authenticode now accepts
> MD5 PE image digests and SHA/MD5 hash-OID aliases under the existing
> signed-digest, CRB-chain and blocklist checks. One generated signed PE
> matches pinned ClamAV 1.5.4 before and after a code-byte change. Full
> EDR tests (with loopback access for two feed packages), focused race,
> vet, Windows/Linux AMD64 builds and the 694-file official-feed safe
> corpus pass; all corpus verdict tuples match v112. The
> [v113 Authenticode audit](../../docs/edr/native-engine/audits/2026-10-27-v113-authenticode-digest/REPORT.md)
> records the controls and limits. The 198 conditional feed rules, native
> Windows runtime, independent efficacy and integrated acceptance remain open.

> **2026-10-27 source candidate `native-go-v112`:** PNG fuzzy hashing now
> matches pinned ClamAV on truncated terminal IEND CRCs and zero-length
> non-IDAT terminal chunks. All 1,910 controlled variants from 191 PNGs
> retain the reference hash and Go matches them, while missing IEND,
> terminal IDAT and bad IDAT controls remain negative. Full EDR tests,
> focused race, vet, sequential Windows/Linux AMD64 builds and the 694-file
> safe corpus pass with zero verdict drift. Focused PNG tests and the
> 369-image reference oracle also pass in a bounded, network-isolated Linux
> ARM64 container. The
> [v112 PNG terminal audit](../../docs/edr/native-engine/audits/2026-10-27-v112-png-terminal/REPORT.md)
> records evidence. The 198 conditional feed rules, native Windows runtime,
> independent efficacy and integrated acceptance remain open.

> **2026-10-27 source candidate `native-go-v111`:** PNG fuzzy hashing now
> matches pinned ClamAV 1.5.4 when only the terminal IEND CRC is bad.
> All 191 checked-in PNG oracle pairs retain identical reference hashes;
> the Go matcher matches them without copying the input or accepting a bad
> IDAT CRC. Full EDR tests, focused race, vet, Windows/Linux AMD64 builds
> and the 694-file official-feed safe corpus pass with zero verdict drift.
> The [v111 PNG audit](../../docs/edr/native-engine/audits/2026-10-27-v111-png-iend/REPORT.md)
> records evidence. The strict feed gate still has 198 conditional rules;
> native Windows runtime, independent efficacy and integrated acceptance
> remain open.

> **2026-10-27 source candidate `native-go-v110`:** XZ now scans bounded
> decoded children when its final footer is missing or partial, after a
> reconstructed-footer preflight checks dictionary and expansion limits.
> A controlled hidden-marker oracle shows pinned ClamAV reports the damaged
> stream clean while Go detects it; the earlier v109 XZ parity inference was
> confounded by visible marker bytes and is corrected. Full EDR tests,
> focused race, vet, Windows/Linux AMD64 builds, a 160,821-input fuzz run,
> and the 694-file full-feed safe corpus passed with zero verdict drift.
> The [v110 XZ audit](../../docs/edr/native-engine/audits/2026-10-27-v110-xz-partial/REPORT.md)
> records the result. The strict feed gate, native Windows runtime,
> independent efficacy, and integrated service acceptance remain open.

> **2026-10-27 source candidate `native-go-v109`:** Gzip now scans decoded
> child bytes despite a late CRC or footer error. Harmless marker oracles
> match pinned ClamAV 1.5.4 on valid, bad-CRC and missing-footer streams.
> Damaged undetected streams remain incomplete and child expansion remains
> bounded. Full EDR tests, focused race, vet, Windows/Linux AMD64 builds,
> and the 694-file full-feed safe corpus pass with zero verdict drift from
> v108. One offline Linux ARM64 scanner-only run under a hard 2 GiB cap
> and 1000 MiB Go soft heap target matched all tuples, peaking at
> 1,367,379,968 bytes without max/OOM events. The
> [v109 gzip audit](../../docs/edr/native-engine/audits/2026-10-27-v109-gzip-partial/REPORT.md)
> records gzip evidence; its XZ inference is corrected in the v110 audit.
> The strict feed gate still has 198 conditional rules; native Windows,
> independent efficacy and integrated service acceptance remain open.

> **2026-10-27 source candidate `native-go-v108`:** A PE icon group whose
> type-3 resource or requested icon ID is definitely absent now yields a
> clean miss, matching pinned ClamAV 1.5.4 on two generated fixtures.
> Malformed directories remain unknown. Full EDR tests, focused race, vet,
> Windows/Linux AMD64 builds, and the 694-file full-feed safe corpus pass
> with zero verdict drift from v107. One offline Linux ARM64 scanner-only
> run under a hard 2 GiB cap and 1000 MiB Go soft heap target matched all
> 694 tuples, peaking at 1,356,857,344 bytes without max/OOM events. The
> [v108 icon audit](../../docs/edr/native-engine/audits/2026-10-27-v108-icon-absent/REPORT.md)
> records details. The strict gate still has 198 conditional rules;
> integrated service, native Windows and efficacy acceptance remain open.

> **2026-10-27 source candidate `native-go-v107`:** Triggerless PE, ELF and
> Mach-O bytecode hooks now run at their parser hook points, including after
> an incremental CVD refresh that preserves hook order. Harmless generated
> fixtures for all four kinds match pinned ClamAV 1.5.4. Full EDR tests,
> focused race, vet, Windows/Linux AMD64 builds and the 694-file safe corpus
> pass with zero verdict drift from v106. The strict feed gate still has 198
> conditional unknown rules. One offline Linux ARM64 scanner-only run with a
> 1000 MiB Go soft heap target matched all 694 tuples under a hard 2 GiB
> cap, peaking at 1,394,647,040 bytes without max/OOM events. Integrated
> service and changed-feed memory remain open. See the
> [v107 triggerless-hook audit](../../docs/edr/native-engine/audits/2026-10-27-v107-triggerless-hooks/REPORT.md).

> **2026-10-27 source candidate `native-go-v106`:** ELF and Mach-O ClamBC
> unpacker hooks (261/262) now run after raw matching and scan emitted
> children through the bounded scanner. Generated trigger, no-trigger and
> child-output cases match pinned ClamAV 1.5.4. Full EDR tests, race, vet,
> Windows/Linux AMD64 builds, and the 694-file official-feed safe corpus
> pass without verdict drift. A scanner-only Linux ARM64 run under a hard
> 2 GiB cap matched all 694 tuples; the unset-heap run touched the cap,
> while `GOMEMLIMIT=1000MiB` left 821 MB headroom. The
> [v106 hook audit](../../docs/edr/native-engine/audits/2026-10-27-v106-executable-hooks/REPORT.md)
> preserves oracle, feed and cgroup receipts. The strict feed gate still
> fails on 198 conditional rules. Other bytecode/parser parity, native
> Windows and integrated efficacy/resource acceptance remain open.

> **2026-10-27 compiler follow-up:** Incremental NDB reload now splices
> replacement rules into one owned table, removing a second full rule-table
> allocation. Two `GOMEMLIMIT=1000MiB` changed-rule Linux ARM64 probes
> allocated about 82 MB less median than the prior build, with overlapping
> cgroup peaks. Two no-target runs finished only 1.8–3.3 MB below a hard
> 2 GiB cap, leaving the resource gate open. Full tests, race, vet and
> Windows/Linux AMD64 builds pass. See the [v107 compiler audit](../../docs/edr/native-engine/audits/2026-10-27-v107-rule-splice-memory/REPORT.md).

> **2026-10-27 changed-rule memory follow-up:** Two current-source Linux
> ARM64 scanner-process reloads under a hard 2 GiB cap and
> `GOMEMLIMIT=1000MiB` passed with 408–521 MB headroom, no concurrent scan
> errors and last-good rollback. The same probe with no Go soft target was
> OOM-killed; forced GC alone left just 51 MB headroom. The
> [v106 receipts](../../docs/edr/native-engine/audits/2026-10-27-v106-reload-memory/REPORT.md)
> do not qualify the integrated service or changed signed CVDs. Full
> replacement remains gated by 198 conditional feed rules, native Windows
> runtime, independent malware efficacy and endpoint resource acceptance.

> **2026-10-27 source candidate `native-go-v105`:** Whole-file hash records
> now sort in their existing buffer during feed load. Three alternating
> macOS full-feed loads per build measured 76.6 MiB less median sampled
> peak heap and 73.2 MiB less cumulative allocation than v104, at a 16.1
> MiB retained-heap cost. Tests, race checks, vet, Windows/Linux AMD64
> builds, the 694-file safe corpus and 500 real-PE comparison pass. One
> offline scanner-only native Linux ARM64 run completed all 694 files under
> a hard 2 GiB limit with 160 MiB of measured headroom and no cgroup
> max/OOM events. See the [v105 memory audit](../../docs/edr/native-engine/audits/2026-10-27-v105-profile/REPORT.md).
> Integrated-service/reload memory, 198 conditional rules, native Windows
> runtime and independent malware efficacy remain open.

> **2026-10-27 source candidate `native-go-v104`:** CRB revocation matching
> now reads SignedData in the complete PE security directory even when
> `WIN_CERTIFICATE.length` is malformed. A real signed Microsoft DLL with
> a local block rule yields the same revoked-certificate alert as ClamAV
> 1.5.4 across six length variants. Trust exemption remains strict on those
> malformed wrappers; raw detections continue to run. The [v104 audit](../../docs/edr/native-engine/audits/2026-10-27-version-info-real-pe/REPORT.md)
> records the controlled divergence and reproducible evidence. Full EDR
> tests, focused race, vet, and Windows/Linux AMD64 cross-builds pass;
> 500 full-feed real PEs agree with pinned ClamAV and the 694-file safe
> corpus has zero verdict changes. The strict feed gate still fails on 198
> conditional rules; native Windows runtime and independent malware efficacy
> remain release gates.

> **2026-10-27 source candidate `native-go-v103`:** Full official-feed
> verdicts agree with pinned ClamAV 1.5.4 on 500 SHA-256-verified real PE
> files (500 clean/clean). The last two historically timestamped DLLs now
> have a CRB-anchored SHA-1 RSA certificate-link check; an inert marker
> rule confirms both engines exempt them. Full EDR tests, focused race,
> vet, Windows/Linux AMD64 cross-builds, and the 694-file safe corpus pass,
> with zero corpus verdict changes from v102. ClamAV 1.5.4 does not verify
> RFC3161 tokens; that was not the source of this gap. The strict feed
> still has 198 conditional unknown rules, and native Windows runtime and
> independent malware-efficacy acceptance remain open. See the
> [v103 signed-file audit](../../docs/edr/native-engine/audits/2026-10-27-version-info-real-pe/REPORT.md).

> **2026-10-27 source candidate `native-go-v102`:** An unmapped PE entry
> point now rejects the header as ClamAV 1.5.4 does. All 2,000 seeded PE
> header mutations and 3,500 independent VERSION_INFO key verdicts on 500
> SHA-256-verified real PE files agree with the pinned reference. Full EDR
> tests, focused race, vet, Windows/Linux AMD64 cross-builds, and the
> 694-file full-feed safe corpus passed; corpus verdicts are unchanged from
> v101. The strict feed gate has 198 conditional unknown rules (109 fuzzy
> image, 81 icon group, eight certificate trust). Native Windows and
> independent malware-efficacy acceptance remain open. See the
> [v102 VERSION_INFO audit](../../docs/edr/native-engine/audits/2026-10-27-version-info-real-pe/REPORT.md).
> A separate full-feed comparison at v102 found 441 conservative
> certificate-trust incompletes in Go while ClamAV left all 500 clean.
> The v103 work above resolves this measured difference.

> **2026-10-27 source candidate `native-go-v101`:** PE image and overlay
> analysis now handles the full 16-bit COFF section count. A 4,097-section
> PE carrying a stored ZIP overlay produces the same inert child-hash
> detection as ClamAV 1.5.4 on macOS and native Linux ARM64. Full EDR tests,
> focused race, vet, Windows/Linux AMD64 cross-builds, and an offline,
> hard-2-GiB Linux ARM64 run of all 694 safe-corpus files passed without a
> verdict change from v100. The strict feed gate still has 208 conditional
> unknown rules; native Windows and independent malware-efficacy acceptance
> remain open. See the
> [v101 PE overlay audit](../../docs/edr/native-engine/audits/2026-10-27-version-info-invalid/REPORT.md).

> **2026-10-27 source candidate `native-go-v100`:** Signature matching now
> supports the full 16-bit PE section count and resolves reference-rejected
> early headers as known false for relative rules. Four large-section PE
> fixtures agree with ClamAV 1.5.4; 2,459 header and 2,000 resource mutation
> cases have no false or missed reference detections, while malformed entry
> mappings remain conservative unknown. The 694-file safe corpus is
> unchanged from v99 on macOS and identical in a hard-2-GiB, offline local
> Linux ARM64 container. The separate overlay/certificate image parser
> retains its 4,096-section limit, the strict feed gate still has 208
> conditional unknown rules, and native Windows and independent malware-
> efficacy acceptance remain open. See the
> [v100 PE audit](../../docs/edr/native-engine/audits/2026-10-27-version-info-invalid/REPORT.md).

> **2026-10-27 source candidate `native-go-v99`:** The PE resource-tree
> walker now uses 16-bit wraparound for name and language entry counts,
> matching ClamAV 1.5.4. Two false detections in an 86-case resource
> comparison are fixed; all 541 single-byte resource mutations also agree.
> The 694-file safe corpus is unchanged from v98. The strict feed gate
> still has 208 conditional unknown rules, and native Windows and malware-
> efficacy acceptance remain open. See the
> [v99 VERSION_INFO audit](../../docs/edr/native-engine/audits/2026-10-27-version-info-invalid/REPORT.md).

> **2026-10-27 source candidate `native-go-v98`:** PE version-info matching
> now ignores a resource directory whose declared size is zero. Pinned
> ClamAV 1.5.4 leaves that inert `VI:` fixture clean; Go previously detected
> it. A 100-case PE header mutation comparison has no remaining false
> detections or missed reference detections, but 31 conservative unknowns.
> The 694-file safe corpus is unchanged from v97 and the strict feed gate
> still has 208 conditional unknown rules. See the
> [v98 VERSION_INFO audit](../../docs/edr/native-engine/audits/2026-10-27-version-info-invalid/REPORT.md).

> **2026-10-27 source candidate `native-go-v97`:** A malformed HTML base URL
> now matters only when a relative protected link needs it. This matches
> pinned ClamAV 1.5.4 on harmless mail parts with no links, an absolute benign
> link, and an absolute spoofed link. One of the 694 safe-corpus fixtures
> moves from incomplete to clean; the other 693 keep their verdicts. The
> strict feed gate remains at 208 conditional unknown rules. See the
> [v97 HTML base audit](../../docs/edr/native-engine/audits/2026-10-27-html-base/REPORT.md).

> **2026-10-27 source candidate `native-go-v96`:** The disk-image partition
> budget is configurable as `scan_limits.max_partitions` (0 selects the
> default 50; explicit values are limited to 1–1000). MBR primary, EBR
> logical, and GPT readers honor it; the overlap alert honors it too. Three
> generated later-partition hashes agree with ClamAV 1.5.4 at limits one
> and two. All 694 safe-corpus verdicts match v95. The strict feed gate
> remains at 208 conditional unknown rules; native Windows and malware-
> efficacy release gates remain open. See the
> [v96 partition-limit audit](../../docs/edr/native-engine/audits/2026-10-27-partition-limit/REPORT.md).

> **2026-10-27 source candidate `native-go-v95`:** An opt-in partition
> intersection alert detects overlapping MBR primary, extended-logical EBR,
> and GPT ranges. Six inert overlap/adjacency images match pinned ClamAV
> 1.5.4 with the alert enabled and disabled. Configure with
> `scan_heuristics.partition_intersection=true` and restart the service;
> the default is off. All 694 safe-corpus verdicts match v94. The strict
> official-feed gate remains at 208 conditional unknown rules; endpoint and
> malware-efficacy release gates remain open. See the
> [v95 partition audit](../../docs/edr/native-engine/audits/2026-10-27-partition-overlap/REPORT.md).

> **2026-10-27 source candidate `native-go-v94`:** The EBR reader now scans
> alternate first-two-entry layouts recognized by ClamAV 1.5.4. Four inert
> exact-partition hash detections agree with the reference; all 694 safe-
> corpus verdicts match v93. The strict feed gate remains at 208 conditional
> unknown rules. See the
> [v94 EBR audit](../../docs/edr/native-engine/audits/2026-10-27-ebr-order/REPORT.md).

> **2026-10-27 source candidate `native-go-v93`:** GPT primary and secondary
> headers and partition tables are CRC-checked and their bounded partitions
> scanned. Seven inert GPT detection paths agree with pinned ClamAV 1.5.4;
> the 694-file safe corpus remains identical to v92. The strict official-feed
> gate still has 208 conditional unknown rules. See the
> [v93 GPT audit](../../docs/edr/native-engine/audits/2026-10-27-gpt/REPORT.md).

> **2026-10-27 source candidate `native-go-v92`:** MBR boot sectors, primary
> partitions and conventional extended-partition chains now receive their own
> bounded hash/rule scans. Four inert MBR detection paths agree with pinned
> ClamAV 1.5.4. The 694-file safe corpus is unchanged from v91. GPT and
> unusual EBR layouts remain undecidable; the strict official-feed gate still
> has 208 conditional unknown rules. See the
> [v92 MBR audit](../../docs/edr/native-engine/audits/2026-10-27-mbr/REPORT.md).

> **2026-10-27 source candidate `native-go-v91`:** GIF fuzzy hashing now reads
> the first image's local palette to preserve RGB at transparent indices.
> Three generated hashes and an exact fuzzy-rule detect/clear pair agree
> with pinned ClamAV 1.5.4. All 694 safe-corpus verdicts match v90; the
> strict full-feed gate still has 208 conditional unknown rules. See the
> [v91 GIF audit](../../docs/edr/native-engine/audits/2026-10-27-gif-local-palette/REPORT.md).

> **2026-10-27 source candidate `native-go-v90`:** Multiple PE resource
> languages under one icon ID now yield each image to the icon matcher.
> Generated-PE hashes and a second-language IDB/LDB detection agree with
> pinned ClamAV 1.5.4; an earlier match also survives a later unreadable
> image. The 694-file safe corpus retains every v89 verdict;
> 208 conditional unknown feed rules and endpoint release gates remain.
> See the [v90 icon-language audit](../../docs/edr/native-engine/audits/2026-10-27-icon-languages/REPORT.md).

> **2026-10-27 source candidate `native-go-v89`:** 32-bit PE icons whose alpha
> bytes are all zero now use their AND mask, or become opaque if the mask is
> absent, as in ClamAV 1.5.4. Generated masked and maskless PEs' full icon
> metric hashes and matching icon-group verdicts agree with the pinned
> reference; a nonmatching IDB leaves the masked PE clean.
> All 694 generated safe-corpus outcomes remain identical to v88. The strict
> feed audit still has 208 conditional unknown rules, and endpoint/release
> gates remain open. See the
> [v89 icon audit](../../docs/edr/native-engine/audits/2026-10-27-icon-zero-alpha/REPORT.md).

> **2026-10-27 v88 changed-local-rule memory audit:** Four current-source
> scanner-only reloads under a hard 2 GiB Linux ARM64 cgroup kept the new
> harmless detection and retained the last-good matcher after a failed update.
> Two `GOMEMLIMIT=1000MiB` runs peaked at 1,707,061,248 and 1,714,757,632
> bytes with no limit hits; two unset runs hit the hard cap 325 and 406 times.
> This local `.ndb` test does not qualify changed official CVDs or the
> integrated endpoint. See the
> [reload audit](../../docs/edr/native-engine/audits/2026-10-27-v88-reload-memory/REPORT.md).

> **2026-10-27 v88 integrated-service memory audit:** An isolated Linux ARM64
> service loaded the frozen full feed and completed a 694-file scheduled scan
> in a hard 2 GiB cgroup. Both `GOMEMLIMIT=1000MiB` and unset runs matched
> all prior Go scanner outcomes (296 detected, 374 clean, 24 incomplete) and
> emitted 296 quarantine events. Their cgroup peaks were 1,291,272,192 and
> 2,072,190,976 bytes respectively, with no OOM. The unset run had only
> 75,292,672 bytes spare; live sensors and changed-feed reload were not
> exercised, and the earlier changed-feed 2 GiB OOM gate remains. See the
> [service audit](../../docs/edr/native-engine/audits/2026-10-27-v88-service-memory/REPORT.md).

> **2026-10-27 source candidate `native-go-v88`:** The escaped-name PDF
> heuristic is default on and can be disabled with
> `scan_heuristics.pdf_obfuscated_name=false` after service restart. The policy
> persists and has a distinct verdict-cache identity. The on/off behavior
> agrees with pinned ClamAV 1.5.4 on a generated PDF; the default-policy
> 694-file corpus is verdict-identical to v87. Other heuristic/category
> controls and full-replacement gates remain open. See the
> [v88 audit](../../docs/edr/native-engine/audits/2026-10-27-v88-heuristic-policy/REPORT.md).

> **2026-10-27 source candidate `native-go-v87`:** A full 694-file generated
> safe-corpus comparison with pinned ClamAV 1.5.4 exposed one PDF heuristic
> detection missing in Go. Hex-escaped spellings of ClamAV's common PDF
> dictionary names now produce `Heuristics.PDF.ObfuscatedNameObject` after
> the object and streams are analyzed; direct ClamAV checks confirm the alert
> also takes precedence over a failed Flate stream or missing PDF EOF and is
> raised for an escaped name in a dictionary comment. This
> changed only that fixture's verdict in the full corpus; all 259
> reference detections now have Go detections in this corpus. The fixture is
> benign, so this heuristic needs explicit response-policy consideration.
> Strict conditional rules, broader bytecode/unpacker/format parity, native
> Windows runtime and independent malware-corpus efficacy remain open.
> See the [v87 audit](../../docs/edr/native-engine/audits/2026-10-27-v87-694/REPORT.md).

> **2026-10-27 `native-go-v86` memory follow-up:** Signature finalization no
> longer clones the large pattern and rule slices. Under a hard 2 GiB ARM64
> container limit with AMD64 emulation, the official-feed scanner peaked at
> 2,124,951,552 bytes without a Go heap target and 1,747,640,320 bytes with
> a 1000 MiB target. The native ARM64 build peaked at 1,889,181,696 and
> 1,312,514,048 bytes under the same settings. All eight runs had identical
> verdicts on the pinned 195-file safe corpus. These single scanner-only
> runs do not qualify integrated service memory or the
> remaining release gates. See the
> [memory audit](../../docs/edr/native-engine/audits/2026-10-27-v86-memory/REPORT.md).

> **2026-10-27 source candidate `native-go-v86`:** 16-bit BMP icons embedded
> in PE resources now use ClamAV 1.5.4's color-word and AND-mask conversion.
> A generated PE's icon hash and matching/nonmatching IDB verdicts agree with
> the pinned reference. Other icon extraction shapes and endpoint gates remain
> open. See the
> [v86 audit](../../docs/edr/native-engine/audits/2026-10-27-icon16/REPORT.md).

> **2026-10-27 source candidate `native-go-v85`:** Certificates from fully
> verified `.cat` sources that independently chain to CRB trust are now
> available to later PE Authenticode checks. A generated PE whose signer
> omits an intermediate is exempted only with the trusted catalog; changed
> code and unverified catalog certificates do not gain trust. Native Windows
> validation and other release
> gates remain open. See the
> [v85 audit](../../docs/edr/native-engine/audits/2026-10-27-catalog-chain/REPORT.md).

> **2026-10-27 source candidate `native-go-v84`:** A trusted CRB RSA key can
> now anchor a PE signer's embedded certificate when the corresponding root
> certificate is omitted from SignedData. The full PKCS#7, signer-path and PE
> image-digest checks still apply. A generated root-omitted PE matches pinned
> ClamAV 1.5.4 in detect/exempt/changed-code outcomes. The same verifier also
> serves signed catalog loading. Other trust-chain shapes, the 208 conditional
> feed unknowns and endpoint release gates remain open. See the
> [v84 audit](../../docs/edr/native-engine/audits/2026-10-27-crb-rootless/REPORT.md).

> **2026-10-27 source candidate `native-go-v83`:** The compiler verifies
> signed `.cat` sources against CRB trust and loads authenticated PE image
> and CAB whole-file hashes as separate allowlist namespaces. The scan path
> uses them after CRB blocklist checks. Synthetic tests cover tampering,
> changed members, audit/limit accounting and CVD reloads. The pinned feed
> has no catalog and still fails strict runtime with 208 conditional unknown
> rules. A byte-pinned OpenVPN release validates real signed CTL parsing and
> an exact catalog-to-PE image hash at a signer-valid historical time; its
> RFC3161-only timestamp does not extend the expired signer's trust today.
> Pinned ClamAV 1.5.4 also denies catalog trust today; it continues scanning
> with other rules, whereas Go rejects that feed generation.
> A separate, currently valid signed DsHidMini catalog matches pinned ClamAV
> in three end-to-end outcomes: detect without catalog, exempt its covered
> driver, and detect changed code. Native Windows acceptance, independent
> malware-corpus efficacy and resource qualification remain
> open. See the [v83 catalog audit](../../docs/edr/native-engine/audits/2026-10-27-authenticode-catalog/REPORT.md).

> **2026-10-27 source candidate `native-go-v82`:** Legacy PKCS#9
> Authenticode countersignatures now authenticate an expired PE code signer
> only when the timestamp RSA signature, binding to the primary signature,
> signing time and CRB-trusted time-signing chain verify. Altered signatures,
> changed code, an untrusted timestamp signer and invalid signing times do
> not receive a trust exemption. The official-feed strict-runtime audit
> still fails with 208 conditional unknown rules. RFC3161 tokens, catalogs,
> pinned-ClamAV and native Windows runtime parity and independent malware
> efficacy remain open. See the [v82 timestamp audit](../../docs/edr/native-engine/audits/2026-10-27-authenticode-legacy-time/REPORT.md).

> **2026-10-27 source candidate `native-go-v81`:** A PE now receives a CRB
> trust exemption only when its PKCS#7 signed attributes and signer signature
> verify, its signer chains to a matching CRB trust certificate for code
> signing at scan time, and its embedded image digest matches the PE regions
> outside the checksum and certificate table. Blocklist matches take
> precedence. Signed synthetic PE tests cover direct and chained trust,
> tampering, expiry, unrelated embedded trust certificates and unsigned
> certificate padding. The pinned-feed strict-runtime gate still fails with
> 208 conditional unknown rules, including eight certificate trust rules:
> timestamps, catalogs, nested signatures and unsupported algorithms remain
> open. Three byte-verified Microsoft Sigcheck PEs passed opt-in Go trust and
> tamper regressions with test-generated CRB rules; pinned-ClamAV and native
> Windows runtime parity remain open. See the [v81 trust audit](../../docs/edr/native-engine/audits/2026-10-27-authenticode-trust/REPORT.md).

> **2026-10-27 source candidate `native-go-v80`:** The PE scan path now
> reads Authenticode `WIN_CERTIFICATE` entries, parses their PKCS#7 SignedData
> certificate set and emits the matching `.crb` blocklist rule as a malware
> verdict. The pinned official feed's 22 blocked identities are evaluated;
> its eight trust identities remain unknown because the signed PE digest and
> certificate chain are not yet verified. Synthetic signed-data tests cover
> revoked certificates, multiple entries, unrelated content types and
> unreadable certificates. The full-feed strict-runtime audit falls from 230
> to 208 unknown rules, still fails, and has zero rejected rules. See the
> [v80 Authenticode blocklist audit](../../docs/edr/native-engine/audits/2026-10-27-authenticode-blocklist/REPORT.md).

> **2026-10-27 source candidate `native-go-v79`:** The compiler now retains
> selected CRB certificate identities and trust/blocklist policies in typed,
> immutable rules. The pinned official `daily.crb` has 22 blocklist entries
> and eight trust entries. Identity and policy comparison are tested, but PE
> Authenticode parsing, signed-digest/chain verification and scan integration
> remain open; all 30 rules still count as runtime unknown. The previous
> description that ignoring CRB could only add detections was incorrect:
> blocklist rules can detect revoked signing certificates. See the
> [v79 certificate-rule audit](../../docs/edr/native-engine/audits/2026-10-27-cert-rules/REPORT.md).

> **2026-10-27 source candidate `native-go-v78`:** Multiple existing CVD/CLD
> sources can now reload incrementally in source order while keeping each
> intermediate generation private. A controlled simultaneous change to pinned
> official `daily.cvd` and `bytecode.cvd` takes this path, matches fresh full
> compilation and all 694 safe-corpus outcomes. Two hard 2 GiB Linux runs
> with `GOMEMLIMIT=1500MiB` peaked at 1.969 and 2.052 GB with zero cgroup
> ceiling/OOM events or concurrent-scan errors. The repeat left only 96 MB
> of headroom. The [v78 audit](../../docs/edr/native-engine/audits/2026-10-27-cvd-multi-source-refresh/REPORT.md)
> records exact evidence and remaining gates. Full ClamAV replacement
> remains unqualified.

> **2026-10-27 source candidate `native-go-v77`:** An existing CVD/CLD can
> now reload changed CBC members with unchanged members and sources shared.
> Logical bytecode rules, PDF hooks, unsupported-hook state and PE-hook order
> are reconstructed for the new generation. A one-member derivative of pinned
> official `bytecode.cvd` takes this path, matches fresh full-compilation
> reports and all 694 safe-corpus outcomes. Two hard 2 GiB Linux runs with
> `GOMEMLIMIT=1500MiB` peaked at 1.733 and 1.731 GB with no OOM or scan
> errors. The [v77 audit](../../docs/edr/native-engine/audits/2026-10-27-cvd-bytecode-refresh/REPORT.md)
> records the exact evidence and remaining gates. Full ClamAV replacement
> remains unqualified.

> **2026-10-27 source candidate `native-go-v76`:** Repeated NDB/LDB source
> replacement now retires unreferenced compiled pattern bodies and reuses
> their stable slots without changing the serving generation. A two-step
> pinned `daily.cvd` derivative agrees with a fresh full compiler; the second
> update agrees on all 694 safe-corpus outcomes. Two hard 2 GiB Linux runs
> with `GOMEMLIMIT=1500MiB` peaked at 2.026 and 1.946 GB, without OOM or
> concurrent-scan errors. The [v76 audit](../../docs/edr/native-engine/audits/2026-10-27-pattern-slot-reuse/REPORT.md)
> records the receipts and remaining gates. The smaller headroom was about
> 122 MB. The [full-feed audit](../../docs/edr/native-engine/audits/2026-10-27-v76-feed-audit/REPORT.md)
> accepts all selected rules but finds 230 that can remain unknown when
> reached; `sigaudit -strict-runtime` fails until those are resolved.
> Production memory policy, bytecode/selection member updates,
> native Windows runtime and full ClamAV parity remain open.

> **2026-10-27 source candidate `native-go-v75`:** An existing CVD/CLD can
> reload with mixed changed NDB, LDB, HDB and HSB members while sharing the
> unchanged compiled generation. A deterministic three-member update to
> pinned `daily.cvd`, including its 76.8 MB logical database, agrees with a
> fresh full compiler and all 694 safe-corpus verdicts. Two hard 2 GiB Linux
> runs with `GOMEMLIMIT=1500MiB` peaked at 1.933 and 1.904 GB, with no OOM or
> concurrent-scan errors. The [v75 audit](../../docs/edr/native-engine/audits/2026-10-27-cvd-logical-refresh/REPORT.md)
> records the validation and remaining gates. Changed bytecode and selection
> members, production memory policy, native Windows runtime and full ClamAV
> parity remain open.

> **2026-10-27 source candidate `native-go-v74`:** One existing CVD/CLD can
> reload with mixed changed NDB and HDB/HSB members while retaining unchanged
> compiled rules and ordered hash-member tables. A deterministic official
> `daily.cvd` derivative changing its real `daily.ndb` and `daily.hsb` members
> agrees with fresh full-compilation reports, both new harmless detections and
> all 694 safe-corpus verdicts. The [v74 audit](../../docs/edr/native-engine/audits/2026-10-27-cvd-mixed-refresh/REPORT.md)
> records a 1.760 GB hard-2-GiB Linux cgroup peak with `GOMEMLIMIT=1500MiB`,
> zero ceiling events and zero errors over 228 concurrent scans. Other member
> formats, production memory policy, native Windows runtime and full ClamAV
> parity remain open.

> **2026-10-27 source candidate `native-go-v73`:** Multiple changed NDB
> members of one existing CVD/CLD now reload in archive order while unchanged
> members and sources remain shared. A deterministic two-member official-feed
> derivative matches fresh full-compilation reports and both new harmless
> detections; all 694 safe-corpus outcomes agree. A hard 2 GiB Linux probe
> with `GOMEMLIMIT=1500MiB` peaked at 1.821 GB with zero ceiling events and
> zero errors over 206 concurrent scans. The [v73 audit](../../docs/edr/native-engine/audits/2026-10-27-cvd-multi-ndb/REPORT.md)
> records the synthetic fixture and limitations. Mixed NDB/hash changes,
> production memory policy, native Windows runtime and full ClamAV parity
> remain open.

> **2026-10-27 source candidate `native-go-v72`:** One NDB member of an existing
> CVD/CLD can reload without recompiling unchanged members, after validating
> archive integrity, member order and source bytes. A controlled one-member
> change in pinned `daily.cvd` agrees with a full compiler and all 694
> official-plus-local safe-corpus outcomes. The [v72 audit](../../docs/edr/native-engine/audits/2026-10-27-cvd-member-refresh/REPORT.md)
> records a 1.865 GB hard-cgroup peak for the final build with `GOMEMLIMIT=1500MiB`; unconfigured
> runs varied from the 2 GiB cap to about 124 MiB below it. Multi-member official feed
> rotation, production memory policy, Windows runtime and full ClamAV parity
> remain open.

> **2026-10-27 source candidate `native-go-v71`:** One existing top-level
> `.ndb` source can reload incrementally after validating every source byte.
> Rule order, shared patterns, active anchor candidates and last-good rollback
> are preserved. The pinned full-feed compiler and 694-file safe-corpus
> comparisons agree with a fresh compile. The [v71 audit](../../docs/edr/native-engine/audits/2026-10-26-ndb-refresh/REPORT.md)
> records two hard-2-GiB Linux OOM kills without a Go soft memory limit and
> one exact-v71 run at the cap, versus three passes with `GOMEMLIMIT=1500MiB`.
> Production memory policy, changed CVD
> reload, Windows native acceptance and full ClamAV parity remain open.

> **2026-10-26 source candidate `native-go-v70`:** Existing changed `.hdb` and
> `.hsb` sources can reload incrementally after byte validation of all inputs.
> Ordered immutable hash-source tables share the unchanged rule/bytecode
> matcher and preserve exact-size and first-source match precedence. The
> [v70 audit](../../docs/edr/native-engine/audits/2026-10-26-hash-refresh/REPORT.md)
> records full-feed compiler equality and hard 2 GiB Linux measurements. The
> larger-feed unconfigured memory margin, changed `.ndb`/CVD reload at v70, Windows
> acceptance and full ClamAV parity remain open.
> [Rule-source ranges, retained pattern keys and the v70 safe-corpus audit](../../docs/edr/native-engine/audits/2026-10-26-rule-provenance/REPORT.md)
> are retained for the next segmented-matcher step; 694 outcomes are unchanged.

> **2026-10-26 source candidate `native-go-v69`:** The ISO scanner now
> scans flagged multi-extent sections individually, matching the pinned
> ClamAV 1.5.4 reader's section behavior. Both engines detect the same
> harmless section hashes; the Go scanner conservatively reports incomplete
> if no section detects because it does not assemble the whole logical file.
> All 694 full-feed safe-corpus outcomes remain unchanged from v68. See the
> [v69 ISO audit](../../docs/edr/native-engine/audits/2026-10-26-iso-multiextent/REPORT.md).
> Full replacement and native endpoint acceptance remain open.

> **2026-10-26 `native-go-v68` Linux memory acceptance:** A real Linux ARM64
> cgroup OOM-killed the changed-feed reload under a hard 2 GiB cap; the same
> workload completed under 4 GiB at 3.31 GiB peak. A 1500 MiB Go soft limit
> passed three repeats at the hard ceiling but also failed once, so it does
> not close the P0 gate. Measured hash-index and compiler capacity-hint
> trials were reverted after resident-memory comparisons.
> See the [cgroup audit](../../docs/edr/native-engine/audits/2026-10-26-hash-freeze-memory/REPORT.md)
> for exact receipts and the required structural compiler work. Full
> replacement, Windows endpoint acceptance and independent malware efficacy
> remain open.

> **2026-10-26 source candidate `native-go-v68`:** Bounded cooked and raw
> ISO 9660 and Joliet directory trees now feed contained files and member
> metadata rules into the native scanner. The pinned ClamAV 1.5.4 oracle and
> Go agree on eighteen harmless child-hash positives, a clean control and a
> metadata rule.
> All 694 full-feed safe-corpus outcomes remain unchanged from v67. The
> [ISO audit](../../docs/edr/native-engine/audits/2026-10-26-iso9660/REPORT.md)
> records limits and remaining disc-image layouts. Full replacement and
> current native endpoint acceptance remain open.

> **2026-10-26 source candidate `native-go-v67`:** The chunked HTML normalizer
> now reproduces encoded JScript/VBScript in script tags and ASP directives,
> including ASP's second pass over a chunk. The page, text and script views
> match 210 ClamAV 1.5.4 oracle pages byte for byte; both engines detect
> encoded-only rules. Supported pages can now be clean. The 694-file safe
> corpus has no changed outcome from v66. The
> [audit](../../docs/edr/native-engine/audits/2026-10-26-html-screnc/REPORT.md)
> records bounds and remaining malformed cases. Full replacement and native
> endpoint acceptance remain open.

> **2026-10-26 source candidate `native-go-v66`:** The Go HTML script view now
> decodes bounded Dean-Edwards six-argument packer calls for validated named
> and anonymous forms. A deterministic 152-case ClamAV 1.5.4 oracle matches
> normalized bytes; both engines detect a decoded-only rule and agree on a
> clean control. The 694-file safe corpus has no changed outcomes from v65.
> See the [packer audit](../../docs/edr/native-engine/audits/2026-10-26-js-packer/REPORT.md).
> Other packer shapes and mixed decoder ordering remain incomplete. Full
> replacement and native endpoint acceptance remain open.

> **2026-10-26 source candidate `native-go-v65`:** Bounded Go code now decodes
> standalone Microsoft Script Encoder files and scans their generated HTML
> child. Encoded scripts embedded in HTML can yield a decoded-body detection,
> while the unverified enclosing page views still report incomplete on a
> non-detection. Nine ClamAV 1.5.4 output vectors match byte for byte and an
> inert decoded-only signature detects in both engines. All 694 retained safe
> files have the same outcome as v64. See the
> [audit](../../docs/edr/native-engine/audits/2026-10-26-screnc/REPORT.md).
> Full feature parity and native endpoint acceptance remain open.

> **2026-10-26 source candidate `native-go-v64`:** The HTML script view now
> decodes bounded ScriptAsylum `dF` bodies. Thirty-two generated ClamAV 1.5.4
> normalizer vectors match byte for byte, and both scanners detect an inert
> marker present only in decoded text. A 694-file safe-corpus comparison changed
> four previously incomplete files to clean, all four of which ClamAV also
> reports clean. The [audit](../../docs/edr/native-engine/audits/2026-10-26-df-decoder/REPORT.md)
> records the exact inputs, checks and remaining encoded-script gaps. Full
> replacement and native endpoint acceptance remain open.

> **2026-10-26 post-v63 memory follow-up:** A byte-verified fast path reuses
> the existing compiled matcher when loose signature files are unchanged;
> unchanged signed generations without local overlays now stream-verify their
> content instead of retaining another complete source copy.
> Three macOS resident full-feed repeats reduced median unchanged-feed reload
> time from 5.601 to 0.049 seconds and median sampled scan-overlap Go heap
> peak from 3.432 to 1.852 GiB, with identical outcomes for all 694 safe
> files. The [audit](../../docs/edr/native-engine/audits/2026-10-26-memory-compact/REPORT.md)
> also records compact storage for the 1,096,438 literal-only patterns in the
> official feed. In three changed-feed runs, sampled Go-heap peaks were
> 2.685–2.686 GiB, versus 3.494 GiB in one pre-change run, with zero
> changed outcomes on the 694-file safe corpus. Actual macOS peak RSS was
> still 3.223 GiB normally and 2.606 GiB under a 2 GiB Go soft limit;
> genuinely changed generations remain a P0 memory blocker. No new native
> endpoint acceptance is claimed.

> **2026-10-26 source candidate `native-go-v63`:** The
> [corrected large-JPEG audit](../../docs/edr/native-engine/audits/2026-10-26-jpeg-large/REPORT.md)
> isolates the one known large-JPEG hash gap to Lanczos3 resize rounding,
> rather than decoded pixels. The guarded exact-confidence profile now covers
> tested JFIF and Adobe CMYK images up to 2048 pixels per side and one million
> total pixels. Two generated probes totaling 855 cases had no wrong certain
> bits; four ClamAV CLI decisions and a 694-file safe-corpus regression agree.
> Exact large-image decisions in the retained set are flat images; nonflat
> cases often remain uncertain. Full replacement is still unverified.

> **2026-10-26 source candidate `native-go-v62`:** The tested JFIF and Adobe
> CMYK JPEG fuzzy-image path now supports restart intervals. The
> [restart audit](../../docs/edr/native-engine/audits/2026-10-26-jpeg-restart/REPORT.md)
> records the 33 retained cases, a 1,191-file broader reference differential
> and four matching/nonmatching CLI decisions. Five retained final-DCT
> boundary cases remain uncertain. The
> [large-JPEG audit](../../docs/edr/native-engine/audits/2026-10-26-jpeg-large/REPORT.md)
> found one mismatch in 285 images. Its decoded-pixel diagnosis and 512-pixel
> cap were superseded by the v63 correction above.

> **2026-10-26 source candidate `native-go-v61`:** A Go-owned JPEG decoder
> supports exact fuzzy-image rule decisions for tested ordinary JFIF and
> subsampled Adobe CMYK layouts
> through 512×512 except at floating-point hash boundaries. The
> [JPEG audit](../../docs/edr/native-engine/audits/2026-10-26-jpeg-native/REPORT.md)
> records 369/369 exact decoded hashes in the pinned mixed-format oracle,
> 1,182/1,191 exact hashes in a wider grayscale/RGB probe with all nine
> differences safely uncertain, 348/348 exact hashes in an independent CMYK
> probe, four ClamAV CLI rule comparisons, and an unchanged
> 694-file synthetic full-feed regression. Other JPEG layouts and the wider
> engine and endpoint acceptance gates remain open.

> **2026-10-26 source candidate `native-go-v60`:** The
> [icon completion audit](../../docs/edr/native-engine/audits/2026-10-26-icon-edge/REPORT.md)
> records exact hashes for 187 retained PE icon oracles, 2,590 exact
> match/nonmatch decisions, two fresh ClamAV 1.5.4 CLI cases, and an unchanged
> 694-file synthetic full-feed regression. Parsed icons now have all edge,
> noedge and grayscale measures; unmodeled resource extraction remains
> conservative. Prior icon entries below describe their historical versions.

> **2026-10-26 feed acceptance:** A new
> [ThreatWinds staging command](cmd/feedstage/README.md) and the existing
> `feedpack` produced a locally signed combined generation from pinned official
> CVDs and five nonempty public hash feeds. The
> [live acceptance report](../../docs/edr/native-engine/audits/2026-10-26-feed-live/REPORT.md)
> records two removed overlaps, 149,505 retained indicators, a verified test
> signature and identical official-only versus merged scan outcomes on 694
> synthetic scenarios. It does not close production delivery or engine parity.

> **2026-10-26 source candidate:** `native-go-v59` decides JPEG 2000
> fuzzy-image rules false where ClamAV 1.5.4 cannot hash the file. The
> [type oracle](../../docs/edr/native-engine/audits/2026-10-26-image-types/REPORT.md)
> also shows that WebP is not reached by image rules in this reference.
> Its v58 predecessor extends exact fuzzy-image
> decisions to BMP, TIFF and PNG through 512×512 pixels. A
> [1,000-image differential and 48 CLI cases](../../docs/edr/native-engine/audits/2026-10-26-large-image/REPORT.md)
> match ClamAV 1.5.4. Its v57 predecessor separates float32 image
> luma and Lanczos3 products from their additions. Its
> [arithmetic audit](../../docs/edr/native-engine/audits/2026-10-26-luma-fma/REPORT.md)
> records 6,000 additional exact lossless image hashes and 74/74 positive,
> 74/74 negative CLI comparisons against ClamAV 1.5.4. Full replacement
> remains pending other semantics and native endpoint validation.
> Its v55 predecessor uses float32 image luma and
> resize arithmetic, narrows the rounding guard for small lossless images and
> bounds image dimensions. A [1,000-image hash oracle and 148 CLI cases](../../docs/edr/native-engine/audits/2026-10-26-image-rounding/REPORT.md)
> give 55 of 74 positive detections, 19 incomplete positives and 74 clean
> negative controls. Its v54 predecessor types and decodes BMP for
> fuzzy-image rules. It computes the same hash as ClamAV on 60 harmless BMPs,
> but only 13 exact-rule CLI cases detect; 47 remain incomplete. See the
> [BMP audit](../../docs/edr/native-engine/audits/2026-10-26-bmp-fuzzy/REPORT.md).
> Its v53 predecessor decodes classic TIFF for
> fuzzy-image rules. Its 142 harmless reference images have identical computed
> hashes, but in 14 CLI probes only five yield detections; nine stay incomplete because of
> conservative uncertain bits. The [TIFF audit](../../docs/edr/native-engine/audits/2026-10-26-tiff-fuzzy/REPORT.md)
> records the exact scope. Its v52 predecessor rejects nonzero
> `fuzzy_img` Hamming distances as ClamAV 1.5.4 does; the
> [distance audit](../../docs/edr/native-engine/audits/2026-10-26-fuzzy-distance/REPORT.md)
> records the comparison. Its v51 predecessor implements bounded
> HandlerType whole-object rescans for all 11 selected official rules. SCRIPT
> normalization and ZIP member scanning precede raw rules. The pinned feed's
> conditional unknown count is 230. The [HandlerType audit](../../docs/edr/native-engine/audits/2026-10-26-handler-type/REPORT.md)
> records nine ClamAV 1.5.4 differential cases. Its v47 predecessor compiles all 185 official
> `daily.wdb` allowlist rules with no conditional unknowns. The
> [WDB X audit](../../docs/edr/native-engine/audits/2026-10-26-wdb-x/REPORT.md)
> and [WDB M audit](../../docs/edr/native-engine/audits/2026-10-26-wdb-m/REPORT.md)
> record 16 differential mail cases and the full-feed rule count. Its
> `native-go-v44` predecessor answers ClamBC engine scan
> and database option queries from an explicit native scanner profile. A pinned
> upstream debug program now returns its published result. The [option API audit](../../docs/edr/native-engine/audits/2026-10-26-bytecode-options/REPORT.md)
> records the verification. Its `native-go-v43` predecessor scans bounded Flate
> output from damaged PDF streams and recovers a stray carriage return or a
> malformed first line before zlib data. The [PDF Flate audit](../../docs/edr/native-engine/audits/2026-10-26-pdf-flate/REPORT.md)
> records the source comparison and tests. Its `native-go-v42` predecessor adds NsPack PE unpacking and
> recursive child scanning, checked against ClamAV 1.5.4 on 120 harmless
> stream mutations. The [NsPack differential](../../docs/edr/native-engine/audits/2026-10-26-nspack/REPORT.md)
> records the evidence. Its v41 predecessor added FSG 2.0 and 1.33 PE
> unpacking and recursive scanning of their rebuilt children, verified against
> ClamAV 1.5.4 on five harmless encoded programs. The [FSG differential](../../docs/edr/native-engine/audits/2026-10-26-fsg2/REPORT.md)
> records the bytes and limits. The earlier v40 change adds bounded ClamBC API 89
> (`matchicon`) handling that reproduces ClamAV 1.5.4's source-proven
> no-match behavior for valid calls and `-1` for invalid calls. It also includes
> the verified binary byte-comparison profile and bounded recursive scanning
> of bytecode-produced output. The [matchicon differential](../../docs/edr/native-engine/audits/2026-10-26-matchicon/REPORT.md),
> [byte-comparison audit](../../docs/edr/native-engine/audits/2026-10-26-bytecompare/REPORT.md),
> [v36 extraction gate](../../docs/edr/native-engine/audits/2026-10-26-extract-gate/REPORT.md)
> [v37 extraction audit](../../docs/edr/native-engine/audits/2026-10-26-extraction/REPORT.md)
> and [v38 pending-output differential](../../docs/edr/native-engine/audits/2026-10-26-pending-output/REPORT.md)
> record the bounded validation. A pending extracted child is scanned even
> when later VM execution fails. Other unpacker output, native Windows
> acceptance and production delivery remain unverified for this candidate.

> **2026-10-21 round6 current:** Native engine `native-go-v35`, package
> `12.0.5-native`. Read the [current report](../../docs/edr/native-engine/audits/2026-10-21-round6/REPORT.md)
> and [capability matrix](../../docs/edr/native-engine/audits/2026-10-21-round6/CAPABILITY_MATRIX.md).
> The approved private cloud tests executed: Windows x64 has24 component/client
> groups plus nine SCM checks passing; Linux has233 top-level native passes,
> one helper skip,84 packet assertions and five service cases passing. Linux
> owner/DHCP/nft-wire and response/startup repairs are in source. Final host
> regression has2,508 pass events/21skips. All owned cloud resources are removed.
> Peak memory remains materially higher than ClamAV:2.125vs1.090GiB mixed median
> RSS,4.606GiB loose-CVD rebuild median; Go OOMs at1/2GiB in the bounded child
> diagnostic. Full AV/EDR parity and signed production delivery remain incomplete.
> The repository already scans in Go; no production rollout or Git mutation
> occurred. Earlier dated checkpoint claims below are historical.


> **2026-10-21 round5:** Current scanner is `native-go-v35`, package
> `12.0.4-native`. Read the [current report](../../docs/edr/native-engine/audits/2026-10-21-round5/REPORT.md)
> and [capability matrix](../../docs/edr/native-engine/audits/2026-10-21-round5/CAPABILITY_MATRIX.md).
> Bounded RTF text views, signature-anchor lifetime cleanup, strict whole-cycle
> network feeds, native host probing and Windows ETW lifecycle/capacity repairs
> are implemented. Retained signature heap improves 5.14%; diagnostic peak RSS
> does not improve. New Windows revision9 remains compiled-only. GCP labs are
> removed; exact private transfer approval is pending after automatic review
> rejection. Full parity and production delivery remain unproven. Historical
> tables below retain their original artifact and acceptance scopes.


Updated 2026-10-21.

The endpoint now calls an in-process Go scanner. There is no `clamd`, `freshclam`, scan socket, external scanning executable, CGo dependency, or linked scanning library in the new analysis/update path. The existing file watcher, launch guard, quarantine, AMSI, ransomware guard, network blocking and event relay remain connected to the scanner.

**This is an experimental replacement with explicitly limited compatibility, not a completed feature-equivalent replacement for ClamAV. Do not deploy it as an equivalent production protection layer.** The full-replacement objective remains open. The complete official signature database now loads, but some of its rules need analysis this engine does not perform yet; they load as "unknown when reached" and make a result incomplete wherever they could apply, so no rule is dropped silently. Analysis errors, unsupported or unreadable encrypted analysis and resource limits return incomplete/error results and are not cached as clean. Some deliberately reference-compatible parser skips still return clean (including the measured missing-EOCD/truncated-ZIP and unknown-PDF-filter probes); clean does not mean every possible decoder ran. Absence of a match is not proof a file is safe; unknown formats and packers can only receive raw signature coverage.

## Historical round4 validation addendum (2026-10-21)

Current engine identity is `native-go-v34`, executable distribution `12.0.3-native`. The [current audit](../../docs/edr/native-engine/audits/2026-10-21-round4/REPORT.md) covers scoped native Linux prevention, kernel process events/inventory, identity-checked canary containment, leased nftables/isolation, resource bounds and actual Windows SYSTEM API/service acceptance. Standard-user Windows, repaired ETW, general system-volume prevention and production trust remain open gates. The [v34 scanner/memory record](../../docs/edr/native-engine/audits/2026-10-21-round3/REPORT.md) retains the previous full-feed evidence; full ClamAV or vendor parity is not established.

Go detects 291/296 safe positives versus ClamAV 258/296, with six gains from v32 and no dropped reference positive. Go raises zero alerts on 366 generated benign controls, leaving six incomplete. Both engines return all 256 verified system controls clean. Peak Go RSS remains higher, and the timing result reverses between fixture and system-file populations. These results do not establish real-malware recall, universal format coverage, Windows runtime acceptance or vendor parity. Earlier current-state paragraphs below are retained as dated history.

### Previous v32 checkpoint

See the [v32 audit and validation](../../docs/edr/native-engine/audits/2026-10-20/OPENEDR_AUDIT.md), [exact fixture results](../../docs/edr/native-engine/audits/2026-10-20/FIXTURE_RESULTS.md), [memory analysis](../../docs/edr/native-engine/audits/2026-10-20/MEMORY_ANALYSIS.md) and [configuration guide](../../docs/edr/native-engine/configuration.md). That checkpoint's identity was `native-go-v32`; earlier version/measurement narrative below is historical. ZIP64/LZMA/Zstandard, HTML data envelopes, typed mailboxes, PDF limit propagation, effective scan policies and compiler memory were improved. Full regression, focused races, official-feed smoke, bounded fuzz and Windows Go cross-builds passed; optional historical corpora were not silently counted as rerun.

On 307 safe files, Go detects 107/108 positives vs 94/108 ClamAV 1.4.3; both flag 0/191 generated benign controls, with 2 Go benign incompletes. All reference detections match. Median process RSS 2.04 GiB remains above ClamAV 0.95 GiB. Full replacement remains open; five-count database discrepancy is now reconciled. See the current audit for exact methods, bytes, hashes and remaining limits rather than inferring production efficacy from these fixtures.

## What changed

- `engine`: bounded recursive Go scanning and immutable signature generations. It loads a local database; scans take one generation snapshot and return that generation's identity with the verdict. Signature identity includes a SHA256 content digest, implementation version and scan limits.
- `signatures`: strict compilation of a defined subset of ClamAV text signature languages, indexed file hashes, literal candidate indexing, bounded pattern matching, logical combinations and bounded CVD/CLD unpacking.
- `scanner`: hashes the same bounded file bytes that it scans; cached verdicts must match the active engine generation. Errors/unknown outcomes are never cached as clean. Revalidates bytes before quarantine and emits deduplicated incomplete-analysis health events.
- `amsi`: buffers go directly to the native engine. Existing fail-open/fail-closed policy remains effective for unavailable/incomplete analysis.
- `feed`: Go HTTPS downloader; pinned Ed25519 manifest, expiry, sequence anti-rollback, SHA256 artifact checks, explicit hash-list adapters, all-or-nothing compilation/activation and cache invalidation. Failed updates retain the working generation.
- `cmd/feedpack`: offline validator and signed feed publisher. It validates the complete prospective database before creating a new release directory. Signing keys are never copied into published output.
- `engine-check`: read-only JSON database validation, including rules unknown when reached by reason and the representations not produced yet. Status includes `engine_implementation`, `engine_compatibility`, `signature_count`, `signature_unknown_rules`, `signature_unknown_by_reason` and `signature_update_error`.
- `native-go-v6`: archive/document extraction (including CAB stored/MSZIP, PDF LZW/predictors, TNEF attachments/RTF compression and uuencoded payloads), phishing checks and the supported logical-bytecode callbacks described below. The implementation identity changed, so earlier cached verdicts are invalidated.
- `native-go-v7`: rules are selected at functionality level 216 with ignore lists and dynamic configuration (see below), so the same database files can load different rules; the identity changed again. Identical subsignatures share one compiled pattern and one count per object, and the literal index covers case-insensitive and ASCII-or-wide literals.
- `native-go-v8`: extraction for formats the upstream harmless test files exposed (cpio, SZDD, BinHex, mbox, RTF embedded objects, HTML data URIs, bzip2 ZIP members, ARJ, Symbian SIS, NSIS and InstallShield installers, executables stored inside executables); executable headers follow what the Windows loader accepts, and known packers are reported as incomplete instead of malformed or clean. Earlier cached verdicts are invalidated.
- Later `native-go-v8` fixes (same identity, verdicts only move from incomplete to clean or detected): only the real Authenticode signature bytes are skipped in a PE overlay and a chance gzip/bzip2 header whose data does not decode is not an archive; link analysis runs only on HTML (the reference's HTML markers), not on any object containing `<`; PDF strings are batched per object and PDF attachments, metadata, functions and scripts are scanned as data; rules are indexed by pattern so an object only evaluates rules it can affect (per-object cost for small files fell from about 13 ms to under 1 ms with the full feed).
- `native-go-v9` (2026-09-24): objects are typed and matched the way the reference does it. Each object is typed from its content (the feed's file-type database or a built-in table, text detection, tar and Office fix-ups, type patterns, UTF-16 and byte-order-mark passes) and pushed on a layer stack that `Container` and `Intermediates` rules read; ignored types (media, databases) are not analyzed. Raw bytes meet the object's typed rule root (targets 0-14) and root 0; raw HTML meets only exact file hashes (a superset of the reference, which skips raw HTML). Text gets a normalized representation matched against target 7. Executables get the reference's PE, ELF and Mach-O header information; a header the reference would reject disables header-relative offsets (EOF offsets still apply), and a PE header this engine cannot model leaves them unknown. The raw scan's type recognition retypes text to HTML or mail and records embedded objects. `.ftm`, `.cdb`, `.idb` and `.crb` files load, and the remaining rule syntax is supported, so every rule of the official feed that the reference loads is accepted. Rules whose analysis is not implemented load as "unknown when reached" (see below). The UTF-16 transcoding and entity-decoded HTML representations of earlier versions are removed for parity; raw HTML is still matched against the generic rules and exact hashes (a superset until normalized HTML exists). Objects found inside other files are handled as the reference handles them: a ZIP entry header is read as one entry on a ZIP layer, RAR, cabinet and ARJ archives are parsed on a layer of their type (a marker that does not parse describes nothing), archives inside a program's image are extracted, and a MIME message is parsed as mail. Icon-group rules are unknown only where the reference would compare an icon, and version-information offsets match at the keys of the version strings. Earlier cached verdicts are invalidated.
- `native-go-v10` (2026-09-26): logical bytecode programs on a parsed PE file receive the reference's PE header record (`__clambc_pedata`: offset, entry point, section count, header size; zero elsewhere) and its nine-field section records, and the interpreter gains the runtime functions they were missing: `malloc`, `memstr`, `read_number`, `pe_rawaddr`, `get_pe_section`, `get_environment` and `disasm_x86`. `disasm_x86` is backed by the new package `x86dis`, a 32-bit x86 decoder written independently whose 64-byte instruction records equal the reference's on about 105 million compared inputs (see the validation notes below). `memset` and `memmove` accept an `i32` length as the reference effectively does. Programs that were unknown when reached now load and run, and programs that ran see real header data, so earlier cached verdicts are invalidated. A program that reads the executable header (the header record, `pe_rawaddr` or `get_pe_section`) does not run on an object whose PE header this engine cannot model, nor on a parsed ELF or Mach-O object, where the reference fills header data this engine leaves zero: those objects are incomplete, never clean. Two newly loadable JavaScript downloader programs (`BC.Js.Downloader.Adodb-5999914-0`, `BC.Js.Downloader.Nemucod-5999936-0`) need more than the per-run step allowance (1,000,000) on real scripts, mostly because `file_find`, `memstr` and `malloc` charge one step per byte, so they still end incomplete there.
- `native-go-v11` (2026-09-26): the official PE hook programs (kinds 257 and 259 with logical triggers) run on the reference's Windows-program parser path: the trigger flags the object, then a hook phase reads the header as the parser does (sections past the end of the file removed, strict checks), runs no hook on a failed reading or a 64-bit program, runs the all-PE hooks, and runs the unpacker hooks unless the parser returns first or a built-in unpacker this engine lacks would be tried (reason `pe-unpacker`). Hooks see zero match counts and offsets and the full header record. `__clambc_kind` reads 0; runtime calls charge per 64 bytes; `memstr` is linear for long needles; a run may use 50 million steps, a failed run is charged the steps it used, and an object's bytecode runs share half of its budget. Checked file by file against the reference's debug output on 3,958 files with no difference. Undecided objects may now be decided, so earlier cached verdicts are invalidated.
- Image fuzzy hashes (2026-09-27): `fuzzy_img#` subsignatures are decided on PNG, GIF and JPEG objects with the reference's hash computed from Go's decoders, and bits the engine cannot be sure of (rounding boundaries, JPEG decoder differences, ties with the median) left uncertain: a rule is false when a certain bit differs, a detection when every bit is certain and equal, unknown otherwise. Identical hashes on 4,022 of 4,078 reference-hashed images and no wrong certain bit (`signatures/imagehash.go`).
- `native-go-v12` (2026-09-27): regex subsignatures are evaluated (they were unknown when reached; 1,185 official rules). `signatures/regex_syntax.go` parses the reference library's 8-bit syntax (no Unicode, C-locale classes, line-feed newline); `regex_match.go` is a backtracking matcher with the library's priorities, captures, backreferences, lookaround and atomic groups; `regex_sub.go` applies the offset window, anchoring, the `g` loop, the 100 MiB object limit and the trigger, checks first the text every match must contain (once per object) and tries only starts whose byte can begin a match; `regex_prefilter.go` builds a looser Go pattern (each byte read as one character, high bytes as private-use characters) that decides, when the matcher reaches its limits, whether any match is possible at all. The library gives up on a call past 100,000 backtracking calls or 2,000 nested frames and the reference then counts no match; the matcher counts at least as much as the library (margins of 1,000 calls and 100 frames cover the largest shortfall seen, 3 calls and 9 frames) and reports a regex it cannot rule on as unknown (`pcre-limit`). Every official regex compiles; the constructs left unknown when reached are those whose meaning depends on the library version or build (recursion, conditionals, control verbs, `\R`, `\X`, `\p`, `\Q`, the `x` option, variable-length lookbehind, `{,m}`). Macro module text (buffer mode) keeps regexes unknown. A Go 1.25 defect in alternation factoring (`A1*|[Aa]` loses the lowercase `a`) is avoided by adding a never-read character to every class of the Go pattern.
- Long pattern gaps (2026-09-27, same identity: verdicts only move from incomplete to clean or detected): a start that exhausts the per-pattern backtracking memory (200,000 remembered states) is decided by a forward search that carries the positions each pattern node can be reached at as ranges (`signatures/pattern_forward.go`), only for rules whose result depends on it. It is identical to the backtracking search on 1.45 million random starts. Windows incompletes from the state limit fell from 1,377 to 13 of 73,051 files, and the Windows scan took 6 min 46 s instead of 12 min 37 s.
- `native-go-v13` (2026-10-03): HTML objects are matched through the reference's normalized views instead of being incomplete: the page without comments and the text outside tags as HTML (text only up to 8 MiB), each script block as normalized JavaScript (as HTML, then as text), and each `data:` URI file and base64 style image as an object; the raw page meets no rule, not even file hashes, as in the reference (checked). `engine/htmlnorm.go` follows the reference's reader (chunks, states, 743 named entities learned from its output); `engine/jsnorm.go` renames declared names, unescapes strings, folds `"a"+"b"` and `unescape("...")`, and decodes `eval("...")` in one pass after the block with the block's own reader, as the reference does (details in the file comment and in `TestJSEvalAsReference`). Encoded scripts, the `dF` and packer decoders, and blocks whose decoded eval strings chain through the reader's unfinished state stay incomplete. Checked against the reference's own normalized files: 2,128 pages identical in every view, 2,000 random script soups with 1 difference left (1,031 identical, 968 incomplete); `TestHTMLNormalizationAgreesWithReference` replays 2,247 generated pages (`testdata/js_oracle.tar.gz`).
- `native-go-v14` (2026-10-03): icon-group rules are decided on a program's first group icon where the measures this engine computes are enough (they were unknown when reached). `signatures/iconpe.go` extracts the group's icons (bitmap depths 1, 4, 8, 24 and 32, AND masks, alpha; PNG and icons narrower than 16 pixels skipped), `iconhash.go` computes the reference's icon hash (compositing onto white, halving or nearest-pixel resampling to 16, 24 or 32 pixels, the bright, dark and colorfulness areas, the color spread) and `icons.go` the reference's confidence (per-measure credits, every division rounded down, a match from 80 minus half the width). The edge areas, and the colorfulness areas of gray icons, are not reproduced and count as bounds from 0 to 100: a rule is true when it matches with them at 0, false when no icon matches with them at 100, unknown otherwise; groups with features never checked against the reference are unknown. The icon database loader now rejects what the reference rejects (widths other than 16, 24, 32, out-of-range values, corners and spreads, more than 256 names per group field). Everything was learned from the reference's printed hashes and confidences (`docs/edr/native-engine/parity/icons.md`); `TestIconHashesAgreeWithReference` and `TestIconMatchesAgreeWithReference` replay 187 generated icons and 2,590 judged entries.
- UPX step (2026-10-03, same identity): where the reference's parser would try UPX before its unpacker hooks and every NRV decoder certainly fails, the hooks run as in the reference.
- `native-go-v15` (2026-10-07): UPX programs are unpacked as the reference unpacks them and the rebuilt program is scanned as a child (`agent/edr/unpack/upx.go`; `docs/edr/native-engine/parity/upx.md`): the decompressor routine the entry code holds (five 13-byte routines at entry+0x69 or entry+0x71) gets a preferred attempt at the section start moved by the `mov esi` skew, the other NRV decoders try the section start and 0x15 bytes in, and UPX 3.91's LZMA stub gets a silent LZMA attempt (`agent/edr/unpack/lzma.go`, written from the public LZMA specification); without the stub's `lea edi` marker the last header copy with sections is rebuilt. Checked on 1,125 programs from nine UPX releases and crafted variants: every outcome and every rebuilt byte as the reference's. The parser model now decides packed programs: a 32-bit program whose header this engine models is incomplete (`pe-unpacker`) only where the reference would use a built-in unpacker this engine lacks; packer names and stubs alone no longer make it incomplete, because the reference unpacks nothing else either. A few clean verdicts can become incomplete, so earlier cached verdicts are invalidated.
- `native-go-v16` (2026-10-08): PDF files are read as the reference reads them (`docs/edr/native-engine/parity/pdf.md`): the PDF hook programs run (at level 216 the official ones return 0 in every phase), streams of files encrypted with an empty user password are decrypted (strings are not, as in the reference), filters stop at image filters and Crypt and skip unknown names, predictors are not undone, LZW uses early change 1, stream lengths are measured to `endstream` when /Length disagrees, repeated keys keep their first value, empty objects are null, PDF objects no longer count as files and the parser's token budget grows with the bytes read. Streams the reference does not extract are still analyzed, without making the file incomplete. Cached verdicts are invalidated.
- `native-go-v17` (2026-10-08): macro projects are matched as the reference matches them (`docs/edr/native-engine/parity/macros.md`): the project dump (its own decompressor, record order, codepage conversion as on Windows and text normalization) as a SCRIPT file against root 0, and module source as buffers against roots 2 and 0, with regexes decided there; a `VBA` storage without a `_VBA_PROJECT` stream is no project. Pictures of Excel drawing groups are extracted and scanned. Typing converts UCS-4 in the 2143 and 3412 byte orders as big-endian, as the reference does, so property-set streams (`FE FF 00 00`) no longer leave Office files incomplete. FAT sectors marked as end of chain or free are accepted, and unterminated PowerPoint storage streams are read. Cached verdicts are invalidated.
- `native-go-v18` (2026-10-14): compound files are read as the reference reads them (`docs/edr/native-engine/parity/cfb.md`): lazily, keeping what can be read; only nonzero bytes at offset 34, bad sector shifts and files under 517 bytes make it read nothing. Chains end at negative entries, block numbers past the reference's bound, repeated blocks or unreadable blocks, keeping what was read; the directory's entry n is the (n%4)-th entry of its (n/4)-th sector; the tree walk ends at a revisited entry or the recursion limit (a limit); names follow the length field and the reference's spelling; long names are skipped when the reference writes streams to a folder. Sector shifts 6-8 and long names in workbooks are undecided (`compound-file`). Cached verdicts are invalidated.
- `native-go-v19` (2026-10-14): a ZIP member found inside another object is scanned as far as its deflate data decodes when the data ends early and not at all when it fails to decode; every PowerPoint ExOleObjStg record is read as zlib data whatever its instance and scanned as far as it decodes (uncompressed storages are not parsed, as in the reference), and a record of length 0 cancels the stream's extraction; workbooks whose records break off before a FilePass record are not malformed (`parity/cfb.md` section 9). Cached verdicts are invalidated.
- `native-go-v20` (2026-10-14): ZIP archives are read as the reference reads them (`docs/edr/native-engine/parity/zip.md`): the last end record anywhere, the directory exactly at its offset, each member's flags, method and size from its local header (from the central entry under flag 8), no CRC or uncompressed-size checks, stored data shorter than declared inflated, deflate data scanned as far as it decodes when it ends early and not at all when it fails, no ZIP64. Cached verdicts are invalidated.
- `native-go-v21` (2026-10-14): MEW programs are unpacked and rebuilt as the reference does it (`docs/edr/native-engine/parity/unpackers.md`), and the rebuilt program is scanned as a child; the reference's generic rebuilt program is reproduced byte for byte. Cached verdicts are invalidated.
- `native-go-v22` (2026-10-14): WWPack32 programs are unpacked and rebuilt as the reference does it (`docs/edr/native-engine/parity/unpackers.md` section 3). Cached verdicts are invalidated.
- `native-go-v23` (2026-10-14): PEtite 2.2 programs are unpacked and rebuilt as the reference does it (`docs/edr/native-engine/parity/unpackers.md` section 4): its section list, the PEtite code stripped, the entry point decrypted through the import list, sections past the stub written empty; a failure lets the parser go on. The generic rebuilt program's image size now rounds the last section's address and size separately. PEtite 2.1 stays incomplete. Cached verdicts are invalidated.
- `native-go-v24` (2026-10-14): yC 1.3 programs are unpacked and rebuilt as the reference does it (`docs/edr/native-engine/parity/unpackers.md` section 5): its two emulated decryption layers with the reference's opcode set, the section name list, the rebuilt header. A failure lets the parser go on. Cached verdicts are invalidated.
- `native-go-v25` (2026-10-15): ASPack 2.12 programs are unpacked and rebuilt as the reference does it (`docs/edr/native-engine/parity/unpackers.md` section 6): the LZX-style decoder written from the stub, the block loop, the first block's call-address filter, the reference's section list. The generic rebuilt program's image size is now 0x1000 plus the sections' rounded sizes. A failure or another version lets the parser go on. Cached verdicts are invalidated.
- `native-go-v26` (2026-10-15): Upack programs with the upstream sample's layout are unpacked and rebuilt as the reference does it (`docs/edr/native-engine/parity/unpackers.md` section 7): standard LZMA through the engine's decoder, the call fixups, the one-section rebuild. Other Upack layouts stay incomplete. Cached verdicts are invalidated.
- `native-go-v27` (2026-10-15): 7-Zip archives are extracted as the reference extracts them (`docs/edr/native-engine/parity/sevenzip.md`, package `sevenzip`): Copy, LZMA, LZMA2, PPMd (variant H, the model reproduced exactly), the x86 filter, BCJ2 and the ARM filter; members under the coders the reference skips (BZip2, Deflate, Delta, the other branch filters) are skipped, an archive whose header does not check gives nothing, encrypted members and damaged data are incomplete. 7-Zip archives inside other objects are extracted the same way. Cached verdicts are invalidated.
- `native-go-v28` (2026-10-15): PESpin programs of the upstream sample's version are unpacked and rebuilt as the reference does it (`docs/edr/native-engine/parity/unpackers.md` section 8): its three stub layers, the file checksum that keys the sections, its emulator for the two generated byte decryptors (13 operations), aPLib depacking into the virtual size. A failure lets the parser go on; sections the parser keeps elsewhere stay incomplete. Cached verdicts are invalidated.
- `native-go-v29` (2026-10-15): Deflate64 and imploded ZIP members are decoded as the reference decodes them (`docs/edr/native-engine/parity/zip.md` section 4, package `zipcodec`): no size or CRC checks, cut Deflate64 data scanned as far as it decoded, imploded data decoded until it ends. Cached verdicts are invalidated.
- `bytecode`: ClamBC interpreter with typed virtual memory, globals, bounded APIs and logical scanner callbacks. Only the validated logical-callback profile runs; programs that need other APIs or hooks load as unknown when reached, and malformed programs reject the generation.
- `x86dis`: decodes one 32-bit x86 instruction (at most 32 bytes) into the reference's instruction record for `disasm_x86`. The opcode maps are written by hand; operation numbers, refused encodings and quirks were learned only from the reference library's outputs.

## Compatibility matrix

| Capability | Implemented behavior | Remaining boundary |
|---|---|---|
| File hashes | MD5, SHA1, SHA256; exact size or wildcard; `.hdb`/`.hsb`; compact sorted tables (the official hash lists retain about 156 MiB instead of 518 MiB with per-entry maps); objects of five bytes or less are never analyzed, as in the reference | No fuzzy hashes, SHA3, reputation inference or malware-family inference from bare hashes |
| PE section hashes | `.mdb`/`.msb` over the reference section layout (file/section alignment with the 0x200 fallback, virtual-only and truncated sections); only sizes named by a rule are hashed | No import-hash/certificate policy engine |
| False-positive hashes | `.fp`/`.sfp`; exact allowlisted objects bypass recursive analysis | Not a general certificate/filename allowlist; an allowlist alone cannot make the engine healthy |
| Extended signatures | `.ndb` targets 0-14 matched on objects of the target's types, as in the reference; absolute, EOF, EP and section offsets for PE, ELF and Mach-O; version-information (`VI`) offsets at the keys of the version strings of the first 16 version resources (identical to the reference on 1,635 Wine programs); byte/nibble wildcards, gaps, alternatives, wide and word-boundary modifiers; a start that exhausts the backtracking memory is decided by a forward search over position ranges when a rule depends on it | A start can still exceed the forward search's own bound (that rule is then undecided); `VI` offsets are unknown on a PE header this engine cannot model |
| Logical signatures | `.ldb`: boolean and supported count combinations; size/entry/section constraints; supported literal modifiers; binary and ASCII hex/decimal/auto byte comparisons; bounded HandlerType rescans for PDF, GRAPHICS, SCRIPT and ZIP; supported `.cbc` logical callbacks receive full subsignature counts and first offsets; `&` and `\|` use the reference's equal-precedence grouping and `,Y` is ignored on a single subsignature, as upstream does; regex subsignatures (trigger, offset window and anchoring, `g` counts, flags `i s m A E U g r e`, backreferences, lookaround, atomic groups) with the reference library's 8-bit meaning, checked against that library (PCRE2 10.47) on 30,002 random and 22,710 feed-pattern cases with identical results; a rule that cannot be decided (a per-pattern state limit, a failed bytecode run, a regex the reference's library could have given up on) makes only that rule unknown — any detection wins, otherwise the object is incomplete and container recursion still runs | Unsupported byte-comparison forms remain unknown when reached; image hashes are undecided for uncertain bits or graphics-typed formats whose reference behavior is not established; JPEG 2000 is known false because the pinned reference cannot hash it; an icon group is undecided only where the edge areas or a gray icon's colorfulness could change the result, or the group holds a feature not checked against the reference (only on a non-DLL PE with a resource directory and a group icon, for a group some icon entry carries); `Container` and `Intermediates` descriptors are evaluated on the layer stack; bytecode callbacks are limited to the profile below |
| Legacy text rules | `.db` name/hex patterns | Same supported body syntax limits |
| Signature containers | Bounded CVD/CLD unpacking in archive order; CVD MD5 integrity check plus verification of the header's upstream publisher signature (the container is then reported as authenticated; `feedpack` refuses unauthenticated containers unless `-allow-unauthenticated-containers` is given); `.ign`/`.ign2` ignore lists and `.cfg` dynamic configuration read before any rule; `.ftm` file types; `.cdb` container metadata rules checked on ZIP (central and local headers) and RAR members before extraction; `.idb` icon data loaded with the reference's checks and used by icon-group rules; `.pwdb` ZIP/ANY archive passwords (cleartext or hex) selected by Engine window and ignore list, tried in the reference's reverse load order and gated by `ARCHIVE.PASSWD`; `.crb` blocklist entries checked against embedded Authenticode certificates, and trust entries applied only after signer, chain and PE-image verification (unsupported certificates remain conditional unknowns); PUA databases skipped and counted | Publisher-side verification of supplied `.cvd.sign` files against the pinned Cisco root is implemented; current production `.sign`/CVD pairs still need acceptance testing. No CDIFF; `.cld` files are never authenticated; remote containers must still be authenticated by the pinned native manifest; unsupported member types/rules reject the generation |
| Archives | ZIP read as the reference reads it (store, deflate, bzip2 method 12; traditional PKWARE decryption with an applicable `.pwdb` password; no CRC or size checks; damaged deflate data scanned as far as it decodes when it ends early; ZIP64 sizes and offsets resolved when present), gzip, bzip2, USTAR/PAX; XZ LZMA2, Zstandard, streaming single-volume RAR, and single-cabinet CAB stored/MSZIP/Quantum/LZX with history, cancellation and expansion bounds, salvage of bad checksums and available member prefixes, validated LZX/Quantum output after later codec errors, and a native MSZIP inflater with CK recovery, reference Huffman/zero repair, dictionary streaming across records and failed-seek boundaries; CAB member metadata checked before extraction in file-table order, with cached referenced folders; advisory root header fields and local continued-to-next data admitted, with bounded declared-size views for embedded CABs; CHM stored/LZX entries; 7-Zip (Copy, LZMA, LZMA2, PPMd, x86, BCJ2 and ARM filters; members under other coders skipped, as in the reference; encrypted members incomplete); cpio (odc, newc/crc, old binary in both byte orders; a missing trailer is incomplete); ARJ stored and methods 1-4 with header and file CRC-32 checks; SZDD (`COMPRESS.EXE`) and KWAJ (methods 0–4); Symbian SIS packages before OS 9 (zlib files, per-language records); recursive analysis without filesystem extraction; every tar entry type that carries data is scanned (a multi-volume continuation is incomplete); ZIP entries marked as directories are scanned when they hold data; overlapping ZIP entries and CAB blocks share one decode allowance; Zstandard streams may start with skippable frames; decoded CAB, RAR, ARJ, SZDD and KWAJ output is charged to the scan budget | No split archives, encrypted RAR/7-Zip decryption or ZIP strong encryption; a ZIP without an applicable password stays incomplete. CAB members needing predecessor history or split CFDATA, plus broader damaged-layout qualification, XZ non-LZMA2 filter chains, Symbian OS 9 SISX or InstallShield `ISc(` cabinets: these are recognized and return incomplete. ARJ methods 0-4 were cross-checked against archives written by the ARJ 3.10 program (a wrong decode fails the CRC and is incomplete). RAR uses a pinned pure-Go decoder; new tests verify stored/nested members, not comprehensive RAR3/4/5 compression-corpus equivalence |
| Mail | Recognized MIME messages, multipart parts, base64 and quoted-printable attachments; mbox files split at "From " postmark lines, each MIME message parsed; BinHex 4.0 introduced by the standard "(This file must be converted with BinHex" line (run-length decoding, data and resource forks, CRC checks); historical and Base64 uuencode envelopes; TNEF legacy attachment data, MAPI property lists and RTF LZFu/MELA content with checksums and limits; `message/rfc822` and `message/global` parts are parsed recursively and `message/partial` is incomplete; a message is recognized from any first header when a MIME field is present (header scan capped at 256 KiB and 8,192 lines) | No complete Outlook parser or full TNEF property/object coverage; message recognition is limited, not full ClamAV mail normalization |
| Scripts/text | Text objects are matched raw against root 0 and as normalized text (the reference's script normalization) against target 7, as a file or as a buffer depending on the database; HTML is matched through its normalized views (page without comments, text, normalized JavaScript, `data:` files), not raw, as in the reference; UTF-16 HTML is converted for those views and also gets the full raw scan; RFC 2397 data URIs (base64 or percent-encoded) in markup attribute values and CSS `url()` are decoded and analyzed | HTML objects with encoded scripts (`#@~^`), the `dF` or packer decoders, or chained eval strings are incomplete (`html-normalization`); one random-input difference from the reference is known (a script inside one long string); no script execution or emulation |
| Phishing | `.pdb` H/R protected domains and all 185 official `daily.wdb` M/X allow rules compile; `.gdb` S/S1/S2 SHA-256 URL hashes, P prefixes and W local hash allows compile; bounded HTML link/form/image extraction, IDNA and domain boundaries; domain/HTTPS spoof checks. GDB checks canonical destination host/path variants before displayed-link checks, including port and bare-query path treatment, four directory prefixes, repeated slashes, nested percent escapes, non-ASCII hostname bytes, and chained dot segments following the pinned buffer walk. X/Y rules use the pinned 1.5.4 pair matcher; text between a PDB/WDB type letter and `:` is ignored as in that parser. X/Y accept scheme expressions, explicit anchors, userinfo and the reference's optional-path suffix removal; M rules retain final encoded octets and root dots | Overlong canonical URLs and other ambiguous forms stay incomplete; no native ClamAV differential corpus has validated full GDB behavior. R uses a restricted POSIX host-oriented subset. The zero-unknown WDB audit is a compile result, not exhaustive URL normalization or phishing parity. Rules outside implemented semantics retain required literals and can make a relevant detection incomplete (`phishing-allowlist`). Original HTML is inspected once |
| OOXML | ZIP members receive recursive generic scans; embedded OLE files are extracted | No complete Excel XLM, Office application semantics, VBA deobfuscation or p-code disassembly/emulation |
| PE / ELF | Bounded structural validation following what the Windows loader accepts (a PE header inside the DOS header, section data inside the headers, SizeOfHeaders past the file, raw pointers rounded down to 512 bytes; COFF symbol fields are ignored for images); raw signatures; PE entry/section signature offsets; installer overlays: NSIS (per-file or solid, LZMA with end marker, raw deflate or its headerless bzip2 variant; the LZMA dictionary is limited to what the block can produce), InstallShield setup records and InstallShield MSI obfuscated zlib files (installer data cut short is incomplete); overlay data before and after installer data gets the same archive and executable search (at most 16 installer segments per overlay); self-extractor archive headers validated beyond their magic (at most 1,024 ARJ header checksums per search); the Authenticode table is not treated as overlay; executables with a loadable PE header stored at any offset are extracted and analyzed (each up to 40 MiB, at most 1,024 header parses per object) | ELF and Mach-O headers follow the reference's readers (entry point, sections; universal Mach-O files are analyzed per architecture). UPX, MEW and WWPack32 programs are unpacked and rebuilt as in the reference and the rebuilt program is scanned. ASPack 2.12, PEtite 2.2, yC 1.3, Upack (one layout) and PESpin (one version) are unpacked as in the reference too. FSG 2.0 and 1.33 and NsPack are decoded and rebuilt on the validated paths; other versions and undecided packer inputs remain incomplete (`pe-unpacker`); other packers (MPRESS, PECompact, Themida and similar) are raw-only, as in the reference; a program whose header the parser model cannot read is incomplete when it carries packer section names, stubs or header layouts; real-world NSIS bzip2 installer corpus validation is pending; compiled AutoIt scripts are recognized and incomplete; unknown packers may remain raw-only |
| OLE2 / VBA | Compound files read as the reference reads them (lazily, keeping what can be read: broken chains keep the bytes read, a final partial sector is zero-filled, shared sectors are allowed, the tree walk follows the reference's order and stops where it stops); RTF `\objdata` objects are decoded (hex text, `\binN`) and their OLE1 native data analyzed, unwrapping the "Package" class like an Ole10Native stream; macro projects (a `dir` stream in a file with a `_VBA_PROJECT` stream) give the reference's project dump (SCRIPT, root 0, file mode) and each module's source through the reference's decompressor (roots 2 and 0, buffer mode, regexes decided); Word 6 macro directories are read and their per-record XOR data is matched as OLE2 buffers; pictures of Excel drawing groups are extracted and scanned; Excel `FILEPASS` and PowerPoint encryption markers make the object incomplete; Ole10Native packages and PowerPoint embedded storages (zlib data whatever the record instance, scanned as far as it decodes) are unwrapped and scanned | Codepage conversion follows the Windows reference (embedded tables for 38 single-byte codepages and 932/936/949/950). Undecided: codepages 0 and 1 (the machine's own) with non-ASCII text, UTF-16 and codepages without a table (`vba-module-text`), workbook pictures in forms not observed (`workbook-pictures`). Not produced: the Excel 4 macro listing (no rule refers to it). Undecided: sector shifts 6-8 and long stream names in workbooks (`compound-file`). Encrypted Office content is incomplete. No p-code disassembly/emulation |
| PDF | Literal/hex/name escapes, stream extraction with lengths measured to `endstream` when /Length disagrees or cannot be resolved, Flate/LZW (early change 1)/ASCIIHex/ASCII85/RunLength filters, decoding stopped at image filters and Crypt (the data decoded so far is the object), unknown filter names skipped, predictors not undone (as in the reference), object streams; objects hidden inside other objects' streams are found; streams of files encrypted with an empty user password decrypted (RC4, AES-128, AES-256 revisions 5 and 6); PDF hook programs run in four phases; output of a failed or over-limit filter is charged to the budget | No full xref resolver, no PDF JavaScript emulation or image/font codecs; a user password leaves the file incomplete (the reference can alert on it); strings of encrypted files are not decrypted (nor by the reference); a damaged Flate stream in an extracted stream is incomplete (the reference may scan its partial output) |
| SWF | FWS tag parsing; CWS/zlib and ZWS/LZMA normalization; binary data, ActionScript payloads, metadata and nested sprites scanned | No ActionScript execution or specialized image/font/audio/video decoding; unimplemented tags remain incomplete |
| Bytecode | Supported ClamBC 6/7 logical callbacks (kind 256), full counts/first offsets, read-only hook metadata, the PE header record and section records on parsed PE files, bounded arithmetic/memory/file/search/diagnostic APIs plus `malloc`, `memstr`, `read_number`, `pe_rawaddr`, `get_pe_section`, `get_environment`, `disasm_x86`, and API 89 `matchicon` with ClamAV 1.5.4's invalid/no-match return profile; strong returned labels detect after successful execution; the `engine_functionality_level` and `engine_dconf_level` APIs return 234; freed memory slots are reused with generation checks; struct-pointer prototypes must equal the reference's compound types; globals are initialized as the reference initializes them | Programs using other hook kinds or APIs load as unknown when reached: with a logical trigger, where the trigger holds; without one, on every object of their hook's type. PE hook programs (kinds 257 and 259) with logical triggers run in a hook phase (see `native-go-v11`). ClamAV 1.5.4's API 89 constructs a resource record without its data-directory count, so its resource walker returns no match even when an ordinary icon rule matches; a future reference version that repairs this requires a separate profile. Programs that write output or call `extract_new` use the bounded recursive child scanner, but their actual unpacker output still needs oracle validation. GEPN and debug metadata remain unsupported. PUA/heuristic/weak bytecode labels produce incomplete analysis until their classification policies are implemented |
| Heuristic issues | Supported phishing matches use explicit detection names; resource, malformed and recognized unsupported outcomes become unknown health telemetry | No general parity with ClamAV heuristics; incomplete-analysis errors alone do not trigger malware quarantine |
| Endpoint response | Existing quarantine/restore, process guard, tree termination, AMSI block, ransomware and WFP paths remain wired; the launch guard kills on a malicious verdict even when a secondary error occurs; verdict-cache failures no longer drop a verdict; files left by the retired updater are moved to a sibling `legacy-database-*` directory at service start | Existing Windows runtime behavior needs reacceptance; Linux monitoring/scheduling now have a tested baseline; central console acceptance, Linux prevention/network/ransomware and WDAC/fapolicyd remain open |

Rules are selected at the pinned ClamAV 1.5.4 functionality level 234: each format's functionality window (checked before any other validation of the rule, with the reference's field parsing), `.ign`/`.ign2` ignore lists from anywhere in the generation (name normalization and optional MD5 of the rule line), the potentially-unwanted-application databases (never loaded here), unknown logical attributes and file types, targets of 15 or more, and the formats a `.cfg` directive switches off. Such rules are skipped and counted by reason in the compile report (`level`, `ignore-list`, `pua`, `unknown-attribute`, `target`, `dconf`, `legacy-format`) and in `sigaudit`; malformed selection fields stay compile errors. The compiler honors the `.cfg` bits PHISHING.ENGINE, BYTECODE, ARCHIVE.PASSWD, PE.MD5SECT, PE.CERTS and PE.CATALOG; a directive that applies at 234 and switches off any other default feature fails compilation, because the scanner cannot switch it off yet (none of the retained current feed's directives applies at 234). This is rule selection, **not an assertion of ClamAV engine/API equivalence**. Syntax the reference rejects fails compilation here too. Selected rules whose analysis this engine does not perform load as **unknown when reached**: each keeps the condition under which the reference could evaluate the missing part, and only an object meeting it gets an incomplete result (never clean, never a detection). `Report.Unknown`, `engine-check`, `status.json` and `sigaudit` count them by reason, so nothing is dropped silently.

Default database limits admit the complete official feed with headroom: 384 MiB per file or container member, 1 GiB for inputs plus expanded members, 8 million rules. Feed downloads allow 256 MiB per artifact and 512 MiB per generation (`feedpack` uses the same bounds). Measured on the official feed of 2026-09-23 (main 63, daily 28132, bytecode 339) with every accepted rule loaded (3,605,099 rules, 391,171 body/logical rules, 1,069,174 distinct patterns): compilation takes about 4.2–4.5 s and retains about 1.05 GiB of heap (peak resident about 2.6 GB while compiling). Matching runs one shared anchor pass per object (rarity-chosen anchors, every occurrence found once, each pattern verified only at the start positions its anchor allows) and was checked to give results identical to the previous per-pattern matcher on about 260,000 generated objects and on real files with every official rule loaded. Real Windows programs, libraries, drivers and installers from 78 KB to 24 MB, including a 14 MB .NET core library and 20 MB graphics libraries, finished clean in 0.04–0.62 s using about 2–12 work steps per byte; before this change most files over 1 MiB exceeded the budget and returned incomplete. The matching budget is max(500 million steps, 16 steps per object byte); an explicit `MaxMatchSteps` without a per-byte value stays a hard cap. A scan can allocate up to about 10 bytes per input byte of short-lived memory for anchor occurrences (stored occurrences are capped at about 24 MB live). With `native-go-v9` the complete official feed of 2026-09-24 (main 63, daily 28133, bytecode 339) compiles: 3,628,077 rules in about 6-7 s, about 1.16 GiB of retained heap and about 2.6-2.8 GB peak resident; `sigaudit` reports 0 rejected rules and 1,457 accepted rules that are unknown when reached (regex 1,185, image hash 109, icon group 80, certificate trust 30, bytecode 21, handler type 11, phishing allow 10, version information 10, byte comparison 1).

Logical bytecode additionally requires a supported LDB trigger, header and LDB functionality windows that include 234 (a program whose header window excludes 234 is skipped before it is parsed; one whose header is selected but whose trigger is not is malformed), and an `i32` function-zero entrypoint with no arguments. A matched trigger alone is not a detection: the VM must finish and explicitly set a valid nonempty detection label. Integer return values alone do not detect. Missing subsignature offsets use `0xfffffffe`; full counts are calculated within the shared matching budget. VM failures, invalid labels and resource exhaustion become incomplete analysis. Compiled bytecode has a conservative aggregate storage budget of at most 128 MiB (or the configured database byte limit, when smaller), and each callback consumes the object-wide matching-work budget as well as its per-execution cap.

## ThreatWinds verification

The live [ThreatWinds feed catalog](https://apis.threatwinds.com/api/feeds/v1/list) inspected on 2026-09-22 advertised MD5, SHA1 and SHA256 malware indicator feeds. A SHA1 snapshot was downloaded and inspected: a gzip-compressed tar containing one plain newline-delimited hash list. The public catalog and [documentation](https://docs.threatwinds.com/feeds/download) did not advertise CVD/NDB/LDB/ClamAV databases. Thus hash-feed availability is verified; the assumption that ThreatWinds redistributes ClamAV signatures is **not verified**.

ThreatWinds hashes supplement compatible ClamAV-format signatures in the same native generation. The existing IP/domain feeds continue through the separate network-blocking pipeline. Bare hash indicators receive generic detection names, not invented malware-family labels. The upstream `bases.sum` MD5 file is a corruption check, not an authentication authority. Daily NDJSON ingestion is not implemented for malware signatures: publish complete verified snapshots instead.

## Local operation

1. Build from the `agent` Go module: `go build -o <output>/utmstack_edr ./edr`. Supported endpoint runtime remains Windows; building/testing Go logic on macOS does not implement macOS EDR.
2. Place approved `.cvd` files or `.hdb`, `.hsb`, `.ndb`, etc. in the configured `signature_db_dir`, or the default `<install>/engine/database`, as regular files (symbolic links are refused). A local database is trusted administrator-provisioned content. The official databases load; objects reached by rules that are unknown when reached get incomplete results.
3. Set `signature_db_dir` through `utmstack_edr config set signature_db_dir <absolute-directory>` or `edr.json`. Run `utmstack_edr engine-check`. Exit 0 means the configured database compiles; the JSON states partial compatibility and lists file-type entries, skipped rules by reason, rules unknown when reached by reason (`unknown_when_reached`) and the representations not produced yet (`representations_not_produced`). Exit 1 and diagnostics mean it cannot be activated.
4. `utmstack_edr scan <path>` retains the existing quarantine-on-detection behavior. Use synthetic test files and an isolated test installation during acceptance. `utmstack_edr scan --dry-run --json [--db <dir>] [--workers N] <path>...` analyzes files and folders without acting: one JSON line per file (path, size, clean, signature, incomplete reason, error, milliseconds) and a summary line; nothing is quarantined and neither the verdict cache nor the event spool is written. Exit 0 requires every file clean; exit 1 means something was detected (other files may be incomplete); exit 2 means incomplete/file errors without detections, or invocation failure. Inspect JSON for complete accounting.
5. Restart the service after changing structural database/key/mirror settings. Loose local file changes need a reload/restart; remotely activated generations update the running engine atomically.

No database is silently seeded with a test rule and reported production-ready. Missing/empty/unsupported databases make engine health false; behavioral/network layers can still operate.

Objects for which the reference produces representations this engine does not yet produce (including OneNote, HWP, the XML Office formats, AutoIt, CryptFF, UDF disk images, HFS+ partition contents, and NSIS and InstallShield data on their own) are incomplete when undetected; `engine-check` lists them under `representations_not_produced`. ALZ, ordinary non-solid EGG and common LHA methods now have dedicated readers, with unsupported members marked incomplete. EGG self-extractors remain separately listed until their embedded extraction path is validated. Conventional MBR and validated GPT partition windows, standard ISO 9660/Joliet images, bounded XAR TOCs and heap extents, and PDF-bearing XDP chunks also have dedicated readers. Word and Workbook XML scan decoded `binData` children after bounded numeric-entity repair, but their remaining metadata representations still make an undetected document incomplete.

Legacy `clamd_addr` and `clamd_path` values are accepted for migration but ignored; saving configuration removes them. Native updates do not use the legacy CDN fallback, HTTP signature port or TLS-skip exception. Keeping a last-known-good database offline is supported; fresh signed updates require validated HTTPS and the pinned key.

## Publish combined feeds

Acquire approved feeds on the publishing server using authenticated HTTPS. Discover available upstream feeds using the catalog. For example, the verified SHA1 snapshot endpoint is `https://apis.threatwinds.com/api/feeds/v1/download/list/level1/accumulative/sha1`. Retain upstream authentication and provenance before signing; signing untrusted input does not make its provenance trustworthy.

Create a spec next to the local inputs:

```json
{
  "version": "edr-native-2026-09-22.1",
  "files": [
    {"name": "main.cvd", "path": "main.cvd", "format": "signature", "official": true},
    {"name": "daily.cvd", "path": "daily.cvd", "format": "signature", "official": true},
    {"name": "bytecode.cvd", "path": "bytecode.cvd", "format": "signature", "official": true},
    {"name": "intel.hsb", "path": "sha1_level1.list.tar.gz", "format": "sha1-list.tar.gz", "dedupe_against_official": true}
  ]
}
```

Provision an Ed25519 key pair using your signing-key process. `feedpack` reads a base64 32-byte seed or 64-byte private key from a private file; endpoints get only the base64 public key. It does not create, print, or distribute private keys.

From the `agent` module:

```sh
go run ./edr/cmd/feedpack \
  -spec /secure/feeds/spec.json \
  -key /secure/keys/native-feed.secret \
  -sequence 1 \
  -expires 2026-09-29T00:00:00Z \
  -out /staging/native-release-1
```

For each official CVD, the spec may include `"detached_signature": "daily-28160.cvd.sign"` beside its `path`; this verifies the exact CVD bytes with a PKCS#7 signature chained to the pinned ClamAV root. Pass `-require-detached-signatures` to reject official CVDs without that field. An invalid supplied signature always rejects the release. The `.sign` file is publisher input, not a published artifact; the signed native manifest authenticates endpoint downloads. The reference fixtures validate the parser and certificate chain, but a current production `.sign`/CVD pair has not yet been tested.

Use a future expiry and a strictly increasing sequence for each changed payload. The output directory must not already exist. It contains `manifest.json` and content-addressed `artifacts/` only. The [local `feedpromote` command](cmd/feedpromote/README.md) verifies the signed output with the endpoint's pinned public key, copies and flushes artifacts, then atomically switches the manifest in a local HTTPS mirror directory. The default native endpoint is `https://<server>:9001/private/edr/native/manifest.json`, separate from the old raw signature mirror. The full server's scheduled acquisition, signing-key custody, deployment wiring and fleet rollout still need operational integration.

Endpoint configuration:

```json
{
  "signature_mirror": "https://server.example:9001/private/edr/native",
  "signature_public_key": "<base64 Ed25519 public key>",
  "signature_fallback": "none",
  "sig_update_hours": 4
}
```

Adapters: `signature` (default), `md5-list`, `sha1-list`, `sha256-list`; use `.gz` or `.tar.gz` on list format names for explicit compression. Tar inputs require one regular member. MD5 output files must end `.hdb`; SHA files `.hsb`. Downloads and individual decompressed artifacts are limited to 256 MiB, a generation to 512 MiB and 128 artifacts. Large feeds require publisher-side partitioning. All endpoints retain the same complete signed generation; failed updates cannot silently replace it with a subset.

For a combined release, mark each verified official signature source `"official": true` and each ThreatWinds hash list `"dedupe_against_official": true`. The publisher compiles the official sources with generation-wide loose ignore/config files, removes a bare hash only when an active official whole-file wildcard-size rule covers the same digest, and removes repeated bare hashes across marked lists. It retains indicators whose official counterpart has a specific size, since the bare indicator also covers other sizes. Unmarked CVD/CLD containers are rejected in this mode because they could carry hidden selection files that change which official rules are active. It prints removed, retained and omitted-file counts after a successful publication. The published deduplicated artifacts are normalized `.hdb`/`.hsb` signature files, signed in the native manifest; they are derived from, rather than byte-for-byte copies of, the original ThreatWinds downloads. Preserve source download receipts and provenance separately. `feedpack` validates and signs local input but does not fetch upstream feeds or establish their legal redistribution rights.

Persistence is `<database>/generations/<identity>/{metadata.json,signatures/*}` plus `CURRENT`. The pointer changes after durable file writes and successful compilation. An OS file lock serializes update writers; each writer checks persisted sequence state before activation. Invalid updates, checksum failure, signature failure, expiry, rollback and same-sequence equivocation preserve the active generation. Immutable generations are retained; automatic pruning and fleet rollback/key-rotation workflows remain future operational work. Anti-rollback protects against older network manifests relative to local state; it is not tamper resistance against an administrator restoring the entire database directory from an old backup.

## Resource and operational boundaries

Input reads and concurrent analyses are bounded. Recursion is capped at 16, object count at 10,000, and analysis time at 120 seconds. Host-derived file/expanded-byte caps remain conservative; each recursive/normalized object charges the aggregate budget. ZIP directory counts are checked before the standard library builds its member list. Executable structures use explicit bounds checks rather than allocating untrusted header counts. Signature matching has independent work/state limits.

Codec dictionaries/windows are capped at 64 MiB. XZ preflight checks actual LZMA2 chunk extents as well as all indexed block headers before decoder allocation: the library's `DictCap` field is a minimum, not a ceiling. RAR's older PPM model has a separate format-limited allocation up to 256 MiB, so only one RAR decoder is admitted concurrently across scans; extracted bytes remain limited by the scan budget. Pure-Go dependencies are pinned: `ulikunitz/xz` v0.5.17, `klauspost/compress` v1.20.0 and `nwaples/rardecode/v2` v2.4.1. They do not introduce a ClamAV process, CGo or native library dependency.

This does not impose a process-wide RSS cap: compiled database memory, simultaneous old/new generations and Go GC overhead can exceed serialized feed size. Full-feed memory/throughput/latency testing is still required. Go makes the core independent of the external scanning daemon; it also moves parsing into the EDR process, so parser fuzzing and resource reviews remain necessary.

The scanner rechecks bytes before quarantine, but a narrow path-check-to-move race remains until the OS quarantine implementation operates on a verified file handle. The existing launch guard is post-launch detection/termination, not a driver-level execution barrier. AMSI only sees content submitted by participating hosts.

## Validation and remaining release gates

The new tests cover supported signatures, unsupported-rule rejection, digest-size behavior, allow hashes, PE sections/offsets, logical rules, malformed/oversized containers, nested archives, MIME, UTF-16, generation reload, rollback/equivocation, failed-update preservation, signed HTTPS feeds, feedpack output, generation-aware cache withdrawal, mutation-before-quarantine and incomplete telemetry. Synthetic markers are used; these are correctness tests, not malware-effectiveness measurements.

Validation completed on 2026-09-22: all EDR and agent tests passed with the race detector; host and Windows amd64/arm64 EDR vet passed; the whole agent module cross-built for both Windows architectures. Container fuzzing passed 30,992 executions; signature-parser fuzzing passed 346,591 executions; bounded executable-validator fuzzing passed 2,047 executions. These are short development fuzz runs, not a production fuzz campaign. No Windows runtime or malware-corpus acceptance was performed for this replacement.

On 2026-09-23, integrated tests additionally exercised PDF filters/object streams, Flash compression/nested binary content, fragmented Office streams, VBA directory/module offsets, phishing false-positive boundaries, ZIP64, XZ/Zstandard and nested RAR. New short fuzz runs passed: documents 171,226 executions; archive preflight 175,941; VBA directory 193,509; ClamBC parser/VM 77,038; phishing 204,511; OLE 56,386. The full EDR/agent race suite passed. Host and Windows amd64/arm64 vet passed; the entire agent cross-built for both Windows architectures. These figures measure parser robustness testing, not detection effectiveness or full ClamAV compatibility.

Subsequent bytecode work was checked against a pinned upstream source revision and four successful upstream fixtures plus an infinite-loop fixture. Source comparison corrected COPY dummy fields, literal-storage widths, void calls, boolean arithmetic and abort terminators. Synthetic tests cover typed memory, initialization and lifetime faults, globals, pointer calls, overlapping copies, file APIs and bounded work. The bytecode race suite and optional hash-verified upstream fixture suite passed. See `bytecode/REFERENCE.md` and `bytecode/README.md`; these checks do not constitute complete official-feed or live C-engine differential validation.

The PDF LZW/predictor extension was checked with synthetic streams crossing the 9-to-10-bit LZW width transition in both `EarlyChange` modes, packed TIFF samples at 1/2/4/8/16 bits, all five PNG row filters, filter-chain parameters, malformed/trailing data, and resource limits. The bytecode, signature and engine race tests passed; host vet and Windows amd64/arm64 EDR builds passed. All EDR packages passed ordinary tests, with the local TLS feed package run separately because its loopback listener is disallowed in the default sandbox. A new fuzz run could not start because instrumented dependency builds exhausted available temporary disk; the earlier short fuzz results above predate this PDF extension.

CAB work uses Microsoft's [MS-CAB](https://download.microsoft.com/download/4/D/A/4DA14F27-B4EF-4170-A6E6-5B1EF85B1BAA/%5BMS-CAB%5D.pdf) and [MS-MCI](https://learn.microsoft.com/en-us/openspecs/exchange_server_protocols/ms-mci/3048c471-f739-4d30-bcb3-a962cfe84a69) specifications and ClamAV's [libmspack integration](https://github.com/Cisco-Talos/clamav/blob/main/libclamav/libmspack.c) as references. Synthetic tests cover stored and MSZIP blocks, MSZIP dictionary continuity, reserved fields, optional and corrupt checksums, multiple files, nested gzip, PE overlays, malformed offsets and budget exhaustion. This is narrower than libmspack's Quantum/LZX and multi-cabinet support; no real-world CAB corpus comparison has been completed.

The TNEF extension uses Microsoft's [MS-OXTNEF](https://learn.microsoft.com/en-us/openspecs/exchange_server_protocols/ms-oxtnef/b03b5704-ef54-4390-b981-dc78f8e79a05) and [MS-OXRTFCP](https://learn.microsoft.com/en-us/openspecs/exchange_server_protocols/ms-oxrtfcp/65dfe2df-1b69-43fc-8ebd-21819a7463fb) specifications and [ClamAV's TNEF parser](https://raw.githubusercontent.com/Cisco-Talos/clamav/main/libclamav/tnef.c) as references. It scans legacy attachment-data fields, encapsulated binary/object and Unicode properties, recipient lists and compressed/uncompressed RTF. The LZFu decoder was checked against Microsoft's published compressed stream and CRC example; synthetic tests cover nested gzip, MIME/base64, named/multivalue properties, malformed lengths, CRC failures and resource limits. Unknown property types and embedded objects that cannot be fully decoded remain incomplete; no real Outlook message corpus comparison has been completed.

Historical and Base64 uuencode envelopes are decoded in Go using the [POSIX uuencode format](https://pubs.opengroup.org/onlinepubs/9699919799/utilities/uuencode.html). Their decoded bytes enter recursive analysis; malformed envelopes return incomplete. Tests include a known `Cat` vector, nested gzip, email bodies and malformed envelopes. This is not a complete implementation of every nonstandard mail encoding variant.

For `native-go-v6`, all `./edr/...` tests passed; `./edr/engine` passed under the race detector, `go vet ./edr/engine` passed, and the EDR binary cross-built for Windows amd64 and arm64 with CGo disabled. The full agent-module race suite, Windows runtime acceptance, real CAB/TNEF/mail corpora and new parser fuzzing remain outstanding for this version. See [HANDOFF.md](HANDOFF.md) for the development continuation plan.

For `native-go-v7` (2026-09-23, continuation): a twelve-agent review of the whole engine reported 43 defects; 39 were confirmed with reproducer tests (the others were refuted or already fixed) and all 39 were fixed test-first and re-checked by independent reviewers, each regression test failing on the earlier code. The fixes cover skipped ZIP/tar members, Zstandard skippable frames, CAB/RAR/PDF budget charging, ZIP directory preflight, encapsulated `message/rfc822` mail, encrypted Excel/PowerPoint, Ole10Native and PowerPoint embedded objects, hidden PDF objects, the compressed-RTF allocation bound, PE section/entry alignment, logical-operator grouping, per-rule incomplete results, bytecode memory/typing/global semantics, feed restart after a damaged generation, redundant update downloads, AMSI incomplete-event deduplication, uuencode false positives on script `begin` blocks and malicious verdicts lost on a failed re-check. Follow-ups in the same session: an undecided rule no longer stops container recursion (a detection inside still wins), objects of five bytes or less are skipped as in the reference, the launch guard kills on a malicious verdict despite a secondary error, verdict-cache failures no longer drop verdicts, a volume root is refused as `signature_db_dir`, and the retired updater's files are moved aside at service start. Rule selection now follows the reference at level 216, and the `sigaudit` gap tool measures the real feed: 96,902 rejected detection rules before this session, 22,927 after (99.4% of the detection rules the reference loads are accepted by count). The full EDR/agent suite passed with and without the race detector, host and Windows amd64/arm64 vet passed, and the whole agent module cross-built for both Windows architectures. Still missing: Windows runtime acceptance of this version, a real clean and malware corpus comparison, and new fuzz runs (disk space on the development host was too low). The reference-semantics research and ordered plan for the remaining rules is in [`docs/edr/native-engine/parity/`](../../docs/edr/native-engine/parity/SPEC.md) at the repository root (`SPEC.md` and topic notes).

For `native-go-v8` (2026-09-23): the 44 upstream harmless test files, all detected by the reference engine, were scanned with the official feed minus the 23,445 rules this engine rejects. `native-go-v7` detected 11, reported 15 clean and 18 incomplete (9 of them malformed); `native-go-v8` detects 29, reports none clean and 15 incomplete, none malformed. The 15 need a decoder or unpacker this engine does not have: ASPack, MEW, PESpin, PEtite, Upack, UPX, WWPack32 and yC packed executables, two compiled AutoIt scripts, 7z, Deflate64 and implode ZIP members, CHM (LZX) and an InstallShield `ISc(` cabinet. The earlier malformed verdicts were parser bugs: executable header checks stricter than the Windows loader, a chance two-byte gzip magic in installer data taken for an archive, and a PowerPoint file with one byte after its last sector. On all 75,858 files under `/usr` of a Debian 13 host (75,814 outside the test-file directory, including 2,151 ELF and 49 PE files and 3,979 gzip files), v7 and v8 gave identical verdicts outside the test-file directory: no new detection and no new incomplete result; total scan time changed by under 1%. New tests generate every new format synthetically, including truncation, oversized declared sizes, run-length and decompression bombs and object-count limits; the upstream samples are not in the repository. The engine, scanner, AMSI and signature packages passed with and without the race detector; host and Windows amd64/arm64 vet and Windows builds of the agent module passed. Short fuzz runs passed: `FuzzExtractionFormats` 3.47 million executions, `FuzzExecutableValidation` about 0.8 million, `FuzzNativeContainers` about 8,600. Still missing for this version: Windows runtime acceptance and a real Windows clean and malware corpus, which matters most for the executable, installer and packer changes.

Independent review of `native-go-v8` (2026-09-24): decoded output was cross-checked against files written by independent tools on the Debian 13 host. Every member came out byte-exact (262 members, compared by MD5) from GNU cpio and libarchive cpio archives (odc, newc, crc, old binary, and big-endian binary), ARJ 3.10 archives of methods 0 to 4, `mscompress` SZDD files, BinHex files (data and resource forks), Info-ZIP bzip2 members, NSIS 3.11 installers (zlib and LZMA, solid and per-file, including a 150 MB file with a 64 MB dictionary), and Python-built mbox files, RTF objects (Package and other classes, hex and `\bin`) and HTML data URIs. NSIS bzip2 installers are incomplete as documented; BinHex files whose first line is not the standard one are not recognized. The review fixed these defects, each with a test that fails on the earlier code: SZDD and ARJ reserved memory for the declared output size (64 MiB for a 9-byte input); NSIS blocks allocated the declared LZMA dictionary whatever their size (64 MiB per 21-byte block); the self-extractor search computed an ARJ header checksum at every candidate (1 s for a 16 MiB overlay); an archive or executable after NSIS or InstallShield MSI data, before NSIS data, or after an InstallShield record that could not be read was never analyzed and the file came back clean; InstallShield data cut short came back clean; RTF objects and mbox messages were charged to the budget twice. The 44-sample result is unchanged. On 101,213 files under `/usr` with Wine and NSIS installed (934 Windows executables, 2,941 ELF files), the reviewed and the fixed builds gave identical verdicts; against v7, 19 Wine DLLs moved from malformed to clean and nothing else changed. `FuzzExtractionFormats` ran 4.57 million executions with installer-segment seeds.

Cloud acceptance (2026-09-23/24, disposable Google Cloud VMs; details in HANDOFF.md). Reference comparison on Debian 13 against ClamAV 1.4.3 with the same official database (the Go engine loads the feed minus the rules it rejects): the 44 upstream harmless test files give 29 detected and 15 incomplete, none clean where the reference detects; 855 wrapped copies give no clean miss either; about 103,000 clean system files give no detection and 100 incomplete results (mostly 16-bit Windows font files); median scan time 0.76 ms per file, and the corpus took 4 min 42 s against 21 min 42 s for the reference. Windows Server 2022 endpoint run with the full feed: service ready in about 20 s with about 2.4-2.5 GB working set; install, on-demand scan and restore, real-time quarantine, launch-guard tree kill with quarantine after the kill (also on a second volume and for a self-spawning program), script blocking through a MinGW-built provider, signed update over HTTPS with tamper/rollback rejection, legacy-file migration and branding all passed; decoy writes, truncation, renames and deletions are contained while reads and attribute queries are not; a scan of 17,472 files under System32 and Program Files gave 0 detections from the official feed, 0 malformed and 1 work-limit result, in 6 minutes. Peak memory is high: about 4.5 GiB for a 4-worker scan run on Linux and about 4 GB on Windows. Not verified at runtime: the live-process image lookup for programs that modify a decoy within a second of starting (unit-tested), arm64 Windows, and event delivery to the platform.

For `native-go-v9` (2026-09-24): every task was written test-first (failing test, then code). The full EDR/agent suite passed with and without the race detector; host and Windows amd64/arm64 vet passed; the whole agent module cross-built for both Windows architectures; the branding check is clean; `sigaudit` on the official feed reports 0 rejected rules. A dry-run scan of 12,607 macOS system files with the full feed gave 0 detections, 0 malformed and 2,004 incomplete results (RTF files reaching a regex rule, images reaching an image-hash rule, HTML files, every PDF through the two PDF hook programs, and 6 firmware files over the per-pattern state limit), in 48 s with 4 workers at 2.8 GB peak resident. It also found two defects, fixed with tests: programs containing the bytes `PK\3\4` were reported malformed, and a self-extractor's archive was re-read once per member.

Cloud comparison of `native-go-v9` (2026-09-24, Debian 13 with ClamAV 1.4.3 and the same database, and Windows Server 2022; details in HANDOFF.md): the 44 upstream test files give 29 detected and 15 incomplete, none clean; 836 wrapped copies give no clean miss (443 of the reference's 722 detected, plus 61 it misses); 78,583 Debian files and 68,460 Windows System32 and Program Files files give no detection. An independent final review and the measurement led to the fixes listed above; with them, incomplete results fell from 6.9% to 2.8% of the Debian files and from 14.8% to 4.4% of the Windows files. One worker scans the Debian corpus in 11 min 40 s against 40 min 22 s for the reference; the Windows service is ready 25 s after start at about 2.5 GB.

For `native-go-v10` (2026-09-26): every task was written test-first. The x86 decoder was compared with the reference library's decoder on about 105 million inputs with zero differences (every opcode with every ModRM byte; prefixed, SIB, 16-bit addressing, sign-pattern, cut-short and prefix-run cases; the opcode x ModRM x SIB corpus under 12 prefix sets; and 32-byte windows at every offset of 86 MB of real code from 793 Windows DLLs and 463 Linux programs); the four behaviors the second probe exposed are golden tests. `sigaudit` on daily 28134 reports 0 rejected rules and 12 bytecode programs unknown when reached (21 before). The measurement found that the feed had failed to compile since `read_number` and `memstr` were added (one official program uses an `i32` memset length); that is fixed and `TestPublishedFeedCompiles` guards it. The full EDR/agent suite passed with and without the race detector; host and Windows amd64/arm64 vet passed; the whole agent module cross-built for both Windows architectures. Cloud comparison with ClamAV 1.4.3 on the same database (details in HANDOFF.md): on 82,806 Debian files (with 32-bit Wine, the 44 test files and 660 wrapped copies) no clean result where the reference detects and no detection it does not make, as with v9; on 17,460 Windows System32 and Program Files files 0 detections, and 447 of the 665 files v9 left incomplete for bytecode are now clean. Most remaining bytecode incompletes come from PE hook programs (kinds 257 and 259), which this engine does not run yet.

Validation commands from `agent`:

```sh
go test -race ./edr/... ./agent/
go vet ./edr/...
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build ./...
GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build ./...
GOOS=windows GOARCH=amd64 go vet ./edr/...
GOOS=windows GOARCH=arm64 go vet ./edr/...
go test ./edr/engine -run '^$' -fuzz=FuzzNativeContainers -fuzztime=10s -parallel=2
go test ./edr/engine -run '^$' -fuzz=FuzzExtractionFormats -fuzztime=10s -parallel=2
```

Still required before calling the original complete-replacement goal achieved:

1. Implement or explicitly renegotiate every missing signature language, extraction engine, unpacker, phishing/heuristic behavior and update protocol in the matrix. Full official feeds must load with measured, accounted-for rule coverage; no silent omissions.
2. Differential test against the existing engine using a curated malicious/benign corpus, real official databases, archived/packed/script samples, rule-count accounting and false-positive outcomes. Establish memory, throughput and latency budgets.
3. Repeat real Windows amd64/arm64 acceptance for USN, WMI, AMSI COM, quarantine/restore, process termination, ransomware, WFP, upgrade/offline behavior and UTMStack event delivery. Cross-compilation does not establish runtime success.
4. Connect scheduled server feed acquisition/signing/promotion and publish the new binaries. Validate expired/offline updates, key provisioning/rotation, least-privilege file permissions and fleet rollback.
5. Address the separate existing integration defects recorded in the architecture audit (including relay loss/backpressure, response PID identity and DNS enforcement exemptions). Their code remains outside this engine migration.

The workspace reports `partial-signature-and-format-support` deliberately. Passing the tests above does not satisfy the missing parity and effectiveness gates.
