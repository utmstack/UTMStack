import {HTTP_INTERCEPTORS, HttpClient, HttpClientModule} from '@angular/common/http';
import {CUSTOM_ELEMENTS_SCHEMA, Injector, NgModule} from '@angular/core';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {BrowserModule} from '@angular/platform-browser';
import {BrowserAnimationsModule} from '@angular/platform-browser/animations';
import {RouterModule} from '@angular/router';
import {NgbDatepickerConfig, NgbModalConfig} from '@ng-bootstrap/ng-bootstrap';
import {TranslateLoader, TranslateModule} from '@ngx-translate/core';
import {TranslateHttpLoader} from '@ngx-translate/http-loader';
import * as moment from 'moment';
import {InlineSVGModule} from 'ng-inline-svg';
import {Ng2TelInputModule} from 'ng2-tel-input';
import {ToastrModule} from 'ng6-toastr-notifications';
import {CookieService} from 'ngx-cookie-service';
import {NgxEchartsModule} from 'ngx-echarts';
import {NgxFlagIconCssModule} from 'ngx-flag-icon-css';
import {InfiniteScrollModule} from 'ngx-infinite-scroll';
import {NgxSpinnerModule} from 'ngx-spinner';
import {LocalStorageService, Ng2Webstorage, SessionStorageService} from 'ngx-webstorage';
import {AppRoutingModule} from './app-routing.module';
import {AppComponent} from './app.component';
import {AuthExpiredInterceptor} from './blocks/interceptor/auth-expired.interceptor';
import {AuthInterceptor} from './blocks/interceptor/auth.interceptor';
import {ErrorHandlerInterceptor} from './blocks/interceptor/errorhandler.interceptor';
import {ManageHttpInterceptor} from './blocks/interceptor/managehttp.interceptor';
import {NotificationInterceptor} from './blocks/interceptor/notification.interceptor';
import {HttpCancelService} from './blocks/service/httpcancel.service';
import {AuthServerProvider} from './core/auth/auth-jwt.service';
import {UtmstackCoreModule} from './core/core.module';
import {UtmDashboardModule} from './dashboard/dashboard.module';
import {AlertIncidentStatusChangeBehavior} from './shared/behaviors/alert-incident-status-change.behavior';
import {GettingStartedBehavior} from './shared/behaviors/getting-started.behavior';
import {NavBehavior} from './shared/behaviors/nav.behavior';
import {NewAlertBehavior} from './shared/behaviors/new-alert.behavior';
import {TimezoneFormatService} from './shared/services/utm-timezone.service';
import {UtmSharedModule} from './shared/utm-shared.module';

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
        loader: {
          provide: TranslateLoader,
          useFactory: (http: HttpClient) => {
            return new TranslateHttpLoader(http);
          },
          deps: [HttpClient]
        }
      },
    ),
    UtmSharedModule,
    UtmDashboardModule,
    UtmstackCoreModule,
    ToastrModule.forRoot(),
    InfiniteScrollModule,
    NgxEchartsModule,
    NgxSpinnerModule,
    Ng2TelInputModule,
    NgxFlagIconCssModule,
    Ng2Webstorage.forRoot(),
  ],
  providers: [
    LocalStorageService,
    CookieService,
    AuthServerProvider,
    SessionStorageService,
    HttpCancelService,
    {
      provide: HTTP_INTERCEPTORS,
      useClass: AuthInterceptor,
      deps: [SessionStorageService, LocalStorageService],
      multi: true
    },
    {
      provide: HTTP_INTERCEPTORS,
      useClass: AuthExpiredInterceptor,
      deps: [AuthServerProvider, HttpCancelService],
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
    NewAlertBehavior,
    NavBehavior,
    AlertIncidentStatusChangeBehavior,
    GettingStartedBehavior,
    TimezoneFormatService
  ],
  bootstrap: [AppComponent],
  schemas: [CUSTOM_ELEMENTS_SCHEMA],
})
export class AppModule {
  constructor(private dpConfig: NgbDatepickerConfig, private config: NgbModalConfig) {
    this.dpConfig.minDate = {year: moment().year() - 100, month: 1, day: 1};
    config.backdrop = 'static';
    //timezoneFormatService.loadTimezoneAndFormat();
  }
}
