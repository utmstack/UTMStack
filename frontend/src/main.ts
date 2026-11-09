import {enableProdMode} from '@angular/core';
import {platformBrowserDynamic} from '@angular/platform-browser-dynamic';
// S17 (echarts 5) wiring: all extensions register on the SAME bundled echarts core
// singleton (echarts/core, shared by every ESM entry point — `index.js`, the
// `lib/export/core.js` re-export used by `lib/echarts.js`, and the lib/* compat
// stubs). ngx-echarts 17's forRoot() defaults its `echarts` provider to
// `() => import('echarts')` = the same core, so these registrations are visible to
// every directive-driven chart:
//   - echarts-leaflet  : UMD, resolves its `require('echarts/lib/echarts')` to the
//                        lib/echarts.js compat stub -> shared core.
//   - echarts-wordcloud: ESM index.js, imports 'echarts' -> shared core. (NOT the
//                        dist/ UMD file — under echarts 5 that would require the
//                        package MAIN (dist/echarts.js), a second prebuilt core.)
//   - lib/chart/{scatter,effectScatter}: echarts-5 compat stubs calling
//                        use(ScatterChart|EffectScatterChart) on the shared core.
//   - theme/macarons.js: UMD registering the 'macarons' theme on the shared core.
// The echarts-4-era angular.json `scripts` UMD entries (dist/echarts.js,
// dist/echarts-wordcloud.js) are removed: they set window.echarts (nothing in src
// reads it) and would instantiate a second, unused core.
import 'echarts-leaflet';
import 'echarts-wordcloud';
import 'echarts/lib/chart/effectScatter';
import 'echarts/lib/chart/scatter';
import 'echarts/theme/macarons.js';
import {AppModule} from './app/app.module';
import {environment} from './environments/environment';

if (environment.production) {
  enableProdMode();
}

platformBrowserDynamic().bootstrapModule(AppModule)
  .catch(err => console.error(err));
