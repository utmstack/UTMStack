# Dependabot frontend close-out memo (FINAL — supersedes all interim artifacts)

**Date:** 2026-11-05 · **Branch:** `fix/ir-role-model-v11.3.1` · **Build:** Angular 17.3.12,
Node 18.20.8 / npm 10.8.2, webpack 5, dart-sass · **dist identity:** `main.0e44dbab6d374ec4.js`

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

Method: marker scan over all 37 JS chunks of `dist/utm-stack` (8,524,405 bytes total).
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

**Verdict: the shipped browser bundle carries 0 open advisories** across the full previously-alerted
frontend inventory.

## 3. Dev-scoped residuals (in install tree, proven NOT bundled)

Still present in the lock at their current toolchain versions — all build-time only:
webpack-dev-server 4.15.1 + middleware, protractor 7.0.0 (+ webdriver-manager 12.1.8, adm-zip
0.4.16, the `request` chain), socket.io 4.6.1 / engine.io 6.4.1 (karma HMR), faye-websocket 0.11.3,
node-forge 1.4.3 (dev-server TLS), postcss 8.4.35, node-gyp, tar 6.2.1, sass 1.32.13 (nested
under build-angular). Several still have open advisories *at their installed versions* (e.g.
engine.io 6.4.1 < 6.6.10, socket.io-parser 4.2.2 < 4.2.7, ws 8.11.0 < 8.21.0,
http-proxy-middleware 2.0.8 < 3.0.7, serialize-javascript 6.0.2 < 7.0.5) — Dependabot will keep
flagging them.

**Disposition (apply only with explicit user approval, per the 2026-10-27 report Part E rule):**
dismiss with reason `not_affected` and the verbatim comment:

```
dev/build-time dependency of the Angular CLI toolchain, not present in the shipped
browser bundle (docs/security/deps/dismiss-memo.md, 2026-11-05 marker scan over
main.0e44dbab6d374ec4.js); no runtime path. Control = isolated egress-restricted
build + dist bundle-scan (this memo).
```

**Do NOT dismiss:** the `sockjs-client`/`stompjs` rows (real runtime deps — 0 advisories today,
re-check if they get one), and anything the final alert pull shows as still *shipped*.

Command template:

```bash
gh api repos/utmstack/UTMStack/dependabot/alerts/<N>/dismiss -X POST \
  -f state=dismissed -f dismiss_reason="not_affected" \
  -f dismiss_comment="dev/build-time dependency of the Angular CLI toolchain, not present in the shipped browser bundle (docs/security/deps/dismiss-memo.md, 2026-11-05 marker scan over main.0e44dbab6d374ec4.js); no runtime path. Control = isolated egress-restricted build + dist bundle-scan (this memo)."
```

## 4. Written accept-risks (the only allowed Phase 4 residual)

- `moment` 2.31.0 — latest 2.x; crafted-locale path traversal (medium); no further patch.
- Angular-17 long-tail matrix libs with no patchable release line (documented per lib in
  `docs/plans/2026-11-02-frontend-dep-remediation.md` Phase 2 exit).

## 5. Standing process

1. Isolated, egress-restricted, credential-free frontend build (CI already replicates it:
   Node 18.20.8, `npm ci` from the committed v2 lock — byte-deterministic, verified
   `INSTALL_RC=0` + `BUILD_RC=0` + identical main hash).
2. Dist integrity: the prod build's `main.*.js` hash is recorded per stage in the staged-upgrade
   plan; a poisoned dev tree cannot silently change shipped JS without changing that hash.
3. Post-build marker scan (this memo, section 2) on every release cut.
