# Frontend disputed-package bundle scan (2026-11-02, post node-sass→dart-sass + websocket-driver 0.7.5)

Build: `ng build --prod`, Node 14.16.1 (docker `node:14.16.1`), dart-sass 1.32.13 active.
Output: `dist/utm-stack` (597 files, 48 top-level .js, `main.f884a8ec…js` 5.04 MB, `styles.18cee4235e…css` 320 kB).

Method: `grep -rlF <marker> dist/utm-stack/*.js`. Marker strings are unique-in-package literals
that survive minification. A package is "shipped" only if its marker is found.

## Marker results

| Marker | Package | In dist? |
|---|---|---|
| `websocket-driver` | websocket-driver | **ABSENT** |
| `faye-websocket` | faye-websocket | **ABSENT** |
| `form-data` | form-data | **ABSENT** |
| `combinedLength` | form-data (internal) | **ABSENT** |
| `tough-cookie` | tough-cookie | **ABSENT** |
| `Tar::` / `createTarStream` / `parseEntry` | tar | **ABSENT** |
| `cross-spawn` | cross-spawn | **ABSENT** |
| `ENOENT.*spawn` | cross-spawn (error path) | **ABSENT** |
| `oauth-sign` | (hard dep of `request`) | **ABSENT** |
| `hawk` | (hard dep of `request`) | **ABSENT** |
| `request` (substring) | — | 21 files (false positive, see below) |
| `html2canvas` | html2canvas | **PRESENT** (main, runtime) |
| `jspdf` / `jsPDF` | jspdf | **PRESENT** (main, runtime) |
| `dompurify` | dompurify | **PRESENT** (main + chunk 47) |
| `canvg` | canvg | **PRESENT** (main + chunk 23) |
| `fflate` | fflate | **ABSENT** (jspdf 2.5.2 optional, not exercised) |

## The `request` false positive, explained

`request` matches 21 files, but those are the English **word** in minified browser
identifiers (`requestAnimationFrame`, `XMLHttpRequest`, `pendingRequest`, …), not the
npm `request@2.88.2` package. The npm package is conclusively **not** bundled: webpack
bundles a package's entire require-graph, and `request`'s hard dependencies
`oauth-sign` and `hawk` are both ABSENT. If `request` shipped, they would ship with it.
`request` is a Node-only HTTP client (uses `http`/`https`/`stream`); no browser code imports it.

## Dispositions

| Package | Version | Declares-it (npm registry) | Shipped in dist? | Disposition |
|---|---|---|---|---|
| websocket-driver | 0.7.5 | faye-websocket (dep of sockjs-client, Node server branch only) | NO (marker absent) | **PATCHED** (Task 1.3, pinned direct at 0.7.5); never bundled |
| tar | 2.2.2 | node-gyp@3.8.0 ← node-sass@4.13.1 ← @angular-devkit/build-angular | NO (marker absent) | **PERSISTS** in dev tree (build-angular hard-depends on node-sass@4.13.1; Task 1.2 removed only the direct 4.14.1) — DISMISS: dev-time install tooling, not bundled; leaves tree in Part 2 when build-angular is upgraded. (Separate tar@4.4.19 ← pacote ← @angular/cli also npm-internal, never bundled.) |
| request | 2.88.2 | node-sass, node-gyp, less@3.9.0, protractor→webdriver-manager | NO (false-positive word; oauth-sign/hawk absent) | **PERSISTS** in dev tree (less/protractor parents survive) — DISMISS: dev-time tooling, not bundled, EOL package |
| form-data | 2.3.3 | request | NO (marker absent) | PERSISTS with request — DISMISS: dev-time, not bundled |
| tough-cookie | 2.5.0 | request | NO (marker absent) | PERSISTS with request — DISMISS: dev-time, not bundled |
| cross-spawn | 3.0.1 (node-sass@4.13.1) / 5.1.0 (compiler-cli→yargs→os-locale→execa) / 6.0.5 (webpack-dev-server) | build-angular→node-sass, compiler-cli, webpack-dev-server | NO (marker absent) | ALL PERSIST (dev parents survive Task 1.2) — DISMISS: dev-only build tooling, not bundled |
| jspdf | 2.5.2 | direct (export-to-pdf-util) | YES | **ACCEPT RISK**: 12 CVEs unreachable per `docs/2026-10-27-security-review-dependabot-and-app-sweep.md` Part B½ per-CVE trace (browser build, PNG-only addImage, no AcroForm/addJS/addSvgAsImage/html); closes in Part 2 (webpack 5 → jspdf 4.x) |
| dompurify / canvg | 2.5.9 / 3.0.11 | jspdf optionalDeps | YES (bundled) | Follow jspdf — patched versions already in the 2.5.2 tree; move with jspdf in Part 2 |
| html2canvas | 1.0.0-rc.7 | direct | YES (bundled) | Latest 1.x is rc.7 (1.4.x exists); re-run dashboard PDF smoke in Part 2 — the Part F½ >120s headless hang is monitored, not a security-scope defect |
| fflate | 0.8.3 | jspdf optional | NO (not exercised) | n/a — not in bundle |
| bootstrap | 4.6.0 → 4.6.2* | direct | YES | Bump in Part 2 Phase 0 (fixes CVE-2024-6531 Carousel XSS, same major, no API change) |

