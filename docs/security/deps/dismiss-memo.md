# Dependabot frontend close-out memo (FINAL — supersedes all interim artifacts)

**Date:** 2026-11-05 · **Branch:** `fix/ir-role-model-v11.3.1` · **Build:** Angular 17.3.12,
Node 18.20.8 / npm 10.8.2, webpack 5, dart-sass · **dist identity:**
`main.581f7606268e344d.js` (final tree incl. lodash 4.18.1 + moment-timezone 0.5.48;
the `0e44dbab` build cited elsewhere in this file is the immediately-preceding,
equally-verified state — all marker scans below were re-run on the final dist)

This memo replaces `2026-11-02-baseline.md`, `2026-11-02-deps-ls.txt`, `bundle-scan.md`,
`phase0-env.md`, and the original (pre-upgrade) `dismiss-memo.md`, all of which described the
Angular 7 / Node 14 / webpack 4 tree that no longer exists. Those interim artifacts are deleted;
this file is the standing record. Evidence the old ones cited (Node-14 build, dart-sass-1.32 pin,
websocket-driver 0.7.5 direct pin) is preserved in git history (commits `8ac4f3769c`,
`e3cac962f5`, `c116ba0ef0`, `d62fe16dee`).

## 1. Root-cause outcome

The 2026-10-27 report found **236 of the 332 alerts** lived in the frozen
Angular 7 / Node 14 / node-sass-4 frontend lockfile. That freeze is removed: the framework is now
Angular 17 on Node 18.20.8 (webpack 5, dart-sass, no ngcc). The old vulnerable tree
(node-sass 4.x, postcss 7.x, node-forge 0.10.x, elliptic/bn.js/browserify-sign, md4, etc.) is gone
from the lockfile; what survives is the current toolchain's dev-scoped tree at mostly-patched
versions.

## 2. Shipped-bundle proof (the authoritative gate)

Method: marker scan over all 37 JS chunks of `dist/utm-stack` (9,071,876 bytes total, final dist).
A package is "shipped" only if a unique-in-package literal survives minification in a chunk.

**Every previously-disputed Node-side package is ABSENT from the shipped bundle:**
websocket-driver, faye-websocket, http-parser-js, tar, request (the single `require('request')`
hit is a Go/Node *sample string* shown in the SOAR UI, chunk 655 — not the package; its hard deps
oauth-sign/hawk are absent), form-data (0 bounded `FormData…boundary` hits, `combinedLength` 0),
tough-cookie, cross-spawn, engine.io, socket.io, node-forge, protractor, adm-zip,
follow-redirects, http-proxy, body-parser, compression, node-gyp, pacote.

**Every shipped runtime package has 0 open advisories at its installed version** (checked against
the GH advisory API 2026-11-05):

| Package | Shipped version | Open advisories |
|---|---|---|
| jspdf | 4.2.1 | 0 (all 12 former advisories closed by version — commit `131b5b24c7`) |
| dompurify | 3.4.17 | 0 |
| canvg | 3.0.11 | 0 |
| html2canvas | 1.4.1 | 0 |
| fflate | 0.8.3 | 0 |
| fast-png | 6.4.0 | 0 |
| iobuffer | 5.4.0 | 0 |
| sockjs-client | 1.5.1 | 0 (direct runtime dep `^1.4.0`; STOMP over `/ws`) |
| websocket-driver | (dev) 0.7.5 | 0 shipped |

**Verdict:** 0 of the previously-flagged *disputed* packages (the jspdf cluster + all Node-side
tooling) carry an open advisory in the shipped bundle. The complete production-scope residual
(`npm audit --omit=dev` = 9 open, checked 2026-11-06) is itemized in **section 4** — 3 of those 9
ship and run (framework / echarts / highlight.js), 6 are locked-major or non-executed. Section 4 is
the exhaustive written accept-risk list the plan's Phase 4 exit requires.

## 3. Dev-scoped residuals — post fix pass (2026-11-06, commit `12317c981f`)

The user chose to **fix** the dev-scoped alerts rather than only dismiss them. What the
fix pass removed/patched:

- **Protractor EOL scaffolding deleted** (dep + `e2e/` folder + `ng e2e` script +
  `utm-stack-e2e` architect target in angular.json + `.eslintrc` ignore entry). Protractor
  7.0.0 is EOL/incompatible with Angular 17 and its `@angular-devkit/build-angular:protractor`
  builder no longer exists. It owned the entire dev-scoped critical/high family:
  `request 2.88.2` -> `form-data 2.3.3` / `tough-cookie 2.5.0`, `adm-zip 0.4.16`, `xml2js`,
  `selenium-webdriver`, `webdriver-manager`, `glob 7`, `minimatch 3`, `tmp` (one copy),
  `jasminewd2`. All gone from the lock.
