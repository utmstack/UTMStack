# Phase 1 notes — accept-risk pins (2026-11-03)

Libs kept at their Angular 7-era versions because **no published release admits
Angular 8** (verified against the npm registry 2026-11-03):

## ngx-draggable-resizer 1.0.4
- Peer: `@angular/core ~7.1.0` (its only 1.x line). npm 6 installs it with a peer
  warning; the component is a small drag-handle wrapper used for resizable table
  columns / panels — low runtime surface, no known CVE surface.
- Disposition: **accept-risk pin**. Revisit in Phase 2 when a later lib (or a
  native `resize` implementation) can replace it; the plan's Phase 2 matrix
  already lists "if unmaintained at Ang 17: accept-risk + pin, or replace with
  native resize" for this lib.

## ng6-toastr-notifications 1.0.4
- Peer: `@angular/core ^6.0.0` (only 1.x line is 6-era; no release admits Angular 8).
  **Actively used** — `app.module.ts` imports `ToastrModule.forRoot()`, and
  `utm-toast.service.ts` drives every app toast through `ToastrManager`
  (success/error/warning/info/custom), so this is high-frequency runtime surface.
- Compile-proven on Angular 8.2.14 (prod build green). Toast display NOT explicitly
  asserted in the Phase 1 smoke (routes booted with 0 pageerrors); confirm on the
  Phase 1 lab smoke pass or first login in the UI. npm 6 installs it with a peer warning.
- Disposition: **accept-risk pin** (documented, no alternative without a replacement
  library). Phase 2 matrix: if still incompatible at 17, hand-roll toasts on
  ngx-toastr or keep as pinned legacy module.

No other lib in the 73-direct-dep tree requires an accept-risk pin at Angular 8.
