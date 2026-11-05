# Phase 2 — Angular 8 → 17 (detailed task plan, measured 2026-11-03)

> Expanded from `docs/plans/2026-11-02-frontend-dep-remediation.md` Part 2 Phase 2
> per its gate. The framework/toolchain ladder below is **registry-measured**
> (build-angular last-patch + its TS peer per major), not pre-estimated.

**Goal:** Move `frontend/` from Angular 8.2 to **Angular 17.x (NgModule mode)** —
webpack 5 (md4 gone, no legacy-provider flag), rxjs 7, zone.js 0.14, TS 5.2 —
with every third-party lib at the master plan's matrix target or formally
accept-risked, and a route-by-route visual diff explained.

**Branch:** `fix/ir-role-model-v11.3.1` (Phase 1 closed at `1e53b79272`).
NEVER stash the main checkout.

**Method (per the master plan):** stage through majors 8→9→…→17, `ng update`-equivalent
manual bumps at each stop, commit per major. Framework + build-angular + TS + rxjs +
zone.js bump at EVERY stage (forced). Third-party libs bump **only when** (a) their
`@angular/core` peer no longer admits the new major AND the build breaks, or (b) the
master plan's 22-lib matrix explicitly stages them at that major. npm 6 warns (not
fails) on peer mismatches, so a lib that still compiles can ride a few majors before
it's forced — bump it then.

**Build host:** Node 18.20 with `NODE_OPTIONS=--openssl-legacy-provider` for
build-angular 0.803–0.1102 (webpack 4, majors 8/9/10/11). From **Angular 12
(build-angular 12.x, webpack 5)** the flag is no longer needed. Install with npm 6
(Node 14) to keep lockfileVersion 1.

## Framework + toolchain ladder (registry-measured, 2026-11-03)

