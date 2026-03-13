import {CommonModule} from '@angular/common';
import {CUSTOM_ELEMENTS_SCHEMA, NgModule, NO_ERRORS_SCHEMA} from '@angular/core';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {NgSelectModule} from '@ng-select/ng-select';
import {NgxEchartsModule} from 'ngx-echarts';
import {NgxJsonViewerModule} from 'ngx-json-viewer';
import {UtmDashboardSharedModule} from '../../dashboard/shared/utm-dashboard-shared.module';
import {
  AlertManagementSharedModule
} from '../../data-management/alert-management/shared/alert-management-shared.module';
import {GraphicBuilderSharedModule} from '../../graphic-builder/shared/graphic-builder-shared.module';
import {UtmSharedModule} from '../../shared/utm-shared.module';
import {
  ComplianceEvaluationViewDetailComponent
} from '../compliance-evaluation-view/compliance-evaluation-view-detail/compliance-evaluation-view-detail.component';
import {ComplianceEvaluationViewComponent} from '../compliance-evaluation-view/compliance-evaluation-view.component';
import {ComplianceEvaluationsViewComponent} from '../compliance-evaluations-view/compliance-evaluations-view.component';
import {
  ComplianceQueryEvaluationsViewComponent
} from '../compliance-query-evaluations-view/compliance-query-evaluations-view.component';
import {
  ComplianceQueryEvaluationDetailComponent
} from '../compliance-query-evaluations-view/components/compliance-query-evaluation-detail/compliance-query-evaluation-detail.component';
import {ComplianceReportsViewComponent} from '../compliance-reports-view/compliance-reports-view.component';
import {
  CompliancePrintViewComponent
} from '../compliance-reports-view/components/compliance-print-view/compliance-print-view.component';
import {
  ComplianceReportDetailComponent
} from '../compliance-reports-view/components/compliance-report-detail/compliance-report-detail.component';
import {
  ComplianceStatusComponent
} from '../compliance-reports-view/components/compliance-status/compliance-status.component';
import {
  ComplianceTimeWindowsComponent
} from '../compliance-reports-view/components/compliance-time-window/compliance-time-windows.component';
import {ReportApplyNoteComponent} from './components/report-apply-note/report-apply-note.component';
import {
  UtmComplianceQueryListComponent
} from './components/utm-compliance-control-config-create/query-config/utm-compliance-query-list.component';
import {
  UtmComplianceQueryComponent
} from './components/utm-compliance-control-config-create/query-config/utm-compliance-query.component';
import {
  UtmComplianceControlConfigCreateComponent
} from './components/utm-compliance-control-config-create/utm-compliance-control-config-create.component';
import {UtmComplianceCreateComponent} from './components/utm-compliance-create/utm-compliance-create.component';
import {
  UtmComplianceScheduleCreateComponent
} from './components/utm-compliance-schedule-create/utm-compliance-schedule-create.component';
import {
  UtmComplianceScheduleDeleteComponent
} from './components/utm-compliance-schedule-delete/utm-compliance-schedule-delete.component';
import {UtmComplianceSelectComponent} from './components/utm-compliance-select/utm-compliance-select.component';
import {ComplianceTimelineComponent} from './components/utm-compliance-timeline/compliance-timeline.component';
import {UtmCpSectionConfigComponent} from './components/utm-cp-section-config/utm-cp-section-config.component';
import { UtmCpSectionComponent } from './components/utm-cp-section/utm-cp-section.component';
import {UtmCpStSectionSelectComponent} from './components/utm-cp-st-section-select/utm-cp-st-section-select.component';
import {UtmCpStandardCreateComponent} from './components/utm-cp-standard-create/utm-cp-standard-create.component';
import {UtmCpStandardSectionCreateComponent} from './components/utm-cp-standard-section-create/utm-cp-standard-section-create.component';
import {UtmCpStandardSelectComponent} from './components/utm-cp-standard-select/utm-cp-standard-select.component';
import {UtmReportInfoViewComponent} from './components/utm-report-info-view/utm-report-info-view.component';
import {UtmSaveAsComplianceComponent} from './components/utm-save-as-compliance/utm-save-as-compliance.component';
import {ComplianceStatusLabelPipe} from './pipes/compliance-status-label.pipe';

@NgModule({
  declarations: [
    UtmSaveAsComplianceComponent,
    UtmCpStandardSelectComponent,
    UtmCpStSectionSelectComponent,
    UtmCpStandardCreateComponent,
    UtmCpStandardSectionCreateComponent,
    UtmReportInfoViewComponent,
    UtmComplianceCreateComponent,
    UtmComplianceScheduleCreateComponent,
    UtmComplianceSelectComponent,
    UtmComplianceScheduleDeleteComponent,
    UtmCpSectionComponent,
    ReportApplyNoteComponent,
    ComplianceStatusComponent,
    ComplianceReportsViewComponent,
    ComplianceReportDetailComponent,
    ComplianceQueryEvaluationDetailComponent,
    ComplianceTimeWindowsComponent,
    CompliancePrintViewComponent,
    UtmComplianceControlConfigCreateComponent,
    UtmComplianceQueryComponent,
    UtmComplianceQueryListComponent,
    ComplianceEvaluationViewComponent,
    ComplianceEvaluationsViewComponent,
    UtmCpSectionConfigComponent,
    ComplianceTimelineComponent,
    ComplianceQueryEvaluationsViewComponent,
    ComplianceEvaluationViewDetailComponent,
    ComplianceStatusLabelPipe
  ],
  imports: [
    CommonModule,
    UtmSharedModule,
    NgSelectModule,
    FormsModule,
    ReactiveFormsModule,
    NgbModule,
    UtmDashboardSharedModule,
    NgxJsonViewerModule,
    GraphicBuilderSharedModule,
    NgxEchartsModule,
    AlertManagementSharedModule
  ],
  schemas: [NO_ERRORS_SCHEMA, CUSTOM_ELEMENTS_SCHEMA],
  entryComponents: [
    UtmSaveAsComplianceComponent,
    UtmCpStandardSelectComponent,
    UtmCpStSectionSelectComponent,
    UtmCpStandardCreateComponent,
    UtmCpStandardSectionCreateComponent,
    UtmComplianceCreateComponent,
    UtmComplianceScheduleCreateComponent,
    UtmComplianceScheduleDeleteComponent,
    UtmComplianceControlConfigCreateComponent
  ],
    exports: [
        UtmSaveAsComplianceComponent,
        UtmCpStandardSelectComponent,
        UtmCpStSectionSelectComponent,
        UtmCpStandardCreateComponent,
        UtmCpStandardSectionCreateComponent,
        UtmReportInfoViewComponent,
        UtmComplianceScheduleCreateComponent,
        UtmComplianceScheduleDeleteComponent,
        UtmCpSectionComponent,
        ReportApplyNoteComponent,
        ComplianceStatusComponent,
        ComplianceReportsViewComponent,
        ComplianceReportDetailComponent,
        ComplianceQueryEvaluationDetailComponent,
        ComplianceTimeWindowsComponent,
        CompliancePrintViewComponent,
        ComplianceEvaluationViewComponent,
        ComplianceEvaluationsViewComponent,
        UtmCpSectionConfigComponent,
        ComplianceStatusLabelPipe
    ]
})
export class ComplianceSharedModule {
}
