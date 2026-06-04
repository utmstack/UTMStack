import {enableProdMode} from '@angular/core';
import {platformBrowserDynamic} from '@angular/platform-browser-dynamic';
import 'echarts-leaflet';


import 'echarts-wordcloud/dist/echarts-wordcloud.js';
import 'echarts/lib/chart/effectScatter';

import 'echarts/lib/chart/scatter';

import 'echarts/theme/macarons.js';
import {AppModule} from './app/app.module';
import {environment} from './environments/environment';
// import 'echarts-leaflet';


if (environment.production) {
  enableProdMode();
}

platformBrowserDynamic().bootstrapModule(AppModule)
  .catch(err => console.error(err));
