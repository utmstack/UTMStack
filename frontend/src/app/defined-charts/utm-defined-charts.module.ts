import {CommonModule} from '@angular/common';
import {CUSTOM_ELEMENTS_SCHEMA, NgModule, NO_ERRORS_SCHEMA} from '@angular/core';
import {FormsModule} from '@angular/forms';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {NgxEchartsModule} from 'ngx-echarts';
import {ActiveDirectorySharedModule} from '../active-directory/shared/active-directory-shared.module';
import {UtmSharedModule} from '../shared/utm-shared.module';
import {ChartAdAdminVsUserComponent} from './components/active-directory/chart-ad-admin-vs-user/chart-ad-admin-vs-user.component';
import {ChartAdEventInTimeComponent} from './components/active-directory/chart-ad-event-in-time/chart-ad-event-in-time.component';
import {ChartAdInactiveAdminComponent} from './components/active-directory/chart-ad-inactive-admin/chart-ad-inactive-admin.component';
import {ChartAdInactiveComponent} from './components/active-directory/chart-ad-inactive/chart-ad-inactive.component';
import {ChartAdMostActiveUserComponent} from './components/active-directory/chart-ad-most-active-user/chart-ad-most-active-user.component';
import {ChartAdPermissionsComponent} from './components/active-directory/chart-ad-permissions/chart-ad-permissions.component';
import {ChartAdQuickInfoComponent} from './components/active-directory/chart-ad-quick-info/chart-ad-quick-info.component';
import {ChartAdUserChangesComponent} from './components/active-directory/chart-ad-user-changes/chart-ad-user-changes.component';
import {ChartAlertByCategoryComponent} from './components/alert/chart-alert-by-category/chart-alert-by-category.component';
import {ChartAlertByStatusComponent} from './components/alert/chart-alert-by-status/chart-alert-by-status.component';
import {ChartAlertDailyWeekComponent} from './components/alert/chart-alert-daily-week/chart-alert-daily-week.component';
import {ChartCommonPieComponent} from './components/common/chart-common-pie/chart-common-pie.component';
import {ChartCommonTableComponent} from './components/common/chart-common-table/chart-common-table.component';
import {DataSourceInputComponent} from './components/data-input/data-source-input/data-source-input.component';
import {ChartEventInTimeComponent} from './components/events/chart-event-in-time/chart-event-in-time.component';

@NgModule({
  declarations: [ChartAlertDailyWeekComponent,
    ChartAlertByStatusComponent,
    ChartAdQuickInfoComponent,
    ChartAdInactiveComponent,
    ChartAdPermissionsComponent,
    ChartCommonTableComponent,
    ChartAlertByCategoryComponent,
    ChartCommonPieComponent,
    ChartEventInTimeComponent,
    ChartAdEventInTimeComponent,
    ChartAdAdminVsUserComponent,
    ChartAdMostActiveUserComponent,
    ChartAdUserChangesComponent,
    ChartAdInactiveAdminComponent,
    DataSourceInputComponent],
  exports: [ChartAlertDailyWeekComponent,
    ChartAlertByStatusComponent,
    ChartAdQuickInfoComponent,
    ChartAdInactiveComponent,
    ChartAdPermissionsComponent,
    ChartCommonTableComponent,
    ChartAlertByCategoryComponent,
    ChartCommonPieComponent,
    ChartEventInTimeComponent,
    ChartAdEventInTimeComponent,
    ChartAdAdminVsUserComponent,
    ChartAdMostActiveUserComponent,
    ChartAdUserChangesComponent,
    ChartAdInactiveAdminComponent, DataSourceInputComponent],
  entryComponents: [ChartAlertDailyWeekComponent,
    ChartAlertByStatusComponent,
    ChartAdQuickInfoComponent,
    ChartAdInactiveComponent,
    ChartAdPermissionsComponent],
  imports: [
    CommonModule,
    UtmSharedModule,
    FormsModule,
    NgbModule,
    NgxEchartsModule,
    ActiveDirectorySharedModule,
  ],
  schemas: [CUSTOM_ELEMENTS_SCHEMA, NO_ERRORS_SCHEMA],
})
export class UtmDefinedChartsModule {
}
