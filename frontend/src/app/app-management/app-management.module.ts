import {CommonModule} from '@angular/common';
import {CUSTOM_ELEMENTS_SCHEMA, NgModule, NO_ERRORS_SCHEMA} from '@angular/core';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {NgSelectModule} from '@ng-select/ng-select';
import {InlineSVGModule} from 'ng-inline-svg';
import {DndModule} from 'ngx-drag-drop';
import {NgxGaugeModule} from 'ngx-gauge';
import {NgxSortableModule} from 'ngx-sortable-2';
import {ComplianceManagementModule} from '../compliance/compliance-management/compliance-management.module';
import {NavBehavior} from '../shared/behaviors/nav.behavior';
import {VersionUpdateBehavior} from '../shared/behaviors/version-update.behavior';
import {TimezoneFormatService} from '../shared/services/utm-timezone.service';
import {UtmSharedModule} from '../shared/utm-shared.module';
import {AppConfigComponent} from './app-config/app-config.component';
import {AppLogsComponent} from './app-logs/app-logs.component';
import {AppManagementRoutingModule} from './app-management-routing.module';
import {AppManagementComponent} from './app-management.component';
import {AppMetricsComponent} from './app-metrics/app-metrics.component';
import {UtmHttpRequestsPreviewComponent} from './app-metrics/utm-http-requests-preview/utm-http-requests-preview.component';
import {UtmJvmMetricsComponent} from './app-metrics/utm-jvm-metrics/utm-jvm-metrics.component';
import {UtmServicesOverviewComponent} from './app-metrics/utm-services-overview/utm-services-overview.component';
import {UtmSystemMetricsComponent} from './app-metrics/utm-system-metrics/utm-system-metrics.component';
import {UtmThreadDetailComponent} from './app-metrics/utm-thread-metrics/utm-thread-detail/utm-thread-detail.component';
import {UtmThreadMetricsComponent} from './app-metrics/utm-thread-metrics/utm-thread-metrics.component';
import {AppThemeConfigComponent} from './app-theme/app-theme-config/app-theme-config.component';
import {AppThemeComponent} from './app-theme/app-theme.component';
import {AuditsComponent} from './audits/audits.component';
import {ConnectionKeyComponent} from './connection-key/connection-key.component';
import {TokenActivateComponent} from './connection-key/token-activate/token-activate.component';
import {HealthChecksComponent} from './health-checks/health-checks.component';
import {HealthClusterComponent} from './health-checks/health-cluster/health-cluster.component';
import {HealthDetailComponent} from './health-checks/health-detail/health-detail.component';
import {IndexDeleteComponent} from './index-management/index-delete/index-delete.component';
import {IndexManagementComponent} from './index-management/index-management.component';
import {IndexPatternDeleteComponent} from './index-pattern/index-pattern-delete/index-pattern-delete.component';
import {IndexPatternListComponent} from './index-pattern/index-pattern-list/index-pattern-list.component';
import {IndexPatternHelpComponent} from './index-pattern/shared/components/index-pattern-help/index-pattern-help.component';
import {AppManagementSidebarComponent} from './layout/app-management-sidebar/app-management-sidebar.component';
import {MenuCardComponent} from './menu/menu-card/menu-card.component';
import {MenuDeleteDialogComponent} from './menu/menu-delete/menu-delete-dialog.component';
import {MenuComponent} from './menu/menu.component';
import {RolloverConfigComponent} from './rollover-config/rollover-config.component';
import {AppManagementSharedModule} from './shared/app-management-shared.module';
import {UtmApiDocComponent} from './utm-api-doc/utm-api-doc.component';


@NgModule({
  declarations: [
    AppManagementComponent,
    AppManagementSidebarComponent,
    IndexPatternHelpComponent,
    IndexPatternListComponent,
    IndexPatternDeleteComponent,
    HealthChecksComponent,
    HealthDetailComponent,
    AuditsComponent,
    AppMetricsComponent,
    UtmJvmMetricsComponent,
    UtmThreadMetricsComponent,
    UtmThreadDetailComponent,
    UtmSystemMetricsComponent,
    MenuComponent,
    MenuDeleteDialogComponent,
    MenuCardComponent,
    AppLogsComponent,
    IndexManagementComponent,
    IndexDeleteComponent,
    AppConfigComponent,
    RolloverConfigComponent,
    UtmApiDocComponent,
    ConnectionKeyComponent,
    AppThemeComponent,
    AppThemeConfigComponent,
    HealthClusterComponent,
    TokenActivateComponent,
    UtmHttpRequestsPreviewComponent,
    UtmServicesOverviewComponent],
  entryComponents: [
    IndexPatternHelpComponent,
    IndexPatternDeleteComponent,
    HealthDetailComponent,
    MenuDeleteDialogComponent,
    TokenActivateComponent,
    IndexDeleteComponent],
  imports: [
    CommonModule,
    AppManagementRoutingModule,
    UtmSharedModule,
    NgbModule,
    FormsModule,
    NgSelectModule,
    ReactiveFormsModule,
    AppManagementSharedModule,
    ComplianceManagementModule,
    DndModule,
    NgxSortableModule,
    InlineSVGModule,
    NgxGaugeModule,
  ],
  exports: [
    HealthChecksComponent,
  ],
  schemas: [CUSTOM_ELEMENTS_SCHEMA, NO_ERRORS_SCHEMA],
  providers: [NavBehavior, VersionUpdateBehavior, TimezoneFormatService]
})
export class AppManagementModule {
}
