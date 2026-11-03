# Phase 1 — Angular 7 → 8 (detailed task plan)

> Expanded from `docs/plans/2026-11-02-frontend-dep-remediation.md` Part 2 Phase 1,
> per that plan's gate ("when a phase starts, it expands into its own detailed task
> plan measured against the actual installed majors"). All targets below were
> registry-verified 2026-11-03, not pre-estimated.

**Goal:** Move `frontend/` from Angular 7.2 to a coherent **Angular 8.2 baseline** on
the Phase-0 toolchain, with a clean build, a passing lab smoke, and a zero-material-change
visual baseline diff. This closes the residual `node-sass@4.13.1 → tar@2.2.2` (build-angular
0.803 dropped its node-sass dep).

**Branch:** `fix/ir-role-model-v11.3.1` (Phase 0 committed at `79db19e111`). NEVER stash
the main checkout (user WIP stash standing rule).

**Measured targets (npm registry, 2026-11-03):**

| Package | Now (lock) | Phase 1 target | Why |
|---|---|---|---|
| @angular/core (+common/compiler/compiler-cli/forms/platform-browser/platform-browser-dynamic/router) | 7.2.0 | **8.2.14** | the framework major |
| @angular/language-service (dev) | 7.2.x | **8.2.14** | matches core |
| @angular/cli (dev) | 7.3.10 | **8.3.29** | last 8.x CLI |
| @angular-devkit/build-angular (dev) | 0.13.10 | **0.803.29** | pairs with cli 8.3; **no node-sass dep** (ships sass@1.22.9) |
| rxjs | 6.3.3 | **6.6.7** | @angular/common 8.2.14 peers rxjs `^6.4.0`; 6.6.7 = last 6.x |
| zone.js | 0.8.26 | **0.9.1** | Angular 8 pairs zone.js 0.9.x |
| typescript (dev) | 3.2.4 | **~3.5.3** | build-angular 0.803 *peer* is `>=3.1 <3.6`, but the Angular 8 **compiler** hard-requires `>=3.4 <3.6` (build error: "The Angular Compiler requires TypeScript >=3.4.0 and <3.6.0 but 3.2.4 was found") — 3.5.3 satisfies both |
| @ng-bootstrap/ng-bootstrap | 4.2.2 (peer ^7) | **5.1.4** | peer ^7 caps below 8; 5.x = Angular 8 line (verified core-peer `^8.0.0`) |
| @kolkov/angular-editor | 0.18.8 (peer >=7 <8) | **1.0.0** | 0.19.x does not exist; 1.0.0 is first admitting 8 (core-peer `>=8.0.0 <9`); `editor` binding API stable |
| @ng-select/ng-select | 2.20.5 (peer >=6 <8) | **3.0.0** | core-peer `>=6.0.0 <9.0.0` (verified); Select API stable; visual QA the selects |
| codelyzer (dev) | 4.5.0 (peer <8) | **5.0.0** | core-peer admits 8 (verified `>7.0.0-beta <9.0.0`); lint-only |
| ng-jhipster | 0.9.3 (peer ^7) | **0.10.1** | core-peer `^8.0.0` (verified); all 5 imported symbols (JhiAlertService/JhiBase64Service/JhiEventManager/JhiParseLinks/JhiResolvePagingParams) confirmed present in the 0.10.1 tarball |
| ngx-spinner | 7.2.0 (peer ^6\|\|^7) | **8.0.2** | first admitting 8 (8.0.0/8.0.1 are `^7.2.0`); `<ngx-spinner>` API stable |
| @fortawesome/angular-fontawesome | 0.3.0 (peer ^6\|\|^7) | **0.4.0** | core-peer `^8.0.0-rc.5` (verified); `fa-icon` inputs stable |
| ngx-draggable-resizer | 1.0.4 (peer ~7.1.0) | **accept-risk pin 1.0.4** | NO published version admits Angular 8; low-surface drag handles |

Libs that already admit Angular 8 (low floors, no upper cap) install **unchanged** and
move only in Phase 2's forward matrix: `angular-gridster2 8.4.2` (peer ^8), `ngx-gauge`
(peer ^8), `ng-inline-svg 8.6.0`, `ngx-echarts 4.2.2`, `ngx-infinite-scroll 7.2.0`,
`ngx-json-viewer 2.4.0`, `ngx-monaco-editor 7.0.0`, `ngx-webstorage 2.0.1`,
`ngx-flag-icon-css 1.0.1`, `ng2-tel-input 1.0.16`, `ng5-slider 1.2.6`,
`ngx-cookie-service 2.4.0`.
(`ng6-toastr-notifications 1.0.4` is NOT in this list — its peer is `^6.0.0` (registry-
verified); it is the second accept-risk pin documented in Task 1.3 / phase1-notes.md.)

