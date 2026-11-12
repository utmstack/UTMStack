import {HTTP_INTERCEPTORS, HttpClientModule} from '@angular/common/http';
import {APP_INITIALIZER, CUSTOM_ELEMENTS_SCHEMA, Injector, NgModule} from '@angular/core';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {BrowserModule} from '@angular/platform-browser';
import {BrowserAnimationsModule} from '@angular/platform-browser/animations';
import {RouterModule} from '@angular/router';
import {NgbDatepickerConfig, NgbModalConfig, NgbTooltipConfig} from '@ng-bootstrap/ng-bootstrap';
import {provideTranslateService} from '@ngx-translate/core';
import {provideTranslateHttpLoader} from '@ngx-translate/http-loader';
import * as moment from 'moment';
import {InlineSVGModule} from 'ng-inline-svg';
import {ToastrModule} from 'ngx-toastr';
import {CookieService} from 'ngx-cookie-service';
import {NgxEchartsModule} from 'ngx-echarts';
import {InfiniteScrollModule} from 'ngx-infinite-scroll';
import {NgxSpinnerModule} from 'ngx-spinner';
import {LocalStorageService, SessionStorageService, provideNgxWebstorage, withLocalStorage, withSessionStorage} from 'ngx-webstorage';
import {first} from 'rxjs/operators';
import {AppRoutingModule} from './app-routing.module';
import {AppComponent} from './app.component';
import {AuthExpiredInterceptor} from './blocks/interceptor/auth-expired.interceptor';
import {AuthInterceptor} from './blocks/interceptor/auth.interceptor';
import {ErrorHandlerInterceptor} from './blocks/interceptor/errorhandler.interceptor';
import {FederationInstanceInterceptor} from './blocks/interceptor/federation-instance.interceptor';
import {ManageHttpInterceptor} from './blocks/interceptor/managehttp.interceptor';
import {NotificationInterceptor} from './blocks/interceptor/notification.interceptor';
import {HttpCancelService} from './blocks/service/httpcancel.service';
import {AccountService} from './core/auth/account.service';
import {ApiServiceCheckerService} from './core/auth/api-checker-service';
import {AuthServerProvider} from './core/auth/auth-jwt.service';
import {UtmstackCoreModule} from './core/core.module';
import {UtmDashboardModule} from './dashboard/dashboard.module';
import {AlertManagementSharedModule} from './data-management/alert-management/shared/alert-management-shared.module';
import {AlertIncidentStatusChangeBehavior} from './shared/behaviors/alert-incident-status-change.behavior';
import {GettingStartedBehavior} from './shared/behaviors/getting-started.behavior';
import {NavBehavior} from './shared/behaviors/nav.behavior';
import {NewAlertBehavior} from './shared/behaviors/new-alert.behavior';
import {initFederationMode} from './federation/hooks/federation-bootstrap';
import {FederationModeService} from './federation/services/federation-mode.service';
import {TimezoneFormatService} from './shared/services/utm-timezone.service';
import {AppVersionService} from './shared/services/version/app-version.service';
import {UtmSharedModule} from './shared/utm-shared.module';

export function initTimezoneFormat(apiChecker: ApiServiceCheckerService,
                                   timezoneService: TimezoneFormatService,
                                   appVersionService: AppVersionService) {
  return () => {
    apiChecker.init();
    return new Promise<void>((resolve, reject) => {
      apiChecker.isOnlineApi$
        .pipe(first(val => val === true))
        .subscribe({
          next: () => {
            timezoneService.loadTimezoneAndFormat()
              .then(() => appVersionService.loadVersionInfo())
              .then(resolve)
              .catch(reject);
          },
          error: reject
        });
    });
  };

}

