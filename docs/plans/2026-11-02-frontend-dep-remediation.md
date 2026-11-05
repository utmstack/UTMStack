# Frontend Unmaintained-Dependency Remediation — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use subagent-driven-development (recommended) or executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Close the open security items from `docs/2026-10-27-security-review-dependabot-and-app-sweep.md` (Part 1: now, no framework change) and define the phased Angular 7→17 upgrade (Part 2) that retires the Angular-7/node-sass freeze — without forking any unmaintained dependency.

**Architecture:** Part 1 works inside the existing Node 14 / Angular 7.3 / webpack 4 toolchain: make `sass` (dart-sass) the active SCSS compiler to retire our direct EOL `node-sass`, pin `websocket-driver` 0.7.4→0.7.5 (direct pin — the one patchable CVE in scope), prove by bundle scan which flagged packages actually ship in `dist/utm-stack`, and produce a GitHub-dismissal-ready evidence memo (dismissals themselves run only with user approval). Part 2 is a phased framework upgrade (7→8→17) executed as its own workstream; this document specifies phases, version targets, the 22-library matrix, entry/exit criteria, and the gate at which Part 2 expands into a detailed task plan.

**Tech Stack:** Angular 7.3 (NgModule, TSLint, Karma/Jasmine), @angular-devkit/build-angular 0.13.10 (webpack 4, sass-loader 7.1.0), Node 14.16.1 (docker build), npm 6 (lockfileVersion 1 — npm 6 predates `overrides`, which arrived in npm 8.3; the websocket-driver fix therefore uses a **direct pin** — see Task 1.3 mechanism note), dart-sass `sass` 1.32.13, `npm ls`/dist grep for bundle proof.

**Locked decisions (do not re-open without user sign-off):**
- **No forks.** node-sass → dart-sass is the sanctioned replacement; jspdf stays 2.5.2 with the Part B½ reachability memo until Part 2's webpack 5 makes 4.x buildable.
- **Bootstrap stays on the 4.x line** — bump the lock 4.6.0 → **4.6.2** (fixes CVE-2024-6531, the Carousel XSS; 4.6.x is EOL so 4.6.2 is the last patched line) rather than a Bootstrap 5 + jQuery-removal rewrite. jQuery 3.7.1 and jquery-ui 1.13.2 are current-patched and stay.
- **Node 14 stays for Part 1** to isolate the sass change; the Node bump is Part 2 Phase 0.
- **GitHub per-alert dispositions are NOT executed by this plan.** The plan produces the evidence memo (Task 1.6); applying dismissals needs explicit user approval per the 2026-10-27 report (Part E).
- Part 2 target is **Angular 17.x (NgModule mode)** — standalone components/signals are opt-in in 17+, so no rewrite; 18/19/20 tracked as follow-on. Stop at 17 for this project.
- `moment` 2.30.1, `echarts` 4.x→5 (Part 2), and the `ngx-gauge`/`ngx-json-viewer`/`ngx-sortable-2`/`echartslayer`/`flat`/`ng-jhipster` long-tail are verified per-library in the Phase 2 matrix; unmaintained-without-alternative libs get a documented accept-risk entry, not a fork.

**Repo ground truth (verified 2026-11-02, frontend/):**
- `package.json`: 73 direct deps. Angular 7.2.0 ×7, `node-sass ^4.0.0` (lock 4.14.1), `rxjs ~6.3.3`, `zone.js ~0.8.26`, `typescript ~3.2.2` (dev), `@angular/cli ~7.3.6`, `@angular-devkit/build-angular ^0.13.8` (installed 0.13.10), `sockjs-client ^1.4.0` (lock 1.5.1, direct, browser), `stompjs ^2.3.3`, `webstomp-client ^1.2.6`, `jspdf ~2.5.2`, `highlight.js ~9.18.2`, `js-yaml ~4.3.2`, `lodash ~4.17.23`, `bootstrap ^4.3.1` (lock 4.6.0), `jquery ^3.7.1`.
- `package-lock.json` is **lockfileVersion 1** (npm 6 flat tree; hoisted transitive deps appear at the root — `websocket-driver 0.7.4`, `request 2.88.2`, `form-data 2.3.3`, `tar 2.2.2`, `cross-spawn 3.0.1/5.1.0/6.0.5`, `dompurify 2.5.9`, `canvg 3.0.11`, `fflate 0.8.3`, `html2canvas 1.0.0-rc.7` are all transitive, none direct).
- Parent chains (npm-6 `npm ls` verified, 2026-11-02 — the authoritative tree):
  - `websocket-driver@0.7.4` ← `faye-websocket@0.11.3` ← `sockjs-client@1.5.1` (direct, browser) [also `faye-websocket@0.10.0` ← `webpack-dev-server→sockjs@0.3.19`]. `faye-websocket` is a Node *server* branch — the browser build of sockjs-client never loads it, so websocket-driver is never bundled.
  - `tar@2.2.2` ← `node-gyp@3.8.0` ← `node-sass@4.14.1` (direct) **and** `node-sass@4.13.1` ← `@angular-devkit/build-angular@0.13.10` (build-angular hard-depends on node-sass). Removing our direct node-sass does NOT remove `tar@2.2.2` — build-angular's nested copy keeps it. A *separate* `tar@4.4.19` ← `pacote@9.4.0` ← `@angular/cli` is npm-internal.
  - `cross-spawn@3.0.1` ← `node-sass@4.14.1` and `node-sass@4.13.1` (build-angular). `cross-spawn@5.1.0` ← `@angular/compiler-cli→yargs@9→os-locale→execa`; `cross-spawn@6.0.5` ← `webpack-dev-server→{internal-ip, os-locale}→execa`. Removing our direct node-sass removes only the 4.14.1 copy; the 4.13.1 copy (and its cross-spawn@3.0.1) stay until Part 2.
  - `request@2.88.2` ← **multiple** parents: `node-sass@4.14.1`, `node-sass→node-gyp@3.8.0`, `less@3.9.0` (via build-angular), `protractor→webdriver-manager`. **NOT unique to node-sass** — it persists after node-sass removal (via less/protractor, both dev-time).
  - `form-data@2.3.3` + `tough-cookie@2.5.0` ← `request` (so they persist with it).
