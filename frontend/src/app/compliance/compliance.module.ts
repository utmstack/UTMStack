import {CommonModule} from '@angular/common';
import {CUSTOM_ELEMENTS_SCHEMA, NgModule, NO_ERRORS_SCHEMA} from '@angular/core';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {RouterModule} from '@angular/router';
import {NgbCollapseModule, NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {NgSelectModule} from '@ng-select/ng-select';
import {InfiniteScrollModule} from 'ngx-infinite-scroll';
import {UtmDashboardSharedModule} from '../dashboard/shared/utm-dashboard-shared.module';
import {GraphicBuilderSharedModule} from '../graphic-builder/shared/graphic-builder-shared.module';
import {VisualizationSharedModule} from '../graphic-builder/visualization/visualization-shared.module';
import {LogAnalyzerModule} from '../log-analyzer/log-analyzer.module';
import {UtmSharedModule} from '../shared/utm-shared.module';
import {ComplianceCustomViewComponent} from './compliance-custom-view/compliance-custom-view.component';
import {ComplianceManagementModule} from './compliance-management/compliance-management.module';

import {ComplianceResultViewComponent} from './compliance-result-view/compliance-result-view.component';
import {ComplianceRoutingModule} from './compliance-routing.module';
import {ComplianceResultParamsComponent} from './compliance-templates/compliance-result-params/compliance-result-params.component';
import {ComplianceTemplatesComponent} from './compliance-templates/compliance-templates.component';
import {ComplianceSharedModule} from './shared/compliance-shared.module';
import {ComplianceScheduleComponent} from "./compliance-schedule/compliance-schedule.component";
import {
  DashboardFilterCreateComponent
} from '../graphic-builder/dashboard-builder/dashboard-filter-create/dashboard-filter-create.component';
import {DashboardBuilderModule} from '../graphic-builder/dashboard-builder/dashboard-builder.module';
import {AlertManagementSharedModule} from '../data-management/alert-management/shared/alert-management-shared.module';
import { ComplianceReportViewerComponent } from './compliance-report-viewer/compliance-report-viewer.component';
import {UtmCpStandardComponent} from './shared/components/utm-cp-standard/utm-cp-standard.component';
import { ComplianceReportsViewComponent } from './compliance-reports-view/compliance-reports-view.component';
import {
  ComplianceStatusComponent
} from './compliance-reports-view/components/compliance-status/compliance-status.component';
import {
  ComplianceReportDetailComponent
} from './compliance-reports-view/components/compliance-report-detail/compliance-report-detail.component';
import {
  ComplianceTimeWindowsComponent
} from './compliance-reports-view/components/compliance-time-window/compliance-time-windows.component';
import {TimeWindowsService} from "./shared/components/utm-cp-section/time-windows.service";
import {NgxJsonViewerModule} from "ngx-json-viewer";

@NgModule({
  declarations: [
    ComplianceResultViewComponent,
    ComplianceTemplatesComponent,
    ComplianceResultParamsComponent,
    ComplianceCustomViewComponent,
    ComplianceScheduleComponent,
    ComplianceReportViewerComponent,
    UtmCpStandardComponent,
    ComplianceReportsViewComponent,
    ComplianceStatusComponent,
    ComplianceReportDetailComponent,
    ComplianceTimeWindowsComponent
  ],
  imports: [
    CommonModule,
    ComplianceRoutingModule,
    RouterModule,
    UtmSharedModule,
    InfiniteScrollModule,
    NgSelectModule,
    FormsModule,
    LogAnalyzerModule,
    NgbModule,
    VisualizationSharedModule,
    GraphicBuilderSharedModule,
    ComplianceManagementModule,
    ComplianceSharedModule,
    UtmDashboardSharedModule,
    DashboardBuilderModule,
    NgbCollapseModule,
    AlertManagementSharedModule,
    ReactiveFormsModule,
    NgxJsonViewerModule
  ],
  schemas: [NO_ERRORS_SCHEMA, CUSTOM_ELEMENTS_SCHEMA],
  entryComponents: [
    ComplianceResultParamsComponent,
    DashboardFilterCreateComponent,
    UtmCpStandardComponent],
  exports: [],
  providers: [
    TimeWindowsService
  ]
})
export class ComplianceModule {
}
