import {CommonModule} from '@angular/common';
import {CUSTOM_ELEMENTS_SCHEMA, NgModule, NO_ERRORS_SCHEMA} from '@angular/core';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {RouterModule} from '@angular/router';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {NgSelectModule} from '@ng-select/ng-select';
import {ResizableModule} from 'angular-resizable-element';
import {Ng2TelInputModule} from 'ng2-tel-input';
import {NgxEchartsModule} from 'ngx-echarts';
import {InfiniteScrollModule} from 'ngx-infinite-scroll';
import {UtmDefinedChartsModule} from '../defined-charts/utm-defined-charts.module';
import {InputClassResolve} from '../shared/util/input-class-resolve';
import {UtmSharedModule} from '../shared/utm-shared.module';
import {ActiveDirectoryRoutingModule} from './active-directory-routing.module';
import {AdComponent} from './active-directory.component';
import {AdDashboardComponent} from './dashboard/active-directory-dashboard/active-directory-dashboard.component';
import {AdUserDetailComponent} from './dashboard/ad-user-detail/ad-user-detail.component';
// tslint:disable-next-line:max-line-length
import {AdNotificationsConfigCreateComponent} from './notifications/ad-notifications-config-create/ad-notifications-config-create.component';
// tslint:disable-next-line:max-line-length
import {AdNotificationsConfigDeleteComponent} from './notifications/ad-notifications-config-delete/ad-notifications-config-delete.component';
// tslint:disable-next-line:max-line-length
import {AdNotificationsConfigListComponent} from './notifications/ad-notifications-config-list/ad-notifications-config-list.component';
import {AdReportsComponent} from './reports/active-directory-reports.component';
import {AdReportCreateComponent} from './reports/ad-report-create/ad-report-create.component';
import {AdReportDeleteComponent} from './reports/ad-report-delete/ad-report-delete.component';
import {AdReportListComponent} from './reports/ad-report-list/ad-report-list.component';
// tslint:disable-next-line:max-line-length
import {AdScheduleCreateComponent} from './reports/schedule/ad-schedule-create/ad-schedule-create.component';
// tslint:disable-next-line:max-line-length
import {AdScheduleDeleteComponent} from './reports/schedule/ad-schedule-delete/ad-schedule-delete.component';
import {AdScheduleListComponent} from './reports/schedule/ad-schedule-list/ad-schedule-list.component';
import {ActiveDirectorySharedModule} from './shared/active-directory-shared.module';
import {AdDetailComponent} from './shared/components/active-directory-detail/active-directory-detail.component';
// tslint:disable-next-line:max-line-length
import {ActiveDirectoryAclComponent} from './shared/components/active-directory-detail/shared/components/active-directory-acl/active-directory-acl.component';
// tslint:disable-next-line:max-line-length
import {ActiveDirectoryFolderAclComponent} from './shared/components/active-directory-detail/shared/components/active-directory-folder-acl/active-directory-folder-acl.component';
// tslint:disable-next-line:max-line-length
import {AdComputerIpsComponent} from './shared/components/active-directory-detail/shared/components/computer/ad-computer-ips/ad-computer-ips.component';
// tslint:disable-next-line:max-line-length
import {AdComputerLocalGroupComponent} from './shared/components/active-directory-detail/shared/components/computer/ad-computer-local-group/ad-computer-local-group.component';
// tslint:disable-next-line:max-line-length
import {AdComputerLocalUsersComponent} from './shared/components/active-directory-detail/shared/components/computer/ad-computer-local-users/ad-computer-local-users.component';
// tslint:disable-next-line:max-line-length
import {AdUserComputerComponent} from './shared/components/active-directory-detail/shared/components/user/ad-user-computer/ad-user-computer.component';
import {ActiveDirectoryFilterComponent} from './shared/components/active-directory-filter/active-directory-filter.component';
import {AdPrivilegesComponent} from './shared/components/active-directory-privileges/active-directory-privileges.component';
import {AdTreeComponent} from './shared/components/active-directory-tree/active-directory-tree.component';
import {AdFolderListComponent} from './shared/components/ad-folder-list/ad-folder-list.component';
import {AdTableTreeSelectedComponent} from './shared/components/ad-table-tree-selected/ad-table-tree-selected.component';
import {UtmMailListComponent} from './shared/components/utm-mail-list/utm-mail-list.component';
import {AdTrackerCreateComponent} from './tracker/ad-tracker-create/ad-tracker-create.component';
import {AdTrackerDeleteComponent} from './tracker/ad-tracker-delete/ad-tracker-delete.component';
import {AdTrackerDetailComponent} from './tracker/ad-tracker-detail/ad-tracker-detail.component';
import {AdTrackerListComponent} from './tracker/ad-tracker-list/ad-tracker-list.component';
import {AdTrackerFilterComponent} from './tracker/shared/components/ad-tracker-filter/ad-tracker-filter.component';
import {AdViewComponent} from './view/active-directory-view/active-directory-view.component';

@NgModule({
  declarations: [
    AdComponent,
    AdDashboardComponent,
    AdPrivilegesComponent,
    AdDetailComponent,
    AdTreeComponent,
    AdViewComponent,
    AdTrackerListComponent,
    AdScheduleListComponent,
    AdScheduleCreateComponent,
    AdScheduleDeleteComponent,
    AdTrackerCreateComponent,
    AdNotificationsConfigCreateComponent,
    AdNotificationsConfigListComponent,
    AdNotificationsConfigDeleteComponent,
    AdReportsComponent,
    AdReportDeleteComponent,
    AdReportListComponent,
    AdTrackerDetailComponent,
    AdTrackerDeleteComponent,
    AdReportCreateComponent,
    AdTableTreeSelectedComponent,
    UtmMailListComponent,
    AdTrackerFilterComponent,
    AdUserDetailComponent,
    ActiveDirectoryFilterComponent,
    ActiveDirectoryAclComponent,
    ActiveDirectoryFolderAclComponent,
    AdComputerLocalGroupComponent,
    AdComputerIpsComponent,
    AdComputerLocalUsersComponent,
    AdFolderListComponent,
    AdUserComputerComponent,
  ],
  imports: [
    CommonModule,
    RouterModule,
    ActiveDirectoryRoutingModule,
    InfiniteScrollModule,
    UtmSharedModule,
    NgbModule,
    NgxEchartsModule,
    ReactiveFormsModule,
    NgSelectModule,
    Ng2TelInputModule,
    FormsModule,
    UtmDefinedChartsModule,
    ActiveDirectorySharedModule,
    ResizableModule,
  ],
  entryComponents: [
    AdTrackerCreateComponent,
    AdNotificationsConfigCreateComponent,
    AdNotificationsConfigDeleteComponent,
    AdNotificationsConfigCreateComponent,
    AdNotificationsConfigDeleteComponent,
    AdScheduleCreateComponent,
    AdScheduleDeleteComponent,
    AdReportDeleteComponent,
    AdTrackerDeleteComponent,
    AdReportCreateComponent
  ],
  schemas: [CUSTOM_ELEMENTS_SCHEMA, NO_ERRORS_SCHEMA],
  providers: [InputClassResolve]
})
export class ActiveDirectoryModule {
}