- SCSS: 11 `~`-import sites; 5 of them use `~src/assets/styles/...` (theme/var/custom-elements) and 6 use `~<node-module>/...`. **Empirically verified 2026-11-02: all 11 resolved unchanged under dart-sass** (build green, zero SCSS edits) — the relative-path rewrites in Task 1.2 Step 3 are a *fallback* for if a future toolchain change breaks them, not a required step. The 3 regex "division" hits are URLs in comments — **no real SCSS division exists**.
- Build command (Node 14 docker): `NODE_OPTIONS=--max_old_space_size=8192 npm run build` → `ng build --prod --base-href /` → `dist/utm-stack`.
- 22 third-party Angular-7-locked libs (matrix in Part 2 Phase 2).

---

# PART 1 — Now (no framework change; ~1–3 days)

## Task 1.1: Baseline evidence — production tree + dist markers (pre-change)

**Files:**
- Create: `docs/security/deps/2026-11-02-baseline.md` (evidence file, committed)
- No source files modified.

- [x] **Step 1: Record the production-tree view of the disputed packages**

Run in the frontend dir with Node 14 (local Node 26 is unusable for npm-6 lock v1 installs — use the docker build env):

```bash
docker run --rm -v /d/Projects/github.com/utmstack/UTMStack/frontend:/app -w /app node:14.16.1 \
  sh -c "npm ls websocket-driver tar request form-data cross-spawn faye-websocket node-sass 2>&1 | tee /app/.deps-baseline.txt"
```

Expected: tree output showing every parent chain (e.g. `sockjs-client → faye-websocket → websocket-driver@0.7.4`; `node-sass → request@2.88.2 → form-data@2.3.3/tough-cookie@2.5.0`; `node-sass → cross-spawn@3.0.1`). `npm ls` exits non-zero for some (invalid deps) — capture stderr, that's the evidence.

- [x] **Step 2: Build the app and record which disputed packages ship in dist**

```bash
docker run --rm -v /d/Projects/github.com/utmstack/UTMStack/frontend:/app -w /app node:14.16.1 \
  sh -c "NODE_OPTIONS=--max_old_space_size=8192 npm run build 2>&1 | tail -5"
```

Then probe `dist/utm-stack` for unique-in-package marker strings (minified code keeps these literals):

```bash
cd frontend
for m in \
  "websocket-driver" "faye-websocket" \
  "form-data" "combinedLength" \
  "request.js" "tough-cookie" \
  "tar.js" "Tar::" \
  "cross-spawn" \
  "html2canvas" "jspdf"; do
  printf "%-22s " "$m"; grep -rlF "$m" dist/utm-stack/*.js 2>/dev/null | tr '\n' ' '; echo
done
```

Expected: `html2canvas`/`jspdf` present (runtime, via `export-to-pdf-util`); the other 8 markers **absent** from `dist` (they are Node-only build/install tooling). Record exact output.

- [x] **Step 3: Write the baseline evidence file**

Create `docs/security/deps/2026-11-02-baseline.md` with the two captured outputs verbatim, one line per disputed package: `package | version | npm-ls parent chain | present in dist? (marker file or none)`. This is the "before" half of the dismissal memo.

- [x] **Step 4: Commit**

```bash
git add docs/security/deps/2026-11-02-baseline.md
git commit --no-gpg-sign -m "chore(security): baseline dep-tree + dist-bundle evidence for disputed frontend packages"
```

## Task 1.2: Make dart-sass the active SCSS compiler (removes our direct EOL node-sass)

**Files:**
- Modify: `frontend/package.json` (deps)
- Modify: `frontend/package-lock.json` (regenerated by the reinstall)
- Modify (only if the Step 2 build flags one): `frontend/src/app/active-directory/shared/components/active-directory-tree/active-directory-tree.component.scss:1`, `frontend/src/app/scanner/scanner-config/target/target-list/target-list.component.scss:1-3`, `frontend/src/app/shared/components/layout/footer/footer.component.scss:1-2`

**What it does (verified 2026-11-02) and does NOT:**
- **Does:** makes `sass` (dart-sass) the **active SCSS compiler** and removes our **direct** `node-sass` (lock 4.14.1, spec `^4.0.0`, EOL — npm deprecates it: "use `sass` or `sass-embedded`") from `package.json`. Proof it took effect: `build-angular require('sass')` resolves and `s.info` reads `dart-sass <ver> ... [Dart]`; the `styles.*.css` chunk hash changes vs the node-sass build; the build compiles our SCSS with no `node-gyp`/binding errors.
- **Does NOT (corrected after live `npm ls`):** remove `tar@2.2.2`, remove node-sass from the tree, or fully unlock the Node-14 freeze. **`@angular-devkit/build-angular@0.13.10` declares `node-sass@4.13.1` as a hard dependency**, and that copy keeps `node-gyp@3.8.0 → tar@2.2.2` + `cross-spawn@3.0.1` in the dev tree and still compiles a native binding on `npm install`. Those leave only when Part 2 upgrades `build-angular` (later majors drop the node-sass dep). So Task 1.2 is a *prerequisite* for leaving Node 14, not the unlock itself.
- `request@2.88.2`/`form-data@2.3.3`/`tough-cookie@2.5.0` persist after the swap (parents: build-angular→node-sass→node-gyp, less, protractor) — dev-only, never bundled (proven in Task 1.1/1.4).