\* bootstrap 4.6.2 not yet applied — Part 2 Phase 0 action.

## Bottom line

Every disputed **Node-side** package (websocket-driver, tar, request, form-data,
tough-cookie, cross-spawn) is **absent from the shipped browser bundle**. They live only
in the install-time / dev-time dependency tree. The only disputed packages that ship are
the jspdf cluster (jspdf + dompurify + canvg + html2canvas) — runtime PDF export,
accepted-risk per the Part B½ reachability trace, to be closed in Part 2 when webpack 5
makes jspdf 4.x buildable.

## Task 1.6 exit verification — ACTUAL RESULT (2026-11-02)

`npm ls` on the post-change tree (Node 14.16.1 docker, committed state = sass 1.32.13 +
websocket-driver 0.7.5):

- **websocket-driver → `0.7.5` in all three tree positions** (`sockjs-client@1.5.1 →
  faye-websocket@0.11.3`, `webpack-dev-server → sockjs@0.3.19 → faye-websocket@0.10.0`,
  and the hoisted root) — **the critical CVE is closed**. Lock diff was exactly
  websocket-driver 0.7.4→0.7.5 + http-parser-js 0.5.3→0.5.10 (its transitive).
- **`npm ls tar` → `2.2.2` still present** (dev tree via build-angular → node-sass@4.13.1
  → node-gyp@3.8.0), plus `4.4.19` (pacote, npm-internal). The original "no 2.2.2"
  expectation was wrong: removing our *direct* node-sass@4.14.1 does not evict
  build-angular's nested node-sass@4.13.1. tar leaves the tree in Part 2 (build-angular
  upgrade). Dismissal stands on the "not shipped" evidence above.
- **`npm ls node-sass` → only the nested build-angular copy 4.13.1** (no direct 4.14.1).
  It is inert: build-angular's `styles.js` passes `implementation: require('sass')`
  (dart-sass) to sass-loader, so the nested node-sass is never invoked for our SCSS.
- **`request@2.88.2` still listed** (less/protractor/node-sass chains) — expected,
  dev-only, never bundled (marker scan above).
- **Karma could not run in this environment**: the `node:14.16.1` docker image has no
  browser binary, and a Windows `chrome.exe` cannot execute inside the Linux container.
  The compile-time gate (full `ng build --prod` green on dart-sass, all SCSS recompiled)
  and the behavioral gate (lab headless smoke — echarts object defined, 2 charts +
  2 tables rendered, 0 pageerrors) are the evidence recorded here and in the lab.