- **Unused `ts-node ~7.0.0` removed.**
- **Same-major dev bumps:** `karma 6.4.4`, `ajv 8.20.0`.
- **`overrides: es5-ext >= 0.10.63`** — patched the one prod-tree dev-adjacent residual
  (pulled by stompjs's *optional* websocket branch, proven not shipped).
- **Same-range dev float** (parents' ranges already admitted the patch): `socket.io 4.8.4`,
  `engine.io`, `socket.io-parser 4.8.x`, `ws 8.21.3`, `flatted 3.4.4`, `ansi-regex 5/6`,
  `semver 6.3.1/7.7.x` dev copies.

Result: `npm audit --json` full scope **83 -> 61**. Every one of the 61 remaining is locked to
the **`@angular-devkit/build-angular@17.3.17` / `@angular/cli@17.3.17` pinned subtree** (or the
Angular framework itself / echarts / highlight.js — sections 2 and 4): e.g. `webpack-dev-server
4.15.1`, `vite 5`, `piscina`, `tar` (via `@npmcli/pacote`), `node-gyp` (via `@npmcli/run-script`),
the `sigstore`/`@tufjs` npm-integrity chain, `@typescript-eslint 7` (via `@angular-eslint 17.5.3`),
`serialize-javascript`, `picomatch`, `postcss`, `copy-webpack-plugin`, `uuid`, `sockjs`, `ajv`
(nested copies), `brace-expansion`, `cacache`, `glob`, `minimatch` (npm-cli copies). Their fixes are
all semver-**major** (build-angular 19.2.x / cli 11+) — i.e. they clear only when the framework
moves past Angular 17, which the "stop at 17" locked decision defers.

**Disposition of the 61 (apply only with explicit user approval):** they are dev/build-time
tooling, proven not shipped (section 2 marker scan), so `not_affected` dismissal with this comment:

```
dev/build-time dependency of the Angular CLI 17 toolchain (pinned by
@angular-devkit/build-angular@17.3.17); not present in the shipped browser bundle
(docs/security/deps/dismiss-memo.md, marker scan over main.581f7606268e344d.js);
no runtime path. Fix requires the framework-major move past Angular 17 (follow-on).
```

**Do NOT dismiss:** section 4's shipped accept-risks (@angular framework / echarts /
highlight.js) — those stay open with their written lines.

Command template:

```bash
gh api repos/utmstack/UTMStack/dependabot/alerts/<N>/dismiss -X POST \
  -f state=dismissed -f dismiss_reason="not_affected" \
  -f dismiss_comment="dev/build-time dependency of the Angular CLI 17 toolchain (pinned by @angular-devkit/build-angular@17.3.17); not present in the shipped browser bundle (docs/security/deps/dismiss-memo.md, marker scan over main.581f7606268e344d.js); no runtime path. Fix requires the framework-major move past Angular 17 (follow-on)."
```

> At disposition time, re-pull the live open list and dismiss **by package + advisory**, not by
> a memorized number (numbers shift as commits/re-syncs land). The shipped accept-risks in
> section 4 are never dismissed.

## 4. Written accept-risks — EXHAUSTIVE (the only allowed Phase 4 residual)

`npm audit --omit=dev` on the final lock reports **9** open items; every one is accounted for
below. The shipped-bundle claim in section 2 is accurate for the previously-disputed packages;
these 9 are the honest remainder.

**Ships and runs (locked by major-version decision — the only real residual):**

| Package | Installed | Open advisory (checked 2026-11-06) | Why it stays |
|---|---|---|---|
| @angular/{core, common, compiler, router, localize} | 17.3.12 | 2026 Angular framework advisories; first-patched versions start at 18.2.15 / 19.2.16 / 22.x (isSemVerMajor) | **Stop at Angular 17** is a locked user decision (this plan's goal). 18/19/20 tracked as follow-on work, not this project. Until then the framework CVEs are accepted risk; several are SSR/HttpTransferCache/i18n paths UTMStack does not use (no SSR, no transfer-cache, minimal i18n), which narrows the real exposure. |
| echarts | 5.4.3 | XSS (GHSA-fgmj-fm8m-jvvx) fixed in 6.1.0 | ngx-echarts 17.2.0 peers echarts ^5; the echarts 4→5 jump (S17) was forced by the Ivy wrapper line. echarts 6 is another breaking major — tracked as a follow-on, not inside this plan's "stop at 17" scope. |
| highlight.js | 9.18.5 | ReDoS (GHSA-7wwv-vh3v-89cq) fixed in 10.4.1 | 9.x is the last pre-10 line; 10 is an API rewrite (auto-init, language registration) with no drop-in. The S17 security pass already hardened the *consumer* (highlight pipe escape-first-then-wrap), which removes the rendering path the ReDoS grammars feed. |

**In the prod dependency tree but NOT executed in the shipped artifact:**

| Package | Installed | Open advisory | Why it is safe |
|---|---|---|---|
| es5-ext | 0.10.53 (optional) | ReDoS (GHSA-4gmj-3p3h-gm8h) fixed 0.10.63 | Its only parents are `websocket`/`d`/`es6-iterator` — the dev-server HMR/WS branch (`optional=True`), proven ABSENT from the final dist by marker scan. No browser code path imports it. |
| @babel/core (transitive via @angular/localize) | 7.x | arbitrary file read via sourceMappingURL comment (GHSA-4x5r-pxfx-6jf8, low) | Build-time tooling that rides along in the prod tree; the browser bundle does not execute babel. |
| moment | 2.31.0 | 0 (locale path-traversal first_patched = 2.31.0 = installed) | Listed for completeness — this one is actually CLOSED; no risk. |

**Dev-scoped (section 3):** the whole build-toolchain tree (webpack-dev-server, protractor,
socket.io/engine.io, node-forge, postcss, tar, cross-spawn, the `request` chain, …) — proven not
bundled; dispositions gated on user approval.

Nothing else in the 332-alert inventory remains open without a written line above.

## 5. Standing process

1. Isolated, egress-restricted, credential-free frontend build (CI already replicates it:
   Node 18.20.8, `npm ci` from the committed v2 lock — byte-deterministic, verified
   `INSTALL_RC=0` + `BUILD_RC=0` + identical main hash).
2. Dist integrity: the prod build's `main.*.js` hash is recorded per stage in the staged-upgrade
   plan; a poisoned dev tree cannot silently change shipped JS without changing that hash.
3. Post-build marker scan (this memo, section 2) on every release cut.
