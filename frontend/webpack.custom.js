/**
 * Custom webpack config (Angular 12+ / webpack 5) — see repo history for the
 * full rationale. Single purpose: alias the `leaflet/src/Leaflet` deep import
 * that echarts-leaflet makes to leaflet's UMD build, so webpack 5 doesn't choke
 * on leaflet's `src/Leaflet.js` line-2 `import {version} from '../package.json'`
 * (a JSON named-import webpack 4 tolerated but webpack 5 rejects).
 *
 * Pointing at the package root makes webpack resolve leaflet's `main`
 * (the UMD dist), which is the same bundle our own `import * as L from 'leaflet'`
 * already uses — so webpack dedupes and there's no second copy of leaflet.
 */
const path = require('path');

module.exports = {
  resolve: {
    alias: {
      'leaflet/src/Leaflet': path.resolve(__dirname, 'node_modules/leaflet'),
    },
  },
};
