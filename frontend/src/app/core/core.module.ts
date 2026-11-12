import {DatePipe, registerLocaleData} from '@angular/common';
import { provideHttpClient, withInterceptorsFromDi } from '@angular/common/http';
import locale from '@angular/common/locales/en';
import {LOCALE_ID, NgModule} from '@angular/core';
import {Title} from '@angular/platform-browser';

@NgModule({ exports: [],
    declarations: [], imports: [], providers: [
        Title,
        {
            provide: LOCALE_ID,
            useValue: 'en'
        },
        DatePipe,
        provideHttpClient(withInterceptorsFromDi())
    ] })
export class UtmstackCoreModule {
  constructor() {
    registerLocaleData(locale);
  }
}
