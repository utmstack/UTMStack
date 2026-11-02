# Dependabot dismissal memo — frontend npm (2026-11-02)

> **APPLY ONLY WITH EXPLICIT USER APPROVAL** (2026-10-27 report Part E rule: never batch
> dismiss without approval). Each row is a proposed disposition; the human reviews and
> approves/rejects per row. Dismiss command template at the bottom.

Evidence base:
- `docs/security/deps/2026-11-02-baseline.md` (pre-change `npm ls` tree + dist marker scan)
- `docs/security/deps/bundle-scan.md` (post-change dist marker scan + dispositions)
- Live alert pull 2026-11-02: 100 open alerts repo-wide → **57 frontend npm** (25 packages).
  Every one of the 57 is dispositioned below (A: 3, B: 16, C: 38; A+B+C = 57, verified by set diff).

## A. Already at the patched version — no action; auto-close on Dependabot re-sync

| Package | Alert(s) | What happened |
|---|---|---|
| websocket-driver | 970, 971 | Pinned to 0.7.5 (Task 1.3 commit `c116ba0ef0`). Do NOT manually dismiss — the re-sync closes it. |
| moment | 1129 | Installed 2.31.0 = advisory first_patched (GHSA-4p3w-j4w9-5jqw, range `< 2.31.0`) → auto-closes on re-sync. Runtime/bundled (date handling in charts). |

## B. Keep open (accepted risk / fix tracked) — do NOT dismiss

| Package | Installed | Alert(s) | Disposition |
|---|---|---|---|
| @angular/core | 7.2.0 | 1035, 1100 | Framework itself; fix = Part 2 (Angular 17). |
| @angular/common | 7.2.0 | 1031, 1096 | Same. |
| @angular/compiler | 7.2.0 | 1033, 1098 | Same. |
| @angular/router | 7.2.0 | 1130, 1141, 1142 | Same. |
| js-yaml (runtime copy) | 4.3.2 at root | 975, 1042, 1073 | Runtime req `4.1.0` resolves to **4.3.2** in the lock, which is ≥ all three advisories' first_patched (4.3.0 / 4.3.1 / 4.3.2) → **auto-closes on re-sync.** (The 2026-10 security pass bumped it.) The three nested **dev** copies (3.14.1 under cosmiconfig / istanbul-api / tslint — alerts 976, 1038, 1074) are genuinely below first_patched but build-time only, not shipped (bundle-scan) → moved to Class C. |
| dompurify | 2.5.9 (via jspdf optionalDeps) | 978, 1039, 1136 | Bundled (jspdf cluster). Accepted risk per 2026-10-27 report Part B½ per-CVE reachability trace: UTMStack's PNG-only `addImage` usage never exercises the vulnerable SVG/canvg pipeline. Closes in Part 2 Phase 3 (webpack 5 → jspdf 4.x with patched dompurify 3.x). |
| fflate | 0.8.3 (via jspdf optionalDeps) | 1070 | Same cluster, same trace; marker scan shows it is NOT even in dist (jspdf 2.5.2 optional, unused). |
| jspdf | 2.5.2 (direct, runtime) | (12 alerts — NOT in the 100 open; already dismissed in the 2026-10 security pass with the Part B½ trace) | Accepted risk; re-checked at Part 2 Phase 4. |

## C. Dev/build-scoped — proposed DISMISS (not shipped in dist; marker scan 2026-11-02)

Dismiss reason for every row (verbatim):
`dev/build-time dependency, not present in the shipped browser bundle (docs/security/deps/bundle-scan.md, 2026-11-02 marker scan); no runtime path`

| Package | Installed (lock) | Alert(s) | Parent chain (why dev-only) |
|---|---|---|---|
| adm-zip | 0.4.16 | 972, 1104, 1121, 1122, 1123, 1124, 1125 | protractor → webdriver-manager (e2e runner) |
| brace-expansion | 1.1.11 | 993, 1036, 1043, 1126, 1127, 1128 | minimatch (npm/glob tooling) |
| browserslist | 4.21.5 | 1068, 1069 | autoprefixer/postcss (build) |
| body-parser | 1.20.2 / 1.20.1 | 974 | express (webpack-dev-server) |
| compression | 1.7.4 | 1134 | express (webpack-dev-server) |
| decode-uri-component | 0.2.0 | 1067 | query-string (build tooling) |
| engine.io | 6.4.1 | 973 | socket.io (webpack-dev-server HMR) |
| js-yaml (dev copies) | 3.14.1 ×3 (cosmiconfig / istanbul-api / tslint) | 976, 1038, 1074 | dev build-time copies (config loaders / coverage / linter); genuinely below first_patched but NOT shipped (bundle-scan); the runtime copy is 4.3.2 (Class B, auto-closes) |
| node-forge | 0.10.0 | 1131 | webpack 4 dev-server self-signed TLS |
| pbkdf2 | 3.1.1 | 1138 | karma (test runner) |
| postcss | 7.0.39 / 7.0.14 | 998, 999, 1037 | @angular-devkit/build-angular (build) |
| proxy-addr | 2.0.7 | 1135 | express (webpack-dev-server) |
| qs | 6.11.0 / 6.5.2 | 1071 | express/body-parser (webpack-dev-server) |
| socket.io-parser | 4.2.2 | 1027 | socket.io (webpack-dev-server HMR) |
| tar | 2.2.2 | 979, 980, 981, 982, 1000 | node-gyp@3.8.0 ← node-sass@4.13.1 ← build-angular; dev install-time, marker-absent from dist. Persists in the dev tree until Part 2 (build-angular upgrade); dismissal stands on "not shipped" evidence |
| webpack-dev-middleware | 3.5.1 / 3.4.0 | 1120 | webpack-dev-server (build) |
| webpack-dev-server | 3.1.14 | 994, 995 | @angular/cli dev server (build) |

## Dismiss command template (per-alert, with approval)

```bash
gh api repos/utmstack/UTMStack/dependabot/alerts/<N>/dismiss \
  -X POST \
  -f state=dismissed \
  -f dismiss_reason="not_affected" \
  -f dismiss_comment="dev/build-time dependency, not present in the shipped browser bundle (docs/security/deps/bundle-scan.md, 2026-11-02 marker scan); no runtime path"
```

**Do NOT dismiss:** Class A (auto-closes), Class B (framework / accepted-risk — Part 2 and Phase 4 own those).