**Why the swap is guaranteed to take effect (verified in the INSTALLED build-angular 0.13.10, `src/angular-cli-files/models/webpack-configs/styles.js`):** it does `let dartSass; try { dartSass = require('sass'); } catch {}` and passes `options.implementation: dartSass` to `sass-loader`. When `sass` is installed, `dartSass` is the dart-sass module and sass-loader 7.1.0 uses it (`getRenderFuncFromSassImpl` accepts `dart-sass ^1.3.0`); it does **not** fall back to node-sass. A stale `node_modules` would still have the old lock, so Step 2's build must start from `rm -rf node_modules`.

**On the 11 `~` imports — handle empirically, don't guess.** All 11 (`~src/assets/styles/...` ×5 and `~<package>/...` ×6) go through sass-loader's `webpackImporter`, which is shared by node-sass and dart-sass; neither form matches its module-import regex, so both defer to webpack's resolver. Whether `~` resolves identically under dart-sass can't be traced statically — so **build first with the imports unchanged**; if it fails, the error names the exact `@import` and you rewrite just that one to a relative/package path. Ready-made replacements for all 11 are in Step 2 (use them only for imports the build flags).

- [x] **Step 1: Replace the dependency**

Edit `frontend/package.json` `dependencies`: remove `"node-sass": "^4.0.0"` and add:

```json
"sass": "1.32.13",
```

