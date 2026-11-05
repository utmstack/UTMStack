#!/usr/bin/env node
'use strict';
/**
 * S16 View Engine dependency compilation shim (REMOVE AT S17 — see
 * docs/plans/2026-11-03-angular17-staged-upgrade.md, S16 break 5).
 *
 * Why this exists:
 *   Angular 16's webpack builder removed in-build ngcc, and the CLI's `ngcc`
 *   bin is a no-op deprecation stub (prints the alert, process.exit(0)). A
 *   clean install therefore leaves every View Engine dep raw and the prod
 *   build fails (TS2314 `ModuleWithProviders<T>` on their .d.ts, NG6002 on
 *   the modules). A tree where only the JS is raw instead green-builds but
 *   crashes at bootstrap with "JIT compiler unavailable".
 *
 * What it does on every `npm install` / `npm ci` (postinstall):
 *   1. Runs the last functional ngcc — @angular/compiler-cli@15.2.10, aliased
 *      as @angular/compiler-cli-ngcc — which compiles the ~20 VE deps in
 *      place. TypeScript must stay ~4.9.5: ngcc@15's ts monkey-patch cannot
 *      handle TS >=5.1 (getExpandoInitializer getter-only conflict).
 *   2. Repairs a known ngcc gap: for UMD-layout packages whose `typings`
 *      entry is a barrel (ngx-sortable-2), ngcc compiles the .js but leaves
 *      the per-class .d.ts without ɵ statics, so ngtsc rejects the NgModule.
 *      This step adds the canonical in-class Ivy declarations, matching
 *      ngcc's own output and Angular 16's type arity:
 *        static ɵfac:  ɵngcc0.ɵɵFactoryDeclaration<X, never>;
 *        static ɵmod:  ɵngcc0.ɵɵNgModuleDeclaration<X, never, never, never>;
 *        static ɵcmp:  ɵngcc0.ɵɵComponentDeclaration<X, 'selector', [], {}, {}, [], []>;
 *        static ɵinj:  ɵngcc0.ɵɵInjectorDeclaration<X>;
 *      (A16's ɵɵComponentDeclaration needs 7+ type args — verified from
 *      @angular/core@16 index.d.ts; a 1-arg form is TS2707.) Idempotent:
 *      a no-op once the statics are present.
 *   3. Normalizes ngcc's InputMap encoding for A16 ngtsc. ngcc@15 writes
 *        "prop": { "alias": undefined; "required": undefined; }
 *      for every input that has no explicit @Input('name'). A16's
 *      readInputsType (compiler-cli ngtsc) resolves bindingPropertyName
 *      via readStringType(alias) — `undefined` reads as null — so the
 *      element still matches by selector (0 NG8001) but NO input binding
 *      resolves (178 NG8002 across ~40 components; skipLibCheck silences
 *      the type error, so 0 TS). The fix rewrites `"alias": undefined`
 *      to the member's own key (the correct binding name) and
 *      `"required": undefined` to `false`, across every d.ts carrying
 *      ɵɵ*Declaration types. Idempotent: no-op once normalized.
 *
 * No forks: only touches node_modules at install time.
 */
const { spawnSync } = require('child_process');
const fs = require('fs');
const path = require('path');

const ROOT = path.resolve(__dirname);
const NM = path.join(ROOT, 'node_modules');
const O = '\u0275'; // ɵ
const log = (m) => process.stdout.write('[ngcc-shim] ' + m + '\n');

// ------------------------------------------------------------- 1. run ngcc
const ngccEntry = path.join(NM, '@angular/compiler-cli-ngcc', 'bundles', 'ngcc', 'main-ngcc.js');
if (!fs.existsSync(ngccEntry)) {
  log('compiler-cli-ngcc not present; nothing to do.');
  process.exit(0);
}
log('running ngcc (@angular/compiler-cli 15.2.10)...');
const r = spawnSync(process.execPath, [ngccEntry], { stdio: 'inherit' });
if (r.status !== 0) {
  log('ngcc exited ' + r.status + ' — aborting (build would fail on raw VE deps).');
  process.exit(r.status);
}

// ------------------------------------------- 2. d.ts ɵ-static repair
function* walkFiles(dir, depth) {
  if (depth > 7) return;
  let entries;
  try { entries = fs.readdirSync(dir, { withFileTypes: true }); } catch (e) { return; }
  for (const e of entries) {
    const p = path.join(dir, e.name);
    if (e.name === 'node_modules' || e.name === '__ivy_ngcc__' || e.name.charAt(0) === '.') continue;
    if (e.isDirectory()) yield* walkFiles(p, depth + 1);
    else yield p;
  }
}
const readIfExists = (p) => { try { return fs.readFileSync(p, 'utf8'); } catch (e) { return null; } };

