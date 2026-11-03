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
| typescript (dev) | 3.2.2 | **3.2.2 (no change)** | build-angular 0.803 peer is `>=3.1 <3.6` — 3.2.2 already satisfies it |
| @ng-bootstrap/ng-bootstrap | 4.2.2 (peer ^7) | **5.1.4** | peer ^7 caps below 8; 5.x = Angular 8 line |
| @kolkov/angular-editor | 0.18.8 (peer >=7 <8) | lowest >= 0.18.8 admitting 8 (verify; else accept-risk pin) | peer caps below 8 |
| @ng-select/ng-select | 2.20.5 (peer >=6 <8) | **3.x** (lowest admitting 8) | peer caps below 8 |
| codelyzer (dev) | 4.5.0 (peer <8) | **5.0.x** | TSLint's Angular 8 ruleset |
| ng-jhipster | 0.9.3 (peer ^7) | **0.10.x** | peer ^7 caps below 8 |
| ngx-spinner | 7.2.0 (peer ^6\|\|^7) | **8.x** (lowest admitting 8) | peer caps below 8 |
| @fortawesome/angular-fontawesome | 0.3.0 (peer ^6\|\|^7) | **0.4.x** (lowest admitting 8) | peer caps below 8 |
| ngx-draggable-resizer | 1.0.4 (peer ~7.1.0) | **accept-risk pin 1.0.4** | NO published version admits Angular 8; low-surface drag handles |

Libs that already admit Angular 8 (low floors, no upper cap) install **unchanged** and
move only in Phase 2's forward matrix: `angular-gridster2 8.4.2` (peer ^8), `ngx-gauge`
(peer ^8), `ng-inline-svg 8.6.0`, `ngx-echarts 4.2.2`, `ngx-infinite-scroll 7.2.0`,
`ngx-json-viewer 2.4.0`, `ngx-monaco-editor 7.0.0`, `ngx-webstorage 2.0.1`,
`ngx-flag-icon-css 1.0.1`, `ng2-tel-input 1.0.16`, `ng5-slider 1.2.6`,
`ngx-cookie-service 2.4.0`, `ng6-toastr-notifications 1.0.4`.

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

## Task 1.1: Framework + toolchain + rxjs/zone bump

**Files:** `frontend/package.json`, `frontend/package-lock.json`

- [ ] **Step 1: Edit `package.json` to the framework targets**

Set (exact resolved values, caret/tilde spec so the lock is authoritative):
- deps: `@angular/core`, `@angular/common`, `@angular/compiler`, `@angular/forms`,
  `@angular/platform-browser`, `@angular/platform-browser-dynamic`, `@angular/router`
  → `~8.2.14`
- devDeps: `@angular/compiler-cli` → `~8.2.14`, `@angular/language-service` → `~8.2.14`,
  `@angular/cli` → `~8.3.29`, `@angular-devkit/build-angular` → `^0.803.29`
- deps: `rxjs` → `~6.6.7`, `zone.js` → `~0.9.1`
- devDeps: `typescript` → leave `~3.2.2`

- [ ] **Step 2: Install + build on Node 18 (legacy flag)**

```bash
docker run --rm -v <frontend>:/app -w /app node:18.20 sh -c \
  "rm -rf node_modules && npm ci --no-audit --no-fund \
   && NODE_OPTIONS='--openssl-legacy-provider --max_old_space_size=8192' npx ng build --prod --base-href / 2>&1 | grep -iE 'error|compiled|successfully|chunk \{19\}|chunk \{21\}' | tail -12"
```