**DI note (measured, not the plan's assumption):** only **35** `@Injectable()` classes lack
`providedIn` (211 already use it). **Angular 8 does NOT require `providedIn`** — a
module-registered `@Injectable()` remains legal. Of the 35, 30 are already in a module
`providers[]`; 5 are not registered anywhere in a scanned `providers:`/`HTTP_INTERCEPTORS:`
block (`TypesService`, `InverHttpAlertManager`, `UtmAlertErrorService`,
`CanDeactivateConfigGuard`, `DialogService`). We do **not** force a `providedIn` refactor in
Phase 1 (YAGNI; 1486-file risk for no build gain). Task 1.4 verifies the 5 unregistered ones
are still provided (component providers / factory / dead) so nothing silently breaks.

**Build host:** Node 18.20 with `NODE_OPTIONS=--openssl-legacy-provider` (Phase 0
recorded — webpack 4 md4 on OpenSSL 3). Node 14 also builds (no flag) as the CI fallback.

---

## Task 1.1: Framework + toolchain + rxjs/zone bump  [DONE 2026-11-03]

**Files:** `frontend/package.json`, `frontend/package-lock.json`

**Executed result (evidence, 2026-11-03):**
- npm-6 install on Node 14 → lockfileVersion stays 1; installed @angular/core 8.2.14,
  build-angular 0.803.29, rxjs 6.6.7, zone.js 0.9.1, cli 8.3.29, typescript 3.5.3.
- **Residual node-sass closed:** `npm ls node-sass` → only `@angular-devkit/build-angular@0.803.29`
  no longer declares it; top-level `tar` is now 6.2.1 (no 2.2.2). The tar@2.2.2 chain that
  Part 1 left (build-angular 0.13 → node-sass@4.13.1 → node-gyp) is gone.
- First Node 18 build failed with the single, expected TS floor error, then the real v7→v8
  compile breaks (see Task 1.1b below).

**Fixes that the build forced (all v7→v8 API breaks, behavior-preserving):**

### Task 1.1b: TypeScript 3.2.4 → 3.5.3
- Angular 8's compiler requires TS `>=3.4 <3.6` (build error: "The Angular Compiler requires
  TypeScript >=3.4.0 and <3.6.0 but 3.2.4 was found"). Bumped dev `typescript` to `~3.5.3`.

### Task 1.1c: Child query decorators require `{static}` for SINGULAR only (40 sites)
- Angular 8.2.14 made the `opts` argument mandatory for the **singular**
  `@ViewChild`/`@ContentChild` with `static: boolean` required (verified in installed
  core.d.ts: the single-`(selector)` overload is gone). But the **plural**
  `@ViewChildren`/`@ContentChildren` still only accept `{read}` — `static` did not land
  until Angular 9 (9 of those sites were wrongly codemodded then reverted).
- **Migration (first pass):** added `{static: false}` to all 40 singular decorators.
  Codemod (paren-balanced, per-line, comment-aware), dry-run-verified on 10 shapes.
- **REGRESSION FOUND IN TASK 1.5 VISUAL DIFF (2026-11-03):** `static: false` defers
  resolution to after change detection, but Angular 7's default resolved *static*
  template queries (plain elements / `ng-template`, not behind `*ngIf`/`*ngFor`)
  BEFORE `ngOnInit`. 16 of the 40 queries are read in `ngOnInit` → they came back
  `null` → e.g. the Log Explorer tab content (`tab-content.component.ts`
  `@ViewChild(TabContentDirective)`, read in ngOnInit, renders the query view via
  `createComponent`) rendered blank. Fix: per-anchor classification of all 40 sites
  (tag-stack walk of each component HTML) and flip to `{static: true}` exactly the
  12 confirmed-static anchors read in ngOnInit (tab-content, dtable-detail-resolver,
  chart-builder, dashboard-grid, logstash editorContainer, notification
  dropNotification, search-input, tag-input, visualization-import content,
  action-conditional dropTracker, asset-group-add tagPopoverSpan,
  notification-view timeFilter). Dynamic anchors (`*ngIf`/`*ngFor` popover/buttons)
  stay `static: false` (static:true would throw at construction). Dead refs
  (`adChanges`, `iconBell`, `popoverQuery`, dynamic-table `#container` — zero
  occurrences in their templates) are null in v7 too and null-guarded; left
  `static: false`. Net effect: v7 resolution timing reproduced exactly per anchor.

### Task 1.1d: Two dead deep imports (TS2307)
- `rule-generic-filter.component.ts:3` `import {errorHandler} from '@angular/platform-browser/src/browser'`
  — `errorHandler` was never used; import deleted.
- `dynamic-table.component.ts:12` `import {container} from '@angular/core/src/render3'`
  — `container` was never used (the local `const container` shadows nothing); import deleted.

- [x] **Step 1: Edit `package.json` to the framework targets** — done (incl. typescript ~3.5.3)
- [x] **Step 2: Install with npm 6 (Node 14) to keep the v1 lock, then build on Node 18** — done
- [x] **Step 3: Fix compile errors** — done (TS floor + 49 child queries + 2 dead imports)
- [x] **Step 4: Confirm the residual node-sass/tar closed** — verified (node-sass ABSENT top-level, tar 2.2.2 gone)
- [ ] **Step 5: Commit (framework + fixes together)**

## Task 1.2: Bump the 7 capped libs to their Angular-8 lines  [DONE 2026-11-03, corrected]

**Executed result (evidence):** the naive caret ranges floated to two bad versions —
`@kolkov/angular-editor ^1.0.0` → **1.2.0** (ships ambient-accessor `.d.ts`, TS ≥4.3-only →
`TS1086` under our TS 3.5.3) and `ng-jhipster ^0.10.1` → **0.10.1** (peers ng-bootstrap
`^4.x`, breaks against ng-bootstrap 5 with "Function calls are not supported in decorators
but 'NgbModule' was called"). Corrected pins, both registry-verified:
- `@kolkov/angular-editor` **exact `1.0.0`** — the Angular-8 line (peer `>=8 <9`); 1.0.0's
  `.d.ts` has no ambient accessors (verified by scanning the published tarball), TS 3.5-safe.
- `ng-jhipster` **`^0.11.5`** — peer ng-bootstrap `^5.1.0` (matches our 5.3.1); tarball scan
  confirmed all 5 imported symbols still exported (JhiAlertService, JhiBase64Service,
  JhiEventManager, JhiParseLinks, JhiResolvePagingParams).
- `@ng-bootstrap ^5.1.4` → 5.3.1, `@ng-select ^3.0.0` → 3.7.3, `codelyzer ~5.0.0` → 5.0.1,
  `ngx-spinner ^8.0.2` → 8.1.0, `@fortawesome ^0.4.0` → 0.4.0 — all within Angular-8 lines,
  API-compatible (ng-bootstrap 5.3.1 still exports every `*Config`/`NgbPopover`/`NgbTooltip`
  class the code uses; the `*Options` rename is v6+).
- `ngx-draggable-resizer` stays 1.0.4 (accept-risk, no version admits 8).
Lock stays v1; install via npm 6 on Node 14.

For each lib, set the spec to the target, `npm install`, rebuild, and fix the named break:

- [x] `@ng-bootstrap/ng-bootstrap` → `^5.1.4`. **Known break:** v5 moved `NgbModal`
  config to an injectable and changed tooltip/popover to components. Grep
  `NgbModalOptions|NgbModal|NgbTooltip|NgbPopover` usage; fix the named files (modal
  open() still returns NgbModalRef; config injection is the change).
- [x] `@kolkov/angular-editor` → lowest version >= 0.18.8 whose `@angular/core` peer admits
  8 (verify via registry `https://registry.npmjs.org/@kolkov/angular-editor/<v>`).
  `editor` binding API is stable. If none admits 8, keep 0.18.8 (accept-risk pin) and note it.
- [x] `@ng-select/ng-select` → `^3.x` (lowest admitting 8). Select API stable; theming
  unchanged. Visual QA the selects in the smoke.
- [x] `codelyzer` (dev) → `^5.0.x`. Lint-only; if `ng lint` surfaces new rules, relax the
  named rule in `tslint.json` (lint is not a build gate).
- [x] `ng-jhipster` → `^0.10.x`. Grep `ng-jhipster` imports; verify each used util is still
  exported (the plan's Phase 2 matrix flags "verify each used util").
- [x] `ngx-spinner` → `^8.x` (lowest admitting 8). `<ngx-spinner>` API stable.
- [x] `@fortawesome/angular-fontawesome` → `^0.4.x` (lowest admitting 8). `fa-icon` inputs
  stable.

Rebuild after each lib (or in one batch if the build isolates errors cleanly). Fix only what
the build names.

- [ ] **Commit:** `chore(frontend): angular-8 peer-lib baseline (ng-bootstrap/ng-select/kolkov/codelyzer/ng-jhipster/spinner/fontawesome)`

## Task 1.3: ngx-draggable-resizer accept-risk pin

- [x] Keep `ngx-draggable-resizer@1.0.4` (no published version admits Angular 8 — peer
  `~7.1.0`). Record a one-line accept-risk in `docs/security/deps/phase1-notes.md`:
  "drag-handle component only; low runtime surface; npm installs with peer warning."
- [x] **Commit** (no code change; the note only).

## Task 1.4: Verify the 5 unregistered `@Injectable` still resolve

- [x] For `TypesService`, `InverHttpAlertManager`, `UtmAlertErrorService`,
  `CanDeactivateConfigGuard`, `DialogService`: grep their class name across `src/` to find how
  they're provided (component `providers`, a factory, `APP_INITIALIZER`, or unused). If any is
  genuinely unprovided AND injected, add the provider; if unused, leave it (dead code, note it).
- [ ] **Commit** if any change was needed.

## Task 1.5: Full build + lab smoke + visual baseline diff

- [x] **Step 1:** clean production build green on Node 18 (legacy flag) AND on Node 14 (CI
  fallback) — both must pass (CI uses Node 14; the branch will be merged to the v11/v12 lines).
- [x] **Step 2:** deploy the dist to the lab (slim image from the current lab frontend base +
  `COPY dist/utm-stack /usr/share/nginx/html/`), force-update, headless smoke the 13 baseline
  routes (login, log explorer, alert list, dashboard-main, dash 7/13/227/403/405, rule editor,
  integrations, soar, compliance) at desktop + mobile → 0 pageerrors.
- [x] **Step 3:** diff the 26 new screenshots against `docs/security/deps/visual-baseline/`
  (the Phase-0 pre-upgrade reference). Every material difference must be explained (the
  intended ones: ng-bootstrap 5 / ng-select 3 / spinner 8 component changes). A difference on a
  route whose libs we didn't touch = regression → fix before exit.
- [x] **Step 4:** Commit `docs/security/deps/visual-baseline-angular8/` (the new reference for
  Phase 2) + any notes.

## Task 1.6: Phase 1 exit

- [ ] Build green (Node 18 legacy + Node 14), lab smoke 0 pageerrors, visual diff explained,
  `npm ls node-sass` empty, residual tar 2.2.2 gone. Commit
  `chore(frontend): angular 8 baseline`. Update the master plan's Phase 1 checkboxes + exit.

**Exit (per master plan):** build green; visual baseline diff = zero *material* (unexplained)
change; residual node-sass/tar closed; commit `chore(frontend): angular 8 baseline`.

## Deviations from the master plan's Phase-1 text (evidence-based, 2026-11-03)

1. **`typescript` → `~3.5.3` (bumped, not "no bump").** build-angular 0.803.29's TS
   *peer* is `>=3.1 <3.6`, but the Angular 8 **compiler** hard-requires `>=3.4.0 <3.6.0`
   (build error: "The Angular Compiler requires TypeScript >=3.4.0 and <3.6.0 but 3.2.4
   was found"). 3.5.3 satisfies both.
2. **`@Injectable` → `providedIn` migration is NOT done.** Angular 8 does not require it and
   only 35 classes are affected (30 already module-registered). Forcing a 1486-file DI
   refactor for a single major bump is disproportionate risk. (Revisit at Phase 2 if a later
   major makes it load-bearing.)
3. **7 capped libs bump at this stop** (ng-bootstrap, kolkov, ng-select, codelyzer,
   ng-jhipster, spinner, fontawesome) + 1 accept-risk pin (draggable-resizer) — the master
   plan's 22-lib matrix is a Phase-2 forward story; these 8 *cap below Angular 8* so a
   coherent 8 baseline must move them now.