/** class name -> {kinds:Set, selector?:string} from compiled .js. */
function jsStatics(pkgDir) {
  const map = new Map();
  const re = new RegExp('\\b([A-Z]\\w*)\\s*\\.\\s*' + O + '(cmp|mod|dir|inj|pipe)\\s*=', 'g');
  const files = [];
  for (const f of walkFiles(pkgDir, 0)) {
    if (f.endsWith('.js') && !f.includes('.min.')) files.push(f);
  }
  for (const f of files) {
    const t = readIfExists(f);
    if (!t || !t.includes(O)) continue;
    let m;
    re.lastIndex = 0;
    while ((m = re.exec(t)) !== null) {
      if (!map.has(m[1])) map.set(m[1], { kinds: new Set(), selector: undefined });
      map.get(m[1]).kinds.add(m[2]);
    }
  }
  // extract selectors for components from ɵcmp = ... selector: "..."
  for (const [cls, info] of map) {
    if (!info.kinds.has('cmp')) continue;
    for (const f of files) {
      const t = readIfExists(f);
      if (!t || !t.includes(O)) continue;
      const m = new RegExp('\\b' + cls + '\\s*\\.\\s*' + O + 'cmp\\s*=\\s*[^;]*?selector\\s*:\\s*["\']([^"\']+)["\']').exec(t);
      if (m) { info.selector = m[1]; break; }
    }
    if (!info.selector) info.selector = cls.charAt(0).toLowerCase() + cls.slice(1);
  }
  return map;
}

/** `export declare class X { ... }` bodies across the package's .d.ts files. */
function dtsClasses(pkgDir) {
  const out = [];
  for (const f of walkFiles(pkgDir, 0)) {
    if (!f.endsWith('.d.ts')) continue;
    const t = readIfExists(f);
    if (!t) continue;
    let m;
    const re = /export\s+declare\s+class\s+([A-Z]\w*)/g;
    while ((m = re.exec(t)) !== null) {
      const ob = t.indexOf('{', m.index + m[0].length);
      if (ob < 0) continue;
      let depth = 0, cb = -1;
      for (let j = ob; j < t.length; j++) {
        if (t[j] === '{') depth++;
        else if (t[j] === '}') { depth--; if (depth === 0) { cb = j; break; } }
      }
      if (cb < 0) continue;
      out.push({ file: f, cls: m[1], bodyOpen: ob, bodyClose: cb });
    }
  }
  return out;
}

/** Inject in-class Ivy static declarations for each missing class. */
function applyPatch(file, clsToInfo) {
  let t = readIfExists(file);
  if (!t) return 0;
  const ns = O + 'ngcc0';
  const classRe = /export\s+declare\s+class\s+([A-Z]\w*)/g;
  let m, patched = 0;
  const insertions = [];
  while ((m = classRe.exec(t)) !== null) {
    const cls = m[1];
    const info = clsToInfo.get(cls);
    if (!info) continue;
    const ob = t.indexOf('{', m.index + m[0].length);
    let depth = 0, cb = -1;
    for (let j = ob; j < t.length; j++) {
      if (t[j] === '{') depth++;
      else if (t[j] === '}') { depth--; if (depth === 0) { cb = j; break; } }
    }
    if (cb < 0) continue;
    if (t.slice(ob, cb).includes(O + 'fac')) continue; // already has statics
    let ins = '\n';
    ins += '    static ' + O + 'fac: ' + ns + '.' + O + O + 'FactoryDeclaration<' + cls + ', never>;\n';
    if (info.kinds.has('mod')) ins += '    static ' + O + 'mod: ' + ns + '.' + O + O + 'NgModuleDeclaration<' + cls + ', never, never, never>;\n';
    if (info.kinds.has('cmp')) ins += "    static " + O + 'cmp: ' + ns + '.' + O + O + 'ComponentDeclaration<' + cls + ", '" + info.selector + "', [], {}, {}, [], []>;\n";
    if (info.kinds.has('dir')) ins += '    static ' + O + 'dir: ' + ns + '.' + O + O + 'DirectiveDeclaration<' + cls + ", '" + info.selector + "', {}>;\n";
    if (info.kinds.has('inj')) ins += '    static ' + O + 'inj: ' + ns + '.' + O + O + 'InjectorDeclaration<' + cls + '>;\n';
    if (info.kinds.has('pipe')) ins += '    static ' + O + 'pipe: ' + ns + '.' + O + O + 'PipeDeclaration<' + cls + '>;\n';
    insertions.push({ index: cb, text: ins });
    patched++;
  }
  if (patched === 0) return 0;
  insertions.sort((a, b) => b.index - a.index);
  for (const it of insertions) t = t.slice(0, it.index) + it.text + t.slice(it.index);
  if (!new RegExp('import \\* as ' + ns + ' from').test(t)) {
    const importRe = /^import[^;]*;/gm;
    let lastEnd = -1, im;
    while ((im = importRe.exec(t)) !== null) lastEnd = im.index + im[0].length;
    const ins = "import * as " + ns + " from '@angular/core';\n";
    t = lastEnd > 0 ? t.slice(0, lastEnd) + '\n' + ins + t.slice(lastEnd) : ins + t;
  }
  fs.writeFileSync(file, t, 'utf8');
  return patched;
}

