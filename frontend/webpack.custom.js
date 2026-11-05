/**
 * Custom webpack config (Angular 12+ / webpack 5) — see repo history for the
 * full rationale.
 *
 * Purpose 1 (S12): alias the `leaflet/src/Leaflet` deep import that
 * echarts-leaflet makes to leaflet's UMD build, so webpack 5 doesn't choke on
 * leaflet's `src/Leaflet.js` line-2 `import {version} from '../package.json'`
 * (a JSON named-import webpack 4 tolerated but webpack 5 rejects). Pointing at
 * the package root makes webpack resolve leaflet's `main` (the UMD dist), which
 * is the same bundle our own `import * as L from 'leaflet'` uses — deduped, no
 * second copy.
 *
 * Purpose 2 (S13): alias `intl-tel-input/build/js/utils` to a local NO-OP STUB.
 * That file is a bare IIFE that registers `intlTelInputUtils` on top-level
 * `this`; webpack 5.76 evaluates it in strict scope (this === undefined) so
 * ng2-tel-input's side-effect import throws at bootstrap ("Cannot use 'in'
 * operator to search for 'intlTelInputUtils' in undefined"). The real file is
 * delivered as a classic <script> (see src/index.html + the angular.json asset
 * entry) where sloppy-mode `this === window` sets the global — the library's
 * README delivery mode. See intl-tel-input.utils.js.
 *
 * Purpose 3 (S17 / echarts 4->5): alias `echarts/lib/echarts` to the FULL echarts
 * module (index.js). `echarts-leaflet@1.1.0` is an echarts-3/4-era UMD: its CJS
 * branch does `require('echarts/lib/echarts')` and then
 * `('default' in m ? m.default : m).extendComponentModel(...)`. In echarts 4
 * `lib/echarts.js` WAS the full build (so `.default.extendComponentModel` worked).
 * In echarts 5 `lib/echarts.js` is a compat STUB whose `default` is only `{init}`
 * — the real `extendComponentModel`/`registerCoordinateSystem`/`util`/`graphic`/
 * `matrix` are NAMED exports of the full module. Without this alias the UMD calls
 * `m.extendComponentModel` on the stub at import time and crashes the whole app
 * at bootstrap ("m.extendComponentModel is not a function"). Pointing at
 * index.js (the ESM full build, the same file `import * as echarts from 'echarts'`
 * resolves to) makes `m.default` absent -> `m` = the full namespace -> the leaflet
 * coordinate system registers on the SAME core that `NgxEchartsModule.forRoot()`
 * lazy-loads and that main.ts's other extension imports (wordcloud, scatter,
 * macarons theme) use. Nothing in src imports `echarts/lib/echarts` directly, so
 * the alias only affects echarts-leaflet.
 */
const path = require('path');

module.exports = {
  resolve: {
    alias: {
      'leaflet/src/Leaflet': path.resolve(__dirname, 'node_modules/leaflet'),
      'intl-tel-input/build/js/utils': path.resolve(__dirname, 'intl-tel-input.utils.js'),
      'echarts/lib/echarts': path.resolve(__dirname, 'node_modules/echarts/index.js'),
    },
  },
};