| Major | @angular/* | cli | build-angular | TS (peer) | zone.js | rxjs |
|---|---|---|---|---|---|---|
| 8 (now) | 8.2.14 | 8.3.29 | 0.803.29 | ~3.5.3 (>=3.1 <3.6) | 0.9.1 | 6.6.7 |
| 9 | 9.1.13 | 9.1.15 | 0.901.15 | ~3.8 (>=3.6 <3.9) | 0.10.3 | 6.6.7 |
| 10 | 10.2.5 | 10.2.4 | 0.1002.4 | ~3.9 (>=3.9 <4.1) | 0.11.4 | 6.6.7 |
| 11 | 11.2.14 | 11.2.19 | 0.1102.19 | ~4.1 (~4.0\|\|~4.1) | 0.11.4 | 6.6.7 |
| 12 | 12.2.17 | 12.2.18 | 12.2.18 | ~4.3 (~4.2.3\|\|~4.3.2) | 0.11.4 | 6.6.7 |
| 13 | 13.4.0 | 13.3.11 | 13.3.11 | ~4.6 (>=4.4.3 <4.7) | 0.11.5 | **7.8** |
| 14 | 14.3.0 | 14.2.13 | 14.2.13 | ~4.8 (>=4.6.2 <4.9) | 0.11.5 | 7.8 |
| 15 | 15.2.10 | 15.2.11 | 15.2.11 | ~4.9 (>=4.8.2 <5.0) | 0.12.0 | 7.8 |
| 16 | 16.2.12 | 16.2.16 | 16.2.16 | ~5.1 (>=4.9.3 <5.2) | 0.13.0 | 7.8 |
| 17 | 17.3.12 | 17.3.17 | 17.3.17 | ~5.4 (>=5.2 <5.5) | 0.14.0 | 7.8 |

## Master plan 22-lib matrix — where each lib MUST move (Angular-compat floor)

| Lib | Now | Angular floor (lowest major whose line it supports) | Target @17 | Bump-at |
|---|---|---|---|---|
| @ng-bootstrap/ng-bootstrap | 5.3.1 (^8) | 6.x(^9), 7.x(^11), 8.x(^12), 9.x(^13), 10(^14), 11-12(^15), 13-14(^16), 15(^17) | 15 | every major (peer tracks major) |
| @ng-select/ng-select | 3.7.3 (<9) | 4.x(<10)... 8.x(^13), 10(^14), 11(^15), 12(^16), 13(^17) | 12 | ~every major |
| @kolkov/angular-editor | 1.0.0 (<9) | 1.x(<13), 2.x(<14), 4.x(^14)... | 4 | 13, 14 |
| @ngx-translate/core+http-loader | 11/4 | 12-13(^9-11), 14(^12), 15(^13)... | 15 / 8 | stepwise |
| @fortawesome/angular-fontawesome | 0.4.0 (^8) | 0.6(^10),0.8(^11),0.9-0.11(^12-14),0.12-0.14(^15-17) | 0.14 | stepwise |
| angular-gridster2 | 8.4.2 (^8) | 9-11(^9-11),12-13(^12),14(^13)... | 14 | stepwise |
| ng-inline-svg | 8.6.0 (>=6) | 9-11(>=6), 12-15(^12)... | 14 | 12 |
| ng-jhipster | 0.11.5 (^8.2.2) | 0.12(^9),0.13-0.14(^10-11),0.15-0.16(^12+) | 0.16 | stepwise |
| ng2-tel-input | 1.0.16 (^6) | 2.x(^8)...2.5(^15) | 2.5 | 15 (per master plan) |
| ng5-slider | 1.2.6 (>=5) | admits all | 1.4 | 17 (API: [(value)]->[value]+(valueChange)) |
| ng6-toastr-notifications | 1.0.4 (^6) | no 8+ line | accept-risk pin | — |
| ngx-cookie-service | 2.4.0 (>=4.2) | 3.x(^9)...17(^17) | 17 | stepwise |
| ngx-drag-drop | 2.x | admits all | 14 | 17 |
| ngx-draggable-resizer | 1.0.4 (~7.1) | no 8+ line | accept-risk pin | — |
| ngx-echarts | 4.2.2 (>=6) | echarts 4→5 at 17 | 17 | 17 (echarts 4→5) |
| echarts-gl/leaflet/stat/wordcloud | 1.1.x | match echarts-5 | matching | 17 |
| echartslayer | 0.1.6 (old) | verify | verify/grep-usage | 17 |
| ngx-flag-icon-css | 1.0.1 (>=2) | admits all | 1 | 17 |
| ngx-gauge | 1.0.0-beta.12 (^8) | 9-14... | verify (hand-roll on echarts if dead) | 17 |
| ngx-infinite-scroll | 7.2.0 (>=7) | 8-11(>=7),12-15... | 14 | stepwise |
| ngx-json-viewer | 2.4.0 (>=5) | 3.x | 3 | 17 |
| ngx-monaco-editor + monaco-editor | 7.0.0 / 0.20 | 8.x(^11), 11-18...; monaco 0.20→0.4x | 18 / 0.45+ | **12-17 (biggest: worker-loader→ESM workers)** |
| ngx-sortable-2 | 1.0.1 (^4) | no 8+ line | accept-risk pin / verify | — |
| ngx-spinner | 8.1.0 (^8) | 9(^9),10-12(^12),13-16(^16)... | 16 | stepwise |
| ngx-webstorage | 2.0.1 (^5) | 3-5...14 | 14 | stepwise |
| popper.js | 1.16.1 | bootstrap 4.6.2 uses v1 internally; verify no direct import | drop if unused | verify |
| interactjs | 1.4-rc.13 | admits all | 1.10 | 17 (RC→stable) |
| html2canvas | 1.0.0-rc.7 | admits all | 1.4.x | 17 (re-run PDF smoke) |

**ESLint** replaces TSLint at the Angular 12 stage (angular-eslint; the
codelyzer rules map to @angular-eslint recommended). TSLint is removed then —
codelyzer 5.x (dev) is the last TSLint line.

## Stages (one commit each, build green + smoke before commit)

- **S9: 8→9** — framework 9.1.13, ba 0.901.15, TS 3.8.3, zone 0.10.3.
  **Measured breaks (2026-11-03):**
  0. **`@angular/localize` is MANDATORY from Angular 9 onward** (canonical upgrade
     step the original matrix missed): core@9's fesm5 references the `$localize`
     global (9 sites) which type-checks (ambient) but is **undefined at runtime**
     unless polyfills import the package → app crashes at bootstrap with
     "$localize is not defined" and app-root never renders (compiles GREEN).
     Fix: dep `~9.1.13` + `import '@angular/localize/init';` in polyfills.ts (after
     zone.js). **The `/init` subpath is essential** — bare `import '@angular/localize'`
     loads the tree-shakeable library entry (not in `sideEffects`) and webpack
     drops it (main.js hash unchanged, crash persists); `init.js` is the
     side-effect entry that runs `_global.$localize = $localize`.
     **Every later stage (S10–S17) must bump @angular/localize to match
     the framework major.**
  1. `Renderer` (v1) removed from `@angular/core` — 4 files (register,
     password-reset finish/init, password-strength-bar). Fix: `Renderer2`
     (`setElementStyle`→`setStyle`, `setElementClass(el,c,false)`→`removeClass`);
     the 3 `renderer.invokeElementMethod(el,'focus',[])` become `el.focus()`
     (Renderer2 has no invokeElementMethod) and drop the now-unused ctor param.
  2. **`ngx-gauge 1.0.0-beta.12` hard-blocks Stage 9** (its gauge.d.ts imports the
     removed v1 `Renderer`). Bumped to **`1.0.0-beta.13`** (peer `^9.0.0`, uses
     Renderer2, identical input API — verified from the published .d.ts). This is the
     "verify ngx-gauge" item from the master matrix forcing itself early; the 1.0.0/1.1.0
     *stable* releases are a rewritten API (thresholds/duration, no max/min/value/size)
     and were NOT used.
  3. kolkov stays 1.0.0 (peer `>=8 <9` admits 9) — no bump.
  Libs whose peer caps at 8 (ng-bootstrap 5.3.1, ng-select 3.7.3, spinner 8.1.0,
  ng-jhipster 0.11.5, fontawesome 0.4.0, gridster2 8.4.2, gauge beta.13) installed with
  npm-6 peer *warnings*; they still compile against 9.1.13 (no AOT break). Bump them at
  the stage their own API first breaks or where the master matrix stages them.
- **S10: 9→10** — 10.2.5, ba 0.1002.4, TS ~3.9.9, zone ~0.11.4, @angular/localize ~10.2.5.
  **Measured breaks (2026-11-03):**
  1. `angular.json`: `es5BrowserSupport` option removed from the schema → delete the key
     (was `"es5BrowserSupport": true` in build options). "Schema validation failed ...
     should NOT have additional properties(es5BrowserSupport)".
  2. **SCSS `url()` resolution switched to file-relative** (build-angular 0.1002). 35
     pre-existing `url(src/assets/...)` refs (34 in `src/assets/styles/icons.scss` mask
     icons, 1 in welcome-to-utmstack `.webp`) now resolve to a doubled path and fail.
     Fix: make them relative to the importing file — icons.scss (in `src/assets/styles/`)
     → `url(../icons/utm/X.svg)`; welcome (in
     `src/app/shared/components/getting-started/welcome-to-utmstack/`) →
     `url(../../../../../assets/img/welcome_integration.webp)`. (icomoon `url(fonts/...)`
     already file-relative — untouched.)
- **S11: 10→11** — 11.2.14, ba 0.1102.19, TS 4.1, zone 0.11.4. (monaco-editor 8.x line.)
- **S11: 10→11** — 11.2.14, ba 0.1102.19, TS 4.1, zone 0.11.4, @angular/localize ~11.2.14.
  **Measured breaks (2026-11-03):**
  1. TS 4.1 strictness: `date.pipe.ts` — Angular 11's `DatePipe.transform` gained a
     `(value: null|undefined) => null` overload; our `: string` override is no longer
     assignable (TS2416). Fix: mirror the base overloads + return `string | null`.
  2. TS 4.1 strictness: `incident-related-alert.component.ts` `@Input() incidentId` used
     in a field initializer before init (TS2729). Fix: `incidentId!: number`
     (definite-assignment; the input is always bound by the parent).
- **S12: 11→12** — 12.2.17, ba 12.2.18, TS 4.3.5. **webpack 5 lands.**
  **Measured breaks (2026-11-04):**
  1. **webpack 5.50 still hashes with md4** (xxhash64 default is 5.54+) → Node 18
     (OpenSSL 3) still fails without `--openssl-legacy-provider`; **S12 is
     validated on Node 14 (OpenSSL 1.1, no flag needed — CI host)**. Node 18
     no-flag builds only from S13 (build-angular 13.3.11 ships webpack 5.76+).
  2. **tslib ^1.14.1 → ^2.3.0** (installs 2.9.1): TS 4.3 emits `__spreadArray`
     (renamed from `__spreadArrays`); with `importHelpers: true` the v1 tslib
     lacks it → dozens of webpack module errors.
  3. **`SectionConfigType.id?: 0` → `id?: number`** (TS2367 ×5): the literal
     `0` made every `section.id !== ApplicationConfigSectionEnum.X` "always
     true" — a latent type bug TS 4.3 surfaces. All usages are numeric.
  4. **critters (NEW in build-angular 12, index.html critical-CSS inliner)
     strict-parses EVERY `<link rel=stylesheet>` in index.html** — including
     `assets/icons/icomoon/styles.css` (a plain `.css` loaded from index.html,
     never run through sass). Its `// padding: 5px;` SCSS-style comments
     (L42-44, a fully commented-out `.icon-bordered` block) fail postcss:
     "Unknown word padding" at `<css input>:42:5`. Fixed: `//` → `/* */`.
     (The 11 angular.json style entries were scanned — intl-tel-input's 3 `//`
     are inside a `/* */` block, safe.)
  5. **lazy routes: string `loadChildren: './x/x.module#XModule'` →
     `() => import('./x/x.module').then(m => m.XModule)`** (30 routes, 5 files).
     webpack 5 no longer code-splits the string form: the bundle had 0
     dynamic `import(` and 7 chunks total, so every lazy nav threw
     "Cannot find module './dashboard/dashboard.module'" at runtime (silent —
     0 pageerrors, empty router-outlet, bodyLen=0). The arrow form is what
     webpack 5 can statically analyze (bundle went 7 → 37 chunks).
     Companion: `tsconfig.json module: es2015 → es2020` — TS1323 forbids
     dynamic `import()` under `es2015` modules. (target stays es5;
     es5BrowserSupport was already removed at S10.)
     NOTE: `ng build --prod` still works but is deprecated in 12 — S13 must
     switch to `--configuration production` (--prod is removed in 13).
  6. **ESLint migration** (drop TSLint/codelyzer): `@angular-eslint` 12.7.0
     + eslint 7.32; `.eslintrc.json` maps the old tslint rules; angular.json
     lint builder `:tslint` → `@angular-eslint/builder:lint`; tslint.json
     deleted. `@typescript-eslint/no-non-null-assertion: warn` (21
     pre-existing `!:` in src).
  **No lib bumps forced at S12** (ng-bootstrap 5.3.1, ng-select 3.7.3 etc.
  still compile against 12.2.17; npm-6 peer warnings only) — the master plan's
  "ng-bootstrap 8, ng-select 8 at S12" was NOT needed to build; they ride
  until their own line is required at a later stage.
