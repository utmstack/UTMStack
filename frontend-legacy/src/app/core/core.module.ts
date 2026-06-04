import {DatePipe, registerLocaleData} from '@angular/common';
import {HttpClientModule} from '@angular/common/http';
import locale from '@angular/common/locales/en';
import {LOCALE_ID, NgModule} from '@angular/core';
import {Title} from '@angular/platform-browser';

@NgModule({
  imports: [HttpClientModule],
  exports: [],
  declarations: [],
  providers: [
    Title,
    {
      provide: LOCALE_ID,
      useValue: 'en'
    },
    DatePipe
  ]
})
export class UtmstackCoreModule {
  constructor() {
    registerLocaleData(locale);
  }
}