(**Exact pin, no caret.** `^1.32.0` would let npm 6 resolve to the latest 1.x — it pulled **1.107.1** on first run, a 2025 release with deprecation behavior and a Node-engine floor that don't belong in this frozen Angular 7 / Node 14 build. `1.32.13` is the last 1.32.x patch: pre-dates the `@use`-vs-`@import` and division-by-slash deprecation *warnings* (those start pinging from 1.33+) while still exposing the legacy `render`/`renderSync` API sass-loader 7.1.0 calls. Build must stay clean, not just green. If the build ever throws a dart-sass incompatibility, this exact pin is the fallback that predates it.)

- [x] **Step 2: Reinstall and build (imports UNCHANGED first)**

Because sass-loader's `webpackImporter` is shared by node-sass and dart-sass (same `~` handling), the build may succeed with zero SCSS edits. Build first; only rewrite if the error names an import.

```bash
docker run --rm -v /d/Projects/github.com/utmstack/UTMStack/frontend:/app -w /app node:14.16.1 \
  sh -c "rm -rf node_modules && NODE_OPTIONS=--max_old_space_size=8192 npm install --no-audit --no-fund 2>&1 | tail -3 && NODE_OPTIONS=--max_old_space_size=8192 npm run build 2>&1 | tail -25"
```

Expected: install completes with `sass` present and **no** `node-gyp`/native build output; build succeeds (`chunk {main} ... [emitted]`, "Compiled successfully").

- [ ] **Step 3: IF the build fails to resolve a `~` import — rewrite only the flagged one**

A dart-sass build can surface `~src/...` (root-relative, non-`node_modules`) as unresolvable where node-sass resolved it. The 5 such sites, with the ready relative-path replacement (all 3 files sit 5 levels below `src/`, so `src/assets/styles/*` = five `../`; four would land at `src/app/assets/...` and fail):

`src/app/active-directory/shared/components/active-directory-tree/active-directory-tree.component.scss` line 1:
```scss
@import "~src/assets/styles/theme";
```
→
```scss
@import "../../../../../assets/styles/theme";
```

`src/app/scanner/scanner-config/target/target-list/target-list.component.scss` lines 1-3:
```scss
@import "~src/assets/styles/theme";
@import "~src/assets/styles/var";
@import "~src/assets/styles/custom-elements";
```
→
```scss
@import "../../../../../assets/styles/theme";
@import "../../../../../assets/styles/var";
@import "../../../../../assets/styles/custom-elements";
```

`src/app/shared/components/layout/footer/footer.component.scss` lines 1-2:
```scss
@import "~src/assets/styles/theme";
@import "~src/assets/styles/var";
```
→
```scss
@import "../../../../../assets/styles/theme";
@import "../../../../../assets/styles/var";
```

The 6 `~<package>/...` sites (`~@ng-select/...`, `~font-awesome/...`, `~highlight.js/...`, `~bootstrap/...` ×2) resolve real `node_modules` paths and are expected to work unchanged; rewrite one to a bare `@import "<package>/..."` (or relative `node_modules/...`) only if the build specifically flags it. Rebuild after each edit. (If a build error names an unresolvable import, the path depth for that one line is off — adjust that line only and rebuild. Target is always `src/assets/styles/*`.)

- [x] **Step 4: Prove the direct node-sass is gone and dart-sass is active**

```bash
docker run --rm -v /d/Projects/github.com/utmstack/UTMStack/frontend:/app -w /app node:14.16.1 \
  sh -c "echo '--- direct node-sass (expect ABSENT at top level):'; npm ls node-sass 2>&1 | grep -E 'node-sass@' ; echo '--- active sass impl build-angular will use:'; node -e \"try{const d=require('sass');console.log(d.info)}catch(e){console.log('sass NOT resolvable:',e.message)}\"; echo '--- sass version (expect 1.32.13):'; node -e \"console.log(require('/app/node_modules/sass/package.json').version)\" 2>/dev/null || grep -m1 '\"version\"' node_modules/sass/package.json"
```

Expected:
- `npm ls node-sass` shows **no top-level `node-sass@4.14.1`**. It MAY still show `@angular-devkit/build-angular@0.13.10 → node-sass@4.13.1` — that's the build system's own hard dep and stays until Part 2. It is **inert**: build-angular's `styles.js` passes `implementation: require('sass')` (dart-sass), so the nested node-sass is never invoked for our SCSS.
- `require('sass').info` prints `dart-sass 1.32.13 … (Sass Compiler) [Dart]` — proves dart-sass is the active implementation and the pin held.
- `dist/utm-stack` still contains `index.html`, `main.*.js`, `styles.*.css` (from Step 2's build).

(Do NOT claim `tar@2.2.2` or `cross-spawn@3.0.1` are gone — they persist via build-angular's nested `node-sass@4.13.1 → node-gyp@3.8.0`; that's a Part 2 outcome. `request`/`form-data`/`tough-cookie` likewise persist via `less`/`protractor`. All are dev-only and never bundled — Task 1.4 settles them by dist evidence, not by tree removal.)

- [x] **Step 5: Visual smoke (lab)**

Slam the new dist onto the lab frontend (established pattern: slim image `FROM utmstack-frontend-local:v11.3.1` + `COPY dist/utm-stack`), `docker service update --force`, then open `/dashboard/render/227/aws-cloud` headlessly and assert `window.echarts` defined + ≥1 chart rendered + 0 pageerrors. (Charts prove the CSS/JS bundles are intact — a broken styles import usually surfaces as a visibly unstyled layout; screenshot it.)

- [x] **Step 6: Commit**

```bash
git add frontend/package.json frontend/package-lock.json \
  frontend/src/app/active-directory/shared/components/active-directory-tree/active-directory-tree.component.scss \
  frontend/src/app/scanner/scanner-config/target/target-list/target-list.component.scss \
  frontend/src/app/shared/components/layout/footer/footer.component.scss
git commit --no-gpg-sign -m "build(frontend): make dart-sass the active SCSS compiler; drop direct EOL node-sass@4.14.1"
```

Note: the nested `node-sass@4.13.1` under `@angular-devkit/build-angular@0.13.10` (and its `node-gyp→tar@2.2.2`) is NOT removed by this task — build-angular hard-depends on it. It is inert (never invoked: `styles.js` passes `implementation: require('sass')`) and leaves the tree in Part 2 when build-angular is upgraded.

## Task 1.3: Patch websocket-driver 0.7.4 → 0.7.5 (the critical)

**Files:**
- Modify: `frontend/package.json` (adds `"websocket-driver": "^0.7.5"` as a pinned direct dep — see mechanism note)
- Modify: `frontend/package-lock.json`

**Why:** CVE on `websocket-driver < 0.7.5` (protocol-level DoS). Chain: `sockjs-client`(direct, browser) → `faye-websocket` (sockjs-client's Node-only *server* branch — the browser build never `require`s faye-websocket, so `websocket-driver` is never bundled; the dist scan in Task 1.4 proves it) → `websocket-driver`. Both `faye-websocket@0.10.0` and `@0.11.3` declare `websocket-driver >=0.5.1` with **no upper cap** (verified against the npm registry 2026-11-02), so 0.7.5 satisfies both.

**Mechanism (corrected during execution, 2026-11-02):** `npm update websocket-driver` did **NOT** work — npm 6's lock had `websocket-driver@0.7.4` recorded as the resolved root entry, and `update` honors the lock over the range. `npm install websocket-driver@0.7.5 --no-save` installed 0.7.5 in `node_modules` but left the lock at 0.7.4 (UNMET DEPENDENCY, inconsistent). The working npm 6 mechanism for forcing a transitive patch is to **declare it as a direct dep**: `npm install websocket-driver@0.7.5` (with save). This adds `"websocket-driver": "^0.7.5"` to `package.json`, rewrites the lock root entry to 0.7.5, and makes both faye-websocket copies dedupe to it. (npm 6 has no `overrides` — that feature arrived in npm 8.3 — so the direct pin is the only reproducible option; a caret range is used so future 0.7.x patches can flow, and the lock records the exact 0.7.5.)

- [x] **Step 1: Install 0.7.5 as a pinned direct dep**

```bash
docker run --rm -v /d/Projects/github.com/utmstack/UTMStack/frontend:/app -w /app node:14.16.1 \
  sh -c "npm install websocket-driver@0.7.5 --no-audit --no-fund 2>&1 | tail -4 && npm ls websocket-driver 2>&1"
```

Expected: `package.json` gains `"websocket-driver": "^0.7.5"`; `npm ls websocket-driver` shows **0.7.5 everywhere** (both faye-websocket copies dedupe to it), no UNMET lines.

- [x] **Step 2: Rebuild + confirm the lock is consistent**

```bash
docker run --rm -v /d/Projects/github.com/utmstack/UTMStack/frontend:/app -w /app node:14.16.1 \
  sh -c "NODE_OPTIONS=--max_old_space_size=8192 npm run build 2>&1 | tail -5"
git -C /d/Projects/github.com/utmstack/UTMStack diff --stat frontend/package.json frontend/package-lock.json
```

Expected: build green; `package.json` diff = the one added `websocket-driver` line; lock diff = the root `websocket-driver` entry 0.7.4→0.7.5 (+ integrity/shasum), no other version churn.

- [x] **Step 3: Commit**

```bash
git add frontend/package-lock.json
git commit --no-gpg-sign -m "chore(deps): websocket-driver 0.7.5 (CVE, via faye-websocket range; Node-only branch, not bundled)"
```

## Task 1.4: Bundle scan — settle tar / request / form-data / cross-spawn

**Files:**
- Create: `docs/security/deps/bundle-scan.md` (evidence)

**Why:** the report left these four "bundle scan decides" (B3 Class 2 rows). Task 1.1 recorded the pre-change dist markers; this task produces the formal scan result and per-package disposition.

- [x] **Step 1: Re-run the dist marker probe from Task 1.2's built dist**

```bash
cd frontend
for m in "faye-websocket" "websocket-driver" "form-data" "combinedLength" "tough-cookie" "request.js" "Tar::" "cross-spawn"; do
  printf "%-20s " "$m"; grep -rlF "$m" dist/utm-stack/*.js 2>/dev/null | tr '\n' ' '; echo
done
```

Expected: all absent (Node-only install/build branches). If any marker IS present, stop and investigate which webpack module pulled it (search the dist chunk for the neighboring module header) before continuing — that changes the disposition from "dismiss" to "fix now".

- [x] **Step 2: Write the disposition table**

Create `docs/security/deps/bundle-scan.md`:

```markdown
# Frontend disputed-package bundle scan (2026-11-02, post node-sass removal)

Build: ng build --prod, Node 14.16.1, dist/utm-stack (main.*.js, chunks, styles)

| Package | Version | Declares-it (npm registry) | Shipped in dist? | Disposition |
|---|---|---|---|---|
| websocket-driver | 0.7.5 | faye-websocket (dep of sockjs-client, Node server branch only) | NO (marker absent) | PATCHED (Task 1.3); never bundled |
| tar | 2.2.2 | node-gyp@3.8.0 ← node-sass@4.13.1 ← @angular-devkit/build-angular | NO (marker absent) | PERSISTS in dev tree (build-angular hard-depends on node-sass@4.13.1; Task 1.2 removes only the direct 4.14.1) — DISMISS: dev-time install tooling, not bundled; leaves tree in Part 2 when build-angular is upgraded. (Separate tar@4.4.19 ← pacote ← @angular/cli also npm-internal, never bundled.) |
| request | 2.88.2 | node-sass, node-gyp, less@3.9.0, protractor→webdriver-manager | NO (marker absent) | PERSISTS in dev tree (less/protractor parents survive) — DISMISS as dev-time tooling, not bundled |
| form-data | 2.3.3 | request | NO (marker absent) | PERSISTS with request — DISMISS as dev-time, not bundled |
| tough-cookie | 2.5.0 | request | NO (marker absent) | PERSISTS with request — DISMISS as dev-time, not bundled |
| cross-spawn | 3.0.1 (node-sass@4.13.1) / 5.1.0+6.0.5 (compiler-cli→yargs→os-locale→execa, webpack-dev-server) | build-angular→node-sass, compiler-cli, webpack-dev-server | NO (marker absent) | ALL PERSIST (dev parents survive Task 1.2) — DISMISS as dev-only build tooling, not bundled |
| jspdf | 2.5.2 | direct (export-to-pdf-util) | YES | ACCEPT RISK: 12 CVEs unreachable per docs/2026-10-27… Part B½ trace; closes in Part 2 (webpack 5 → jspdf 4.x) |
| bootstrap | 4.6.0 → 4.6.2* | direct | YES | Bump in Part 2 Phase 0 (fixes CVE-2024-6531 Carousel XSS, same major, no API change) |

*bootstrap 4.6.2 not yet applied — listed as Phase 0 Part 2 action.
```

(Fill the `request`/`form-data` chain cells with the actual Task 1.1 `npm ls` output, not the parenthetical.)

- [x] **Step 3: Commit**

```bash
git add docs/security/deps/bundle-scan.md
git commit --no-gpg-sign -m "docs(security): frontend bundle-scan evidence + dispositions for disputed transitive deps"
```

## Task 1.5: GitHub dismissal memo (approval-gated)

**Files:**
- Create: `docs/security/deps/dismiss-memo.md`

- [x] **Step 1: Produce the per-alert dismissal text**

**Executed 2026-11-02:** live Dependabot pull (100 open alerts repo-wide; 57 frontend npm across 25 packages) dispositioned in `docs/security/deps/dismiss-memo.md` as A=2 (websocket-driver 970/971 — patched, auto-closes) / B=20 (framework + accepted-risk, keep open) / C=35 (dev/build-scoped, proposed dismiss). Set diff verified A+B+C=57 exactly.

```markdown
# Dependabot dismissal memo (apply with user approval only)

Evidence: docs/security/deps/2026-11-02-baseline.md + bundle-scan.md

| Dep | CVEs | Dismiss reason |
|---|---|---|
| <pkg> | <ids> | dev-time tooling only; not present in dist/utm-stack (marker scan 2026-11-02); no runtime path |
...
```

- [ ] **Step 2: Present to user; apply ONLY on explicit approval**

If approved: `gh api repos/utmstack/UTMStack/dependabot/alerts/<n>/dismiss -f state=dismissed -f reason=...` per alert. Never batch without approval (report Part E rule).

## Task 1.6: Part 1 exit verification

- [x] **Step 1: Full frontend build + test in Node 14 docker**

```bash
docker run --rm -v /d/Projects/github.com/utmstack/UTMStack/frontend:/app -w /app node:14.16.1 \
  sh -c "NODE_OPTIONS=--max_old_space_size=8192 npm run build 2>&1 | tail -5 && npm test 2>&1 | tail -8"
```

Expected: build green; Karma suite same pass/fail profile as the pre-change baseline (record both — a NEW failure is a stop-the-world regression).

- [x] **Step 2: Confirm the Part 1 CVEs are closed**

`npm ls websocket-driver` → 0.7.5 (the protocol-level critical, via `faye-websocket >=0.5.1`); `npm ls tar` → no `tar@2.2.2` (its parent `node-gyp` left with node-sass); `npm ls node-sass` → empty. `request`/`form-data`/`tough-cookie` may still list (their `less`/`protractor` parents survive) — that's expected; they are DISMISSED as dev-only (bundle-scan proves they don't ship), not claimed absent. Record results in `docs/security/deps/bundle-scan.md` footer.

**ACTUAL RESULT (2026-11-02):** `npm ls websocket-driver` → 0.7.5 in all three tree positions (PASS). `npm ls tar` → **2.2.2 still present** (dev tree via build-angular→node-sass@4.13.1→node-gyp) — the "no 2.2.2" expectation is wrong, see correction below; 4.4.19 (pacote) also present, npm-internal. `npm ls node-sass` → only the nested build-angular copy 4.13.1 (no direct 4.14.1). `request@2.88.2` still listed (less/protractor/node-sass chains) — expected, dev-only, never bundled.

> **Plan correction (2026-11-02, found at exit verification):** the original expectation "`npm ls tar` → no 2.2.2" is **not achievable in Part 1**. `tar@2.2.2` is held by `node-gyp@3.8.0`, whose only project-tree parent is `node-sass@4.13.1` — which is a **hard dependency of `@angular-devkit/build-angular@0.13.10` itself**, not of our removed direct `node-sass@4.14.1`. Removing our direct node-sass does not evict build-angular's nested copy, so `tar@2.2.2` and `cross-spawn@3.0.1` remain in the dev tree until Part 2 upgrades build-angular to a version that dropped node-sass. The security claim that matters is unaffected and is proven: none of these packages ship in `dist/utm-stack` (bundle-scan marker scan, all ABSENT). Part 1 exit criteria are met on the CVEs that Part 1 can actually close: **websocket-driver 0.7.5 (the critical) patched**; the rest are dev-only/non-bundled and tracked for Part 2.

- [x] **Step 3: Commit any stragglers, tag the Part 1 state**

```bash
git add -A docs/security/deps/
git commit --no-gpg-sign -m "chore(security): Part 1 exit state - criticals closed, bundle-scan evidence complete"
```

---

# PART 2 — Framework upgrade (tracked project; ~3–6 weeks)

**Not task-expanded here.** Each phase below has entry/exit criteria; when a phase starts, it expands into its own detailed task plan (same skill, same branch strategy) because the per-library API diffs must be measured against the actual installed majors, not pre-estimated. This document is the contract: targets, ordering, matrix, gates.

**Working branch:** off `fix/ir-role-model-v11.3.1` (or a fresh `feat/frontend-17` off v11 head once Part 1 lands) — decided at Part 2 kickoff. NEVER stash the main checkout (user WIP stash standing rule).

## Phase 0 — De-risk foundation (2–4 days)

- **Entry:** Part 1 complete (dart-sass in, build green on Node 14).
- **Actions:**
  1. `bootstrap ^4.3.1` (lock 4.6.0) → `^4.6.2` (fixes CVE-2024-6531 Carousel XSS; same major, zero API change for the components UTMStack uses). Build + smoke.
  2. Move the build to **Node 18.20** docker image (dart-sass already removed the Node-14 lock; ng 7 build on Node 18 may need `--openssl-legacy-provider` via `NODE_OPTIONS` for webpack 4's md4 hash — if so, record it as the Node-18 build flag; if it hard-fails, Phase 0 stays on Node 14 and Node bumps in Phase 2 alongside Angular 12+ which dropped md4).
  3. Establish the visual regression baseline: screenshot the 10 highest-traffic routes (login, log explorer, alert detail, dashboards 7/13/227/403/405, rule editor, integrations, SOAR, compliance) at desktop+mobile widths — stored under `docs/security/deps/visual-baseline/`.
  4. Karma suite green on the new Node (jasmine 2.99/jasminewd2 may need a compat pin — fix in this phase, not mid-upgrade).
- **Exit:** build + tests green on the Phase-0 Node; visual baseline captured; bootstrap 4.6.2 committed.

## Phase 1 — Angular 7 → 8 (the one true break; 3–5 days)

> **Expanded 2026-11-03 into a measured task plan:** `docs/plans/2026-11-03-angular8-upgrade.md`
> (registry-verified targets, the capped libs, the DI note, and 3 evidence-based
> deviations from this phase's text — TS 3.2.2→3.5.3 (v8 compiler floor), no providedIn
> refactor, 7 libs bump now). DONE 2026-11-03 (commits 5841b46916 → 450c44e229, incl. the
> child-query `{static}` regression fix 106b1faa8e); residual node-sass/tar@2.2.2 closed.

- **Entry:** Phase 0 exit.
- **Actions (all via `ng update @angular/cli @angular/core @angular/compiler @angular/compiler-cli @angular/common @angular/forms @angular/platform-browser @angular/platform-browser-dynamic @angular/router @angular/language-service` then hand-fixes):**
  1. Static providers: `@Injectable()` → `@Injectable({providedIn: 'root'})` for the ~service set (codemod `ng generate provider`/manual; 1486 .ts files but only `@Injectable` declarations change).
  2. `ViewEncapsulation` / `ngModule` metadata unchanged; v8 removed `@angular/http` (already unused — verify `grep -r "@angular/http" src` = 0) and the old `ng` package (verify unused).
  3. `rxjs 6.3.3` stays (v8 is rxjs 6.x — no change); `zone.js ~0.9.1`; `typescript ~3.8`; `@angular-devkit/build-angular 0.13.10` → `8.3.x` (still webpack 4).
  4. TSLint → **keep** (ESLint migration is Phase 2, not 8 — TSLint still ships in 8).
- **Exit:** build green; Karma green; visual baseline diff = zero material change; commit `chore(frontend): angular 8 baseline`.

## Phase 2 — Angular 8 → 17 via staged `ng update` (1–3 weeks)

- **Entry:** Phase 1 exit.
- **Actions:**
  1. Stage through majors (8→9→…→17) running `ng update` at each stop, committing per major. Angular's codemods handle DI injection tokens, `HttpClient` generics (9), `NgIf`/trackBy (12+), standalone-off compat (13–17 keep NgModule).
  2. **TypeScript 5.2/5.3, rxjs 7.5, zone.js 0.14, karma/jasmine modernization** ride along at their respective majors.
  3. **The 22-library matrix** — bump at the stage where Angular compat allows; known breaking notes:

  | Lib (installed) | Target (Ang 17) | Known breaks |
  |---|---|---|
  | @ng-bootstrap/ng-bootstrap 4 | 15 | v5+ requires `ngbModal` config inject, tooltips changed to components |
  | @ng-select/ng-select 2 | 12 | v5+ `ng-select` config API stable; theming rework — visual QA pass |
  | angular-gridster2 8 | 14 | config object shape stable; v11+ dropped `ngDefault` — grep usage |
  | @kolkov/angular-editor 0.18 | 4 | v2+ `editor` binding API stable; quill version bump internal |
  | @ngx-translate/core 11 / http-loader 4 | 15 / 8 | loader injectable provided change (14+) — codemod |
  | @fortawesome/angular-fontawesome 0.3 | 0.14 | `fa-icon` inputs stable |
  | ng-inline-svg 8 | 14 | module name stable |
  | ng-jhipster 0.9 | latest (0.11) | JHipster helpers — verify each used util still exported |
  | ng2-tel-input 1 | 2.5 | v2 requires Angular 15+ — bump at 15 stage |
  | ng5-slider 0.1.x | 1.4 | v1 API: `[(value)]` → `[value]`+`(valueChange)` — grep ~N uses |
  | ng6-toastr-notifications 1.0 | 2.0 | v2 config token — small |
  | ngx-cookie-service 2.4 | 17 | API stable |
  | ngx-drag-drop 2 | 14 | API stable |
  | ngx-draggable-resizer 1.0 | verify latest (possibly unmaintained) | if unmaintained at Ang 17: accept-risk + pin, or replace drag handles with native `ng-resize` — decision at Phase 2 |
  | ngx-echarts 4 | 17 | **echarts 4→5** API: `echarts.init` same; `bindGlobal` removed; visual QA on all 9 map charts + scatter |
  | echarts-gl 1.1 / echarts-leaflet 1.1 / echarts-stat 1.1 / echarts-wordcloud 1.1 | matching echarts-5 versions | load-order with echarts 5 — verify extensions register |
  | echartslayer 0.1.6 | verify (old) | if incompatible with echarts 5: check usage sites; possibly dead — grep first |
  | ngx-flag-icon-css 0.1 | 1 | trivial |
  | ngx-gauge 0.1-beta.12 | verify (low activity) | if incompatible: the 2–3 gauge usages can be hand-rolled on ngx-echarts — decision at Phase 2 |
  | ngx-infinite-scroll 7 | 14 | API stable |
  | ngx-json-viewer 2.4 | verify latest | small surface |
  | ngx-monaco-editor 7 | 18 | monaco-editor 0.20 → 0.4x: worker loading via `worker-loader`→`monaco-editor/esm/vs/...` ESM workers — biggest single lib effort; Log Explorer + rule editor are the consumers |
  | ngx-sortable-2 1.0.1 | verify (low activity) | sortable grid in dashboard builder — QA gate |
  | ngx-spinner 7 | 16 | API stable |
  | ngx-webstorage 2 | 14 | API stable |
  | monaco-editor 0.20 | 0.45+ | same as ngx-monaco-editor row |
  | popper.js 1.15 | 2.11 | **only if** any code imports popper directly (bootstrap 4.6 uses it internally via its own bundle — verify; if unused direct, drop the dep) |
  | interactjs 1.4-rc.13 | 1.10 | API stable; RC→stable |
  | html2canvas 1.0.0-rc.7 | 1.4.x | same API; re-run the dashboard PDF smoke (the Part F½ hang is monitored, not fixed) |

  4. **ESLint** replaces TSLint at the 12+ stage (angular-eslint; codelyzer rules → eslint-plugin-jsx? no — `@angular-eslint` has the equivalents; tslint.json `codelyzer` rules map 1:1 to `@angular-eslint` recommended).
- **Exit:** Angular 17 build green on **Node 18+** (webpack 5 — md4 gone, no legacy-provider flag needed), Karma green, **visual baseline diff reviewed route-by-route** (every difference explained or fixed), all 22 libs at target or formally accept-risked with a written line. Commit per major + one per library-cluster.
- **STATUS: COMPLETE (2026-11-05).** S9-S17 all committed on `fix/ir-role-model-v11.3.1`
   (S17 = `bf5e811343`). Final framework 17.3.12 / cli+ba 17.3.17 / TS ~5.4.5 /
   Node 18.20.8. Clean CI-replica `npm ci`+build GREEN and byte-deterministic
   (`main.1abf6bd72256cff4.js`, 0 raw-VE, lockfileVersion 2 preserved); lab smoke
   SMOKE_PASS on the final tree (9 routes ng=17.3.12 0 pageerrors, echarts 5
   charts, monaco SQL completion, toastr, AD phone, flags, resize handles,
   multi-page pagination). Bootstrap stays 4.6.2 + ng-bootstrap 11.0.1 (last
   Bootstrap-4 line) per decision; `legacy-peer-deps` retained for 4 sub-17 core
   peers (documented in `frontend/.npmrc`). Per-stage detail in
   `docs/plans/2026-11-03-angular17-staged-upgrade.md`.

## Phase 3 — Close the jspdf cluster (2–4 days)

- **Entry:** Phase 2 exit (webpack 5 is what blocked jspdf 3/4 — `fast-png@6` ES2022 now parses).
- **Actions:** `jspdf ~2.5.2` → `^2.5.2` re-evaluate: if 2.5.2's 12 CVEs are fixed in a 2.x patch, take it; otherwise `jspdf 4.x` + re-run `export-to-pdf-util.ts` (API: `new jsPDF()`, `addImage`, `output('blob')` — stable across 2→4) + the Part F½ in-browser PDF E2E (valid `%PDF`, real dashboard export). `dompurify/canvg/fflate/html2canvas` follow the lock diff; re-run the bundle scan.
- **Exit:** 0 open jspdf-cluster alerts (or written accept-risk for any residual); PDF export E2E green.
- **STATUS: COMPLETE (2026-11-05, after S17 = webpack 5 + Node 18).** `jspdf ~2.5.2`
   -> `~4.2.1`. The 2.x line ended at 2.5.2 (no 2.x patch exists) and all 12 open
   jspdf advisories are only fixed in 3.x/4.x - programmatically verified 12/12
   CLOSED by 4.2.1 (0 remain; 2 were critical). Build blocker resolved: S17 put us
   on webpack 5, which parses fast-png@6 ES2022 (the 2.5.2-pin rationale was
   webpack 4 only). fast-png@6.4.0 = 0 advisories. API is a drop-in for our usage
   (export-to-pdf-util.ts: import jsPDF, new jsPDF(p,mm,a4), addImage(PNG),
   output(blob), save - exact ctor signature + members confirmed against 4.2.1
   types). 4.2.1 also DROPS canvg/dompurify optionalDeps (we never used fromHTML)
   and keeps fflate - cluster SHRINKS. Runtime E2E: the exact bundled 4.2.1 UMD
   driven through the app call sequence in a real Chromium produced a valid
   %PDF-1.3 (111,671 bytes, blob type application/pdf). NOTE: the server-side
   Save-to-PDF dashboard button (exportToPdf -> /api/generate-pdf-report
   web-pdf service) does NOT use client jspdf; the client path is
   customize-export (captureScreen/pdfPreview) whose html2canvas-on-chart-grid
   headless hang is the known pre-existing #2934 behavior, not a jspdf
   regression. Clean CI-replica on the final Phase 3 tree: npm ci (lockfile
   stays v2) + build 0 errors, deterministic.

## Phase 4 — Close-out (1–2 days)

- **Entry:** Phase 3 exit.
- **Actions:** final full bundle scan (all 332-alert inventory re-checked); GitHub dispositions (user approval gate, per Task 1.5 pattern); delete `docs/security/deps/` interim artifacts except the final memo; update the 2026-10-27 report's Part D to DONE-state with commit links; **release** as a v12-line (or v11.x-frontend-major, decided with user — the Angular 17 frontend changes the dist contract, so it wants its own release tag).
- **Exit:** 0 open findings in the report except the written accept-risks (moment, any long-tail lib from the matrix); Part 2 documented complete.

## Effort roll-up
| Phase | Est. |
|---|---|
| 0 | 2–4 days |
| 1 | 3–5 days |
| 2 | 1–3 weeks (monaco + echarts + ng-bootstrap are the heavy ones) |
| 3 | 2–4 days |
| 4 | 1–2 days |
| **Total** | **~3–6 weeks**, one engineer |

## Explicit non-goals
- No fork of any third-party package (user decision, 2026-11-02).
- No Bootstrap 5 / jQuery removal (bootstrap 4.6.2 is the CVE fix).
- No standalone-component / signals conversion (Angular 17 stays NgModule).
- No Angular 18/19/20 (tracked as follow-on after 17 lands).
- No backend changes in this plan.
