import {HTTP_INTERCEPTORS, HttpClientModule} from '@angular/common/http';
import {APP_INITIALIZER, CUSTOM_ELEMENTS_SCHEMA, Injector, NgModule} from '@angular/core';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {BrowserModule} from '@angular/platform-browser';
import {BrowserAnimationsModule} from '@angular/platform-browser/animations';
import {RouterModule} from '@angular/router';
import {NgbDatepickerConfig, NgbModalConfig, NgbTooltipConfig} from '@ng-bootstrap/ng-bootstrap';
import {TranslateLoader, TranslateModule} from '@ngx-translate/core';
import {TRANSLATE_HTTP_LOADER_CONFIG, TranslateHttpLoader} from '@ngx-translate/http-loader';
import * as moment from 'moment';
import {InlineSVGModule} from 'ng-inline-svg';
import {ToastrModule} from 'ngx-toastr';
import {CookieService} from 'ngx-cookie-service';
import {NgxEchartsModule} from 'ngx-echarts';
import {InfiniteScrollModule} from 'ngx-infinite-scroll';
import {NgxSpinnerModule} from 'ngx-spinner';
import {LocalStorageService, NgxWebstorageModule, SessionStorageService} from 'ngx-webstorage';
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
    TranslateModule.forRoot({
      // http-loader 17: TranslateHttpLoader injects HttpClient itself (no-arg ctor)
      loader: {provide: TranslateLoader, useClass: TranslateHttpLoader},
    }),
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
    NgxWebstorageModule.forRoot(),
    AlertManagementSharedModule,
  ],
  providers: [
    // S17 (ngx-webstorage 13): LocalStorageService/SessionStorageService are provided
    // by NgxWebstorageModule.forRoot() (above) via factory — do NOT re-declare them
    // bare here; v13 classes require a StorageStrategy ctor arg and the bare form
    // throws NullInjectorError on every HttpClient call.
    CookieService,
    AuthServerProvider,
    HttpCancelService,
    // S17 (@ngx-translate/http-loader 17): TranslateHttpLoader's ctor calls
    // inject(TRANSLATE_HTTP_LOADER_CONFIG); forRoot's defaultProviders never register
    // that token, so without this the first HttpClient call (TranslateService ->
    // TranslateHttpLoader) throws NullInjectorError. Providing it lets the loader
    // read its config (defaults: /assets/i18n/ prefix, .json suffix).
    {provide: TRANSLATE_HTTP_LOADER_CONFIG, useValue: {}},
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
