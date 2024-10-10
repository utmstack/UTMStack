import {CommonModule} from '@angular/common';
import {CUSTOM_ELEMENTS_SCHEMA, NgModule, NO_ERRORS_SCHEMA} from '@angular/core';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {RouterModule} from '@angular/router';
import {NgbActiveModal, NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {NgSelectModule} from '@ng-select/ng-select';
import {TranslateModule} from '@ngx-translate/core';
import {InlineSVGModule} from 'ng-inline-svg';
import {NgxFlagIconCssModule} from 'ngx-flag-icon-css';
import {InfiniteScrollModule} from 'ngx-infinite-scroll';
import {NgxJsonViewerModule} from 'ngx-json-viewer';
import {NgxSortableModule} from 'ngx-sortable-2';
import {LocalStorageService} from 'ngx-webstorage';
import {AssetsGroupAddModule} from '../assets-discover/shared/components/asset-group-add/assets-group-add.module';
import {AssetsApplyNoteModule} from '../assets-discover/shared/components/assets-apply-note/assets-apply-note.module';
import {AssetsApplyTypeModule} from '../assets-discover/shared/components/assets-apply-type/assets-apply-type.module';
import {AuthServerProvider} from '../core/auth/auth-jwt.service';
import {UtmToastService} from './alert/utm-toast.service';
import {DashboardBehavior} from './behaviors/dashboard.behavior';
import {MenuBehavior} from './behaviors/menu.behavior';
import {VersionUpdateBehavior} from './behaviors/version-update.behavior';
import {AppFilterComponent} from './components/app-filter/app-filter.component';
import {ConfirmIdentityComponent} from './components/auth/confirm-identity/confirm-identity.component';
import {LoginComponent} from './components/auth/login/login.component';
import {PasswordResetFinishComponent} from './components/auth/password-reset/finish/password-reset-finish.component';
import {PasswordResetInitComponent} from './components/auth/password-reset/init/password-reset-init.component';
import {PasswordStrengthBarComponent} from './components/auth/password-strength/password-strength-bar.component';
import {TotpComponent} from './components/auth/totp/totp.component';
import {ContactUsComponent} from './components/contact-us/contact-us.component';
import {
  EmailSettingNotificactionComponent
} from './components/email-setting-notification/email-setting-notificaction.component';
import {
  UtmAdminChangeEmailComponent
} from './components/getting-started/utm-admin-change-email/utm-admin-change-email.component';
import {
  WelcomeToUtmstackComponent
} from './components/getting-started/welcome-to-utmstack/welcome-to-utmstack.component';
import {IncidentSeverityComponent} from './components/incident/incident-severity/incident-severity.component';
import {
  CompleteIncidentModalComponent,
  IncidentStatusComponent
} from './components/incident/incident-status/incident-status.component';
import {FooterComponent} from './components/layout/footer/footer.component';
import {
  HeaderMenuNavigationComponent
} from './components/layout/header/header-menu-navigation/header-menu-navigation.component';
import {HeaderComponent} from './components/layout/header/header.component';
import {MobileHeaderComponent} from './components/layout/header/mobile-header/mobile-header.component';
// tslint:disable-next-line:max-line-length
import {
  UtmDateFormatInfoComponent
} from './components/layout/header/shared/components/utm-date-format-info/utm-date-format-info.component';
import {
  GettingStartedFinishedModalComponent,
  GettingStartedModalComponent,
  UtmGettingStartedComponent
} from './components/layout/header/shared/components/utm-getting-started/utm-getting-started.component';
import {
  UtmLicenseInfoComponent
} from './components/layout/header/shared/components/utm-license-info/utm-license-info.component';
import {
  UtmMenuBurgerComponent
} from './components/layout/header/shared/components/utm-menu-burger/utm-menu-burger.component';
import {
  UtmVaultStatusComponent
} from './components/layout/header/shared/components/utm-vault-status/utm-vault-status.component';
import {
  UtmVersionInfoComponent
} from './components/layout/header/shared/components/utm-version-info/utm-version-info.component';
import {
  UtmNotificationAdComponent
} from './components/layout/header/shared/notification/utm-notification-ad/utm-notification-ad.component';
// tslint:disable-next-line:max-line-length
import {
  UtmNotificationAlertComponent
} from './components/layout/header/shared/notification/utm-notification-alert/utm-notification-alert.component';
// tslint:disable-next-line:max-line-length
import {
  UtmNotificationAssetsComponent
} from './components/layout/header/shared/notification/utm-notification-assets/utm-notification-assets.component';
import {
  PasswordComponent
} from './components/layout/header/shared/notification/utm-notification-user-setting/password/password.component';
import {
  SettingsComponent
} from './components/layout/header/shared/notification/utm-notification-user-setting/settings/settings.component';
// tslint:disable-next-line:max-line-length
import {
  UtmNotificationUserSettingComponent
} from './components/layout/header/shared/notification/utm-notification-user-setting/utm-notification-user-setting.component';
import {UtmNotificationComponent} from './components/layout/header/utm-header-notification/utm-notification.component';
import {SidebarComponent} from './components/layout/sidebar/sidebar.component';
import {NotFoundComponent} from './components/not-found/not-found.component';
import {
  UtmHeaderHealthWarningComponent
} from './components/utm-header-health-warning/utm-header-health-warning.component';
import {UtmLiteVersionComponent} from './components/utm-lite-version/utm-lite-version.component';
import {UtmModuleDisabledComponent} from './components/utm-module-disabled/utm-module-disabled.component';
import {KpiComponent} from './components/utm/charts/kpi/kpi.component';
import {NoDataChartComponent} from './components/utm/charts/no-data-chart/no-data-chart.component';
import {
  AppConfigDeleteConfirmComponent
} from './components/utm/config/app-config-delete-confirm/app-config-delete-confirm.component';
import {AppConfigParamsComponent} from './components/utm/config/app-config-params/app-config-params.component';
import {AppConfigSectionsComponent} from './components/utm/config/app-config-sections/app-config-sections.component';
import {
  AppModuleDisabledWarningComponent
} from './components/utm/config/app-module-disabled-warning/app-module-disabled-warning.component';
import {UtmEmailConfCheckComponent} from './components/utm/config/utm-email-conf-check/utm-email-conf-check.component';
import {
  ElasticMetricHealthComponent
} from './components/utm/elastic/elastic-metric-health/elastic-metric-health.component';
// tslint:disable-next-line:max-line-length
import {
  DashboardFilterSelectComponent
} from './components/utm/filters/dashboard-filter-view/dashboard-filter-select/dashboard-filter-select.component';
import {
  DashboardFilterViewComponent
} from './components/utm/filters/dashboard-filter-view/dashboard-filter-view.component';
import {DateRangeComponent} from './components/utm/filters/date-range/date-range.component';
import {ElasticFilterTimeComponent} from './components/utm/filters/elastic-filter-time/elastic-filter-time.component';
import {RefreshFilterComponent} from './components/utm/filters/refresh-filter/refresh-filter.component';
import {TimeFilterComponent} from './components/utm/filters/time-filter/time-filter.component';
import {UtmDataLimitComponent} from './components/utm/filters/utm-data-limit/utm-data-limit.component';
import {
  ElasticFilterAddComponent
} from './components/utm/filters/utm-elastic-filter/elastic-filter-add/elastic-filter-add.component';
import {ElasticFilterComponent} from './components/utm/filters/utm-elastic-filter/elastic-filter.component';
import {FormcontrolErrorComponent} from './components/utm/form/formcontrol-error/formcontrol-error.component';
import {IncidentVariableSelectComponent} from './components/utm/incident-variables/incident-variable-select.component';
import {
  IrVariableCreateComponent
} from './components/utm/incident-variables/ir-variable-create/ir-variable-create.component';
import {
  IndexPatternCreateComponent
} from './components/utm/index-pattern/index-pattern-create/index-pattern-create.component';
import {
  IndexPatternSelectComponent
} from './components/utm/index-pattern/index-pattern-select/index-pattern-select.component';
import {LogstashStatsComponent} from './components/utm/logstash/logstash-stats/logstash-stats.component';
import {UtmServerSelectComponent} from './components/utm/server/utm-server-select/utm-server-select.component';
import {NoDataFoundComponent} from './components/utm/table/no-data-found/no-data-found.component';
import {SortByComponent} from './components/utm/table/sort-by/sort-by.component';
import {UtmItemsPerPageComponent} from './components/utm/table/utm-items-per-page/utm-items-per-page.component';
import {UtmDynamicTableComponent} from './components/utm/table/utm-table/dynamic-table/dynamic-table.component';
// tslint:disable-next-line:max-line-length
import {
  UtmDtableDetailResolverComponent
} from './components/utm/table/utm-table/shared/component/utm-dtable-detail-resolver/utm-dtable-detail-resolver.component';
import {
  TableDetailContentDirective
} from './components/utm/table/utm-table/shared/directive/table-detail-content.directive';
import {
  UtmDtableColumnsComponent
} from './components/utm/table/utm-table/utm-dtable-columns/utm-dtable-columns.component';
// tslint:disable-next-line:max-line-length
import {
  UtmJsonDetailViewComponent
} from './components/utm/table/utm-table/utm-table-detail/utm-json-detail-view/utm-json-detail-view.component';
// tslint:disable-next-line:max-line-length
import {
  UtmTableDetailViewComponent
} from './components/utm/table/utm-table/utm-table-detail/utm-table-detail-view/utm-table-detail-view.component';
import {UtmTableDetailComponent} from './components/utm/table/utm-table/utm-table-detail/utm-table-detail.component';
import {AppRestartApiComponent} from './components/utm/util/app-restart-api/app-restart-api.component';
import {ColorSelectComponent} from './components/utm/util/color-select/color-select.component';
import {ColorSelectorComponent} from './components/utm/util/color-select/color-selector/color-selector.component';
import {CountdownComponent} from './components/utm/util/countdown/countdown.component';
import {GenericFilerSortComponent} from './components/utm/util/generic-filer-sort/generic-filer-sort.component';
import {IconSelectComponent} from './components/utm/util/icon-select/icon-select.component';
import {MenuCreateComponent} from './components/utm/util/menu-create/menu-create.component';
import {MenuIconSelectComponent} from './components/utm/util/menu-icon-select/menu-icon-select.component';
import {ModalConfirmationComponent} from './components/utm/util/modal-confirmation/modal-confirmation.component';
import {UtmAgentConnectComponent} from './components/utm/util/utm-agent-connect/utm-agent-connect.component';
import {UtmAgentConsoleComponent} from './components/utm/util/utm-agent-console/utm-agent-console.component';
import {UtmAgentDetailComponent} from './components/utm/util/utm-agent-detail/utm-agent-detail.component';
import {UtmAgentSelectComponent} from './components/utm/util/utm-agent-select/utm-agent-select.component';
import {
  UtmChangeDashboardTimeComponent
} from './components/utm/util/utm-change-dashboard-time/utm-change-dashboard-time.component';
import {UtmCodeHighlightComponent} from './components/utm/util/utm-code-highlight/utm-code-highlight.component';
import {UtmCodeViewComponent} from './components/utm/util/utm-code-view/utm-code-view.component';
import {UtmCollectorDetailComponent} from './components/utm/util/utm-collector-detail/utm-collector-detail.component';
import {UtmColorsOrderComponent} from './components/utm/util/utm-colors-order/utm-colors-order.component';
import {UtmConsoleCheckComponent} from './components/utm/util/utm-console-check/utm-console-check.component';
import {UtmCountryFlagComponent} from './components/utm/util/utm-country-flag/utm-country-flag.component';
import {
  UtmFileDragDropDirective
} from './components/utm/util/utm-file-upload/shared/directives/utm-file-drag-drop.directive';
import {UtmFileUploadComponent} from './components/utm/util/utm-file-upload/utm-file-upload.component';
import {UtmInputFileUploadComponent} from './components/utm/util/utm-input-file-upload/utm-input-file-upload.component';
import {UtmInsideModalComponent} from './components/utm/util/utm-inside-modal/utm-inside-modal.component';
import {UtmModalHeaderComponent} from './components/utm/util/utm-modal-header/utm-modal-header.component';
import {
  UtmOnlineDocumentationComponent
} from './components/utm/util/utm-online-documentation/utm-online-documentation.component';
import {UtmPdfPreviewComponent} from './components/utm/util/utm-pdf-preview/utm-pdf-preview.component';
import {UtmProgressbarComponent} from './components/utm/util/utm-progressbar/utm-progressbar.component';
import {UtmReportHeaderComponent} from './components/utm/util/utm-report-header/utm-report-header.component';
import {UtmScrollTopComponent} from './components/utm/util/utm-scroll-top/utm-scroll-top.component';
import {UtmSearchInputComponent} from './components/utm/util/utm-search-input/utm-search-input.component';
import {UtmSecretViewComponent} from './components/utm/util/utm-secret-view/utm-secret-view.component';
import {
  UtmSingleColorSelectComponent
} from './components/utm/util/utm-single-color-select/utm-single-color-select.component';
import {UtmSpinnerComponent} from './components/utm/util/utm-spinner/utm-spinner.component';
import {UtmTagInputComponent} from './components/utm/util/utm-tag-input/utm-tag-input.component';
import {UtmTerminalInputComponent} from './components/utm/util/utm-terminal-input/utm-terminal-input.component';
import {UtmTimeDataRefreshComponent} from './components/utm/util/utm-time-data-refresh/utm-time-data-refresh.component';
import {UtmToggleComponent} from './components/utm/util/utm-toggle/utm-toggle.component';
import {UtmUserSelectComponent} from './components/utm/util/utm-user-select/utm-user-select.component';
import {HasAnyAuthorityDirective} from './directives/auth/has-any-authority.directive';
import {HasEnterpriseLicenseDirective} from './directives/auth/has-enterprise-license.directive';
import { BadgeTypeDirective } from './directives/badge-type/badge-type.directive';
import {UtmInputErrorDirective} from './directives/input-error/utm-input-error.directive';
import {SortableDirective} from './directives/sortable/sortable.directive';
import { TemplateSelectorDirective } from './directives/template-selector/template-selector.directive';
import {ZoomDirective} from './directives/zoom.directive';
import {CapitalizePipe} from './pipes/capitalize.pipe';
import {UtmDatePipe} from './pipes/date.pipe';
import {ThousandSuffPipe} from './pipes/numbers/thousand-suff.pipe';
import {KeysPipe} from './pipes/object-keys/keys.pipe';
import {SafePipe} from './pipes/safe.pipe';
import {HighlightPipe} from './pipes/text/highlight.pipe';
import {TimezoneOffsetPipe} from './pipes/timezone-offset.pipe';
import {UtmNotifier} from './websocket/utm-notifier';


@NgModule({
  imports: [
    InlineSVGModule,
    CommonModule,
    ReactiveFormsModule,
    TranslateModule,
    NgbModule,
    FormsModule,
    RouterModule,
    NgSelectModule,
    NgxSortableModule,
    NgxFlagIconCssModule,
    NgxJsonViewerModule,
    AssetsApplyTypeModule,
    AssetsApplyNoteModule,
    AssetsGroupAddModule,
    InfiniteScrollModule
  ],
  declarations: [
    ElasticFilterComponent,
    ElasticFilterAddComponent,
    ElasticFilterTimeComponent,
    LoginComponent,
    HeaderComponent,
    HasAnyAuthorityDirective,
    NotFoundComponent,
    SidebarComponent,
    PasswordResetInitComponent,
    PasswordResetFinishComponent,
    PasswordStrengthBarComponent,
    SortByComponent,
    UtmSpinnerComponent,
    FormcontrolErrorComponent,
    UtmInputErrorDirective,
    ContactUsComponent,
    TimeFilterComponent,
    DateRangeComponent,
    NoDataFoundComponent,
    UtmModalHeaderComponent,
    NoDataChartComponent,
    FooterComponent,
    KpiComponent,
    IconSelectComponent,
    ColorSelectComponent,
    UtmToggleComponent,
    ColorSelectorComponent,
    SortableDirective,
    UtmInsideModalComponent,
    KeysPipe,
    UtmFileUploadComponent,
    UtmFileDragDropDirective,
    UtmInputErrorDirective,
    UtmSearchInputComponent,
    UtmDtableColumnsComponent,
    UtmItemsPerPageComponent,
    UtmDataLimitComponent,
    MenuCreateComponent,
    ThousandSuffPipe,
    UtmDynamicTableComponent,
    UtmDtableDetailResolverComponent,
    TableDetailContentDirective,
    HighlightPipe,
    ModalConfirmationComponent,
    UtmTagInputComponent,
    UtmCountryFlagComponent,
    ConfirmIdentityComponent,
    UtmTableDetailComponent,
    UtmTableDetailViewComponent,
    UtmJsonDetailViewComponent,
    UtmColorsOrderComponent,
    UtmSingleColorSelectComponent,
    UtmModuleDisabledComponent,
    UtmPdfPreviewComponent,
    HeaderMenuNavigationComponent,
    UtmVaultStatusComponent,
    UtmReportHeaderComponent,
    UtmTerminalInputComponent,
    UtmChangeDashboardTimeComponent,
    UtmTimeDataRefreshComponent,
    AppRestartApiComponent,
    UtmCodeViewComponent,
    UtmNotificationAssetsComponent,
    UtmNotificationAlertComponent,
    UtmNotificationAdComponent,
    UtmNotificationUserSettingComponent,
    UtmMenuBurgerComponent,
    UtmLicenseInfoComponent,
    HasEnterpriseLicenseDirective,
    MobileHeaderComponent,
    GenericFilerSortComponent,
    AppConfigParamsComponent,
    AppConfigSectionsComponent,
    AppModuleDisabledWarningComponent,
    AppConfigDeleteConfirmComponent,
    UtmScrollTopComponent,
    TotpComponent,
    UtmInputFileUploadComponent,
    SafePipe,
    IndexPatternCreateComponent,
    IndexPatternSelectComponent,
    MenuIconSelectComponent,
    UtmProgressbarComponent,
    ElasticMetricHealthComponent,
    SettingsComponent,
    PasswordComponent,
    CountdownComponent,
    UtmVersionInfoComponent,
    UtmLiteVersionComponent,
    UtmEmailConfCheckComponent,
    DashboardFilterViewComponent,
    DashboardFilterSelectComponent,
    UtmServerSelectComponent,
    UtmHeaderHealthWarningComponent,
    UtmUserSelectComponent,
    IncidentSeverityComponent,
    CompleteIncidentModalComponent,
    IncidentStatusComponent,
    UtmDatePipe,
    UtmDateFormatInfoComponent,
    CapitalizePipe,
    UtmAgentConsoleComponent,
    UtmConsoleCheckComponent,
    UtmAgentSelectComponent,
    UtmAgentConnectComponent,
    UtmAgentDetailComponent,
    UtmCollectorDetailComponent,
    UtmGettingStartedComponent,
    GettingStartedModalComponent,
    GettingStartedFinishedModalComponent,
    WelcomeToUtmstackComponent,
    UtmAdminChangeEmailComponent,
    UtmOnlineDocumentationComponent,
    UtmSecretViewComponent,
    LogstashStatsComponent,
    UtmCodeHighlightComponent,
    ZoomDirective,
    IrVariableCreateComponent,
    IncidentVariableSelectComponent,
    EmailSettingNotificactionComponent,
    TimezoneOffsetPipe,
    RefreshFilterComponent,
    UtmNotificationComponent,
    BadgeTypeDirective,
    AppFilterComponent,
    TemplateSelectorDirective
  ],
  exports: [
    IndexPatternCreateComponent,
    LoginComponent,
    HeaderComponent,
    HasAnyAuthorityDirective,
    SidebarComponent,
    PasswordResetInitComponent,
    PasswordResetFinishComponent,
    PasswordStrengthBarComponent,
    SortByComponent,
    UtmSpinnerComponent,
    FormcontrolErrorComponent,
    UtmInputErrorDirective,
    TimeFilterComponent,
    NoDataFoundComponent,
    UtmModalHeaderComponent,
    NoDataChartComponent,
    FooterComponent,
    KpiComponent,
    IconSelectComponent,
    ColorSelectComponent,
    UtmToggleComponent,
    ColorSelectorComponent,
    DateRangeComponent,
    SortableDirective,
    UtmInsideModalComponent,
    KeysPipe,
    UtmFileUploadComponent,
    UtmInputErrorDirective,
    UtmSearchInputComponent,
    ElasticFilterComponent,
    ElasticFilterAddComponent,
    ElasticFilterTimeComponent,
    UtmItemsPerPageComponent,
    UtmDtableColumnsComponent,
    MenuCreateComponent,
    UtmDataLimitComponent,
    UtmDynamicTableComponent,
    HighlightPipe,
    ModalConfirmationComponent,
    UtmTagInputComponent,
    UtmCountryFlagComponent,
    UtmTableDetailComponent,
    UtmTableDetailViewComponent,
    UtmJsonDetailViewComponent,
    UtmColorsOrderComponent,
    ColorSelectComponent,
    UtmSingleColorSelectComponent,
    UtmPdfPreviewComponent,
    UtmReportHeaderComponent,
    UtmTerminalInputComponent,
    UtmTimeDataRefreshComponent,
    AppRestartApiComponent,
    UtmCodeViewComponent,
    HasEnterpriseLicenseDirective,
    MobileHeaderComponent,
    GenericFilerSortComponent,
    ThousandSuffPipe,
    AppConfigParamsComponent,
    AppConfigSectionsComponent,
    AppModuleDisabledWarningComponent,
    UtmScrollTopComponent,
    TotpComponent,
    UtmInputFileUploadComponent,
    SafePipe,
    IndexPatternSelectComponent,
    MenuIconSelectComponent,
    UtmProgressbarComponent,
    ElasticMetricHealthComponent,
    DashboardFilterViewComponent,
    UtmServerSelectComponent,
    UtmHeaderHealthWarningComponent,
    UtmUserSelectComponent,
    IncidentSeverityComponent,
    CompleteIncidentModalComponent,
    IncidentStatusComponent,
    UtmNotificationAlertComponent,
    UtmDatePipe,
    CapitalizePipe,
    UtmAgentConsoleComponent,
    UtmConsoleCheckComponent,
    UtmAgentConnectComponent,
    UtmAgentDetailComponent,
    UtmCollectorDetailComponent,
    UtmAgentSelectComponent,
    UtmSecretViewComponent,
    LogstashStatsComponent,
    UtmCodeHighlightComponent,
    UtmVersionInfoComponent,
    IrVariableCreateComponent,
    IncidentVariableSelectComponent,
    EmailSettingNotificactionComponent,
    TimezoneOffsetPipe,
    RefreshFilterComponent,
    UtmNotificationComponent,
    BadgeTypeDirective,
    AppFilterComponent,
    TemplateSelectorDirective
  ],
  entryComponents: [
    LoginComponent,
    HeaderComponent,
    PasswordResetInitComponent,
    PasswordResetFinishComponent,
    PasswordStrengthBarComponent,
    SortByComponent,
    UtmSpinnerComponent,
    FormcontrolErrorComponent,
    ContactUsComponent,
    TimeFilterComponent,
    DateRangeComponent,
    NoDataFoundComponent,
    UtmModalHeaderComponent,
    NoDataChartComponent,
    FooterComponent,
    IconSelectComponent,
    ColorSelectComponent,
    UtmToggleComponent,
    ColorSelectorComponent,
    UtmInsideModalComponent,
    ElasticFilterComponent,
    ElasticFilterAddComponent,
    ElasticFilterTimeComponent,
    MenuCreateComponent,
    ModalConfirmationComponent,
    UtmCountryFlagComponent,
    UtmTableDetailComponent,
    UtmTableDetailViewComponent,
    UtmJsonDetailViewComponent,
    AppConfigParamsComponent,
    IndexPatternCreateComponent,
    AppConfigDeleteConfirmComponent,
    MenuIconSelectComponent,
    SettingsComponent,
    PasswordComponent,
    CompleteIncidentModalComponent,
    UtmDateFormatInfoComponent,
    GettingStartedModalComponent,
    GettingStartedFinishedModalComponent,
    UtmAdminChangeEmailComponent,
    IrVariableCreateComponent,
    EmailSettingNotificactionComponent],
  providers: [
    UtmToastService,
    MenuBehavior,
    DashboardBehavior,
    UtmNotifier,
    VersionUpdateBehavior,
    AuthServerProvider,
    NgbActiveModal,
    LocalStorageService],
  schemas: [NO_ERRORS_SCHEMA, CUSTOM_ELEMENTS_SCHEMA]
})
export class UtmSharedModule {
}