@NgModule({
  declarations: [
    AppComponent,
  ],
  imports: [
    InlineSVGModule.forRoot(),
    BrowserModule,
    BrowserAnimationsModule,
    AppRoutingModule,
    RouterModule,
    FormsModule,
    ReactiveFormsModule,
    HttpClientModule,
    UtmSharedModule,
    UtmDashboardModule,
    UtmstackCoreModule,
    ToastrModule.forRoot(),
    InfiniteScrollModule,
    // v17: the directive's NGX_ECHARTS_CONFIG token is NOT @Optional — forRoot()
    // must be called exactly once (root) or every `echarts` element throws
    // NullInjectorError. We pass the default echarts loader explicitly: it
    // lazy-imports the bundled 'echarts' core, the SAME singleton that main.ts's
    // extension imports (echarts-leaflet, effectScatter, macarons theme)
    // register on — so every extension is visible to the charts.
    NgxEchartsModule.forRoot({echarts: () => import('echarts')}),
    NgxSpinnerModule,
    AlertManagementSharedModule,
  ],
  providers: [
    // ngx-webstorage 20: NgxWebstorageModule was removed. Root registration is now
    // the provideNgxWebstorage(...features) environment provider (must be in
    // `providers`, not `imports`), which registers the LocalStorageService /
    // SessionStorageService singletons the interceptors inject.
    provideNgxWebstorage(withLocalStorage(), withSessionStorage()),
    // @ngx-translate/core 18: TranslateModule was removed. The service + HTTP
    // loader are registered as bare providers. provideTranslateHttpLoader()
    // supplies both the TranslateLoader class and its TRANSLATE_HTTP_LOADER_CONFIG
    // token (default prefix /assets/i18n/, suffix .json), which TranslateHttpLoader's
    // ctor injects. We deliberately do NOT pass `lang` here: with `lang` set, the
    // root TranslateService calls use(lang) in its CONSTRUCTOR (v18), forcing an
    // eager translation load during bootstrap that throws NG0200. Instead the
    // language is set lazily in AppComponent (the v17 setDefaultLang('en') spot).
    provideTranslateService(),
    provideTranslateHttpLoader(),
    CookieService,
    AuthServerProvider,
    HttpCancelService,
    {
      provide: HTTP_INTERCEPTORS,
      useClass: AuthInterceptor,
      deps: [SessionStorageService, LocalStorageService],
      multi: true
    },
    {
      provide: HTTP_INTERCEPTORS,
      useClass: FederationInstanceInterceptor,
      multi: true
    },
    {
      provide: HTTP_INTERCEPTORS,
      useClass: AuthExpiredInterceptor,
      deps: [AuthServerProvider, HttpCancelService, AccountService],
      multi: true
    },
    {
      provide: HTTP_INTERCEPTORS,
      useClass: ErrorHandlerInterceptor,
      deps: [AuthServerProvider, HttpCancelService],
      multi: true
    },
    {
      provide: HTTP_INTERCEPTORS,
      useClass: NotificationInterceptor,
      deps: [SessionStorageService, Injector],
      multi: true
    },
    {
      provide: HTTP_INTERCEPTORS,
      useClass: ManageHttpInterceptor,
      multi: true,
    },
    {
      provide: APP_INITIALIZER,
      useFactory: initFederationMode,
      deps: [FederationModeService],
      multi: true
    },
    {
      provide: APP_INITIALIZER,
      useFactory: initTimezoneFormat,
      deps: [
        ApiServiceCheckerService,
        TimezoneFormatService,
        AppVersionService
      ],
      multi: true
    },
    NewAlertBehavior,
    NavBehavior,
    AlertIncidentStatusChangeBehavior,
    GettingStartedBehavior
  ],
  bootstrap: [AppComponent],
  schemas: [CUSTOM_ELEMENTS_SCHEMA],
})
export class AppModule {
  constructor(private dpConfig: NgbDatepickerConfig, private config: NgbModalConfig, private tooltipConfig: NgbTooltipConfig) {
    this.dpConfig.minDate = {year: moment().year() - 100, month: 1, day: 1};
    config.backdrop = 'static';
    this.tooltipConfig.container = 'body';
  }
}
