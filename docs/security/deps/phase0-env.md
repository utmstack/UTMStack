# Phase 0 — environment findings (2026-11-03)

De-risk foundation for the Part 2 upgrade. Verified on `frontend/` with the
docker `node:18.20` image (Node 18.20.8).

## 1. Node 18 build flag (Phase 0 Action 2) — VERIFIED

Angular 7's webpack 4 uses the **md4** hash for chunk/bundle naming. Node 18 ships
OpenSSL 3, which no longer exposes md4 in the default provider, so a plain
`ng build --prod` on Node 18 hard-fails:

```
Error: error:0308010C:digital envelope routines::unsupported
    at handleParseError (/app/node_modules/webpack/lib/NormalModule.js:449:10)
```

With the legacy provider flag it builds green:

```
NODE_OPTIONS="--openssl-legacy-provider --max_old_space_size=8192" npx ng build --prod --base-href /
# -> chunk {19} main.<hash>.js (main) 5.04 MB [initial] [rendered]
# -> chunk {21} styles.<hash>.css (styles) 320 kB [initial] [rendered]
# -> chunk {scripts} scripts.<hash>.js (scripts) 1.38 MB [entry] [rendered]
```

**Node-18 build flag (recorded): `NODE_OPTIONS=--openssl-legacy-provider`.**

Notes:
- `npm ci` on Node 18 succeeds (1332 packages, deprecation warnings only — the
  node-sass@4.13.1 nested under build-angular installs via a downloaded prebuilt
  binding, no native compile needed on this tree).
- The flag is **Node-18-only** and must NOT be baked into `package.json`'s `build`
  script: `--openssl-legacy-provider` is an invalid Node flag on Node 14, so it
  would break the existing Node 14 CI (reusable-node.yml) and any Node 14 build.
  It is set by the caller (CI / build script) once the toolchain moves to Node 18.
  It becomes **unnecessary at Angular 12+** (webpack 5), where md4 is gone — so it
  is a Phase-0/Phase-1/early-Phase-2 bridge flag, not a permanent one.

## 2. Karma (Phase 0 Action 4) — limitation, same as Part 1

`src/karma.conf.js` uses `frameworks: ['jasmine', '@angular-devkit/build-angular']`,
`browsers: ['Chrome']`, `jasmine-core ~2.99.1`, `karma ^6.4.1`, `karma-chrome-launcher ~2.2.0`.
All pure-JS and install fine on Node 18. But:
- Running the suite needs a **browser binary**. The `node:18.20` docker image has
  none, and a Windows `chrome.exe` cannot execute inside the Linux container — the
  same environment limitation that blocked Karma in Part 1 (Node 14).
- The Angular 7 karma builder is webpack-4-based, so on Node 18 it ALSO needs
  `NODE_OPTIONS=--openssl-legacy-provider` to compile the spec bundle.

Resolution (consistent with Part 1 and the plan's exit criteria): the **compile
gate** is `ng build --prod` (every TS spec is type-checked/compiled through the
same webpack pipeline, plus all SCSS), and the **behavioral gate** is the lab
headless smoke (0 pageerrors across the baseline routes). Karma is run in
CI/where a browser exists, not in this box.

## 2b. bootstrap 4.6.0 → 4.6.2 — known cosmetic delta (note for Phase 1 visual review)

The styles bundle hash changes across the bump (`styles.18cee4235e…css` →
`styles.3ac07fb5…css`, 320→321 kB), i.e. 4.6.2 emits different CSS than 4.6.0. This is
expected: 4.6.1/4.6.2 carry minor CSS fixes (carousel XSS patch in 4.6.2 = CVE-2024-6531;
community reports note small-font sizing shifts between 4.6.0 and 4.6.2). **At the Phase 1
visual-baseline diff, any layout/font-size delta on these routes is the intentional
bootstrap change, not an Angular-upgrade regression** — the upgrade diff should be measured
against the *post-bootstrap* baseline, not the pre-Phase-0 one. (The 26 screenshots in
`visual-baseline/` are the pre-bootstrap reference; re-capture after the 4.6.2 deploy is
committed if a tighter Phase-1 baseline is wanted.)

## 3. Visual baseline (Phase 0 Action 3) — captured

26 screenshots (13 routes × desktop 1440×900 + mobile 390×844) in
`docs/security/deps/visual-baseline/`, 0 pageerrors on every route.
Dashboard render route is `/dashboard/render/:id/:dashboard` (two params; the
resolver reads only `.id`). Correct baseline routes:
`/7/threat-activity`, `/13/o365-overview`, `/227/aws-cloud`, `/403/pfsense-firewall`,
`/405/gcp-activity`. (dash-403 renders empty — the lab has no pfSense data; that is
the correct pre-upgrade reference.)
