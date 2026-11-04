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
 */
const path = require('path');

module.exports = {
  resolve: {
    alias: {
      'leaflet/src/Leaflet': path.resolve(__dirname, 'node_modules/leaflet'),
      'intl-tel-input/build/js/utils': path.resolve(__dirname, 'intl-tel-input.utils.js'),
    },
  },
};