Expected: install clean (npm 6-style v1 lock preserved — verify `lockfileVersion: 1`); build
may FAIL on a specific Angular 8 compile error (that's the point — it names the file). If it
fails, record the exact error and fix that file in Step 3. If it SUCCEEDS, skip to 1.2.

- [ ] **Step 3: Fix compile errors one at a time**

Angular 7→8 known compile breaks (fix only those the build names):
- `@angular/http` imports — measured **0** refs, so none expected.
- `NgModule`/`ViewEncapsulation` — unchanged.
- Any `rxjs` 6.4+ operator signature change (e.g. `concatMapTo` removed → `concatMap` to
  constant). Grep `concatMapTo|skipUntil|exhaustMapTo` if the build names an rxjs error.
- Template type-check tightening in 8 (strict-ish) — fix the named binding.

Rebuild after each fix. Stop when the build is green.

- [ ] **Step 4: Confirm the residual node-sass/tar closed**

```bash
npm ls node-sass tar 2>&1 | grep -E 'node-sass@|tar@'
```
Expected: **no** `node-sass@4.13.1`, **no** `tar@2.2.2` (build-angular 0.803 dropped node-sass).
The only `tar` may be `tar@4.4.19` (pacote ← @angular/cli, npm-internal, not bundled).

- [ ] **Step 5: Commit (framework only, libs still on 7 — expect build green or not-yet-8 libs)**

```
chore(frontend): angular 8 framework + toolchain baseline
```

## Task 1.2: Bump the 7 capped libs to their Angular-8 lines

**Files:** `frontend/package.json`, `frontend/package-lock.json`, (any component fixing an API break)

For each lib, set the spec to the target, `npm install`, rebuild, and fix the named break:

- [ ] `@ng-bootstrap/ng-bootstrap` → `^5.1.4`. **Known break:** v5 moved `NgbModal`
  config to an injectable and changed tooltip/popover to components. Grep
  `NgbModalOptions|NgbModal|NgbTooltip|NgbPopover` usage; fix the named files (modal
  open() still returns NgbModalRef; config injection is the change).
- [ ] `@kolkov/angular-editor` → lowest version >= 0.18.8 whose `@angular/core` peer admits
  8 (verify via registry `https://registry.npmjs.org/@kolkov/angular-editor/<v>`).
  `editor` binding API is stable. If none admits 8, keep 0.18.8 (accept-risk pin) and note it.
- [ ] `@ng-select/ng-select` → `^3.x` (lowest admitting 8). Select API stable; theming
  unchanged. Visual QA the selects in the smoke.
- [ ] `codelyzer` (dev) → `^5.0.x`. Lint-only; if `ng lint` surfaces new rules, relax the
  named rule in `tslint.json` (lint is not a build gate).
- [ ] `ng-jhipster` → `^0.10.x`. Grep `ng-jhipster` imports; verify each used util is still
  exported (the plan's Phase 2 matrix flags "verify each used util").
- [ ] `ngx-spinner` → `^8.x` (lowest admitting 8). `<ngx-spinner>` API stable.
- [ ] `@fortawesome/angular-fontawesome` → `^0.4.x` (lowest admitting 8). `fa-icon` inputs
  stable.

Rebuild after each lib (or in one batch if the build isolates errors cleanly). Fix only what
the build names.

- [ ] **Commit:** `chore(frontend): angular-8 peer-lib baseline (ng-bootstrap/ng-select/kolkov/codelyzer/ng-jhipster/spinner/fontawesome)`

## Task 1.3: ngx-draggable-resizer accept-risk pin

- [ ] Keep `ngx-draggable-resizer@1.0.4` (no published version admits Angular 8 — peer
  `~7.1.0`). Record a one-line accept-risk in `docs/security/deps/phase1-notes.md`:
  "drag-handle component only; low runtime surface; npm installs with peer warning."
- [ ] **Commit** (no code change; the note only).

## Task 1.4: Verify the 5 unregistered `@Injectable` still resolve

- [ ] For `TypesService`, `InverHttpAlertManager`, `UtmAlertErrorService`,
  `CanDeactivateConfigGuard`, `DialogService`: grep their class name across `src/` to find how
  they're provided (component `providers`, a factory, `APP_INITIALIZER`, or unused). If any is
  genuinely unprovided AND injected, add the provider; if unused, leave it (dead code, note it).
- [ ] **Commit** if any change was needed.

## Task 1.5: Full build + lab smoke + visual baseline diff

- [ ] **Step 1:** clean production build green on Node 18 (legacy flag) AND on Node 14 (CI
  fallback) — both must pass (CI uses Node 14; the branch will be merged to the v11/v12 lines).
- [ ] **Step 2:** deploy the dist to the lab (slim image from the current lab frontend base +
  `COPY dist/utm-stack /usr/share/nginx/html/`), force-update, headless smoke the 13 baseline
  routes (login, log explorer, alert list, dashboard-main, dash 7/13/227/403/405, rule editor,
  integrations, soar, compliance) at desktop + mobile → 0 pageerrors.
- [ ] **Step 3:** diff the 26 new screenshots against `docs/security/deps/visual-baseline/`
  (the Phase-0 pre-upgrade reference). Every material difference must be explained (the
  intended ones: ng-bootstrap 5 / ng-select 3 / spinner 8 component changes). A difference on a
  route whose libs we didn't touch = regression → fix before exit.
- [ ] **Step 4:** Commit `docs/security/deps/visual-baseline-angular8/` (the new reference for
  Phase 2) + any notes.

## Task 1.6: Phase 1 exit

- [ ] Build green (Node 18 legacy + Node 14), lab smoke 0 pageerrors, visual diff explained,
  `npm ls node-sass` empty, residual tar 2.2.2 gone. Commit
  `chore(frontend): angular 8 baseline`. Update the master plan's Phase 1 checkboxes + exit.

**Exit (per master plan):** build green; visual baseline diff = zero *material* (unexplained)
change; residual node-sass/tar closed; commit `chore(frontend): angular 8 baseline`.

## Deviations from the master plan's Phase-1 text (evidence-based, 2026-11-03)

1. **`typescript ~3.8` → no bump (keep 3.2.2).** build-angular 0.803.29's TS peer is
   `>=3.1 <3.6`; 3.8 would violate it. 3.2.2 already satisfies it.
2. **`@Injectable` → `providedIn` migration is NOT done.** Angular 8 does not require it and
   only 35 classes are affected (30 already module-registered). Forcing a 1486-file DI
   refactor for a single major bump is disproportionate risk. (Revisit at Phase 2 if a later
   major makes it load-bearing.)
3. **7 capped libs bump at this stop** (ng-bootstrap, kolkov, ng-select, codelyzer,
   ng-jhipster, spinner, fontawesome) + 1 accept-risk pin (draggable-resizer) — the master
   plan's 22-lib matrix is a Phase-2 forward story; these 8 *cap below Angular 8* so a
   coherent 8 baseline must move them now.