- **S13: 12→13** — 13.4.0, ba 13.3.11, TS ~4.6.4, **rxjs 7.8.2**,
  @angular-eslint 13.5.0, custom-webpack 13.1.0, webpack 5.76.1.
  **Measured breaks (2026-11-04):**
  1. **`--prod` REMOVED in CLI 13** → build script + all invocations now use
     `--configuration production`.
  2. **ES5 differential loading removed** → tsconfig `target: es5 → es2017`,
     `lib: es2018 → es2020` (the polyfills-es5 chunk disappears).
  3. rxjs 6→7 — measured breaks (2026-11-04):
     a. `next()`/`next(value)` now REQUIRES an argument (rxjs 7 tightened the
        Subscriber signature). Of 67 arg-less `.next()` in src, 65 are on
        `Subject<void>` (still valid); 2 are on `Subscriber<any>`/
        `Subject<any>` → `observer.next()`/`this.destroy$.next()` in
        auth-jwt.service.ts:106 + goal-view.component.ts:200. Fixed to
        `.next(null)` (both emit nothing downstream — logout/destroy signals).
     b. `forkJoin(...)` over an array of observables now infers `unknown[]`
        (was `any[]` in rxjs 6) → `.filter` on the emitted value is TS2339.
        2 sites: add-rule.component.ts:275 + import-rule.component.ts:159.
        Fixed by annotating the subscribe/map param `(filteredDataTypes: any)`.
     c. 1 `.toPromise()` (account.service) — deprecated but works in rxjs 7,
        left as-is (0 removed-statics Observable.of/fromPromise/empty in src).
  4. webpack 5.76.1 ≥5.54 → **Node 18 needs no `--openssl-legacy-provider`**
     (verified by the S13 build running flag-free).
  5. angular.json: `extractCss` REMOVED from the build schema (like
     es5BrowserSupport was at S10) → schema validation fails. `aot`,
     `namedChunks`, `vendorChunk` still valid in 13 (verified against the
     13.3.11 browser schema.json: 41 properties).
  6. `--prod` flag removed from the CLI at 13 → `package.json build` script
     now `ng build --configuration production` (the production config already
     existed; --prod was only deprecated-accepted at 12).
  7. tsconfig: `target es5 → es2017`, `lib es2018 → es2020` (Angular 13
     dropped ES5 differential loading — es5 target no longer builds).
  8. **PITFALL (hit S13, affects S14–S17):** Angular 12+ keeps a persistent
     webpack filesystem cache at `frontend/.angular/cache`, keyed on
     angular.json / package.json / tsconfig — NOT on webpack.custom.js. Editing
     the custom webpack config (aliases!) with the cache present is a SILENT
     no-op: the build serves the stale compilation and the bundle hash stays
     byte-identical. After any webpack.custom.js change: `rm -rf frontend/.angular`
     (and `node_modules/.cache`) before rebuilding. Symptom check: main hash
     unchanged across two builds that differ only in webpack.custom.js.
  9. **intl-tel-input utils bootstrap crash (webpack 5.76):** ng2-tel-input
     (root module) side-effect-imports `intl-tel-input/build/js/utils` — a
     BARE IIFE `(function(){var aa=this; m("intlTelInputUtils",…)})();` whose
     `this` comes from the bare call inside the file. webpack 5.50 (S12)
     evaluated the module sloppy (`this === module.exports` object → attaches
     harmlessly, global never set, plugin degrades gracefully — app booted);
     webpack 5.76 (S13) evaluates it strict (`this === undefined`) →
     `"intlTelInputUtils" in undefined` throws at bootstrap (root module =
     app dies). PROVEN DEAD END: aliasing the import to a shim that
     `require()`s the real file does NOT help — the IIFE's own `this` is
     independent of the requiring module. Fix (no fork): (a) alias the
     import to `intl-tel-input.utils.js` = `export {}` no-op stub (stops
     webpack from evaluating the IIFE); (b) angular.json asset entry copies
     `node_modules/intl-tel-input/build/js/utils.js` → `dist/lib/
     intl-tel-input/utils.js`; (c) raw `<script src="lib/intl-tel-input/
     utils.js">` in src/index.html — a classic script is sloppy
     (`this === window`), so the IIFE sets `window.intlTelInputUtils`
     (the library's README delivery mode; restores FULL tel-input
     functionality, better than S12's degraded state). ng2-tel-input itself
     never touches intlTelInputUtils (only the plugin reads it, guarded).
- **S14: 13→14** — 14.3.0, ba 14.2.13, cli 14.2.13, TS ~4.8.2 (4.8.4),
  @angular-eslint 14.4.0, custom-webpack 14.1.0. zone.js stays 0.11.8
  (14 peers `~0.11.4 || ~0.12.0`; 0.12 comes at S15).
  **Measured breaks (2026-11-04), all TS 4.8:**
  1. **TS2839 (new rule — object reference compares):**
     a. `utm-thread-metrics.component.ts:38` `if (this.threadData !== {})`
        (always true; `threadData` is an `any` array `.forEach`-ed) →
        `if (this.threadData && this.threadData.length > 0)`.
     b. `sddl-parser.util.ts:163` `if (this.aceTypes === {})` (always false;
        `aceTypes` is a string map, ctor already calls `initialize()`) →
        `if (Object.keys(this.aceTypes).length === 0)` (defensive re-init guard,
        same intent, now actually checks emptiness).
  2. **TS2345 Routes inference:** `rule-management.routing.module.ts` declared
     `const routes = [...]` with NO `: Routes` annotation; TS 4.8 infers a
     per-object union that Angular 14's `Routes` rejects. Fix: annotate
     `const routes: Routes = [...]` + import `Routes` (the app's own
     convention in app-routing.module.ts).
  3. angular.json: only `extractCss` removed from the 14 schema (already gone
     at S13); `aot`/`vendorChunk`/`namedChunks` still valid (verified against
     the 14.2.13 schema.json: 40 properties).
  4. **npm-6 hoisting conflict (build crashes at SETUP, RC=127, before any
     compile):** `Cannot find module 'ajv/dist/compile/codegen'` via
     `ajv-keywords@5 → schema-utils → webpack-dev-server`. build-angular 15.2.11
     pulls schema-utils 4.8.2 + ajv-keywords 5.1.0 (peer `ajv ^8.8.2`), but npm 6
     hoisted **ajv 6.12.6** to top-level (eslint 7 depends on ajv ^6), so
     ajv-keywords resolved the top-level ajv6 which has no `dist/compile/codegen`.
     Fix: direct devDependency `"ajv": "^8.12.0"` → ajv8 hoists to top (satisfies
     ajv-keywords/schema-utils), eslint 7 gets a nested ajv6. ajv is build-time
     only (never bundled). Reinstall + verify top ajv=8.x before building.
- **S15: 14→15** — 15.2.10, ba 15.2.11, TS ~4.9.5, zone 0.12.0.
  **Measured breaks (2026-11-04):**
  1. **npm-6 hoisting conflict (build crashes at SETUP, RC=127):** build-angular
     15.2.11 pulls schema-utils 4.8.2 + ajv-keywords 5.1.0 (peer `ajv ^8`),
     but npm 6 hoisted ajv 6.12.6 (eslint 7's dep) to top-level →
     `ajv-keywords` resolved top ajv6 → `Cannot find module
     'ajv/dist/compile/codegen'`. Fix: direct devDep `"ajv": "^8.12.0"` →
     ajv8 hoists (top = 8.20.0), eslint 7 gets nested ajv 6.15.0. Build-time
     only, never bundled.
  2. **sass `~` prefix is dead (build-angular 15 switched sass-loader to
     `api: 'modern'`, `webpackImporter: false` + custom
     `getSassResolutionImporter`).** The legacy webpack importer that stripped
     `~` is gone; the modern importer resolves bare requests via webpack's
     enhanced-resolve (node_modules + project root via `preferRelative`).
     Fix: strip `~` from all 11 `@import` targets (4 node_modules: bootstrap,
     @ng-select, font-awesome, highlight.js; 7 project: src/assets/styles/{
     theme,var,custom-elements}) across 6 scss files. No includePaths needed.
  3. **CI Node floor raised:** Angular CLI 15 requires
     "minimum Node.js version of either v14.20, v16.13 or v18.10" — the
     pinned CI node 14.16.1 (reusable-node.yml) now REJECTS the build at
     startup (compilation itself succeeds; the CLI version gate fires first).
     Fix: bump reusable-node.yml `node-version` 14.16.1 → **14.20.0** (the only
     Node that both clears the CLI floor AND keeps npm 6 / lockfileVersion 1;
     16/18 = npm 8/9 = v3 lock break).
  4. ng2-tel-input stays 1.0.16 (no template usage found — only module import;
     its `intl-tel-input` side-effect import is stubbed by the S13 alias +
     classic-script delivery).
- **S16: 15→16** — 16.2.12, ba 16.2.16, cli 16.2.16, **TS ~4.9.5** (peer
  `>=4.9.3 <5.2`; deliberately NOT 5.1 — 4.9.5 is the highest TS the ngcc@15
  shim supports, see break 5), zone.js 0.13.3, webpack 5.94.0,
  @angular-eslint 16.3.1, custom-webpack 16.0.1.
  **Measured breaks (2026-11-04):**
  1. **`entryComponents` REMOVED from the NgModule type** (Angular 16) →
     TS2345 "not assignable to NgModule" on ~every module that declared it.
     39 `entryComponents: [...]` statements removed across 39 files
     (codemod, balanced-bracket, unit-tested 4/4; diff = 210 deletions, all
     entryComponents array lines). Safe: Ivy (default since v9) auto-registers
     declared components; entryComponents only mattered for View Engine /
     dynamic `createComponent`, which this app does not rely on for
     declared-in-same-module components (the 2 dynamic uses — log-analyzer
     tab content, dtable-detail-resolver — are `@ViewChild`-anchored
     directives, not entryComponents).
  2. **ng6-toastr-notifications is UNBUILDABLE at 16**: it imports
     `ReflectiveInjector` (v1 symbol removed from @angular/core in 16) — a
     hard named-import error in its FESM. No version admits 16 (peer ^6.0.0,
     last release 1.0.4). **Migrated to ngx-toastr 17.0.2** (peer
     `>=16.0.0`), per the plan's own sanction ("replace on ngx-toastr if
     incompatible"). Scope was contained: the ONLY consumer is
     `UtmToastService` (app.module.ts + utm-toast.service.ts; no template/CSS
     references toastr). Public `UtmToastService` method names kept identical
     so the 94 call sites (showError x63, showSuccessBottom x17, showWarning
     x7, showInfo x3, showSuccess x2, showSuccessProcess x2) are untouched.
     Mapping: `successToastr(msg,title,opts)`→`success(...)`,
     `customToastr(msg,null,{enableHTML,position})`→`show(...,{enableHtml,
     positionClass})`, `position:'bottom-right'`→`positionClass:
     'toast-bottom-right'`, `toastTimeout:N`→`timeOut:N`, default (no
     position) = top-right = both libraries' default. `ToastrModule.forRoot()`
     kept (ngx-toastr has forRoot(config?) — called with no config). ng6-toastr
     removed from package.json.
   3. **CI Node floor: Angular CLI 16 DROPS Node 14 entirely** (message: "The
      Angular CLI requires a minimum of v16.13" — verified: Node 14.20.0 build
      RC=3, compilation itself passes, gate fires first). **The v1-lock rule
      ends at S15.** S16+ host = **Node 16.20.2** (last Node 16, npm 8.19.4)
      → lockfileVersion 2 is canonical (npm 8 reads/writes v2; `npm ci` on CI
      stays reproducible). Both workflows bumped 14.20.0 → 16.20.2
      (reusable-node.yml, frontend-federation-build.yml) + AGENTS.md sync.
      Verified: npm install on 16.20.2 (lockfileVersion 2), prod build green,
      main hash IDENTICAL to the Node 18 build (702f53ddbe609fee).
    4. **npm 8 ERESOLVE (the Node 16 host change surfaces it):** npm 8
       enforces peer deps that npm 6 only warned about. `npm install` on
       16.20.2 hard-fails on `@fortawesome/angular-fontawesome@0.4.0` (peers
       `@angular/common ^8`, an Angular-8-era lib that rides until the S17
       sweep). Fix: `frontend/.npmrc` with `legacy-peer-deps=true` (repo-scoped,
       deterministic for CI + local) — restores npm 6's resolution behavior.
       Re-evaluate the flag after the S17 lib sweep retires the old-line libs.
    5. **Angular 16 removed in-build ngcc from the webpack builder** (the
       JavaScriptTransformer now only exists under `src/tools/esbuild/`; the
       CLI 16 `ngcc` bin is a no-op deprecation stub — verified by source
       read: it prints the alert and `process.exit(0)`, no `main-ngcc.js`).
       A clean install therefore leaves all ~20 View Engine deps raw: prod
       build fails `TS2314 ModuleWithProviders<T>` on their `.d.ts` +
       `NG6002`; a tree where only the JS is raw GREEN-BUILDS but ships raw
       VE `.decorators=[` into prod -> silent bootstrap crash "JIT compiler
       unavailable" (0 pageerrors, empty app-root). S15 only built because
       its libs were still CLI-15-era in-place-compiled and ngcc ran
       in-memory per-build.
       **Fix (committed shim `ngcc-shim.cjs`, REMOVE AT S17):** TS pinned
       ~4.9.5 (highest TS ngcc@15 supports — it monkey-patches `ts`, breaks
       on 5.1+ with a `getExpandoInitializer` getter-only conflict); devDep
       alias `@angular/compiler-cli-ngcc: npm:@angular/compiler-cli@15.2.10`;
       `postinstall: node ngcc-shim.cjs` which (a) runs that ngcc, then
       (b) repairs ngcc's d.ts output to A16-acceptable shape. No forks.
       Two distinct A16-vs-ngcc@15 gates were hit, both proven from source:
       (1) **UMD barrel gap** — ngcc@15 compiles the `.js` of UMD-layout
           packages whose `typings` is a barrel (only `ngx-sortable-2`) but
           leaves the per-class `.d.ts` without ɵ statics -> ngtsc rejects
           the NgModule (NG6002) -> ~110-error cascade. The shim injects the
           canonical in-class `static ɵfac/ɵmod/ɵcmp` (A16's
           `ɵɵComponentDeclaration` needs 7+ type args; 1-arg = TS2707).
       (2) **InputMap encoding gap** — ngcc@15 writes
           `"prop": { "alias": undefined; "required": undefined; }` for
           every input without an explicit `@Input('name')`. A16's
           `readInputsType` (compiler-cli) resolves the binding name via
           `readStringType(alias)`, and `undefined` reads as `null` -> the
           element still matches by selector (0 NG8001) but NO input
           binding resolves (178 NG8002 across ~40 components; skipLibCheck
           silences the type error, so 0 TS). The shim rewrites
           `"alias": undefined` -> the member key and `"required": undefined`
           -> `false` across every d.ts carrying `ɵɵ*Declaration`. Both
           passes are idempotent no-ops once the tree is correct.
- **S16: 15→16 — DONE (commit a994f1a437, 51 files).** Verified: clean
  CI-replica build green on Node 16.20.2 (full wipe -> `npm install` fires
  the postinstall shim -> `BUILD_RC=0`, 0 TS/NG errors, 0 raw-VE `.decorators=[`
  in the bundle, `main.343696e8e26a497a.js`); lab smoke all 9 routes
  `ng=16.2.12` 0 pageerrors; ngx-toastr migration proven at runtime (real
  `.ngx-toastr` toast "Query saved successfully" via log-analyzer Save-query,
  ROLE_USER-reachable + reversible); intl-tel-input classic-script check
  green (`window.intlTelInputUtils` = object + `.intl-tel-input` widget
  initialized in AD notification phone box). Lab deploy: slim image
  `utmstack-frontend-local:v11.3.1-ang16` (FROM the v11.3.1fix3 nginx
  image, fresh dist). PITFALL HIT: building under MSYS2 bash rewrote
  `--base-href /` to `<base href="C:/Program Files/MSYS2/">` (path
  conversion) — every deep-route asset/lazy-chunk 404'd to the SPA
  fallback (MIME text/html) while JS hashes stayed correct; rebuild under
  PowerShell fixed it. `src/index.html` double-`<head>` is pre-existing
  (S13 era, benign).
- **S17: 16->17 - DONE (commit pending this session).** Framework 17.3.12 /
  cli+ba 17.3.17 / TS ~5.4.5 / zone ~0.14.10 / @angular-eslint 17.5.3 /
  custom-webpack 17.0.2. **ngcc shim REMOVED** (A17 has no ngcc); final lib
  sweep to the measured Ivy matrix (all tarball-verified Ivy; ng-bootstrap
  11.0.1 = last Bootstrap-4 line, honors the no-Bootstrap-5 decision).
  **Measured S17 breaks (all fixed, all root-caused to library changes, none to A17 itself):**
  1. **zone.js 0.14** dropped the `./dist/zone` export-map path ->
     `import "zone.js"` in polyfills.ts.
  2. **`@ngx-translate/http-loader@17`** - the REAL boot-hang. `TranslateHttpLoader`
     has a no-arg ctor that calls `inject(TRANSLATE_HTTP_LOADER_CONFIG)`;
     `forRoot` defaultProviders never registers that token, so
     `loader: {provide: TranslateLoader, useClass: TranslateHttpLoader}` threw
     `NullInjectorError: No provider for InjectionToken TRANSLATE_HTTP_LOADER_CONFIG!`
     on the FIRST HttpClient call (interceptor chain -> UtmToastService ->
     TranslateService). `detect()` swallowed it via catchError, isOnlineApi$ never
     hit /api/ping 200, and initTimezoneFormat first(val=>val===true) hung the
     whole bootstrap SILENTLY (0 pageerrors, ng-version never set, "Preparing
     your workspace" spinner forever; /api works fine from curl/python).
     Fix: `{provide: TRANSLATE_HTTP_LOADER_CONFIG, useValue: {}}` in AppModule
     providers; v17 defaults (/assets/i18n/ + .json) match the app i18n path.
     **Lesson: a swallowed catchError in an APP_INITIALIZER dependency turns a
     DI error into a silent infinite spinner - instrument the catch before
     theorizing about interceptors/webstorage (first suspected bare
     LocalStorageService providers, which was wrong).**
  3. **angular-resizable-element 3->7**: `resizeEdges` moved off the
     `[mwlResizable]` container onto the new `[mwlResizeHandle]` child directive
     (confirmed via ɵdir metadata in 5.0.0/6.0.2/7.0.2 - none keep the container
     API; v5 is DUAL/partial-Ivy so 7.0.2 is the only clean A17 line). All 11
     split-pane templates migrated: `[resizeEdges]` moved to a child
     `<div mwlResizeHandle [resizeEdges]="{ bottom: false, right: true, top:
     false, left: false }"></div>` as first child + global CSS in styles.scss
     (`[mwlResizeHandle]` = absolute right-edge strip, ew-resize, z-index 10).
     ResizeEvent.rectangle unchanged (all 11 TS onResize handlers untouched).
     Only 3 of 11 surfaced as NG8002 - the other 8 are in NO_ERRORS_SCHEMA
     modules and would have broken SILENTLY at runtime; migrated all 11.
  4. **echarts 4->5** (forced: every Ivy ngx-echarts line requires echarts>=5; first
     Ivy ngx-echarts is 8.0.0). (a) `EChartOption`->`EChartsOption` in
     chart-view.component.ts. (b) echarts-leaflet 1.1.0 (only release,
     echarts-3/4-era UMD) calls `(m.default ?? m).extendComponentModel` on
     `require("echarts/lib/echarts")`; in echarts 5 that path is a compat STUB
     whose default is only `{init}` -> `m.extendComponentModel is not a
     function` at IMPORT TIME -> whole app crashed at bootstrap. Fix:
     webpack alias `echarts/lib/echarts` -> `node_modules/echarts/index.js`
     (webpack.custom.js, the sanctioned alias mechanism, same pattern as the
     S12 leaflet + S13 intl-tel-input aliases): the UMD factory then gets the
     full echarts namespace (no `export default`, so `m.default` is undefined ->
     falls through to `m` = full core with extendComponentModel/register
     CoordinateSystem/util/graphic/matrix). The alias dedupes to the SAME core
     that NgxEchartsModule.forRoot({echarts: () => import("echarts")}) and the
     other main.ts extension imports use. Verified: "not supported anymore" stub
     string absent from all dist chunks.
  5. **ngx-monaco-editor 7->12**: default AMD baseUrl changed
     `config.baseUrl || "/assets"` + `/monaco/vs` to
     `(config.baseUrl || "./assets") + "/monaco-editor/min/vs"`; angular.json
     asset glob output changed to `/assets/monaco-editor/min/vs` to match (a
     loader.js at the wrong path = silent editor failure). monaco-editor 0.20
     -> 0.24 (the last ngx-monaco-editor line peers ^0.24, so the plan's 0.45
     target is unattainable without a different wrapper - accepted). (c)
     monaco 0.24 typings = the ESM MODULE esm/vs/editor/editor.api.d.ts
     (no ambient global `monaco`) -> `import * as monaco from "monaco-editor"`
     in code-editor.component.ts. **RISK (verified OK in smoke):** that is a
     VALUE import, so webpack bundles the ESM monaco while the editor loads the
     AMD monaco from /assets - if the SQL completion provider registered on the
     ESM instance it would be dead. Smoke PROVED completion works (suggestions
     show "SELECT" + real custom fields origin.windowsScheduledTask etc), so
     the provider registered on the right instance. (d) 0.24 made
     CompletionItem.range REQUIRED - now computed via
     model.getWordUntilPosition(position) + position.lineNumber.
  6. **VE-only libs replaced (no Ivy version exists - A17 has no ngcc to compile
     them):** ngx-sortable-2 -> ngx-sortable 2.1.0 (drop-in: same `ngx-sortable
     selector, items/name inputs, @ContentChild(TemplateRef) itemTemplate -
     template untouched beyond import path); ng2-tel-input -> NEW local Ivy
     UtmTelInputDirective (same public API: selector [ng2TelInput],
     ng2TelInputOptions input, hasError/ng2TelOutput/countryChange/intlTelInput
     Object outputs; drives intl-tel-input v12 jQuery plugin - `import
     "intl-tel-input"` + `import * as $ from "jquery"` in the directive, both
     allowlisted in allowedCommonJsDependencies; the plugin patches the SAME
     jquery module the directive imports (webpack dedupes), and lands in the
     AD lazy chunk 459). ngx-flag-icon-css -> inline span with class
     flag-icon + ngClass "flag-icon-" + countryCode.toLowerCase() (flag-icon-css
     3.x was already a direct dep + in angular.json styles; .flag-icon-gb/
     .flag-icon-us confirmed in the CSS). ngx-draggable-resizer -> DROPPED (dead
     import in dashboard-builder.module.ts, zero template usage; its css removed
     from angular.json).
  7. **ngx-webstorage 2->13**: `Ng2Webstorage.forRoot()` ->
     `NgxWebstorageModule.forRoot()` (root removed; LocalStorageService/
     SessionStorageService names kept).
  8. **ngx-echarts 4->17**: `NGX_ECHARTS_CONFIG` is NOT @Optional in v17 ->
     `NgxEchartsModule.forRoot({echarts: () => import("echarts")})` mandatory
     once (AppModule); directive inputs options/theme unchanged; dropped
     `@types/echarts` (echarts 5 ships its own index.d.ts); `export =` in
     echarts/index.d.ts + wordcloud default import -> tsconfig
     allowSyntheticDefaultImports: true.
  9. **ng-bootstrap 5->11** (11.0.1 = last Bootstrap-4 line that is fully Ivy;
     12.x = Bootstrap 5; 10.x = VE-only DUAL): template codemod across 73
     files - `[collectionSize]`->`[pageSize]`, `startViewMode=`->`startView=`
     (datepicker), dropped `showNavigationBar=`/`selectionMode=` (removed in
     v8). tooltipClass/popoverClass KEPT in v11 (verified via ɵdir dumps: both
     still inputs; 207 + 50 usages untouched).
   **VERIFIED (lab smoke on FINAL tree, main.1abf6bd72256cff4, DEPLOY_OK):**
   all 9 routes ng=17.3.12 with 0 pageerrors (boot + dash-227 canvases = echarts
   5 charts render); monaco SQL completion WORKS (suggestions include "SELECT" +
   real custom fields origin.windowsScheduledTask etc) = no dual-instance bug;
   toast "Query saved successfully" (test query auto-deleted); AD phone box
   initializes (iti:true, utils:object); flag-icon spans render (flag-icon-us);
   mwlResizeHandle present; ngb-pagination shows multi-page lists (page 1..4 +
   prev/next on the alert list) = collectionSize intact. **REGRESSION CAUGHT +
   FIXED during verification:** the ng-bootstrap codemod blindly renamed
   [collectionSize]->[pageSize] (assumed the 16.x API), but 11.0.1 keeps BOTH as
   distinct inputs -> 73 duplicate-[pageSize] bindings, pagination silently
   broken (1 page only; hidden in NO_ERRORS_SCHEMA modules, build stayed green).
   Reverted exactly 73/73 via a diff-verified inverse; re-verified: 0 collectionSize
   / 0 stray pageSize deltas vs the S16 baseline.
   **Build:** PowerShell prod build BUILD_RC=0, base href="/", 0 raw-VE across
   37 chunks, monaco loader.js + editor.main.css at the v12 path, monaco CSS in
   the styles bundle. Clean CI-replica (full wipe -> npm ci Node 18.20.8 / npm
   10.8.2 -> build) was GREEN and byte-deterministic on the pre-pagination-fix
   tree (main.8c8cc2a9b76e31ee, INSTALL_RC=0, 0 errors, lockfileVersion stays 2).
   Re-run on the FINAL tree (after the pagination revert) was GREEN and
   byte-identical to the smoke-passing PowerShell build
   (main.1abf6bd72256cff4, INSTALL_RC=0, 0 errors, 0 raw-VE across 37 chunks). CI-replica dist re-verified: intl-tel-input plugin in
   lazy chunk 459, registerCoordinateSystem(leaflet) present in main, echarts-5
   compat stub absent.
  **CI host:** Node 16.20.2 -> 18.20.8 (CLI 17 floor v18.13) in both workflows;
  npm 10.8.2 preserves lockfileVersion 2 (verified). **legacy-peer-deps
  RETAINED** (re-evaluated): 4 libs still pin core peers <17 - ng-bootstrap
  ^13 (11.0.1), ng-inline-svg ^12, ng-jhipster ^10, ngx-monaco-editor ^12
  (all = highest available Ivy line; ng-bootstrap capped by the Bootstrap-4
  decision). Documented in .npmrc.
## Exit (master plan): Angular 17 build green on Node 18+ (webpack 5, NO legacy flag),
Karma green (compile gate = prod build; browser gate = lab smoke), visual baseline diff
reviewed route-by-route, all 22 libs at target or accept-risked (ng6-toastr,
ngx-draggable-resizer, ngx-sortable-2, echartslayer, ngx-gauge pending), commit per major.

## Risk notes
- **ngx-monaco-editor + monaco-editor** (Log Explorer + rule editor) is the single
  biggest effort — worker loading changes 0.20→0.4x. Budget the most QA here.
- **echarts 4→5** at S17 changes the chart API surface (bindGlobal removed); all 9
  map charts + scatter + every dashboard chart must be re-verified visually.
- **rxjs 6→7** at S13: mostly compatible, but `Observable.of`/`fromPromise` etc. are
  gone (measured: 0 refs in Phase 1 scan, so low risk).
- Each stage is independently buildable + smoke-able, so a break at stage N does not
  corrupt stages < N (already committed).
