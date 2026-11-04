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
- **S14: 13→14** — 14.3.0, ba 14.2.13, TS 4.8.
- **S15: 14→15** — 15.2.10, ba 15.2.11, TS 4.9, zone 0.12. ng2-tel-input→2.5.
- **S16: 15→16** — 16.2.12, ba 16.2.16, TS 5.1, zone 0.13.
- **S17: 16→17** — 17.3.12, ba 17.3.17, TS 5.4, zone 0.14. Final lib sweep to
  matrix targets (echarts 4→5 cluster, monaco 0.45, ng5-slider API, html2canvas 1.4,
  interactjs 1.10). **ngx-echarts 17 + echarts-gl/leaflet/stat/wordcloud matching 5.x** —
  the heaviest cluster; Log Explorer + all 9 map charts + dashboard scatter must be
  visual-QA'd.

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