function processPkg(pkgDir, name) {
  let pj;
  try { pj = JSON.parse(fs.readFileSync(path.join(pkgDir, 'package.json'), 'utf8')); } catch (e) { return; }
  if (!pj.__processed_by_ivy_ngcc__) return; // ngcc didn't touch it
  const statics = jsStatics(pkgDir);
  if (statics.size === 0) return; // Ivy-native, nothing to repair
  const classes = dtsClasses(pkgDir);
  const byFile = new Map();
  for (const [cls, info] of statics) {
    const decl = classes.find((c) => c.cls === cls);
    if (!decl) continue; // declared elsewhere
    const body = readIfExists(decl.file) || '';
    const ci = body.indexOf('class ' + cls);
    const seg = ci >= 0 ? body.slice(ci, decl.bodyClose) : '';
    if (seg.includes(O + 'fac') || seg.includes(O + 'mod') || seg.includes(O + 'cmp') || seg.includes(O + 'dir')) continue;
    if (!byFile.has(decl.file)) byFile.set(decl.file, new Map());
    byFile.get(decl.file).set(cls, info);
  }
  let patched = 0;
  for (const [file, map] of byFile) patched += applyPatch(file, map);
  if (patched) log('repaired d.ts ɵstatics in ' + name + ': ' + patched + ' class(es)');
}

const tops = fs.readdirSync(NM, { withFileTypes: true });
for (const top of tops) {
  if (!top.isDirectory()) continue;
  if (top.name === '@angular/compiler-cli-ngcc') continue; // the ngcc tool itself
  if (top.name.startsWith('@')) {
    let subs = [];
    try { subs = fs.readdirSync(path.join(NM, top.name), { withFileTypes: true }); } catch (e) {}
    for (const s of subs) if (s.isDirectory()) processPkg(path.join(NM, top.name, s.name), top.name + '/' + s.name);
  } else {
    processPkg(path.join(NM, top.name), top.name);
  }
}

// ------------------------------------------- 3. InputMap normalization
// ngcc@15 emits  "prop": { "alias": undefined; "required": undefined; }
// for every input WITHOUT an explicit @Input('name'). A16's ngtsc reads the
// binding name as readStringType(alias) -> null, so no input binding resolves
// (NG8002) even though the element selector does (0 NG8001). Rewrite alias to
// the member key (the correct binding name) and required undefined -> false.
// Only touches d.ts files that carry ɵɵ*Declaration statics (i.e. compiled
// Ivy/VE libs); idempotent because the pattern disappears after the first run.
function normalizeInputMaps(pkgDir) {
  let filesTouched = 0, entriesFixed = 0;
  for (const f of walkFiles(pkgDir, 0)) {
    if (!f.endsWith('.d.ts')) continue;
    let t = readIfExists(f);
    if (!t || t.indexOf('"alias": undefined') < 0) continue;
    if (t.indexOf(O + O + 'ComponentDeclaration') < 0 && t.indexOf(O + O + 'DirectiveDeclaration') < 0) continue;
    let fixed = 0;
    t = t.replace(/"([A-Za-z_$][A-Za-z0-9_$]*)":\s*\{\s*"alias":\s*undefined/g, (m, k) => {
      fixed++;
      return '"' + k + '": { "alias": "' + k + '"';
    });
    t = t.replace(/"required":\s*undefined/g, () => { fixed++; return '"required": false'; });
    if (fixed) {
      fs.writeFileSync(f, t, 'utf8');
      filesTouched++;
      entriesFixed += fixed;
    }
  }
  return { filesTouched, entriesFixed };
}
let totalFiles = 0, totalEntries = 0;
const imTops = fs.readdirSync(NM, { withFileTypes: true });
for (const top of imTops) {
  if (!top.isDirectory()) continue;
  if (top.name === '@angular') continue; // framework + compiler-cli-ngcc tool: Ivy-native / not libs
  if (top.name.startsWith('@')) {
    let subs = [];
    try { subs = fs.readdirSync(path.join(NM, top.name), { withFileTypes: true }); } catch (e) {}
    for (const s of subs) {
      if (!s.isDirectory()) continue;
      const r = normalizeInputMaps(path.join(NM, top.name, s.name));
      totalFiles += r.filesTouched; totalEntries += r.entriesFixed;
    }
  } else {
    const r = normalizeInputMaps(path.join(NM, top.name));
    totalFiles += r.filesTouched; totalEntries += r.entriesFixed;
  }
}
if (totalFiles) log('normalized InputMap alias/required in ' + totalFiles + ' d.ts file(s), ' + totalEntries + ' entr(y/ies).');
log('done — all ngcc-processed packages carry ɵstatics in their declarations.');
